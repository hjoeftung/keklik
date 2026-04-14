## Story 2: Family Onboarding


### TASK-008: Implement edit family use case and API
- Size: `S`
- Goal: Allow family settings updates after creation.
- Scope:
  - Implement `EditFamily`
  - Support updating family name, baby name
  - Expose HTTP endpoint
- Dependencies:
  - [TASK-006](#task-006-implement-create-family-use-case-and-api)
  - [TASK-007](#task-007-implement-google-oauth-identity-flow)
- Acceptance criteria:
  - Family settings can be updated by any family member

### TASK-008a: Implement update night window

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
