# Plan: Healthz integration test (testcontainers + toxiproxy)

## Approach
- Add an integration test that spins up Postgres and Toxiproxy containers, wires the app's DB connection through the proxy, and exercises `/healthz` over HTTP.
- Verify healthy (200/`ok`), unhealthy (503) when proxy disabled, and healthy again when re-enabled.
- Use direct `httptest` server with `handlers.NewHealthHandler` and `httpserver.NewMux` to keep the test focused.

## Steps
1) Add a new integration test under `internal/httpserver/` that starts Postgres + Toxiproxy with testcontainers.
2) Use the proxy endpoint in the DB URL, start the HTTP server, and assert status transitions (200 -> 503 -> 200).
3) Update `go.mod`/`go.sum` with testcontainers and toxiproxy dependencies.

## Verification
- `go test ./...`

## Rollback
- Remove the new test file and revert `go.mod`/`go.sum` changes.
