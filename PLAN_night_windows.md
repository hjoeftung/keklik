# Plan: NightWindow Refactor

## Status

Implemented and verified.

Implementation order status:
- [x] 1. SQL migration (`000009`)
- [x] 2. Domain: `NightWindow` fields + `FindWindowForSession` + remove `SleepProfile`
- [x] 3. Domain: `Baby.NightWindows`, `NightWindowRepository` interface
- [x] 4. Infrastructure: `PostgresNightWindowRepository`
- [x] 5. Infrastructure: update `PostgresFamilyRepository` (JOIN baby_night_windows)
- [x] 6. Infrastructure: update `PostgresSleepSessionRepository` (drop classification columns)
- [x] 7. Use cases: `SetNightWindow`, update `StopSleep`, update `GetSleepHistory`
- [x] 8. HTTP handlers: route rename, add `?timezone=`, drop timezone from bodies
- [x] 9. Wire up in `main.go`
- [x] 10. Tests

## Summary of changes

1. **Timezone removed from storage entirely** — accepted only as a query param on read endpoints
2. **NightWindow becomes a versioned entity on Baby** — replaces SleepProfile
3. **SleepSession stores no classification or night window snapshot** — both are computed on-the-fly at read time
4. **SleepProfile entity disappears**

---

## Domain model changes

### `sleep/sleep.go`

**Remove entirely:**
- `SleepProfile` struct and `NewSleepProfile()`
- `SleepProfileRepository` interface
- `classifiedWithNightWindow *NightWindow` field on `SleepSession`
- `classification SleepClassification` field on `SleepSession`
- `ClassifiedWithNightWindow()` accessor
- `Classification()` accessor
- `Stop(stoppedAt, classification, classifiedWith)` — replace with `Stop(stoppedAt time.Time)`

**Modify `NightWindow`:**
```go
type NightWindow struct {
    id            NightWindowID   // new
    babyID        BabyID          // new
    start         LocalTime
    end           LocalTime
    effectiveFrom time.Time       // new
    effectiveTo   *time.Time      // new, nil = currently active
}
```

Add `NightWindowID` type alias (string/UUID).

**New repository interface:**
```go
type NightWindowRepository interface {
    Save(ctx context.Context, nw NightWindow) error
    DeleteByIDs(ctx context.Context, ids []NightWindowID) error
    FindByBabyID(ctx context.Context, babyID BabyID) ([]NightWindow, error)
}
```

**Classification stays pure** — `Classify(session SleepSession, timezone string, nw NightWindow)` signature unchanged. Callers are responsible for finding the right NightWindow.

Add helper:
```go
// FindWindowForSession returns the NightWindow whose effective range covers session.startedAt.
func FindWindowForSession(windows []NightWindow, session SleepSession) (NightWindow, bool)
```

### `family/family.go`

**Modify `Baby`:**
```go
type Baby struct {
    ID           BabyID
    FamilyID     FamilyID
    Name         string
    NightWindows []sleep.NightWindow   // ordered by effective_from ASC
}
```

`Family` aggregate loads babies with their full NightWindows collection via `FamilyRepository`.

---

## Use case changes

### `StopSleep`
- Remove `timezone` param from input
- Remove classification logic entirely
- `session.Stop(stoppedAt)` — just records the time
- No NightWindow lookup needed

### `GetSleepHistory`
- Add `timezone string` to input
- After loading sessions, load baby's NightWindows (via `NightWindowRepository` or from baby aggregate)
- For each session: call `FindWindowForSession(windows, session)` then `Classify(session, timezone, nw)`
- Return sessions with computed `classification`

### `EditSleepSession`
- If `stopped_at` changes, no reclassification is stored (classification is always computed at read time, nothing to update)

### `CreateSleepProfile` → **`SetNightWindow`** (rename + rework)
- Input: `babyID`, `nightWindow{start, end}`, `effectiveFrom time.Time`
- Replacement logic (enforced in application layer):
  1. Load existing windows for baby (ordered by `effective_from` ASC)
  2. Delete all windows where `effective_from >= input.effectiveFrom`
  3. Find the window immediately before `input.effectiveFrom` (if any) and set its `effective_to = input.effectiveFrom`
  4. Save the new window with `effective_to = nil`
- No `timezone` param

---

## HTTP handler changes

### Write endpoints — remove timezone

| Endpoint | Change |
|---|---|
| `POST /babies/{id}/sleep-profiles` → `POST /babies/{id}/night-windows` | Remove `timezone` from body; body becomes `{start_hour, start_minute, end_hour, end_minute, effective_from}` |
| `DELETE /babies/{id}/sleep-sessions/active` | Remove `timezone` from body (body stays `{stopped_at?}`) |

### Read endpoints — add `?timezone=`

| Endpoint | Change |
|---|---|
| `GET /babies/{id}/sleep-sessions` | Add required `?timezone=America/New_York` query param; classify sessions on the fly before responding |
| `DELETE /babies/{id}/sleep-sessions/active` (stop) | No longer returns `classification` in response (or returns it as `""`) |

---

## Infrastructure changes

### New: `PostgresNightWindowRepository`
Implements `NightWindowRepository`. Table: `baby_night_windows`.

### Remove: `PostgresSleepProfileRepository`
Delete the file entirely.

### Update: `PostgresFamilyRepository`
- `reconstruct()` LEFT JOINs `baby_night_windows` on `babies.id`
- Deduplicates and assembles `Baby.NightWindows []NightWindow` (similar to existing member/baby deduplication)

### Update: `PostgresSleepSessionRepository`
- `Save()` no longer writes `classification` or `classified_with_nw_*` columns
- `scan*` functions no longer read those columns
- Response DTOs compute `classification` at the HTTP layer, not from DB

---

## Database migration (`000009_night_windows`)

```sql
-- up

CREATE TABLE baby_night_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    baby_id UUID NOT NULL REFERENCES babies(id),
    start_hour SMALLINT NOT NULL CHECK (start_hour BETWEEN 0 AND 23),
    start_minute SMALLINT NOT NULL CHECK (start_minute BETWEEN 0 AND 59),
    end_hour SMALLINT NOT NULL CHECK (end_hour BETWEEN 0 AND 23),
    end_minute SMALLINT NOT NULL CHECK (end_minute BETWEEN 0 AND 59),
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT chk_effective_range CHECK (effective_to IS NULL OR effective_to > effective_from)
);

CREATE INDEX idx_baby_night_windows_baby_id ON baby_night_windows (baby_id, effective_from ASC);

-- Migrate existing sleep_profiles → baby_night_windows (epoch sentinel for "always")
INSERT INTO baby_night_windows (baby_id, start_hour, start_minute, end_hour, end_minute, effective_from)
SELECT baby_id,
       night_window_start_hour, night_window_start_minute,
       night_window_end_hour, night_window_end_minute,
       '1970-01-01T00:00:00Z'
FROM sleep_profiles;

DROP TABLE sleep_profiles;

ALTER TABLE sleep_sessions
    DROP COLUMN classification,
    DROP COLUMN classified_with_nw_start_hour,
    DROP COLUMN classified_with_nw_start_minute,
    DROP COLUMN classified_with_nw_end_hour,
    DROP COLUMN classified_with_nw_end_minute;
```
