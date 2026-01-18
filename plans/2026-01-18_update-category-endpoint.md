# Plan: add PUT /categories/{categoryId}

## Approach
- Extend OpenAPI with update category endpoint.
- Regenerate oapi-codegen server/types.
- Implement update handler in categories HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with PUT /categories/{id} and schema.
2) Run `go generate ./internal/api`.
3) Implement `UpdateCategory` handler.
4) Run curl test to create then update a category.

## Verification
- Curl: POST then PUT `/categories/{id}` returns 200.

## Rollback
- Revert spec + generated code + handler changes.
