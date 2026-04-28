CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    keycloak_id TEXT        UNIQUE NOT NULL,
    email       TEXT        UNIQUE NOT NULL,
    name        TEXT        NOT NULL,
    role        TEXT        NOT NULL DEFAULT 'student',
    points      INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    lab_id       TEXT        NOT NULL,
    user_id      UUID        NOT NULL REFERENCES users(id),
    status       TEXT        NOT NULL DEFAULT 'provisioning',
    vm_provider  TEXT,
    vm_id        TEXT,
    vm_ip        TEXT,
    vm_port      INT         NOT NULL DEFAULT 22,
    current_step INT         NOT NULL DEFAULT 0,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '3 hours',
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS step_progress (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID        NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    step_id    INT         NOT NULL,
    status     TEXT        NOT NULL DEFAULT 'pending',
    attempts   INT         NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, step_id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status     ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_step_session        ON step_progress(session_id);
