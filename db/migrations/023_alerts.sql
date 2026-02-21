-- Migration: 023_alerts.sql
-- Description: Alerts table — generated from test failures with full lifecycle
-- Created: 2026-02-20
-- Sprint: 4 — Continuous Monitoring Engine

-- ============================================================================
-- ALERTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS alerts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    alert_number            SERIAL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,

    -- Classification
    severity                alert_severity NOT NULL,
    status                  alert_status NOT NULL DEFAULT 'open',

    -- Source (what triggered this alert)
    test_id                 UUID REFERENCES tests(id) ON DELETE SET NULL,
    test_result_id          UUID REFERENCES test_results(id) ON DELETE SET NULL,
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    alert_rule_id           UUID,  -- FK added in 025_sprint4_fk_cross_refs.sql

    -- Assignment
    assigned_to             UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at             TIMESTAMPTZ,
    assigned_by             UUID REFERENCES users(id) ON DELETE SET NULL,

    -- SLA tracking
    sla_deadline            TIMESTAMPTZ,
    sla_breached            BOOLEAN NOT NULL DEFAULT FALSE,

    -- Resolution
    resolved_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_at             TIMESTAMPTZ,
    resolution_notes        TEXT,

    -- Suppression (spec §3.3.2: snooze/suppress with justification)
    suppressed_until        TIMESTAMPTZ,
    suppression_reason      TEXT,

    -- Delivery tracking
    delivery_channels       alert_delivery_channel[] DEFAULT '{in_app}',
    delivered_at            JSONB NOT NULL DEFAULT '{}',

    -- Metadata
    tags                    TEXT[] DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_alerts_org ON alerts (org_id);
CREATE INDEX IF NOT EXISTS idx_alerts_org_status ON alerts (org_id, status);
CREATE INDEX IF NOT EXISTS idx_alerts_org_severity ON alerts (org_id, severity);
CREATE INDEX IF NOT EXISTS idx_alerts_org_status_severity ON alerts (org_id, status, severity);
CREATE INDEX IF NOT EXISTS idx_alerts_control ON alerts (control_id);
CREATE INDEX IF NOT EXISTS idx_alerts_test ON alerts (test_id)
    WHERE test_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_test_result ON alerts (test_result_id)
    WHERE test_result_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_assigned_to ON alerts (assigned_to)
    WHERE assigned_to IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_sla_deadline ON alerts (org_id, sla_deadline)
    WHERE status NOT IN ('resolved', 'closed', 'suppressed') AND sla_deadline IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_sla_breached ON alerts (org_id, sla_breached)
    WHERE sla_breached = TRUE AND status NOT IN ('resolved', 'closed');
CREATE INDEX IF NOT EXISTS idx_alerts_created ON alerts (org_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_open ON alerts (org_id, created_at DESC)
    WHERE status IN ('open', 'acknowledged', 'in_progress');
CREATE INDEX IF NOT EXISTS idx_alerts_suppressed ON alerts (org_id, suppressed_until)
    WHERE status = 'suppressed' AND suppressed_until IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_tags ON alerts USING gin (tags);

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_alerts_updated_at ON alerts;
CREATE TRIGGER trg_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE alerts IS 'Generated alerts from test failures with full lifecycle tracking (spec §3.3.2)';
COMMENT ON COLUMN alerts.alert_number IS 'Auto-incrementing for human-readable display (ALT-42)';
COMMENT ON COLUMN alerts.sla_deadline IS 'Computed at creation from severity + org SLA config';
COMMENT ON COLUMN alerts.sla_breached IS 'Set TRUE by periodic worker when deadline passes without resolution';
COMMENT ON COLUMN alerts.suppressed_until IS 'Expiration of snooze — alert auto-reopens after this time';
COMMENT ON COLUMN alerts.suppression_reason IS 'Mandatory justification when suppressing (spec §3.3.2)';
COMMENT ON COLUMN alerts.delivery_channels IS 'Array of delivery methods for this alert';
COMMENT ON COLUMN alerts.delivered_at IS 'Tracks successful delivery per channel: {"slack": "...", "email": "..."}';
COMMENT ON COLUMN alerts.alert_rule_id IS 'FK to alert_rules — added via deferred constraint in 025_sprint4_fk_cross_refs.sql';
