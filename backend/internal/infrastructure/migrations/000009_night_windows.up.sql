CREATE TABLE baby_night_windows (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    baby_id         UUID        NOT NULL REFERENCES babies(id),
    start_hour      SMALLINT    NOT NULL CHECK (start_hour   BETWEEN 0 AND 23),
    start_minute    SMALLINT    NOT NULL CHECK (start_minute BETWEEN 0 AND 59),
    end_hour        SMALLINT    NOT NULL CHECK (end_hour     BETWEEN 0 AND 23),
    end_minute      SMALLINT    NOT NULL CHECK (end_minute   BETWEEN 0 AND 59),
    effective_from  TIMESTAMPTZ NOT NULL,
    effective_to    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_effective_range CHECK (effective_to IS NULL OR effective_to > effective_from)
);

CREATE INDEX idx_baby_night_windows_baby_id ON baby_night_windows (baby_id, effective_from ASC);

INSERT INTO baby_night_windows (baby_id, start_hour, start_minute, end_hour, end_minute, effective_from)
SELECT baby_id,
       night_window_start_hour, night_window_start_minute,
       night_window_end_hour,   night_window_end_minute,
       '1970-01-01T00:00:00Z'
FROM sleep_profiles;

DROP TABLE sleep_profiles;

ALTER TABLE sleep_sessions
    DROP COLUMN classification,
    DROP COLUMN classified_with_nw_start_hour,
    DROP COLUMN classified_with_nw_start_minute,
    DROP COLUMN classified_with_nw_end_hour,
    DROP COLUMN classified_with_nw_end_minute;
