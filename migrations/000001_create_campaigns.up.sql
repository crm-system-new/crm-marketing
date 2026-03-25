CREATE TABLE IF NOT EXISTS campaigns (
    id              VARCHAR(36) PRIMARY KEY,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    channel         VARCHAR(50) NOT NULL,
    segment_id      VARCHAR(36),
    scheduled_at    TIMESTAMPTZ,
    sent_count      INTEGER NOT NULL DEFAULT 0,
    open_rate       DOUBLE PRECISION NOT NULL DEFAULT 0,
    click_rate      DOUBLE PRECISION NOT NULL DEFAULT 0,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_channel ON campaigns(channel);
CREATE INDEX idx_campaigns_segment ON campaigns(segment_id);
