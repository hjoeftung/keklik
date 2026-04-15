package family

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

type inviteFamilyRepository struct {
	families map[FamilyID]Family
}

func newInviteFamilyRepository(families ...Family) *inviteFamilyRepository {
	repo := &inviteFamilyRepository{families: make(map[FamilyID]Family, len(families))}
	for _, f := range families {
		repo.families[f.ID()] = f
	}
	return repo
}

func (r *inviteFamilyRepository) Save(_ context.Context, f Family) error {
	r.families[f.ID()] = f
	return nil
}

func (r *inviteFamilyRepository) FindByID(_ context.Context, id FamilyID) (Family, error) {
	f, ok := r.families[id]
	if !ok {
		return Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	return f, nil
}

func (r *inviteFamilyRepository) FindByMemberID(_ context.Context, memberID FamilyMemberID) (Family, error) {
	for _, f := range r.families {
		for _, member := range f.Members() {
			if member.ID == memberID {
				return f, nil
			}
		}
	}
	return Family{}, apperror.New(apperror.CodeNotFound, "family not found")
}

func (r *inviteFamilyRepository) FindByInviteToken(_ context.Context, token InviteToken) (Family, error) {
	for _, f := range r.families {
		for _, inviteLink := range f.InviteLinks() {
			if inviteLink.Token == token {
				return f, nil
			}
		}
	}
	return Family{}, apperror.New(apperror.CodeNotFound, "family not found")
}

type inviteMemberRepository struct {
	families *inviteFamilyRepository
}

func (r *inviteMemberRepository) Save(_ context.Context, _ FamilyMember) error {
	return nil
}

func (r *inviteMemberRepository) FindByID(_ context.Context, id FamilyMemberID) (FamilyMember, error) {
	for _, f := range r.families.families {
		for _, member := range f.Members() {
			if member.ID == id {
				return member, nil
			}
		}
	}
	return FamilyMember{}, apperror.New(apperror.CodeNotFound, "family member not found")
}

func (r *inviteMemberRepository) FindByGoogleSubjectID(_ context.Context, googleSubjectID string) (FamilyMember, error) {
	for _, f := range r.families.families {
		for _, member := range f.Members() {
			if member.GoogleSubjectID == googleSubjectID {
				return member, nil
			}
		}
	}
	return FamilyMember{}, apperror.New(apperror.CodeNotFound, "family member not found")
}

func TestCreateFamilyInviteLinkPersistsInviteAndReturnsURL(t *testing.T) {
	t.Parallel()

	f := mustFamily(t)
	families := newInviteFamilyRepository(f)
	members := &inviteMemberRepository{families: families}
	handler := NewCreateFamilyInviteLinkHandler(families, members, "http://localhost:8080/", 24*time.Hour)
	handler.now = func() time.Time {
		return time.Date(2026, time.April, 15, 9, 0, 0, 0, time.UTC)
	}

	result, err := handler.Handle(context.Background(), CreateFamilyInviteLinkCommand{
		CreatorGoogleSubjectID: "google-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.InviteLink.Token == "" {
		t.Fatal("expected non-empty invite token")
	}
	if result.InviteURL == "" {
		t.Fatal("expected non-empty invite URL")
	}
	if want := "http://localhost:8080/join-family?token="; result.InviteURL[:len(want)] != want {
		t.Fatalf("expected invite URL to start with %q, got %q", want, result.InviteURL)
	}

	saved, err := families.FindByID(context.Background(), f.ID())
	if err != nil {
		t.Fatalf("find saved family: %v", err)
	}
	if len(saved.InviteLinks()) != 1 {
		t.Fatalf("expected 1 invite link, got %d", len(saved.InviteLinks()))
	}
	if saved.InviteLinks()[0].ExpiresAt != handler.now().Add(24*time.Hour) {
		t.Fatalf("unexpected invite expiry: %s", saved.InviteLinks()[0].ExpiresAt)
	}
}

func TestCreateFamilyInviteLinkRejectsAccountOutsideFamily(t *testing.T) {
	t.Parallel()

	families := newInviteFamilyRepository(mustFamily(t))
	members := &inviteMemberRepository{families: families}
	handler := NewCreateFamilyInviteLinkHandler(families, members, "http://localhost:8080", 24*time.Hour)

	_, err := handler.Handle(context.Background(), CreateFamilyInviteLinkCommand{
		CreatorGoogleSubjectID: "google-missing",
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeForbidden {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestJoinFamilyByInviteLinkAddsMemberExactlyOnce(t *testing.T) {
	t.Parallel()

	f := mustFamily(t)
	if err := f.AddInviteLink(InviteLink{
		Token:             InviteToken("invite-1"),
		FamilyID:          f.ID(),
		CreatedByMemberID: FamilyMemberID("member-1"),
		ExpiresAt:         time.Date(2026, time.April, 20, 8, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed invite link: %v", err)
	}

	families := newInviteFamilyRepository(f)
	members := &inviteMemberRepository{families: families}
	handler := NewJoinFamilyByInviteLinkHandler(families, members)
	handler.now = func() time.Time {
		return time.Date(2026, time.April, 15, 8, 0, 0, 0, time.UTC)
	}

	result, err := handler.Handle(context.Background(), JoinFamilyByInviteLinkCommand{
		InviteToken:     InviteToken("invite-1"),
		GoogleSubjectID: "google-2",
		Email:           "parent.two@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FamilyID != f.ID() {
		t.Fatalf("expected family %q, got %q", f.ID(), result.FamilyID)
	}
	if result.MemberID == "" {
		t.Fatal("expected non-empty member ID")
	}

	saved, err := families.FindByID(context.Background(), f.ID())
	if err != nil {
		t.Fatalf("find saved family: %v", err)
	}
	if len(saved.Members()) != 2 {
		t.Fatalf("expected 2 members after join, got %d", len(saved.Members()))
	}
	if saved.Members()[1].Name != "parent.two" {
		t.Fatalf("expected fallback member name from email, got %q", saved.Members()[1].Name)
	}
}

func TestJoinFamilyByInviteLinkRejectsExpiredInvite(t *testing.T) {
	t.Parallel()

	f := mustFamily(t)
	if err := f.AddInviteLink(InviteLink{
		Token:             InviteToken("invite-expired"),
		FamilyID:          f.ID(),
		CreatedByMemberID: FamilyMemberID("member-1"),
		ExpiresAt:         time.Date(2026, time.April, 14, 8, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed invite link: %v", err)
	}

	families := newInviteFamilyRepository(f)
	members := &inviteMemberRepository{families: families}
	handler := NewJoinFamilyByInviteLinkHandler(families, members)
	handler.now = func() time.Time {
		return time.Date(2026, time.April, 15, 8, 0, 0, 0, time.UTC)
	}

	_, err := handler.Handle(context.Background(), JoinFamilyByInviteLinkCommand{
		InviteToken:     InviteToken("invite-expired"),
		GoogleSubjectID: "google-2",
		Email:           "parent.two@example.com",
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInviteLink {
		t.Fatalf("expected invalid invite link error, got %v", err)
	}
}

func TestJoinFamilyByInviteLinkRejectsDuplicateMembership(t *testing.T) {
	t.Parallel()

	f := mustFamily(t)
	if err := f.AddInviteLink(InviteLink{
		Token:             InviteToken("invite-1"),
		FamilyID:          f.ID(),
		CreatedByMemberID: FamilyMemberID("member-1"),
		ExpiresAt:         time.Date(2026, time.April, 20, 8, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed invite link: %v", err)
	}

	families := newInviteFamilyRepository(f)
	members := &inviteMemberRepository{families: families}
	handler := NewJoinFamilyByInviteLinkHandler(families, members)

	_, err := handler.Handle(context.Background(), JoinFamilyByInviteLinkCommand{
		InviteToken:     InviteToken("invite-1"),
		GoogleSubjectID: "google-1",
		Email:           "parent.one@example.com",
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeConflict {
		t.Fatalf("expected conflict error, got %v", err)
	}
}
