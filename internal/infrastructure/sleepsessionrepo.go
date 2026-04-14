package infrastructure

import (
	"context"
	"database/sql"
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
