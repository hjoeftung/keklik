package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubSleepSessionHistoryRepository struct {
	sessions []SleepSession
	err      error
}

func (r *stubSleepSessionHistoryRepository) FindByBabyIDAndDateRange(_ context.Context, _ BabyID, _ DateRange) ([]SleepSession, error) {
	return r.sessions, r.err
}

type stubNightWindowRepository struct {
	windows []NightWindow
	err     error
}

func (r *stubNightWindowRepository) Save(_ context.Context, _ NightWindow) error {
	return nil
}

func (r *stubNightWindowRepository) DeleteByIDs(_ context.Context, _ []NightWindowID) error {
	return nil
}

func (r *stubNightWindowRepository) FindByBabyID(_ context.Context, _ BabyID) ([]NightWindow, error) {
	return r.windows, r.err
}

func TestPeriodToDateRangeToday(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 16, 14, 30, 0, 0, time.UTC)
	dr, err := periodToDateRange("today", "UTC", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	todayMidnight := time.Date(2026, time.April, 16, 0, 0, 0, 0, time.UTC)
	if !dr.Start().Equal(todayMidnight) {
		t.Fatalf("expected start %v, got %v", todayMidnight, dr.Start())
	}
}

func TestGetSleepHistoryHandlerReturnsEmptySliceWhenNoSessions(t *testing.T) {
	t.Parallel()

	repo := &stubSleepSessionHistoryRepository{}
	windows := &stubNightWindowRepository{windows: []NightWindow{mustNightWindow(t, 21, 0, 7, 0)}}
	h := NewGetSleepHistoryHandler(repo, windows)

	entries, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID:   BabyID("baby-1"),
		Timezone: "UTC",
		Period:   "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestGetSleepHistoryHandlerClassifiesSessionsFromApplicableWindow(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.April, 16, 22, 0, 0, 0, time.UTC)
	stop := start.Add(8 * time.Hour)
	session, err := NewCompletedSleepSession(
		SleepSessionID("session-1"),
		BabyID("baby-1"),
		FamilyMemberID("member-1"),
		start,
		stop,
	)
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}

	h := NewGetSleepHistoryHandler(
		&stubSleepSessionHistoryRepository{sessions: []SleepSession{session}},
		&stubNightWindowRepository{windows: []NightWindow{mustNightWindow(t, 21, 0, 7, 0)}},
	)

	entries, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID:   BabyID("baby-1"),
		Timezone: "UTC",
		Period:   "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Classification != SleepClassificationNight {
		t.Fatalf("expected night classification, got %q", entries[0].Classification)
	}
}

func TestGetSleepHistoryHandlerRequiresTimezone(t *testing.T) {
	t.Parallel()

	h := NewGetSleepHistoryHandler(&stubSleepSessionHistoryRepository{}, &stubNightWindowRepository{})
	_, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID: BabyID("baby-1"),
		Period: "7d",
	})
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestGetSleepHistoryHandlerInvalidPeriodReturnsError(t *testing.T) {
	t.Parallel()

	h := NewGetSleepHistoryHandler(&stubSleepSessionHistoryRepository{}, &stubNightWindowRepository{})
	_, err := h.Handle(context.Background(), GetSleepHistoryQuery{
		BabyID:   BabyID("baby-1"),
		Timezone: "UTC",
		Period:   "30d",
	})
	if !errors.Is(err, ErrInvalidSleepHistoryPeriod) {
		t.Fatalf("expected ErrInvalidSleepHistoryPeriod, got %v", err)
	}
}
