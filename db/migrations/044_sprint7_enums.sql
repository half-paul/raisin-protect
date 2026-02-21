-- Migration: 044_sprint7_enums.sql
-- Description: Enum types for Sprint 7 (Audit Hub)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub

-- ============================================================================
-- NEW ENUM TYPES (9)
-- ============================================================================

-- Audit engagement lifecycle
DO $$ BEGIN
    CREATE TYPE audit_status AS ENUM (
        'planning',
        'fieldwork',
        'review',
        'draft_report',
        'management_response',
        'final_report',
        'completed',
        'cancelled'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_status IS 'Audit engagement lifecycle: planning → fieldwork → review → draft_report → management_response → final_report → completed | cancelled';

-- Audit engagement type
DO $$ BEGIN
    CREATE TYPE audit_type AS ENUM (
        'soc2_type1',
        'soc2_type2',
        'iso27001_certification',
        'iso27001_surveillance',
        'pci_dss_roc',
        'pci_dss_saq',
        'gdpr_dpia',
        'internal',
        'custom'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_type IS 'Type of audit engagement: SOC 2 Type I/II, ISO 27001, PCI DSS ROC/SAQ, GDPR DPIA, internal, custom';

-- Evidence request lifecycle
DO $$ BEGIN
    CREATE TYPE audit_request_status AS ENUM (
        'open',
        'in_progress',
        'submitted',
        'accepted',
        'rejected',
        'closed',
        'overdue'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_request_status IS 'Evidence request lifecycle: open → in_progress → submitted → accepted/rejected/closed/overdue';

-- Priority for evidence requests
DO $$ BEGIN
    CREATE TYPE audit_request_priority AS ENUM (
        'critical',
        'high',
        'medium',
        'low'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_request_priority IS 'Priority level for evidence requests: critical, high, medium, low';

-- Audit finding severity
DO $$ BEGIN
    CREATE TYPE audit_finding_severity AS ENUM (
        'critical',
        'high',
        'medium',
        'low',
        'informational'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_finding_severity IS 'Severity: critical=material weakness, high=significant deficiency, medium=observation, low=minor, informational=advisory';

-- Finding lifecycle with remediation
DO $$ BEGIN
    CREATE TYPE audit_finding_status AS ENUM (
        'identified',
        'acknowledged',
        'remediation_planned',
        'remediation_in_progress',
        'remediation_complete',
        'verified',
        'risk_accepted',
        'closed'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_finding_status IS 'Finding lifecycle: identified → acknowledged → remediation_planned → in_progress → complete → verified | risk_accepted → closed';

-- Finding category aligned with audit standards
DO $$ BEGIN
    CREATE TYPE audit_finding_category AS ENUM (
        'control_deficiency',
        'control_gap',
        'documentation_gap',
        'process_gap',
        'configuration_issue',
        'access_control',
        'monitoring_gap',
        'policy_violation',
        'vendor_risk',
        'data_handling',
        'other'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_finding_category IS 'Finding category: control_deficiency, control_gap, documentation_gap, process_gap, configuration_issue, access_control, monitoring_gap, policy_violation, vendor_risk, data_handling, other';

-- Polymorphic comment target type
DO $$ BEGIN
    CREATE TYPE audit_comment_target_type AS ENUM (
        'audit',
        'request',
        'finding'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_comment_target_type IS 'Polymorphic target for audit comments: audit, request, or finding';

-- Auditor review status for submitted evidence
DO $$ BEGIN
    CREATE TYPE audit_evidence_link_status AS ENUM (
        'pending_review',
        'accepted',
        'rejected',
        'needs_clarification'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_evidence_link_status IS 'Auditor review status for submitted evidence: pending_review → accepted/rejected/needs_clarification';

-- ============================================================================
-- EXTEND EXISTING ENUMS
-- ============================================================================

-- Add Sprint 7 actions to audit_action enum (20 new values)
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.created'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.updated'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.status_changed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.completed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.cancelled'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.auditor_added'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.auditor_removed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.created'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.assigned'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.submitted'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.accepted'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.rejected'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.closed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.created'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.updated'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.status_changed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.remediation_planned'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.verified'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_evidence.submitted'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_evidence.reviewed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- Extend evidence_link_target_type to support audits
DO $$ BEGIN ALTER TYPE evidence_link_target_type ADD VALUE IF NOT EXISTS 'audit'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
