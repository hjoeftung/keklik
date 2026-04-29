package sleep

import (
	"strings"
	"time"
)

type NightWindowID string

type LocalTime struct {
	hour   int
	minute int
}

type NightWindow struct {
	id            NightWindowID
	babyID        BabyID
	start         LocalTime
	end           LocalTime
	effectiveFrom time.Time
	effectiveTo   *time.Time
}

func NewLocalTime(hour int, minute int) (LocalTime, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return LocalTime{}, ErrInvalidLocalTime
	}

	return LocalTime{hour: hour, minute: minute}, nil
}

func NewNightWindow(id NightWindowID, babyID BabyID, start LocalTime, end LocalTime, effectiveFrom time.Time, effectiveTo *time.Time) (NightWindow, error) {
	trimmedID := strings.TrimSpace(string(id))
	if trimmedID == "" {
		return NightWindow{}, ErrEmptyNightWindowID
	}

	trimmedBabyID := strings.TrimSpace(string(babyID))
	if trimmedBabyID == "" {
		return NightWindow{}, ErrEmptyBabyID
	}

	if start == end {
		return NightWindow{}, ErrInvalidNightWindow
	}

	if effectiveTo != nil && !effectiveTo.After(effectiveFrom) {
		return NightWindow{}, ErrInvalidNightWindow
	}

	return NightWindow{
		id:            NightWindowID(trimmedID),
		babyID:        BabyID(trimmedBabyID),
		start:         start,
		end:           end,
		effectiveFrom: effectiveFrom,
		effectiveTo:   effectiveTo,
	}, nil
}

// FindWindowForSession returns the NightWindow whose effective range covers
// session.StartedAt. windows must be ordered by effective_from ASC.
func FindWindowForSession(windows []NightWindow, session SleepSession) (NightWindow, bool) {
	sessionStart := session.StartedAt()
	var best NightWindow
	found := false
	for _, w := range windows {
		if !w.effectiveFrom.After(sessionStart) {
			if w.effectiveTo == nil || w.effectiveTo.After(sessionStart) {
				best = w
				found = true
			}
		}
	}
	return best, found
}

func (t LocalTime) Hour() int {
	return t.hour
}

func (t LocalTime) Minute() int {
	return t.minute
}

func (n NightWindow) ID() NightWindowID {
	return n.id
}

func (n NightWindow) BabyID() BabyID {
	return n.babyID
}

func (n NightWindow) Start() LocalTime {
	return n.start
}

func (n NightWindow) End() LocalTime {
	return n.end
}

func (n NightWindow) EffectiveFrom() time.Time {
	return n.effectiveFrom
}

func (n NightWindow) EffectiveTo() *time.Time {
	return n.effectiveTo
}
