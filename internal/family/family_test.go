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
		ID:              FamilyMemberID("member-1"),
		FamilyID:        familyID,
		Name:            "Parent One",
		GoogleSubjectID: "google-1",
	}
	nightWindow := mustNightWindow(t, 19, 30, 7, 0)

	_, err := NewFamily(
		familyID,
		"The Owls",
		"Europe/Oslo",
		nightWindow,
		[]FamilyMember{member},
		nil,
	)
	if !errors.Is(err, ErrFamilyMustHaveExactlyOneBaby) {
		t.Fatalf("expected ErrFamilyMustHaveExactlyOneBaby, got %v", err)
	}

	_, err = NewFamily(
		familyID,
		"The Owls",
		"Europe/Oslo",
		nightWindow,
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

func TestNewFamilyRejectsInvalidTimezone(t *testing.T) {
	t.Parallel()

	_, err := NewFamily(
		FamilyID("family-1"),
		"The Owls",
		"Mars/Phobos",
		mustNightWindow(t, 20, 0, 6, 0),
		[]FamilyMember{{
			ID:              FamilyMemberID("member-1"),
			FamilyID:        FamilyID("family-1"),
			Name:            "Parent One",
			GoogleSubjectID: "google-1",
		}},
		[]Baby{{ID: BabyID("baby-1"), FamilyID: FamilyID("family-1"), Name: "Mika"}},
	)
	if !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}

func TestNewNightWindowRejectsEqualBounds(t *testing.T) {
	t.Parallel()

	start, err := NewLocalTime(20, 0)
	if err != nil {
		t.Fatalf("NewLocalTime returned error: %v", err)
	}

	_, err = NewNightWindow(start, start)
	if !errors.Is(err, ErrInvalidNightWindow) {
		t.Fatalf("expected ErrInvalidNightWindow, got %v", err)
	}
}

func TestFamilyAddMemberRejectsDuplicateAccount(t *testing.T) {
	t.Parallel()

	aggregate := mustFamily(t)

	err := aggregate.AddMember(FamilyMember{
		ID:              FamilyMemberID("member-1"),
		FamilyID:        aggregate.ID(),
		Name:            "Parent One",
		GoogleSubjectID: "google-2",
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
		"The Owls",
		"Europe/Oslo",
		mustNightWindow(t, 19, 30, 7, 0),
		[]FamilyMember{{
			ID:              FamilyMemberID("member-1"),
			FamilyID:        FamilyID("family-1"),
			Name:            "Parent One",
			GoogleSubjectID: "google-1",
		}},
		[]Baby{{ID: BabyID("baby-1"), FamilyID: FamilyID("family-1"), Name: "Mika"}},
	)
	if err != nil {
		t.Fatalf("NewFamily returned error: %v", err)
	}

	return aggregate
}

func mustNightWindow(t *testing.T, startHour int, startMinute int, endHour int, endMinute int) NightWindow {
	t.Helper()

	start, err := NewLocalTime(startHour, startMinute)
	if err != nil {
		t.Fatalf("NewLocalTime returned error: %v", err)
	}

	end, err := NewLocalTime(endHour, endMinute)
	if err != nil {
		t.Fatalf("NewLocalTime returned error: %v", err)
	}

	nightWindow, err := NewNightWindow(start, end)
	if err != nil {
		t.Fatalf("NewNightWindow returned error: %v", err)
	}

	return nightWindow
}
