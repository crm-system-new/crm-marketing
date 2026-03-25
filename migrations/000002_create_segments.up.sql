CREATE TABLE IF NOT EXISTS segments (
    id                  VARCHAR(36) PRIMARY KEY,
    name                VARCHAR(200) NOT NULL,
    criteria            TEXT,
    subscriber_count    INTEGER NOT NULL DEFAULT 0,
    version             INTEGER NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
