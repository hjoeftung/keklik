package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- stub repositories ---

type stubSleepSessionHistoryRepository struct {
	sessions []SleepSession
	err      error
	// captured arguments from last call
	capturedBabyID    BabyID
	capturedDateRange DateRange
}

func (r *stubSleepSessionHistoryRepository) FindByBabyIDAndDateRange(_ context.Context, babyID BabyID, dr DateRange) ([]SleepSession, error) {
	r.capturedBabyID = babyID
	r.capturedDateRange = dr
	return r.sessions, r.err
}

type stubSleepProfileRepository struct {
	profile SleepProfile
	err     error
}

func (r *stubSleepProfileRepository) Save(_ context.Context, _ SleepProfile) error {
	return nil
}

func (r *stubSleepProfileRepository) FindByBabyID(_ context.Context, _ BabyID) (SleepProfile, error) {
	return r.profile, r.err
}

// --- periodToDateRange unit tests ---

func TestPeriodToDateRangeToday(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	dr, err := periodToDateRange("today", "UTC", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	todayMidnight := time.Date(2026, time.April, 16, 0, 0, 0, 0, time.UTC)
	tomorrowMidnight := todayMidnight.AddDate(0, 0, 1)

	if !dr.Start().Equal(todayMidnight) {
		t.Errorf("start: want %v, got %v", todayMidnight, dr.Start())
	}
	if !dr.End().Equal(tomorrowMidnight) {
		t.Errorf("end: want %v, got %v", tomorrowMidnight, dr.End())
	}
}

func TestPeriodToDateRange7d(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	dr, err := periodToDateRange("7d", "UTC", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, time.April, 9, 0, 0, 0, 0, time.UTC)
	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
	if !dr.End().Equal(now) {
		t.Errorf("end: want %v, got %v", now, dr.End())
	}
}

func TestPeriodToDateRange14d(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	dr, err := periodToDateRange("14d", "UTC", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)
	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
	if !dr.End().Equal(now) {
		t.Errorf("end: want %v, got %v", now, dr.End())
	}
}

func TestPeriodToDateRangeInvalidPeriodReturnsError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	_, err := periodToDateRange("30d", "UTC", now)
	if !errors.Is(err, ErrInvalidSleepHistoryPeriod) {
		t.Fatalf("expected ErrInvalidSleepHistoryPeriod, got %v", err)
	}
}

func TestPeriodToDateRangeInvalidTimezoneReturnsError(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	_, err := periodToDateRange("7d", "Not/ATimezone", now)
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}

// TestPeriodToDateRangeTodayInEasternTimezone verifies that "today" uses the
// local calendar day of the supplied timezone, not UTC.
func TestPeriodToDateRangeTodayInEasternTimezone(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-04-16 14:30 UTC = 2026-04-16 10:30 America/New_York.
	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	dr, err := periodToDateRange("today", "America/New_York", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, time.April, 16, 0, 0, 0, 0, loc).UTC()
	expectedEnd := expectedStart.AddDate(0, 0, 1)

	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
	if !dr.End().Equal(expectedEnd) {
		t.Errorf("end: want %v, got %v", expectedEnd, dr.End())
	}
}

// TestPeriodToDateRangeDSTSpringForwardToday verifies that periodToDateRange
// produces a 23-hour UTC window for "today" on the spring-forward night
// (2026-03-08 in New York).
func TestPeriodToDateRangeDSTSpringForwardToday(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// Noon local time on the spring-forward day.
	now := time.Date(2026, time.March, 8, 12, 0, 0, 0, loc).UTC()
	dr, err := periodToDateRange("today", "America/New_York", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Midnight starts at UTC-5; next midnight is at UTC-4 → 23h window.
	expectedStart := time.Date(2026, time.March, 8, 0, 0, 0, 0, loc).UTC()
	expectedEnd := time.Date(2026, time.March, 9, 0, 0, 0, 0, loc).UTC()
	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
	if !dr.End().Equal(expectedEnd) {
		t.Errorf("end: want %v, got %v", expectedEnd, dr.End())
	}
	if dr.End().Sub(dr.Start()) != 23*time.Hour {
		t.Errorf("expected 23h window, got %v", dr.End().Sub(dr.Start()))
	}
}

// TestPeriodToDateRangeDSTFallBackToday verifies that periodToDateRange
// produces a 25-hour UTC window for "today" on the fall-back night
// (2026-11-01 in New York).
func TestPeriodToDateRangeDSTFallBackToday(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// Noon local time on the fall-back day.
	now := time.Date(2026, time.November, 1, 12, 0, 0, 0, loc).UTC()
	dr, err := periodToDateRange("today", "America/New_York", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Midnight starts at UTC-4; next midnight is at UTC-5 → 25h window.
	expectedStart := time.Date(2026, time.November, 1, 0, 0, 0, 0, loc).UTC()
	expectedEnd := time.Date(2026, time.November, 2, 0, 0, 0, 0, loc).UTC()
	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
	if !dr.End().Equal(expectedEnd) {
		t.Errorf("end: want %v, got %v", expectedEnd, dr.End())
	}
	if dr.End().Sub(dr.Start()) != 25*time.Hour {
		t.Errorf("expected 25h window, got %v", dr.End().Sub(dr.Start()))
	}
}

// --- GetSleepHistoryHandler tests ---

func TestGetSleepHistoryHandlerReturnsEmptySliceWhenNoSessions(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	profile, _ := NewSleepProfile(BabyID("baby-1"), "UTC", nw)

	repo := &stubSleepSessionHistoryRepository{sessions: nil}
	profileRepo := &stubSleepProfileRepository{profile: profile}
	h := NewGetSleepHistoryHandler(repo, profileRepo)

	sessions, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID: BabyID("baby-1"),
		Period: "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestGetSleepHistoryHandlerReturnsSessionsForPeriod(t *testing.T) {
	t.Parallel()

	startedAt := time.Now().UTC().Add(-2 * time.Hour)
	session := mustSleepSession(t, startedAt)

	nw := mustNightWindow(t, 21, 0, 7, 0)
	profile, _ := NewSleepProfile(BabyID("baby-1"), "UTC", nw)

	repo := &stubSleepSessionHistoryRepository{sessions: []SleepSession{session}}
	profileRepo := &stubSleepProfileRepository{profile: profile}
	h := NewGetSleepHistoryHandler(repo, profileRepo)

	sessions, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID: BabyID("baby-1"),
		Period: "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
}

func TestGetSleepHistoryHandlerUsesProfileTimezoneWhenQueryTimezoneEmpty(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	profile, _ := NewSleepProfile(BabyID("baby-1"), "America/New_York", nw)

	repo := &stubSleepSessionHistoryRepository{}
	profileRepo := &stubSleepProfileRepository{profile: profile}
	h := NewGetSleepHistoryHandler(repo, profileRepo)

	_, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID:   BabyID("baby-1"),
		Timezone: "", // should be resolved from profile
		Period:   "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the date range passed to the repo uses New York local midnight.
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)
	sevenAgo := now.AddDate(0, 0, -7)
	expectedStart := time.Date(sevenAgo.Year(), sevenAgo.Month(), sevenAgo.Day(), 0, 0, 0, 0, loc).UTC()

	if !repo.capturedDateRange.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, repo.capturedDateRange.Start())
	}
}

func TestGetSleepHistoryHandlerInvalidPeriodReturnsError(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	profile, _ := NewSleepProfile(BabyID("baby-1"), "UTC", nw)

	repo := &stubSleepSessionHistoryRepository{}
	profileRepo := &stubSleepProfileRepository{profile: profile}
	h := NewGetSleepHistoryHandler(repo, profileRepo)

	_, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID: BabyID("baby-1"),
		Period: "30d",
	})
	if !errors.Is(err, ErrInvalidSleepHistoryPeriod) {
		t.Fatalf("expected ErrInvalidSleepHistoryPeriod, got %v", err)
	}
}
