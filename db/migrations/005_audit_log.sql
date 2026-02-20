-- Migration: 005_audit_log.sql
-- Description: Immutable, append-only audit log for compliance evidence
-- Created: 2026-02-20
-- Sprint: 1 — Project Scaffolding & Auth

-- ============================================================================
-- TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    action          audit_action NOT NULL,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     UUID,
    metadata        JSONB NOT NULL DEFAULT '{}',
    ip_address      INET,
    user_agent      VARCHAR(500),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- No updated_at: audit logs are immutable per spec §11.3
);

COMMENT ON TABLE audit_log IS 'Immutable, append-only log of security-relevant actions. Compliance backbone.';
COMMENT ON COLUMN audit_log.actor_id IS 'NULL for system-initiated actions. ON DELETE SET NULL preserves history.';
COMMENT ON COLUMN audit_log.resource_type IS 'Affected entity type: user, organization, token';
COMMENT ON COLUMN audit_log.resource_id IS 'ID of affected resource for generic lookups';
COMMENT ON COLUMN audit_log.metadata IS 'Action-specific context (e.g., old_role/new_role, IP, reason)';

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_audit_log_org_id ON audit_log (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor_id ON audit_log (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_org_action ON audit_log (org_id, action);
CREATE INDEX IF NOT EXISTS idx_audit_log_org_resource ON audit_log (org_id, resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_org_created_at ON audit_log (org_id, created_at DESC);

-- ============================================================================
-- IMMUTABILITY PROTECTION
-- ============================================================================

-- Prevent UPDATE and DELETE on audit_log rows
CREATE OR REPLACE FUNCTION audit_log_immutable()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_log is immutable: % operations are not permitted', TG_OP;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_audit_log_no_update ON audit_log;
CREATE TRIGGER trg_audit_log_no_update
    BEFORE UPDATE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_immutable();

DROP TRIGGER IF EXISTS trg_audit_log_no_delete ON audit_log;
CREATE TRIGGER trg_audit_log_no_delete
    BEFORE DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_immutable();

-- Note: For production with 7-year retention (spec §11.3), partition by month:
-- CREATE TABLE audit_log (...) PARTITION BY RANGE (created_at);
