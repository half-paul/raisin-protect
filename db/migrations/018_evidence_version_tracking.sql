-- Migration: 018_evidence_version_tracking.sql
-- Description: Evidence version history tracking (view + helper function)
-- Created: 2026-02-20
-- Sprint: 3 — Evidence Management
--
-- The version chain is stored in evidence_artifacts via parent_artifact_id.
-- This migration adds:
-- 1. A view for easy version history queries
-- 2. A function to supersede old versions when a new one is uploaded

-- ============================================================================
-- VERSION HISTORY VIEW
-- ============================================================================

CREATE OR REPLACE VIEW evidence_version_history AS
SELECT
    ea.id,
    ea.org_id,
    -- The root artifact ID: either itself (if no parent) or the parent
    COALESCE(ea.parent_artifact_id, ea.id) AS root_artifact_id,
    ea.parent_artifact_id,
    ea.version,
    ea.is_current,
    ea.title,
    ea.status,
    ea.evidence_type,
    ea.file_name,
    ea.file_size,
    ea.mime_type,
    ea.object_key,
    ea.checksum_sha256,
    ea.collection_date,
    ea.expires_at,
    ea.uploaded_by,
    ea.created_at
FROM evidence_artifacts ea
ORDER BY ea.org_id, COALESCE(ea.parent_artifact_id, ea.id), ea.version DESC;

COMMENT ON VIEW evidence_version_history IS 'Flattened view of evidence artifact version chains, grouped by root_artifact_id';

-- ============================================================================
-- FUNCTION: Supersede previous versions when a new version is created
-- ============================================================================
-- Call this after inserting a new version to mark old ones as superseded.
-- Idempotent: only updates rows that aren't already superseded.

CREATE OR REPLACE FUNCTION supersede_evidence_versions(
    p_parent_artifact_id UUID,
    p_new_artifact_id UUID,
    p_org_id UUID
) RETURNS INT AS $$
DECLARE
    rows_updated INT;
BEGIN
    UPDATE evidence_artifacts
    SET
        is_current = FALSE,
        status = 'superseded'
    WHERE org_id = p_org_id
      AND is_current = TRUE
      AND status != 'superseded'
      AND (
          id = p_parent_artifact_id
          OR parent_artifact_id = p_parent_artifact_id
      )
      AND id != p_new_artifact_id;

    GET DIAGNOSTICS rows_updated = ROW_COUNT;
    RETURN rows_updated;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION supersede_evidence_versions IS 'Mark all previous versions of an artifact as superseded when a new version is uploaded';

-- ============================================================================
-- FUNCTION: Calculate evidence freshness status
-- ============================================================================
-- Returns: 'fresh', 'expiring_soon' (within 30 days), 'expired', or 'no_expiry'

CREATE OR REPLACE FUNCTION evidence_freshness_status(
    p_expires_at TIMESTAMPTZ,
    p_warning_days INT DEFAULT 30
) RETURNS TEXT AS $$
BEGIN
    IF p_expires_at IS NULL THEN
        RETURN 'no_expiry';
    ELSIF p_expires_at < NOW() THEN
        RETURN 'expired';
    ELSIF p_expires_at < NOW() + (p_warning_days || ' days')::INTERVAL THEN
        RETURN 'expiring_soon';
    ELSE
        RETURN 'fresh';
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION evidence_freshness_status IS 'Calculate freshness status for an evidence artifact based on expiration date';
