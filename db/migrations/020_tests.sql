-- Migration: 020_tests.sql
-- Description: Tests table — automated test definitions with schedules
-- Created: 2026-02-20
-- Sprint: 4 — Continuous Monitoring Engine

-- ============================================================================
-- TESTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS tests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    identifier              VARCHAR(50) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,

    -- Classification
    test_type               test_type NOT NULL,
    severity                test_severity NOT NULL DEFAULT 'medium',
    status                  test_status NOT NULL DEFAULT 'draft',

    -- What it validates
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Schedule (cron expression or interval)
    schedule_cron           VARCHAR(100),
    schedule_interval_min   INT,
    next_run_at             TIMESTAMPTZ,
    last_run_at             TIMESTAMPTZ,

    -- Test logic (Compliance as Code — spec §3.2.2)
    test_script             TEXT,
    test_script_language    VARCHAR(30),
    test_config             JSONB NOT NULL DEFAULT '{}',

    -- Execution parameters
    timeout_seconds         INT NOT NULL DEFAULT 300,
    retry_count             INT NOT NULL DEFAULT 0,
    retry_delay_seconds     INT NOT NULL DEFAULT 60,

    -- Metadata
    tags                    TEXT[] DEFAULT '{}',
    source_integration_id   UUID,

    -- Ownership
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_test_identifier UNIQUE (org_id, identifier),
    CONSTRAINT chk_test_schedule CHECK (
        (schedule_cron IS NOT NULL AND schedule_interval_min IS NULL) OR
        (schedule_cron IS NULL AND schedule_interval_min IS NOT NULL) OR
        (schedule_cron IS NULL AND schedule_interval_min IS NULL)
    ),
    CONSTRAINT chk_test_interval_positive CHECK (schedule_interval_min IS NULL OR schedule_interval_min > 0),
    CONSTRAINT chk_test_timeout_positive CHECK (timeout_seconds > 0 AND timeout_seconds <= 3600),
    CONSTRAINT chk_test_retry CHECK (retry_count >= 0 AND retry_count <= 5)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_tests_org ON tests (org_id);
CREATE INDEX IF NOT EXISTS idx_tests_org_status ON tests (org_id, status);
CREATE INDEX IF NOT EXISTS idx_tests_control ON tests (control_id);
CREATE INDEX IF NOT EXISTS idx_tests_type ON tests (org_id, test_type);
CREATE INDEX IF NOT EXISTS idx_tests_severity ON tests (org_id, severity);
CREATE INDEX IF NOT EXISTS idx_tests_next_run ON tests (next_run_at)
    WHERE status = 'active' AND next_run_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tests_identifier ON tests (org_id, identifier);
CREATE INDEX IF NOT EXISTS idx_tests_tags ON tests USING gin (tags);
CREATE INDEX IF NOT EXISTS idx_tests_created_by ON tests (created_by)
    WHERE created_by IS NOT NULL;

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_tests_updated_at ON tests;
CREATE TRIGGER trg_tests_updated_at
    BEFORE UPDATE ON tests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE tests IS 'Automated test definitions — what to check and when';
COMMENT ON COLUMN tests.identifier IS 'Human-readable code like TST-CFG-001 (org-scoped unique)';
COMMENT ON COLUMN tests.schedule_cron IS 'Cron expression for scheduling, e.g. 0 * * * * (every hour)';
COMMENT ON COLUMN tests.schedule_interval_min IS 'Alternative: run every N minutes (mutually exclusive with cron)';
COMMENT ON COLUMN tests.next_run_at IS 'Precomputed next execution time for worker polling';
COMMENT ON COLUMN tests.test_script IS 'Script content for custom test type (shell, Python, JS)';
COMMENT ON COLUMN tests.test_config IS 'Test-specific configuration: thresholds, params, endpoints';
COMMENT ON COLUMN tests.source_integration_id IS 'Forward reference to Sprint 9 integrations table';
