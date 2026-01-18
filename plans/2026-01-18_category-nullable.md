# Plan: allow transactions without categories

## Approach
- Add a goose migration to make `transactions.category_id` nullable and adjust the FK to `ON DELETE SET NULL`.

## Steps
1) Create a new migration altering the FK constraint and dropping NOT NULL on `category_id`.
2) Provide a down migration to restore NOT NULL and `ON DELETE RESTRICT`.

## Verification
- `goose -dir migrations postgres "$DEV_DATABASE_URL" status`
- `goose -dir migrations postgres "$DEV_DATABASE_URL" up`

## Rollback
- `goose -dir migrations postgres "$DEV_DATABASE_URL" down`
