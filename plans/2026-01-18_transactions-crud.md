# Plan: transactions repository CRUD

## Approach
- Add a small `internal/transactions` package with a model and repository using `database/sql` and plain SQL.
- Represent money as `int64` cents in Go and convert to/from `NUMERIC(12,2)` in SQL.
- Provide basic CRUD methods (Create/Get/List/Update/Delete) and a sentinel `ErrNotFound`.

## Steps
1) Create `internal/transactions/model.go` with structs using `AmountCents int64` and `time.Time`.
2) Create `internal/transactions/repository.go` implementing SQL CRUD with amount conversion.
3) Add `internal/transactions/errors.go` with `ErrNotFound`.

## Verification
- `go test ./...`

## Rollback
- Remove `internal/transactions/` and this plan file.
