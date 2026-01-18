package categories

import "time"

type Category struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type CreateInput struct {
	Name string
}

type UpdateInput struct {
	Name string
}
