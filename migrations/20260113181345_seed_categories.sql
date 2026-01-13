-- +goose Up
-- +goose StatementBegin
INSERT INTO categories (name) VALUES
  ('Groceries'),
  ('Rent'),
  ('Restaurants'),
  ('Transport'),
  ('Utilities'),
  ('Subscriptions'),
  ('Health'),
  ('Shopping'),
  ('Travel'),
  ('Other')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM categories
WHERE name IN (
  'Groceries','Rent','Restaurants','Transport','Utilities',
  'Subscriptions','Health','Shopping','Travel','Other'
);
-- +goose StatementEnd
