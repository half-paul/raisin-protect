# Sprint 2 — API Specification: Frameworks & Controls

## Overview

Sprint 2 adds the core GRC API surface: a system framework catalog (read-only for tenants), per-org framework activation, a control library with CRUD and ownership, cross-framework control mappings, requirement scoping, and coverage analysis.

**Base URL:** `http://localhost:8090/api/v1`

All endpoints follow Sprint 1 response patterns (see Sprint 1 API_SPEC.md for common response shapes, error codes, and auth model).

---

## New Resource Patterns

| Resource | Scope | Description |
|----------|-------|-------------|
| `frameworks` | System catalog | Compliance frameworks (SOC 2, PCI DSS, etc.) |
| `framework_versions` | System catalog | Versions of each framework |
| `requirements` | System catalog | Individual requirements within a version |
| `org-frameworks` | Per-org | Which frameworks an org has activated |
| `controls` | Per-org | Org's control library |
| `control-mappings` | Per-org | Links between controls and requirements |
| `requirement-scopes` | Per-org | Org's in/out-of-scope decisions |

---

## Endpoints

---

### 1. Framework Catalog (System-Level)

These endpoints expose the shared framework catalog. All authenticated users can read; only system admins could create/modify (out of scope for Sprint 2 — frameworks are seeded).

---

#### `GET /api/v1/frameworks`

List all available compliance frameworks.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `category` | string | — | Filter by category: `security_privacy`, `payment`, `data_privacy`, `ai_governance`, `industry`, `custom` |
| `search` | string | — | Search by name or identifier (case-insensitive, partial match) |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "pci_dss",
      "name": "PCI DSS",
      "description": "Payment Card Industry Data Security Standard...",
      "category": "payment",
      "website_url": "https://www.pcisecuritystandards.org/",
      "logo_url": null,
      "versions_count": 1,
      "created_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 5,
    "request_id": "uuid"
  }
}
```

**Notes:**
- No pagination needed (framework count is small, <50).
- `versions_count` is a convenience join showing how many versions exist.
- Returns all frameworks regardless of org activation status.

---

#### `GET /api/v1/frameworks/:framework_id`

Get a single framework with all its versions.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "pci_dss",
    "name": "PCI DSS",
    "description": "Payment Card Industry Data Security Standard...",
    "category": "payment",
    "website_url": "https://www.pcisecuritystandards.org/",
    "logo_url": null,
    "versions": [
      {
        "id": "uuid",
        "version": "4.0.1",
        "display_name": "PCI DSS v4.0.1",
        "status": "active",
        "effective_date": "2024-06-11",
        "sunset_date": null,
        "total_requirements": 280,
        "created_at": "2026-02-20T10:00:00Z"
      }
    ],
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Framework ID doesn't exist

---

#### `GET /api/v1/frameworks/:framework_id/versions/:version_id`

Get a specific framework version with metadata.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "framework_id": "uuid",
    "framework_identifier": "pci_dss",
    "framework_name": "PCI DSS",
    "version": "4.0.1",
    "display_name": "PCI DSS v4.0.1",
    "status": "active",
    "effective_date": "2024-06-11",
    "sunset_date": null,
    "changelog": null,
    "total_requirements": 280,
    "created_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Framework or version doesn't exist

---

#### `GET /api/v1/frameworks/:framework_id/versions/:version_id/requirements`

List requirements for a framework version. Returns a flat list by default, or a nested tree with `?format=tree`.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | string | `flat` | `flat` for paginated list, `tree` for nested hierarchy |
| `assessable_only` | boolean | `false` | If true, only return assessable (leaf) requirements |
| `parent_id` | uuid | — | Filter to children of a specific requirement |
| `search` | string | — | Search by identifier or title |
| `page` | int | 1 | Page number (flat mode only) |
| `per_page` | int | 50 | Items per page, max 200 (flat mode only) |

**Response 200 (flat):**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "6.4.3",
      "title": "All payment page scripts that are loaded and executed...",
      "description": "...",
      "guidance": "...",
      "parent_id": "uuid",
      "depth": 2,
      "section_order": 3,
      "is_assessable": true,
      "created_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 280,
    "page": 1,
    "per_page": 50,
    "request_id": "uuid"
  }
}
```

**Response 200 (tree):**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "6",
      "title": "Develop and Maintain Secure Systems and Software",
      "depth": 0,
      "is_assessable": false,
      "children": [
        {
          "id": "uuid",
          "identifier": "6.4",
          "title": "Public-Facing Web Applications are Protected Against Attacks",
          "depth": 1,
          "is_assessable": false,
          "children": [
            {
              "id": "uuid",
              "identifier": "6.4.3",
              "title": "All payment page scripts...",
              "depth": 2,
              "is_assessable": true,
              "children": []
            }
          ]
        }
      ]
    }
  ],
  "meta": {
    "request_id": "uuid"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Framework or version doesn't exist

**Notes:**
- Tree format uses a recursive CTE and returns the full hierarchy in one response (no pagination).
- For large frameworks (PCI DSS 280 reqs), tree format is still manageable (<100KB response).
- `parent_id` filter is useful for lazy-loading tree branches in the UI.

---

### 2. Org Frameworks (Per-Org Activation)

Manage which frameworks an org is tracking for compliance.

---

#### `GET /api/v1/org-frameworks`

List the org's activated frameworks with compliance posture summary.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | — | Filter: `active`, `inactive` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "framework": {
        "id": "uuid",
        "identifier": "pci_dss",
        "name": "PCI DSS",
        "category": "payment"
      },
      "active_version": {
        "id": "uuid",
        "version": "4.0.1",
        "display_name": "PCI DSS v4.0.1",
        "total_requirements": 280
      },
      "status": "active",
      "target_date": "2026-12-31",
      "notes": null,
      "stats": {
        "total_requirements": 280,
        "in_scope": 265,
        "out_of_scope": 15,
        "mapped": 210,
        "unmapped": 55,
        "coverage_pct": 79.2
      },
      "activated_at": "2026-02-20T10:00:00Z",
      "created_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 5,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `stats` is computed on-the-fly from control_mappings and requirement_scopes.
- `coverage_pct` = (mapped / in_scope) * 100, showing how many in-scope assessable requirements have at least one control mapped.

---

#### `POST /api/v1/org-frameworks`

Activate a framework for the org.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "framework_id": "uuid",
  "version_id": "uuid",
  "target_date": "2026-12-31",
  "notes": "PCI DSS compliance required for payment processing",
  "seed_controls": true
}
```

**Validation:**
- `framework_id`: required, must exist in frameworks table
- `version_id`: required, must be an active version of the specified framework
- `target_date`: optional, ISO date
- `notes`: optional, max 2000 chars
- `seed_controls`: optional, default `true` — if true, creates pre-built controls from the library template and maps them to requirements

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "framework": {
      "id": "uuid",
      "identifier": "soc2",
      "name": "SOC 2"
    },
    "active_version": {
      "id": "uuid",
      "version": "2024",
      "display_name": "SOC 2 (2024 TSC)"
    },
    "status": "active",
    "target_date": "2026-12-31",
    "notes": "PCI DSS compliance required for payment processing",
    "controls_seeded": 45,
    "mappings_seeded": 89,
    "activated_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `409 CONFLICT` — Framework already activated for this org
- `404 NOT_FOUND` — Framework or version doesn't exist
- `422 UNPROCESSABLE` — Version is not in `active` status

**Audit log:** `framework.activated` with `{"framework": "soc2", "version": "2024", "seed_controls": true}`

**Notes:**
- When `seed_controls` is true, the API creates org-specific controls from the control library templates and auto-maps them to the framework's requirements. This gives orgs a head start.
- Seeding is idempotent — controls with the same `source_template_id` won't be duplicated.

---

#### `PUT /api/v1/org-frameworks/:id`

Update an org framework activation (change version, target date, notes).

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "version_id": "uuid",
  "target_date": "2027-03-31",
  "notes": "Extended timeline due to scope change",
  "status": "active"
}
```

**Validation:**
- `version_id`: optional, must be an active version of the same framework
- `target_date`: optional, ISO date (null to clear)
- `notes`: optional, max 2000 chars
- `status`: optional, `active` or `inactive`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "framework": {
      "id": "uuid",
      "identifier": "pci_dss",
      "name": "PCI DSS"
    },
    "active_version": {
      "id": "uuid",
      "version": "4.0.1",
      "display_name": "PCI DSS v4.0.1"
    },
    "status": "active",
    "target_date": "2027-03-31",
    "notes": "Extended timeline due to scope change",
    "updated_at": "2026-02-20T11:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Org framework activation doesn't exist
- `403 FORBIDDEN` — Not authorized
- `422 UNPROCESSABLE` — Version is not in `active` status or doesn't belong to this framework

**Audit log:** `framework.version_changed` (if version changed) or `framework.activated`/`framework.deactivated` (if status changed)

---

#### `DELETE /api/v1/org-frameworks/:id`

Deactivate a framework for the org. Sets status to `inactive` — does NOT delete controls or mappings.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "inactive",
    "message": "Framework deactivated. Controls and mappings preserved."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Org framework not found
- `403 FORBIDDEN` — Not authorized

**Audit log:** `framework.deactivated`

**Notes:**
- Soft deactivation — all controls, mappings, and scoping decisions are preserved.
- Reactivate by `PUT` with `"status": "active"`.

---

#### `GET /api/v1/org-frameworks/:id/coverage`

Coverage gap analysis for a specific activated framework. Shows which in-scope requirements have controls mapped and which are gaps.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | — | Filter: `covered` (has mapping), `gap` (no mapping) |
| `page` | int | 1 | Page number |
| `per_page` | int | 50 | Items per page (max 200) |

**Response 200:**

```json
{
  "data": {
    "framework": {
      "identifier": "pci_dss",
      "name": "PCI DSS",
      "version": "4.0.1"
    },
    "summary": {
      "total_requirements": 280,
      "assessable_requirements": 251,
      "in_scope": 240,
      "out_of_scope": 11,
      "covered": 195,
      "gaps": 45,
      "coverage_pct": 81.25
    },
    "requirements": [
      {
        "id": "uuid",
        "identifier": "6.4.3",
        "title": "All payment page scripts...",
        "depth": 2,
        "in_scope": true,
        "status": "covered",
        "controls": [
          {
            "id": "uuid",
            "identifier": "CTRL-SD-012",
            "title": "Payment Page Script Monitoring",
            "strength": "primary",
            "status": "active"
          }
        ]
      },
      {
        "id": "uuid",
        "identifier": "11.3.1",
        "title": "Internal vulnerability scans...",
        "depth": 2,
        "in_scope": true,
        "status": "gap",
        "controls": []
      }
    ]
  },
  "meta": {
    "total": 240,
    "page": 1,
    "per_page": 50,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Only includes assessable requirements (section headers excluded).
- Out-of-scope requirements are excluded from coverage calculation but can be included with a future `include_out_of_scope` param.
- `coverage_pct` = covered / in_scope * 100.

---

### 3. Requirement Scoping (Per-Org)

Manage which requirements are in-scope or out-of-scope for the org.

---

#### `GET /api/v1/org-frameworks/:id/scoping`

Get all scoping decisions for a framework's requirements.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `in_scope` | boolean | — | Filter: `true` for in-scope, `false` for out-of-scope |
| `page` | int | 1 | Page number |
| `per_page` | int | 100 | Items per page (max 200) |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "requirement": {
        "id": "uuid",
        "identifier": "9.1",
        "title": "Processes and mechanisms for restricting physical access..."
      },
      "in_scope": false,
      "justification": "Cloud-only environment — no physical cardholder data access points.",
      "scoped_by": {
        "id": "uuid",
        "name": "Alice Compliance"
      },
      "updated_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 15,
    "page": 1,
    "per_page": 100,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Only returns requirements with explicit scoping decisions. Requirements without entries are implicitly in-scope.

---

#### `PUT /api/v1/org-frameworks/:id/requirements/:requirement_id/scope`

Set or update the scoping decision for a specific requirement.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "in_scope": false,
  "justification": "Cloud-only environment — no physical cardholder data access points."
}
```

**Validation:**
- `in_scope`: required, boolean
- `justification`: required when `in_scope` is `false`, max 2000 chars. Optional when `in_scope` is `true`.

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "requirement_id": "uuid",
    "requirement_identifier": "9.1",
    "in_scope": false,
    "justification": "Cloud-only environment — no physical cardholder data access points.",
    "scoped_by": {
      "id": "uuid",
      "name": "Alice Compliance"
    },
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Org framework or requirement doesn't exist
- `403 FORBIDDEN` — Not authorized
- `422 UNPROCESSABLE` — `justification` required when marking out-of-scope; requirement doesn't belong to this framework version

**Audit log:** `requirement.scoped` with `{"requirement": "9.1", "in_scope": false}`

**Notes:**
- Upsert behavior — creates if no scoping decision exists, updates if one does.
- Scoping a parent (section header) does NOT automatically scope children. Each assessable requirement must be scoped individually for audit trail clarity.
- To reset to in-scope, send `{"in_scope": true}`.

---

#### `DELETE /api/v1/org-frameworks/:id/requirements/:requirement_id/scope`

Remove a scoping decision, returning the requirement to the default (in-scope).

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "message": "Scoping decision removed. Requirement is now implicitly in-scope."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — No scoping decision exists for this requirement

**Audit log:** `requirement.scoped` with `{"requirement": "9.1", "in_scope": true, "action": "reset"}`

---

### 4. Controls (Per-Org Library)

Full CRUD for the org's control library.

---

#### `GET /api/v1/controls`

List the org's controls with filtering, search, and pagination.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter: `draft`, `active`, `under_review`, `deprecated` |
| `category` | string | — | Filter: `technical`, `administrative`, `physical`, `operational` |
| `owner_id` | uuid | — | Filter by primary owner |
| `is_custom` | boolean | — | Filter: `true` for custom, `false` for library |
| `framework_id` | uuid | — | Filter controls mapped to a specific framework |
| `unmapped` | boolean | `false` | If true, only show controls with zero mappings |
| `search` | string | — | Full-text search on title + description |
| `sort` | string | `identifier` | Sort: `identifier`, `title`, `category`, `status`, `created_at`, `updated_at` |
| `order` | string | `asc` | Sort order: `asc`, `desc` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "identifier": "CTRL-AC-001",
      "title": "Multi-Factor Authentication",
      "description": "Enforce MFA for all user accounts...",
      "category": "technical",
      "status": "active",
      "is_custom": false,
      "owner": {
        "id": "uuid",
        "name": "Bob Security",
        "email": "security@acme.example.com"
      },
      "secondary_owner": null,
      "mappings_count": 4,
      "frameworks": ["SOC 2", "PCI DSS", "ISO 27001", "GDPR"],
      "created_at": "2026-02-20T10:00:00Z",
      "updated_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 312,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `mappings_count` is a denormalized count of control_mappings for this control.
- `frameworks` is an array of distinct framework names this control is mapped to (convenience for the control library browser).
- `search` uses PostgreSQL full-text search (GIN index) for good performance on 300+ controls.
- `framework_id` filter joins through control_mappings → requirements → framework_versions to find controls mapped to a specific framework.
- Results always scoped to caller's `org_id`.

---

#### `POST /api/v1/controls`

Create a new control in the org's library.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Request:**

```json
{
  "identifier": "CTRL-NW-015",
  "title": "Web Application Firewall",
  "description": "Deploy and maintain WAF in front of all public-facing web applications...",
  "implementation_guidance": "1. Deploy WAF (Cloudflare, AWS WAF, or equivalent)\n2. Configure OWASP Top 10 rule set\n3. Enable logging to SIEM\n4. Review blocked requests weekly",
  "category": "technical",
  "status": "draft",
  "owner_id": "uuid",
  "secondary_owner_id": "uuid",
  "evidence_requirements": "WAF configuration export, weekly review logs, SIEM integration proof",
  "test_criteria": "WAF is active and blocking OWASP Top 10 attacks. Logs are forwarded to SIEM within 5 minutes.",
  "metadata": {
    "risk_level": "high",
    "implementation_cost": "medium"
  }
}
```

**Validation:**
- `identifier`: required, max 50 chars, unique within org, alphanumeric + hyphens
- `title`: required, max 500 chars
- `description`: required, max 10000 chars
- `implementation_guidance`: optional, max 10000 chars
- `category`: required, must be valid `control_category`
- `status`: optional, default `draft`
- `owner_id`: optional, must be a user in the same org
- `secondary_owner_id`: optional, must be a user in the same org, different from `owner_id`
- `evidence_requirements`: optional, max 5000 chars
- `test_criteria`: optional, max 5000 chars
- `metadata`: optional, valid JSON object, max 10KB

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "CTRL-NW-015",
    "title": "Web Application Firewall",
    "description": "Deploy and maintain WAF in front of all public-facing web applications...",
    "implementation_guidance": "...",
    "category": "technical",
    "status": "draft",
    "is_custom": true,
    "owner": {
      "id": "uuid",
      "name": "Bob Security"
    },
    "secondary_owner": {
      "id": "uuid",
      "name": "Eve DevOps"
    },
    "evidence_requirements": "...",
    "test_criteria": "...",
    "metadata": {
      "risk_level": "high",
      "implementation_cost": "medium"
    },
    "mappings_count": 0,
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `409 CONFLICT` — Identifier already exists in this org
- `403 FORBIDDEN` — Role not authorized
- `404 NOT_FOUND` — `owner_id` or `secondary_owner_id` not found in org

**Audit log:** `control.created` with `{"identifier": "CTRL-NW-015", "category": "technical"}`

**Notes:**
- Custom controls are automatically flagged with `is_custom: true`.
- Controls start in `draft` status by default. Use the status change endpoint to move through the lifecycle.

---

#### `GET /api/v1/controls/:id`

Get a single control with full details including its requirement mappings.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "CTRL-AC-001",
    "title": "Multi-Factor Authentication",
    "description": "Enforce MFA for all user accounts...",
    "implementation_guidance": "...",
    "category": "technical",
    "status": "active",
    "is_custom": false,
    "source_template_id": "TPL-AC-001",
    "owner": {
      "id": "uuid",
      "name": "Bob Security",
      "email": "security@acme.example.com"
    },
    "secondary_owner": null,
    "evidence_requirements": "...",
    "test_criteria": "...",
    "metadata": {},
    "mappings": [
      {
        "id": "uuid",
        "requirement": {
          "id": "uuid",
          "identifier": "8.3.1",
          "title": "All user access to system components...",
          "framework": "PCI DSS",
          "framework_version": "4.0.1"
        },
        "strength": "primary",
        "notes": null
      },
      {
        "id": "uuid",
        "requirement": {
          "id": "uuid",
          "identifier": "CC6.1",
          "title": "Logical access security...",
          "framework": "SOC 2",
          "framework_version": "2024"
        },
        "strength": "primary",
        "notes": null
      }
    ],
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Control doesn't exist in this org

**Notes:**
- `mappings` includes the full requirement context (framework name, version) for UI display.
- This is the "control detail" view used in the dashboard's control library browser.

---

#### `PUT /api/v1/controls/:id`

Update a control's details.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer` (or the control's owner)

**Request:**

```json
{
  "title": "Multi-Factor Authentication (MFA)",
  "description": "Updated description...",
  "implementation_guidance": "Updated guidance...",
  "category": "technical",
  "evidence_requirements": "Updated requirements...",
  "test_criteria": "Updated criteria...",
  "metadata": {
    "risk_level": "critical"
  }
}
```

**Validation:**
- Same rules as POST, all fields optional
- `identifier` is NOT updatable after creation (stable reference)
- `status` is NOT updatable via this endpoint (use the status change endpoint)

**Response 200:**

Returns updated control (same shape as GET response, without `mappings`).

**Errors:**
- `403 FORBIDDEN` — Not authorized
- `404 NOT_FOUND` — Control not found

**Audit log:** `control.updated` with changed fields

**Notes:**
- Owner can update their own controls regardless of admin role.
- `metadata` is merged (not replaced) — send only keys to update. To remove a key, send `null` value.

---

#### `PUT /api/v1/controls/:id/owner`

Change control ownership.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "owner_id": "uuid",
  "secondary_owner_id": "uuid"
}
```

**Validation:**
- `owner_id`: optional, must be a user in the org
- `secondary_owner_id`: optional, must be different from `owner_id`. Send `null` to clear.
- At least one field required.

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "CTRL-AC-001",
    "owner": {
      "id": "uuid",
      "name": "Carol IT"
    },
    "secondary_owner": null,
    "message": "Ownership updated"
  }
}
```

**Audit log:** `control.owner_changed` with `{"previous_owner": "uuid", "new_owner": "uuid"}`

---

#### `PUT /api/v1/controls/:id/status`

Change a control's lifecycle status.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Request:**

```json
{
  "status": "active"
}
```

**Validation:**
- `status`: required, must be valid `control_status`
- Allowed transitions:
  - `draft` → `active`, `deprecated`
  - `active` → `under_review`, `deprecated`
  - `under_review` → `active`, `deprecated`
  - `deprecated` → `draft` (reactivation)

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "CTRL-AC-001",
    "status": "active",
    "previous_status": "draft",
    "message": "Status updated"
  }
}
```

**Errors:**
- `422 UNPROCESSABLE` — Invalid status transition

**Audit log:** `control.status_changed` with `{"old_status": "draft", "new_status": "active"}`

---

#### `DELETE /api/v1/controls/:id`

Deprecate a control (soft delete). Sets status to `deprecated`.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "identifier": "CTRL-AC-001",
    "status": "deprecated",
    "message": "Control deprecated. Mappings preserved for audit trail."
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Control not found
- `403 FORBIDDEN` — Not authorized

**Audit log:** `control.deprecated`

**Notes:**
- Soft delete — control and all its mappings are preserved.
- Deprecated controls are excluded from coverage calculations by default.
- To hard-delete, a future admin endpoint can be added.

---

### 5. Control Mappings

Link controls to framework requirements. This is the cross-framework mapping mechanism.

---

#### `GET /api/v1/controls/:control_id/mappings`

List all requirement mappings for a specific control.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "requirement": {
        "id": "uuid",
        "identifier": "8.3.1",
        "title": "All user access to system components...",
        "framework": {
          "identifier": "pci_dss",
          "name": "PCI DSS"
        },
        "version": "4.0.1"
      },
      "strength": "primary",
      "notes": null,
      "mapped_by": {
        "id": "uuid",
        "name": "Alice Compliance"
      },
      "created_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 4,
    "request_id": "uuid"
  }
}
```

---

#### `POST /api/v1/controls/:control_id/mappings`

Map a control to one or more requirements. Supports bulk mapping.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Request (single):**

```json
{
  "requirement_id": "uuid",
  "strength": "primary",
  "notes": "MFA directly satisfies this authentication requirement."
}
```

**Request (bulk):**

```json
{
  "mappings": [
    {
      "requirement_id": "uuid",
      "strength": "primary",
      "notes": "Direct match"
    },
    {
      "requirement_id": "uuid",
      "strength": "supporting",
      "notes": "Contributes to this control objective"
    }
  ]
}
```

**Validation:**
- `requirement_id`: required, must exist and be assessable
- `strength`: optional, default `primary`. Must be: `primary`, `supporting`, `partial`
- `notes`: optional, max 2000 chars
- Bulk: max 50 mappings per request
- Duplicate (control_id, requirement_id) pairs are rejected

**Response 201:**

```json
{
  "data": {
    "created": 2,
    "mappings": [
      {
        "id": "uuid",
        "control_id": "uuid",
        "requirement_id": "uuid",
        "strength": "primary",
        "created_at": "2026-02-20T10:00:00Z"
      }
    ]
  }
}
```

**Errors:**
- `409 CONFLICT` — Mapping already exists for this control-requirement pair
- `404 NOT_FOUND` — Control or requirement not found
- `422 UNPROCESSABLE` — Requirement is not assessable (section header), or requirement doesn't belong to an activated framework

**Audit log:** `control_mapping.created` with `{"control": "CTRL-AC-001", "requirement": "8.3.1", "framework": "pci_dss"}`

**Notes:**
- Bulk mapping is the primary flow — when mapping a control, you typically map it to requirements across multiple frameworks at once.
- Requirement must belong to a framework version that the org has activated. Can't map to frameworks you're not tracking.

---

#### `DELETE /api/v1/controls/:control_id/mappings/:mapping_id`

Remove a single control-to-requirement mapping.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `security_engineer`

**Response 200:**

```json
{
  "data": {
    "message": "Mapping removed"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — Mapping not found or doesn't belong to this control/org

**Audit log:** `control_mapping.deleted` with `{"control": "CTRL-AC-001", "requirement": "8.3.1", "framework": "pci_dss"}`

---

### 6. Cross-Framework Mapping Matrix

The signature view — "Which controls satisfy which requirements across all my frameworks?"

---

#### `GET /api/v1/mapping-matrix`

Generate the cross-framework mapping matrix for the org.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `framework_ids` | uuid[] | — | Filter to specific frameworks (comma-separated). Default: all activated. |
| `control_category` | string | — | Filter controls by category |
| `control_status` | string | `active` | Filter controls by status |
| `search` | string | — | Search controls by title/identifier |
| `page` | int | 1 | Page number (paginated by controls) |
| `per_page` | int | 20 | Controls per page (max 50) |

**Response 200:**

```json
{
  "data": {
    "frameworks": [
      {
        "id": "uuid",
        "identifier": "soc2",
        "name": "SOC 2",
        "version": "2024"
      },
      {
        "id": "uuid",
        "identifier": "pci_dss",
        "name": "PCI DSS",
        "version": "4.0.1"
      },
      {
        "id": "uuid",
        "identifier": "iso27001",
        "name": "ISO 27001",
        "version": "2022"
      }
    ],
    "controls": [
      {
        "id": "uuid",
        "identifier": "CTRL-AC-001",
        "title": "Multi-Factor Authentication",
        "category": "technical",
        "status": "active",
        "mappings_by_framework": {
          "soc2": [
            { "requirement_id": "uuid", "identifier": "CC6.1", "strength": "primary" }
          ],
          "pci_dss": [
            { "requirement_id": "uuid", "identifier": "8.3.1", "strength": "primary" },
            { "requirement_id": "uuid", "identifier": "8.3.2", "strength": "primary" }
          ],
          "iso27001": [
            { "requirement_id": "uuid", "identifier": "A.8.5", "strength": "primary" }
          ]
        }
      },
      {
        "id": "uuid",
        "identifier": "CTRL-AC-002",
        "title": "Role-Based Access Control",
        "category": "technical",
        "status": "active",
        "mappings_by_framework": {
          "soc2": [
            { "requirement_id": "uuid", "identifier": "CC6.3", "strength": "primary" }
          ],
          "pci_dss": [
            { "requirement_id": "uuid", "identifier": "7.2.1", "strength": "primary" }
          ],
          "iso27001": [
            { "requirement_id": "uuid", "identifier": "A.5.15", "strength": "primary" },
            { "requirement_id": "uuid", "identifier": "A.8.3", "strength": "supporting" }
          ]
        }
      }
    ]
  },
  "meta": {
    "total_controls": 312,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- `mappings_by_framework` uses framework `identifier` as key for stable references.
- This powers the dashboard's "mapping matrix" view — a table where rows are controls and columns are frameworks.
- Paginated by controls to keep response size manageable.
- Only includes frameworks the org has activated.
- Default `control_status=active` excludes deprecated controls.

---

### 7. Bulk Operations

---

#### `POST /api/v1/controls/bulk-status`

Change status for multiple controls at once.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "control_ids": ["uuid", "uuid", "uuid"],
  "status": "active"
}
```

**Validation:**
- `control_ids`: required, max 100 IDs
- `status`: required, valid `control_status`
- All controls must exist and belong to the org
- All status transitions must be valid

**Response 200:**

```json
{
  "data": {
    "updated": 3,
    "failed": 0,
    "results": [
      { "id": "uuid", "identifier": "CTRL-AC-001", "status": "active", "success": true },
      { "id": "uuid", "identifier": "CTRL-AC-002", "status": "active", "success": true },
      { "id": "uuid", "identifier": "CTRL-AC-003", "status": "active", "success": true }
    ]
  }
}
```

**Errors:**
- Partial success is allowed — `failed > 0` with error details per control

**Audit log:** One `control.status_changed` entry per control

---

### 8. Statistics

---

#### `GET /api/v1/controls/stats`

Aggregate statistics about the org's control library.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "total": 312,
    "by_status": {
      "draft": 15,
      "active": 280,
      "under_review": 12,
      "deprecated": 5
    },
    "by_category": {
      "technical": 145,
      "administrative": 98,
      "physical": 22,
      "operational": 47
    },
    "custom_count": 18,
    "library_count": 294,
    "unowned_count": 3,
    "unmapped_count": 8,
    "frameworks_coverage": [
      {
        "framework": "SOC 2",
        "version": "2024",
        "in_scope": 60,
        "covered": 55,
        "gaps": 5,
        "coverage_pct": 91.67
      },
      {
        "framework": "PCI DSS",
        "version": "4.0.1",
        "in_scope": 240,
        "covered": 195,
        "gaps": 45,
        "coverage_pct": 81.25
      }
    ]
  }
}
```

**Notes:**
- Designed for the dashboard home page stat cards and compliance posture display.
- `frameworks_coverage` provides per-framework compliance posture at a glance.

---

## Summary of All Endpoints

| # | Method | Path | Roles | Description |
|---|--------|------|-------|-------------|
| 1 | GET | `/api/v1/frameworks` | All | List framework catalog |
| 2 | GET | `/api/v1/frameworks/:id` | All | Get framework + versions |
| 3 | GET | `/api/v1/frameworks/:fid/versions/:vid` | All | Get version details |
| 4 | GET | `/api/v1/frameworks/:fid/versions/:vid/requirements` | All | List/tree requirements |
| 5 | GET | `/api/v1/org-frameworks` | All | List activated frameworks |
| 6 | POST | `/api/v1/org-frameworks` | CISO, CM | Activate framework |
| 7 | PUT | `/api/v1/org-frameworks/:id` | CISO, CM | Update activation |
| 8 | DELETE | `/api/v1/org-frameworks/:id` | CISO, CM | Deactivate framework |
| 9 | GET | `/api/v1/org-frameworks/:id/coverage` | All | Coverage gap analysis |
| 10 | GET | `/api/v1/org-frameworks/:id/scoping` | All | List scoping decisions |
| 11 | PUT | `/api/v1/org-frameworks/:ofid/requirements/:rid/scope` | CISO, CM | Set scope |
| 12 | DELETE | `/api/v1/org-frameworks/:ofid/requirements/:rid/scope` | CISO, CM | Reset scope |
| 13 | GET | `/api/v1/controls` | All | List controls |
| 14 | POST | `/api/v1/controls` | CISO, CM, SE | Create control |
| 15 | GET | `/api/v1/controls/:id` | All | Get control + mappings |
| 16 | PUT | `/api/v1/controls/:id` | CISO, CM, SE, Owner | Update control |
| 17 | PUT | `/api/v1/controls/:id/owner` | CISO, CM | Change ownership |
| 18 | PUT | `/api/v1/controls/:id/status` | CISO, CM, SE | Change status |
| 19 | DELETE | `/api/v1/controls/:id` | CISO, CM | Deprecate control |
| 20 | GET | `/api/v1/controls/:cid/mappings` | All | List control mappings |
| 21 | POST | `/api/v1/controls/:cid/mappings` | CISO, CM, SE | Create mappings (bulk) |
| 22 | DELETE | `/api/v1/controls/:cid/mappings/:mid` | CISO, CM, SE | Remove mapping |
| 23 | GET | `/api/v1/mapping-matrix` | All | Cross-framework matrix |
| 24 | POST | `/api/v1/controls/bulk-status` | CISO, CM | Bulk status change |
| 25 | GET | `/api/v1/controls/stats` | All | Control statistics |

**Role abbreviations:** CISO = `ciso`, CM = `compliance_manager`, SE = `security_engineer`, Owner = control owner (any role)

**Total: 25 new endpoints**

---

## Go Implementation Notes

### New Files

```
api/internal/
├── handlers/
│   ├── frameworks.go         # Endpoints 1-4
│   ├── org_frameworks.go     # Endpoints 5-9
│   ├── requirement_scopes.go # Endpoints 10-12
│   ├── controls.go           # Endpoints 13-19, 24-25
│   ├── control_mappings.go   # Endpoints 20-22
│   └── mapping_matrix.go     # Endpoint 23
├── models/
│   ├── framework.go          # Framework, FrameworkVersion structs
│   ├── requirement.go        # Requirement struct + tree builder
│   ├── org_framework.go      # OrgFramework struct + stats
│   ├── requirement_scope.go  # RequirementScope struct
│   ├── control.go            # Control struct + search
│   └── control_mapping.go    # ControlMapping struct
└── services/
    ├── framework.go          # Framework activation + seeding logic
    ├── control.go            # Control lifecycle + bulk ops
    └── coverage.go           # Coverage calculation + gap analysis
```

### Route Registration

```go
// Framework catalog (system-level, read-only)
fw := v1.Group("/frameworks")
fw.Use(middleware.Auth())
{
    fw.GET("", handlers.ListFrameworks)
    fw.GET("/:id", handlers.GetFramework)
    fw.GET("/:id/versions/:vid", handlers.GetFrameworkVersion)
    fw.GET("/:id/versions/:vid/requirements", handlers.ListRequirements)
}

// Org frameworks
of := v1.Group("/org-frameworks")
of.Use(middleware.Auth(), middleware.Org())
{
    of.GET("", handlers.ListOrgFrameworks)
    of.POST("", middleware.RBAC("ciso", "compliance_manager"), handlers.ActivateFramework)
    of.PUT("/:id", middleware.RBAC("ciso", "compliance_manager"), handlers.UpdateOrgFramework)
    of.DELETE("/:id", middleware.RBAC("ciso", "compliance_manager"), handlers.DeactivateFramework)
    of.GET("/:id/coverage", handlers.GetCoverage)
    of.GET("/:id/scoping", handlers.ListScoping)
    of.PUT("/:id/requirements/:rid/scope", middleware.RBAC("ciso", "compliance_manager"), handlers.SetScope)
    of.DELETE("/:id/requirements/:rid/scope", middleware.RBAC("ciso", "compliance_manager"), handlers.ResetScope)
}

// Controls
ctrl := v1.Group("/controls")
ctrl.Use(middleware.Auth(), middleware.Org())
{
    ctrl.GET("", handlers.ListControls)
    ctrl.POST("", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.CreateControl)
    ctrl.GET("/stats", handlers.GetControlStats)
    ctrl.POST("/bulk-status", middleware.RBAC("ciso", "compliance_manager"), handlers.BulkControlStatus)
    ctrl.GET("/:id", handlers.GetControl)
    ctrl.PUT("/:id", handlers.UpdateControl) // owner check in handler
    ctrl.PUT("/:id/owner", middleware.RBAC("ciso", "compliance_manager"), handlers.ChangeControlOwner)
    ctrl.PUT("/:id/status", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.ChangeControlStatus)
    ctrl.DELETE("/:id", middleware.RBAC("ciso", "compliance_manager"), handlers.DeprecateControl)
    ctrl.GET("/:id/mappings", handlers.ListControlMappings)
    ctrl.POST("/:id/mappings", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.CreateControlMappings)
    ctrl.DELETE("/:id/mappings/:mid", middleware.RBAC("ciso", "compliance_manager", "security_engineer"), handlers.DeleteControlMapping)
}

// Mapping matrix
v1.GET("/mapping-matrix", middleware.Auth(), middleware.Org(), handlers.GetMappingMatrix)
```
