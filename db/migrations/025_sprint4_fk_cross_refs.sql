-- Migration: 025_sprint4_fk_cross_refs.sql
-- Description: Deferred foreign keys between Sprint 4 tables
-- Created: 2026-02-20
-- Sprint: 4 — Continuous Monitoring Engine

-- ============================================================================
-- test_results.alert_id → alerts(id)
-- ============================================================================

DO $$ BEGIN
    ALTER TABLE test_results
        ADD CONSTRAINT fk_test_results_alert
        FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_test_results_alert ON test_results (alert_id)
    WHERE alert_id IS NOT NULL;

-- ============================================================================
-- alerts.alert_rule_id → alert_rules(id)
-- ============================================================================

DO $$ BEGIN
    ALTER TABLE alerts
        ADD CONSTRAINT fk_alerts_alert_rule
        FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_alerts_rule ON alerts (alert_rule_id)
    WHERE alert_rule_id IS NOT NULL;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON CONSTRAINT fk_test_results_alert ON test_results IS 'Back-reference: which alert was generated from this test result';
COMMENT ON CONSTRAINT fk_alerts_alert_rule ON alerts IS 'Which alert rule generated this alert';
