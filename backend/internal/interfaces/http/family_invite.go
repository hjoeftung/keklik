package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
)

type createFamilyInviteLinkResponse struct {
	Token     string `json:"token"`
	InviteURL string `json:"invite_url"`
	ExpiresAt string `json:"expires_at"`
}

type joinFamilyByInviteLinkRequest struct {
	Token      string `json:"token"`
	MemberName string `json:"member_name"`
}

type joinFamilyByInviteLinkResponse struct {
	FamilyID string `json:"family_id"`
	MemberID string `json:"member_id"`
}

// createFamilyInviteLinkHandler generates a one-time invite link for the caller's family.
//
// @Summary   Create family invite link
// @Tags      families
// @Produce   json
// @Security  BearerAuth
// @Success   201  {object}  createFamilyInviteLinkResponse
// @Failure   401  {object}  errorResponse
// @Failure   403  {object}  errorResponse
// @Router    /families/invite-links [post]
func createFamilyInviteLinkHandler(w http.ResponseWriter, r *http.Request, h *family.CreateFamilyInviteLinkHandler) {
	accountID, ok := auth.AccountIDFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	result, err := h.Handle(r.Context(), family.CreateFamilyInviteLinkCommand{
		CreatorAccountID: accountID,
	})
	if err != nil {
		writeError(w, mapFamilyError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createFamilyInviteLinkResponse{
		Token:     string(result.InviteLink.Token),
		InviteURL: result.InviteURL,
		ExpiresAt: result.InviteLink.ExpiresAt.UTC().Format(http.TimeFormat),
	})
}

// revokeInviteLinkHandler deletes an invite link so the token can no longer be used to join.
//
// @Summary   Revoke family invite link
// @Tags      families
// @Produce   json
// @Security  BearerAuth
// @Param     token  path      string  true  "Invite token"
// @Success   204
// @Failure   401  {object}  errorResponse
// @Failure   403  {object}  errorResponse
// @Failure   404  {object}  errorResponse
// @Router    /families/invite-links/{token} [delete]
func revokeInviteLinkHandler(w http.ResponseWriter, r *http.Request, h *family.RevokeInviteLinkHandler) {
	accountID, ok := auth.AccountIDFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	token := r.PathValue("token")
	if token == "" {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "token is required"))
		return
	}

	if err := h.Handle(r.Context(), family.RevokeInviteLinkCommand{
		AccountID: accountID,
		Token:     family.InviteToken(token),
	}); err != nil {
		writeError(w, mapFamilyError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// joinFamilyByInviteLinkHandler adds the authenticated user to a family using an invite token.
//
// @Summary   Join family by invite link
// @Tags      families
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      joinFamilyByInviteLinkRequest   true  "Invite token and member name"
// @Success   201   {object}  joinFamilyByInviteLinkResponse
// @Failure   400   {object}  errorResponse
// @Failure   401   {object}  errorResponse
// @Router    /families/join-by-invite-link [post]
func joinFamilyByInviteLinkHandler(w http.ResponseWriter, r *http.Request, h *family.JoinFamilyByInviteLinkHandler, accounts auth.AccountRepository) {
	accountID, ok := auth.AccountIDFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	var req joinFamilyByInviteLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		return
	}

	account, err := accounts.FindByID(r.Context(), accountID)
	if err != nil {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "account not found"))
		return
	}

	result, err := h.Handle(r.Context(), family.JoinFamilyByInviteLinkCommand{
		InviteToken: family.InviteToken(req.Token),
		AccountID:   accountID,
		Email:       account.Email,
		MemberName:  req.MemberName,
	})
	if err != nil {
		writeError(w, mapFamilyError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(joinFamilyByInviteLinkResponse{
		FamilyID: string(result.FamilyID),
		MemberID: string(result.MemberID),
	})
}
