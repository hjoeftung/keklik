package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/hjoeftung/keklik/internal/sleep"
)

// PostgresNightWindowRepository implements sleep.NightWindowRepository using PostgreSQL.
type PostgresNightWindowRepository struct {
	db *sql.DB
}

// NewPostgresNightWindowRepository returns a repository backed by the given database connection.
func NewPostgresNightWindowRepository(db *sql.DB) *PostgresNightWindowRepository {
	return &PostgresNightWindowRepository{db: db}
}

// Save inserts or updates a NightWindow record.
func (r *PostgresNightWindowRepository) Save(ctx context.Context, nw sleep.NightWindow) error {
	var effectiveTo *time.Time
	if t := nw.EffectiveTo(); t != nil {
		utc := t.UTC()
		effectiveTo = &utc
	}

	_, err := querierFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO baby_night_windows (
			id, baby_id,
			start_hour, start_minute,
			end_hour, end_minute,
			effective_from, effective_to
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			start_hour     = EXCLUDED.start_hour,
			start_minute   = EXCLUDED.start_minute,
			end_hour       = EXCLUDED.end_hour,
			end_minute     = EXCLUDED.end_minute,
			effective_from = EXCLUDED.effective_from,
			effective_to   = EXCLUDED.effective_to`,
		string(nw.ID()), string(nw.BabyID()),
		nw.Start().Hour(), nw.Start().Minute(),
		nw.End().Hour(), nw.End().Minute(),
		nw.EffectiveFrom().UTC(), effectiveTo,
	)
	if err != nil {
		return fmt.Errorf("upsert night window: %w", err)
	}
	return nil
}

// DeleteByIDs removes NightWindows by their IDs. A no-op when ids is empty.
func (r *PostgresNightWindowRepository) DeleteByIDs(ctx context.Context, ids []sleep.NightWindowID) error {
	if len(ids) == 0 {
		return nil
	}
	raw := make([]string, len(ids))
	for i, id := range ids {
		raw[i] = string(id)
	}
	_, err := querierFromContext(ctx, r.db).ExecContext(ctx,
		`DELETE FROM baby_night_windows WHERE id = ANY($1)`, pq.Array(raw))
	if err != nil {
		return fmt.Errorf("delete night windows: %w", err)
	}
	return nil
}

// FindByBabyID returns all NightWindows for a baby, ordered by effective_from ASC.
func (r *PostgresNightWindowRepository) FindByBabyID(ctx context.Context, babyID sleep.BabyID) ([]sleep.NightWindow, error) {
	rows, err := querierFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, baby_id,
		       start_hour, start_minute,
		       end_hour, end_minute,
		       effective_from, effective_to
		FROM baby_night_windows
		WHERE baby_id = $1
		ORDER BY effective_from ASC`, string(babyID))
	if err != nil {
		return nil, fmt.Errorf("query night windows: %w", err)
	}
	defer rows.Close()

	var windows []sleep.NightWindow
	for rows.Next() {
		nw, err := scanNightWindow(rows)
		if err != nil {
			return nil, err
		}
		windows = append(windows, nw)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate night windows: %w", err)
	}
	return windows, nil
}

func scanNightWindow(s interface{ Scan(...any) error }) (sleep.NightWindow, error) {
	var (
		id, babyID             string
		startHour, startMinute int
		endHour, endMinute     int
		rawEffectiveFrom       time.Time
		rawEffectiveTo         sql.NullTime
	)

	err := s.Scan(
		&id, &babyID,
		&startHour, &startMinute,
		&endHour, &endMinute,
		&rawEffectiveFrom, &rawEffectiveTo,
	)
	if err != nil {
		return sleep.NightWindow{}, fmt.Errorf("scan night window: %w", err)
	}

	start, err := sleep.NewLocalTime(startHour, startMinute)
	if err != nil {
		return sleep.NightWindow{}, fmt.Errorf("reconstruct night window start: %w", err)
	}

	end, err := sleep.NewLocalTime(endHour, endMinute)
	if err != nil {
		return sleep.NightWindow{}, fmt.Errorf("reconstruct night window end: %w", err)
	}

	var effectiveTo *time.Time
	if rawEffectiveTo.Valid {
		t := rawEffectiveTo.Time
		effectiveTo = &t
	}

	nw, err := sleep.NewNightWindow(
		sleep.NightWindowID(id),
		sleep.BabyID(babyID),
		start, end,
		rawEffectiveFrom,
		effectiveTo,
	)
	if err != nil {
		return sleep.NightWindow{}, fmt.Errorf("reconstruct night window: %w", err)
	}
	return nw, nil
}
