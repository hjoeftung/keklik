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

func createFamilyInviteLinkHandler(w http.ResponseWriter, r *http.Request, h *family.CreateFamilyInviteLinkHandler) {
	account, ok := auth.AccountFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	result, err := h.Handle(r.Context(), family.CreateFamilyInviteLinkCommand{
		CreatorGoogleSubjectID: account.GoogleSubjectID,
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

func joinFamilyByInviteLinkHandler(w http.ResponseWriter, r *http.Request, h *family.JoinFamilyByInviteLinkHandler) {
	account, ok := auth.AccountFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	var req joinFamilyByInviteLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		return
	}

	result, err := h.Handle(r.Context(), family.JoinFamilyByInviteLinkCommand{
		InviteToken:     family.InviteToken(req.Token),
		GoogleSubjectID: account.GoogleSubjectID,
		Email:           account.Email,
		MemberName:      req.MemberName,
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
