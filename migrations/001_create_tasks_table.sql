-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(50) NOT NULL DEFAULT 'medium',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);

---- create above / drop below ----

-- Drop indexes
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_status;

-- Drop table
DROP TABLE IF EXISTS tasks;
