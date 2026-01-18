# Plan: add GET /categories/{categoryId}

## Approach
- Extend OpenAPI with get category endpoint.
- Regenerate oapi-codegen server/types.
- Implement get handler in categories HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with GET /categories/{id}.
2) Run `go generate ./internal/api`.
3) Implement `GetCategory` handler.
4) Run curl test to create then get a category.

## Verification
- Curl: POST then GET `/categories/{id}` returns 200.

## Rollback
- Revert spec + generated code + handler changes.
