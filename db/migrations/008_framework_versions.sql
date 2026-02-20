-- Migration: 008_framework_versions.sql
-- Description: Framework versions table (supports multiple active versions per framework)
-- Created: 2026-02-20
-- Sprint: 2 — Core Entities (Frameworks & Controls)

CREATE TABLE IF NOT EXISTS framework_versions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_id        UUID NOT NULL REFERENCES frameworks(id) ON DELETE CASCADE,
    version             VARCHAR(50) NOT NULL,
    display_name        VARCHAR(255) NOT NULL,
    status              framework_version_status NOT NULL DEFAULT 'draft',
    effective_date      DATE,
    sunset_date         DATE,
    changelog           TEXT,
    total_requirements  INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_framework_version UNIQUE (framework_id, version),
    CONSTRAINT chk_framework_version_dates CHECK (sunset_date IS NULL OR effective_date IS NULL OR sunset_date > effective_date),
    CONSTRAINT chk_total_requirements_positive CHECK (total_requirements >= 0)
);

COMMENT ON TABLE framework_versions IS 'Versioned releases of each framework (e.g. PCI DSS v3.2.1 and v4.0.1)';
COMMENT ON COLUMN framework_versions.total_requirements IS 'Denormalized count of assessable requirements for performance';
COMMENT ON COLUMN framework_versions.effective_date IS 'When this version became officially active';
COMMENT ON COLUMN framework_versions.sunset_date IS 'When this version is no longer supported';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_framework_versions_framework_id ON framework_versions (framework_id);
CREATE INDEX IF NOT EXISTS idx_framework_versions_status ON framework_versions (status);

-- Trigger
DROP TRIGGER IF EXISTS trg_framework_versions_updated_at ON framework_versions;
CREATE TRIGGER trg_framework_versions_updated_at
    BEFORE UPDATE ON framework_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
