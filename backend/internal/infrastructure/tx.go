package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
)

// dbQuerier is the common subset of *sql.DB and *sql.Tx used by repository methods.
type dbQuerier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type txContextKey struct{}

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// querierFromContext returns the *sql.Tx stored in ctx by WithTransaction, or
// falls back to db when no transaction is active.
func querierFromContext(ctx context.Context, db *sql.DB) dbQuerier {
	if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}

// PostgresTransactor implements sleep.Transactor using a *sql.DB.
type PostgresTransactor struct {
	db *sql.DB
}

// NewPostgresTransactor returns a PostgresTransactor backed by the given connection pool.
func NewPostgresTransactor(db *sql.DB) *PostgresTransactor {
	return &PostgresTransactor{db: db}
}

// WithTransaction begins a transaction, passes a transactional context to fn,
// and commits on success or rolls back on any error.
func (t *PostgresTransactor) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(withTx(ctx, tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
