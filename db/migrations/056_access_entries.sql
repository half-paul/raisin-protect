-- Sprint 8 — Table 3: access_entries

CREATE TABLE IF NOT EXISTS access_entries (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Source
    identity_provider_id    UUID REFERENCES identity_providers(id) ON DELETE SET NULL,
    resource_id             UUID NOT NULL REFERENCES access_resources(id) ON DELETE CASCADE,
    external_user_id        VARCHAR(500),              -- User ID in the source IdP

    -- User identity (from IdP)
    user_email              VARCHAR(255) NOT NULL,
    user_display_name       VARCHAR(255) NOT NULL,
    user_department         VARCHAR(255),
    user_title              VARCHAR(255),
    user_manager_email      VARCHAR(255),

    -- Internal user mapping (if email matches a platform user)
    internal_user_id        UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Access details
    role_name               VARCHAR(255) NOT NULL,     -- Role/permission name (e.g., "Admin", "Viewer", "Write")
    access_level            VARCHAR(100),              -- Normalized level: "admin", "write", "read", "custom"
    permissions             JSONB DEFAULT '{}',        -- Detailed permission set (provider-specific)
    is_privileged           BOOLEAN NOT NULL DEFAULT FALSE,  -- Flagged as elevated/admin access

    -- Expected access (from role definition or baseline)
    expected_role           VARCHAR(255),              -- What role this user SHOULD have based on their job title/department
    expected_access_level   VARCHAR(100),              -- Expected normalized level
    has_role_drift          BOOLEAN NOT NULL DEFAULT FALSE,  -- actual ≠ expected

    -- Temporal data
    granted_at              TIMESTAMPTZ,               -- When access was granted
    last_used_at            TIMESTAMPTZ,               -- When access was last used (from IdP logs)
    last_login_at           TIMESTAMPTZ,               -- When user last logged into this resource

    -- Status
    status                  access_entry_status NOT NULL DEFAULT 'active',
    mfa_enabled             BOOLEAN,                   -- MFA status for this resource (null = unknown)

    -- Anomaly detection results
    anomalies               JSONB NOT NULL DEFAULT '[]',   -- Array of { "type": "stale_access", "detected_at": "...", "details": "..." }

    -- Metadata
    last_sync_at            TIMESTAMPTZ,               -- When this entry was last synced from IdP
    is_service_account      BOOLEAN NOT NULL DEFAULT FALSE,
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique: one entry per user per role per resource
    CONSTRAINT uq_access_entries_resource_user_role UNIQUE (org_id, resource_id, user_email, role_name),
    CONSTRAINT chk_access_entries_email_format CHECK (user_email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT chk_access_entries_name_not_empty CHECK (length(trim(user_display_name)) > 0),
    CONSTRAINT chk_access_entries_role_not_empty CHECK (length(trim(role_name)) > 0)
);

COMMENT ON TABLE access_entries IS 'Individual access records: user X has role Y on resource Z. Source of truth for access reviews.';
COMMENT ON COLUMN access_entries.expected_role IS 'Expected role based on user job title/department. Compared against actual role_name for drift detection.';
COMMENT ON COLUMN access_entries.anomalies IS 'Detected anomalies as JSONB array: [{"type":"stale_access","detected_at":"2026-02-01","details":"No login in 120 days"}]';
COMMENT ON COLUMN access_entries.internal_user_id IS 'References platform users table when access entry email matches a known user. Enables linking reviews to internal identities.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_access_entries_org_id ON access_entries (org_id);
CREATE INDEX IF NOT EXISTS idx_access_entries_resource ON access_entries (resource_id);
CREATE INDEX IF NOT EXISTS idx_access_entries_org_status ON access_entries (org_id, status);
CREATE INDEX IF NOT EXISTS idx_access_entries_org_email ON access_entries (org_id, user_email);
CREATE INDEX IF NOT EXISTS idx_access_entries_provider ON access_entries (identity_provider_id) WHERE identity_provider_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_access_entries_internal_user ON access_entries (internal_user_id) WHERE internal_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_access_entries_org_privileged ON access_entries (org_id, is_privileged) WHERE is_privileged = TRUE;
CREATE INDEX IF NOT EXISTS idx_access_entries_org_drift ON access_entries (org_id, has_role_drift) WHERE has_role_drift = TRUE;
CREATE INDEX IF NOT EXISTS idx_access_entries_org_orphaned ON access_entries (org_id, status) WHERE status = 'orphaned';
CREATE INDEX IF NOT EXISTS idx_access_entries_last_used ON access_entries (org_id, last_used_at);
CREATE INDEX IF NOT EXISTS idx_access_entries_anomalies ON access_entries USING GIN (anomalies jsonb_path_ops);
CREATE INDEX IF NOT EXISTS idx_access_entries_department ON access_entries (org_id, user_department) WHERE user_department IS NOT NULL;

-- Trigger
DROP TRIGGER IF EXISTS trg_access_entries_updated_at ON access_entries;
CREATE TRIGGER trg_access_entries_updated_at
    BEFORE UPDATE ON access_entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
