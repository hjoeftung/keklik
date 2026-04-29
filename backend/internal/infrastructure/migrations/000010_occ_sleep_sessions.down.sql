ALTER TABLE sleep_sessions DROP CONSTRAINT IF EXISTS sleep_sessions_no_overlap;
ALTER TABLE sleep_sessions DROP COLUMN IF EXISTS version;
