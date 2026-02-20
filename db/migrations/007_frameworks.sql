-- Migration: 007_frameworks.sql
-- Description: Frameworks catalog table (system-level, no org_id)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

CREATE TABLE IF NOT EXISTS frameworks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier      VARCHAR(50) NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    category        framework_category NOT NULL,
    website_url     VARCHAR(500),
    logo_url        VARCHAR(500),
    is_custom       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE frameworks IS 'Master catalog of compliance frameworks (system-level, shared across all tenants)';
COMMENT ON COLUMN frameworks.identifier IS 'Stable machine key (e.g. soc2, pci_dss) used in APIs';
COMMENT ON COLUMN frameworks.is_custom IS 'TRUE for user-created frameworks (future feature)';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_frameworks_category ON frameworks (category);
CREATE INDEX IF NOT EXISTS idx_frameworks_identifier ON frameworks (identifier);

-- Trigger
DROP TRIGGER IF EXISTS trg_frameworks_updated_at ON frameworks;
CREATE TRIGGER trg_frameworks_updated_at
    BEFORE UPDATE ON frameworks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
