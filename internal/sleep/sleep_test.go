package sleep

import (
	"errors"
	"testing"
	"time"
)

func TestNewSleepSessionCreatesActiveSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	session, err := NewSleepSession(SleepSessionID("session-1"), BabyID("baby-1"), startedAt)
	if err != nil {
		t.Fatalf("NewSleepSession returned error: %v", err)
	}

	if session.ID() != SleepSessionID("session-1") {
		t.Fatalf("expected session id %q, got %q", SleepSessionID("session-1"), session.ID())
	}

	if session.BabyID() != BabyID("baby-1") {
		t.Fatalf("expected baby id %q, got %q", BabyID("baby-1"), session.BabyID())
	}

	if !session.IsActive() {
		t.Fatal("expected session to be active")
	}

	if session.Classification() != SleepClassificationUnknown {
		t.Fatalf("expected empty classification for active session, got %q", session.Classification())
	}

	if _, ok := session.Duration(); ok {
		t.Fatal("expected active session duration to be unavailable")
	}

	if _, ok := session.StoppedAt(); ok {
		t.Fatal("expected active session to have no stop timestamp")
	}
}

func TestSleepSessionStopCompletesSessionAndDerivesDuration(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(90 * time.Minute)
	session := mustSleepSession(t, startedAt)

	if err := session.Stop(stoppedAt, SleepClassificationNight); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	if session.IsActive() {
		t.Fatal("expected session to be completed")
	}

	storedStoppedAt, ok := session.StoppedAt()
	if !ok {
		t.Fatal("expected stopped timestamp to be available")
	}

	if !storedStoppedAt.Equal(stoppedAt) {
		t.Fatalf("expected stop time %v, got %v", stoppedAt, storedStoppedAt)
	}

	duration, ok := session.Duration()
	if !ok {
		t.Fatal("expected completed session duration to be available")
	}

	if duration != 90*time.Minute {
		t.Fatalf("expected duration %v, got %v", 90*time.Minute, duration)
	}

	if session.Classification() != SleepClassificationNight {
		t.Fatalf("expected classification %q, got %q", SleepClassificationNight, session.Classification())
	}
}

func TestSleepSessionAllowsEqualStartAndStop(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	session := mustSleepSession(t, startedAt)

	if err := session.Stop(startedAt, SleepClassificationNap); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	duration, ok := session.Duration()
	if !ok {
		t.Fatal("expected completed session duration to be available")
	}

	if duration != 0 {
		t.Fatalf("expected zero duration, got %v", duration)
	}
}

func TestSleepSessionStopRejectsEarlierStop(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	session := mustSleepSession(t, startedAt)

	err := session.Stop(startedAt.Add(-time.Second), SleepClassificationNap)
	if !errors.Is(err, ErrInvalidSleepSessionStop) {
		t.Fatalf("expected ErrInvalidSleepSessionStop, got %v", err)
	}
}

func TestSleepSessionStopRejectsSecondStop(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	session := mustSleepSession(t, startedAt)

	if err := session.Stop(startedAt.Add(30*time.Minute), SleepClassificationNap); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	err := session.Stop(startedAt.Add(45*time.Minute), SleepClassificationNight)
	if !errors.Is(err, ErrSleepSessionAlreadyStopped) {
		t.Fatalf("expected ErrSleepSessionAlreadyStopped, got %v", err)
	}
}

func TestNewCompletedSleepSessionRejectsUnknownClassification(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 12, 19, 0, 0, 0, time.UTC)
	_, err := NewCompletedSleepSession(
		SleepSessionID("session-1"),
		BabyID("baby-1"),
		startedAt,
		startedAt.Add(time.Hour),
		SleepClassification("catnap"),
	)
	if !errors.Is(err, ErrUnknownSleepClassification) {
		t.Fatalf("expected ErrUnknownSleepClassification, got %v", err)
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

	session, err := NewSleepSession(SleepSessionID("session-1"), BabyID("baby-1"), startedAt)
	if err != nil {
		t.Fatalf("NewSleepSession returned error: %v", err)
	}

	return session
}
