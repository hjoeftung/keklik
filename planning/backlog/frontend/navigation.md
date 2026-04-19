## Story 4: Navigation

### TASK-056: Implement persistent navigation bar
- Size: `S`
- Goal: Give authenticated users persistent access to all main screens and a sign-out action.
- Scope:
  - `src/components/NavBar/` — rendered on all authenticated routes via a layout wrapper
  - Links: Dashboard (`/dashboard`), Sleep timeline (`/timeline`), Settings (`/settings`)
  - Sign-out action calls `signOut()` from auth context, then redirects to `/`
  - Responsive: collapses to icons or a bottom bar on mobile (360 px)
  - Active route is visually highlighted
- Dependencies:
  - [TASK-050](#task-050-implement-route-guards-and-auth-context-provider)
- Acceptance criteria:
  - Nav is visible on dashboard, timeline, and settings screens
  - Nav is not visible on sign-in, callback, or onboarding screens
  - Sign-out clears session and redirects to `/`
