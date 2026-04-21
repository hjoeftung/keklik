package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
)

type createFamilyRequest struct {
	BabyName    string `json:"baby_name"`
	CreatorName string `json:"creator_name"`
}

type createFamilyResponse struct {
	FamilyID string `json:"family_id"`
	MemberID string `json:"member_id"`
	BabyID   string `json:"baby_id"`
}

// createFamilyHandler creates a new family and returns IDs for the family, first member, and baby.
//
// @Summary   Create family
// @Tags      families
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      createFamilyRequest   true  "Family creation payload"
// @Success   201   {object}  createFamilyResponse
// @Failure   400   {object}  errorResponse
// @Failure   401   {object}  errorResponse
// @Router    /families [post]
func createFamilyHandler(w http.ResponseWriter, r *http.Request, h *family.CreateFamilyHandler) {
	accountID, ok := auth.AccountIDFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	var req createFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		return
	}

	result, err := h.Handle(r.Context(), family.CreateFamilyCommand{
		BabyName:         req.BabyName,
		CreatorName:      req.CreatorName,
		CreatorAccountID: accountID,
	})
	if err != nil {
		writeError(w, mapFamilyError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createFamilyResponse{
		FamilyID: string(result.FamilyID),
		MemberID: string(result.MemberID),
		BabyID:   string(result.BabyID),
	})
}

type getFamilyMemberResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type getFamilyBabyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type getFamilyResponse struct {
	FamilyID string                    `json:"family_id"`
	Members  []getFamilyMemberResponse `json:"members"`
	Baby     getFamilyBabyResponse     `json:"baby"`
}

// getFamilyHandler returns the authenticated member's family with its members and baby.
//
// @Summary   Get family
// @Tags      families
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  getFamilyResponse
// @Failure   401  {object}  errorResponse
// @Failure   404  {object}  errorResponse
// @Router    /family [get]
func getFamilyHandler(w http.ResponseWriter, r *http.Request, h *family.GetFamilyHandler) {
	accountID, ok := auth.AccountIDFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	result, err := h.Handle(r.Context(), family.GetFamilyQuery{
		AccountID: accountID,
	})
	if err != nil {
		writeError(w, mapFamilyError(err))
		return
	}

	members := make([]getFamilyMemberResponse, 0, len(result.Members))
	for _, m := range result.Members {
		members = append(members, getFamilyMemberResponse{ID: string(m.ID), Name: m.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getFamilyResponse{
		FamilyID: string(result.FamilyID),
		Members:  members,
		Baby:     getFamilyBabyResponse{ID: string(result.Baby.ID), Name: result.Baby.Name},
	})
}

func mapFamilyError(err error) apperror.AppError {
	switch {
	case errors.Is(err, family.ErrInvalidBabyName),
		errors.Is(err, family.ErrInvalidFamilyMemberName),
		errors.Is(err, family.ErrEmptyAccountID),
		errors.Is(err, family.ErrInvalidInviteToken):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	case errors.Is(err, family.ErrInviteLinkCreatorNotMember):
		return apperror.New(apperror.CodeForbidden, err.Error())
	case errors.Is(err, family.ErrDuplicateFamilyMember),
		errors.Is(err, family.ErrDuplicateInviteToken),
		errors.Is(err, family.ErrMemberAlreadyHasFamily):
		return apperror.New(apperror.CodeConflict, err.Error())
	case errors.Is(err, family.ErrInviteLinkFamilyMismatch):
		return apperror.New(apperror.CodeInvalidInviteLink, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.New(apperror.CodeInternalError, "unexpected error")
	}
}
