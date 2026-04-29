package sleep

import "errors"

var (
	ErrEmptySleepSessionID          = errors.New("sleep session id must not be empty")
	ErrEmptyBabyID                  = errors.New("baby id must not be empty")
	ErrEmptyFamilyMemberID          = errors.New("family member id must not be empty")
	ErrZeroSleepSessionStart        = errors.New("sleep session start must not be zero")
	ErrMissingSleepSessionEdit      = errors.New("sleep session edit requires started_at or stopped_at")
	ErrSleepSessionAlreadyStopped   = errors.New("sleep session already stopped")
	ErrInvalidSleepSessionStop      = errors.New("sleep session stop must not be before start")
	ErrInvalidSleepSessionDateRange = errors.New("sleep session date range is invalid")
	ErrActiveSleepSessionExists     = errors.New("active sleep session already exists for this baby")
	ErrSleepSessionOverlap          = errors.New("sleep session overlaps an existing session")
	ErrSleepSessionConflict         = errors.New("sleep session was modified concurrently")
	ErrInvalidSleepHistoryPeriod    = errors.New("period must be one of: today, 7d, 14d")
	ErrEffectiveFromTooOld          = errors.New("effective_from must not be earlier than 30 days ago")

	ErrInvalidTimezone    = errors.New("invalid timezone")
	ErrInvalidLocalTime   = errors.New("invalid local time")
	ErrInvalidNightWindow = errors.New("invalid night window")
	ErrEmptyNightWindowID = errors.New("night window id must not be empty")
)
