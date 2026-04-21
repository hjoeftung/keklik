package family

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
)

var (
	ErrInvalidFamilyMemberName        = errors.New("family member name must not be empty")
	ErrInvalidBabyName                = errors.New("baby name must not be empty")
	ErrEmptyAccountID                 = errors.New("account id must not be empty")
	ErrFamilyMustHaveAtLeastOneMember = errors.New("family must have at least one member")
	ErrFamilyMustHaveExactlyOneBaby   = errors.New("family must have exactly one baby in the mvp")
	ErrFamilyMemberFamilyMismatch     = errors.New("family member belongs to another family")
	ErrBabyFamilyMismatch             = errors.New("baby belongs to another family")
	ErrDuplicateFamilyMember          = errors.New("family member already belongs to family")
	ErrInviteLinkFamilyMismatch       = errors.New("invite link belongs to another family")
	ErrDuplicateInviteToken           = errors.New("invite token already exists")
	ErrMemberAlreadyHasFamily         = errors.New("account already belongs to a family")
	ErrInviteLinkCreatorNotMember     = errors.New("invite link creator must belong to family")
	ErrInvalidInviteToken             = errors.New("invite token must not be empty")
)

type FamilyID string

type FamilyMemberID string

type BabyID string

type InviteToken string

type Family struct {
	id          FamilyID
	members     []FamilyMember
	babies      []Baby
	inviteLinks []InviteLink
}

type FamilyMember struct {
	ID        FamilyMemberID
	FamilyID  FamilyID
	Name      string
	AccountID auth.AccountID
}

type Baby struct {
	ID       BabyID
	FamilyID FamilyID
	Name     string
}

type InviteLink struct {
	Token             InviteToken
	FamilyID          FamilyID
	ExpiresAt         time.Time
	CreatedByMemberID FamilyMemberID
}

type FamilyRepository interface {
	Save(ctx context.Context, family Family) error
	FindByID(ctx context.Context, id FamilyID) (Family, error)
	FindByMemberID(ctx context.Context, memberID FamilyMemberID) (Family, error)
	FindByAccountID(ctx context.Context, accountID auth.AccountID) (Family, error)
	FindByInviteToken(ctx context.Context, token InviteToken) (Family, error)
}

type FamilyMemberRepository interface {
	Save(ctx context.Context, member FamilyMember) error
	FindByID(ctx context.Context, id FamilyMemberID) (FamilyMember, error)
	FindByAccountID(ctx context.Context, accountID auth.AccountID) (FamilyMember, error)
}

func NewFamily(id FamilyID, members []FamilyMember, babies []Baby) (Family, error) {
	if len(members) == 0 {
		return Family{}, ErrFamilyMustHaveAtLeastOneMember
	}

	validatedMembers := make([]FamilyMember, 0, len(members))
	seenMembers := make(map[FamilyMemberID]struct{}, len(members))
	for _, member := range members {
		if strings.TrimSpace(member.Name) == "" {
			return Family{}, ErrInvalidFamilyMemberName
		}
		if strings.TrimSpace(string(member.AccountID)) == "" {
			return Family{}, ErrEmptyAccountID
		}
		if member.FamilyID != id {
			return Family{}, ErrFamilyMemberFamilyMismatch
		}
		if _, exists := seenMembers[member.ID]; exists {
			return Family{}, ErrDuplicateFamilyMember
		}

		seenMembers[member.ID] = struct{}{}
		validatedMembers = append(validatedMembers, member)
	}

	if len(babies) != 1 {
		return Family{}, ErrFamilyMustHaveExactlyOneBaby
	}

	validatedBabies := make([]Baby, 0, len(babies))
	for _, baby := range babies {
		if strings.TrimSpace(baby.Name) == "" {
			return Family{}, ErrInvalidBabyName
		}
		if baby.FamilyID != id {
			return Family{}, ErrBabyFamilyMismatch
		}

		validatedBabies = append(validatedBabies, baby)
	}

	return Family{
		id:          id,
		members:     validatedMembers,
		babies:      validatedBabies,
		inviteLinks: nil,
	}, nil
}

func (f Family) ID() FamilyID {
	return f.id
}

func (f Family) Members() []FamilyMember {
	return append([]FamilyMember(nil), f.members...)
}

func (f Family) Babies() []Baby {
	return append([]Baby(nil), f.babies...)
}

func (f Family) InviteLinks() []InviteLink {
	return append([]InviteLink(nil), f.inviteLinks...)
}

func (f Family) HasMember(memberID FamilyMemberID) bool {
	for _, member := range f.members {
		if member.ID == memberID {
			return true
		}
	}

	return false
}

func (f *Family) AddMember(member FamilyMember) error {
	if strings.TrimSpace(member.Name) == "" {
		return ErrInvalidFamilyMemberName
	}
	if strings.TrimSpace(string(member.AccountID)) == "" {
		return ErrEmptyAccountID
	}
	if member.FamilyID != f.id {
		return ErrFamilyMemberFamilyMismatch
	}
	if f.HasMember(member.ID) {
		return ErrDuplicateFamilyMember
	}

	f.members = append(f.members, member)

	return nil
}

func (f *Family) AddInviteLink(inviteLink InviteLink) error {
	if strings.TrimSpace(string(inviteLink.Token)) == "" {
		return ErrInvalidInviteToken
	}
	if inviteLink.FamilyID != f.id {
		return ErrInviteLinkFamilyMismatch
	}
	if !f.HasMember(inviteLink.CreatedByMemberID) {
		return ErrInviteLinkCreatorNotMember
	}

	for _, existing := range f.inviteLinks {
		if existing.Token == inviteLink.Token {
			return ErrDuplicateInviteToken
		}
	}

	f.inviteLinks = append(f.inviteLinks, inviteLink)

	return nil
}

func (i InviteLink) IsExpired(now time.Time) bool {
	return !i.ExpiresAt.After(now)
}
