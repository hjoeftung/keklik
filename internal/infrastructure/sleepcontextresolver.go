package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// PostgresSleepContextResolver resolves the baby ID and family member ID needed
// to start or stop a sleep session from the caller's Google subject ID.
type PostgresSleepContextResolver struct {
	db *sql.DB
}

// NewPostgresSleepContextResolver returns a resolver backed by the given database connection.
func NewPostgresSleepContextResolver(db *sql.DB) *PostgresSleepContextResolver {
	return &PostgresSleepContextResolver{db: db}
}

// ResolveSleepContext returns the baby ID and family member ID for the given Google subject ID.
func (r *PostgresSleepContextResolver) ResolveSleepContext(ctx context.Context, googleSubjectID string) (sleep.BabyID, sleep.FamilyMemberID, error) {
	var babyID, memberID string

	err := r.db.QueryRowContext(ctx, `
		SELECT b.id, fm.id
		FROM family_members fm
		JOIN babies b ON b.family_id = fm.family_id
		WHERE fm.google_subject_id = $1`, googleSubjectID).
		Scan(&babyID, &memberID)
	if err == sql.ErrNoRows {
		return "", "", apperror.New(apperror.CodeNotFound, "no family found for this account")
	}
	if err != nil {
		return "", "", fmt.Errorf("resolve sleep context: %w", err)
	}

	return sleep.BabyID(babyID), sleep.FamilyMemberID(memberID), nil
}
