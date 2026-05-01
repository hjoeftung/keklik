package sleep

import "context"

// DeleteSleepSessionCommand holds the inputs for deleting a sleep session.
type DeleteSleepSessionCommand struct {
	SessionID       SleepSessionID
	ExpectedVersion *int
}

// DeleteSleepSessionHandler executes the DeleteSleepSession use case.
type DeleteSleepSessionHandler struct {
	sessions EditableSleepSessionRepository
}

// NewDeleteSleepSessionHandler returns a DeleteSleepSessionHandler backed by the given repository.
func NewDeleteSleepSessionHandler(sessions EditableSleepSessionRepository) *DeleteSleepSessionHandler {
	return &DeleteSleepSessionHandler{sessions: sessions}
}

// Handle hard-deletes the requested sleep session after verifying it belongs to
// the caller's family.
func (h *DeleteSleepSessionHandler) Handle(ctx context.Context, cmd DeleteSleepSessionCommand) error {
	if cmd.ExpectedVersion == nil {
		return ErrMissingSleepSessionVersion
	}

	current, err := h.sessions.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}
	if current.Version() != *cmd.ExpectedVersion {
		return NewStaleSleepSessionConflict(current)
	}

	if err := h.sessions.DeleteByIDAndVersion(ctx, cmd.SessionID, *cmd.ExpectedVersion); err != nil {
		current, findErr := h.sessions.FindByID(ctx, cmd.SessionID)
		if findErr == nil {
			return NewStaleSleepSessionConflict(current)
		}
		return findErr
	}
	return nil
}
