-- Migration: 003_users.sql
-- Description: Users table with GRC roles — one user belongs to one org
-- Created: 2026-02-20
-- Sprint: 1 — Project Scaffolding & Auth

-- ============================================================================
-- TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email               VARCHAR(255) NOT NULL,
    password_hash       VARCHAR(255) NOT NULL,
    first_name          VARCHAR(100) NOT NULL,
    last_name           VARCHAR(100) NOT NULL,
    role                grc_role NOT NULL DEFAULT 'compliance_manager',
    status              user_status NOT NULL DEFAULT 'active',
    mfa_enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret          VARCHAR(255),
    last_login_at       TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Email uniqueness scoped to org (same person may exist across orgs)
    CONSTRAINT uq_users_org_email UNIQUE (org_id, email),

    -- Data integrity checks
    CONSTRAINT chk_users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT chk_users_first_name_not_empty CHECK (length(trim(first_name)) > 0),
    CONSTRAINT chk_users_last_name_not_empty CHECK (length(trim(last_name)) > 0),
    CONSTRAINT chk_users_password_hash_not_empty CHECK (length(password_hash) > 0),
    CONSTRAINT chk_users_mfa_secret_when_enabled CHECK (
        (mfa_enabled = FALSE) OR (mfa_enabled = TRUE AND mfa_secret IS NOT NULL)
    )
);

COMMENT ON TABLE users IS 'Users belong to exactly one org. Single GRC role per user (Sprint 1 simplicity)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash — hashing done at app layer';
COMMENT ON COLUMN users.mfa_secret IS 'TOTP secret — encrypted at app level. Column present but MFA enforcement comes later';
COMMENT ON COLUMN users.role IS 'Single GRC role. Future: migrate to user_roles junction for multi-role';

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_users_org_id ON users (org_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_org_role ON users (org_id, role);
CREATE INDEX IF NOT EXISTS idx_users_org_status ON users (org_id, status);

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
