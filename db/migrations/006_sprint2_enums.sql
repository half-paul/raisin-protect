-- Migration: 006_sprint2_enums.sql
-- Description: Enum types for Sprint 2 (Frameworks & Controls)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

-- ============================================================================
-- NEW ENUM TYPES
-- ============================================================================

-- Framework taxonomy (spec §3.1.1)
DO $$ BEGIN
    CREATE TYPE framework_category AS ENUM (
        'security_privacy',
        'payment',
        'data_privacy',
        'ai_governance',
        'industry',
        'custom'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE framework_category IS 'Framework taxonomy from spec §3.1.1';

-- Framework version lifecycle
DO $$ BEGIN
    CREATE TYPE framework_version_status AS ENUM (
        'draft',
        'active',
        'deprecated',
        'sunset'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE framework_version_status IS 'Framework version lifecycle states';

-- Control categories (spec §3.2.1)
DO $$ BEGIN
    CREATE TYPE control_category AS ENUM (
        'technical',
        'administrative',
        'physical',
        'operational'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE control_category IS 'Control categories from spec §3.2.1';

-- Control lifecycle (spec §3.2.1)
DO $$ BEGIN
    CREATE TYPE control_status AS ENUM (
        'draft',
        'active',
        'under_review',
        'deprecated'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE control_status IS 'Control lifecycle states';

-- Org framework activation status
DO $$ BEGIN
    CREATE TYPE org_framework_status AS ENUM (
        'active',
        'inactive'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE org_framework_status IS 'Org-level framework activation states';

-- ============================================================================
-- EXTEND EXISTING ENUMS (Sprint 2 audit actions)
-- ============================================================================
-- Note: ALTER TYPE ... ADD VALUE is not transactional, but IS idempotent
-- (Postgres ignores ADD VALUE IF NOT EXISTS since v12)

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'framework.activated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'framework.deactivated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'framework.version_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'control.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'control.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'control.status_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'control.deprecated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'control.owner_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'control_mapping.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'control_mapping.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'requirement.scoped';
