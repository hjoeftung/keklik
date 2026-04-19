# Keklik Frontend Requirements

## 1. Purpose

Keklik is a simple baby sleep tracker for families. This document covers the frontend MVP — a simple web application that lets family members track a baby's sleep from any browser.

The frontend consumes the Keklik HTTP JSON API described in the backend requirements.

## 2. Scope

### In scope for the first version
- Google OAuth sign-in
- Family creation and family join via invite link
- Starting and stopping a baby's sleep session
- Viewing time since last sleep and last awakening
- Viewing today's sleep and active totals
- Viewing 7-day and 14-day average sleep and active durations
- Weekly timeline chart showing sleep blocks across a 24-hour axis per day
- Basic family settings management (timezone, night window, baby name, invite link)

### Out of scope for the first version
- Editing or deleting individual sleep sessions
- Telegram bot or native Android app
- Push notifications or reminders
- Multi-language support
- Offline support or service workers
- Advanced charts or analytics
- Multi-baby support
- Role-based UI differences between family members

## 3. Product Goal

A family member should be able to open the application on any device with a browser, sign in with Google, and immediately see the baby's current sleep state and key metrics. Starting or stopping a sleep should require as few taps as possible.

## 4. Users and Roles

### Family member
- Has signed in with Google
- Is linked to a family (either by creating one or accepting an invite)
- Can perform all actions: start/stop sleep, view metrics, manage family settings
- All family members have identical permissions in the MVP

### Unauthenticated visitor
- Can only see the sign-in screen
- All other routes must redirect to sign-in

## 5. Screens and Functional Requirements

### 5.1 Sign-in screen

1. The application must show a sign-in screen to unauthenticated visitors.
2. The sign-in screen must offer a single "Sign in with Google" action.
3. After successful Google OAuth, the application must redirect to onboarding if the user has no family, or to the dashboard otherwise.
4. An invite-link URL must survive the OAuth redirect so the user lands on the accept-invite screen after sign-in.

### 5.2 Onboarding — create family

1. A signed-in user with no family must be shown an onboarding screen.
2. The onboarding screen must let the user create a new family by providing:
   - User name
   - Baby name
   - Night window start and end (local time, e.g. 20:00–08:00)
   - Timezone (default to the browser's detected timezone, allow override via a searchable dropdown)
3. On success the user must land on the dashboard.

### 5.3 Onboarding — join family via invite link

1. When a signed-in user navigates to a valid invite URL, the application must show a confirmation screen summarising what family they are about to join.
2. The user must be able to confirm or cancel.
3. On confirmation the application must call the join-family API and then redirect to the dashboard.
4. If the invite link is invalid or expired, the application must display an explanatory error and offer the option to create a new family instead.

### 5.4 Dashboard

The dashboard is the primary screen. It must be accessible on a single page load after sign-in.

#### 5.4.1 Sleep state control

1. The dashboard must show whether the baby is currently sleeping or awake.
2. When the baby is awake, the dashboard must show a prominent "Start sleep" button.
3. When the baby is sleeping, the dashboard must show a prominent "Stop sleep" button.
4. Tapping Start or Stop must call the corresponding API endpoint and update the UI immediately on success.
5. An in-progress sleep session must display the elapsed time since it started, updating in real time (client-side timer, no polling required).

#### 5.4.2 Time since last event

1. The dashboard must show the time elapsed since the last sleep started.
2. The dashboard must show the time elapsed since the last awakening.
3. If either metric is unavailable (no data yet), the application must show a clear placeholder rather than zero or an error.

#### 5.4.3 Today's summary

1. The dashboard must show total sleep time for today.
2. The dashboard must show total active time for today.
3. "Today" must be interpreted using the family's configured timezone.

#### 5.4.4 Rolling averages

1. The dashboard must show average daily sleep time for the last 7 days.
2. The dashboard must show average daily sleep time for the last 14 days.
3. The dashboard must show average daily active time for the last 7 days.
4. The dashboard must show average daily active time for the last 14 days.

### 5.5 Sleep timeline screen

The sleep timeline is a visual chart showing sleep blocks across a horizontal 24-hour axis, with one row per day.

1. The screen must render a grid where:
   - Each row represents one calendar day, labeled with a short date (e.g. "Wed 16 Apr").
   - The horizontal axis represents 24 hours from 00:00 to 24:00 in the family timezone.
   - Sleep sessions are drawn as filled rectangles positioned by their start and stop times.
2. Night sleep and naps must be visually distinct (different fill colors).
3. An active (ongoing) session must be shown as a block extending from its start time to the current moment, visually distinguished from completed sessions (e.g. a pulsing or striped fill).
4. Sessions that cross midnight must be split across the two affected day rows.
5. The user must be able to switch between a 7-day view and a 14-day view.
6. The most recent day must appear at the top.
7. Tapping or clicking a sleep block must show a small overlay with the session's start time, end time, duration, and classification.
8. The timeline must scroll vertically when there are more days than fit on screen.
9. The chart must remain legible on a 360 px wide mobile screen; the time axis labels may be sparse (e.g. every 6 hours: 00, 06, 12, 18, 24).

### 5.6 Settings screen

1. The application must provide a settings screen accessible from the dashboard.
2. The settings screen must allow editing:
   - User name
   - Baby name
   - Timezone (searchable dropdown)
   - Night window (start and end local time)
3. Changes must be saved explicitly (a Save button), not auto-saved.
4. The settings screen must allow generating a new invite link.
5. The generated invite link must be displayable in a copyable format.

## 6. Navigation

1. The application must have a persistent navigation element visible on all authenticated screens.
2. Navigation must link to: Dashboard, Sleep timeline, Settings.
3. Navigation must include a sign-out action.
4. The application must be a single-page application (SPA) with client-side routing.
5. All routes except sign-in must redirect to sign-in for unauthenticated users.

## 7. UI and UX Requirements

1. The application must be usable on mobile-sized screens (minimum 360 px width) as well as desktop.
2. The Start/Stop sleep button must be large enough to tap comfortably on a phone screen.
3. The dashboard must display the most important information (current state, elapsed time, today's totals) without scrolling on a typical phone.
4. Duration values must be formatted as hours and minutes (e.g. 2 h 15 min), never as raw seconds or ISO durations.
5. Times must always be displayed in the family's configured timezone, not the browser's local time.
6. The application must give the user clear feedback for all async actions (loading state, success, and error).
7. Error messages returned by the API must be surfaced to the user in plain language rather than raw codes.

## 8. API Integration

The frontend communicates exclusively with the Keklik backend HTTP JSON API.

### 8.1 Authentication

1. After Google OAuth callback, the frontend must exchange the Google credential for a session with the backend.
2. The frontend must include the session credential on all subsequent API requests (e.g. via an Authorization header or cookie, depending on the backend session design).
3. On 401 responses the frontend must clear local session state and redirect to sign-in.

### 8.2 Endpoints used

| Action | Method | Path |
|---|---|---|
| Start Google OAuth flow | GET | /auth/google/start |
| Create family | POST | /families |
| Create sleep profile | POST | /sleep-profiles |
| Create invite link | POST | /families/invite-links |
| Join family by invite link | POST | /families/join-by-invite-link |
| Start sleep | POST | /sleep-sessions |
| Stop active sleep | DELETE | /sleep-sessions/active |
| Get dashboard summary | GET | /sleep-sessions/summary |
| Get sleep timeline data | GET | /sleep-sessions?period=7d\|14d |

The dashboard summary endpoint returns all metrics needed by the dashboard in a single response, eliminating the need for multiple requests on load or after a start/stop action:

```json
{
  "active_session": {
    "id": "...",
    "started_at": "2026-04-15T09:30:00Z"
  },
  "since_last": {
    "since_sleep_start_seconds": 4500,
    "since_awakening_seconds": null
  },
  "today": {
    "total_sleep_seconds": 28800,
    "total_active_seconds": 18000
  },
  "rolling_7d": {
    "avg_daily_sleep_seconds": 30000,
    "avg_daily_active_seconds": 17000
  },
  "rolling_14d": {
    "avg_daily_sleep_seconds": 29500,
    "avg_daily_active_seconds": 17500
  }
}
```

`active_session` is `null` when the baby is awake. `since_last` fields are `null` when there is not enough data yet.

### 8.3 Error handling

1. Network errors must show a generic retry message.
2. 400 and 422 responses must surface the API's machine-readable error code in a human-readable form.
3. 409 responses (e.g. duplicate active session) must explain the conflict to the user clearly.
4. The frontend must not silently swallow errors.

## 9. Technical Requirements

### 9.1 Recommended technology choices

- Framework: React with TypeScript
- Build tool: Vite
- Styling: plain CSS or a lightweight utility library (e.g. Tailwind CSS); no heavy UI component framework required
- Routing: React Router
- State management: React context or a lightweight store (e.g. Zustand); avoid heavy solutions like Redux for the MVP
- HTTP client: native `fetch` with a thin wrapper; no full SDK required
- No server-side rendering required for the MVP

Rationale: React + TypeScript + Vite gives a fast development loop, good tooling, and broad familiarity. The lightweight choices keep the initial bundle and complexity low.

### 9.2 Project structure

Suggested top-level layout:
- `src/api/` — API client and type definitions
- `src/components/` — shared UI components
- `src/screens/` — one directory per screen
- `src/hooks/` — shared custom hooks for data fetching and business logic
- `src/utils/` — time formatting and other helpers

### 9.3 Time and timezone handling

1. The frontend must not use the browser's local timezone for any displayed timestamps; use the family timezone returned by the API.
2. Duration formatting must be client-side only (convert seconds to h/min display).
3. The real-time elapsed timer on the dashboard must be driven by a `setInterval` using the session's start time from the API, not a server-pushed clock.
4. Timezone-aware formatting should use the Intl API (`Intl.DateTimeFormat`) rather than a heavyweight date library.

## 10. Non-Functional Requirements

### Reliability
1. The application must not lose user input on API failure; form state must be preserved so the user can retry.
2. Optimistic UI updates are not required; the application may wait for API confirmation before updating state.

### Performance
1. Initial load must not require a large bundle; code-splitting per route is recommended.
2. The dashboard must be interactive within 3 seconds on a typical mobile connection.

### Security
1. Google OAuth must be the only sign-in method.
2. The frontend must not store raw Google tokens in localStorage; session management must follow the backend's design.
3. No family data may be cached in a way that persists across sign-out (localStorage, IndexedDB).

### Accessibility
1. Interactive controls (Start/Stop button, forms) must be keyboard-navigable.
2. All images and icons must have descriptive alt text or aria labels.

## 11. Acceptance Criteria

The first frontend version is acceptable when:

1. A new user can sign in with Google and create a family with a baby, timezone, and night window.
2. An existing family member can sign in and land on the dashboard showing current sleep state, elapsed time, and today's totals.
3. A family member can start a sleep session from the dashboard and see the elapsed timer begin.
4. A family member can stop an active sleep session from the dashboard.
5. A family member can view the sleep timeline chart for the last 7 days and 14 days.
6. Sleep blocks are visually distinct between night sleep and naps.
7. An active sleep session appears as an ongoing block on the timeline.
8. Tapping a sleep block shows a detail overlay with start time, end time, duration, and classification.
9. A family member can copy an invite link from the settings screen.
6. A new user can follow an invite link, sign in with Google, and land on the dashboard as a member of that family.
7. A family member can update the baby name, timezone, and night window from the settings screen.
13. All times are displayed in the family's configured timezone.
14. Duration values are displayed in a human-readable h/min format.
15. The application redirects unauthenticated users to the sign-in screen.
16. API errors are surfaced to the user with a legible message.
17. The application is usable on a mobile phone screen without horizontal scrolling.
