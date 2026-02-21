-- Migration: 041_sprint6_fk_cross_refs.sql
-- Description: Deferred FKs — extend evidence_links to support risk evidence (Sprint 6 — Risk Register)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register

-- ============================================================================
-- EXTEND evidence_links TO SUPPORT RISKS
-- ============================================================================

-- Add risk_id column to evidence_links (if not exists)
DO $$ BEGIN
    ALTER TABLE evidence_links ADD COLUMN risk_id UUID REFERENCES risks(id) ON DELETE CASCADE;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- Index for risk evidence lookups
CREATE INDEX IF NOT EXISTS idx_evidence_links_risk ON evidence_links (risk_id)
    WHERE risk_id IS NOT NULL;

-- Update the existing CHECK constraint on evidence_links to include risk_id
-- Drop old constraint safely and recreate with risk support
DO $$ BEGIN
    ALTER TABLE evidence_links DROP CONSTRAINT IF EXISTS chk_evidence_link_target;
EXCEPTION WHEN undefined_object THEN NULL;
END $$;

ALTER TABLE evidence_links ADD CONSTRAINT chk_evidence_link_target CHECK (
    (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL AND policy_id IS NULL AND risk_id IS NULL) OR
    (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL AND policy_id IS NULL AND risk_id IS NULL) OR
    (target_type = 'policy' AND policy_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL AND risk_id IS NULL) OR
    (target_type = 'risk' AND risk_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL AND policy_id IS NULL)
);

-- Add uniqueness constraint for risk evidence links (idempotent via IF NOT EXISTS on index)
DO $$ BEGIN
    ALTER TABLE evidence_links ADD CONSTRAINT uq_evidence_link_risk
        UNIQUE (org_id, artifact_id, risk_id) DEFERRABLE INITIALLY DEFERRED;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
