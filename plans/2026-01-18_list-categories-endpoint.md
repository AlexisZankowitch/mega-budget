# Plan: add GET /categories (list)

## Approach
- Extend OpenAPI with list categories endpoint.
- Regenerate oapi-codegen server/types.
- Implement list handler in categories HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with GET /categories.
2) Run `go generate ./internal/api`.
3) Implement `ListCategories` handler.
4) Run curl test for list endpoint.

## Verification
- Curl: GET `/categories` returns 200 with JSON list.

## Rollback
- Revert spec + generated code + handler changes.
