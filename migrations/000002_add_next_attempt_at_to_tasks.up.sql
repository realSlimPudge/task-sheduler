ALTER TABLE tasks
ADD COLUMN next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX idx_tasks_pending_next_attempt
ON tasks (next_attempt_at)
WHERE status = 'pending';
