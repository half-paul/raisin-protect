-- Sprint 9 -- Table 1: integration_definitions (system-level catalog)
CREATE TABLE IF NOT EXISTS integration_definitions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255) NOT NULL,
    slug                VARCHAR(100) NOT NULL UNIQUE,
    provider            VARCHAR(100) NOT NULL UNIQUE,
    category            integration_category NOT NULL,
    description         TEXT,
    short_description   VARCHAR(500),
    icon_url            VARCHAR(500),
    documentation_url   VARCHAR(500),
    website_url         VARCHAR(500),
    auth_type           integration_auth_type NOT NULL DEFAULT 'api_key',
    config_schema       JSONB NOT NULL DEFAULT '{}',
    capabilities        TEXT[] NOT NULL DEFAULT '{}',
    is_active           BOOLEAN NOT NULL DEFAULT true,
    is_beta             BOOLEAN NOT NULL DEFAULT false,
    version             VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    tags                TEXT[] NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE integration_definitions IS 'System-level catalog of available integration types';

CREATE INDEX IF NOT EXISTS idx_integration_definitions_category ON integration_definitions (category);
CREATE INDEX IF NOT EXISTS idx_integration_definitions_capabilities ON integration_definitions USING GIN (capabilities);
CREATE INDEX IF NOT EXISTS idx_integration_definitions_tags ON integration_definitions USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_integration_definitions_is_active ON integration_definitions (is_active);

DROP TRIGGER IF EXISTS trg_integration_definitions_updated_at ON integration_definitions;
CREATE TRIGGER trg_integration_definitions_updated_at
    BEFORE UPDATE ON integration_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
