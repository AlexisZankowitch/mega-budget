# Plan: add POST /categories

## Approach
- Extend OpenAPI with create category endpoint.
- Regenerate oapi-codegen server/types.
- Implement create handler in categories HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with POST /categories and schemas.
2) Run `go generate ./internal/api`.
3) Implement `CreateCategory` handler.
4) Run curl test to create category.

## Verification
- Curl: POST `/categories` returns 201 with JSON body.

## Rollback
- Revert spec + generated code + handler changes.
