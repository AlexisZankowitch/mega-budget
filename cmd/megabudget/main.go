package main

import (
	"net/http"
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"zankowitch.com/go-db-app/internal/api"
	"zankowitch.com/go-db-app/internal/categories"
	"zankowitch.com/go-db-app/internal/config"
	"zankowitch.com/go-db-app/internal/db"
	"zankowitch.com/go-db-app/internal/handlers"
	"zankowitch.com/go-db-app/internal/httpapi"
	"zankowitch.com/go-db-app/internal/httpserver"
	"zankowitch.com/go-db-app/internal/logging"
	"zankowitch.com/go-db-app/internal/transactions"
)

func main() {
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ConsoleLogger{W: os.Stdout}
		}),
		fx.Provide(
			config.Load,
			db.New,
			logging.New,
			handlers.NewHealthHandler,
			categories.NewRepository,
			httpapi.NewCategoriesHandler,
			transactions.NewRepository,
			httpapi.NewTransactionsHandler,
			httpapi.NewAnalyticsHandler,
			httpapi.NewHandler,
			func(h *httpapi.Handler) api.StrictServerInterface { return h },
			httpserver.NewMux,
			httpserver.NewServer,
		),
		fx.Invoke(func(*http.Server) {}),
	)

	app.Run()
}
