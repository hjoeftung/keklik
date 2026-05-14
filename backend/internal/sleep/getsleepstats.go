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

	userDays, err := buildUserDays(sessions, windows, loc, now, 91)
	if err != nil {
		return SleepStats{}, fmt.Errorf("get sleep stats: %w", err)
	}

	localNow := now.In(loc)
	numDays := min(7, len(userDays))
	days := make([]DayStats, numDays)
	for i := range days {
		d := localNow.AddDate(0, 0, -i)
		days[i] = DayStats{
			Date:               fmt.Sprintf("%04d-%02d-%02d", d.Year(), int(d.Month()), d.Day()),
			TotalNapSeconds:    userDays[i].NapDuration(now).Seconds(),
			TotalSleepSeconds:  userDays[i].NightDuration(now).Seconds() + userDays[i].NapDuration(now).Seconds(),
			TotalActiveSeconds: userDays[i].ActiveDuration(now).Seconds(),
		}
	}
	summary := summaryAveragesFrom(userDays[1:], now)

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

func summaryAveragesFrom(days []*UserDay, now time.Time) map[string]PeriodAverage {
	avg := func(n int) PeriodAverage {
		var sleep, nap, active float64
		for i := 0; i < n && i < len(days); i++ {
			sleep += days[i].NightDuration(now).Seconds() + days[i].NapDuration(now).Seconds()
			nap += days[i].NapDuration(now).Seconds()
			active += days[i].ActiveDuration(now).Seconds()
		}
		count := min(n, len(days))
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
