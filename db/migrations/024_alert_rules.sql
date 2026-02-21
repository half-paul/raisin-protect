-- Migration: 024_alert_rules.sql
-- Description: Alert rules table — org-configurable alert generation rules
-- Created: 2026-02-20
-- Sprint: 4 — Continuous Monitoring Engine

-- ============================================================================
-- ALERT RULES TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS alert_rules (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    enabled                 BOOLEAN NOT NULL DEFAULT TRUE,

    -- Matching conditions (all must match — AND logic; NULL = wildcard)
    match_test_types        test_type[],
    match_severities        test_severity[],
    match_result_statuses   test_result_status[],
    match_control_ids       UUID[],
    match_tags              TEXT[],

    -- Threshold (suppress noise)
    consecutive_failures    INT NOT NULL DEFAULT 1,
    cooldown_minutes        INT NOT NULL DEFAULT 0,

    -- Alert generation
    alert_severity          alert_severity NOT NULL,
    alert_title_template    VARCHAR(500),
    auto_assign_to          UUID REFERENCES users(id) ON DELETE SET NULL,

    -- SLA
    sla_hours               INT,

    -- Delivery
    delivery_channels       alert_delivery_channel[] NOT NULL DEFAULT '{in_app}',
    slack_webhook_url       TEXT,
    email_recipients        TEXT[],
    webhook_url             TEXT,
    webhook_headers         JSONB NOT NULL DEFAULT '{}',

    -- Priority / ordering
    priority                INT NOT NULL DEFAULT 100,

    -- Ownership
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_alert_rule_name UNIQUE (org_id, name),
    CONSTRAINT chk_sla_hours_positive CHECK (sla_hours IS NULL OR sla_hours > 0),
    CONSTRAINT chk_consecutive_positive CHECK (consecutive_failures > 0 AND consecutive_failures <= 100),
    CONSTRAINT chk_cooldown_nonneg CHECK (cooldown_minutes >= 0 AND cooldown_minutes <= 10080),
    CONSTRAINT chk_priority_range CHECK (priority >= 0 AND priority <= 1000)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_alert_rules_org ON alert_rules (org_id);
CREATE INDEX IF NOT EXISTS idx_alert_rules_org_enabled ON alert_rules (org_id, enabled, priority)
    WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_alert_rules_created_by ON alert_rules (created_by)
    WHERE created_by IS NOT NULL;

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_alert_rules_updated_at ON alert_rules;
CREATE TRIGGER trg_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE alert_rules IS 'Org-configurable rules defining when and how alerts are generated';
COMMENT ON COLUMN alert_rules.match_test_types IS 'NULL = match all types; array uses OR logic within';
COMMENT ON COLUMN alert_rules.match_severities IS 'NULL = match all severities; array uses OR logic within';
COMMENT ON COLUMN alert_rules.match_result_statuses IS 'Which result statuses trigger this rule (default: fail)';
COMMENT ON COLUMN alert_rules.consecutive_failures IS 'Require N consecutive failures before alerting (suppress flaps)';
COMMENT ON COLUMN alert_rules.cooldown_minutes IS 'Don''t re-alert for same test within N minutes';
COMMENT ON COLUMN alert_rules.alert_title_template IS 'Mustache-style template: {{test.title}} failed on {{control.identifier}}';
COMMENT ON COLUMN alert_rules.priority IS 'Rule evaluation order — lower = checked first, first match wins';
COMMENT ON COLUMN alert_rules.slack_webhook_url IS 'Should be encrypted at rest (application-level encryption)';
