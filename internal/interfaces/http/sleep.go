package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// sleepContextResolver resolves the baby and member IDs from the caller's Google subject ID
// in a single database round-trip.
type sleepContextResolver interface {
	ResolveSleepContext(ctx context.Context, googleSubjectID string) (sleep.BabyID, sleep.FamilyMemberID, error)
}

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

type startSleepRequest struct {
	StartedAt time.Time `json:"started_at"`
}

type startSleepResponse struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
}

func startSleepHandler(
	w http.ResponseWriter,
	r *http.Request,
	resolver sleepContextResolver,
	h *sleep.StartSleepHandler,
) {
	account, ok := auth.AccountFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	babyID, memberID, err := resolver.ResolveSleepContext(r.Context(), account.GoogleSubjectID)
	if err != nil {
		writeError(w, mapStartSleepError(err))
		return
	}

	var req startSleepRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.Handle(r.Context(), sleep.StartSleepCommand{
		BabyID:            babyID,
		CreatedByMemberID: memberID,
		StartedAt:         req.StartedAt,
	})
	if err != nil {
		writeError(w, mapStartSleepError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(startSleepResponse{
		ID:        string(result.ID),
		StartedAt: result.StartedAt,
	})
}

func mapStartSleepError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrActiveSleepSessionExists):
		return apperror.New(apperror.CodeActiveSleepExists, err.Error())
	case errors.Is(err, sleep.ErrEmptyBabyID),
		errors.Is(err, sleep.ErrEmptyFamilyMemberID),
		errors.Is(err, sleep.ErrZeroSleepSessionStart):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.New(apperror.CodeInternalError, "unexpected error")
	}
}

type stopSleepRequest struct {
	StoppedAt time.Time `json:"stopped_at"`
}

type stopSleepResponse struct {
	ID             string    `json:"id"`
	StartedAt      time.Time `json:"started_at"`
	StoppedAt      time.Time `json:"stopped_at"`
	Classification string    `json:"classification"`
}

func stopSleepHandler(
	w http.ResponseWriter,
	r *http.Request,
	resolver sleepContextResolver,
	h *sleep.StopSleepHandler,
) {
	account, ok := auth.AccountFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "authorization required"))
		return
	}

	babyID, _, err := resolver.ResolveSleepContext(r.Context(), account.GoogleSubjectID)
	if err != nil {
		writeError(w, mapStopSleepError(err))
		return
	}

	var req stopSleepRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.Handle(r.Context(), sleep.StopSleepCommand{
		BabyID:    babyID,
		StoppedAt: req.StoppedAt,
	})
	if err != nil {
		writeError(w, mapStopSleepError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stopSleepResponse{
		ID:             string(result.ID),
		StartedAt:      result.StartedAt,
		StoppedAt:      result.StoppedAt,
		Classification: string(result.Classification),
	})
}

func mapStopSleepError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrSleepSessionAlreadyStopped),
		errors.Is(err, sleep.ErrInvalidSleepSessionStop):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.New(apperror.CodeInternalError, "unexpected error")
	}
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
