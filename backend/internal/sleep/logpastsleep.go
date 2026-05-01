package sleep

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LogPastSleepCommand struct {
	BabyID            BabyID
	CreatedByMemberID FamilyMemberID
	StartedAt         time.Time
	StoppedAt         time.Time
}

type LogPastSleepResult struct {
	ID        SleepSessionID
	StartedAt time.Time
	StoppedAt time.Time
	Version   int
}

// logPastSleepRepository combines the interfaces required by LogPastSleepHandler.
type logPastSleepRepository interface {
	SleepSessionRepository
	HasOverlappingByBabyID(ctx context.Context, babyID BabyID, startedAt time.Time, stoppedAt time.Time) (bool, error)
}

// LogPastSleepHandler executes the LogPastSleep use case.
type LogPastSleepHandler struct {
	sessions logPastSleepRepository
}

// NewLogPastSleepHandler returns a LogPastSleepHandler backed by the given repository.
func NewLogPastSleepHandler(sessions logPastSleepRepository) *LogPastSleepHandler {
	return &LogPastSleepHandler{sessions: sessions}
}

// Handle creates and persists a completed sleep session from past times.
// It rejects the session if it overlaps any existing session for the same baby.
func (h *LogPastSleepHandler) Handle(ctx context.Context, cmd LogPastSleepCommand) (LogPastSleepResult, error) {
	now := time.Now()
	if cmd.StartedAt.After(now) || cmd.StoppedAt.After(now) {
		return LogPastSleepResult{}, ErrSleepSessionInFuture
	}

	id := SleepSessionID(uuid.New().String())

	session, err := NewCompletedSleepSession(id, cmd.BabyID, cmd.CreatedByMemberID, cmd.StartedAt, cmd.StoppedAt)
	if err != nil {
		return LogPastSleepResult{}, err
	}

	overlaps, err := h.sessions.HasOverlappingByBabyID(ctx, cmd.BabyID, cmd.StartedAt, cmd.StoppedAt)
	if err != nil {
		return LogPastSleepResult{}, err
	}
	if overlaps {
		return LogPastSleepResult{}, ErrSleepSessionOverlap
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return LogPastSleepResult{}, err
	}

	return LogPastSleepResult{ID: id, StartedAt: cmd.StartedAt, StoppedAt: cmd.StoppedAt, Version: 1}, nil
}
