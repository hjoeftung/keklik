package sleep

import (
	"context"
	"errors"
	"testing"
)

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

	repo := &inMemorySleepProfileRepository{}
	h := NewCreateSleepProfileHandler(repo)

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

	repo := &inMemorySleepProfileRepository{}
	h := NewCreateSleepProfileHandler(repo)

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

	repo := &inMemorySleepProfileRepository{}
	h := NewCreateSleepProfileHandler(repo)

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

	repo := &inMemorySleepProfileRepository{}
	h := NewCreateSleepProfileHandler(repo)

	err := h.Handle(context.Background(), CreateSleepProfileCommand{
		BabyID:                 "baby-1",
		Timezone:               "Europe/Berlin",
		NightWindowStartHour:   20,
		NightWindowStartMinute: 30,
		NightWindowEndHour:     7,
		NightWindowEndMinute:   0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved profile, got %d", len(repo.saved))
	}

	saved := repo.saved[0]
	if saved.BabyID() != BabyID("baby-1") {
		t.Errorf("expected baby id %q, got %q", "baby-1", saved.BabyID())
	}
	if saved.Timezone() != "Europe/Berlin" {
		t.Errorf("expected timezone %q, got %q", "Europe/Berlin", saved.Timezone())
	}
}

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
