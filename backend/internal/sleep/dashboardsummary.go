package sleep

import (
	"context"
	"errors"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

// DashboardSummary holds all metrics needed for a single dashboard screen load.
type DashboardSummary struct {
	ActiveSession       *ActiveSessionInfo
	TimeSinceSleepStart *time.Duration
	TimeSinceAwakening  *time.Duration
	Today               DailySummary
	Rolling7d           RollingAverage
	Rolling14d          RollingAverage
}

// ActiveSessionInfo holds identifying data for an ongoing sleep session.
type ActiveSessionInfo struct {
	ID        SleepSessionID
	StartedAt time.Time
	Duration  time.Duration
}

// RollingAverage holds averaged daily sleep and active durations over a window.
type RollingAverage struct {
	AvgDailySleep  time.Duration
	AvgDailyActive time.Duration
}

// GetDashboardSummaryQuery holds the inputs for the dashboard summary.
type GetDashboardSummaryQuery struct {
	BabyID BabyID
}

// GetDashboardSummaryHandler computes all dashboard metrics in a single use case.
type GetDashboardSummaryHandler struct {
	profiles SleepProfileRepository
	sessions SleepSessionHistoryRepository
	active   ActiveSleepSessionRepository
	elapsed  SleepElapsedTimeRepository
	now      func() time.Time
}

// NewGetDashboardSummaryHandler returns a handler backed by the given repositories.
func NewGetDashboardSummaryHandler(
	profiles SleepProfileRepository,
	sessions SleepSessionHistoryRepository,
	active ActiveSleepSessionRepository,
	elapsed SleepElapsedTimeRepository,
) *GetDashboardSummaryHandler {
	return &GetDashboardSummaryHandler{
		profiles: profiles,
		sessions: sessions,
		active:   active,
		elapsed:  elapsed,
		now:      time.Now,
	}
}

// Handle computes and returns all dashboard metrics for the given baby.
func (h *GetDashboardSummaryHandler) Handle(ctx context.Context, q GetDashboardSummaryQuery) (DashboardSummary, error) {
	profile, err := h.profiles.FindByBabyID(ctx, q.BabyID)
	if err != nil {
		return DashboardSummary{}, err
	}

	loc, err := time.LoadLocation(profile.Timezone())
	if err != nil {
		return DashboardSummary{}, ErrInvalidTimezone
	}

	now := h.now().UTC()
	localNow := now.In(loc)

	todayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc).UTC()
	todayEnd := time.Date(localNow.Year(), localNow.Month(), localNow.Day()+1, 0, 0, 0, 0, loc).UTC()

	// Single query covers 14 days plus a 24-hour buffer so naps that started
	// just before a day boundary are captured by ComputeDailySummary.
	windowStart := todayStart.AddDate(0, 0, -14).Add(-24 * time.Hour)
	dr, err := NewDateRange(windowStart, todayEnd)
	if err != nil {
		return DashboardSummary{}, err
	}
	allSessions, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dr)
	if err != nil {
		return DashboardSummary{}, err
	}

	today := ComputeDailySummary(allSessions, todayStart, todayEnd, now)
	rolling7d := computeRollingAverage(allSessions, loc, localNow, 7, now)
	rolling14d := computeRollingAverage(allSessions, loc, localNow, 14, now)

	var activeSess *ActiveSessionInfo
	activeSession, err := h.active.FindActiveByBabyID(ctx, q.BabyID)
	if err == nil {
		d := now.Sub(activeSession.StartedAt())
		activeSess = &ActiveSessionInfo{
			ID:        activeSession.ID(),
			StartedAt: activeSession.StartedAt(),
			Duration:  d,
		}
	} else if !isNotFoundErr(err) {
		return DashboardSummary{}, err
	}

	var timeSinceSleepStart *time.Duration
	mostRecent, err := h.elapsed.FindMostRecentByBabyID(ctx, q.BabyID)
	if err == nil {
		d := now.Sub(mostRecent.StartedAt())
		timeSinceSleepStart = &d
	} else if !isNotFoundErr(err) {
		return DashboardSummary{}, err
	}

	var timeSinceAwakening *time.Duration
	mostRecentCompleted, err := h.elapsed.FindMostRecentCompletedByBabyID(ctx, q.BabyID)
	if err == nil {
		if stoppedAt, ok := mostRecentCompleted.StoppedAt(); ok {
			d := now.Sub(stoppedAt)
			timeSinceAwakening = &d
		}
	} else if !isNotFoundErr(err) {
		return DashboardSummary{}, err
	}

	return DashboardSummary{
		ActiveSession:       activeSess,
		TimeSinceSleepStart: timeSinceSleepStart,
		TimeSinceAwakening:  timeSinceAwakening,
		Today:               today,
		Rolling7d:           rolling7d,
		Rolling14d:          rolling14d,
	}, nil
}

// computeRollingAverage averages ComputeDailySummary over the given number of
// days ending on localNow's calendar day (inclusive).
func computeRollingAverage(sessions []SleepSession, loc *time.Location, localNow time.Time, days int, now time.Time) RollingAverage {
	var totalSleep, totalActive time.Duration
	for i := 0; i < days; i++ {
		ref := localNow.AddDate(0, 0, -i)
		dayStart := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, loc).UTC()
		dayEnd := time.Date(ref.Year(), ref.Month(), ref.Day()+1, 0, 0, 0, 0, loc).UTC()
		s := ComputeDailySummary(sessions, dayStart, dayEnd, now)
		totalSleep += s.TotalSleep
		totalActive += s.TotalActive
	}
	return RollingAverage{
		AvgDailySleep:  totalSleep / time.Duration(days),
		AvgDailyActive: totalActive / time.Duration(days),
	}
}

func isNotFoundErr(err error) bool {
	var appErr apperror.AppError
	return errors.As(err, &appErr) && appErr.Code == apperror.CodeNotFound
}
