-- Migration: 002_organizations.sql
-- Description: Organizations table — root tenant entity
-- Created: 2026-02-20
-- Sprint: 1 — Project Scaffolding & Auth

-- ============================================================================
-- TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS organizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) NOT NULL,
    domain          VARCHAR(255),
    status          org_status NOT NULL DEFAULT 'active',
    settings        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_organizations_slug UNIQUE (slug),
    CONSTRAINT chk_organizations_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_organizations_slug_not_empty CHECK (length(trim(slug)) > 0),
    CONSTRAINT chk_organizations_slug_format CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$' OR length(slug) = 1)
);

COMMENT ON TABLE organizations IS 'Root tenant entity. Every tenant-scoped table references organizations.id';
COMMENT ON COLUMN organizations.slug IS 'URL-friendly org identifier (e.g., /orgs/acme-corp)';
COMMENT ON COLUMN organizations.domain IS 'Optional: for SSO domain matching / email auto-provisioning';
COMMENT ON COLUMN organizations.settings IS 'Org-level config (timezone, locale, etc.) — extensible without migrations';

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations (slug);
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations (status);
CREATE INDEX IF NOT EXISTS idx_organizations_domain ON organizations (domain) WHERE domain IS NOT NULL;

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_organizations_updated_at ON organizations;
CREATE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
