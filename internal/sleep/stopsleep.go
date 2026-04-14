package sleep

import (
	"context"
	"time"
)

// CurrentClassificationRuleVersion is the version of the classification rules
// currently in use. Increment this when the classification logic changes so
// that past sessions can be detected and reclassified if needed.
const CurrentClassificationRuleVersion ClassificationRuleVersion = 1

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
}

// NewStopSleepHandler returns a StopSleepHandler backed by the given repositories.
func NewStopSleepHandler(sessions stopSleepSessionRepository, profiles SleepProfileRepository) *StopSleepHandler {
	return &StopSleepHandler{sessions: sessions, profiles: profiles}
}

// Handle stops the active sleep session for the baby, classifies it, and
// persists the result. When StoppedAt is zero it defaults to time.Now().UTC().
func (h *StopSleepHandler) Handle(ctx context.Context, cmd StopSleepCommand) (StopSleepResult, error) {
	stoppedAt := cmd.StoppedAt
	if stoppedAt.IsZero() {
		stoppedAt = time.Now().UTC()
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
		CurrentClassificationRuleVersion,
	)
	if err != nil {
		// NewCompletedSleepSession returns ErrInvalidSleepSessionStop when
		// stoppedAt < startedAt, which is the expected validation error.
		return StopSleepResult{}, err
	}

	classification, err := Classify(tentative, profile.Timezone(), profile.NightWindow())
	if err != nil {
		return StopSleepResult{}, err
	}

	if err := session.Stop(stoppedAt, classification, CurrentClassificationRuleVersion); err != nil {
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
