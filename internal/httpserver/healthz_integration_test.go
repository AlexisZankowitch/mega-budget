package httpserver_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"zankowitch.com/go-db-app/internal/config"
	"zankowitch.com/go-db-app/internal/handlers"
	"zankowitch.com/go-db-app/internal/httpserver"
)

const (
	toxiproxyImage = "ghcr.io/shopify/toxiproxy:2.5.0"
	postgresImage  = "postgres:latest"
	proxyPort      = "8666"
)

type toxiPostgres struct {
	postgres *postgres.PostgresContainer
	toxi     testcontainers.Container
	proxy    *toxiproxy.Proxy
	host     string
	port     string
}

func TestHealthzEndpointReportsDbAvailability(t *testing.T) {
	ctx := context.Background()
	tp := startPostgresBehindToxi(ctx, t)

	dbURL := fmt.Sprintf(
		"postgres://postgres:mysecretpassword@%s:%s/postgres?sslmode=disable",
		tp.host,
		tp.port,
	)
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	handler := handlers.NewHealthHandler(db, config.Config{HealthTimeout: 2 * time.Second})
	mux := httpserver.NewMux(handler)
	server := httptest.NewServer(mux)
	client := &http.Client{Timeout: 5 * time.Second}
	endpoint := server.URL + "/healthz"

	t.Cleanup(server.Close)

	waitForStatus(t, client, endpoint, http.StatusOK, "ok")

	if err := tp.proxy.Disable(); err != nil {
		t.Fatal(err)
	}

	waitForStatus(t, client, endpoint, http.StatusServiceUnavailable, "")

	if err := tp.proxy.Enable(); err != nil {
		t.Fatal(err)
	}

	waitForStatus(t, client, endpoint, http.StatusOK, "ok")
}

func startPostgresBehindToxi(ctx context.Context, t *testing.T) *toxiPostgres {
	t.Helper()

	toxiContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        toxiproxyImage,
			ExposedPorts: []string{"8474/tcp", proxyPort + "/tcp"},
			WaitingFor:   wait.ForHTTP("/version").WithPort("8474/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = toxiContainer.Terminate(ctx)
	})

	pgContainer, err := postgres.Run(
		ctx,
		postgresImage,
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("mysecretpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	toxiHost, err := toxiContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	apiPort, err := toxiContainer.MappedPort(ctx, "8474/tcp")
	if err != nil {
		t.Fatal(err)
	}

	pgIP, err := pgContainer.ContainerIP(ctx)
	if err != nil {
		t.Fatal(err)
	}

	proxyClient := toxiproxy.NewClient(fmt.Sprintf("%s:%s", toxiHost, apiPort.Port()))
	proxy, err := proxyClient.CreateProxy(
		fmt.Sprintf("postgres-%d", time.Now().UnixNano()),
		"0.0.0.0:"+proxyPort,
		fmt.Sprintf("%s:5432", pgIP),
	)
	if err != nil {
		t.Fatal(err)
	}

	mappedProxyPort, err := toxiContainer.MappedPort(ctx, proxyPort+"/tcp")
	if err != nil {
		t.Fatal(err)
	}

	return &toxiPostgres{
		postgres: pgContainer,
		toxi:     toxiContainer,
		proxy:    proxy,
		host:     toxiHost,
		port:     mappedProxyPort.Port(),
	}
}

func waitForStatus(t *testing.T, client *http.Client, url string, status int, body string) {
	t.Helper()

	var lastStatus int
	var lastBody string
	var lastErr error

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			return
		}

		payload, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		lastStatus = resp.StatusCode
		lastBody = strings.TrimSpace(string(payload))

		assert.Equal(c, status, resp.StatusCode)
		if body != "" {
			assert.Equal(c, body, lastBody)
		}
	}, 30*time.Second, 200*time.Millisecond, "expected status %d from %s (body %q). last status: %d body: %q last error: %v", status, url, body, lastStatus, lastBody, lastErr)
}
