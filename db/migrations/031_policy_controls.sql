-- Migration: 031_policy_controls.sql
-- Description: Policy controls junction table (many-to-many: policy ↔ control)
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management

-- ============================================================================
-- POLICY CONTROLS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS policy_controls (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id               UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Link metadata
    notes                   TEXT,
    coverage                VARCHAR(20) NOT NULL DEFAULT 'full'
                            CHECK (coverage IN ('full', 'partial')),

    -- Who created the link
    linked_by               UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each policy links to each control only once per org
    CONSTRAINT uq_policy_control UNIQUE (org_id, policy_id, control_id)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_policy_controls_org
    ON policy_controls (org_id);

CREATE INDEX IF NOT EXISTS idx_policy_controls_policy
    ON policy_controls (policy_id);

CREATE INDEX IF NOT EXISTS idx_policy_controls_control
    ON policy_controls (control_id);

CREATE INDEX IF NOT EXISTS idx_policy_controls_linked_by
    ON policy_controls (linked_by)
    WHERE linked_by IS NOT NULL;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE policy_controls IS 'Many-to-many junction: policies ↔ controls. Foundation for policy gap detection (spec §6.1).';
COMMENT ON COLUMN policy_controls.coverage IS 'Whether the policy fully or partially addresses the control''s governance needs';
COMMENT ON COLUMN policy_controls.notes IS 'Explanation of why this policy governs this control';
