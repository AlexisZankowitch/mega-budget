# Plan: transactions repository integration tests (testcontainers)

## Approach
- Use `testcontainers-go` Postgres module to spin up an ephemeral DB.
- Run existing goose migrations from `migrations/` to create schema.
- Exercise Create/Get/List/Update/Delete and verify cents conversion and not-found behavior.

## Steps
1) Add integration test file under `internal/transactions`.
2) Implement container setup + schema creation + seed category helper.
3) Add CRUD tests using the repository.

## Verification
- `go test ./...`

## Rollback
- Remove `internal/transactions/*_test.go` and this plan file.
