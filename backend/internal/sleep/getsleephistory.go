package sleep

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GetSleepHistoryQuery holds the inputs for querying a baby's sleep history.
type GetSleepHistoryQuery struct {
	BabyID   BabyID
	Timezone string // IANA timezone name
	Period   string // "today", "7d", "14d", "30d", or "90d"
}

type SleepHistoryEntry struct {
	Session        SleepSession
	Classification SleepClassification
}

// GetSleepHistoryHandler executes the GetSleepHistory use case.
type GetSleepHistoryHandler struct {
	sessions SleepSessionHistoryRepository
	windows  NightWindowRepository
}

// NewGetSleepHistoryHandler returns a GetSleepHistoryHandler backed by the given repositories.
func NewGetSleepHistoryHandler(sessions SleepSessionHistoryRepository, windows NightWindowRepository) *GetSleepHistoryHandler {
	return &GetSleepHistoryHandler{sessions: sessions, windows: windows}
}

// Handle returns sleep sessions for the given baby and period, ordered by
// started_at descending. It always returns a non-nil slice (empty when there
// are no results).
func (h *GetSleepHistoryHandler) Handle(ctx context.Context, q GetSleepHistoryQuery) ([]SleepHistoryEntry, error) {
	if q.Timezone == "" {
		return nil, ErrInvalidTimezone
	}

	dateRange, err := periodToDateRange(q.Period, q.Timezone, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	sessions, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("get sleep history: %w", err)
	}

	windows, err := h.windows.FindByBabyID(ctx, q.BabyID)
	if err != nil {
		return nil, fmt.Errorf("load night windows: %w", err)
	}

	if sessions == nil {
		return []SleepHistoryEntry{}, nil
	}

	result := make([]SleepHistoryEntry, 0, len(sessions))
	for _, session := range sessions {
		classification := SleepClassificationUnknown
		if window, ok := FindWindowForSession(windows, session); ok {
			classification, err = Classify(session, q.Timezone, window)
			if err != nil {
				return nil, err
			}
		}

		result = append(result, SleepHistoryEntry{
			Session:        session,
			Classification: classification,
		})
	}

	return result, nil
}

// periodToDateRange converts a period string and IANA timezone name into a
// DateRange expressed in UTC, using the local calendar in the given timezone.
//
//   - "today" → local midnight today .. local midnight tomorrow
//   - "Nd"    → local midnight N days ago .. now (UTC), where 1 ≤ N ≤ 120
func periodToDateRange(period, timezone string, now time.Time) (DateRange, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return DateRange{}, ErrInvalidTimezone
	}

	localNow := now.In(loc)

	if period == "today" {
		startLocal := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
		endLocal := startLocal.AddDate(0, 0, 1)
		return NewDateRange(startLocal.UTC(), endLocal.UTC())
	}

	if strings.HasSuffix(period, "d") {
		n, parseErr := strconv.Atoi(strings.TrimSuffix(period, "d"))
		if parseErr == nil && n >= 1 && n <= 120 {
			daysAgoLocal := localNow.AddDate(0, 0, -n)
			startLocal := time.Date(daysAgoLocal.Year(), daysAgoLocal.Month(), daysAgoLocal.Day(), 0, 0, 0, 0, loc)
			return NewDateRange(startLocal.UTC(), now)
		}
	}

	return DateRange{}, ErrInvalidSleepHistoryPeriod
}
