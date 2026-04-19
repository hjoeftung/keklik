## Story 6: Sleep Timeline

### TASK-061: Build sleep timeline chart grid
- Size: `S`
- Goal: Render a scrollable grid with one row per calendar day and a 24-hour time axis.
- Background:
  - Day rows are grouped by calendar day in the **family timezone** (so all family members see the same row structure). Block positions within a row are calculated from UTC timestamps relative to family-timezone midnight, not the browser clock.
- Scope:
  - `src/screens/Timeline/TimelineChart.tsx`
  - Fetch `GET /sleep-sessions?period=7d` on mount
  - Group sessions into rows using `getLocalDayBoundaries` with the family timezone; label each row with `formatDate` (also family timezone)
  - Horizontal axis spans 00:00–24:00 in the family timezone; tick labels at 00, 06, 12, 18, 24
  - Position sleep blocks as `<div>` or SVG `<rect>` elements using percentage-based offsets relative to family-timezone midnight
  - Cross-midnight sessions (in the family timezone) are split across the two affected rows
  - Scrolls vertically; legible at 360 px width
- Dependencies:
  - [TASK-048](#task-048-implement-api-client-with-auth-header-error-normalization-and-401-redirect)
  - [TASK-049](#task-049-implement-duration-and-timezone-formatting-utilities)
- Acceptance criteria:
  - Days render in correct order, newest at top, grouped by family timezone calendar day
  - A session crossing midnight in the family timezone appears as two partial blocks on the correct rows
  - Chart is usable on a 360 px wide screen without horizontal scroll

### TASK-062: Implement sleep block visual distinction and active session indication
- Size: `S`
- Goal: Make night sleep, naps, and the active session visually distinct.
- Scope:
  - Night sleep blocks use one fill color; nap blocks use a different fill color; a legend is provided
  - Active (ongoing) session block extends from its start time to `now`, re-computed via the same `setInterval` as the dashboard timer
  - Active block uses a pulsing or striped CSS animation to distinguish it from completed sessions
- Dependencies:
  - [TASK-061](#task-061-build-sleep-timeline-chart-grid)
  - [TASK-058](#task-058-implement-real-time-elapsed-timer-for-active-session)
- Acceptance criteria:
  - Night and nap blocks are visually different and a legend explains the colors
  - Active block grows in real time and is visually distinct from closed blocks

### TASK-063: Implement session detail overlay on tap/click
- Size: `S`
- Goal: Let users tap a sleep block to see its details.
- Scope:
  - Clicking/tapping any sleep block opens a small overlay (tooltip or modal) showing:
    - Start time — formatted with `formatTime` (browser local timezone)
    - End time — formatted with `formatTime`; shows "ongoing" for active session
    - Duration (`formatDuration`)
    - Classification (night sleep / nap)
  - Overlay dismisses on outside click or a close button
- Dependencies:
  - [TASK-061](#task-061-build-sleep-timeline-chart-grid)
  - [TASK-049](#task-049-implement-duration-and-timezone-formatting-utilities)
- Acceptance criteria:
  - Tapping a block shows all four fields
  - Start/end times are in the browser's local timezone, not the family timezone
  - Active session shows "ongoing" for end time
  - Overlay is keyboard-dismissible (Escape key)

### TASK-064: Add 7-day / 14-day view switcher to timeline
- Size: `S`
- Goal: Let users switch between a 7-day and 14-day view of the timeline.
- Scope:
  - Add a toggle control (two buttons or a segmented control) above the chart
  - Switching re-fetches `GET /sleep-sessions?period=7d` or `?period=14d` accordingly
  - Show a loading state during re-fetch; preserve the selected period across re-renders
- Dependencies:
  - [TASK-061](#task-061-build-sleep-timeline-chart-grid)
- Acceptance criteria:
  - Toggling to 14-day shows 14 rows; toggling back to 7-day shows 7 rows
  - Selection is visually indicated on the toggle control
