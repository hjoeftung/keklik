package sleep_test

import (
	"context"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/sleep"
)

type stubLogPastSleepRepo struct {
	hasOverlap  bool
	overlapErr  error
	saveErr     error
	savedSession sleep.SleepSession
}

func (r *stubLogPastSleepRepo) Save(_ context.Context, s sleep.SleepSession) error {
	r.savedSession = s
	return r.saveErr
}

func (r *stubLogPastSleepRepo) SaveAll(_ context.Context, _ []sleep.SleepSession) error {
	return nil
}

func (r *stubLogPastSleepRepo) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, nil
}

func (r *stubLogPastSleepRepo) HasOverlappingByBabyID(_ context.Context, _ sleep.BabyID, _, _ time.Time) (bool, error) {
	return r.hasOverlap, r.overlapErr
}

func TestLogPastSleepReturnsCompletedSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(8 * time.Hour)
	repo := &stubLogPastSleepRepo{}
	h := sleep.NewLogPastSleepHandler(repo)

	result, err := h.Handle(context.Background(), sleep.LogPastSleepCommand{
		BabyID:            "baby-1",
		CreatedByMemberID: "member-1",
		StartedAt:         startedAt,
		StoppedAt:         stoppedAt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	if !result.StartedAt.Equal(startedAt) {
		t.Fatalf("expected StartedAt %v, got %v", startedAt, result.StartedAt)
	}
	if !result.StoppedAt.Equal(stoppedAt) {
		t.Fatalf("expected StoppedAt %v, got %v", stoppedAt, result.StoppedAt)
	}
}

func TestLogPastSleepRejectsOverlappingSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(8 * time.Hour)
	repo := &stubLogPastSleepRepo{hasOverlap: true}
	h := sleep.NewLogPastSleepHandler(repo)

	_, err := h.Handle(context.Background(), sleep.LogPastSleepCommand{
		BabyID:            "baby-1",
		CreatedByMemberID: "member-1",
		StartedAt:         startedAt,
		StoppedAt:         stoppedAt,
	})
	if err == nil {
		t.Fatal("expected overlap error, got nil")
	}
	if err != sleep.ErrSleepSessionOverlap {
		t.Fatalf("expected ErrSleepSessionOverlap, got %v", err)
	}
}

func TestLogPastSleepRejectsStopBeforeStart(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(-1 * time.Hour)
	repo := &stubLogPastSleepRepo{}
	h := sleep.NewLogPastSleepHandler(repo)

	_, err := h.Handle(context.Background(), sleep.LogPastSleepCommand{
		BabyID:            "baby-1",
		CreatedByMemberID: "member-1",
		StartedAt:         startedAt,
		StoppedAt:         stoppedAt,
	})
	if err == nil {
		t.Fatal("expected error for stop before start, got nil")
	}
}
