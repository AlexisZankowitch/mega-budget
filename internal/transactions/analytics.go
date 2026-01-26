package transactions

import "context"

type MonthlyCategoryTotal struct {
	CategoryID  int64
	Month       int
	AmountCents int64
}

func (r *Repository) ListMonthlySpendingByCategory(ctx context.Context, year int) ([]MonthlyCategoryTotal, error) {
	const query = `
		SELECT
			category_id,
			EXTRACT(MONTH FROM transaction_date)::int AS month,
			(SUM(CASE WHEN amount < 0 THEN -amount ELSE 0 END) * 100)::bigint AS amount_cents
		FROM transactions
		WHERE category_id IS NOT NULL
			AND transaction_date >= make_date($1, 1, 1)
			AND transaction_date < make_date($1 + 1, 1, 1)
		GROUP BY category_id, month
		HAVING SUM(CASE WHEN amount < 0 THEN -amount ELSE 0 END) <> 0
		ORDER BY category_id, month
	`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]MonthlyCategoryTotal, 0)
	for rows.Next() {
		var row MonthlyCategoryTotal
		if err := rows.Scan(&row.CategoryID, &row.Month, &row.AmountCents); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListMonthlyIncomeByCategory(ctx context.Context, year int) ([]MonthlyCategoryTotal, error) {
	const query = `
		SELECT
			category_id,
			EXTRACT(MONTH FROM transaction_date)::int AS month,
			(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) * 100)::bigint AS amount_cents
		FROM transactions
		WHERE category_id IS NOT NULL
			AND transaction_date >= make_date($1, 1, 1)
			AND transaction_date < make_date($1 + 1, 1, 1)
		GROUP BY category_id, month
		HAVING SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) <> 0
		ORDER BY category_id, month
	`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]MonthlyCategoryTotal, 0)
	for rows.Next() {
		var row MonthlyCategoryTotal
		if err := rows.Scan(&row.CategoryID, &row.Month, &row.AmountCents); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
