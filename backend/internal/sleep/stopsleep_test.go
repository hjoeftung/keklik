package sleep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

type stubSleepSessionRepo struct {
	active  SleepSession
	findErr error
	saveErr error
	saved   *SleepSession
}

func (r *stubSleepSessionRepo) Save(_ context.Context, s SleepSession) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = &s
	return nil
}

func (r *stubSleepSessionRepo) SaveAll(_ context.Context, _ []SleepSession) error {
	return nil
}

func (r *stubSleepSessionRepo) FindByID(_ context.Context, _ SleepSessionID) (SleepSession, error) {
	return SleepSession{}, errors.New("not implemented")
}

func (r *stubSleepSessionRepo) FindActiveByBabyID(_ context.Context, _ BabyID) (SleepSession, error) {
	if r.findErr != nil {
		return SleepSession{}, r.findErr
	}
	if r.active.id == "" {
		return SleepSession{}, apperror.New(apperror.CodeNotFound, "no active session")
	}
	return r.active, nil
}

func mustActiveSession(t *testing.T, startedAt time.Time) SleepSession {
	t.Helper()

	session, err := NewSleepSession(
		SleepSessionID("session-1"),
		BabyID("baby-1"),
		FamilyMemberID("member-1"),
		startedAt,
	)
	if err != nil {
		t.Fatalf("NewSleepSession: %v", err)
	}

	return session
}

func TestStopSleepStopsAndSavesSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(8 * time.Hour)
	sessRepo := &stubSleepSessionRepo{active: mustActiveSession(t, startedAt)}

	h := NewStopSleepHandler(sessRepo)
	result, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID:    BabyID("baby-1"),
		StoppedAt: stoppedAt,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if !result.StoppedAt.Equal(stoppedAt) {
		t.Fatalf("expected stop time %v, got %v", stoppedAt, result.StoppedAt)
	}
	if sessRepo.saved == nil {
		t.Fatal("expected session to be saved")
	}
	if sessRepo.saved.IsActive() {
		t.Fatal("expected saved session to be completed")
	}
}

func TestStopSleepDefaultsStoppedAtToNowWhenZero(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC()
	sessRepo := &stubSleepSessionRepo{active: mustActiveSession(t, before.Add(-30*time.Minute))}

	h := NewStopSleepHandler(sessRepo)
	result, err := h.Handle(context.Background(), StopSleepCommand{BabyID: BabyID("baby-1")})
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if result.StoppedAt.Before(before) || result.StoppedAt.After(after) {
		t.Fatalf("StoppedAt %v is not between %v and %v", result.StoppedAt, before, after)
	}
}

func TestStopSleepNoActiveSessionReturnsNotFound(t *testing.T) {
	t.Parallel()

	notFound := apperror.New(apperror.CodeNotFound, "sleep session not found")
	h := NewStopSleepHandler(&stubSleepSessionRepo{findErr: notFound})

	_, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID:    BabyID("baby-1"),
		StoppedAt: time.Now().UTC(),
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}
}

func TestStopSleepStopBeforeStartReturnsError(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	h := NewStopSleepHandler(&stubSleepSessionRepo{active: mustActiveSession(t, startedAt)})

	_, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID:    BabyID("baby-1"),
		StoppedAt: startedAt.Add(-time.Second),
	})
	if !errors.Is(err, ErrInvalidSleepSessionStop) {
		t.Fatalf("expected ErrInvalidSleepSessionStop, got %v", err)
	}
}
