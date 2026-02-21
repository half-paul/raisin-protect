# Sprint 5 — API Specification: Policy Management

## Overview

Sprint 5 adds policy management endpoints: full CRUD for policies, version management with rich text content, sign-off workflows for policy approvals, policy-to-control mapping for gap detection, a template library for quick policy creation, and policy search/filtering.

This implements spec §6.1 (Policy Management) and lays the foundation for spec §5.1 (AI Policy Agent). Policies are the governance layer — they define *why* controls exist and *what* standards they enforce.

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `policies` | Per-org | Policy definitions with lifecycle |
| `policies/:id/versions` | Per-org | Version history with content |
| `policies/:id/signoffs` | Per-org | Sign-off workflow per version |
| `policies/:id/controls` | Per-org | Policy-to-control mappings |
| `policy-templates` | Per-org | Template library (read-only + clone) |
| `policy-gap` | Per-org | Gap detection analysis |
| `signoffs/pending` | Per-user | Cross-policy pending approvals |

---

## Endpoints

---

### 1. Policy CRUD

---

#### `GET /api/v1/policies`

List the org's policies with filtering, search, and pagination.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter: `draft`, `in_review`, `approved`, `published`, `archived` |
| `category` | string | — | Filter by policy_category enum value |
| `owner_id` | uuid | — | Filter by policy owner |
| `is_template` | bool | `false` | Include templates (default: only non-templates) |
| `framework_id` | uuid | — | Filter templates by framework |
| `tags` | string | — | Comma-separated tags (AND logic) |
| `review_status` | string | — | Filter: `overdue`, `due_soon`, `on_track`, `no_schedule` |
| `search` | string | — | Full-text search in title and description |
| `sort` | string | `identifier` | Sort: `identifier`, `title`, `category`, `status`, `next_review_at`, `published_at`, `created_at`, `updated_at` |
| `order` | string | `asc` | Sort order: `asc`, `desc` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "POL-AC-001",
      "title": "Access Control Policy",
      "description": "Defines requirements for user access management...",
      "category": "access_control",
      "status": "published",
      "owner": {
        "id": "uuid",
        "name": "Bob Security",
        "email": "security@acme.example.com"
      },
      "secondary_owner": null,
      "current_version": {
        "id": "uuid",
        "version_number": 2,
        "change_summary": "Updated MFA requirements for PCI DSS v4.0.1",
        "word_count": 3200,
        "created_at": "2026-02-15T10:00:00Z"
      },
      "review_frequency_days": 365,
      "next_review_at": "2027-02-01",
      "last_reviewed_at": "2026-02-01",
      "review_status": "on_track",
      "approved_at": "2026-02-01T14:00:00Z",
      "published_at": "2026-02-01T14:30:00Z",
      "is_template": false,
      "cloned_from_policy_id": "uuid",
      "linked_controls_count": 3,
      "pending_signoffs_count": 0,
      "tags": ["annual", "access", "production"],
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-02-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 12,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `review_status` is computed: `overdue` (past due), `due_soon` (within 30 days), `on_track` (beyond 30 days), `no_schedule` (no review configured).
- `linked_controls_count` and `pending_signoffs_count` are aggregated for the list view.
- Templates are excluded by default (`is_template=false`). Use `is_template=true` to browse templates.

**Error Codes:**
- `400` — Invalid filter/sort parameters
- `401` — Missing/invalid token
- `403` — Insufficient role permissions

---

#### `GET /api/v1/policies/:id`

Get a single policy with full details including current version content.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "POL-AC-001",
    "title": "Access Control Policy",
    "description": "Defines requirements for user access management...",
    "category": "access_control",
    "status": "published",
    "owner": {
      "id": "uuid",
      "name": "Bob Security",
      "email": "security@acme.example.com"
    },
    "secondary_owner": {
      "id": "uuid",
      "name": "Alice Compliance",
      "email": "compliance@acme.example.com"
    },
    "current_version": {
      "id": "uuid",
      "version_number": 2,
      "content": "<h1>Access Control Policy</h1><p>Version 2.0...</p>",
      "content_format": "html",
      "content_summary": "Access control policy with updated MFA requirements",
      "change_summary": "Updated MFA requirements for PCI DSS v4.0.1",
      "change_type": "minor",
      "word_count": 3200,
      "character_count": 18500,
      "created_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "created_at": "2026-02-15T10:00:00Z"
    },
    "review_frequency_days": 365,
    "next_review_at": "2027-02-01",
    "last_reviewed_at": "2026-02-01",
    "review_status": "on_track",
    "approved_at": "2026-02-01T14:00:00Z",
    "approved_version": 2,
    "published_at": "2026-02-01T14:30:00Z",
    "is_template": false,
    "template_framework": null,
    "cloned_from_policy_id": "uuid",
    "linked_controls": [
      {
        "id": "uuid",
        "identifier": "CTRL-AC-001",
        "title": "Multi-Factor Authentication",
        "category": "technical",
        "coverage": "full"
      }
    ],
    "signoff_summary": {
      "total": 2,
      "approved": 2,
      "pending": 0,
      "rejected": 0
    },
    "tags": ["annual", "access", "production"],
    "metadata": {},
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-02-15T10:00:00Z"
  }
}
```

**Notes:**
- `current_version` includes the full content. For list views, content is omitted for performance.
- `linked_controls` is embedded with basic info. Full control data available via `GET /api/v1/policies/:id/controls`.
- `signoff_summary` aggregates sign-off status for the current version.

**Error Codes:**
- `401` — Missing/invalid token
- `404` — Policy not found or belongs to different org

---

#### `POST /api/v1/policies`

Create a new policy.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`

**Request Body:**

```json
{
  "identifier": "POL-VM-001",
  "title": "Vulnerability Management Policy",
  "description": "Defines the process for identifying, assessing, and remediating vulnerabilities.",
  "category": "vulnerability_management",
  "owner_id": "uuid",
  "secondary_owner_id": "uuid",
  "review_frequency_days": 180,
  "tags": ["vulnerability", "quarterly"],
  "content": "<h1>Vulnerability Management Policy</h1><p>Version 1.0...</p>",
  "content_format": "html",
  "content_summary": "Initial vulnerability management policy"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `identifier` | string | Yes | Policy identifier (org-unique, e.g., `POL-VM-001`) |
| `title` | string | Yes | Policy title (max 500 chars) |
| `description` | string | No | Policy summary/abstract |
| `category` | string | Yes | policy_category enum value |
| `owner_id` | uuid | No | Primary policy owner (defaults to creator) |
| `secondary_owner_id` | uuid | No | Secondary owner |
| `review_frequency_days` | int | No | Review cadence in days (e.g., 365) |
| `tags` | string[] | No | Free-form tags |
| `content` | string | Yes | Initial version content (HTML/Markdown/plain text) |
| `content_format` | string | No | Content format: `html` (default), `markdown`, `plain_text` |
| `content_summary` | string | No | Brief summary of the content |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "POL-VM-001",
    "title": "Vulnerability Management Policy",
    "status": "draft",
    "current_version": {
      "id": "uuid",
      "version_number": 1,
      "change_type": "initial"
    },
    "created_at": "2026-02-20T19:00:00Z"
  }
}
```

**Behavior:**
1. Creates the policy with `status = 'draft'`
2. Creates the first `policy_version` (version 1, `change_type = 'initial'`)
3. Sets `current_version_id` to the new version
4. Sanitizes HTML content (strips scripts, event handlers, iframes)
5. Computes `word_count` and `character_count`
6. Logs `policy.created` and `policy_version.created` to audit log

**Error Codes:**
- `400` — Validation error (missing fields, invalid category, identifier already exists)
- `401` — Missing/invalid token
- `403` — Role not authorized to create policies
- `409` — Identifier already exists for this org

---

#### `PUT /api/v1/policies/:id`

Update policy metadata (not content — use version endpoints for content changes).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, policy owner

**Request Body:**

```json
{
  "title": "Vulnerability Management Policy (Updated)",
  "description": "Updated description...",
  "category": "vulnerability_management",
  "owner_id": "uuid",
  "secondary_owner_id": "uuid",
  "review_frequency_days": 365,
  "next_review_at": "2027-02-20",
  "tags": ["vulnerability", "annual", "updated"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | No | Updated title |
| `description` | string | No | Updated description |
| `category` | string | No | Updated category |
| `owner_id` | uuid | No | New primary owner |
| `secondary_owner_id` | uuid | No | New secondary owner (null to clear) |
| `review_frequency_days` | int | No | Updated review cadence |
| `next_review_at` | date | No | Manually set next review date |
| `tags` | string[] | No | Replace tags |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "POL-VM-001",
    "title": "Vulnerability Management Policy (Updated)",
    "status": "draft",
    "updated_at": "2026-02-20T19:10:00Z"
  }
}
```

**Notes:**
- Cannot change `identifier` or `is_template` after creation.
- Status transitions are handled by dedicated endpoints (see Status Transitions below).
- If `owner_id` changes, logs `policy.owner_changed` to audit log.

**Error Codes:**
- `400` — Invalid field values
- `401` — Missing/invalid token
- `403` — Not authorized (not owner, compliance_manager, or ciso)
- `404` — Policy not found

---

#### `POST /api/v1/policies/:id/archive`

Archive a policy (soft-delete — moves to archived status).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "POL-VM-001",
    "status": "archived",
    "updated_at": "2026-02-20T19:15:00Z"
  }
}
```

**Behavior:**
1. Sets `status = 'archived'`
2. Withdraws any pending sign-offs
3. Logs `policy.archived` to audit log

**Error Codes:**
- `400` — Policy is already archived
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Policy not found

---

#### `POST /api/v1/policies/:id/submit-for-review`

Submit a draft policy for review (transitions to `in_review`).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, policy owner

**Request Body:**

```json
{
  "signer_ids": ["uuid-1", "uuid-2"],
  "due_date": "2026-03-01",
  "message": "Please review the updated vulnerability management policy."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `signer_ids` | uuid[] | Yes | Users who must sign off (1-10 signers) |
| `due_date` | date | No | Optional deadline for all sign-offs |
| `message` | string | No | Optional message to include in notification |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "POL-VM-001",
    "status": "in_review",
    "signoffs_created": 2,
    "signoffs": [
      {
        "id": "uuid",
        "signer": { "id": "uuid", "name": "David CISO" },
        "status": "pending",
        "due_date": "2026-03-01"
      },
      {
        "id": "uuid",
        "signer": { "id": "uuid", "name": "Alice Compliance" },
        "status": "pending",
        "due_date": "2026-03-01"
      }
    ]
  }
}
```

**Behavior:**
1. Validates policy is in `draft` or `approved` status (re-review after changes)
2. Sets `status = 'in_review'`
3. Creates `policy_signoff` records for each signer (linked to `current_version_id`)
4. Snapshots each signer's current `grc_role` into `signer_role`
5. Logs `policy.status_changed` and `policy_signoff.requested` to audit log
6. *(Future: triggers email/Slack notification to signers)*

**Error Codes:**
- `400` — No signers specified, policy not in valid state for review, signer_ids empty or >10
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Policy not found
- `422` — One or more signer_ids are invalid or not in this org

---

#### `POST /api/v1/policies/:id/publish`

Publish an approved policy (make it active/in-effect).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "POL-VM-001",
    "status": "published",
    "published_at": "2026-02-20T19:30:00Z",
    "current_version": {
      "version_number": 1
    }
  }
}
```

**Behavior:**
1. Validates policy is in `approved` status
2. Sets `status = 'published'`, `published_at = NOW()`
3. Computes `next_review_at = NOW() + review_frequency_days` if review_frequency_days is set
4. Sets `last_reviewed_at = NOW()`
5. Logs `policy.status_changed` to audit log

**Error Codes:**
- `400` — Policy is not in `approved` status
- `401` — Missing/invalid token
- `403` — Not authorized (only compliance_manager or ciso can publish)
- `404` — Policy not found

---

### 2. Policy Version Management

---

#### `GET /api/v1/policies/:id/versions`

List all versions of a policy.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "version_number": 2,
      "is_current": true,
      "content_format": "html",
      "content_summary": "Updated MFA requirements",
      "change_summary": "Section 3.1 updated for PCI DSS v4.0.1 MFA requirements",
      "change_type": "minor",
      "word_count": 3200,
      "character_count": 18500,
      "created_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "signoff_summary": {
        "total": 2,
        "approved": 2,
        "pending": 0,
        "rejected": 0
      },
      "created_at": "2026-02-15T10:00:00Z"
    },
    {
      "id": "uuid",
      "version_number": 1,
      "is_current": false,
      "content_format": "html",
      "content_summary": "Initial access control policy",
      "change_summary": "Initial version",
      "change_type": "initial",
      "word_count": 2800,
      "character_count": 16200,
      "created_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "signoff_summary": {
        "total": 1,
        "approved": 1,
        "pending": 0,
        "rejected": 0
      },
      "created_at": "2026-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 2,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Content is NOT included in the list response (performance). Use `GET /api/v1/policies/:id/versions/:version_number` for full content.
- `signoff_summary` is aggregated per version.
- Ordered by `version_number DESC` (newest first).

---

#### `GET /api/v1/policies/:id/versions/:version_number`

Get a specific version with full content.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "policy_id": "uuid",
    "version_number": 2,
    "is_current": true,
    "content": "<h1>Access Control Policy</h1><h2>1. Purpose</h2><p>This policy defines...</p>",
    "content_format": "html",
    "content_summary": "Updated MFA requirements for PCI DSS v4.0.1",
    "change_summary": "Section 3.1 updated for PCI DSS v4.0.1 MFA requirements",
    "change_type": "minor",
    "word_count": 3200,
    "character_count": 18500,
    "created_by": {
      "id": "uuid",
      "name": "Bob Security",
      "email": "security@acme.example.com"
    },
    "signoffs": [
      {
        "id": "uuid",
        "signer": { "id": "uuid", "name": "David CISO" },
        "status": "approved",
        "decided_at": "2026-02-01T14:00:00Z",
        "comments": "Approved — aligns with requirements."
      }
    ],
    "created_at": "2026-02-15T10:00:00Z"
  }
}
```

**Error Codes:**
- `401` — Missing/invalid token
- `404` — Policy or version not found

---

#### `POST /api/v1/policies/:id/versions`

Create a new version of a policy (edit content).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, policy owner

**Request Body:**

```json
{
  "content": "<h1>Access Control Policy</h1><h2>1. Purpose</h2><p>Updated content...</p>",
  "content_format": "html",
  "content_summary": "Updated MFA requirements for PCI DSS v4.0.1",
  "change_summary": "Section 3.1 updated — added MFA requirements per PCI DSS v4.0.1",
  "change_type": "minor"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | string | Yes | New version content |
| `content_format` | string | No | Format: `html` (default), `markdown`, `plain_text` |
| `content_summary` | string | No | Brief summary of this version |
| `change_summary` | string | Yes | What changed from previous version |
| `change_type` | string | No | Change magnitude: `major`, `minor` (default), `patch` |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "policy_id": "uuid",
    "version_number": 3,
    "is_current": true,
    "content_format": "html",
    "change_summary": "Section 3.1 updated — added MFA requirements per PCI DSS v4.0.1",
    "change_type": "minor",
    "word_count": 3400,
    "created_at": "2026-02-20T19:00:00Z"
  }
}
```

**Behavior:**
1. Sets previous current version's `is_current = FALSE`
2. Creates new version with incremented `version_number` and `is_current = TRUE`
3. Sanitizes HTML content
4. Computes `word_count` and `character_count`
5. Updates `policies.current_version_id` to new version
6. If policy was in `approved` or `published` status, reverts to `draft` (content changed since approval)
7. Logs `policy_version.created` to audit log

**Error Codes:**
- `400` — Content is empty or exceeds max size (1MB)
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Policy not found
- `422` — Policy is archived (cannot create new versions)

---

#### `GET /api/v1/policies/:id/versions/compare`

Compare two versions side-by-side (returns both versions' content for client-side diffing).

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `v1` | int | Yes | First version number (typically older) |
| `v2` | int | Yes | Second version number (typically newer) |

**Response 200:**

```json
{
  "data": {
    "policy_id": "uuid",
    "policy_identifier": "POL-AC-001",
    "versions": [
      {
        "version_number": 1,
        "content": "<h1>Access Control Policy</h1><p>Original content...</p>",
        "content_format": "html",
        "content_summary": "Initial version",
        "change_summary": "Initial version",
        "change_type": "initial",
        "word_count": 2800,
        "created_by": { "id": "uuid", "name": "Bob Security" },
        "created_at": "2026-01-15T10:00:00Z"
      },
      {
        "version_number": 2,
        "content": "<h1>Access Control Policy</h1><p>Updated content...</p>",
        "content_format": "html",
        "content_summary": "Updated MFA requirements",
        "change_summary": "Section 3.1 updated for PCI DSS v4.0.1",
        "change_type": "minor",
        "word_count": 3200,
        "created_by": { "id": "uuid", "name": "Bob Security" },
        "created_at": "2026-02-15T10:00:00Z"
      }
    ],
    "word_count_delta": 400
  }
}
```

**Notes:**
- The API returns both versions' full content. Diff computation is done client-side using a library like `diff-match-patch`.
- `word_count_delta` shows the net change in words between versions.

**Error Codes:**
- `400` — `v1` or `v2` missing, `v1 == v2`, invalid version numbers
- `401` — Missing/invalid token
- `404` — Policy not found or version doesn't exist

---

### 3. Policy Sign-Off Workflow

---

#### `GET /api/v1/policies/:id/signoffs`

List all sign-offs for a policy, optionally filtered by version.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `version_number` | int | — | Filter sign-offs for a specific version |
| `status` | string | — | Filter: `pending`, `approved`, `rejected`, `withdrawn` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "policy_id": "uuid",
      "policy_version": {
        "id": "uuid",
        "version_number": 2
      },
      "signer": {
        "id": "uuid",
        "name": "David CISO",
        "email": "ciso@acme.example.com",
        "role": "ciso"
      },
      "signer_role": "ciso",
      "requested_by": {
        "id": "uuid",
        "name": "Alice Compliance"
      },
      "requested_at": "2026-02-18T09:00:00Z",
      "due_date": "2026-03-01",
      "status": "pending",
      "decided_at": null,
      "comments": null,
      "reminder_count": 1,
      "reminder_sent_at": "2026-02-19T09:00:00Z"
    }
  ],
  "meta": {
    "total": 2,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

---

#### `POST /api/v1/policies/:id/signoffs/:signoff_id/approve`

Approve a sign-off request.

- **Auth:** Bearer token required
- **Roles:** The designated signer only

**Request Body:**

```json
{
  "comments": "Reviewed and approved. Meets all framework requirements."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `comments` | string | No | Optional approval comments |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "approved",
    "decided_at": "2026-02-20T19:30:00Z",
    "comments": "Reviewed and approved. Meets all framework requirements.",
    "policy_status": "approved",
    "all_signoffs_complete": true
  }
}
```

**Behavior:**
1. Validates the authenticated user is the designated signer
2. Sets `status = 'approved'`, `decided_at = NOW()`
3. Logs `policy_signoff.approved` to audit log
4. Checks if ALL sign-offs for this version are now approved
5. If all approved: automatically transitions `policies.status` to `approved`, sets `approved_at`, `approved_version`
6. `policy_status` in the response indicates the policy's new status
7. `all_signoffs_complete` indicates whether this was the final required approval

**Error Codes:**
- `400` — Sign-off is not in `pending` status
- `401` — Missing/invalid token
- `403` — Authenticated user is not the designated signer
- `404` — Policy or sign-off not found

---

#### `POST /api/v1/policies/:id/signoffs/:signoff_id/reject`

Reject a sign-off request.

- **Auth:** Bearer token required
- **Roles:** The designated signer only

**Request Body:**

```json
{
  "comments": "Section 4.2 does not address the new PCI DSS v4.0.1 requirements for privileged access."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `comments` | string | **Yes** | Mandatory rejection reason |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "rejected",
    "decided_at": "2026-02-20T19:30:00Z",
    "comments": "Section 4.2 does not address the new PCI DSS v4.0.1 requirements for privileged access.",
    "policy_status": "in_review"
  }
}
```

**Behavior:**
1. Validates the authenticated user is the designated signer
2. Validates `comments` is not empty (mandatory for rejections)
3. Sets `status = 'rejected'`, `decided_at = NOW()`
4. Logs `policy_signoff.rejected` to audit log
5. Policy status remains `in_review` — owner must address feedback, create new version, and re-submit

**Error Codes:**
- `400` — Missing comments, sign-off not in `pending` status
- `401` — Missing/invalid token
- `403` — Authenticated user is not the designated signer
- `404` — Policy or sign-off not found

---

#### `POST /api/v1/policies/:id/signoffs/:signoff_id/withdraw`

Withdraw a pending sign-off request (cancel it).

- **Auth:** Bearer token required
- **Roles:** The original requester, `compliance_manager`, `ciso`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "withdrawn",
    "decided_at": "2026-02-20T19:35:00Z"
  }
}
```

**Behavior:**
1. Validates sign-off is in `pending` status
2. Sets `status = 'withdrawn'`, `decided_at = NOW()`
3. Logs `policy_signoff.withdrawn` to audit log

**Error Codes:**
- `400` — Sign-off is not in `pending` status
- `401` — Missing/invalid token
- `403` — Not the original requester, compliance_manager, or ciso
- `404` — Sign-off not found

---

#### `GET /api/v1/signoffs/pending`

List all pending sign-offs for the authenticated user across all policies.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `urgency` | string | — | Filter: `overdue`, `due_soon`, `on_time` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "policy": {
        "id": "uuid",
        "identifier": "POL-IR-001",
        "title": "Incident Response Plan",
        "category": "incident_response"
      },
      "policy_version": {
        "id": "uuid",
        "version_number": 1,
        "content_summary": "Initial incident response plan",
        "word_count": 2500
      },
      "requested_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "requested_at": "2026-02-18T09:00:00Z",
      "due_date": "2026-03-01",
      "urgency": "on_time",
      "reminder_count": 0
    }
  ],
  "meta": {
    "total": 3,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Returns sign-offs where `signer_id = authenticated_user AND status = 'pending'`.
- `urgency` is computed: `overdue` (past due_date), `due_soon` (within 3 days), `on_time` (beyond 3 days or no due_date).
- Useful for the "My Pending Approvals" dashboard widget.

---

### 4. Policy-to-Control Mapping

---

#### `GET /api/v1/policies/:id/controls`

List controls linked to a policy.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "policy_control_id": "uuid",
      "identifier": "CTRL-AC-001",
      "title": "Multi-Factor Authentication",
      "description": "Enforce MFA for all user accounts...",
      "category": "technical",
      "status": "active",
      "coverage": "full",
      "notes": "Section 3.1 — MFA enforcement requirements",
      "linked_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "linked_at": "2026-02-01T10:00:00Z",
      "frameworks": ["SOC 2", "PCI DSS", "ISO 27001"]
    }
  ],
  "meta": {
    "total": 3,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `frameworks` is a convenience aggregation showing which frameworks this control maps to (via `control_mappings` → `requirements` → `framework_versions` → `frameworks`).
- `policy_control_id` is the ID of the junction table record (for unlinking).

---

#### `POST /api/v1/policies/:id/controls`

Link a control to a policy.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, policy owner

**Request Body:**

```json
{
  "control_id": "uuid",
  "coverage": "full",
  "notes": "Section 3.1 fully addresses MFA requirements for this control"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `control_id` | uuid | Yes | Control to link |
| `coverage` | string | No | Coverage level: `full` (default), `partial` |
| `notes` | string | No | Why this policy governs this control |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "policy_id": "uuid",
    "control": {
      "id": "uuid",
      "identifier": "CTRL-AC-001",
      "title": "Multi-Factor Authentication"
    },
    "coverage": "full",
    "notes": "Section 3.1 fully addresses MFA requirements for this control",
    "linked_by": {
      "id": "uuid",
      "name": "Bob Security"
    },
    "created_at": "2026-02-20T19:00:00Z"
  }
}
```

**Behavior:**
1. Validates control exists and belongs to the same org
2. Creates `policy_controls` junction record
3. Logs `policy_control.linked` to audit log

**Error Codes:**
- `400` — Invalid control_id or coverage value
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Policy or control not found
- `409` — Control is already linked to this policy

---

#### `DELETE /api/v1/policies/:id/controls/:control_id`

Unlink a control from a policy.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, policy owner

**Response 204:** No content

**Behavior:**
1. Deletes the `policy_controls` junction record
2. Logs `policy_control.unlinked` to audit log

**Error Codes:**
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Policy-control link not found

---

#### `POST /api/v1/policies/:id/controls/bulk`

Bulk link multiple controls to a policy.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, policy owner

**Request Body:**

```json
{
  "links": [
    { "control_id": "uuid-1", "coverage": "full", "notes": "Section 3.1" },
    { "control_id": "uuid-2", "coverage": "partial", "notes": "Section 4.2 partially" },
    { "control_id": "uuid-3", "coverage": "full" }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `links` | array | Yes | Array of link objects (max 50) |
| `links[].control_id` | uuid | Yes | Control to link |
| `links[].coverage` | string | No | `full` (default) or `partial` |
| `links[].notes` | string | No | Justification |

**Response 201:**

```json
{
  "data": {
    "created": 3,
    "skipped": 0,
    "errors": []
  }
}
```

**Notes:**
- Skips controls already linked (returns count in `skipped`).
- Processes all valid links even if some fail (partial success).

**Error Codes:**
- `400` — Empty links array or exceeds max 50
- `401` — Missing/invalid token
- `403` — Not authorized

---

### 5. Policy Templates

---

#### `GET /api/v1/policy-templates`

List available policy templates, optionally filtered by framework.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `framework_id` | uuid | — | Filter templates by framework |
| `category` | string | — | Filter by policy_category |
| `search` | string | — | Search in title and description |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "TPL-IS-001",
      "title": "Information Security Policy",
      "description": "Comprehensive information security policy establishing...",
      "category": "information_security",
      "framework": {
        "id": "uuid",
        "identifier": "soc2",
        "name": "SOC 2"
      },
      "current_version": {
        "id": "uuid",
        "version_number": 1,
        "word_count": 280,
        "content_summary": "Comprehensive information security policy template..."
      },
      "review_frequency_days": 365,
      "tags": ["soc2", "iso27001", "pci", "template", "mandatory"]
    }
  ],
  "meta": {
    "total": 15,
    "request_id": "uuid"
  }
}
```

**Notes:**
- This is a convenience alias for `GET /api/v1/policies?is_template=true` with a simpler response shape.
- Content is NOT included — use `GET /api/v1/policies/:id/versions/:version` to preview template content.

---

#### `POST /api/v1/policy-templates/:id/clone`

Clone a template into a new policy for the organization.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`

**Request Body:**

```json
{
  "identifier": "POL-IS-001",
  "title": "Acme Corp Information Security Policy",
  "description": "Acme Corporation's information security policy.",
  "owner_id": "uuid",
  "review_frequency_days": 365,
  "tags": ["mandatory", "annual", "all-employees"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `identifier` | string | Yes | New policy identifier (org-unique) |
| `title` | string | No | New title (defaults to template title) |
| `description` | string | No | New description (defaults to template description) |
| `owner_id` | uuid | No | Owner (defaults to creator) |
| `review_frequency_days` | int | No | Review cadence (defaults to template value) |
| `tags` | string[] | No | Tags (defaults to template tags minus 'template') |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "POL-IS-001",
    "title": "Acme Corp Information Security Policy",
    "status": "draft",
    "cloned_from_policy_id": "uuid",
    "current_version": {
      "id": "uuid",
      "version_number": 1,
      "word_count": 280
    },
    "created_at": "2026-02-20T19:00:00Z"
  }
}
```

**Behavior:**
1. Creates new policy with `status = 'draft'`, `is_template = FALSE`, `cloned_from_policy_id` set to template ID
2. Copies the template's current version content into a new `policy_version` (version 1)
3. Sets `change_summary = "Cloned from template: {template_identifier}"`
4. Copies `category` and `review_frequency_days` from template (unless overridden)
5. Logs `policy.created` and `policy.cloned_from_template` to audit log

**Error Codes:**
- `400` — Template not found, identifier already exists
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Template not found
- `409` — Identifier already exists

---

### 6. Policy Gap Detection

---

#### `GET /api/v1/policy-gap`

Identify controls without policy coverage.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, `auditor`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `framework_id` | uuid | — | Filter to controls mapped to a specific framework |
| `category` | string | — | Filter by control category |
| `include_partial` | bool | `false` | If true, also include controls with only `partial` policy coverage |

**Response 200:**

```json
{
  "data": {
    "summary": {
      "total_active_controls": 318,
      "controls_with_full_coverage": 245,
      "controls_with_partial_coverage": 28,
      "controls_without_coverage": 45,
      "coverage_percentage": 77.04
    },
    "gaps": [
      {
        "control": {
          "id": "uuid",
          "identifier": "CTRL-NW-001",
          "title": "Network Segmentation",
          "category": "technical",
          "status": "active",
          "owner": {
            "id": "uuid",
            "name": "Eve DevOps"
          }
        },
        "mapped_frameworks": ["PCI DSS", "SOC 2", "ISO 27001"],
        "mapped_requirements_count": 5,
        "policy_coverage": "none",
        "suggested_categories": ["network_security"]
      },
      {
        "control": {
          "id": "uuid",
          "identifier": "CTRL-DP-001",
          "title": "Encryption at Rest",
          "category": "technical",
          "status": "active",
          "owner": {
            "id": "uuid",
            "name": "Eve DevOps"
          }
        },
        "mapped_frameworks": ["PCI DSS", "GDPR"],
        "mapped_requirements_count": 3,
        "policy_coverage": "none",
        "suggested_categories": ["encryption", "data_classification"]
      }
    ]
  },
  "meta": {
    "total_gaps": 45,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `suggested_categories` recommends which `policy_category` would likely cover this control, based on the control's category and common governance patterns.
- Gaps are ordered by `mapped_requirements_count DESC` (most-impactful gaps first).
- When `include_partial=true`, controls with only `partial` policy coverage are included in the gaps list with `policy_coverage: "partial"`.
- When `framework_id` is specified, only controls mapped to that framework are considered.

**Error Codes:**
- `401` — Missing/invalid token
- `403` — Not authorized

---

#### `GET /api/v1/policy-gap/by-framework`

Gap analysis grouped by framework.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, `auditor`

**Response 200:**

```json
{
  "data": [
    {
      "framework": {
        "id": "uuid",
        "identifier": "pci_dss",
        "name": "PCI DSS",
        "version": "4.0.1"
      },
      "total_requirements": 280,
      "requirements_with_controls": 265,
      "controls_with_policy_coverage": 220,
      "controls_without_policy_coverage": 45,
      "policy_coverage_percentage": 83.02,
      "gap_count": 45
    },
    {
      "framework": {
        "id": "uuid",
        "identifier": "soc2",
        "name": "SOC 2",
        "version": "2024"
      },
      "total_requirements": 64,
      "requirements_with_controls": 62,
      "controls_with_policy_coverage": 55,
      "controls_without_policy_coverage": 7,
      "policy_coverage_percentage": 88.71,
      "gap_count": 7
    }
  ],
  "meta": {
    "request_id": "uuid"
  }
}
```

**Notes:**
- Only includes frameworks activated by the org (via `org_frameworks`).
- `policy_coverage_percentage` is the percentage of controls (mapped to this framework's requirements) that have at least one policy with `full` coverage.

---

### 7. Policy Search and Filtering

---

#### `GET /api/v1/policies/search`

Advanced search across policies and their content.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `q` | string | — | **Required.** Search query |
| `scope` | string | `metadata` | Search scope: `metadata` (title + description), `content` (includes version content), `all` |
| `status` | string | — | Filter by status |
| `category` | string | — | Filter by category |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "POL-AC-001",
      "title": "Access Control Policy",
      "description": "Defines requirements for user access management...",
      "category": "access_control",
      "status": "published",
      "match_context": "...multi-factor authentication (MFA) shall be required for all user accounts...",
      "match_source": "content",
      "current_version_number": 2,
      "owner": {
        "id": "uuid",
        "name": "Bob Security"
      }
    }
  ],
  "meta": {
    "total": 3,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `match_context` provides a snippet of the matching text for display.
- `match_source` indicates where the match was found: `title`, `description`, or `content`.
- When `scope=content` or `scope=all`, searches within `policy_versions.content` of the current version.
- Uses PostgreSQL full-text search (`to_tsvector`/`to_tsquery`).
- Templates are excluded from search results by default.

**Error Codes:**
- `400` — Missing `q` parameter
- `401` — Missing/invalid token

---

### 8. Policy Approval Notifications

---

#### `POST /api/v1/policies/:id/signoffs/remind`

Send reminders for pending sign-offs.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, policy owner, original requester

**Request Body:**

```json
{
  "signoff_ids": ["uuid-1", "uuid-2"],
  "message": "Gentle reminder: please review and sign off on the updated policy."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `signoff_ids` | uuid[] | No | Specific sign-offs to remind (default: all pending for this policy) |
| `message` | string | No | Custom reminder message |

**Response 200:**

```json
{
  "data": {
    "reminders_sent": 2,
    "signers": [
      { "id": "uuid", "name": "David CISO", "reminder_count": 2 },
      { "id": "uuid", "name": "Alice Compliance", "reminder_count": 1 }
    ]
  }
}
```

**Behavior:**
1. Identifies pending sign-offs (all or specified IDs)
2. Updates `reminder_sent_at = NOW()` and increments `reminder_count`
3. *(Future: triggers email/Slack notification to each signer)*
4. Rate-limited: max 1 reminder per signer per 24 hours

**Error Codes:**
- `400` — No pending sign-offs found
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Policy not found
- `429` — Reminder already sent within the last 24 hours

---

### 9. Policy Statistics

---

#### `GET /api/v1/policies/stats`

Get policy management statistics for the dashboard.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "total_policies": 12,
    "by_status": {
      "draft": 2,
      "in_review": 1,
      "approved": 1,
      "published": 7,
      "archived": 1
    },
    "by_category": {
      "information_security": 1,
      "access_control": 2,
      "incident_response": 1,
      "data_privacy": 2,
      "network_security": 1,
      "encryption": 1,
      "vulnerability_management": 1,
      "change_management": 1,
      "business_continuity": 1,
      "secure_development": 1
    },
    "review_status": {
      "overdue": 1,
      "due_within_30_days": 2,
      "on_track": 8,
      "no_schedule": 1
    },
    "signoff_summary": {
      "total_pending": 3,
      "overdue_signoffs": 1
    },
    "gap_summary": {
      "total_active_controls": 318,
      "controls_with_policy_coverage": 273,
      "coverage_percentage": 85.85
    },
    "templates_available": 15,
    "recent_activity": [
      {
        "policy_identifier": "POL-IR-001",
        "action": "submitted_for_review",
        "actor": "Bob Security",
        "timestamp": "2026-02-18T09:00:00Z"
      },
      {
        "policy_identifier": "POL-AC-001",
        "action": "new_version_created",
        "actor": "Bob Security",
        "timestamp": "2026-02-15T10:00:00Z"
      }
    ]
  }
}
```

**Notes:**
- `recent_activity` shows the last 5 policy-related actions from the audit log.
- `gap_summary` gives a quick view of policy coverage without running the full gap analysis.
- Templates are NOT counted in `total_policies`.

---

## Status Transition Diagram

```
                    ┌──────────────────────────────────────────────┐
                    │                                              │
                    ▼                                              │
               ┌─────────┐      submit-for-review       ┌────────────┐
       ────▶   │  DRAFT  │  ─────────────────────────▶   │ IN_REVIEW  │
  create       │         │  (creates sign-off requests)  │            │
               └────┬────┘                               └─────┬──────┘
                    │                                          │
                    │ ◄───── new version created ──────────────│
                    │        (auto-revert on content change)   │
                    │                                    ┌─────┴──────┐
                    │                                    │            │
                    │                              all approved  rejected
                    │                                    │       (stays in_review)
                    │                                    ▼
                    │                               ┌──────────┐
                    │                               │ APPROVED │
                    │                               └────┬─────┘
                    │                                    │
                    │                                 publish
                    │                                    │
                    │                                    ▼
                    │                              ┌───────────┐
                    │                              │ PUBLISHED │
                    │                              └─────┬─────┘
                    │                                    │
                    ▼                                    ▼
               ┌──────────┐                        ┌──────────┐
               │ ARCHIVED │ ◄──────── archive ───── │ (any)    │
               └──────────┘                        └──────────┘
```

**Transition Rules:**
| From | To | Trigger | Conditions |
|------|----|---------|------------|
| — | `draft` | `POST /policies` | Creating a new policy |
| `draft` | `in_review` | `POST /policies/:id/submit-for-review` | At least one signer specified |
| `in_review` | `approved` | Automatic | All sign-offs approved |
| `in_review` | `in_review` | Rejection | Sign-off rejected (stays in review) |
| `approved` | `published` | `POST /policies/:id/publish` | Only compliance_manager or ciso |
| `approved`/`published` | `draft` | `POST /policies/:id/versions` | New version reverts to draft |
| `draft` | `in_review` | `POST /policies/:id/submit-for-review` | Re-submit after changes |
| Any (except archived) | `archived` | `POST /policies/:id/archive` | Compliance_manager or ciso |

---

## Content Security

### HTML Sanitization Rules

All HTML content is sanitized before storage using an allowlist approach:

**Allowed tags:** `h1`, `h2`, `h3`, `h4`, `h5`, `h6`, `p`, `br`, `hr`, `ul`, `ol`, `li`, `table`, `thead`, `tbody`, `tr`, `th`, `td`, `strong`, `em`, `u`, `s`, `blockquote`, `pre`, `code`, `a`, `img`, `span`, `div`, `sub`, `sup`

**Allowed attributes:**
- Global: `class`, `id`, `style` (restricted — see below)
- `a`: `href` (http/https only), `title`, `target`
- `img`: `src` (http/https only), `alt`, `title`, `width`, `height`
- `td`/`th`: `colspan`, `rowspan`

**Stripped:**
- `<script>`, `<iframe>`, `<object>`, `<embed>`, `<form>`, `<input>`
- Event handlers: `onclick`, `onload`, `onerror`, etc.
- `javascript:`, `data:`, `vbscript:` URLs
- `style` attributes with `expression()`, `url()`, `import`

**Recommended library:** `bluemonday` (Go) with `UGCPolicy()` + custom additions for table support.

---

## Authentication & Authorization Summary

| Endpoint | Method | Roles Allowed |
|----------|--------|---------------|
| `GET /policies` | GET | All |
| `GET /policies/:id` | GET | All |
| `POST /policies` | POST | compliance_manager, ciso, security_engineer |
| `PUT /policies/:id` | PUT | compliance_manager, ciso, security_engineer, owner |
| `POST /policies/:id/archive` | POST | compliance_manager, ciso |
| `POST /policies/:id/submit-for-review` | POST | compliance_manager, ciso, security_engineer, owner |
| `POST /policies/:id/publish` | POST | compliance_manager, ciso |
| `GET /policies/:id/versions` | GET | All |
| `GET /policies/:id/versions/:vn` | GET | All |
| `POST /policies/:id/versions` | POST | compliance_manager, ciso, security_engineer, owner |
| `GET /policies/:id/versions/compare` | GET | All |
| `GET /policies/:id/signoffs` | GET | All |
| `POST /policies/:id/signoffs/:id/approve` | POST | Designated signer only |
| `POST /policies/:id/signoffs/:id/reject` | POST | Designated signer only |
| `POST /policies/:id/signoffs/:id/withdraw` | POST | Requester, compliance_manager, ciso |
| `GET /signoffs/pending` | GET | All (filtered to own) |
| `GET /policies/:id/controls` | GET | All |
| `POST /policies/:id/controls` | POST | compliance_manager, ciso, security_engineer, owner |
| `DELETE /policies/:id/controls/:cid` | DELETE | compliance_manager, ciso, security_engineer, owner |
| `POST /policies/:id/controls/bulk` | POST | compliance_manager, ciso, security_engineer, owner |
| `GET /policy-templates` | GET | All |
| `POST /policy-templates/:id/clone` | POST | compliance_manager, ciso, security_engineer |
| `GET /policy-gap` | GET | compliance_manager, ciso, security_engineer, auditor |
| `GET /policy-gap/by-framework` | GET | compliance_manager, ciso, security_engineer, auditor |
| `GET /policies/search` | GET | All |
| `POST /policies/:id/signoffs/remind` | POST | compliance_manager, ciso, owner, requester |
| `GET /policies/stats` | GET | All |

---

## Audit Log Events

All policy-related actions are logged to `audit_log`:

| Action | Resource Type | When |
|--------|--------------|------|
| `policy.created` | policy | New policy created |
| `policy.updated` | policy | Policy metadata updated |
| `policy.status_changed` | policy | Status transition (draft→in_review, etc.) |
| `policy.archived` | policy | Policy archived |
| `policy.owner_changed` | policy | Owner reassigned |
| `policy.cloned_from_template` | policy | Policy cloned from template |
| `policy_version.created` | policy_version | New version created |
| `policy_version.published` | policy_version | Version published |
| `policy_signoff.requested` | policy_signoff | Sign-off requested |
| `policy_signoff.approved` | policy_signoff | Sign-off approved |
| `policy_signoff.rejected` | policy_signoff | Sign-off rejected |
| `policy_signoff.withdrawn` | policy_signoff | Sign-off withdrawn |
| `policy_control.linked` | policy_control | Control linked to policy |
| `policy_control.unlinked` | policy_control | Control unlinked from policy |

---

## Error Response Format

All errors follow the standard format from Sprint 1:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Policy identifier already exists",
    "details": {
      "field": "identifier",
      "value": "POL-AC-001"
    },
    "request_id": "uuid"
  }
}
```

**Sprint 5 Error Codes:**

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `VALIDATION_ERROR` | Invalid input data |
| 400 | `INVALID_STATUS_TRANSITION` | Status change not allowed (e.g., draft → published) |
| 400 | `CONTENT_TOO_LARGE` | Policy content exceeds 1MB |
| 400 | `NO_PENDING_SIGNOFFS` | No pending sign-offs to remind |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 403 | `FORBIDDEN` | Role not permitted for this action |
| 403 | `NOT_SIGNER` | User is not the designated signer |
| 403 | `NOT_OWNER` | User is not the policy owner |
| 404 | `NOT_FOUND` | Resource not found or wrong org |
| 409 | `DUPLICATE_IDENTIFIER` | Policy identifier already exists |
| 409 | `ALREADY_LINKED` | Control already linked to this policy |
| 422 | `POLICY_ARCHIVED` | Cannot modify archived policy |
| 422 | `REJECTION_REQUIRES_COMMENTS` | Rejection must include comments |
| 429 | `REMINDER_RATE_LIMITED` | Reminder already sent within 24 hours |
