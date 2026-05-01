package sleep

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hjoeftung/keklik/internal/apperror"
)

// StartSleepCommand holds the inputs for starting a sleep session.
type StartSleepCommand struct {
	BabyID            BabyID
	CreatedByMemberID FamilyMemberID
	StartedAt         time.Time // if zero, defaults to time.Now()
}

// StartSleepResult holds the identifiers returned after a sleep session is started.
type StartSleepResult struct {
	ID        SleepSessionID
	StartedAt time.Time
	Version   int
}

// startSleepSessionRepository combines the interfaces required by StartSleepHandler.
type startSleepSessionRepository interface {
	SleepSessionRepository
	ActiveSleepSessionRepository
}

// StartSleepHandler executes the StartSleep use case.
type StartSleepHandler struct {
	sessions startSleepSessionRepository
	now      func() time.Time
}

// NewStartSleepHandler returns a StartSleepHandler backed by the given repository.
func NewStartSleepHandler(sessions startSleepSessionRepository) *StartSleepHandler {
	return &StartSleepHandler{sessions: sessions, now: time.Now}
}

// Handle creates and persists a new active sleep session. It first checks for an
// existing active session and returns a rich conflict error if one exists. The DB
// unique partial index remains as a safety net for the race window.
func (h *StartSleepHandler) Handle(ctx context.Context, cmd StartSleepCommand) (StartSleepResult, error) {
	startedAt := cmd.StartedAt
	if startedAt.IsZero() {
		startedAt = h.now().UTC()
	}

	existing, err := h.sessions.FindActiveByBabyID(ctx, cmd.BabyID)
	if err != nil {
		var appErr apperror.AppError
		if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
			return StartSleepResult{}, err
		}
	} else {
		return StartSleepResult{}, NewActiveSessionConflict(existing)
	}

	id := SleepSessionID(uuid.New().String())

	session, err := NewSleepSession(id, cmd.BabyID, cmd.CreatedByMemberID, startedAt)
	if err != nil {
		return StartSleepResult{}, err
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return StartSleepResult{}, err
	}

	return StartSleepResult{ID: id, StartedAt: startedAt, Version: 1}, nil
}
