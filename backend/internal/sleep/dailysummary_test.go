package sleep

import (
	"context"
	"testing"
	"time"
)

// mustNap builds a completed nap session between start and stop.
func mustNap(t *testing.T, id string, start, stop time.Time) SleepSession {
	t.Helper()
	s, err := NewCompletedSleepSession(
		SleepSessionID(id), BabyID("baby-1"), FamilyMemberID("member-1"),
		start, stop, SleepClassificationNap, nil,
	)
	if err != nil {
		t.Fatalf("mustNap: %v", err)
	}
	return s
}

// mustNight builds a completed night-sleep session between start and stop.
func mustNight(t *testing.T, id string, start, stop time.Time) SleepSession {
	t.Helper()
	nw := mustNightWindow(t, 21, 0, 7, 0)
	s, err := NewCompletedSleepSession(
		SleepSessionID(id), BabyID("baby-1"), FamilyMemberID("member-1"),
		start, stop, SleepClassificationNight, &nw,
	)
	if err != nil {
		t.Fatalf("mustNight: %v", err)
	}
	return s
}

// mustActiveSessionWithID builds an active (not yet stopped) session with the given id.
func mustActiveSessionWithID(t *testing.T, id string, start time.Time) SleepSession {
	t.Helper()
	s, err := NewSleepSession(SleepSessionID(id), BabyID("baby-1"), FamilyMemberID("member-1"), start)
	if err != nil {
		t.Fatalf("mustActiveSessionWithID: %v", err)
	}
	return s
}

// stubDailySummaryRepo is a stub for SleepSessionHistoryRepository used by handler-level tests.
type stubDailySummaryRepo struct {
	sessions []SleepSession
	err      error
}

func (r *stubDailySummaryRepo) FindByBabyIDAndDateRange(_ context.Context, _ BabyID, _ DateRange) ([]SleepSession, error) {
	return r.sessions, r.err
}

// day returns [dayStart, dayEnd) in UTC for the given year/month/day.
func day(year int, month time.Month, d int) (dayStart, dayEnd time.Time) {
	dayStart = time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
	dayEnd = dayStart.AddDate(0, 0, 1)
	return
}

// --- ComputeDailySummary unit tests ---

func TestComputeDailySummaryNoSessionsAllActive(t *testing.T) {
	t.Parallel()

	dayStart, dayEnd := day(2026, time.April, 10)
	now := dayStart.Add(18 * time.Hour)

	got := ComputeDailySummary(nil, dayStart, dayEnd, now)

	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	if got.TotalActive != 18*time.Hour {
		t.Errorf("TotalActive: want 18h, got %v", got.TotalActive)
	}
}

func TestComputeDailySummaryNapWithinDay(t *testing.T) {
	t.Parallel()

	dayStart, dayEnd := day(2026, time.April, 10)
	nap := mustNap(t, "nap-1", dayStart.Add(10*time.Hour), dayStart.Add(11*time.Hour+30*time.Minute))
	now := dayStart.Add(20 * time.Hour)

	got := ComputeDailySummary([]SleepSession{nap}, dayStart, dayEnd, now)

	if got.TotalSleep != 90*time.Minute {
		t.Errorf("TotalSleep: want 90m, got %v", got.TotalSleep)
	}
	if got.TotalActive != 20*time.Hour-90*time.Minute {
		t.Errorf("TotalActive: want %v, got %v", 20*time.Hour-90*time.Minute, got.TotalActive)
	}
}

func TestComputeDailySummaryNightSleepStartingTodayFullDurationCredited(t *testing.T) {
	t.Parallel()

	// Night sleep 21:00 today → 06:00 next day (9h). Now is just after midnight.
	dayStart, dayEnd := day(2026, time.April, 10)
	night := mustNight(t, "night-1", dayStart.Add(21*time.Hour), dayStart.Add(30*time.Hour))
	now := dayEnd.Add(time.Second)

	got := ComputeDailySummary([]SleepSession{night}, dayStart, dayEnd, now)

	// Full 9h attributed; only the 3h before midnight clips into the active window.
	if got.TotalSleep != 9*time.Hour {
		t.Errorf("TotalSleep: want 9h, got %v", got.TotalSleep)
	}
	if got.TotalActive != 21*time.Hour {
		t.Errorf("TotalActive: want 21h, got %v", got.TotalActive)
	}
}

func TestComputeDailySummaryNightSleepStartingYesterdayExcluded(t *testing.T) {
	t.Parallel()

	dayStart, dayEnd := day(2026, time.April, 10)
	// Night started yesterday, ends 06:00 today — must not appear in today's total.
	night := mustNight(t, "night-1", dayStart.Add(-9*time.Hour), dayStart.Add(6*time.Hour))
	now := dayStart.Add(18 * time.Hour)

	got := ComputeDailySummary([]SleepSession{night}, dayStart, dayEnd, now)

	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	if got.TotalActive != 18*time.Hour {
		t.Errorf("TotalActive: want 18h, got %v", got.TotalActive)
	}
}

func TestComputeDailySummaryCrossMidnightNapOverlapsDay(t *testing.T) {
	t.Parallel()

	// Nap 23:30 yesterday → 00:30 today: full 1h attributed; 30m clips into active window.
	dayStart, dayEnd := day(2026, time.April, 10)
	nap := mustNap(t, "nap-1", dayStart.Add(-30*time.Minute), dayStart.Add(30*time.Minute))
	now := dayStart.Add(18 * time.Hour)

	got := ComputeDailySummary([]SleepSession{nap}, dayStart, dayEnd, now)

	if got.TotalSleep != time.Hour {
		t.Errorf("TotalSleep: want 1h, got %v", got.TotalSleep)
	}
	want := 18*time.Hour - 30*time.Minute
	if got.TotalActive != want {
		t.Errorf("TotalActive: want %v, got %v", want, got.TotalActive)
	}
}

func TestComputeDailySummaryActiveSessionClippedToNow(t *testing.T) {
	t.Parallel()

	// Active session started 09:00; now is 10:30.
	dayStart, dayEnd := day(2026, time.April, 10)
	active := mustActiveSessionWithID(t, "active-1", dayStart.Add(9*time.Hour))
	now := dayStart.Add(10*time.Hour + 30*time.Minute)

	got := ComputeDailySummary([]SleepSession{active}, dayStart, dayEnd, now)

	if got.TotalSleep != 90*time.Minute {
		t.Errorf("TotalSleep: want 90m, got %v", got.TotalSleep)
	}
	if got.TotalActive != 9*time.Hour {
		t.Errorf("TotalActive: want 9h, got %v", got.TotalActive)
	}
}

func TestComputeDailySummaryMultipleSessions(t *testing.T) {
	t.Parallel()

	// nap1 09:00–10:00 (1h), nap2 13:00–14:30 (1.5h), night 21:00–06:00 next day (9h).
	dayStart, dayEnd := day(2026, time.April, 10)
	nap1 := mustNap(t, "nap-1", dayStart.Add(9*time.Hour), dayStart.Add(10*time.Hour))
	nap2 := mustNap(t, "nap-2", dayStart.Add(13*time.Hour), dayStart.Add(14*time.Hour+30*time.Minute))
	night := mustNight(t, "night-1", dayStart.Add(21*time.Hour), dayStart.Add(30*time.Hour))
	now := dayStart.Add(25 * time.Hour) // 01:00 next day

	got := ComputeDailySummary([]SleepSession{nap1, nap2, night}, dayStart, dayEnd, now)

	// Sleep: 1h + 1.5h + 9h = 11.5h
	wantSleep := time.Hour + 90*time.Minute + 9*time.Hour
	if got.TotalSleep != wantSleep {
		t.Errorf("TotalSleep: want %v, got %v", wantSleep, got.TotalSleep)
	}
	// Active: 24h window − (1h + 1.5h + 3h night before midnight) = 18.5h
	wantActive := 24*time.Hour - (time.Hour + 90*time.Minute + 3*time.Hour)
	if got.TotalActive != wantActive {
		t.Errorf("TotalActive: want %v, got %v", wantActive, got.TotalActive)
	}
}

func TestComputeDailySummaryFutureDayReturnsZero(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
	dayStart, dayEnd := day(2026, time.April, 10)

	got := ComputeDailySummary(nil, dayStart, dayEnd, now)

	if got.TotalSleep != 0 || got.TotalActive != 0 {
		t.Errorf("expected zero totals for future day, got sleep=%v active=%v", got.TotalSleep, got.TotalActive)
	}
}

// --- DST transition tests ---

// TestComputeDailySummaryDSTSpringForwardNoSessions verifies a 23-hour active window.
func TestComputeDailySummaryDSTSpringForwardNoSessions(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-03-08: clocks spring forward → 23-hour day.
	dayStart := time.Date(2026, time.March, 8, 0, 0, 0, 0, loc).UTC()
	dayEnd := time.Date(2026, time.March, 9, 0, 0, 0, 0, loc).UTC()
	now := dayEnd.Add(time.Second)

	got := ComputeDailySummary(nil, dayStart, dayEnd, now)

	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	if got.TotalActive != 23*time.Hour {
		t.Errorf("TotalActive: want 23h, got %v", got.TotalActive)
	}
}

// TestComputeDailySummaryDSTFallBackNoSessions verifies a 25-hour active window.
func TestComputeDailySummaryDSTFallBackNoSessions(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-11-01: clocks fall back → 25-hour day.
	dayStart := time.Date(2026, time.November, 1, 0, 0, 0, 0, loc).UTC()
	dayEnd := time.Date(2026, time.November, 2, 0, 0, 0, 0, loc).UTC()
	now := dayEnd.Add(time.Second)

	got := ComputeDailySummary(nil, dayStart, dayEnd, now)

	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	if got.TotalActive != 25*time.Hour {
		t.Errorf("TotalActive: want 25h, got %v", got.TotalActive)
	}
}

// TestComputeDailySummaryDSTSpringForwardWithNap verifies totals within a 23-hour day.
func TestComputeDailySummaryDSTSpringForwardWithNap(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	napStart := time.Date(2026, time.March, 8, 10, 0, 0, 0, loc)
	napEnd := time.Date(2026, time.March, 8, 11, 30, 0, 0, loc)
	nap := mustNap(t, "nap-1", napStart.UTC(), napEnd.UTC())

	dayStart := time.Date(2026, time.March, 8, 0, 0, 0, 0, loc).UTC()
	dayEnd := time.Date(2026, time.March, 9, 0, 0, 0, 0, loc).UTC()
	now := dayEnd.Add(time.Second)

	got := ComputeDailySummary([]SleepSession{nap}, dayStart, dayEnd, now)

	if got.TotalSleep != 90*time.Minute {
		t.Errorf("TotalSleep: want 90m, got %v", got.TotalSleep)
	}
	if got.TotalActive != 23*time.Hour-90*time.Minute {
		t.Errorf("TotalActive: want %v, got %v", 23*time.Hour-90*time.Minute, got.TotalActive)
	}
}

// --- GetDailySummaryHandler integration test (verifies the repo query path) ---

func TestGetDailySummaryHandlerQueriesCorrectRange(t *testing.T) {
	t.Parallel()

	dayStart := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	nap := mustNap(t, "nap-1", dayStart.Add(10*time.Hour), dayStart.Add(11*time.Hour))
	now := dayStart.Add(20 * time.Hour)

	repo := &stubDailySummaryRepo{sessions: []SleepSession{nap}}
	h := NewGetDailySummaryHandler(repo)
	h.now = func() time.Time { return now }

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != time.Hour {
		t.Errorf("TotalSleep: want 1h, got %v", got.TotalSleep)
	}
}
