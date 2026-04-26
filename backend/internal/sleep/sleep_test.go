package sleep

import (
	"errors"
	"testing"
	"time"
)

func TestNewSleepSessionCreatesActiveSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	session, err := NewSleepSession(SleepSessionID("session-1"), BabyID("baby-1"), FamilyMemberID("member-1"), startedAt)
	if err != nil {
		t.Fatalf("NewSleepSession returned error: %v", err)
	}

	if !session.IsActive() {
		t.Fatal("expected session to be active")
	}
	if _, ok := session.Duration(); ok {
		t.Fatal("expected no duration for active session")
	}
}

func TestSleepSessionStopCompletesSessionAndDerivesDuration(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(90 * time.Minute)
	session := mustSleepSession(t, startedAt)

	if err := session.Stop(stoppedAt); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	duration, ok := session.Duration()
	if !ok || duration != 90*time.Minute {
		t.Fatalf("expected duration %v, got %v (ok=%v)", 90*time.Minute, duration, ok)
	}
}

func TestSleepSessionStopRejectsEarlierStop(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	session := mustSleepSession(t, startedAt)

	err := session.Stop(startedAt.Add(-time.Second))
	if !errors.Is(err, ErrInvalidSleepSessionStop) {
		t.Fatalf("expected ErrInvalidSleepSessionStop, got %v", err)
	}
}

func TestSleepSessionStopRejectsSecondStop(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	session := mustSleepSession(t, startedAt)
	if err := session.Stop(startedAt.Add(30 * time.Minute)); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	err := session.Stop(startedAt.Add(45 * time.Minute))
	if !errors.Is(err, ErrSleepSessionAlreadyStopped) {
		t.Fatalf("expected ErrSleepSessionAlreadyStopped, got %v", err)
	}
}

func TestNewDateRangeRejectsEndBeforeStart(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC)
	_, err := NewDateRange(start, start.Add(-time.Second))
	if !errors.Is(err, ErrInvalidSleepSessionDateRange) {
		t.Fatalf("expected ErrInvalidSleepSessionDateRange, got %v", err)
	}
}

func mustSleepSession(t *testing.T, startedAt time.Time) SleepSession {
	t.Helper()

	session, err := NewSleepSession(SleepSessionID("session-1"), BabyID("baby-1"), FamilyMemberID("member-1"), startedAt)
	if err != nil {
		t.Fatalf("NewSleepSession returned error: %v", err)
	}

	return session
}
