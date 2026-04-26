package sleep

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrZeroNightWindowEffectiveFrom = errors.New("effective_from must not be zero")

// SetNightWindowCommand holds the inputs for setting a baby's active night window.
type SetNightWindowCommand struct {
	BabyID                 BabyID
	NightWindowStartHour   int
	NightWindowStartMinute int
	NightWindowEndHour     int
	NightWindowEndMinute   int
	EffectiveFrom          time.Time
}

// SetNightWindowHandler executes the SetNightWindow use case.
type SetNightWindowHandler struct {
	windows    NightWindowRepository
	transactor Transactor
}

// NewSetNightWindowHandler returns a SetNightWindowHandler backed by the given repositories.
func NewSetNightWindowHandler(windows NightWindowRepository, transactor Transactor) *SetNightWindowHandler {
	return &SetNightWindowHandler{windows: windows, transactor: transactor}
}

// Handle replaces the active night-window timeline from cmd.EffectiveFrom onward.
func (h *SetNightWindowHandler) Handle(ctx context.Context, cmd SetNightWindowCommand) error {
	start, err := NewLocalTime(cmd.NightWindowStartHour, cmd.NightWindowStartMinute)
	if err != nil {
		return err
	}

	end, err := NewLocalTime(cmd.NightWindowEndHour, cmd.NightWindowEndMinute)
	if err != nil {
		return err
	}

	if cmd.EffectiveFrom.IsZero() {
		return ErrZeroNightWindowEffectiveFrom
	}

	return h.transactor.WithTransaction(ctx, func(ctx context.Context) error {
		existing, err := h.windows.FindByBabyID(ctx, cmd.BabyID)
		if err != nil {
			return err
		}

		var (
			deleteIDs []NightWindowID
			previous  *NightWindow
		)

		for _, window := range existing {
			if !window.EffectiveFrom().Before(cmd.EffectiveFrom) {
				deleteIDs = append(deleteIDs, window.ID())
				continue
			}

			if previous == nil || previous.EffectiveFrom().Before(window.EffectiveFrom()) {
				w := window
				previous = &w
			}
		}

		if err := h.windows.DeleteByIDs(ctx, deleteIDs); err != nil {
			return err
		}

		if previous != nil {
			effectiveTo := cmd.EffectiveFrom.UTC()
			updatedPrevious, err := NewNightWindow(
				previous.ID(),
				previous.BabyID(),
				previous.Start(),
				previous.End(),
				previous.EffectiveFrom(),
				&effectiveTo,
			)
			if err != nil {
				return err
			}
			if err := h.windows.Save(ctx, updatedPrevious); err != nil {
				return err
			}
		}

		newWindow, err := NewNightWindow(
			NightWindowID(uuid.New().String()),
			cmd.BabyID,
			start,
			end,
			cmd.EffectiveFrom.UTC(),
			nil,
		)
		if err != nil {
			return err
		}

		return h.windows.Save(ctx, newWindow)
	})
}
