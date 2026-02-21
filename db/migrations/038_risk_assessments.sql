-- Migration: 038_risk_assessments.sql
-- Description: Risk assessments table — point-in-time scoring snapshots (Sprint 6 — Risk Register)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register

-- ============================================================================
-- RISK ASSESSMENTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS risk_assessments (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,

    -- Assessment classification
    assessment_type             risk_assessment_type NOT NULL,

    -- Scoring
    likelihood                  likelihood_level NOT NULL,
    impact                      impact_level NOT NULL,
    likelihood_score            INT NOT NULL,
    impact_score                INT NOT NULL,
    overall_score               NUMERIC(5,2) NOT NULL,
    scoring_formula             VARCHAR(100) NOT NULL DEFAULT 'likelihood_x_impact',

    -- Severity classification (derived from overall_score)
    severity                    VARCHAR(20) NOT NULL,

    -- Context
    justification               TEXT,
    assumptions                 TEXT,
    data_sources                TEXT[],

    -- Assessor
    assessed_by                 UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assessment_date             DATE NOT NULL DEFAULT CURRENT_DATE,

    -- Validity
    valid_until                 DATE,
    superseded_by               UUID REFERENCES risk_assessments(id) ON DELETE SET NULL,
    is_current                  BOOLEAN NOT NULL DEFAULT TRUE,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_ra_likelihood_score CHECK (likelihood_score >= 1 AND likelihood_score <= 5),
    CONSTRAINT chk_ra_impact_score CHECK (impact_score >= 1 AND impact_score <= 5),
    CONSTRAINT chk_ra_overall_score CHECK (overall_score >= 1 AND overall_score <= 25)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_risk_assessments_org ON risk_assessments (org_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_risk ON risk_assessments (risk_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_risk_type ON risk_assessments (risk_id, assessment_type);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_current ON risk_assessments (risk_id, assessment_type, is_current)
    WHERE is_current = TRUE;
CREATE INDEX IF NOT EXISTS idx_risk_assessments_assessed_by ON risk_assessments (assessed_by);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_date ON risk_assessments (risk_id, assessment_date DESC);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_severity ON risk_assessments (org_id, severity)
    WHERE is_current = TRUE;
CREATE INDEX IF NOT EXISTS idx_risk_assessments_expiry ON risk_assessments (org_id, valid_until)
    WHERE valid_until IS NOT NULL AND is_current = TRUE;

-- Each risk can have only one current assessment per type
CREATE UNIQUE INDEX IF NOT EXISTS uq_risk_assessment_current ON risk_assessments (risk_id, assessment_type)
    WHERE is_current = TRUE;
