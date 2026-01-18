package transactions

import "time"

// AmountCents represents monetary values in cents (can be negative).
type Transaction struct {
	ID              int64
	TransactionDate time.Time
	CategoryID      *int64
	AmountCents     int64
	Description     *string
	CreatedAt       time.Time
}

type CreateInput struct {
	TransactionDate time.Time
	CategoryID      *int64
	AmountCents     int64
	Description     *string
}

type UpdateInput struct {
	TransactionDate time.Time
	CategoryID      *int64
	AmountCents     int64
	Description     *string
}
