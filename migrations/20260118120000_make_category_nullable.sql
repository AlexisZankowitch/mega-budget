-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_category_id_fkey;

ALTER TABLE transactions
  ALTER COLUMN category_id DROP NOT NULL;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_category_id_fkey
  FOREIGN KEY (category_id)
  REFERENCES categories(id)
  ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions
  DROP CONSTRAINT IF EXISTS transactions_category_id_fkey;

ALTER TABLE transactions
  ALTER COLUMN category_id SET NOT NULL;

ALTER TABLE transactions
  ADD CONSTRAINT transactions_category_id_fkey
  FOREIGN KEY (category_id)
  REFERENCES categories(id)
  ON DELETE RESTRICT;
-- +goose StatementEnd
