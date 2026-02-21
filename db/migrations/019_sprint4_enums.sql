-- Migration: 019_sprint4_enums.sql
-- Description: Enum types for Sprint 4 (Continuous Monitoring Engine)
-- Created: 2026-02-20
-- Sprint: 4 — Continuous Monitoring Engine

-- ============================================================================
-- NEW ENUM TYPES
-- ============================================================================

-- Categories of automated tests (spec §3.3.1)
DO $$ BEGIN
    CREATE TYPE test_type AS ENUM (
        'configuration',
        'access_control',
        'endpoint',
        'vulnerability',
        'data_protection',
        'network',
        'logging',
        'custom'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE test_type IS 'Categories of automated tests from spec §3.3.1';

-- Test definition lifecycle
DO $$ BEGIN
    CREATE TYPE test_status AS ENUM (
        'draft',
        'active',
        'paused',
        'deprecated'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE test_status IS 'Test definition lifecycle states';

-- Severity of a test (determines alert severity on failure)
DO $$ BEGIN
    CREATE TYPE test_severity AS ENUM (
        'critical',
        'high',
        'medium',
        'low',
        'informational'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE test_severity IS 'Test severity levels determining alert severity on failure';

-- Outcome of an individual test execution
DO $$ BEGIN
    CREATE TYPE test_result_status AS ENUM (
        'pass',
        'fail',
        'error',
        'skip',
        'warning'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE test_result_status IS 'Individual test execution outcome states';

-- Lifecycle of a test run (batch sweep)
DO $$ BEGIN
    CREATE TYPE test_run_status AS ENUM (
        'pending',
        'running',
        'completed',
        'failed',
        'cancelled'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE test_run_status IS 'Batch test run (sweep) lifecycle states';

-- What triggered a test run
DO $$ BEGIN
    CREATE TYPE test_run_trigger AS ENUM (
        'scheduled',
        'manual',
        'on_change',
        'webhook'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE test_run_trigger IS 'Test run trigger sources';

-- Alert severity (spec §3.3.2)
DO $$ BEGIN
    CREATE TYPE alert_severity AS ENUM (
        'critical',
        'high',
        'medium',
        'low'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE alert_severity IS 'Alert severity levels from spec §3.3.2';

-- Alert lifecycle (spec §3.3.2 workflow)
DO $$ BEGIN
    CREATE TYPE alert_status AS ENUM (
        'open',
        'acknowledged',
        'in_progress',
        'resolved',
        'suppressed',
        'closed'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE alert_status IS 'Alert lifecycle states from spec §3.3.2';

-- How alerts are delivered
DO $$ BEGIN
    CREATE TYPE alert_delivery_channel AS ENUM (
        'slack',
        'email',
        'webhook',
        'in_app'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE alert_delivery_channel IS 'Alert delivery channel types';

-- ============================================================================
-- EXTEND EXISTING ENUMS (Sprint 4 audit actions)
-- ============================================================================

ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test.status_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test_run.started';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test_run.completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test_run.failed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'test_run.cancelled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.acknowledged';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.assigned';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.status_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.resolved';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.suppressed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.closed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert.reopened';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert_rule.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert_rule.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'alert_rule.deleted';
