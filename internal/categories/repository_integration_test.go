package categories

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

	// Subtests share state and must not be run in isolation or parallel.
	var created Category
	baseName := uniqueName("Category")
	updatedName := baseName + "-updated"
	secondName := baseName + "-second"

	t.Run("create category", func(t *testing.T) {
		var err error
		created, err = repo.Create(ctx, CreateInput{Name: baseName})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.Name != baseName {
			t.Fatalf("create name = %q, want %q", created.Name, baseName)
		}
	})

	t.Run("read category", func(t *testing.T) {
		got, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("get id = %d, want %d", got.ID, created.ID)
		}
		if got.Name != baseName {
			t.Fatalf("get name = %q, want %q", got.Name, baseName)
		}
	})

	t.Run("read all categories", func(t *testing.T) {
		_, err := repo.Create(ctx, CreateInput{Name: secondName})
		if err != nil {
			t.Fatalf("create second: %v", err)
		}

		list, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) < 2 {
			t.Fatalf("list len = %d, want at least 2", len(list))
		}
		if !containsCategoryName(list, baseName) || !containsCategoryName(list, secondName) {
			t.Fatalf("list does not include expected categories")
		}
	})

	t.Run("update category", func(t *testing.T) {
		updated, err := repo.Update(ctx, created.ID, UpdateInput{Name: updatedName})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Name != updatedName {
			t.Fatalf("update name = %q, want %q", updated.Name, updatedName)
		}
	})

	t.Run("delete category", func(t *testing.T) {
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

func uniqueName(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("20060102150405.000000000")
}

func containsCategoryName(list []Category, name string) bool {
	for _, c := range list {
		if c.Name == name {
			return true
		}
	}
	return false
}
