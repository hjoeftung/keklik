## Story 4: Reporting and Summaries

### TASK-020: Implement GetDashboardSummary query and API
- Size: `M`
- Goal: Expose a single dashboard endpoint returning all metrics needed for one screen load.
- Scope:
  - Implement `GetDashboardSummary` use case
  - Response must include:
    - Active session (if any)
    - Time since last sleep start
    - Time since last awakening
    - Today's total sleep time
    - Today's total active time
    - 7-day average daily sleep time
    - 7-day average daily active time
    - 14-day average daily sleep time
    - 14-day average daily active time
  - Expose as a single HTTP endpoint to avoid multiple round-trips per screen load
- Dependencies:
  - [TASK-018](#task-018-implement-elapsed-time-calculation-service)
  - [TASK-019](#task-019-implement-daily-summary-calculation-service)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - Single endpoint returns all dashboard metrics in one response
  - Today's totals use family timezone semantics
  - Rolling averages are returned for both 7-day and 14-day periods
  - Elapsed-time fields return explicit empty values when no relevant data exists
  - All values are scoped to the caller's family only
