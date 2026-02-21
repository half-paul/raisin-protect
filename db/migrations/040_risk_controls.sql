-- Migration: 040_risk_controls.sql
-- Description: Risk-to-control junction table with effectiveness tracking (Sprint 6 — Risk Register)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register

-- ============================================================================
-- RISK CONTROLS TABLE (many-to-many: risks ↔ controls)
-- ============================================================================

CREATE TABLE IF NOT EXISTS risk_controls (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    control_id                  UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Effectiveness assessment
    effectiveness               control_effectiveness NOT NULL DEFAULT 'not_assessed',

    -- Link metadata
    notes                       TEXT,
    mitigation_percentage       INT,

    -- Who created the link
    linked_by                   UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Effectiveness review tracking
    last_effectiveness_review   DATE,
    reviewed_by                 UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each risk links to each control only once per org
    CONSTRAINT uq_risk_control UNIQUE (org_id, risk_id, control_id),
    CONSTRAINT chk_rc_mitigation_percentage CHECK (
        mitigation_percentage IS NULL OR
        (mitigation_percentage >= 0 AND mitigation_percentage <= 100)
    )
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_risk_controls_org ON risk_controls (org_id);
CREATE INDEX IF NOT EXISTS idx_risk_controls_risk ON risk_controls (risk_id);
CREATE INDEX IF NOT EXISTS idx_risk_controls_control ON risk_controls (control_id);
CREATE INDEX IF NOT EXISTS idx_risk_controls_effectiveness ON risk_controls (org_id, effectiveness);
CREATE INDEX IF NOT EXISTS idx_risk_controls_linked_by ON risk_controls (linked_by) WHERE linked_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_risk_controls_not_assessed ON risk_controls (org_id, risk_id, effectiveness)
    WHERE effectiveness = 'not_assessed';

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_risk_controls_updated_at ON risk_controls;
CREATE TRIGGER trg_risk_controls_updated_at
    BEFORE UPDATE ON risk_controls
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
