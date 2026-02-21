-- Migration: 037_risks.sql
-- Description: Core risk register table (Sprint 6 — Risk Register)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register

-- ============================================================================
-- RISKS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS risks (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    identifier                  VARCHAR(50) NOT NULL,
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT,

    -- Classification
    category                    risk_category NOT NULL,
    status                      risk_status NOT NULL DEFAULT 'identified',

    -- Ownership
    owner_id                    UUID REFERENCES users(id) ON DELETE SET NULL,
    secondary_owner_id          UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Denormalized scores (synced via assessment creation / recalculation)
    -- Inherent: risk before any controls or treatments
    inherent_likelihood         likelihood_level,
    inherent_impact             impact_level,
    inherent_score              NUMERIC(5,2),

    -- Residual: risk after controls and treatments applied
    residual_likelihood         likelihood_level,
    residual_impact             impact_level,
    residual_score              NUMERIC(5,2),

    -- Risk appetite
    risk_appetite_threshold     NUMERIC(5,2),
    accepted_at                 TIMESTAMPTZ,
    accepted_by                 UUID REFERENCES users(id) ON DELETE SET NULL,
    acceptance_expiry           DATE,
    acceptance_justification    TEXT,

    -- Assessment scheduling
    assessment_frequency_days   INT,
    next_assessment_at          DATE,
    last_assessed_at            DATE,

    -- Source and context
    source                      VARCHAR(200),
    affected_assets             TEXT[],

    -- Template support
    is_template                 BOOLEAN NOT NULL DEFAULT FALSE,
    template_source             VARCHAR(200),

    -- Tags and metadata
    tags                        TEXT[] DEFAULT '{}',
    metadata                    JSONB NOT NULL DEFAULT '{}',

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_risk_identifier UNIQUE (org_id, identifier),
    CONSTRAINT chk_assessment_frequency CHECK (assessment_frequency_days IS NULL OR assessment_frequency_days > 0),
    CONSTRAINT chk_inherent_score CHECK (inherent_score IS NULL OR (inherent_score >= 1 AND inherent_score <= 25)),
    CONSTRAINT chk_residual_score CHECK (residual_score IS NULL OR (residual_score >= 1 AND residual_score <= 25)),
    CONSTRAINT chk_appetite_threshold CHECK (risk_appetite_threshold IS NULL OR (risk_appetite_threshold >= 1 AND risk_appetite_threshold <= 25)),
    CONSTRAINT chk_acceptance CHECK (
        (status != 'accepted') OR
        (accepted_by IS NOT NULL AND acceptance_justification IS NOT NULL)
    )
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_risks_org ON risks (org_id);
CREATE INDEX IF NOT EXISTS idx_risks_org_status ON risks (org_id, status);
CREATE INDEX IF NOT EXISTS idx_risks_org_category ON risks (org_id, category);
CREATE INDEX IF NOT EXISTS idx_risks_owner ON risks (owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_risks_secondary_owner ON risks (secondary_owner_id) WHERE secondary_owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_risks_inherent_score ON risks (org_id, inherent_score DESC NULLS LAST)
    WHERE is_template = FALSE;
CREATE INDEX IF NOT EXISTS idx_risks_residual_score ON risks (org_id, residual_score DESC NULLS LAST)
    WHERE is_template = FALSE;
CREATE INDEX IF NOT EXISTS idx_risks_templates ON risks (org_id, is_template) WHERE is_template = TRUE;
CREATE INDEX IF NOT EXISTS idx_risks_assessment_due ON risks (org_id, next_assessment_at)
    WHERE next_assessment_at IS NOT NULL AND status NOT IN ('closed', 'archived') AND is_template = FALSE;
CREATE INDEX IF NOT EXISTS idx_risks_acceptance_expiry ON risks (org_id, acceptance_expiry)
    WHERE acceptance_expiry IS NOT NULL AND status = 'accepted';
CREATE INDEX IF NOT EXISTS idx_risks_identifier ON risks (org_id, identifier);
CREATE INDEX IF NOT EXISTS idx_risks_tags ON risks USING gin (tags);
CREATE INDEX IF NOT EXISTS idx_risks_affected_assets ON risks USING gin (affected_assets);

-- Heat map indexes: aggregation by likelihood × impact
CREATE INDEX IF NOT EXISTS idx_risks_heat_map_inherent ON risks (org_id, inherent_likelihood, inherent_impact)
    WHERE is_template = FALSE AND status NOT IN ('closed', 'archived');
CREATE INDEX IF NOT EXISTS idx_risks_heat_map_residual ON risks (org_id, residual_likelihood, residual_impact)
    WHERE is_template = FALSE AND status NOT IN ('closed', 'archived');

-- Full-text search on title + description
CREATE INDEX IF NOT EXISTS idx_risks_search ON risks
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_risks_updated_at ON risks;
CREATE TRIGGER trg_risks_updated_at
    BEFORE UPDATE ON risks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
