-- Migration: 045_audits.sql
-- Description: Audit engagements table — core record for audit lifecycle (Sprint 7)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub

CREATE TABLE IF NOT EXISTS audits (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Engagement identity
    title               VARCHAR(255) NOT NULL,
    description         TEXT,
    audit_type          audit_type NOT NULL,
    status              audit_status NOT NULL DEFAULT 'planning',

    -- Framework linkage (which framework is being audited)
    org_framework_id    UUID REFERENCES org_frameworks(id) ON DELETE SET NULL,

    -- Audit period (what time range is being audited)
    period_start        DATE,
    period_end          DATE,

    -- Engagement timeline
    planned_start       DATE,
    planned_end         DATE,
    actual_start        DATE,
    actual_end          DATE,

    -- Auditor information
    audit_firm          VARCHAR(255),
    lead_auditor_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    auditor_ids         UUID[] NOT NULL DEFAULT '{}',

    -- Internal team
    internal_lead_id    UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Milestones (lightweight, embedded JSONB)
    -- Array of: { "name": "Kickoff", "target_date": "2026-03-01", "completed_at": null }
    milestones          JSONB NOT NULL DEFAULT '[]',

    -- Report information
    report_type         VARCHAR(100),
    report_url          TEXT,
    report_issued_at    TIMESTAMPTZ,

    -- Summary statistics (denormalized for dashboard performance)
    total_requests      INTEGER NOT NULL DEFAULT 0,
    open_requests       INTEGER NOT NULL DEFAULT 0,
    total_findings      INTEGER NOT NULL DEFAULT 0,
    open_findings       INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    tags                TEXT[] NOT NULL DEFAULT '{}',
    metadata            JSONB NOT NULL DEFAULT '{}',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_audit_period CHECK (
        period_start IS NULL OR period_end IS NULL OR period_start <= period_end
    ),
    CONSTRAINT chk_audit_timeline CHECK (
        planned_start IS NULL OR planned_end IS NULL OR planned_start <= planned_end
    ),
    CONSTRAINT chk_audit_actual CHECK (
        actual_start IS NULL OR actual_end IS NULL OR actual_start <= actual_end
    )
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audits_org ON audits (org_id);
CREATE INDEX IF NOT EXISTS idx_audits_org_status ON audits (org_id, status);
CREATE INDEX IF NOT EXISTS idx_audits_org_framework ON audits (org_id, org_framework_id) WHERE org_framework_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audits_lead_auditor ON audits (lead_auditor_id) WHERE lead_auditor_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audits_internal_lead ON audits (internal_lead_id) WHERE internal_lead_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audits_org_type ON audits (org_id, audit_type);
CREATE INDEX IF NOT EXISTS idx_audits_org_planned_end ON audits (org_id, planned_end) WHERE planned_end IS NOT NULL;

-- Trigger
DROP TRIGGER IF EXISTS trg_audits_updated_at ON audits;
CREATE TRIGGER trg_audits_updated_at
    BEFORE UPDATE ON audits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audits IS 'Audit engagements: SOC 2, ISO 27001, PCI DSS, etc. (spec §6.4)';
COMMENT ON COLUMN audits.org_framework_id IS 'Which activated framework this audit covers';
COMMENT ON COLUMN audits.period_start IS 'Start of the audit observation period';
COMMENT ON COLUMN audits.period_end IS 'End of the audit observation period';
COMMENT ON COLUMN audits.auditor_ids IS 'Array of user IDs with auditor role who can access this engagement';
COMMENT ON COLUMN audits.milestones IS 'JSONB array of {name, target_date, completed_at} milestone objects';
COMMENT ON COLUMN audits.total_requests IS 'Denormalized count of evidence requests (updated by app)';
COMMENT ON COLUMN audits.open_requests IS 'Denormalized count of non-closed requests (updated by app)';
