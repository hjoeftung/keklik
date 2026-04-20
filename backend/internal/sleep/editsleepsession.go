package sleep

import (
	"context"
	"time"
)

// EditSleepSessionCommand holds the inputs for editing a sleep session.
type EditSleepSessionCommand struct {
	SessionID SleepSessionID
	StartedAt *time.Time
	StoppedAt *time.Time
}

// EditSleepSessionHandler executes the EditSleepSession use case.
type EditSleepSessionHandler struct {
	sessions EditableSleepSessionRepository
	profiles SleepProfileRepository
}

// NewEditSleepSessionHandler returns an EditSleepSessionHandler backed by the given repositories.
func NewEditSleepSessionHandler(sessions EditableSleepSessionRepository, profiles SleepProfileRepository) *EditSleepSessionHandler {
	return &EditSleepSessionHandler{sessions: sessions, profiles: profiles}
}

// Handle updates the requested sleep session and reclassifies it when the
// resulting session is completed.
func (h *EditSleepSessionHandler) Handle(ctx context.Context, cmd EditSleepSessionCommand) (SleepSession, error) {
	if cmd.StartedAt == nil && cmd.StoppedAt == nil {
		return SleepSession{}, ErrMissingSleepSessionEdit
	}

	existing, err := h.sessions.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return SleepSession{}, err
	}

	startedAt := existing.StartedAt()
	if cmd.StartedAt != nil {
		startedAt = *cmd.StartedAt
	}

	stoppedAt, completed := existing.StoppedAt()
	if cmd.StoppedAt != nil {
		stoppedAt = *cmd.StoppedAt
		completed = true
	}

	updated, err := rebuildSleepSession(ctx, existing, startedAt, stoppedAt, completed, h.profiles)
	if err != nil {
		return SleepSession{}, err
	}

	if err := h.sessions.Save(ctx, updated); err != nil {
		return SleepSession{}, err
	}

	return updated, nil
}

func rebuildSleepSession(ctx context.Context, existing SleepSession, startedAt, stoppedAt time.Time, completed bool, profiles SleepProfileRepository) (SleepSession, error) {
	if !completed {
		return NewSleepSession(existing.ID(), existing.BabyID(), existing.CreatedByMemberID(), startedAt)
	}

	profile, err := profiles.FindByBabyID(ctx, existing.BabyID())
	if err != nil {
		return SleepSession{}, err
	}

	tentative, err := NewCompletedSleepSession(
		existing.ID(),
		existing.BabyID(),
		existing.CreatedByMemberID(),
		startedAt,
		stoppedAt,
		SleepClassificationNap,
		nil,
	)
	if err != nil {
		return SleepSession{}, err
	}

	nightWindow := profile.NightWindow()
	classification, err := Classify(tentative, profile.Timezone(), nightWindow)
	if err != nil {
		return SleepSession{}, err
	}

	return NewCompletedSleepSession(
		existing.ID(),
		existing.BabyID(),
		existing.CreatedByMemberID(),
		startedAt,
		stoppedAt,
		classification,
		&nightWindow,
	)
}
