package sleep

import (
	"context"
	"time"

	"github.com/google/uuid"
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

// StartSleepHandler executes the StartSleep use case.
type StartSleepHandler struct {
	sessions SleepSessionRepository
	now      func() time.Time
}

// NewStartSleepHandler returns a StartSleepHandler backed by the given repository.
func NewStartSleepHandler(sessions SleepSessionRepository) *StartSleepHandler {
	return &StartSleepHandler{sessions: sessions, now: time.Now}
}

// Handle creates and persists a new active sleep session. Duplicate active sessions
// for the same baby are rejected by the database unique partial index; the repository
// maps that violation to an AppError with CodeActiveSleepExists.
func (h *StartSleepHandler) Handle(ctx context.Context, cmd StartSleepCommand) (StartSleepResult, error) {
	startedAt := cmd.StartedAt
	if startedAt.IsZero() {
		startedAt = h.now().UTC()
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
