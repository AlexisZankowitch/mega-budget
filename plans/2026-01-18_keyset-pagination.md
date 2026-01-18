# Plan: replace offset pagination with keyset pagination

## Approach
- Replace transactions `List` with `ListAfter` for keyset pagination.
- Use `(transaction_date, id)` tuple cursor.
- Update integration tests to call the new method.

## Steps
1) Update transactions repository to add `ListAfter` and remove `List`.
2) Adjust transactions integration tests to use `ListAfter`.

## Verification
- `go test ./internal/transactions`
- `go test ./internal/categories`

## Rollback
- Restore prior `List` implementations and tests.
