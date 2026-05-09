package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hjoeftung/keklik/internal/infrastructure"
	"github.com/hjoeftung/keklik/internal/sleep"
)

func main() {
	babyID := flag.String("baby-id", "", "baby UUID to seed sessions for (defaults to first baby in DB)")
	memberID := flag.String("member-id", "", "family member UUID to attribute sessions to (defaults to first member in DB)")
	days := flag.Int("days", 30, "number of past days to generate sessions for")
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := infrastructure.OpenDB(databaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if *babyID == "" {
		if err := db.QueryRowContext(ctx, `SELECT id FROM babies ORDER BY created_at LIMIT 1`).Scan(babyID); err != nil {
			log.Fatalf("no baby found in DB (create a family first): %v", err)
		}
	}

	if *memberID == "" {
		if err := db.QueryRowContext(ctx, `SELECT id FROM family_members ORDER BY created_at LIMIT 1`).Scan(memberID); err != nil {
			log.Fatalf("no family member found in DB: %v", err)
		}
	}

	log.Printf("seeding %d days of sleep data for baby=%s member=%s", *days, *babyID, *memberID)

	nightWindowRepo := infrastructure.NewPostgresNightWindowRepository(db)
	transactor := infrastructure.NewPostgresTransactor(db)
	setNightWindow := sleep.NewSetNightWindowHandler(nightWindowRepo, transactor)

	// Backdate the night window so it covers all seeded sessions.
	effectiveFrom := time.Now().UTC().AddDate(0, 0, -(*days + 1))
	if err := setNightWindow.Handle(ctx, sleep.SetNightWindowCommand{
		BabyID:                 sleep.BabyID(*babyID),
		NightWindowStartHour:   21,
		NightWindowStartMinute: 0,
		NightWindowEndHour:     7,
		NightWindowEndMinute:   0,
		EffectiveFrom:          effectiveFrom,
	}); err != nil {
		log.Fatalf("failed to set night window: %v", err)
	}
	log.Printf("night window set: 21:00–07:00 effective from %s", effectiveFrom.Format(time.DateOnly))

	handler := sleep.NewLogPastSleepHandler(infrastructure.NewPostgresSleepSessionRepository(db))

	rng := rand.New(rand.NewSource(42))
	inserted := 0
	skipped := 0

	now := time.Now().UTC()
	loc := now.Location()

	for d := *days; d >= 1; d-- {
		midnight := time.Date(now.Year(), now.Month(), now.Day()-d, 0, 0, 0, 0, loc)

		// Night sleep: starts between 19:00–21:00, ends 8–11 hours later.
		nightStart := midnight.Add(time.Duration(19+rng.Intn(3)) * time.Hour).Add(time.Duration(rng.Intn(60)) * time.Minute)
		nightDuration := time.Duration(8+rng.Intn(4)) * time.Hour
		nightEnd := nightStart.Add(nightDuration)

		if nightEnd.Before(now) {
			if err := logSession(ctx, handler, *babyID, *memberID, nightStart, nightEnd); err != nil {
				log.Printf("skip night sleep on day -%d: %v", d, err)
				skipped++
			} else {
				inserted++
			}
		}

		// Naps: 1–3 per day, each 30–120 minutes, spread through daytime hours.
		napCount := 1 + rng.Intn(3)
		napStartHour := 8
		for n := 0; n < napCount; n++ {
			napStart := midnight.Add(time.Duration(napStartHour+rng.Intn(3)) * time.Hour).Add(time.Duration(rng.Intn(60)) * time.Minute)
			napDuration := time.Duration(30+rng.Intn(90)) * time.Minute
			napEnd := napStart.Add(napDuration)
			napStartHour += 3 + rng.Intn(2)

			if napEnd.Before(now) && napEnd.Before(nightStart) {
				if err := logSession(ctx, handler, *babyID, *memberID, napStart, napEnd); err != nil {
					log.Printf("skip nap %d on day -%d: %v", n+1, d, err)
					skipped++
				} else {
					inserted++
				}
			}
		}
	}

	fmt.Printf("done: %d sessions inserted, %d skipped\n", inserted, skipped)
}

func logSession(ctx context.Context, h *sleep.LogPastSleepHandler, babyID, memberID string, start, end time.Time) error {
	_, err := h.Handle(ctx, sleep.LogPastSleepCommand{
		BabyID:            sleep.BabyID(babyID),
		CreatedByMemberID: sleep.FamilyMemberID(memberID),
		StartedAt:         start,
		StoppedAt:         end,
	})
	return err
}
