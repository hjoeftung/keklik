package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	var stoppedAt *string
	if t, ok := s.StoppedAt(); ok {
		ts := t.UTC().Format("2006-01-02T15:04:05.999999999Z")
		stoppedAt = &ts
	}

	var classification string
	if c := s.Classification(); c != sleep.SleepClassificationUnknown {
		classification = string(c)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sleep_sessions (
			id, baby_id, created_by_member_id,
			started_at, stopped_at,
			classification, classification_rule_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			stopped_at                  = EXCLUDED.stopped_at,
			classification              = EXCLUDED.classification,
			classification_rule_version = EXCLUDED.classification_rule_version,
			updated_by_member_id        = EXCLUDED.created_by_member_id,
			updated_at                  = now()`,
		string(s.ID()),
		string(s.BabyID()),
		string(s.CreatedByMemberID()),
		s.StartedAt().UTC(),
		stoppedAt,
		classification,
		int(s.ClassificationRuleVersion()),
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
		       classification, classification_rule_version
		FROM sleep_sessions WHERE id = $1`, string(id))

	return scanSleepSession(row)
}

// FindActiveByBabyID loads the active (not yet stopped) sleep session for a baby.
// Returns apperror with CodeNotFound when no active session exists.
func (r *PostgresSleepSessionRepository) FindActiveByBabyID(ctx context.Context, babyID sleep.BabyID) (sleep.SleepSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id,
		       started_at, stopped_at,
		       classification, classification_rule_version
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
		       classification, classification_rule_version
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

// FindMostRecentByBabyID returns the most recently started session for a baby
// (active or stopped). The bool is false when no session exists.
func (r *PostgresSleepSessionRepository) FindMostRecentByBabyID(ctx context.Context, babyID sleep.BabyID) (sleep.SleepSession, bool, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, baby_id, created_by_member_id,
		       started_at, stopped_at,
		       classification, classification_rule_version
		FROM sleep_sessions
		WHERE baby_id = $1
		ORDER BY started_at DESC
		LIMIT 1`, string(babyID))

	s, err := scanSleepSession(row)
	if err != nil {
		var appErr apperror.AppError
		if errors.As(err, &appErr) && appErr.Code == apperror.CodeNotFound {
			return sleep.SleepSession{}, false, nil
		}
		return sleep.SleepSession{}, false, err
	}
	return s, true, nil
}

func scanSleepSession(row *sql.Row) (sleep.SleepSession, error) {
	var (
		id, babyID, memberID string
		classification       string
		ruleVersion          int
		startedAt            sql.NullTime
		stoppedAt            sql.NullTime
	)

	err := row.Scan(&id, &babyID, &memberID, &startedAt, &stoppedAt, &classification, &ruleVersion)
	if err == sql.ErrNoRows {
		return sleep.SleepSession{}, apperror.New(apperror.CodeNotFound, "sleep session not found")
	}
	if err != nil {
		return sleep.SleepSession{}, fmt.Errorf("scan sleep session: %w", err)
	}

	return assembleSleepSession(id, babyID, memberID, startedAt, stoppedAt, classification, ruleVersion)
}

func scanSleepSessionRows(rows *sql.Rows) (sleep.SleepSession, error) {
	var (
		id, babyID, memberID string
		classification       string
		ruleVersion          int
		startedAt            sql.NullTime
		stoppedAt            sql.NullTime
	)

	if err := rows.Scan(&id, &babyID, &memberID, &startedAt, &stoppedAt, &classification, &ruleVersion); err != nil {
		return sleep.SleepSession{}, fmt.Errorf("scan sleep session row: %w", err)
	}

	return assembleSleepSession(id, babyID, memberID, startedAt, stoppedAt, classification, ruleVersion)
}

func assembleSleepSession(id, babyID, memberID string, startedAt, stoppedAt sql.NullTime, classification string, ruleVersion int) (sleep.SleepSession, error) {
	if stoppedAt.Valid {
		return sleep.NewCompletedSleepSession(
			sleep.SleepSessionID(id),
			sleep.BabyID(babyID),
			sleep.FamilyMemberID(memberID),
			startedAt.Time,
			stoppedAt.Time,
			sleep.SleepClassification(classification),
			sleep.ClassificationRuleVersion(ruleVersion),
		)
	}

	return sleep.NewSleepSession(
		sleep.SleepSessionID(id),
		sleep.BabyID(babyID),
		sleep.FamilyMemberID(memberID),
		startedAt.Time,
	)
}
