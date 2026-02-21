-- Migration: 047_audit_findings.sql
-- Description: Audit findings — deficiencies with remediation tracking (Sprint 7)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub

CREATE TABLE IF NOT EXISTS audit_findings (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id                UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,

    -- Finding content
    title                   VARCHAR(500) NOT NULL,
    description             TEXT NOT NULL,
    severity                audit_finding_severity NOT NULL,
    category                audit_finding_category NOT NULL,
    status                  audit_finding_status NOT NULL DEFAULT 'identified',

    -- What this finding relates to
    control_id              UUID REFERENCES controls(id) ON DELETE SET NULL,
    requirement_id          UUID REFERENCES requirements(id) ON DELETE SET NULL,

    -- People
    found_by                UUID REFERENCES users(id) ON DELETE SET NULL,
    remediation_owner_id    UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Remediation tracking
    remediation_plan        TEXT,
    remediation_due_date    DATE,
    remediation_started_at  TIMESTAMPTZ,
    remediation_completed_at TIMESTAMPTZ,
    verification_notes      TEXT,
    verified_at             TIMESTAMPTZ,
    verified_by             UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Risk acceptance (if org accepts instead of remediating)
    risk_accepted           BOOLEAN NOT NULL DEFAULT FALSE,
    risk_acceptance_reason  TEXT,
    risk_accepted_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    risk_accepted_at        TIMESTAMPTZ,

    -- Finding metadata
    reference_number        VARCHAR(50),
    recommendation          TEXT,
    management_response     TEXT,
    tags                    TEXT[] NOT NULL DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_finding_risk_accepted CHECK (
        (risk_accepted = FALSE) OR
        (risk_accepted = TRUE AND risk_acceptance_reason IS NOT NULL)
    ),
    CONSTRAINT chk_finding_verification CHECK (
        verified_at IS NULL OR verified_by IS NOT NULL
    )
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_findings_org ON audit_findings (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_findings_audit ON audit_findings (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_findings_audit_status ON audit_findings (audit_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_findings_audit_severity ON audit_findings (audit_id, severity);
CREATE INDEX IF NOT EXISTS idx_audit_findings_remediation_owner ON audit_findings (remediation_owner_id) WHERE remediation_owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_findings_control ON audit_findings (control_id) WHERE control_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_findings_requirement ON audit_findings (requirement_id) WHERE requirement_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_findings_org_status ON audit_findings (org_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_findings_due_date ON audit_findings (audit_id, remediation_due_date) WHERE remediation_due_date IS NOT NULL AND status NOT IN ('verified', 'closed');

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_findings_updated_at ON audit_findings;
CREATE TRIGGER trg_audit_findings_updated_at
    BEFORE UPDATE ON audit_findings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_findings IS 'Audit deficiencies and observations with remediation tracking (spec §6.4)';
COMMENT ON COLUMN audit_findings.severity IS 'critical=material weakness, high=significant deficiency, medium=observation, low=minor, informational=advisory';
COMMENT ON COLUMN audit_findings.management_response IS 'Organization''s formal response to the finding (spec §6.4: management response phase)';
COMMENT ON COLUMN audit_findings.recommendation IS 'Auditor''s recommended remediation action';
