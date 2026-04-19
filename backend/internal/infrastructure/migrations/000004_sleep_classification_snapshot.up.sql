ALTER TABLE sleep_sessions
    ADD COLUMN classified_with_nw_start_hour   SMALLINT CHECK (classified_with_nw_start_hour   BETWEEN 0 AND 23),
    ADD COLUMN classified_with_nw_start_minute SMALLINT CHECK (classified_with_nw_start_minute BETWEEN 0 AND 59),
    ADD COLUMN classified_with_nw_end_hour     SMALLINT CHECK (classified_with_nw_end_hour     BETWEEN 0 AND 23),
    ADD COLUMN classified_with_nw_end_minute   SMALLINT CHECK (classified_with_nw_end_minute   BETWEEN 0 AND 59),
    DROP COLUMN classification_rule_version;
