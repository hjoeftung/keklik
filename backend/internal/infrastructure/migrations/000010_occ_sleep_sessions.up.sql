-- btree_gist is required to mix btree operators (= on baby_id) with gist operators
-- (&& on tstzrange) in a single exclusion constraint.
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- OCC version counter: starts at 1 for every existing row.
ALTER TABLE sleep_sessions ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

-- Prevent overlapping sleep sessions for the same baby at the database level.
-- Active sessions (stopped_at IS NULL) are treated as open-ended via 'infinity'.
ALTER TABLE sleep_sessions
    ADD CONSTRAINT sleep_sessions_no_overlap
    EXCLUDE USING GIST (
        baby_id WITH =,
        tstzrange(started_at, COALESCE(stopped_at, 'infinity'::TIMESTAMPTZ), '[)') WITH &&
    );
