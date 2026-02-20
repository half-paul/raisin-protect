-- Migration: 010_org_frameworks.sql
-- Description: Org-level framework activations (which frameworks each org is pursuing)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

CREATE TABLE IF NOT EXISTS org_frameworks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    framework_id        UUID NOT NULL REFERENCES frameworks(id) ON DELETE RESTRICT,
    active_version_id   UUID NOT NULL REFERENCES framework_versions(id) ON DELETE RESTRICT,
    status              org_framework_status NOT NULL DEFAULT 'active',
    target_date         DATE,
    notes               TEXT,
    activated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deactivated_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_org_framework UNIQUE (org_id, framework_id)
);

COMMENT ON TABLE org_frameworks IS 'Which frameworks each organization has activated for compliance tracking';
COMMENT ON COLUMN org_frameworks.target_date IS 'Target compliance date for planning';
COMMENT ON COLUMN org_frameworks.active_version_id IS 'Which version the org is working against';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_org_frameworks_org_id ON org_frameworks (org_id);
CREATE INDEX IF NOT EXISTS idx_org_frameworks_status ON org_frameworks (org_id, status);
CREATE INDEX IF NOT EXISTS idx_org_frameworks_framework ON org_frameworks (framework_id);

-- Trigger
DROP TRIGGER IF EXISTS trg_org_frameworks_updated_at ON org_frameworks;
CREATE TRIGGER trg_org_frameworks_updated_at
    BEFORE UPDATE ON org_frameworks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
