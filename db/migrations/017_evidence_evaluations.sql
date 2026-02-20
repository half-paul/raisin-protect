-- Migration: 017_evidence_evaluations.sql
-- Description: Evidence evaluations table (review/approve/reject history)
-- Created: 2026-02-20
-- Sprint: 3 — Evidence Management

-- ============================================================================
-- EVIDENCE EVALUATIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS evidence_evaluations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    artifact_id         UUID NOT NULL REFERENCES evidence_artifacts(id) ON DELETE CASCADE,
    evidence_link_id    UUID REFERENCES evidence_links(id) ON DELETE SET NULL,

    -- Evaluation
    verdict             evidence_evaluation_verdict NOT NULL,
    confidence          VARCHAR(20) NOT NULL DEFAULT 'medium'
                        CHECK (confidence IN ('high', 'medium', 'low')),
    comments            TEXT NOT NULL,

    -- What was found lacking (if not sufficient)
    missing_elements    TEXT[],
    remediation_notes   TEXT,

    -- Who evaluated
    evaluated_by        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Immutable: no updated_at (create new evaluation to re-evaluate)
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_evidence_evaluations_org
    ON evidence_evaluations (org_id);

CREATE INDEX IF NOT EXISTS idx_evidence_evaluations_artifact
    ON evidence_evaluations (artifact_id);

CREATE INDEX IF NOT EXISTS idx_evidence_evaluations_link
    ON evidence_evaluations (evidence_link_id)
    WHERE evidence_link_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_evidence_evaluations_verdict
    ON evidence_evaluations (org_id, verdict);

CREATE INDEX IF NOT EXISTS idx_evidence_evaluations_evaluator
    ON evidence_evaluations (evaluated_by);

CREATE INDEX IF NOT EXISTS idx_evidence_evaluations_created
    ON evidence_evaluations (artifact_id, created_at DESC);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE evidence_evaluations IS 'Immutable evaluation history for evidence artifacts (spec §3.4.2). Append-only — no updates.';
COMMENT ON COLUMN evidence_evaluations.evidence_link_id IS 'Optional: evaluate against a specific control/requirement link. NULL = general quality evaluation.';
COMMENT ON COLUMN evidence_evaluations.confidence IS 'Evaluation confidence: high, medium, low (aligns with spec §3.4.2 AI scoring)';
COMMENT ON COLUMN evidence_evaluations.missing_elements IS 'Structured tracking of missing elements: dates, signatures, scope statements, etc.';
