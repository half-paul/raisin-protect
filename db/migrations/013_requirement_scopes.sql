-- Migration: 013_requirement_scopes.sql
-- Description: Per-org requirement scoping decisions (in/out of scope)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

CREATE TABLE IF NOT EXISTS requirement_scopes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requirement_id  UUID NOT NULL REFERENCES requirements(id) ON DELETE CASCADE,
    in_scope        BOOLEAN NOT NULL DEFAULT TRUE,
    justification   TEXT,
    scoped_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_requirement_scope UNIQUE (org_id, requirement_id)
);

COMMENT ON TABLE requirement_scopes IS 'Per-org scoping decisions. No row = in-scope by default. Only stores explicit decisions.';
COMMENT ON COLUMN requirement_scopes.justification IS 'Required when marking out-of-scope (enforced at API layer)';
COMMENT ON COLUMN requirement_scopes.scoped_by IS 'User who made the scoping decision (audit trail)';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_requirement_scopes_org ON requirement_scopes (org_id);
CREATE INDEX IF NOT EXISTS idx_requirement_scopes_requirement ON requirement_scopes (requirement_id);
CREATE INDEX IF NOT EXISTS idx_requirement_scopes_out_of_scope ON requirement_scopes (org_id, in_scope)
    WHERE in_scope = FALSE;

-- Trigger
DROP TRIGGER IF EXISTS trg_requirement_scopes_updated_at ON requirement_scopes;
CREATE TRIGGER trg_requirement_scopes_updated_at
    BEFORE UPDATE ON requirement_scopes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
