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

func toSleepSessionResponse(s sleep.SleepSession) sleepSessionResponse {
	resp := sleepSessionResponse{
		ID:             string(s.ID()),
		BabyID:         string(s.BabyID()),
		StartedAt:      s.StartedAt().UTC().Format(time.RFC3339),
		Classification: string(s.Classification()),
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
// @Param     period   query     string  false  "History window: today, 7d (default), or 14d"
// @Success   200      {array}   sleepSessionResponse
// @Failure   400      {object}  errorResponse
// @Failure   401      {object}  errorResponse
// @Failure   403      {object}  errorResponse
// @Failure   404      {object}  errorResponse
// @Router    /babies/{baby_id}/sleep-sessions [get]
func getSleepHistoryHandler(w http.ResponseWriter, r *http.Request, h *sleep.GetSleepHistoryHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}
	switch period {
	case "today", "7d", "14d":
		// valid
	default:
		writeError(w, apperror.New(apperror.CodeInvalidArgument, sleep.ErrInvalidSleepHistoryPeriod.Error()))
		return
	}

	sessions, err := h.Handle(r.Context(), sleep.GetSleepHistoryQuery{
		BabyID: bc.BabyID,
		Period: period,
	})
	if err != nil {
		writeError(w, mapSleepHistoryError(err))
		return
	}

	resp := make([]sleepSessionResponse, len(sessions))
	for i, s := range sessions {
		resp[i] = toSleepSessionResponse(s)
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
		return apperror.New(apperror.CodeInternalError, "unexpected error")
	}
}

type nightWindowRequest struct {
	StartHour   int `json:"start_hour"`
	StartMinute int `json:"start_minute"`
	EndHour     int `json:"end_hour"`
	EndMinute   int `json:"end_minute"`
}

type createSleepProfileRequest struct {
	Timezone      string             `json:"timezone"`
	NightWindow   nightWindowRequest `json:"night_window"`
	EffectiveFrom *time.Time         `json:"effective_from,omitempty"`
}

// createSleepProfileHandler creates or updates the sleep profile (timezone + night window) for a baby.
//
// @Summary   Create sleep profile
// @Tags      sleep
// @Accept    json
// @Security  BearerAuth
// @Param     baby_id  path  string                     true  "Baby ID"
// @Param     body     body  createSleepProfileRequest  true  "Sleep profile configuration. Set effective_from to retroactively reclassify completed sessions whose started_at >= effective_from (max 30 days back)."
// @Success   204
// @Failure   400  {object}  errorResponse  "Invalid input or effective_from earlier than 30 days ago"
// @Failure   401  {object}  errorResponse
// @Failure   403  {object}  errorResponse
// @Failure   404  {object}  errorResponse
// @Router    /babies/{baby_id}/sleep-profiles [post]
func createSleepProfileHandler(w http.ResponseWriter, r *http.Request, h *sleep.CreateSleepProfileHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req createSleepProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var timeErr *time.ParseError
		if errors.As(err, &timeErr) {
			writeError(w, apperror.New(apperror.CodeInvalidArgument, "effective_from must be a valid RFC3339 timestamp (e.g. 2006-01-02T15:04:05Z)"))
		} else {
			writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		}
		return
	}

	err := h.Handle(r.Context(), sleep.CreateSleepProfileCommand{
		BabyID:                 bc.BabyID,
		Timezone:               req.Timezone,
		NightWindowStartHour:   req.NightWindow.StartHour,
		NightWindowStartMinute: req.NightWindow.StartMinute,
		NightWindowEndHour:     req.NightWindow.EndHour,
		NightWindowEndMinute:   req.NightWindow.EndMinute,
		EffectiveFrom:          req.EffectiveFrom,
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
		writeError(w, apperror.New(apperror.CodeInternalError, "baby context missing"))
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
		writeError(w, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req stopSleepRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.Handle(r.Context(), sleep.StopSleepCommand{
		BabyID:    bc.BabyID,
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
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	var req editSleepSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(apperror.CodeInvalidArgument, "invalid request body"))
		return
	}

	session, err := h.Handle(r.Context(), sleep.EditSleepSessionCommand{
		SessionID:      sleep.SleepSessionID(r.PathValue("id")),
		FamilyMemberID: bc.MemberID,
		StartedAt:      req.StartedAt,
		StoppedAt:      req.StoppedAt,
	})
	if err != nil {
		writeError(w, mapEditSleepSessionError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toSleepSessionResponse(session))
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
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	if err := h.Handle(r.Context(), sleep.DeleteSleepSessionCommand{
		SessionID:      sleep.SleepSessionID(r.PathValue("id")),
		FamilyMemberID: bc.MemberID,
	}); err != nil {
		writeError(w, mapDeleteSleepSessionError(err))
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
		return apperror.New(apperror.CodeInternalError, "unexpected error")
	}
}

func mapDeleteSleepSessionError(err error) apperror.AppError {
	var appErr apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.New(apperror.CodeInternalError, "unexpected error")
}

type activeSleepSessionResponse struct {
	ID              string  `json:"id"`
	StartedAt       string  `json:"started_at"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type dailySummaryResponse struct {
	TotalSleepSeconds  float64 `json:"total_sleep_seconds"`
	TotalActiveSeconds float64 `json:"total_active_seconds"`
}

type rollingAverageResponse struct {
	AvgDailySleepSeconds  float64 `json:"avg_daily_sleep_seconds"`
	AvgDailyActiveSeconds float64 `json:"avg_daily_active_seconds"`
}

type dashboardSummaryResponse struct {
	ActiveSession              *activeSleepSessionResponse `json:"active_session"`
	TimeSinceSleepStartSeconds *float64                    `json:"time_since_sleep_start_seconds"`
	TimeSinceAwakeningSeconds  *float64                    `json:"time_since_awakening_seconds"`
	Today                      dailySummaryResponse        `json:"today"`
	Rolling7d                  rollingAverageResponse      `json:"rolling_7d"`
	Rolling14d                 rollingAverageResponse      `json:"rolling_14d"`
}

// getDashboardSummaryHandler returns all dashboard metrics for a single screen load.
//
// @Summary   Get dashboard summary
// @Tags      sleep
// @Produce   json
// @Security  BearerAuth
// @Param     baby_id  path      string  true  "Baby ID"
// @Success   200      {object}  dashboardSummaryResponse
// @Failure   401      {object}  errorResponse
// @Failure   403      {object}  errorResponse
// @Failure   404      {object}  errorResponse
// @Router    /babies/{baby_id}/dashboard [get]
func getDashboardSummaryHandler(w http.ResponseWriter, r *http.Request, h *sleep.GetDashboardSummaryHandler) {
	bc, ok := babyContextFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeInternalError, "baby context missing"))
		return
	}

	summary, err := h.Handle(r.Context(), sleep.GetDashboardSummaryQuery{
		BabyID: bc.BabyID,
	})
	if err != nil {
		writeError(w, mapDashboardSummaryError(err))
		return
	}

	resp := dashboardSummaryResponse{
		Today: dailySummaryResponse{
			TotalSleepSeconds:  summary.Today.TotalSleep.Seconds(),
			TotalActiveSeconds: summary.Today.TotalActive.Seconds(),
		},
		Rolling7d: rollingAverageResponse{
			AvgDailySleepSeconds:  summary.Rolling7d.AvgDailySleep.Seconds(),
			AvgDailyActiveSeconds: summary.Rolling7d.AvgDailyActive.Seconds(),
		},
		Rolling14d: rollingAverageResponse{
			AvgDailySleepSeconds:  summary.Rolling14d.AvgDailySleep.Seconds(),
			AvgDailyActiveSeconds: summary.Rolling14d.AvgDailyActive.Seconds(),
		},
	}

	if summary.ActiveSession != nil {
		secs := summary.ActiveSession.Duration.Seconds()
		resp.ActiveSession = &activeSleepSessionResponse{
			ID:              string(summary.ActiveSession.ID),
			StartedAt:       summary.ActiveSession.StartedAt.UTC().Format(time.RFC3339),
			DurationSeconds: secs,
		}
	}

	if summary.TimeSinceSleepStart != nil {
		secs := summary.TimeSinceSleepStart.Seconds()
		resp.TimeSinceSleepStartSeconds = &secs
	}

	if summary.TimeSinceAwakening != nil {
		secs := summary.TimeSinceAwakening.Seconds()
		resp.TimeSinceAwakeningSeconds = &secs
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func mapDashboardSummaryError(err error) apperror.AppError {
	switch {
	case errors.Is(err, sleep.ErrInvalidTimezone):
		return apperror.New(apperror.CodeInvalidTimezone, err.Error())
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
		errors.Is(err, sleep.ErrEmptyBabyID),
		errors.Is(err, sleep.ErrEffectiveFromTooOld):
		return apperror.New(apperror.CodeInvalidArgument, err.Error())
	default:
		var appErr apperror.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return apperror.New(apperror.CodeInternalError, "unexpected error")
	}
}
