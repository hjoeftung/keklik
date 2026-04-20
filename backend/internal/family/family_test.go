package family

import (
	"errors"
	"testing"
	"time"
)

func TestNewFamilyRequiresExactlyOneBaby(t *testing.T) {
	t.Parallel()

	familyID := FamilyID("family-1")
	member := FamilyMember{
		ID:        FamilyMemberID("member-1"),
		FamilyID:  familyID,
		Name:      "Parent One",
		AccountID: "google-1",
	}

	_, err := NewFamily(
		familyID,
		[]FamilyMember{member},
		nil,
	)
	if !errors.Is(err, ErrFamilyMustHaveExactlyOneBaby) {
		t.Fatalf("expected ErrFamilyMustHaveExactlyOneBaby, got %v", err)
	}

	_, err = NewFamily(
		familyID,
		[]FamilyMember{member},
		[]Baby{
			{ID: BabyID("baby-1"), FamilyID: familyID, Name: "Mika"},
			{ID: BabyID("baby-2"), FamilyID: familyID, Name: "Ada"},
		},
	)
	if !errors.Is(err, ErrFamilyMustHaveExactlyOneBaby) {
		t.Fatalf("expected ErrFamilyMustHaveExactlyOneBaby, got %v", err)
	}
}

func TestFamilyAddMemberRejectsDuplicateAccount(t *testing.T) {
	t.Parallel()

	aggregate := mustFamily(t)

	err := aggregate.AddMember(FamilyMember{
		ID:        FamilyMemberID("member-1"),
		FamilyID:  aggregate.ID(),
		Name:      "Parent One",
		AccountID: "google-2",
	})
	if !errors.Is(err, ErrDuplicateFamilyMember) {
		t.Fatalf("expected ErrDuplicateFamilyMember, got %v", err)
	}
}

func TestFamilyAddInviteLinkRejectsOtherFamilyOwnership(t *testing.T) {
	t.Parallel()

	aggregate := mustFamily(t)

	err := aggregate.AddInviteLink(InviteLink{
		Token:             InviteToken("invite-1"),
		FamilyID:          FamilyID("family-2"),
		ExpiresAt:         time.Date(2026, time.April, 20, 8, 0, 0, 0, time.UTC),
		CreatedByMemberID: FamilyMemberID("member-1"),
	})
	if !errors.Is(err, ErrInviteLinkFamilyMismatch) {
		t.Fatalf("expected ErrInviteLinkFamilyMismatch, got %v", err)
	}
}

func mustFamily(t *testing.T) Family {
	t.Helper()

	aggregate, err := NewFamily(
		FamilyID("family-1"),
		[]FamilyMember{{
			ID:        FamilyMemberID("member-1"),
			FamilyID:  FamilyID("family-1"),
			Name:      "Parent One",
			AccountID: "google-1",
		}},
		[]Baby{{ID: BabyID("baby-1"), FamilyID: FamilyID("family-1"), Name: "Mika"}},
	)
	if err != nil {
		t.Fatalf("NewFamily returned error: %v", err)
	}

	return aggregate
}
