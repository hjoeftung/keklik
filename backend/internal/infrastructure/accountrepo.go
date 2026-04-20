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

