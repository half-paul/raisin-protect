-- Sprint 9 -- Table 4: integration_logs (append-only)
CREATE TABLE IF NOT EXISTS integration_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    run_id          UUID NOT NULL REFERENCES integration_runs(id) ON DELETE CASCADE,
    connection_id   UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    level           log_level NOT NULL DEFAULT 'info',
    message         TEXT NOT NULL,
    details         JSONB,
    source          VARCHAR(255),
    item_ref        VARCHAR(500),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE integration_logs IS 'Append-only logs for integration runs';

CREATE INDEX IF NOT EXISTS idx_integration_logs_run_created ON integration_logs (run_id, created_at);
CREATE INDEX IF NOT EXISTS idx_integration_logs_run_level ON integration_logs (run_id, level);
CREATE INDEX IF NOT EXISTS idx_integration_logs_connection_id ON integration_logs (connection_id);
