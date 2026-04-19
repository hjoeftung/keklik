package sleep

import (
	"context"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

// ElapsedTimeQuery holds the inputs for elapsed-time calculations.
type ElapsedTimeQuery struct {
	BabyID BabyID
}

// ElapsedTimeResult holds the computed duration. Present is false when no
// relevant event exists (empty dataset), distinguishing it from a zero duration.
type ElapsedTimeResult struct {
	Duration time.Duration
	Present  bool
}

// GetElapsedTimeHandler computes time-since-event metrics for use by GetDashboardSummary.
type GetElapsedTimeHandler struct {
	sessions SleepElapsedTimeRepository
	now      func() time.Time
}

// NewGetElapsedTimeHandler returns a handler backed by the given repository.
func NewGetElapsedTimeHandler(sessions SleepElapsedTimeRepository) *GetElapsedTimeHandler {
	return &GetElapsedTimeHandler{sessions: sessions, now: time.Now}
}

// GetTimeSinceLastSleepStart returns the elapsed time since the most recent sleep
// session started for the given baby. Returns Present=false when no sessions exist.
func (h *GetElapsedTimeHandler) GetTimeSinceLastSleepStart(ctx context.Context, q ElapsedTimeQuery) (ElapsedTimeResult, error) {
	session, err := h.sessions.FindMostRecentByBabyID(ctx, q.BabyID)
	if err != nil {
		if appErr, ok := err.(apperror.AppError); ok && appErr.Code == apperror.CodeNotFound {
			return ElapsedTimeResult{}, nil
		}
		return ElapsedTimeResult{}, err
	}
	return ElapsedTimeResult{
		Duration: h.now().Sub(session.StartedAt()),
		Present:  true,
	}, nil
}

// GetTimeSinceLastAwakening returns the elapsed time since the most recent sleep
// session ended for the given baby. Returns Present=false when no completed sessions exist.
func (h *GetElapsedTimeHandler) GetTimeSinceLastAwakening(ctx context.Context, q ElapsedTimeQuery) (ElapsedTimeResult, error) {
	session, err := h.sessions.FindMostRecentCompletedByBabyID(ctx, q.BabyID)
	if err != nil {
		if appErr, ok := err.(apperror.AppError); ok && appErr.Code == apperror.CodeNotFound {
			return ElapsedTimeResult{}, nil
		}
		return ElapsedTimeResult{}, err
	}
	stoppedAt, _ := session.StoppedAt()
	return ElapsedTimeResult{
		Duration: h.now().Sub(stoppedAt),
		Present:  true,
	}, nil
}
