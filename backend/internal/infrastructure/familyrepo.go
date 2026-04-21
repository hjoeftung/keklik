package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
)

// PostgresFamilyRepository implements family.FamilyRepository using PostgreSQL.
type PostgresFamilyRepository struct {
	db *sql.DB
}

// NewPostgresFamilyRepository returns a repository backed by the given database connection.
func NewPostgresFamilyRepository(db *sql.DB) *PostgresFamilyRepository {
	return &PostgresFamilyRepository{db: db}
}

// Save persists the full family aggregate (family row, member rows, and baby rows)
// in a single transaction.
func (r *PostgresFamilyRepository) Save(ctx context.Context, f family.Family) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx, `
		INSERT INTO families (id)
		VALUES ($1)
		ON CONFLICT (id) DO UPDATE SET
			updated_at = now()`,
		string(f.ID()),
	)
	if err != nil {
		return fmt.Errorf("upsert family: %w", err)
	}

	for _, m := range f.Members() {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO family_members (id, family_id, name, account_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
				name             = EXCLUDED.name,
				account_id = EXCLUDED.account_id,
				updated_at       = now()`,
			string(m.ID), string(m.FamilyID), m.Name, m.AccountID,
		)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" && pqErr.Constraint == "idx_family_members_account_id" {
				return family.ErrMemberAlreadyHasFamily
			}
			return fmt.Errorf("upsert family member %s: %w", m.ID, err)
		}
	}

	for _, b := range f.Babies() {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO babies (id, family_id, name)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE SET
				name       = EXCLUDED.name,
				updated_at = now()`,
			string(b.ID), string(b.FamilyID), b.Name,
		)
		if err != nil {
			return fmt.Errorf("upsert baby %s: %w", b.ID, err)
		}
	}

	for _, l := range f.InviteLinks() {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO invite_links (token, family_id, created_by_member_id, expires_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (token) DO UPDATE SET
				expires_at = EXCLUDED.expires_at`,
			string(l.Token), string(l.FamilyID), string(l.CreatedByMemberID), l.ExpiresAt,
		)
		if err != nil {
			return fmt.Errorf("upsert invite link %s: %w", l.Token, err)
		}
	}

	return tx.Commit()
}

// FindByID loads a family aggregate by its ID.
func (r *PostgresFamilyRepository) FindByID(ctx context.Context, id family.FamilyID) (family.Family, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM families WHERE id = $1)`, string(id)).
		Scan(&exists)
	if err != nil {
		return family.Family{}, fmt.Errorf("query family: %w", err)
	}
	if !exists {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}

	return r.reconstruct(ctx, id)
}

// FindByMemberID loads the family that contains the given member.
func (r *PostgresFamilyRepository) FindByMemberID(ctx context.Context, memberID family.FamilyMemberID) (family.Family, error) {
	var familyID family.FamilyID

	err := r.db.QueryRowContext(ctx, `
		SELECT f.id
		FROM families f
		JOIN family_members m ON m.family_id = f.id
		WHERE m.id = $1`, string(memberID)).
		Scan(&familyID)
	if err == sql.ErrNoRows {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	if err != nil {
		return family.Family{}, fmt.Errorf("query family by member: %w", err)
	}

	return r.reconstruct(ctx, familyID)
}

// FindByAccountID loads the family whose members include the given account.
func (r *PostgresFamilyRepository) FindByAccountID(ctx context.Context, accountID auth.AccountID) (family.Family, error) {
	var familyID family.FamilyID

	err := r.db.QueryRowContext(ctx, `
		SELECT f.id
		FROM families f
		JOIN family_members m ON m.family_id = f.id
		WHERE m.account_id = $1`, accountID).
		Scan(&familyID)
	if err == sql.ErrNoRows {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	if err != nil {
		return family.Family{}, fmt.Errorf("query family by account: %w", err)
	}

	return r.reconstruct(ctx, familyID)
}

// FindByInviteToken loads the family associated with the given invite token.
func (r *PostgresFamilyRepository) FindByInviteToken(ctx context.Context, token family.InviteToken) (family.Family, error) {
	var familyID family.FamilyID

	err := r.db.QueryRowContext(ctx, `
		SELECT f.id
		FROM families f
		JOIN invite_links l ON l.family_id = f.id
		WHERE l.token = $1`, string(token)).
		Scan(&familyID)
	if err == sql.ErrNoRows {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	if err != nil {
		return family.Family{}, fmt.Errorf("query family by invite token: %w", err)
	}

	return r.reconstruct(ctx, familyID)
}

// reconstruct builds a Family aggregate from a single JOIN across family_members, babies, and invite_links.
// The JOIN produces a Cartesian product (member × baby × invite_link), so each entity is deduplicated by ID.
func (r *PostgresFamilyRepository) reconstruct(
	ctx context.Context,
	id family.FamilyID,
) (family.Family, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			fm.id                   AS member_id,
			fm.family_id            AS member_family_id,
			fm.name                 AS member_name,
			fm.account_id           AS member_account_id,
			b.id                    AS baby_id,
			b.family_id             AS baby_family_id,
			b.name                  AS baby_name,
			il.token                AS invite_token,
			il.created_by_member_id AS invite_created_by,
			il.expires_at           AS invite_expires_at
		FROM families f
		LEFT JOIN family_members fm ON fm.family_id = f.id
		LEFT JOIN babies b          ON b.family_id  = f.id
		LEFT JOIN invite_links il   ON il.family_id = f.id
		WHERE f.id = $1`, string(id))
	if err != nil {
		return family.Family{}, fmt.Errorf("query family: %w", err)
	}
	defer rows.Close()

	seenMembers := make(map[family.FamilyMemberID]struct{})
	seenBabies := make(map[family.BabyID]struct{})
	seenLinks := make(map[family.InviteToken]struct{})
	var members []family.FamilyMember
	var babies []family.Baby
	var links []family.InviteLink

	for rows.Next() {
		var (
			memberID, memberFamilyID, memberName, memberAccountID sql.NullString
			babyID, babyFamilyID, babyName                       sql.NullString
			inviteToken, inviteCreatedBy                          sql.NullString
			inviteExpiresAt                                       sql.NullTime
		)
		if err := rows.Scan(
			&memberID, &memberFamilyID, &memberName, &memberAccountID,
			&babyID, &babyFamilyID, &babyName,
			&inviteToken, &inviteCreatedBy, &inviteExpiresAt,
		); err != nil {
			return family.Family{}, fmt.Errorf("scan family row: %w", err)
		}

		if memberID.Valid {
			mid := family.FamilyMemberID(memberID.String)
			if _, seen := seenMembers[mid]; !seen {
				seenMembers[mid] = struct{}{}
				members = append(members, family.FamilyMember{
					ID:        mid,
					FamilyID:  family.FamilyID(memberFamilyID.String),
					Name:      memberName.String,
					AccountID: auth.AccountID(memberAccountID.String),
				})
			}
		}

		if babyID.Valid {
			bid := family.BabyID(babyID.String)
			if _, seen := seenBabies[bid]; !seen {
				seenBabies[bid] = struct{}{}
				babies = append(babies, family.Baby{
					ID:       bid,
					FamilyID: family.FamilyID(babyFamilyID.String),
					Name:     babyName.String,
				})
			}
		}

		if inviteToken.Valid {
			tok := family.InviteToken(inviteToken.String)
			if _, seen := seenLinks[tok]; !seen {
				seenLinks[tok] = struct{}{}
				links = append(links, family.InviteLink{
					Token:             tok,
					FamilyID:          id,
					CreatedByMemberID: family.FamilyMemberID(inviteCreatedBy.String),
					ExpiresAt:         inviteExpiresAt.Time,
				})
			}
		}
	}
	if err := rows.Err(); err != nil {
		return family.Family{}, fmt.Errorf("iterate family rows: %w", err)
	}

	f, err := family.NewFamily(id, members, babies)
	if err != nil {
		return family.Family{}, fmt.Errorf("reconstruct family aggregate: %w", err)
	}

	for _, link := range links {
		if err := f.AddInviteLink(link); err != nil {
			return family.Family{}, fmt.Errorf("reconstruct invite link: %w", err)
		}
	}

	return f, nil
}

// PostgresFamilyMemberRepository implements family.FamilyMemberRepository using PostgreSQL.
type PostgresFamilyMemberRepository struct {
	db *sql.DB
}

// NewPostgresFamilyMemberRepository returns a repository backed by the given database connection.
func NewPostgresFamilyMemberRepository(db *sql.DB) *PostgresFamilyMemberRepository {
	return &PostgresFamilyMemberRepository{db: db}
}

// Save persists a FamilyMember record.
func (r *PostgresFamilyMemberRepository) Save(ctx context.Context, m family.FamilyMember) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO family_members (id, family_id, name, account_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name              = EXCLUDED.name,
			account_id = EXCLUDED.account_id,
			updated_at        = now()`,
		string(m.ID), string(m.FamilyID), m.Name, m.AccountID,
	)
	if err != nil {
		return fmt.Errorf("upsert family member: %w", err)
	}
	return nil
}

// FindByID loads a FamilyMember by its ID.
func (r *PostgresFamilyMemberRepository) FindByID(ctx context.Context, id family.FamilyMemberID) (family.FamilyMember, error) {
	var m family.FamilyMember
	var rawID, rawFamilyID string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, family_id, name, account_id
		FROM family_members WHERE id = $1`, string(id)).
		Scan(&rawID, &rawFamilyID, &m.Name, &m.AccountID)
	if err == sql.ErrNoRows {
		return family.FamilyMember{}, apperror.New(apperror.CodeNotFound, "family member not found")
	}
	if err != nil {
		return family.FamilyMember{}, fmt.Errorf("query family member by id: %w", err)
	}

	m.ID = family.FamilyMemberID(rawID)
	m.FamilyID = family.FamilyID(rawFamilyID)
	return m, nil
}

// FindByAccountID loads the FamilyMember associated with the given Google subject ID.
// This is the explicit link between an auth.Account and a family-domain FamilyMember.
func (r *PostgresFamilyMemberRepository) FindByAccountID(ctx context.Context, accountID auth.AccountID) (family.FamilyMember, error) {
	var m family.FamilyMember
	var rawID, rawFamilyID string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, family_id, name, account_id
		FROM family_members WHERE account_id = $1`, accountID).
		Scan(&rawID, &rawFamilyID, &m.Name, &m.AccountID)
	if err == sql.ErrNoRows {
		return family.FamilyMember{}, apperror.New(apperror.CodeNotFound, "family member not found")
	}
	if err != nil {
		return family.FamilyMember{}, fmt.Errorf("query family member by google subject id: %w", err)
	}

	m.ID = family.FamilyMemberID(rawID)
	m.FamilyID = family.FamilyID(rawFamilyID)
	return m, nil
}

