-- Индекс для быстрого поиска задач по userid
CREATE INDEX IF NOT EXISTS idx_tasks_userid ON tasks(userid);

-- Индекс для поиска по статусу
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
