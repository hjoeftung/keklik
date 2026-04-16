package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- test doubles ---

type inMemorySleepProfileRepository struct {
	saved []SleepProfile
	err   error
}

func (r *inMemorySleepProfileRepository) Save(_ context.Context, p SleepProfile) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, p)
	return nil
}

func (r *inMemorySleepProfileRepository) FindByBabyID(_ context.Context, _ BabyID) (SleepProfile, error) {
	return SleepProfile{}, errors.New("not implemented")
}

type inMemoryCompletedSessionsRepo struct {
	sessions []SleepSession
	err      error
}

func (r *inMemoryCompletedSessionsRepo) FindCompletedByBabyIDSince(_ context.Context, _ BabyID, _ time.Time) ([]SleepSession, error) {
	return r.sessions, r.err
}

type inMemorySleepSessionWriter struct {
	saved []SleepSession
	err   error
}

func (r *inMemorySleepSessionWriter) Save(_ context.Context, s SleepSession) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, s)
	return nil
}

func (r *inMemorySleepSessionWriter) FindByID(_ context.Context, _ SleepSessionID) (SleepSession, error) {
	return SleepSession{}, errors.New("not implemented")
}

// --- helpers ---

func mustNightWindow(t *testing.T, startHour, startMinute, endHour, endMinute int) NightWindow {
	t.Helper()

	start, err := NewLocalTime(startHour, startMinute)
	if err != nil {
		t.Fatalf("NewLocalTime returned error: %v", err)
	}

	end, err := NewLocalTime(endHour, endMinute)
	if err != nil {
		t.Fatalf("NewLocalTime returned error: %v", err)
	}

	nw, err := NewNightWindow(start, end)
	if err != nil {
		t.Fatalf("NewNightWindow returned error: %v", err)
	}

	return nw
}

func newTestHandler(profiles *inMemorySleepProfileRepository, sessions *inMemoryCompletedSessionsRepo, writer *inMemorySleepSessionWriter) *CreateSleepProfileHandler {
	h := NewCreateSleepProfileHandler(profiles, sessions, writer)
	return h
}

func validCmd() CreateSleepProfileCommand {
	return CreateSleepProfileCommand{
		BabyID:                 "baby-1",
		Timezone:               "Europe/Berlin",
		NightWindowStartHour:   20,
		NightWindowStartMinute: 30,
		NightWindowEndHour:     7,
		NightWindowEndMinute:   0,
	}
}

// --- existing validation tests ---

func TestNewSleepProfileRejectsInvalidTimezone(t *testing.T) {
	t.Parallel()

	nw := mustNightWindow(t, 20, 0, 6, 0)
	_, err := NewSleepProfile(BabyID("baby-1"), "Mars/Phobos", nw)
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestNewNightWindowRejectsEqualBounds(t *testing.T) {
	t.Parallel()

	start, err := NewLocalTime(20, 0)
	if err != nil {
		t.Fatalf("NewLocalTime returned error: %v", err)
	}

	_, err = NewNightWindow(start, start)
	if !errors.Is(err, ErrInvalidNightWindow) {
		t.Fatalf("expected ErrInvalidNightWindow, got %v", err)
	}
}

func TestCreateSleepProfileRejectsInvalidTimezone(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&inMemorySleepProfileRepository{}, &inMemoryCompletedSessionsRepo{}, &inMemorySleepSessionWriter{})

	err := h.Handle(context.Background(), CreateSleepProfileCommand{
		BabyID:               "baby-1",
		Timezone:             "Not/ATimezone",
		NightWindowStartHour: 20,
		NightWindowEndHour:   6,
	})
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Errorf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestCreateSleepProfileRejectsInvalidNightWindow(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&inMemorySleepProfileRepository{}, &inMemoryCompletedSessionsRepo{}, &inMemorySleepSessionWriter{})

	err := h.Handle(context.Background(), CreateSleepProfileCommand{
		BabyID:                 "baby-1",
		Timezone:               "Europe/Berlin",
		NightWindowStartHour:   20,
		NightWindowStartMinute: 0,
		NightWindowEndHour:     20,
		NightWindowEndMinute:   0,
	})
	if !errors.Is(err, ErrInvalidNightWindow) {
		t.Errorf("expected ErrInvalidNightWindow, got %v", err)
	}
}

func TestCreateSleepProfileRejectsInvalidLocalTime(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&inMemorySleepProfileRepository{}, &inMemoryCompletedSessionsRepo{}, &inMemorySleepSessionWriter{})

	err := h.Handle(context.Background(), CreateSleepProfileCommand{
		BabyID:               "baby-1",
		Timezone:             "Europe/Berlin",
		NightWindowStartHour: 25,
	})
	if !errors.Is(err, ErrInvalidLocalTime) {
		t.Errorf("expected ErrInvalidLocalTime, got %v", err)
	}
}

func TestCreateSleepProfilePersistsProfile(t *testing.T) {
	t.Parallel()

	profiles := &inMemorySleepProfileRepository{}
	h := newTestHandler(profiles, &inMemoryCompletedSessionsRepo{}, &inMemorySleepSessionWriter{})

	err := h.Handle(context.Background(), validCmd())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(profiles.saved) != 1 {
		t.Fatalf("expected 1 saved profile, got %d", len(profiles.saved))
	}

	saved := profiles.saved[0]
	if saved.BabyID() != BabyID("baby-1") {
		t.Errorf("expected baby id %q, got %q", "baby-1", saved.BabyID())
	}
	if saved.Timezone() != "Europe/Berlin" {
		t.Errorf("expected timezone %q, got %q", "Europe/Berlin", saved.Timezone())
	}
}

// --- effective_from tests ---

func TestCreateSleepProfileWithoutEffectiveFromDoesNotReclassify(t *testing.T) {
	t.Parallel()

	sessions := &inMemoryCompletedSessionsRepo{}
	writer := &inMemorySleepSessionWriter{}
	h := newTestHandler(&inMemorySleepProfileRepository{}, sessions, writer)

	cmd := validCmd()
	if err := h.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(writer.saved) != 0 {
		t.Errorf("expected no sessions reclassified, got %d", len(writer.saved))
	}
}

func TestCreateSleepProfileEffectiveFromTooOld(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	tooOld := fixedNow.AddDate(0, -1, 0).Add(-time.Second) // 1 second before the 30-day boundary

	h := newTestHandler(&inMemorySleepProfileRepository{}, &inMemoryCompletedSessionsRepo{}, &inMemorySleepSessionWriter{})
	h.now = func() time.Time { return fixedNow }

	cmd := validCmd()
	cmd.EffectiveFrom = &tooOld

	err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrEffectiveFromTooOld) {
		t.Errorf("expected ErrEffectiveFromTooOld, got %v", err)
	}
}

func TestCreateSleepProfileEffectiveFromAtBoundaryIsValid(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	boundary := fixedNow.AddDate(0, -1, 0) // exactly 30 days ago

	h := newTestHandler(&inMemorySleepProfileRepository{}, &inMemoryCompletedSessionsRepo{}, &inMemorySleepSessionWriter{})
	h.now = func() time.Time { return fixedNow }

	cmd := validCmd()
	cmd.EffectiveFrom = &boundary

	if err := h.Handle(context.Background(), cmd); err != nil {
		t.Errorf("expected no error at 30-day boundary, got %v", err)
	}
}

func TestCreateSleepProfileEffectiveFromZeroSessionsIsNoOp(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	effectiveFrom := fixedNow.AddDate(0, 0, -7)

	sessions := &inMemoryCompletedSessionsRepo{sessions: nil}
	writer := &inMemorySleepSessionWriter{}
	h := newTestHandler(&inMemorySleepProfileRepository{}, sessions, writer)
	h.now = func() time.Time { return fixedNow }

	cmd := validCmd()
	cmd.EffectiveFrom = &effectiveFrom

	if err := h.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(writer.saved) != 0 {
		t.Errorf("expected 0 reclassified sessions, got %d", len(writer.saved))
	}
}

func TestCreateSleepProfileEffectiveFromReclassifiesSessions(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	effectiveFrom := fixedNow.AddDate(0, 0, -7)

	// A nap-hour session (12:00–13:00) — should be classified as nap with any standard night window.
	napStart := fixedNow.AddDate(0, 0, -3).Add(12 * time.Hour)
	napStop := napStart.Add(time.Hour)
	nw := mustNightWindow(t, 21, 0, 7, 0)
	session, err := NewCompletedSleepSession("s-1", "baby-1", "member-1", napStart, napStop, SleepClassificationNight, &nw)
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}

	sessions := &inMemoryCompletedSessionsRepo{sessions: []SleepSession{session}}
	writer := &inMemorySleepSessionWriter{}
	h := newTestHandler(&inMemorySleepProfileRepository{}, sessions, writer)
	h.now = func() time.Time { return fixedNow }

	// Profile has night window 21:00–07:00 UTC; the session is 12:00–13:00 UTC so it should become a nap.
	cmd := CreateSleepProfileCommand{
		BabyID:               "baby-1",
		Timezone:             "UTC",
		NightWindowStartHour: 21,
		NightWindowEndHour:   7,
		EffectiveFrom:        &effectiveFrom,
	}

	if err := h.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(writer.saved) != 1 {
		t.Fatalf("expected 1 reclassified session, got %d", len(writer.saved))
	}

	if writer.saved[0].Classification() != SleepClassificationNap {
		t.Errorf("expected classification %q, got %q", SleepClassificationNap, writer.saved[0].Classification())
	}
}

func TestCreateSleepProfileEffectiveFromNightReclassification(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	effectiveFrom := fixedNow.AddDate(0, 0, -7)

	// A night-hour session (22:00–06:00) — should be classified as night with a 21:00–07:00 window.
	nightStart := time.Date(2026, 4, 13, 22, 0, 0, 0, time.UTC)
	nightStop := time.Date(2026, 4, 14, 6, 0, 0, 0, time.UTC)
	nw := mustNightWindow(t, 9, 0, 10, 0) // wrong window stored with session
	session, err := NewCompletedSleepSession("s-2", "baby-1", "member-1", nightStart, nightStop, SleepClassificationNap, &nw)
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}

	sessions := &inMemoryCompletedSessionsRepo{sessions: []SleepSession{session}}
	writer := &inMemorySleepSessionWriter{}
	h := newTestHandler(&inMemorySleepProfileRepository{}, sessions, writer)
	h.now = func() time.Time { return fixedNow }

	cmd := CreateSleepProfileCommand{
		BabyID:               "baby-1",
		Timezone:             "UTC",
		NightWindowStartHour: 21,
		NightWindowEndHour:   7,
		EffectiveFrom:        &effectiveFrom,
	}

	if err := h.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(writer.saved) != 1 {
		t.Fatalf("expected 1 reclassified session, got %d", len(writer.saved))
	}

	if writer.saved[0].Classification() != SleepClassificationNight {
		t.Errorf("expected classification %q, got %q", SleepClassificationNight, writer.saved[0].Classification())
	}
}
