## Story 2: Family Onboarding

### TASK-005: Model family aggregate and repository interfaces
- Status: Done
- Size: `S`
- Goal: Define the family domain model before implementing commands.
- Scope:
  - Add `Family`, `FamilyMember`, `Baby`, `NightWindow`, and `InviteLink` domain types
  - Define repository interfaces
  - Encode MVP invariant of exactly one baby per family while keeping future extension possible
- Boundary note:
  - `FamilyMember` belongs to the family domain and represents membership in a family
  - `Account` belongs to the auth domain and represents an authenticated identity
  - A `FamilyMember` may be linked to an `Account`, but the concepts should remain separate in code and docs
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
- Status: Done
- Size: `S`
- Goal: Allow creation of a family with the initial baby and settings.
- Scope:
  - Implement `CreateFamily`
  - Validate family name, baby name, timezone, and night window
  - Persist family, creator family member, and initial baby in one transaction
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
  - Keep auth `Account` separate from family-domain `FamilyMember`
  - Define how an authenticated `Account` is linked to an existing or newly created `FamilyMember`
  - Define how authenticated identity is attached to requests
- Decision notes:
  - Keep session mechanism simple for MVP
  - Preserve Google subject identifier in account data
  - Avoid leaking auth terminology into the family aggregate
- Dependencies:
  - [TASK-002](#task-002-add-configuration-and-environment-contract)
  - [TASK-004](#task-004-establish-shared-error-model-and-http-error-mapping)
- Acceptance criteria:
  - Unauthenticated requests are rejected on protected endpoints
  - Google-authenticated identity can be resolved to internal account data
  - Authenticated account-to-family-member linking is explicit in the application flow
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
  - Link the authenticated `Account` to the family by creating or associating a `FamilyMember`
  - Prevent duplicate membership
  - Expose endpoint
- Dependencies:
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
  - [TASK-009](#task-009-implement-family-invite-link-creation)
  - [TASK-005](#task-005-model-family-aggregate-and-repository-interfaces)
- Acceptance criteria:
  - Only a Google-authenticated user can accept an invite
  - Invalid or expired links are rejected
  - A valid invite links the authenticated account to one family member exactly once
