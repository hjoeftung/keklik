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
	if stats.DiaryWindow.Start != (time.Time{}) || stats.DiaryWindow.End != (time.Time{}) {
		t.Fatalf("expected zero diary window, got %+v", stats.DiaryWindow)
	}
}

func TestComputeDiaryWindow(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0) // 21:00–07:00
	loc := time.UTC
	// now = April 29 2026 at 12:00 UTC
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	dw := computeDiaryWindow(nw, loc, now)

	// night window end today = April 29 07:00 UTC
	// diary_window.end = 07:00 - 2h = 05:00 UTC
	// diary_window.start = 05:00 - 24h = April 28 05:00 UTC
	wantEnd := time.Date(2026, time.April, 29, 5, 0, 0, 0, time.UTC)
	wantStart := time.Date(2026, time.April, 28, 5, 0, 0, 0, time.UTC)

	if !dw.End.Equal(wantEnd) {
		t.Fatalf("diary window end: want %v, got %v", wantEnd, dw.End)
	}
	if !dw.Start.Equal(wantStart) {
		t.Fatalf("diary window start: want %v, got %v", wantStart, dw.Start)
	}
}

func TestGetSleepStatsHandlerTodayTotals(t *testing.T) {
	t.Parallel()

	// Night window 21:00–07:00 UTC, diary window April 28 05:00 – April 29 05:00 UTC.
	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	// Nap: April 28 10:00–11:00 (1h, within diary window)
	nap, _ := NewCompletedSleepSession("nap-1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 28, 11, 0, 0, 0, time.UTC),
	)
	// Night sleep: April 28 21:00 – April 29 07:00 (10h), overlaps diary window end
	night, _ := NewCompletedSleepSession("night-1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)
	// Session outside diary window (April 27 10:00–11:00): should be excluded
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
	if stats.Today.TotalNapSeconds != wantNap {
		t.Fatalf("TotalNapSeconds: want %.0f, got %.0f", wantNap, stats.Today.TotalNapSeconds)
	}

	wantSleep := (1*time.Hour + 10*time.Hour).Seconds()
	if stats.Today.TotalSleepSeconds != wantSleep {
		t.Fatalf("TotalSleepSeconds: want %.0f, got %.0f", wantSleep, stats.Today.TotalSleepSeconds)
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
