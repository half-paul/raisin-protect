-- Sprint 9 -- Table 2: integration_connections (per-org)
CREATE TABLE IF NOT EXISTS integration_connections (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    definition_id           UUID NOT NULL REFERENCES integration_definitions(id),
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    instance_label          VARCHAR(255),
    status                  connection_status NOT NULL DEFAULT 'pending',
    health                  connection_health NOT NULL DEFAULT 'unknown',
    config                  JSONB NOT NULL DEFAULT '{}',
    sync_enabled            BOOLEAN NOT NULL DEFAULT false,
    sync_interval_mins      INTEGER NOT NULL DEFAULT 360,
    sync_cron               VARCHAR(100),
    next_sync_at            TIMESTAMPTZ,
    last_sync_at            TIMESTAMPTZ,
    last_sync_status        run_status,
    last_sync_error         TEXT,
    last_health_check_at    TIMESTAMPTZ,
    last_health_status      connection_health,
    consecutive_failures    INTEGER NOT NULL DEFAULT 0,
    total_runs              INTEGER NOT NULL DEFAULT 0,
    successful_runs         INTEGER NOT NULL DEFAULT 0,
    failed_runs             INTEGER NOT NULL DEFAULT 0,
    created_by              UUID REFERENCES users(id),
    tags                    TEXT[] NOT NULL DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_connection_org_name UNIQUE (org_id, name),
    CONSTRAINT chk_sync_interval CHECK (sync_interval_mins >= 5 AND sync_interval_mins <= 10080)
);

COMMENT ON TABLE integration_connections IS 'Per-org integration connection instances';

CREATE INDEX IF NOT EXISTS idx_integration_connections_org_id ON integration_connections (org_id);
CREATE INDEX IF NOT EXISTS idx_integration_connections_definition_id ON integration_connections (definition_id);
CREATE INDEX IF NOT EXISTS idx_integration_connections_status ON integration_connections (org_id, status);
CREATE INDEX IF NOT EXISTS idx_integration_connections_health ON integration_connections (org_id, health);
CREATE INDEX IF NOT EXISTS idx_integration_connections_next_sync ON integration_connections (next_sync_at) WHERE sync_enabled = true;

DROP TRIGGER IF EXISTS trg_integration_connections_updated_at ON integration_connections;
CREATE TRIGGER trg_integration_connections_updated_at
    BEFORE UPDATE ON integration_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
