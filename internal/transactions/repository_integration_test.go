package transactions

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestRepositoryCRUD(t *testing.T) {
	t.Parallel()

	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	repo := NewRepository(db)
	ctx := context.Background()

	date := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)

	// Subtests share state and must not be run in isolation or parallel.
	var created Transaction

	t.Run("create transaction", func(t *testing.T) {
		var err error
		created, err = repo.Create(ctx, CreateInput{
			TransactionDate: date,
			CategoryID:      nil,
			AmountCents:     -1257,
			Description:     strPtr("lunch"),
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.AmountCents != -1257 {
			t.Fatalf("create amount cents = %d, want -1257", created.AmountCents)
		}
	})

	t.Run("read transaction", func(t *testing.T) {
		got, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("get id = %d, want %d", got.ID, created.ID)
		}
		if !got.TransactionDate.Equal(date) {
			t.Fatalf("get date = %v, want %v", got.TransactionDate, date)
		}
	})

	t.Run("read all transactions", func(t *testing.T) {
		_, err := repo.Create(ctx, CreateInput{
			TransactionDate: date.AddDate(0, 0, -1),
			CategoryID:      nil,
			AmountCents:     2500,
			Description:     nil,
		})
		if err != nil {
			t.Fatalf("create second: %v", err)
		}

		list, err := repo.ListAfter(ctx, 10, nil, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("list len = %d, want 2", len(list))
		}
	})

	t.Run("update transaction", func(t *testing.T) {
		updated, err := repo.Update(ctx, created.ID, UpdateInput{
			TransactionDate: date,
			CategoryID:      nil,
			AmountCents:     -999,
			Description:     strPtr("updated"),
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.AmountCents != -999 {
			t.Fatalf("update amount cents = %d, want -999", updated.AmountCents)
		}
	})

	t.Run("read next page with cursor", func(t *testing.T) {
		list, err := repo.ListAfter(ctx, 1, nil, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("list first page: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("list first page len = %d, want 1", len(list))
		}

		cursorDate := list[0].TransactionDate
		cursorID := list[0].ID
		next, err := repo.ListAfter(ctx, 10, nil, nil, nil, nil, &cursorDate, &cursorID)
		if err != nil {
			t.Fatalf("list next page: %v", err)
		}
		if len(next) != 1 {
			t.Fatalf("list next page len = %d, want 1", len(next))
		}
	})

	t.Run("read from start date", func(t *testing.T) {
		startDate := date.AddDate(0, 0, -1)
		list, err := repo.ListAfter(ctx, 10, nil, &startDate, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("list from date: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("list from date len = %d, want 1", len(list))
		}
		if !list[0].TransactionDate.Equal(startDate) {
			t.Fatalf("list from date got %v, want %v", list[0].TransactionDate, startDate)
		}
	})

	t.Run("delete transaction", func(t *testing.T) {
		if err := repo.Delete(ctx, created.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
	})

	t.Run("read after delete returns not found", func(t *testing.T) {
		_, err := repo.Get(ctx, created.ID)
		if err != sql.ErrNoRows {
			t.Fatalf("get after delete = %v, want sql.ErrNoRows", err)
		}
	})
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

func strPtr(v string) *string {
	return &v
}
