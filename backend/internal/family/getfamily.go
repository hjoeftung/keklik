package family

import (
	"context"

	"github.com/hjoeftung/keklik/internal/apperror"
)

// GetFamilyQuery identifies the authenticated account requesting family data.
type GetFamilyQuery struct {
	GoogleSubjectID string
}

// GetFamilyMemberResult holds a member's public fields.
type GetFamilyMemberResult struct {
	ID   FamilyMemberID
	Name string
}

// GetFamilyBabyResult holds a baby's public fields.
type GetFamilyBabyResult struct {
	ID   BabyID
	Name string
}

// GetFamilyResult holds the family data returned to the caller.
type GetFamilyResult struct {
	FamilyID FamilyID
	Members  []GetFamilyMemberResult
	Baby     GetFamilyBabyResult
}

// GetFamilyHandler executes the GetFamily use case.
type GetFamilyHandler struct {
	families FamilyRepository
	members  FamilyMemberRepository
}

// NewGetFamilyHandler returns a GetFamilyHandler backed by the given repositories.
func NewGetFamilyHandler(families FamilyRepository, members FamilyMemberRepository) *GetFamilyHandler {
	return &GetFamilyHandler{families: families, members: members}
}

// Handle resolves the authenticated account's family and returns its data.
// Returns CodeNotFound if the account has no family yet.
func (h *GetFamilyHandler) Handle(ctx context.Context, q GetFamilyQuery) (GetFamilyResult, error) {
	member, err := h.members.FindByGoogleSubjectID(ctx, q.GoogleSubjectID)
	if err != nil {
		if isFamilyMemberNotFound(err) {
			return GetFamilyResult{}, apperror.New(apperror.CodeNotFound, "no family found for this account")
		}
		return GetFamilyResult{}, err
	}

	f, err := h.families.FindByMemberID(ctx, member.ID)
	if err != nil {
		return GetFamilyResult{}, err
	}

	memberResults := make([]GetFamilyMemberResult, 0, len(f.Members()))
	for _, m := range f.Members() {
		memberResults = append(memberResults, GetFamilyMemberResult{ID: m.ID, Name: m.Name})
	}

	babies := f.Babies()
	baby := GetFamilyBabyResult{ID: babies[0].ID, Name: babies[0].Name}

	return GetFamilyResult{
		FamilyID: f.ID(),
		Members:  memberResults,
		Baby:     baby,
	}, nil
}
