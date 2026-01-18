# Plan: add start-date filter to transaction pagination

## Approach
- Extend transactions `ListAfter` to accept an optional `startDate` filter.
- Apply `transaction_date <= startDate` before keyset cursor logic.
- Add a small integration test to validate the filter.

## Steps
1) Update repository method signature and query builder.
2) Adjust integration tests to pass the new argument and verify behavior.

## Verification
- `go test ./internal/transactions`

## Rollback
- Revert the `ListAfter` signature and test changes.
