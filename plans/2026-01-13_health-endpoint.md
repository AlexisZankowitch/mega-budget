# Plan: Health endpoint + Fx bootstrap

## Approach
- Introduce an Fx-based app skeleton with config and DB wiring.
- Implement `/healthz` to verify DB connectivity using a short timeout.
- Start an HTTP server via Fx lifecycle hooks.

## Steps
1. Add `cmd/megabudget/main.go` with Fx app wiring.
2. Add internal packages for config, DB connection, HTTP server, and health handler.
3. Update `go.mod`/`go.sum` with Fx dependencies.

## Verification
- `go test ./...`
- `go run ./cmd/megabudget` with `DATABASE_URL` or `DEV_DATABASE_URL` set, then `curl -i localhost:8080/healthz`.

## Rollback
- Remove `cmd/megabudget` and `internal` packages.
- Revert `go.mod`/`go.sum` changes.
