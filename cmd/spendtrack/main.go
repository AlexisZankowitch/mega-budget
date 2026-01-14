package main

import (
	"net/http"
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"zankowitch.com/go-db-app/internal/config"
	"zankowitch.com/go-db-app/internal/db"
	"zankowitch.com/go-db-app/internal/handlers"
	"zankowitch.com/go-db-app/internal/httpserver"
)

func main() {
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ConsoleLogger{W: os.Stdout}
		}),
		fx.Provide(
			config.Load,
			db.New,
			handlers.NewHealthHandler,
			httpserver.NewMux,
			httpserver.NewServer,
		),
		fx.Invoke(func(*http.Server) {}),
	)

	app.Run()
}
