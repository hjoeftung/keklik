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

