package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/hjoeftung/keklik/internal/apperror"
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
			INSERT INTO family_members (id, family_id, name, google_subject_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
				name             = EXCLUDED.name,
				google_subject_id = EXCLUDED.google_subject_id,
				updated_at       = now()`,
			string(m.ID), string(m.FamilyID), m.Name, m.GoogleSubjectID,
		)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" && pqErr.Constraint == "idx_family_members_google_subject_id" {
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

// reconstruct builds a Family aggregate from raw DB columns, loading members, babies, and invite links.
func (r *PostgresFamilyRepository) reconstruct(
	ctx context.Context,
	id family.FamilyID,
) (family.Family, error) {
	members, err := r.queryMembers(ctx, id)
	if err != nil {
		return family.Family{}, err
	}

	babies, err := r.queryBabies(ctx, id)
	if err != nil {
		return family.Family{}, err
	}

	f, err := family.NewFamily(id, members, babies)
	if err != nil {
		return family.Family{}, fmt.Errorf("reconstruct family aggregate: %w", err)
	}

	links, err := r.queryInviteLinks(ctx, id)
	if err != nil {
		return family.Family{}, err
	}

	for _, link := range links {
		if err := f.AddInviteLink(link); err != nil {
			return family.Family{}, fmt.Errorf("reconstruct invite link: %w", err)
		}
	}

	return f, nil
}

func (r *PostgresFamilyRepository) queryMembers(ctx context.Context, familyID family.FamilyID) ([]family.FamilyMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, family_id, name, google_subject_id
		FROM family_members WHERE family_id = $1`, string(familyID))
	if err != nil {
		return nil, fmt.Errorf("query family members: %w", err)
	}
	defer rows.Close()

	var members []family.FamilyMember
	for rows.Next() {
		var m family.FamilyMember
		var id, fid string
		if err := rows.Scan(&id, &fid, &m.Name, &m.GoogleSubjectID); err != nil {
			return nil, fmt.Errorf("scan family member: %w", err)
		}
		m.ID = family.FamilyMemberID(id)
		m.FamilyID = family.FamilyID(fid)
		members = append(members, m)
	}

	return members, rows.Err()
}

func (r *PostgresFamilyRepository) queryBabies(ctx context.Context, familyID family.FamilyID) ([]family.Baby, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, family_id, name
		FROM babies WHERE family_id = $1`, string(familyID))
	if err != nil {
		return nil, fmt.Errorf("query babies: %w", err)
	}
	defer rows.Close()

	var babies []family.Baby
	for rows.Next() {
		var b family.Baby
		var id, fid string
		if err := rows.Scan(&id, &fid, &b.Name); err != nil {
			return nil, fmt.Errorf("scan baby: %w", err)
		}
		b.ID = family.BabyID(id)
		b.FamilyID = family.FamilyID(fid)
		babies = append(babies, b)
	}

	return babies, rows.Err()
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
		INSERT INTO family_members (id, family_id, name, google_subject_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name              = EXCLUDED.name,
			google_subject_id = EXCLUDED.google_subject_id,
			updated_at        = now()`,
		string(m.ID), string(m.FamilyID), m.Name, m.GoogleSubjectID,
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
		SELECT id, family_id, name, google_subject_id
		FROM family_members WHERE id = $1`, string(id)).
		Scan(&rawID, &rawFamilyID, &m.Name, &m.GoogleSubjectID)
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

// FindByGoogleSubjectID loads the FamilyMember associated with the given Google subject ID.
// This is the explicit link between an auth.Account and a family-domain FamilyMember.
func (r *PostgresFamilyMemberRepository) FindByGoogleSubjectID(ctx context.Context, googleSubjectID string) (family.FamilyMember, error) {
	var m family.FamilyMember
	var rawID, rawFamilyID string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, family_id, name, google_subject_id
		FROM family_members WHERE google_subject_id = $1`, googleSubjectID).
		Scan(&rawID, &rawFamilyID, &m.Name, &m.GoogleSubjectID)
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

func (r *PostgresFamilyRepository) queryInviteLinks(ctx context.Context, familyID family.FamilyID) ([]family.InviteLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT token, family_id, created_by_member_id, expires_at
		FROM invite_links WHERE family_id = $1`, string(familyID))
	if err != nil {
		return nil, fmt.Errorf("query invite links: %w", err)
	}
	defer rows.Close()

	var links []family.InviteLink
	for rows.Next() {
		var token, fid, creatorID string
		var expiresAt time.Time
		if err := rows.Scan(&token, &fid, &creatorID, &expiresAt); err != nil {
			return nil, fmt.Errorf("scan invite link: %w", err)
		}
		links = append(links, family.InviteLink{
			Token:             family.InviteToken(token),
			FamilyID:          family.FamilyID(fid),
			CreatedByMemberID: family.FamilyMemberID(creatorID),
			ExpiresAt:         expiresAt,
		})
	}

	return links, rows.Err()
}
