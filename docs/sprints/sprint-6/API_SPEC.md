# Sprint 6 — API Specification: Risk Register

## Overview

Sprint 6 adds risk management endpoints: full CRUD for risks, point-in-time assessments with configurable scoring formulas, treatment plan lifecycle, risk-to-control linkage with effectiveness tracking, heat map data for visualization, gap detection, and risk search/filtering.

This implements spec §4.1 (Internal Risk Management) — the risk backbone of the GRC platform. Risks are scored on a 5×5 likelihood × impact grid (range 1–25), with inherent scores (before controls) and residual scores (after controls/treatments).

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `risks` | Per-org | Risk definitions with scoring and lifecycle |
| `risks/:id/assessments` | Per-org | Point-in-time likelihood × impact assessments |
| `risks/:id/treatments` | Per-org | Treatment plans (mitigate/accept/transfer/avoid) |
| `risks/:id/controls` | Per-org | Risk-to-control mappings with effectiveness |
| `risks/heat-map` | Per-org | Aggregated heat map data (5×5 grid) |
| `risks/gaps` | Per-org | Gap detection (risks without treatments/controls) |
| `risks/stats` | Per-org | Dashboard statistics |

---

## Scoring Model

### Likelihood Levels → Numeric Scores

| Enum Value | Score | Probability Range |
|-----------|-------|------------------|
| `rare` | 1 | <5% |
| `unlikely` | 2 | 5–25% |
| `possible` | 3 | 25–50% |
| `likely` | 4 | 50–75% |
| `almost_certain` | 5 | >75% |

### Impact Levels → Numeric Scores

| Enum Value | Score | Severity |
|-----------|-------|----------|
| `negligible` | 1 | Insignificant |
| `minor` | 2 | Brief disruption |
| `moderate` | 3 | Moderate disruption |
| `major` | 4 | Significant disruption |
| `severe` | 5 | Catastrophic |

### Scoring Formulas

| Formula | Calculation | Range | Description |
|---------|-------------|-------|-------------|
| `likelihood_x_impact` | L × I | 1–25 | Default. Simple multiplication. |

### Severity Bands (derived from overall_score)

| Score Range | Severity | Color |
|------------|----------|-------|
| 20–25 | `critical` | Red |
| 12–19 | `high` | Orange |
| 6–11 | `medium` | Yellow |
| 1–5 | `low` | Green |

---

## Endpoints

---

### 1. Risk CRUD

---

#### `GET /api/v1/risks`

List the org's risks with filtering, search, and pagination.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter: `identified`, `open`, `assessing`, `treating`, `monitoring`, `accepted`, `closed`, `archived` |
| `category` | string | — | Filter by risk_category enum value |
| `owner_id` | uuid | — | Filter by risk owner |
| `is_template` | bool | `false` | Include templates (default: only active risks) |
| `severity` | string | — | Filter by severity band: `critical`, `high`, `medium`, `low` |
| `score_min` | number | — | Minimum residual score (inclusive) |
| `score_max` | number | — | Maximum residual score (inclusive) |
| `score_type` | string | `residual` | Which score to filter on: `inherent`, `residual` |
| `has_treatments` | bool | — | Filter: `true` = has treatments, `false` = no treatments |
| `has_controls` | bool | — | Filter: `true` = has linked controls, `false` = no controls |
| `overdue_assessment` | bool | — | Filter: `true` = assessment overdue |
| `tags` | string | — | Comma-separated tags (AND logic) |
| `search` | string | — | Full-text search in title and description |
| `sort` | string | `residual_score` | Sort: `identifier`, `title`, `category`, `status`, `inherent_score`, `residual_score`, `next_assessment_at`, `created_at`, `updated_at` |
| `order` | string | `desc` | Sort order: `asc`, `desc` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "RISK-CY-001",
      "title": "Ransomware Attack on Production Systems",
      "description": "Risk of ransomware encrypting production databases...",
      "category": "cyber_security",
      "status": "treating",
      "owner": {
        "id": "uuid",
        "name": "Bob Security",
        "email": "security@acme.example.com"
      },
      "secondary_owner": null,
      "inherent_score": {
        "likelihood": "likely",
        "impact": "severe",
        "score": 20.00,
        "severity": "critical"
      },
      "residual_score": {
        "likelihood": "possible",
        "impact": "major",
        "score": 12.00,
        "severity": "high"
      },
      "risk_appetite_threshold": 10.00,
      "appetite_breached": true,
      "assessment_frequency_days": 90,
      "next_assessment_at": "2026-05-18",
      "last_assessed_at": "2026-02-18",
      "assessment_status": "on_track",
      "source": "threat_assessment",
      "affected_assets": ["payment-api", "customer-db", "erp-system"],
      "linked_controls_count": 2,
      "active_treatments_count": 3,
      "tags": ["critical", "ransomware", "q1-review"],
      "created_at": "2026-02-18T10:00:00Z",
      "updated_at": "2026-02-20T14:00:00Z"
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
- `appetite_breached` is computed: `true` when `residual_score > risk_appetite_threshold`.
- `assessment_status` is computed: `overdue` (past due), `due_soon` (within 30 days), `on_track` (beyond 30 days), `no_schedule` (no frequency configured).
- Default sort is `residual_score DESC` — highest risks first.
- Templates are excluded by default. Use `is_template=true` to browse the risk library.

**Error Codes:**
- `400` — Invalid filter/sort parameters
- `401` — Missing/invalid token
- `403` — Insufficient role permissions

---

#### `GET /api/v1/risks/:id`

Get a single risk with full details.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "RISK-CY-001",
    "title": "Ransomware Attack on Production Systems",
    "description": "Risk of ransomware encrypting production databases and application servers, with potential double extortion.",
    "category": "cyber_security",
    "status": "treating",
    "owner": {
      "id": "uuid",
      "name": "Bob Security",
      "email": "security@acme.example.com"
    },
    "secondary_owner": null,
    "inherent_score": {
      "likelihood": "likely",
      "likelihood_score": 4,
      "impact": "severe",
      "impact_score": 5,
      "score": 20.00,
      "severity": "critical"
    },
    "residual_score": {
      "likelihood": "possible",
      "likelihood_score": 3,
      "impact": "major",
      "impact_score": 4,
      "score": 12.00,
      "severity": "high"
    },
    "risk_appetite_threshold": 10.00,
    "appetite_breached": true,
    "acceptance": null,
    "assessment_frequency_days": 90,
    "next_assessment_at": "2026-05-18",
    "last_assessed_at": "2026-02-18",
    "assessment_status": "on_track",
    "source": "threat_assessment",
    "affected_assets": ["payment-api", "customer-db", "erp-system"],
    "is_template": false,
    "template_source": null,
    "linked_controls": [
      {
        "id": "uuid",
        "identifier": "CTRL-EP-001",
        "title": "Endpoint Detection & Response",
        "effectiveness": "effective",
        "mitigation_percentage": 35
      },
      {
        "id": "uuid",
        "identifier": "CTRL-NW-001",
        "title": "Network Segmentation",
        "effectiveness": "partially_effective",
        "mitigation_percentage": 20
      }
    ],
    "treatment_summary": {
      "total": 3,
      "planned": 0,
      "in_progress": 1,
      "implemented": 0,
      "verified": 2,
      "cancelled": 0
    },
    "latest_assessments": {
      "inherent": {
        "id": "uuid",
        "assessment_date": "2026-02-18",
        "assessor": "Bob Security",
        "justification": "Ransomware attacks are increasing...",
        "valid_until": "2026-05-18"
      },
      "residual": {
        "id": "uuid",
        "assessment_date": "2026-02-18",
        "assessor": "Bob Security",
        "justification": "EDR, network segmentation, and immutable backups...",
        "valid_until": "2026-05-18"
      }
    },
    "tags": ["critical", "ransomware", "q1-review"],
    "metadata": {},
    "created_at": "2026-02-18T10:00:00Z",
    "updated_at": "2026-02-20T14:00:00Z"
  }
}
```

**Notes:**
- `linked_controls` shows basic info with effectiveness. Full details via `GET /risks/:id/controls`.
- `treatment_summary` aggregates treatment statuses. Full list via `GET /risks/:id/treatments`.
- `latest_assessments` shows the most recent inherent and residual assessments.
- `acceptance` is populated when `status = 'accepted'`:
  ```json
  {
    "accepted_at": "2026-02-15T10:00:00Z",
    "accepted_by": { "id": "uuid", "name": "David CISO" },
    "expiry": "2027-02-15",
    "justification": "ERP migration scheduled for Q3 2026..."
  }
  ```

**Error Codes:**
- `401` — Missing/invalid token
- `404` — Risk not found or belongs to different org

---

#### `POST /api/v1/risks`

Create a new risk.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`

**Request Body:**

```json
{
  "identifier": "RISK-CY-010",
  "title": "API Authentication Bypass",
  "description": "Risk of attackers bypassing API authentication mechanisms to access protected resources.",
  "category": "cyber_security",
  "owner_id": "uuid",
  "secondary_owner_id": "uuid",
  "risk_appetite_threshold": 8.00,
  "assessment_frequency_days": 90,
  "source": "penetration_test",
  "affected_assets": ["payment-api", "admin-api"],
  "tags": ["api", "authentication", "critical"],
  "initial_assessment": {
    "inherent_likelihood": "likely",
    "inherent_impact": "major",
    "residual_likelihood": "unlikely",
    "residual_impact": "moderate",
    "justification": "Recent pentest identified weak token validation in legacy endpoints."
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `identifier` | string | Yes | Risk identifier (org-unique, e.g., `RISK-CY-010`) |
| `title` | string | Yes | Risk title (max 500 chars) |
| `description` | string | No | Detailed risk description |
| `category` | string | Yes | risk_category enum value |
| `owner_id` | uuid | No | Primary risk owner (defaults to creator) |
| `secondary_owner_id` | uuid | No | Secondary owner |
| `risk_appetite_threshold` | number | No | Acceptable score threshold (1–25) |
| `assessment_frequency_days` | int | No | Re-assessment cadence in days |
| `source` | string | No | How this risk was discovered |
| `affected_assets` | string[] | No | Impacted systems/assets |
| `tags` | string[] | No | Free-form tags |
| `initial_assessment` | object | No | If provided, creates initial assessments |
| `initial_assessment.inherent_likelihood` | string | Cond. | Required if initial_assessment is provided |
| `initial_assessment.inherent_impact` | string | Cond. | Required if initial_assessment is provided |
| `initial_assessment.residual_likelihood` | string | No | Optional initial residual assessment |
| `initial_assessment.residual_impact` | string | No | Optional initial residual assessment |
| `initial_assessment.justification` | string | No | Assessment justification |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "RISK-CY-010",
    "title": "API Authentication Bypass",
    "status": "identified",
    "inherent_score": {
      "likelihood": "likely",
      "impact": "major",
      "score": 16.00,
      "severity": "high"
    },
    "residual_score": {
      "likelihood": "unlikely",
      "impact": "moderate",
      "score": 6.00,
      "severity": "medium"
    },
    "created_at": "2026-02-20T23:00:00Z"
  }
}
```

**Behavior:**
1. Creates the risk with `status = 'identified'`
2. If `initial_assessment` is provided:
   - Creates inherent `risk_assessment` with computed scores
   - Optionally creates residual `risk_assessment`
   - Denormalizes scores onto `risks` table
   - Sets `last_assessed_at = TODAY`, computes `next_assessment_at`
3. Logs `risk.created` to audit log
4. **Notification trigger**: If `inherent_score >= 20` (critical) or `residual_score >= 12` (high), queue notification to CISO and risk owner (per spec: "Risk notifications when high/critical risks created")

**Error Codes:**
- `400` — Validation error (missing fields, invalid category, invalid score range)
- `401` — Missing/invalid token
- `403` — Role not authorized to create risks
- `409` — Identifier already exists for this org

---

#### `PUT /api/v1/risks/:id`

Update risk metadata (not scores — use assessment endpoints for scoring changes).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner

**Request Body:**

```json
{
  "title": "API Authentication Bypass (Legacy Endpoints)",
  "description": "Updated description...",
  "category": "cyber_security",
  "owner_id": "uuid",
  "secondary_owner_id": "uuid",
  "risk_appetite_threshold": 6.00,
  "assessment_frequency_days": 60,
  "source": "penetration_test",
  "affected_assets": ["payment-api", "admin-api", "user-api"],
  "tags": ["api", "authentication", "critical", "pentest-finding"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | No | Updated title |
| `description` | string | No | Updated description |
| `category` | string | No | Updated category |
| `owner_id` | uuid | No | New primary owner |
| `secondary_owner_id` | uuid | No | New secondary owner (null to clear) |
| `risk_appetite_threshold` | number | No | Updated threshold (1–25) |
| `assessment_frequency_days` | int | No | Updated cadence |
| `next_assessment_at` | date | No | Manually set next assessment date |
| `source` | string | No | Updated source |
| `affected_assets` | string[] | No | Replace affected assets list |
| `tags` | string[] | No | Replace tags |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "RISK-CY-010",
    "title": "API Authentication Bypass (Legacy Endpoints)",
    "status": "identified",
    "updated_at": "2026-02-20T23:10:00Z"
  }
}
```

**Notes:**
- Cannot change `identifier` or `is_template` after creation.
- Status transitions are handled by dedicated status change logic (see below).
- If `owner_id` changes, logs `risk.owner_changed` to audit log.

**Error Codes:**
- `400` — Invalid field values
- `401` — Missing/invalid token
- `403` — Not authorized (not owner, compliance_manager, or ciso)
- `404` — Risk not found

---

#### `POST /api/v1/risks/:id/archive`

Archive a risk (soft-delete — moves to archived status).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "RISK-CY-010",
    "status": "archived",
    "updated_at": "2026-02-20T23:15:00Z"
  }
}
```

**Behavior:**
1. Sets `status = 'archived'`
2. Cancels any `planned` or `in_progress` treatments (sets to `cancelled`)
3. Logs `risk.archived` to audit log

**Error Codes:**
- `400` — Risk is already archived
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk not found

---

#### `PUT /api/v1/risks/:id/status`

Transition risk status.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner

**Request Body:**

```json
{
  "status": "accepted",
  "justification": "ERP migration scheduled for Q3 2026. Risk is within appetite given compensating controls.",
  "acceptance_expiry": "2027-02-15"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | Yes | Target status |
| `justification` | string | Cond. | Required when status = `accepted` |
| `acceptance_expiry` | date | Cond. | Required when status = `accepted` |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "RISK-CY-010",
    "status": "accepted",
    "acceptance": {
      "accepted_at": "2026-02-20T23:20:00Z",
      "accepted_by": { "id": "uuid", "name": "David CISO" },
      "expiry": "2027-02-15",
      "justification": "ERP migration scheduled for Q3 2026..."
    },
    "updated_at": "2026-02-20T23:20:00Z"
  }
}
```

**Allowed transitions:**

| From | To | Conditions |
|------|----|-----------|
| `identified` | `open`, `assessing` | — |
| `open` | `assessing`, `treating`, `accepted`, `closed` | `accepted` requires justification + expiry |
| `assessing` | `open`, `treating`, `accepted` | `accepted` requires justification + expiry |
| `treating` | `monitoring`, `accepted`, `closed` | `accepted` requires justification + expiry |
| `monitoring` | `treating`, `accepted`, `closed` | — |
| `accepted` | `open`, `assessing`, `treating` | Re-opens the risk |
| `closed` | `open` | Re-opens if threat re-emerges |
| Any | `archived` | Use `POST /risks/:id/archive` instead |

**Error Codes:**
- `400` — Invalid transition, missing required fields for accepted status
- `401` — Missing/invalid token
- `403` — Not authorized. `accepted` status requires `ciso` or `compliance_manager` role.
- `404` — Risk not found

---

### 2. Risk Assessments

---

#### `GET /api/v1/risks/:id/assessments`

List all assessments for a risk (trend data).

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `assessment_type` | string | — | Filter: `inherent`, `residual`, `target` |
| `current_only` | bool | `false` | If true, only return current assessments |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "risk_id": "uuid",
      "assessment_type": "residual",
      "likelihood": "possible",
      "impact": "major",
      "likelihood_score": 3,
      "impact_score": 4,
      "overall_score": 12.00,
      "scoring_formula": "likelihood_x_impact",
      "severity": "high",
      "justification": "EDR, network segmentation, and immutable backups reduce likelihood...",
      "assumptions": null,
      "data_sources": ["edr-deployment-report", "backup-verification-test"],
      "assessed_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "assessment_date": "2026-02-18",
      "valid_until": "2026-05-18",
      "is_current": true,
      "superseded_by": null,
      "created_at": "2026-02-18T10:00:00Z"
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
- Ordered by `assessment_date DESC` (newest first).
- Contains full history for trend analysis charts.

---

#### `POST /api/v1/risks/:id/assessments`

Create a new assessment (scores a risk at a point in time).

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner

**Request Body:**

```json
{
  "assessment_type": "residual",
  "likelihood": "unlikely",
  "impact": "moderate",
  "scoring_formula": "likelihood_x_impact",
  "justification": "After EDR deployment and network segmentation, residual risk has decreased.",
  "assumptions": "Assumes 100% EDR coverage is maintained and backups are tested monthly.",
  "data_sources": ["edr-coverage-report", "monthly-backup-test-feb-2026"],
  "valid_until": "2026-05-18"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `assessment_type` | string | Yes | `inherent`, `residual`, or `target` |
| `likelihood` | string | Yes | likelihood_level enum value |
| `impact` | string | Yes | impact_level enum value |
| `scoring_formula` | string | No | Formula identifier (default: `likelihood_x_impact`) |
| `justification` | string | No | Why these levels were chosen |
| `assumptions` | string | No | Key assumptions |
| `data_sources` | string[] | No | Evidence/references used |
| `valid_until` | date | No | When this assessment expires |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "risk_id": "uuid",
    "assessment_type": "residual",
    "likelihood": "unlikely",
    "impact": "moderate",
    "likelihood_score": 2,
    "impact_score": 3,
    "overall_score": 6.00,
    "severity": "medium",
    "scoring_formula": "likelihood_x_impact",
    "is_current": true,
    "assessed_by": {
      "id": "uuid",
      "name": "Bob Security"
    },
    "assessment_date": "2026-02-20",
    "risk_updated": {
      "residual_likelihood": "unlikely",
      "residual_impact": "moderate",
      "residual_score": 6.00,
      "appetite_breached": false
    },
    "created_at": "2026-02-20T23:00:00Z"
  }
}
```

**Behavior:**
1. Computes `likelihood_score` and `impact_score` from enum values
2. Computes `overall_score` using the specified formula
3. Computes `severity` from `overall_score`
4. Sets previous current assessment of same type to `is_current = FALSE`, links via `superseded_by`
5. Denormalizes scores onto the `risks` table:
   - If `assessment_type = 'inherent'` → updates `risks.inherent_likelihood/impact/score`
   - If `assessment_type = 'residual'` → updates `risks.residual_likelihood/impact/score`
6. Updates `risks.last_assessed_at = TODAY`
7. Computes `risks.next_assessment_at = TODAY + assessment_frequency_days`
8. Logs `risk_assessment.created` to audit log
9. **Notification trigger**: If new severity is `critical` or `high`, notify CISO and risk owner

**Error Codes:**
- `400` — Invalid likelihood/impact values, invalid formula
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk not found

---

#### `POST /api/v1/risks/:id/recalculate`

Recalculate denormalized scores from the latest current assessments. Useful after bulk changes or data corrections.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "RISK-CY-001",
    "inherent_score": {
      "likelihood": "likely",
      "impact": "severe",
      "score": 20.00,
      "severity": "critical"
    },
    "residual_score": {
      "likelihood": "possible",
      "impact": "major",
      "score": 12.00,
      "severity": "high"
    },
    "appetite_breached": true,
    "recalculated_at": "2026-02-20T23:30:00Z"
  }
}
```

**Behavior:**
1. Reads the latest `is_current = TRUE` assessments for `inherent` and `residual` types
2. Updates denormalized fields on `risks` table
3. Logs `risk.score_recalculated` to audit log

**Error Codes:**
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk not found

---

### 3. Risk Treatments

---

#### `GET /api/v1/risks/:id/treatments`

List treatment plans for a risk.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | — | Filter: `planned`, `in_progress`, `implemented`, `verified`, `ineffective`, `cancelled` |
| `treatment_type` | string | — | Filter: `mitigate`, `accept`, `transfer`, `avoid` |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "risk_id": "uuid",
      "treatment_type": "mitigate",
      "title": "Deploy EDR to All Endpoints",
      "description": "Deploy CrowdStrike Falcon EDR to 100% of endpoints...",
      "status": "verified",
      "owner": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "priority": "critical",
      "due_date": "2026-03-15",
      "started_at": "2026-02-01T09:00:00Z",
      "completed_at": "2026-02-15T16:00:00Z",
      "estimated_effort_hours": 40.00,
      "actual_effort_hours": 52.00,
      "effectiveness_rating": "highly_effective",
      "effectiveness_notes": "CrowdStrike deployed to 100% of endpoints...",
      "expected_residual": {
        "likelihood": "unlikely",
        "impact": "major",
        "score": 8.00
      },
      "target_control": {
        "id": "uuid",
        "identifier": "CTRL-EP-001",
        "title": "Endpoint Detection & Response"
      },
      "created_by": {
        "id": "uuid",
        "name": "David CISO"
      },
      "created_at": "2026-02-01T09:00:00Z",
      "updated_at": "2026-02-20T10:00:00Z"
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

---

#### `POST /api/v1/risks/:id/treatments`

Create a treatment plan.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner

**Request Body:**

```json
{
  "treatment_type": "mitigate",
  "title": "Implement Web Application Firewall",
  "description": "Deploy WAF in front of all public-facing APIs to detect and block common attack patterns.",
  "owner_id": "uuid",
  "priority": "high",
  "due_date": "2026-04-15",
  "estimated_effort_hours": 80,
  "expected_residual_likelihood": "unlikely",
  "expected_residual_impact": "minor",
  "target_control_id": "uuid",
  "notes": "Evaluate Cloudflare WAF vs AWS WAF. Budget approved."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `treatment_type` | string | Yes | `mitigate`, `accept`, `transfer`, `avoid` |
| `title` | string | Yes | Treatment title (max 500 chars) |
| `description` | string | No | Detailed plan |
| `owner_id` | uuid | No | Treatment owner (defaults to risk owner) |
| `priority` | string | No | `critical`, `high`, `medium` (default), `low` |
| `due_date` | date | No | Target completion date |
| `estimated_effort_hours` | number | No | Estimated hours |
| `expected_residual_likelihood` | string | No | Expected likelihood after treatment |
| `expected_residual_impact` | string | No | Expected impact after treatment |
| `target_control_id` | uuid | No | Control that implements this treatment |
| `notes` | string | No | Additional context |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "risk_id": "uuid",
    "treatment_type": "mitigate",
    "title": "Implement Web Application Firewall",
    "status": "planned",
    "priority": "high",
    "due_date": "2026-04-15",
    "expected_residual": {
      "likelihood": "unlikely",
      "impact": "minor",
      "score": 4.00
    },
    "created_at": "2026-02-20T23:00:00Z"
  }
}
```

**Behavior:**
1. Creates treatment with `status = 'planned'`
2. Computes `expected_residual_score` from expected likelihood × impact
3. If risk status is `identified` or `open`, transitions to `treating`
4. Logs `risk_treatment.created` to audit log

**Error Codes:**
- `400` — Invalid treatment_type, priority, or expected levels
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk not found
- `422` — Risk is archived

---

#### `PUT /api/v1/risks/:id/treatments/:treatment_id`

Update a treatment plan.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner, treatment owner

**Request Body:**

```json
{
  "title": "Implement Web Application Firewall (Cloudflare)",
  "description": "Updated: using Cloudflare WAF based on evaluation.",
  "status": "in_progress",
  "priority": "high",
  "due_date": "2026-04-30",
  "started_at": "2026-03-01T09:00:00Z",
  "estimated_effort_hours": 60,
  "actual_effort_hours": 12,
  "notes": "Cloudflare selected. DNS migration underway."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | No | Updated title |
| `description` | string | No | Updated description |
| `status` | string | No | Status transition (see allowed transitions) |
| `owner_id` | uuid | No | Reassign owner |
| `priority` | string | No | Updated priority |
| `due_date` | date | No | Updated due date |
| `started_at` | datetime | No | When work began |
| `estimated_effort_hours` | number | No | Updated estimate |
| `actual_effort_hours` | number | No | Hours spent so far |
| `notes` | string | No | Updated notes |

**Allowed status transitions:**

| From | To |
|------|----|
| `planned` | `in_progress`, `cancelled` |
| `in_progress` | `implemented`, `cancelled` |
| `implemented` | `verified`, `ineffective` |
| `verified` | `ineffective` (if re-evaluated) |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "treatment_type": "mitigate",
    "title": "Implement Web Application Firewall (Cloudflare)",
    "status": "in_progress",
    "started_at": "2026-03-01T09:00:00Z",
    "updated_at": "2026-03-01T09:00:00Z"
  }
}
```

**Error Codes:**
- `400` — Invalid status transition, invalid field values
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk or treatment not found

---

#### `POST /api/v1/risks/:id/treatments/:treatment_id/complete`

Mark a treatment as implemented and optionally record effectiveness.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner, treatment owner

**Request Body:**

```json
{
  "actual_effort_hours": 65,
  "effectiveness_rating": "effective",
  "effectiveness_notes": "WAF blocked 200+ attack attempts in first week. False positive rate below 0.5%."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `actual_effort_hours` | number | No | Total hours spent |
| `effectiveness_rating` | string | No | `highly_effective`, `effective`, `partially_effective`, `ineffective` |
| `effectiveness_notes` | string | No | Effectiveness explanation |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "verified",
    "completed_at": "2026-02-20T23:00:00Z",
    "effectiveness_rating": "effective",
    "effectiveness_notes": "WAF blocked 200+ attack attempts...",
    "effectiveness_reviewed_at": "2026-02-20T23:00:00Z"
  }
}
```

**Behavior:**
1. If `effectiveness_rating` is provided: sets `status = 'verified'` (skips `implemented`)
2. If no `effectiveness_rating`: sets `status = 'implemented'`
3. Sets `completed_at = NOW()`
4. Sets `effectiveness_reviewed_at = NOW()` and `effectiveness_reviewed_by = current_user`
5. Logs `risk_treatment.completed` to audit log
6. Checks if all treatments for this risk are now complete — if yes, transitions risk status to `monitoring`

**Error Codes:**
- `400` — Treatment is already completed/cancelled, invalid effectiveness_rating
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk or treatment not found

---

### 4. Risk-to-Control Linkage

---

#### `GET /api/v1/risks/:id/controls`

List controls linked to a risk with effectiveness data.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "risk_control_id": "uuid",
      "identifier": "CTRL-EP-001",
      "title": "Endpoint Detection & Response",
      "description": "Deploy EDR to all endpoints...",
      "category": "technical",
      "status": "active",
      "effectiveness": "effective",
      "mitigation_percentage": 35,
      "notes": "EDR detects and blocks ransomware payloads at endpoint level",
      "last_effectiveness_review": "2026-02-18",
      "reviewed_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "linked_by": {
        "id": "uuid",
        "name": "Bob Security"
      },
      "linked_at": "2026-02-18T10:00:00Z",
      "latest_test_status": "pass",
      "frameworks": ["SOC 2", "ISO 27001"]
    }
  ],
  "meta": {
    "total": 2,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `latest_test_status` comes from Sprint 4's test_results for this control.
- `frameworks` shows which frameworks this control maps to.
- `risk_control_id` is the junction table record ID (for unlinking/updating).

---

#### `POST /api/v1/risks/:id/controls`

Link a control to a risk.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner

**Request Body:**

```json
{
  "control_id": "uuid",
  "effectiveness": "partially_effective",
  "mitigation_percentage": 25,
  "notes": "Email filtering reduces phishing delivery but doesn't prevent all attempts"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `control_id` | uuid | Yes | Control to link |
| `effectiveness` | string | No | `effective`, `partially_effective`, `ineffective`, `not_assessed` (default) |
| `mitigation_percentage` | int | No | Estimated % of risk mitigated (0–100) |
| `notes` | string | No | Why this control mitigates this risk |

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "risk_id": "uuid",
    "control": {
      "id": "uuid",
      "identifier": "CTRL-EF-001",
      "title": "Email Filtering"
    },
    "effectiveness": "partially_effective",
    "mitigation_percentage": 25,
    "notes": "Email filtering reduces phishing delivery...",
    "linked_by": {
      "id": "uuid",
      "name": "Bob Security"
    },
    "created_at": "2026-02-20T23:00:00Z"
  }
}
```

**Behavior:**
1. Validates control exists and belongs to the same org
2. Creates `risk_controls` junction record
3. Logs `risk_control.linked` to audit log

**Error Codes:**
- `400` — Invalid control_id, effectiveness, or mitigation_percentage
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk or control not found
- `409` — Control is already linked to this risk

---

#### `PUT /api/v1/risks/:id/controls/:control_id`

Update effectiveness assessment for a risk-control linkage.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner

**Request Body:**

```json
{
  "effectiveness": "effective",
  "mitigation_percentage": 40,
  "notes": "After tuning, email filtering now catches 99.5% of phishing attempts"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `effectiveness` | string | No | Updated effectiveness |
| `mitigation_percentage` | int | No | Updated mitigation estimate |
| `notes` | string | No | Updated justification |

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "risk_id": "uuid",
    "control_id": "uuid",
    "effectiveness": "effective",
    "mitigation_percentage": 40,
    "last_effectiveness_review": "2026-02-20",
    "reviewed_by": {
      "id": "uuid",
      "name": "Bob Security"
    },
    "updated_at": "2026-02-20T23:10:00Z"
  }
}
```

**Behavior:**
1. Updates `risk_controls` record
2. Sets `last_effectiveness_review = TODAY`, `reviewed_by = current_user`
3. Logs `risk_control.effectiveness_updated` to audit log

**Error Codes:**
- `400` — Invalid effectiveness or mitigation_percentage
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk-control link not found

---

#### `DELETE /api/v1/risks/:id/controls/:control_id`

Unlink a control from a risk.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, risk owner

**Response 204:** No content

**Behavior:**
1. Deletes the `risk_controls` junction record
2. Logs `risk_control.unlinked` to audit log

**Error Codes:**
- `401` — Missing/invalid token
- `403` — Not authorized
- `404` — Risk-control link not found

---

### 5. Risk Heat Map

---

#### `GET /api/v1/risks/heat-map`

Get aggregated risk data for the 5×5 heat map visualization.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `score_type` | string | `residual` | Which scores to map: `inherent`, `residual` |
| `category` | string | — | Filter by risk category |
| `status` | string | — | Filter by status (comma-separated for multiple) |

**Response 200:**

```json
{
  "data": {
    "score_type": "residual",
    "grid": [
      {
        "likelihood": "almost_certain",
        "likelihood_score": 5,
        "impact": "severe",
        "impact_score": 5,
        "score": 25,
        "severity": "critical",
        "count": 0,
        "risks": []
      },
      {
        "likelihood": "likely",
        "likelihood_score": 4,
        "impact": "severe",
        "impact_score": 5,
        "score": 20,
        "severity": "critical",
        "count": 0,
        "risks": []
      },
      {
        "likelihood": "possible",
        "likelihood_score": 3,
        "impact": "major",
        "impact_score": 4,
        "score": 12,
        "severity": "high",
        "count": 1,
        "risks": [
          {
            "id": "uuid",
            "identifier": "RISK-CY-001",
            "title": "Ransomware Attack on Production Systems",
            "status": "treating"
          }
        ]
      }
    ],
    "summary": {
      "total_risks": 5,
      "by_severity": {
        "critical": 0,
        "high": 2,
        "medium": 2,
        "low": 1
      },
      "average_score": 9.60,
      "appetite_breaches": 2
    }
  }
}
```

**Notes:**
- `grid` contains all 25 cells of the 5×5 matrix (likelihood × impact). Empty cells have `count: 0` and `risks: []`.
- `risks` within each cell are sorted by score descending.
- `summary` provides aggregate statistics across all active risks.
- `appetite_breaches` counts risks where `residual_score > risk_appetite_threshold`.
- Only includes active risks (excludes `closed`, `archived`, and templates).

---

### 6. Risk Gap Detection

---

#### `GET /api/v1/risks/gaps`

Identify risks with missing treatments or controls.

- **Auth:** Bearer token required
- **Roles:** `compliance_manager`, `ciso`, `security_engineer`, `auditor`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `gap_type` | string | `all` | Filter: `no_treatments`, `no_controls`, `high_without_controls`, `overdue_assessment`, `expired_acceptance`, `all` |
| `min_severity` | string | — | Minimum severity: `low`, `medium`, `high`, `critical` |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page |

**Response 200:**

```json
{
  "data": {
    "summary": {
      "total_active_risks": 42,
      "risks_without_treatments": 8,
      "risks_without_controls": 5,
      "high_risks_without_controls": 2,
      "overdue_assessments": 3,
      "expired_acceptances": 1
    },
    "gaps": [
      {
        "risk": {
          "id": "uuid",
          "identifier": "RISK-OP-005",
          "title": "Change Management Failure",
          "category": "operational",
          "status": "open",
          "residual_score": 12.00,
          "severity": "high",
          "owner": {
            "id": "uuid",
            "name": "Eve DevOps"
          }
        },
        "gap_types": ["no_treatments", "no_controls"],
        "days_open": 45,
        "recommendation": "High-severity risk with no treatments or controls. Immediate action needed: create mitigation plan and link relevant controls."
      },
      {
        "risk": {
          "id": "uuid",
          "identifier": "RISK-TE-001",
          "title": "Legacy ERP System Dependency",
          "category": "technology",
          "status": "accepted",
          "residual_score": 12.00,
          "severity": "high",
          "owner": {
            "id": "uuid",
            "name": "David CISO"
          }
        },
        "gap_types": ["expired_acceptance"],
        "acceptance_expiry": "2027-02-15",
        "days_until_expiry": -5,
        "recommendation": "Risk acceptance has expired. Re-assess and either renew acceptance or create treatment plan."
      }
    ]
  },
  "meta": {
    "total_gaps": 12,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Gaps are prioritized by severity then age (highest/oldest first).
- `recommendation` is generated based on gap type and risk attributes.
- `days_open` = days since risk was created.
- `days_until_expiry` = negative means expired.

**Error Codes:**
- `401` — Missing/invalid token
- `403` — Not authorized

---

### 7. Risk Search

---

#### `GET /api/v1/risks/search`

Advanced search across risks.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `q` | string | — | **Required.** Search query |
| `category` | string | — | Filter by category |
| `status` | string | — | Filter by status |
| `severity` | string | — | Filter by severity band |
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "RISK-CY-001",
      "title": "Ransomware Attack on Production Systems",
      "description": "Risk of ransomware encrypting production databases...",
      "category": "cyber_security",
      "status": "treating",
      "residual_score": 12.00,
      "severity": "high",
      "match_context": "...ransomware encrypting critical business data and systems...",
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
- Uses PostgreSQL full-text search on `title || description`.
- Templates are excluded from search results by default.
- `match_context` shows a snippet around the matching text.

**Error Codes:**
- `400` — Missing `q` parameter
- `401` — Missing/invalid token

---

### 8. Risk Statistics

---

#### `GET /api/v1/risks/stats`

Get risk management statistics for the dashboard.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "total_risks": 42,
    "by_status": {
      "identified": 3,
      "open": 8,
      "assessing": 2,
      "treating": 12,
      "monitoring": 10,
      "accepted": 4,
      "closed": 2,
      "archived": 1
    },
    "by_category": {
      "cyber_security": 12,
      "operational": 8,
      "compliance": 6,
      "technology": 5,
      "data_privacy": 4,
      "third_party": 3,
      "financial": 2,
      "legal": 1,
      "reputational": 1
    },
    "by_severity": {
      "critical": 2,
      "high": 10,
      "medium": 18,
      "low": 12
    },
    "scoring_summary": {
      "average_inherent_score": 12.40,
      "average_residual_score": 7.80,
      "average_risk_reduction": 37.10,
      "highest_residual": {
        "id": "uuid",
        "identifier": "RISK-CY-001",
        "title": "Ransomware Attack on Production Systems",
        "score": 12.00,
        "severity": "high"
      }
    },
    "treatment_summary": {
      "total_treatments": 28,
      "planned": 5,
      "in_progress": 10,
      "implemented": 3,
      "verified": 8,
      "ineffective": 1,
      "cancelled": 1,
      "overdue": 2
    },
    "control_coverage": {
      "risks_with_controls": 35,
      "risks_without_controls": 7,
      "average_controls_per_risk": 3.2
    },
    "assessment_health": {
      "overdue_assessments": 3,
      "due_within_30_days": 5,
      "expired_acceptances": 1
    },
    "appetite_summary": {
      "within_appetite": 30,
      "breaching_appetite": 8,
      "no_threshold_set": 4
    },
    "templates_available": 225,
    "recent_activity": [
      {
        "risk_identifier": "RISK-CY-001",
        "action": "risk_treatment.completed",
        "actor": "Bob Security",
        "timestamp": "2026-02-20T23:00:00Z"
      }
    ]
  }
}
```

**Notes:**
- `average_risk_reduction` = average percentage reduction from inherent to residual scores: `(1 - avg_residual/avg_inherent) × 100`.
- `treatment_summary.overdue` = treatments with `due_date < TODAY` and status in (`planned`, `in_progress`).
- `recent_activity` shows the last 5 risk-related actions from the audit log.
- Templates are NOT counted in `total_risks`.

---

## Status Transition Diagram

```
                    ┌──────────────────────────────────────────────────────────┐
                    │                                                          │
                    ▼                                                          │
             ┌─────────────┐                    ┌──────────┐                   │
  ────▶      │ IDENTIFIED  │ ──────────────────▶│  OPEN    │ ◄────────────────│
 create      │             │                    │          │                   │
             └──────┬──────┘                    └────┬─────┘                   │
                    │                                │                         │
                    │          ┌──────────────────────┤                         │
                    ▼          ▼                      ▼                         │
             ┌──────────┐   assessing          ┌───────────┐                   │
             │ ASSESSING │ ◄───────────────────│ TREATING  │──▶ MONITORING ───┘
             └────┬─────┘                      └─────┬─────┘       │
                  │                                  │              │
                  ├──▶ treating                      │              │
                  ├──▶ accepted                      ▼              ▼
                  │                            ┌──────────┐   ┌──────────┐
                  │                            │ ACCEPTED │   │  CLOSED  │
                  │                            │ (time-   │   │          │
                  │                            │  bound)  │   └──────────┘
                  │                            └──────────┘
                  │
                  ▼
             ┌──────────┐
             │ ARCHIVED │ ◄──── (via POST /risks/:id/archive from any status)
             └──────────┘
```

**Key rules:**
- `accepted` requires CISO or compliance_manager role + justification + expiry date
- `accepted` → any active status: re-opens the risk (acceptance revoked)
- When all treatments are `verified`/`cancelled`: risk auto-transitions to `monitoring`
- `closed` can be re-opened to `open` if the threat re-emerges
- `archived` is a terminal state (use archive endpoint)

---

## Notification Rules

The following events trigger notifications (delivered via Slack webhook and/or email):

| Event | Recipients | Condition |
|-------|-----------|-----------|
| Risk created | CISO, risk owner | When inherent_score ≥ 20 (critical) |
| Risk created | Risk owner | When inherent_score ≥ 12 (high) |
| Risk score increased | CISO, risk owner | When new severity > previous severity |
| Appetite breached | CISO, risk owner | When residual_score > risk_appetite_threshold |
| Assessment overdue | Risk owner | When next_assessment_at < TODAY |
| Acceptance expiring | CISO, risk owner | When acceptance_expiry is within 30 days |
| Treatment overdue | Treatment owner, risk owner | When treatment due_date < TODAY and status in (planned, in_progress) |
| Treatment completed | Risk owner | When any treatment is marked complete |

**Implementation note:** Notifications are queued asynchronously. The API response is not delayed by notification delivery. Notification preferences are stored in the user's profile (future Sprint 9 integration engine will formalize delivery channels).

---

## Authentication & Authorization Summary

| Endpoint | Method | Roles Allowed |
|----------|--------|---------------|
| `GET /risks` | GET | All |
| `GET /risks/:id` | GET | All |
| `POST /risks` | POST | compliance_manager, ciso, security_engineer |
| `PUT /risks/:id` | PUT | compliance_manager, ciso, security_engineer, owner |
| `POST /risks/:id/archive` | POST | compliance_manager, ciso |
| `PUT /risks/:id/status` | PUT | compliance_manager, ciso, security_engineer, owner (accepted: ciso/compliance_manager only) |
| `GET /risks/:id/assessments` | GET | All |
| `POST /risks/:id/assessments` | POST | compliance_manager, ciso, security_engineer, owner |
| `POST /risks/:id/recalculate` | POST | compliance_manager, ciso, security_engineer |
| `GET /risks/:id/treatments` | GET | All |
| `POST /risks/:id/treatments` | POST | compliance_manager, ciso, security_engineer, owner |
| `PUT /risks/:id/treatments/:tid` | PUT | compliance_manager, ciso, security_engineer, owner, treatment_owner |
| `POST /risks/:id/treatments/:tid/complete` | POST | compliance_manager, ciso, security_engineer, owner, treatment_owner |
| `GET /risks/:id/controls` | GET | All |
| `POST /risks/:id/controls` | POST | compliance_manager, ciso, security_engineer, owner |
| `PUT /risks/:id/controls/:cid` | PUT | compliance_manager, ciso, security_engineer, owner |
| `DELETE /risks/:id/controls/:cid` | DELETE | compliance_manager, ciso, security_engineer, owner |
| `GET /risks/heat-map` | GET | All |
| `GET /risks/gaps` | GET | compliance_manager, ciso, security_engineer, auditor |
| `GET /risks/search` | GET | All |
| `GET /risks/stats` | GET | All |

---

## Audit Log Events

All risk-related actions are logged to `audit_log`:

| Action | Resource Type | When |
|--------|--------------|------|
| `risk.created` | risk | New risk created |
| `risk.updated` | risk | Risk metadata updated |
| `risk.status_changed` | risk | Status transition |
| `risk.archived` | risk | Risk archived |
| `risk.owner_changed` | risk | Owner reassigned |
| `risk.score_recalculated` | risk | Scores refreshed from assessments |
| `risk_assessment.created` | risk_assessment | New assessment recorded |
| `risk_treatment.created` | risk_treatment | Treatment plan created |
| `risk_treatment.updated` | risk_treatment | Treatment plan updated |
| `risk_treatment.status_changed` | risk_treatment | Status transition |
| `risk_treatment.completed` | risk_treatment | Treatment marked complete |
| `risk_treatment.cancelled` | risk_treatment | Treatment cancelled |
| `risk_control.linked` | risk_control | Control linked to risk |
| `risk_control.unlinked` | risk_control | Control unlinked from risk |
| `risk_control.effectiveness_updated` | risk_control | Effectiveness re-evaluated |

---

## Endpoint Count Summary

| Category | Endpoints |
|----------|-----------|
| Risk CRUD | 5 (list, get, create, update, archive) |
| Risk Status | 1 (status transition) |
| Risk Assessments | 3 (list, create, recalculate) |
| Risk Treatments | 4 (list, create, update, complete) |
| Risk-Control Linkage | 4 (list, link, update effectiveness, unlink) |
| Heat Map | 1 |
| Gap Detection | 1 |
| Search | 1 |
| Statistics | 1 |
| **Total** | **21** |
