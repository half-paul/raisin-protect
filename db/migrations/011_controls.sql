-- Migration: 011_controls.sql
-- Description: Controls table (org-scoped control library)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

CREATE TABLE IF NOT EXISTS controls (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    identifier              VARCHAR(50) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT NOT NULL,
    implementation_guidance TEXT,
    category                control_category NOT NULL,
    status                  control_status NOT NULL DEFAULT 'draft',
    owner_id                UUID REFERENCES users(id) ON DELETE SET NULL,
    secondary_owner_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    evidence_requirements   TEXT,
    test_criteria           TEXT,
    is_custom               BOOLEAN NOT NULL DEFAULT FALSE,
    source_template_id      VARCHAR(100),
    metadata                JSONB NOT NULL DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_control_identifier UNIQUE (org_id, identifier),
    CONSTRAINT chk_control_owner_different CHECK (owner_id IS NULL OR secondary_owner_id IS NULL OR owner_id <> secondary_owner_id)
);

COMMENT ON TABLE controls IS 'Org-scoped control library — safeguards to meet compliance requirements';
COMMENT ON COLUMN controls.source_template_id IS 'Reference to seed template (e.g. TPL-AC-001) for library updates';
COMMENT ON COLUMN controls.is_custom IS 'TRUE if org-created, FALSE if seeded from control library';
COMMENT ON COLUMN controls.metadata IS 'Extensible custom fields (spec §3.2.2)';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_controls_org_id ON controls (org_id);
CREATE INDEX IF NOT EXISTS idx_controls_status ON controls (org_id, status);
CREATE INDEX IF NOT EXISTS idx_controls_category ON controls (org_id, category);
CREATE INDEX IF NOT EXISTS idx_controls_owner ON controls (owner_id);
CREATE INDEX IF NOT EXISTS idx_controls_secondary_owner ON controls (secondary_owner_id);
CREATE INDEX IF NOT EXISTS idx_controls_source_template ON controls (source_template_id)
    WHERE source_template_id IS NOT NULL;

-- Full-text search on title + description (for control library browser)
CREATE INDEX IF NOT EXISTS idx_controls_search ON controls
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Trigger
DROP TRIGGER IF EXISTS trg_controls_updated_at ON controls;
CREATE TRIGGER trg_controls_updated_at
    BEFORE UPDATE ON controls
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
