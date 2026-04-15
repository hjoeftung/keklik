package sleep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

type stubEditableSleepSessionRepo struct {
	session          SleepSession
	findErr          error
	saveErr          error
	deleteErr        error
	saved            *SleepSession
	deletedSessionID SleepSessionID
}

func (r *stubEditableSleepSessionRepo) Save(_ context.Context, s SleepSession) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = &s
	return nil
}

func (r *stubEditableSleepSessionRepo) FindByID(_ context.Context, _ SleepSessionID) (SleepSession, error) {
	return r.session, r.findErr
}

func (r *stubEditableSleepSessionRepo) FindByIDForFamilyMember(_ context.Context, _ SleepSessionID, _ FamilyMemberID) (SleepSession, error) {
	return r.session, r.findErr
}

func (r *stubEditableSleepSessionRepo) DeleteByIDForFamilyMember(_ context.Context, id SleepSessionID, _ FamilyMemberID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deletedSessionID = id
	return nil
}

func mustEditableCompletedSession(t *testing.T, startedAt, stoppedAt time.Time) SleepSession {
	t.Helper()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	session, err := NewCompletedSleepSession(
		SleepSessionID("session-1"),
		BabyID("baby-1"),
		FamilyMemberID("member-1"),
		startedAt,
		stoppedAt,
		SleepClassificationNap,
		&nw,
	)
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}

	return session
}

func TestEditSleepSessionUpdatesActiveSessionStart(t *testing.T) {
	t.Parallel()

	originalStart := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	editedStart := originalStart.Add(-30 * time.Minute)

	repo := &stubEditableSleepSessionRepo{session: mustActiveSession(t, originalStart)}
	profiles := &stubSleepProfileRepo{profile: mustProfile(t)}

	h := NewEditSleepSessionHandler(repo, profiles)
	updated, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:      SleepSessionID("session-1"),
		FamilyMemberID: FamilyMemberID("member-2"),
		StartedAt:      &editedStart,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if !updated.StartedAt().Equal(editedStart) {
		t.Fatalf("expected started_at %v, got %v", editedStart, updated.StartedAt())
	}
	if !updated.IsActive() {
		t.Fatal("expected updated session to remain active")
	}
	if repo.saved == nil {
		t.Fatal("expected updated session to be saved")
	}
}

func TestEditSleepSessionReclassifiesCompletedSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 19, 30, 0, 0, time.UTC)
	stoppedAt := time.Date(2026, time.April, 15, 4, 30, 0, 0, time.UTC)

	repo := &stubEditableSleepSessionRepo{session: mustEditableCompletedSession(t, startedAt, stoppedAt)}
	profiles := &stubSleepProfileRepo{profile: mustProfile(t)}

	h := NewEditSleepSessionHandler(repo, profiles)
	updated, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:      SleepSessionID("session-1"),
		FamilyMemberID: FamilyMemberID("member-2"),
		StoppedAt:      &stoppedAt,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if updated.Classification() != SleepClassificationNight {
		t.Fatalf("expected reclassified session to be night, got %q", updated.Classification())
	}
	if updated.ClassifiedWithNightWindow() == nil {
		t.Fatal("expected classified night window to be stored")
	}
}

func TestEditSleepSessionRejectsMissingChanges(t *testing.T) {
	t.Parallel()

	repo := &stubEditableSleepSessionRepo{session: mustActiveSession(t, time.Now().UTC())}
	profiles := &stubSleepProfileRepo{profile: mustProfile(t)}

	h := NewEditSleepSessionHandler(repo, profiles)
	_, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:      SleepSessionID("session-1"),
		FamilyMemberID: FamilyMemberID("member-1"),
	})

	if !errors.Is(err, ErrMissingSleepSessionEdit) {
		t.Fatalf("expected ErrMissingSleepSessionEdit, got %v", err)
	}
}

func TestEditSleepSessionRejectsInvalidInterval(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(-time.Minute)

	repo := &stubEditableSleepSessionRepo{session: mustEditableCompletedSession(t, startedAt, startedAt.Add(time.Hour))}
	profiles := &stubSleepProfileRepo{profile: mustProfile(t)}

	h := NewEditSleepSessionHandler(repo, profiles)
	_, err := h.Handle(context.Background(), EditSleepSessionCommand{
		SessionID:      SleepSessionID("session-1"),
		FamilyMemberID: FamilyMemberID("member-1"),
		StoppedAt:      &stoppedAt,
	})

	if !errors.Is(err, ErrInvalidSleepSessionStop) {
		t.Fatalf("expected ErrInvalidSleepSessionStop, got %v", err)
	}
}

func TestDeleteSleepSessionDelegatesToRepository(t *testing.T) {
	t.Parallel()

	repo := &stubEditableSleepSessionRepo{}
	h := NewDeleteSleepSessionHandler(repo)

	if err := h.Handle(context.Background(), DeleteSleepSessionCommand{
		SessionID:      SleepSessionID("session-9"),
		FamilyMemberID: FamilyMemberID("member-1"),
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if repo.deletedSessionID != SleepSessionID("session-9") {
		t.Fatalf("expected session-9 to be deleted, got %q", repo.deletedSessionID)
	}
}

func TestDeleteSleepSessionReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := &stubEditableSleepSessionRepo{
		deleteErr: apperror.New(apperror.CodeNotFound, "sleep session not found"),
	}
	h := NewDeleteSleepSessionHandler(repo)

	err := h.Handle(context.Background(), DeleteSleepSessionCommand{
		SessionID:      SleepSessionID("missing"),
		FamilyMemberID: FamilyMemberID("member-1"),
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeNotFound {
		t.Fatalf("expected not_found, got %q", appErr.Code)
	}
}
