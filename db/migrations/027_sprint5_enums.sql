-- Migration: 027_sprint5_enums.sql
-- Description: Enum types for Sprint 5 (Policy Management)
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management

-- ============================================================================
-- NEW ENUM TYPES
-- ============================================================================

-- Categories of policies (from spec §6.1 + industry standard taxonomy)
DO $$ BEGIN
    CREATE TYPE policy_category AS ENUM (
        'information_security',
        'acceptable_use',
        'access_control',
        'data_classification',
        'data_privacy',
        'data_retention',
        'incident_response',
        'business_continuity',
        'change_management',
        'vulnerability_management',
        'vendor_management',
        'physical_security',
        'encryption',
        'network_security',
        'secure_development',
        'human_resources',
        'compliance',
        'risk_management',
        'asset_management',
        'logging_monitoring',
        'custom'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE policy_category IS 'Categories of policies from spec §6.1 + industry standard taxonomy';

-- Policy lifecycle (from spec §6.1)
DO $$ BEGIN
    CREATE TYPE policy_status AS ENUM (
        'draft',
        'in_review',
        'approved',
        'published',
        'archived'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE policy_status IS 'Policy lifecycle states: draft → in_review → approved → published → archived';

-- Sign-off decision status
DO $$ BEGIN
    CREATE TYPE signoff_status AS ENUM (
        'pending',
        'approved',
        'rejected',
        'withdrawn'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE signoff_status IS 'Sign-off decision states for policy approval workflow';

-- Policy content format
DO $$ BEGIN
    CREATE TYPE policy_content_format AS ENUM (
        'html',
        'markdown',
        'plain_text'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE policy_content_format IS 'Policy content storage format: html (rich text editor), markdown, or plain_text';

-- ============================================================================
-- EXTEND EXISTING ENUMS (Sprint 5 audit actions)
-- ============================================================================

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy.status_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy.archived';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy.owner_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy.cloned_from_template';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_version.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_version.published';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_signoff.requested';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_signoff.approved';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_signoff.rejected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_signoff.withdrawn';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_control.linked';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'policy_control.unlinked';
