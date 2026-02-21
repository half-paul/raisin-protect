-- Migration: 032_sprint5_fk_cross_refs.sql
-- Description: Deferred foreign keys for Sprint 5 (policies ↔ policy_versions, evidence_links ↔ policies)
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management

-- ============================================================================
-- policies.current_version_id → policy_versions(id)
-- ============================================================================

DO $$ BEGIN
    ALTER TABLE policies
        ADD CONSTRAINT fk_policies_current_version
        FOREIGN KEY (current_version_id) REFERENCES policy_versions(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_policies_current_version
    ON policies (current_version_id)
    WHERE current_version_id IS NOT NULL;

-- ============================================================================
-- evidence_links.policy_id column + FK → policies(id)
-- Completing Sprint 3's forward declaration for target_type = 'policy'
-- ============================================================================

-- Add the policy_id column if it doesn't exist
DO $$ BEGIN
    ALTER TABLE evidence_links
        ADD COLUMN policy_id UUID REFERENCES policies(id) ON DELETE CASCADE;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_evidence_links_policy
    ON evidence_links (policy_id)
    WHERE policy_id IS NOT NULL;

-- Drop and recreate the CHECK constraint to include policy_id validation
ALTER TABLE evidence_links DROP CONSTRAINT IF EXISTS chk_evidence_link_target;
ALTER TABLE evidence_links ADD CONSTRAINT chk_evidence_link_target CHECK (
    (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL AND policy_id IS NULL) OR
    (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL AND policy_id IS NULL) OR
    (target_type = 'policy' AND policy_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL)
);

-- Uniqueness constraint for policy evidence links
DO $$ BEGIN
    ALTER TABLE evidence_links ADD CONSTRAINT uq_evidence_link_policy
        UNIQUE (org_id, artifact_id, policy_id) DEFERRABLE INITIALLY DEFERRED;
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON CONSTRAINT fk_policies_current_version ON policies IS 'Denormalized FK to the current published policy version';
COMMENT ON COLUMN evidence_links.policy_id IS 'FK to policies — set when target_type = ''policy'' (completing Sprint 3 forward declaration)';
