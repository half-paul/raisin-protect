# Sprint 9 — API Specification: Integration Engine (Foundation)

## Overview

Sprint 9 adds the Integration Engine — the extensible connector framework that powers automated evidence collection, identity sync, and third-party monitoring. Endpoints cover the integration catalog (system-level definitions), per-org connection management (CRUD + lifecycle + health), sync execution with run history and logs, inbound webhooks for real-time events, starter integration helpers (AWS Config, GitHub, Okta, Slack), and a connection health dashboard.

This implements spec §8 (Integration Engine): pluggable integration framework with health monitoring, 5 starter integrations, connection management UI, and integration status dashboard.

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## Access Control Model

### Role Permissions Matrix

| Action | compliance_manager | security_engineer | it_admin | ciso | devops_engineer | auditor | vendor_manager |
|--------|-------------------|-------------------|----------|------|-----------------|---------|----------------|
| Browse integration catalog | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Create connections | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Manage connections (update/delete) | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| View connections | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Trigger sync / health check | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| View run history / logs | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Manage webhooks | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| View dashboard | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Starter integration data (preview) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅* | ❌ |

\* Auditors can view integration data when it's linked as evidence in audits (Sprint 7), but cannot directly access integration endpoints.

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `integrations` | System | Integration catalog (definitions) |
| `integration-connections` | Per-org | Org's configured connections |
| `integration-connections/:id/runs` | Per-connection | Sync execution history |
| `integration-connections/:id/runs/:rid/logs` | Per-run | Structured log entries |
| `integration-connections/:id/webhooks` | Per-connection | Inbound webhook endpoints |
| `integrations/dashboard` | Per-org | Connection health dashboard |

---

## Endpoints

---

### 1. Integration Catalog (System-Level)

---

#### `GET /api/v1/integrations`

List available integration definitions. System-level catalog — not org-scoped.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `category` | string | No | Filter by category: `cloud_infrastructure`, `identity_access`, `code_devops`, `communication`, `custom`, etc. |
| `capability` | string | No | Filter by capability: `sync_users`, `sync_resources`, `collect_evidence`, `send_notifications`, `receive_webhooks`, etc. |
| `search` | string | No | Search name and description |
| `is_beta` | boolean | No | Filter for beta integrations |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "AWS Config",
      "slug": "aws-config",
      "provider": "aws_config",
      "category": "cloud_infrastructure",
      "short_description": "Cloud resource compliance monitoring via AWS Config rules",
      "icon_url": "/icons/integrations/aws-config.svg",
      "auth_type": "aws_iam",
      "capabilities": ["sync_configurations", "collect_evidence", "run_scans"],
      "is_active": true,
      "is_beta": false,
      "version": "1.0.0",
      "tags": ["aws", "cloud", "compliance"],
      "org_connections": 0,
      "created_at": "2026-02-21T00:00:00Z"
    },
    {
      "id": "uuid",
      "name": "GitHub",
      "slug": "github",
      "provider": "github",
      "category": "code_devops",
      "short_description": "Code repository monitoring and security compliance",
      "icon_url": "/icons/integrations/github.svg",
      "auth_type": "api_key",
      "capabilities": ["sync_resources", "sync_configurations", "collect_evidence", "receive_webhooks"],
      "is_active": true,
      "is_beta": false,
      "version": "1.0.0",
      "tags": ["github", "devops", "code"],
      "org_connections": 1,
      "created_at": "2026-02-21T00:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

---

#### `GET /api/v1/integrations/:id`

Get detailed information about an integration definition, including config schema for form rendering.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "name": "Okta",
    "slug": "okta",
    "provider": "okta",
    "category": "identity_access",
    "description": "Sync users, groups, and applications from Okta. Monitors MFA enrollment, password policies, and application access.",
    "short_description": "Identity and access management via Okta",
    "icon_url": "/icons/integrations/okta.svg",
    "documentation_url": "https://docs.raisinprotect.com/integrations/okta",
    "website_url": "https://www.okta.com",
    "auth_type": "api_key",
    "config_schema": {
      "type": "object",
      "required": ["domain", "api_token"],
      "properties": {
        "domain": {
          "type": "string",
          "title": "Okta Domain",
          "description": "Your Okta domain (e.g., acme.okta.com)",
          "pattern": "^[a-z0-9.-]+\\.okta\\.com$"
        },
        "api_token": {
          "type": "string",
          "title": "API Token",
          "description": "Okta API token with read-only admin access",
          "format": "password"
        },
        "sync_deprovisioned": {
          "type": "boolean",
          "title": "Sync Deprovisioned Users",
          "default": false
        }
      }
    },
    "capabilities": ["sync_users", "sync_resources", "sync_configurations", "collect_evidence"],
    "is_active": true,
    "is_beta": false,
    "version": "1.0.0",
    "tags": ["okta", "identity", "sso", "mfa"],
    "org_connections": 1,
    "created_at": "2026-02-21T00:00:00Z",
    "updated_at": "2026-02-21T00:00:00Z"
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Integration not found

---

### 2. Integration Connection Management

---

#### `GET /api/v1/integration-connections`

List the org's integration connections with filtering.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | Filter: `pending_setup`, `configuring`, `connected`, `error`, `disabled`, `syncing` |
| `health` | string | No | Filter: `healthy`, `degraded`, `unhealthy`, `unknown` |
| `definition_id` | UUID | No | Filter by integration definition |
| `category` | string | No | Filter by integration category |
| `search` | string | No | Search connection name and description |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort: `name`, `status`, `health`, `last_sync_at`, `created_at` |
| `order` | string | No | `asc`, `desc` (default: `asc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "definition_id": "uuid",
      "integration_name": "Okta",
      "integration_slug": "okta",
      "integration_category": "identity_access",
      "integration_icon_url": "/icons/integrations/okta.svg",
      "name": "Okta SSO (Production)",
      "description": "Primary identity provider for all employees",
      "instance_label": "Production",
      "status": "connected",
      "health": "healthy",
      "sync_enabled": true,
      "sync_interval_mins": 360,
      "last_sync_at": "2026-02-21T12:00:00Z",
      "last_sync_status": "completed",
      "last_sync_duration_ms": 3200,
      "last_sync_stats": {
        "items_synced": 156,
        "items_created": 2,
        "items_updated": 5,
        "items_failed": 0
      },
      "last_health_check_at": "2026-02-21T15:00:00Z",
      "consecutive_failures": 0,
      "total_runs": 42,
      "successful_runs": 41,
      "failed_runs": 1,
      "success_rate_pct": 97.6,
      "tags": ["production", "identity"],
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-02-21T15:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 3, "total_pages": 1 }
}
```

---

#### `GET /api/v1/integration-connections/:id`

Get detailed connection information, including config (with secrets masked).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "definition_id": "uuid",
    "integration_name": "Okta",
    "integration_slug": "okta",
    "integration_category": "identity_access",
    "capabilities": ["sync_users", "sync_resources", "sync_configurations", "collect_evidence"],
    "name": "Okta SSO (Production)",
    "description": "Primary identity provider for all employees",
    "instance_label": "Production",
    "status": "connected",
    "health": "healthy",
    "config": {
      "domain": "acme.okta.com",
      "api_token": "***masked***",
      "sync_deprovisioned": false
    },
    "sync_enabled": true,
    "sync_interval_mins": 360,
    "sync_cron": null,
    "next_sync_at": "2026-02-21T18:00:00Z",
    "last_sync_at": "2026-02-21T12:00:00Z",
    "last_sync_status": "completed",
    "last_sync_error": null,
    "last_sync_duration_ms": 3200,
    "last_sync_stats": {
      "items_synced": 156,
      "items_created": 2,
      "items_updated": 5,
      "items_failed": 0
    },
    "last_health_check_at": "2026-02-21T15:00:00Z",
    "last_health_error": null,
    "health_check_interval_mins": 60,
    "consecutive_failures": 0,
    "total_runs": 42,
    "successful_runs": 41,
    "failed_runs": 1,
    "created_by": "uuid",
    "created_by_name": "Alice Admin",
    "tags": ["production", "identity"],
    "metadata": {
      "okta_org_id": "00o1234567890",
      "okta_org_url": "https://acme.okta.com"
    },
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-02-21T15:00:00Z"
  }
}
```

---

#### `POST /api/v1/integration-connections`

Create a new integration connection.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Request Body:**

```json
{
  "definition_id": "uuid",
  "name": "Okta SSO (Production)",
  "description": "Primary identity provider for all employees",
  "instance_label": "Production",
  "config": {
    "domain": "acme.okta.com",
    "api_token": "00abcdef...",
    "sync_deprovisioned": false
  },
  "sync_enabled": true,
  "sync_interval_mins": 360,
  "sync_cron": null,
  "health_check_interval_mins": 60,
  "tags": ["production", "identity"]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `definition_id` | UUID | Yes | Must reference active integration_definitions |
| `name` | string | Yes | 1-255 chars, unique per org |
| `config` | object | Yes | Validated against definition's config_schema |
| `instance_label` | string | No | Human-friendly label for multi-instance |
| `sync_enabled` | boolean | No | Default: true |
| `sync_interval_mins` | integer | No | 5-10080, default: 360 |
| `sync_cron` | string | No | Valid cron expression (overrides interval) |
| `health_check_interval_mins` | integer | No | 5-1440, default: 60 |
| `description` | string | No | Free text |
| `tags` | string[] | No | Array of tags |

**Response (201):**

```json
{
  "data": {
    "id": "uuid",
    "name": "Okta SSO (Production)",
    "status": "pending_setup",
    "health": "unknown",
    "message": "Connection created. Use POST /integration-connections/:id/test to validate credentials, then POST /integration-connections/:id/enable to start syncing.",
    "...": "..."
  }
}
```

**Error Codes:**
- `400` Config does not match definition's config_schema
- `401` Unauthorized
- `403` Forbidden
- `404` Integration definition not found
- `409` Connection name already exists for this org

---

#### `PUT /api/v1/integration-connections/:id`

Update connection configuration. Config fields are merged (not replaced) — send only changed fields. Secrets must be resent if changed (masked values are ignored).

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Request Body:** Same fields as POST, all optional.

**Response (200):** Updated connection object.

**Error Codes:**
- `400` Invalid config
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `409` Name conflict
- `422` Cannot update config while sync is in progress

---

#### `DELETE /api/v1/integration-connections/:id`

Delete a connection. Cascades to all runs, logs, and webhooks. Cannot delete while sync is running.

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Response (204):** No content.

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden (only it_admin, ciso)
- `404` Not found
- `409` Cannot delete while sync is in progress

---

#### `POST /api/v1/integration-connections/:id/test`

Test connection credentials by performing a lightweight health check against the provider. Does not sync data — just validates authentication.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "connection_id": "uuid",
    "test_result": "success",
    "message": "Successfully authenticated with Okta (acme.okta.com). Organization: Acme Corp. Users: 156.",
    "details": {
      "provider_response_time_ms": 245,
      "provider_version": "2024.01.0",
      "permissions_verified": ["users.read", "apps.read", "groups.read"]
    },
    "status": "connected",
    "health": "healthy",
    "tested_at": "2026-02-21T15:30:00Z"
  }
}
```

**Error Response (200 — test failed but endpoint succeeded):**

```json
{
  "data": {
    "connection_id": "uuid",
    "test_result": "failed",
    "message": "Authentication failed: Invalid API token. Verify your Okta API token has admin read access.",
    "details": {
      "error_type": "auth_error",
      "http_status": 401,
      "provider_message": "Invalid token provided"
    },
    "status": "error",
    "health": "unhealthy",
    "tested_at": "2026-02-21T15:30:00Z"
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Connection not found
- `422` Connection in `disabled` status (must enable first)

---

#### `POST /api/v1/integration-connections/:id/enable`

Enable a connection (from `disabled` or `pending_setup` → `connected`). Requires successful test first.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "status": "connected",
    "sync_enabled": true,
    "next_sync_at": "2026-02-21T16:00:00Z",
    "message": "Connection enabled. First sync scheduled for 2026-02-21T16:00:00Z."
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `422` Connection has never passed a test (must test first)

---

#### `POST /api/v1/integration-connections/:id/disable`

Disable a connection. Stops all syncs and health checks. Does not delete data.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Request Body:** (optional)

```json
{
  "reason": "Rotating API credentials, will re-enable after update"
}
```

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "status": "disabled",
    "sync_enabled": false,
    "message": "Connection disabled. Syncs and health checks paused."
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `409` Cannot disable while sync is in progress

---

### 3. Sync Execution & Run History

---

#### `POST /api/v1/integration-connections/:id/sync`

Trigger a manual sync for a connected integration.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Request Body:** (optional)

```json
{
  "full_sync": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|------------|
| `full_sync` | boolean | No | If true, full re-sync (not incremental). Default: false. |

**Response (202):**

```json
{
  "data": {
    "run_id": "uuid",
    "connection_id": "uuid",
    "trigger": "manual",
    "status": "pending",
    "message": "Sync queued. Track progress via GET /integration-connections/:id/runs/:run_id"
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `409` Sync already in progress
- `422` Connection not in `connected` status

---

#### `GET /api/v1/integration-connections/:id/runs`

List sync runs for a connection with filtering.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | Filter: `pending`, `running`, `completed`, `partial`, `failed`, `cancelled` |
| `trigger` | string | No | Filter: `scheduled`, `manual`, `webhook`, `retry`, `health_check`, `connection_test` |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort: `created_at`, `started_at`, `duration_ms` |
| `order` | string | No | `asc`, `desc` (default: `desc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "connection_id": "uuid",
      "trigger": "scheduled",
      "triggered_by": null,
      "triggered_by_name": null,
      "status": "completed",
      "queued_at": "2026-02-21T12:00:00Z",
      "started_at": "2026-02-21T12:00:02Z",
      "completed_at": "2026-02-21T12:00:05Z",
      "duration_ms": 3200,
      "stats": {
        "items_synced": 156,
        "items_created": 2,
        "items_updated": 5,
        "items_failed": 0,
        "evidence_collected": 3
      },
      "error_message": null,
      "retry_count": 0,
      "log_counts": {
        "info": 8,
        "warn": 1,
        "error": 0
      },
      "created_at": "2026-02-21T12:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 42, "total_pages": 3 }
}
```

---

#### `GET /api/v1/integration-connections/:id/runs/:rid`

Get detailed information about a specific run, including full stats and error details.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "connection_id": "uuid",
    "connection_name": "Okta SSO (Production)",
    "integration_name": "Okta",
    "trigger": "manual",
    "triggered_by": "uuid",
    "triggered_by_name": "Alice Admin",
    "status": "partial",
    "queued_at": "2026-02-21T14:00:00Z",
    "started_at": "2026-02-21T14:00:01Z",
    "completed_at": "2026-02-21T14:00:08Z",
    "duration_ms": 7500,
    "stats": {
      "items_synced": 148,
      "items_created": 0,
      "items_updated": 3,
      "items_deleted": 0,
      "items_failed": 2,
      "items_skipped": 6,
      "evidence_collected": 1,
      "bytes_transferred": 524288
    },
    "error_message": "Partial sync: 2 items failed during processing",
    "error_details": {
      "failed_items": [
        { "item": "user:john@acme.com", "error": "Missing required field: department" },
        { "item": "app:legacy-tool", "error": "Application not found in Okta" }
      ]
    },
    "retry_count": 0,
    "max_retries": 3,
    "config_snapshot": {
      "domain": "acme.okta.com",
      "sync_deprovisioned": false
    },
    "log_counts": {
      "debug": 0,
      "info": 12,
      "warn": 3,
      "error": 2
    },
    "created_at": "2026-02-21T14:00:00Z",
    "updated_at": "2026-02-21T14:00:08Z"
  }
}
```

---

#### `POST /api/v1/integration-connections/:id/runs/:rid/cancel`

Cancel a pending or running sync.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "status": "cancelled",
    "message": "Sync cancelled. Partial data from this run may remain."
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Run not found
- `409` Run already completed/failed/cancelled

---

#### `GET /api/v1/integration-connections/:id/runs/:rid/logs`

Get structured log entries for a specific run.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `level` | string | No | Filter: `debug`, `info`, `warn`, `error` |
| `search` | string | No | Search log messages |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 50, max: 200) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "level": "info",
      "message": "Starting Okta user sync",
      "details": { "domain": "acme.okta.com" },
      "source": "okta_connector.sync_users",
      "item_ref": null,
      "created_at": "2026-02-21T14:00:01Z"
    },
    {
      "id": "uuid",
      "level": "info",
      "message": "Synced 148 users (3 updated, 6 skipped, 2 failed)",
      "details": { "total": 156, "synced": 148, "updated": 3, "skipped": 6, "failed": 2 },
      "source": "okta_connector.sync_users",
      "item_ref": null,
      "created_at": "2026-02-21T14:00:05Z"
    },
    {
      "id": "uuid",
      "level": "error",
      "message": "Failed to process user: Missing required field 'department'",
      "details": { "user_email": "john@acme.com", "missing_fields": ["department"] },
      "source": "okta_connector.process_user",
      "item_ref": "user:john@acme.com",
      "created_at": "2026-02-21T14:00:04Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 50, "total": 17, "total_pages": 1 }
}
```

---

### 4. Health Monitoring

---

#### `GET /api/v1/integration-connections/:id/health`

Get detailed health status for a connection, including recent check history.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "connection_id": "uuid",
    "connection_name": "Okta SSO (Production)",
    "integration_name": "Okta",
    "status": "connected",
    "health": "healthy",
    "last_health_check_at": "2026-02-21T15:00:00Z",
    "last_health_error": null,
    "consecutive_failures": 0,
    "health_check_interval_mins": 60,
    "next_health_check_at": "2026-02-21T16:00:00Z",
    "sync_state": {
      "sync_enabled": true,
      "last_sync_at": "2026-02-21T12:00:00Z",
      "last_sync_status": "completed",
      "next_sync_at": "2026-02-21T18:00:00Z"
    },
    "recent_runs": [
      { "id": "uuid", "status": "completed", "started_at": "2026-02-21T12:00:00Z", "duration_ms": 3200 },
      { "id": "uuid", "status": "completed", "started_at": "2026-02-21T06:00:00Z", "duration_ms": 3100 },
      { "id": "uuid", "status": "completed", "started_at": "2026-02-21T00:00:00Z", "duration_ms": 3400 }
    ],
    "uptime": {
      "last_24h_success_rate": 100.0,
      "last_7d_success_rate": 97.6,
      "last_30d_success_rate": 98.2
    }
  }
}
```

---

#### `POST /api/v1/integration-connections/:id/health-check`

Trigger an immediate health check (lightweight credentials + connectivity test).

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "connection_id": "uuid",
    "health": "healthy",
    "check_result": "success",
    "response_time_ms": 180,
    "message": "Health check passed. Okta API responding normally.",
    "checked_at": "2026-02-21T15:35:00Z"
  }
}
```

---

### 5. Webhook Management

---

#### `GET /api/v1/integration-connections/:id/webhooks`

List webhooks for a connection.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "connection_id": "uuid",
      "name": "GitHub Push Events",
      "description": "Receive push and PR events from GitHub",
      "webhook_url": "https://api.raisinprotect.com/api/v1/webhooks/receive/uuid",
      "status": "active",
      "event_types": ["push", "pull_request"],
      "signature_header": "X-Hub-Signature-256",
      "signature_algo": "sha256",
      "total_received": 234,
      "total_processed": 231,
      "total_failed": 3,
      "last_received_at": "2026-02-21T15:20:00Z",
      "last_error": null,
      "created_at": "2026-01-20T10:00:00Z",
      "updated_at": "2026-02-21T15:20:00Z"
    }
  ]
}
```

---

#### `POST /api/v1/integration-connections/:id/webhooks`

Create a new webhook endpoint for a connection.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Request Body:**

```json
{
  "name": "GitHub Push Events",
  "description": "Receive push and PR events",
  "event_types": ["push", "pull_request"],
  "signature_header": "X-Hub-Signature-256",
  "signature_algo": "sha256"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `name` | string | Yes | 1-255 chars, unique per connection |
| `event_types` | string[] | No | Empty = accept all events |
| `signature_header` | string | No | Default: `X-Webhook-Signature` |
| `signature_algo` | string | No | `sha256` (default) or `sha1` |
| `description` | string | No | Free text |

**Response (201):**

```json
{
  "data": {
    "id": "uuid",
    "name": "GitHub Push Events",
    "webhook_url": "https://api.raisinprotect.com/api/v1/webhooks/receive/uuid",
    "webhook_secret": "whsec_abc123...",
    "status": "active",
    "message": "Webhook created. Configure your provider to send events to the webhook_url with the secret for HMAC verification. The secret is shown once — save it now."
  }
}
```

**Note:** The `webhook_secret` is only returned on creation. It cannot be retrieved afterward — only rotated.

---

#### `DELETE /api/v1/integration-connections/:id/webhooks/:wid`

Delete a webhook endpoint.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso, devops_engineer

**Response (204):** No content.

---

#### `POST /api/v1/integration-connections/:id/webhooks/:wid/rotate-secret`

Rotate the webhook secret. Returns the new secret (shown once).

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "webhook_secret": "whsec_newkey456...",
    "message": "Webhook secret rotated. Update your provider configuration with the new secret. The old secret is now invalid."
  }
}
```

---

#### `POST /api/v1/webhooks/receive/:id`

Inbound webhook receiver endpoint. This is the public-facing URL that receives events from external providers. Authenticated via HMAC signature verification (not JWT).

- **Auth:** HMAC signature in configured header
- **Roles:** N/A (external endpoint, no user auth)

**Request Headers:**
- `Content-Type: application/json`
- `X-Hub-Signature-256: sha256=...` (or configured signature header)

**Request Body:** Provider-specific event payload (JSON).

**Response (200):**

```json
{
  "status": "accepted",
  "event_id": "uuid"
}
```

**Error Codes:**
- `400` Invalid payload
- `401` Invalid signature
- `404` Webhook not found
- `410` Webhook is inactive
- `429` Rate limited

---

### 6. Dashboard & Reports

---

#### `GET /api/v1/integrations/dashboard`

Integration health dashboard with summary statistics across all connections.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Response (200):**

```json
{
  "data": {
    "summary": {
      "total_connections": 3,
      "connected": 3,
      "healthy": 2,
      "degraded": 0,
      "unhealthy": 1,
      "disabled": 0,
      "total_runs_24h": 12,
      "successful_runs_24h": 11,
      "failed_runs_24h": 1,
      "success_rate_24h": 91.7,
      "total_items_synced_24h": 523,
      "total_webhooks_received_24h": 47
    },
    "connections": [
      {
        "id": "uuid",
        "name": "Okta SSO (Production)",
        "integration_name": "Okta",
        "integration_icon_url": "/icons/integrations/okta.svg",
        "category": "identity_access",
        "status": "connected",
        "health": "healthy",
        "last_sync_at": "2026-02-21T12:00:00Z",
        "last_sync_status": "completed",
        "next_sync_at": "2026-02-21T18:00:00Z",
        "success_rate_7d": 97.6,
        "items_synced_24h": 312
      },
      {
        "id": "uuid",
        "name": "GitHub (Engineering)",
        "integration_name": "GitHub",
        "integration_icon_url": "/icons/integrations/github.svg",
        "category": "code_devops",
        "status": "connected",
        "health": "healthy",
        "last_sync_at": "2026-02-21T14:00:00Z",
        "last_sync_status": "completed",
        "next_sync_at": "2026-02-21T20:00:00Z",
        "success_rate_7d": 100.0,
        "items_synced_24h": 156
      },
      {
        "id": "uuid",
        "name": "Slack (Compliance)",
        "integration_name": "Slack",
        "integration_icon_url": "/icons/integrations/slack.svg",
        "category": "communication",
        "status": "connected",
        "health": "unhealthy",
        "last_sync_at": null,
        "last_sync_status": null,
        "next_sync_at": null,
        "success_rate_7d": 0.0,
        "items_synced_24h": 0
      }
    ],
    "recent_failures": [
      {
        "run_id": "uuid",
        "connection_name": "Slack (Compliance)",
        "integration_name": "Slack",
        "status": "failed",
        "error_message": "Slack API rate limit exceeded",
        "failed_at": "2026-02-21T13:00:00Z"
      }
    ],
    "data_freshness": {
      "fresh": 2,
      "stale": 1,
      "stale_connections": [
        {
          "connection_id": "uuid",
          "connection_name": "Slack (Compliance)",
          "integration_name": "Slack",
          "last_sync_at": null,
          "expected_interval_mins": 360
        }
      ]
    },
    "category_coverage": {
      "cloud_infrastructure": { "available": 1, "connected": 0 },
      "identity_access": { "available": 1, "connected": 1 },
      "code_devops": { "available": 1, "connected": 1 },
      "communication": { "available": 1, "connected": 1 },
      "custom": { "available": 1, "connected": 0 }
    }
  }
}
```

---

#### `GET /api/v1/integrations/dashboard/sync-activity`

Sync activity timeline for the past 24 hours, grouped by hour. For activity charts.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `hours` | int | No | Number of hours to look back (default: 24, max: 168) |
| `connection_id` | UUID | No | Filter by specific connection |

**Response (200):**

```json
{
  "data": {
    "timeline": [
      {
        "hour": "2026-02-21T15:00:00Z",
        "runs_started": 1,
        "runs_completed": 1,
        "runs_failed": 0,
        "items_synced": 156,
        "webhooks_received": 5
      },
      {
        "hour": "2026-02-21T14:00:00Z",
        "runs_started": 2,
        "runs_completed": 1,
        "runs_failed": 1,
        "items_synced": 148,
        "webhooks_received": 8
      }
    ],
    "totals": {
      "runs_started": 12,
      "runs_completed": 11,
      "runs_failed": 1,
      "items_synced": 523,
      "webhooks_received": 47
    }
  }
}
```

---

### 7. Starter Integration Previews

These endpoints provide quick access to data from specific connected integrations. They're convenience wrappers around the sync data — the actual data lives in Sprint 2-8 tables (controls, evidence, access entries, etc.) after sync processing.

---

#### `GET /api/v1/integration-connections/:id/preview`

Get a preview of data available from a connected integration. The response shape varies by integration provider.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, devops_engineer

**Response (200) — Okta example:**

```json
{
  "data": {
    "provider": "okta",
    "preview": {
      "users": {
        "total": 156,
        "active": 148,
        "deprovisioned": 8,
        "mfa_enrolled": 143,
        "mfa_enrolled_pct": 91.7
      },
      "applications": {
        "total": 23,
        "active": 20,
        "inactive": 3
      },
      "groups": {
        "total": 12
      },
      "security_policies": {
        "password_policy": "Minimum 12 chars, complexity required",
        "mfa_policy": "MFA required for all users",
        "session_lifetime_mins": 60
      }
    },
    "last_synced_at": "2026-02-21T12:00:00Z"
  }
}
```

**Response (200) — GitHub example:**

```json
{
  "data": {
    "provider": "github",
    "preview": {
      "repositories": {
        "total": 45,
        "public": 3,
        "private": 42,
        "archived": 8
      },
      "security": {
        "repos_with_branch_protection": 38,
        "repos_with_dependabot": 35,
        "repos_with_code_scanning": 20,
        "open_vulnerability_alerts": 12
      },
      "members": {
        "total": 52,
        "admins": 5,
        "outside_collaborators": 3
      }
    },
    "last_synced_at": "2026-02-21T14:00:00Z"
  }
}
```

**Response (200) — AWS Config example:**

```json
{
  "data": {
    "provider": "aws_config",
    "preview": {
      "resources": {
        "total_tracked": 847,
        "compliant": 792,
        "non_compliant": 43,
        "not_applicable": 12,
        "compliance_rate": 93.5
      },
      "rules": {
        "total": 35,
        "compliant": 28,
        "non_compliant": 7
      },
      "top_violations": [
        { "rule": "s3-bucket-server-side-encryption-enabled", "non_compliant_resources": 5 },
        { "rule": "ec2-instance-no-public-ip", "non_compliant_resources": 3 }
      ]
    },
    "last_synced_at": "2026-02-21T10:00:00Z"
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Connection not found
- `422` Connection not synced yet (no preview data available)

---

## Endpoint Summary

| # | Method | Path | Description |
|---|--------|------|-------------|
| 1 | GET | `/api/v1/integrations` | List integration catalog |
| 2 | GET | `/api/v1/integrations/:id` | Get integration definition details |
| 3 | GET | `/api/v1/integration-connections` | List org's connections |
| 4 | GET | `/api/v1/integration-connections/:id` | Get connection details |
| 5 | POST | `/api/v1/integration-connections` | Create connection |
| 6 | PUT | `/api/v1/integration-connections/:id` | Update connection |
| 7 | DELETE | `/api/v1/integration-connections/:id` | Delete connection |
| 8 | POST | `/api/v1/integration-connections/:id/test` | Test connection credentials |
| 9 | POST | `/api/v1/integration-connections/:id/enable` | Enable connection |
| 10 | POST | `/api/v1/integration-connections/:id/disable` | Disable connection |
| 11 | POST | `/api/v1/integration-connections/:id/sync` | Trigger manual sync |
| 12 | GET | `/api/v1/integration-connections/:id/runs` | List runs for connection |
| 13 | GET | `/api/v1/integration-connections/:id/runs/:rid` | Get run details |
| 14 | POST | `/api/v1/integration-connections/:id/runs/:rid/cancel` | Cancel running sync |
| 15 | GET | `/api/v1/integration-connections/:id/runs/:rid/logs` | Get run logs |
| 16 | GET | `/api/v1/integration-connections/:id/health` | Get health status |
| 17 | POST | `/api/v1/integration-connections/:id/health-check` | Trigger health check |
| 18 | GET | `/api/v1/integration-connections/:id/webhooks` | List webhooks |
| 19 | POST | `/api/v1/integration-connections/:id/webhooks` | Create webhook |
| 20 | DELETE | `/api/v1/integration-connections/:id/webhooks/:wid` | Delete webhook |
| 21 | POST | `/api/v1/integration-connections/:id/webhooks/:wid/rotate-secret` | Rotate webhook secret |
| 22 | POST | `/api/v1/webhooks/receive/:id` | Inbound webhook receiver (public) |
| 23 | GET | `/api/v1/integrations/dashboard` | Integration health dashboard |
| 24 | GET | `/api/v1/integrations/dashboard/sync-activity` | Sync activity timeline |
| 25 | GET | `/api/v1/integration-connections/:id/preview` | Integration data preview |

**Total: 25 endpoints**
