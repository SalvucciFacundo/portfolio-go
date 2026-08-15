-- 0003_add_project_status.sql — Estado del proyecto: production | development | demo
-- Muestra un badge en la card. Default: development (neutral).

ALTER TABLE projects ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'development';
