# Plan: add PUT /transactions/{transactionId}

## Approach
- Extend OpenAPI with update transaction endpoint.
- Regenerate oapi-codegen server/types.
- Implement update handler in transactions HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with PUT transaction path + responses.
2) Run `go generate ./internal/api`.
3) Implement `UpdateTransaction` handler.
4) Run curl test to create then update a transaction.

## Verification
- Curl: POST then PUT `/transactions/{id}` returns 200 with updated body.

## Rollback
- Revert spec + generated code + handler changes.
