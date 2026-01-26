# Plan: Monthly savings analytics endpoint

## Approach
- Add a new analytics endpoint that returns monthly net savings (income - spending) for a given year.
- Extend the OpenAPI spec with a minimal response schema.
- Implement repository aggregation (sum of amounts per month in cents).
- Implement handler that builds months array, monthly values, and total.
- Add integration test using a future year to avoid seed data collisions.

## Steps
1) Update `internal/api/openapi.yaml` and regenerate `internal/api/api.gen.go`.
2) Add repository query for monthly net totals in `internal/transactions`.
3) Add handler + wiring in `internal/httpapi`.
4) Add integration test in `internal/httpapi`.

## Verification
- `go test ./internal/httpapi`
- Manual: `curl "http://localhost:8080/analytics/monthly-savings?year=2030"`

## Rollback
- Remove the endpoint and delete added repo/handler/test code.
