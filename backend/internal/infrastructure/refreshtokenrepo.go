package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
)

// PostgresRefreshTokenRepository implements auth.RefreshTokenRepository using PostgreSQL.
type PostgresRefreshTokenRepository struct {
	db *sql.DB
}

// NewPostgresRefreshTokenRepository returns a repository backed by the given database connection.
func NewPostgresRefreshTokenRepository(db *sql.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

// Save persists a new refresh token record.
func (r *PostgresRefreshTokenRepository) Save(ctx context.Context, t auth.RefreshToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (token, account_id, expires_at)
		VALUES ($1, $2, $3)`,
		t.Token, string(t.AccountID), t.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// FindByToken loads a refresh token by its opaque value. Returns ErrRefreshTokenNotFound
// if no row exists. Revoked tokens are still returned so callers can inspect RevokedAt.
func (r *PostgresRefreshTokenRepository) FindByToken(ctx context.Context, token string) (auth.RefreshToken, error) {
	var t auth.RefreshToken
	var rawAccountID string
	var expiresAt time.Time
	var revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT token, account_id, expires_at, revoked_at
		FROM refresh_tokens WHERE token = $1`, token).
		Scan(&t.Token, &rawAccountID, &expiresAt, &revokedAt)
	if err == sql.ErrNoRows {
		return auth.RefreshToken{}, auth.ErrRefreshTokenNotFound
	}
	if err != nil {
		return auth.RefreshToken{}, fmt.Errorf("query refresh token: %w", err)
	}

	t.AccountID = auth.AccountID(rawAccountID)
	t.ExpiresAt = expiresAt
	if revokedAt.Valid {
		t.RevokedAt = &revokedAt.Time
	}
	return t, nil
}

// Revoke marks a single refresh token as revoked. Returns ErrRefreshTokenNotFound if
// the token does not exist.
func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE token = $1 AND revoked_at IS NULL`,
		token,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke refresh token rows affected: %w", err)
	}
	if n == 0 {
		return auth.ErrRefreshTokenNotFound
	}
	return nil
}

// RevokeAllForAccount marks all active refresh tokens for an account as revoked.
func (r *PostgresRefreshTokenRepository) RevokeAllForAccount(ctx context.Context, accountID auth.AccountID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE account_id = $1 AND revoked_at IS NULL`,
		string(accountID),
	)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens for account: %w", err)
	}
	return nil
}
