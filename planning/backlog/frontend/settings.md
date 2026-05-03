# Settings backlog

The Settings screen (`frontend/src/screens/Settings/SettingsScreen.tsx`) currently has a design shell with no live functionality. These stories implement each row in order of the sections shown.

---

## SETTINGS-6 — Accept invite flow

Let a new caregiver join a family by opening an invite link.

**Context**

`InviteScreen.tsx` at `frontend/src/screens/Invite/InviteScreen.tsx` is a stub. The route extracts `token` from the URL. The backend endpoint `POST /families/join-by-invite-link` accepts `{ token: string, member_name: string }` and returns the new family.

This screen is reached by an unauthenticated (or already-authenticated) user clicking the invite URL. If the user is not logged in they must authenticate first — check `AuthContext` and redirect to `/` (login) with the invite token preserved so they can be redirected back after auth.

**What to change**

1. **Complete `InviteScreen.tsx`**:
   - If not authenticated, store the invite token in `sessionStorage` then redirect to `/`.
   - If authenticated, show a "You've been invited to join [baby name]'s family" card. Show a name input field pre-populated with the current user's name if known.
   - On submit, call `POST /families/join-by-invite-link` with `{ token, member_name }`. On success call `refreshFamily()` and navigate to `/`.
   - Handle error states: expired token (show a friendly message), already a member (redirect to home), invalid token (show error).

2. **Post-login redirect** — in `AuthCallbackScreen.tsx`, after successful auth check `sessionStorage` for a pending invite token. If found, redirect to `/invite/{token}` instead of `/`.

---

## SETTINGS-7 — Help link

Provide caregivers with a way to get support.

**Context**

The "Help" row is a stub chevron. No help content exists yet.

**What to change**

Keep this minimal for now. Replace the chevron row with a plain anchor tag (`<a href="mailto:support@keklik.app" target="_blank">`) or a link to a future help page. No new screen needed.
