-- Migration: 016_evidence_links.sql
-- Description: Evidence links table (many-to-many: evidence ↔ controls/requirements/policies)
-- Created: 2026-02-20
-- Sprint: 3 — Evidence Management

-- ============================================================================
-- EVIDENCE LINKS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS evidence_links (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    artifact_id         UUID NOT NULL REFERENCES evidence_artifacts(id) ON DELETE CASCADE,

    -- Polymorphic target
    target_type         evidence_link_target_type NOT NULL,
    control_id          UUID REFERENCES controls(id) ON DELETE CASCADE,
    requirement_id      UUID REFERENCES requirements(id) ON DELETE CASCADE,
    -- policy_id        UUID REFERENCES policies(id) ON DELETE CASCADE,  -- Sprint 5

    -- Link metadata
    notes               TEXT,
    strength            VARCHAR(20) NOT NULL DEFAULT 'primary'
                        CHECK (strength IN ('primary', 'supporting', 'supplementary')),
    linked_by           UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure each artifact links to each target only once
    CONSTRAINT uq_evidence_link_control UNIQUE (org_id, artifact_id, control_id)
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT uq_evidence_link_requirement UNIQUE (org_id, artifact_id, requirement_id)
        DEFERRABLE INITIALLY DEFERRED,

    -- Enforce correct FK populated for target_type
    CONSTRAINT chk_evidence_link_target CHECK (
        (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL) OR
        (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL) OR
        (target_type = 'policy' AND control_id IS NULL AND requirement_id IS NULL)
    )
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_evidence_links_org
    ON evidence_links (org_id);

CREATE INDEX IF NOT EXISTS idx_evidence_links_artifact
    ON evidence_links (artifact_id);

CREATE INDEX IF NOT EXISTS idx_evidence_links_control
    ON evidence_links (control_id)
    WHERE control_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_evidence_links_requirement
    ON evidence_links (requirement_id)
    WHERE requirement_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_evidence_links_target_type
    ON evidence_links (org_id, target_type);

CREATE INDEX IF NOT EXISTS idx_evidence_links_linked_by
    ON evidence_links (linked_by)
    WHERE linked_by IS NOT NULL;

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_evidence_links_updated_at ON evidence_links;
CREATE TRIGGER trg_evidence_links_updated_at
    BEFORE UPDATE ON evidence_links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE evidence_links IS 'Many-to-many links between evidence artifacts and controls/requirements/policies (spec §3.4.3)';
COMMENT ON COLUMN evidence_links.target_type IS 'Polymorphic discriminator: control, requirement, or policy';
COMMENT ON COLUMN evidence_links.strength IS 'How strongly evidence supports the target: primary, supporting, supplementary';
COMMENT ON COLUMN evidence_links.notes IS 'Explanation of why this evidence supports this control/requirement';
