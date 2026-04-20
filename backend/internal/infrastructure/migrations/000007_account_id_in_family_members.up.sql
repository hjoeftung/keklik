ALTER TABLE family_members
    RENAME COLUMN google_subject_id TO account_id;

ALTER TABLE family_members
    ADD CONSTRAINT fk_family_members_account_id
        FOREIGN KEY (account_id) REFERENCES accounts (id);

ALTER INDEX idx_family_members_google_subject_id
    RENAME TO idx_family_members_account_id;
