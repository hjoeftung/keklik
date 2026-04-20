package family

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
)

// CreateFamilyInviteLinkCommand identifies the authenticated member creating the invite.
type CreateFamilyInviteLinkCommand struct {
	CreatorAccountID auth.AccountID
}

// CreateFamilyInviteLinkResult returns the persisted invite and its shareable URL.
type CreateFamilyInviteLinkResult struct {
	InviteLink InviteLink
	InviteURL  string
}

// CreateFamilyInviteLinkHandler issues family invite links for existing family members.
type CreateFamilyInviteLinkHandler struct {
	families FamilyRepository
	members  FamilyMemberRepository
	baseURL  string
	ttl      time.Duration
	now      func() time.Time
}

// NewCreateFamilyInviteLinkHandler returns a handler with runtime-configured invite settings.
func NewCreateFamilyInviteLinkHandler(
	families FamilyRepository,
	members FamilyMemberRepository,
	baseURL string,
	ttl time.Duration,
) *CreateFamilyInviteLinkHandler {
	return &CreateFamilyInviteLinkHandler{
		families: families,
		members:  members,
		baseURL:  strings.TrimRight(baseURL, "/"),
		ttl:      ttl,
		now:      time.Now,
	}
}

// Handle creates a reusable invite link that remains valid until expiry.
// The token is multi-use by design for the MVP; duplicate membership is still rejected per account.
func (h *CreateFamilyInviteLinkHandler) Handle(
	ctx context.Context,
	cmd CreateFamilyInviteLinkCommand,
) (CreateFamilyInviteLinkResult, error) {
	member, err := h.members.FindByAccountID(ctx, cmd.CreatorAccountID)
	if err != nil {
		if isFamilyMemberNotFound(err) {
			return CreateFamilyInviteLinkResult{}, apperror.New(apperror.CodeForbidden, "account is not part of a family")
		}
		return CreateFamilyInviteLinkResult{}, err
	}

	f, err := h.families.FindByMemberID(ctx, member.ID)
	if err != nil {
		return CreateFamilyInviteLinkResult{}, err
	}

	token, err := generateInviteToken()
	if err != nil {
		return CreateFamilyInviteLinkResult{}, err
	}

	inviteLink := InviteLink{
		Token:             token,
		FamilyID:          f.ID(),
		CreatedByMemberID: member.ID,
		ExpiresAt:         h.now().Add(h.ttl),
	}
	if err := f.AddInviteLink(inviteLink); err != nil {
		return CreateFamilyInviteLinkResult{}, err
	}

	if err := h.families.Save(ctx, f); err != nil {
		return CreateFamilyInviteLinkResult{}, err
	}

	return CreateFamilyInviteLinkResult{
		InviteLink: inviteLink,
		InviteURL:  buildInviteURL(h.baseURL, inviteLink.Token),
	}, nil
}

// JoinFamilyByInviteLinkCommand contains the authenticated account and invite token.
type JoinFamilyByInviteLinkCommand struct {
	InviteToken InviteToken
	AccountID   auth.AccountID
	Email       string
	MemberName  string
}

// JoinFamilyByInviteLinkResult returns the linked membership identifiers.
type JoinFamilyByInviteLinkResult struct {
	FamilyID FamilyID
	MemberID FamilyMemberID
}

// JoinFamilyByInviteLinkHandler links an authenticated account to a family via invite token.
type JoinFamilyByInviteLinkHandler struct {
	families FamilyRepository
	members  FamilyMemberRepository
	now      func() time.Time
}

// NewJoinFamilyByInviteLinkHandler returns a handler backed by family repositories.
func NewJoinFamilyByInviteLinkHandler(
	families FamilyRepository,
	members FamilyMemberRepository,
) *JoinFamilyByInviteLinkHandler {
	return &JoinFamilyByInviteLinkHandler{
		families: families,
		members:  members,
		now:      time.Now,
	}
}

// Handle validates the invite token and creates exactly one family membership for the account.
func (h *JoinFamilyByInviteLinkHandler) Handle(
	ctx context.Context,
	cmd JoinFamilyByInviteLinkCommand,
) (JoinFamilyByInviteLinkResult, error) {
	f, err := h.families.FindByInviteToken(ctx, cmd.InviteToken)
	if err != nil {
		return JoinFamilyByInviteLinkResult{}, invalidInviteLinkError(err)
	}

	inviteLink, ok := findInviteLink(f, cmd.InviteToken)
	if !ok || inviteLink.IsExpired(h.now()) {
		return JoinFamilyByInviteLinkResult{}, apperror.New(apperror.CodeInvalidInviteLink, "invite link is invalid or expired")
	}

	existingMember, err := h.members.FindByAccountID(ctx, cmd.AccountID)
	switch {
	case err == nil:
		if existingMember.FamilyID == f.ID() {
			return JoinFamilyByInviteLinkResult{}, apperror.New(apperror.CodeConflict, "account already belongs to this family")
		}
		return JoinFamilyByInviteLinkResult{}, apperror.New(apperror.CodeConflict, "account already belongs to another family")
	case !isFamilyMemberNotFound(err):
		return JoinFamilyByInviteLinkResult{}, err
	}

	member := FamilyMember{
		ID:        FamilyMemberID(uuid.New().String()),
		FamilyID:  f.ID(),
		Name:      resolveMemberName(cmd.MemberName, cmd.Email),
		AccountID: cmd.AccountID,
	}
	if err := f.AddMember(member); err != nil {
		return JoinFamilyByInviteLinkResult{}, err
	}

	if err := h.families.Save(ctx, f); err != nil {
		return JoinFamilyByInviteLinkResult{}, err
	}

	return JoinFamilyByInviteLinkResult{
		FamilyID: f.ID(),
		MemberID: member.ID,
	}, nil
}

func generateInviteToken() (InviteToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate invite token: %w", err)
	}

	return InviteToken(base64.RawURLEncoding.EncodeToString(tokenBytes)), nil
}

func buildInviteURL(baseURL string, token InviteToken) string {
	return fmt.Sprintf("%s/join-family?token=%s", baseURL, url.QueryEscape(string(token)))
}

func findInviteLink(f Family, token InviteToken) (InviteLink, bool) {
	for _, inviteLink := range f.InviteLinks() {
		if inviteLink.Token == token {
			return inviteLink, true
		}
	}

	return InviteLink{}, false
}

func invalidInviteLinkError(err error) error {
	var appErr apperror.AppError
	if errors.As(err, &appErr) && appErr.Code == apperror.CodeNotFound {
		return apperror.New(apperror.CodeInvalidInviteLink, "invite link is invalid or expired")
	}

	return err
}

func isFamilyMemberNotFound(err error) bool {
	var appErr apperror.AppError
	return errors.As(err, &appErr) && appErr.Code == apperror.CodeNotFound
}

func resolveMemberName(memberName, email string) string {
	if trimmed := strings.TrimSpace(memberName); trimmed != "" {
		return trimmed
	}

	address := strings.TrimSpace(email)
	if localPart, _, found := strings.Cut(address, "@"); found && strings.TrimSpace(localPart) != "" {
		return localPart
	}

	return address
}
