# Sprint 8 — Database Schema: User Access Reviews

## Overview

Sprint 8 introduces User Access Reviews — the governance layer for identity and access management. This implements spec §6.2: automated access review campaigns on configurable schedules, identity provider integration stubs, side-by-side current-vs-expected access views, reviewer assignment, one-click approve/revoke decisions with audit trail, anomaly detection (orphaned accounts, excessive privileges, role drift), and certification reports for auditors.

**Key design decisions:**

- **Identity providers are lightweight stubs**: Sprint 9 builds the full Integration Engine. Here we create `identity_providers` as a minimal table for IdP connection metadata (Okta, Azure AD, Google Workspace, JumpCloud, OneLogin) with JSONB config. This is enough to track sync state and pull access data, while leaving the full connector framework to Sprint 9.
- **Access resources represent reviewable targets**: Applications, systems, and services pulled from identity providers or created manually. Each resource has a criticality tier and owner, driving review prioritization. Resources are org-scoped — the same SaaS app at different orgs is tracked independently.
- **Access entries are the source of truth for current access**: Each entry represents "user X has role/permission Y on resource Z." Entries are pulled from identity providers during sync or created manually. They carry both the *actual* access state and the *expected* access (from role definitions) — enabling side-by-side comparison per spec §6.2.
- **Campaigns batch reviews into governed cycles**: A campaign is a scheduled batch (e.g., "Q1 2026 Quarterly Access Review") with scope rules (which resources, departments, or criticality tiers to include), deadline, and escalation settings. Campaigns transition through a lifecycle: draft → active → in_review → completed → cancelled.
- **Reviews are individual decisions within a campaign**: When a campaign launches, the system creates one `access_review` per in-scope `access_entry`. Each review is assigned to a reviewer (by resource owner, department head, or explicit assignment). Reviewers make approve/revoke/flag decisions with mandatory justification. This implements the "one-click approve/revoke decisions with audit trail" from spec §6.2.
- **Anomaly detection is tag-based**: Rather than a separate findings table, access entries carry an `anomalies` JSONB array with detected issues (orphaned accounts, excessive privileges, role drift, stale access, no MFA, departed users). This keeps the data model simple while surfacing issues in the review UI. Anomalies are recalculated during sync and on-demand.
- **Certification reports leverage campaign completion data**: When a campaign completes, all decisions, justifications, and reviewer identities are preserved for auditor-facing certification reports — a key requirement for SOC 2 and ISO 27001 evidence.

---

## Entity Relationship Diagram

```
                    ┌──────────────────────────────────────────────────────────┐
                    │   ACCESS REVIEW DOMAIN (org-scoped)                       │
                    │                                                          │
 organizations ─┬──▶ identity_providers                                        │
                │   │   (IdP stubs: Okta, Azure AD, Google Workspace, etc.)    │
                │   │                                                          │
                │   │   1──∞ access_resources                                  │
                │   │         (apps/systems/services from IdPs or manual)      │
                │   │         ∞──1 users (owner_id)                            │
                │   │                                                          │
                │   │         1──∞ access_entries                              │
                │   │               (user X has role Y on resource Z)          │
                │   │               ∞──1 identity_providers (provider_id)      │
                │   │               ∞──1 users (internal_user_id, optional)    │
                │   │                                                          │
                ├──▶ access_review_campaigns                                   │
                │   │   (governed review cycles: Q1 review, annual review)     │
                │   │   ∞──1 users (created_by)                                │
                │   │                                                          │
                │   │   1──∞ access_reviews                                    │
                │   │         (individual approve/revoke decisions)             │
                │   │         ∞──1 access_entries (entry_id)                   │
                │   │         ∞──1 users (reviewer_id)                         │
                │   │         ∞──1 users (decided_by)                          │
                │   │                                                          │
                └──▶ audit_log (extended with Sprint 8 actions)                │
                    └──────────────────────────────────────────────────────────┘
```

**Cross-domain relationships:**
```
access_resources.owner_id ──▶ users (Sprint 1)
access_entries.internal_user_id ──▶ users (Sprint 1)
access_review_campaigns.created_by ──▶ users (Sprint 1)
access_reviews.reviewer_id ──▶ users (Sprint 1)
access_reviews.decided_by ──▶ users (Sprint 1)
evidence_links.access_review_campaign_id ──▶ access_review_campaigns (new FK, via evidence_link_target_type extension)
```

---

## New Enum Types

### `identity_provider_type`
Supported identity provider systems (stubs for Sprint 8, full connectors in Sprint 9).

```sql
CREATE TYPE identity_provider_type AS ENUM (
    'okta',               -- Okta Identity
    'azure_ad',           -- Azure AD / Entra ID
    'google_workspace',   -- Google Workspace
    'jumpcloud',          -- JumpCloud
    'onelogin',           -- OneLogin
    'custom'              -- Custom SCIM/API provider
);
```

### `identity_provider_status`
Connection lifecycle for identity providers.

```sql
CREATE TYPE identity_provider_status AS ENUM (
    'pending_setup',   -- Configuration started, not yet connected
    'connected',       -- Successfully connected, sync operational
    'syncing',         -- Currently pulling access data
    'error',           -- Connection failed, needs attention
    'disconnected'     -- Intentionally disconnected / disabled
);
```

### `resource_criticality`
Criticality tier for access resources (drives review prioritization).

```sql
CREATE TYPE resource_criticality AS ENUM (
    'critical',   -- Production databases, payment systems, admin consoles
    'high',       -- Source code repos, CI/CD, cloud consoles
    'medium',     -- Internal tools, collaboration platforms
    'low'         -- Read-only resources, public tools
);
```

### `resource_type`
Category of access resource.

```sql
CREATE TYPE resource_type AS ENUM (
    'application',     -- SaaS application (Slack, GitHub, etc.)
    'infrastructure',  -- Cloud accounts, servers, databases
    'directory',       -- Identity directories (AD groups, Okta groups)
    'custom'           -- Manually added resources
);
```

### `campaign_status`
Access review campaign lifecycle.

```sql
CREATE TYPE campaign_status AS ENUM (
    'draft',       -- Campaign being configured, not yet launched
    'active',      -- Campaign launched, reviews being assigned
    'in_review',   -- All reviews assigned, decisions being collected
    'completed',   -- All reviews decided, campaign closed
    'cancelled'    -- Campaign cancelled before completion
);
```

### `campaign_cadence`
Configurable review frequency per spec §6.2.

```sql
CREATE TYPE campaign_cadence AS ENUM (
    'monthly',
    'quarterly',
    'semi_annual',
    'annual',
    'custom'
);
```

### `review_decision`
Reviewer's decision on an access entry.

```sql
CREATE TYPE review_decision AS ENUM (
    'pending',     -- Not yet reviewed
    'approved',    -- Access confirmed appropriate
    'revoked',     -- Access should be removed
    'flagged',     -- Requires further investigation
    'delegated',   -- Delegated to another reviewer
    'expired'      -- Review deadline passed without decision
);
```

### `access_entry_status`
Status of an access record pulled from an identity provider.

```sql
CREATE TYPE access_entry_status AS ENUM (
    'active',              -- Currently active access
    'inactive',            -- Access exists but not used recently
    'orphaned',            -- No matching employee record found
    'suspended',           -- Access temporarily suspended
    'pending_revocation'   -- Revocation approved, awaiting execution
);
```

### `access_anomaly_type`
Types of access anomalies detected during sync/analysis (spec §6.2: "Auto-detect orphaned accounts, excessive privileges, and role drift").

```sql
CREATE TYPE access_anomaly_type AS ENUM (
    'orphaned_account',       -- No matching employee / departed user
    'excessive_privileges',   -- More access than role requires
    'role_drift',             -- Actual access diverges from expected
    'stale_access',           -- Access not used within threshold (90+ days)
    'no_mfa',                 -- Account lacks MFA on critical resource
    'departed_user'           -- User marked as departed in HRIS
);
```

---

## Extended Enums

### `audit_action` Extensions (18 new values)

```sql
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.connected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.disconnected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.sync_started';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'identity_provider.sync_completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_resource.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_resource.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_resource.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.launched';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'campaign.cancelled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.assigned';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.decided';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.delegated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.escalated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'access_review.bulk_decided';
```

### `evidence_link_target_type` Extension

```sql
ALTER TYPE evidence_link_target_type ADD VALUE IF NOT EXISTS 'access_review_campaign';
```

---

## Tables

### Table 1: `identity_providers`

Lightweight identity provider connection stubs. Stores configuration metadata and sync state. The full connector implementation (polling, transformation, error handling) comes in Sprint 9's Integration Engine — this table provides the data model foundation.

```sql
CREATE TABLE IF NOT EXISTS identity_providers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Provider identity
    name                VARCHAR(255) NOT NULL,
    provider_type       identity_provider_type NOT NULL,
    status              identity_provider_status NOT NULL DEFAULT 'pending_setup',

    -- Connection configuration (JSONB for flexibility across providers)
    -- Okta: { "domain": "acme.okta.com", "api_token": "encrypted:..." }
    -- Azure AD: { "tenant_id": "...", "client_id": "...", "client_secret": "encrypted:..." }
    -- Google: { "domain": "acme.com", "service_account_key": "encrypted:..." }
    config              JSONB NOT NULL DEFAULT '{}',

    -- Sync state
    last_sync_at        TIMESTAMPTZ,
    last_sync_status    VARCHAR(50),                   -- 'success', 'partial', 'failed'
    last_sync_error     TEXT,                          -- Error message if sync failed
    last_sync_stats     JSONB DEFAULT '{}',            -- { "users_synced": 150, "resources_synced": 23, "duration_ms": 4500 }
    sync_interval_mins  INTEGER NOT NULL DEFAULT 360,  -- Default: sync every 6 hours

    -- Metadata
    description         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One active provider per type per org (can reconnect different instance)
    CONSTRAINT uq_identity_providers_org_name UNIQUE (org_id, name),
    CONSTRAINT chk_identity_providers_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_identity_providers_sync_interval CHECK (sync_interval_mins >= 15 AND sync_interval_mins <= 10080)
);

COMMENT ON TABLE identity_providers IS 'Identity provider connection stubs. Full connector logic in Sprint 9 Integration Engine.';
COMMENT ON COLUMN identity_providers.config IS 'Provider-specific config (JSONB). Secrets should be encrypted at application layer before storage.';
COMMENT ON COLUMN identity_providers.last_sync_stats IS 'Stats from last sync run: users/resources synced, duration, errors.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_identity_providers_org_id ON identity_providers (org_id);
CREATE INDEX IF NOT EXISTS idx_identity_providers_org_status ON identity_providers (org_id, status);
CREATE INDEX IF NOT EXISTS idx_identity_providers_org_type ON identity_providers (org_id, provider_type);

-- Trigger
DROP TRIGGER IF EXISTS trg_identity_providers_updated_at ON identity_providers;
CREATE TRIGGER trg_identity_providers_updated_at
    BEFORE UPDATE ON identity_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

### Table 2: `access_resources`

Applications, systems, and services that are subject to access review. Pulled from identity providers during sync or created manually. Resources have criticality tiers that drive review prioritization and campaign scoping.

```sql
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
```

---

### Table 3: `access_entries`

Individual access records: "user X has role/permission Y on resource Z." This is the core data that reviewers see during campaigns. Entries carry both actual and expected access for side-by-side comparison (spec §6.2). Anomalies are detected during sync and stored as JSONB for flexible querying.

```sql
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

    -- Unique: one entry per user per role per resource (a user can have multiple roles)
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
```

---

### Table 4: `access_review_campaigns`

Governed review cycles. A campaign is a scheduled batch of access reviews (e.g., "Q1 2026 Quarterly Access Review") with scope rules, deadlines, and escalation settings. Campaigns drive the access review workflow from spec §6.2.

```sql
CREATE TABLE IF NOT EXISTS access_review_campaigns (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Campaign identity
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    status              campaign_status NOT NULL DEFAULT 'draft',
    cadence             campaign_cadence NOT NULL DEFAULT 'quarterly',

    -- Scope configuration (which access entries to include)
    scope               JSONB NOT NULL DEFAULT '{}',
    -- Example scope:
    -- {
    --   "resource_ids": ["uuid1", "uuid2"],          -- Specific resources (empty = all)
    --   "resource_criticalities": ["critical", "high"], -- By criticality tier
    --   "resource_types": ["application"],            -- By resource type
    --   "departments": ["Engineering", "Finance"],    -- By user department
    --   "include_privileged_only": false,             -- Only review admin/privileged access
    --   "include_service_accounts": true,             -- Include service accounts
    --   "stale_days_threshold": 90                    -- Include entries unused for N+ days
    -- }

    -- Reviewer assignment strategy
    reviewer_strategy   VARCHAR(50) NOT NULL DEFAULT 'resource_owner',
    -- 'resource_owner' — Assign reviews to the resource's owner_id
    -- 'department_head' — Assign to the user's manager/department head
    -- 'explicit' — Manually assign reviewers during campaign creation
    -- 'mixed' — Resource owners review, with fallback to explicit assignment

    default_reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,  -- Fallback reviewer if strategy can't assign

    -- Timeline
    started_at          TIMESTAMPTZ,
    deadline            TIMESTAMPTZ NOT NULL,          -- All reviews must be completed by this date
    completed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,

    -- Escalation settings
    escalation_config   JSONB NOT NULL DEFAULT '{}',
    -- {
    --   "reminder_days_before_deadline": [7, 3, 1],   -- Send reminders N days before deadline
    --   "escalate_after_days": 5,                      -- Escalate to manager after N days unreviewed
    --   "escalation_recipient_id": "uuid",             -- Who gets escalation notifications
    --   "auto_expire_days": null                        -- Auto-expire pending reviews after N days (null = never)
    -- }

    -- Denormalized counts
    total_reviews       INTEGER NOT NULL DEFAULT 0,
    completed_reviews   INTEGER NOT NULL DEFAULT 0,
    approved_count      INTEGER NOT NULL DEFAULT 0,
    revoked_count       INTEGER NOT NULL DEFAULT 0,
    flagged_count       INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    created_by          UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    tags                TEXT[] NOT NULL DEFAULT '{}',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_campaigns_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_campaigns_total_reviews CHECK (total_reviews >= 0),
    CONSTRAINT chk_campaigns_completed_reviews CHECK (completed_reviews >= 0 AND completed_reviews <= total_reviews),
    CONSTRAINT chk_campaigns_approved CHECK (approved_count >= 0),
    CONSTRAINT chk_campaigns_revoked CHECK (revoked_count >= 0),
    CONSTRAINT chk_campaigns_flagged CHECK (flagged_count >= 0),
    CONSTRAINT chk_campaigns_deadline_future CHECK (
        status = 'draft' OR deadline IS NOT NULL
    ),
    CONSTRAINT chk_campaigns_completed_after_started CHECK (
        completed_at IS NULL OR (started_at IS NOT NULL AND completed_at >= started_at)
    )
);

COMMENT ON TABLE access_review_campaigns IS 'Governed access review cycles. Scope rules determine which access entries are reviewed.';
COMMENT ON COLUMN access_review_campaigns.scope IS 'JSONB scope config: resource_ids, criticalities, types, departments, privileged_only, service_accounts, stale_threshold.';
COMMENT ON COLUMN access_review_campaigns.reviewer_strategy IS 'How reviewers are assigned: resource_owner, department_head, explicit, or mixed.';
COMMENT ON COLUMN access_review_campaigns.escalation_config IS 'Escalation settings: reminder schedule, escalation delay, recipient, auto-expire.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_campaigns_org_id ON access_review_campaigns (org_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_org_status ON access_review_campaigns (org_id, status);
CREATE INDEX IF NOT EXISTS idx_campaigns_org_cadence ON access_review_campaigns (org_id, cadence);
CREATE INDEX IF NOT EXISTS idx_campaigns_deadline ON access_review_campaigns (deadline) WHERE status IN ('active', 'in_review');
CREATE INDEX IF NOT EXISTS idx_campaigns_created_by ON access_review_campaigns (created_by);
CREATE INDEX IF NOT EXISTS idx_campaigns_tags ON access_review_campaigns USING GIN (tags);

-- Trigger
DROP TRIGGER IF EXISTS trg_campaigns_updated_at ON access_review_campaigns;
CREATE TRIGGER trg_campaigns_updated_at
    BEFORE UPDATE ON access_review_campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

### Table 5: `access_reviews`

Individual access review decisions within a campaign. One review is created per in-scope `access_entry` when a campaign launches. Reviewers make approve/revoke/flag decisions with mandatory justification.

```sql
CREATE TABLE IF NOT EXISTS access_reviews (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    campaign_id         UUID NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    entry_id            UUID NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,

    -- Assignment
    reviewer_id         UUID REFERENCES users(id) ON DELETE SET NULL,  -- Who is assigned to review
    assigned_at         TIMESTAMPTZ,

    -- Decision
    decision            review_decision NOT NULL DEFAULT 'pending',
    justification       TEXT,                          -- Required for revoke/flag decisions
    decided_by          UUID REFERENCES users(id) ON DELETE SET NULL,  -- Who actually made the decision (may differ from reviewer if delegated)
    decided_at          TIMESTAMPTZ,

    -- Delegation
    delegated_to_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    delegated_at        TIMESTAMPTZ,
    delegation_reason   TEXT,

    -- Escalation
    is_escalated        BOOLEAN NOT NULL DEFAULT FALSE,
    escalated_at        TIMESTAMPTZ,
    escalated_to_id     UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Context snapshot (frozen at review time for audit trail)
    access_snapshot     JSONB NOT NULL DEFAULT '{}',
    -- {
    --   "resource_name": "GitHub",
    --   "resource_criticality": "high",
    --   "user_email": "alice@acme.com",
    --   "user_department": "Engineering",
    --   "role_name": "Admin",
    --   "access_level": "admin",
    --   "is_privileged": true,
    --   "expected_role": "Write",
    --   "has_role_drift": true,
    --   "granted_at": "2025-06-15T...",
    --   "last_used_at": "2026-02-10T...",
    --   "anomalies": [{"type": "excessive_privileges", "details": "..."}]
    -- }

    -- Revocation tracking
    revocation_executed BOOLEAN NOT NULL DEFAULT FALSE,
    revocation_executed_at TIMESTAMPTZ,
    revocation_notes    TEXT,

    -- Metadata
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One review per entry per campaign
    CONSTRAINT uq_access_reviews_campaign_entry UNIQUE (campaign_id, entry_id),
    CONSTRAINT chk_reviews_justification_on_revoke CHECK (
        decision NOT IN ('revoked', 'flagged') OR justification IS NOT NULL
    ),
    CONSTRAINT chk_reviews_decided_at CHECK (
        decision = 'pending' OR decided_at IS NOT NULL
    ),
    CONSTRAINT chk_reviews_decided_by CHECK (
        decision = 'pending' OR decided_by IS NOT NULL
    ),
    CONSTRAINT chk_reviews_revocation CHECK (
        revocation_executed = FALSE OR (decision = 'revoked' AND revocation_executed_at IS NOT NULL)
    )
);

COMMENT ON TABLE access_reviews IS 'Individual access review decisions. One review per access entry per campaign. Decisions are immutable once made.';
COMMENT ON COLUMN access_reviews.access_snapshot IS 'Frozen snapshot of access state at review time. Ensures audit trail survives access entry changes.';
COMMENT ON COLUMN access_reviews.decided_by IS 'May differ from reviewer_id when review was delegated.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_access_reviews_org_id ON access_reviews (org_id);
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign ON access_reviews (campaign_id);
CREATE INDEX IF NOT EXISTS idx_access_reviews_entry ON access_reviews (entry_id);
CREATE INDEX IF NOT EXISTS idx_access_reviews_reviewer ON access_reviews (reviewer_id) WHERE reviewer_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign_decision ON access_reviews (campaign_id, decision);
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign_pending ON access_reviews (campaign_id) WHERE decision = 'pending';
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign_revoked ON access_reviews (campaign_id) WHERE decision = 'revoked';
CREATE INDEX IF NOT EXISTS idx_access_reviews_escalated ON access_reviews (campaign_id) WHERE is_escalated = TRUE;
CREATE INDEX IF NOT EXISTS idx_access_reviews_revocation_pending ON access_reviews (org_id) WHERE decision = 'revoked' AND revocation_executed = FALSE;
CREATE INDEX IF NOT EXISTS idx_access_reviews_decided_by ON access_reviews (decided_by) WHERE decided_by IS NOT NULL;

-- Trigger
DROP TRIGGER IF EXISTS trg_access_reviews_updated_at ON access_reviews;
CREATE TRIGGER trg_access_reviews_updated_at
    BEFORE UPDATE ON access_reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

## Cross-Reference FK Extension

### `evidence_links` Extension

Add support for linking evidence to access review campaigns (for auditor certification reports).

```sql
-- Add FK column for access_review_campaigns
ALTER TABLE evidence_links
    ADD COLUMN IF NOT EXISTS access_review_campaign_id UUID REFERENCES access_review_campaigns(id) ON DELETE SET NULL;

-- Update the target type CHECK constraint to include the new type
-- (evidence_link_target_type already extended above with 'access_review_campaign')

-- Index for the new FK
CREATE INDEX IF NOT EXISTS idx_evidence_links_access_review_campaign
    ON evidence_links (access_review_campaign_id)
    WHERE access_review_campaign_id IS NOT NULL;
```

---

## Query Patterns

### Side-by-Side Access Comparison (spec §6.2)

```sql
-- Current vs. expected access for a user across all resources
SELECT
    ar.name AS resource_name,
    ar.criticality,
    ae.role_name AS actual_role,
    ae.access_level AS actual_level,
    ae.expected_role,
    ae.expected_access_level,
    ae.has_role_drift,
    ae.last_used_at,
    ae.anomalies
FROM access_entries ae
JOIN access_resources ar ON ar.id = ae.resource_id
WHERE ae.org_id = $1
  AND ae.user_email = $2
  AND ae.status = 'active'
ORDER BY ar.criticality, ar.name;
```

### Anomaly Detection Summary

```sql
-- Count anomalies by type across all active entries
SELECT
    anomaly->>'type' AS anomaly_type,
    COUNT(*) AS count
FROM access_entries ae,
     jsonb_array_elements(ae.anomalies) AS anomaly
WHERE ae.org_id = $1
  AND ae.status != 'pending_revocation'
GROUP BY anomaly->>'type'
ORDER BY count DESC;
```

### Campaign Progress

```sql
-- Campaign completion dashboard
SELECT
    c.id,
    c.name,
    c.status,
    c.deadline,
    c.total_reviews,
    c.completed_reviews,
    CASE WHEN c.total_reviews > 0
         THEN ROUND(100.0 * c.completed_reviews / c.total_reviews, 1)
         ELSE 0 END AS completion_pct,
    c.approved_count,
    c.revoked_count,
    c.flagged_count,
    COUNT(CASE WHEN r.decision = 'pending' AND r.is_escalated = TRUE THEN 1 END) AS escalated_count,
    c.deadline - NOW() AS time_remaining
FROM access_review_campaigns c
LEFT JOIN access_reviews r ON r.campaign_id = c.id
WHERE c.org_id = $1
  AND c.status IN ('active', 'in_review')
GROUP BY c.id
ORDER BY c.deadline;
```

### Reviewer Workload

```sql
-- Pending reviews by reviewer
SELECT
    u.id AS reviewer_id,
    u.first_name || ' ' || u.last_name AS reviewer_name,
    COUNT(*) AS pending_reviews,
    COUNT(CASE WHEN snap->>'is_privileged' = 'true' THEN 1 END) AS privileged_reviews,
    MIN(c.deadline) AS earliest_deadline
FROM access_reviews r
JOIN users u ON u.id = r.reviewer_id
JOIN access_review_campaigns c ON c.id = r.campaign_id
CROSS JOIN LATERAL (SELECT r.access_snapshot AS snap) s
WHERE r.org_id = $1
  AND r.decision = 'pending'
  AND c.status IN ('active', 'in_review')
GROUP BY u.id, u.first_name, u.last_name
ORDER BY pending_reviews DESC;
```

### Certification Report Data

```sql
-- All decisions for a completed campaign (auditor-facing)
SELECT
    r.access_snapshot->>'resource_name' AS resource,
    r.access_snapshot->>'user_email' AS user_email,
    r.access_snapshot->>'user_department' AS department,
    r.access_snapshot->>'role_name' AS role,
    r.access_snapshot->>'is_privileged' AS privileged,
    r.decision,
    r.justification,
    du.first_name || ' ' || du.last_name AS decided_by_name,
    r.decided_at,
    r.revocation_executed,
    r.revocation_executed_at
FROM access_reviews r
LEFT JOIN users du ON du.id = r.decided_by
WHERE r.campaign_id = $1
  AND r.org_id = $2
ORDER BY
    r.access_snapshot->>'resource_name',
    r.access_snapshot->>'user_email';
```

---

## Migration Files

| File | Contents |
|------|----------|
| `053_sprint8_enums.sql` | 9 new enums + 18 audit_action extensions + evidence_link_target_type extension |
| `054_identity_providers.sql` | identity_providers table with indexes, trigger |
| `055_access_resources.sql` | access_resources table with indexes, trigger, GIN index on tags |
| `056_access_entries.sql` | access_entries table with indexes, trigger, GIN index on anomalies |
| `057_access_review_campaigns.sql` | access_review_campaigns table with indexes, trigger |
| `058_access_reviews.sql` | access_reviews table with indexes, trigger, unique constraint |
| `059_sprint8_fk_cross_refs.sql` | evidence_links extension (access_review_campaign_id FK) |
| `060_sprint8_seed_templates.sql` | IdP connection templates, sample resources, demo anomaly patterns |
| `061_sprint8_seed_demo.sql` | Demo data: 2 IdPs, 6 resources, 25 access entries, 1 campaign, 15 reviews, anomalies |

---

## Seed Data

### Identity Provider Templates

Pre-configured templates for quick IdP setup:

```sql
-- Demo: Okta connection (connected)
INSERT INTO identity_providers (org_id, name, provider_type, status, config, last_sync_at, last_sync_status, last_sync_stats, sync_interval_mins, description)
VALUES (
    (SELECT id FROM organizations LIMIT 1),
    'Okta SSO',
    'okta',
    'connected',
    '{"domain": "acme.okta.com", "api_token": "encrypted:demo_token_okta"}',
    NOW() - INTERVAL '2 hours',
    'success',
    '{"users_synced": 156, "resources_synced": 12, "duration_ms": 3200}',
    360,
    'Primary identity provider — all employees'
);

-- Demo: Azure AD (connected)
INSERT INTO identity_providers (org_id, name, provider_type, status, config, last_sync_at, last_sync_status, last_sync_stats, sync_interval_mins, description)
VALUES (
    (SELECT id FROM organizations LIMIT 1),
    'Azure AD',
    'azure_ad',
    'connected',
    '{"tenant_id": "demo-tenant-id", "client_id": "demo-client-id", "client_secret": "encrypted:demo_secret_azure"}',
    NOW() - INTERVAL '4 hours',
    'success',
    '{"users_synced": 89, "resources_synced": 8, "duration_ms": 5100}',
    360,
    'Microsoft 365 and Azure cloud services'
);
```

### Demo Access Resources

```sql
-- 6 resources across criticality tiers
-- Critical: Production Database, AWS Console
-- High: GitHub Organization, Stripe Dashboard
-- Medium: Slack Workspace, Jira
```

### Demo Access Entries (25 entries)

```sql
-- Spread across 6 resources with realistic patterns:
-- 3 admin/privileged accounts on critical resources
-- 5 entries with anomalies (2 stale_access, 1 orphaned_account, 1 excessive_privileges, 1 role_drift)
-- 4 service accounts
-- Mix of departments: Engineering (10), Finance (5), Marketing (4), HR (3), DevOps (3)
```

### Demo Campaign

```sql
-- 1 active campaign: "Q1 2026 Quarterly Access Review"
-- Scope: critical + high resources, all departments
-- 15 reviews: 8 pending, 3 approved, 2 revoked (with justification), 1 flagged, 1 delegated
-- Deadline: 2 weeks from now
-- 2 escalated reviews (overdue)
```

---

## Design Notes

### Why Identity Providers are Stubs (Not Full Connectors)

Sprint 9 builds the Integration Engine with proper connector architecture (base connector class, health monitoring, polling, retry logic, transformation pipeline). Sprint 8 creates the data model and minimal CRUD — enough to manually sync or use basic API token authentication. The `identity_providers` table is designed to be subsumed by Sprint 9's `integrations` table (same fields, plus connector-specific extensions).

### Why Access Snapshots in Reviews

Access entries are mutable — they change every sync cycle. But audit evidence requires an immutable record of what was reviewed. The `access_snapshot` JSONB field in `access_reviews` freezes the state at review time, ensuring the certification report accurately reflects what the reviewer saw, even if the access entry has since changed.

### Why JSONB for Anomalies

Rather than a junction table with anomaly types, JSONB `anomalies` on access_entries allows:
1. Multiple anomalies per entry (common — stale access + no MFA)
2. Rich detail strings per anomaly
3. Efficient GIN index querying
4. Easy extension without schema changes
5. Direct inclusion in access_snapshot for reviews

### Denormalization Strategy

- `access_resources.total_users` / `total_roles` — updated during sync and entry CRUD
- `access_review_campaigns.total_reviews` / `completed_reviews` / `approved_count` / `revoked_count` / `flagged_count` — updated on review decisions via helper functions (same pattern as Sprint 7's audit request counts)
