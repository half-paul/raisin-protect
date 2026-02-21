-- Migration: 048_audit_evidence_links.sql
-- Description: Audit evidence links — chain-of-custody for evidence submission (Sprint 7)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub

CREATE TABLE IF NOT EXISTS audit_evidence_links (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id            UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,
    request_id          UUID NOT NULL REFERENCES audit_requests(id) ON DELETE CASCADE,
    artifact_id         UUID NOT NULL REFERENCES evidence_artifacts(id) ON DELETE CASCADE,

    -- Submission info
    submitted_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    submitted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submission_notes    TEXT,

    -- Auditor review
    status              audit_evidence_link_status NOT NULL DEFAULT 'pending_review',
    reviewed_by         UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    review_notes        TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Prevent duplicate submissions of same artifact to same request
    CONSTRAINT uq_audit_evidence_request_artifact UNIQUE (request_id, artifact_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_org ON audit_evidence_links (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_audit ON audit_evidence_links (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_request ON audit_evidence_links (request_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_artifact ON audit_evidence_links (artifact_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_status ON audit_evidence_links (request_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_submitted_by ON audit_evidence_links (submitted_by) WHERE submitted_by IS NOT NULL;

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_evidence_links_updated_at ON audit_evidence_links;
CREATE TRIGGER trg_audit_evidence_links_updated_at
    BEFORE UPDATE ON audit_evidence_links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_evidence_links IS 'Evidence artifacts submitted for audit requests — chain-of-custody (spec §3.4.3, §6.4)';
COMMENT ON COLUMN audit_evidence_links.status IS 'Auditor review status: pending_review → accepted/rejected/needs_clarification';
COMMENT ON COLUMN audit_evidence_links.submission_notes IS 'Internal team notes explaining what this evidence demonstrates';
