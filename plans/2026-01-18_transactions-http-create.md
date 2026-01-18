# Plan: add OpenAPI + create transaction endpoint (stdlib net/http)

## Approach
- Add an OpenAPI spec with a `POST /transactions` endpoint and related schemas.
- Use oapi-codegen (strict server, stdlib) to generate server interface + models.
- Implement a transactions HTTP handler that maps the request to the repository.
- Wire the handler + request validation middleware into the existing httpserver mux.

## Steps
1) Create `internal/api/openapi.yaml` and `internal/api/oapi-codegen.yaml` plus `go:generate`.
2) Generate `internal/api/api.gen.go`.
3) Implement `internal/transactions/http/handler.go` for CreateTransaction.
4) Update DI + mux wiring to register the generated handler and validator.

## Verification
- `go test ./...`

## Rollback
- Remove `internal/api`, new handler, and DI/mux wiring changes.
