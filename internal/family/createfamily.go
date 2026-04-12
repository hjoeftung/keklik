package family

import (
	"context"

	"github.com/google/uuid"
)

// CreateFamilyCommand holds the inputs for creating a new family.
type CreateFamilyCommand struct {
	FamilyName             string
	BabyName               string
	Timezone               string
	NightWindowStartHour   int
	NightWindowStartMinute int
	NightWindowEndHour     int
	NightWindowEndMinute   int
	// CreatorName and CreatorGoogleSubjectID will be sourced from auth context once
	// TASK-007 is implemented. For now they are supplied explicitly by the caller.
	CreatorName            string
	CreatorGoogleSubjectID string
}

// CreateFamilyResult holds the identifiers returned after successful family creation.
type CreateFamilyResult struct {
	FamilyID FamilyID
	MemberID FamilyMemberID
	BabyID   BabyID
}

// CreateFamilyHandler executes the CreateFamily use case.
type CreateFamilyHandler struct {
	families FamilyRepository
}

// NewCreateFamilyHandler returns a CreateFamilyHandler backed by the given repository.
func NewCreateFamilyHandler(families FamilyRepository) *CreateFamilyHandler {
	return &CreateFamilyHandler{families: families}
}

// Handle validates the command, builds the family aggregate, and persists it atomically.
func (h *CreateFamilyHandler) Handle(ctx context.Context, cmd CreateFamilyCommand) (CreateFamilyResult, error) {
	start, err := NewLocalTime(cmd.NightWindowStartHour, cmd.NightWindowStartMinute)
	if err != nil {
		return CreateFamilyResult{}, err
	}

	end, err := NewLocalTime(cmd.NightWindowEndHour, cmd.NightWindowEndMinute)
	if err != nil {
		return CreateFamilyResult{}, err
	}

	nightWindow, err := NewNightWindow(start, end)
	if err != nil {
		return CreateFamilyResult{}, err
	}

	familyID := FamilyID(uuid.New().String())
	memberID := FamilyMemberID(uuid.New().String())
	babyID := BabyID(uuid.New().String())

	member := FamilyMember{
		ID:              memberID,
		FamilyID:        familyID,
		Name:            cmd.CreatorName,
		GoogleSubjectID: cmd.CreatorGoogleSubjectID,
	}

	baby := Baby{
		ID:       babyID,
		FamilyID: familyID,
		Name:     cmd.BabyName,
	}

	f, err := NewFamily(familyID, cmd.FamilyName, cmd.Timezone, nightWindow, []FamilyMember{member}, []Baby{baby})
	if err != nil {
		return CreateFamilyResult{}, err
	}

	if err := h.families.Save(ctx, f); err != nil {
		return CreateFamilyResult{}, err
	}

	return CreateFamilyResult{
		FamilyID: familyID,
		MemberID: memberID,
		BabyID:   babyID,
	}, nil
}
