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
	account, ok := auth.AccountFromContext(r.Context())
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
		BabyName:               req.BabyName,
		CreatorName:            req.CreatorName,
		CreatorGoogleSubjectID: account.GoogleSubjectID,
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

func mapFamilyError(err error) apperror.AppError {
	switch {
	case errors.Is(err, family.ErrInvalidBabyName),
		errors.Is(err, family.ErrInvalidFamilyMemberName),
		errors.Is(err, family.ErrEmptyGoogleSubjectID),
		errors.Is(err, family.ErrInvalidInviteToken):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	case errors.Is(err, family.ErrInviteLinkCreatorNotMember):
		return apperror.New(apperror.CodeForbidden, err.Error())
	case errors.Is(err, family.ErrDuplicateFamilyMember),
		errors.Is(err, family.ErrDuplicateInviteToken):
		return apperror.New(apperror.CodeConflict, err.Error())
	case errors.Is(err, family.ErrInviteLinkFamilyMismatch):
		return apperror.New(apperror.CodeInvalidInviteLink, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.New(apperror.CodeInvalidArgument, "unexpected error")
	}
}
