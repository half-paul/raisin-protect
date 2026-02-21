-- Migration: 022_test_results.sql
-- Description: Test results table — individual test outcomes within a sweep
-- Created: 2026-02-20
-- Sprint: 4 — Continuous Monitoring Engine

-- ============================================================================
-- TEST RESULTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS test_results (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    test_run_id             UUID NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
    test_id                 UUID NOT NULL REFERENCES tests(id) ON DELETE CASCADE,

    -- Denormalized for query performance
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Result
    status                  test_result_status NOT NULL,
    severity                test_severity NOT NULL,

    -- Details
    message                 TEXT,
    details                 JSONB NOT NULL DEFAULT '{}',
    output_log              TEXT,
    error_message           TEXT,

    -- Timing
    started_at              TIMESTAMPTZ NOT NULL,
    completed_at            TIMESTAMPTZ,
    duration_ms             INT,

    -- Alert linkage (FK added in 025_sprint4_fk_cross_refs.sql)
    alert_generated         BOOLEAN NOT NULL DEFAULT FALSE,
    alert_id                UUID,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- No updated_at: test results are immutable once written
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_test_results_org ON test_results (org_id);
CREATE INDEX IF NOT EXISTS idx_test_results_run ON test_results (test_run_id);
CREATE INDEX IF NOT EXISTS idx_test_results_test ON test_results (test_id);
CREATE INDEX IF NOT EXISTS idx_test_results_control ON test_results (control_id);
CREATE INDEX IF NOT EXISTS idx_test_results_status ON test_results (org_id, status);
CREATE INDEX IF NOT EXISTS idx_test_results_severity ON test_results (org_id, severity);
CREATE INDEX IF NOT EXISTS idx_test_results_run_status ON test_results (test_run_id, status);
CREATE INDEX IF NOT EXISTS idx_test_results_control_latest ON test_results (control_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_results_test_latest ON test_results (test_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_results_failures ON test_results (org_id, created_at DESC)
    WHERE status IN ('fail', 'error');

-- Unique constraint: each test runs once per test_run
CREATE UNIQUE INDEX IF NOT EXISTS uq_test_results_run_test ON test_results (test_run_id, test_id);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE test_results IS 'Individual test outcomes within a test run — immutable once written';
COMMENT ON COLUMN test_results.control_id IS 'Denormalized from tests.control_id for fast control health queries';
COMMENT ON COLUMN test_results.severity IS 'Snapshotted from test definition at execution time';
COMMENT ON COLUMN test_results.details IS 'Structured output: actual vs expected values (test-type-specific)';
COMMENT ON COLUMN test_results.output_log IS 'Raw script output for custom tests (capped at 64KB in API layer)';
COMMENT ON COLUMN test_results.alert_id IS 'FK to alerts table — added via deferred constraint in 025_sprint4_fk_cross_refs.sql';
