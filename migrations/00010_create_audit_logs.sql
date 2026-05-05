-- +goose Up
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    event_name  VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50)  NOT NULL,
    entity_id   UUID NULL,
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_occurred_at ON audit_logs(occurred_at DESC);
CREATE INDEX idx_audit_logs_actor_id    ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_event_name  ON audit_logs(event_name);
CREATE INDEX idx_audit_logs_entity      ON audit_logs(entity_type, entity_id);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
