## Story 5: Dashboard

### TASK-057: Build dashboard sleep state control (Start/Stop button)
- Size: `S`
- Goal: Show the baby's current sleep state and let the user toggle it with a prominent button.
- Scope:
  - `src/screens/Dashboard/SleepControl.tsx`
  - Fetch dashboard summary from `GET /babies/{baby_id}/sleep-sessions/summary` on mount and after each toggle
  - Awake state: show large "Start sleep" button → calls `POST /babies/{baby_id}/sleep-sessions`
  - Sleeping state: show large "Stop sleep" button → calls `DELETE /babies/{baby_id}/sleep-sessions/active`
  - Show loading state during the API call; disable button to prevent double-submit
  - Surface API errors inline
- Dependencies:
  - [TASK-048](#task-048-implement-api-client-with-auth-header-error-normalization-and-401-redirect)
  - [TASK-056](#task-056-implement-persistent-navigation-bar)
- Acceptance criteria:
  - Button label and action reflect actual backend state on load
  - Double-tapping before response does not send two requests
  - API error is shown without navigating away

### TASK-058: Implement real-time elapsed timer for active session
- Size: `S`
- Goal: Show how long the baby has been sleeping, updating every second client-side.
- Scope:
  - `src/components/ElapsedTimer.tsx` — accepts `startedAt: string` ISO timestamp, displays elapsed time formatted as `formatDuration`
  - Uses `setInterval` with 1 s interval; drives display from `(Date.now() - startedAt)` delta
  - Cleans up interval on unmount or when session ends
  - If `active_session` is null, component renders nothing (parent decides placeholder)
- Dependencies:
  - [TASK-049](#task-049-implement-duration-and-timezone-formatting-utilities)
  - [TASK-057](#task-057-build-dashboard-sleep-state-control-startstop-button)
- Acceptance criteria:
  - Timer increments every second without drift accumulation
  - Timer stops and unmounts cleanly when Stop is pressed

### TASK-059: Display time-since-last-event metrics on dashboard
- Size: `S`
- Goal: Show elapsed time since the last sleep started and since the last awakening.
- Scope:
  - `src/screens/Dashboard/SinceLastPanel.tsx`
  - Reads `since_last` from the summary response
  - Formats both values using `formatDuration`
  - When a field is `null`, shows a clear placeholder (e.g. "No data yet") rather than zero or blank
- Dependencies:
  - [TASK-049](#task-049-implement-duration-and-timezone-formatting-utilities)
  - [TASK-057](#task-057-build-dashboard-sleep-state-control-startstop-button)
- Acceptance criteria:
  - Both metrics visible on dashboard
  - Null values show a placeholder, not zero or an error

### TASK-060: Display today's summary and rolling averages on dashboard
- Size: `S`
- Goal: Show today's total sleep/active time and 7-day/14-day rolling averages.
- Scope:
  - `src/screens/Dashboard/SummaryPanel.tsx` — renders today's totals and rolling averages from the summary response
  - All values formatted with `formatDuration`
  - Labels clarify the period (e.g. "7-day avg", "14-day avg")
  - Layout must fit without scrolling on a standard phone screen alongside the sleep control
- Dependencies:
  - [TASK-049](#task-049-implement-duration-and-timezone-formatting-utilities)
  - [TASK-057](#task-057-build-dashboard-sleep-state-control-startstop-button)
- Acceptance criteria:
  - All eight metric values (today ×2, 7d ×2, 14d ×2, since-last ×2) visible on one phone screen without scroll
  - Values update immediately after Start/Stop completes
