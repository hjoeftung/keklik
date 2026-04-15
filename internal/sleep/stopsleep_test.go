package sleep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

// --- test doubles ---

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

func (r *stubSleepSessionRepo) FindByID(_ context.Context, _ SleepSessionID) (SleepSession, error) {
	return SleepSession{}, errors.New("not implemented")
}

func (r *stubSleepSessionRepo) FindActiveByBabyID(_ context.Context, _ BabyID) (SleepSession, error) {
	return r.active, r.findErr
}

type stubSleepProfileRepo struct {
	profile SleepProfile
	err     error
}

func (r *stubSleepProfileRepo) Save(_ context.Context, _ SleepProfile) error {
	return errors.New("not implemented")
}

func (r *stubSleepProfileRepo) FindByBabyID(_ context.Context, _ BabyID) (SleepProfile, error) {
	return r.profile, r.err
}

// --- helpers ---

func mustProfile(t *testing.T) SleepProfile {
	t.Helper()

	nw := mustNightWindow(t, 21, 0, 7, 0)
	profile, err := NewSleepProfile(BabyID("baby-1"), "UTC", nw)
	if err != nil {
		t.Fatalf("NewSleepProfile: %v", err)
	}

	return profile
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

// --- tests ---

func TestStopSleepHappyPathClassifiesAndSaves(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC) // inside night window 21:00–07:00
	stoppedAt := startedAt.Add(8 * time.Hour)

	sessRepo := &stubSleepSessionRepo{active: mustActiveSession(t, startedAt)}
	profRepo := &stubSleepProfileRepo{profile: mustProfile(t)}

	h := NewStopSleepHandler(sessRepo, profRepo)
	result, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID:    BabyID("baby-1"),
		StoppedAt: stoppedAt,
	})

	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if result.ID != SleepSessionID("session-1") {
		t.Errorf("expected session ID %q, got %q", "session-1", result.ID)
	}

	if !result.StartedAt.Equal(startedAt) {
		t.Errorf("expected StartedAt %v, got %v", startedAt, result.StartedAt)
	}

	if !result.StoppedAt.Equal(stoppedAt) {
		t.Errorf("expected StoppedAt %v, got %v", stoppedAt, result.StoppedAt)
	}

	if result.Classification != SleepClassificationNight {
		t.Errorf("expected night classification, got %q", result.Classification)
	}

	if sessRepo.saved == nil {
		t.Fatal("expected session to be saved")
	}

	if sessRepo.saved.IsActive() {
		t.Error("expected saved session to be completed (not active)")
	}
}

func TestStopSleepDefaultsStoppedAtToNowWhenZero(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC()
	startedAt := before.Add(-30 * time.Minute)

	sessRepo := &stubSleepSessionRepo{active: mustActiveSession(t, startedAt)}
	profRepo := &stubSleepProfileRepo{profile: mustProfile(t)}

	h := NewStopSleepHandler(sessRepo, profRepo)
	result, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID: BabyID("baby-1"),
		// StoppedAt is zero
	})

	after := time.Now().UTC()

	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if result.StoppedAt.Before(before) || result.StoppedAt.After(after) {
		t.Errorf("StoppedAt %v is not between %v and %v", result.StoppedAt, before, after)
	}
}

func TestStopSleepNoActiveSessionReturnsNotFound(t *testing.T) {
	t.Parallel()

	notFound := apperror.New(apperror.CodeNotFound, "sleep session not found")
	sessRepo := &stubSleepSessionRepo{findErr: notFound}
	profRepo := &stubSleepProfileRepo{}

	h := NewStopSleepHandler(sessRepo, profRepo)
	_, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID:    BabyID("baby-1"),
		StoppedAt: time.Now().UTC(),
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}

	if appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected CodeNotFound, got %q", appErr.Code)
	}
}

func TestStopSleepStopBeforeStartReturnsError(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(-time.Second)

	sessRepo := &stubSleepSessionRepo{active: mustActiveSession(t, startedAt)}
	profRepo := &stubSleepProfileRepo{profile: mustProfile(t)}

	h := NewStopSleepHandler(sessRepo, profRepo)
	_, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID:    BabyID("baby-1"),
		StoppedAt: stoppedAt,
	})

	if !errors.Is(err, ErrInvalidSleepSessionStop) {
		t.Fatalf("expected ErrInvalidSleepSessionStop, got %v", err)
	}
}

func TestStopSleepProfileNotFoundReturnsError(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(time.Hour)

	notFound := apperror.New(apperror.CodeNotFound, "sleep profile not found")
	sessRepo := &stubSleepSessionRepo{active: mustActiveSession(t, startedAt)}
	profRepo := &stubSleepProfileRepo{err: notFound}

	h := NewStopSleepHandler(sessRepo, profRepo)
	_, err := h.Handle(context.Background(), StopSleepCommand{
		BabyID:    BabyID("baby-1"),
		StoppedAt: stoppedAt,
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}

	if appErr.Code != apperror.CodeNotFound {
		t.Errorf("expected CodeNotFound, got %q", appErr.Code)
	}
}
