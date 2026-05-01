package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDeleteSleepSessionRejectsMissingVersion(t *testing.T) {
	t.Parallel()

	h := NewDeleteSleepSessionHandler(&stubEditableSleepSessionRepo{})
	err := h.Handle(context.Background(), DeleteSleepSessionCommand{SessionID: "session-1"})
	if !errors.Is(err, ErrMissingSleepSessionVersion) {
		t.Fatalf("expected ErrMissingSleepSessionVersion, got %v", err)
	}
}

func TestDeleteSleepSessionStaleVersionReturnsConflict(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	current := sleepSessionWithVersion(mustActiveSession(t, startedAt), 4)
	expectedVersion := 3
	h := NewDeleteSleepSessionHandler(&stubEditableSleepSessionRepo{session: current})

	err := h.Handle(context.Background(), DeleteSleepSessionCommand{
		SessionID:       "session-1",
		ExpectedVersion: &expectedVersion,
	})

	var conflict SleepSessionConflictError
	if !errors.As(err, &conflict) || conflict.Type != SleepSessionConflictStaleVersion || conflict.CurrentSession == nil {
		t.Fatalf("expected stale version conflict with current session, got %v", err)
	}
}

func TestDeleteSleepSessionDeletesMatchingVersion(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	current := sleepSessionWithVersion(mustActiveSession(t, startedAt), 4)
	expectedVersion := current.Version()
	h := NewDeleteSleepSessionHandler(&stubEditableSleepSessionRepo{session: current})

	err := h.Handle(context.Background(), DeleteSleepSessionCommand{
		SessionID:       "session-1",
		ExpectedVersion: &expectedVersion,
	})
	if err != nil {
		t.Fatalf("expected successful delete, got %v", err)
	}
}
