-- Migration: 004_refresh_tokens.sql
-- Description: Refresh tokens for JWT rotation (single-use, revocable)
-- Created: 2026-02-20
-- Sprint: 1 — Project Scaffolding & Auth

-- ============================================================================
-- TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL,
    user_agent      VARCHAR(500),
    ip_address      INET,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Token hash must be globally unique
    CONSTRAINT uq_refresh_tokens_hash UNIQUE (token_hash),

    -- Expiry must be in the future at creation (app enforces, DB validates)
    CONSTRAINT chk_refresh_tokens_expires_future CHECK (expires_at > created_at),

    -- Revoked timestamp, if set, must not precede creation
    CONSTRAINT chk_refresh_tokens_revoked_after_created CHECK (
        revoked_at IS NULL OR revoked_at >= created_at
    )
);

COMMENT ON TABLE refresh_tokens IS 'Active refresh tokens for JWT rotation. Single-use: old token revoked on refresh';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA-256 of the raw token — raw token is only ever sent to client';
COMMENT ON COLUMN refresh_tokens.revoked_at IS 'NULL = active, timestamp = revoked (explicit logout or rotation)';
COMMENT ON COLUMN refresh_tokens.org_id IS 'Denormalized from user for efficient org-scoped queries and cleanup';

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_org_id ON refresh_tokens (org_id);

-- Active (non-revoked) tokens expiring soon — for cleanup jobs
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_active_expires ON refresh_tokens (expires_at)
    WHERE revoked_at IS NULL;

-- Cleanup query support: expired or revoked tokens older than 30 days
-- DELETE FROM refresh_tokens
--   WHERE (revoked_at IS NOT NULL OR expires_at < NOW())
--     AND created_at < NOW() - INTERVAL '30 days';
