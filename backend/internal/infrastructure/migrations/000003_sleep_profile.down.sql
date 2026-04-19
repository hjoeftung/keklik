DROP TABLE sleep_profiles;

ALTER TABLE families ADD COLUMN timezone                    TEXT     NOT NULL DEFAULT 'UTC';
ALTER TABLE families ADD COLUMN night_window_start_hour     SMALLINT NOT NULL DEFAULT 20 CHECK (night_window_start_hour   BETWEEN 0 AND 23);
ALTER TABLE families ADD COLUMN night_window_start_minute   SMALLINT NOT NULL DEFAULT 0  CHECK (night_window_start_minute BETWEEN 0 AND 59);
ALTER TABLE families ADD COLUMN night_window_end_hour       SMALLINT NOT NULL DEFAULT 6  CHECK (night_window_end_hour     BETWEEN 0 AND 23);
ALTER TABLE families ADD COLUMN night_window_end_minute     SMALLINT NOT NULL DEFAULT 0  CHECK (night_window_end_minute   BETWEEN 0 AND 59);
