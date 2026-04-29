package sleep

import (
	"strings"
	"time"
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
	id                SleepSessionID
	babyID            BabyID
	createdByMemberID FamilyMemberID
	startedAt         time.Time
	stoppedAt         *time.Time
	version           int
}

type DateRange struct {
	start time.Time
	end   time.Time
}

func NewSleepSession(id SleepSessionID, babyID BabyID, createdByMemberID FamilyMemberID, startedAt time.Time) (SleepSession, error) {
	return newSleepSession(id, babyID, createdByMemberID, startedAt, nil)
}

func NewCompletedSleepSession(id SleepSessionID, babyID BabyID, createdByMemberID FamilyMemberID, startedAt time.Time, stoppedAt time.Time) (SleepSession, error) {
	return newSleepSession(id, babyID, createdByMemberID, startedAt, &stoppedAt)
}

// RestoreSleepSession recreates a SleepSession from persisted data, including
// the version counter required for optimistic concurrency control.
func RestoreSleepSession(id SleepSessionID, babyID BabyID, createdByMemberID FamilyMemberID, startedAt time.Time, stoppedAt *time.Time, version int) (SleepSession, error) {
	s, err := newSleepSession(id, babyID, createdByMemberID, startedAt, stoppedAt)
	if err != nil {
		return SleepSession{}, err
	}
	s.version = version
	return s, nil
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

func (s SleepSession) Version() int {
	return s.version
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

func (s SleepSession) IsActive() bool {
	return s.stoppedAt == nil
}

func (s SleepSession) Duration() (time.Duration, bool) {
	if s.stoppedAt == nil {
		return 0, false
	}

	return s.stoppedAt.Sub(s.startedAt), true
}

func (s *SleepSession) Stop(stoppedAt time.Time) error {
	if s.stoppedAt != nil {
		return ErrSleepSessionAlreadyStopped
	}

	if stoppedAt.Before(s.startedAt) {
		return ErrInvalidSleepSessionStop
	}

	s.stoppedAt = &stoppedAt

	return nil
}

func (r DateRange) Start() time.Time {
	return r.start
}

func (r DateRange) End() time.Time {
	return r.end
}

func newSleepSession(id SleepSessionID, babyID BabyID, createdByMemberID FamilyMemberID, startedAt time.Time, stoppedAt *time.Time) (SleepSession, error) {
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

	if stoppedAt != nil && stoppedAt.Before(startedAt) {
		return SleepSession{}, ErrInvalidSleepSessionStop
	}

	session := SleepSession{
		id:                SleepSessionID(trimmedID),
		babyID:            BabyID(trimmedBabyID),
		createdByMemberID: FamilyMemberID(trimmedMemberID),
		startedAt:         startedAt,
	}

	if stoppedAt != nil {
		stoppedAtCopy := *stoppedAt
		session.stoppedAt = &stoppedAtCopy
	}

	return session, nil
}
