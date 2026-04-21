package auth

import "context"

type contextKey int

const accountContextKey contextKey = iota

// WithAccountID returns a new context carrying the authenticated account ID.
func WithAccountID(ctx context.Context, id AccountID) context.Context {
	return context.WithValue(ctx, accountContextKey, id)
}

// AccountIDFromContext retrieves the authenticated account ID from the context.
// The second return value is false when no account ID is present.
func AccountIDFromContext(ctx context.Context) (AccountID, bool) {
	id, ok := ctx.Value(accountContextKey).(AccountID)
	return id, ok
}
