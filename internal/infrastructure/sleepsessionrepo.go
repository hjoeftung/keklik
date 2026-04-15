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

// Save persists a SleepSession. When an active session already exists for the
// same baby (unique partial index violation), it returns an AppError with
// CodeActiveSleepExists so the use case can surface the correct conflict response.
func (r *PostgresSleepSessionRepository) Save(ctx context.Context, s sleep.SleepSession) error {
	var stoppedAt *time.Time
	if t, ok := s.StoppedAt(); ok {
		ts := t.UTC()
		stoppedAt = &ts
	}

	var classification string
	if c := s.Classification(); c != sleep.SleepClassificationUnknown {
		classification = string(c)
	}

	var (
		nwStartHour, nwStartMinute *int
		nwEndHour, nwEndMinute     *int
	)
	if nw := s.ClassifiedWithNightWindow(); nw != nil {
		sh, sm, eh, em := nw.Start().Hour(), nw.Start().Minute(), nw.End().Hour(), nw.End().Minute()
		nwStartHour, nwStartMinute = &sh, &sm
		nwEndHour, nwEndMinute = &eh, &em
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sleep_sessions (
			id, baby_id, created_by_member_id,
			started_at, stopped_at,
			classification,
			classified_with_nw_start_hour, classified_with_nw_start_minute,
			classified_with_nw_end_hour,   classified_with_nw_end_minute
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			started_at                      = EXCLUDED.started_at,
			stopped_at                     = EXCLUDED.stopped_at,
			classification                 = EXCLUDED.classification,
			classified_with_nw_start_hour  = EXCLUDED.classified_with_nw_start_hour,
			classified_with_nw_start_minute= EXCLUDED.classified_with_nw_start_minute,
			classified_with_nw_end_hour    = EXCLUDED.classified_with_nw_end_hour,
			classified_with_nw_end_minute  = EXCLUDED.classified_with_nw_end_minute,
			updated_by_member_id           = EXCLUDED.created_by_member_id,
			updated_at                     = now()`,
		string(s.ID()),
		string(s.BabyID()),
		string(s.CreatedByMemberID()),
		s.StartedAt().UTC(),
		stoppedAt,
		classification,
		nwStartHour, nwStartMinute,
		nwEndHour, nwEndMinute,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return apperror.New(apperror.CodeActiveSleepExists, sleep.ErrActiveSleepSessionExists.Error())
		}
		return fmt.Errorf("upsert sleep session: %w", err)
	}
	return nil
}

// FindByID loads a SleepSession by its ID.
func (r *PostgresSleepSessionRepository) FindByID(ctx context.Context, id sleep.SleepSessionID) (sleep.SleepSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id,
		       started_at, stopped_at,
		       classification,
		       classified_with_nw_start_hour, classified_with_nw_start_minute,
		       classified_with_nw_end_hour,   classified_with_nw_end_minute
		FROM sleep_sessions WHERE id = $1`, string(id))

	return scanSleepSession(row)
}

// FindByIDForFamilyMember loads a sleep session by ID after verifying the
// given family member belongs to the same family as the session's baby.
func (r *PostgresSleepSessionRepository) FindByIDForFamilyMember(ctx context.Context, id sleep.SleepSessionID, memberID sleep.FamilyMemberID) (sleep.SleepSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.baby_id, s.created_by_member_id,
		       s.started_at, s.stopped_at,
		       s.classification,
		       s.classified_with_nw_start_hour, s.classified_with_nw_start_minute,
		       s.classified_with_nw_end_hour,   s.classified_with_nw_end_minute
		FROM sleep_sessions s
		JOIN babies b ON b.id = s.baby_id
		JOIN family_members fm ON fm.family_id = b.family_id
		WHERE s.id = $1 AND fm.id = $2`,
		string(id),
		string(memberID),
	)

	return scanSleepSession(row)
}

// FindActiveByBabyID loads the active (not yet stopped) sleep session for a baby.
// Returns apperror with CodeNotFound when no active session exists.
func (r *PostgresSleepSessionRepository) FindActiveByBabyID(ctx context.Context, babyID sleep.BabyID) (sleep.SleepSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id,
		       started_at, stopped_at,
		       classification,
		       classified_with_nw_start_hour, classified_with_nw_start_minute,
		       classified_with_nw_end_hour,   classified_with_nw_end_minute
		FROM sleep_sessions
		WHERE baby_id = $1 AND stopped_at IS NULL`, string(babyID))

	return scanSleepSession(row)
}

// FindByBabyIDAndDateRange returns all sessions for a baby whose started_at falls
// within [dateRange.Start, dateRange.End), ordered by started_at descending.
func (r *PostgresSleepSessionRepository) FindByBabyIDAndDateRange(ctx context.Context, babyID sleep.BabyID, dateRange sleep.DateRange) ([]sleep.SleepSession, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, baby_id, created_by_member_id,
		       started_at, stopped_at,
		       classification,
		       classified_with_nw_start_hour, classified_with_nw_start_minute,
		       classified_with_nw_end_hour,   classified_with_nw_end_minute
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

// DeleteByIDForFamilyMember hard-deletes a sleep session after verifying the
// given family member belongs to the same family as the session's baby.
func (r *PostgresSleepSessionRepository) DeleteByIDForFamilyMember(ctx context.Context, id sleep.SleepSessionID, memberID sleep.FamilyMemberID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM sleep_sessions s
		USING babies b, family_members fm
		WHERE s.id = $1
		  AND s.baby_id = b.id
		  AND fm.id = $2
		  AND fm.family_id = b.family_id`,
		string(id),
		string(memberID),
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

func scanSleepSession(row *sql.Row) (sleep.SleepSession, error) {
	var (
		id, babyID, memberID                               string
		classification                                     string
		nwStartHour, nwStartMinute, nwEndHour, nwEndMinute sql.NullInt32
		startedAt, stoppedAt                               sql.NullTime
	)

	err := row.Scan(
		&id, &babyID, &memberID, &startedAt, &stoppedAt,
		&classification,
		&nwStartHour, &nwStartMinute, &nwEndHour, &nwEndMinute,
	)
	if err == sql.ErrNoRows {
		return sleep.SleepSession{}, apperror.New(apperror.CodeNotFound, "sleep session not found")
	}
	if err != nil {
		return sleep.SleepSession{}, fmt.Errorf("scan sleep session: %w", err)
	}

	return assembleSleepSession(id, babyID, memberID, startedAt, stoppedAt, classification, nwStartHour, nwStartMinute, nwEndHour, nwEndMinute)
}

func scanSleepSessionRows(rows *sql.Rows) (sleep.SleepSession, error) {
	var (
		id, babyID, memberID                               string
		classification                                     string
		nwStartHour, nwStartMinute, nwEndHour, nwEndMinute sql.NullInt32
		startedAt, stoppedAt                               sql.NullTime
	)

	if err := rows.Scan(
		&id, &babyID, &memberID, &startedAt, &stoppedAt,
		&classification,
		&nwStartHour, &nwStartMinute, &nwEndHour, &nwEndMinute,
	); err != nil {
		return sleep.SleepSession{}, fmt.Errorf("scan sleep session row: %w", err)
	}

	return assembleSleepSession(id, babyID, memberID, startedAt, stoppedAt, classification, nwStartHour, nwStartMinute, nwEndHour, nwEndMinute)
}

func assembleSleepSession(id, babyID, memberID string, startedAt, stoppedAt sql.NullTime, classification string, nwStartHour, nwStartMinute, nwEndHour, nwEndMinute sql.NullInt32) (sleep.SleepSession, error) {
	if stoppedAt.Valid {
		var classifiedWith *sleep.NightWindow
		if nwStartHour.Valid {
			start, err := sleep.NewLocalTime(int(nwStartHour.Int32), int(nwStartMinute.Int32))
			if err != nil {
				return sleep.SleepSession{}, fmt.Errorf("reconstruct night window start: %w", err)
			}
			end, err := sleep.NewLocalTime(int(nwEndHour.Int32), int(nwEndMinute.Int32))
			if err != nil {
				return sleep.SleepSession{}, fmt.Errorf("reconstruct night window end: %w", err)
			}
			nw, err := sleep.NewNightWindow(start, end)
			if err != nil {
				return sleep.SleepSession{}, fmt.Errorf("reconstruct night window: %w", err)
			}
			classifiedWith = &nw
		}
		return sleep.NewCompletedSleepSession(
			sleep.SleepSessionID(id),
			sleep.BabyID(babyID),
			sleep.FamilyMemberID(memberID),
			startedAt.Time,
			stoppedAt.Time,
			sleep.SleepClassification(classification),
			classifiedWith,
		)
	}

	return sleep.NewSleepSession(
		sleep.SleepSessionID(id),
		sleep.BabyID(babyID),
		sleep.FamilyMemberID(memberID),
		startedAt.Time,
	)
}
