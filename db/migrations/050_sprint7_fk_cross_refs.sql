-- Migration: 050_sprint7_fk_cross_refs.sql
-- Description: Extend evidence_links to support audit evidence (Sprint 7 — Audit Hub)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub

-- ============================================================================
-- EXTEND evidence_links TO SUPPORT AUDITS
-- ============================================================================

-- Add audit_id column to evidence_links (if not exists)
DO $$ BEGIN
    ALTER TABLE evidence_links ADD COLUMN audit_id UUID REFERENCES audits(id) ON DELETE CASCADE;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

-- Index for audit evidence lookups
CREATE INDEX IF NOT EXISTS idx_evidence_links_audit ON evidence_links (audit_id)
    WHERE audit_id IS NOT NULL;

-- Update the existing CHECK constraint to include audit_id
-- Must drop old and recreate since CHECK constraints cannot be ALTERed in-place
DO $$ BEGIN
    ALTER TABLE evidence_links DROP CONSTRAINT IF EXISTS chk_evidence_link_target;
EXCEPTION WHEN undefined_object THEN NULL;
END $$;

ALTER TABLE evidence_links ADD CONSTRAINT chk_evidence_link_target CHECK (
    (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL AND policy_id IS NULL AND risk_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL AND policy_id IS NULL AND risk_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'policy' AND policy_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL AND risk_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'risk' AND risk_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL AND policy_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'audit' AND audit_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL AND policy_id IS NULL AND risk_id IS NULL)
);

-- Add uniqueness constraint for audit evidence links (idempotent)
DO $$ BEGIN
    ALTER TABLE evidence_links ADD CONSTRAINT uq_evidence_link_audit
        UNIQUE (org_id, artifact_id, audit_id) DEFERRABLE INITIALLY DEFERRED;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
