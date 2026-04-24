package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// PostgresBabyAccessChecker verifies that the caller is a family member of a given baby
// in a single database round-trip.
type PostgresBabyAccessChecker struct {
	db *sql.DB
}

// NewPostgresBabyAccessChecker returns a checker backed by the given database connection.
func NewPostgresBabyAccessChecker(db *sql.DB) *PostgresBabyAccessChecker {
	return &PostgresBabyAccessChecker{db: db}
}

// CheckBabyAccess verifies the caller (identified by accountID) belongs to the family
// of babyID. Returns CodeNotFound if the baby does not exist, CodeForbidden if the caller
// is not a member.
func (c *PostgresBabyAccessChecker) CheckBabyAccess(ctx context.Context, accountID auth.AccountID, babyID sleep.BabyID) (sleep.FamilyMemberID, error) {
	var babyRowID string
	var memberID sql.NullString

	err := c.db.QueryRowContext(ctx, `
		SELECT b.id, fm.id
		FROM babies b
		LEFT JOIN family_members fm ON fm.family_id = b.family_id AND fm.account_id = $1
		WHERE b.id = $2`,
		string(accountID), string(babyID)).
		Scan(&babyRowID, &memberID)
	if err == sql.ErrNoRows {
		return "", apperror.New(apperror.CodeNotFound, "baby not found")
	}
	if err != nil {
		return "", fmt.Errorf("check baby access: %w", err)
	}
	if !memberID.Valid {
		return "", apperror.New(apperror.CodeForbidden, "access to this baby is not allowed")
	}

	return sleep.FamilyMemberID(memberID.String), nil
}
