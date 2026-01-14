package httpserver

import (
	"context"
	"log"
	"net/http"

	"go.uber.org/fx"

	"zankowitch.com/go-db-app/internal/config"
)

func NewMux(healthHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthHandler)
	return mux
}

func NewServer(cfg config.Config, mux *http.ServeMux, lc fx.Lifecycle) *http.Server {
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}
