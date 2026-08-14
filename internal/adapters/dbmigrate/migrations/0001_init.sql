-- 0001_init.sql — Esquema inicial del portafolio
-- Contrato: api_contract.md

BEGIN;

-- ==========================================================================
-- Profile (una sola fila, id fijo 1)
-- ==========================================================================
CREATE TABLE profile (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL DEFAULT '',
    role_es         TEXT NOT NULL DEFAULT '',
    role_en         TEXT NOT NULL DEFAULT '',
    headline_es     TEXT NOT NULL DEFAULT '',
    headline_en     TEXT NOT NULL DEFAULT '',
    summary_es      TEXT NOT NULL DEFAULT '',
    summary_en      TEXT NOT NULL DEFAULT '',
    email           TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT NOT NULL DEFAULT '',
    resume_url      TEXT NOT NULL DEFAULT '',
    resume_filename TEXT NOT NULL DEFAULT '',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ==========================================================================
-- Social links (sidebar izquierdo)
-- ==========================================================================
CREATE TABLE social_links (
    id       BIGSERIAL PRIMARY KEY,
    position INTEGER NOT NULL DEFAULT 0,
    name     TEXT NOT NULL,
    url      TEXT NOT NULL DEFAULT '',
    icon_key TEXT NOT NULL DEFAULT ''
);

-- ==========================================================================
-- Skills
-- ==========================================================================
CREATE TABLE skills (
    id       BIGSERIAL PRIMARY KEY,
    position INTEGER NOT NULL DEFAULT 0,
    name     TEXT NOT NULL,
    icon_url TEXT NOT NULL DEFAULT '',
    is_tool  BOOLEAN NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX idx_skills_name ON skills (name);

-- ==========================================================================
-- Projects
-- ==========================================================================
CREATE TABLE projects (
    id                  BIGSERIAL PRIMARY KEY,
    position            INTEGER NOT NULL DEFAULT 0,
    title_es            TEXT NOT NULL,
    title_en            TEXT NOT NULL,
    description_es      TEXT NOT NULL DEFAULT '',
    description_en      TEXT NOT NULL DEFAULT '',
    tech_description_es TEXT NOT NULL DEFAULT '',
    tech_description_en TEXT NOT NULL DEFAULT '',
    category            TEXT NOT NULL DEFAULT '',
    tags                TEXT[] NOT NULL DEFAULT '{}',
    link                TEXT NOT NULL DEFAULT '',
    repo_link           TEXT NOT NULL DEFAULT '',
    cover_url           TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_projects_position ON projects (position);

-- ==========================================================================
-- Project images (screenshots)
-- ==========================================================================
CREATE TABLE project_images (
    id         BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    position   INTEGER NOT NULL DEFAULT 0,
    url        TEXT NOT NULL
);

CREATE INDEX idx_project_images_project ON project_images (project_id);

-- ==========================================================================
-- Experience
-- ==========================================================================
CREATE TABLE experience (
    id            BIGSERIAL PRIMARY KEY,
    position      INTEGER NOT NULL DEFAULT 0,
    company       TEXT NOT NULL,
    position_es   TEXT NOT NULL DEFAULT '',
    position_en   TEXT NOT NULL DEFAULT '',
    period_es     TEXT NOT NULL DEFAULT '',
    period_en     TEXT NOT NULL DEFAULT '',
    description_es TEXT NOT NULL DEFAULT '',
    description_en TEXT NOT NULL DEFAULT ''
);

-- ==========================================================================
-- Education
-- ==========================================================================
CREATE TABLE education (
    id             BIGSERIAL PRIMARY KEY,
    position       INTEGER NOT NULL DEFAULT 0,
    title_es       TEXT NOT NULL,
    title_en       TEXT NOT NULL,
    school         TEXT NOT NULL DEFAULT '',
    date           TEXT NOT NULL DEFAULT '',
    is_course      BOOLEAN NOT NULL DEFAULT false,
    description_es TEXT NOT NULL DEFAULT '',
    description_en TEXT NOT NULL DEFAULT ''
);

-- ==========================================================================
-- Auth
-- ==========================================================================
CREATE TABLE admin_users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,  -- argon2id encoded
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id             BIGSERIAL PRIMARY KEY,
    token_hash     TEXT NOT NULL UNIQUE,     -- SHA-256(token)
    admin_user_id  BIGINT NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    user_agent_hash TEXT NOT NULL DEFAULT '',
    expires_at     TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_expires ON sessions (expires_at);

COMMIT;
