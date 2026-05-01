# Warm Redesign — Backlog

Design reference: design/
Requirements: `sleep-and-stats-requirements-v2.md`

---

## Backend

### B2 — Extend sleep history to 90d and consolidate stats into one endpoint

---

## Frontend

---

### F3 — Stats: Today tab

**What:** Replace the current horizontal timeline with a vertical 24h diary view.

**Layout:**
- Summary header: three pill cards (Total Sleep / Total Nap / Total Active) — values from B5
- Vertical timeline below, scrollable, 24h window from B5 (`window_start` → `window_end`)
- Hour labels on the left at each full hour
- Horizontal grid lines at each hour
- Sessions rendered as vertical colored bars spanning full column width, height proportional to duration
  - Night sleep: `night` color (#5B7BB8); Nap: `nap` color (#E8B86E)
  - Corner radius scales with session height (short naps = very small radius)
  - Duration label inside (or above) the bar
  - Start/end time shown when bar is tall enough
- Awake gap pills between consecutive sessions ("awake · 2h 17m"), dashed border
- Active sessions are **not shown** in the diary
- Tapping a session bar opens a detail panel (bottom sheet): start, end, duration, classification explanation, Edit and Delete actions
  - Edit saves via existing PATCH endpoint
  - Delete saves via existing DELETE endpoint

**Scope:** New `TodayTab` component, reuses `EditSession` bottom sheet design from the design file. Fetch via B5 for stats + B2 for window. Session list from the existing history endpoint with `period=today`.

---

### F4 — Stats: Week tab

**What:** 7-column grid showing the past 7 days side by side.

**Layout:**
- Column headers: day name + date number (e.g. "Mon\n21")
- Vertical axis: hour ticks (06, 12, 18, 24) shared on the left
- Each column: same 24h window as Today tab
- Sessions as colored blocks (night/nap) positioned and sized by time of day
- Corner radius: scales with block height (very short blocks nearly square)
- Legend: Night / Nap color swatches
- Display-only — no tap interaction required

**Scope:** New `WeekTab` component. Shares the 24h window from B2. Takes 7 days of session data from the existing history endpoint with `period=7d`.

---

### F5 — Stats: Summary tab

**What:** Aggregated averages for a selected period.

**Layout:**
- Four sub-tabs: 7d / 14d / 30d / 90d (pill-style segmented control)
- Date range label below: e.g. "21 April 2026 – 28 April 2026"
- Three stat cards stacked vertically, each with a colored icon circle:
  - Avg Sleep (night icon, `night` color)
  - Avg Nap (sun icon, `nap` color)
  - Avg Active (star icon, `primary` color)
- Numbers in Fraunces display font

**Scope:** New `SummaryTab` component. Data from B4. Switching sub-tabs triggers a new fetch (or cache hit).

---

### F6 — Stats screen container: tabs, data loading, skeletons, pull-to-refresh

**What:** The outer `TimelineScreen` (to be renamed `StatsScreen`) that:
- Renders the three top-level tabs (Today / Week / Summary)
- Fetches session data **optimistically on app load**, before the user navigates to Stats
- Caches data in a React context or module-level store for the app session lifetime
- Passes cached data to each tab; tabs show skeleton placeholders while the cache is unpopulated
- Pull-to-refresh re-fetches and updates the cache
- Skeleton loaders: match the approximate shape of each tab's content

**Scope:** Refactor `TimelineScreen` into `StatsScreen`, add an `AppDataCache` context (or similar), wire up pull-to-refresh with the browser's native scroll-up gesture.

---

### F7 — Bottom navigation bar redesign

**What:** Replace the current `NavBar` with a warm Keklik-styled tab bar.

**Design:**
- Three tabs: Sleep (pillow icon) / Stats (chart icon) / Settings (gear icon)
- Surface: white with a top border in `border` color, subtle shadow
- Active tab: primary color fill on icon + label
- Inactive: inkMuted color
- Large touch targets (~64px height)
- Primary interaction area at the bottom per mobile-first design principle

**Scope:** Update `NavBar.tsx` and its CSS. The pillow icon for Sleep tab is a mini version of the pillow SVG from F2.

---

### F8 — Onboarding and Settings screens (visual pass)

**What:** Apply the new design system to the existing Onboarding and Settings screens. These screens already exist functionally; this story is purely a visual update.

**Onboarding:**
- Welcome screen: hero illustration (pillow + moon + cloud + stars), warm headline, "Continue with Google" big button
- Setup screen: baby name + birthday + night window form, warm card styling

**Settings:**
- Baby card at top (avatar + name + age + caregiver count)
- Grouped rows in rounded cards: Sleep section (Night window, Day starts at, Time format), Family section (Caregivers, Invite partner), About section (Help, Sign out)
- Version string at bottom

**Scope:** `OnboardingScreen.tsx`, `SettingsScreen.tsx` CSS modules. No functional changes — only styling.
