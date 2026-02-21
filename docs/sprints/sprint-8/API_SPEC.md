# Sprint 8 — API Specification: User Access Reviews

## Overview

Sprint 8 adds User Access Reviews — the governance layer for identity and access management. Endpoints cover identity provider management (connection stubs), access resource inventory, access entry tracking with anomaly detection, review campaign lifecycle, individual review decisions with approve/revoke/flag workflow, delegation and escalation, and certification reports for auditors.

This implements spec §6.2: automated access review campaigns on configurable schedules, identity provider integration, side-by-side current-vs-expected access views, reviewer assignment, one-click approve/revoke decisions with audit trail, anomaly detection, and certification reports.

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## Access Control Model

### Role Permissions Matrix

| Action | compliance_manager | security_engineer | it_admin | ciso | auditor | vendor_manager |
|--------|-------------------|-------------------|----------|------|---------|----------------|
| Manage identity providers | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ |
| View identity providers | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Manage access resources | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| View access resources | ✅ | ✅ | ✅ | ✅ | ✅* | ❌ |
| View access entries | ✅ | ✅ | ✅ | ✅ | ✅* | ❌ |
| Create campaigns | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| Manage campaigns | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| View campaigns | ✅ | ✅ | ✅ | ✅ | ✅* | ❌ |
| Submit review decision | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Delegate review | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| View certification report | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ |

\* Auditor access is read-only and limited to completed campaigns (for certification evidence).

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `identity-providers` | Per-org | Identity provider connection stubs |
| `access-resources` | Per-org | Applications/systems subject to review |
| `access-entries` | Per-org | Current access records from IdPs |
| `access-reviews/campaigns` | Per-org | Review campaign management |
| `access-reviews/campaigns/:id/reviews` | Per-campaign | Individual review decisions |
| `access-reviews/dashboard` | Per-org | Access review dashboard statistics |

---

## Endpoints

---

### 1. Identity Provider Management

---

#### `GET /api/v1/identity-providers`

List connected identity providers for the org.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | Filter by status: `pending_setup`, `connected`, `syncing`, `error`, `disconnected` |
| `provider_type` | string | No | Filter by type: `okta`, `azure_ad`, `google_workspace`, `jumpcloud`, `onelogin`, `custom` |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Okta SSO",
      "provider_type": "okta",
      "status": "connected",
      "description": "Primary identity provider — all employees",
      "last_sync_at": "2026-02-21T06:04:00Z",
      "last_sync_status": "success",
      "last_sync_stats": {
        "users_synced": 156,
        "resources_synced": 12,
        "duration_ms": 3200
      },
      "sync_interval_mins": 360,
      "total_resources": 12,
      "total_entries": 487,
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-02-21T06:04:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 2,
    "total_pages": 1
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden (auditor, vendor_manager)

---

#### `GET /api/v1/identity-providers/:id`

Get details of a specific identity provider, including config (with secrets masked).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "name": "Okta SSO",
    "provider_type": "okta",
    "status": "connected",
    "description": "Primary identity provider — all employees",
    "config": {
      "domain": "acme.okta.com",
      "api_token": "***masked***"
    },
    "last_sync_at": "2026-02-21T06:04:00Z",
    "last_sync_status": "success",
    "last_sync_error": null,
    "last_sync_stats": {
      "users_synced": 156,
      "resources_synced": 12,
      "duration_ms": 3200
    },
    "sync_interval_mins": 360,
    "total_resources": 12,
    "total_entries": 487,
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-02-21T06:04:00Z"
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Identity provider not found

---

#### `POST /api/v1/identity-providers`

Connect a new identity provider.

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Request Body:**

```json
{
  "name": "Okta SSO",
  "provider_type": "okta",
  "config": {
    "domain": "acme.okta.com",
    "api_token": "00abcdef..."
  },
  "description": "Primary identity provider",
  "sync_interval_mins": 360
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `name` | string | Yes | 1-255 chars, unique per org |
| `provider_type` | string | Yes | One of: `okta`, `azure_ad`, `google_workspace`, `jumpcloud`, `onelogin`, `custom` |
| `config` | object | Yes | Provider-specific configuration (see below) |
| `description` | string | No | Free text |
| `sync_interval_mins` | integer | No | 15-10080, default 360 |

**Provider Config Requirements:**

| Provider | Required Fields |
|----------|----------------|
| `okta` | `domain`, `api_token` |
| `azure_ad` | `tenant_id`, `client_id`, `client_secret` |
| `google_workspace` | `domain`, `service_account_key` |
| `jumpcloud` | `api_key`, `org_id` |
| `onelogin` | `subdomain`, `client_id`, `client_secret` |
| `custom` | `base_url`, `auth_type` (`api_key` or `oauth2`), auth credentials |

**Response (201):**

```json
{
  "data": {
    "id": "uuid",
    "name": "Okta SSO",
    "provider_type": "okta",
    "status": "pending_setup",
    "...": "..."
  }
}
```

**Error Codes:**
- `400` Invalid config for provider type
- `401` Unauthorized
- `403` Forbidden (only it_admin, ciso)
- `409` Provider name already exists

---

#### `PUT /api/v1/identity-providers/:id`

Update identity provider configuration.

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Request Body:**

```json
{
  "name": "Okta SSO (Production)",
  "config": {
    "domain": "acme.okta.com",
    "api_token": "00newtoken..."
  },
  "sync_interval_mins": 180,
  "description": "Updated description"
}
```

All fields optional. Config is merged (not replaced) — send only changed fields.

**Response (200):** Updated identity provider object.

**Error Codes:**
- `400` Invalid config
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `409` Name conflict

---

#### `DELETE /api/v1/identity-providers/:id`

Disconnect and remove an identity provider. Access resources and entries from this provider will have their `identity_provider_id` set to NULL (orphaned but preserved for review).

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Response (204):** No content.

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `409` Cannot delete while sync is in progress

---

#### `POST /api/v1/identity-providers/:id/sync`

Trigger a manual sync of access data from the identity provider.

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Request Body:** (optional)

```json
{
  "full_sync": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|------------|
| `full_sync` | boolean | No | If true, re-sync all data (not just changes). Default: false (incremental). |

**Response (202):**

```json
{
  "data": {
    "sync_id": "uuid",
    "status": "started",
    "message": "Sync started. Check GET /identity-providers/:id/sync-status for progress."
  }
}
```

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `409` Sync already in progress
- `422` Provider not in `connected` status

---

#### `GET /api/v1/identity-providers/:id/sync-status`

Get the current sync status for an identity provider.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Response (200):**

```json
{
  "data": {
    "provider_id": "uuid",
    "status": "connected",
    "is_syncing": false,
    "last_sync_at": "2026-02-21T06:04:00Z",
    "last_sync_status": "success",
    "last_sync_error": null,
    "last_sync_stats": {
      "users_synced": 156,
      "resources_synced": 12,
      "new_entries": 3,
      "updated_entries": 12,
      "removed_entries": 1,
      "anomalies_detected": 5,
      "duration_ms": 3200
    },
    "next_sync_at": "2026-02-21T12:04:00Z"
  }
}
```

---

### 2. Access Resource Management

---

#### `GET /api/v1/access-resources`

List access resources (applications/systems) with filtering.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (read-only)

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `criticality` | string | No | Filter: `critical`, `high`, `medium`, `low` |
| `resource_type` | string | No | Filter: `application`, `infrastructure`, `directory`, `custom` |
| `provider_id` | UUID | No | Filter by identity provider |
| `department` | string | No | Filter by owning department |
| `is_active` | boolean | No | Filter by active status (default: true) |
| `search` | string | No | Search name and description |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort: `name`, `criticality`, `total_users`, `last_reviewed_at`, `created_at` |
| `order` | string | No | `asc`, `desc` (default: `asc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "GitHub Organization",
      "description": "Source code repositories and CI/CD",
      "resource_type": "application",
      "criticality": "high",
      "department": "Engineering",
      "category": "Developer Tools",
      "tags": ["source-code", "ci-cd"],
      "owner_id": "uuid",
      "owner_name": "Alice Engineer",
      "identity_provider_id": "uuid",
      "identity_provider_name": "Okta SSO",
      "total_users": 45,
      "total_roles": 4,
      "url": "https://github.com/acme-corp",
      "last_sync_at": "2026-02-21T06:04:00Z",
      "last_reviewed_at": "2025-12-15T00:00:00Z",
      "review_cadence": "quarterly",
      "is_active": true,
      "anomaly_count": 3,
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-02-21T06:04:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 6, "total_pages": 1 }
}
```

---

#### `GET /api/v1/access-resources/:id`

Get detailed information about a specific access resource, including summary statistics and recent anomalies.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (read-only)

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "name": "GitHub Organization",
    "description": "Source code repositories and CI/CD",
    "resource_type": "application",
    "criticality": "high",
    "department": "Engineering",
    "category": "Developer Tools",
    "tags": ["source-code", "ci-cd"],
    "owner_id": "uuid",
    "owner_name": "Alice Engineer",
    "identity_provider_id": "uuid",
    "identity_provider_name": "Okta SSO",
    "external_id": "0oa1234567890",
    "total_users": 45,
    "total_roles": 4,
    "url": "https://github.com/acme-corp",
    "last_sync_at": "2026-02-21T06:04:00Z",
    "last_reviewed_at": "2025-12-15T00:00:00Z",
    "review_cadence": "quarterly",
    "is_active": true,
    "stats": {
      "active_entries": 42,
      "inactive_entries": 3,
      "privileged_entries": 5,
      "service_accounts": 2,
      "entries_with_anomalies": 3,
      "anomaly_breakdown": {
        "stale_access": 1,
        "excessive_privileges": 1,
        "no_mfa": 1
      },
      "entries_with_drift": 2,
      "departments": ["Engineering", "DevOps", "Security"]
    },
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-02-21T06:04:00Z"
  }
}
```

---

#### `POST /api/v1/access-resources`

Create an access resource manually (not synced from an IdP).

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso

**Request Body:**

```json
{
  "name": "Internal Wiki",
  "description": "Confluence instance for internal documentation",
  "resource_type": "application",
  "criticality": "medium",
  "department": "Engineering",
  "category": "Collaboration",
  "tags": ["documentation", "internal"],
  "owner_id": "uuid",
  "url": "https://wiki.acme.com",
  "review_cadence": "semi_annual"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `name` | string | Yes | 1-255 chars |
| `resource_type` | string | No | Default: `application` |
| `criticality` | string | No | Default: `medium` |
| `department` | string | No | Free text |
| `category` | string | No | Free text |
| `tags` | string[] | No | Array of tags |
| `owner_id` | UUID | No | Must be active user in org |
| `url` | string | No | Valid URL |
| `review_cadence` | string | No | Campaign cadence enum |
| `description` | string | No | Free text |

**Response (201):** Created resource object.

**Error Codes:**
- `400` Validation error
- `401` Unauthorized
- `403` Forbidden
- `409` Resource name already exists for this org

---

#### `PUT /api/v1/access-resources/:id`

Update an access resource.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso

**Request Body:** Same fields as POST, all optional.

**Response (200):** Updated resource object.

**Error Codes:**
- `400` Validation error
- `401` Unauthorized
- `403` Forbidden
- `404` Not found

---

#### `DELETE /api/v1/access-resources/:id`

Delete an access resource and all associated access entries. Cannot delete if resource has active (non-completed) review campaigns.

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Response (204):** No content.

**Error Codes:**
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `409` Resource has entries in active review campaigns

---

### 3. Access Entry Management

---

#### `GET /api/v1/access-entries`

List access entries with comprehensive filtering. Supports the side-by-side comparison view from spec §6.2.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (read-only, completed campaigns only)

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `resource_id` | UUID | No | Filter by resource |
| `provider_id` | UUID | No | Filter by identity provider |
| `status` | string | No | Filter: `active`, `inactive`, `orphaned`, `suspended`, `pending_revocation` |
| `user_email` | string | No | Filter by user email (partial match) |
| `department` | string | No | Filter by user department |
| `is_privileged` | boolean | No | Filter for privileged/admin access |
| `has_role_drift` | boolean | No | Filter for entries with role drift |
| `has_anomalies` | boolean | No | Filter for entries with detected anomalies |
| `anomaly_type` | string | No | Filter by anomaly type: `orphaned_account`, `excessive_privileges`, `role_drift`, `stale_access`, `no_mfa`, `departed_user` |
| `is_service_account` | boolean | No | Filter for service accounts |
| `search` | string | No | Search user_email, user_display_name, role_name |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort: `user_email`, `resource_name`, `role_name`, `last_used_at`, `granted_at`, `created_at` |
| `order` | string | No | `asc`, `desc` (default: `asc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "resource_id": "uuid",
      "resource_name": "GitHub Organization",
      "resource_criticality": "high",
      "identity_provider_id": "uuid",
      "identity_provider_name": "Okta SSO",
      "user_email": "alice@acme.com",
      "user_display_name": "Alice Engineer",
      "user_department": "Engineering",
      "user_title": "Senior Engineer",
      "user_manager_email": "bob@acme.com",
      "internal_user_id": "uuid",
      "role_name": "Admin",
      "access_level": "admin",
      "is_privileged": true,
      "expected_role": "Write",
      "expected_access_level": "write",
      "has_role_drift": true,
      "granted_at": "2025-06-15T10:00:00Z",
      "last_used_at": "2026-02-20T14:30:00Z",
      "last_login_at": "2026-02-20T09:00:00Z",
      "status": "active",
      "mfa_enabled": true,
      "is_service_account": false,
      "anomalies": [
        {
          "type": "excessive_privileges",
          "detected_at": "2026-02-21T06:04:00Z",
          "details": "User has Admin role but expected Write based on Senior Engineer title"
        }
      ],
      "last_sync_at": "2026-02-21T06:04:00Z",
      "created_at": "2025-06-15T10:00:00Z",
      "updated_at": "2026-02-21T06:04:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 487, "total_pages": 25 }
}
```

---

#### `GET /api/v1/access-entries/:id`

Get detailed access entry with full context (resource, provider, anomalies, review history).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "resource_id": "uuid",
    "resource_name": "GitHub Organization",
    "resource_criticality": "high",
    "identity_provider_id": "uuid",
    "identity_provider_name": "Okta SSO",
    "external_user_id": "00u1234567890",
    "user_email": "alice@acme.com",
    "user_display_name": "Alice Engineer",
    "user_department": "Engineering",
    "user_title": "Senior Engineer",
    "user_manager_email": "bob@acme.com",
    "internal_user_id": "uuid",
    "role_name": "Admin",
    "access_level": "admin",
    "permissions": {
      "repos": "all",
      "teams": "manage",
      "billing": "read",
      "members": "manage"
    },
    "is_privileged": true,
    "expected_role": "Write",
    "expected_access_level": "write",
    "has_role_drift": true,
    "granted_at": "2025-06-15T10:00:00Z",
    "last_used_at": "2026-02-20T14:30:00Z",
    "last_login_at": "2026-02-20T09:00:00Z",
    "status": "active",
    "mfa_enabled": true,
    "is_service_account": false,
    "anomalies": [
      {
        "type": "excessive_privileges",
        "detected_at": "2026-02-21T06:04:00Z",
        "details": "User has Admin role but expected Write based on Senior Engineer title"
      }
    ],
    "review_history": [
      {
        "campaign_id": "uuid",
        "campaign_name": "Q4 2025 Quarterly Access Review",
        "decision": "approved",
        "decided_by_name": "Bob Manager",
        "decided_at": "2025-12-15T14:00:00Z",
        "justification": "Confirmed admin access needed for CI/CD management"
      }
    ],
    "last_sync_at": "2026-02-21T06:04:00Z",
    "notes": null,
    "created_at": "2025-06-15T10:00:00Z",
    "updated_at": "2026-02-21T06:04:00Z"
  }
}
```

---

#### `GET /api/v1/access-entries/anomalies`

Get aggregated anomaly summary across all access entries. Used for the anomaly detection dashboard.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `resource_id` | UUID | No | Filter by resource |
| `department` | string | No | Filter by user department |
| `criticality` | string | No | Filter by resource criticality |

**Response (200):**

```json
{
  "data": {
    "total_entries": 487,
    "entries_with_anomalies": 23,
    "anomaly_breakdown": [
      {
        "type": "stale_access",
        "count": 8,
        "description": "Access not used in 90+ days",
        "severity": "medium",
        "entries": [
          {
            "entry_id": "uuid",
            "user_email": "departed@acme.com",
            "resource_name": "AWS Console",
            "role_name": "PowerUser",
            "last_used_at": "2025-10-01T00:00:00Z",
            "days_stale": 143
          }
        ]
      },
      {
        "type": "orphaned_account",
        "count": 3,
        "description": "No matching employee record found",
        "severity": "high",
        "entries": [
          {
            "entry_id": "uuid",
            "user_email": "unknown@acme.com",
            "resource_name": "Production Database",
            "role_name": "Admin",
            "detected_at": "2026-02-21T06:04:00Z"
          }
        ]
      },
      {
        "type": "excessive_privileges",
        "count": 5,
        "description": "More access than role requires",
        "severity": "high",
        "entries": []
      },
      {
        "type": "role_drift",
        "count": 4,
        "description": "Actual access diverges from expected",
        "severity": "medium",
        "entries": []
      },
      {
        "type": "no_mfa",
        "count": 2,
        "description": "Account lacks MFA on critical resource",
        "severity": "critical",
        "entries": []
      },
      {
        "type": "departed_user",
        "count": 1,
        "description": "User marked as departed in HRIS",
        "severity": "critical",
        "entries": []
      }
    ],
    "by_resource_criticality": {
      "critical": 8,
      "high": 9,
      "medium": 4,
      "low": 2
    },
    "by_department": {
      "Engineering": 10,
      "Finance": 5,
      "Marketing": 4,
      "HR": 2,
      "Unknown": 2
    }
  }
}
```

---

#### `POST /api/v1/access-entries/detect-anomalies`

Trigger anomaly detection across all access entries (or a subset). Analyzes entries against rules: stale access (90+ days no use), orphaned accounts (no matching employee), excessive privileges (actual > expected role), role drift (actual ≠ expected), no MFA on critical resources, departed users.

- **Auth:** Bearer token required
- **Roles:** security_engineer, it_admin, ciso

**Request Body:** (optional)

```json
{
  "resource_ids": ["uuid1", "uuid2"],
  "stale_days_threshold": 90,
  "include_service_accounts": false
}
```

**Response (202):**

```json
{
  "data": {
    "status": "started",
    "message": "Anomaly detection started for 487 access entries",
    "entries_scanned": 487,
    "anomalies_detected": 23,
    "new_anomalies": 3,
    "resolved_anomalies": 1
  }
}
```

---

### 4. Access Review Campaigns

---

#### `GET /api/v1/access-reviews/campaigns`

List access review campaigns with filtering and pagination.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (completed only)

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | Filter: `draft`, `active`, `in_review`, `completed`, `cancelled` |
| `cadence` | string | No | Filter: `monthly`, `quarterly`, `semi_annual`, `annual`, `custom` |
| `search` | string | No | Search name and description |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort: `name`, `deadline`, `status`, `created_at`, `total_reviews` |
| `order` | string | No | `asc`, `desc` (default: `desc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Q1 2026 Quarterly Access Review",
      "description": "Quarterly review of all critical and high-criticality resources",
      "status": "in_review",
      "cadence": "quarterly",
      "reviewer_strategy": "resource_owner",
      "deadline": "2026-03-07T00:00:00Z",
      "started_at": "2026-02-21T10:00:00Z",
      "completed_at": null,
      "total_reviews": 45,
      "completed_reviews": 27,
      "completion_pct": 60.0,
      "approved_count": 22,
      "revoked_count": 3,
      "flagged_count": 2,
      "pending_count": 18,
      "escalated_count": 2,
      "overdue": false,
      "days_remaining": 14,
      "created_by_name": "John Compliance",
      "tags": ["quarterly", "q1-2026"],
      "created_at": "2026-02-14T09:00:00Z",
      "updated_at": "2026-02-21T08:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

---

#### `GET /api/v1/access-reviews/campaigns/:id`

Get detailed campaign information including scope, escalation config, and summary statistics.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (completed only)

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "name": "Q1 2026 Quarterly Access Review",
    "description": "Quarterly review of all critical and high-criticality resources",
    "status": "in_review",
    "cadence": "quarterly",
    "scope": {
      "resource_criticalities": ["critical", "high"],
      "include_privileged_only": false,
      "include_service_accounts": true,
      "stale_days_threshold": 90
    },
    "reviewer_strategy": "resource_owner",
    "default_reviewer_id": "uuid",
    "default_reviewer_name": "John Compliance",
    "deadline": "2026-03-07T00:00:00Z",
    "started_at": "2026-02-21T10:00:00Z",
    "completed_at": null,
    "escalation_config": {
      "reminder_days_before_deadline": [7, 3, 1],
      "escalate_after_days": 5,
      "escalation_recipient_id": "uuid",
      "auto_expire_days": null
    },
    "total_reviews": 45,
    "completed_reviews": 27,
    "completion_pct": 60.0,
    "approved_count": 22,
    "revoked_count": 3,
    "flagged_count": 2,
    "stats": {
      "by_resource": [
        {
          "resource_id": "uuid",
          "resource_name": "GitHub Organization",
          "total": 15,
          "completed": 10,
          "approved": 8,
          "revoked": 1,
          "flagged": 1,
          "pending": 5
        }
      ],
      "by_reviewer": [
        {
          "reviewer_id": "uuid",
          "reviewer_name": "Alice Engineer",
          "total": 12,
          "completed": 8,
          "pending": 4,
          "escalated": 1
        }
      ],
      "by_department": {
        "Engineering": { "total": 20, "completed": 14 },
        "Finance": { "total": 10, "completed": 7 },
        "DevOps": { "total": 15, "completed": 6 }
      },
      "anomalies_in_scope": {
        "total": 8,
        "stale_access": 3,
        "excessive_privileges": 2,
        "role_drift": 2,
        "orphaned_account": 1
      }
    },
    "created_by": "uuid",
    "created_by_name": "John Compliance",
    "tags": ["quarterly", "q1-2026"],
    "notes": null,
    "created_at": "2026-02-14T09:00:00Z",
    "updated_at": "2026-02-21T08:00:00Z"
  }
}
```

---

#### `POST /api/v1/access-reviews/campaigns`

Create a new access review campaign.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, it_admin, ciso

**Request Body:**

```json
{
  "name": "Q1 2026 Quarterly Access Review",
  "description": "Quarterly review of all critical and high-criticality resources",
  "cadence": "quarterly",
  "scope": {
    "resource_ids": [],
    "resource_criticalities": ["critical", "high"],
    "resource_types": [],
    "departments": [],
    "include_privileged_only": false,
    "include_service_accounts": true,
    "stale_days_threshold": 90
  },
  "reviewer_strategy": "resource_owner",
  "default_reviewer_id": "uuid",
  "deadline": "2026-03-07T00:00:00Z",
  "escalation_config": {
    "reminder_days_before_deadline": [7, 3, 1],
    "escalate_after_days": 5,
    "escalation_recipient_id": "uuid"
  },
  "tags": ["quarterly", "q1-2026"],
  "notes": "Focus on engineering resources this quarter"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `name` | string | Yes | 1-255 chars |
| `cadence` | string | Yes | Campaign cadence enum |
| `scope` | object | Yes | At least one scope filter required |
| `reviewer_strategy` | string | No | Default: `resource_owner` |
| `default_reviewer_id` | UUID | No | Fallback reviewer (recommended) |
| `deadline` | datetime | Yes | Must be in the future |
| `escalation_config` | object | No | Escalation settings |
| `description` | string | No | Free text |
| `tags` | string[] | No | Array of tags |
| `notes` | string | No | Internal notes |

**Response (201):**

```json
{
  "data": {
    "id": "uuid",
    "name": "Q1 2026 Quarterly Access Review",
    "status": "draft",
    "in_scope_entries": 45,
    "message": "Campaign created in draft status. Use POST /campaigns/:id/launch to start.",
    "...": "..."
  }
}
```

**Error Codes:**
- `400` Validation error (empty scope, past deadline)
- `401` Unauthorized
- `403` Forbidden
- `409` Campaign name already exists
- `422` No access entries match the scope criteria

---

#### `PUT /api/v1/access-reviews/campaigns/:id`

Update a campaign. Only `draft` campaigns can be fully edited. `active`/`in_review` campaigns allow updating deadline, escalation_config, default_reviewer_id, description, notes, and tags.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, it_admin, ciso

**Request Body:** Same fields as POST, all optional.

**Response (200):** Updated campaign object.

**Error Codes:**
- `400` Validation error
- `401` Unauthorized
- `403` Forbidden (or campaign is completed/cancelled)
- `404` Not found
- `422` Cannot change scope on active/in_review campaign

---

#### `POST /api/v1/access-reviews/campaigns/:id/launch`

Launch a draft campaign. This:
1. Queries access entries matching the campaign scope
2. Creates `access_review` records for each in-scope entry
3. Assigns reviewers per the `reviewer_strategy`
4. Snapshots current access state into `access_snapshot`
5. Transitions campaign to `active` status

- **Auth:** Bearer token required
- **Roles:** compliance_manager, it_admin, ciso

**Request Body:** (optional)

```json
{
  "notify_reviewers": true
}
```

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "status": "active",
    "started_at": "2026-02-21T10:00:00Z",
    "total_reviews": 45,
    "reviewers_assigned": 8,
    "unassigned_reviews": 0,
    "message": "Campaign launched. 45 reviews created, 8 reviewers notified."
  }
}
```

**Error Codes:**
- `400` Campaign is not in `draft` status
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `422` No entries match scope (campaign would be empty)

---

#### `POST /api/v1/access-reviews/campaigns/:id/complete`

Complete an active/in_review campaign. All pending reviews are marked as `expired`. Campaign transitions to `completed` status.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:** (optional)

```json
{
  "expire_pending": true,
  "completion_notes": "Completed with 3 pending reviews auto-expired."
}
```

| Field | Type | Required | Description |
|-------|------|----------|------------|
| `expire_pending` | boolean | No | If true, auto-expire pending reviews. If false, reject if reviews still pending. Default: true. |
| `completion_notes` | string | No | Notes about campaign completion |

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "status": "completed",
    "completed_at": "2026-03-07T10:00:00Z",
    "total_reviews": 45,
    "completed_reviews": 42,
    "expired_reviews": 3,
    "approved_count": 35,
    "revoked_count": 5,
    "flagged_count": 2,
    "message": "Campaign completed. 3 pending reviews expired."
  }
}
```

**Error Codes:**
- `400` Campaign is not in `active` or `in_review` status
- `401` Unauthorized
- `403` Forbidden
- `404` Not found
- `422` `expire_pending` is false and there are still pending reviews

---

#### `POST /api/v1/access-reviews/campaigns/:id/cancel`

Cancel a campaign. All pending reviews are discarded. Only `draft`, `active`, or `in_review` campaigns can be cancelled.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:**

```json
{
  "reason": "Scope needs to be revised to include new resources"
}
```

| Field | Type | Required | Description |
|-------|------|----------|------------|
| `reason` | string | Yes | Cancellation reason (required for audit trail) |

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "status": "cancelled",
    "cancelled_at": "2026-02-25T10:00:00Z",
    "message": "Campaign cancelled. All pending reviews discarded."
  }
}
```

**Error Codes:**
- `400` Campaign already completed or cancelled
- `401` Unauthorized
- `403` Forbidden
- `404` Not found

---

#### `GET /api/v1/access-reviews/campaigns/:id/stats`

Get detailed statistics for a campaign, including per-resource breakdown, reviewer workload, decision distribution, and timeline analysis.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (completed campaigns only)

**Response (200):**

```json
{
  "data": {
    "campaign_id": "uuid",
    "campaign_name": "Q1 2026 Quarterly Access Review",
    "status": "in_review",
    "completion_pct": 60.0,
    "timeline": {
      "started_at": "2026-02-21T10:00:00Z",
      "deadline": "2026-03-07T00:00:00Z",
      "days_elapsed": 7,
      "days_remaining": 14,
      "is_overdue": false,
      "avg_decision_time_hours": 18.5
    },
    "decisions": {
      "total": 45,
      "approved": 22,
      "revoked": 3,
      "flagged": 2,
      "delegated": 1,
      "expired": 0,
      "pending": 17
    },
    "by_resource": [
      {
        "resource_id": "uuid",
        "resource_name": "Production Database",
        "resource_criticality": "critical",
        "total": 8,
        "approved": 5,
        "revoked": 2,
        "pending": 1,
        "anomalies": 3
      }
    ],
    "by_reviewer": [
      {
        "reviewer_id": "uuid",
        "reviewer_name": "Alice Engineer",
        "assigned": 12,
        "completed": 8,
        "pending": 4,
        "avg_decision_time_hours": 12.3,
        "is_escalated": false
      }
    ],
    "by_department": {
      "Engineering": { "total": 20, "approved": 14, "revoked": 2, "pending": 4 },
      "Finance": { "total": 10, "approved": 5, "revoked": 1, "pending": 4 }
    },
    "anomaly_impact": {
      "entries_with_anomalies": 8,
      "revoked_with_anomalies": 3,
      "anomaly_revocation_rate": 37.5
    },
    "privileged_access": {
      "total_privileged_reviews": 12,
      "approved": 8,
      "revoked": 2,
      "flagged": 1,
      "pending": 1
    }
  }
}
```

---

### 5. Individual Access Reviews

---

#### `GET /api/v1/access-reviews/campaigns/:id/reviews`

List individual access reviews within a campaign with filtering.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (completed campaigns only)

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `decision` | string | No | Filter: `pending`, `approved`, `revoked`, `flagged`, `delegated`, `expired` |
| `reviewer_id` | UUID | No | Filter by assigned reviewer |
| `resource_id` | UUID | No | Filter by resource |
| `is_privileged` | boolean | No | Filter by privileged access (from snapshot) |
| `is_escalated` | boolean | No | Filter for escalated reviews |
| `has_anomalies` | boolean | No | Filter for reviews with anomalies (from snapshot) |
| `department` | string | No | Filter by user department (from snapshot) |
| `search` | string | No | Search user_email, resource_name from snapshot |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort: `user_email`, `resource_name`, `decision`, `decided_at`, `created_at` |
| `order` | string | No | `asc`, `desc` (default: `asc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "campaign_id": "uuid",
      "entry_id": "uuid",
      "reviewer_id": "uuid",
      "reviewer_name": "Alice Engineer",
      "assigned_at": "2026-02-21T10:00:00Z",
      "decision": "pending",
      "justification": null,
      "decided_by": null,
      "decided_at": null,
      "is_escalated": false,
      "delegated_to_id": null,
      "access_snapshot": {
        "resource_name": "GitHub Organization",
        "resource_criticality": "high",
        "user_email": "bob@acme.com",
        "user_department": "Engineering",
        "user_title": "Junior Developer",
        "role_name": "Admin",
        "access_level": "admin",
        "is_privileged": true,
        "expected_role": "Write",
        "has_role_drift": true,
        "granted_at": "2025-06-15T10:00:00Z",
        "last_used_at": "2026-02-20T14:30:00Z",
        "anomalies": [
          {
            "type": "excessive_privileges",
            "detected_at": "2026-02-21T06:04:00Z",
            "details": "User has Admin role but expected Write"
          }
        ]
      },
      "revocation_executed": false,
      "notes": null,
      "created_at": "2026-02-21T10:00:00Z",
      "updated_at": "2026-02-21T10:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 45, "total_pages": 3 }
}
```

---

#### `GET /api/v1/access-reviews/campaigns/:id/reviews/:rid`

Get detailed information about a specific access review, including full access snapshot and audit trail.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (completed campaigns only)

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "campaign_id": "uuid",
    "campaign_name": "Q1 2026 Quarterly Access Review",
    "entry_id": "uuid",
    "reviewer_id": "uuid",
    "reviewer_name": "Alice Engineer",
    "assigned_at": "2026-02-21T10:00:00Z",
    "decision": "revoked",
    "justification": "User is a Junior Developer with Admin access. Downgrade to Write per least privilege principle.",
    "decided_by": "uuid",
    "decided_by_name": "Alice Engineer",
    "decided_at": "2026-02-22T14:30:00Z",
    "delegated_to_id": null,
    "delegated_to_name": null,
    "delegated_at": null,
    "delegation_reason": null,
    "is_escalated": false,
    "escalated_at": null,
    "escalated_to_id": null,
    "escalated_to_name": null,
    "access_snapshot": {
      "resource_name": "GitHub Organization",
      "resource_criticality": "high",
      "user_email": "bob@acme.com",
      "user_display_name": "Bob Developer",
      "user_department": "Engineering",
      "user_title": "Junior Developer",
      "user_manager_email": "alice@acme.com",
      "role_name": "Admin",
      "access_level": "admin",
      "is_privileged": true,
      "expected_role": "Write",
      "expected_access_level": "write",
      "has_role_drift": true,
      "granted_at": "2025-06-15T10:00:00Z",
      "last_used_at": "2026-02-20T14:30:00Z",
      "last_login_at": "2026-02-20T09:00:00Z",
      "mfa_enabled": true,
      "anomalies": [
        {
          "type": "excessive_privileges",
          "detected_at": "2026-02-21T06:04:00Z",
          "details": "User has Admin role but expected Write based on Junior Developer title"
        }
      ]
    },
    "revocation_executed": false,
    "revocation_executed_at": null,
    "revocation_notes": null,
    "notes": null,
    "created_at": "2026-02-21T10:00:00Z",
    "updated_at": "2026-02-22T14:30:00Z"
  }
}
```

---

#### `POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/decide`

Submit a decision on an access review. Justification is mandatory for `revoked` and `flagged` decisions.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso
- **Additional check:** User must be the assigned reviewer, or have compliance_manager/ciso role (override capability)

**Request Body:**

```json
{
  "decision": "revoked",
  "justification": "User is a Junior Developer with Admin access. Downgrade to Write per least privilege principle.",
  "notes": "Recommended downgrade to Write role in GitHub"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `decision` | string | Yes | `approved`, `revoked`, `flagged` |
| `justification` | string | Conditional | Required for `revoked` and `flagged` |
| `notes` | string | No | Additional notes |

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "decision": "revoked",
    "justification": "User is a Junior Developer with Admin access...",
    "decided_by": "uuid",
    "decided_by_name": "Alice Engineer",
    "decided_at": "2026-02-22T14:30:00Z",
    "campaign_progress": {
      "total_reviews": 45,
      "completed_reviews": 28,
      "completion_pct": 62.2
    }
  }
}
```

**Error Codes:**
- `400` Invalid decision or missing justification
- `401` Unauthorized
- `403` Not assigned reviewer and not compliance_manager/ciso
- `404` Review not found
- `409` Review already decided
- `422` Campaign is not in `active` or `in_review` status

---

#### `POST /api/v1/access-reviews/campaigns/:id/reviews/bulk-decide`

Submit decisions for multiple reviews at once (batch approve is common).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Request Body:**

```json
{
  "reviews": [
    {
      "review_id": "uuid1",
      "decision": "approved",
      "justification": "Access confirmed appropriate"
    },
    {
      "review_id": "uuid2",
      "decision": "approved",
      "justification": "Access confirmed appropriate"
    },
    {
      "review_id": "uuid3",
      "decision": "revoked",
      "justification": "User no longer needs access to this resource"
    }
  ]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `reviews` | array | Yes | 1-100 review decisions |
| `reviews[].review_id` | UUID | Yes | Must belong to this campaign |
| `reviews[].decision` | string | Yes | `approved`, `revoked`, `flagged` |
| `reviews[].justification` | string | Conditional | Required for `revoked` and `flagged` |

**Response (200):**

```json
{
  "data": {
    "processed": 3,
    "succeeded": 3,
    "failed": 0,
    "results": [
      { "review_id": "uuid1", "status": "success", "decision": "approved" },
      { "review_id": "uuid2", "status": "success", "decision": "approved" },
      { "review_id": "uuid3", "status": "success", "decision": "revoked" }
    ],
    "campaign_progress": {
      "total_reviews": 45,
      "completed_reviews": 30,
      "completion_pct": 66.7
    }
  }
}
```

**Error Codes:**
- `400` Empty reviews array or exceeds 100
- `401` Unauthorized
- `403` Forbidden
- `422` Campaign not active/in_review, or some reviews already decided (partial success possible)

---

#### `POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/delegate`

Delegate a review to another user. The original reviewer is preserved in the audit trail.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Request Body:**

```json
{
  "delegate_to_id": "uuid",
  "reason": "I'm not familiar with this resource. Delegating to the team lead who manages it."
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `delegate_to_id` | UUID | Yes | Must be active user in org with review-capable role |
| `reason` | string | Yes | Delegation reason (required for audit trail) |

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "decision": "delegated",
    "delegated_to_id": "uuid",
    "delegated_to_name": "Charlie TeamLead",
    "delegated_at": "2026-02-23T09:00:00Z",
    "delegation_reason": "I'm not familiar with this resource...",
    "message": "Review delegated to Charlie TeamLead. They will receive a notification."
  }
}
```

**Error Codes:**
- `400` Missing delegate_to_id or reason
- `401` Unauthorized
- `403` Not assigned reviewer and not compliance_manager/ciso
- `404` Review or delegate user not found
- `409` Review already decided (cannot delegate a decided review)
- `422` Cannot delegate to self, or delegate user is not active

---

#### `POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/escalate`

Manually escalate a review. Typically triggered automatically by the escalation rules, but can be done manually by campaign managers.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:**

```json
{
  "escalate_to_id": "uuid",
  "reason": "Reviewer has not responded for 5 days. Escalating to department head."
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `escalate_to_id` | UUID | Yes | Must be active user in org |
| `reason` | string | Yes | Escalation reason |

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "is_escalated": true,
    "escalated_at": "2026-02-26T10:00:00Z",
    "escalated_to_id": "uuid",
    "escalated_to_name": "Diana Director",
    "reviewer_id": "uuid",
    "reviewer_name": "Diana Director",
    "message": "Review escalated to Diana Director. Original reviewer (Alice Engineer) preserved in audit trail."
  }
}
```

**Error Codes:**
- `400` Missing fields
- `401` Unauthorized
- `403` Forbidden (only compliance_manager, ciso)
- `404` Not found
- `409` Review already decided
- `422` Review already escalated to this user

---

#### `POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/revocation`

Mark a revoked review's revocation as executed (access actually removed in the identity provider).

- **Auth:** Bearer token required
- **Roles:** it_admin, ciso

**Request Body:**

```json
{
  "executed": true,
  "notes": "Admin access removed from GitHub. User downgraded to Write role."
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `executed` | boolean | Yes | Must be true |
| `notes` | string | No | Execution details |

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "decision": "revoked",
    "revocation_executed": true,
    "revocation_executed_at": "2026-02-28T14:00:00Z",
    "revocation_notes": "Admin access removed from GitHub..."
  }
}
```

**Error Codes:**
- `400` Review decision is not `revoked`
- `401` Unauthorized
- `403` Forbidden (only it_admin, ciso)
- `404` Not found
- `409` Revocation already executed

---

### 6. Dashboard & Reports

---

#### `GET /api/v1/access-reviews/dashboard`

Access review dashboard with summary statistics across all campaigns, resources, and anomalies.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Response (200):**

```json
{
  "data": {
    "summary": {
      "total_identity_providers": 2,
      "connected_providers": 2,
      "total_resources": 6,
      "total_entries": 487,
      "entries_with_anomalies": 23,
      "privileged_entries": 45,
      "active_campaigns": 1,
      "pending_reviews": 18,
      "pending_revocations": 3
    },
    "active_campaigns": [
      {
        "id": "uuid",
        "name": "Q1 2026 Quarterly Access Review",
        "status": "in_review",
        "completion_pct": 60.0,
        "deadline": "2026-03-07T00:00:00Z",
        "days_remaining": 14,
        "pending_reviews": 18,
        "escalated_reviews": 2
      }
    ],
    "anomaly_summary": {
      "total": 23,
      "critical": 3,
      "high": 8,
      "medium": 12,
      "by_type": {
        "orphaned_account": 3,
        "excessive_privileges": 5,
        "stale_access": 8,
        "role_drift": 4,
        "no_mfa": 2,
        "departed_user": 1
      }
    },
    "recent_decisions": [
      {
        "review_id": "uuid",
        "campaign_name": "Q1 2026 Quarterly Access Review",
        "user_email": "bob@acme.com",
        "resource_name": "GitHub Organization",
        "decision": "revoked",
        "decided_by_name": "Alice Engineer",
        "decided_at": "2026-02-22T14:30:00Z"
      }
    ],
    "overdue_reviews": [
      {
        "review_id": "uuid",
        "campaign_name": "Q1 2026 Quarterly Access Review",
        "reviewer_name": "Charlie TeamLead",
        "user_email": "dave@acme.com",
        "resource_name": "Production Database",
        "days_overdue": 3
      }
    ],
    "resource_coverage": {
      "total_resources": 6,
      "reviewed_last_quarter": 4,
      "never_reviewed": 1,
      "overdue_for_review": 1
    },
    "revocation_tracking": {
      "total_revocations": 5,
      "executed": 2,
      "pending_execution": 3,
      "avg_execution_time_hours": 48.5
    }
  }
}
```

---

#### `GET /api/v1/access-reviews/campaigns/:id/certification-report`

Generate a certification report for a completed campaign. Provides all decisions with full audit trail, formatted for auditor consumption. Used as evidence for SOC 2 and ISO 27001 audits.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, auditor

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `format` | string | No | `json` (default), `csv` |

**Response (200):**

```json
{
  "data": {
    "report": {
      "campaign_id": "uuid",
      "campaign_name": "Q4 2025 Quarterly Access Review",
      "cadence": "quarterly",
      "period": {
        "started_at": "2025-12-01T10:00:00Z",
        "completed_at": "2025-12-20T16:00:00Z",
        "deadline": "2025-12-31T00:00:00Z"
      },
      "scope": {
        "resource_criticalities": ["critical", "high"],
        "include_privileged_only": false,
        "include_service_accounts": true
      },
      "summary": {
        "total_reviews": 38,
        "approved": 32,
        "revoked": 4,
        "flagged": 1,
        "expired": 1,
        "completion_rate": 97.4,
        "avg_decision_time_hours": 24.2,
        "unique_reviewers": 6,
        "resources_reviewed": 4
      },
      "decisions": [
        {
          "resource_name": "AWS Console",
          "resource_criticality": "critical",
          "user_email": "alice@acme.com",
          "user_department": "Engineering",
          "user_title": "Senior Engineer",
          "role_name": "Admin",
          "is_privileged": true,
          "decision": "approved",
          "justification": "Admin access required for infrastructure management",
          "reviewer_name": "Bob Manager",
          "reviewer_email": "bob@acme.com",
          "decided_at": "2025-12-05T14:30:00Z",
          "anomalies_at_review": [],
          "revocation_executed": null,
          "revocation_executed_at": null
        },
        {
          "resource_name": "Stripe Dashboard",
          "resource_criticality": "critical",
          "user_email": "charlie@acme.com",
          "user_department": "Marketing",
          "user_title": "Marketing Analyst",
          "role_name": "Admin",
          "is_privileged": true,
          "decision": "revoked",
          "justification": "Marketing Analyst should not have Admin access to payment systems. Downgrade to Viewer.",
          "reviewer_name": "Diana Finance",
          "reviewer_email": "diana@acme.com",
          "decided_at": "2025-12-08T10:00:00Z",
          "anomalies_at_review": [
            { "type": "excessive_privileges", "details": "Admin access for non-finance role" }
          ],
          "revocation_executed": true,
          "revocation_executed_at": "2025-12-10T09:00:00Z"
        }
      ],
      "revocation_summary": {
        "total_revocations": 4,
        "executed": 3,
        "pending": 1,
        "avg_execution_time_hours": 36.5
      }
    },
    "generated_at": "2026-02-21T08:00:00Z"
  }
}
```

**Error Codes:**
- `400` Campaign is not in `completed` status
- `401` Unauthorized
- `403` Forbidden
- `404` Campaign not found

---

#### `GET /api/v1/access-reviews/my-reviews`

Get reviews assigned to the current user across all active campaigns. Provides the reviewer's personal queue.

- **Auth:** Bearer token required
- **Roles:** All roles except vendor_manager

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `decision` | string | No | Filter: `pending`, `approved`, `revoked`, `flagged`, `delegated` |
| `campaign_id` | UUID | No | Filter by campaign |
| `is_privileged` | boolean | No | Filter by privileged access |
| `has_anomalies` | boolean | No | Filter for reviews with anomalies |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort: `deadline`, `resource_criticality`, `created_at` |
| `order` | string | No | `asc`, `desc` (default: `asc`) |

**Response (200):**

```json
{
  "data": {
    "summary": {
      "total_pending": 8,
      "total_completed": 12,
      "campaigns_with_pending": 1,
      "earliest_deadline": "2026-03-07T00:00:00Z"
    },
    "reviews": [
      {
        "id": "uuid",
        "campaign_id": "uuid",
        "campaign_name": "Q1 2026 Quarterly Access Review",
        "campaign_deadline": "2026-03-07T00:00:00Z",
        "days_remaining": 14,
        "decision": "pending",
        "access_snapshot": {
          "resource_name": "Production Database",
          "resource_criticality": "critical",
          "user_email": "eve@acme.com",
          "user_department": "Engineering",
          "role_name": "ReadWrite",
          "is_privileged": false,
          "has_role_drift": false,
          "last_used_at": "2026-02-19T16:00:00Z",
          "anomalies": []
        },
        "assigned_at": "2026-02-21T10:00:00Z",
        "created_at": "2026-02-21T10:00:00Z"
      }
    ]
  },
  "pagination": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

---

## Endpoint Summary

| # | Method | Path | Description |
|---|--------|------|-------------|
| 1 | GET | `/api/v1/identity-providers` | List identity providers |
| 2 | GET | `/api/v1/identity-providers/:id` | Get identity provider details |
| 3 | POST | `/api/v1/identity-providers` | Connect identity provider |
| 4 | PUT | `/api/v1/identity-providers/:id` | Update identity provider |
| 5 | DELETE | `/api/v1/identity-providers/:id` | Disconnect identity provider |
| 6 | POST | `/api/v1/identity-providers/:id/sync` | Trigger manual sync |
| 7 | GET | `/api/v1/identity-providers/:id/sync-status` | Get sync status |
| 8 | GET | `/api/v1/access-resources` | List access resources |
| 9 | GET | `/api/v1/access-resources/:id` | Get resource details |
| 10 | POST | `/api/v1/access-resources` | Create resource |
| 11 | PUT | `/api/v1/access-resources/:id` | Update resource |
| 12 | DELETE | `/api/v1/access-resources/:id` | Delete resource |
| 13 | GET | `/api/v1/access-entries` | List access entries |
| 14 | GET | `/api/v1/access-entries/:id` | Get entry details |
| 15 | GET | `/api/v1/access-entries/anomalies` | Anomaly summary |
| 16 | POST | `/api/v1/access-entries/detect-anomalies` | Trigger anomaly detection |
| 17 | GET | `/api/v1/access-reviews/campaigns` | List campaigns |
| 18 | GET | `/api/v1/access-reviews/campaigns/:id` | Get campaign details |
| 19 | POST | `/api/v1/access-reviews/campaigns` | Create campaign |
| 20 | PUT | `/api/v1/access-reviews/campaigns/:id` | Update campaign |
| 21 | POST | `/api/v1/access-reviews/campaigns/:id/launch` | Launch campaign |
| 22 | POST | `/api/v1/access-reviews/campaigns/:id/complete` | Complete campaign |
| 23 | POST | `/api/v1/access-reviews/campaigns/:id/cancel` | Cancel campaign |
| 24 | GET | `/api/v1/access-reviews/campaigns/:id/stats` | Campaign statistics |
| 25 | GET | `/api/v1/access-reviews/campaigns/:id/reviews` | List reviews in campaign |
| 26 | GET | `/api/v1/access-reviews/campaigns/:id/reviews/:rid` | Get review details |
| 27 | POST | `/api/v1/access-reviews/campaigns/:id/reviews/:rid/decide` | Submit decision |
| 28 | POST | `/api/v1/access-reviews/campaigns/:id/reviews/bulk-decide` | Bulk decisions |
| 29 | POST | `/api/v1/access-reviews/campaigns/:id/reviews/:rid/delegate` | Delegate review |
| 30 | POST | `/api/v1/access-reviews/campaigns/:id/reviews/:rid/escalate` | Escalate review |
| 31 | POST | `/api/v1/access-reviews/campaigns/:id/reviews/:rid/revocation` | Mark revocation executed |
| 32 | GET | `/api/v1/access-reviews/dashboard` | Dashboard statistics |
| 33 | GET | `/api/v1/access-reviews/campaigns/:id/certification-report` | Certification report |
| 34 | GET | `/api/v1/access-reviews/my-reviews` | Reviewer's personal queue |

**Total: 34 endpoints**
