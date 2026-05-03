package family

import (
	"context"

	"github.com/hjoeftung/keklik/internal/auth"
)

// GetFamilyQuery identifies the authenticated account requesting family data.
type GetFamilyQuery struct {
	AccountID auth.AccountID
}

// GetFamilyMemberResult holds a member's public fields.
type GetFamilyMemberResult struct {
	ID        FamilyMemberID
	Name      string
	AccountID auth.AccountID
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
}

// NewGetFamilyHandler returns a GetFamilyHandler backed by the given repository.
func NewGetFamilyHandler(families FamilyRepository) *GetFamilyHandler {
	return &GetFamilyHandler{families: families}
}

// Handle resolves the authenticated account's family and returns its data.
// Returns CodeNotFound if the account has no family yet.
func (h *GetFamilyHandler) Handle(ctx context.Context, q GetFamilyQuery) (GetFamilyResult, error) {
	f, err := h.families.FindByAccountID(ctx, q.AccountID)
	if err != nil {
		return GetFamilyResult{}, err
	}

	memberResults := make([]GetFamilyMemberResult, 0, len(f.Members()))
	for _, m := range f.Members() {
		memberResults = append(memberResults, GetFamilyMemberResult{ID: m.ID, Name: m.Name, AccountID: m.AccountID})
	}

	babies := f.Babies()
	baby := GetFamilyBabyResult{ID: babies[0].ID, Name: babies[0].Name}

	return GetFamilyResult{
		FamilyID: f.ID(),
		Members:  memberResults,
		Baby:     baby,
	}, nil
}
