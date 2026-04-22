// Package reqctx provides context helpers for per-request tracing metadata.
package reqctx

import "context"

type contextKey int

const requestIDKey contextKey = iota

// WithRequestID returns a new context carrying the given request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext retrieves the request ID from the context.
// Returns ("", false) when no request ID is present.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok && id != ""
}
