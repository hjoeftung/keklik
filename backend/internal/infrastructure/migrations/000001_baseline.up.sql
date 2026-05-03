CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE families (
    id          UUID        PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE accounts (
    id                  UUID    PRIMARY KEY,
    google_subject_id   TEXT    NOT NULL UNIQUE,
    email               TEXT    NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_google_subject_id ON accounts (google_subject_id);

CREATE TABLE family_members (
    id          UUID        PRIMARY KEY,
    family_id   UUID        NOT NULL REFERENCES families(id),
    name        TEXT        NOT NULL,
    account_id  UUID        REFERENCES accounts(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_family_members_account_id ON family_members (account_id);
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

CREATE TABLE refresh_tokens (
    token       UUID        PRIMARY KEY,
    account_id  UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ
);

CREATE INDEX refresh_tokens_account_id_idx ON refresh_tokens (account_id);

CREATE TABLE baby_night_windows (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    baby_id         UUID        NOT NULL REFERENCES babies(id),
    start_hour      SMALLINT    NOT NULL CHECK (start_hour   BETWEEN 0 AND 23),
    start_minute    SMALLINT    NOT NULL CHECK (start_minute BETWEEN 0 AND 59),
    end_hour        SMALLINT    NOT NULL CHECK (end_hour     BETWEEN 0 AND 23),
    end_minute      SMALLINT    NOT NULL CHECK (end_minute   BETWEEN 0 AND 59),
    effective_from  TIMESTAMPTZ NOT NULL,
    effective_to    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_effective_range CHECK (effective_to IS NULL OR effective_to > effective_from)
);

CREATE INDEX idx_baby_night_windows_baby_id ON baby_night_windows (baby_id, effective_from ASC);

CREATE TABLE sleep_sessions (
    id                      UUID        PRIMARY KEY,
    baby_id                 UUID        NOT NULL REFERENCES babies(id),
    started_at              TIMESTAMPTZ NOT NULL,
    stopped_at              TIMESTAMPTZ,
    created_by_member_id    UUID        NOT NULL REFERENCES family_members(id),
    updated_by_member_id    UUID        REFERENCES family_members(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    version                 BIGINT      NOT NULL DEFAULT 1,
    CONSTRAINT chk_sleep_stopped_after_started CHECK (stopped_at IS NULL OR stopped_at >= started_at)
);

-- Partial index for active sleep lookups: quickly find the one active session per baby.
CREATE UNIQUE INDEX idx_sleep_sessions_one_active_per_baby ON sleep_sessions (baby_id) WHERE stopped_at IS NULL;

-- Composite index for date-range history queries ordered by start time descending.
CREATE INDEX idx_sleep_sessions_baby_started_at ON sleep_sessions (baby_id, started_at DESC);

-- Prevent overlapping sleep sessions for the same baby at the database level.
-- Active sessions (stopped_at IS NULL) are treated as open-ended via 'infinity'.
ALTER TABLE sleep_sessions
    ADD CONSTRAINT sleep_sessions_no_overlap
    EXCLUDE USING GIST (
        baby_id WITH =,
        tstzrange(started_at, COALESCE(stopped_at, 'infinity'::TIMESTAMPTZ), '[)') WITH &&
    );
