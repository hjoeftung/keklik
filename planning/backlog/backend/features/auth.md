### TASK-033: Add member roles (viewer vs editor) for baby sleep operations
- Size: `M`
- Goal: Let families distinguish read-only observers from members who can log and edit sleep sessions. Permission enforcement lives in the HTTP layer introduced by TASK-032.
- Background:
  - Today all family members have symmetric write access. The intended use case is that one primary caregiver logs sleep while grandparents or other observers can view. There is no mechanism to express this distinction today.
- Scope:
  - Add `role` column to `family_members`: enum `admin` | `viewer`, not null. Default existing rows to `admin`.
  - `requireBabyAccess` middleware (TASK-032) attaches the caller's role to context alongside `(baby_id, member_id)`
  - Add a separate `requireEditor` middleware gate that reads the role from context and returns 403 for viewers on mutating endpoints (start, stop, edit, delete sleep)
  - Invite flow: invite creator can choose role for the invitee; default is `viewer`
  - Family creator gets `admin` on creation
  - Update OpenAPI spec and integration tests
- Dependencies: [TASK-032](#task-032-add-babiesbaby_id-to-sleep-endpoints-with-family-membership-middleware)
- Acceptance criteria:
  - A viewer can GET sleep history but receives 403 on POST/PATCH/DELETE sleep endpoints
  - An admin can perform all operations
  - Role is set at invite-link creation and honoured on join
  - Role is visible in a family-member listing endpoint (or returned on join)

