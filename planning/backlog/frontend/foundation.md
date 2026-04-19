## Story 1: Project Foundation

### TASK-049: Implement duration and timezone formatting utilities
- Size: `S`
- Goal: Centralize all time and duration formatting so no screen does ad-hoc conversion.
- Background:
  - The requirements separate two timezone concerns: the **display timezone** (browser local, used for all clock times shown to the user) and the **calculation timezone** (family timezone, used only for day-boundary calculations and sleep classification). These must not be conflated.
- Scope:
  - `src/utils/time.ts`
  - `formatDuration(seconds: number): string` — returns `"2 h 15 min"` style; never raw seconds
  - `formatTime(isoString: string): string` — formats a timestamp in the **browser's local timezone** using `Intl.DateTimeFormat`; no timezone parameter, uses `Intl.DateTimeFormat().resolvedOptions().timeZone`
  - `formatDate(isoString: string, familyTimezone: string): string` — short date label (e.g. `"Wed 16 Apr"`) in the **family timezone**; used only for day-row labels in the timeline where the calendar day must be consistent across family members
  - `getLocalDayBoundaries(date: Date, familyTimezone: string): { start: Date; end: Date }` — returns UTC timestamps for midnight-to-midnight in the family timezone; used for today's totals and timeline row grouping
  - Unit tests for edge cases: zero seconds, exactly 1 hour, DST boundary input, midnight crossing in a non-UTC family timezone
- Acceptance criteria:
  - `formatDuration(0)` returns `"0 min"` or similar clear placeholder
  - `formatTime` always uses browser local timezone; passing a family timezone to it is a type error
  - `formatDate` and `getLocalDayBoundaries` accept an explicit `familyTimezone` parameter and never fall back to the browser timezone

### TASK-050: Implement route guards and auth context provider
- Size: `S`
- Goal: Protect all authenticated routes and provide session state to the component tree.
- Scope:
  - `src/hooks/useAuth.ts` — exposes `{ user, signOut, isLoading }`; reads session from a cookie or memory store (no localStorage)
  - `AuthProvider` wraps the app and fetches initial session state on mount
  - `RequireAuth` wrapper component: redirects to `/` if session is absent
  - `RequireNoFamily` wrapper: redirects to `/onboarding` if user has no family; redirects to `/dashboard` if they do
  - No family data cached across sign-out
- Acceptance criteria:
  - Navigating to `/dashboard` without a session redirects to `/`
  - Signing out clears session state and redirects to `/`
  - Family data is not readable in localStorage or IndexedDB after sign-out
