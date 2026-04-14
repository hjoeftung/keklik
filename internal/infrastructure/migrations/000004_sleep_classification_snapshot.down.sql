ALTER TABLE sleep_sessions
    ADD COLUMN classification_rule_version INTEGER NOT NULL DEFAULT 0,
    DROP COLUMN classified_with_nw_start_hour,
    DROP COLUMN classified_with_nw_start_minute,
    DROP COLUMN classified_with_nw_end_hour,
    DROP COLUMN classified_with_nw_end_minute;
