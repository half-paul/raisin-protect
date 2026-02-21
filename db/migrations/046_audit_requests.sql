-- Migration: 046_audit_requests.sql
-- Description: Audit evidence requests — PBC items from auditors to internal teams (Sprint 7)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub

CREATE TABLE IF NOT EXISTS audit_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id            UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,

    -- Request content
    title               VARCHAR(500) NOT NULL,
    description         TEXT NOT NULL,
    priority            audit_request_priority NOT NULL DEFAULT 'medium',
    status              audit_request_status NOT NULL DEFAULT 'open',

    -- What this request relates to
    control_id          UUID REFERENCES controls(id) ON DELETE SET NULL,
    requirement_id      UUID REFERENCES requirements(id) ON DELETE SET NULL,

    -- People
    requested_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_to         UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Timing
    due_date            DATE,
    submitted_at        TIMESTAMPTZ,
    reviewed_at         TIMESTAMPTZ,

    -- Auditor feedback (when rejected or needing clarification)
    reviewer_notes      TEXT,

    -- Request metadata
    reference_number    VARCHAR(50),
    tags                TEXT[] NOT NULL DEFAULT '{}',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_requests_org ON audit_requests (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_requests_audit ON audit_requests (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_requests_audit_status ON audit_requests (audit_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_requests_assigned ON audit_requests (assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_requested_by ON audit_requests (requested_by) WHERE requested_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_control ON audit_requests (control_id) WHERE control_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_requirement ON audit_requests (requirement_id) WHERE requirement_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_due_date ON audit_requests (audit_id, due_date) WHERE due_date IS NOT NULL AND status NOT IN ('accepted', 'closed');
CREATE INDEX IF NOT EXISTS idx_audit_requests_org_status ON audit_requests (org_id, status);

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_requests_updated_at ON audit_requests;
CREATE TRIGGER trg_audit_requests_updated_at
    BEFORE UPDATE ON audit_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_requests IS 'Evidence requests from auditors to internal teams (spec §6.4)';
COMMENT ON COLUMN audit_requests.reference_number IS 'Auditor''s own reference number (e.g., PBC list item)';
COMMENT ON COLUMN audit_requests.reviewer_notes IS 'Auditor feedback when rejecting or requesting clarification';
COMMENT ON COLUMN audit_requests.due_date IS 'Deadline for evidence submission; used for SLA/overdue tracking';
