## Story 7: Settings

### TASK-065: Build settings screen with editable family fields
- Size: `S`
- Goal: Let users update their name, baby name, timezone, and night window from a single form.
- Scope:
  - `src/screens/Settings/SettingsForm.tsx`
  - Load current values from a family/profile endpoint on mount
  - Fields: user name, baby name, timezone (searchable dropdown), night window start/end (time inputs)
  - A single "Save" button submits all changes explicitly (no auto-save)
  - Preserve form state on API error; show success confirmation on save
- Dependencies:
  - [TASK-048](#task-048-implement-api-client-with-auth-header-error-normalization-and-401-redirect)
  - [TASK-056](#task-056-implement-persistent-navigation-bar)
- Acceptance criteria:
  - All four field types editable and pre-populated from the API
  - Save button disabled while submitting
  - API errors shown inline without losing form state

### TASK-066: Implement invite link generation and copyable display in settings
- Size: `S`
- Goal: Let family members generate a new invite link and copy it to the clipboard.
- Scope:
  - Add an "Invite link" section to the settings screen
  - "Generate new link" button calls `POST /families/invite-links`
  - Display the returned URL in a read-only text input with a "Copy" button that uses the Clipboard API
  - Show feedback on successful copy ("Copied!")
  - Warn the user that generating a new link invalidates the previous one (if that is the backend behavior)
- Dependencies:
  - [TASK-065](#task-065-build-settings-screen-with-editable-family-fields)
- Acceptance criteria:
  - Generated link appears and is copyable with one tap
  - "Copied!" feedback appears after successful clipboard write
  - If Clipboard API is unavailable, the link remains selectable for manual copy
