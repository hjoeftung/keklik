package sleep

import (
	"context"
	"testing"
	"time"
)

type inMemoryNightWindowRepository struct {
	windows   []NightWindow
	saved     []NightWindow
	deleted   []NightWindowID
	findErr   error
	saveErr   error
	deleteErr error
}

func (r *inMemoryNightWindowRepository) Save(_ context.Context, nw NightWindow) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = append(r.saved, nw)
	return nil
}

func (r *inMemoryNightWindowRepository) DeleteByIDs(_ context.Context, ids []NightWindowID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deleted = append(r.deleted, ids...)
	return nil
}

func (r *inMemoryNightWindowRepository) FindByBabyID(_ context.Context, _ BabyID) ([]NightWindow, error) {
	return r.windows, r.findErr
}

type noopTransactor struct{}

func (noopTransactor) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestSetNightWindowRejectsZeroEffectiveFrom(t *testing.T) {
	t.Parallel()

	h := NewSetNightWindowHandler(&inMemoryNightWindowRepository{}, noopTransactor{})
	err := h.Handle(context.Background(), SetNightWindowCommand{
		BabyID:                 BabyID("baby-1"),
		NightWindowStartHour:   21,
		NightWindowStartMinute: 0,
		NightWindowEndHour:     7,
		NightWindowEndMinute:   0,
	})
	if err != ErrZeroNightWindowEffectiveFrom {
		t.Fatalf("expected ErrZeroNightWindowEffectiveFrom, got %v", err)
	}
}

func TestSetNightWindowReplacesFutureWindowsAndClosesPrevious(t *testing.T) {
	t.Parallel()

	effectiveFrom := time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	old1, _ := NewNightWindow("w1", "baby-1", mustLocalTime(t, 20, 0), mustLocalTime(t, 6, 0), effectiveFrom.AddDate(0, 0, -10), nil)
	old2, _ := NewNightWindow("w2", "baby-1", mustLocalTime(t, 21, 0), mustLocalTime(t, 7, 0), effectiveFrom.AddDate(0, 0, 1), nil)
	repo := &inMemoryNightWindowRepository{windows: []NightWindow{old1, old2}}

	h := NewSetNightWindowHandler(repo, noopTransactor{})
	err := h.Handle(context.Background(), SetNightWindowCommand{
		BabyID:                 BabyID("baby-1"),
		NightWindowStartHour:   22,
		NightWindowStartMinute: 0,
		NightWindowEndHour:     8,
		NightWindowEndMinute:   0,
		EffectiveFrom:          effectiveFrom,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if len(repo.deleted) != 1 || repo.deleted[0] != NightWindowID("w2") {
		t.Fatalf("expected future window w2 to be deleted, got %v", repo.deleted)
	}
	if len(repo.saved) != 2 {
		t.Fatalf("expected previous+new window to be saved, got %d saves", len(repo.saved))
	}
	if repo.saved[0].EffectiveTo() == nil || !repo.saved[0].EffectiveTo().Equal(effectiveFrom) {
		t.Fatalf("expected previous window effective_to to be %v, got %v", effectiveFrom, repo.saved[0].EffectiveTo())
	}
	if repo.saved[1].EffectiveTo() != nil {
		t.Fatalf("expected new window to be open-ended, got %v", repo.saved[1].EffectiveTo())
	}
}

func mustLocalTime(t *testing.T, hour, minute int) LocalTime {
	t.Helper()

	lt, err := NewLocalTime(hour, minute)
	if err != nil {
		t.Fatalf("NewLocalTime: %v", err)
	}
	return lt
}
