# Plan: categories repository CRUD

## Approach
- Add `internal/categories` package with model, repository, and `ErrNotFound`.
- Use plain SQL with `database/sql`.
- Add testcontainers integration test using existing migrations.

## Steps
1) Create `internal/categories/model.go`, `errors.go`, `repository.go`.
2) Add `internal/categories/repository_integration_test.go` with subtests for CRUD.

## Verification
- `go test ./internal/categories`

## Rollback
- Remove `internal/categories/` and this plan file.
