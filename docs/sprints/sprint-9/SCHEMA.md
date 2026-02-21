# Sprint 9 — Database Schema: Integration Engine (Foundation)

## Overview

Sprint 9 introduces the Integration Engine — the extensible connector framework that powers automated evidence collection, identity sync, and third-party monitoring across the platform. This implements spec §8 (Integration Engine): a pluggable integration framework with health monitoring, 5 starter integrations (AWS Config, GitHub, Okta, Slack, custom webhook), connection management, sync execution tracking, webhook ingestion, and a connection health dashboard.

**Key design decisions:**

- **System-level catalog + org-level connections**: Like frameworks (Sprint 2), integration definitions are system-level (shared across all tenants). The `integration_definitions` table is the catalog of available integrations (AWS Config, GitHub, Okta, Slack, custom webhook, and future 300+ connectors). Per-org connections (`integration_connections`) store credentials, sync schedules, and health state. This separation means adding new integration types requires no migration — just a seed INSERT.
- **Unifies with Sprint 8 identity providers**: Sprint 8 created `identity_providers` as lightweight IdP stubs. Sprint 9 adds an optional `integration_connection_id` FK to `identity_providers`, linking IdP configs to the full integration framework when a connection is established. This allows Sprint 8's access review flow to continue working standalone while adding the richer lifecycle management (health checks, run history, logs) from the integration engine.
- **Config schemas validate connection setup**: Each `integration_definition` carries a `config_schema` (JSON Schema) that defines what fields are required for connection setup (API keys, domains, OAuth credentials, etc.). The frontend uses this to render dynamic configuration forms. The backend validates `integration_connections.config` against the schema before saving.
- **Run-based execution model**: Every sync operation creates an `integration_run` record that tracks execution lifecycle (pending → running → completed/failed/partial). Runs carry JSONB `stats` for items synced/created/updated/failed. This provides full auditability of what data was pulled from each integration and when.
- **Structured log entries per run**: `integration_logs` captures detailed per-run logs with severity levels (debug/info/warn/error). Logs are queryable by level, message content, and time range. This replaces ad-hoc logging and provides integration troubleshooting directly in the dashboard.
- **Inbound webhooks for real-time events**: `integration_webhooks` supports inbound webhook endpoints for integrations that push data (GitHub events, Slack events, AWS SNS notifications). Each webhook gets a unique URL with HMAC secret verification. This enables real-time evidence collection without polling.
- **Capabilities model for connectors**: Each integration definition declares its `capabilities` (e.g., `sync_users`, `sync_resources`, `collect_evidence`, `send_notifications`, `receive_webhooks`). This drives the UI — showing only relevant features for each integration type — and enables capability-based queries (e.g., "list all integrations that can sync users").
- **5 starter integrations**: AWS Config (cloud resource compliance), GitHub (code repository monitoring), Okta (identity/access management — extends Sprint 8's IdP stubs), Slack (notification delivery — extends Sprint 4's alert delivery), and Custom Webhook (generic inbound/outbound). Each has a specific `config_schema` and `capabilities` set.

---

## Entity Relationship Diagram

```
                    ┌──────────────────────────────────────────────────────────────┐
                    │   INTEGRATION ENGINE                                          │
                    │                                                              │
                    │   ┌──────────────────────────────────────────────┐            │
                    │   │ SYSTEM CATALOG (no org_id)                   │            │
                    │   │                                              │            │
                    │   │   integration_definitions                    │            │
                    │   │     (AWS Config, GitHub, Okta, Slack, etc.)  │            │
                    │   │     config_schema, capabilities, category    │            │
                    │   └──────────────────┬───────────────────────────┘            │
                    │                      │                                        │
                    │   ┌──────────────────┴───────────────────────────┐            │
                    │   │ ORG-SCOPED (has org_id)                      │            │
                    │   │                                              │            │
 organizations ─┬──▶   │   integration_connections                     │            │
                │   │   │     (per-org config, credentials, health)    │            │
                │   │   │     ∞──1 integration_definitions             │            │
                │   │   │     ∞──1 users (created_by)                  │            │
                │   │   │                                              │            │
                │   │   │     1──∞ integration_runs                    │            │
                │   │   │           (sync execution history)           │            │
                │   │   │                                              │            │
                │   │   │           1──∞ integration_logs              │            │
                │   │   │                 (per-run log entries)        │            │
                │   │   │                                              │            │
                │   │   │     1──∞ integration_webhooks                │            │
                │   │   │           (inbound webhook endpoints)        │            │
                │   │   └──────────────────────────────────────────────┘            │
                │   │                                                              │
                └──▶ audit_log (extended with Sprint 9 actions)                    │
                    └──────────────────────────────────────────────────────────────┘
```

**Cross-domain relationships:**
```
integration_connections.created_by ──▶ users (Sprint 1)
identity_providers.integration_connection_id ──▶ integration_connections (new FK, linking Sprint 8 IdP stubs to full integration engine)
integration_runs triggered by ──▶ monitoring worker (Sprint 4) — integration sync can be scheduled via the existing worker infrastructure
```

---

## New Enum Types

### `integration_category`
Category of integration aligned with spec §8.1.

```sql
CREATE TYPE integration_category AS ENUM (
    'cloud_infrastructure',   -- AWS, Azure, GCP
    'identity_access',        -- Okta, Azure AD, Google Workspace, JumpCloud
    'hris',                   -- Workday, BambooHR, Gusto, Rippling
    'endpoint_mdm',           -- CrowdStrike, Jamf, Intune, Kandji
    'siem_logging',           -- Wazuh, Splunk, Datadog, Elastic
    'vulnerability_scanning', -- Qualys, Nessus, Snyk, Rapid7
    'code_devops',            -- GitHub, GitLab, Bitbucket, Azure DevOps
    'task_tracking',          -- Jira, Asana, Linear, Monday.com
    'communication',          -- Slack, Teams, Email
    'password_management',    -- 1Password, LastPass, Bitwarden
    'security_training',      -- KnowBe4, Curricula, Proofpoint
    'background_checks',      -- Checkr, Sterling, GoodHire
    'pen_testing',            -- Cobalt, HackerOne, Bugcrowd
    'antimalware',            -- Malwarebytes, SentinelOne, CrowdStrike
    'waf_ddos',               -- Cloudflare, AWS WAF, Akamai
    'file_integrity',         -- Wazuh FIM, Tripwire, OSSEC
    'network',                -- Meraki, Palo Alto, Fortinet
    'custom'                  -- Custom webhook or private integrations
);
```

### `integration_auth_type`
Authentication method for the integration connection.

```sql
CREATE TYPE integration_auth_type AS ENUM (
    'api_key',          -- Simple API key/token
    'oauth2',           -- OAuth 2.0 authorization code or client credentials
    'basic_auth',       -- Username + password (legacy systems)
    'aws_iam',          -- AWS IAM role assumption (cross-account)
    'service_account',  -- GCP/Azure service account JSON key
    'webhook_secret',   -- Inbound webhook with HMAC verification
    'none'              -- No auth required (public endpoints)
);
```

### `connection_status`
Connection lifecycle for integration connections.

```sql
CREATE TYPE connection_status AS ENUM (
    'pending_setup',  -- Created but credentials not configured
    'configuring',    -- Credentials entered, awaiting validation
    'connected',      -- Successfully connected and validated
    'error',          -- Connection failed (bad credentials, network error, etc.)
    'disabled',       -- Intentionally disabled by user
    'syncing'         -- Currently executing a sync operation
);
```

### `connection_health`
Health state derived from recent run history and health checks.

```sql
CREATE TYPE connection_health AS ENUM (
    'healthy',    -- Last health check / sync succeeded
    'degraded',   -- Partial failures or slow response
    'unhealthy',  -- Last health check / sync failed
    'unknown'     -- No health data yet (new connection)
);
```

### `run_status`
Integration run execution lifecycle.

```sql
CREATE TYPE run_status AS ENUM (
    'pending',     -- Queued, waiting to execute
    'running',     -- Currently executing
    'completed',   -- Finished successfully
    'partial',     -- Finished with some failures (partial sync)
    'failed',      -- Failed entirely
    'cancelled'    -- Cancelled by user or system
);
```

### `run_trigger`
What initiated the integration run.

```sql
CREATE TYPE run_trigger AS ENUM (
    'scheduled',         -- Triggered by sync schedule (cron/interval)
    'manual',            -- Triggered by user via API/UI
    'webhook',           -- Triggered by inbound webhook event
    'retry',             -- Automatic retry after failure
    'health_check',      -- Triggered as a health check probe
    'connection_test'    -- Triggered during connection setup validation
);
```

### `log_level`
Severity level for integration log entries.

```sql
CREATE TYPE log_level AS ENUM (
    'debug',   -- Detailed diagnostic information
    'info',    -- Normal operation events
    'warn',    -- Potential issues that didn't cause failure
    'error'    -- Failures requiring attention
);
```

### `webhook_status`
Status of an inbound webhook endpoint.

```sql
CREATE TYPE webhook_status AS ENUM (
    'active',     -- Receiving events
    'inactive',   -- Created but not active
    'error'       -- Receiving errors (bad signature, etc.)
);
```

---

## Extended Enums

### `audit_action` Extensions (16 new values)

```sql
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.connected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.disconnected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.disabled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.enabled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_run.started';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_run.completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_run.failed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_run.cancelled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_webhook.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_webhook.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_webhook.rotated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_health.check_passed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_health.check_failed';
```

---

## Tables

### Table 1: `integration_definitions`

System-level catalog of available integrations. No `org_id` — this is shared across all tenants, like `frameworks` in Sprint 2. New integrations are added via seed data, not code changes (spec §3.1.1: "add new frameworks via configuration (not code changes)" — same principle applies here).

```sql
CREATE TABLE IF NOT EXISTS integration_definitions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Identity
    name                VARCHAR(255) NOT NULL,
    slug                VARCHAR(100) NOT NULL,         -- URL-friendly identifier (e.g., 'aws-config', 'github', 'okta')
    provider            VARCHAR(100) NOT NULL,         -- Provider key used in code (e.g., 'aws_config', 'github', 'okta', 'slack', 'custom_webhook')
    category            integration_category NOT NULL,
    description         TEXT,
    short_description   VARCHAR(500),                  -- One-liner for catalog listings

    -- Branding
    icon_url            VARCHAR(2048),                 -- URL to integration icon/logo
    documentation_url   VARCHAR(2048),                 -- Link to integration setup docs
    website_url         VARCHAR(2048),                 -- Link to provider's website

    -- Configuration schema
    auth_type           integration_auth_type NOT NULL DEFAULT 'api_key',
    config_schema       JSONB NOT NULL DEFAULT '{}',
    -- JSON Schema defining required/optional config fields. Example for Okta:
    -- {
    --   "type": "object",
    --   "required": ["domain", "api_token"],
    --   "properties": {
    --     "domain": { "type": "string", "title": "Okta Domain", "description": "e.g., acme.okta.com", "pattern": "^[a-z0-9.-]+\\.okta\\.com$" },
    --     "api_token": { "type": "string", "title": "API Token", "description": "Okta API token with read access", "format": "password" }
    --   }
    -- }

    -- Capabilities
    capabilities        TEXT[] NOT NULL DEFAULT '{}',
    -- Supported capabilities for this integration:
    -- 'sync_users'           — Pull user/identity data
    -- 'sync_resources'       — Pull resource/app data
    -- 'sync_configurations'  — Pull configuration state (compliance checks)
    -- 'collect_evidence'     — Collect evidence artifacts
    -- 'send_notifications'   — Send alerts/notifications
    -- 'receive_webhooks'     — Accept inbound webhook events
    -- 'create_tickets'       — Create tickets in task trackers
    -- 'run_scans'            — Trigger vulnerability scans
    -- 'export_logs'          — Export security/audit logs

    -- Availability
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    is_beta             BOOLEAN NOT NULL DEFAULT FALSE,  -- Beta integrations shown with badge
    min_plan            VARCHAR(50),                     -- Minimum subscription plan (future use)

    -- Metadata
    version             VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    tags                TEXT[] NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_integration_definitions_slug UNIQUE (slug),
    CONSTRAINT uq_integration_definitions_provider UNIQUE (provider),
    CONSTRAINT chk_integration_definitions_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_integration_definitions_slug_format CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$' OR length(slug) = 1)
);

COMMENT ON TABLE integration_definitions IS 'System-level catalog of available integrations. No org_id — shared across all tenants.';
COMMENT ON COLUMN integration_definitions.provider IS 'Provider key used in code to dispatch to the correct connector implementation.';
COMMENT ON COLUMN integration_definitions.config_schema IS 'JSON Schema defining the required configuration fields. Used by frontend for dynamic form rendering and backend for validation.';
COMMENT ON COLUMN integration_definitions.capabilities IS 'Array of capability strings this integration supports. Drives feature gating in the UI.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_integration_definitions_category ON integration_definitions (category);
CREATE INDEX IF NOT EXISTS idx_integration_definitions_active ON integration_definitions (is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_integration_definitions_capabilities ON integration_definitions USING GIN (capabilities);
CREATE INDEX IF NOT EXISTS idx_integration_definitions_tags ON integration_definitions USING GIN (tags);

-- Trigger
DROP TRIGGER IF EXISTS trg_integration_definitions_updated_at ON integration_definitions;
CREATE TRIGGER trg_integration_definitions_updated_at
    BEFORE UPDATE ON integration_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

### Table 2: `integration_connections`

Per-org connection to an integration. Stores credentials (encrypted at application layer), sync schedule, health state, and lifecycle status. This is the primary operational table — users interact with connections, not definitions.

```sql
CREATE TABLE IF NOT EXISTS integration_connections (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    definition_id           UUID NOT NULL REFERENCES integration_definitions(id) ON DELETE RESTRICT,

    -- Connection identity
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    instance_label          VARCHAR(255),              -- For multi-instance: "AWS Account (Production)", "AWS Account (Staging)"

    -- Status
    status                  connection_status NOT NULL DEFAULT 'pending_setup',
    health                  connection_health NOT NULL DEFAULT 'unknown',

    -- Configuration (JSONB — secrets encrypted at application layer)
    config                  JSONB NOT NULL DEFAULT '{}',
    -- Validated against definition.config_schema before saving.
    -- Secrets stored as "encrypted:..." values, decrypted only at runtime.

    -- Sync schedule
    sync_enabled            BOOLEAN NOT NULL DEFAULT TRUE,
    sync_interval_mins      INTEGER NOT NULL DEFAULT 360,  -- Default: every 6 hours
    sync_cron               VARCHAR(100),                  -- Optional: cron expression (overrides interval)
    next_sync_at            TIMESTAMPTZ,                   -- Computed: when the next sync should run

    -- Sync state
    last_sync_at            TIMESTAMPTZ,
    last_sync_status        run_status,
    last_sync_error         TEXT,
    last_sync_duration_ms   INTEGER,
    last_sync_stats         JSONB DEFAULT '{}',            -- { "items_synced": 150, "items_created": 5, "items_updated": 12, "items_failed": 0 }

    -- Health check state
    last_health_check_at    TIMESTAMPTZ,
    last_health_error       TEXT,
    health_check_interval_mins INTEGER NOT NULL DEFAULT 60,  -- Default: check health every hour
    consecutive_failures    INTEGER NOT NULL DEFAULT 0,

    -- Access control
    created_by              UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Denormalized counts
    total_runs              INTEGER NOT NULL DEFAULT 0,
    successful_runs         INTEGER NOT NULL DEFAULT 0,
    failed_runs             INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    tags                    TEXT[] NOT NULL DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',   -- Provider-specific metadata (e.g., AWS account ID, GitHub org name)
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- An org can have multiple connections to the same integration (multi-instance)
    -- but names must be unique per org
    CONSTRAINT uq_integration_connections_org_name UNIQUE (org_id, name),
    CONSTRAINT chk_integration_connections_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_integration_connections_sync_interval CHECK (sync_interval_mins >= 5 AND sync_interval_mins <= 10080),
    CONSTRAINT chk_integration_connections_health_interval CHECK (health_check_interval_mins >= 5 AND health_check_interval_mins <= 1440),
    CONSTRAINT chk_integration_connections_consecutive_failures CHECK (consecutive_failures >= 0),
    CONSTRAINT chk_integration_connections_total_runs CHECK (total_runs >= 0),
    CONSTRAINT chk_integration_connections_successful_runs CHECK (successful_runs >= 0),
    CONSTRAINT chk_integration_connections_failed_runs CHECK (failed_runs >= 0)
);

COMMENT ON TABLE integration_connections IS 'Per-org connection to an integration. Stores credentials (encrypted), sync schedule, health state.';
COMMENT ON COLUMN integration_connections.config IS 'Provider-specific config validated against definition.config_schema. Secrets encrypted at application layer.';
COMMENT ON COLUMN integration_connections.instance_label IS 'Human-friendly label for multi-instance connections (e.g., "AWS Prod", "AWS Staging").';
COMMENT ON COLUMN integration_connections.sync_cron IS 'Cron expression for sync schedule. If set, overrides sync_interval_mins.';
COMMENT ON COLUMN integration_connections.consecutive_failures IS 'Consecutive failed syncs/health checks. Used for health degradation logic.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_integration_connections_org_id ON integration_connections (org_id);
CREATE INDEX IF NOT EXISTS idx_integration_connections_org_status ON integration_connections (org_id, status);
CREATE INDEX IF NOT EXISTS idx_integration_connections_org_health ON integration_connections (org_id, health);
CREATE INDEX IF NOT EXISTS idx_integration_connections_definition ON integration_connections (definition_id);
CREATE INDEX IF NOT EXISTS idx_integration_connections_next_sync ON integration_connections (next_sync_at)
    WHERE sync_enabled = TRUE AND status = 'connected';
CREATE INDEX IF NOT EXISTS idx_integration_connections_health_check ON integration_connections (last_health_check_at)
    WHERE status = 'connected';
CREATE INDEX IF NOT EXISTS idx_integration_connections_created_by ON integration_connections (created_by);
CREATE INDEX IF NOT EXISTS idx_integration_connections_tags ON integration_connections USING GIN (tags);

-- Trigger
DROP TRIGGER IF EXISTS trg_integration_connections_updated_at ON integration_connections;
CREATE TRIGGER trg_integration_connections_updated_at
    BEFORE UPDATE ON integration_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

### Table 3: `integration_runs`

Execution history for each sync operation. Every sync (scheduled, manual, webhook-triggered) creates a run record. Provides full audit trail of what data was pulled, when, and what happened.

```sql
CREATE TABLE IF NOT EXISTS integration_runs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connection_id       UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,

    -- Run identity
    trigger             run_trigger NOT NULL,
    triggered_by        UUID REFERENCES users(id) ON DELETE SET NULL,  -- NULL for scheduled/webhook/retry triggers
    status              run_status NOT NULL DEFAULT 'pending',

    -- Execution timeline
    queued_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    duration_ms         INTEGER,

    -- Results
    stats               JSONB NOT NULL DEFAULT '{}',
    -- {
    --   "items_synced": 156,
    --   "items_created": 5,
    --   "items_updated": 12,
    --   "items_deleted": 0,
    --   "items_failed": 1,
    --   "items_skipped": 3,
    --   "evidence_collected": 8,
    --   "bytes_transferred": 1048576
    -- }

    -- Error tracking
    error_message       TEXT,
    error_details       JSONB DEFAULT '{}',            -- Structured error data (stack trace, HTTP status, etc.)
    retry_count         INTEGER NOT NULL DEFAULT 0,
    max_retries         INTEGER NOT NULL DEFAULT 3,

    -- Metadata
    config_snapshot     JSONB DEFAULT '{}',            -- Snapshot of connection config at run time (for debugging, secrets excluded)
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_integration_runs_duration CHECK (duration_ms IS NULL OR duration_ms >= 0),
    CONSTRAINT chk_integration_runs_retry CHECK (retry_count >= 0 AND retry_count <= max_retries),
    CONSTRAINT chk_integration_runs_completed_after_started CHECK (
        completed_at IS NULL OR (started_at IS NOT NULL AND completed_at >= started_at)
    )
);

COMMENT ON TABLE integration_runs IS 'Execution history for integration syncs. One run per sync operation. Provides full auditability.';
COMMENT ON COLUMN integration_runs.stats IS 'JSONB stats: items_synced/created/updated/deleted/failed/skipped, evidence_collected, bytes_transferred.';
COMMENT ON COLUMN integration_runs.config_snapshot IS 'Frozen connection config at run time (secrets excluded). For debugging config-related failures.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_integration_runs_org_id ON integration_runs (org_id);
CREATE INDEX IF NOT EXISTS idx_integration_runs_connection ON integration_runs (connection_id);
CREATE INDEX IF NOT EXISTS idx_integration_runs_connection_status ON integration_runs (connection_id, status);
CREATE INDEX IF NOT EXISTS idx_integration_runs_connection_created ON integration_runs (connection_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_integration_runs_org_status ON integration_runs (org_id, status);
CREATE INDEX IF NOT EXISTS idx_integration_runs_pending ON integration_runs (status, queued_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_integration_runs_running ON integration_runs (status, started_at)
    WHERE status = 'running';

-- Trigger
DROP TRIGGER IF EXISTS trg_integration_runs_updated_at ON integration_runs;
CREATE TRIGGER trg_integration_runs_updated_at
    BEFORE UPDATE ON integration_runs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

### Table 4: `integration_logs`

Structured log entries per integration run. Provides detailed diagnostic information for troubleshooting integration issues. Queryable by level, message content, and time range.

```sql
CREATE TABLE IF NOT EXISTS integration_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    run_id              UUID NOT NULL REFERENCES integration_runs(id) ON DELETE CASCADE,
    connection_id       UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,

    -- Log entry
    level               log_level NOT NULL DEFAULT 'info',
    message             TEXT NOT NULL,
    details             JSONB DEFAULT '{}',            -- Structured context (HTTP request/response, parsed data, etc.)

    -- Source tracking
    source              VARCHAR(255),                  -- e.g., 'okta_connector.sync_users', 'github_connector.list_repos'
    item_ref            VARCHAR(500),                  -- Reference to the specific item being processed (e.g., user email, repo name)

    -- Metadata
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_integration_logs_message_not_empty CHECK (length(trim(message)) > 0)
);

COMMENT ON TABLE integration_logs IS 'Structured log entries per integration run. For troubleshooting integration issues.';
COMMENT ON COLUMN integration_logs.source IS 'Code location or connector method that generated this log entry.';
COMMENT ON COLUMN integration_logs.item_ref IS 'Reference to the specific item being processed (e.g., user email, repo URL).';

-- Indexes (optimized for log querying patterns)
CREATE INDEX IF NOT EXISTS idx_integration_logs_run ON integration_logs (run_id, created_at);
CREATE INDEX IF NOT EXISTS idx_integration_logs_connection ON integration_logs (connection_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_integration_logs_org_level ON integration_logs (org_id, level);
CREATE INDEX IF NOT EXISTS idx_integration_logs_run_level ON integration_logs (run_id, level);
CREATE INDEX IF NOT EXISTS idx_integration_logs_created ON integration_logs (created_at);

-- Note: No updated_at trigger — log entries are immutable (append-only).
```

---

### Table 5: `integration_webhooks`

Inbound webhook endpoints for integrations that push data (GitHub events, Slack events, AWS SNS, etc.). Each webhook gets a unique URL with HMAC secret verification.

```sql
CREATE TABLE IF NOT EXISTS integration_webhooks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connection_id       UUID NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,

    -- Webhook identity
    name                VARCHAR(255) NOT NULL,
    description         TEXT,

    -- Endpoint configuration
    -- The webhook URL is deterministic: POST /api/v1/webhooks/receive/{id}
    -- No need to store the full URL — it's derived from the webhook ID.
    webhook_secret      VARCHAR(500) NOT NULL,         -- HMAC secret for request verification (encrypted at app layer)
    signature_header    VARCHAR(100) NOT NULL DEFAULT 'X-Webhook-Signature',  -- HTTP header containing the HMAC signature
    signature_algo      VARCHAR(50) NOT NULL DEFAULT 'sha256',              -- HMAC algorithm (sha256, sha1)

    -- Status
    status              webhook_status NOT NULL DEFAULT 'active',

    -- Event filtering
    event_types         TEXT[] NOT NULL DEFAULT '{}',   -- Event types to accept (empty = all). e.g., ['push', 'pull_request', 'issues']

    -- Stats
    total_received      INTEGER NOT NULL DEFAULT 0,
    total_processed     INTEGER NOT NULL DEFAULT 0,
    total_failed        INTEGER NOT NULL DEFAULT 0,
    last_received_at    TIMESTAMPTZ,
    last_error          TEXT,

    -- Metadata
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_integration_webhooks_connection_name UNIQUE (connection_id, name),
    CONSTRAINT chk_integration_webhooks_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_integration_webhooks_secret_not_empty CHECK (length(webhook_secret) > 0),
    CONSTRAINT chk_integration_webhooks_total_received CHECK (total_received >= 0),
    CONSTRAINT chk_integration_webhooks_total_processed CHECK (total_processed >= 0),
    CONSTRAINT chk_integration_webhooks_total_failed CHECK (total_failed >= 0)
);

COMMENT ON TABLE integration_webhooks IS 'Inbound webhook endpoints for real-time event ingestion. HMAC-verified.';
COMMENT ON COLUMN integration_webhooks.webhook_secret IS 'HMAC secret for verifying inbound requests. Encrypted at application layer.';
COMMENT ON COLUMN integration_webhooks.event_types IS 'Event type filter. Empty array = accept all events. Provider-specific (e.g., GitHub: push, pull_request).';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_integration_webhooks_org ON integration_webhooks (org_id);
CREATE INDEX IF NOT EXISTS idx_integration_webhooks_connection ON integration_webhooks (connection_id);
CREATE INDEX IF NOT EXISTS idx_integration_webhooks_active ON integration_webhooks (status) WHERE status = 'active';

-- Trigger
DROP TRIGGER IF EXISTS trg_integration_webhooks_updated_at ON integration_webhooks;
CREATE TRIGGER trg_integration_webhooks_updated_at
    BEFORE UPDATE ON integration_webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

## Cross-Reference FK Extensions

### `identity_providers` Extension (Sprint 8 → Sprint 9 bridge)

Link Sprint 8's identity provider stubs to the full integration engine when a connection is established.

```sql
-- Add FK column linking identity_providers to their integration_connection
ALTER TABLE identity_providers
    ADD COLUMN IF NOT EXISTS integration_connection_id UUID REFERENCES integration_connections(id) ON DELETE SET NULL;

-- Index for the new FK
CREATE INDEX IF NOT EXISTS idx_identity_providers_integration_connection
    ON identity_providers (integration_connection_id)
    WHERE integration_connection_id IS NOT NULL;

COMMENT ON COLUMN identity_providers.integration_connection_id IS 'Optional link to Sprint 9 integration connection. When set, IdP health/sync managed by integration engine.';
```

---

## Query Patterns

### Connection Health Dashboard

```sql
-- Integration health summary for an org
SELECT
    id.name AS integration_name,
    id.category,
    ic.name AS connection_name,
    ic.instance_label,
    ic.status,
    ic.health,
    ic.last_sync_at,
    ic.last_sync_status,
    ic.last_sync_duration_ms,
    ic.last_health_check_at,
    ic.consecutive_failures,
    ic.total_runs,
    ic.successful_runs,
    ic.failed_runs,
    CASE WHEN ic.total_runs > 0
         THEN ROUND(100.0 * ic.successful_runs / ic.total_runs, 1)
         ELSE 0 END AS success_rate_pct
FROM integration_connections ic
JOIN integration_definitions id ON id.id = ic.definition_id
WHERE ic.org_id = $1
  AND ic.status != 'disabled'
ORDER BY
    CASE ic.health
        WHEN 'unhealthy' THEN 1
        WHEN 'degraded' THEN 2
        WHEN 'unknown' THEN 3
        WHEN 'healthy' THEN 4
    END,
    ic.last_sync_at DESC NULLS LAST;
```

### Due Syncs (Worker Query)

```sql
-- Find connections due for sync (used by monitoring worker)
SELECT
    ic.id AS connection_id,
    ic.org_id,
    id.provider,
    ic.config,
    ic.sync_interval_mins,
    ic.sync_cron,
    ic.next_sync_at,
    ic.last_sync_at
FROM integration_connections ic
JOIN integration_definitions id ON id.id = ic.definition_id
WHERE ic.sync_enabled = TRUE
  AND ic.status = 'connected'
  AND (ic.next_sync_at IS NULL OR ic.next_sync_at <= NOW())
ORDER BY ic.next_sync_at NULLS FIRST
LIMIT 50;
```

### Run History with Stats

```sql
-- Recent runs for a connection with duration and stats
SELECT
    ir.id,
    ir.trigger,
    ir.status,
    ir.queued_at,
    ir.started_at,
    ir.completed_at,
    ir.duration_ms,
    ir.stats,
    ir.error_message,
    ir.retry_count,
    u.first_name || ' ' || u.last_name AS triggered_by_name,
    (SELECT COUNT(*) FROM integration_logs il WHERE il.run_id = ir.id AND il.level = 'error') AS error_count,
    (SELECT COUNT(*) FROM integration_logs il WHERE il.run_id = ir.id AND il.level = 'warn') AS warn_count
FROM integration_runs ir
LEFT JOIN users u ON u.id = ir.triggered_by
WHERE ir.connection_id = $1
  AND ir.org_id = $2
ORDER BY ir.created_at DESC
LIMIT 20;
```

### Capability-Based Search

```sql
-- Find all integrations with a specific capability (e.g., sync_users for access reviews)
SELECT
    id.id,
    id.name,
    id.slug,
    id.category,
    id.short_description,
    id.icon_url,
    id.auth_type,
    id.capabilities,
    id.is_beta,
    (SELECT COUNT(*) FROM integration_connections ic
     WHERE ic.definition_id = id.id AND ic.org_id = $1) AS org_connections
FROM integration_definitions id
WHERE id.is_active = TRUE
  AND 'sync_users' = ANY(id.capabilities)
ORDER BY id.name;
```

### Integration Data Freshness

```sql
-- Data freshness check: connections that haven't synced recently
SELECT
    ic.id,
    ic.name,
    id.name AS integration_name,
    ic.last_sync_at,
    ic.sync_interval_mins,
    EXTRACT(EPOCH FROM (NOW() - ic.last_sync_at)) / 60 AS minutes_since_sync,
    ic.health,
    ic.consecutive_failures
FROM integration_connections ic
JOIN integration_definitions id ON id.id = ic.definition_id
WHERE ic.org_id = $1
  AND ic.status = 'connected'
  AND ic.sync_enabled = TRUE
  AND (
    ic.last_sync_at IS NULL
    OR ic.last_sync_at < NOW() - (ic.sync_interval_mins * 2 || ' minutes')::INTERVAL
  )
ORDER BY ic.last_sync_at NULLS FIRST;
```

---

## Migration Files

| File | Contents |
|------|----------|
| `062_sprint9_enums.sql` | 8 new enums + 16 audit_action extensions |
| `063_integration_definitions.sql` | integration_definitions table with indexes, trigger |
| `064_integration_connections.sql` | integration_connections table with indexes, trigger |
| `065_integration_runs.sql` | integration_runs table with indexes, trigger |
| `066_integration_logs.sql` | integration_logs table with indexes (no updated_at — append-only) |
| `067_integration_webhooks.sql` | integration_webhooks table with indexes, trigger |
| `068_sprint9_fk_cross_refs.sql` | identity_providers.integration_connection_id FK extension |
| `069_sprint9_seed_definitions.sql` | 5 starter integration definitions (AWS Config, GitHub, Okta, Slack, Custom Webhook) |
| `070_sprint9_seed_demo.sql` | Demo data: 3 connections (Okta connected + GitHub connected + Slack connected), 5 runs, 15 logs, 1 webhook |

---

## Seed Data

### 5 Starter Integration Definitions

#### 1. AWS Config

```sql
INSERT INTO integration_definitions (name, slug, provider, category, description, short_description, auth_type, config_schema, capabilities, tags, version)
VALUES (
    'AWS Config',
    'aws-config',
    'aws_config',
    'cloud_infrastructure',
    'Monitor AWS resource configurations for compliance. Pulls Config rules, evaluations, and resource inventory. Supports multi-account via IAM cross-account role assumption.',
    'Cloud resource compliance monitoring via AWS Config rules',
    'aws_iam',
    '{
        "type": "object",
        "required": ["account_id", "region", "role_arn"],
        "properties": {
            "account_id": { "type": "string", "title": "AWS Account ID", "description": "12-digit AWS account ID", "pattern": "^[0-9]{12}$" },
            "region": { "type": "string", "title": "AWS Region", "description": "Primary region for Config", "default": "us-east-1", "enum": ["us-east-1","us-west-2","eu-west-1","eu-central-1","ap-southeast-1"] },
            "role_arn": { "type": "string", "title": "IAM Role ARN", "description": "Cross-account IAM role for Config read access", "pattern": "^arn:aws:iam::[0-9]{12}:role/.+" },
            "external_id": { "type": "string", "title": "External ID", "description": "External ID for role assumption (recommended)" }
        }
    }',
    ARRAY['sync_configurations', 'collect_evidence', 'run_scans'],
    ARRAY['aws', 'cloud', 'compliance', 'infrastructure'],
    '1.0.0'
);
```

#### 2. GitHub

```sql
INSERT INTO integration_definitions (name, slug, provider, category, description, short_description, auth_type, config_schema, capabilities, tags, version)
VALUES (
    'GitHub',
    'github',
    'github',
    'code_devops',
    'Monitor GitHub organizations for code security compliance. Pulls repository settings, branch protection rules, vulnerability alerts, and contributor access. Supports GitHub Enterprise.',
    'Code repository monitoring and security compliance',
    'api_key',
    '{
        "type": "object",
        "required": ["org_name", "access_token"],
        "properties": {
            "org_name": { "type": "string", "title": "Organization Name", "description": "GitHub organization slug" },
            "access_token": { "type": "string", "title": "Personal Access Token", "description": "PAT with org:read, repo:read scopes", "format": "password" },
            "enterprise_url": { "type": "string", "title": "Enterprise URL", "description": "GitHub Enterprise base URL (leave empty for github.com)", "format": "uri" },
            "include_archived": { "type": "boolean", "title": "Include Archived Repos", "default": false }
        }
    }',
    ARRAY['sync_resources', 'sync_configurations', 'collect_evidence', 'receive_webhooks'],
    ARRAY['github', 'devops', 'code', 'security'],
    '1.0.0'
);
```

#### 3. Okta

```sql
INSERT INTO integration_definitions (name, slug, provider, category, description, short_description, auth_type, config_schema, capabilities, tags, version)
VALUES (
    'Okta',
    'okta',
    'okta',
    'identity_access',
    'Sync users, groups, and applications from Okta. Monitors MFA enrollment, password policies, and application access. Feeds data into access reviews (Sprint 8) and automated testing (Sprint 4).',
    'Identity and access management via Okta',
    'api_key',
    '{
        "type": "object",
        "required": ["domain", "api_token"],
        "properties": {
            "domain": { "type": "string", "title": "Okta Domain", "description": "Your Okta domain (e.g., acme.okta.com)", "pattern": "^[a-z0-9.-]+\\\\.(okta\\\\.com|oktapreview\\\\.com)$" },
            "api_token": { "type": "string", "title": "API Token", "description": "Okta API token with read-only admin access", "format": "password" },
            "sync_deprovisioned": { "type": "boolean", "title": "Sync Deprovisioned Users", "description": "Include deactivated users in sync", "default": false }
        }
    }',
    ARRAY['sync_users', 'sync_resources', 'sync_configurations', 'collect_evidence'],
    ARRAY['okta', 'identity', 'sso', 'mfa', 'access'],
    '1.0.0'
);
```

#### 4. Slack

```sql
INSERT INTO integration_definitions (name, slug, provider, category, description, short_description, auth_type, config_schema, capabilities, tags, version)
VALUES (
    'Slack',
    'slack',
    'slack',
    'communication',
    'Send compliance notifications, alerts, and workflow requests to Slack channels. Supports channel-based routing, severity-based escalation, and interactive approvals. Extends Sprint 4 alert delivery.',
    'Compliance notifications and workflow automation via Slack',
    'oauth2',
    '{
        "type": "object",
        "required": ["workspace_name", "bot_token"],
        "properties": {
            "workspace_name": { "type": "string", "title": "Workspace Name", "description": "Slack workspace display name" },
            "bot_token": { "type": "string", "title": "Bot Token", "description": "Slack bot OAuth token (xoxb-...)", "format": "password", "pattern": "^xoxb-" },
            "default_channel": { "type": "string", "title": "Default Channel", "description": "Default channel for notifications (e.g., #compliance-alerts)" },
            "signing_secret": { "type": "string", "title": "Signing Secret", "description": "Slack app signing secret for webhook verification", "format": "password" }
        }
    }',
    ARRAY['send_notifications', 'receive_webhooks'],
    ARRAY['slack', 'notifications', 'communication', 'alerts'],
    '1.0.0'
);
```

#### 5. Custom Webhook

```sql
INSERT INTO integration_definitions (name, slug, provider, category, description, short_description, auth_type, config_schema, capabilities, tags, version)
VALUES (
    'Custom Webhook',
    'custom-webhook',
    'custom_webhook',
    'custom',
    'Generic webhook integration for custom systems. Supports both inbound (receive events via webhook URL) and outbound (send data to external endpoints) patterns. Use for private integrations and unsupported systems.',
    'Generic inbound/outbound webhook for custom integrations',
    'webhook_secret',
    '{
        "type": "object",
        "required": ["webhook_direction"],
        "properties": {
            "webhook_direction": { "type": "string", "title": "Direction", "enum": ["inbound", "outbound", "bidirectional"], "description": "Inbound: receive events. Outbound: send data. Bidirectional: both." },
            "outbound_url": { "type": "string", "title": "Outbound URL", "description": "URL to send data to (for outbound/bidirectional)", "format": "uri" },
            "outbound_auth_type": { "type": "string", "title": "Outbound Auth", "enum": ["none", "api_key", "bearer_token", "basic_auth"], "default": "none" },
            "outbound_auth_value": { "type": "string", "title": "Auth Value", "description": "API key, bearer token, or base64-encoded credentials", "format": "password" },
            "outbound_headers": { "type": "object", "title": "Custom Headers", "description": "Additional HTTP headers for outbound requests" },
            "payload_format": { "type": "string", "title": "Payload Format", "enum": ["json", "form"], "default": "json" }
        }
    }',
    ARRAY['receive_webhooks', 'collect_evidence', 'send_notifications'],
    ARRAY['custom', 'webhook', 'private', 'generic'],
    '1.0.0'
);
```

### Demo Data

```sql
-- 3 demo connections: Okta (connected/healthy), GitHub (connected/healthy), Slack (connected/healthy)
-- For each connection:
--   - Realistic config (with encrypted: prefixed secrets)
--   - Last sync stats
--   - Health state

-- 5 demo runs across the 3 connections:
--   - 2 Okta runs (1 completed, 1 completed with stats)
--   - 2 GitHub runs (1 completed, 1 partial — some repos private)
--   - 1 Slack run (connection test)

-- 15 demo log entries across the 5 runs:
--   - Mix of info, warn, error levels
--   - Realistic messages (e.g., "Synced 156 users from Okta", "Repository X is archived, skipping")

-- 1 demo webhook for GitHub (push events)
```

---

## Design Notes

### Relationship to Sprint 8 Identity Providers

Sprint 8's `identity_providers` table was explicitly designed as a stub for Sprint 9. The bridge works as follows:

1. Sprint 8 `identity_providers` continues to function standalone for access reviews
2. Sprint 9 adds `integration_connection_id` FK to `identity_providers`
3. When an Okta/Azure AD/Google Workspace integration is connected in Sprint 9, the system:
   - Creates an `integration_connection` record
   - Links the corresponding `identity_provider` via `integration_connection_id`
   - IdP sync is now managed by the integration engine (health checks, run history, structured logs)
4. Sprint 8's access review flow is unchanged — it reads from `identity_providers` and `access_entries` regardless of whether the integration engine is connected

### Health Degradation Logic

Connection health is computed from recent run history:
- **healthy**: Last run succeeded, no consecutive failures
- **degraded**: Last run partially succeeded OR 1-2 consecutive failures
- **unhealthy**: 3+ consecutive failures OR last health check failed
- **unknown**: No runs or health checks yet

Health is updated after each run and each health check. The worker polls for connections due for health checks separately from sync operations.

### Log Retention Strategy

Integration logs can grow rapidly (especially at debug level). Recommended retention:
- Error/warn logs: 90 days
- Info logs: 30 days
- Debug logs: 7 days

A future background job should handle log cleanup. For Sprint 9, no automatic pruning — logs accumulate until manually cleaned or a cleanup job is added.

### Worker Integration

Sprint 9 extends Sprint 4's monitoring worker with integration sync scheduling:
- The worker polls `integration_connections` for due syncs (using the `next_sync_at` index)
- For each due connection, it creates a `pending` run, dispatches to the appropriate connector, and updates run status
- Connectors are Go interfaces: `type IntegrationConnector interface { Sync(ctx, config) (RunStats, error); HealthCheck(ctx, config) error }`
- The 5 starter connectors implement this interface with provider-specific logic

### Multi-Instance Support

An org can have multiple connections to the same integration definition (e.g., AWS Config for production account + staging account). Each connection has its own credentials, sync schedule, and health state. The `instance_label` field helps users differentiate instances in the UI.

### Secret Handling

Connection configs often contain secrets (API tokens, OAuth credentials, etc.). These are:
1. Encrypted at the application layer before storage (AES-256)
2. Stored as `"encrypted:..."` values in the JSONB `config` column
3. Decrypted only at runtime when executing a sync
4. Never included in `config_snapshot` on runs
5. Masked in API responses (replaced with `"***masked***"`)
6. Sprint 4's existing SSRF protections apply to outbound webhook URLs
