# Sprint 3 — API Specification: Evidence Management

## Overview

Sprint 3 adds the evidence management API: upload and store evidence artifacts in MinIO, link evidence to controls and requirements, track versioning and freshness, detect staleness, and evaluate evidence sufficiency. This is the bridge between "we have controls" and "we can prove they work" (spec §3.4).

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `evidence` | Per-org | Evidence artifacts with file metadata |
| `evidence/:id/versions` | Per-org | Version history of an artifact |
| `evidence/:id/links` | Per-org | Links between evidence and controls/requirements |
| `evidence/:id/evaluations` | Per-org | Review/approval evaluations |
| `evidence/staleness` | Per-org | Staleness alerts dashboard |

---

## Endpoints

---

### 1. Evidence Artifacts — CRUD

---

#### `GET /api/v1/evidence`

List the org's evidence artifacts with filtering, search, and pagination. Returns only current versions by default.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter: `draft`, `pending_review`, `approved`, `rejected`, `expired`, `superseded` |
| `evidence_type` | string | — | Filter: `screenshot`, `api_response`, `configuration_export`, `log_sample`, `policy_document`, `access_list`, `vulnerability_report`, `certificate`, `training_record`, `penetration_test`, `audit_report`, `other` |
| `collection_method` | string | — | Filter: `manual_upload`, `automated_pull`, `api_ingestion`, `screenshot_capture`, `system_export` |
| `freshness` | string | — | Filter: `fresh`, `expiring_soon` (within 30 days), `expired` |
| `control_id` | uuid | — | Filter evidence linked to a specific control |
| `requirement_id` | uuid | — | Filter evidence linked to a specific requirement |
| `uploaded_by` | uuid | — | Filter by uploader |
| `tags` | string | — | Comma-separated tags to filter by (AND logic) |
| `include_versions` | boolean | `false` | If true, include superseded versions |
| `search` | string | — | Full-text search on title + description |
| `sort` | string | `created_at` | Sort: `title`, `evidence_type`, `status`, `collection_date`, `expires_at`, `created_at`, `file_size` |
| `order` | string | `desc` | Sort order: `asc`, `desc` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Okta MFA Configuration Export",
      "description": "Export of MFA policy settings from Okta...",
      "evidence_type": "configuration_export",
      "status": "approved",
      "collection_method": "system_export",
      "file_name": "okta-mfa-config-2026-02.json",
      "file_size": 15234,
      "mime_type": "application/json",
      "version": 1,
      "is_current": true,
      "collection_date": "2026-02-15",
      "expires_at": "2026-05-15T00:00:00Z",
      "freshness_period_days": 90,
      "freshness_status": "fresh",
      "source_system": "okta",
      "uploaded_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "tags": ["mfa", "okta", "access-control", "q1-2026"],
      "links_count": 1,
      "evaluations_count": 1,
      "latest_evaluation": {
        "verdict": "sufficient",
        "confidence": "high",
        "evaluated_at": "2026-02-16T10:00:00Z"
      },
      "created_at": "2026-02-15T09:00:00Z",
      "updated_at": "2026-02-15T09:00:00Z"
    }
  ],
  "meta": {
    "total": 42,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `freshness_status` is computed on the fly: `fresh` (no expiry or >30 days remaining), `expiring_soon` (≤30 days remaining), `expired` (past expiry date).
- By default, only `is_current = TRUE` artifacts are returned. Set `include_versions=true` to see all versions.
- `links_count` and `evaluations_count` are computed from joins for the list view.
- `latest_evaluation` is the most recent evaluation (by `created_at`) for quick status display.
- Results always scoped to caller's `org_id`.

---

#### `POST /api/v1/evidence`

Create an evidence artifact record and get a presigned upload URL for MinIO.

This is a **two-step upload flow**:
1. Call this endpoint with metadata → receive a presigned upload URL
2. Upload the file directly to MinIO using the presigned URL
3. Call `POST /api/v1/evidence/:id/confirm` to finalize

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`, `it_admin`, `devops_engineer`

**Request:**

```json
{
  "title": "Okta MFA Configuration Export",
  "description": "Export of MFA policy settings from Okta showing enforcement for all users.",
  "evidence_type": "configuration_export",
  "collection_method": "system_export",
  "file_name": "okta-mfa-config-2026-02.json",
  "file_size": 15234,
  "mime_type": "application/json",
  "collection_date": "2026-02-15",
  "freshness_period_days": 90,
  "source_system": "okta",
  "tags": ["mfa", "okta", "access-control"]
}
```

**Validation:**
- `title`: required, max 500 chars
- `description`: optional, max 10000 chars
- `evidence_type`: required, valid enum value
- `collection_method`: optional, default `manual_upload`
- `file_name`: required, max 255 chars, sanitized (no path separators, control chars)
- `file_size`: required, positive integer, max 104857600 (100 MB)
- `mime_type`: required, must be in the allowed MIME types list (see schema doc)
- `collection_date`: required, ISO date, cannot be in the future
- `freshness_period_days`: optional, positive integer (1–3650). If provided, `expires_at` is auto-computed as `collection_date + freshness_period_days`
- `source_system`: optional, max 255 chars
- `tags`: optional, array of strings, max 20 tags, each max 50 chars

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "title": "Okta MFA Configuration Export",
    "evidence_type": "configuration_export",
    "status": "draft",
    "file_name": "okta-mfa-config-2026-02.json",
    "object_key": "a0000000.../e0000000.../1/okta-mfa-config-2026-02.json",
    "version": 1,
    "collection_date": "2026-02-15",
    "expires_at": "2026-05-16T00:00:00Z",
    "upload": {
      "presigned_url": "https://minio:9000/rp-evidence/...?X-Amz-Signature=...",
      "method": "PUT",
      "expires_in": 900,
      "max_size": 104857600,
      "content_type": "application/json"
    },
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `400 BAD_REQUEST` — Invalid file type, file too large, or validation failure
- `403 FORBIDDEN` — Role not authorized
- `422 UNPROCESSABLE` — collection_date in the future

**Audit log:** `evidence.uploaded` with `{"title": "...", "type": "configuration_export", "file_name": "..."}`

**Notes:**
- The artifact is created in `draft` status immediately but no file is stored yet.
- The presigned URL allows the client to upload directly to MinIO (no API proxy overhead for large files).
- `expires_in` is seconds until the presigned URL expires (15 minutes).
- Content-Type in the presigned URL must match `mime_type` or MinIO will reject the upload.
- The client must call `POST /evidence/:id/confirm` after uploading to finalize the artifact.

---

#### `POST /api/v1/evidence/:id/confirm`

Confirm that file upload to MinIO is complete. Verifies the file exists in MinIO and optionally stores the checksum.

- **Auth:** Bearer token required
- **Roles:** Same as POST (uploader or admin roles)

**Request:**

```json
{
  "checksum_sha256": "a1b2c3d4e5f6..."
}
```

**Validation:**
- `checksum_sha256`: optional, 64-char hex string. If provided, stored for integrity verification.
- Artifact must be in `draft` status.
- File must exist in MinIO at the expected `object_key`.

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "draft",
    "file_verified": true,
    "file_size_actual": 15234,
    "checksum_sha256": "a1b2c3d4e5f6...",
    "message": "Upload confirmed. Artifact ready for review."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Artifact doesn't exist or not in draft status
- `409 CONFLICT` — Already confirmed
- `422 UNPROCESSABLE` — File not found in MinIO (upload may have failed)

**Notes:**
- After confirmation, the artifact can be linked, submitted for review, etc.
- If the checksum doesn't match a later verification, the API will flag it.
- The API verifies `file_size_actual` from MinIO matches `file_size` from the initial POST.

---

#### `GET /api/v1/evidence/:id`

Get a single evidence artifact with full details, version history summary, links, and latest evaluation.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "title": "Okta MFA Configuration Export",
    "description": "Export of MFA policy settings from Okta...",
    "evidence_type": "configuration_export",
    "status": "approved",
    "collection_method": "system_export",
    "file_name": "okta-mfa-config-2026-02.json",
    "file_size": 15234,
    "mime_type": "application/json",
    "object_key": "a0000000.../e0000000.../1/okta-mfa-config-2026-02.json",
    "checksum_sha256": "a1b2c3d4...",
    "version": 1,
    "is_current": true,
    "total_versions": 1,
    "collection_date": "2026-02-15",
    "expires_at": "2026-05-15T00:00:00Z",
    "freshness_period_days": 90,
    "freshness_status": "fresh",
    "days_until_expiry": 84,
    "source_system": "okta",
    "uploaded_by": {
      "id": "uuid",
      "name": "Bob Security",
      "email": "security@acme.example.com"
    },
    "tags": ["mfa", "okta", "access-control", "q1-2026"],
    "metadata": {},
    "links": [
      {
        "id": "uuid",
        "target_type": "control",
        "control": {
          "id": "uuid",
          "identifier": "CTRL-AC-001",
          "title": "Multi-Factor Authentication",
          "status": "active"
        },
        "strength": "primary",
        "notes": null,
        "linked_by": {
          "id": "uuid",
          "name": "Alice Compliance"
        },
        "created_at": "2026-02-15T10:00:00Z"
      }
    ],
    "latest_evaluation": {
      "id": "uuid",
      "verdict": "sufficient",
      "confidence": "high",
      "comments": "MFA is enforced for all user types...",
      "evaluated_by": {
        "id": "uuid",
        "name": "Alice Compliance"
      },
      "created_at": "2026-02-16T10:00:00Z"
    },
    "created_at": "2026-02-15T09:00:00Z",
    "updated_at": "2026-02-15T09:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Artifact doesn't exist in this org

**Notes:**
- `days_until_expiry` is computed: NULL if no `expires_at`, negative if expired.
- `total_versions` counts all versions sharing the same `parent_artifact_id`.
- `links` and `latest_evaluation` are embedded for the detail view.
- This is the primary "evidence detail page" data source.

---

#### `PUT /api/v1/evidence/:id`

Update evidence artifact metadata (not the file — file updates use versioning).

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer` (or the uploader)

**Request:**

```json
{
  "title": "Okta MFA Configuration Export (Updated)",
  "description": "Updated description...",
  "evidence_type": "configuration_export",
  "collection_date": "2026-02-15",
  "freshness_period_days": 60,
  "source_system": "okta",
  "tags": ["mfa", "okta", "access-control", "q1-2026", "updated"]
}
```

**Validation:**
- Same rules as POST, all fields optional
- `file_name`, `file_size`, `mime_type`, `object_key` are NOT updatable (immutable after upload)
- `status` is NOT updatable via this endpoint (use status change endpoint)

**Response 200:**

Returns updated artifact (same shape as GET response).

**Errors:**
- `403 FORBIDDEN` — Not authorized (not uploader or admin role)
- `404 NOT_FOUND` — Artifact not found

**Audit log:** `evidence.updated` with changed fields

**Notes:**
- Updating `freshness_period_days` recalculates `expires_at` from `collection_date`.
- Tags are replaced (not merged) — send the complete tag array.
- The uploader can always update their own artifacts.

---

#### `PUT /api/v1/evidence/:id/status`

Change an evidence artifact's lifecycle status.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "status": "pending_review"
}
```

**Validation:**
- `status`: required, valid enum value
- Allowed transitions:
  - `draft` → `pending_review`
  - `pending_review` → `approved`, `rejected`
  - `rejected` → `pending_review` (resubmit after fixes)
  - `approved` → `expired` (manual expiration)
  - `expired` → `pending_review` (refresh and resubmit)
  - Any → `superseded` (system-only, when new version is uploaded)

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "pending_review",
    "previous_status": "draft",
    "message": "Status updated"
  }
}
```

**Errors:**
- `422 UNPROCESSABLE` — Invalid status transition

**Audit log:** `evidence.status_changed` with `{"old_status": "draft", "new_status": "pending_review"}`

---

#### `DELETE /api/v1/evidence/:id`

Soft-delete an evidence artifact. Marks as superseded and removes from active views. Does NOT delete the file from MinIO (retained for audit trail).

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "superseded",
    "message": "Evidence artifact removed from active view. File retained for audit trail."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Artifact not found
- `403 FORBIDDEN` — Not authorized

**Audit log:** `evidence.deleted` with `{"title": "...", "artifact_id": "uuid"}`

**Notes:**
- Files are NEVER deleted from MinIO via the API — only via manual/lifecycle cleanup.
- Soft delete preserves chain-of-custody for regulatory requirements.

---

### 2. Evidence Upload & Download

---

#### `POST /api/v1/evidence/:id/upload`

Get a fresh presigned upload URL for an artifact that was created but whose upload failed or expired.

- **Auth:** Bearer token required
- **Roles:** Same as POST (uploader or admin roles)

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "upload": {
      "presigned_url": "https://minio:9000/rp-evidence/...?X-Amz-Signature=...",
      "method": "PUT",
      "expires_in": 900,
      "max_size": 104857600,
      "content_type": "application/json"
    }
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Artifact not found
- `409 CONFLICT` — File already confirmed (upload complete)

---

#### `GET /api/v1/evidence/:id/download`

Get a presigned download URL for the artifact's file.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `version` | int | — | Specific version number to download (default: current version) |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "file_name": "okta-mfa-config-2026-02.json",
    "file_size": 15234,
    "mime_type": "application/json",
    "download": {
      "presigned_url": "https://minio:9000/rp-evidence/...?X-Amz-Signature=...",
      "method": "GET",
      "expires_in": 3600
    }
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Artifact or requested version not found
- `422 UNPROCESSABLE` — File not yet uploaded (draft with no confirm)

**Notes:**
- Download URLs expire in 1 hour (3600 seconds).
- The `Content-Disposition` header on the presigned URL will suggest the original filename.
- Auditor role has read access to all evidence (spec §1.2).

---

### 3. Evidence Versioning

---

#### `POST /api/v1/evidence/:id/versions`

Upload a new version of an existing evidence artifact. Creates a new artifact record linked to the original, marks the old version as superseded.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`, `it_admin`, `devops_engineer`

**Request:**

```json
{
  "title": "Okta MFA Configuration Export (March 2026)",
  "description": "Updated MFA config export for Q1 refresh.",
  "file_name": "okta-mfa-config-2026-03.json",
  "file_size": 16102,
  "mime_type": "application/json",
  "collection_date": "2026-03-15",
  "freshness_period_days": 90,
  "tags": ["mfa", "okta", "access-control", "q1-2026"]
}
```

**Validation:**
- Same as `POST /evidence`, except:
- `evidence_type` and `collection_method` are inherited from the parent (can be overridden)
- `source_system` is inherited from the parent (can be overridden)
- The referenced artifact must exist and belong to the org

**Response 201:**

```json
{
  "data": {
    "id": "uuid-new",
    "parent_artifact_id": "uuid-original",
    "version": 2,
    "is_current": true,
    "status": "draft",
    "title": "Okta MFA Configuration Export (March 2026)",
    "file_name": "okta-mfa-config-2026-03.json",
    "previous_version": {
      "id": "uuid-original",
      "version": 1,
      "status": "superseded"
    },
    "upload": {
      "presigned_url": "https://minio:9000/rp-evidence/...?X-Amz-Signature=...",
      "method": "PUT",
      "expires_in": 900,
      "max_size": 104857600,
      "content_type": "application/json"
    },
    "created_at": "2026-03-15T10:00:00Z"
  }
}
```

**Side effects:**
- Previous current version: `is_current = FALSE`, `status = 'superseded'`
- Evidence links from the previous version are **automatically copied** to the new version (evidence links follow the latest version)
- A `POST /evidence/:new_id/confirm` call is needed to finalize the upload

**Errors:**
- `404 NOT_FOUND` — Original artifact not found
- `403 FORBIDDEN` — Not authorized

**Audit log:** `evidence.version_created` with `{"parent_id": "uuid", "version": 2}`

**Notes:**
- The new version gets its own `id`, `object_key`, and file, but shares `parent_artifact_id` with the original for version chain tracking.
- Link copying preserves the evidence-to-control relationships across version refreshes. The user can then modify links on the new version independently.

---

#### `GET /api/v1/evidence/:id/versions`

Get the version history for an evidence artifact.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid-v2",
      "version": 2,
      "is_current": true,
      "title": "Okta MFA Configuration Export (March 2026)",
      "status": "draft",
      "file_name": "okta-mfa-config-2026-03.json",
      "file_size": 16102,
      "collection_date": "2026-03-15",
      "uploaded_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "created_at": "2026-03-15T10:00:00Z"
    },
    {
      "id": "uuid-v1",
      "version": 1,
      "is_current": false,
      "title": "Okta MFA Configuration Export",
      "status": "superseded",
      "file_name": "okta-mfa-config-2026-02.json",
      "file_size": 15234,
      "collection_date": "2026-02-15",
      "uploaded_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "created_at": "2026-02-15T09:00:00Z"
    }
  ],
  "meta": {
    "total_versions": 2,
    "current_version": 2,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Ordered by version number descending (newest first).
- Querying any version in the chain returns the full chain.
- `total_versions` and `current_version` in meta for quick reference.

---

### 4. Evidence Linking

---

#### `GET /api/v1/evidence/:id/links`

List all entities linked to this evidence artifact.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "target_type": "control",
      "control": {
        "id": "uuid",
        "identifier": "CTRL-AC-001",
        "title": "Multi-Factor Authentication",
        "category": "technical",
        "status": "active"
      },
      "requirement": null,
      "strength": "primary",
      "notes": "MFA config directly proves this control is in place.",
      "linked_by": {
        "id": "uuid",
        "name": "Alice Compliance"
      },
      "created_at": "2026-02-15T10:00:00Z"
    },
    {
      "id": "uuid",
      "target_type": "requirement",
      "control": null,
      "requirement": {
        "id": "uuid",
        "identifier": "8.3.1",
        "title": "All user access to system components...",
        "framework": "PCI DSS",
        "framework_version": "4.0.1"
      },
      "strength": "supporting",
      "notes": "Supplementary evidence for PCI DSS 8.3.1.",
      "linked_by": {
        "id": "uuid",
        "name": "Alice Compliance"
      },
      "created_at": "2026-02-15T11:00:00Z"
    }
  ],
  "meta": {
    "total": 2,
    "request_id": "uuid"
  }
}
```

---

#### `POST /api/v1/evidence/:id/links`

Link an evidence artifact to one or more controls or requirements. Supports bulk linking.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Request (single):**

```json
{
  "target_type": "control",
  "control_id": "uuid",
  "strength": "primary",
  "notes": "MFA config directly proves this control is in place."
}
```

**Request (bulk):**

```json
{
  "links": [
    {
      "target_type": "control",
      "control_id": "uuid",
      "strength": "primary",
      "notes": "Direct evidence"
    },
    {
      "target_type": "requirement",
      "requirement_id": "uuid",
      "strength": "supporting",
      "notes": "Supplementary evidence"
    }
  ]
}
```

**Validation:**
- `target_type`: required, must be `control` or `requirement`
- `control_id`: required when `target_type = 'control'`, must exist in the org's control library
- `requirement_id`: required when `target_type = 'requirement'`, must exist and belong to an activated framework
- `strength`: optional, default `primary`. Must be: `primary`, `supporting`, `supplementary`
- `notes`: optional, max 2000 chars
- Bulk: max 50 links per request
- Duplicate (artifact_id, control_id) or (artifact_id, requirement_id) pairs are rejected

**Response 201:**

```json
{
  "data": {
    "created": 2,
    "links": [
      {
        "id": "uuid",
        "target_type": "control",
        "control_id": "uuid",
        "strength": "primary",
        "created_at": "2026-02-20T10:00:00Z"
      },
      {
        "id": "uuid",
        "target_type": "requirement",
        "requirement_id": "uuid",
        "strength": "supporting",
        "created_at": "2026-02-20T10:00:00Z"
      }
    ]
  }
}
```

**Errors:**
- `409 CONFLICT` — Link already exists for this artifact-target pair
- `404 NOT_FOUND` — Artifact, control, or requirement not found
- `403 FORBIDDEN` — Not authorized
- `422 UNPROCESSABLE` — Control doesn't belong to org, requirement not in an activated framework

**Audit log:** `evidence.linked` with `{"artifact_id": "uuid", "target_type": "control", "target_id": "uuid"}`

**Notes:**
- When linking to a control, the evidence implicitly covers all requirements mapped to that control (cross-framework reuse).
- Direct-to-requirement links are for edge cases where evidence doesn't map through a control.

---

#### `DELETE /api/v1/evidence/:id/links/:link_id`

Remove an evidence link.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Response 200:**

```json
{
  "data": {
    "message": "Evidence link removed"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Link not found or doesn't belong to this artifact/org

**Audit log:** `evidence.unlinked` with `{"artifact_id": "uuid", "target_type": "control", "target_id": "uuid"}`

---

### 5. Evidence Relationship Queries

These endpoints provide cross-referencing views — "what evidence exists for X?"

---

#### `GET /api/v1/controls/:id/evidence`

List all evidence artifacts linked to a specific control.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | — | Filter evidence status: `approved`, `pending_review`, etc. |
| `freshness` | string | — | Filter: `fresh`, `expiring_soon`, `expired` |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |

**Response 200:**

```json
{
  "data": {
    "control": {
      "id": "uuid",
      "identifier": "CTRL-AC-001",
      "title": "Multi-Factor Authentication"
    },
    "evidence_summary": {
      "total": 3,
      "approved": 2,
      "pending_review": 1,
      "fresh": 2,
      "expiring_soon": 1,
      "expired": 0
    },
    "evidence": [
      {
        "id": "uuid",
        "title": "Okta MFA Configuration Export",
        "evidence_type": "configuration_export",
        "status": "approved",
        "collection_date": "2026-02-15",
        "expires_at": "2026-05-15T00:00:00Z",
        "freshness_status": "fresh",
        "link": {
          "id": "uuid",
          "strength": "primary",
          "notes": "Direct MFA evidence"
        },
        "latest_evaluation": {
          "verdict": "sufficient",
          "confidence": "high"
        }
      }
    ]
  },
  "meta": {
    "total": 3,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `evidence_summary` provides quick stats for the control detail page.
- Only returns current versions (`is_current = TRUE`).
- Enhances the Sprint 2 control detail page with evidence coverage visibility.

---

#### `GET /api/v1/requirements/:id/evidence`

List all evidence artifacts linked to a specific requirement (both direct links and transitive through controls).

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `include_transitive` | boolean | `true` | Include evidence linked via controls (not just direct links) |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |

**Response 200:**

```json
{
  "data": {
    "requirement": {
      "id": "uuid",
      "identifier": "8.3.1",
      "title": "All user access to system components...",
      "framework": "PCI DSS",
      "framework_version": "4.0.1"
    },
    "evidence": [
      {
        "id": "uuid",
        "title": "Okta MFA Configuration Export",
        "evidence_type": "configuration_export",
        "status": "approved",
        "freshness_status": "fresh",
        "link_type": "transitive",
        "via_control": {
          "id": "uuid",
          "identifier": "CTRL-AC-001",
          "title": "Multi-Factor Authentication"
        },
        "strength": "primary"
      },
      {
        "id": "uuid",
        "title": "MFA Enforcement Screenshot",
        "evidence_type": "screenshot",
        "status": "approved",
        "freshness_status": "fresh",
        "link_type": "direct",
        "via_control": null,
        "strength": "supporting"
      }
    ]
  },
  "meta": {
    "total": 2,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- **Transitive evidence**: If control CTRL-AC-001 is mapped to requirement 8.3.1, and evidence is linked to CTRL-AC-001, then that evidence transitively covers 8.3.1. `link_type: "transitive"` with `via_control` shows the path.
- **Direct evidence**: Evidence linked directly to the requirement via `evidence_links.requirement_id`. `link_type: "direct"`.
- Set `include_transitive=false` to see only direct links.

---

### 6. Freshness & Staleness

---

#### `GET /api/v1/evidence/staleness`

Staleness alert dashboard — list all evidence artifacts that are expired or expiring soon.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `alert_level` | string | — | Filter: `expired`, `expiring_soon` (within 30 days) |
| `days_ahead` | int | 30 | How many days ahead to look for expiring evidence |
| `evidence_type` | string | — | Filter by evidence type |
| `sort` | string | `expires_at` | Sort: `expires_at`, `title`, `evidence_type` |
| `order` | string | `asc` | Sort order (asc = most urgent first) |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |

**Response 200:**

```json
{
  "data": {
    "summary": {
      "total_alerts": 8,
      "expired": 3,
      "expiring_soon": 5,
      "affected_controls": 12
    },
    "alerts": [
      {
        "id": "uuid",
        "title": "Penetration Test Report - 2025",
        "evidence_type": "penetration_test",
        "status": "approved",
        "collection_date": "2025-06-15",
        "expires_at": "2026-02-12T00:00:00Z",
        "freshness_period_days": 365,
        "alert_level": "expired",
        "days_overdue": 8,
        "linked_controls": [
          {
            "id": "uuid",
            "identifier": "CTRL-VM-005",
            "title": "Annual Penetration Testing"
          }
        ],
        "linked_controls_count": 1,
        "uploaded_by": {
          "id": "uuid",
          "name": "Bob Security"
        }
      },
      {
        "id": "uuid",
        "title": "Vulnerability Scan Results - January 2026",
        "evidence_type": "vulnerability_report",
        "status": "approved",
        "collection_date": "2026-01-15",
        "expires_at": "2026-02-14T00:00:00Z",
        "freshness_period_days": 30,
        "alert_level": "expiring_soon",
        "days_until_expiry": 6,
        "linked_controls": [
          {
            "id": "uuid",
            "identifier": "CTRL-VM-001",
            "title": "Monthly Vulnerability Scanning"
          },
          {
            "id": "uuid",
            "identifier": "CTRL-VM-002",
            "title": "Vulnerability Remediation SLA"
          }
        ],
        "linked_controls_count": 2,
        "uploaded_by": {
          "id": "uuid",
          "name": "Bob Security"
        }
      }
    ]
  },
  "meta": {
    "total": 8,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `expired`: `expires_at < NOW()` and status is not `superseded`/`draft`.
- `expiring_soon`: `expires_at` is within `days_ahead` days from now.
- `days_overdue` (for expired) or `days_until_expiry` (for expiring) — one or the other is present.
- `affected_controls` counts distinct controls linked to any alerting evidence (shows blast radius).
- Sorted by `expires_at ASC` by default (most urgent first).
- Only includes current versions (`is_current = TRUE`) with status `approved` or `pending_review`.

---

#### `GET /api/v1/evidence/freshness-summary`

High-level freshness overview for the evidence library — suitable for dashboard widgets.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "total_evidence": 42,
    "by_freshness": {
      "fresh": 30,
      "expiring_soon": 5,
      "expired": 3,
      "no_expiry": 4
    },
    "by_status": {
      "draft": 2,
      "pending_review": 5,
      "approved": 30,
      "rejected": 1,
      "expired": 3,
      "superseded": 1
    },
    "by_type": {
      "configuration_export": 10,
      "screenshot": 8,
      "policy_document": 6,
      "vulnerability_report": 5,
      "access_list": 4,
      "training_record": 3,
      "penetration_test": 2,
      "certificate": 2,
      "log_sample": 1,
      "audit_report": 1
    },
    "coverage": {
      "total_active_controls": 280,
      "controls_with_evidence": 210,
      "controls_without_evidence": 70,
      "evidence_coverage_pct": 75.0
    }
  }
}
```

**Notes:**
- Only counts current versions (`is_current = TRUE`).
- `coverage` shows what percentage of active controls have at least one current, non-expired, approved evidence artifact linked.
- Designed for the main dashboard or evidence library header.

---

### 7. Evidence Evaluations

---

#### `GET /api/v1/evidence/:id/evaluations`

List all evaluations for an evidence artifact (chronological history).

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 50) |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "verdict": "sufficient",
      "confidence": "high",
      "comments": "MFA is enforced for all user types. Configuration export shows Okta MFA policy is set to 'Always' with no exceptions.",
      "missing_elements": [],
      "remediation_notes": null,
      "evidence_link": {
        "id": "uuid",
        "target_type": "control",
        "control_identifier": "CTRL-AC-001"
      },
      "evaluated_by": {
        "id": "uuid",
        "name": "Alice Compliance",
        "role": "compliance_manager"
      },
      "created_at": "2026-02-16T10:00:00Z"
    }
  ],
  "meta": {
    "total": 1,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Evaluations are ordered by `created_at DESC` (newest first).
- `evidence_link` shows which specific link the evaluation was against (optional — may be null for general evaluations).

---

#### `POST /api/v1/evidence/:id/evaluations`

Submit an evaluation of an evidence artifact.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `auditor`

**Request:**

```json
{
  "evidence_link_id": "uuid",
  "verdict": "sufficient",
  "confidence": "high",
  "comments": "MFA is enforced for all user types. Configuration export shows Okta MFA policy is set to 'Always' with no exceptions. Meets PCI DSS 8.3 and SOC 2 CC6.1 requirements.",
  "missing_elements": [],
  "remediation_notes": null
}
```

**Validation:**
- `evidence_link_id`: optional, must be a valid link for this artifact. If provided, evaluation is specific to that link. If null, evaluation is a general quality assessment.
- `verdict`: required, must be `sufficient`, `partial`, `insufficient`, or `needs_update`
- `confidence`: optional, default `medium`. Must be `high`, `medium`, or `low`
- `comments`: required, max 5000 chars
- `missing_elements`: optional, array of strings, max 20 items, each max 200 chars
- `remediation_notes`: optional, max 5000 chars. Recommended when verdict is `insufficient` or `partial`

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "artifact_id": "uuid",
    "evidence_link_id": "uuid",
    "verdict": "sufficient",
    "confidence": "high",
    "comments": "MFA is enforced for all user types...",
    "missing_elements": [],
    "remediation_notes": null,
    "evaluated_by": {
      "id": "uuid",
      "name": "Alice Compliance"
    },
    "created_at": "2026-02-16T10:00:00Z"
  }
}
```

**Side effects:**
- If `verdict` is `sufficient` and artifact status is `pending_review`, the artifact status is **automatically changed to `approved`**.
- If `verdict` is `insufficient` and artifact status is `pending_review`, the artifact status is **automatically changed to `rejected`**.
- These automatic status changes are logged as separate audit events.

**Errors:**
- `404 NOT_FOUND` — Artifact or evidence_link_id not found
- `403 FORBIDDEN` — Role not authorized
- `422 UNPROCESSABLE` — evidence_link_id doesn't belong to this artifact

**Audit log:** `evidence.evaluated` with `{"verdict": "sufficient", "confidence": "high", "link_id": "uuid"}`

**Notes:**
- Evaluations are immutable — create a new one to re-evaluate (full history preserved).
- `auditor` role can evaluate evidence (spec §1.2: auditors review evidence).
- The automatic status change on evaluation is a convenience — it can be overridden by manually changing the status.

---

### 8. Evidence Search

---

#### `GET /api/v1/evidence/search`

Advanced search across evidence artifacts with combined filters. Functionally similar to `GET /evidence` but with additional search capabilities.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `q` | string | — | Full-text search query (title, description, tags, source_system) |
| `evidence_type` | string[] | — | Multiple types (comma-separated) |
| `status` | string[] | — | Multiple statuses (comma-separated) |
| `collection_method` | string[] | — | Multiple methods (comma-separated) |
| `freshness` | string | — | `fresh`, `expiring_soon`, `expired` |
| `date_from` | date | — | Collection date range start (ISO) |
| `date_to` | date | — | Collection date range end (ISO) |
| `uploaded_by` | uuid | — | Filter by uploader |
| `has_links` | boolean | — | `true` = only with links, `false` = only without |
| `has_evaluations` | boolean | — | `true` = only evaluated, `false` = only unevaluated |
| `control_id` | uuid | — | Evidence linked to specific control |
| `framework_id` | uuid | — | Evidence linked to controls mapped to specific framework |
| `tags` | string[] | — | Comma-separated tags (AND logic) |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `sort` | string | `relevance` | Sort: `relevance` (when q provided), `collection_date`, `expires_at`, `created_at`, `title` |
| `order` | string | `desc` | Sort order |

**Response 200:**

Same shape as `GET /evidence` with an additional `search_meta` field:

```json
{
  "data": [...],
  "meta": {
    "total": 15,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  },
  "search_meta": {
    "query": "mfa okta",
    "matched_fields": ["title", "tags"],
    "suggestion": null
  }
}
```

**Notes:**
- `relevance` sort uses PostgreSQL's `ts_rank` from the full-text search index.
- `framework_id` filter joins through evidence_links → controls → control_mappings → requirements → framework_versions to find evidence supporting a specific framework.
- Multiple values for type/status/method filters use OR logic within the same field, AND logic between fields.

---

## Summary of All Endpoints

| # | Method | Path | Roles | Description |
|---|--------|------|-------|-------------|
| 1 | GET | `/api/v1/evidence` | All | List evidence artifacts |
| 2 | POST | `/api/v1/evidence` | CISO, CM, SE, IT, DE | Create artifact + get upload URL |
| 3 | POST | `/api/v1/evidence/:id/confirm` | Uploader/Admin | Confirm upload complete |
| 4 | GET | `/api/v1/evidence/:id` | All | Get artifact detail |
| 5 | PUT | `/api/v1/evidence/:id` | CISO, CM, SE, Uploader | Update artifact metadata |
| 6 | PUT | `/api/v1/evidence/:id/status` | CISO, CM | Change artifact status |
| 7 | DELETE | `/api/v1/evidence/:id` | CISO, CM | Soft-delete artifact |
| 8 | POST | `/api/v1/evidence/:id/upload` | Uploader/Admin | Get fresh upload URL |
| 9 | GET | `/api/v1/evidence/:id/download` | All | Get download URL |
| 10 | POST | `/api/v1/evidence/:id/versions` | CISO, CM, SE, IT, DE | Upload new version |
| 11 | GET | `/api/v1/evidence/:id/versions` | All | Get version history |
| 12 | GET | `/api/v1/evidence/:id/links` | All | List evidence links |
| 13 | POST | `/api/v1/evidence/:id/links` | CISO, CM, SE | Create evidence links (bulk) |
| 14 | DELETE | `/api/v1/evidence/:id/links/:lid` | CISO, CM, SE | Remove evidence link |
| 15 | GET | `/api/v1/controls/:id/evidence` | All | List evidence for a control |
| 16 | GET | `/api/v1/requirements/:id/evidence` | All | List evidence for a requirement |
| 17 | GET | `/api/v1/evidence/staleness` | All | Staleness alert dashboard |
| 18 | GET | `/api/v1/evidence/freshness-summary` | All | Freshness overview stats |
| 19 | GET | `/api/v1/evidence/:id/evaluations` | All | List evaluations |
| 20 | POST | `/api/v1/evidence/:id/evaluations` | CISO, CM, Auditor | Submit evaluation |
| 21 | GET | `/api/v1/evidence/search` | All | Advanced search |

**Role abbreviations:** CISO = `ciso`, CM = `compliance_manager`, SE = `security_engineer`, IT = `it_admin`, DE = `devops_engineer`, Uploader = user who created the artifact

**Total: 21 new endpoints**

---

## Go Implementation Notes

### New Files

```
api/internal/
├── handlers/
│   ├── evidence.go              # Endpoints 1-7, 21
│   ├── evidence_upload.go       # Endpoints 2 (upload flow), 3, 8, 9
│   ├── evidence_versions.go     # Endpoints 10, 11
│   ├── evidence_links.go        # Endpoints 12-14
│   ├── evidence_relations.go    # Endpoints 15, 16
│   ├── evidence_staleness.go    # Endpoints 17, 18
│   └── evidence_evaluations.go  # Endpoints 19, 20
├── models/
│   ├── evidence.go              # EvidenceArtifact struct + queries
│   ├── evidence_link.go         # EvidenceLink struct + queries
│   └── evidence_evaluation.go   # EvidenceEvaluation struct + queries
├── services/
│   ├── evidence.go              # Evidence business logic (versioning, freshness)
│   └── minio.go                 # MinIO client: presigned URLs, bucket mgmt, file verification
└── ...
```

### Route Registration

```go
// Evidence artifacts
ev := v1.Group("/evidence")
ev.Use(middleware.Auth(), middleware.Org())
{
    ev.GET("", handlers.ListEvidence)
    ev.POST("", middleware.RBAC("ciso", "compliance_manager", "security_engineer", "it_admin", "devops_engineer"), handlers.CreateEvidence)
    ev.GET("/staleness", handlers.GetStalenessAlerts)
    ev.GET("/freshness-summary", handlers.GetFreshnessSummary)
    ev.GET("/search", handlers.SearchEvidence)
    
    ev.GET("/:id", handlers.GetEvidence)
    ev.PUT("/:id", handlers.UpdateEvidence) // uploader check in handler
    ev.DELETE("/:id", middleware.RBAC("ciso", "compliance_manager"), handlers.DeleteEvidence)
    ev.PUT("/:id/status", middleware.RBAC("ciso", "compliance_manager"), handlers.ChangeEvidenceStatus)
    
    // Upload flow
    ev.POST("/:id/confirm", handlers.ConfirmEvidenceUpload)
    ev.POST("/:id/upload", handlers.GetUploadURL)
    ev.GET("/:id/download", handlers.GetDownloadURL)
    
    // Versioning
    ev.POST("/:id/versions", middleware.RBAC("ciso", "compliance_manager", "security_engineer", "it_admin", "devops_engineer"), handlers.CreateEvidenceVersion)
    ev.GET("/:id/versions", handlers.ListEvidenceVersions)
    
    // Links
    ev.GET("/:id/links", handlers.ListEvidenceLinks)
    ev.POST("/:id/links", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.CreateEvidenceLinks)
    ev.DELETE("/:id/links/:lid", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.DeleteEvidenceLink)
    
    // Evaluations
    ev.GET("/:id/evaluations", handlers.ListEvidenceEvaluations)
    ev.POST("/:id/evaluations", middleware.RBAC("ciso", "compliance_manager", "auditor"), handlers.CreateEvidenceEvaluation)
}

// Evidence on existing resources
ctrl := v1.Group("/controls")
ctrl.Use(middleware.Auth(), middleware.Org())
{
    // ... existing Sprint 2 routes ...
    ctrl.GET("/:id/evidence", handlers.ListControlEvidence)  // NEW
}

req := v1.Group("/requirements")
req.Use(middleware.Auth())
{
    req.GET("/:id/evidence", middleware.Org(), handlers.ListRequirementEvidence)  // NEW
}
```

### MinIO Client Library

```go
// api/internal/services/minio.go

type MinIOService struct {
    client     *minio.Client
    bucket     string
    region     string
    uploadTTL  time.Duration  // 15 min
    downloadTTL time.Duration // 1 hour
}

func (s *MinIOService) GenerateUploadURL(objectKey, contentType string, maxSize int64) (string, error)
func (s *MinIOService) GenerateDownloadURL(objectKey, fileName string) (string, error)
func (s *MinIOService) VerifyObjectExists(objectKey string) (int64, error)  // returns actual size
func (s *MinIOService) EnsureBucket() error
func (s *MinIOService) DeleteObject(objectKey string) error
```

### Docker Compose Addition

MinIO was defined in the Sprint 1 Docker topology but may not be in docker-compose.yml yet. Add:

```yaml
minio:
  image: minio/minio:latest
  command: server /data --console-address ":9001"
  ports:
    - "9000:9000"
    - "9001:9001"
  environment:
    MINIO_ROOT_USER: ${MINIO_ROOT_USER:-rp-admin}
    MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD:-changeme-minio}
  volumes:
    - rp_minio_data:/data
  healthcheck:
    test: ["CMD", "mc", "ready", "local"]
    interval: 30s
    timeout: 20s
    retries: 3
  networks:
    - rp-net
```

Add to API service environment:
```yaml
MINIO_ENDPOINT: minio:9000
MINIO_ACCESS_KEY: ${MINIO_ROOT_USER:-rp-admin}
MINIO_SECRET_KEY: ${MINIO_ROOT_PASSWORD:-changeme-minio}
MINIO_BUCKET: rp-evidence
MINIO_USE_SSL: "false"
```
