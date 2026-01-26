# Plan: Transactions analytics endpoint (monthly spending/income by category)

## Approach
- Extend the OpenAPI spec with a new analytics endpoint that accepts a year and returns monthly totals by category for spending and income.
- Add a repository query that aggregates transactions into monthly sums per category, separated by spending (amount < 0) and income (amount > 0).
- Implement a handler that builds the response shape (months array, row values, column totals, total) and plugs into the existing httpapi handler wiring.
- Add API integration tests similar to existing transaction endpoints.

## Steps
1) Update `internal/api/openapi.yaml` with new endpoint and schemas, then regenerate `internal/api/api.gen.go`.
2) Implement repository aggregation helpers in `internal/transactions`.
3) Add handler logic in `internal/httpapi` and wire into `internal/httpapi/handler.go`.
4) Add integration tests for the new endpoint.

## Verification
- `go test ./internal/httpapi -run Analytics` (or full package tests)
- Manual: `curl "http://localhost:8080/analytics/transactions-summary?year=2025"`

## Rollback
- Remove the endpoint from OpenAPI and delete new handler/repo/test code.
