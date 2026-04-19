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
// so that naps beginning just before midnight are captured, then filters them
// with includeInDay before computing totals.
func (h *GetDailySummaryHandler) Handle(ctx context.Context, q GetDailySummaryQuery) (DailySummary, error) {
	loc, err := time.LoadLocation(q.Timezone)
	if err != nil {
		return DailySummary{}, ErrInvalidTimezone
	}

	localDate := q.Date.In(loc)
	dayStart := time.Date(localDate.Year(), localDate.Month(), localDate.Day(), 0, 0, 0, 0, loc).UTC()
	// AddDate(0,0,1) handles DST correctly: Go resolves the new wall-clock date
	// in the given location, yielding a 23h or 25h window on transition days.
	dayEnd := time.Date(localDate.Year(), localDate.Month(), localDate.Day()+1, 0, 0, 0, 0, loc).UTC()

	now := h.now().UTC()
	if !now.After(dayStart) {
		return DailySummary{}, nil
	}
	effectiveEnd := dayEnd
	if now.Before(dayEnd) {
		effectiveEnd = now
	}

	// Look back 24 hours before dayStart to capture naps that started just before
	// the day boundary and overlap it (e.g. a nap starting at 23:30 the prior day).
	lookback := dayStart.Add(-24 * time.Hour)
	dr, err := NewDateRange(lookback, dayEnd)
	if err != nil {
		return DailySummary{}, err
	}
	all, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dr)
	if err != nil {
		return DailySummary{}, err
	}

	included := make([]SleepSession, 0, len(all))
	for _, s := range all {
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

	totalActive := activeTimeInWindow(included, dayStart, effectiveEnd, now)

	return DailySummary{
		TotalSleep:  totalSleep,
		TotalActive: totalActive,
	}, nil
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
		// Clip to window.
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
