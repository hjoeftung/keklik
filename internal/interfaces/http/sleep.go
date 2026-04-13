package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/sleep"
)

type nightWindowRequest struct {
	StartHour   int `json:"start_hour"`
	StartMinute int `json:"start_minute"`
	EndHour     int `json:"end_hour"`
	EndMinute   int `json:"end_minute"`
}

type createSleepProfileRequest struct {
	BabyID      string             `json:"baby_id"`
	Timezone    string             `json:"timezone"`
	NightWindow nightWindowRequest `json:"night_window"`
}

func createSleepProfileHandler(w http.ResponseWriter, r *http.Request, h *sleep.CreateSleepProfileHandler) {
	var req createSleepProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		return
	}

	err := h.Handle(r.Context(), sleep.CreateSleepProfileCommand{
		BabyID:                 sleep.BabyID(req.BabyID),
		Timezone:               req.Timezone,
		NightWindowStartHour:   req.NightWindow.StartHour,
		NightWindowStartMinute: req.NightWindow.StartMinute,
		NightWindowEndHour:     req.NightWindow.EndHour,
		NightWindowEndMinute:   req.NightWindow.EndMinute,
	})
	if err != nil {
		writeError(w, mapSleepProfileError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapSleepProfileError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrInvalidTimezone):
		return apperror.New(apperror.CodeInvalidTimezone, err.Error())
	case errors.Is(err, sleep.ErrInvalidNightWindow),
		errors.Is(err, sleep.ErrInvalidLocalTime),
		errors.Is(err, sleep.ErrEmptyBabyID):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.New(apperror.CodeInvalidArgument, "unexpected error")
	}
}
