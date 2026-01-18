# Plan: add DELETE /categories/{categoryId}

## Approach
- Extend OpenAPI with delete category endpoint.
- Regenerate oapi-codegen server/types.
- Implement delete handler in categories HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with DELETE /categories/{id}.
2) Run `go generate ./internal/api`.
3) Implement `DeleteCategory` handler.
4) Run curl test to create then delete a category.

## Verification
- Curl: POST then DELETE `/categories/{id}` returns 204.

## Rollback
- Revert spec + generated code + handler changes.
