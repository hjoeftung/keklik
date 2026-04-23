## Story 1: Authentication

### TASK-045: Redirect to frontend after OAuth callback instead of returning JSON
- Size: `S`
- Goal: Complete the browser-based OAuth flow by redirecting the user back to the frontend with session tokens, rather than responding with raw JSON that the browser cannot consume.
- Background:
  - The current `GET /auth/google/callback` handler calls `writeAuthSessionResponse`, which writes a 200 JSON body. When the browser follows the Google redirect to this backend URL it receives JSON with no further navigation — the frontend never learns the session tokens and the app stays broken.
  - The frontend expects to land on `/auth/callback?access_token=...&refresh_token=...&account_id=...` after sign-in (see `src/screens/AuthCallback/AuthCallbackScreen.tsx`).
- Scope:
  - Add a `FRONTEND_URL` config value (e.g. `http://localhost:5173` in development).
  - After a successful OAuth exchange, replace `writeAuthSessionResponse` with an `http.Redirect` to `FRONTEND_URL/auth/callback?access_token=...&refresh_token=...&account_id=...`.
  - On OAuth error (invalid state, missing code, Google error), redirect to `FRONTEND_URL/?error=<code>` instead of returning a JSON error body, so the frontend can surface a readable message on the sign-in screen.
  - Update the existing auth handler tests to assert a redirect response instead of a JSON body.
- Acceptance criteria:
  - Completing the Google OAuth flow in a browser lands the user on `FRONTEND_URL/auth/callback` with all three token params present.
  - OAuth errors redirect to `FRONTEND_URL/` with an `error` query param.
  - No session token is exposed in a JSON response body accessible to arbitrary HTTP clients without a browser context.
