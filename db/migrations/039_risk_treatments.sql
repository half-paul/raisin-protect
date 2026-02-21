-- Migration: 039_risk_treatments.sql
-- Description: Risk treatments table — mitigation/acceptance/transfer/avoidance plans (Sprint 6 — Risk Register)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register

-- ============================================================================
-- RISK TREATMENTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS risk_treatments (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,

    -- Treatment definition
    treatment_type              treatment_type NOT NULL,
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT,
    status                      treatment_status NOT NULL DEFAULT 'planned',

    -- Ownership
    owner_id                    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by                  UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Priority and scheduling
    priority                    VARCHAR(20) NOT NULL DEFAULT 'medium'
                                CHECK (priority IN ('critical', 'high', 'medium', 'low')),
    due_date                    DATE,
    started_at                  TIMESTAMPTZ,
    completed_at                TIMESTAMPTZ,

    -- Effort tracking
    estimated_effort_hours      NUMERIC(8,2),
    actual_effort_hours         NUMERIC(8,2),

    -- Effectiveness review (post-implementation, spec §4.1.3)
    effectiveness_rating        VARCHAR(20)
                                CHECK (effectiveness_rating IS NULL OR
                                       effectiveness_rating IN ('highly_effective', 'effective',
                                                                 'partially_effective', 'ineffective')),
    effectiveness_notes         TEXT,
    effectiveness_reviewed_at   TIMESTAMPTZ,
    effectiveness_reviewed_by   UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Expected risk reduction
    expected_residual_likelihood likelihood_level,
    expected_residual_impact    impact_level,
    expected_residual_score     NUMERIC(5,2),

    -- Target control (optional: which control implements this treatment)
    target_control_id           UUID REFERENCES controls(id) ON DELETE SET NULL,

    -- Notes and metadata
    notes                       TEXT,
    metadata                    JSONB NOT NULL DEFAULT '{}',

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_rt_expected_residual_score CHECK (
        expected_residual_score IS NULL OR
        (expected_residual_score >= 1 AND expected_residual_score <= 25)
    ),
    CONSTRAINT chk_rt_completed CHECK (
        (status NOT IN ('verified', 'implemented') OR completed_at IS NOT NULL)
    ),
    CONSTRAINT chk_rt_effort CHECK (
        estimated_effort_hours IS NULL OR estimated_effort_hours >= 0
    ),
    CONSTRAINT chk_rt_actual_effort CHECK (
        actual_effort_hours IS NULL OR actual_effort_hours >= 0
    )
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_risk_treatments_org ON risk_treatments (org_id);
CREATE INDEX IF NOT EXISTS idx_risk_treatments_risk ON risk_treatments (risk_id);
CREATE INDEX IF NOT EXISTS idx_risk_treatments_risk_status ON risk_treatments (risk_id, status);
CREATE INDEX IF NOT EXISTS idx_risk_treatments_owner ON risk_treatments (owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_risk_treatments_created_by ON risk_treatments (created_by) WHERE created_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_risk_treatments_status ON risk_treatments (org_id, status);
CREATE INDEX IF NOT EXISTS idx_risk_treatments_due ON risk_treatments (org_id, due_date)
    WHERE due_date IS NOT NULL AND status IN ('planned', 'in_progress');
CREATE INDEX IF NOT EXISTS idx_risk_treatments_overdue ON risk_treatments (org_id, due_date, status)
    WHERE due_date IS NOT NULL AND status IN ('planned', 'in_progress');
CREATE INDEX IF NOT EXISTS idx_risk_treatments_type ON risk_treatments (org_id, treatment_type);
CREATE INDEX IF NOT EXISTS idx_risk_treatments_priority ON risk_treatments (org_id, priority)
    WHERE status IN ('planned', 'in_progress');
CREATE INDEX IF NOT EXISTS idx_risk_treatments_control ON risk_treatments (target_control_id)
    WHERE target_control_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_risk_treatments_effectiveness ON risk_treatments (org_id, effectiveness_rating)
    WHERE effectiveness_rating IS NOT NULL;

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_risk_treatments_updated_at ON risk_treatments;
CREATE TRIGGER trg_risk_treatments_updated_at
    BEFORE UPDATE ON risk_treatments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
