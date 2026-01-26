package httpapi_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"zankowitch.com/go-db-app/internal/categories"
	"zankowitch.com/go-db-app/internal/config"
	"zankowitch.com/go-db-app/internal/handlers"
	"zankowitch.com/go-db-app/internal/httpapi"
	"zankowitch.com/go-db-app/internal/httpserver"
	"zankowitch.com/go-db-app/internal/logging"
	"zankowitch.com/go-db-app/internal/transactions"
)

var (
	testClient *http.Client
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	db, cleanup := setupTestDB()
	defer cleanup()

	server := startTestServer(db)
	testServer = server
	testClient = server.Client()
	defer server.Close()

	os.Exit(m.Run())
}

func startTestServer(db *sql.DB) *httptest.Server {
	logger, err := logging.New()
	if err != nil {
		panic(err)
	}

	health := handlers.NewHealthHandler(db, config.Config{HealthTimeout: 2 * time.Second})
	txRepo := transactions.NewRepository(db)
	catRepo := categories.NewRepository(db)
	txHandler := httpapi.NewTransactionsHandler(txRepo, logger)
	catHandler := httpapi.NewCategoriesHandler(catRepo, logger)
	analyticsHandler := httpapi.NewAnalyticsHandler(txRepo, catRepo, logger)
	apiHandler := httpapi.NewHandler(txHandler, catHandler, analyticsHandler)

	mux, err := httpserver.NewMux(health, apiHandler)
	if err != nil {
		panic(err)
	}

	handler := logging.RequestIDAndLogger(logger)(mux)
	return httptest.NewServer(handler)
}

func setupTestDB() (*sql.DB, func()) {
	ctx := context.Background()
	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.BasicWaitStrategies(),
		postgres.WithDatabase("megabudget_test"),
		postgres.WithUsername("megabudget_app"),
		postgres.WithPassword("megabudget_pass"),
	)
	if err != nil {
		panic(err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		panic(err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		_ = container.Terminate(ctx)
		panic(err)
	}

	if err := runMigrations(ctx, db); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		panic(err)
	}

	cleanup := func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}

	return db, cleanup
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetDialect("postgres")
	goose.SetBaseFS(os.DirFS(migrationsDir()))
	return goose.UpContext(ctx, db, ".")
}

func migrationsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "migrations"
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "migrations"))
}

func doRequest(t *testing.T, method, url string, body []byte) *http.Response {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := testClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
