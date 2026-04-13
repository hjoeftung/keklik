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
	FamilyName  string `json:"family_name"`
	BabyName    string `json:"baby_name"`
	CreatorName string `json:"creator_name"`
}

type createFamilyResponse struct {
	FamilyID string `json:"family_id"`
	MemberID string `json:"member_id"`
	BabyID   string `json:"baby_id"`
}

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
		FamilyName:             req.FamilyName,
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
	case errors.Is(err, family.ErrInvalidFamilyName),
		errors.Is(err, family.ErrInvalidBabyName),
		errors.Is(err, family.ErrInvalidFamilyMemberName),
		errors.Is(err, family.ErrEmptyGoogleSubjectID):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.New(apperror.CodeInvalidArgument, "unexpected error")
	}
}
