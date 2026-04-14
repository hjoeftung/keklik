package sleep

import (
	"context"
	"fmt"
	"time"
)

// GetSleepHistoryQuery holds the inputs for querying a baby's sleep history.
type GetSleepHistoryQuery struct {
	BabyID   BabyID
	Timezone string // IANA timezone name; if empty, resolved from the sleep profile
	Period   string // "today", "7d", or "14d"
}

// GetSleepHistoryHandler executes the GetSleepHistory use case.
type GetSleepHistoryHandler struct {
	sessions SleepSessionHistoryRepository
	profiles SleepProfileRepository
}

// NewGetSleepHistoryHandler returns a GetSleepHistoryHandler backed by the given repositories.
func NewGetSleepHistoryHandler(sessions SleepSessionHistoryRepository, profiles SleepProfileRepository) *GetSleepHistoryHandler {
	return &GetSleepHistoryHandler{sessions: sessions, profiles: profiles}
}

// Handle returns sleep sessions for the given baby and period, ordered by
// started_at descending. It always returns a non-nil slice (empty when there
// are no results).
func (h *GetSleepHistoryHandler) Handle(ctx context.Context, q GetSleepHistoryQuery) ([]SleepSession, error) {
	tz := q.Timezone
	if tz == "" {
		profile, err := h.profiles.FindByBabyID(ctx, q.BabyID)
		if err != nil {
			return nil, fmt.Errorf("load sleep profile for timezone: %w", err)
		}
		tz = profile.Timezone()
	}

	dateRange, err := periodToDateRange(q.Period, tz)
	if err != nil {
		return nil, err
	}

	sessions, err := h.sessions.FindByBabyIDAndDateRange(ctx, q.BabyID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("get sleep history: %w", err)
	}

	if sessions == nil {
		sessions = []SleepSession{}
	}

	return sessions, nil
}

// periodToDateRange converts a period string and IANA timezone name into a
// DateRange expressed in UTC, using the local calendar in the given timezone.
//
//   - "today" → local midnight today .. local midnight tomorrow
//   - "7d"    → local midnight 7 days ago .. now (UTC)
//   - "14d"   → local midnight 14 days ago .. now (UTC)
func periodToDateRange(period, timezone string) (DateRange, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return DateRange{}, ErrInvalidTimezone
	}

	now := time.Now().UTC()
	localNow := now.In(loc)

	switch period {
	case "today":
		// Midnight at start of today in local time.
		startLocal := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
		endLocal := startLocal.AddDate(0, 0, 1)
		return NewDateRange(startLocal.UTC(), endLocal.UTC())

	case "7d":
		// Midnight 7 days ago in local time.
		daysAgoLocal := localNow.AddDate(0, 0, -7)
		startLocal := time.Date(daysAgoLocal.Year(), daysAgoLocal.Month(), daysAgoLocal.Day(), 0, 0, 0, 0, loc)
		return NewDateRange(startLocal.UTC(), now)

	case "14d":
		// Midnight 14 days ago in local time.
		daysAgoLocal := localNow.AddDate(0, 0, -14)
		startLocal := time.Date(daysAgoLocal.Year(), daysAgoLocal.Month(), daysAgoLocal.Day(), 0, 0, 0, 0, loc)
		return NewDateRange(startLocal.UTC(), now)

	default:
		return DateRange{}, ErrInvalidSleepHistoryPeriod
	}
}
