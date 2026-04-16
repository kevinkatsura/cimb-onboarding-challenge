-- Notification schema

CREATE TABLE notifications (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type    VARCHAR(100) NOT NULL,
    event_key     VARCHAR(200) NOT NULL,
    payload       JSONB NOT NULL,
    callback_url  TEXT NOT NULL DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    response_code INT,
    response_body TEXT,
    attempts      INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at       TIMESTAMPTZ
);

CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_event  ON notifications(event_type);
