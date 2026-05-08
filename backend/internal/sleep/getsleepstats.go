package sleep

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type GetSleepStatsQuery struct {
	BabyID   BabyID
	Timezone string
}

type TodayStats struct {
	TotalSleepSeconds  float64
	TotalNapSeconds    float64
	TotalActiveSeconds float64
}

type PeriodAverage struct {
	AvgSleepSeconds  float64
	AvgNapSeconds    float64
	AvgActiveSeconds float64
}

type NightWindowInfo struct {
	StartHHMM string
	EndHHMM   string
}

type SleepStats struct {
	Today       TodayStats
	Summary     map[string]PeriodAverage
	NightWindow *NightWindowInfo
}

type GetSleepStatsHandler struct {
	sessions SleepSessionHistoryRepository
	windows  NightWindowRepository
	now      func() time.Time
}

func NewGetSleepStatsHandler(sessions SleepSessionHistoryRepository, windows NightWindowRepository) *GetSleepStatsHandler {
	return &GetSleepStatsHandler{sessions: sessions, windows: windows, now: time.Now}
}

func (h *GetSleepStatsHandler) Handle(ctx context.Context, q GetSleepStatsQuery) (SleepStats, error) {
	if q.Timezone == "" {
		return SleepStats{}, ErrInvalidTimezone
	}
	loc, err := time.LoadLocation(q.Timezone)
	if err != nil {
		return SleepStats{}, ErrInvalidTimezone
	}

	now := h.now().UTC()

	windows, err := h.windows.FindByBabyID(ctx, q.BabyID)
	if err != nil {
		return SleepStats{}, fmt.Errorf("load night windows: %w", err)
	}

	w, ok := findWindowAt(windows, now)
	if !ok {
		return SleepStats{}, nil
	}
	nightWindow := &NightWindowInfo{
		StartHHMM: fmt.Sprintf("%02d:%02d", w.Start().Hour(), w.Start().Minute()),
		EndHHMM:   fmt.Sprintf("%02d:%02d", w.End().Hour(), w.End().Minute()),
	}

	// Fetch sessions covering 90d plus a 24h buffer for cross-midnight naps.
	ninetyDaysAgoLocal := now.In(loc).AddDate(0, 0, -90)
	periodStart := time.Date(ninetyDaysAgoLocal.Year(), ninetyDaysAgoLocal.Month(), ninetyDaysAgoLocal.Day(), 0, 0, 0, 0, loc).UTC()
	fetchStart := periodStart.Add(-24 * time.Hour)

	dr, err := NewDateRange(fetchStart, now)
	if err != nil {
		return SleepStats{}, err
	}

	sessions, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dr)
	if err != nil {
		return SleepStats{}, fmt.Errorf("get sleep stats: %w", err)
	}

	// Classify all completed sessions once; reuse for today and summary.
	classified := make([]statsCS, 0, len(sessions))
	for _, s := range sessions {
		if s.IsActive() {
			continue
		}
		cls := SleepClassificationUnknown
		if w, ok := FindWindowForSession(windows, s); ok {
			cls, err = Classify(s, q.Timezone, w)
			if err != nil {
				return SleepStats{}, err
			}
		}
		classified = append(classified, statsCS{s: s, cls: cls})
	}

	// Find the single active session (if any) and classify it.
	var activeSession *statsCS
	for _, s := range sessions {
		if !s.IsActive() {
			continue
		}
		cls := SleepClassificationUnknown
		if sw, ok := FindWindowForSession(windows, s); ok {
			cls, err = Classify(s, q.Timezone, sw)
			if err != nil {
				return SleepStats{}, err
			}
		}
		cs := statsCS{s: s, cls: cls}
		activeSession = &cs
		break
	}

	today := statsTodayTotals(classified, activeSession, loc, now)
	summary := statsSummaryAverages(classified, loc, now)

	return SleepStats{
		Today:       today,
		Summary:     summary,
		NightWindow: nightWindow,
	}, nil
}

// findWindowAt returns the night window effective at time t.
// windows must be ordered by effective_from ASC.
func findWindowAt(windows []NightWindow, t time.Time) (NightWindow, bool) {
	var best NightWindow
	found := false
	for _, w := range windows {
		if !w.effectiveFrom.After(t) && (w.effectiveTo == nil || w.effectiveTo.After(t)) {
			best = w
			found = true
		}
	}
	return best, found
}

type statsCS struct {
	s   SleepSession
	cls SleepClassification
}

// statsTodayTotals computes active, nap, and total sleep seconds for today.
//
// Active is awake time since the last completed night sleep that ended today,
// or since dayStart if none exists. If the baby is currently in a night sleep
// (activeSession is a night classification), active is 0.
// Naps counts completed nap sessions that overlap [anchor, now).
// Total sleep is naps + completed night sessions whose StartedAt falls in today.
func statsTodayTotals(classified []statsCS, activeSession *statsCS, loc *time.Location, now time.Time) TodayStats {
	localNow := now.In(loc)
	dayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc).UTC()
	dayEnd := dayStart.Add(24 * time.Hour)

	// a. Anchor detection.
	anchor := dayStart
	if activeSession != nil && activeSession.cls == SleepClassificationNight {
		anchor = now // active = 0
	} else {
		for _, cs := range classified {
			if cs.cls != SleepClassificationNight {
				continue
			}
			stoppedAt, _ := cs.s.StoppedAt()
			if !stoppedAt.Before(dayStart) && stoppedAt.Before(dayEnd) && stoppedAt.After(anchor) {
				anchor = stoppedAt
			}
		}
	}

	// b. Naps since anchor.
	var napSessions []SleepSession
	var totalNap time.Duration
	for _, cs := range classified {
		if cs.cls != SleepClassificationNap {
			continue
		}
		stoppedAt, _ := cs.s.StoppedAt()
		if cs.s.StartedAt().Before(now) && stoppedAt.After(anchor) {
			napSessions = append(napSessions, cs.s)
			totalNap += attributedDuration(cs.s, now)
		}
	}

	// c. Active since anchor.
	sort.Slice(napSessions, func(i, j int) bool {
		return napSessions[i].StartedAt().Before(napSessions[j].StartedAt())
	})
	active := activeTimeInWindow(napSessions, anchor, now, now)

	// d. Night sleep started today (StartedAt within [dayStart, dayEnd)).
	var totalNight time.Duration
	for _, cs := range classified {
		if cs.cls != SleepClassificationNight {
			continue
		}
		startedAt := cs.s.StartedAt()
		if !startedAt.Before(dayStart) && startedAt.Before(dayEnd) {
			totalNight += attributedDuration(cs.s, now)
		}
	}

	return TodayStats{
		TotalSleepSeconds:  (totalNap + totalNight).Seconds(),
		TotalNapSeconds:    totalNap.Seconds(),
		TotalActiveSeconds: active.Seconds(),
	}
}

// statsSummaryAverages computes rolling averages for 7d, 14d, 30d, 90d.
// Today (the partial calendar day) is excluded from all averages.
func statsSummaryAverages(classified []statsCS, loc *time.Location, now time.Time) map[string]PeriodAverage {
	localNow := now.In(loc)
	today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)

	type dayStats struct{ sleep, nap, active float64 }
	daily := make([]dayStats, 90)

	for i := 0; i < 90; i++ {
		dayStart := today.AddDate(0, 0, -(i + 1))
		dayEnd := dayStart.AddDate(0, 0, 1)
		dayStartUTC := dayStart.UTC()
		dayEndUTC := dayEnd.UTC()

		var inDay []SleepSession
		var totalSleep, totalNap time.Duration

		for _, cs := range classified {
			if !includeInDay(cs.s, dayStartUTC, dayEndUTC) {
				continue
			}
			inDay = append(inDay, cs.s)
			d := attributedDuration(cs.s, now)
			totalSleep += d
			if cs.cls == SleepClassificationNap {
				totalNap += d
			}
		}

		sort.Slice(inDay, func(a, b int) bool {
			return inDay[a].StartedAt().Before(inDay[b].StartedAt())
		})

		effectiveEnd := dayEndUTC
		if now.Before(dayEndUTC) {
			effectiveEnd = now
		}

		var active time.Duration
		if now.After(dayStartUTC) {
			active = activeTimeInWindow(inDay, dayStartUTC, effectiveEnd, now)
		}

		daily[i] = dayStats{
			sleep:  totalSleep.Seconds(),
			nap:    totalNap.Seconds(),
			active: active.Seconds(),
		}
	}

	avg := func(n int) PeriodAverage {
		var s, p, a float64
		for i := 0; i < n; i++ {
			s += daily[i].sleep
			p += daily[i].nap
			a += daily[i].active
		}
		f := float64(n)
		return PeriodAverage{
			AvgSleepSeconds:  s / f,
			AvgNapSeconds:    p / f,
			AvgActiveSeconds: a / f,
		}
	}

	return map[string]PeriodAverage{
		"7d":  avg(7),
		"14d": avg(14),
		"30d": avg(30),
		"90d": avg(90),
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
	return start.Before(dayEnd) && stoppedAt.After(dayStart)
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
