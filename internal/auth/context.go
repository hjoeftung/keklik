package auth

import "context"

type contextKey int

const accountContextKey contextKey = iota

// WithAccount returns a new context carrying the authenticated account.
func WithAccount(ctx context.Context, account Account) context.Context {
	return context.WithValue(ctx, accountContextKey, account)
}

// AccountFromContext retrieves the authenticated account from the context.
// The second return value is false when no account is present.
func AccountFromContext(ctx context.Context) (Account, bool) {
	acc, ok := ctx.Value(accountContextKey).(Account)
	return acc, ok
}
