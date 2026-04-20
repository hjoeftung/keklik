package sleep

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptySleepSessionID          = errors.New("sleep session id must not be empty")
	ErrEmptyBabyID                  = errors.New("baby id must not be empty")
	ErrEmptyFamilyMemberID          = errors.New("family member id must not be empty")
	ErrZeroSleepSessionStart        = errors.New("sleep session start must not be zero")
	ErrMissingSleepSessionEdit      = errors.New("sleep session edit requires started_at or stopped_at")
	ErrSleepSessionAlreadyStopped   = errors.New("sleep session already stopped")
	ErrInvalidSleepSessionStop      = errors.New("sleep session stop must not be before start")
	ErrUnknownSleepClassification   = errors.New("unknown sleep classification")
	ErrInvalidSleepSessionDateRange = errors.New("sleep session date range is invalid")
	ErrActiveSleepSessionExists     = errors.New("active sleep session already exists for this baby")
	ErrInvalidSleepHistoryPeriod    = errors.New("period must be one of: today, 7d, 14d")
	ErrEffectiveFromTooOld          = errors.New("effective_from must not be earlier than 30 days ago")

	ErrInvalidTimezone    = errors.New("invalid timezone")
	ErrInvalidLocalTime   = errors.New("invalid local time")
	ErrInvalidNightWindow = errors.New("invalid night window")
)

type SleepSessionID string

type BabyID string

type FamilyMemberID string

type SleepClassification string

const (
	SleepClassificationUnknown SleepClassification = ""
	SleepClassificationNap     SleepClassification = "nap"
	SleepClassificationNight   SleepClassification = "night"
)

type SleepSession struct {
	id                        SleepSessionID
	babyID                    BabyID
	createdByMemberID         FamilyMemberID
	startedAt                 time.Time
	stoppedAt                 *time.Time
	classification            SleepClassification
	classifiedWithNightWindow *NightWindow
}

type DateRange struct {
	start time.Time
	end   time.Time
}

type LocalTime struct {
	hour   int
	minute int
}

type NightWindow struct {
	start LocalTime
	end   LocalTime
}

type SleepProfile struct {
	babyID      BabyID
	timezone    string
	nightWindow NightWindow
}

type SleepSessionRepository interface {
	Save(ctx context.Context, session SleepSession) error
	SaveAll(ctx context.Context, sessions []SleepSession) error
	FindByID(ctx context.Context, id SleepSessionID) (SleepSession, error)
}

// Transactor executes a function within a database transaction, committing on
// success and rolling back on any error.
type Transactor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type EditableSleepSessionRepository interface {
	SleepSessionRepository
	DeleteByID(ctx context.Context, id SleepSessionID) error
}

type ActiveSleepSessionRepository interface {
	FindActiveByBabyID(ctx context.Context, babyID BabyID) (SleepSession, error)
}

type SleepSessionHistoryRepository interface {
	FindByBabyIDAndDateRange(ctx context.Context, babyID BabyID, dateRange DateRange) ([]SleepSession, error)
}

type SleepProfileRepository interface {
	Save(ctx context.Context, profile SleepProfile) error
	FindByBabyID(ctx context.Context, babyID BabyID) (SleepProfile, error)
}

// CompletedSleepSessionsSinceRepository finds completed (stopped) sessions from a given point in time.
type CompletedSleepSessionsSinceRepository interface {
	FindCompletedByBabyIDSince(ctx context.Context, babyID BabyID, since time.Time) ([]SleepSession, error)
}

// SleepElapsedTimeRepository finds the most recent sessions for elapsed-time calculations.
// Both methods return an AppError with CodeNotFound when no matching session exists.
type SleepElapsedTimeRepository interface {
	FindMostRecentByBabyID(ctx context.Context, babyID BabyID) (SleepSession, error)
	FindMostRecentCompletedByBabyID(ctx context.Context, babyID BabyID) (SleepSession, error)
}

func NewSleepSession(id SleepSessionID, babyID BabyID, createdByMemberID FamilyMemberID, startedAt time.Time) (SleepSession, error) {
	return newSleepSession(id, babyID, createdByMemberID, startedAt, nil, SleepClassificationUnknown, nil)
}

func NewCompletedSleepSession(id SleepSessionID, babyID BabyID, createdByMemberID FamilyMemberID, startedAt time.Time, stoppedAt time.Time, classification SleepClassification, classifiedWith *NightWindow) (SleepSession, error) {
	return newSleepSession(id, babyID, createdByMemberID, startedAt, &stoppedAt, classification, classifiedWith)
}

func NewDateRange(start time.Time, end time.Time) (DateRange, error) {
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return DateRange{}, ErrInvalidSleepSessionDateRange
	}

	return DateRange{start: start, end: end}, nil
}

func NewLocalTime(hour int, minute int) (LocalTime, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return LocalTime{}, ErrInvalidLocalTime
	}

	return LocalTime{hour: hour, minute: minute}, nil
}

func NewNightWindow(start LocalTime, end LocalTime) (NightWindow, error) {
	if start == end {
		return NightWindow{}, ErrInvalidNightWindow
	}

	return NightWindow{start: start, end: end}, nil
}

func NewSleepProfile(babyID BabyID, timezone string, nightWindow NightWindow) (SleepProfile, error) {
	trimmedBabyID := strings.TrimSpace(string(babyID))
	if trimmedBabyID == "" {
		return SleepProfile{}, ErrEmptyBabyID
	}

	trimmedTimezone := strings.TrimSpace(timezone)
	if _, err := time.LoadLocation(trimmedTimezone); err != nil {
		return SleepProfile{}, ErrInvalidTimezone
	}

	return SleepProfile{
		babyID:      BabyID(trimmedBabyID),
		timezone:    trimmedTimezone,
		nightWindow: nightWindow,
	}, nil
}

func (s SleepSession) ID() SleepSessionID {
	return s.id
}

func (s SleepSession) BabyID() BabyID {
	return s.babyID
}

func (s SleepSession) CreatedByMemberID() FamilyMemberID {
	return s.createdByMemberID
}

func (s SleepSession) StartedAt() time.Time {
	return s.startedAt
}

func (s SleepSession) StoppedAt() (time.Time, bool) {
	if s.stoppedAt == nil {
		return time.Time{}, false
	}

	return *s.stoppedAt, true
}

func (s SleepSession) Classification() SleepClassification {
	return s.classification
}

func (s SleepSession) ClassifiedWithNightWindow() *NightWindow {
	return s.classifiedWithNightWindow
}

func (s SleepSession) IsActive() bool {
	return s.stoppedAt == nil
}

func (s SleepSession) Duration() (time.Duration, bool) {
	if s.stoppedAt == nil {
		return 0, false
	}

	return s.stoppedAt.Sub(s.startedAt), true
}

func (s *SleepSession) Stop(stoppedAt time.Time, classification SleepClassification, classifiedWith NightWindow) error {
	if s.stoppedAt != nil {
		return ErrSleepSessionAlreadyStopped
	}

	normalizedClassification, err := normalizeClassification(classification)
	if err != nil {
		return err
	}

	if stoppedAt.Before(s.startedAt) {
		return ErrInvalidSleepSessionStop
	}

	s.stoppedAt = &stoppedAt
	s.classification = normalizedClassification
	s.classifiedWithNightWindow = &classifiedWith

	return nil
}

func (r DateRange) Start() time.Time {
	return r.start
}

func (r DateRange) End() time.Time {
	return r.end
}

func (t LocalTime) Hour() int {
	return t.hour
}

func (t LocalTime) Minute() int {
	return t.minute
}

func (n NightWindow) Start() LocalTime {
	return n.start
}

func (n NightWindow) End() LocalTime {
	return n.end
}

func (p SleepProfile) BabyID() BabyID {
	return p.babyID
}

func (p SleepProfile) Timezone() string {
	return p.timezone
}

func (p SleepProfile) NightWindow() NightWindow {
	return p.nightWindow
}

func newSleepSession(id SleepSessionID, babyID BabyID, createdByMemberID FamilyMemberID, startedAt time.Time, stoppedAt *time.Time, classification SleepClassification, classifiedWith *NightWindow) (SleepSession, error) {
	trimmedID := strings.TrimSpace(string(id))
	if trimmedID == "" {
		return SleepSession{}, ErrEmptySleepSessionID
	}

	trimmedBabyID := strings.TrimSpace(string(babyID))
	if trimmedBabyID == "" {
		return SleepSession{}, ErrEmptyBabyID
	}

	trimmedMemberID := strings.TrimSpace(string(createdByMemberID))
	if trimmedMemberID == "" {
		return SleepSession{}, ErrEmptyFamilyMemberID
	}

	if startedAt.IsZero() {
		return SleepSession{}, ErrZeroSleepSessionStart
	}

	normalizedClassification, err := normalizeClassification(classification)
	if err != nil {
		return SleepSession{}, err
	}

	if stoppedAt != nil && stoppedAt.Before(startedAt) {
		return SleepSession{}, ErrInvalidSleepSessionStop
	}

	if stoppedAt == nil {
		normalizedClassification = SleepClassificationUnknown
	}

	session := SleepSession{
		id:                        SleepSessionID(trimmedID),
		babyID:                    BabyID(trimmedBabyID),
		createdByMemberID:         FamilyMemberID(trimmedMemberID),
		startedAt:                 startedAt,
		classification:            normalizedClassification,
		classifiedWithNightWindow: classifiedWith,
	}

	if stoppedAt != nil {
		stoppedAtCopy := *stoppedAt
		session.stoppedAt = &stoppedAtCopy
	}

	return session, nil
}

func normalizeClassification(classification SleepClassification) (SleepClassification, error) {
	switch classification {
	case SleepClassificationUnknown, SleepClassificationNap, SleepClassificationNight:
		return classification, nil
	default:
		return SleepClassificationUnknown, ErrUnknownSleepClassification
	}
}
