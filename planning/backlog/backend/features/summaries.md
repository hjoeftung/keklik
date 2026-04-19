## Story 4: Reporting and Summaries

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
  - Daily sleep totals follow the nap and night-sleep rule from [requirements.md](requirements.md)
  - Active time is derived strictly from gaps between sessions
  - Tests cover DST forward and backward days

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
