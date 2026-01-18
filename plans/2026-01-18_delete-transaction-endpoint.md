# Plan: add DELETE /transactions/{transactionId}

## Approach
- Extend OpenAPI with delete transaction endpoint.
- Regenerate oapi-codegen server/types.
- Implement delete handler in transactions HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with DELETE transaction path + responses.
2) Run `go generate ./internal/api`.
3) Implement `DeleteTransaction` handler.
4) Run curl test to create then delete a transaction.

## Verification
- Curl: POST then DELETE `/transactions/{id}` returns 204.

## Rollback
- Revert spec + generated code + handler changes.
