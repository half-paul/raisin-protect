# ADR-001: CDE Scoping Module Architecture

## Status

Accepted

## Date

2026-03-15

## Context

PCI DSS v4.0.1 Requirements 1 and 11.4 mandate that organizations explicitly document and maintain their Cardholder Data Environment (CDE) scope. This includes:

- Identifying all system components in scope (those that store, process, or transmit CHD/SAD, or are connected to them)
- Documenting out-of-scope systems with justification for exclusion
- Mapping data flows between in-scope and connected systems
- Tracking network segmentation testing to verify isolation of out-of-scope systems
- Producing network topology documentation for QSA review

The current `requirement_scopes` table (migration 013) is a simple boolean flag against PCI DSS requirement items. It does not represent the CDE asset inventory — the physical/logical systems, applications, and network segments that are in or out of scope. These are fundamentally different concerns:

| Current `requirement_scopes` | New CDE Scoping Module |
|---|---|
| "Is PCI DSS Req 1.2.1 applicable to us?" | "Is this firewall appliance in CDE scope?" |
| Requirement-level scoping decision | System/asset-level inventory and classification |
| Single boolean + justification | Rich asset model with data flows and topology |

This ADR defines the architecture for a dedicated CDE Scoping Module that satisfies QSA expectations for PCI DSS v4.0.1 assessments.

---

## Decision

Implement the CDE Scoping Module as a first-class domain within Raisin Protect, comprising:

1. A **CDE asset inventory** with asset types (system, application, network segment, device)
2. A **scope classification workflow** with explicit statuses and mandatory justifications
3. A **data flow registry** documenting how CHD/SAD moves between assets
4. A **network segmentation test tracker** for quarterly and on-demand verification
5. A **REST API** following existing codebase patterns (Gin, JWT auth, RBAC middleware)

---

## Data Model

### New Tables (Sprint 11 migrations)

#### `cde_assets` (migration 053)

Represents any system component relevant to CDE scoping.

```sql
CREATE TABLE cde_assets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    asset_type      TEXT NOT NULL CHECK (asset_type IN (
                        'server', 'workstation', 'network_device', 'application',
                        'virtual_machine', 'container_cluster', 'cloud_service',
                        'network_segment', 'storage_system', 'terminal'
                    )),
    -- Scope classification
    scope_status    TEXT NOT NULL DEFAULT 'under_review' CHECK (scope_status IN (
                        'in_scope',      -- Stores, processes, or transmits CHD/SAD
                        'connected_to',  -- Connected to CDE but does not handle CHD/SAD
                        'out_of_scope',  -- Isolated from CDE with verified segmentation
                        'under_review'   -- Pending classification decision
                    )),
    scope_justification TEXT,           -- Required when out_of_scope or connected_to
    scope_decided_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    scope_decided_at    TIMESTAMPTZ,
    -- Asset metadata
    hostname        TEXT,
    ip_address      INET,
    location        TEXT,              -- Physical or logical location (e.g., "AWS us-east-1", "DC Rack 12")
    os_or_platform  TEXT,
    version         TEXT,
    owner_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    tags            TEXT[] NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    -- Audit
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_cde_assets_org ON cde_assets (org_id);
CREATE INDEX idx_cde_assets_org_scope ON cde_assets (org_id, scope_status);
CREATE INDEX idx_cde_assets_org_type ON cde_assets (org_id, asset_type);
CREATE INDEX idx_cde_assets_org_active ON cde_assets (org_id, is_active);
```

**Design notes:**
- `scope_status = 'in_scope'` does NOT require justification (default state for CDE systems)
- `scope_status = 'out_of_scope'` REQUIRES justification AND a linked segmentation test (enforced at API layer)
- `scope_status = 'connected_to'` REQUIRES justification describing the nature of connectivity and why CHD/SAD does not transit this system
- `scope_status = 'under_review'` is the initial state for newly added assets; blocks report generation until resolved

#### `cde_data_flows` (migration 054)

Documents how CHD/SAD flows between assets. Required by PCI DSS Req 1.2.4 (accurate data-flow diagrams).

```sql
CREATE TABLE cde_data_flows (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_asset_id UUID NOT NULL REFERENCES cde_assets(id) ON DELETE CASCADE,
    target_asset_id UUID NOT NULL REFERENCES cde_assets(id) ON DELETE CASCADE,
    -- Flow classification
    flow_type       TEXT NOT NULL CHECK (flow_type IN (
                        'chd_transmission',  -- Cardholder Data (PAN, etc.)
                        'sad_transmission',  -- Sensitive Authentication Data
                        'management',        -- Admin/management traffic (no CHD)
                        'authentication',    -- Auth/SSO traffic
                        'monitoring',        -- Logging/monitoring traffic
                        'backup',            -- Backup/replication
                        'other'
                    )),
    protocol        TEXT,               -- e.g., "TLS 1.3", "HTTPS", "SSH"
    port            INTEGER,
    direction       TEXT NOT NULL DEFAULT 'unidirectional' CHECK (direction IN ('unidirectional', 'bidirectional')),
    encryption      BOOLEAN NOT NULL DEFAULT TRUE,
    description     TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    -- Audit
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT no_self_flow CHECK (source_asset_id != target_asset_id)
);

CREATE INDEX idx_cde_data_flows_org ON cde_data_flows (org_id);
CREATE INDEX idx_cde_data_flows_source ON cde_data_flows (source_asset_id);
CREATE INDEX idx_cde_data_flows_target ON cde_data_flows (target_asset_id);
```

#### `cde_segmentation_tests` (migration 055)

Tracks network segmentation testing results. PCI DSS Req 11.4.1 requires testing at least every 6 months and after significant changes.

```sql
CREATE TABLE cde_segmentation_tests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    -- Test scope: which asset pair was tested, or which out-of-scope segment
    tested_asset_id UUID REFERENCES cde_assets(id) ON DELETE SET NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    -- Test details
    test_method     TEXT NOT NULL CHECK (test_method IN (
                        'internal_scan',
                        'external_penetration',
                        'firewall_rule_review',
                        'manual_verification',
                        'automated_tool'
                    )),
    tool_used       TEXT,
    performed_by    TEXT NOT NULL,      -- Tester name or company (may be external QSA/pentester)
    performed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    -- Results
    test_status     TEXT NOT NULL DEFAULT 'planned' CHECK (test_status IN (
                        'planned', 'in_progress', 'passed', 'failed', 'remediated'
                    )),
    result_summary  TEXT,
    findings_count  INTEGER NOT NULL DEFAULT 0,
    -- Dates
    scheduled_date  DATE,
    performed_date  DATE,
    next_due_date   DATE,              -- 6 months from performed_date (set by API)
    -- Evidence linkage (reuses existing evidence system)
    evidence_note   TEXT,              -- Before evidence_links table link (Sprint 11 shortcut)
    -- Audit
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cde_seg_tests_org ON cde_segmentation_tests (org_id);
CREATE INDEX idx_cde_seg_tests_asset ON cde_segmentation_tests (tested_asset_id);
CREATE INDEX idx_cde_seg_tests_due ON cde_segmentation_tests (org_id, next_due_date);
CREATE INDEX idx_cde_seg_tests_status ON cde_segmentation_tests (org_id, test_status);
```

#### `cde_scope_history` (migration 055, same file)

Immutable audit trail for scope classification changes. PCI DSS requires demonstrating that scoping decisions are controlled and reviewed.

```sql
CREATE TABLE cde_scope_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    asset_id        UUID NOT NULL REFERENCES cde_assets(id) ON DELETE CASCADE,
    previous_status TEXT,
    new_status      TEXT NOT NULL,
    justification   TEXT,
    changed_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cde_scope_history_asset ON cde_scope_history (asset_id, changed_at DESC);
```

### Relationship to Existing `requirement_scopes`

The `requirement_scopes` table is **not replaced** — it addresses a different concern (whether a PCI DSS requirement applies to the organization). The new CDE tables address **which systems** are in scope. These two concepts are complementary:

- `requirement_scopes`: "Does PCI DSS Req 8.3.6 (MFA) apply to our organization?"
- `cde_assets`: "Is the Salesforce CRM in our CDE? Connected to it? Or fully isolated?"

---

## Scope Justification Workflow

```
Asset Created
      │
      ▼
  under_review ──────────────────────────────────────────────────────────────┐
      │                                                                       │
      ├──[scope decision: stores/processes CHD]──► in_scope                 │
      │                                               │                      │
      ├──[scope decision: connected, no CHD]──► connected_to                │
      │                         (justification required)                     │
      │                                                                      │
      └──[scope decision: fully isolated]──► out_of_scope                   │
                               (justification + segmentation test required)  │
                                                                             │
  Any status ──[material change to system]──────────────────────────────────┘
                                       (triggers re-review, history recorded)
```

**API enforcement rules:**
1. Transition to `out_of_scope` requires non-empty `scope_justification` AND at least one linked `cde_segmentation_tests` record with `test_status = 'passed'`
2. Transition to `connected_to` requires non-empty `scope_justification`
3. All status transitions are recorded in `cde_scope_history` (immutable)
4. Re-classifying a previously `out_of_scope` asset back to `under_review` sends an alert to the org admin (via existing alert engine)

---

## Network Segmentation Testing Tracking

Segmentation test lifecycle:
1. **Plan**: Create a `cde_segmentation_tests` record with `scheduled_date` and `test_method`
2. **Perform**: Update `test_status → in_progress`, record `performed_date`
3. **Result**: Update `test_status → passed | failed`, set `next_due_date = performed_date + 6 months`
4. **Remediation** (on failure): Update to `remediated` after fixing findings, triggers re-test requirement

A background worker (extending the existing monitoring worker) checks for overdue segmentation tests (`next_due_date < NOW()`) and fires alerts via the existing alert engine.

---

## Data Flow Documentation Approach

Data flows are managed as explicit graph edges between `cde_assets` nodes. The API exposes:
- CRUD for individual flow records
- A `GET /api/v1/cde/data-flows/diagram` endpoint returning a D3-compatible adjacency list for frontend rendering
- Flows with `flow_type IN ('chd_transmission', 'sad_transmission')` are highlighted as "critical paths" in the UI

We do **not** attempt to auto-discover data flows in Sprint 11. Manual documentation is required — auto-discovery from network scans or cloud APIs is deferred to a future sprint.

---

## Multi-Tenant Isolation

All CDE tables include `org_id` with `REFERENCES organizations(id) ON DELETE CASCADE` and are indexed by `org_id`. The existing PostgreSQL Row Level Security (RLS) pattern used elsewhere in the platform will be applied to all CDE tables, ensuring:

- Queries are automatically scoped to the authenticated user's organization
- No org_id filtering is needed at the application layer (defense in depth)
- Cross-tenant data leakage is prevented at the database layer

The existing `middleware.AuthRequired()` + `handlers.SetDB(database)` pattern provides the JWT-based org context that RLS policies consume.

---

## API Endpoint Design

Following existing REST patterns in `main.go`:

```
// CDE Scoping Module (Sprint 11)
cde := protected.Group("/cde")

// CDE Assets
cde.GET("/assets",              handlers.ListCDEAssets)
cde.POST("/assets",             middleware.RequireRoles(models.CDEManageRoles...), handlers.CreateCDEAsset)
cde.GET("/assets/summary",      handlers.GetCDEScopeSummary)      // counts by status/type
cde.GET("/assets/:id",          handlers.GetCDEAsset)
cde.PUT("/assets/:id",          handlers.UpdateCDEAsset)          // owner/admin check in handler
cde.PUT("/assets/:id/scope",    middleware.RequireRoles(models.CDEManageRoles...), handlers.SetCDEAssetScope)
cde.GET("/assets/:id/history",  handlers.GetCDEAssetScopeHistory)

// Data Flows
cde.GET("/data-flows",          handlers.ListCDEDataFlows)
cde.POST("/data-flows",         middleware.RequireRoles(models.CDEManageRoles...), handlers.CreateCDEDataFlow)
cde.GET("/data-flows/diagram",  handlers.GetCDEDataFlowDiagram)   // adjacency list for D3 rendering
cde.GET("/data-flows/:id",      handlers.GetCDEDataFlow)
cde.PUT("/data-flows/:id",      middleware.RequireRoles(models.CDEManageRoles...), handlers.UpdateCDEDataFlow)
cde.DELETE("/data-flows/:id",   middleware.RequireRoles(models.CDEManageRoles...), handlers.DeleteCDEDataFlow)

// Segmentation Tests
cde.GET("/segmentation-tests",             handlers.ListSegmentationTests)
cde.POST("/segmentation-tests",            middleware.RequireRoles(models.CDEManageRoles...), handlers.CreateSegmentationTest)
cde.GET("/segmentation-tests/overdue",     handlers.GetOverdueSegmentationTests)
cde.GET("/segmentation-tests/:id",         handlers.GetSegmentationTest)
cde.PUT("/segmentation-tests/:id",         middleware.RequireRoles(models.CDEManageRoles...), handlers.UpdateSegmentationTest)
cde.PUT("/segmentation-tests/:id/result",  middleware.RequireRoles(models.CDEManageRoles...), handlers.RecordSegmentationTestResult)
```

**RBAC roles for CDE module:**
- `CDEManageRoles`: `compliance_manager`, `admin`, `owner` — can create/update assets and scope decisions
- All authenticated users can `GET` CDE data (read-only visibility for auditors and contributors)

**Key response models:**

`CDEAssetResponse`:
```json
{
  "id": "uuid",
  "name": "Payment Processing Server (PPS-01)",
  "asset_type": "server",
  "scope_status": "in_scope",
  "scope_justification": null,
  "scope_decided_by": { "id": "uuid", "name": "Alice Chen" },
  "scope_decided_at": "2026-03-15T10:00:00Z",
  "hostname": "pps-01.internal",
  "ip_address": "10.1.5.10",
  "location": "AWS us-east-1 / VPC-PCI",
  "is_active": true,
  "data_flow_count": 4,
  "open_segmentation_tests": 0
}
```

`CDEScopeSummaryResponse`:
```json
{
  "total_assets": 47,
  "in_scope": 12,
  "connected_to": 8,
  "out_of_scope": 24,
  "under_review": 3,
  "overdue_segmentation_tests": 1,
  "data_flows": { "total": 31, "chd_flows": 9, "unencrypted": 0 }
}
```

---

## Alternatives Considered

### Option A: Extend `requirement_scopes` (Rejected)

Adding asset tracking columns to the existing `requirement_scopes` table was considered and rejected. The two concerns (requirement applicability vs. asset inventory) have different cardinalities (1 requirement vs. N assets), different lifecycles, and different reporting outputs. Merging them would create a confused, hard-to-query table that satisfies neither use case cleanly.

### Option B: Use a Graph Database (Rejected)

A graph database (e.g., Neo4j) would model data flows more elegantly. Rejected because: (a) it adds significant operational complexity, (b) the graph is relatively small (<1000 nodes for most merchants), (c) PostgreSQL handles the adjacency list pattern adequately, and (d) the team does not have graph DB expertise. PostgreSQL with proper indexes is sufficient.

### Option C: Import from Network Scanner APIs (Deferred)

Automated discovery from AWS Config, Nessus, or Qualys would reduce manual effort. Deferred to Sprint 13+. Sprint 11 focuses on the data model and manual entry — this is the critical unblocking path for QSA review.

---

## Consequences

**Positive:**
- Satisfies PCI DSS v4.0.1 Req 1.2.4 (data flow diagrams) and Req 11.4.1 (segmentation testing)
- Provides structured CDE inventory that can be referenced by AOC/ROC document generator (ADR-002)
- Immutable scope history supports QSA audit trail requirements
- Re-uses existing evidence, alert, and RBAC infrastructure — no new dependencies

**Negative / Trade-offs:**
- Manual data entry burden — no auto-discovery in Sprint 11
- `under_review` assets will block AOC generation until classified (acceptable: classification is required by PCI DSS anyway)
- Segmentation test overdue alerts may generate noise if test scheduling is not maintained

**Implementation order (unblocks other work):**
1. Migrations 053-055 (Dana)
2. Handler stubs + models (Logan)
3. Background worker extension for overdue segmentation tests (Logan)
4. Frontend CDE asset inventory page (Alex)
5. Data flow diagram visualization (Alex)

---

## References

- PCI DSS v4.0.1 Requirements 1.2.4, 11.4.1
- Existing migration pattern: `db/migrations/013_requirement_scopes.sql`
- Existing RLS pattern: `db/migrations/003_users.sql`
- Monitoring worker: `api/internal/workers/`
- Alert engine: `api/internal/handlers/alerts.go`
