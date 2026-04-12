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
	ErrZeroSleepSessionStart        = errors.New("sleep session start must not be zero")
	ErrSleepSessionAlreadyStopped   = errors.New("sleep session already stopped")
	ErrInvalidSleepSessionStop      = errors.New("sleep session stop must not be before start")
	ErrUnknownSleepClassification   = errors.New("unknown sleep classification")
	ErrInvalidSleepSessionDateRange = errors.New("sleep session date range is invalid")
)

type SleepSessionID string

type BabyID string

type SleepClassification string

const (
	SleepClassificationUnknown SleepClassification = ""
	SleepClassificationNap     SleepClassification = "nap"
	SleepClassificationNight   SleepClassification = "night"
)

type SleepSession struct {
	id             SleepSessionID
	babyID         BabyID
	startedAt      time.Time
	stoppedAt      *time.Time
	classification SleepClassification
}

type DateRange struct {
	start time.Time
	end   time.Time
}

type SleepSessionRepository interface {
	Save(ctx context.Context, session SleepSession) error
	FindByID(ctx context.Context, id SleepSessionID) (SleepSession, error)
}

type ActiveSleepSessionRepository interface {
	FindActiveByBabyID(ctx context.Context, babyID BabyID) (SleepSession, error)
}

type SleepSessionHistoryRepository interface {
	FindByBabyIDAndDateRange(ctx context.Context, babyID BabyID, dateRange DateRange) ([]SleepSession, error)
}

func NewSleepSession(id SleepSessionID, babyID BabyID, startedAt time.Time) (SleepSession, error) {
	return newSleepSession(id, babyID, startedAt, nil, SleepClassificationUnknown)
}

func NewCompletedSleepSession(id SleepSessionID, babyID BabyID, startedAt time.Time, stoppedAt time.Time, classification SleepClassification) (SleepSession, error) {
	return newSleepSession(id, babyID, startedAt, &stoppedAt, classification)
}

func NewDateRange(start time.Time, end time.Time) (DateRange, error) {
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return DateRange{}, ErrInvalidSleepSessionDateRange
	}

	return DateRange{start: start, end: end}, nil
}

func (s SleepSession) ID() SleepSessionID {
	return s.id
}

func (s SleepSession) BabyID() BabyID {
	return s.babyID
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

func (s SleepSession) IsActive() bool {
	return s.stoppedAt == nil
}

func (s SleepSession) Duration() (time.Duration, bool) {
	if s.stoppedAt == nil {
		return 0, false
	}

	return s.stoppedAt.Sub(s.startedAt), true
}

func (s *SleepSession) Stop(stoppedAt time.Time, classification SleepClassification) error {
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

	return nil
}

func (r DateRange) Start() time.Time {
	return r.start
}

func (r DateRange) End() time.Time {
	return r.end
}

func newSleepSession(id SleepSessionID, babyID BabyID, startedAt time.Time, stoppedAt *time.Time, classification SleepClassification) (SleepSession, error) {
	trimmedID := strings.TrimSpace(string(id))
	if trimmedID == "" {
		return SleepSession{}, ErrEmptySleepSessionID
	}

	trimmedBabyID := strings.TrimSpace(string(babyID))
	if trimmedBabyID == "" {
		return SleepSession{}, ErrEmptyBabyID
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
		id:             SleepSessionID(trimmedID),
		babyID:         BabyID(trimmedBabyID),
		startedAt:      startedAt,
		classification: normalizedClassification,
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
