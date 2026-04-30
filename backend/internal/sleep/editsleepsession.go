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
}

// NewEditSleepSessionHandler returns an EditSleepSessionHandler backed by the given repositories.
func NewEditSleepSessionHandler(sessions EditableSleepSessionRepository) *EditSleepSessionHandler {
	return &EditSleepSessionHandler{sessions: sessions}
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

	updated, err := rebuildSleepSession(existing, startedAt, stoppedAt, completed)
	if err != nil {
		return SleepSession{}, err
	}

	if err := h.sessions.Save(ctx, updated); err != nil {
		return SleepSession{}, err
	}

	return updated, nil
}

func rebuildSleepSession(existing SleepSession, startedAt, stoppedAt time.Time, completed bool) (SleepSession, error) {
	var stoppedAtPtr *time.Time
	if completed {
		stoppedAtPtr = &stoppedAt
	}
	return RestoreSleepSession(
		existing.ID(),
		existing.BabyID(),
		existing.CreatedByMemberID(),
		startedAt,
		stoppedAtPtr,
		existing.Version(),
	)
}
