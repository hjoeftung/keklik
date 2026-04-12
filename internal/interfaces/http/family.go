package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/family"
)

type nightWindowRequest struct {
	StartHour   int `json:"start_hour"`
	StartMinute int `json:"start_minute"`
	EndHour     int `json:"end_hour"`
	EndMinute   int `json:"end_minute"`
}

type createFamilyRequest struct {
	FamilyName             string             `json:"family_name"`
	BabyName               string             `json:"baby_name"`
	Timezone               string             `json:"timezone"`
	NightWindow            nightWindowRequest `json:"night_window"`
	CreatorName            string             `json:"creator_name"`
	CreatorGoogleSubjectID string             `json:"creator_google_subject_id"`
}

type createFamilyResponse struct {
	FamilyID string `json:"family_id"`
	MemberID string `json:"member_id"`
	BabyID   string `json:"baby_id"`
}

func createFamilyHandler(w http.ResponseWriter, r *http.Request, h *family.CreateFamilyHandler) {
	var req createFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		return
	}

	result, err := h.Handle(r.Context(), family.CreateFamilyCommand{
		FamilyName:             req.FamilyName,
		BabyName:               req.BabyName,
		Timezone:               req.Timezone,
		NightWindowStartHour:   req.NightWindow.StartHour,
		NightWindowStartMinute: req.NightWindow.StartMinute,
		NightWindowEndHour:     req.NightWindow.EndHour,
		NightWindowEndMinute:   req.NightWindow.EndMinute,
		CreatorName:            req.CreatorName,
		CreatorGoogleSubjectID: req.CreatorGoogleSubjectID,
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
	case errors.Is(err, family.ErrInvalidTimezone):
		return apperror.New(apperror.CodeInvalidTimezone, err.Error())
	case errors.Is(err, family.ErrInvalidNightWindow),
		errors.Is(err, family.ErrInvalidLocalTime):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
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
