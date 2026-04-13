CREATE TABLE sleep_profiles (
    baby_id                     UUID        PRIMARY KEY REFERENCES babies(id),
    timezone                    TEXT        NOT NULL,
    night_window_start_hour     SMALLINT    NOT NULL CHECK (night_window_start_hour   BETWEEN 0 AND 23),
    night_window_start_minute   SMALLINT    NOT NULL CHECK (night_window_start_minute BETWEEN 0 AND 59),
    night_window_end_hour       SMALLINT    NOT NULL CHECK (night_window_end_hour     BETWEEN 0 AND 23),
    night_window_end_minute     SMALLINT    NOT NULL CHECK (night_window_end_minute   BETWEEN 0 AND 59),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE families DROP COLUMN timezone;
ALTER TABLE families DROP COLUMN night_window_start_hour;
ALTER TABLE families DROP COLUMN night_window_start_minute;
ALTER TABLE families DROP COLUMN night_window_end_hour;
ALTER TABLE families DROP COLUMN night_window_end_minute;
