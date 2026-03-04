-- Sprint 8 — Enum Types for User Access Reviews

-- identity_provider_type: Supported identity provider systems (stubs for Sprint 8)
DO $$ BEGIN
    CREATE TYPE identity_provider_type AS ENUM (
        'okta',
        'azure_ad',
        'google_workspace',
        'jumpcloud',
        'onelogin',
        'custom'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- identity_provider_status: Connection lifecycle for identity providers
DO $$ BEGIN
    CREATE TYPE identity_provider_status AS ENUM (
        'pending_setup',
        'connected',
        'syncing',
        'error',
        'disconnected'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- resource_criticality: Criticality tier for access resources
DO $$ BEGIN
    CREATE TYPE resource_criticality AS ENUM (
        'critical',
        'high',
        'medium',
        'low'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- resource_type: Category of access resource
DO $$ BEGIN
    CREATE TYPE resource_type AS ENUM (
        'application',
        'infrastructure',
        'directory',
        'custom'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- campaign_status: Access review campaign lifecycle
DO $$ BEGIN
    CREATE TYPE campaign_status AS ENUM (
        'draft',
        'active',
        'in_review',
        'completed',
        'cancelled'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- campaign_cadence: Configurable review frequency
DO $$ BEGIN
    CREATE TYPE campaign_cadence AS ENUM (
        'monthly',
        'quarterly',
        'semi_annual',
        'annual',
        'custom'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- review_decision: Reviewer's decision on an access entry
DO $$ BEGIN
    CREATE TYPE review_decision AS ENUM (
        'pending',
        'approved',
        'revoked',
        'flagged',
        'delegated',
        'expired'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- access_entry_status: Status of an access record pulled from an IdP
DO $$ BEGIN
    CREATE TYPE access_entry_status AS ENUM (
        'active',
        'inactive',
        'orphaned',
        'suspended',
        'pending_revocation'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- access_anomaly_type: Types of access anomalies detected
DO $$ BEGIN
    CREATE TYPE access_anomaly_type AS ENUM (
        'orphaned_account',
        'excessive_privileges',
        'role_drift',
        'stale_access',
        'no_mfa',
        'departed_user'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- audit_action Extensions (18 new values)
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.connected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.disconnected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.sync_started';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.sync_completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_resource.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_resource.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_resource.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.launched';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.cancelled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.assigned';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.decided';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.delegated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.escalated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.bulk_decided';

-- evidence_link_target_type Extension
ALTER TYPE evidence_link_target_type ADD VALUE IF NOT EXISTS 'access_review_campaign';
