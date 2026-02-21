# Sprint 7 — API Specification: Audit Hub

## Overview

Sprint 7 adds the Audit Hub — the collaboration workspace for external audits. Endpoints cover audit engagement management, evidence request/response workflows, finding management with remediation tracking, evidence submission with chain-of-custody, threaded comments, and dashboard statistics.

This implements spec §6.4 (Audit Hub): dedicated auditor workspace with controlled access, engagement management, evidence request/response workflow, finding management with remediation tracking, and multi-audit support.

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## Access Control Model

### Auditor Isolation

Users with the `auditor` role can **only** access audits where their user ID appears in `audits.auditor_ids`. All audit hub endpoints enforce this in addition to org_id isolation.

### Role Permissions Matrix

| Action | compliance_manager | security_engineer | it_admin | ciso | auditor | vendor_manager |
|--------|-------------------|-------------------|----------|------|---------|----------------|
| Create audit | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| Update audit | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| View audit | ✅ | ✅ | ✅ | ✅ | ✅* | ❌ |
| Create request | ✅ | ❌ | ❌ | ✅ | ✅* | ❌ |
| Assign request | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| Submit evidence | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Review evidence | ❌ | ❌ | ❌ | ❌ | ✅* | ❌ |
| Create finding | ❌ | ❌ | ❌ | ❌ | ✅* | ❌ |
| Update finding status | ✅ | ✅ | ❌ | ✅ | ✅* | ❌ |
| Create comment | ✅ | ✅ | ✅ | ✅ | ✅* | ❌ |
| View internal comments | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |

\* Auditor access limited to audits where their ID is in `auditor_ids`.

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `audits` | Per-org | Audit engagements (SOC 2, PCI, ISO, etc.) |
| `audits/:id/requests` | Per-audit | Evidence requests from auditors |
| `audits/:id/requests/:rid/evidence` | Per-request | Evidence submitted for a request |
| `audits/:id/findings` | Per-audit | Audit findings / deficiencies |
| `audits/:id/comments` | Per-audit | Threaded comments on audit entities |
| `audits/dashboard` | Per-org | Audit hub dashboard statistics |

---

## Endpoints

---

### 1. Audit Engagement CRUD

---

#### `GET /api/v1/audits`

List the org's audits with filtering, search, and pagination.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (filtered)

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | Filter by status: `planning`, `fieldwork`, `review`, `draft_report`, `management_response`, `final_report`, `completed`, `cancelled` |
| `audit_type` | string | No | Filter by type: `soc2_type1`, `soc2_type2`, `iso27001_certification`, etc. |
| `framework_id` | UUID | No | Filter by org_framework_id |
| `search` | string | No | Search title and description |
| `page` | int | No | Page number (default: 1) |
| `per_page` | int | No | Items per page (default: 20, max: 100) |
| `sort` | string | No | Sort field: `created_at`, `planned_end`, `title`, `status` (default: `created_at`) |
| `order` | string | No | Sort order: `asc`, `desc` (default: `desc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "SOC 2 Type II — 2026 Annual",
      "description": "Annual SOC 2 Type II examination...",
      "audit_type": "soc2_type2",
      "status": "fieldwork",
      "org_framework_id": "uuid",
      "framework_name": "SOC 2",
      "period_start": "2026-01-01",
      "period_end": "2026-12-31",
      "planned_start": "2026-02-01",
      "planned_end": "2026-04-30",
      "actual_start": "2026-02-15",
      "actual_end": null,
      "audit_firm": "Deloitte & Touche LLP",
      "lead_auditor_id": "uuid",
      "lead_auditor_name": "Jane Auditor",
      "internal_lead_id": "uuid",
      "internal_lead_name": "John Compliance",
      "total_requests": 8,
      "open_requests": 4,
      "total_findings": 4,
      "open_findings": 3,
      "tags": ["annual", "soc2"],
      "created_at": "2026-02-01T00:00:00Z",
      "updated_at": "2026-02-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 3,
    "total_pages": 1
  }
}
```

---

#### `GET /api/v1/audits/:id`

Get a single audit with full detail including milestones and summary stats.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (if in auditor_ids)

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "title": "SOC 2 Type II — 2026 Annual",
    "description": "Annual SOC 2 Type II examination covering the Trust Services Criteria...",
    "audit_type": "soc2_type2",
    "status": "fieldwork",
    "org_framework_id": "uuid",
    "framework_name": "SOC 2",
    "period_start": "2026-01-01",
    "period_end": "2026-12-31",
    "planned_start": "2026-02-01",
    "planned_end": "2026-04-30",
    "actual_start": "2026-02-15",
    "actual_end": null,
    "audit_firm": "Deloitte & Touche LLP",
    "lead_auditor_id": "uuid",
    "lead_auditor_name": "Jane Auditor",
    "internal_lead_id": "uuid",
    "internal_lead_name": "John Compliance",
    "auditor_ids": ["uuid1", "uuid2", "uuid3"],
    "milestones": [
      {
        "name": "Kickoff Meeting",
        "target_date": "2026-02-01",
        "completed_at": "2026-02-01T14:00:00Z"
      },
      {
        "name": "Fieldwork Start",
        "target_date": "2026-02-15",
        "completed_at": "2026-02-15T09:00:00Z"
      },
      {
        "name": "Interim Testing",
        "target_date": "2026-03-01",
        "completed_at": null
      },
      {
        "name": "Final Testing",
        "target_date": "2026-03-15",
        "completed_at": null
      },
      {
        "name": "Draft Report",
        "target_date": "2026-04-01",
        "completed_at": null
      },
      {
        "name": "Final Report",
        "target_date": "2026-04-30",
        "completed_at": null
      }
    ],
    "report_type": "SOC 2 Type II",
    "report_url": null,
    "report_issued_at": null,
    "total_requests": 8,
    "open_requests": 4,
    "total_findings": 4,
    "open_findings": 3,
    "tags": ["annual", "soc2"],
    "metadata": {},
    "created_at": "2026-02-01T00:00:00Z",
    "updated_at": "2026-02-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `404 Not Found` — Audit doesn't exist or user lacks access

---

#### `POST /api/v1/audits`

Create a new audit engagement.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:**

```json
{
  "title": "SOC 2 Type II — 2026 Annual",
  "description": "Annual SOC 2 Type II examination...",
  "audit_type": "soc2_type2",
  "org_framework_id": "uuid",
  "period_start": "2026-01-01",
  "period_end": "2026-12-31",
  "planned_start": "2026-02-01",
  "planned_end": "2026-04-30",
  "audit_firm": "Deloitte & Touche LLP",
  "lead_auditor_id": "uuid",
  "auditor_ids": ["uuid1", "uuid2"],
  "internal_lead_id": "uuid",
  "milestones": [
    { "name": "Kickoff Meeting", "target_date": "2026-02-01" },
    { "name": "Fieldwork Start", "target_date": "2026-02-15" },
    { "name": "Draft Report", "target_date": "2026-04-01" },
    { "name": "Final Report", "target_date": "2026-04-30" }
  ],
  "report_type": "SOC 2 Type II",
  "tags": ["annual", "soc2"]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `title` | string | ✅ | 1-255 chars |
| `description` | string | No | |
| `audit_type` | enum | ✅ | Valid `audit_type` enum value |
| `org_framework_id` | UUID | No | Must be active org_framework |
| `period_start` | date | No | ISO 8601 date |
| `period_end` | date | No | Must be ≥ period_start |
| `planned_start` | date | No | ISO 8601 date |
| `planned_end` | date | No | Must be ≥ planned_start |
| `audit_firm` | string | No | 1-255 chars |
| `lead_auditor_id` | UUID | No | Must be user with `auditor` role |
| `auditor_ids` | UUID[] | No | All must be users with `auditor` role |
| `internal_lead_id` | UUID | No | Must be user in org |
| `milestones` | array | No | Array of `{ name, target_date }` objects |
| `report_type` | string | No | 1-100 chars |
| `tags` | string[] | No | |

**Response (201):** Full audit object (same as GET response)

**Error Responses:**
- `400 Bad Request` — Validation error
- `403 Forbidden` — Role not authorized
- `404 Not Found` — Referenced user or framework not found

**Audit Log:** `audit.created`

---

#### `PUT /api/v1/audits/:id`

Update an audit engagement.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:** Same fields as POST (all optional). Cannot update `audit_type` after creation.

**Response (200):** Updated audit object

**Audit Log:** `audit.updated`

---

#### `PUT /api/v1/audits/:id/status`

Transition an audit's status.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:**

```json
{
  "status": "fieldwork",
  "notes": "Auditors onsite, fieldwork beginning"
}
```

**Valid Transitions:**

| From | To |
|------|----|
| `planning` | `fieldwork`, `cancelled` |
| `fieldwork` | `review`, `cancelled` |
| `review` | `draft_report`, `fieldwork` (reopen), `cancelled` |
| `draft_report` | `management_response`, `cancelled` |
| `management_response` | `final_report`, `draft_report` (revision needed) |
| `final_report` | `completed` |
| `completed` | — (terminal) |
| `cancelled` | — (terminal) |

**Response (200):** Updated audit object

**Error Responses:**
- `400 Bad Request` — Invalid transition
- `404 Not Found` — Audit not found

**Audit Log:** `audit.status_changed` (metadata: `{ "old_status": "...", "new_status": "...", "notes": "..." }`)

---

#### `POST /api/v1/audits/:id/auditors`

Add an auditor to the engagement.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:**

```json
{
  "user_id": "uuid"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `user_id` | UUID | ✅ | Must be user with `auditor` role in org |

**Response (200):** Updated audit object

**Error Responses:**
- `400 Bad Request` — User is not an auditor, or already in auditor_ids
- `404 Not Found` — User or audit not found

**Audit Log:** `audit.auditor_added`

---

#### `DELETE /api/v1/audits/:id/auditors/:user_id`

Remove an auditor from the engagement.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Response (200):** Updated audit object

**Audit Log:** `audit.auditor_removed`

---

### 2. Audit Requests (Evidence Request/Response Workflow)

---

#### `GET /api/v1/audits/:id/requests`

List evidence requests for an audit.

- **Auth:** Bearer token required
- **Roles:** All audit-access roles

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | Filter by status |
| `priority` | string | No | Filter by priority |
| `assigned_to` | UUID | No | Filter by assignee |
| `control_id` | UUID | No | Filter by linked control |
| `requirement_id` | UUID | No | Filter by linked requirement |
| `overdue` | bool | No | If `true`, only return overdue requests |
| `search` | string | No | Search title and description |
| `page` | int | No | Default: 1 |
| `per_page` | int | No | Default: 20, max: 100 |
| `sort` | string | No | `created_at`, `due_date`, `priority`, `status` (default: `due_date`) |
| `order` | string | No | `asc`, `desc` (default: `asc`) |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "audit_id": "uuid",
      "title": "Information Security Policy (current, approved)",
      "description": "Please provide the current information security policy...",
      "priority": "high",
      "status": "accepted",
      "control_id": "uuid",
      "control_title": "Information Security Policy Management",
      "requirement_id": "uuid",
      "requirement_title": "CC1.1 — COSO Principle 1",
      "requested_by": "uuid",
      "requested_by_name": "Jane Auditor",
      "assigned_to": "uuid",
      "assigned_to_name": "John Compliance",
      "due_date": "2026-02-28",
      "submitted_at": "2026-02-20T14:00:00Z",
      "reviewed_at": "2026-02-21T09:30:00Z",
      "reviewer_notes": null,
      "reference_number": "PBC-001",
      "evidence_count": 2,
      "tags": ["policy"],
      "created_at": "2026-02-15T10:00:00Z",
      "updated_at": "2026-02-21T09:30:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

---

#### `GET /api/v1/audits/:id/requests/:rid`

Get a single request with linked evidence.

- **Auth:** Bearer token required

**Response (200):**

```json
{
  "data": {
    "id": "uuid",
    "audit_id": "uuid",
    "title": "Information Security Policy (current, approved)",
    "description": "Please provide the current approved information security policy including effective date, approval signatures, and version history.",
    "priority": "high",
    "status": "accepted",
    "control_id": "uuid",
    "control_title": "Information Security Policy Management",
    "requirement_id": "uuid",
    "requirement_title": "CC1.1 — COSO Principle 1",
    "requested_by": "uuid",
    "requested_by_name": "Jane Auditor",
    "assigned_to": "uuid",
    "assigned_to_name": "John Compliance",
    "due_date": "2026-02-28",
    "submitted_at": "2026-02-20T14:00:00Z",
    "reviewed_at": "2026-02-21T09:30:00Z",
    "reviewer_notes": null,
    "reference_number": "PBC-001",
    "tags": ["policy"],
    "evidence": [
      {
        "link_id": "uuid",
        "artifact_id": "uuid",
        "artifact_title": "Information Security Policy v3.2",
        "file_name": "ISP-v3.2-approved.pdf",
        "file_size": 245760,
        "mime_type": "application/pdf",
        "submitted_by": "uuid",
        "submitted_by_name": "John Compliance",
        "submitted_at": "2026-02-20T14:00:00Z",
        "submission_notes": "Current approved version, signed by CISO on 2026-01-15",
        "status": "accepted",
        "reviewed_by": "uuid",
        "reviewed_by_name": "Jane Auditor",
        "reviewed_at": "2026-02-21T09:30:00Z",
        "review_notes": "Accepted — policy is current, signed, and covers all required areas."
      }
    ],
    "created_at": "2026-02-15T10:00:00Z",
    "updated_at": "2026-02-21T09:30:00Z"
  }
}
```

---

#### `POST /api/v1/audits/:id/requests`

Create a new evidence request.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, auditor (if in auditor_ids)

**Request Body:**

```json
{
  "title": "Vulnerability Scan Reports Q1-Q2",
  "description": "Please provide internal and external vulnerability scan reports for Q1 and Q2 2026, including remediation evidence for critical/high findings.",
  "priority": "high",
  "control_id": "uuid",
  "requirement_id": "uuid",
  "assigned_to": "uuid",
  "due_date": "2026-03-15",
  "reference_number": "PBC-009",
  "tags": ["vulnerability", "scanning"]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `title` | string | ✅ | 1-500 chars |
| `description` | string | ✅ | Non-empty |
| `priority` | enum | No | Default: `medium` |
| `control_id` | UUID | No | Must exist in org |
| `requirement_id` | UUID | No | Must exist |
| `assigned_to` | UUID | No | Must be internal user in org |
| `due_date` | date | No | ISO 8601 date, must be in the future |
| `reference_number` | string | No | 1-50 chars |
| `tags` | string[] | No | |

**Response (201):** Full request object

**Side Effects:**
- Increments `audits.total_requests` and `audits.open_requests`
- If `assigned_to` is set, status auto-transitions to `in_progress`

**Audit Log:** `audit_request.created`

---

#### `PUT /api/v1/audits/:id/requests/:rid`

Update a request's metadata (title, description, priority, due_date, tags, reference_number).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, auditor (if creator)

**Response (200):** Updated request object

**Audit Log:** Updates are tracked in metadata

---

#### `PUT /api/v1/audits/:id/requests/:rid/assign`

Assign or reassign a request to an internal team member.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:**

```json
{
  "assigned_to": "uuid"
}
```

**Response (200):** Updated request object

**Side Effects:**
- If status is `open`, auto-transitions to `in_progress`

**Audit Log:** `audit_request.assigned`

---

#### `PUT /api/v1/audits/:id/requests/:rid/submit`

Mark a request as submitted (internal team signals evidence is ready for auditor review).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Request Body:**

```json
{
  "notes": "All requested scan reports attached with remediation evidence."
}
```

**Response (200):** Updated request object with `status: "submitted"`, `submitted_at` set

**Error Responses:**
- `400 Bad Request` — No evidence linked to this request yet
- `409 Conflict` — Request is not in `open` or `in_progress` status

**Audit Log:** `audit_request.submitted`

---

#### `PUT /api/v1/audits/:id/requests/:rid/review`

Auditor reviews submitted evidence — accept or reject.

- **Auth:** Bearer token required
- **Roles:** auditor (if in auditor_ids)

**Request Body:**

```json
{
  "decision": "accepted",
  "notes": "Evidence is complete and covers the requested period."
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `decision` | string | ✅ | `accepted` or `rejected` |
| `notes` | string | No | Auditor feedback (required if rejected) |

**Response (200):** Updated request object

**Side Effects:**
- If `accepted`: status → `accepted`, `reviewed_at` set, decrement `audits.open_requests`
- If `rejected`: status → `rejected`, `reviewed_at` set, `reviewer_notes` updated

**Error Responses:**
- `400 Bad Request` — Request not in `submitted` status
- `400 Bad Request` — `rejected` without notes

**Audit Log:** `audit_request.accepted` or `audit_request.rejected`

---

#### `PUT /api/v1/audits/:id/requests/:rid/close`

Close a request (withdrawn, superseded, or no longer needed).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, auditor (if in auditor_ids)

**Request Body:**

```json
{
  "reason": "Superseded by PBC-012"
}
```

**Response (200):** Updated request with `status: "closed"`

**Side Effects:** Decrement `audits.open_requests`

**Audit Log:** `audit_request.closed`

---

#### `POST /api/v1/audits/:id/requests/bulk`

Create multiple requests at once (e.g., from a PBC template).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, auditor (if in auditor_ids)

**Request Body:**

```json
{
  "requests": [
    {
      "title": "Information Security Policy",
      "description": "...",
      "priority": "high",
      "due_date": "2026-03-01",
      "reference_number": "PBC-001"
    },
    {
      "title": "Access Control Procedures",
      "description": "...",
      "priority": "high",
      "due_date": "2026-03-01",
      "reference_number": "PBC-002"
    }
  ]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `requests` | array | ✅ | 1-100 items, each follows POST /requests schema |

**Response (201):**

```json
{
  "data": {
    "created": 25,
    "requests": [ /* array of created request objects */ ]
  }
}
```

---

### 3. Evidence Submission (for Requests)

---

#### `GET /api/v1/audits/:id/requests/:rid/evidence`

List evidence submitted for a specific request.

- **Auth:** Bearer token required

**Response (200):**

```json
{
  "data": [
    {
      "link_id": "uuid",
      "artifact_id": "uuid",
      "artifact_title": "Information Security Policy v3.2",
      "file_name": "ISP-v3.2-approved.pdf",
      "file_size": 245760,
      "mime_type": "application/pdf",
      "evidence_type": "policy_document",
      "evidence_status": "approved",
      "submitted_by": "uuid",
      "submitted_by_name": "John Compliance",
      "submitted_at": "2026-02-20T14:00:00Z",
      "submission_notes": "Current approved version",
      "status": "accepted",
      "reviewed_by": "uuid",
      "reviewed_by_name": "Jane Auditor",
      "reviewed_at": "2026-02-21T09:30:00Z",
      "review_notes": "Accepted — policy is current and complete."
    }
  ]
}
```

---

#### `POST /api/v1/audits/:id/requests/:rid/evidence`

Submit evidence for a request by linking an existing evidence artifact.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso

**Request Body:**

```json
{
  "artifact_id": "uuid",
  "notes": "Current approved version, signed by CISO on 2026-01-15"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `artifact_id` | UUID | ✅ | Must be an evidence_artifact in the same org |
| `notes` | string | No | Explanation of what this evidence demonstrates |

**Response (201):**

```json
{
  "data": {
    "link_id": "uuid",
    "artifact_id": "uuid",
    "artifact_title": "Information Security Policy v3.2",
    "submitted_by": "uuid",
    "submitted_at": "2026-02-20T14:00:00Z",
    "submission_notes": "Current approved version, signed by CISO on 2026-01-15",
    "status": "pending_review"
  }
}
```

**Error Responses:**
- `400 Bad Request` — Artifact doesn't exist in org
- `409 Conflict` — Artifact already linked to this request

**Audit Log:** `audit_evidence.submitted`

---

#### `PUT /api/v1/audits/:id/requests/:rid/evidence/:lid/review`

Auditor reviews a specific evidence submission.

- **Auth:** Bearer token required
- **Roles:** auditor (if in auditor_ids)

**Request Body:**

```json
{
  "status": "accepted",
  "notes": "Accepted — policy is current, signed, and covers all required areas."
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `status` | enum | ✅ | `accepted`, `rejected`, `needs_clarification` |
| `notes` | string | No | Required if `rejected` or `needs_clarification` |

**Response (200):** Updated evidence link object

**Audit Log:** `audit_evidence.reviewed`

---

#### `DELETE /api/v1/audits/:id/requests/:rid/evidence/:lid`

Remove an evidence submission (unlink artifact from request).

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, or the user who submitted it

**Response (204):** No content

---

### 4. Audit Findings

---

#### `GET /api/v1/audits/:id/findings`

List findings for an audit.

- **Auth:** Bearer token required

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `severity` | string | No | Filter by severity |
| `status` | string | No | Filter by status |
| `category` | string | No | Filter by category |
| `remediation_owner_id` | UUID | No | Filter by remediation owner |
| `search` | string | No | Search title and description |
| `page` | int | No | Default: 1 |
| `per_page` | int | No | Default: 20 |
| `sort` | string | No | `created_at`, `severity`, `status`, `remediation_due_date` |
| `order` | string | No | `asc`, `desc` |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "audit_id": "uuid",
      "title": "Missing MFA on admin accounts",
      "description": "Administrative accounts for the production AWS environment do not enforce multi-factor authentication...",
      "severity": "critical",
      "category": "access_control",
      "status": "remediation_complete",
      "control_id": "uuid",
      "control_title": "Multi-Factor Authentication",
      "requirement_id": "uuid",
      "requirement_title": "CC6.1 — Logical Access Security",
      "found_by": "uuid",
      "found_by_name": "Jane Auditor",
      "remediation_owner_id": "uuid",
      "remediation_owner_name": "Mike DevOps",
      "remediation_plan": "Enable MFA on all IAM admin users and enforce via SCP...",
      "remediation_due_date": "2026-03-01",
      "remediation_started_at": "2026-02-18T10:00:00Z",
      "remediation_completed_at": "2026-02-20T16:00:00Z",
      "verified_at": null,
      "reference_number": "F-001",
      "recommendation": "Implement hardware security keys for all admin accounts",
      "management_response": "Agreed. MFA enforcement policy deployed on 2026-02-20.",
      "risk_accepted": false,
      "comment_count": 3,
      "tags": ["mfa", "aws", "iam"],
      "created_at": "2026-02-17T14:00:00Z",
      "updated_at": "2026-02-20T16:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 20, "total": 4, "total_pages": 1 }
}
```

---

#### `GET /api/v1/audits/:id/findings/:fid`

Get a single finding with full detail.

- **Auth:** Bearer token required

**Response (200):** Full finding object including `verification_notes`, `risk_acceptance_reason`, `metadata`, and `comments` (last 5).

---

#### `POST /api/v1/audits/:id/findings`

Create a new finding.

- **Auth:** Bearer token required
- **Roles:** auditor (if in auditor_ids)

**Request Body:**

```json
{
  "title": "Incomplete access reviews",
  "description": "Access reviews for the production database were not completed for Q3 2025...",
  "severity": "high",
  "category": "access_control",
  "control_id": "uuid",
  "requirement_id": "uuid",
  "remediation_owner_id": "uuid",
  "remediation_due_date": "2026-03-15",
  "reference_number": "F-005",
  "recommendation": "Implement quarterly automated access review campaigns...",
  "tags": ["access-review", "database"]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `title` | string | ✅ | 1-500 chars |
| `description` | string | ✅ | Non-empty |
| `severity` | enum | ✅ | Valid `audit_finding_severity` |
| `category` | enum | ✅ | Valid `audit_finding_category` |
| `control_id` | UUID | No | Must exist in org |
| `requirement_id` | UUID | No | Must exist |
| `remediation_owner_id` | UUID | No | Must be internal user in org |
| `remediation_due_date` | date | No | ISO 8601 date |
| `reference_number` | string | No | 1-50 chars |
| `recommendation` | string | No | |
| `tags` | string[] | No | |

**Response (201):** Full finding object

**Side Effects:** Increments `audits.total_findings` and `audits.open_findings`

**Audit Log:** `audit_finding.created`

---

#### `PUT /api/v1/audits/:id/findings/:fid`

Update finding metadata (title, description, severity, category, recommendation, tags, reference_number).

- **Auth:** Bearer token required
- **Roles:** auditor (if in auditor_ids)

**Response (200):** Updated finding object

**Audit Log:** `audit_finding.updated`

---

#### `PUT /api/v1/audits/:id/findings/:fid/status`

Transition a finding's status through the remediation lifecycle.

- **Auth:** Bearer token required
- **Roles:** Depends on transition (see below)

**Request Body:**

```json
{
  "status": "remediation_planned",
  "remediation_plan": "Implement quarterly access review campaigns using automated tooling...",
  "remediation_due_date": "2026-03-15",
  "remediation_owner_id": "uuid",
  "notes": "Plan approved by CISO"
}
```

**Valid Transitions and Required Roles:**

| From | To | Roles | Additional Fields |
|------|----|-------|-------------------|
| `identified` | `acknowledged` | compliance_manager, security_engineer, ciso | — |
| `acknowledged` | `remediation_planned` | compliance_manager, security_engineer, ciso | `remediation_plan` (required), `remediation_due_date`, `remediation_owner_id` |
| `remediation_planned` | `remediation_in_progress` | compliance_manager, security_engineer, ciso | — (sets `remediation_started_at`) |
| `remediation_in_progress` | `remediation_complete` | compliance_manager, security_engineer, ciso | `management_response` (optional) |
| `remediation_complete` | `verified` | auditor | `verification_notes` (optional, sets `verified_at`, `verified_by`) |
| `remediation_complete` | `remediation_in_progress` | auditor | `notes` (required — reason for reopening) |
| `any non-terminal` | `risk_accepted` | ciso | `risk_acceptance_reason` (required) |
| `verified` | `closed` | compliance_manager, ciso, auditor | — |
| `risk_accepted` | `closed` | compliance_manager, ciso, auditor | — |

**Response (200):** Updated finding object

**Side Effects:**
- Transitions to `verified`, `closed`, or `risk_accepted` decrement `audits.open_findings`
- Transitions from `closed`/`risk_accepted` back (reopening) increment `audits.open_findings`

**Audit Log:** `audit_finding.status_changed` (metadata includes `old_status`, `new_status`, and transition-specific fields)

---

#### `PUT /api/v1/audits/:id/findings/:fid/management-response`

Submit the organization's formal management response to a finding.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso

**Request Body:**

```json
{
  "management_response": "Management agrees with the finding. We have implemented quarterly automated access review campaigns starting Q1 2026. The first campaign completed on 2026-01-15 with 100% coverage."
}
```

**Response (200):** Updated finding object

**Audit Log:** Included in `audit_finding.updated` metadata

---

### 5. Audit Comments

---

#### `GET /api/v1/audits/:id/comments`

List comments for an audit (optionally filtered by target).

- **Auth:** Bearer token required

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `target_type` | string | No | `audit`, `request`, `finding` |
| `target_id` | UUID | No | ID of the specific target entity |
| `page` | int | No | Default: 1 |
| `per_page` | int | No | Default: 50 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "audit_id": "uuid",
      "target_type": "request",
      "target_id": "uuid",
      "author_id": "uuid",
      "author_name": "Jane Auditor",
      "author_role": "auditor",
      "body": "Can you also include the exception log for Q2? I see some gaps in coverage.",
      "parent_comment_id": null,
      "is_internal": false,
      "edited_at": null,
      "replies": [
        {
          "id": "uuid",
          "author_id": "uuid",
          "author_name": "John Compliance",
          "author_role": "compliance_manager",
          "body": "Sure — uploading the exception log now. Note: there was a planned system migration in April which accounts for the gap.",
          "parent_comment_id": "uuid",
          "is_internal": false,
          "edited_at": null,
          "created_at": "2026-02-21T10:15:00Z"
        }
      ],
      "created_at": "2026-02-21T10:00:00Z",
      "updated_at": "2026-02-21T10:00:00Z"
    }
  ],
  "pagination": { "page": 1, "per_page": 50, "total": 6, "total_pages": 1 }
}
```

**Note:** Comments with `is_internal = true` are excluded for users with `auditor` role.

---

#### `POST /api/v1/audits/:id/comments`

Create a comment on an audit, request, or finding.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, it_admin, ciso, auditor (if in auditor_ids)

**Request Body:**

```json
{
  "target_type": "request",
  "target_id": "uuid",
  "body": "Can you also include the exception log for Q2?",
  "parent_comment_id": null,
  "is_internal": false
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `target_type` | enum | ✅ | `audit`, `request`, `finding` |
| `target_id` | UUID | ✅ | Must be valid entity within this audit |
| `body` | string | ✅ | 1-10000 chars |
| `parent_comment_id` | UUID | No | Must be existing comment on same target |
| `is_internal` | bool | No | Default: `false`. Auditors cannot set this to `true`. |

**Response (201):** Comment object

**Error Responses:**
- `400 Bad Request` — Auditor trying to create internal comment
- `404 Not Found` — Target entity not found

---

#### `PUT /api/v1/audits/:id/comments/:cid`

Edit a comment (only author can edit).

- **Auth:** Bearer token required

**Request Body:**

```json
{
  "body": "Updated comment text..."
}
```

**Response (200):** Updated comment with `edited_at` set

---

#### `DELETE /api/v1/audits/:id/comments/:cid`

Delete a comment (only author or compliance_manager/ciso).

- **Auth:** Bearer token required

**Response (204):** No content

---

### 6. Request Templates (PBC List)

---

#### `GET /api/v1/audit-request-templates`

List available PBC (Prepared by Client) request templates.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, auditor

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `audit_type` | string | No | Filter by audit type (e.g., `soc2_type2`, `pci_dss_roc`) |
| `framework` | string | No | Filter by framework slug |
| `search` | string | No | Search title and description |

**Response (200):**

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Information Security Policy (current, approved)",
      "description": "Please provide the current information security policy including effective date, approval signatures, and version history. Policy must be approved within the last 12 months.",
      "audit_type": "soc2_type2",
      "framework": "soc2",
      "category": "governance",
      "default_priority": "high",
      "tags": ["policy", "governance"]
    }
  ]
}
```

---

#### `POST /api/v1/audits/:id/requests/from-template`

Create requests from one or more templates.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, ciso, auditor (if in auditor_ids)

**Request Body:**

```json
{
  "template_ids": ["uuid1", "uuid2", "uuid3"],
  "default_due_date": "2026-03-15",
  "auto_number": true,
  "number_prefix": "PBC"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `template_ids` | UUID[] | ✅ | 1-100 template IDs |
| `default_due_date` | date | No | Applied to all created requests |
| `auto_number` | bool | No | Auto-assign reference_number (PBC-001, PBC-002...) |
| `number_prefix` | string | No | Prefix for auto-numbering (default: `PBC`) |

**Response (201):**

```json
{
  "data": {
    "created": 25,
    "requests": [ /* array of created request objects */ ]
  }
}
```

---

### 7. Dashboard & Analytics

---

#### `GET /api/v1/audits/dashboard`

Get audit hub dashboard statistics.

- **Auth:** Bearer token required
- **Roles:** compliance_manager, security_engineer, ciso

**Response (200):**

```json
{
  "data": {
    "summary": {
      "active_audits": 2,
      "completed_audits": 5,
      "total_open_requests": 12,
      "total_overdue_requests": 3,
      "total_open_findings": 7,
      "critical_findings": 1,
      "high_findings": 2
    },
    "active_audits": [
      {
        "id": "uuid",
        "title": "SOC 2 Type II — 2026 Annual",
        "audit_type": "soc2_type2",
        "status": "fieldwork",
        "planned_end": "2026-04-30",
        "days_remaining": 68,
        "readiness_pct": 63,
        "total_requests": 8,
        "open_requests": 4,
        "total_findings": 4,
        "open_findings": 3,
        "next_milestone": {
          "name": "Interim Testing",
          "target_date": "2026-03-01",
          "days_until": 8
        }
      }
    ],
    "overdue_requests": [
      {
        "id": "uuid",
        "title": "Employee Training Records",
        "audit_title": "SOC 2 Type II — 2026 Annual",
        "due_date": "2026-02-15",
        "days_overdue": 6,
        "assigned_to_name": "Sarah HR",
        "priority": "medium"
      }
    ],
    "critical_findings": [
      {
        "id": "uuid",
        "title": "Missing MFA on admin accounts",
        "audit_title": "SOC 2 Type II — 2026 Annual",
        "severity": "critical",
        "status": "remediation_complete",
        "remediation_due_date": "2026-03-01",
        "remediation_owner_name": "Mike DevOps"
      }
    ],
    "recent_activity": [
      {
        "type": "request_submitted",
        "title": "Access Control Procedures",
        "audit_title": "SOC 2 Type II — 2026 Annual",
        "actor_name": "John Compliance",
        "timestamp": "2026-02-21T09:00:00Z"
      },
      {
        "type": "finding_status_changed",
        "title": "Missing MFA on admin accounts",
        "audit_title": "SOC 2 Type II — 2026 Annual",
        "old_status": "remediation_in_progress",
        "new_status": "remediation_complete",
        "actor_name": "Mike DevOps",
        "timestamp": "2026-02-20T16:00:00Z"
      }
    ]
  }
}
```

---

#### `GET /api/v1/audits/:id/stats`

Get statistics for a specific audit engagement.

- **Auth:** Bearer token required

**Response (200):**

```json
{
  "data": {
    "audit_id": "uuid",
    "title": "SOC 2 Type II — 2026 Annual",
    "status": "fieldwork",
    "readiness": {
      "total_requests": 8,
      "accepted": 2,
      "submitted": 1,
      "in_progress": 1,
      "open": 2,
      "rejected": 1,
      "overdue": 1,
      "readiness_pct": 25
    },
    "findings": {
      "total": 4,
      "by_severity": {
        "critical": 1,
        "high": 1,
        "medium": 1,
        "low": 1
      },
      "by_status": {
        "identified": 1,
        "acknowledged": 1,
        "remediation_in_progress": 1,
        "remediation_complete": 1
      },
      "overdue_remediation": 0
    },
    "evidence": {
      "total_submitted": 5,
      "accepted": 3,
      "pending_review": 1,
      "rejected": 1
    },
    "timeline": {
      "planned_start": "2026-02-01",
      "planned_end": "2026-04-30",
      "actual_start": "2026-02-15",
      "days_elapsed": 6,
      "days_remaining": 68,
      "milestones_completed": 2,
      "milestones_total": 6,
      "next_milestone": {
        "name": "Interim Testing",
        "target_date": "2026-03-01",
        "days_until": 8
      }
    },
    "activity": {
      "comments_count": 15,
      "last_activity_at": "2026-02-21T10:15:00Z"
    }
  }
}
```

---

#### `GET /api/v1/audits/:id/readiness`

Get audit readiness breakdown per control/requirement — how much evidence has been accepted.

- **Auth:** Bearer token required

**Response (200):**

```json
{
  "data": {
    "audit_id": "uuid",
    "overall_readiness_pct": 25,
    "by_requirement": [
      {
        "requirement_id": "uuid",
        "requirement_title": "CC1.1 — COSO Principle 1",
        "total_requests": 3,
        "accepted_requests": 2,
        "readiness_pct": 67
      }
    ],
    "by_control": [
      {
        "control_id": "uuid",
        "control_title": "Information Security Policy Management",
        "total_requests": 1,
        "accepted_requests": 1,
        "readiness_pct": 100
      }
    ],
    "gaps": [
      {
        "requirement_id": "uuid",
        "requirement_title": "CC6.3 — Access Review",
        "issue": "no_requests",
        "description": "No evidence requests created for this requirement"
      }
    ]
  }
}
```

---

## Error Codes (Sprint 7 additions)

| Code | HTTP | Description |
|------|------|-------------|
| `AUDIT_NOT_FOUND` | 404 | Audit engagement doesn't exist or user lacks access |
| `AUDIT_REQUEST_NOT_FOUND` | 404 | Evidence request doesn't exist |
| `AUDIT_FINDING_NOT_FOUND` | 404 | Audit finding doesn't exist |
| `AUDIT_INVALID_TRANSITION` | 400 | Invalid status transition for audit/request/finding |
| `AUDIT_NO_EVIDENCE` | 400 | Cannot submit request with no evidence attached |
| `AUDIT_DUPLICATE_EVIDENCE` | 409 | Evidence artifact already linked to this request |
| `AUDIT_REJECTION_REQUIRES_NOTES` | 400 | Rejection must include notes/reason |
| `AUDIT_NOT_AUDITOR` | 403 | User is not an auditor for this engagement |
| `AUDIT_COMMENT_NOT_FOUND` | 404 | Comment doesn't exist |
| `AUDIT_INTERNAL_COMMENT_DENIED` | 400 | Auditors cannot create internal comments |
| `AUDIT_RISK_ACCEPT_REQUIRES_REASON` | 400 | Risk acceptance requires a justification |
| `AUDIT_COMPLETED` | 400 | Cannot modify a completed audit |
| `AUDIT_CANCELLED` | 400 | Cannot modify a cancelled audit |

---

## Endpoint Summary

| # | Method | Path | Description |
|---|--------|------|-------------|
| 1 | GET | `/api/v1/audits` | List audits |
| 2 | GET | `/api/v1/audits/:id` | Get audit detail |
| 3 | POST | `/api/v1/audits` | Create audit |
| 4 | PUT | `/api/v1/audits/:id` | Update audit |
| 5 | PUT | `/api/v1/audits/:id/status` | Transition audit status |
| 6 | POST | `/api/v1/audits/:id/auditors` | Add auditor |
| 7 | DELETE | `/api/v1/audits/:id/auditors/:user_id` | Remove auditor |
| 8 | GET | `/api/v1/audits/:id/requests` | List requests |
| 9 | GET | `/api/v1/audits/:id/requests/:rid` | Get request detail |
| 10 | POST | `/api/v1/audits/:id/requests` | Create request |
| 11 | PUT | `/api/v1/audits/:id/requests/:rid` | Update request |
| 12 | PUT | `/api/v1/audits/:id/requests/:rid/assign` | Assign request |
| 13 | PUT | `/api/v1/audits/:id/requests/:rid/submit` | Submit request |
| 14 | PUT | `/api/v1/audits/:id/requests/:rid/review` | Review request (auditor) |
| 15 | PUT | `/api/v1/audits/:id/requests/:rid/close` | Close request |
| 16 | POST | `/api/v1/audits/:id/requests/bulk` | Bulk create requests |
| 17 | GET | `/api/v1/audits/:id/requests/:rid/evidence` | List request evidence |
| 18 | POST | `/api/v1/audits/:id/requests/:rid/evidence` | Submit evidence |
| 19 | PUT | `/api/v1/audits/:id/requests/:rid/evidence/:lid/review` | Review evidence (auditor) |
| 20 | DELETE | `/api/v1/audits/:id/requests/:rid/evidence/:lid` | Remove evidence |
| 21 | GET | `/api/v1/audits/:id/findings` | List findings |
| 22 | GET | `/api/v1/audits/:id/findings/:fid` | Get finding detail |
| 23 | POST | `/api/v1/audits/:id/findings` | Create finding |
| 24 | PUT | `/api/v1/audits/:id/findings/:fid` | Update finding |
| 25 | PUT | `/api/v1/audits/:id/findings/:fid/status` | Transition finding status |
| 26 | PUT | `/api/v1/audits/:id/findings/:fid/management-response` | Submit management response |
| 27 | GET | `/api/v1/audits/:id/comments` | List comments |
| 28 | POST | `/api/v1/audits/:id/comments` | Create comment |
| 29 | PUT | `/api/v1/audits/:id/comments/:cid` | Edit comment |
| 30 | DELETE | `/api/v1/audits/:id/comments/:cid` | Delete comment |
| 31 | GET | `/api/v1/audit-request-templates` | List PBC templates |
| 32 | POST | `/api/v1/audits/:id/requests/from-template` | Create from templates |
| 33 | GET | `/api/v1/audits/dashboard` | Audit hub dashboard |
| 34 | GET | `/api/v1/audits/:id/stats` | Audit engagement stats |
| 35 | GET | `/api/v1/audits/:id/readiness` | Audit readiness breakdown |

**Total: 35 endpoints**

---

## Notes for DEV-BE

1. **Auditor isolation middleware:** Create a middleware or helper that, for `auditor` role users, filters queries to only audits where `user_id = ANY(auditor_ids)`. Apply to ALL audit hub endpoints.
2. **Internal comment filtering:** When role is `auditor`, exclude comments where `is_internal = true` from all responses.
3. **Denormalized counts:** When requests/findings are created, updated, or status-changed, update `audits.total_requests`, `audits.open_requests`, `audits.total_findings`, `audits.open_findings`. Use a helper function for consistency.
4. **Overdue detection:** Requests past `due_date` with status not in `(accepted, closed)` should be flagged. Consider a periodic check (in the existing monitoring worker) or on-read detection.
5. **Completed/cancelled guard:** All mutating endpoints on audits with status `completed` or `cancelled` should return `400 AUDIT_COMPLETED` or `400 AUDIT_CANCELLED`.
6. **Template seeding:** PBC templates are stored in `db/seeds/` and loaded into a `audit_request_templates` in-memory map or lightweight table. Keep it simple — templates are read-only reference data.
7. **Evidence chain-of-custody:** The `audit_evidence_links` table provides full traceability. Ensure `submitted_by` is always set from the JWT token, never from request body.
8. **Reference Sprint 3 patterns** for evidence artifact queries (joins with `evidence_artifacts` table).
9. **Reference Sprint 4 patterns** for the alert/monitoring worker if adding overdue request detection.
