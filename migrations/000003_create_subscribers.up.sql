CREATE TABLE IF NOT EXISTS subscribers (
    id              VARCHAR(36) PRIMARY KEY,
    email           VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    preferences     TEXT NOT NULL DEFAULT '{}',
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_subscribers_email ON subscribers(email);
CREATE INDEX idx_subscribers_status ON subscribers(status);

CREATE TABLE IF NOT EXISTS segment_subscribers (
    segment_id      VARCHAR(36) NOT NULL REFERENCES segments(id),
    subscriber_id   VARCHAR(36) NOT NULL REFERENCES subscribers(id),
    added_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (segment_id, subscriber_id)
);

CREATE INDEX idx_segment_subscribers_subscriber ON segment_subscribers(subscriber_id);
