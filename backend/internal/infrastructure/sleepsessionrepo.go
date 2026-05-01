package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// PostgresSleepSessionRepository implements sleep.SleepSessionRepository and
// sleep.ActiveSleepSessionRepository using PostgreSQL.
type PostgresSleepSessionRepository struct {
	db *sql.DB
}

// NewPostgresSleepSessionRepository returns a repository backed by the given database connection.
func NewPostgresSleepSessionRepository(db *sql.DB) *PostgresSleepSessionRepository {
	return &PostgresSleepSessionRepository{db: db}
}

// Save persists a SleepSession using optimistic concurrency control.
//
// New sessions (Version == 0) are inserted with version=1. Existing sessions
// (Version > 0) are updated only when the stored version matches; if another
// writer incremented the version first, Save returns ErrSleepSessionConflict.
// An overlap exclusion-constraint violation is mapped to ErrSleepSessionOverlap,
// and a duplicate-active-session violation is mapped to ErrActiveSleepSessionExists.
func (r *PostgresSleepSessionRepository) Save(ctx context.Context, s sleep.SleepSession) error {
	if s.Version() == 0 {
		return r.insert(ctx, s)
	}
	return r.update(ctx, s)
}

func (r *PostgresSleepSessionRepository) insert(ctx context.Context, s sleep.SleepSession) error {
	var stoppedAt *time.Time
	if t, ok := s.StoppedAt(); ok {
		ts := t.UTC()
		stoppedAt = &ts
	}

	_, err := querierFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO sleep_sessions (
			id, baby_id, created_by_member_id,
			started_at, stopped_at, version
		) VALUES ($1, $2, $3, $4, $5, 1)`,
		string(s.ID()),
		string(s.BabyID()),
		string(s.CreatedByMemberID()),
		s.StartedAt().UTC(),
		stoppedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case pgUniqueViolation:
				return apperror.Wrap(apperror.CodeActiveSleepExists, sleep.ErrActiveSleepSessionExists.Error(), sleep.ErrActiveSleepSessionExists)
			case pgExclusionViolation:
				return apperror.Wrap(apperror.CodeConflict, sleep.ErrSleepSessionOverlap.Error(), sleep.ErrSleepSessionOverlap)
			}
		}
		return fmt.Errorf("insert sleep session: %w", err)
	}
	return nil
}

func (r *PostgresSleepSessionRepository) update(ctx context.Context, s sleep.SleepSession) error {
	var stoppedAt *time.Time
	if t, ok := s.StoppedAt(); ok {
		ts := t.UTC()
		stoppedAt = &ts
	}

	result, err := querierFromContext(ctx, r.db).ExecContext(ctx, `
		UPDATE sleep_sessions
		SET started_at           = $2,
		    stopped_at           = $3,
		    updated_by_member_id = $4,
		    updated_at           = now(),
		    version              = version + 1
		WHERE id = $1
		  AND version = $5`,
		string(s.ID()),
		s.StartedAt().UTC(),
		stoppedAt,
		string(s.CreatedByMemberID()),
		s.Version(),
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pgExclusionViolation {
			return apperror.Wrap(apperror.CodeConflict, sleep.ErrSleepSessionOverlap.Error(), sleep.ErrSleepSessionOverlap)
		}
		return fmt.Errorf("update sleep session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update sleep session rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return apperror.Wrap(apperror.CodeConflict, sleep.ErrSleepSessionConflict.Error(), sleep.ErrSleepSessionConflict)
	}
	return nil
}

// SaveAll persists multiple SleepSessions. Each session must already exist in
// the database (Version > 0); it is updated with a conditional version check
// and the version counter is incremented. Returns ErrSleepSessionConflict if
// any row was concurrently modified.
func (r *PostgresSleepSessionRepository) SaveAll(ctx context.Context, sessions []sleep.SleepSession) error {
	for _, s := range sessions {
		if err := r.Save(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

// FindByID loads a SleepSession by its ID.
func (r *PostgresSleepSessionRepository) FindByID(ctx context.Context, id sleep.SleepSessionID) (sleep.SleepSession, error) {
	row := querierFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id, started_at, stopped_at, version
		FROM sleep_sessions WHERE id = $1`, string(id))

	return scanSleepSession(row)
}

// FindActiveByBabyID loads the active (not yet stopped) sleep session for a baby.
// Returns apperror with CodeNotFound when no active session exists.
func (r *PostgresSleepSessionRepository) FindActiveByBabyID(ctx context.Context, babyID sleep.BabyID) (sleep.SleepSession, error) {
	row := querierFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id, started_at, stopped_at, version
		FROM sleep_sessions
		WHERE baby_id = $1 AND stopped_at IS NULL`, string(babyID))

	return scanSleepSession(row)
}

// FindByBabyIDAndDateRange returns all sessions for a baby whose started_at falls
// within [dateRange.Start, dateRange.End), ordered by started_at descending.
func (r *PostgresSleepSessionRepository) FindByBabyIDAndDateRange(ctx context.Context, babyID sleep.BabyID, dateRange sleep.DateRange) ([]sleep.SleepSession, error) {
	rows, err := querierFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, baby_id, created_by_member_id, started_at, stopped_at, version
		FROM sleep_sessions
		WHERE baby_id = $1 AND started_at >= $2 AND started_at < $3
		ORDER BY started_at DESC`,
		string(babyID),
		dateRange.Start().UTC(),
		dateRange.End().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("query sleep sessions by date range: %w", err)
	}
	defer rows.Close()

	var sessions []sleep.SleepSession
	for rows.Next() {
		s, err := scanSleepSessionRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sleep sessions: %w", err)
	}

	return sessions, nil
}

// DeleteByID hard-deletes a sleep session by ID.
// Returns apperror with CodeNotFound when no session with that ID exists.
func (r *PostgresSleepSessionRepository) DeleteByID(ctx context.Context, id sleep.SleepSessionID) error {
	result, err := querierFromContext(ctx, r.db).ExecContext(ctx, `
		DELETE FROM sleep_sessions WHERE id = $1`,
		string(id),
	)
	if err != nil {
		return fmt.Errorf("delete sleep session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete sleep session rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return apperror.New(apperror.CodeNotFound, "sleep session not found")
	}

	return nil
}

// DeleteByIDAndVersion hard-deletes a sleep session only if the version still matches.
func (r *PostgresSleepSessionRepository) DeleteByIDAndVersion(ctx context.Context, id sleep.SleepSessionID, version int) error {
	result, err := querierFromContext(ctx, r.db).ExecContext(ctx, `
		DELETE FROM sleep_sessions WHERE id = $1 AND version = $2`,
		string(id),
		version,
	)
	if err != nil {
		return fmt.Errorf("delete sleep session by version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete sleep session by version rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return apperror.Wrap(apperror.CodeConflict, sleep.ErrSleepSessionConflict.Error(), sleep.ErrSleepSessionConflict)
	}

	return nil
}

// FindMostRecentByBabyID returns the most recently started sleep session for a
// baby, regardless of whether it is active or completed.
// Returns apperror with CodeNotFound when no sessions exist.
func (r *PostgresSleepSessionRepository) FindMostRecentByBabyID(ctx context.Context, babyID sleep.BabyID) (sleep.SleepSession, error) {
	row := querierFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id, started_at, stopped_at, version
		FROM sleep_sessions
		WHERE baby_id = $1
		ORDER BY started_at DESC
		LIMIT 1`, string(babyID))

	return scanSleepSession(row)
}

// FindMostRecentCompletedByBabyID returns the most recently ended sleep session
// for a baby (the one with the latest stopped_at).
// Returns apperror with CodeNotFound when no completed sessions exist.
func (r *PostgresSleepSessionRepository) FindMostRecentCompletedByBabyID(ctx context.Context, babyID sleep.BabyID) (sleep.SleepSession, error) {
	row := querierFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id, started_at, stopped_at, version
		FROM sleep_sessions
		WHERE baby_id = $1 AND stopped_at IS NOT NULL
		ORDER BY stopped_at DESC
		LIMIT 1`, string(babyID))

	return scanSleepSession(row)
}

// FindCompletedByBabyIDSince returns all completed (stopped_at IS NOT NULL) sessions
// for a baby whose started_at is >= since, ordered by started_at ascending.
func (r *PostgresSleepSessionRepository) FindCompletedByBabyIDSince(ctx context.Context, babyID sleep.BabyID, since time.Time) ([]sleep.SleepSession, error) {
	rows, err := querierFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, baby_id, created_by_member_id, started_at, stopped_at, version
		FROM sleep_sessions
		WHERE baby_id = $1
		  AND started_at >= $2
		  AND stopped_at IS NOT NULL
		ORDER BY started_at ASC`,
		string(babyID),
		since.UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("query completed sleep sessions since: %w", err)
	}
	defer rows.Close()

	var sessions []sleep.SleepSession
	for rows.Next() {
		s, err := scanSleepSessionRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate completed sleep sessions: %w", err)
	}

	return sessions, nil
}

// HasOverlappingByBabyID reports whether any existing session for the baby
// intersects the interval [startedAt, stoppedAt). Active sessions (stopped_at IS NULL)
// are treated as open-ended and always overlap if they started before stoppedAt.
func (r *PostgresSleepSessionRepository) HasOverlappingByBabyID(ctx context.Context, babyID sleep.BabyID, startedAt time.Time, stoppedAt time.Time) (bool, error) {
	var count int
	err := querierFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sleep_sessions
		WHERE baby_id = $1
		  AND started_at < $3
		  AND (stopped_at IS NULL OR stopped_at > $2)`,
		string(babyID),
		startedAt.UTC(),
		stoppedAt.UTC(),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check overlapping sleep sessions: %w", err)
	}
	return count > 0, nil
}

// FindOverlappingByBabyID returns the first session that intersects
// [startedAt, stoppedAt), optionally excluding a known session ID.
func (r *PostgresSleepSessionRepository) FindOverlappingByBabyID(ctx context.Context, babyID sleep.BabyID, startedAt time.Time, stoppedAt time.Time, excludeID *sleep.SleepSessionID) (sleep.SleepSession, error) {
	var excludedID *string
	if excludeID != nil {
		id := string(*excludeID)
		excludedID = &id
	}

	row := querierFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id, started_at, stopped_at, version
		FROM sleep_sessions
		WHERE baby_id = $1
		  AND started_at < $3
		  AND (stopped_at IS NULL OR stopped_at > $2)
		  AND ($4::uuid IS NULL OR id <> $4::uuid)
		ORDER BY started_at ASC
		LIMIT 1`,
		string(babyID),
		startedAt.UTC(),
		stoppedAt.UTC(),
		excludedID,
	)

	return scanSleepSession(row)
}

func scanSleepSession(row *sql.Row) (sleep.SleepSession, error) {
	var (
		id, babyID, memberID string
		startedAt, stoppedAt sql.NullTime
		version              int
	)

	err := row.Scan(&id, &babyID, &memberID, &startedAt, &stoppedAt, &version)
	if err == sql.ErrNoRows {
		return sleep.SleepSession{}, apperror.New(apperror.CodeNotFound, "sleep session not found")
	}
	if err != nil {
		return sleep.SleepSession{}, fmt.Errorf("scan sleep session: %w", err)
	}

	return assembleSleepSession(id, babyID, memberID, startedAt, stoppedAt, version)
}

func scanSleepSessionRows(rows *sql.Rows) (sleep.SleepSession, error) {
	var (
		id, babyID, memberID string
		startedAt, stoppedAt sql.NullTime
		version              int
	)

	if err := rows.Scan(&id, &babyID, &memberID, &startedAt, &stoppedAt, &version); err != nil {
		return sleep.SleepSession{}, fmt.Errorf("scan sleep session row: %w", err)
	}

	return assembleSleepSession(id, babyID, memberID, startedAt, stoppedAt, version)
}

func assembleSleepSession(id, babyID, memberID string, startedAt, stoppedAt sql.NullTime, version int) (sleep.SleepSession, error) {
	var stoppedAtPtr *time.Time
	if stoppedAt.Valid {
		t := stoppedAt.Time
		stoppedAtPtr = &t
	}
	return sleep.RestoreSleepSession(
		sleep.SleepSessionID(id),
		sleep.BabyID(babyID),
		sleep.FamilyMemberID(memberID),
		startedAt.Time,
		stoppedAtPtr,
		version,
	)
}
