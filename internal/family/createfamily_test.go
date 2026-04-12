package family

import (
	"context"
	"errors"
	"testing"
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

func (r *inMemoryFamilyRepository) FindByInviteToken(_ context.Context, _ InviteToken) (Family, error) {
	return Family{}, errors.New("not implemented")
}

func validCreateFamilyCommand() CreateFamilyCommand {
	return CreateFamilyCommand{
		FamilyName:             "Smith Family",
		BabyName:               "Emma",
		Timezone:               "Europe/Berlin",
		NightWindowStartHour:   20,
		NightWindowStartMinute: 30,
		NightWindowEndHour:     7,
		NightWindowEndMinute:   0,
		CreatorName:            "Alice",
		CreatorGoogleSubjectID: "google-subject-123",
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
	if saved.Name() != "Smith Family" {
		t.Errorf("expected family name %q, got %q", "Smith Family", saved.Name())
	}
	if len(saved.Members()) != 1 {
		t.Errorf("expected 1 member, got %d", len(saved.Members()))
	}
	if len(saved.Babies()) != 1 {
		t.Errorf("expected 1 baby, got %d", len(saved.Babies()))
	}
}

func TestCreateFamilyRejectsInvalidTimezone(t *testing.T) {
	t.Parallel()

	repo := &inMemoryFamilyRepository{}
	h := NewCreateFamilyHandler(repo)

	cmd := validCreateFamilyCommand()
	cmd.Timezone = "Not/ATimezone"

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Errorf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestCreateFamilyRejectsInvalidNightWindow(t *testing.T) {
	t.Parallel()

	repo := &inMemoryFamilyRepository{}
	h := NewCreateFamilyHandler(repo)

	cmd := validCreateFamilyCommand()
	// start == end is invalid
	cmd.NightWindowStartHour = 20
	cmd.NightWindowStartMinute = 0
	cmd.NightWindowEndHour = 20
	cmd.NightWindowEndMinute = 0

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidNightWindow) {
		t.Errorf("expected ErrInvalidNightWindow, got %v", err)
	}
}

func TestCreateFamilyRejectsEmptyFamilyName(t *testing.T) {
	t.Parallel()

	repo := &inMemoryFamilyRepository{}
	h := NewCreateFamilyHandler(repo)

	cmd := validCreateFamilyCommand()
	cmd.FamilyName = "  "

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidFamilyName) {
		t.Errorf("expected ErrInvalidFamilyName, got %v", err)
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

func TestCreateFamilyRejectsInvalidLocalTime(t *testing.T) {
	t.Parallel()

	repo := &inMemoryFamilyRepository{}
	h := NewCreateFamilyHandler(repo)

	cmd := validCreateFamilyCommand()
	cmd.NightWindowStartHour = 25 // out of range

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidLocalTime) {
		t.Errorf("expected ErrInvalidLocalTime, got %v", err)
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
