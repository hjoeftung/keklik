package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/sleep"
)

type babyContextKey struct{}

type babyCtx struct {
	BabyID   sleep.BabyID
	MemberID sleep.FamilyMemberID
}

func withBabyContext(ctx context.Context, babyID sleep.BabyID, memberID sleep.FamilyMemberID) context.Context {
	return context.WithValue(ctx, babyContextKey{}, babyCtx{BabyID: babyID, MemberID: memberID})
}

func babyContextFromContext(ctx context.Context) (babyCtx, bool) {
	v, ok := ctx.Value(babyContextKey{}).(babyCtx)
	return v, ok
}

// sleepSessionResponse is the JSON shape for a single sleep session.
type sleepSessionResponse struct {
	ID              string   `json:"id"`
	BabyID          string   `json:"baby_id"`
	StartedAt       string   `json:"started_at"`
	StoppedAt       *string  `json:"stopped_at,omitempty"`
	Classification  string   `json:"classification,omitempty"`
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
}

func toSleepSessionResponse(s sleep.SleepSession, classification sleep.SleepClassification) sleepSessionResponse {
	resp := sleepSessionResponse{
		ID:             string(s.ID()),
		BabyID:         string(s.BabyID()),
		StartedAt:      s.StartedAt().UTC().Format(time.RFC3339),
		Classification: string(classification),
	}
	if t, ok := s.StoppedAt(); ok {
		ts := t.UTC().Format(time.RFC3339)
		resp.StoppedAt = &ts
	}
	if d, ok := s.Duration(); ok {
		secs := d.Seconds()
		resp.DurationSeconds = &secs
	}
	return resp
}

// getSleepHistoryHandler returns the sleep session history for the caller's baby.
//
// @Summary   Get sleep history
// @Tags      sleep
// @Produce   json
// @Security  BearerAuth
// @Param     baby_id  path      string  true   "Baby ID"
// @Param     period   query     string  false  "History window: today, or Nd where 1 ≤ N ≤ 120 (default: 7d)"
// @Param     timezone query     string  true   "IANA timezone, e.g. America/New_York"
// @Success   200      {array}   sleepSessionResponse
// @Failure   400      {object}  errorResponse
// @Failure   401      {object}  errorResponse
// @Failure   403      {object}  errorResponse
// @Failure   404      {object}  errorResponse
// @Router    /babies/{baby_id}/sleep-sessions [get]
func getSleepHistoryHandler(w http.ResponseWriter, r *http.Request, h *sleep.GetSleepHistoryHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}
	timezone := r.URL.Query().Get("timezone")
	if timezone == "" {
		writeError(w, r, apperror.New(apperror.CodeInvalidTimezone, sleep.ErrInvalidTimezone.Error()))
		return
	}
	sessions, err := h.Handle(r.Context(), sleep.GetSleepHistoryQuery{
		BabyID:   bc.BabyID,
		Timezone: timezone,
		Period:   period,
	})
	if err != nil {
		writeError(w, r, mapSleepHistoryError(err))
		return
	}

	resp := make([]sleepSessionResponse, len(sessions))
	for i, s := range sessions {
		resp[i] = toSleepSessionResponse(s.Session, s.Classification)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func mapSleepHistoryError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrInvalidSleepHistoryPeriod):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	case errors.Is(err, sleep.ErrInvalidTimezone):
		return apperror.New(apperror.CodeInvalidTimezone, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
	}
}

type setNightWindowRequest struct {
	StartHour     int       `json:"start_hour"`
	StartMinute   int       `json:"start_minute"`
	EndHour       int       `json:"end_hour"`
	EndMinute     int       `json:"end_minute"`
	EffectiveFrom time.Time `json:"effective_from"`
}

// setNightWindowHandler creates or updates the baby's night-window timeline.
//
// @Summary   Set night window
// @Tags      sleep
// @Accept    json
// @Security  BearerAuth
// @Param     baby_id  path  string                true  "Baby ID"
// @Param     body     body  setNightWindowRequest true  "Night window configuration"
// @Success   204
// @Failure   400  {object}  errorResponse
// @Failure   401  {object}  errorResponse
// @Failure   403  {object}  errorResponse
// @Failure   404  {object}  errorResponse
// @Router    /babies/{baby_id}/night-windows [post]
func setNightWindowHandler(w http.ResponseWriter, r *http.Request, h *sleep.SetNightWindowHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req setNightWindowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var timeErr *time.ParseError
		if errors.As(err, &timeErr) {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "effective_from must be a valid RFC3339 timestamp (e.g. 2006-01-02T15:04:05Z)"))
		} else {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		}
		return
	}

	err := h.Handle(r.Context(), sleep.SetNightWindowCommand{
		BabyID:                 bc.BabyID,
		NightWindowStartHour:   req.StartHour,
		NightWindowStartMinute: req.StartMinute,
		NightWindowEndHour:     req.EndHour,
		NightWindowEndMinute:   req.EndMinute,
		EffectiveFrom:          req.EffectiveFrom,
	})
	if err != nil {
		writeError(w, r, mapNightWindowError(err))
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

// startSleepHandler starts a new sleep session for the caller's baby.
//
// @Summary   Start sleep session
// @Tags      sleep
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     baby_id  path      string             true   "Baby ID"
// @Param     body     body      startSleepRequest  false  "Optional explicit start time (defaults to now)"
// @Success   201      {object}  startSleepResponse
// @Failure   400      {object}  errorResponse
// @Failure   401      {object}  errorResponse
// @Failure   403      {object}  errorResponse
// @Failure   404      {object}  errorResponse
// @Failure   409      {object}  errorResponse  "Active sleep session already exists"
// @Router    /babies/{baby_id}/sleep-sessions/active [post]
func startSleepHandler(w http.ResponseWriter, r *http.Request, h *sleep.StartSleepHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req startSleepRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.Handle(r.Context(), sleep.StartSleepCommand{
		BabyID:            bc.BabyID,
		CreatedByMemberID: bc.MemberID,
		StartedAt:         req.StartedAt,
	})
	if err != nil {
		writeError(w, r, mapStartSleepError(err))
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
		return apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
	}
}

type stopSleepRequest struct {
	StoppedAt time.Time `json:"stopped_at"`
}

type stopSleepResponse struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
}

type editSleepSessionRequest struct {
	StartedAt *time.Time `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at"`
}

// stopSleepHandler stops the active sleep session for the caller's baby.
//
// @Summary   Stop active sleep session
// @Tags      sleep
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     baby_id  path      string            true   "Baby ID"
// @Param     body     body      stopSleepRequest  false  "Optional explicit stop time (defaults to now)"
// @Success   200      {object}  stopSleepResponse
// @Failure   400      {object}  errorResponse
// @Failure   401      {object}  errorResponse
// @Failure   403      {object}  errorResponse
// @Failure   404      {object}  errorResponse
// @Router    /babies/{baby_id}/sleep-sessions/active [delete]
func stopSleepHandler(w http.ResponseWriter, r *http.Request, h *sleep.StopSleepHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req stopSleepRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.Handle(r.Context(), sleep.StopSleepCommand{
		BabyID:    bc.BabyID,
		StoppedAt: req.StoppedAt,
	})
	if err != nil {
		writeError(w, r, mapStopSleepError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stopSleepResponse{
		ID:        string(result.ID),
		StartedAt: result.StartedAt,
		StoppedAt: result.StoppedAt,
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
		return apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
	}
}

// editSleepSessionHandler updates the start or stop time of an existing sleep session.
//
// @Summary   Edit sleep session
// @Tags      sleep
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     baby_id  path      string                   true  "Baby ID"
// @Param     id       path      string                   true  "Sleep session UUID"
// @Param     body     body      editSleepSessionRequest  true  "Fields to update (at least one required)"
// @Success   200      {object}  sleepSessionResponse
// @Failure   400      {object}  errorResponse
// @Failure   401      {object}  errorResponse
// @Failure   403      {object}  errorResponse
// @Failure   404      {object}  errorResponse
// @Router    /babies/{baby_id}/sleep-sessions/{id} [patch]
func editSleepSessionHandler(w http.ResponseWriter, r *http.Request, h *sleep.EditSleepSessionHandler) {
	if _, ok := babyContextFromContext(r.Context()); !ok {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req editSleepSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		return
	}

	session, err := h.Handle(r.Context(), sleep.EditSleepSessionCommand{
		SessionID: sleep.SleepSessionID(r.PathValue("id")),
		StartedAt: req.StartedAt,
		StoppedAt: req.StoppedAt,
	})
	if err != nil {
		writeError(w, r, mapEditSleepSessionError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toSleepSessionResponse(session, sleep.SleepClassificationUnknown))
}

// deleteSleepSessionHandler permanently removes a sleep session.
//
// @Summary   Delete sleep session
// @Tags      sleep
// @Security  BearerAuth
// @Param     baby_id  path  string  true  "Baby ID"
// @Param     id       path  string  true  "Sleep session UUID"
// @Success   204
// @Failure   401  {object}  errorResponse
// @Failure   403  {object}  errorResponse
// @Failure   404  {object}  errorResponse
// @Router    /babies/{baby_id}/sleep-sessions/{id} [delete]
func deleteSleepSessionHandler(w http.ResponseWriter, r *http.Request, h *sleep.DeleteSleepSessionHandler) {
	if _, ok := babyContextFromContext(r.Context()); !ok {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	if err := h.Handle(r.Context(), sleep.DeleteSleepSessionCommand{
		SessionID: sleep.SleepSessionID(r.PathValue("id")),
	}); err != nil {
		writeError(w, r, mapDeleteSleepSessionError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapEditSleepSessionError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrMissingSleepSessionEdit),
		errors.Is(err, sleep.ErrZeroSleepSessionStart),
		errors.Is(err, sleep.ErrInvalidSleepSessionStop):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	case errors.Is(err, sleep.ErrActiveSleepSessionExists):
		return apperror.New(apperror.CodeActiveSleepExists, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
	}
}

func mapDeleteSleepSessionError(err error) apperror.AppError {
	var appErr apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
}

type logPastSleepRequest struct {
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
}

type logPastSleepResponse struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
}

// logPastSleepHandler creates a completed sleep session from explicit start and end times.
//
// @Summary   Log past sleep session
// @Tags      sleep
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     baby_id  path      string               true  "Baby ID"
// @Param     body     body      logPastSleepRequest  true  "Start and end times of the completed session"
// @Success   201      {object}  logPastSleepResponse
// @Failure   400      {object}  errorResponse
// @Failure   401      {object}  errorResponse
// @Failure   403      {object}  errorResponse
// @Failure   409      {object}  errorResponse  "Session overlaps an existing session"
// @Router    /babies/{baby_id}/sleep-sessions [post]
func logPastSleepHandler(w http.ResponseWriter, r *http.Request, h *sleep.LogPastSleepHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, r, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req logPastSleepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var timeErr *time.ParseError
		if errors.As(err, &timeErr) {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "started_at and stopped_at must be valid RFC3339 timestamps"))
		} else {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		}
		return
	}

	result, err := h.Handle(r.Context(), sleep.LogPastSleepCommand{
		BabyID:            bc.BabyID,
		CreatedByMemberID: bc.MemberID,
		StartedAt:         req.StartedAt,
		StoppedAt:         req.StoppedAt,
	})
	if err != nil {
		writeError(w, r, mapLogPastSleepError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(logPastSleepResponse{
		ID:        string(result.ID),
		StartedAt: result.StartedAt,
		StoppedAt: result.StoppedAt,
	})
}

func mapLogPastSleepError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrZeroSleepSessionStart),
		errors.Is(err, sleep.ErrInvalidSleepSessionStop),
		errors.Is(err, sleep.ErrEmptyBabyID),
		errors.Is(err, sleep.ErrEmptyFamilyMemberID):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	case errors.Is(err, sleep.ErrSleepSessionOverlap):
		return apperror.New(apperror.CodeConflict, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
	}
}

func mapNightWindowError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrInvalidNightWindow),
		errors.Is(err, sleep.ErrInvalidLocalTime),
		errors.Is(err, sleep.ErrEmptyBabyID),
		errors.Is(err, sleep.ErrZeroNightWindowEffectiveFrom):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.Wrap(apperror.CodeInternalError, "unexpected error", err)
	}
}
