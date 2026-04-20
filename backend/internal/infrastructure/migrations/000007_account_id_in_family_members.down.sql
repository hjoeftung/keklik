ALTER INDEX idx_family_members_account_id
    RENAME TO idx_family_members_google_subject_id;

ALTER TABLE family_members
    DROP CONSTRAINT fk_family_members_account_id;

ALTER TABLE family_members
    RENAME COLUMN account_id TO google_subject_id;
