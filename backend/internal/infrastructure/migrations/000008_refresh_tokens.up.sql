CREATE TABLE refresh_tokens (
    token      UUID        PRIMARY KEY,
    account_id UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX refresh_tokens_account_id_idx ON refresh_tokens (account_id);
