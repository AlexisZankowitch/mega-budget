# Plan: add HTTP integration tests with testcontainers

## Approach
- Spin up Postgres with testcontainers and run goose migrations.
- Start an httptest server with the real mux, validation, and logging.
- Add CRUD tests for transactions and categories using t.Run subtests.

## Steps
1) Add `internal/httpserver/api_integration_test.go` with shared setup helpers.
2) Implement transactions and categories CRUD tests with named subtests.
3) Run `go test ./internal/httpserver`.

## Verification
- `go test ./internal/httpserver`

## Rollback
- Remove the test file and this plan.
