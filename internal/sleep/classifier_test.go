package sleep

import (
	"testing"
	"time"
)

// mustCompletedSession is a test helper that builds a completed SleepSession.
func mustCompletedSession(t *testing.T, start, stop time.Time) SleepSession {
	t.Helper()

	session, err := NewSleepSession(SleepSessionID("s1"), BabyID("b1"), FamilyMemberID("member-1"), start)
	if err != nil {
		t.Fatalf("NewSleepSession: %v", err)
	}

	if err := session.Stop(stop, SleepClassificationNap, 0); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	return session
}

// --- basic cases ---

func TestClassifyNightSleepMoreThanHalfInWindow(t *testing.T) {
	t.Parallel()

	// Window 21:00–07:00 (crosses midnight). Session 22:00–06:00 is fully inside.
	nw := mustNightWindow(t, 21, 0, 7, 0)
	start := time.Date(2026, time.January, 10, 22, 0, 0, 0, time.UTC)
	stop := time.Date(2026, time.January, 11, 6, 0, 0, 0, time.UTC)

	got, err := Classify(mustCompletedSession(t, start, stop), "UTC", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationNight {
		t.Fatalf("expected night, got %q", got)
	}
}

func TestClassifyNapLessThanHalfInWindow(t *testing.T) {
	t.Parallel()

	// Window 21:00–07:00. Session 14:00–15:00 is entirely outside.
	nw := mustNightWindow(t, 21, 0, 7, 0)
	start := time.Date(2026, time.January, 10, 14, 0, 0, 0, time.UTC)
	stop := time.Date(2026, time.January, 10, 15, 0, 0, 0, time.UTC)

	got, err := Classify(mustCompletedSession(t, start, stop), "UTC", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationNap {
		t.Fatalf("expected nap, got %q", got)
	}
}

func TestClassifyExactlyHalfInWindowIsNap(t *testing.T) {
	t.Parallel()

	// Session 21:00–23:00 with window 22:00–06:00: overlap = 60 min out of
	// 120 min total → not MORE than half → nap.
	nw := mustNightWindow(t, 22, 0, 6, 0)
	start := time.Date(2026, time.January, 10, 21, 0, 0, 0, time.UTC)
	stop := time.Date(2026, time.January, 10, 23, 0, 0, 0, time.UTC)

	got, err := Classify(mustCompletedSession(t, start, stop), "UTC", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationNap {
		t.Fatalf("expected nap for exactly-half overlap, got %q", got)
	}
}

func TestClassifyActiveSessionReturnsUnknown(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	start := time.Date(2026, time.January, 10, 22, 0, 0, 0, time.UTC)
	session, err := NewSleepSession(SleepSessionID("s1"), BabyID("b1"), FamilyMemberID("member-1"), start)
	if err != nil {
		t.Fatalf("NewSleepSession: %v", err)
	}

	got, err := Classify(session, "UTC", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationUnknown {
		t.Fatalf("expected unknown for active session, got %q", got)
	}
}

func TestClassifyInvalidTimezoneReturnsError(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	start := time.Date(2026, time.January, 10, 22, 0, 0, 0, time.UTC)
	stop := start.Add(8 * time.Hour)

	_, err := Classify(mustCompletedSession(t, start, stop), "Not/ATimezone", nw)
	if err != ErrInvalidTimezone {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}

// --- midnight-crossing window ---

func TestClassifyMidnightCrossingWindowNightSleep(t *testing.T) {
	t.Parallel()

	// Window 20:00–04:00. Session 19:00–03:00: overlap = 20:00–03:00 = 7h
	// out of 8h total → night.
	nw := mustNightWindow(t, 20, 0, 4, 0)
	start := time.Date(2026, time.March, 15, 19, 0, 0, 0, time.UTC)
	stop := time.Date(2026, time.March, 16, 3, 0, 0, 0, time.UTC)

	got, err := Classify(mustCompletedSession(t, start, stop), "UTC", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationNight {
		t.Fatalf("expected night, got %q", got)
	}
}

func TestClassifyMidnightCrossingWindowNap(t *testing.T) {
	t.Parallel()

	// Window 22:00–05:00. Session 05:30–07:30: entirely outside → nap.
	nw := mustNightWindow(t, 22, 0, 5, 0)
	start := time.Date(2026, time.March, 15, 5, 30, 0, 0, time.UTC)
	stop := time.Date(2026, time.March, 15, 7, 30, 0, 0, time.UTC)

	got, err := Classify(mustCompletedSession(t, start, stop), "UTC", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationNap {
		t.Fatalf("expected nap, got %q", got)
	}
}

// --- DST transitions ---

// TestClassifyDSTSpringForwardNightSleep covers the North-American spring
// forward (second Sunday in March) where clocks jump from 02:00 to 03:00.
// A session from 22:00 to 06:00 local on the DST night should still be night
// sleep even though the night is one hour shorter than usual.
func TestClassifyDSTSpringForwardNightSleep(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-03-08: spring-forward night in New York (clocks move forward at 02:00).
	nw := mustNightWindow(t, 21, 0, 7, 0)
	start := time.Date(2026, time.March, 8, 22, 0, 0, 0, loc) // 22:00 EST
	stop := time.Date(2026, time.March, 9, 6, 0, 0, 0, loc)   // 06:00 EDT

	got, err := Classify(mustCompletedSession(t, start, stop), "America/New_York", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationNight {
		t.Fatalf("expected night on DST spring-forward night, got %q", got)
	}
}

// TestClassifyDSTFallBackNightSleep covers the North-American fall back
// (first Sunday in November) where clocks repeat 01:00–02:00.
// A session from 22:00 to 06:00 local on the fall-back night should still
// be night sleep even though the night is one hour longer than usual.
func TestClassifyDSTFallBackNightSleep(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-11-01: fall-back night in New York (clocks move back at 02:00).
	nw := mustNightWindow(t, 21, 0, 7, 0)
	start := time.Date(2026, time.November, 1, 22, 0, 0, 0, loc) // 22:00 EDT
	stop := time.Date(2026, time.November, 2, 6, 0, 0, 0, loc)   // 06:00 EST

	got, err := Classify(mustCompletedSession(t, start, stop), "America/New_York", nw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != SleepClassificationNight {
		t.Fatalf("expected night on DST fall-back night, got %q", got)
	}
}

// --- forward-only night-window changes ---

// TestClassifyForwardOnlyWindowChange documents that Classify operates purely
// on the supplied window — it has no memory of prior classifications.
// The version stored with a session is the mechanism a use-case layer uses to
// detect whether a past session needs reclassification; sessions whose version
// predates the current window version are left untouched.
func TestClassifyForwardOnlyWindowChange(t *testing.T) {
	t.Parallel()

	// Session 22:30–05:30 (7 hours).
	start := time.Date(2026, time.January, 10, 22, 30, 0, 0, time.UTC)
	stop := time.Date(2026, time.January, 11, 5, 30, 0, 0, time.UTC)
	session := mustCompletedSession(t, start, stop)

	// Old window 22:00–06:00: full session inside → night.
	oldWindow := mustNightWindow(t, 22, 0, 6, 0)
	gotOld, err := Classify(session, "UTC", oldWindow)
	if err != nil {
		t.Fatalf("unexpected error with old window: %v", err)
	}

	if gotOld != SleepClassificationNight {
		t.Fatalf("expected night with old window, got %q", gotOld)
	}

	// Narrowed window 01:00–04:00: overlap = 3h out of 7h → <half → nap.
	narrowWindow := mustNightWindow(t, 1, 0, 4, 0)
	gotNarrow, err := Classify(session, "UTC", narrowWindow)
	if err != nil {
		t.Fatalf("unexpected error with narrow window: %v", err)
	}

	if gotNarrow != SleepClassificationNap {
		t.Fatalf("expected nap with narrow window, got %q", gotNarrow)
	}

	// The stored ClassificationRuleVersion on the session (0) signals that it
	// was classified under the old rule. A use-case layer increments the version
	// on window changes and skips reclassification for sessions with older versions.
	if session.ClassificationRuleVersion() != 0 {
		t.Fatalf("expected rule version 0 on original session, got %d", session.ClassificationRuleVersion())
	}
}
