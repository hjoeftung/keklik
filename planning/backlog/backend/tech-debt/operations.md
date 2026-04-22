## Story 5: Operational Hardening

### TASK-021: Add repository transaction boundaries and concurrency tests
- Size: `S`
- Goal: Verify atomic behavior for state-changing commands.
- Scope:
  - Ensure create family, start sleep, stop sleep, edit sleep, join family, and delete sleep run transactionally where needed
  - Add concurrency-focused tests for active sleep conflicts
- Dependencies:
  - [TASK-006](#task-006-implement-create-family-use-case-and-api)
  - [TASK-010](#task-010-implement-join-family-by-invite-link)
  - [TASK-013](#task-013-implement-start-sleep-use-case-and-api)
  - [TASK-014](#task-014-implement-stop-sleep-use-case-and-api)
  - [TASK-015](#task-015-implement-edit-sleep-session-use-case-and-api)
  - [TASK-016](#task-016-implement-hard-delete-sleep-session-use-case-and-api)
- Acceptance criteria:
  - Concurrent start-sleep requests cannot create duplicate active sessions
  - Multi-row write operations rollback cleanly on failure
  - Tests prove atomic behavior for core commands

### TASK-022: Add structured logging and request tracing basics
- Size: `S`
- Goal: Make failures diagnosable in local development and production.
- Scope:
  - Add structured logs for command failures and validation errors
  - Attach request identifier to logs
  - Avoid logging secrets or OAuth credentials
- Acceptance criteria:
  - Logs include request context and stable error code where possible
  - Validation failures are visible in logs without leaking sensitive data
  - Logging is used consistently in handlers and application services

### TASK-023: Add integration test path for core MVP flows
- Size: `S`
- Goal: Cover the main end-to-end backend behaviors with automated tests.
- Scope:
  - Add integration tests for:
    - create family
    - create invite link
    - join family
    - start sleep
    - stop sleep
    - edit sleep
    - delete sleep
    - get sleep history
- Acceptance criteria:
  - The MVP acceptance criteria in [requirements.md](/home/hjoeftung/code/projects/keklik/requirements.md) are covered by automated tests or explicitly mapped to unit coverage
  - The test suite runs locally with documented setup


---

## Story 6: DDD Architecture Cleanup (from 2026-04-16 review)

### TASK-036: Move invite URL construction out of the family application layer
- Size: `S`
- Goal: Remove the HTTP route string and `baseURL` from the family domain, restoring correct dependency direction.
- Scope:
  - Remove `InviteURL` from `CreateFamilyInviteLinkResult` in `internal/family/invitelinks.go`
  - Remove `buildInviteURL` and the `baseURL` field from `CreateFamilyInviteLinkHandler`
  - Build the full invite URL in `internal/interfaces/http/family_invite.go` using `baseURL` injected into the HTTP handler
  - Update the HTTP response DTO to include the constructed URL
- Acceptance criteria:
  - `internal/family` imports nothing from `net/http` or any HTTP-related package
  - `CreateFamilyInviteLinkResult` contains only `InviteLink` (token + expiry) — no URL string
  - Invite URL is assembled in the HTTP handler and returned correctly in the API response

### TASK-037: Export composite sleep session repository interface
- Size: `XS`
- Goal: Make `StopSleepHandler`'s full repository dependency discoverable and nameable.
- Scope:
  - Export `stopSleepSessionRepository` as `ActiveWritableSleepSessionRepository` in `internal/sleep/sleep.go` alongside the other repository interfaces
  - Remove the unexported composite from `internal/sleep/stopsleep.go`
  - Update `NewStopSleepHandler` to accept the exported type
- Acceptance criteria:
  - No unexported interface types used in constructor signatures in `internal/sleep`
  - `ActiveWritableSleepSessionRepository` appears in `internal/sleep/sleep.go` with the other repository role definitions

### TASK-038: Enforce sleep classification invariant inside `SleepSession.Stop`
- Size: `M`
- Goal: Make the aggregate the single enforcer of the "stopped session has correct classification" rule.
- Scope:
  - Change `SleepSession.Stop` signature to accept `stoppedAt time.Time`, `timezone string`, and `nightWindow NightWindow`; call `Classify` internally
  - Remove the `classification SleepClassification` and `classifiedWith NightWindow` parameters from the public signature
  - Update `StopSleepHandler` and `EditSleepSessionHandler` to pass the required inputs rather than a pre-computed classification
  - Keep `classifiedWith NightWindow` as an internal audit field populated by `Stop`
- Acceptance criteria:
  - No caller outside `internal/sleep` can pass an arbitrary classification to `Stop`
  - Existing stop-sleep and edit-sleep tests pass with updated call sites
  - `Classify` is only called from within `SleepSession.Stop`

### TASK-039: Apply private-field discipline to `FamilyMember` and `Baby` entities
- Size: `S`
- Goal: Prevent direct mutation of owned entities and align them with `Family` aggregate's encapsulation model.
- Scope:
  - Convert `FamilyMember` and `Baby` public fields to private fields with accessor methods
  - Introduce `family.ReconstituteFamilyMember(...)` and `family.ReconstituteBaby(...)` factories for the infrastructure reconstruct pattern
  - Update all read sites (infrastructure mappers, application layer) to use accessors
- Acceptance criteria:
  - No code outside `internal/family` can mutate `FamilyMember` or `Baby` fields directly
  - Infrastructure reconstitution uses named factory functions, not struct literals with public fields

### TASK-040: Parse sleep history period at the HTTP boundary
- Size: `XS`
- Goal: Remove the stringly-typed `Period` field from `GetSleepHistoryQuery` and surface parse errors at the correct layer.
- Scope:
  - Replace `GetSleepHistoryQuery.Period string` with `GetSleepHistoryQuery.Range DateRange`
  - Move `periodToDateRange` parsing logic into the HTTP handler in `internal/interfaces/http`
  - Return HTTP 400 from the handler on unrecognised period values before invoking the use case
- Acceptance criteria:
  - `internal/sleep` has no string-to-date-range parsing logic
  - Invalid `period` query parameters produce a 400 response at the HTTP layer

### TASK-041: Name the member display-name fallback policy
- Size: `XS`
- Goal: Give the "name defaults to email prefix" policy a home in the ubiquitous language.
- Scope:
  - Introduce `MemberNameFrom(displayName, email string) string` in `internal/family`
  - Replace the private `resolveMemberName` helper in `internal/family/invitelinks.go` with a call to `MemberNameFrom`
- Acceptance criteria:
  - The fallback rule is expressed through a named, exported function
  - No anonymous string-manipulation logic for member names remains in invite handling code

### TASK-042: Rename repository interfaces to reflect domain role
- Size: `XS`
- Goal: Make repository interface names answer "why does this exist" rather than "what query does it run".
- Scope:
  - Rename `CompletedSleepSessionsSinceRepository` → `ReclassifiableSessionRepository` in `internal/sleep/sleep.go`
  - Rename `EditableSleepSessionRepository` → `OwnedSleepSessionRepository` (or `MemberSleepSessionRepository`) in `internal/sleep/sleep.go`
  - Update all references in use-case files and constructors
- Acceptance criteria:
  - No interface name in `internal/sleep` encodes query mechanics rather than domain role
  - All call sites compile with the new names
