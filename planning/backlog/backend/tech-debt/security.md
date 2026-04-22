## Story 6: Security Hardening



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
