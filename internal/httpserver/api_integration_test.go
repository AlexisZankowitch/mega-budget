package httpserver_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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
	categorieshttp "zankowitch.com/go-db-app/internal/categories/http"
	"zankowitch.com/go-db-app/internal/config"
	"zankowitch.com/go-db-app/internal/handlers"
	"zankowitch.com/go-db-app/internal/httpapi"
	"zankowitch.com/go-db-app/internal/httpserver"
	"zankowitch.com/go-db-app/internal/logging"
	"zankowitch.com/go-db-app/internal/transactions"
	transactionshttp "zankowitch.com/go-db-app/internal/transactions/http"
)

type transactionResponse struct {
	ID              int64   `json:"id"`
	TransactionDate string  `json:"transaction_date"`
	CategoryID      *int64  `json:"category_id"`
	AmountCents     int64   `json:"amount_cents"`
	Description     *string `json:"description"`
	CreatedAt       string  `json:"created_at"`
}

type transactionListResponse struct {
	Items []transactionResponse `json:"items"`
}

type categoryResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type categoryListResponse struct {
	Items []categoryResponse `json:"items"`
}

func TestTransactionsHTTPCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	server := startTestServer(t, db)
	t.Cleanup(server.Close)

	client := server.Client()

	// Subtests share state and must not be run in isolation or parallel.
	var created transactionResponse

	t.Run("create transaction", func(t *testing.T) {
		body := []byte(`{"transaction_date":"2026-02-01","amount_cents":-1257,"description":"lunch"}`)
		resp := doRequest(t, client, http.MethodPost, server.URL+"/transactions", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Fatalf("missing X-Request-ID header")
		}

		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if created.ID == 0 {
			t.Fatalf("expected id to be set")
		}
	})

	t.Run("get transaction", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodGet, server.URL+"/transactions/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("list transactions", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodGet, server.URL+"/transactions", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var list transactionListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(list.Items) == 0 {
			t.Fatalf("expected at least one transaction")
		}
	})

	t.Run("update transaction", func(t *testing.T) {
		body := []byte(`{"transaction_date":"2026-02-02","amount_cents":-2000,"description":"updated"}`)
		resp := doRequest(t, client, http.MethodPut, server.URL+"/transactions/"+itoa(created.ID), body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("delete transaction", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodDelete, server.URL+"/transactions/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", resp.StatusCode)
		}
	})

	t.Run("get after delete returns not found", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodGet, server.URL+"/transactions/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
	})
}

func TestCategoriesHTTPCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	server := startTestServer(t, db)
	t.Cleanup(server.Close)

	client := server.Client()

	// Subtests share state and must not be run in isolation or parallel.
	var created categoryResponse

	t.Run("create category", func(t *testing.T) {
		name := "Category-" + time.Now().UTC().Format("150405.000000000")
		body := []byte(`{"name":"` + name + `"}`)
		resp := doRequest(t, client, http.MethodPost, server.URL+"/categories", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Fatalf("missing X-Request-ID header")
		}

		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if created.ID == 0 {
			t.Fatalf("expected id to be set")
		}
	})

	t.Run("get category", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodGet, server.URL+"/categories/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("list categories", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodGet, server.URL+"/categories", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		var list categoryListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(list.Items) == 0 {
			t.Fatalf("expected at least one category")
		}
	})

	t.Run("update category", func(t *testing.T) {
		body := []byte(`{"name":"Updated Category"}`)
		resp := doRequest(t, client, http.MethodPut, server.URL+"/categories/"+itoa(created.ID), body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("delete category", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodDelete, server.URL+"/categories/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", resp.StatusCode)
		}
	})

	t.Run("get after delete returns not found", func(t *testing.T) {
		resp := doRequest(t, client, http.MethodGet, server.URL+"/categories/"+itoa(created.ID), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
	})
}

func startTestServer(t *testing.T, db *sql.DB) *httptest.Server {
	t.Helper()

	logger, err := logging.New()
	if err != nil {
		t.Fatalf("logger: %v", err)
	}

	health := handlers.NewHealthHandler(db, config.Config{HealthTimeout: 2 * time.Second})
	txRepo := transactions.NewRepository(db)
	catRepo := categories.NewRepository(db)
	txHandler := transactionshttp.NewHandler(txRepo, logger)
	catHandler := categorieshttp.NewHandler(catRepo, logger)
	apiHandler := httpapi.NewHandler(txHandler, catHandler)

	mux, err := httpserver.NewMux(health, apiHandler)
	if err != nil {
		t.Fatalf("new mux: %v", err)
	}

	handler := logging.RequestIDAndLogger(logger)(mux)
	return httptest.NewServer(handler)
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

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
		t.Fatalf("start container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("open db: %v", err)
	}

	if err := runMigrations(ctx, db); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("run migrations: %v", err)
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

func doRequest(t *testing.T, client *http.Client, method, url string, body []byte) *http.Response {
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

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
