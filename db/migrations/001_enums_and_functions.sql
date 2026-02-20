-- Migration: 001_enums_and_functions.sql
-- Description: Enum types and helper functions for Raisin Protect
-- Created: 2026-02-20
-- Sprint: 1 — Project Scaffolding & Auth

-- ============================================================================
-- EXTENSIONS
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

-- GRC roles (spec §1.2)
DO $$ BEGIN
    CREATE TYPE grc_role AS ENUM (
        'compliance_manager',
        'security_engineer',
        'it_admin',
        'ciso',
        'devops_engineer',
        'auditor',
        'vendor_manager'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE grc_role IS 'GRC platform user roles from spec §1.2';

-- Organization status
DO $$ BEGIN
    CREATE TYPE org_status AS ENUM (
        'active',
        'suspended',
        'deactivated'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE org_status IS 'Organization lifecycle states';

-- User account status
DO $$ BEGIN
    CREATE TYPE user_status AS ENUM (
        'active',
        'invited',
        'deactivated',
        'locked'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE user_status IS 'User account lifecycle states';

-- Audit log action categories
DO $$ BEGIN
    CREATE TYPE audit_action AS ENUM (
        'user.register',
        'user.login',
        'user.logout',
        'user.login_failed',
        'user.password_changed',
        'user.updated',
        'user.deactivated',
        'user.reactivated',
        'user.role_assigned',
        'user.role_revoked',
        'org.created',
        'org.updated',
        'org.suspended',
        'org.deactivated',
        'token.refreshed',
        'token.revoked'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE audit_action IS 'Security-relevant actions for immutable audit log';

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Auto-update updated_at timestamp on row modification
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at() IS 'Trigger function: sets updated_at to NOW() before update';

-- Generate URL-safe slug from text input
CREATE OR REPLACE FUNCTION generate_slug(input TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN lower(regexp_replace(trim(input), '[^a-zA-Z0-9]+', '-', 'g'));
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_slug(TEXT) IS 'Generates URL-safe slug from arbitrary text';
