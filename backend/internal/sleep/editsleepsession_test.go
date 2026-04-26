package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubEditableSleepSessionRepo struct {
	session   SleepSession
	findErr   error
	saveErr   error
	deleteErr error
	saved     *SleepSession
}

func (r *stubEditableSleepSessionRepo) Save(_ context.Context, s SleepSession) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = &s
	return nil
}

func (r *stubEditableSleepSessionRepo) SaveAll(_ context.Context, _ []SleepSession) error {
	return nil
}

func (r *stubEditableSleepSessionRepo) FindByID(_ context.Context, _ SleepSessionID) (SleepSession, error) {
	return r.session, r.findErr
}

func (r *stubEditableSleepSessionRepo) DeleteByID(_ context.Context, _ SleepSessionID) error {
	return r.deleteErr
}

func TestEditSleepSessionUpdatesStartedAt(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	editedStart := startedAt.Add(-15 * time.Minute)
	repo := &stubEditableSleepSessionRepo{session: mustActiveSession(t, startedAt)}

	h := NewEditSleepSessionHandler(repo)
	updated, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID: SleepSessionID("session-1"),
		StartedAt: &editedStart,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !updated.StartedAt().Equal(editedStart) {
		t.Fatalf("expected edited start %v, got %v", editedStart, updated.StartedAt())
	}
}

func TestEditSleepSessionCanCompleteSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(2 * time.Hour)
	repo := &stubEditableSleepSessionRepo{session: mustActiveSession(t, startedAt)}

	h := NewEditSleepSessionHandler(repo)
	updated, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID: SleepSessionID("session-1"),
		StoppedAt: &stoppedAt,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if updated.IsActive() {
		t.Fatal("expected completed session")
	}
}

func TestEditSleepSessionRejectsMissingFields(t *testing.T) {
	t.Parallel()

	h := NewEditSleepSessionHandler(&stubEditableSleepSessionRepo{})
	_, err := h.Handle(context.Background(), EditSleepSessionCommand{SessionID: SleepSessionID("session-1")})
	if !errors.Is(err, ErrMissingSleepSessionEdit) {
		t.Fatalf("expected ErrMissingSleepSessionEdit, got %v", err)
	}
}
