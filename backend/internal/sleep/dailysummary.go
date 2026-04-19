package sleep

import (
	"context"
	"sort"
	"time"
)

// DailySummary holds the computed sleep and active totals for a single local calendar day.
//
// TotalSleep is the sum of the full attributed durations: all naps overlapping
// the day plus the night sleep that started on the day. An active session is
// measured up to the current time.
//
// TotalActive is the awake time within [dayStart, min(dayEnd, now)] — the
// calendar window minus the portion of attributed sessions that falls inside it.
// Because night sleep often extends past midnight, TotalSleep + TotalActive may
// exceed the calendar window; that is expected and correct.
type DailySummary struct {
	TotalSleep  time.Duration
	TotalActive time.Duration
}

// GetDailySummaryQuery holds the inputs for the daily summary calculation.
type GetDailySummaryQuery struct {
	BabyID   BabyID
	Timezone string    // IANA timezone name
	Date     time.Time // any instant within the target day; the local date is resolved in Timezone
}

// GetDailySummaryHandler computes sleep and active totals for a single local calendar day.
type GetDailySummaryHandler struct {
	sessions SleepSessionHistoryRepository
	now      func() time.Time
}

// NewGetDailySummaryHandler returns a handler backed by the given repository.
func NewGetDailySummaryHandler(sessions SleepSessionHistoryRepository) *GetDailySummaryHandler {
	return &GetDailySummaryHandler{sessions: sessions, now: time.Now}
}

// Handle returns sleep and active time totals for the local calendar day that
// contains q.Date in q.Timezone.
//
// The method queries sessions whose started_at falls in [dayStart−24h, dayEnd)
// so that naps beginning just before midnight are captured, then delegates to
// ComputeDailySummary for the actual calculation.
func (h *GetDailySummaryHandler) Handle(ctx context.Context, q GetDailySummaryQuery) (DailySummary, error) {
	loc, err := time.LoadLocation(q.Timezone)
	if err != nil {
		return DailySummary{}, ErrInvalidTimezone
	}

	now := h.now().UTC()
	localDate := q.Date.In(loc)
	dayStart := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), 0, 0, 0, 0, loc).UTC()
	dayEnd := time.Date(localDate.Year(), localDate.Month(), localDate.Day()+1, 0, 0, 0, 0, loc).UTC()

	if !now.After(dayStart) {
		return DailySummary{}, nil
	}

	// Look back 24 hours before dayStart to capture naps that started just before
	// the day boundary and overlap it (e.g. a nap starting at 23:30 the prior day).
	dr, err := NewDateRange(dayStart.Add(-24*time.Hour), dayEnd)
	if err != nil {
		return DailySummary{}, err
	}
	sessions, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dr)
	if err != nil {
		return DailySummary{}, err
	}

	return ComputeDailySummary(sessions, dayStart, dayEnd, now), nil
}

// ComputeDailySummary calculates sleep and active totals for the calendar day
// [dayStart, dayEnd) from an already-fetched slice of sessions.
//
// Callers that need summaries for multiple days (e.g. rolling averages) should
// fetch all sessions for the full window once and call this function per day,
// rather than issuing one database query per day.
//
// sessions must contain all sessions that could be relevant: at minimum those
// whose started_at falls in [dayStart−24h, dayEnd). They need not be pre-sorted.
// now is used to cap active-session durations and the active-time window.
func ComputeDailySummary(sessions []SleepSession, dayStart, dayEnd, now time.Time) DailySummary {
	if !now.After(dayStart) {
		return DailySummary{}
	}
	effectiveEnd := dayEnd
	if now.Before(dayEnd) {
		effectiveEnd = now
	}

	included := make([]SleepSession, 0, len(sessions))
	for _, s := range sessions {
		if includeInDay(s, dayStart, dayEnd) {
			included = append(included, s)
		}
	}
	sort.Slice(included, func(i, j int) bool {
		return included[i].StartedAt().Before(included[j].StartedAt())
	})

	var totalSleep time.Duration
	for _, s := range included {
		totalSleep += attributedDuration(s, now)
	}

	return DailySummary{
		TotalSleep:  totalSleep,
		TotalActive: activeTimeInWindow(included, dayStart, effectiveEnd, now),
	}
}

// includeInDay reports whether s contributes to the daily summary for [dayStart, dayEnd).
//
// Naps are included when they overlap the window (started before dayEnd and, if
// completed, ended after dayStart). Night sleeps are included only when they
// started within the window. Active (unclassified) sessions are treated like naps.
func includeInDay(s SleepSession, dayStart, dayEnd time.Time) bool {
	start := s.StartedAt()
	if s.IsActive() {
		return start.Before(dayEnd)
	}
	stoppedAt, _ := s.StoppedAt()
	switch s.Classification() {
	case SleepClassificationNap:
		return start.Before(dayEnd) && stoppedAt.After(dayStart)
	case SleepClassificationNight:
		return !start.Before(dayStart) && start.Before(dayEnd)
	default:
		// Completed but unclassified: use overlap rule.
		return start.Before(dayEnd) && stoppedAt.After(dayStart)
	}
}

// attributedDuration returns the full duration to credit to s in the daily sleep
// total. Completed sessions use their stored duration; active sessions are
// measured from start to now.
func attributedDuration(s SleepSession, now time.Time) time.Duration {
	if d, ok := s.Duration(); ok {
		return d
	}
	if d := now.Sub(s.StartedAt()); d > 0 {
		return d
	}
	return 0
}

// activeTimeInWindow returns the awake time within [windowStart, windowEnd] by
// computing the window duration minus the total sleep clipped to that window.
//
// Sessions are expected to be sorted by start time and non-overlapping (the
// domain allows only one active session per baby at a time). Overlapping
// intervals are merged defensively before subtraction.
func activeTimeInWindow(sessions []SleepSession, windowStart, windowEnd, now time.Time) time.Duration {
	if !windowEnd.After(windowStart) {
		return 0
	}

	type interval struct{ start, end time.Time }
	intervals := make([]interval, 0, len(sessions))
	for _, s := range sessions {
		start := s.StartedAt()
		var end time.Time
		if stopped, ok := s.StoppedAt(); ok {
			end = stopped
		} else {
			end = now
		}
		if start.Before(windowStart) {
			start = windowStart
		}
		if end.After(windowEnd) {
			end = windowEnd
		}
		if end.After(start) {
			intervals = append(intervals, interval{start, end})
		}
	}

	if len(intervals) == 0 {
		return windowEnd.Sub(windowStart)
	}

	// Merge overlapping intervals.
	merged := []interval{intervals[0]}
	for _, iv := range intervals[1:] {
		last := &merged[len(merged)-1]
		if !iv.start.After(last.end) {
			if iv.end.After(last.end) {
				last.end = iv.end
			}
		} else {
			merged = append(merged, iv)
		}
	}

	var sleepInWindow time.Duration
	for _, iv := range merged {
		sleepInWindow += iv.end.Sub(iv.start)
	}
	return windowEnd.Sub(windowStart) - sleepInWindow
}
