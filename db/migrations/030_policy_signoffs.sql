-- Migration: 030_policy_signoffs.sql
-- Description: Policy signoffs table (approval workflow tracking)
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management

-- ============================================================================
-- POLICY SIGNOFFS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS policy_signoffs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id               UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    policy_version_id       UUID NOT NULL REFERENCES policy_versions(id) ON DELETE CASCADE,

    -- Who needs to sign
    signer_id               UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    signer_role             grc_role,

    -- Request details
    requested_by            UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    requested_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date                DATE,

    -- Sign-off decision
    status                  signoff_status NOT NULL DEFAULT 'pending',
    decided_at              TIMESTAMPTZ,
    comments                TEXT,

    -- Notification tracking
    reminder_sent_at        TIMESTAMPTZ,
    reminder_count          INT NOT NULL DEFAULT 0,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each signer signs each version only once
    CONSTRAINT uq_policy_signoff UNIQUE (policy_version_id, signer_id),
    -- Rejection requires comments
    CONSTRAINT chk_rejection_comments CHECK (
        status != 'rejected' OR comments IS NOT NULL
    ),
    -- Decided_at must be set when status is not pending
    CONSTRAINT chk_decided_at CHECK (
        (status = 'pending' AND decided_at IS NULL) OR
        (status != 'pending' AND decided_at IS NOT NULL) OR
        (status = 'withdrawn')
    )
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_org
    ON policy_signoffs (org_id);

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_policy
    ON policy_signoffs (policy_id);

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_version
    ON policy_signoffs (policy_version_id);

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_signer
    ON policy_signoffs (signer_id);

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_signer_pending
    ON policy_signoffs (signer_id, status)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_requested_by
    ON policy_signoffs (requested_by);

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_status
    ON policy_signoffs (org_id, status);

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_pending
    ON policy_signoffs (org_id, policy_id, status)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_policy_signoffs_due
    ON policy_signoffs (org_id, due_date)
    WHERE status = 'pending' AND due_date IS NOT NULL;

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_policy_signoffs_updated_at ON policy_signoffs;
CREATE TRIGGER trg_policy_signoffs_updated_at
    BEFORE UPDATE ON policy_signoffs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE policy_signoffs IS 'Approval workflow tracking for policy versions. Each row = one signer''s status for one version.';
COMMENT ON COLUMN policy_signoffs.signer_role IS 'Snapshot of signer''s role at request time (audit trail integrity)';
COMMENT ON COLUMN policy_signoffs.due_date IS 'Optional deadline for signing — enables SLA tracking';
COMMENT ON COLUMN policy_signoffs.reminder_count IS 'Number of reminder notifications sent for this sign-off request';
