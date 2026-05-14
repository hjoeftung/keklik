package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGetSleepStatsHandlerRequiresTimezone(t *testing.T) {
	t.Parallel()

	h := NewGetSleepStatsHandler(&stubSleepSessionHistoryRepository{}, &stubNightWindowRepository{})
	_, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1"})
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestGetSleepStatsHandlerReturnsZeroStatsWhenNoNightWindow(t *testing.T) {
	t.Parallel()

	h := NewGetSleepStatsHandler(&stubSleepSessionHistoryRepository{}, &stubNightWindowRepository{})
	stats, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats.Days) > 0 && (stats.Days[0].TotalSleepSeconds != 0 || stats.Days[0].TotalNapSeconds != 0) {
		t.Fatalf("expected zero today stats, got %+v", stats.Days[0])
	}
}

func TestGetSleepStatsHandlerTodayTotals(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	// now = April 29 2026 at 12:00 UTC; calendar day = April 29 00:00–24:00 UTC
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	// Nap: April 29 09:00–10:00 (1h, within today's calendar day)
	nap, _ := NewCompletedSleepSession("nap-1", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 10, 0, 0, 0, time.UTC),
	)
	// Night sleep: April 28 21:00 – April 29 07:00 (10h), overlaps today
	night, _ := NewCompletedSleepSession("night-1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)
	// Session from two days ago: should be excluded
	outside, _ := NewCompletedSleepSession("old-1", "baby-1", "m-1",
		time.Date(2026, time.April, 27, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 27, 11, 0, 0, 0, time.UTC),
	)

	sessionRepo := &stubSleepSessionHistoryRepository{sessions: []SleepSession{nap, night, outside}}
	windowRepo := &stubNightWindowRepository{windows: []NightWindow{nw}}

	h := &GetSleepStatsHandler{sessions: sessionRepo, windows: windowRepo, now: func() time.Time { return now }}

	stats, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantNap := (1 * time.Hour).Seconds()
	if stats.Days[0].TotalNapSeconds != wantNap {
		t.Fatalf("TotalNapSeconds: want %.0f, got %.0f", wantNap, stats.Days[0].TotalNapSeconds)
	}

	// Night sleep (Apr 28 21:00–Apr 29 07:00) is in today's previous night window → sets WokeAt, not counted.
	wantSleep := (1 * time.Hour).Seconds() // 0h night + 1h nap
	if stats.Days[0].TotalSleepSeconds != wantSleep {
		t.Fatalf("TotalSleepSeconds: want %.0f, got %.0f", wantSleep, stats.Days[0].TotalSleepSeconds)
	}
}

func TestGetSleepStatsSummaryExcludesToday(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	// Single nap today (should not appear in averages)
	todayNap, _ := NewCompletedSleepSession("today-nap", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 10, 0, 0, 0, time.UTC),
	)

	sessionRepo := &stubSleepSessionHistoryRepository{sessions: []SleepSession{todayNap}}
	windowRepo := &stubNightWindowRepository{windows: []NightWindow{nw}}

	h := &GetSleepStatsHandler{sessions: sessionRepo, windows: windowRepo, now: func() time.Time { return now }}

	stats, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All averages should be zero because the only session is today's.
	for _, key := range []string{"7d", "14d", "30d", "90d"} {
		p := stats.Summary[key]
		if p.AvgSleepSeconds != 0 || p.AvgNapSeconds != 0 {
			t.Fatalf("%s: expected zero averages, got sleep=%.0f nap=%.0f", key, p.AvgSleepSeconds, p.AvgNapSeconds)
		}
	}
}

// TestGetSleepStatsSummaryAnchorConsistency verifies that historical active time
// is anchored at the wake-up time from night sleep, not at midnight. Without the
// fix the anchor would be 00:00 and active would be 23h instead of 16h.
func TestGetSleepStatsSummaryAnchorConsistency(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	// Night sleep Apr 27 21:00 – Apr 28 07:00; anchor for Apr 28 is 07:00.
	night, _ := NewCompletedSleepSession("night-1", "baby-1", "m-1",
		time.Date(2026, time.April, 27, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 28, 7, 0, 0, 0, time.UTC),
	)
	// Nap Apr 28 09:00–10:00 (1h).
	nap, _ := NewCompletedSleepSession("nap-1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 28, 10, 0, 0, 0, time.UTC),
	)

	h := &GetSleepStatsHandler{
		sessions: &stubSleepSessionHistoryRepository{sessions: []SleepSession{night, nap}},
		windows:  &stubNightWindowRepository{windows: []NightWindow{nw}},
		now:      func() time.Time { return now },
	}

	stats, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Apr 28: WokeAt=07:00, NightStartedAt=nil → active = dayEnd−07:00 − 1h nap = 17h−1h = 16h.
	// Apr 27: WokeAt=midnight, NightStartedAt=Apr 27 21:00 (night starts tonight) → active = 21h.
	// Apr 22–26: no sessions → active = 24h each.
	want7dAvg := (16*3600.0 + 21*3600.0 + 5*86400.0) / 7.0
	got := stats.Summary["7d"].AvgActiveSeconds
	if got != want7dAvg {
		t.Fatalf("7d AvgActiveSeconds: want %.2f (anchor at 07:00), got %.2f", want7dAvg, got)
	}
}

// TestGetSleepStatsSummaryNightSleepAttributionByStartedAt verifies that a night
// sleep is attributed to the day it started on (StartedAt), not the day it ended
// on. A session starting Apr 27 and ending Apr 28 should count only for Apr 27,
// not twice (old overlap-based code would include it in both Apr 27 and Apr 28).
func TestGetSleepStatsSummaryNightSleepAttributionByStartedAt(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	// Night sleep started Apr 27 (before Apr 28 dayStart) — counted for Apr 27 only.
	night, _ := NewCompletedSleepSession("night-1", "baby-1", "m-1",
		time.Date(2026, time.April, 27, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 28, 7, 0, 0, 0, time.UTC),
	)

	h := &GetSleepStatsHandler{
		sessions: &stubSleepSessionHistoryRepository{sessions: []SleepSession{night}},
		windows:  &stubNightWindowRepository{windows: []NightWindow{nw}},
		now:      func() time.Time { return now },
	}

	stats, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Night counted once for Apr 27 (36000s); Apr 28 has 0 sleep (night started Apr 27).
	// Old overlap-based code would give 72000/7 (counted for both Apr 27 and Apr 28).
	want7dAvg := 36000.0 / 7.0
	got := stats.Summary["7d"].AvgSleepSeconds
	if got != want7dAvg {
		t.Fatalf("7d AvgSleepSeconds: want %.4f (night on Apr 27 only), got %.4f", want7dAvg, got)
	}
}

// TestGetSleepStatsSummaryCrossMidnightNapCaptured verifies that a nap starting
// just before midnight on day D-1 and ending just after midnight on day D is
// attributed to day D via the 24-hour pre-filter buffer.
func TestGetSleepStatsSummaryCrossMidnightNapCaptured(t *testing.T) {
	t.Parallel()

	// Night window 00:00–06:00 means a session starting at 23:30 has only 15 min
	// overlap with the window (< half of 45 min) and is therefore a Nap.
	nw := mustNightWindow(t, 0, 0, 6, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	// Nap Apr 27 23:30 – Apr 28 00:15 (45 min); stoppedAt is after Apr 28 dayStart.
	crossMidnightNap, _ := NewCompletedSleepSession("nap-1", "baby-1", "m-1",
		time.Date(2026, time.April, 27, 23, 30, 0, 0, time.UTC),
		time.Date(2026, time.April, 28, 0, 15, 0, 0, time.UTC),
	)

	h := &GetSleepStatsHandler{
		sessions: &stubSleepSessionHistoryRepository{sessions: []SleepSession{crossMidnightNap}},
		windows:  &stubNightWindowRepository{windows: []NightWindow{nw}},
		now:      func() time.Time { return now },
	}

	stats, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Apr 28 nap total = 45 min = 2700s; all other days = 0.
	want7dAvg := 2700.0 / 7.0
	got := stats.Summary["7d"].AvgNapSeconds
	if got != want7dAvg {
		t.Fatalf("7d AvgNapSeconds: want %.4f (cross-midnight nap), got %.4f", want7dAvg, got)
	}
}

// TestGetSleepStatsSummaryZeroSessions verifies that all sleep and nap averages
// are zero when there are no historical sessions.
func TestGetSleepStatsSummaryZeroSessions(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	h := &GetSleepStatsHandler{
		sessions: &stubSleepSessionHistoryRepository{sessions: nil},
		windows:  &stubNightWindowRepository{windows: []NightWindow{nw}},
		now:      func() time.Time { return now },
	}

	stats, err := h.Handle(context.Background(), GetSleepStatsQuery{BabyID: "baby-1", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, key := range []string{"7d", "14d", "30d", "90d"} {
		p := stats.Summary[key]
		if p.AvgSleepSeconds != 0 || p.AvgNapSeconds != 0 {
			t.Fatalf("%s: expected zero sleep/nap averages, got sleep=%.0f nap=%.0f", key, p.AvgSleepSeconds, p.AvgNapSeconds)
		}
	}
}
