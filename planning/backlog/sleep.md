## Story 3: Sleep Lifecycle

### TASK-011: Model sleep session aggregate and repository interfaces
- Status: Done
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
