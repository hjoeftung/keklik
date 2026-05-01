package sleep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

type stubEditableSleepSessionRepo struct {
	session     SleepSession
	overlapping SleepSession
	findErr     error
	overlapErr  error
	saveErr     error
	deleteErr   error
	saved       *SleepSession
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
func (r *stubEditableSleepSessionRepo) DeleteByIDAndVersion(_ context.Context, _ SleepSessionID, _ int) error {
	return r.deleteErr
}
func (r *stubEditableSleepSessionRepo) FindOverlappingByBabyID(_ context.Context, _ BabyID, _, _ time.Time, _ *SleepSessionID) (SleepSession, error) {
	if r.overlapping.ID() == "" && r.overlapErr == nil {
		return SleepSession{}, apperror.New(apperror.CodeNotFound, "sleep session not found")
	}
	return r.overlapping, r.overlapErr
}

func TestEditSleepSessionUpdatesStartedAt(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	editedStart := startedAt.Add(-15 * time.Minute)
	repo := &stubEditableSleepSessionRepo{session: mustActiveSession(t, startedAt)}
	version := repo.session.Version()

	h := NewEditSleepSessionHandler(repo)
	updated, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:       SleepSessionID("session-1"),
		StartedAt:       &editedStart,
		ExpectedVersion: &version,
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
	version := repo.session.Version()

	h := NewEditSleepSessionHandler(repo)
	updated, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:       SleepSessionID("session-1"),
		StoppedAt:       &stoppedAt,
		ExpectedVersion: &version,
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

	version := 1
	h := NewEditSleepSessionHandler(&stubEditableSleepSessionRepo{})
	_, err := h.Handle(context.Background(), EditSleepSessionCommand{SessionID: SleepSessionID("session-1"), ExpectedVersion: &version})
	if !errors.Is(err, ErrMissingSleepSessionEdit) {
		t.Fatalf("expected ErrMissingSleepSessionEdit, got %v", err)
	}
}

func TestEditSleepSessionRejectsMissingVersion(t *testing.T) {
	t.Parallel()

	h := NewEditSleepSessionHandler(&stubEditableSleepSessionRepo{})
	_, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID: SleepSessionID("session-1"),
		StartedAt: ptrTime(time.Now().UTC()),
	})
	if !errors.Is(err, ErrMissingSleepSessionVersion) {
		t.Fatalf("expected ErrMissingSleepSessionVersion, got %v", err)
	}
}

func TestEditSleepSessionStaleVersionReturnsConflict(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	current := sleepSessionWithVersion(mustActiveSession(t, startedAt), 4)
	expectedVersion := 3
	h := NewEditSleepSessionHandler(&stubEditableSleepSessionRepo{session: current})

	_, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:       SleepSessionID("session-1"),
		StartedAt:       ptrTime(startedAt.Add(-15 * time.Minute)),
		ExpectedVersion: &expectedVersion,
	})

	var conflict SleepSessionConflictError
	if !errors.As(err, &conflict) || conflict.Type != SleepSessionConflictStaleVersion || conflict.CurrentSession == nil {
		t.Fatalf("expected stale version conflict with current session, got %v", err)
	}
}

func TestEditSleepSessionOverlapReturnsConflict(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(2 * time.Hour)
	existing := sleepSessionWithVersion(mustActiveSession(t, startedAt), 2)
	conflicting, err := NewCompletedSleepSession("session-2", "baby-1", "member-1", startedAt.Add(time.Hour), startedAt.Add(3*time.Hour))
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}
	conflicting = sleepSessionWithVersion(conflicting, 1)
	expectedVersion := existing.Version()
	h := NewEditSleepSessionHandler(&stubEditableSleepSessionRepo{
		session:     existing,
		overlapping: conflicting,
	})

	_, err = h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:       SleepSessionID("session-1"),
		StoppedAt:       &stoppedAt,
		ExpectedVersion: &expectedVersion,
	})

	var conflict SleepSessionConflictError
	if !errors.As(err, &conflict) || conflict.Type != SleepSessionConflictOverlap || conflict.ConflictingSession == nil {
		t.Fatalf("expected overlap conflict with conflicting session, got %v", err)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
