-- Sprint 9 -- Table 3: integration_runs (sync execution history)
CREATE TABLE IF NOT EXISTS integration_runs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connection_id       UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
    trigger             run_trigger NOT NULL DEFAULT 'manual',
    triggered_by        UUID REFERENCES users(id),
    status              run_status NOT NULL DEFAULT 'pending',
    queued_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    duration_ms         INTEGER,
    stats               JSONB NOT NULL DEFAULT '{}',
    error_message       TEXT,
    error_details       JSONB,
    retry_count         INTEGER NOT NULL DEFAULT 0,
    max_retries         INTEGER NOT NULL DEFAULT 3,
    config_snapshot     JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE integration_runs IS 'History of integration sync runs';

CREATE INDEX IF NOT EXISTS idx_integration_runs_org_id ON integration_runs (org_id);
CREATE INDEX IF NOT EXISTS idx_integration_runs_connection_id ON integration_runs (connection_id);
CREATE INDEX IF NOT EXISTS idx_integration_runs_status_pending ON integration_runs (status) WHERE status IN ('pending', 'running');
CREATE INDEX IF NOT EXISTS idx_integration_runs_connection_created ON integration_runs (connection_id, created_at DESC);

DROP TRIGGER IF EXISTS trg_integration_runs_updated_at ON integration_runs;
CREATE TRIGGER trg_integration_runs_updated_at
    BEFORE UPDATE ON integration_runs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
