CREATE TABLE notification.events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic       VARCHAR(100) NOT NULL,
    event_key   VARCHAR(200) NOT NULL,
    payload     JSONB        NOT NULL,
    webhook_url TEXT         NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'sent',
    http_code   INT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_topic ON notification.events(topic);
