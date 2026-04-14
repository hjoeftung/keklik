package sleep

import (
	"context"
	"testing"
	"time"
)

// stubSleepSessionQueryRepository is a test double for SleepSessionQueryRepository.
type stubSleepSessionQueryRepository struct {
	session SleepSession
	found   bool
	err     error
}

func (r *stubSleepSessionQueryRepository) FindMostRecentByBabyID(_ context.Context, _ BabyID) (SleepSession, bool, error) {
	return r.session, r.found, r.err
}

func TestGetElapsedTimeHandlerNoSessionsReturnsBothNil(t *testing.T) {
	t.Parallel()

	repo := &stubSleepSessionQueryRepository{found: false}
	h := NewGetElapsedTimeHandler(repo)

	result, err := h.Handle(context.Background(), BabyID("baby-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TimeSinceLastSleepStart != nil {
		t.Errorf("expected TimeSinceLastSleepStart to be nil, got %v", result.TimeSinceLastSleepStart)
	}
	if result.TimeSinceLastAwakening != nil {
		t.Errorf("expected TimeSinceLastAwakening to be nil, got %v", result.TimeSinceLastAwakening)
	}
}

func TestGetElapsedTimeHandlerActiveSessionReturnsSinceStartOnly(t *testing.T) {
	t.Parallel()

	startedAt := time.Now().UTC().Add(-30 * time.Minute)
	session := mustSleepSession(t, startedAt) // active — no StoppedAt

	repo := &stubSleepSessionQueryRepository{session: session, found: true}
	h := NewGetElapsedTimeHandler(repo)

	result, err := h.Handle(context.Background(), BabyID("baby-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TimeSinceLastSleepStart == nil {
		t.Fatal("expected TimeSinceLastSleepStart to be non-nil")
	}
	// Should be approximately 30 minutes (allow generous tolerance for test runtime).
	const tolerance = 5 * time.Second
	if *result.TimeSinceLastSleepStart < 30*time.Minute-tolerance || *result.TimeSinceLastSleepStart > 30*time.Minute+tolerance {
		t.Errorf("TimeSinceLastSleepStart: want ~30m, got %v", *result.TimeSinceLastSleepStart)
	}

	if result.TimeSinceLastAwakening != nil {
		t.Errorf("expected TimeSinceLastAwakening to be nil for active session, got %v", result.TimeSinceLastAwakening)
	}
}

func TestGetElapsedTimeHandlerCompletedSessionReturnsBothFields(t *testing.T) {
	t.Parallel()

	startedAt := time.Now().UTC().Add(-2 * time.Hour)
	stoppedAt := time.Now().UTC().Add(-45 * time.Minute)
	session := mustCompletedSession(t, startedAt, stoppedAt)

	repo := &stubSleepSessionQueryRepository{session: session, found: true}
	h := NewGetElapsedTimeHandler(repo)

	result, err := h.Handle(context.Background(), BabyID("baby-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TimeSinceLastSleepStart == nil {
		t.Fatal("expected TimeSinceLastSleepStart to be non-nil")
	}
	if result.TimeSinceLastAwakening == nil {
		t.Fatal("expected TimeSinceLastAwakening to be non-nil")
	}

	const tolerance = 5 * time.Second
	if *result.TimeSinceLastSleepStart < 2*time.Hour-tolerance || *result.TimeSinceLastSleepStart > 2*time.Hour+tolerance {
		t.Errorf("TimeSinceLastSleepStart: want ~2h, got %v", *result.TimeSinceLastSleepStart)
	}
	if *result.TimeSinceLastAwakening < 45*time.Minute-tolerance || *result.TimeSinceLastAwakening > 45*time.Minute+tolerance {
		t.Errorf("TimeSinceLastAwakening: want ~45m, got %v", *result.TimeSinceLastAwakening)
	}
}
