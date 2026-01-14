package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"zankowitch.com/go-db-app/internal/config"
)

func NewHealthHandler(db *sql.DB, cfg config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), cfg.HealthTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Printf("health check failed: %v", err)
			http.Error(w, "unhealthy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
