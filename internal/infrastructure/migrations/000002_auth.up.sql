CREATE TABLE accounts (
    id                  UUID        PRIMARY KEY,
    google_subject_id   TEXT        NOT NULL UNIQUE,
    email               TEXT        NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_google_subject_id ON accounts (google_subject_id);

CREATE TABLE sessions (
    token       TEXT        PRIMARY KEY,
    account_id  UUID        NOT NULL REFERENCES accounts(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_account_id ON sessions (account_id);
