package sleep

import (
	"context"
	"fmt"
	"time"
)

type GetSleepStatsQuery struct {
	BabyID   BabyID
	Timezone string
}

type DayStats struct {
	Date               string
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
	Days        []DayStats
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
	currentDayStart := statsDayStartForInstant(now, w, loc)
	fetchStart := currentDayStart.AddDate(0, 0, -90).Add(-24 * time.Hour)

	dr, err := NewDateRange(fetchStart, now)
	if err != nil {
		return SleepStats{}, err
	}

	sessions, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dr)
	if err != nil {
		return SleepStats{}, fmt.Errorf("get sleep stats: %w", err)
	}

	dayStats, err := buildDayStats(sessions, windows, loc, now, 91)
	if err != nil {
		return SleepStats{}, fmt.Errorf("get sleep stats: %w", err)
	}

	numDays := min(7, len(dayStats))
	days := make([]DayStats, numDays)
	copy(days, dayStats[:numDays])
	summary := summaryAveragesFromDayStats(dayStats[1:])

	return SleepStats{
		Days:        days,
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

func buildDayStats(
	sessions []SleepSession,
	windows []NightWindow,
	loc *time.Location,
	now time.Time,
	days int,
) ([]DayStats, error) {
	windowNow, ok := findWindowAt(windows, now)
	if !ok {
		return nil, nil
	}

	currentDayStart := statsDayStartForInstant(now, windowNow, loc)
	result := make([]DayStats, 0, days)
	for i := 0; i < days; i++ {
		dayStart := currentDayStart.AddDate(0, 0, -i)
		windowAtDayStart, ok := findWindowAt(windows, dayStart)
		if !ok {
			continue
		}
		dayStart = statsDayStartForDate(dayStart.In(loc), windowAtDayStart, loc)
		dayEnd := dayStart.Add(24 * time.Hour)
		sleep, nap := overlapTotalsForWindow(sessions, windows, loc, dayStart, dayEnd, now)
		active := dayEnd.Sub(dayStart).Seconds() - sleep
		if i == 0 && now.Before(dayEnd) {
			active = now.Sub(dayStart).Seconds() - sleep
		}
		localDayStart := dayStart.In(loc)
		result = append(result, DayStats{
			Date:               fmt.Sprintf("%04d-%02d-%02d", localDayStart.Year(), int(localDayStart.Month()), localDayStart.Day()),
			TotalSleepSeconds:  sleep,
			TotalNapSeconds:    nap,
			TotalActiveSeconds: maxFloat64(0, active),
		})
	}
	return result, nil
}

func overlapTotalsForWindow(
	sessions []SleepSession,
	windows []NightWindow,
	loc *time.Location,
	windowStart time.Time,
	windowEnd time.Time,
	now time.Time,
) (sleepSeconds float64, napSeconds float64) {
	for _, session := range sessions {
		classification, ok, err := classifyForBuild(session, windows, loc)
		if err != nil || !ok {
			continue
		}

		seconds := overlapSeconds(session, windowStart, windowEnd, now)
		if seconds <= 0 {
			continue
		}

		sleepSeconds += seconds
		if classification == SleepClassificationNap {
			napSeconds += seconds
		}
	}
	return sleepSeconds, napSeconds
}

func overlapSeconds(session SleepSession, start time.Time, end time.Time, now time.Time) float64 {
	sessionStart := session.StartedAt()
	sessionEnd, ok := session.StoppedAt()
	if !ok {
		sessionEnd = now
	}
	overlapStart := maxTime(sessionStart, start)
	overlapEnd := minTime(sessionEnd, end)
	if !overlapEnd.After(overlapStart) {
		return 0
	}
	return overlapEnd.Sub(overlapStart).Seconds()
}

func statsDayStartForInstant(now time.Time, nw NightWindow, loc *time.Location) time.Time {
	localNow := now.In(loc)
	dayStart := statsDayStartForDate(localNow, nw, loc)
	if localNow.Before(dayStart.In(loc)) {
		return dayStart.AddDate(0, 0, -1)
	}
	return dayStart
}

func statsDayStartForDate(localDate time.Time, nw NightWindow, loc *time.Location) time.Time {
	return time.Date(
		localDate.Year(),
		localDate.Month(),
		localDate.Day(),
		nw.End().Hour(),
		nw.End().Minute(),
		0,
		0,
		loc,
	).UTC()
}

func summaryAveragesFromDayStats(days []DayStats) map[string]PeriodAverage {
	avg := func(n int) PeriodAverage {
		var sleep, nap, active float64
		var count int
		for i := 0; i < n && i < len(days); i++ {
			if days[i].TotalSleepSeconds == 0 {
				continue
			}
			sleep += days[i].TotalSleepSeconds
			nap += days[i].TotalNapSeconds
			active += days[i].TotalActiveSeconds
			count++
		}
		if count == 0 {
			return PeriodAverage{}
		}
		return PeriodAverage{
			AvgSleepSeconds:  sleep / float64(count),
			AvgNapSeconds:    nap / float64(count),
			AvgActiveSeconds: active / float64(count),
		}
	}
	return map[string]PeriodAverage{
		"7d":  avg(7),
		"14d": avg(14),
		"30d": avg(30),
		"90d": avg(90),
	}
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
