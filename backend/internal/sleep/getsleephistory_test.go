package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubSleepSessionHistoryRepository struct {
	sessions []SleepSession
	err      error
}

func (r *stubSleepSessionHistoryRepository) FindByBabyIDAndDateRange(_ context.Context, _ BabyID, _ DateRange) ([]SleepSession, error) {
	return r.sessions, r.err
}

type stubNightWindowRepository struct {
	windows []NightWindow
	err     error
}

func (r *stubNightWindowRepository) Save(_ context.Context, _ NightWindow) error {
	return nil
}

func (r *stubNightWindowRepository) DeleteByIDs(_ context.Context, _ []NightWindowID) error {
	return nil
}

func (r *stubNightWindowRepository) FindByBabyID(_ context.Context, _ BabyID) ([]NightWindow, error) {
	return r.windows, r.err
}

func TestPeriodToDateRangeToday(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	dr, err := periodToDateRange("today", "UTC", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	todayMidnight := time.Date(2026, time.April, 16, 0, 0, 0, 0, time.UTC)
	if !dr.Start().Equal(todayMidnight) {
		t.Fatalf("expected start %v, got %v", todayMidnight, dr.Start())
	}
}

func TestGetSleepHistoryHandlerReturnsEmptySliceWhenNoSessions(t *testing.T) {
	t.Parallel()

	repo := &stubSleepSessionHistoryRepository{}
	windows := &stubNightWindowRepository{windows: []NightWindow{mustNightWindow(t, 21, 0, 7, 0)}}
	h := NewGetSleepHistoryHandler(repo, windows)

	entries, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID:   BabyID("baby-1"),
		Timezone: "UTC",
		Period:   "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestGetSleepHistoryHandlerClassifiesSessionsFromApplicableWindow(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.April, 16, 22, 0, 0, 0, time.UTC)
	stop := start.Add(8 * time.Hour)
	session, err := NewCompletedSleepSession(
		SleepSessionID("session-1"),
		BabyID("baby-1"),
		FamilyMemberID("member-1"),
		start,
		stop,
	)
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}

	h := NewGetSleepHistoryHandler(
		&stubSleepSessionHistoryRepository{sessions: []SleepSession{session}},
		&stubNightWindowRepository{windows: []NightWindow{mustNightWindow(t, 21, 0, 7, 0)}},
	)

	entries, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID:   BabyID("baby-1"),
		Timezone: "UTC",
		Period:   "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Classification != SleepClassificationNight {
		t.Fatalf("expected night classification, got %q", entries[0].Classification)
	}
}

func TestGetSleepHistoryHandlerRequiresTimezone(t *testing.T) {
	t.Parallel()

	h := NewGetSleepHistoryHandler(&stubSleepSessionHistoryRepository{}, &stubNightWindowRepository{})
	_, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID: BabyID("baby-1"),
		Period: "7d",
	})
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestPeriodToDateRangeNd(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)

	cases := []struct {
		period    string
		wantStart time.Time
	}{
		{"1d", time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC)},
		{"7d", time.Date(2026, time.April, 9, 0, 0, 0, 0, time.UTC)},
		{"30d", time.Date(2026, time.March, 17, 0, 0, 0, 0, time.UTC)},
		{"120d", time.Date(2025, time.December, 17, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		dr, err := periodToDateRange(tc.period, "UTC", now)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.period, err)
		}
		if !dr.Start().Equal(tc.wantStart) {
			t.Fatalf("%s: expected start %v, got %v", tc.period, tc.wantStart, dr.Start())
		}
		if !dr.End().Equal(now) {
			t.Fatalf("%s: expected end %v, got %v", tc.period, now, dr.End())
		}
	}
}

func TestPeriodToDateRangeInvalidValues(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	invalid := []string{"0d", "121d", "999d", "d", "-1d", "7", "week"}
	for _, period := range invalid {
		_, err := periodToDateRange(period, "UTC", now)
		if !errors.Is(err, ErrInvalidSleepHistoryPeriod) {
			t.Fatalf("%q: expected ErrInvalidSleepHistoryPeriod, got %v", period, err)
		}
	}
}

func TestGetSleepHistoryHandlerInvalidPeriodReturnsError(t *testing.T) {
	t.Parallel()

	h := NewGetSleepHistoryHandler(&stubSleepSessionHistoryRepository{}, &stubNightWindowRepository{})
	_, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID:   BabyID("baby-1"),
		Timezone: "UTC",
		Period:   "121d",
	})
	if !errors.Is(err, ErrInvalidSleepHistoryPeriod) {
		t.Fatalf("expected ErrInvalidSleepHistoryPeriod, got %v", err)
	}
}
