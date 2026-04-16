package sleep

import (
	"context"
	"time"
)

// CreateSleepProfileCommand holds the inputs for creating a sleep profile for a baby.
type CreateSleepProfileCommand struct {
	BabyID                 BabyID
	Timezone               string
	NightWindowStartHour   int
	NightWindowStartMinute int
	NightWindowEndHour     int
	NightWindowEndMinute   int
	// EffectiveFrom, when set, triggers retroactive reclassification of all completed
	// sessions whose started_at falls within [EffectiveFrom, now). Must not be
	// earlier than 30 days ago.
	EffectiveFrom *time.Time
}

// CreateSleepProfileHandler executes the CreateSleepProfile use case.
type CreateSleepProfileHandler struct {
	profiles      SleepProfileRepository
	sessions      CompletedSleepSessionsSinceRepository
	sessionWriter SleepSessionRepository
	now           func() time.Time
}

// NewCreateSleepProfileHandler returns a CreateSleepProfileHandler backed by the given repositories.
func NewCreateSleepProfileHandler(
	profiles SleepProfileRepository,
	sessions CompletedSleepSessionsSinceRepository,
	sessionWriter SleepSessionRepository,
) *CreateSleepProfileHandler {
	return &CreateSleepProfileHandler{
		profiles:      profiles,
		sessions:      sessions,
		sessionWriter: sessionWriter,
		now:           time.Now,
	}
}

// Handle validates the command, builds the sleep profile, persists it, and optionally
// reclassifies all completed sessions since effective_from using the new profile settings.
func (h *CreateSleepProfileHandler) Handle(ctx context.Context, cmd CreateSleepProfileCommand) error {
	start, err := NewLocalTime(cmd.NightWindowStartHour, cmd.NightWindowStartMinute)
	if err != nil {
		return err
	}

	end, err := NewLocalTime(cmd.NightWindowEndHour, cmd.NightWindowEndMinute)
	if err != nil {
		return err
	}

	nightWindow, err := NewNightWindow(start, end)
	if err != nil {
		return err
	}

	profile, err := NewSleepProfile(cmd.BabyID, cmd.Timezone, nightWindow)
	if err != nil {
		return err
	}

	if cmd.EffectiveFrom != nil {
		limit := h.now().AddDate(0, -1, 0)
		if cmd.EffectiveFrom.Before(limit) {
			return ErrEffectiveFromTooOld
		}
	}

	if err := h.profiles.Save(ctx, profile); err != nil {
		return err
	}

	if cmd.EffectiveFrom == nil {
		return nil
	}

	toReclassify, err := h.sessions.FindCompletedByBabyIDSince(ctx, cmd.BabyID, *cmd.EffectiveFrom)
	if err != nil {
		return err
	}

	for _, session := range toReclassify {
		stoppedAt, ok := session.StoppedAt()
		if !ok {
			continue
		}

		classification, err := Classify(session, profile.Timezone(), profile.NightWindow())
		if err != nil {
			return err
		}

		nw := profile.NightWindow()
		rebuilt, err := NewCompletedSleepSession(
			session.ID(),
			session.BabyID(),
			session.CreatedByMemberID(),
			session.StartedAt(),
			stoppedAt,
			classification,
			&nw,
		)
		if err != nil {
			return err
		}

		if err := h.sessionWriter.Save(ctx, rebuilt); err != nil {
			return err
		}
	}

	return nil
}
