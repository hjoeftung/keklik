ALTER TABLE family_members
    ADD COLUMN account_id UUID REFERENCES accounts(id);

ALTER TABLE family_members
    DROP COLUMN google_subject_id;

CREATE UNIQUE INDEX idx_family_members_account_id ON family_members (account_id);
