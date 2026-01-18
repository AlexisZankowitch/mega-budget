# Plan: add GET /transactions (list)

## Approach
- Extend OpenAPI with list transactions endpoint + query params.
- Regenerate oapi-codegen server/types.
- Implement list handler in transactions HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with GET /transactions and params.
2) Run `go generate ./internal/api`.
3) Implement `ListTransactions` handler.
4) Run curl test for list endpoint.

## Verification
- Curl: GET `/transactions` returns 200 with JSON list.

## Rollback
- Revert spec + generated code + handler changes.
