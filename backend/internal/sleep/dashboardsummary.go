package sleep

import (
	"context"
	"time"
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
	BabyID   BabyID
	Timezone string // IANA timezone name
}

// GetDashboardSummaryHandler computes all dashboard metrics in a single use case.
type GetDashboardSummaryHandler struct {
	sessions SleepSessionHistoryRepository
	now      func() time.Time
}

// NewGetDashboardSummaryHandler returns a handler backed by the given repositories.
func NewGetDashboardSummaryHandler(
	sessions SleepSessionHistoryRepository,
) *GetDashboardSummaryHandler {
	return &GetDashboardSummaryHandler{
		sessions: sessions,
		now:      time.Now,
	}
}

// Handle computes and returns all dashboard metrics for the given baby.
//
// All elapsed-time and active-session fields are derived from the same
// 14-day session slice — no extra DB queries are needed for those fields.
func (h *GetDashboardSummaryHandler) Handle(ctx context.Context, q GetDashboardSummaryQuery) (DashboardSummary, error) {
	loc, err := time.LoadLocation(q.Timezone)
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
	// allSessions is ordered by started_at DESC.
	allSessions, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dr)
	if err != nil {
		return DashboardSummary{}, err
	}

	return DashboardSummary{
		ActiveSession:       findActiveSession(allSessions, now),
		TimeSinceSleepStart: timeSinceSleepStart(allSessions, now),
		TimeSinceAwakening:  timeSinceAwakening(allSessions, now),
		Today:               ComputeDailySummary(allSessions, todayStart, todayEnd, now),
		Rolling7d:           computeRollingAverage(allSessions, loc, localNow, 7, now),
		Rolling14d:          computeRollingAverage(allSessions, loc, localNow, 14, now),
	}, nil
}

func findActiveSession(sessions []SleepSession, now time.Time) *ActiveSessionInfo {
	for _, s := range sessions {
		if s.IsActive() {
			return &ActiveSessionInfo{
				ID:        s.ID(),
				StartedAt: s.StartedAt(),
				Duration:  now.Sub(s.StartedAt()),
			}
		}
	}
	return nil
}

// timeSinceSleepStart returns elapsed time since the most recently started session.
// sessions must be ordered by started_at DESC.
func timeSinceSleepStart(sessions []SleepSession, now time.Time) *time.Duration {
	if len(sessions) == 0 {
		return nil
	}
	d := now.Sub(sessions[0].StartedAt())
	return &d
}

func timeSinceAwakening(sessions []SleepSession, now time.Time) *time.Duration {
	var latestStop time.Time
	for _, s := range sessions {
		if stoppedAt, ok := s.StoppedAt(); ok && stoppedAt.After(latestStop) {
			latestStop = stoppedAt
		}
	}
	if latestStop.IsZero() {
		return nil
	}
	d := now.Sub(latestStop)
	return &d
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
