# Keklik Backend Requirements

## 1. Purpose

Keklik is a simple baby sleep tracker for families. The first implementation target is the backend only.

The backend must support future clients such as:
- Telegram bot
- Simple web application
- Android application

The backend should be designed using Domain-Driven Design (DDD) and implemented in Go.

## 2. Scope

### In scope for the first version
- Family account management
- Shared access for multiple family members
- Baby sleep tracking
- Night sleep and nap classification
- Sleep history and summary queries
- Backend APIs and domain model

### Out of scope for the first version
- Frontend implementation
- Push notifications and reminders
- Multi-language support
- Advanced analytics beyond the stories in [stories.md](/home/hjoeftung/code/projects/keklik/stories.md)

## 3. Product Goal

The system should allow a family to track a baby's sleep and awake periods from multiple devices while keeping a single shared source of truth.

## 4. Users and Roles

### Family member
- Belongs to a family account
- Can create, start, stop, and edit sleep records for babies in the family
- Can view sleep history and summary metrics for babies in the family
- Has the same permissions as every other family member in the MVP

### Family
- Represents the shared ownership boundary for babies and member accounts
- All family members work with the same underlying data

## 5. Functional Requirements

### 5.1 Family and account management
1. The system must allow creation of a family account.
2. The system must allow multiple accounts to belong to the same family.
3. The system must treat family members as collaborators with access to the same babies and sleep data.
4. The system must require the family's first baby to be created together with the family.
5. The system must allow editing family settings after creation.
6. The MVP must support exactly one baby per family.
7. The domain model must leave room to support more than one baby per family in a future version.
8. The system must allow a family to generate invite links for additional family members.
9. The system must allow a Google-authenticated user to join a family through a valid invite link.

### 5.2 Authentication and identity
1. The MVP must support OAuth sign-in with Google.
2. Google OAuth must be the only authentication option in the MVP.
3. The backend must map authenticated users to family member accounts.
4. All family members must have identical permissions in the MVP.
5. Invite-link acceptance must require successful Google authentication before a user is linked to a family.

### 5.3 Baby sleep tracking
1. A family member must be able to start a baby's sleep session.
2. A family member must be able to stop an active sleep session.
3. A family member must be able to edit an existing sleep session.
4. The system must calculate the duration of each sleep session.
5. The system must prevent more than one active sleep session for the same baby at the same time.
6. The system must reject invalid sleep intervals where stop time is earlier than start time.
7. Sleep edits must be allowed for both active and completed sleep sessions.

### 5.4 Sleep classification
1. A family must define a configurable local time period that represents the baby's night window.
2. The night window must be stored and evaluated in the family's configured timezone.
3. A completed sleep session must be classified as a night sleep when more than half of its duration falls within the configured night window.
4. A completed sleep session that is not classified as a night sleep must be classified as a nap.
5. A sleep session that is still active may have a provisional classification, but final classification must be recalculated when the session is completed or edited.

Additional user story:
- I, as a user, want to configure the baby's night period so the app can distinguish between night sleep and daily naps.

### 5.5 Sleep history queries
1. A family member must be able to retrieve sleep sessions for a selected baby for:
   - Today
   - Last 7 days
   - Last 14 days
2. The system must return sessions ordered by start time descending unless another ordering is explicitly requested.

### 5.6 Time since last event
1. A family member must be able to view the time elapsed since the last sleep started.
2. A family member must be able to view the time elapsed since the last awakening.
3. If there is not enough data to calculate a metric, the system must return an explicit empty result rather than an incorrect duration.

### 5.7 Daily and rolling summaries
1. A family member must be able to view total sleep time for today.
2. A family member must be able to view total active time for today.
3. A family member must be able to view average daily sleep time for the last 7 days.
4. A family member must be able to view average daily active time for the last 7 days.
5. A family member must be able to view average daily sleep time for the last 14 days.
6. A family member must be able to view average daily active time for the last 14 days.
7. Daily summaries must include all naps that overlap the local day according to the family timezone.
8. Daily summaries must include the night sleep that started on that local day.
9. The definition of "today" must be based on the family's configured timezone, including correct handling of daylight saving transitions.
10. Active time must be calculated strictly as the gaps between sleep sessions.

### 5.8 Timezone handling
1. Each family must have a configured IANA timezone identifier.
2. All timestamps must be stored in UTC.
3. All day-based calculations and night-window calculations must use the family's configured timezone.
4. The system must correctly handle daylight saving time changes, including days with 23 or 25 hours.
5. The system must not assume that day boundaries or configured night windows map to fixed UTC offsets throughout the year.
6. Changes to the configured night window must apply only to sleep sessions created or updated after the change takes effect.
7. Changes to the configured night window must not retroactively reclassify historical sleep sessions.

## 6. Domain Model Requirements

The existing notes in [entities.md](/home/hjoeftung/code/projects/keklik/entities.md) imply the following domain concepts.

### 6.1 Aggregates and entities

#### Family aggregate
- Aggregate root: Family
- Fields:
  - FamilyId
  - Name
  - Members
  - Baby
  - Timezone
  - NightWindow
  - InviteLinks
- Responsibilities:
  - Own family membership
  - Own baby registration inside the family boundary
  - Own family settings such as timezone and night window
  - Enforce that only family members can act on family data

#### Account entity
- Fields:
  - AccountId
  - Name
  - GoogleSubjectId
- Responsibilities:
  - Represent a family member identity inside the domain
  - Represent an authenticated user linked from Google OAuth

#### Baby entity
- Fields:
  - BabyId
  - Name
  - FamilyId
- Responsibilities:
  - Represent the tracked child

Note: A Baby entity is required even though it is only implied in [entities.md](/home/hjoeftung/code/projects/keklik/entities.md) via `[]babies`. The MVP stores one baby per family, but the model should not make future multi-baby support impossible.

#### NightWindow value object
- Fields:
  - StartLocalTime
  - EndLocalTime
  - Timezone
- Responsibilities:
  - Represent the family-configured local time range used to classify night sleep
  - Resolve correctly across midnight boundaries and daylight saving transitions

#### InviteLink entity or value object
- Fields:
  - InviteToken
  - FamilyId
  - ExpiresAt
  - CreatedByAccountId
- Responsibilities:
  - Represent a shareable link that allows a new member to join a family
  - Support validation before a Google-authenticated user is linked to the family

#### Sleep aggregate or entity
- Preferred name: SleepSession
- Fields:
  - SleepSessionId
  - BabyId
  - StartTime
  - StopTime
  - Duration
  - Classification
  - ClassificationRuleVersion
  - CreatedByAccountId
  - UpdatedByAccountId
- Responsibilities:
  - Represent one sleep interval
  - Enforce valid lifecycle transitions
  - Support duration recalculation after edits
  - Support nap versus night-sleep classification
  - Preserve the classification decision that applied when the session was created or last edited

### 6.2 Domain invariants
1. A sleep session belongs to exactly one baby.
2. A baby belongs to exactly one family.
3. Only one active sleep session may exist per baby.
4. A completed sleep session must have `StopTime >= StartTime`.
5. Duration must be derived from start and stop times, not manually entered.
6. Accounts may access only data belonging to their family.
7. A family has exactly one baby in the MVP.
8. A family must have a valid timezone and night window configuration.
9. Sleep classification must be derived from the configured night window and the session interval, not entered manually.
10. Historical sleep-session classification must remain stable after later night-window changes unless the session itself is edited.

## 7. Bounded Contexts

### 7.1 Family Management
- Manage families
- Manage accounts within a family
- Manage the baby created with the family
- Manage family timezone and night-window settings
- Manage family invite links

### 7.2 Sleep Tracking
- Start sleep
- Stop sleep
- Edit sleep sessions
- Enforce sleep-related domain invariants

### 7.3 Reporting
- Query sleep history
- Query nap and night-sleep distinctions
- Query elapsed time since last sleep or awakening
- Query daily totals and 7-day and 14-day averages

For a first version, these contexts may live in one deployable backend service, but they should remain separate at the domain and application layers.

## 8. Application Layer Requirements

The backend should expose application use cases aligned with the stories.

### Required commands
- CreateFamily
- AddFamilyMember
- CreateFamilyInviteLink
- JoinFamilyByInviteLink
- EditFamily
- StartSleep
- StopSleep
- EditSleepSession

### Required queries
- GetSleepSessionsByPeriod
- GetTimeSinceLastSleepStart
- GetTimeSinceLastAwakening
- GetTodaySleepSummary
- GetTodayActiveSummary
- GetRollingAverageSummary

CreateFamily must include:
- Family name
- Initial baby data
- Family timezone
- Night-window configuration

## 9. API Requirements

The first version should provide a simple client-agnostic API suitable for Telegram, web, and Android clients.

### API style
- Preferred: HTTP JSON API
- Alternative: gRPC internally with HTTP gateway if needed later

### Minimum API capabilities
- Google OAuth authentication
- Family creation and family membership management
- Invite-link creation and invite-link acceptance
- Family editing for timezone, night-window, and baby details
- Start, stop, and edit sleep sessions
- Retrieve sleep sessions by date range
- Retrieve summaries and elapsed time metrics

### API behavior requirements
- All timestamps must be stored in UTC.
- The API must require a family timezone for day-based summaries and night-window calculations.
- The API must use IANA timezone identifiers.
- The API must behave correctly across daylight saving time changes and timezone offset changes.
- Errors must be returned with stable machine-readable codes.
- Query endpoints must be idempotent.
- Command endpoints must validate business rules before persistence.
- The authentication layer must accept Google OAuth identity and resolve it to an internal account.

## 10. Persistence Requirements

1. The system must persist families, accounts, babies, family settings, and sleep sessions.
2. The storage design must support querying by baby and date range efficiently.
3. The storage design must support detection of an active sleep session for a baby.
4. The system should keep audit metadata for create and update operations.
5. The system may hard-delete sleep sessions in the MVP.
6. The storage design must preserve the data needed for timezone-aware and daylight-saving-aware calculations.
7. The storage design must support unrestricted hard deletion of sleep sessions.
8. The storage design must preserve the classification result used for historical sleep sessions so later night-window changes do not alter past records implicitly.

Preferred initial choice:
- PostgreSQL

Rationale:
- Strong consistency for shared family data
- Good support for date-range queries and indexing
- Simple operational model for a first backend

## 11. Non-Functional Requirements

### Architecture
- The codebase must be structured according to DDD layers:
  - Domain
  - Application
  - Infrastructure
  - Interfaces
- Domain logic must not depend on transport or persistence frameworks.
- Business rules must be expressed in domain types and domain services, not in handlers.

### Language and implementation
- The backend must be implemented in Go.
- The code should prefer standard library and small, explicit dependencies.
- The project should be easy to run locally for a single developer.

### Reliability
- Commands that change sleep state must be atomic.
- The system must behave correctly when multiple family members act nearly simultaneously.
- The backend must protect against duplicate active sleep sessions caused by race conditions.
- Time calculations must remain correct across daylight saving transitions and timezone changes.

### Observability
- The backend should log command execution failures and domain validation errors.
- The backend should expose basic health checks.

## 12. Recommended Go Project Direction

The implementation should favor a modular monolith.

Recommended high-level package layout:
- `internal/family/domain`
- `internal/family/application`
- `internal/auth/application`
- `internal/sleep/domain`
- `internal/sleep/application`
- `internal/reporting/application`
- `internal/infrastructure/persistence`
- `internal/infrastructure/oauth`
- `internal/interfaces/http`

This keeps the first version simple while preserving clear domain boundaries.

## 13. Acceptance Criteria

The first backend version is acceptable when:
1. A family can be created.
2. Multiple accounts can operate under the same family.
3. The initial baby is created together with the family.
4. Families can create invite links for additional family members.
5. A Google-authenticated user can join a family through a valid invite link.
6. A family member can start a sleep session for a baby.
7. A family member can stop that sleep session.
8. A family member can edit a sleep session.
9. The system rejects overlapping active sleep sessions for the same baby.
10. The system returns sleep history for 1-day, 7-day, and 14-day windows.
11. The system returns time since last sleep start and last awakening.
12. The system distinguishes naps from night sleep using the configured night window.
13. The system returns today's sleep and active totals using the family timezone and night-sleep rules.
14. Active time is calculated strictly as the gaps between sleep sessions.
15. The system handles daylight saving transitions correctly for daily summaries and sleep classification.
16. Night-window changes do not retroactively change historical sleep-session classifications.
17. Users can authenticate with Google OAuth.
18. All family members have identical permissions.
19. Sleep sessions can be hard-deleted without role restrictions.
20. The system returns 7-day and 14-day average sleep and active durations.
21. All operations respect family data boundaries.
