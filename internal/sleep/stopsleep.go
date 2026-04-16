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
	ID             SleepSessionID
	StartedAt      time.Time
	StoppedAt      time.Time
	Classification SleepClassification
}

// stopSleepSessionRepository combines the interfaces required by StopSleepHandler.
type stopSleepSessionRepository interface {
	SleepSessionRepository
	ActiveSleepSessionRepository
}

// StopSleepHandler executes the StopSleep use case.
type StopSleepHandler struct {
	sessions stopSleepSessionRepository
	profiles SleepProfileRepository
	now      func() time.Time
}

// NewStopSleepHandler returns a StopSleepHandler backed by the given repositories.
func NewStopSleepHandler(sessions stopSleepSessionRepository, profiles SleepProfileRepository) *StopSleepHandler {
	return &StopSleepHandler{sessions: sessions, profiles: profiles, now: time.Now}
}

// Handle stops the active sleep session for the baby, classifies it, and
// persists the result. When StoppedAt is zero it defaults to time.Now().UTC().
func (h *StopSleepHandler) Handle(ctx context.Context, cmd StopSleepCommand) (StopSleepResult, error) {
	stoppedAt := cmd.StoppedAt
	if stoppedAt.IsZero() {
		stoppedAt = h.now().UTC()
	}

	session, err := h.sessions.FindActiveByBabyID(ctx, cmd.BabyID)
	if err != nil {
		return StopSleepResult{}, err
	}

	profile, err := h.profiles.FindByBabyID(ctx, cmd.BabyID)
	if err != nil {
		return StopSleepResult{}, err
	}

	// Build a temporary stopped session to classify without mutating the real
	// session first. This lets us pass the stoppedAt time to Classify.
	tentative, err := NewCompletedSleepSession(
		session.ID(),
		session.BabyID(),
		session.CreatedByMemberID(),
		session.StartedAt(),
		stoppedAt,
		SleepClassificationNap, // placeholder; replaced below
		nil,
	)
	if err != nil {
		// NewCompletedSleepSession returns ErrInvalidSleepSessionStop when
		// stoppedAt < startedAt, which is the expected validation error.
		return StopSleepResult{}, err
	}

	nightWindow := profile.NightWindow()
	classification, err := Classify(tentative, profile.Timezone(), nightWindow)
	if err != nil {
		return StopSleepResult{}, err
	}

	if err := session.Stop(stoppedAt, classification, nightWindow); err != nil {
		return StopSleepResult{}, err
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return StopSleepResult{}, err
	}

	return StopSleepResult{
		ID:             session.ID(),
		StartedAt:      session.StartedAt(),
		StoppedAt:      stoppedAt,
		Classification: classification,
	}, nil
}
