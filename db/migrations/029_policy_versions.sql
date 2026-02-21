-- Migration: 029_policy_versions.sql
-- Description: Policy versions table (rich text content, change tracking, immutable)
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management

-- ============================================================================
-- POLICY VERSIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS policy_versions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id               UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,

    -- Version tracking
    version_number          INT NOT NULL,
    is_current              BOOLEAN NOT NULL DEFAULT TRUE,

    -- Content
    content                 TEXT NOT NULL,
    content_format          policy_content_format NOT NULL DEFAULT 'html',
    content_summary         TEXT,

    -- Change tracking
    change_summary          TEXT,
    change_type             VARCHAR(50) NOT NULL DEFAULT 'minor'
                            CHECK (change_type IN ('major', 'minor', 'patch', 'initial')),

    -- Content metrics (denormalized for UI)
    word_count              INT,
    character_count         INT,

    -- Authorship
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- No updated_at: versions are immutable once created
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_policy_versions_org
    ON policy_versions (org_id);

CREATE INDEX IF NOT EXISTS idx_policy_versions_policy
    ON policy_versions (policy_id);

CREATE INDEX IF NOT EXISTS idx_policy_versions_current
    ON policy_versions (policy_id, is_current)
    WHERE is_current = TRUE;

CREATE INDEX IF NOT EXISTS idx_policy_versions_policy_number
    ON policy_versions (policy_id, version_number DESC);

CREATE INDEX IF NOT EXISTS idx_policy_versions_created_by
    ON policy_versions (created_by)
    WHERE created_by IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_policy_versions_created
    ON policy_versions (policy_id, created_at DESC);

-- Each policy can have only one version with a given number
CREATE UNIQUE INDEX IF NOT EXISTS uq_policy_version_number
    ON policy_versions (policy_id, version_number);

-- Each policy can have only one current version
CREATE UNIQUE INDEX IF NOT EXISTS uq_policy_version_current
    ON policy_versions (policy_id)
    WHERE is_current = TRUE;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE policy_versions IS 'Immutable policy version content with change tracking. Each edit creates a new version.';
COMMENT ON COLUMN policy_versions.version_number IS 'Sequential version number: 1, 2, 3...';
COMMENT ON COLUMN policy_versions.is_current IS 'TRUE = latest version. Enforced unique via partial index.';
COMMENT ON COLUMN policy_versions.content IS 'Full policy body text (HTML, Markdown, or plain text). HTML is sanitized on write.';
COMMENT ON COLUMN policy_versions.change_type IS 'Edit category: initial, major, minor, patch';
COMMENT ON COLUMN policy_versions.change_summary IS 'What changed in this version — for version comparison views';
