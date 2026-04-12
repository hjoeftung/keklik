package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
)

// PostgresAccountRepository implements auth.AccountRepository using PostgreSQL.
type PostgresAccountRepository struct {
	db *sql.DB
}

// NewPostgresAccountRepository returns a repository backed by the given database connection.
func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

// Save persists a new Account record. Conflicts on ID are not expected in normal flow.
func (r *PostgresAccountRepository) Save(ctx context.Context, a auth.Account) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO accounts (id, google_subject_id, email)
		VALUES ($1, $2, $3)`,
		string(a.ID), a.GoogleSubjectID, a.Email,
	)
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

// FindByID loads an Account by its internal ID.
func (r *PostgresAccountRepository) FindByID(ctx context.Context, id auth.AccountID) (auth.Account, error) {
	var a auth.Account
	var rawID string
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, `
		SELECT id, google_subject_id, email, created_at
		FROM accounts WHERE id = $1`, string(id)).
		Scan(&rawID, &a.GoogleSubjectID, &a.Email, &createdAt)
	if err == sql.ErrNoRows {
		return auth.Account{}, auth.ErrAccountNotFound
	}
	if err != nil {
		return auth.Account{}, fmt.Errorf("query account by id: %w", err)
	}

	a.ID = auth.AccountID(rawID)
	a.CreatedAt = createdAt
	return a, nil
}

// FindByGoogleSubjectID loads an Account by its Google subject identifier.
func (r *PostgresAccountRepository) FindByGoogleSubjectID(ctx context.Context, googleSubjectID string) (auth.Account, error) {
	var a auth.Account
	var rawID string
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, `
		SELECT id, google_subject_id, email, created_at
		FROM accounts WHERE google_subject_id = $1`, googleSubjectID).
		Scan(&rawID, &a.GoogleSubjectID, &a.Email, &createdAt)
	if err == sql.ErrNoRows {
		return auth.Account{}, auth.ErrAccountNotFound
	}
	if err != nil {
		return auth.Account{}, fmt.Errorf("query account by google subject id: %w", err)
	}

	a.ID = auth.AccountID(rawID)
	a.CreatedAt = createdAt
	return a, nil
}

// PostgresSessionRepository implements auth.SessionRepository using PostgreSQL.
type PostgresSessionRepository struct {
	db *sql.DB
}

// NewPostgresSessionRepository returns a repository backed by the given database connection.
func NewPostgresSessionRepository(db *sql.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

// Save persists a new Session record.
func (r *PostgresSessionRepository) Save(ctx context.Context, s auth.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (token, account_id, expires_at)
		VALUES ($1, $2, $3)`,
		string(s.Token), string(s.AccountID), s.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// FindByToken loads a Session by its bearer token.
func (r *PostgresSessionRepository) FindByToken(ctx context.Context, token auth.SessionToken) (auth.Session, error) {
	var s auth.Session
	var rawToken, rawAccountID string

	err := r.db.QueryRowContext(ctx, `
		SELECT token, account_id, expires_at
		FROM sessions WHERE token = $1`, string(token)).
		Scan(&rawToken, &rawAccountID, &s.ExpiresAt)
	if err == sql.ErrNoRows {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	if err != nil {
		return auth.Session{}, fmt.Errorf("query session by token: %w", err)
	}

	s.Token = auth.SessionToken(rawToken)
	s.AccountID = auth.AccountID(rawAccountID)
	return s, nil
}
