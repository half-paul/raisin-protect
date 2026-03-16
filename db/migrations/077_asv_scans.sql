-- Migration: 077_asv_scans.sql
-- Description: ASV Scan Management — quarterly external vulnerability scan tracking
--              per PCI DSS v4.0.1 Requirement 11.3.2
-- Created: 2026-03-15
-- Sprint: 12 — ASV Scan Import + Quarterly Tracking (LEX-94)
--
-- PCI DSS Req 11.3.2: External vulnerability scans are performed by an ASV
-- (Approved Scanning Vendor) at least quarterly, with all findings of high
-- severity or higher remediated and re-scanned.

-- ============================================================================
-- TABLE: asv_scans
-- One row per ASV scan engagement. Tracks quarterly compliance across years
-- and stores import metadata for CSV/XML scan reports.
-- ============================================================================

CREATE TABLE IF NOT EXISTS asv_scans (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Vendor and scan classification
    asv_vendor              TEXT NOT NULL,          -- e.g., "Qualys", "Tenable", "Rapid7", "SecurityMetrics"
    scan_type               TEXT NOT NULL DEFAULT 'external'
                                CHECK (scan_type IN ('external', 'internal')),

    -- Quarterly tracking (for compliance calendar)
    quarter                 INTEGER NOT NULL        -- 1–4
                                CHECK (quarter BETWEEN 1 AND 4),
    year                    INTEGER NOT NULL        -- e.g., 2026
                                CHECK (year >= 2000 AND year <= 2100),
    scan_date               DATE NOT NULL,

    -- Status
    status                  TEXT NOT NULL DEFAULT 'in_progress'
                                CHECK (status IN (
                                    'in_progress',  -- Scan running or results not yet imported
                                    'pass',         -- No high/critical findings, or all remediated
                                    'fail',         -- High/critical findings remain unresolved
                                    'remediated'    -- Previously failed, now remediated + re-scanned
                                )),

    -- Finding counts (summary from imported report)
    findings_count          INTEGER NOT NULL DEFAULT 0,
    critical_count          INTEGER NOT NULL DEFAULT 0,
    high_count              INTEGER NOT NULL DEFAULT 0,
    medium_count            INTEGER NOT NULL DEFAULT 0,
    low_count               INTEGER NOT NULL DEFAULT 0,
    informational_count     INTEGER NOT NULL DEFAULT 0,

    -- Remediation
    remediation_deadline    DATE,           -- Typically 30 days from scan_date for critical/high

    -- Import metadata
    report_path             TEXT,           -- MinIO object key for uploaded scan report (PDF/CSV/XML)
    import_format           TEXT
                                CHECK (import_format IS NULL OR import_format IN (
                                    'csv', 'xml', 'pdf', 'json', 'manual'
                                )),
    raw_findings            JSONB,          -- Structured findings imported from ASV report
    import_notes            TEXT,

    -- Audit
    imported_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One scan per (org, type, quarter, year) enforces quarterly tracking discipline
    CONSTRAINT uq_asv_scan_quarter UNIQUE (org_id, scan_type, quarter, year),
    -- Finding counts are non-negative
    CONSTRAINT chk_asv_findings_nonneg CHECK (
        findings_count >= 0 AND critical_count >= 0 AND high_count >= 0
        AND medium_count >= 0 AND low_count >= 0 AND informational_count >= 0
    )
);

COMMENT ON TABLE asv_scans IS
    'ASV (Approved Scanning Vendor) scan records for PCI DSS v4.0.1 Req 11.3.2. '
    'Enforces quarterly tracking with one external scan record per (org, quarter, year). '
    'Scan findings are summarized here; detailed findings live in raw_findings JSONB.';
COMMENT ON COLUMN asv_scans.quarter IS
    'Calendar quarter (1=Jan-Mar, 2=Apr-Jun, 3=Jul-Sep, 4=Oct-Dec). '
    'Combined with year, allows compliance calendar queries.';
COMMENT ON COLUMN asv_scans.status IS
    'pass=no high/critical findings remain; fail=unresolved findings; '
    'remediated=was fail, now passing after re-scan.';
COMMENT ON COLUMN asv_scans.raw_findings IS
    'Structured array of findings imported from ASV report. Each finding: '
    '{"id":"...", "host":"...", "severity":"high", "cve":"CVE-...", "description":"...", "remediated":false}';
COMMENT ON COLUMN asv_scans.remediation_deadline IS
    'Target date by which all critical/high findings must be remediated and re-scanned. '
    'Typically 30 days from scan_date per PCI DSS Req 11.3.2.1.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_asv_scans_org                ON asv_scans (org_id);
CREATE INDEX IF NOT EXISTS idx_asv_scans_org_status         ON asv_scans (org_id, status);
CREATE INDEX IF NOT EXISTS idx_asv_scans_org_year_quarter   ON asv_scans (org_id, year DESC, quarter DESC);
CREATE INDEX IF NOT EXISTS idx_asv_scans_org_type           ON asv_scans (org_id, scan_type);
CREATE INDEX IF NOT EXISTS idx_asv_scans_remediation        ON asv_scans (org_id, remediation_deadline)
    WHERE remediation_deadline IS NOT NULL AND status = 'fail';

-- RLS
ALTER TABLE asv_scans ENABLE ROW LEVEL SECURITY;
CREATE POLICY asv_scans_org_isolation ON asv_scans
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_asv_scans_updated_at ON asv_scans;
CREATE TRIGGER trg_asv_scans_updated_at
    BEFORE UPDATE ON asv_scans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- Extend audit_action enum for ASV scan events
-- ============================================================================

DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'asv_scan.created';     EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'asv_scan.imported';    EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'asv_scan.status_changed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'asv_scan.overdue';     EXCEPTION WHEN duplicate_object THEN NULL; END $$;
