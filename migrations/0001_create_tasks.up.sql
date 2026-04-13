CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	recurrence_type TEXT NOT NULL DEFAULT 'none',
	recurrence_config JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
