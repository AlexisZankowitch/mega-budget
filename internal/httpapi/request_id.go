package httpapi

import (
	"context"

	"zankowitch.com/go-db-app/internal/logging"
)

func requestIDFromContext(ctx context.Context) string {
	if id, ok := logging.RequestIDFromContext(ctx); ok {
		return id
	}
	return ""
}
