package db

import (
	"context"
	"database/sql"

	"go.uber.org/fx"

	"zankowitch.com/go-db-app/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(cfg config.Config, lc fx.Lifecycle) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
