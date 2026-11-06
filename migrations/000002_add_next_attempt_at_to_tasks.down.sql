DROP INDEX IF EXISTS idx_tasks_pending_next_attempt;
ALTER TABLE tasks DROP COLUMN IF EXISTS next_attempt_at;
