ALTER TABLE sleep_sessions
    ADD COLUMN classification              TEXT     NOT NULL DEFAULT '',
    ADD COLUMN classified_with_nw_start_hour   SMALLINT CHECK (classified_with_nw_start_hour   BETWEEN 0 AND 23),
    ADD COLUMN classified_with_nw_start_minute SMALLINT CHECK (classified_with_nw_start_minute BETWEEN 0 AND 59),
    ADD COLUMN classified_with_nw_end_hour     SMALLINT CHECK (classified_with_nw_end_hour     BETWEEN 0 AND 23),
    ADD COLUMN classified_with_nw_end_minute   SMALLINT CHECK (classified_with_nw_end_minute   BETWEEN 0 AND 59);

CREATE TABLE sleep_profiles (
    baby_id                     UUID        PRIMARY KEY REFERENCES babies(id),
    timezone                    TEXT        NOT NULL DEFAULT 'UTC',
    night_window_start_hour     SMALLINT    NOT NULL CHECK (night_window_start_hour   BETWEEN 0 AND 23),
    night_window_start_minute   SMALLINT    NOT NULL CHECK (night_window_start_minute BETWEEN 0 AND 59),
    night_window_end_hour       SMALLINT    NOT NULL CHECK (night_window_end_hour     BETWEEN 0 AND 23),
    night_window_end_minute     SMALLINT    NOT NULL CHECK (night_window_end_minute   BETWEEN 0 AND 59),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO sleep_profiles (baby_id, timezone, night_window_start_hour, night_window_start_minute, night_window_end_hour, night_window_end_minute)
SELECT DISTINCT ON (baby_id) baby_id, 'UTC', start_hour, start_minute, end_hour, end_minute
FROM baby_night_windows
ORDER BY baby_id, effective_from DESC;

DROP TABLE baby_night_windows;
