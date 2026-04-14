## Story 3: Sleep Lifecycle

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
