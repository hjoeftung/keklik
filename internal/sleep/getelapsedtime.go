package sleep

import (
	"context"
	"fmt"
	"time"
)

// GetElapsedTimeResult holds the elapsed-time values returned by GetElapsedTimeHandler.
// Fields are nil when the underlying data is absent.
type GetElapsedTimeResult struct {
	// TimeSinceLastSleepStart is the duration since the most recent session started.
	// Nil when there are no sessions at all.
	TimeSinceLastSleepStart *time.Duration

	// TimeSinceLastAwakening is the duration since the most recent session stopped
	// (i.e. when the baby last woke up). Nil when there are no completed sessions
	// or when the baby is currently sleeping.
	TimeSinceLastAwakening *time.Duration
}

// GetElapsedTimeHandler executes the GetElapsedTime use case.
type GetElapsedTimeHandler struct {
	sessions SleepSessionQueryRepository
}

// NewGetElapsedTimeHandler returns a GetElapsedTimeHandler backed by the given repository.
func NewGetElapsedTimeHandler(sessions SleepSessionQueryRepository) *GetElapsedTimeHandler {
	return &GetElapsedTimeHandler{sessions: sessions}
}

// Handle returns elapsed-time data for the given baby's most recent sleep activity.
func (h *GetElapsedTimeHandler) Handle(ctx context.Context, babyID BabyID) (GetElapsedTimeResult, error) {
	session, found, err := h.sessions.FindMostRecentByBabyID(ctx, babyID)
	if err != nil {
		return GetElapsedTimeResult{}, fmt.Errorf("get elapsed time: %w", err)
	}

	if !found {
		return GetElapsedTimeResult{}, nil
	}

	sinceStart := time.Since(session.StartedAt())
	result := GetElapsedTimeResult{
		TimeSinceLastSleepStart: &sinceStart,
	}

	if stoppedAt, ok := session.StoppedAt(); ok {
		sinceAwakening := time.Since(stoppedAt)
		result.TimeSinceLastAwakening = &sinceAwakening
	}

	return result, nil
}
