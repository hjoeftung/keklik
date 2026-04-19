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

// stubDailySummaryRepo is a stub for SleepSessionHistoryRepository.
type stubDailySummaryRepo struct {
	sessions []SleepSession
	err      error
}

func (r *stubDailySummaryRepo) FindByBabyIDAndDateRange(_ context.Context, _ BabyID, _ DateRange) ([]SleepSession, error) {
	return r.sessions, r.err
}

// fixedNow returns a func() time.Time that always returns t.
func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

// newHandler wires up a GetDailySummaryHandler with a fixed clock and provided sessions.
func newHandler(sessions []SleepSession, now time.Time) *GetDailySummaryHandler {
	h := NewGetDailySummaryHandler(&stubDailySummaryRepo{sessions: sessions})
	h.now = fixedNow(now)
	return h
}

// --- basic cases ---

func TestGetDailySummaryNoSessionsReturnsZeroTotals(t *testing.T) {
	t.Parallel()

	// Day: 2026-04-10 UTC. now is 18:00 (18h elapsed).
	now := time.Date(2026, time.April, 10, 18, 0, 0, 0, time.UTC)
	h := newHandler(nil, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID:   "baby-1",
		Timezone: "UTC",
		Date:     now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	if got.TotalActive != 18*time.Hour {
		t.Errorf("TotalActive: want 18h, got %v", got.TotalActive)
	}
}

func TestGetDailySummaryNapWithinDayIsIncluded(t *testing.T) {
	t.Parallel()

	// Day: 2026-04-10. Nap 10:00–11:30 (1.5h).
	dayStart := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	nap := mustNap(t, "nap-1",
		dayStart.Add(10*time.Hour),
		dayStart.Add(11*time.Hour+30*time.Minute),
	)
	now := dayStart.Add(20 * time.Hour)
	h := newHandler([]SleepSession{nap}, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != 90*time.Minute {
		t.Errorf("TotalSleep: want 90m, got %v", got.TotalSleep)
	}
	if got.TotalActive != 20*time.Hour-90*time.Minute {
		t.Errorf("TotalActive: want %v, got %v", 20*time.Hour-90*time.Minute, got.TotalActive)
	}
}

func TestGetDailySummaryNightSleepStartingTodayIsIncluded(t *testing.T) {
	t.Parallel()

	// Day: 2026-04-10. Night sleep starts 21:00 today, ends 06:00 next day (9h).
	dayStart := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	night := mustNight(t, "night-1",
		dayStart.Add(21*time.Hour),
		dayStart.Add(30*time.Hour), // 06:00 next day
	)
	// now is midnight+1s (day complete), but night sleep extends past it.
	dayEnd := dayStart.Add(24 * time.Hour)
	now := dayEnd.Add(time.Second) // just after midnight
	h := newHandler([]SleepSession{night}, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: dayStart.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Full 9h night sleep attributed to this day.
	if got.TotalSleep != 9*time.Hour {
		t.Errorf("TotalSleep: want 9h, got %v", got.TotalSleep)
	}
	// Active time uses only the portion of the night sleep within [dayStart, dayEnd):
	// 21:00–00:00 = 3h sleeping; 24h day − 3h = 21h active.
	if got.TotalActive != 21*time.Hour {
		t.Errorf("TotalActive: want 21h, got %v", got.TotalActive)
	}
}

func TestGetDailySummaryNightSleepStartingYesterdayIsExcluded(t *testing.T) {
	t.Parallel()

	// Night sleep starts yesterday evening and ends this morning.
	dayStart := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	night := mustNight(t, "night-1",
		dayStart.Add(-9*time.Hour), // 15:00 yesterday
		dayStart.Add(6*time.Hour),  // 06:00 today
	)
	now := dayStart.Add(18 * time.Hour)
	h := newHandler([]SleepSession{night}, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Night sleep started yesterday → not included in today's sleep total.
	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	// Active time: the night sleep clips to [dayStart, dayEnd ∩ now] = 6h sleeping; 18h − 6h = 12h.
	// But the night sleep is excluded by the inclusion rule, so 0 sleep in window → 18h active.
	if got.TotalActive != 18*time.Hour {
		t.Errorf("TotalActive: want 18h, got %v", got.TotalActive)
	}
}

func TestGetDailySummaryCrossMidnightNapOverlapsDay(t *testing.T) {
	t.Parallel()

	// Nap 23:30 (yesterday) → 00:30 (today). Overlaps today by 30 minutes.
	dayStart := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	nap := mustNap(t, "nap-1",
		dayStart.Add(-30*time.Minute), // 23:30 yesterday
		dayStart.Add(30*time.Minute),  // 00:30 today
	)
	now := dayStart.Add(18 * time.Hour)
	h := newHandler([]SleepSession{nap}, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Full nap duration (1h) is attributed to today.
	if got.TotalSleep != time.Hour {
		t.Errorf("TotalSleep: want 1h, got %v", got.TotalSleep)
	}
	// Active within window [00:00, 18:00]: nap clips to [00:00, 00:30] = 30m sleeping; 18h − 30m = 17.5h.
	want := 18*time.Hour - 30*time.Minute
	if got.TotalActive != want {
		t.Errorf("TotalActive: want %v, got %v", want, got.TotalActive)
	}
}

func TestGetDailySummaryActiveSessionCountedUpToNow(t *testing.T) {
	t.Parallel()

	// Day: 2026-04-10. Active session started at 09:00; now is 10:30.
	dayStart := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	active := mustActiveSessionWithID(t, "active-1", dayStart.Add(9*time.Hour))
	now := dayStart.Add(10*time.Hour + 30*time.Minute)
	h := newHandler([]SleepSession{active}, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != 90*time.Minute {
		t.Errorf("TotalSleep: want 90m, got %v", got.TotalSleep)
	}
	// Active: window is [00:00, 10:30] = 10.5h; sleep in window = 1.5h → active = 9h.
	if got.TotalActive != 9*time.Hour {
		t.Errorf("TotalActive: want 9h, got %v", got.TotalActive)
	}
}

func TestGetDailySummaryMultipleSessionsCombined(t *testing.T) {
	t.Parallel()

	// Day: 2026-04-10.
	//   nap 1:  09:00–10:00 (1h)
	//   nap 2:  13:00–14:30 (1.5h)
	//   night:  21:00–06:00 next day (9h)
	dayStart := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	nap1 := mustNap(t, "nap-1", dayStart.Add(9*time.Hour), dayStart.Add(10*time.Hour))
	nap2 := mustNap(t, "nap-2", dayStart.Add(13*time.Hour), dayStart.Add(14*time.Hour+30*time.Minute))
	night := mustNight(t, "night-1", dayStart.Add(21*time.Hour), dayStart.Add(30*time.Hour))

	now := dayStart.Add(25 * time.Hour) // 01:00 next day, night still active in DB but completed
	h := newHandler([]SleepSession{nap1, nap2, night}, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: dayStart.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Sleep total: 1h + 1.5h + 9h = 11.5h
	wantSleep := time.Hour + 90*time.Minute + 9*time.Hour
	if got.TotalSleep != wantSleep {
		t.Errorf("TotalSleep: want %v, got %v", wantSleep, got.TotalSleep)
	}
	// Active: window [00:00, 00:00 next day] = 24h;
	// sleep clipped to window: nap1(1h) + nap2(1.5h) + night(21:00–00:00 = 3h) = 5.5h
	// active = 24h − 5.5h = 18.5h
	wantActive := 24*time.Hour - (time.Hour + 90*time.Minute + 3*time.Hour)
	if got.TotalActive != wantActive {
		t.Errorf("TotalActive: want %v, got %v", wantActive, got.TotalActive)
	}
}

// --- DST transition tests ---

// TestGetDailySummaryDSTSpringForwardDay verifies that on the 23-hour spring-forward
// day, a session-free day yields TotalActive = 23h (not 24h).
func TestGetDailySummaryDSTSpringForwardDay(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-03-08: clocks spring forward at 02:00 → 03:00 (23-hour day).
	dayStart := time.Date(2026, time.March, 8, 0, 0, 0, 0, loc).UTC()
	dayEnd := time.Date(2026, time.March, 9, 0, 0, 0, 0, loc).UTC()
	wantWindow := dayEnd.Sub(dayStart) // 23h

	// now = just after local midnight (day complete, no sessions).
	now := dayEnd.Add(time.Second)
	h := newHandler(nil, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID:   "baby-1",
		Timezone: "America/New_York",
		Date:     dayStart.Add(time.Hour), // noon-ish local to pick the right day
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	if got.TotalActive != wantWindow {
		t.Errorf("TotalActive: want %v (23h), got %v", wantWindow, got.TotalActive)
	}
}

// TestGetDailySummaryDSTFallBackDay verifies that on the 25-hour fall-back day,
// a session-free day yields TotalActive = 25h (not 24h).
func TestGetDailySummaryDSTFallBackDay(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-11-01: clocks fall back at 02:00 → 01:00 (25-hour day).
	dayStart := time.Date(2026, time.November, 1, 0, 0, 0, 0, loc).UTC()
	dayEnd := time.Date(2026, time.November, 2, 0, 0, 0, 0, loc).UTC()
	wantWindow := dayEnd.Sub(dayStart) // 25h

	now := dayEnd.Add(time.Second)
	h := newHandler(nil, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID:   "baby-1",
		Timezone: "America/New_York",
		Date:     dayStart.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != 0 {
		t.Errorf("TotalSleep: want 0, got %v", got.TotalSleep)
	}
	if got.TotalActive != wantWindow {
		t.Errorf("TotalActive: want %v (25h), got %v", wantWindow, got.TotalActive)
	}
}

// TestGetDailySummaryDSTSpringForwardWithNap verifies sleep/active totals are
// computed correctly within a 23-hour day.
func TestGetDailySummaryDSTSpringForwardWithNap(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// Day: 2026-03-08 (23h). Nap 10:00–11:30 local (1.5h).
	napStart := time.Date(2026, time.March, 8, 10, 0, 0, 0, loc)
	napEnd := time.Date(2026, time.March, 8, 11, 30, 0, 0, loc)
	nap := mustNap(t, "nap-1", napStart.UTC(), napEnd.UTC())

	dayEnd := time.Date(2026, time.March, 9, 0, 0, 0, 0, loc).UTC()
	now := dayEnd.Add(time.Second)
	h := newHandler([]SleepSession{nap}, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID:   "baby-1",
		Timezone: "America/New_York",
		Date:     napStart.UTC(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != 90*time.Minute {
		t.Errorf("TotalSleep: want 90m, got %v", got.TotalSleep)
	}
	wantActive := 23*time.Hour - 90*time.Minute
	if got.TotalActive != wantActive {
		t.Errorf("TotalActive: want %v, got %v", wantActive, got.TotalActive)
	}
}

// TestGetDailySummaryFutureDayReturnsZero verifies that querying a day that
// hasn't started yet returns zero totals without error.
func TestGetDailySummaryFutureDayReturnsZero(t *testing.T) {
	t.Parallel()

	// now is 2026-04-09 noon; query for 2026-04-10.
	now := time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
	futureDate := time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC)
	h := newHandler(nil, now)

	got, err := h.Handle(context.Background(), GetDailySummaryQuery{
		BabyID: "baby-1", Timezone: "UTC", Date: futureDate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalSleep != 0 || got.TotalActive != 0 {
		t.Errorf("expected zero totals for future day, got sleep=%v active=%v", got.TotalSleep, got.TotalActive)
	}
}
