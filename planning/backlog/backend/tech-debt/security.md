## Story 6: Security Hardening

### TASK-036: Fix race condition on concurrent first-time OAuth logins
- Size: `S`
- Goal: Prevent duplicate-account constraint errors when two OAuth callbacks for the same Google subject ID arrive simultaneously.
- Scope:
  - Replace the check-then-insert pattern in `findOrCreateAccount` with `INSERT ... ON CONFLICT (google_subject_id) DO UPDATE` (upsert), or catch the constraint violation and retry with a fetch
  - Add a test that simulates concurrent callbacks for the same subject ID
- Files: [internal/auth/handleoauthcallback.go](internal/auth/handleoauthcallback.go)
- Acceptance criteria:
  - Concurrent callbacks for the same `google_subject_id` both resolve to the same account without DB errors
  - No duplicate rows in `accounts` under concurrent load

### TASK-037: Add session invalidation (logout) endpoint
- Size: `S`
- Goal: Allow users to revoke their session and prevent unbounded accumulation of expired tokens in the DB.
- Scope:
  - Add `POST /auth/logout` handler that deletes the caller's token row from `sessions`
  - Add a periodic cleanup job (or a DB cron query) to purge expired sessions
  - Consider reducing session TTL from 30 days to 8 hours
- Acceptance criteria:
  - After logout, the token returns 401 on any authenticated endpoint
  - Expired sessions are removed from the DB over time
  - Logout is documented in the OpenAPI spec

### TASK-038: Add Secure flag to OAuth state cookie
- Size: `XS`
- Goal: Prevent the state cookie from being transmitted over plain HTTP in production.
- Scope:
  - Set `Secure: true` on the state cookie in `internal/interfaces/http/auth.go:62-69`
  - Gate it on environment: `!isDev` or `ENVIRONMENT=production`, so local dev keeps working
- Files: [internal/interfaces/http/auth.go](internal/interfaces/http/auth.go)
- Acceptance criteria:
  - Cookie has `Secure` flag in any non-dev environment
  - Local dev flow is unaffected

### TASK-039: Add timestamp verification to OAuth state parameter
- Size: `XS`
- Goal: Prevent replayed state cookies — a stolen cookie replayed after the 5-minute window should be rejected.
- Scope:
  - Embed a Unix timestamp in the state value alongside the nonce (e.g., `nonce.base64(ts)`)
  - On callback, verify the timestamp is within 5 minutes before accepting the state
- Files: [internal/interfaces/http/auth.go](internal/interfaces/http/auth.go)
- Acceptance criteria:
  - State values older than 5 minutes are rejected with 400
  - Valid state values within the window continue to work

### TASK-040: Return 401 (not 404) from disabled test auth endpoint
- Size: `XS`
- Goal: Avoid fingerprinting the test-auth feature through 404 responses when it is disabled.
- Scope:
  - In `internal/interfaces/http/auth.go:155-158`, return `401 Unauthorized` instead of `http.NotFound` when `ENABLE_TEST_AUTH=false`
  - Alternatively, skip registering the route entirely when the flag is off
- Files: [internal/interfaces/http/auth.go](internal/interfaces/http/auth.go)
- Acceptance criteria:
  - Disabled test auth endpoint returns 401, not 404
  - Enabled test auth endpoint continues to work as before

### TASK-041: Add invite token revocation endpoint and rate-limit the join endpoint
- Size: `S`
- Goal: Limit the blast radius of a leaked invite link and prevent brute-force joins.
- Scope:
  - Add `DELETE /families/{family_id}/invite-links/{token}` (or equivalent) to revoke a specific token before expiry
  - Add rate limiting to the invite join endpoint: 5 attempts per IP per minute
  - Use `golang.org/x/time/rate` or equivalent middleware
- Files: [internal/family/invitelinks.go](internal/family/invitelinks.go)
- Acceptance criteria:
  - A revoked token returns 404/410 on the join endpoint
  - More than 5 join attempts from the same IP in 60 seconds returns 429

### TASK-042: Enforce test account silo on family joins
- Size: `XS`
- Goal: Prevent test accounts (created via `ENABLE_TEST_AUTH`) from joining production families.
- Scope:
  - In `JoinFamilyByInviteLinkHandler`, reject accounts whose `google_subject_id` starts with `test:` unless `ENABLE_TEST_AUTH=true`
  - Alternatively, add an `is_test_account` column to `accounts` and gate family joins on it
- Files: [internal/auth/handletestlogin.go](internal/auth/handletestlogin.go)
- Acceptance criteria:
  - A test account cannot join a family when `ENABLE_TEST_AUTH=false`
  - Normal accounts are unaffected

### TASK-043: Gate `sslmode=disable` on an explicit dev-only flag
- Size: `XS`
- Goal: Prevent accidental unencrypted DB connections in production.
- Scope:
  - Change the default DB connection string to `sslmode=require`
  - Allow override only via explicit `ALLOW_INSECURE_DB=true` env var
  - Add a startup panic (or logged fatal) if `sslmode=disable` and `ENVIRONMENT=production`
- Files: [compose.yaml](compose.yaml)
- Acceptance criteria:
  - Production deploy fails fast if `sslmode=disable` is present without the override flag
  - Local dev with `ALLOW_INSECURE_DB=true` continues to work

### TASK-044: Normalize 403/404 responses for cross-family baby access
- Size: `XS`
- Goal: Prevent enumeration of valid baby IDs by other families.
- Scope:
  - In `internal/infrastructure/babyaccesschecker.go:38-44`, return 404 for both "baby doesn't exist" and "baby belongs to another family"
  - Update tests to assert 404 in both cases
- Files: [internal/infrastructure/babyaccesschecker.go](internal/infrastructure/babyaccesschecker.go)
- Acceptance criteria:
  - A request for a valid baby UUID owned by a different family returns 404, not 403
  - A request for a random UUID still returns 404

### TASK-045: Add rate limiting middleware to auth and invite endpoints
- Size: `S`
- Goal: Prevent brute-force and resource-exhaustion attacks on unauthenticated endpoints.
- Scope:
  - Apply per-IP rate limiting to `POST /auth/test/login` and the invite join endpoint at minimum
  - Use `golang.org/x/time/rate` token-bucket middleware; return 429 on breach
  - Consider applying a broader limit to all write endpoints
- Dependencies: [TASK-041](#task-041-add-invite-token-revocation-endpoint-and-rate-limit-the-join-endpoint)
- Acceptance criteria:
  - Repeated requests to rate-limited endpoints from the same IP are throttled with 429
  - Legitimate traffic within limits is unaffected

### TASK-046: Add audit logging for sensitive operations
- Size: `S`
- Goal: Create an observable trail for account creation, family membership changes, and invite generation.
- Scope:
  - Use `slog` structured logging on: account creation, family creation, invite link creation, family join, and logout
  - Log `account_id`, `family_id`, action, and timestamp; never log tokens or credentials
- Dependencies: [TASK-022](#task-022-add-structured-logging-and-request-tracing-basics)
- Acceptance criteria:
  - Each listed event emits a structured log line at INFO level with the required fields
  - No secrets (tokens, OAuth codes) appear in logs
