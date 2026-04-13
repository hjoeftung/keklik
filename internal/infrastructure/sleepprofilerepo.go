package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// PostgresSleepProfileRepository implements sleep.SleepProfileRepository using PostgreSQL.
type PostgresSleepProfileRepository struct {
	db *sql.DB
}

// NewPostgresSleepProfileRepository returns a repository backed by the given database connection.
func NewPostgresSleepProfileRepository(db *sql.DB) *PostgresSleepProfileRepository {
	return &PostgresSleepProfileRepository{db: db}
}

// Save persists a SleepProfile record.
func (r *PostgresSleepProfileRepository) Save(ctx context.Context, p sleep.SleepProfile) error {
	nw := p.NightWindow()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sleep_profiles (
			baby_id, timezone,
			night_window_start_hour, night_window_start_minute,
			night_window_end_hour,   night_window_end_minute
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (baby_id) DO UPDATE SET
			timezone                  = EXCLUDED.timezone,
			night_window_start_hour   = EXCLUDED.night_window_start_hour,
			night_window_start_minute = EXCLUDED.night_window_start_minute,
			night_window_end_hour     = EXCLUDED.night_window_end_hour,
			night_window_end_minute   = EXCLUDED.night_window_end_minute,
			updated_at                = now()`,
		string(p.BabyID()), p.Timezone(),
		nw.Start().Hour(), nw.Start().Minute(),
		nw.End().Hour(), nw.End().Minute(),
	)
	if err != nil {
		return fmt.Errorf("upsert sleep profile: %w", err)
	}
	return nil
}

// FindByBabyID loads the SleepProfile for the given baby.
func (r *PostgresSleepProfileRepository) FindByBabyID(ctx context.Context, babyID sleep.BabyID) (sleep.SleepProfile, error) {
	var timezone string
	var startHour, startMinute, endHour, endMinute int

	err := r.db.QueryRowContext(ctx, `
		SELECT timezone,
			night_window_start_hour, night_window_start_minute,
			night_window_end_hour,   night_window_end_minute
		FROM sleep_profiles WHERE baby_id = $1`, string(babyID)).
		Scan(&timezone, &startHour, &startMinute, &endHour, &endMinute)
	if err == sql.ErrNoRows {
		return sleep.SleepProfile{}, apperror.New(apperror.CodeNotFound, "sleep profile not found")
	}
	if err != nil {
		return sleep.SleepProfile{}, fmt.Errorf("query sleep profile: %w", err)
	}

	start, err := sleep.NewLocalTime(startHour, startMinute)
	if err != nil {
		return sleep.SleepProfile{}, fmt.Errorf("reconstruct night window start: %w", err)
	}

	end, err := sleep.NewLocalTime(endHour, endMinute)
	if err != nil {
		return sleep.SleepProfile{}, fmt.Errorf("reconstruct night window end: %w", err)
	}

	nightWindow, err := sleep.NewNightWindow(start, end)
	if err != nil {
		return sleep.SleepProfile{}, fmt.Errorf("reconstruct night window: %w", err)
	}

	return sleep.NewSleepProfile(babyID, timezone, nightWindow)
}
