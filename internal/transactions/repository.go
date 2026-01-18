package transactions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, in CreateInput) (Transaction, error) {
	const query = `
		INSERT INTO transactions (transaction_date, category_id, amount, description)
		VALUES ($1, $2, $3::numeric / 100, $4)
		RETURNING id, transaction_date, category_id, (amount * 100)::bigint, description, created_at
	`

	var t Transaction
	var categoryID sql.NullInt64
	err := r.db.QueryRowContext(
		ctx,
		query,
		in.TransactionDate,
		in.CategoryID,
		in.AmountCents,
		in.Description,
	).Scan(
		&t.ID,
		&t.TransactionDate,
		&categoryID,
		&t.AmountCents,
		&t.Description,
		&t.CreatedAt,
	)
	if err != nil {
		return Transaction{}, err
	}

	t.CategoryID = nullableInt64(categoryID)

	return t, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (Transaction, error) {
	const query = `
		SELECT id, transaction_date, category_id, (amount * 100)::bigint, description, created_at
		FROM transactions
		WHERE id = $1
	`

	var t Transaction
	var categoryID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.TransactionDate,
		&categoryID,
		&t.AmountCents,
		&t.Description,
		&t.CreatedAt,
	)
	if err != nil {
		return Transaction{}, err
	}

	t.CategoryID = nullableInt64(categoryID)

	return t, nil
}

func (r *Repository) ListAfter(ctx context.Context, limit int, startDate *time.Time, afterDate *time.Time, afterID *int64) ([]Transaction, error) {
	const baseQuery = `
		SELECT id, transaction_date, category_id, (amount * 100)::bigint, description, created_at
		FROM transactions
	`

	clauses := make([]string, 0, 2)
	args := make([]any, 0, 4)

	if startDate != nil {
		clauses = append(clauses, fmt.Sprintf("transaction_date <= $%d", len(args)+1))
		args = append(args, *startDate)
	}
	if afterDate != nil && afterID != nil {
		clauses = append(clauses, fmt.Sprintf("(transaction_date, id) < ($%d, $%d)", len(args)+1, len(args)+2))
		args = append(args, *afterDate, *afterID)
	}

	query := baseQuery
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY transaction_date DESC, id DESC LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var t Transaction
		var categoryID sql.NullInt64
		if err := rows.Scan(
			&t.ID,
			&t.TransactionDate,
			&categoryID,
			&t.AmountCents,
			&t.Description,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		t.CategoryID = nullableInt64(categoryID)
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *Repository) Update(ctx context.Context, id int64, in UpdateInput) (Transaction, error) {
	const query = `
		UPDATE transactions
		SET transaction_date = $1,
			category_id = $2,
			amount = $3::numeric / 100,
			description = $4
		WHERE id = $5
		RETURNING id, transaction_date, category_id, (amount * 100)::bigint, description, created_at
	`

	var t Transaction
	var categoryID sql.NullInt64
	err := r.db.QueryRowContext(
		ctx,
		query,
		in.TransactionDate,
		in.CategoryID,
		in.AmountCents,
		in.Description,
		id,
	).Scan(
		&t.ID,
		&t.TransactionDate,
		&categoryID,
		&t.AmountCents,
		&t.Description,
		&t.CreatedAt,
	)
	if err != nil {
		return Transaction{}, err
	}

	t.CategoryID = nullableInt64(categoryID)

	return t, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM transactions WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func nullableInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}
