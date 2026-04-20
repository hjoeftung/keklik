CREATE TABLE sessions (
    token       TEXT        PRIMARY KEY,
    account_id  UUID        NOT NULL REFERENCES accounts(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_account_id ON sessions (account_id);
