package sleep

import (
	"context"
)

// CreateSleepProfileCommand holds the inputs for creating a sleep profile for a baby.
type CreateSleepProfileCommand struct {
	BabyID                 BabyID
	Timezone               string
	NightWindowStartHour   int
	NightWindowStartMinute int
	NightWindowEndHour     int
	NightWindowEndMinute   int
}

// CreateSleepProfileHandler executes the CreateSleepProfile use case.
type CreateSleepProfileHandler struct {
	profiles SleepProfileRepository
}

// NewCreateSleepProfileHandler returns a CreateSleepProfileHandler backed by the given repository.
func NewCreateSleepProfileHandler(profiles SleepProfileRepository) *CreateSleepProfileHandler {
	return &CreateSleepProfileHandler{profiles: profiles}
}

// Handle validates the command, builds the sleep profile, and persists it.
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

	return h.profiles.Save(ctx, profile)
}
