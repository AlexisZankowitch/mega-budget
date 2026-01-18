package logging

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func RequestIDAndLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get(requestIDHeader)
			if reqID == "" {
				reqID = uuid.NewString()
			}

			w.Header().Set(requestIDHeader, reqID)
			r = r.WithContext(WithRequestID(r.Context(), reqID))

			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			logger.Info(
				"http request",
				zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rec.status),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
