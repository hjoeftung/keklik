# Sleep & Statistics Screens — Detailed Requirements

## General

- The app is mobile-first. Buttons must be large and touch-friendly. Navigation must be adapted for portrait mobile.
- Show loading skeletons on all screens while data has not yet been received from the backend.
- Pull-to-refresh triggers a manual re-fetch on all screens.

---

## Sleep Screen

### Primary actions

Three actions must be immediately accessible without additional navigation:

1. **Start / Stop session** — a large prominent button. Tapping it starts a session if none is active, or stops the currently active session.
2. **Edit active session start time** — visible only when a session is active. Allows the user to correct the session's start time via a time picker. Not available when no session is running.
3. **Log past session** — always visible. Opens a form with two time pickers (start and end). Allows the user to record a session that was not tracked in real time. The form must validate that end time is after start time and that the session does not overlap an existing session.

### Status display

- **When a session is active:** display a live duration counter (time elapsed since session start). Updates every minute.
- **When no session is active:** display the time elapsed since the last session ended ("active for X h Y min"). Updates every minute.

---

## Statistics Screen (Timeline)

### Data loading

- Fetch all session data from the backend optimistically on app load (before the user opens the Statistics screen).
- Cache the result for the duration of the app session.
- Display skeletons on each tab while the cache has not yet been populated.
- Pull-to-refresh re-fetches and updates the cache.
- Session classification (night sleep vs. nap) is provided by the backend and must not be computed on the client.

### Tabs

Three top-level tabs: **Today**, **Week**, **Summary**.

---

### Today tab

#### Summary header

Displayed above the diary view. Three values, received from the backend:

- **Total Sleep** — total duration of all completed sleep sessions in the diary window.
- **Total Nap** — total duration of sessions classified as naps.
- **Total Active** — total time in the diary window not covered by any sleep session.

#### Diary view

A vertical timeline representing a fixed 24-hour window:

- **Start:** end of the baby's night window minus 2 hours. Example: if the night window ends at 08:00, the diary starts at 06:00 of the current date.
- **End:** the same clock time on the following date (06:00 the next day in the example above).
- **Hour labels** on the left side at each full hour.

Sleep sessions are rendered as **horizontal bars**:

- Each bar spans the full width of the timeline column.
- Height is proportional to the session's duration relative to the total window height.
- Color must be clearly distinct from the background. Night sleep and nap sessions may use different colors (based on backend classification).
- Each bar displays its **duration** as a text label inside or adjacent to the bar.
- Tapping a bar opens a **detail panel** for that session. The detail panel shows: start time, end time, duration, classification. It also provides an **Edit** action that lets the user modify start and end times.

Active (not yet finished) sessions are **not shown** in the diary view. There is no indicator of a currently active session on this tab.

---

### Week tab

A grid showing 7 days of sleep sessions.

- **Columns:** one per day, ordered left to right from 6 days ago to today (7 columns total).
- **Rows:** hours, using the same 24-hour window definition as the Today tab (same start/end logic). Time increases top to bottom.
- Column headers show the date (e.g. "Mon 21").
- Sleep sessions are rendered as colored blocks within the appropriate column, positioned vertically by time of day and sized by duration. Night sleep and nap sessions use different colors.
- No tap interaction required on the week grid (display only).

---

### Summary tab

#### Period selector

Four sub-tabs: **7d**, **14d**, **30d**, **90d**.

Below the sub-tabs, display the start and end dates of the selected period. Example: if 7d is selected and today is 23 April 2026, display "17 April 2026 – 23 April 2026".

#### Content

Three daily averages for the selected period, received pre-calculated from the backend:

- **Avg Active** — average active time per day.
- **Avg Sleep** — average total sleep per day.
- **Avg Nap** — average nap time per day.

No charts or visualizations in this tab for now — numbers only.
