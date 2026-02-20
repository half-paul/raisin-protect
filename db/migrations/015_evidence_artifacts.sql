-- Migration: 015_evidence_artifacts.sql
-- Description: Evidence artifacts table (core evidence storage metadata)
-- Created: 2026-02-20
-- Sprint: 3 — Evidence Management

-- ============================================================================
-- EVIDENCE ARTIFACTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS evidence_artifacts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity & display
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    evidence_type           evidence_type NOT NULL,
    status                  evidence_status NOT NULL DEFAULT 'draft',
    collection_method       evidence_collection_method NOT NULL DEFAULT 'manual_upload',

    -- File storage (MinIO)
    file_name               VARCHAR(500) NOT NULL,
    file_size               BIGINT NOT NULL CHECK (file_size > 0),
    mime_type               VARCHAR(255) NOT NULL,
    object_key              VARCHAR(1000) NOT NULL UNIQUE,
    checksum_sha256         VARCHAR(64),

    -- Versioning
    parent_artifact_id      UUID REFERENCES evidence_artifacts(id) ON DELETE SET NULL,
    version                 INT NOT NULL DEFAULT 1 CHECK (version >= 1),
    is_current              BOOLEAN NOT NULL DEFAULT TRUE,

    -- Freshness tracking (spec §3.4.1)
    collection_date         DATE NOT NULL,
    expires_at              TIMESTAMPTZ,
    freshness_period_days   INT CHECK (freshness_period_days IS NULL OR freshness_period_days > 0),

    -- Source attribution
    source_system           VARCHAR(255),
    source_integration_id   UUID,  -- FK to integrations table (Sprint 9, nullable for now)

    -- Ownership
    uploaded_by             UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Tags for search/filter
    tags                    TEXT[] DEFAULT '{}',

    -- Metadata
    metadata                JSONB NOT NULL DEFAULT '{}',

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_org
    ON evidence_artifacts (org_id);

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_status
    ON evidence_artifacts (org_id, status);

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_type
    ON evidence_artifacts (org_id, evidence_type);

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_collection_method
    ON evidence_artifacts (org_id, collection_method);

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_parent
    ON evidence_artifacts (parent_artifact_id)
    WHERE parent_artifact_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_current
    ON evidence_artifacts (org_id, is_current)
    WHERE is_current = TRUE;

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_uploaded_by
    ON evidence_artifacts (uploaded_by);

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_collection_date
    ON evidence_artifacts (org_id, collection_date DESC);

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_expires
    ON evidence_artifacts (org_id, expires_at)
    WHERE expires_at IS NOT NULL AND status NOT IN ('expired', 'superseded');

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_tags
    ON evidence_artifacts USING gin (tags);

CREATE INDEX IF NOT EXISTS idx_evidence_artifacts_search
    ON evidence_artifacts
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_evidence_artifacts_updated_at ON evidence_artifacts;
CREATE TRIGGER trg_evidence_artifacts_updated_at
    BEFORE UPDATE ON evidence_artifacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE evidence_artifacts IS 'Evidence artifacts with file metadata, MinIO object keys, versioning, and freshness tracking (spec §3.4)';
COMMENT ON COLUMN evidence_artifacts.object_key IS 'MinIO object key: {org_id}/{artifact_id}/{version}/{filename}';
COMMENT ON COLUMN evidence_artifacts.parent_artifact_id IS 'Points to the original artifact for version chains';
COMMENT ON COLUMN evidence_artifacts.is_current IS 'TRUE = latest version of this artifact';
COMMENT ON COLUMN evidence_artifacts.freshness_period_days IS 'How often evidence should be refreshed (e.g., 90 for quarterly)';
COMMENT ON COLUMN evidence_artifacts.checksum_sha256 IS 'SHA-256 of file contents for chain-of-custody integrity (spec §3.4.3)';
COMMENT ON COLUMN evidence_artifacts.source_integration_id IS 'FK to integrations table (Sprint 9) — nullable until then';
