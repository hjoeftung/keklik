package sleep

import (
	"testing"
	"time"
)

func mustNightWindow(t *testing.T, startHour, startMinute, endHour, endMinute int) NightWindow {
	t.Helper()

	start, err := NewLocalTime(startHour, startMinute)
	if err != nil {
		t.Fatalf("NewLocalTime start: %v", err)
	}

	end, err := NewLocalTime(endHour, endMinute)
	if err != nil {
		t.Fatalf("NewLocalTime end: %v", err)
	}

	window, err := NewNightWindow(
		NightWindowID("night-window-1"),
		BabyID("baby-1"),
		start,
		end,
		time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		nil,
	)
	if err != nil {
		t.Fatalf("NewNightWindow: %v", err)
	}

	return window
}
