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

	dr, err := periodToDateRange("today", "UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now().UTC()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
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

	dr, err := periodToDateRange("7d", "UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now().UTC()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	expectedStart := time.Date(sevenDaysAgo.Year(), sevenDaysAgo.Month(), sevenDaysAgo.Day(), 0, 0, 0, 0, time.UTC)

	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
	// end should be approximately now (within a second)
	if dr.End().Before(now.Add(-time.Second)) || dr.End().After(now.Add(time.Second)) {
		t.Errorf("end: want approximately %v, got %v", now, dr.End())
	}
}

func TestPeriodToDateRange14d(t *testing.T) {
	t.Parallel()

	dr, err := periodToDateRange("14d", "UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now().UTC()
	fourteenDaysAgo := now.AddDate(0, 0, -14)
	expectedStart := time.Date(fourteenDaysAgo.Year(), fourteenDaysAgo.Month(), fourteenDaysAgo.Day(), 0, 0, 0, 0, time.UTC)

	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
}

func TestPeriodToDateRangeInvalidPeriodReturnsError(t *testing.T) {
	t.Parallel()

	_, err := periodToDateRange("30d", "UTC")
	if !errors.Is(err, ErrInvalidSleepHistoryPeriod) {
		t.Fatalf("expected ErrInvalidSleepHistoryPeriod, got %v", err)
	}
}

func TestPeriodToDateRangeInvalidTimezoneReturnsError(t *testing.T) {
	t.Parallel()

	_, err := periodToDateRange("7d", "Not/ATimezone")
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

	dr, err := periodToDateRange("today", "America/New_York")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now().In(loc)
	expectedStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
	expectedEnd := expectedStart.AddDate(0, 0, 1)

	if !dr.Start().Equal(expectedStart) {
		t.Errorf("start: want %v, got %v", expectedStart, dr.Start())
	}
	if !dr.End().Equal(expectedEnd) {
		t.Errorf("end: want %v, got %v", expectedEnd, dr.End())
	}
}

// TestPeriodToDateRangeDSTSpringForwardToday verifies that "today" boundary
// computation is correct on a DST spring-forward night (2026-03-08 in New York).
// The local midnight-to-midnight span is 23 hours in UTC on that day.
func TestPeriodToDateRangeDSTSpringForwardToday(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-03-08: spring-forward night. Midnight starts at UTC-5; after the
	// transition the next midnight is at UTC-4, so the UTC window is 23h long.
	startLocal := time.Date(2026, time.March, 8, 0, 0, 0, 0, loc)
	endLocal := time.Date(2026, time.March, 9, 0, 0, 0, 0, loc)
	expectedDuration := endLocal.UTC().Sub(startLocal.UTC())

	// 23 hours because clocks spring forward.
	if expectedDuration != 23*time.Hour {
		t.Fatalf("expected DST spring-forward day to be 23h in UTC, got %v", expectedDuration)
	}
}

// TestPeriodToDateRangeDSTFallBackToday verifies that "today" boundary
// computation is correct on a DST fall-back night (2026-11-01 in New York).
// The local midnight-to-midnight span is 25 hours in UTC on that day.
func TestPeriodToDateRangeDSTFallBackToday(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-11-01: fall-back night. Midnight starts at UTC-4; after the
	// transition the next midnight is at UTC-5, so the UTC window is 25h long.
	startLocal := time.Date(2026, time.November, 1, 0, 0, 0, 0, loc)
	endLocal := time.Date(2026, time.November, 2, 0, 0, 0, 0, loc)
	expectedDuration := endLocal.UTC().Sub(startLocal.UTC())

	// 25 hours because clocks fall back.
	if expectedDuration != 25*time.Hour {
		t.Fatalf("expected DST fall-back day to be 25h in UTC, got %v", expectedDuration)
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

