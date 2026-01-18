package logging

import (
	"context"
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok
}
