-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS categories (
  id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name       TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
  id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  transaction_date DATE NOT NULL,
  category_id      BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  amount           NUMERIC(12,2) NOT NULL,
  description      TEXT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_date
  ON transactions (transaction_date);

CREATE INDEX IF NOT EXISTS idx_transactions_category
  ON transactions (category_id);

CREATE INDEX IF NOT EXISTS idx_transactions_transaction_date
  ON transactions (transaction_date);

CREATE INDEX IF NOT EXISTS idx_transactions_category_date
  ON transactions (category_id, transaction_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
