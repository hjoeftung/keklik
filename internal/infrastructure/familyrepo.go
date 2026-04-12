package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

	nw := f.NightWindow()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO families (id, name, timezone,
			night_window_start_hour, night_window_start_minute,
			night_window_end_hour,   night_window_end_minute)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name                      = EXCLUDED.name,
			timezone                  = EXCLUDED.timezone,
			night_window_start_hour   = EXCLUDED.night_window_start_hour,
			night_window_start_minute = EXCLUDED.night_window_start_minute,
			night_window_end_hour     = EXCLUDED.night_window_end_hour,
			night_window_end_minute   = EXCLUDED.night_window_end_minute,
			updated_at                = now()`,
		string(f.ID()), f.Name(), f.Timezone(),
		nw.Start().Hour(), nw.Start().Minute(),
		nw.End().Hour(), nw.End().Minute(),
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

	return tx.Commit()
}

// FindByID loads a family aggregate by its ID.
func (r *PostgresFamilyRepository) FindByID(ctx context.Context, id family.FamilyID) (family.Family, error) {
	var name, timezone string
	var startHour, startMinute, endHour, endMinute int

	err := r.db.QueryRowContext(ctx, `
		SELECT name, timezone,
			night_window_start_hour, night_window_start_minute,
			night_window_end_hour,   night_window_end_minute
		FROM families WHERE id = $1`, string(id)).
		Scan(&name, &timezone, &startHour, &startMinute, &endHour, &endMinute)
	if err == sql.ErrNoRows {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	if err != nil {
		return family.Family{}, fmt.Errorf("query family: %w", err)
	}

	return r.reconstruct(ctx, id, name, timezone, startHour, startMinute, endHour, endMinute)
}

// FindByMemberID loads the family that contains the given member.
func (r *PostgresFamilyRepository) FindByMemberID(ctx context.Context, memberID family.FamilyMemberID) (family.Family, error) {
	var familyID family.FamilyID
	var name, timezone string
	var startHour, startMinute, endHour, endMinute int

	err := r.db.QueryRowContext(ctx, `
		SELECT f.id, f.name, f.timezone,
			f.night_window_start_hour, f.night_window_start_minute,
			f.night_window_end_hour,   f.night_window_end_minute
		FROM families f
		JOIN family_members m ON m.family_id = f.id
		WHERE m.id = $1`, string(memberID)).
		Scan(&familyID, &name, &timezone, &startHour, &startMinute, &endHour, &endMinute)
	if err == sql.ErrNoRows {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	if err != nil {
		return family.Family{}, fmt.Errorf("query family by member: %w", err)
	}

	return r.reconstruct(ctx, familyID, name, timezone, startHour, startMinute, endHour, endMinute)
}

// FindByInviteToken loads the family associated with the given invite token.
func (r *PostgresFamilyRepository) FindByInviteToken(ctx context.Context, token family.InviteToken) (family.Family, error) {
	var familyID family.FamilyID
	var name, timezone string
	var startHour, startMinute, endHour, endMinute int

	err := r.db.QueryRowContext(ctx, `
		SELECT f.id, f.name, f.timezone,
			f.night_window_start_hour, f.night_window_start_minute,
			f.night_window_end_hour,   f.night_window_end_minute
		FROM families f
		JOIN invite_links l ON l.family_id = f.id
		WHERE l.token = $1`, string(token)).
		Scan(&familyID, &name, &timezone, &startHour, &startMinute, &endHour, &endMinute)
	if err == sql.ErrNoRows {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	if err != nil {
		return family.Family{}, fmt.Errorf("query family by invite token: %w", err)
	}

	return r.reconstruct(ctx, familyID, name, timezone, startHour, startMinute, endHour, endMinute)
}

// reconstruct builds a Family aggregate from raw DB columns, loading members, babies, and invite links.
func (r *PostgresFamilyRepository) reconstruct(
	ctx context.Context,
	id family.FamilyID, name, timezone string,
	startHour, startMinute, endHour, endMinute int,
) (family.Family, error) {
	start, err := family.NewLocalTime(startHour, startMinute)
	if err != nil {
		return family.Family{}, fmt.Errorf("reconstruct night window start: %w", err)
	}

	end, err := family.NewLocalTime(endHour, endMinute)
	if err != nil {
		return family.Family{}, fmt.Errorf("reconstruct night window end: %w", err)
	}

	nightWindow, err := family.NewNightWindow(start, end)
	if err != nil {
		return family.Family{}, fmt.Errorf("reconstruct night window: %w", err)
	}

	members, err := r.queryMembers(ctx, id)
	if err != nil {
		return family.Family{}, err
	}

	babies, err := r.queryBabies(ctx, id)
	if err != nil {
		return family.Family{}, err
	}

	f, err := family.NewFamily(id, name, timezone, nightWindow, members, babies)
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
