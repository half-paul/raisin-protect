-- Migration: 012_control_mappings.sql
-- Description: Cross-framework control mappings (links controls to requirements)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

CREATE TABLE IF NOT EXISTS control_mappings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    control_id      UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    requirement_id  UUID NOT NULL REFERENCES requirements(id) ON DELETE CASCADE,
    strength        VARCHAR(20) NOT NULL DEFAULT 'primary'
                    CHECK (strength IN ('primary', 'supporting', 'partial')),
    notes           TEXT,
    mapped_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_control_mapping UNIQUE (org_id, control_id, requirement_id)
);

COMMENT ON TABLE control_mappings IS 'Maps org controls to framework requirements (one control can satisfy many frameworks)';
COMMENT ON COLUMN control_mappings.strength IS 'How strongly the control addresses the requirement: primary, supporting, or partial';
COMMENT ON COLUMN control_mappings.org_id IS 'Denormalized from control_id for efficient org-scoped queries';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_control_mappings_org ON control_mappings (org_id);
CREATE INDEX IF NOT EXISTS idx_control_mappings_control ON control_mappings (control_id);
CREATE INDEX IF NOT EXISTS idx_control_mappings_requirement ON control_mappings (requirement_id);
CREATE INDEX IF NOT EXISTS idx_control_mappings_strength ON control_mappings (org_id, strength);

-- Composite index for cross-framework matrix query
CREATE INDEX IF NOT EXISTS idx_control_mappings_cross_fw ON control_mappings (org_id, requirement_id, control_id);
