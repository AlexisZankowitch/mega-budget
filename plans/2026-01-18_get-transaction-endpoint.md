# Plan: add GET /transactions/{transactionId}

## Approach
- Extend OpenAPI with get transaction endpoint.
- Regenerate oapi-codegen server/types.
- Implement get handler in transactions HTTP package.
- Verify via curl against running server.

## Steps
1) Update `internal/api/openapi.yaml` with GET transaction path + responses.
2) Run `go generate ./internal/api`.
3) Implement `GetTransaction` handler.
4) Run curl test to create then get a transaction.

## Verification
- Curl: POST then GET `/transactions/{id}` returns 200 with JSON body.

## Rollback
- Revert spec + generated code + handler changes.
