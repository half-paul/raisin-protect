-- Migration: 021_test_runs.sql
-- Description: Test runs table — batch execution sweeps
-- Created: 2026-02-20
-- Sprint: 4 — Continuous Monitoring Engine

-- ============================================================================
-- TEST RUNS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS test_runs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Run metadata
    run_number              SERIAL,
    status                  test_run_status NOT NULL DEFAULT 'pending',
    trigger_type            test_run_trigger NOT NULL DEFAULT 'scheduled',

    -- Timing
    started_at              TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    duration_ms             INT,

    -- Results summary (denormalized for fast dashboard queries)
    total_tests             INT NOT NULL DEFAULT 0,
    passed                  INT NOT NULL DEFAULT 0,
    failed                  INT NOT NULL DEFAULT 0,
    errors                  INT NOT NULL DEFAULT 0,
    skipped                 INT NOT NULL DEFAULT 0,
    warnings                INT NOT NULL DEFAULT 0,

    -- Who/what triggered it
    triggered_by            UUID REFERENCES users(id) ON DELETE SET NULL,
    trigger_metadata        JSONB NOT NULL DEFAULT '{}',

    -- Worker tracking
    worker_id               VARCHAR(100),
    error_message           TEXT,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_test_runs_org ON test_runs (org_id);
CREATE INDEX IF NOT EXISTS idx_test_runs_org_status ON test_runs (org_id, status);
CREATE INDEX IF NOT EXISTS idx_test_runs_org_created ON test_runs (org_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_runs_trigger ON test_runs (org_id, trigger_type);
CREATE INDEX IF NOT EXISTS idx_test_runs_started ON test_runs (org_id, started_at DESC)
    WHERE started_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_test_runs_pending ON test_runs (status, created_at)
    WHERE status = 'pending';

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_test_runs_updated_at ON test_runs;
CREATE TRIGGER trg_test_runs_updated_at
    BEFORE UPDATE ON test_runs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE test_runs IS 'Batch test execution records — a sweep containing multiple test results';
COMMENT ON COLUMN test_runs.run_number IS 'Auto-incrementing run counter for human-readable display';
COMMENT ON COLUMN test_runs.duration_ms IS 'Total wall-clock time computed as completed_at - started_at';
COMMENT ON COLUMN test_runs.worker_id IS 'Which worker instance processed this run (for debugging)';
COMMENT ON COLUMN test_runs.trigger_metadata IS 'Additional context: webhook source, change details, etc.';
