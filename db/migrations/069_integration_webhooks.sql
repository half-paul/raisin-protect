-- Sprint 9 -- Table 5: integration_webhooks
CREATE TABLE IF NOT EXISTS integration_webhooks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connection_id       UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    webhook_secret      VARCHAR(500) NOT NULL,
    signature_header    VARCHAR(100) NOT NULL DEFAULT 'X-Hub-Signature-256',
    signature_algo      VARCHAR(50) NOT NULL DEFAULT 'sha256',
    status              webhook_status NOT NULL DEFAULT 'active',
    event_types         TEXT[] NOT NULL DEFAULT '{}',
    total_received      INTEGER NOT NULL DEFAULT 0,
    total_processed     INTEGER NOT NULL DEFAULT 0,
    total_errors        INTEGER NOT NULL DEFAULT 0,
    last_received_at    TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_webhook_connection_name UNIQUE (connection_id, name)
);

COMMENT ON TABLE integration_webhooks IS 'Webhook endpoints for receiving integration events';

CREATE INDEX IF NOT EXISTS idx_integration_webhooks_org_id ON integration_webhooks (org_id);
CREATE INDEX IF NOT EXISTS idx_integration_webhooks_connection_id ON integration_webhooks (connection_id);
CREATE INDEX IF NOT EXISTS idx_integration_webhooks_status ON integration_webhooks (status);

DROP TRIGGER IF EXISTS trg_integration_webhooks_updated_at ON integration_webhooks;
CREATE TRIGGER trg_integration_webhooks_updated_at
    BEFORE UPDATE ON integration_webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
