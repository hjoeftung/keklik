package sleep

import "time"

// Classify derives the sleep classification for a completed session by
// calculating how much of the session's duration falls inside the family's
// night window (expressed in the family's local timezone). If more than half
// of the duration is inside the night window the session is classified as
// night sleep; otherwise it is a nap.
//
// DST transitions are handled naturally because all window boundary
// calculations use time.In to convert UTC instants to local wall-clock time —
// Go's time package adjusts for clock changes automatically.
//
// Windows that cross midnight (e.g. 21:00–06:00) are supported: if the local
// wall-clock time of a boundary would land on the following calendar day it is
// shifted forward by 24 hours relative to the night-window start anchor.
func Classify(session SleepSession, timezone string, nightWindow NightWindow) (SleepClassification, error) {
	stoppedAt, ok := session.StoppedAt()
	if !ok {
		return SleepClassificationUnknown, nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return SleepClassificationUnknown, ErrInvalidTimezone
	}

	overlap := nightOverlap(session.StartedAt(), stoppedAt, loc, nightWindow)

	duration := stoppedAt.Sub(session.StartedAt())
	if duration == 0 {
		// Zero-duration session: classify by whether start falls in night window.
		if overlap > 0 {
			return SleepClassificationNight, nil
		}
		return SleepClassificationNap, nil
	}

	if overlap*2 > duration {
		return SleepClassificationNight, nil
	}
	return SleepClassificationNap, nil
}

// nightOverlap returns how much of [sessionStart, sessionEnd) overlaps with
// the night window on each calendar day covered by the session. It iterates
// day by day so that multi-night sessions (or sessions spanning a DST
// boundary) are handled correctly.
func nightOverlap(sessionStart, sessionEnd time.Time, loc *time.Location, nw NightWindow) time.Duration {
	var total time.Duration

	// Anchor to the start of the local calendar day that contains sessionStart.
	localStart := sessionStart.In(loc)
	dayStart := time.Date(
		localStart.Year(), localStart.Month(), localStart.Day(),
		0, 0, 0, 0, loc,
	)

	for {
		windowStart, windowEnd := nightWindowBounds(dayStart, nw, loc)

		// Intersect [windowStart, windowEnd) with [sessionStart, sessionEnd).
		overlapStart := maxTime(windowStart, sessionStart)
		overlapEnd := minTime(windowEnd, sessionEnd)
		if overlapEnd.After(overlapStart) {
			total += overlapEnd.Sub(overlapStart)
		}

		// Advance to next calendar day.
		dayStart = dayStart.AddDate(0, 0, 1)
		if !dayStart.Before(sessionEnd) {
			break
		}
	}

	return total
}

// nightWindowBounds returns the absolute [start, end) instants for the night
// window anchored on the given local calendar day. When the window crosses
// midnight (end hour:minute <= start hour:minute) the end is placed on the
// following calendar day.
func nightWindowBounds(dayStart time.Time, nw NightWindow, loc *time.Location) (time.Time, time.Time) {
	ws := time.Date(
		dayStart.Year(), dayStart.Month(), dayStart.Day(),
		nw.Start().Hour(), nw.Start().Minute(), 0, 0, loc,
	)
	we := time.Date(
		dayStart.Year(), dayStart.Month(), dayStart.Day(),
		nw.End().Hour(), nw.End().Minute(), 0, 0, loc,
	)

	// Window crosses midnight when the end wall-clock is not after the start.
	if !we.After(ws) {
		we = we.AddDate(0, 0, 1)
	}

	return ws, we
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
