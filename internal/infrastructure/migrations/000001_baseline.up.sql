CREATE TABLE families (
    id                          UUID        PRIMARY KEY,
    name                        TEXT        NOT NULL,
    timezone                    TEXT        NOT NULL,
    night_window_start_hour     SMALLINT    NOT NULL CHECK (night_window_start_hour   BETWEEN 0 AND 23),
    night_window_start_minute   SMALLINT    NOT NULL CHECK (night_window_start_minute BETWEEN 0 AND 59),
    night_window_end_hour       SMALLINT    NOT NULL CHECK (night_window_end_hour     BETWEEN 0 AND 23),
    night_window_end_minute     SMALLINT    NOT NULL CHECK (night_window_end_minute   BETWEEN 0 AND 59),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE family_members (
    id                  UUID        PRIMARY KEY,
    family_id           UUID        NOT NULL REFERENCES families(id),
    name                TEXT        NOT NULL,
    google_subject_id   TEXT        NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_family_members_google_subject_id ON family_members (google_subject_id);
CREATE INDEX idx_family_members_family_id ON family_members (family_id);

CREATE TABLE babies (
    id          UUID        PRIMARY KEY,
    family_id   UUID        NOT NULL REFERENCES families(id),
    name        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_babies_family_id ON babies (family_id);

CREATE TABLE invite_links (
    token                   TEXT        PRIMARY KEY,
    family_id               UUID        NOT NULL REFERENCES families(id),
    created_by_member_id    UUID        NOT NULL REFERENCES family_members(id),
    expires_at              TIMESTAMPTZ NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invite_links_family_id ON invite_links (family_id);

CREATE TABLE sleep_sessions (
    id                          UUID        PRIMARY KEY,
    baby_id                     UUID        NOT NULL REFERENCES babies(id),
    started_at                  TIMESTAMPTZ NOT NULL,
    stopped_at                  TIMESTAMPTZ,
    classification              TEXT        NOT NULL DEFAULT '',
    classification_rule_version INTEGER     NOT NULL DEFAULT 0,
    created_by_member_id        UUID        NOT NULL REFERENCES family_members(id),
    updated_by_member_id        UUID        REFERENCES family_members(id),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_sleep_stopped_after_started CHECK (stopped_at IS NULL OR stopped_at >= started_at)
);

-- Partial index for active sleep lookups: quickly find the one active session per baby.
CREATE UNIQUE INDEX idx_sleep_sessions_one_active_per_baby ON sleep_sessions (baby_id) WHERE stopped_at IS NULL;

-- Composite index for date-range history queries ordered by start time descending.
CREATE INDEX idx_sleep_sessions_baby_started_at ON sleep_sessions (baby_id, started_at DESC);
