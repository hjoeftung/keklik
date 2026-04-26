package sleep

import (
	"context"
	"time"
)

// StopSleepCommand holds the inputs for stopping the active sleep session.
type StopSleepCommand struct {
	BabyID    BabyID
	StoppedAt time.Time // if zero, defaults to time.Now().UTC()
}

// StopSleepResult holds the data returned after a sleep session is stopped.
type StopSleepResult struct {
	ID        SleepSessionID
	StartedAt time.Time
	StoppedAt time.Time
}

// stopSleepSessionRepository combines the interfaces required by StopSleepHandler.
type stopSleepSessionRepository interface {
	SleepSessionRepository
	ActiveSleepSessionRepository
}

// StopSleepHandler executes the StopSleep use case.
type StopSleepHandler struct {
	sessions stopSleepSessionRepository
	now      func() time.Time
}

// NewStopSleepHandler returns a StopSleepHandler backed by the given repositories.
func NewStopSleepHandler(sessions stopSleepSessionRepository) *StopSleepHandler {
	return &StopSleepHandler{sessions: sessions, now: time.Now}
}

// Handle stops the active sleep session for the baby and persists the result.
// When StoppedAt is zero it defaults to time.Now().UTC().
func (h *StopSleepHandler) Handle(ctx context.Context, cmd StopSleepCommand) (StopSleepResult, error) {
	stoppedAt := cmd.StoppedAt
	if stoppedAt.IsZero() {
		stoppedAt = h.now().UTC()
	}

	session, err := h.sessions.FindActiveByBabyID(ctx, cmd.BabyID)
	if err != nil {
		return StopSleepResult{}, err
	}

	if err := session.Stop(stoppedAt); err != nil {
		return StopSleepResult{}, err
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return StopSleepResult{}, err
	}

	return StopSleepResult{
		ID:        session.ID(),
		StartedAt: session.StartedAt(),
		StoppedAt: stoppedAt,
	}, nil
}
