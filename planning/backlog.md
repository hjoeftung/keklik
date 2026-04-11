# Keklik Backend Backlog

This document decomposes [requirements.md](/home/hjoeftung/code/projects/keklik/requirements.md) into S-sized implementation tasks.

## S Size Definition

- Size: `S`
- Expected effort: about 0.5 to 2 developer days
- A task is complete only when code, tests, and minimal documentation are included

## Recommended Delivery Order

1. Foundation and architecture
2. Authentication and family onboarding
3. Sleep lifecycle
4. Reporting and summaries
5. Operational hardening

## Story 1: Foundation and Architecture

### TASK-001: Bootstrap Go service skeleton
- Size: `S`
- Goal: Create the initial Go backend structure as a modular monolith with DDD boundaries.
- Scope:
  - Initialize Go module
  - Create package layout from the requirements
  - Add entrypoint for HTTP server
  - Add config loading skeleton
  - Add health check endpoint
- Suggested packages:
  - `cmd/api`
  - `internal/family`
  - `internal/auth`
  - `internal/sleep`
  - `internal/reporting`
  - `internal/infrastructure`
  - `internal/interfaces/http`
- Dependencies: none
- Acceptance criteria:
  - The project builds successfully
  - The HTTP server starts locally
  - A health endpoint returns success
  - The package structure reflects the bounded contexts in [requirements.md](/home/hjoeftung/code/projects/keklik/requirements.md)

### TASK-002: Add configuration and environment contract
- Size: `S`
- Goal: Define runtime configuration for the backend.
- Scope:
  - Add config struct and loader
  - Define required environment variables
  - Cover HTTP port, database DSN, Google OAuth settings, and base URL for invite links
- Required config keys:
  - `HTTP_PORT`
  - `DATABASE_URL`
  - `GOOGLE_OAUTH_CLIENT_ID`
  - `GOOGLE_OAUTH_CLIENT_SECRET`
  - `GOOGLE_OAUTH_REDIRECT_URL`
  - `APP_BASE_URL`
- Dependencies: [TASK-001](#task-001-bootstrap-go-service-skeleton)
- Acceptance criteria:
  - Service fails fast on missing required configuration
  - Local development defaults are documented where appropriate
  - Config is injectable into application and infrastructure layers

### TASK-003: Add PostgreSQL migration framework and baseline schema
- Size: `S`
- Goal: Establish schema management for the MVP.
- Scope:
  - Add migration tooling
  - Create initial schema for families, accounts, babies, invite links, and sleep sessions
  - Add indexes for active sleep lookups and date-range queries
- Minimum schema expectations:
  - Store timestamps in UTC
  - Store family timezone as IANA identifier
  - Persist sleep classification and classification rule version
- Dependencies: [TASK-001](#task-001-bootstrap-go-service-skeleton)
- Acceptance criteria:
  - Migrations run cleanly on an empty database
  - Migrations are repeatable in local development
  - Schema supports every persistence requirement in [requirements.md](/home/hjoeftung/code/projects/keklik/requirements.md)

### TASK-004: Establish shared error model and HTTP error mapping
- Size: `S`
- Goal: Standardize machine-readable error handling.
- Scope:
  - Define domain and application error codes
  - Add HTTP mapping for validation, auth, conflict, not-found, and forbidden cases
  - Add consistent JSON error response shape
- Example codes:
  - `invalid_argument`
  - `unauthenticated`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `invalid_timezone`
  - `active_sleep_exists`
  - `invalid_sleep_interval`
  - `invalid_invite_link`
- Dependencies: [TASK-001](#task-001-bootstrap-go-service-skeleton)
- Acceptance criteria:
  - All handlers can return the shared error model
  - Error responses include stable code and human-readable message
  - Conflict scenarios map to HTTP 409

## Story 2: Authentication and Family Onboarding

### TASK-005: Model family aggregate and repository interfaces
- Size: `S`
- Goal: Define the family domain model before implementing commands.
- Scope:
  - Add `Family`, `Account`, `Baby`, `NightWindow`, and `InviteLink` domain types
  - Define repository interfaces
  - Encode MVP invariant of exactly one baby per family while keeping future extension possible
- Key rules to encode:
  - Family owns baby, members, timezone, and night window
  - Family members have identical permissions
  - Invite links belong to a family
- Dependencies: [TASK-001](#task-001-bootstrap-go-service-skeleton)
- Acceptance criteria:
  - Domain types compile without infrastructure dependencies
  - Repository interfaces support the planned commands and queries
  - Unit tests cover key family invariants

### TASK-006: Implement create family use case and API
- Size: `S`
- Goal: Allow creation of a family with the initial baby and settings.
- Scope:
  - Implement `CreateFamily`
  - Validate family name, baby name, timezone, and night window
  - Persist family, creator account, and initial baby in one transaction
  - Expose HTTP endpoint
- Request must include:
  - Family name
  - Baby name
  - Family timezone
  - Night window start and end local time
- Dependencies:
  - [TASK-003](#task-003-add-postgresql-migration-framework-and-baseline-schema)
  - [TASK-005](#task-005-model-family-aggregate-and-repository-interfaces)
  - [TASK-004](#task-004-establish-shared-error-model-and-http-error-mapping)
- Acceptance criteria:
  - Family creation stores the first baby together with the family
  - Invalid IANA timezones are rejected
  - Invalid night windows are rejected
  - API returns identifiers needed by clients

### TASK-007: Implement Google OAuth identity flow
- Size: `S`
- Goal: Support Google OAuth as the only authentication option for the MVP.
- Scope:
  - Implement OAuth start and callback flow
  - Verify Google identity token or callback response
  - Resolve or provision internal account identity record
  - Define how authenticated identity is attached to requests
- Decision notes:
  - Keep session mechanism simple for MVP
  - Preserve Google subject identifier in account data
- Dependencies:
  - [TASK-002](#task-002-add-configuration-and-environment-contract)
  - [TASK-004](#task-004-establish-shared-error-model-and-http-error-mapping)
- Acceptance criteria:
  - Unauthenticated requests are rejected on protected endpoints
  - Google-authenticated identity can be resolved to internal account data
  - OAuth failure modes return stable API errors

### TASK-008: Implement edit family use case and API
- Size: `S`
- Goal: Allow family settings updates after creation.
- Scope:
  - Implement `EditFamily`
  - Support updating family name, baby name, timezone, and night window
  - Ensure night-window changes apply only forward in time
  - Expose HTTP endpoint
- Important rule:
  - Past sleep sessions must keep their stored classification unless edited later
- Dependencies:
  - [TASK-006](#task-006-implement-create-family-use-case-and-api)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - Family settings can be updated by any family member
  - Existing historical session classifications remain unchanged
  - Updated settings are used for future sleep creation and reporting

### TASK-009: Implement family invite link creation
- Size: `S`
- Goal: Let family members create shareable invite links.
- Scope:
  - Implement `CreateFamilyInviteLink`
  - Generate secure token
  - Persist expiration and creator metadata
  - Expose endpoint that returns a full invite URL
- Suggested defaults:
  - Configurable expiry duration
  - One-time use or multi-use should be decided in code comments if still open internally
- Dependencies:
  - [TASK-006](#task-006-implement-create-family-use-case-and-api)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - Only authenticated family members can generate links
  - Invite links contain enough information to join the family later
  - Expired links are not considered valid

### TASK-010: Implement join family by invite link
- Size: `S`
- Goal: Allow a Google-authenticated user to join a family through a valid invite link.
- Scope:
  - Implement `JoinFamilyByInviteLink`
  - Validate token, expiration, and family state
  - Link authenticated user to the family as a member
  - Prevent duplicate membership
  - Expose endpoint
- Dependencies:
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
  - [TASK-009](#task-009-implement-family-invite-link-creation)
  - [TASK-005](#task-005-model-family-aggregate-and-repository-interfaces)
- Acceptance criteria:
  - Only a Google-authenticated user can accept an invite
  - Invalid or expired links are rejected
  - A valid invite links the user to the family exactly once

## Story 3: Sleep Lifecycle

### TASK-011: Model sleep session aggregate and repository interfaces
- Size: `S`
- Goal: Define the core sleep domain model with business invariants.
- Scope:
  - Add `SleepSession` domain type
  - Add classification enum or value object for nap versus night sleep
  - Add repository interfaces for active session lookup and date-range queries
- Key rules to encode:
  - Only one active sleep session per baby
  - `stop >= start`
  - Duration is derived
  - Classification is derived, not user-entered
- Dependencies: [TASK-001](#task-001-bootstrap-go-service-skeleton)
- Acceptance criteria:
  - Domain model is independent from transport and persistence
  - Unit tests cover aggregate lifecycle and invariants

### TASK-012: Implement timezone-aware sleep classification service
- Size: `S`
- Goal: Derive nap versus night-sleep classification safely across DST and midnight boundaries.
- Scope:
  - Implement calculation based on family timezone and night window
  - Support windows that cross midnight
  - Persist classification and classification rule version
  - Ensure future night-window changes do not silently reclassify past sessions
- Required behaviors:
  - Night sleep if more than half of session duration falls inside the night window
  - Nap otherwise
  - Recalculate on session edit
- Dependencies:
  - [TASK-005](#task-005-model-family-aggregate-and-repository-interfaces)
  - [TASK-011](#task-011-model-sleep-session-aggregate-and-repository-interfaces)
- Acceptance criteria:
  - Unit tests cover DST forward and backward cases
  - Unit tests cover night windows spanning midnight
  - Unit tests cover forward-only night-window changes

### TASK-013: Implement start sleep use case and API
- Size: `S`
- Goal: Start a sleep session for the family's baby.
- Scope:
  - Implement `StartSleep`
  - Validate family membership
  - Check for existing active sleep session
  - Persist active session atomically
  - Expose endpoint
- Dependencies:
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
  - [TASK-011](#task-011-model-sleep-session-aggregate-and-repository-interfaces)
  - [TASK-003](#task-003-add-postgresql-migration-framework-and-baseline-schema)
  - [TASK-004](#task-004-establish-shared-error-model-and-http-error-mapping)
- Acceptance criteria:
  - Starting a sleep creates an active session
  - Starting a second active sleep for the same baby returns conflict
  - Concurrent requests do not create duplicate active sessions

### TASK-014: Implement stop sleep use case and API
- Size: `S`
- Goal: Stop the current active sleep session.
- Scope:
  - Implement `StopSleep`
  - Load active session
  - Validate stop time
  - Calculate duration and final classification
  - Persist atomically
  - Expose endpoint
- Dependencies:
  - [TASK-012](#task-012-implement-timezone-aware-sleep-classification-service)
  - [TASK-013](#task-013-implement-start-sleep-use-case-and-api)
- Acceptance criteria:
  - Stopping an active sleep produces a completed session
  - Invalid stop time is rejected
  - Classification is stored with the completed session

### TASK-015: Implement edit sleep session use case and API
- Size: `S`
- Goal: Allow correction of both active and completed sleep sessions.
- Scope:
  - Implement `EditSleepSession`
  - Support changing start and stop values
  - Recompute duration and classification as needed
  - Preserve forward-only classification behavior by using current rule only on edited session
  - Expose endpoint
- Dependencies:
  - [TASK-012](#task-012-implement-timezone-aware-sleep-classification-service)
  - [TASK-014](#task-014-implement-stop-sleep-use-case-and-api)
- Acceptance criteria:
  - Active sessions can be edited
  - Completed sessions can be edited
  - Invalid intervals are rejected
  - Edited sessions receive updated derived values

### TASK-016: Implement hard delete sleep session use case and API
- Size: `S`
- Goal: Allow unrestricted deletion of sleep sessions by family members.
- Scope:
  - Implement delete command
  - Validate membership and family boundary
  - Perform hard delete
  - Expose endpoint
- Dependencies:
  - [TASK-015](#task-015-implement-edit-sleep-session-use-case-and-api)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - Any family member can delete a sleep session in their family
  - Deleted sessions are removed from subsequent queries
  - Deleting a non-existent session returns not found

## Story 4: Reporting and Summaries

### TASK-017: Implement sleep history query and API
- Size: `S`
- Goal: Return sleep sessions for today, 7 days, and 14 days.
- Scope:
  - Implement `GetSleepSessionsByPeriod`
  - Support period presets and explicit date ranges if useful internally
  - Order by start time descending
  - Return stored classification with each session
- Dependencies:
  - [TASK-015](#task-015-implement-edit-sleep-session-use-case-and-api)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - API returns sessions for requested period in family timezone context
  - Results are ordered by start time descending
  - Sessions are scoped to the caller's family only

### TASK-018: Implement elapsed time queries and API
- Size: `S`
- Goal: Return time since last sleep start and last awakening.
- Scope:
  - Implement `GetTimeSinceLastSleepStart`
  - Implement `GetTimeSinceLastAwakening`
  - Return explicit empty result when no data exists
- Dependencies:
  - [TASK-015](#task-015-implement-edit-sleep-session-use-case-and-api)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - Queries return durations based on the most recent relevant event
  - Empty dataset returns explicit empty result, not zero masquerading as real data
  - Only family-scoped data is used

### TASK-019: Implement daily summary calculation service
- Size: `S`
- Goal: Produce correct daily totals for sleep and active time.
- Scope:
  - Implement local-day calculation in family timezone
  - Include naps overlapping the day
  - Include the night sleep that started on that day
  - Calculate active time strictly as gaps between sleep sessions
- Important edge cases:
  - Daylight saving transition days
  - Cross-midnight sleep sessions
  - Active session overlapping current day
- Dependencies:
  - [TASK-012](#task-012-implement-timezone-aware-sleep-classification-service)
  - [TASK-017](#task-017-implement-sleep-history-query-and-api)
- Acceptance criteria:
  - Daily sleep totals follow the nap and night-sleep rule from [requirements.md](/home/hjoeftung/code/projects/keklik/requirements.md)
  - Active time is derived strictly from gaps between sessions
  - Tests cover DST forward and backward days

### TASK-020: Implement summary query APIs for today, 7 days, and 14 days
- Size: `S`
- Goal: Expose reporting endpoints for today and rolling averages.
- Scope:
  - Implement `GetTodaySleepSummary`
  - Implement `GetTodayActiveSummary`
  - Implement `GetRollingAverageSummary`
  - Return structured values for sleep time, active time, and period metadata
- Dependencies:
  - [TASK-019](#task-019-implement-daily-summary-calculation-service)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - Today summary uses family timezone semantics
  - Rolling averages are available for 7-day and 14-day periods
  - Results match the daily summary calculation rules

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
- Dependencies: [TASK-001](#task-001-bootstrap-go-service-skeleton)
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
    - query summaries
- Dependencies:
  - [TASK-020](#task-020-implement-summary-query-apis-for-today-7-days-and-14-days)
  - [TASK-021](#task-021-add-repository-transaction-boundaries-and-concurrency-tests)
- Acceptance criteria:
  - The MVP acceptance criteria in [requirements.md](/home/hjoeftung/code/projects/keklik/requirements.md) are covered by automated tests or explicitly mapped to unit coverage
  - The test suite runs locally with documented setup

## Suggested First Milestone

If you want an initial deliverable that already has end-user value for a simple client, stop after:
- [TASK-001](#task-001-bootstrap-go-service-skeleton)
- [TASK-003](#task-003-add-postgresql-migration-framework-and-baseline-schema)
- [TASK-005](#task-005-model-family-aggregate-and-repository-interfaces)
- [TASK-006](#task-006-implement-create-family-use-case-and-api)
- [TASK-007](#task-007-implement-google-oauth-identity-flow)
- [TASK-008](#task-008-implement-edit-family-use-case-and-api)
- [TASK-011](#task-011-model-sleep-session-aggregate-and-repository-interfaces)
- [TASK-012](#task-012-implement-timezone-aware-sleep-classification-service)
- [TASK-013](#task-013-implement-start-sleep-use-case-and-api)
- [TASK-014](#task-014-implement-stop-sleep-use-case-and-api)
- [TASK-017](#task-017-implement-sleep-history-query-and-api)

That milestone would support:
- authentication
- family creation
- family settings
- sleep start and stop
- basic sleep history

## Notes for Planning

- Keep tasks small. If any task starts expanding beyond roughly 2 days, split by transport layer, application service, or repository work.
- Prefer finishing domain model and tests before wiring HTTP handlers.
- Treat timezone and daylight saving tests as mandatory, not optional cleanup.
- Preserve the stored classification result on historical sleep sessions so later family-setting changes do not rewrite history implicitly.