## Story 2: Authentication

### TASK-051: Build sign-in screen
- Size: `S`
- Goal: Show unauthenticated visitors a sign-in screen with a single Google OAuth action.
- Scope:
  - `src/screens/SignIn/` — full-screen layout, centered "Sign in with Google" button
  - Button triggers a GET to `/auth/google/start`; follow the redirect returned by the backend
  - No form fields, no other actions
  - Show a loading state while the redirect is in flight
- Acceptance criteria:
  - Unauthenticated users landing on any route see this screen
  - Clicking the button initiates the Google OAuth flow

### TASK-052: Handle Google OAuth callback and post-sign-in routing
- Size: `S`
- Goal: Exchange the OAuth response for a backend session, then route the user correctly.
- Scope:
  - `src/screens/AuthCallback/` — handles the redirect from Google via the backend callback URL
  - On success: if user has no family, redirect to `/onboarding`; otherwise redirect to `/dashboard`
  - If an invite token was preserved (see TASK-054), redirect to `/invite/:token` instead
  - On error, display the error and offer a retry link to sign-in
- Dependencies:
  - [TASK-050](#task-050-implement-route-guards-and-auth-context-provider)
- Acceptance criteria:
  - New user with no family lands on `/onboarding`
  - Returning family member lands on `/dashboard`
  - OAuth error shows a readable message

### TASK-053: Preserve invite-link URL across OAuth redirect
- Size: `S`
- Goal: Ensure a user who follows an invite link and is not signed in lands on the join-family screen after signing in.
- Scope:
  - When an unauthenticated user navigates to `/invite/:token`, store the token in `sessionStorage` before redirecting to sign-in
  - After OAuth callback completes, read and clear the stored token; redirect to `/invite/:token` if present
- Dependencies:
  - [TASK-052](#task-052-handle-google-oauth-callback-and-post-sign-in-routing)
- Acceptance criteria:
  - Visiting `/invite/abc123` without a session, signing in with Google, lands on `/invite/abc123`
  - Token is cleared from sessionStorage after redirect
