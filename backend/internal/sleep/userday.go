package sleep

import (
	"errors"
	"slices"
	"time"
)

var ErrSleepOutOfDayBound = errors.New("Sleep session does not belong to the UserDay")

func NewUserDay(
	dayStart time.Time,
	dayEnd time.Time,
	nightWindow NightWindow,
	loc *time.Location,
) *UserDay {
	return &UserDay{
		dayStart: dayStart,
		dayEnd: dayEnd,
		nightWindow: nightWindow,
		loc: loc,
		wokeAt: dayStart,
		nightStartedAt: dayEnd,
	}
}

// UserDay models one parent-perceived day: from the last wake-up after night
// sleep to the start of the next night sleep.
type UserDay struct {
	dayStart time.Time
	dayEnd time.Time
	nightWindow NightWindow
	loc *time.Location
	wokeAt time.Time
	nightStartedAt time.Time

	// Sessions whose StartedAt falls within tonight's night window.
	// May include an active session (baby currently sleeping).
	nightSessions []SleepSession

	// Nap sessions between WokeAt and NightStartedAt (or dayEnd).
	// Always empty when WokeAt is nil.
	naps []SleepSession
}

func (d *UserDay) AddSleep(sleep SleepSession) error {
	var classification SleepClassification
	if sleep.IsActive() {
		classification = classifyActive(sleep, d.loc, d.nightWindow)
	} else {
		classification = classifyFromLocation(sleep, d.loc, d.nightWindow)
	}

	if classification == SleepClassificationNight {
		return d.AddNightSleep(sleep)
	} else {
		return d.AddNap(sleep)
	}
}

func (d *UserDay) AddNightSleep(sleep SleepSession) error {
	stoppedAt, ok := sleep.StoppedAt()
	if !ok {
		completedSleep, err := NewCompletedSleepSession(sleep.id, sleep.babyID, sleep.createdByMemberID, sleep.startedAt, time.Now())
		if err != nil {
			return err
		}
		return d.AddNightSleep(completedSleep)
	}

	// A UserDay is anchored by wake-up, not by the previous night's sleep total.
	// So a night session contributes to NightDuration only when it belongs to
	// this day's night window. A night session from yesterday may still define
	// today's wake-up anchor, but it must not be counted as today's night sleep.
	if d.belongsToThisDayNightWindow(sleep) {
		d.nightSessions = append(d.nightSessions, sleep)
		if d.nightStartedAt.After(sleep.startedAt) {
			d.nightStartedAt = sleep.startedAt
		}
		return nil
	}

	if d.wakesIntoThisDay(sleep, stoppedAt) && stoppedAt.After(d.wokeAt) {
		d.wokeAt = stoppedAt
	}

	return ErrSleepOutOfDayBound
}

func (d *UserDay) AddNap(sleep SleepSession) error {
	_, ok := sleep.StoppedAt()
	if !ok {
		completedSleep, err := NewCompletedSleepSession(sleep.id, sleep.babyID, sleep.createdByMemberID, sleep.startedAt, time.Now())
		if err != nil {
			return err
		}
		return d.AddNap(completedSleep)
	}

	if intersects(sleep, d.dayStart, d.dayEnd) {
		d.naps = append(d.naps, sleep)
	}
	return nil
}

func (d *UserDay) NightDuration(now time.Time) time.Duration {
	var total time.Duration
	for _, s := range d.nightSessions {
		stoppedAt, ok := s.StoppedAt()
		if !ok {
			stoppedAt = now
		}
		total += stoppedAt.Sub(s.StartedAt())
	}
	return total
}

func (d *UserDay) NapDuration(now time.Time) time.Duration {
	var total time.Duration
	for _, s := range d.naps {
		stoppedAt, ok := s.StoppedAt()
		if !ok {
			stoppedAt = now
		}
		total += stoppedAt.Sub(s.StartedAt())
	}
	return total
}

func (d *UserDay) ActiveDuration(now time.Time) time.Duration {
	finish := d.nightStartedAt
	if now.Before(finish) {
		finish = now
	}
	return finish.Sub(d.wokeAt) - d.NapDuration(now)
}

func (d *UserDay) NightStartedAt() (time.Time, bool) {
	if len(d.nightSessions) == 0 {
		return time.Time{}, false
	}
	slices.SortFunc(d.nightSessions, func(a, b SleepSession) int {
		if a.StartedAt().Before(b.StartedAt()) {return -1}
		if a.StartedAt().After(b.StartedAt()) {return 1}
		return 0
	})
	return d.nightSessions[0].StartedAt(), true
}

func (d *UserDay) NightFinishedAt() (time.Time, bool) {
	if len(d.nightSessions) == 0 {
		return time.Time{}, false
	}
	slices.SortFunc(d.nightSessions, func(a, b SleepSession) int {
		if a.StartedAt().Before(b.StartedAt()) {return -1}
		if a.StartedAt().After(b.StartedAt()) {return 1}
		return 0
	})
	stoppedAt, ok := d.nightSessions[len(d.nightSessions)-1].StoppedAt()
	if !ok {
		return time.Time{}, false
	}
	return stoppedAt, true
}

func (d *UserDay) belongsToThisDayNightWindow(sleep SleepSession) bool {
	nwStart, nwEnd := nightWindowBounds(d.dayStart, d.nightWindow, d.loc)
	return timeWithinRangeStartInclusive(sleep.StartedAt(), d.dayStart, d.dayEnd) &&
		sleepOverlapsRange(sleep, nwStart, nwEnd)
}

func (d *UserDay) wakesIntoThisDay(sleep SleepSession, stoppedAt time.Time) bool {
	previousNightStart, previousNightEnd := nightWindowBounds(d.dayStart.AddDate(0, 0, -1), d.nightWindow, d.loc)
	return sleep.StartedAt().Before(d.dayStart) &&
		sleepOverlapsRange(sleep, previousNightStart, previousNightEnd) &&
		timeWithinRangeStartInclusive(stoppedAt, d.dayStart, d.dayEnd) &&
		stoppedAt.Before(d.dayEnd)
}


// buildUserDays constructs one UserDay per calendar date in [today, today−(days−1)],
// index 0 = today, in reverse-chronological order.
//
// classified contains all sessions pre-classified by the caller; no
// classification logic is performed here.
func buildUserDays(
	sleeps []SleepSession,
	windows []NightWindow,
	loc *time.Location,
	now time.Time,
	days int,
) ([]*UserDay, error) {
	localNow := now.In(loc)
	today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)

	userDays := make([]*UserDay, days)
	for i := 0; i < days; i++ {
		d := today.AddDate(0, 0, -i)
		dayStart := d.UTC()
		dayEnd := d.Add(24 * time.Hour).UTC()
		nw, ok := findWindowAt(windows, dayStart)
		if !ok {

		}
		userDays[i] = NewUserDay(dayStart, dayEnd, nw, loc)
	}

	userDayIdx := 0
	for i := 0; i<len(sleeps) && userDayIdx < len(userDays); i++ {
		err := userDays[userDayIdx].AddSleep(sleeps[i])
		if err != nil {
			if err == ErrSleepOutOfDayBound {
				userDayIdx++
			} else {
				return []*UserDay{}, err
			}
		}
	}

	return userDays, nil
}

func intersects(sleep SleepSession, start time.Time, end time.Time) bool {
	stoppedAt, ok := sleep.StoppedAt()
	if !ok {
		stoppedAt = time.Now()
	}
	return startedWithin(sleep, start, end) || (stoppedAt.After(start) && stoppedAt.Before(end))
}

func startedWithin(sleep SleepSession, start time.Time, end time.Time) bool {
	return sleep.StartedAt().After(start) && sleep.StartedAt().Before(end) 
}

func sleepOverlapsRange(sleep SleepSession, start time.Time, end time.Time) bool {
	stoppedAt, ok := sleep.StoppedAt()
	if !ok {
		return sleep.StartedAt().Before(end)
	}
	return sleep.StartedAt().Before(end) && stoppedAt.After(start)
}

func timeWithinRangeStartInclusive(t time.Time, start time.Time, end time.Time) bool {
	return !t.Before(start) && t.Before(end)
}
