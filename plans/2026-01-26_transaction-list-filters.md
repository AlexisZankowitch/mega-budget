# Plan: Transaction list filters for month/category/type

## Approach
- Extend the list transactions endpoint to accept from/to date filters, category_id, and type (spending/income).
- Keep backward compatibility with existing start_date/after cursor; map start_date to to_date when to_date is absent.
- Update repository list query to include optional filters.
- Add integration tests covering month/category/type filters and start_date mapping.

## Steps
1) Update `internal/api/openapi.yaml` with new query parameters and regenerate `internal/api/api.gen.go`.
2) Update transactions repository and handler to accept/apply filters.
3) Add integration tests for filtered listing.

## Verification
- `go test ./internal/httpapi`

## Rollback
- Remove new params and revert repo/handler/test changes.
