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

