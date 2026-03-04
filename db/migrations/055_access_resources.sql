-- Sprint 8 — Table 2: access_resources

CREATE TABLE IF NOT EXISTS access_resources (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Source tracking
    identity_provider_id    UUID REFERENCES identity_providers(id) ON DELETE SET NULL,
    external_id             VARCHAR(500),              -- ID in the source IdP (e.g., Okta app ID)

    -- Resource identity
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    resource_type           resource_type NOT NULL DEFAULT 'application',
    criticality             resource_criticality NOT NULL DEFAULT 'medium',

    -- Classification
    department              VARCHAR(255),              -- Owning department (e.g., "Engineering", "Finance")
    category                VARCHAR(255),              -- Custom category (e.g., "Developer Tools", "HR Systems")
    tags                    TEXT[] NOT NULL DEFAULT '{}',

    -- Ownership
    owner_id                UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Access metadata
    total_users             INTEGER NOT NULL DEFAULT 0,    -- Denormalized: count of active access entries
    total_roles             INTEGER NOT NULL DEFAULT 0,    -- Denormalized: distinct roles/permissions
    last_sync_at            TIMESTAMPTZ,                   -- When access data was last pulled
    url                     VARCHAR(2048),                 -- Resource URL (e.g., https://app.slack.com)

    -- Review tracking
    last_reviewed_at        TIMESTAMPTZ,                   -- When this resource was last included in a campaign
    review_cadence          campaign_cadence,               -- Recommended review frequency for this resource

    -- Metadata
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique external ID per provider per org
    CONSTRAINT uq_access_resources_org_provider_external UNIQUE (org_id, identity_provider_id, external_id),
    CONSTRAINT chk_access_resources_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_access_resources_total_users CHECK (total_users >= 0),
    CONSTRAINT chk_access_resources_total_roles CHECK (total_roles >= 0)
);

COMMENT ON TABLE access_resources IS 'Applications/systems/services subject to access review. Pulled from IdPs or created manually.';
COMMENT ON COLUMN access_resources.external_id IS 'Resource identifier in the source identity provider (e.g., Okta app_id, Azure AD app registration ID).';
COMMENT ON COLUMN access_resources.total_users IS 'Denormalized count of active access entries — updated during sync and on entry changes.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_access_resources_org_id ON access_resources (org_id);
CREATE INDEX IF NOT EXISTS idx_access_resources_org_criticality ON access_resources (org_id, criticality);
CREATE INDEX IF NOT EXISTS idx_access_resources_org_type ON access_resources (org_id, resource_type);
CREATE INDEX IF NOT EXISTS idx_access_resources_org_active ON access_resources (org_id, is_active);
CREATE INDEX IF NOT EXISTS idx_access_resources_provider ON access_resources (identity_provider_id) WHERE identity_provider_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_access_resources_owner ON access_resources (owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_access_resources_department ON access_resources (org_id, department) WHERE department IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_access_resources_tags ON access_resources USING GIN (tags);

-- Trigger
DROP TRIGGER IF EXISTS trg_access_resources_updated_at ON access_resources;
CREATE TRIGGER trg_access_resources_updated_at
    BEFORE UPDATE ON access_resources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
