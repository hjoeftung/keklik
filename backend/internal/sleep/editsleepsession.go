package sleep

import (
	"context"
	"errors"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

// EditSleepSessionCommand holds the inputs for editing a sleep session.
type EditSleepSessionCommand struct {
	SessionID       SleepSessionID
	StartedAt       *time.Time
	StoppedAt       *time.Time
	ExpectedVersion *int
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
	if cmd.ExpectedVersion == nil {
		return SleepSession{}, ErrMissingSleepSessionVersion
	}
	if cmd.StartedAt == nil && cmd.StoppedAt == nil {
		return SleepSession{}, ErrMissingSleepSessionEdit
	}

	existing, err := h.sessions.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return SleepSession{}, err
	}
	if existing.Version() != *cmd.ExpectedVersion {
		return SleepSession{}, NewStaleSleepSessionConflict(existing)
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

	if completed {
		excludeID := existing.ID()
		conflicting, err := h.sessions.FindOverlappingByBabyID(ctx, existing.BabyID(), startedAt, stoppedAt, &excludeID)
		if err == nil {
			return SleepSession{}, NewOverlapSleepSessionConflict(conflicting)
		}
		if !isNotFound(err) {
			return SleepSession{}, err
		}
	}

	if err := h.sessions.Save(ctx, updated); err != nil {
		current, findErr := h.sessions.FindByID(ctx, existing.ID())
		if findErr == nil {
			if current.Version() != *cmd.ExpectedVersion {
				return SleepSession{}, NewStaleSleepSessionConflict(current)
			}
		}
		if completed {
			excludeID := existing.ID()
			conflicting, overlapErr := h.sessions.FindOverlappingByBabyID(ctx, existing.BabyID(), startedAt, stoppedAt, &excludeID)
			if overlapErr == nil {
				return SleepSession{}, NewOverlapSleepSessionConflict(conflicting)
			}
			if !isNotFound(overlapErr) {
				return SleepSession{}, err
			}
		}
		return SleepSession{}, err
	}

	return sleepSessionWithVersion(updated, updated.Version()+1), nil
}

func isNotFound(err error) bool {
	var appErr apperror.AppError
	return errors.As(err, &appErr) && appErr.Code == apperror.CodeNotFound
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

func sleepSessionWithVersion(s SleepSession, version int) SleepSession {
	stoppedAt, completed := s.StoppedAt()
	var stoppedAtPtr *time.Time
	if completed {
		stoppedAtPtr = &stoppedAt
	}
	restored, err := RestoreSleepSession(s.ID(), s.BabyID(), s.CreatedByMemberID(), s.StartedAt(), stoppedAtPtr, version)
	if err != nil {
		return s
	}
	return restored
}
