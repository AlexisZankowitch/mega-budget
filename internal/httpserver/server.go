package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	middleware "github.com/oapi-codegen/nethttp-middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"zankowitch.com/go-db-app/internal/api"
	"zankowitch.com/go-db-app/internal/config"
	"zankowitch.com/go-db-app/internal/logging"
)

func NewMux(healthHandler http.Handler, transactionsHandler api.StrictServerInterface) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthHandler)

	if transactionsHandler == nil {
		return mux, nil
	}

	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, err
	}

	handler := api.NewStrictHandler(transactionsHandler, nil)
	api.HandlerWithOptions(handler, api.StdHTTPServerOptions{
		BaseRouter:  mux,
		Middlewares: []api.MiddlewareFunc{middleware.OapiRequestValidator(swagger)},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.Error{Message: err.Error()})
		},
	})

	return mux, nil
}

func NewServer(cfg config.Config, mux *http.ServeMux, logger *zap.Logger, lc fx.Lifecycle) *http.Server {
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: logging.RequestIDAndLogger(logger)(mux),
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
