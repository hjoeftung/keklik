## Story 3: Onboarding

### TASK-054: Build create-family onboarding screen
- Size: `S`
- Goal: Let a signed-in user without a family create one by providing the required details.
- Scope:
  - `src/screens/Onboarding/` — form with: user name, baby name, night window start/end (time inputs), timezone (searchable dropdown defaulting to `Intl.DateTimeFormat().resolvedOptions().timeZone`)
  - On submit, call `POST /families` then `POST /sleep-profiles` as needed; redirect to `/dashboard` on success
  - Preserve form state on API failure so user can retry
  - Show field-level validation errors from the API
- Dependencies:
  - [TASK-048](#task-048-implement-api-client-with-auth-header-error-normalization-and-401-redirect)
  - [TASK-050](#task-050-implement-route-guards-and-auth-context-provider)
- Acceptance criteria:
  - Timezone dropdown is searchable; browser timezone is pre-selected
  - Successful submission redirects to `/dashboard`
  - Form state is not lost on a 422 response

### TASK-055: Build join-family confirmation screen
- Size: `S`
- Goal: Show users arriving via an invite link which family they are about to join, with confirm and cancel actions.
- Scope:
  - `src/screens/JoinFamily/` — read `:token` from the URL, show family name and baby name (from a lightweight invite-info API call or from the join response preview if the backend supports it)
  - "Confirm" calls `POST /families/join-by-invite-link`; on success redirect to `/dashboard`
  - "Cancel" redirects to `/onboarding` (create-family path)
  - If the token is invalid or expired, show an explanatory error and offer "Create a new family instead"
- Dependencies:
  - [TASK-048](#task-048-implement-api-client-with-auth-header-error-normalization-and-401-redirect)
  - [TASK-053](#task-053-preserve-invite-link-url-across-oauth-redirect)
- Acceptance criteria:
  - Valid token shows family details and two actions
  - Confirming joins the family and lands on `/dashboard`
  - Invalid/expired token shows an error with a link to `/onboarding`
