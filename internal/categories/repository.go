package categories

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, in CreateInput) (Category, error) {
	const query = `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, name, created_at
	`

	var c Category
	err := r.db.QueryRowContext(ctx, query, in.Name).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)
	if err != nil {
		return Category{}, err
	}

	return c, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (Category, error) {
	const query = `
		SELECT id, name, created_at
		FROM categories
		WHERE id = $1
	`

	var c Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)
	if err != nil {
		return Category{}, err
	}

	return c, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Category, error) {
	const query = `
		SELECT id, name, created_at
		FROM categories
		ORDER BY name ASC, id ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *Repository) Update(ctx context.Context, id int64, in UpdateInput) (Category, error) {
	const query = `
		UPDATE categories
		SET name = $1
		WHERE id = $2
		RETURNING id, name, created_at
	`

	var c Category
	err := r.db.QueryRowContext(ctx, query, in.Name, id).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)
	if err != nil {
		return Category{}, err
	}

	return c, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM categories WHERE id = $1`

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
