package family

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/hjoeftung/keklik/internal/auth"
)

// CreateFamilyCommand holds the inputs for creating a new family.
type CreateFamilyCommand struct {
	BabyName         string
	CreatorName      string
	CreatorAccountID auth.AccountID
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
	familyID := FamilyID(uuid.New().String())
	memberID := FamilyMemberID(uuid.New().String())
	babyID := BabyID(uuid.New().String())

	member := FamilyMember{
		ID:        memberID,
		FamilyID:  familyID,
		Name:      cmd.CreatorName,
		AccountID: cmd.CreatorAccountID,
	}

	baby := Baby{
		ID:       babyID,
		FamilyID: familyID,
		Name:     cmd.BabyName,
	}

	f, err := NewFamily(familyID, []FamilyMember{member}, []Baby{baby})
	if err != nil {
		return CreateFamilyResult{}, err
	}

	if err := h.families.Save(ctx, f); err != nil {
		return CreateFamilyResult{}, err
	}
	slog.InfoContext(ctx, "family_created", "account_id", string(cmd.CreatorAccountID), "family_id", string(familyID))

	return CreateFamilyResult{
		FamilyID: familyID,
		MemberID: memberID,
		BabyID:   babyID,
	}, nil
}
