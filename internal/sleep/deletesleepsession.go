package sleep

import "context"

// DeleteSleepSessionCommand holds the inputs for deleting a sleep session.
type DeleteSleepSessionCommand struct {
	SessionID      SleepSessionID
	FamilyMemberID FamilyMemberID
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
	return h.sessions.DeleteByIDForFamilyMember(ctx, cmd.SessionID, cmd.FamilyMemberID)
}
