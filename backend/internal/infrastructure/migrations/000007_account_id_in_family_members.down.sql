DROP INDEX IF EXISTS idx_family_members_account_id;

ALTER TABLE family_members
    ADD COLUMN google_subject_id TEXT NOT NULL DEFAULT '';

ALTER TABLE family_members
    DROP COLUMN account_id;

CREATE UNIQUE INDEX idx_family_members_google_subject_id ON family_members (google_subject_id);
