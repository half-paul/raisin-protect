-- Sprint 8 — Table 1: identity_providers

CREATE TABLE IF NOT EXISTS identity_providers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Provider identity
    name                VARCHAR(255) NOT NULL,
    provider_type       identity_provider_type NOT NULL,
    status              identity_provider_status NOT NULL DEFAULT 'pending_setup',

    -- Connection configuration (JSONB for flexibility across providers)
    config              JSONB NOT NULL DEFAULT '{}',

    -- Sync state
    last_sync_at        TIMESTAMPTZ,
    last_sync_status    VARCHAR(50),                   -- 'success', 'partial', 'failed'
    last_sync_error     TEXT,                          -- Error message if sync failed
    last_sync_stats     JSONB DEFAULT '{}',            -- { "users_synced": 150, "resources_synced": 23, "duration_ms": 4500 }
    sync_interval_mins  INTEGER NOT NULL DEFAULT 360,  -- Default: sync every 6 hours

    -- Metadata
    description         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One active provider per type per org (can reconnect different instance)
    CONSTRAINT uq_identity_providers_org_name UNIQUE (org_id, name),
    CONSTRAINT chk_identity_providers_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_identity_providers_sync_interval CHECK (sync_interval_mins >= 15 AND sync_interval_mins <= 10080)
);

COMMENT ON TABLE identity_providers IS 'Identity provider connection stubs. Full connector logic in Sprint 9 Integration Engine.';
COMMENT ON COLUMN identity_providers.config IS 'Provider-specific config (JSONB). Secrets should be encrypted at application layer before storage.';
COMMENT ON COLUMN identity_providers.last_sync_stats IS 'Stats from last sync run: users/resources synced, duration, errors.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_identity_providers_org_id ON identity_providers (org_id);
CREATE INDEX IF NOT EXISTS idx_identity_providers_org_status ON identity_providers (org_id, status);
CREATE INDEX IF NOT EXISTS idx_identity_providers_org_type ON identity_providers (org_id, provider_type);

-- Trigger
DROP TRIGGER IF EXISTS trg_identity_providers_updated_at ON identity_providers;
CREATE TRIGGER trg_identity_providers_updated_at
    BEFORE UPDATE ON identity_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
