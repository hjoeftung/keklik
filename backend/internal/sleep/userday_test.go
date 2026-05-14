package sleep

import (
	"testing"
	"time"
)

// Fixtures: window 21:00–07:00, loc=UTC, now=Apr 29 12:00 UTC unless stated.

func TestBuildUserDaysActiveNightSleep(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 6, 0, 0, 0, time.UTC)

	// Night started Apr 28 22:00, still active.
	night, _ := NewSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 22, 0, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ night, nap, active float64 }{28800, 0, 0}
	got := struct{ night, nap, active float64 }{
		days[0].NightDuration(now).Seconds(),
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysNoSessions(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	days, err := buildUserDays(nil, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ night, nap, active float64 }{0, 0, 43200}
	got := struct{ night, nap, active float64 }{
		days[0].NightDuration(now).Seconds(),
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysYesterdayNightAnchorsWakeButIsNotTodaysNightSleep(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	night, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ night, nap, active float64 }{0, 0, 18000}
	got := struct{ night, nap, active float64 }{
		days[0].NightDuration(now).Seconds(),
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysNapBetweenWokeAtAndNow(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	night, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)
	nap, _ := NewCompletedSleepSession("s2", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 10, 0, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night, nap}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ nap, active float64 }{3600, 14400}
	got := struct{ nap, active float64 }{
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysTonightStartedCapsActive(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 22, 0, 0, 0, time.UTC)

	night, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)
	nap, _ := NewCompletedSleepSession("s2", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 9, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 10, 0, 0, 0, time.UTC),
	)
	tonight, _ := NewSleepSession("s3", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 21, 0, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night, nap, tonight}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ nap, active float64 }{3600, 46800}
	got := struct{ nap, active float64 }{
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysNapBeforeWokeAtExcluded(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	night, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)
	// Nap starting before WokeAt (06:30 < 07:00).
	earlyNap, _ := NewCompletedSleepSession("s2", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 6, 30, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 6, 50, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night, earlyNap}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ nap, active float64 }{0, 18000}
	got := struct{ nap, active float64 }{
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysNapAtExactWokeAtIncluded(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	night, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)
	// Nap starting exactly at WokeAt.
	nap, _ := NewCompletedSleepSession("s2", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 8, 0, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night, nap}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ nap, active float64 }{3600, 14400}
	got := struct{ nap, active float64 }{
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysNapAtNightStartedAtExcluded(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 22, 0, 0, 0, time.UTC)

	night, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)
	tonight, _ := NewSleepSession("s2", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 21, 0, 0, 0, time.UTC),
	)
	// Nap-classified session starting exactly at NightStartedAt — excluded.
	napAtNight, _ := NewCompletedSleepSession("s3", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 21, 30, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night, tonight, napAtNight}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ nap, active float64 }{0, 50400}
	got := struct{ nap, active float64 }{
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysMultipleNightSessionsWokeAtIsLatest(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	// Two night sessions: 21:00–01:00 (4h) and 02:00–07:00 (5h).
	night1, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 1, 0, 0, 0, time.UTC),
	)
	night2, _ := NewCompletedSleepSession("s2", "baby-1", "m-1",
		time.Date(2026, time.April, 29, 2, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night1, night2}, []NightWindow{nw}, time.UTC, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := struct{ night, nap, active float64 }{32400, 0, 18000}
	got := struct{ night, nap, active float64 }{
		days[0].NightDuration(now).Seconds(),
		days[0].NapDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBuildUserDaysMultiDayIndependence(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	now := time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC)

	night1, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 27, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 28, 7, 0, 0, 0, time.UTC),
	)
	night2, _ := NewCompletedSleepSession("s2", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 21, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 29, 7, 0, 0, 0, time.UTC),
	)

	days, err := buildUserDays([]SleepSession{night1, night2}, []NightWindow{nw}, time.UTC, now, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// days[0] = Apr 29: previous window Apr 28 21:00–Apr 29 07:00 → night2.
	wantDay0 := struct{ night, active float64 }{36000, 18000}
	gotDay0 := struct{ night, active float64 }{
		days[0].NightDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if gotDay0 != wantDay0 {
		t.Fatalf("days[0]: want %+v, got %+v", wantDay0, gotDay0)
	}

	// days[1] = Apr 28: previous window Apr 27 21:00–Apr 28 07:00 → night1;
	//   tonight window Apr 28 21:00–Apr 29 07:00 → NightStartedAt = Apr 28 21:00.
	wantDay1 := struct{ night, active float64 }{36000, 50400}
	gotDay1 := struct{ night, active float64 }{
		days[1].NightDuration(now).Seconds(),
		days[1].ActiveDuration(now).Seconds(),
	}
	if gotDay1 != wantDay1 {
		t.Fatalf("days[1]: want %+v, got %+v", wantDay1, gotDay1)
	}
}

func TestBuildUserDaysNonUTCTimezoneYesterdayNightSetsWakeAnchor(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// now = Apr 29 03:00 UTC = Apr 28 23:00 EDT → "today" is Apr 28 in NY.
	now := time.Date(2026, time.April, 29, 3, 0, 0, 0, time.UTC)

	// Night Apr 27 21:00 EDT – Apr 28 07:00 EDT.
	night, _ := NewCompletedSleepSession("s1", "baby-1", "m-1",
		time.Date(2026, time.April, 28, 1, 0, 0, 0, time.UTC),  // Apr 27 21:00 EDT
		time.Date(2026, time.April, 28, 11, 0, 0, 0, time.UTC), // Apr 28 07:00 EDT
	)

	days, err := buildUserDays([]SleepSession{night}, []NightWindow{nw}, loc, now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WokeAt = Apr 28 07:00 EDT; active = now(23:00 EDT) − 07:00 EDT = 16h.
	want := struct{ night, active float64 }{0, 57600}
	got := struct{ night, active float64 }{
		days[0].NightDuration(now).Seconds(),
		days[0].ActiveDuration(now).Seconds(),
	}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}
