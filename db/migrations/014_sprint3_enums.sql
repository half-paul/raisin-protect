-- Migration: 014_sprint3_enums.sql
-- Description: Enum types for Sprint 3 (Evidence Management)
-- Created: 2026-02-20
-- Sprint: 3 — Evidence Management

-- ============================================================================
-- NEW ENUM TYPES
-- ============================================================================

-- Types of evidence artifacts (spec §3.4.1)
DO $$ BEGIN
    CREATE TYPE evidence_type AS ENUM (
        'screenshot',
        'api_response',
        'configuration_export',
        'log_sample',
        'policy_document',
        'access_list',
        'vulnerability_report',
        'certificate',
        'training_record',
        'penetration_test',
        'audit_report',
        'other'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE evidence_type IS 'Types of evidence artifacts from spec §3.4.1';

-- Evidence lifecycle status
DO $$ BEGIN
    CREATE TYPE evidence_status AS ENUM (
        'draft',
        'pending_review',
        'approved',
        'rejected',
        'expired',
        'superseded'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE evidence_status IS 'Evidence artifact lifecycle states';

-- How evidence was collected (spec §3.4.1)
DO $$ BEGIN
    CREATE TYPE evidence_collection_method AS ENUM (
        'manual_upload',
        'automated_pull',
        'api_ingestion',
        'screenshot_capture',
        'system_export'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE evidence_collection_method IS 'Evidence collection methods from spec §3.4.1';

-- Evaluation verdict
DO $$ BEGIN
    CREATE TYPE evidence_evaluation_verdict AS ENUM (
        'sufficient',
        'partial',
        'insufficient',
        'needs_update'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE evidence_evaluation_verdict IS 'Evidence evaluation verdicts';

-- What entity type an evidence link targets
DO $$ BEGIN
    CREATE TYPE evidence_link_target_type AS ENUM (
        'control',
        'requirement',
        'policy'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE evidence_link_target_type IS 'Polymorphic target types for evidence links';

-- ============================================================================
-- EXTEND EXISTING ENUMS (Sprint 3 audit actions)
-- ============================================================================

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.uploaded';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.version_created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.status_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.linked';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.unlinked';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.evaluated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'evidence.expired';
