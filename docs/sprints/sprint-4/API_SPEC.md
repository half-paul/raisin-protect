# Sprint 4 — API Specification: Continuous Monitoring Engine

## Overview

Sprint 4 adds the continuous monitoring API: define and manage automated tests, execute test sweeps, track results, generate severity-based alerts, manage alert lifecycle (assign, resolve, suppress, close), configure alert rules, and power the monitoring dashboard with real-time compliance posture data.

This is the operational heart of the GRC platform — moving from "we documented our controls" (Sprint 2) and "we stored evidence" (Sprint 3) to "we continuously prove our controls work" (spec §3.3).

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `tests` | Per-org | Test definitions with schedules |
| `test-runs` | Per-org | Batch execution sweeps |
| `test-runs/:id/results` | Per-org | Individual test outcomes within a sweep |
| `alerts` | Per-org | Generated alerts from test failures |
| `alert-rules` | Per-org | Configurable alert generation rules |
| `monitoring/heatmap` | Per-org | Control health heatmap data |
| `monitoring/posture` | Per-org | Compliance posture score per framework |
| `monitoring/summary` | Per-org | Monitoring dashboard summary stats |

---

## Endpoints

---

### 1. Test Definitions — CRUD

---

#### `GET /api/v1/tests`

List the org's test definitions with filtering and pagination.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter: `draft`, `active`, `paused`, `deprecated` |
| `test_type` | string | — | Filter: `configuration`, `access_control`, `endpoint`, `vulnerability`, `data_protection`, `network`, `logging`, `custom` |
| `severity` | string | — | Filter: `critical`, `high`, `medium`, `low`, `informational` |
| `control_id` | uuid | — | Filter tests for a specific control |
| `tags` | string | — | Comma-separated tags (AND logic) |
| `search` | string | — | Search in title and description |
| `sort` | string | `identifier` | Sort: `identifier`, `title`, `test_type`, `severity`, `status`, `last_run_at`, `created_at` |
| `order` | string | `asc` | Sort order: `asc`, `desc` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "TST-AC-001",
      "title": "MFA Enforcement Verification",
      "description": "Verifies that MFA is enforced for all users...",
      "test_type": "access_control",
      "severity": "critical",
      "status": "active",
      "control": {
        "id": "uuid",
        "identifier": "CTRL-AC-001",
        "title": "Multi-Factor Authentication"
      },
      "schedule_cron": "0 * * * *",
      "schedule_interval_min": null,
      "next_run_at": "2026-02-20T17:00:00Z",
      "last_run_at": "2026-02-20T16:00:00Z",
      "latest_result": {
        "status": "pass",
        "tested_at": "2026-02-20T16:00:12Z"
      },
      "tags": ["mfa", "access-control", "pci", "soc2"],
      "created_at": "2026-02-20T10:00:00Z",
      "updated_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 8,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `latest_result` shows the most recent test result (pass/fail/error) for quick status display.
- `next_run_at` shows when the test is next scheduled to execute.
- Results always scoped to caller's `org_id`.

---

#### `POST /api/v1/tests`

Create a new test definition.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`, `devops_engineer`

**Request:**

```json
{
  "identifier": "TST-AC-001",
  "title": "MFA Enforcement Verification",
  "description": "Verifies that multi-factor authentication is enforced for all users in the identity provider.",
  "test_type": "access_control",
  "severity": "critical",
  "control_id": "uuid",
  "schedule_cron": "0 * * * *",
  "timeout_seconds": 120,
  "retry_count": 1,
  "retry_delay_seconds": 30,
  "test_config": {
    "provider": "okta",
    "check": "mfa_enforced",
    "expected": true
  },
  "tags": ["mfa", "access-control", "pci", "soc2"]
}
```

**Validation:**
- `identifier`: required, max 50 chars, alphanumeric + hyphens, unique per org
- `title`: required, max 500 chars
- `description`: optional, max 10000 chars
- `test_type`: required, valid enum value
- `severity`: optional, default `medium`, valid enum value
- `control_id`: required, must exist in org's control library, control must be `active`
- `schedule_cron`: optional, valid cron expression (5-field)
- `schedule_interval_min`: optional, positive integer (1–10080, i.e., max 1 week). Mutually exclusive with `schedule_cron`
- `timeout_seconds`: optional, default 300, range 1–3600
- `retry_count`: optional, default 0, range 0–5
- `retry_delay_seconds`: optional, default 60, range 1–3600
- `test_config`: optional, JSONB, max 100KB
- `test_script`: optional (for `custom` type), max 64KB
- `test_script_language`: required if `test_script` provided, must be `shell`, `python`, or `javascript`
- `tags`: optional, max 20 tags, each max 50 chars

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "TST-AC-001",
    "title": "MFA Enforcement Verification",
    "test_type": "access_control",
    "severity": "critical",
    "status": "draft",
    "control": {
      "id": "uuid",
      "identifier": "CTRL-AC-001",
      "title": "Multi-Factor Authentication"
    },
    "schedule_cron": "0 * * * *",
    "next_run_at": null,
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `400 BAD_REQUEST` — Validation failure
- `403 FORBIDDEN` — Role not authorized
- `409 CONFLICT` — Identifier already exists in this org
- `422 UNPROCESSABLE` — Control not found or not active

**Audit log:** `test.created` with `{"identifier": "TST-AC-001", "test_type": "access_control", "control_id": "uuid"}`

**Notes:**
- Tests are created in `draft` status. Use `PUT /tests/:id/status` to activate.
- `next_run_at` is NULL until the test is activated (status = `active`). The worker computes it on activation.

---

#### `GET /api/v1/tests/:id`

Get a single test definition with full details.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "TST-AC-001",
    "title": "MFA Enforcement Verification",
    "description": "Verifies that MFA is enforced for all users...",
    "test_type": "access_control",
    "severity": "critical",
    "status": "active",
    "control": {
      "id": "uuid",
      "identifier": "CTRL-AC-001",
      "title": "Multi-Factor Authentication",
      "category": "technical",
      "status": "active"
    },
    "schedule_cron": "0 * * * *",
    "schedule_interval_min": null,
    "next_run_at": "2026-02-20T17:00:00Z",
    "last_run_at": "2026-02-20T16:00:00Z",
    "timeout_seconds": 120,
    "retry_count": 1,
    "retry_delay_seconds": 30,
    "test_config": {
      "provider": "okta",
      "check": "mfa_enforced",
      "expected": true
    },
    "test_script": null,
    "test_script_language": null,
    "tags": ["mfa", "access-control", "pci", "soc2"],
    "created_by": {
      "id": "uuid",
      "name": "Bob Security"
    },
    "latest_result": {
      "id": "uuid",
      "status": "pass",
      "message": "MFA is enforced for all 142 users. No exceptions found.",
      "tested_at": "2026-02-20T16:00:12Z",
      "duration_ms": 1234
    },
    "result_summary": {
      "total_runs": 48,
      "last_24h": {
        "pass": 23,
        "fail": 0,
        "error": 1,
        "warning": 0,
        "skip": 0
      }
    },
    "active_alerts": 0,
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Test not found in this org

**Notes:**
- `result_summary.last_24h` provides a quick pass/fail breakdown for the last 24 hours.
- `active_alerts` counts unresolved alerts generated from this test.

---

#### `PUT /api/v1/tests/:id`

Update a test definition.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`, `devops_engineer`

**Request:**

```json
{
  "title": "MFA Enforcement Verification (Updated)",
  "description": "Updated description...",
  "severity": "critical",
  "schedule_cron": "*/30 * * * *",
  "timeout_seconds": 180,
  "test_config": {
    "provider": "okta",
    "check": "mfa_enforced",
    "expected": true,
    "include_service_accounts": true
  },
  "tags": ["mfa", "access-control", "pci", "soc2", "updated"]
}
```

**Validation:**
- Same rules as POST, all fields optional
- `identifier`, `test_type`, `control_id` are NOT updatable after creation (immutable identity)
- `status` is NOT updatable via this endpoint (use status endpoint)

**Response 200:**

Returns updated test (same shape as GET).

**Errors:**
- `403 FORBIDDEN` — Not authorized
- `404 NOT_FOUND` — Test not found

**Audit log:** `test.updated` with changed fields

**Notes:**
- Updating `schedule_cron` or `schedule_interval_min` recalculates `next_run_at`.
- Tags are replaced (not merged) — send the complete array.

---

#### `PUT /api/v1/tests/:id/status`

Change a test's lifecycle status.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Request:**

```json
{
  "status": "active"
}
```

**Validation:**
- `status`: required, valid enum value
- Allowed transitions:
  - `draft` → `active`
  - `active` → `paused`, `deprecated`
  - `paused` → `active`, `deprecated`
  - `deprecated` → (terminal, no transitions out)

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "active",
    "previous_status": "draft",
    "next_run_at": "2026-02-20T17:00:00Z",
    "message": "Test activated. First run scheduled."
  }
}
```

**Side effects:**
- On activation: `next_run_at` is computed from schedule. `last_run_at` stays NULL until first execution.
- On pause: `next_run_at` is set to NULL (worker will skip it).
- On deprecation: `next_run_at` is set to NULL. Existing results and alerts are preserved.

**Errors:**
- `422 UNPROCESSABLE` — Invalid status transition

**Audit log:** `test.status_changed` with `{"old_status": "draft", "new_status": "active"}`

---

#### `DELETE /api/v1/tests/:id`

Soft-delete a test by marking it as deprecated. Preserves all historical results and alerts.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "deprecated",
    "message": "Test deprecated. Historical results preserved."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Test not found
- `403 FORBIDDEN` — Not authorized

**Audit log:** `test.deleted` with `{"identifier": "TST-AC-001"}`

---

### 2. Test Execution & Runs

---

#### `POST /api/v1/test-runs`

Trigger a manual test run. Creates a new sweep and queues it for the worker.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`, `devops_engineer`

**Request:**

```json
{
  "test_ids": ["uuid1", "uuid2"],
  "trigger_metadata": {
    "reason": "Pre-audit verification"
  }
}
```

**Validation:**
- `test_ids`: optional, array of test UUIDs. If empty/null, runs ALL active tests (full sweep).
- Max 500 test IDs per request.
- All specified tests must exist, belong to the org, and be in `active` or `paused` status.
- `trigger_metadata`: optional JSONB, max 10KB.

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "run_number": 42,
    "status": "pending",
    "trigger_type": "manual",
    "total_tests": 8,
    "triggered_by": {
      "id": "uuid",
      "name": "Alice Compliance"
    },
    "created_at": "2026-02-20T16:00:00Z"
  }
}
```

**Errors:**
- `400 BAD_REQUEST` — No active tests to run, or test IDs invalid
- `403 FORBIDDEN` — Not authorized
- `409 CONFLICT` — A run is already in progress for this org (max 1 concurrent run)
- `422 UNPROCESSABLE` — One or more test_ids not found or not active

**Audit log:** `test_run.started` with `{"run_number": 42, "trigger": "manual", "test_count": 8}`

**Notes:**
- The run is created in `pending` status. The worker picks it up and transitions to `running`.
- Only one run per org at a time (prevents resource exhaustion). The `409 CONFLICT` is returned if a `pending` or `running` run exists.
- Paused tests can be manually triggered but won't run on their schedule.

---

#### `GET /api/v1/test-runs`

List test runs with filtering and pagination.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter: `pending`, `running`, `completed`, `failed`, `cancelled` |
| `trigger_type` | string | — | Filter: `scheduled`, `manual`, `on_change`, `webhook` |
| `date_from` | datetime | — | Filter runs after this time (ISO 8601) |
| `date_to` | datetime | — | Filter runs before this time (ISO 8601) |
| `sort` | string | `created_at` | Sort: `run_number`, `status`, `trigger_type`, `started_at`, `created_at`, `total_tests` |
| `order` | string | `desc` | Sort order |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "run_number": 42,
      "status": "completed",
      "trigger_type": "scheduled",
      "started_at": "2026-02-20T16:00:00Z",
      "completed_at": "2026-02-20T16:02:34Z",
      "duration_ms": 154000,
      "total_tests": 8,
      "passed": 7,
      "failed": 1,
      "errors": 0,
      "skipped": 0,
      "warnings": 0,
      "triggered_by": null,
      "created_at": "2026-02-20T16:00:00Z"
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

---

#### `GET /api/v1/test-runs/:id`

Get a single test run with full details.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "run_number": 42,
    "status": "completed",
    "trigger_type": "scheduled",
    "started_at": "2026-02-20T16:00:00Z",
    "completed_at": "2026-02-20T16:02:34Z",
    "duration_ms": 154000,
    "total_tests": 8,
    "passed": 7,
    "failed": 1,
    "errors": 0,
    "skipped": 0,
    "warnings": 0,
    "triggered_by": null,
    "trigger_metadata": {},
    "worker_id": "worker-01",
    "error_message": null,
    "created_at": "2026-02-20T16:00:00Z",
    "updated_at": "2026-02-20T16:02:34Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Run not found in this org

---

#### `POST /api/v1/test-runs/:id/cancel`

Cancel a pending or running test run.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "cancelled",
    "previous_status": "running",
    "message": "Test run cancelled."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Run not found
- `422 UNPROCESSABLE` — Run already completed, failed, or cancelled

**Audit log:** `test_run.cancelled` with `{"run_number": 42}`

---

#### `GET /api/v1/test-runs/:id/results`

List individual test results within a test run.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 50 | Items per page (max 200) |
| `status` | string | — | Filter: `pass`, `fail`, `error`, `skip`, `warning` |
| `severity` | string | — | Filter by test severity |
| `sort` | string | `status` | Sort: `status`, `severity`, `test_identifier`, `duration_ms`, `started_at` |
| `order` | string | `asc` | Sort order (default: failures first) |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "test": {
        "id": "uuid",
        "identifier": "TST-NW-001",
        "title": "Security Group — No Open Inbound 0.0.0.0/0",
        "test_type": "network"
      },
      "control": {
        "id": "uuid",
        "identifier": "CTRL-NW-001",
        "title": "Network Access Controls"
      },
      "status": "fail",
      "severity": "critical",
      "message": "Found 2 security groups with unrestricted inbound access on port 22.",
      "details": {
        "violations": [
          {
            "resource": "sg-abc123",
            "name": "dev-ssh-access",
            "port": 22,
            "source": "0.0.0.0/0"
          },
          {
            "resource": "sg-def456",
            "name": "legacy-admin",
            "port": 22,
            "source": "0.0.0.0/0"
          }
        ],
        "total_checked": 15,
        "violations_found": 2
      },
      "duration_ms": 2340,
      "alert_generated": true,
      "alert_id": "uuid",
      "started_at": "2026-02-20T16:01:15Z",
      "completed_at": "2026-02-20T16:01:17Z",
      "created_at": "2026-02-20T16:01:17Z"
    },
    {
      "id": "uuid",
      "test": {
        "id": "uuid",
        "identifier": "TST-AC-001",
        "title": "MFA Enforcement Verification",
        "test_type": "access_control"
      },
      "control": {
        "id": "uuid",
        "identifier": "CTRL-AC-001",
        "title": "Multi-Factor Authentication"
      },
      "status": "pass",
      "severity": "critical",
      "message": "MFA is enforced for all 142 users. No exceptions found.",
      "details": {
        "total_users": 142,
        "mfa_enabled": 142,
        "mfa_disabled": 0,
        "exceptions": []
      },
      "duration_ms": 1234,
      "alert_generated": false,
      "alert_id": null,
      "started_at": "2026-02-20T16:00:01Z",
      "completed_at": "2026-02-20T16:00:02Z",
      "created_at": "2026-02-20T16:00:02Z"
    }
  ],
  "meta": {
    "total": 8,
    "page": 1,
    "per_page": 50,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Default sort puts failures first (`status ASC`: error, fail, skip, warning, pass).
- `details` JSONB contains structured test output — schema varies by test type.
- `alert_generated` and `alert_id` show if this result triggered an alert.

---

#### `GET /api/v1/test-runs/:run_id/results/:result_id`

Get a single test result with full details including output log.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "test_run_id": "uuid",
    "test": {
      "id": "uuid",
      "identifier": "TST-NW-001",
      "title": "Security Group — No Open Inbound 0.0.0.0/0",
      "test_type": "network",
      "severity": "critical"
    },
    "control": {
      "id": "uuid",
      "identifier": "CTRL-NW-001",
      "title": "Network Access Controls"
    },
    "status": "fail",
    "severity": "critical",
    "message": "Found 2 security groups with unrestricted inbound access on port 22.",
    "details": {
      "violations": [
        {
          "resource": "sg-abc123",
          "name": "dev-ssh-access",
          "port": 22,
          "source": "0.0.0.0/0"
        }
      ],
      "total_checked": 15,
      "violations_found": 2
    },
    "output_log": "2026-02-20T16:01:15Z [INFO] Starting network security check...\n2026-02-20T16:01:15Z [INFO] Checking 15 security groups...\n2026-02-20T16:01:17Z [FAIL] sg-abc123 (dev-ssh-access): port 22 open to 0.0.0.0/0\n...",
    "error_message": null,
    "duration_ms": 2340,
    "alert_generated": true,
    "alert_id": "uuid",
    "started_at": "2026-02-20T16:01:15Z",
    "completed_at": "2026-02-20T16:01:17Z",
    "created_at": "2026-02-20T16:01:17Z"
  }
}
```

**Notes:**
- `output_log` is only included in the single-result detail view (not in list view, to keep payloads small).
- `output_log` is truncated to 64KB.

---

### 3. Test Results — Cross-Resource Queries

---

#### `GET /api/v1/tests/:id/results`

List execution history for a specific test (across all test runs).

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter by result status |
| `date_from` | datetime | — | Results after this time |
| `date_to` | datetime | — | Results before this time |
| `sort` | string | `created_at` | Sort: `created_at`, `status`, `duration_ms` |
| `order` | string | `desc` | Sort order (newest first) |

**Response 200:**

```json
{
  "data": {
    "test": {
      "id": "uuid",
      "identifier": "TST-AC-001",
      "title": "MFA Enforcement Verification"
    },
    "results": [
      {
        "id": "uuid",
        "test_run_id": "uuid",
        "run_number": 42,
        "status": "pass",
        "severity": "critical",
        "message": "MFA is enforced for all 142 users.",
        "duration_ms": 1234,
        "alert_generated": false,
        "started_at": "2026-02-20T16:00:01Z",
        "created_at": "2026-02-20T16:00:02Z"
      }
    ]
  },
  "meta": {
    "total": 48,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

---

#### `GET /api/v1/controls/:id/test-results`

List test execution history for a specific control (across all tests and runs).

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter by result status |
| `date_from` | datetime | — | Results after this time |
| `date_to` | datetime | — | Results before this time |

**Response 200:**

```json
{
  "data": {
    "control": {
      "id": "uuid",
      "identifier": "CTRL-AC-001",
      "title": "Multi-Factor Authentication"
    },
    "health_status": "healthy",
    "tests_count": 2,
    "results": [
      {
        "id": "uuid",
        "test": {
          "id": "uuid",
          "identifier": "TST-AC-001",
          "title": "MFA Enforcement Verification"
        },
        "run_number": 42,
        "status": "pass",
        "severity": "critical",
        "message": "MFA is enforced for all 142 users.",
        "duration_ms": 1234,
        "started_at": "2026-02-20T16:00:01Z",
        "created_at": "2026-02-20T16:00:02Z"
      }
    ]
  },
  "meta": {
    "total": 96,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `health_status` is computed from the most recent test result: `healthy`, `failing`, `error`, `warning`, `untested`.
- `tests_count` shows how many tests are defined for this control.

---

### 4. Alerts — CRUD & Lifecycle

---

#### `GET /api/v1/alerts`

List alerts with filtering, search, and pagination.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter: `open`, `acknowledged`, `in_progress`, `resolved`, `suppressed`, `closed`. Comma-separated for multiple. |
| `severity` | string | — | Filter: `critical`, `high`, `medium`, `low`. Comma-separated for multiple. |
| `control_id` | uuid | — | Filter alerts for a specific control |
| `test_id` | uuid | — | Filter alerts from a specific test |
| `assigned_to` | uuid | — | Filter by assignee. Use `unassigned` for unassigned alerts. |
| `sla_breached` | boolean | — | Filter by SLA breach status |
| `search` | string | — | Search in title and description |
| `date_from` | datetime | — | Alerts created after this time |
| `date_to` | datetime | — | Alerts created before this time |
| `sort` | string | `created_at` | Sort: `alert_number`, `severity`, `status`, `sla_deadline`, `created_at`, `updated_at` |
| `order` | string | `desc` | Sort order |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "alert_number": 7,
      "title": "Security Group — No Open Inbound 0.0.0.0/0 failed on CTRL-NW-001",
      "description": "Found 2 security groups with unrestricted inbound access on port 22.",
      "severity": "critical",
      "status": "open",
      "control": {
        "id": "uuid",
        "identifier": "CTRL-NW-001",
        "title": "Network Access Controls"
      },
      "test": {
        "id": "uuid",
        "identifier": "TST-NW-001",
        "title": "Security Group — No Open Inbound 0.0.0.0/0"
      },
      "assigned_to": null,
      "sla_deadline": "2026-02-20T20:00:00Z",
      "sla_breached": false,
      "hours_remaining": 3.5,
      "created_at": "2026-02-20T16:01:17Z",
      "updated_at": "2026-02-20T16:01:17Z"
    }
  ],
  "meta": {
    "total": 7,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `hours_remaining` is computed: positive = time left before SLA breach; negative = hours overdue. NULL if no SLA.
- Default sort is `created_at DESC` (newest first). For the alert queue dashboard, use `sort=severity&order=asc` to see critical alerts first.

---

#### `GET /api/v1/alerts/:id`

Get a single alert with full details.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "alert_number": 7,
    "title": "Security Group — No Open Inbound 0.0.0.0/0 failed on CTRL-NW-001",
    "description": "Found 2 security groups with unrestricted inbound access on port 22.",
    "severity": "critical",
    "status": "open",
    "control": {
      "id": "uuid",
      "identifier": "CTRL-NW-001",
      "title": "Network Access Controls",
      "category": "technical"
    },
    "test": {
      "id": "uuid",
      "identifier": "TST-NW-001",
      "title": "Security Group — No Open Inbound 0.0.0.0/0",
      "test_type": "network"
    },
    "test_result": {
      "id": "uuid",
      "status": "fail",
      "message": "Found 2 security groups with unrestricted inbound access on port 22.",
      "details": {
        "violations": [
          {"resource": "sg-abc123", "name": "dev-ssh-access", "port": 22, "source": "0.0.0.0/0"},
          {"resource": "sg-def456", "name": "legacy-admin", "port": 22, "source": "0.0.0.0/0"}
        ]
      },
      "tested_at": "2026-02-20T16:01:17Z"
    },
    "alert_rule": {
      "id": "uuid",
      "name": "Critical Test Failures"
    },
    "assigned_to": null,
    "assigned_at": null,
    "assigned_by": null,
    "sla_deadline": "2026-02-20T20:00:00Z",
    "sla_breached": false,
    "hours_remaining": 3.5,
    "resolved_by": null,
    "resolved_at": null,
    "resolution_notes": null,
    "suppressed_until": null,
    "suppression_reason": null,
    "delivery_channels": ["slack", "email", "in_app"],
    "delivered_at": {
      "slack": "2026-02-20T16:01:18Z",
      "email": "2026-02-20T16:01:22Z",
      "in_app": "2026-02-20T16:01:17Z"
    },
    "tags": [],
    "metadata": {},
    "created_at": "2026-02-20T16:01:17Z",
    "updated_at": "2026-02-20T16:01:17Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Alert not found in this org

---

#### `PUT /api/v1/alerts/:id/status`

Change an alert's lifecycle status.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`, `it_admin`, `devops_engineer`

**Request:**

```json
{
  "status": "acknowledged"
}
```

**Validation:**
- `status`: required, valid enum value
- Allowed transitions:
  - `open` → `acknowledged`, `in_progress`, `suppressed`, `closed`
  - `acknowledged` → `in_progress`, `suppressed`, `closed`
  - `in_progress` → `resolved`, `suppressed`, `closed`
  - `resolved` → `closed`, `open` (reopen if fix didn't hold)
  - `suppressed` → `open` (auto-reopen on expiration), `closed`
  - `closed` → `open` (reopen if issue recurs)

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "alert_number": 7,
    "status": "acknowledged",
    "previous_status": "open",
    "message": "Alert acknowledged."
  }
}
```

**Errors:**
- `422 UNPROCESSABLE` — Invalid status transition

**Audit log:** `alert.status_changed` with `{"alert_number": 7, "old_status": "open", "new_status": "acknowledged"}`

---

#### `PUT /api/v1/alerts/:id/assign`

Assign an alert to a user.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Request:**

```json
{
  "assigned_to": "uuid"
}
```

**Validation:**
- `assigned_to`: required, must be a valid user ID in the same org
- Alert must not be in `closed` or `resolved` status

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "alert_number": 7,
    "assigned_to": {
      "id": "uuid",
      "name": "Charlie DevOps",
      "email": "devops@acme.example.com"
    },
    "assigned_at": "2026-02-20T16:05:00Z",
    "assigned_by": {
      "id": "uuid",
      "name": "Alice Compliance"
    },
    "message": "Alert assigned to Charlie DevOps."
  }
}
```

**Side effects:**
- If alert is `open`, automatically transitions to `acknowledged`.
- Sends notification to the assignee via configured delivery channels.

**Errors:**
- `404 NOT_FOUND` — Alert or user not found
- `422 UNPROCESSABLE` — Alert is closed/resolved, can't assign

**Audit log:** `alert.assigned` with `{"alert_number": 7, "assigned_to": "uuid", "assigned_by": "uuid"}`

---

#### `PUT /api/v1/alerts/:id/resolve`

Mark an alert as resolved with resolution notes.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`, `it_admin`, `devops_engineer`

**Request:**

```json
{
  "resolution_notes": "Removed 0.0.0.0/0 ingress rules from sg-abc123 and sg-def456. Replaced with VPN CIDR 10.0.0.0/8 for SSH access."
}
```

**Validation:**
- `resolution_notes`: required, max 10000 chars
- Alert must be in `open`, `acknowledged`, or `in_progress` status

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "alert_number": 7,
    "status": "resolved",
    "previous_status": "in_progress",
    "resolved_by": {
      "id": "uuid",
      "name": "Charlie DevOps"
    },
    "resolved_at": "2026-02-20T17:30:00Z",
    "resolution_notes": "Removed 0.0.0.0/0 ingress rules...",
    "message": "Alert resolved. Will be verified on next test run."
  }
}
```

**Side effects:**
- Sets `resolved_by` to current user and `resolved_at` to now.
- Status transitions to `resolved`.
- The next passing test result for the same control can auto-close the alert (worker logic).

**Errors:**
- `422 UNPROCESSABLE` — Alert not in a resolvable status

**Audit log:** `alert.resolved` with `{"alert_number": 7, "resolution_notes": "..."}`

---

#### `PUT /api/v1/alerts/:id/suppress`

Suppress (snooze) an alert with mandatory justification and expiration (spec §3.3.2).

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "suppressed_until": "2026-02-27T16:00:00Z",
  "suppression_reason": "Known issue during scheduled infrastructure migration. Expected to be resolved by Feb 25. Tracking in JIRA-1234."
}
```

**Validation:**
- `suppressed_until`: required, ISO 8601 datetime, must be in the future, max 90 days from now
- `suppression_reason`: required, max 5000 chars, min 20 chars (enforce meaningful justification)
- Alert must not be in `closed` status

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "alert_number": 7,
    "status": "suppressed",
    "previous_status": "open",
    "suppressed_until": "2026-02-27T16:00:00Z",
    "suppression_reason": "Known issue during scheduled infrastructure migration...",
    "message": "Alert suppressed until Feb 27, 2026 4:00 PM UTC."
  }
}
```

**Side effects:**
- Status transitions to `suppressed`.
- When `suppressed_until` passes, the worker auto-transitions the alert back to `open`.
- Suppressed alerts are excluded from SLA tracking.

**Errors:**
- `422 UNPROCESSABLE` — suppressed_until in the past, or too far in the future (>90 days)

**Audit log:** `alert.suppressed` with `{"alert_number": 7, "until": "2026-02-27T16:00:00Z", "reason": "..."}`

---

#### `PUT /api/v1/alerts/:id/close`

Close an alert (verified fixed or accepted as risk).

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "resolution_notes": "Verified fixed in test run #43. All security groups now comply."
}
```

**Validation:**
- `resolution_notes`: optional (recommended), max 10000 chars
- Alert can be closed from any status

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "alert_number": 7,
    "status": "closed",
    "previous_status": "resolved",
    "message": "Alert closed."
  }
}
```

**Audit log:** `alert.closed` with `{"alert_number": 7}`

---

### 5. Alert Rules — CRUD

---

#### `GET /api/v1/alert-rules`

List the org's alert rules.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 50) |
| `enabled` | boolean | — | Filter by enabled/disabled |
| `sort` | string | `priority` | Sort: `name`, `priority`, `alert_severity`, `created_at` |
| `order` | string | `asc` | Sort order |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Critical Test Failures",
      "description": "Alert immediately on any critical test failure...",
      "enabled": true,
      "match_test_types": null,
      "match_severities": ["critical"],
      "match_result_statuses": ["fail"],
      "match_control_ids": null,
      "match_tags": null,
      "consecutive_failures": 1,
      "cooldown_minutes": 60,
      "alert_severity": "critical",
      "alert_title_template": null,
      "auto_assign_to": null,
      "sla_hours": 4,
      "delivery_channels": ["slack", "email", "in_app"],
      "priority": 10,
      "alerts_generated": 3,
      "created_at": "2026-02-20T10:00:00Z",
      "updated_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 4,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `alerts_generated` is a computed count of alerts that reference this rule.
- Sensitive fields (`slack_webhook_url`, `email_recipients`, `webhook_url`) are included for authorized roles.
- Default sort by priority (most important rules first).

---

#### `POST /api/v1/alert-rules`

Create a new alert rule.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "name": "Critical Test Failures",
  "description": "Alert immediately on any critical test failure. Delivered via Slack and email.",
  "enabled": true,
  "match_test_types": null,
  "match_severities": ["critical"],
  "match_result_statuses": ["fail"],
  "match_control_ids": null,
  "match_tags": null,
  "consecutive_failures": 1,
  "cooldown_minutes": 60,
  "alert_severity": "critical",
  "alert_title_template": "{{test.title}} failed on {{control.identifier}}",
  "auto_assign_to": null,
  "sla_hours": 4,
  "delivery_channels": ["slack", "email", "in_app"],
  "slack_webhook_url": "https://hooks.slack.com/services/T00000/B00000/XXXXXXXXX",
  "email_recipients": ["security@acme.example.com", "ciso@acme.example.com"],
  "priority": 10
}
```

**Validation:**
- `name`: required, max 255 chars, unique per org
- `description`: optional, max 5000 chars
- `enabled`: optional, default `true`
- `match_test_types`: optional, array of valid test_type values (NULL = match all)
- `match_severities`: optional, array of valid test_severity values (NULL = match all)
- `match_result_statuses`: optional, array of valid test_result_status values (default: `['fail']`)
- `match_control_ids`: optional, array of UUIDs (each must exist in org). Max 100.
- `match_tags`: optional, array of strings, max 20 tags
- `consecutive_failures`: optional, default 1, range 1–100
- `cooldown_minutes`: optional, default 0, range 0–10080 (1 week)
- `alert_severity`: required, valid alert_severity value
- `alert_title_template`: optional, max 500 chars. Supports `{{test.title}}`, `{{test.identifier}}`, `{{control.title}}`, `{{control.identifier}}`, `{{severity}}`
- `auto_assign_to`: optional, valid user UUID in the org
- `sla_hours`: optional, positive integer (1–8760, i.e., max 1 year). NULL = no SLA.
- `delivery_channels`: required, array of valid alert_delivery_channel values, must include at least one
- `slack_webhook_url`: required if `delivery_channels` includes `slack`. Valid HTTPS URL.
- `email_recipients`: required if `delivery_channels` includes `email`. Array of valid email addresses, max 20.
- `webhook_url`: required if `delivery_channels` includes `webhook`. Valid HTTPS URL.
- `webhook_headers`: optional, JSONB key-value pairs. Max 10 headers.
- `priority`: optional, default 100, range 0–1000

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "name": "Critical Test Failures",
    "enabled": true,
    "alert_severity": "critical",
    "priority": 10,
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `400 BAD_REQUEST` — Validation failure
- `403 FORBIDDEN` — Not authorized
- `409 CONFLICT` — Rule name already exists in this org
- `422 UNPROCESSABLE` — Invalid control IDs or user ID

**Audit log:** `alert_rule.created` with `{"name": "Critical Test Failures", "severity": "critical"}`

---

#### `GET /api/v1/alert-rules/:id`

Get a single alert rule with full details.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Response 200:**

Full rule object (same fields as list response + `slack_webhook_url`, `email_recipients`, `webhook_url`, `webhook_headers`, `created_by`).

**Errors:**
- `404 NOT_FOUND` — Rule not found

---

#### `PUT /api/v1/alert-rules/:id`

Update an alert rule.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

Same shape as POST, all fields optional.

**Validation:**
- Same rules as POST.

**Response 200:**

Returns updated rule.

**Errors:**
- `403 FORBIDDEN` — Not authorized
- `404 NOT_FOUND` — Rule not found
- `409 CONFLICT` — Name conflict

**Audit log:** `alert_rule.updated` with changed fields

---

#### `DELETE /api/v1/alert-rules/:id`

Delete an alert rule. Existing alerts generated by this rule are preserved.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "message": "Alert rule deleted. Existing alerts preserved."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Rule not found
- `403 FORBIDDEN` — Not authorized

**Audit log:** `alert_rule.deleted` with `{"name": "...", "rule_id": "uuid"}`

**Notes:**
- Hard delete. Alerts that reference this rule will have `alert_rule_id = NULL` (SET NULL FK).
- Consider offering a disable toggle before delete.

---

### 6. Alert Delivery

---

#### `POST /api/v1/alerts/:id/deliver`

Manually re-deliver an alert's notifications. Useful when initial delivery failed or new channels were configured.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Request:**

```json
{
  "channels": ["slack", "email"]
}
```

**Validation:**
- `channels`: optional, array of valid delivery channels. If omitted, re-delivers to all configured channels.

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "alert_number": 7,
    "delivery_results": {
      "slack": {"success": true, "delivered_at": "2026-02-20T17:00:00Z"},
      "email": {"success": true, "delivered_at": "2026-02-20T17:00:02Z"}
    },
    "message": "Alert re-delivered to 2 channels."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Alert not found
- `422 UNPROCESSABLE` — No delivery channels configured for this alert

**Notes:**
- Delivery is synchronous for simplicity (Slack webhook + SMTP are fast). For high-volume, this would be queued.
- `delivered_at` is updated in the alert's `delivered_at` JSONB.

---

#### `POST /api/v1/alerts/test-delivery`

Test alert delivery channels without creating a real alert. Sends a test notification.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "channel": "slack",
  "slack_webhook_url": "https://hooks.slack.com/services/T00000/B00000/XXXXXXXXX"
}
```

or

```json
{
  "channel": "email",
  "email_recipients": ["test@acme.example.com"]
}
```

or

```json
{
  "channel": "webhook",
  "webhook_url": "https://api.example.com/hooks/alerts",
  "webhook_headers": {"Authorization": "Bearer ..."}
}
```

**Response 200:**

```json
{
  "data": {
    "channel": "slack",
    "success": true,
    "message": "Test notification delivered successfully."
  }
}
```

**Errors:**
- `400 BAD_REQUEST` — Invalid channel configuration
- `422 UNPROCESSABLE` — Delivery failed (timeout, invalid URL, etc.)

---

### 7. Monitoring Dashboard

---

#### `GET /api/v1/monitoring/heatmap`

Control health heatmap data — for each active control, its current health status based on the latest test result.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `framework_id` | uuid | — | Filter controls for a specific framework (via control mappings) |
| `category` | string | — | Filter by control category: `technical`, `administrative`, `physical`, `operational` |

**Response 200:**

```json
{
  "data": {
    "summary": {
      "total_controls": 280,
      "healthy": 210,
      "failing": 8,
      "error": 2,
      "warning": 5,
      "untested": 55
    },
    "controls": [
      {
        "id": "uuid",
        "identifier": "CTRL-NW-001",
        "title": "Network Access Controls",
        "category": "technical",
        "health_status": "failing",
        "latest_result": {
          "status": "fail",
          "severity": "critical",
          "message": "Found 2 security groups with unrestricted inbound access.",
          "tested_at": "2026-02-20T16:01:17Z"
        },
        "active_alerts": 1,
        "tests_count": 1
      },
      {
        "id": "uuid",
        "identifier": "CTRL-AC-001",
        "title": "Multi-Factor Authentication",
        "category": "technical",
        "health_status": "healthy",
        "latest_result": {
          "status": "pass",
          "severity": "critical",
          "message": "MFA is enforced for all 142 users.",
          "tested_at": "2026-02-20T16:00:02Z"
        },
        "active_alerts": 0,
        "tests_count": 1
      },
      {
        "id": "uuid",
        "identifier": "CTRL-POL-001",
        "title": "Information Security Policy",
        "category": "administrative",
        "health_status": "untested",
        "latest_result": null,
        "active_alerts": 0,
        "tests_count": 0
      }
    ]
  }
}
```

**Notes:**
- Controls are sorted by health status (failing first, then error, warning, untested, healthy).
- `health_status` values: `healthy` (last result passed), `failing` (last result failed), `error` (last result errored), `warning` (last result warning), `untested` (no test results ever).
- `active_alerts` counts open/acknowledged/in_progress alerts for each control.
- This is the primary data source for the control health heatmap visualization.

---

#### `GET /api/v1/monitoring/posture`

Compliance posture score per activated framework — percentage of mapped controls passing.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "overall_score": 82.5,
    "frameworks": [
      {
        "framework_id": "uuid",
        "framework_name": "SOC 2",
        "framework_version": "2017",
        "org_framework_id": "uuid",
        "total_mapped_controls": 120,
        "passing": 105,
        "failing": 5,
        "untested": 10,
        "posture_score": 87.5,
        "trend": {
          "7d_ago": 85.0,
          "30d_ago": 78.0,
          "direction": "improving"
        }
      },
      {
        "framework_id": "uuid",
        "framework_name": "PCI DSS",
        "framework_version": "4.0.1",
        "org_framework_id": "uuid",
        "total_mapped_controls": 180,
        "passing": 140,
        "failing": 8,
        "untested": 32,
        "posture_score": 77.8,
        "trend": {
          "7d_ago": 75.0,
          "30d_ago": 70.0,
          "direction": "improving"
        }
      }
    ]
  }
}
```

**Notes:**
- `overall_score` is the weighted average across all frameworks (weighted by number of mapped controls).
- `posture_score` = `passing / total_mapped_controls * 100`.
- `trend` compares current score to 7 days ago and 30 days ago. `direction` is `improving`, `declining`, or `stable` (±1% threshold).
- Trend data requires historical posture snapshots. The worker can compute and cache this daily.

---

#### `GET /api/v1/monitoring/summary`

Top-level monitoring dashboard summary — high-level stats for the monitoring home page.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "overall_posture_score": 82.5,
    "controls": {
      "total_active": 280,
      "healthy": 210,
      "failing": 8,
      "untested": 55,
      "health_rate": 75.0
    },
    "tests": {
      "total_active": 45,
      "last_run": {
        "run_number": 42,
        "status": "completed",
        "completed_at": "2026-02-20T16:02:34Z",
        "passed": 43,
        "failed": 1,
        "errors": 1
      },
      "pass_rate_24h": 96.5
    },
    "alerts": {
      "open": 3,
      "acknowledged": 1,
      "in_progress": 2,
      "sla_breached": 0,
      "resolved_today": 5,
      "by_severity": {
        "critical": 1,
        "high": 2,
        "medium": 3,
        "low": 0
      }
    },
    "recent_activity": [
      {
        "type": "alert_created",
        "alert_number": 7,
        "title": "Security Group — No Open Inbound 0.0.0.0/0 failed",
        "severity": "critical",
        "timestamp": "2026-02-20T16:01:17Z"
      },
      {
        "type": "test_run_completed",
        "run_number": 42,
        "passed": 43,
        "failed": 1,
        "timestamp": "2026-02-20T16:02:34Z"
      },
      {
        "type": "alert_resolved",
        "alert_number": 5,
        "title": "Endpoint encryption non-compliant",
        "resolved_by": "Charlie DevOps",
        "timestamp": "2026-02-20T15:30:00Z"
      }
    ]
  }
}
```

**Notes:**
- `health_rate` = percentage of active controls that are `healthy` (controls with at least one test, latest result is pass).
- `pass_rate_24h` = percentage of test results in the last 24 hours that passed.
- `recent_activity` shows the last 10 monitoring events (alerts created/resolved, test runs completed, etc.).
- This is the single-query data source for the monitoring dashboard home page.

---

#### `GET /api/v1/monitoring/alert-queue`

Dedicated alert queue view — active alerts sorted by urgency for the operations team.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `queue` | string | `active` | Queue view: `active` (open+acknowledged+in_progress), `resolved`, `suppressed`, `all` |

**Response 200:**

```json
{
  "data": {
    "queue_summary": {
      "active": 6,
      "resolved": 12,
      "suppressed": 1,
      "closed": 45,
      "sla_breached": 0
    },
    "alerts": [
      {
        "id": "uuid",
        "alert_number": 7,
        "title": "Security Group — No Open Inbound 0.0.0.0/0 failed on CTRL-NW-001",
        "severity": "critical",
        "status": "open",
        "control_identifier": "CTRL-NW-001",
        "test_identifier": "TST-NW-001",
        "assigned_to_name": null,
        "sla_deadline": "2026-02-20T20:00:00Z",
        "sla_breached": false,
        "hours_remaining": 3.5,
        "created_at": "2026-02-20T16:01:17Z"
      }
    ]
  },
  "meta": {
    "total": 6,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Default `active` queue sorts by: severity (critical first), then SLA urgency (most urgent first), then creation time.
- `queue_summary` provides counts for all queues regardless of current filter.
- This is the primary data source for the alert queue dashboard page.

---

## Summary of All Endpoints

| # | Method | Path | Roles | Description |
|---|--------|------|-------|-------------|
| **Tests** | | | | |
| 1 | GET | `/api/v1/tests` | All | List test definitions |
| 2 | POST | `/api/v1/tests` | CISO, CM, SE, DE | Create test definition |
| 3 | GET | `/api/v1/tests/:id` | All | Get test detail |
| 4 | PUT | `/api/v1/tests/:id` | CISO, CM, SE, DE | Update test definition |
| 5 | PUT | `/api/v1/tests/:id/status` | CISO, CM, SE | Change test status |
| 6 | DELETE | `/api/v1/tests/:id` | CISO, CM | Soft-delete (deprecate) test |
| **Test Runs** | | | | |
| 7 | POST | `/api/v1/test-runs` | CISO, CM, SE, DE | Trigger manual test run |
| 8 | GET | `/api/v1/test-runs` | All | List test runs |
| 9 | GET | `/api/v1/test-runs/:id` | All | Get test run detail |
| 10 | POST | `/api/v1/test-runs/:id/cancel` | CISO, CM, SE | Cancel a test run |
| 11 | GET | `/api/v1/test-runs/:id/results` | All | List results in a run |
| 12 | GET | `/api/v1/test-runs/:rid/results/:id` | All | Get single result detail |
| **Test Results (Cross-Resource)** | | | | |
| 13 | GET | `/api/v1/tests/:id/results` | All | Test execution history |
| 14 | GET | `/api/v1/controls/:id/test-results` | All | Control test history |
| **Alerts** | | | | |
| 15 | GET | `/api/v1/alerts` | All | List alerts |
| 16 | GET | `/api/v1/alerts/:id` | All | Get alert detail |
| 17 | PUT | `/api/v1/alerts/:id/status` | CISO, CM, SE, IT, DE | Change alert status |
| 18 | PUT | `/api/v1/alerts/:id/assign` | CISO, CM, SE | Assign alert |
| 19 | PUT | `/api/v1/alerts/:id/resolve` | CISO, CM, SE, IT, DE | Resolve alert |
| 20 | PUT | `/api/v1/alerts/:id/suppress` | CISO, CM | Suppress (snooze) alert |
| 21 | PUT | `/api/v1/alerts/:id/close` | CISO, CM | Close alert |
| 22 | POST | `/api/v1/alerts/:id/deliver` | CISO, CM, SE | Re-deliver alert |
| 23 | POST | `/api/v1/alerts/test-delivery` | CISO, CM | Test delivery channels |
| **Alert Rules** | | | | |
| 24 | GET | `/api/v1/alert-rules` | CISO, CM, SE | List alert rules |
| 25 | POST | `/api/v1/alert-rules` | CISO, CM | Create alert rule |
| 26 | GET | `/api/v1/alert-rules/:id` | CISO, CM, SE | Get alert rule detail |
| 27 | PUT | `/api/v1/alert-rules/:id` | CISO, CM | Update alert rule |
| 28 | DELETE | `/api/v1/alert-rules/:id` | CISO, CM | Delete alert rule |
| **Monitoring Dashboard** | | | | |
| 29 | GET | `/api/v1/monitoring/heatmap` | All | Control health heatmap |
| 30 | GET | `/api/v1/monitoring/posture` | All | Compliance posture scores |
| 31 | GET | `/api/v1/monitoring/summary` | All | Dashboard summary stats |
| 32 | GET | `/api/v1/monitoring/alert-queue` | All | Alert queue view |

**Role abbreviations:** CISO = `ciso`, CM = `compliance_manager`, SE = `security_engineer`, IT = `it_admin`, DE = `devops_engineer`

**Total: 32 new endpoints**

---

## Go Implementation Notes

### New Files

```
api/internal/
├── handlers/
│   ├── tests.go                 # Endpoints 1-6 (test CRUD)
│   ├── test_runs.go             # Endpoints 7-12 (test execution)
│   ├── test_results.go          # Endpoints 13-14 (cross-resource queries)
│   ├── alerts.go                # Endpoints 15-21 (alert CRUD + lifecycle)
│   ├── alert_delivery.go        # Endpoints 22-23 (alert delivery)
│   ├── alert_rules.go           # Endpoints 24-28 (alert rule CRUD)
│   └── monitoring.go            # Endpoints 29-32 (dashboard data)
├── models/
│   ├── test.go                  # Test struct + queries
│   ├── test_run.go              # TestRun struct + queries
│   ├── test_result.go           # TestResult struct + queries
│   ├── alert.go                 # Alert struct + queries
│   └── alert_rule.go            # AlertRule struct + queries
├── services/
│   ├── test_runner.go           # Test execution engine
│   ├── alert_engine.go          # Alert generation + rule evaluation
│   └── alert_delivery.go        # Slack/email/webhook delivery
├── workers/
│   ├── monitoring_worker.go     # Background job: poll tests, execute, evaluate alerts
│   └── sla_worker.go            # Background job: check SLA breaches, unsuppress expired alerts
└── ...
```

### Route Registration

```go
// Tests
tests := v1.Group("/tests")
tests.Use(middleware.Auth(), middleware.Org())
{
    tests.GET("", handlers.ListTests)
    tests.POST("", middleware.RBAC("ciso", "compliance_manager", "security_engineer", "devops_engineer"), handlers.CreateTest)
    tests.GET("/:id", handlers.GetTest)
    tests.PUT("/:id", middleware.RBAC("ciso", "compliance_manager", "security_engineer", "devops_engineer"), handlers.UpdateTest)
    tests.PUT("/:id/status", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.ChangeTestStatus)
    tests.DELETE("/:id", middleware.RBAC("ciso", "compliance_manager"), handlers.DeleteTest)
    tests.GET("/:id/results", handlers.ListTestResultsByTest)
}

// Test Runs
runs := v1.Group("/test-runs")
runs.Use(middleware.Auth(), middleware.Org())
{
    runs.POST("", middleware.RBAC("ciso", "compliance_manager", "security_engineer", "devops_engineer"), handlers.CreateTestRun)
    runs.GET("", handlers.ListTestRuns)
    runs.GET("/:id", handlers.GetTestRun)
    runs.POST("/:id/cancel", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.CancelTestRun)
    runs.GET("/:id/results", handlers.ListTestRunResults)
    runs.GET("/:id/results/:rid", handlers.GetTestRunResult)
}

// Alerts
alerts := v1.Group("/alerts")
alerts.Use(middleware.Auth(), middleware.Org())
{
    alerts.GET("", handlers.ListAlerts)
    alerts.GET("/:id", handlers.GetAlert)
    alerts.PUT("/:id/status", middleware.RBAC("ciso", "compliance_manager", "security_engineer", "it_admin", "devops_engineer"), handlers.ChangeAlertStatus)
    alerts.PUT("/:id/assign", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.AssignAlert)
    alerts.PUT("/:id/resolve", middleware.RBAC("ciso", "compliance_manager", "security_engineer", "it_admin", "devops_engineer"), handlers.ResolveAlert)
    alerts.PUT("/:id/suppress", middleware.RBAC("ciso", "compliance_manager"), handlers.SuppressAlert)
    alerts.PUT("/:id/close", middleware.RBAC("ciso", "compliance_manager"), handlers.CloseAlert)
    alerts.POST("/:id/deliver", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.RedeliverAlert)
    alerts.POST("/test-delivery", middleware.RBAC("ciso", "compliance_manager"), handlers.TestAlertDelivery)
}

// Alert Rules
rules := v1.Group("/alert-rules")
rules.Use(middleware.Auth(), middleware.Org())
{
    rules.GET("", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.ListAlertRules)
    rules.POST("", middleware.RBAC("ciso", "compliance_manager"), handlers.CreateAlertRule)
    rules.GET("/:id", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.GetAlertRule)
    rules.PUT("/:id", middleware.RBAC("ciso", "compliance_manager"), handlers.UpdateAlertRule)
    rules.DELETE("/:id", middleware.RBAC("ciso", "compliance_manager"), handlers.DeleteAlertRule)
}

// Monitoring Dashboard
monitoring := v1.Group("/monitoring")
monitoring.Use(middleware.Auth(), middleware.Org())
{
    monitoring.GET("/heatmap", handlers.GetControlHealthHeatmap)
    monitoring.GET("/posture", handlers.GetCompliancePosture)
    monitoring.GET("/summary", handlers.GetMonitoringSummary)
    monitoring.GET("/alert-queue", handlers.GetAlertQueue)
}

// Add to existing controls routes
ctrl.GET("/:id/test-results", handlers.ListControlTestResults)  // NEW
```

### Worker Architecture

```go
// api/internal/workers/monitoring_worker.go

type MonitoringWorker struct {
    db          *sql.DB
    testRunner  *services.TestRunner
    alertEngine *services.AlertEngine
    interval    time.Duration  // poll interval (e.g., 30 seconds)
}

func (w *MonitoringWorker) Run(ctx context.Context) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.processDueTests(ctx)
            w.checkSLABreaches(ctx)
            w.unsuppressExpiredAlerts(ctx)
        }
    }
}

func (w *MonitoringWorker) processDueTests(ctx context.Context) {
    // 1. Find tests WHERE status='active' AND next_run_at <= NOW()
    // 2. Group by org_id, create test_run per org
    // 3. Execute each test, write test_results
    // 4. Update test_run summary counters
    // 5. Evaluate alert rules for failed tests
    // 6. Update next_run_at for executed tests
}
```

### Docker Compose Addition

Add worker service to `docker-compose.yml`:

```yaml
worker:
  build:
    context: ./api
    dockerfile: Dockerfile
  command: ["./api", "--mode=worker"]  # or separate binary
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
  environment:
    <<: *api-env
    WORKER_MODE: "true"
    WORKER_POLL_INTERVAL: "30s"
  networks:
    - rp-net
  restart: unless-stopped
```

### Alert Delivery Implementations

```go
// Slack webhook
func (s *AlertDeliveryService) SendSlack(webhookURL string, alert *models.Alert) error {
    payload := map[string]interface{}{
        "text": fmt.Sprintf("🚨 *%s Alert* — %s", alert.Severity, alert.Title),
        "blocks": []map[string]interface{}{
            // Rich Slack block kit message
        },
    }
    // POST to webhookURL with JSON payload
}

// Email via SMTP
func (s *AlertDeliveryService) SendEmail(recipients []string, alert *models.Alert) error {
    // Use Go's net/smtp or a library like gomail
    // Subject: [Raisin Protect] Critical Alert: {title}
    // Body: HTML template with alert details, SLA deadline, and action links
}

// Custom webhook
func (s *AlertDeliveryService) SendWebhook(url string, headers map[string]string, alert *models.Alert) error {
    // POST alert as JSON to the webhook URL with custom headers
    // Include signature header for webhook verification
}
```
