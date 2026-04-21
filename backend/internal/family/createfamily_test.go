package family

import (
	"context"
	"errors"
	"testing"

	"github.com/hjoeftung/keklik/internal/auth"
)

// inMemoryFamilyRepository is a test double for FamilyRepository.
type inMemoryFamilyRepository struct {
	saved []Family
	err   error
}

func (r *inMemoryFamilyRepository) Save(_ context.Context, f Family) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, f)
	return nil
}

func (r *inMemoryFamilyRepository) FindByID(_ context.Context, _ FamilyID) (Family, error) {
	return Family{}, errors.New("not implemented")
}

func (r *inMemoryFamilyRepository) FindByMemberID(_ context.Context, _ FamilyMemberID) (Family, error) {
	return Family{}, errors.New("not implemented")
}

func (r *inMemoryFamilyRepository) FindByAccountID(_ context.Context, _ auth.AccountID) (Family, error) {
	return Family{}, errors.New("not implemented")
}

func (r *inMemoryFamilyRepository) FindByInviteToken(_ context.Context, _ InviteToken) (Family, error) {
	return Family{}, errors.New("not implemented")
}

func validCreateFamilyCommand() CreateFamilyCommand {
	return CreateFamilyCommand{
		BabyName:         "Emma",
		CreatorName:      "Alice",
		CreatorAccountID: "google-subject-123",
	}
}

func TestCreateFamilyPersistsFamilyMemberAndBaby(t *testing.T) {
	t.Parallel()

	repo := &inMemoryFamilyRepository{}
	h := NewCreateFamilyHandler(repo)

	result, err := h.Handle(context.Background(), validCreateFamilyCommand())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(result.FamilyID) == "" {
		t.Error("expected non-empty family ID")
	}
	if string(result.MemberID) == "" {
		t.Error("expected non-empty member ID")
	}
	if string(result.BabyID) == "" {
		t.Error("expected non-empty baby ID")
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved family, got %d", len(repo.saved))
	}

	saved := repo.saved[0]
	if len(saved.Members()) != 1 {
		t.Errorf("expected 1 member, got %d", len(saved.Members()))
	}
	if len(saved.Babies()) != 1 {
		t.Errorf("expected 1 baby, got %d", len(saved.Babies()))
	}
}

func TestCreateFamilyRejectsEmptyBabyName(t *testing.T) {
	t.Parallel()

	repo := &inMemoryFamilyRepository{}
	h := NewCreateFamilyHandler(repo)

	cmd := validCreateFamilyCommand()
	cmd.BabyName = ""

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidBabyName) {
		t.Errorf("expected ErrInvalidBabyName, got %v", err)
	}
}

func TestCreateFamilyIdsAreUnique(t *testing.T) {
	t.Parallel()

	repo := &inMemoryFamilyRepository{}
	h := NewCreateFamilyHandler(repo)

	r1, _ := h.Handle(context.Background(), validCreateFamilyCommand())
	r2, _ := h.Handle(context.Background(), validCreateFamilyCommand())

	if r1.FamilyID == r2.FamilyID {
		t.Error("expected unique family IDs across calls")
	}
}
