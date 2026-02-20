-- Migration: 009_requirements.sql
-- Description: Requirements table (hierarchical, self-referential via parent_id)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

CREATE TABLE IF NOT EXISTS requirements (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_version_id    UUID NOT NULL REFERENCES framework_versions(id) ON DELETE CASCADE,
    parent_id               UUID REFERENCES requirements(id) ON DELETE CASCADE,
    identifier              VARCHAR(50) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    guidance                TEXT,
    section_order           INT NOT NULL DEFAULT 0,
    depth                   INT NOT NULL DEFAULT 0,
    is_assessable           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_requirement_identifier UNIQUE (framework_version_id, identifier),
    CONSTRAINT chk_requirement_depth CHECK (depth >= 0),
    CONSTRAINT chk_requirement_section_order CHECK (section_order >= 0)
);

COMMENT ON TABLE requirements IS 'Individual requirements within a framework version (hierarchical tree)';
COMMENT ON COLUMN requirements.identifier IS 'Requirement number (e.g. 6.4.3, CC6.1, A.8.1)';
COMMENT ON COLUMN requirements.is_assessable IS 'FALSE for section headers, TRUE for testable leaf requirements';
COMMENT ON COLUMN requirements.guidance IS 'Implementation guidance to help users understand the requirement';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_requirements_framework_version ON requirements (framework_version_id);
CREATE INDEX IF NOT EXISTS idx_requirements_parent ON requirements (parent_id);
CREATE INDEX IF NOT EXISTS idx_requirements_assessable ON requirements (framework_version_id, is_assessable)
    WHERE is_assessable = TRUE;
CREATE INDEX IF NOT EXISTS idx_requirements_depth ON requirements (framework_version_id, depth, section_order);

-- Trigger
DROP TRIGGER IF EXISTS trg_requirements_updated_at ON requirements;
CREATE TRIGGER trg_requirements_updated_at
    BEFORE UPDATE ON requirements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
