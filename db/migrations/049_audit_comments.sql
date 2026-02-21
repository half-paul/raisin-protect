-- Migration: 049_audit_comments.sql
-- Description: Audit comments — threaded discussion with visibility control (Sprint 7)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub

CREATE TABLE IF NOT EXISTS audit_comments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id            UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,

    -- Polymorphic target
    target_type         audit_comment_target_type NOT NULL,
    target_id           UUID NOT NULL,

    -- Comment content
    author_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body                TEXT NOT NULL,

    -- Threading
    parent_comment_id   UUID REFERENCES audit_comments(id) ON DELETE CASCADE,

    -- Metadata
    is_internal         BOOLEAN NOT NULL DEFAULT FALSE,
    edited_at           TIMESTAMPTZ,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_comments_org ON audit_comments (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_audit ON audit_comments (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_target ON audit_comments (target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_author ON audit_comments (author_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_parent ON audit_comments (parent_comment_id) WHERE parent_comment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_comments_target_created ON audit_comments (target_type, target_id, created_at);

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_comments_updated_at ON audit_comments;
CREATE TRIGGER trg_audit_comments_updated_at
    BEFORE UPDATE ON audit_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_comments IS 'Threaded comments on audit engagements, requests, and findings (spec §6.4)';
COMMENT ON COLUMN audit_comments.target_type IS 'Which entity this comment is on: audit, request, or finding';
COMMENT ON COLUMN audit_comments.target_id IS 'UUID of the target entity (audit, audit_request, or audit_finding)';
COMMENT ON COLUMN audit_comments.is_internal IS 'If true, only visible to internal team (not auditor role users)';
COMMENT ON COLUMN audit_comments.parent_comment_id IS 'Self-referencing FK for threaded replies';
