# Sprint 2 — Database Schema: Frameworks & Controls

## Overview

Sprint 2 introduces the core GRC data model: compliance frameworks with versioning, hierarchical requirements, a control library with ownership, and cross-framework control mappings. This is the backbone of the compliance engine (spec §3).

**Key design decisions:**
- **System-level catalog vs. org-level state**: Frameworks, versions, and requirements are **system-level** (shared across all tenants — SOC 2 is SOC 2 for everyone). Org-specific state lives in `org_frameworks`, `controls`, `control_mappings`, and `requirement_scopes`.
- **Hierarchical requirements**: Self-referential `parent_id` enables arbitrary nesting (e.g., PCI DSS 6 → 6.4 → 6.4.3 → 6.4.3.1).
- **Cross-framework mapping**: A single control can satisfy requirements across multiple frameworks via `control_mappings`. This is the key multiplier — one piece of evidence covers many frameworks.

---

## Entity Relationship Diagram

```
                    ┌─────────────────────────────┐
                    │   SYSTEM CATALOG (no org_id) │
                    │                             │
                    │   frameworks                │
                    │       1──∞ framework_versions│
                    │              1──∞ requirements│
                    │              (self-ref parent)│
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────┴──────────────┐
                    │   ORG-SCOPED (has org_id)    │
                    │                             │
 organizations ─┬──▶ org_frameworks               │
                │   │   (activates version)       │
                │   │                             │
                ├──▶ requirement_scopes           │
                │   │   (in/out scope per req)    │
                │   │                             │
                ├──▶ controls                     │
                │   │   1──∞ control_mappings ──▶ requirements
                │   │                             │
                └──▶ audit_log (extended)          │
                    └─────────────────────────────┘
```

**Relationships:**
```
frameworks          1──∞  framework_versions
framework_versions  1──∞  requirements
requirements        1──∞  requirements          (parent_id self-ref)
requirements        1──∞  control_mappings
controls            1──∞  control_mappings
organizations       1──∞  org_frameworks
organizations       1──∞  controls
organizations       1──∞  requirement_scopes
org_frameworks      ∞──1  frameworks
org_frameworks      ∞──1  framework_versions
controls            ∞──1  users                 (owner_id)
controls            ∞──1  users                 (secondary_owner_id)
requirement_scopes  ∞──1  users                 (scoped_by)
```

---

## New Enums

```sql
-- Framework taxonomy (from spec §3.1.1)
CREATE TYPE framework_category AS ENUM (
    'security_privacy',    -- SOC 2, ISO 27001, ISO 27701
    'payment',             -- PCI DSS
    'data_privacy',        -- GDPR, CCPA/CPRA
    'ai_governance',       -- ISO 42001, EU AI Act, NIST AI RMF
    'industry',            -- TISAX, CSA STAR, CIS Controls
    'custom'               -- User-defined frameworks
);

-- Framework version lifecycle
CREATE TYPE framework_version_status AS ENUM (
    'draft',               -- Being prepared, not yet usable
    'active',              -- Current and available for activation
    'deprecated',          -- Superseded but still supported
    'sunset'               -- No longer supported, read-only
);

-- Control categories (from spec §3.2.1)
CREATE TYPE control_category AS ENUM (
    'technical',
    'administrative',
    'physical',
    'operational'
);

-- Control lifecycle (from spec §3.2.1)
CREATE TYPE control_status AS ENUM (
    'draft',
    'active',
    'under_review',
    'deprecated'
);

-- Org framework activation status
CREATE TYPE org_framework_status AS ENUM (
    'active',              -- Framework is being tracked
    'inactive'             -- Paused/deactivated (preserves history)
);
```

### Extend Existing Enums

```sql
-- Add Sprint 2 actions to audit_action enum
ALTER TYPE audit_action ADD VALUE 'framework.activated';
ALTER TYPE audit_action ADD VALUE 'framework.deactivated';
ALTER TYPE audit_action ADD VALUE 'framework.version_changed';
ALTER TYPE audit_action ADD VALUE 'control.created';
ALTER TYPE audit_action ADD VALUE 'control.updated';
ALTER TYPE audit_action ADD VALUE 'control.status_changed';
ALTER TYPE audit_action ADD VALUE 'control.deprecated';
ALTER TYPE audit_action ADD VALUE 'control.owner_changed';
ALTER TYPE audit_action ADD VALUE 'control_mapping.created';
ALTER TYPE audit_action ADD VALUE 'control_mapping.deleted';
ALTER TYPE audit_action ADD VALUE 'requirement.scoped';
```

---

## Tables

### frameworks

The master catalog of compliance frameworks. System-level — no `org_id`. These are the same SOC 2, PCI DSS, etc. for every tenant.

```sql
CREATE TABLE frameworks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier      VARCHAR(50) NOT NULL UNIQUE,           -- machine key: 'soc2', 'pci_dss', 'iso27001'
    name            VARCHAR(255) NOT NULL,                 -- display: 'SOC 2', 'PCI DSS', 'ISO 27001'
    description     TEXT,
    category        framework_category NOT NULL,
    website_url     VARCHAR(500),                          -- official framework URL
    logo_url        VARCHAR(500),                          -- framework logo for UI
    is_custom       BOOLEAN NOT NULL DEFAULT FALSE,        -- TRUE for user-created frameworks (future)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_frameworks_category ON frameworks (category);
CREATE INDEX idx_frameworks_identifier ON frameworks (identifier);

-- Trigger
CREATE TRIGGER trg_frameworks_updated_at
    BEFORE UPDATE ON frameworks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `identifier` is the stable machine key used in APIs and code (e.g., `pci_dss`). `name` is the human display name.
- `is_custom` reserved for Sprint 2+ when orgs can define their own frameworks (spec §3.1.1 "Custom" category).
- No `org_id` — frameworks are shared. Org-level activation is in `org_frameworks`.

---

### framework_versions

Each framework can have multiple versions active simultaneously (spec §3.1.2). E.g., PCI DSS v3.2.1 and v4.0.1 can both be active while orgs transition.

```sql
CREATE TABLE framework_versions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_id        UUID NOT NULL REFERENCES frameworks(id) ON DELETE CASCADE,
    version             VARCHAR(50) NOT NULL,              -- '4.0.1', '2022', 'Type II'
    display_name        VARCHAR(255) NOT NULL,             -- 'PCI DSS v4.0.1', 'SOC 2 Type II (2024)'
    status              framework_version_status NOT NULL DEFAULT 'draft',
    effective_date      DATE,                              -- when this version became official
    sunset_date         DATE,                              -- when this version expires
    changelog           TEXT,                              -- what changed from previous version
    total_requirements  INT NOT NULL DEFAULT 0,            -- denormalized count of assessable requirements
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One version string per framework
    CONSTRAINT uq_framework_version UNIQUE (framework_id, version)
);

-- Indexes
CREATE INDEX idx_framework_versions_framework_id ON framework_versions (framework_id);
CREATE INDEX idx_framework_versions_status ON framework_versions (status);

-- Trigger
CREATE TRIGGER trg_framework_versions_updated_at
    BEFORE UPDATE ON framework_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `total_requirements` is denormalized for performance (updated when requirements are seeded/modified).
- `effective_date` / `sunset_date` support transition tracking (spec §3.1.2: "Transition period tracking with deadline alerts").
- `changelog` captures what changed between versions for gap analysis.

---

### requirements

The individual requirements within a framework version. Hierarchical via `parent_id` — supports arbitrary nesting depth.

```sql
CREATE TABLE requirements (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework_version_id    UUID NOT NULL REFERENCES framework_versions(id) ON DELETE CASCADE,
    parent_id               UUID REFERENCES requirements(id) ON DELETE CASCADE,
    identifier              VARCHAR(50) NOT NULL,          -- '6.4.3', 'CC6.1', 'A.8.1'
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    guidance                TEXT,                          -- implementation guidance
    section_order           INT NOT NULL DEFAULT 0,        -- ordering within same parent
    depth                   INT NOT NULL DEFAULT 0,        -- nesting level (0 = top-level)
    is_assessable           BOOLEAN NOT NULL DEFAULT TRUE, -- FALSE for section headers, TRUE for leaf requirements
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique identifier within a framework version
    CONSTRAINT uq_requirement_identifier UNIQUE (framework_version_id, identifier)
);

-- Indexes
CREATE INDEX idx_requirements_framework_version ON requirements (framework_version_id);
CREATE INDEX idx_requirements_parent ON requirements (parent_id);
CREATE INDEX idx_requirements_assessable ON requirements (framework_version_id, is_assessable)
    WHERE is_assessable = TRUE;
CREATE INDEX idx_requirements_depth ON requirements (framework_version_id, depth, section_order);

-- Trigger
CREATE TRIGGER trg_requirements_updated_at
    BEFORE UPDATE ON requirements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `is_assessable` distinguishes section headers (e.g., "Requirement 6: Develop and maintain secure systems") from actual testable requirements (e.g., "6.4.3: All payment page scripts..."). Only assessable requirements can have control mappings.
- `section_order` determines display order among siblings. Combined with `depth`, enables proper tree rendering.
- `guidance` stores implementation notes that help users understand what the requirement means in practice.
- No `org_id` — requirements are the same for everyone. Org-specific scoping lives in `requirement_scopes`.

---

### org_frameworks

Which frameworks (and which version) each organization has activated. This is the org-level "I'm pursuing SOC 2 compliance" declaration.

```sql
CREATE TABLE org_frameworks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    framework_id        UUID NOT NULL REFERENCES frameworks(id) ON DELETE RESTRICT,
    active_version_id   UUID NOT NULL REFERENCES framework_versions(id) ON DELETE RESTRICT,
    status              org_framework_status NOT NULL DEFAULT 'active',
    target_date         DATE,                              -- target compliance date
    notes               TEXT,                              -- org notes about this framework
    activated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deactivated_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One activation per framework per org
    CONSTRAINT uq_org_framework UNIQUE (org_id, framework_id)
);

-- Indexes
CREATE INDEX idx_org_frameworks_org_id ON org_frameworks (org_id);
CREATE INDEX idx_org_frameworks_status ON org_frameworks (org_id, status);
CREATE INDEX idx_org_frameworks_framework ON org_frameworks (framework_id);

-- Trigger
CREATE TRIGGER trg_org_frameworks_updated_at
    BEFORE UPDATE ON org_frameworks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `active_version_id` tracks which version the org is working against. Version switching is an explicit action (spec §3.1.2: "Configurable activation").
- `target_date` is for compliance program planning ("We need SOC 2 by Q3 2026").
- `ON DELETE RESTRICT` on framework/version FKs prevents accidentally deleting a framework that orgs depend on.

---

### requirement_scopes

Per-org, per-requirement scoping decisions. Not every requirement applies to every org — e.g., PCI DSS requirements about physical access may not apply to a cloud-only company.

```sql
CREATE TABLE requirement_scopes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requirement_id  UUID NOT NULL REFERENCES requirements(id) ON DELETE CASCADE,
    in_scope        BOOLEAN NOT NULL DEFAULT TRUE,
    justification   TEXT,                                  -- required when out-of-scope
    scoped_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One scoping decision per org per requirement
    CONSTRAINT uq_requirement_scope UNIQUE (org_id, requirement_id)
);

-- Indexes
CREATE INDEX idx_requirement_scopes_org ON requirement_scopes (org_id);
CREATE INDEX idx_requirement_scopes_requirement ON requirement_scopes (requirement_id);
CREATE INDEX idx_requirement_scopes_out_of_scope ON requirement_scopes (org_id, in_scope)
    WHERE in_scope = FALSE;

-- Trigger
CREATE TRIGGER trg_requirement_scopes_updated_at
    BEFORE UPDATE ON requirement_scopes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- If no row exists for a requirement, it's **in-scope by default**. This table only stores explicit scoping decisions.
- `justification` is effectively required when `in_scope = FALSE` (enforced at API layer, not DB — allows flexibility for drafts).
- `scoped_by` tracks who made the scoping decision for audit purposes.
- Spec §3.1.3: "Framework scoping — mark requirements in-scope or out-of-scope per framework with justification tracking."

---

### controls

The org's control library. Each control is a safeguard the org implements to meet compliance requirements. Controls are org-scoped — each tenant has their own library (seeded from pre-built templates).

```sql
CREATE TABLE controls (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    identifier              VARCHAR(50) NOT NULL,          -- 'CTRL-AC-001', org-unique
    title                   VARCHAR(500) NOT NULL,
    description             TEXT NOT NULL,
    implementation_guidance TEXT,                          -- how to implement this control
    category                control_category NOT NULL,
    status                  control_status NOT NULL DEFAULT 'draft',
    owner_id                UUID REFERENCES users(id) ON DELETE SET NULL,
    secondary_owner_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    evidence_requirements   TEXT,                          -- what evidence is needed
    test_criteria           TEXT,                          -- how to verify this control works
    is_custom               BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE if org-created, FALSE if from library
    source_template_id      VARCHAR(100),                  -- reference to seed template (e.g., 'TPL-AC-001')
    metadata                JSONB NOT NULL DEFAULT '{}',   -- extensible custom fields
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique identifier per org
    CONSTRAINT uq_control_identifier UNIQUE (org_id, identifier)
);

-- Indexes
CREATE INDEX idx_controls_org_id ON controls (org_id);
CREATE INDEX idx_controls_status ON controls (org_id, status);
CREATE INDEX idx_controls_category ON controls (org_id, category);
CREATE INDEX idx_controls_owner ON controls (owner_id);
CREATE INDEX idx_controls_secondary_owner ON controls (secondary_owner_id);
CREATE INDEX idx_controls_source_template ON controls (source_template_id)
    WHERE source_template_id IS NOT NULL;

-- Full-text search on title + description (for control library browser)
CREATE INDEX idx_controls_search ON controls
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Trigger
CREATE TRIGGER trg_controls_updated_at
    BEFORE UPDATE ON controls
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- Controls are per-org (`org_id`). Pre-built library controls are seeded into each org on framework activation.
- `source_template_id` tracks which seed template a control was cloned from (enables future "library updates" — detect when templates change and suggest updates to org controls).
- `is_custom` distinguishes org-created controls from library-seeded ones.
- `metadata` JSONB supports "Custom fields and formulas on controls" (spec §3.2.2).
- Full-text search index enables the control library browser.
- `owner_id` and `secondary_owner_id` per spec §3.2.3 (Control Ownership).
- `evidence_requirements` and `test_criteria` are text fields — structured evidence/test config comes in Sprints 3-4.

---

### control_mappings

The cross-framework mapping table. Links org controls to framework requirements. This is where "one control satisfies multiple frameworks" lives (spec §3.1.3).

```sql
CREATE TABLE control_mappings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    control_id      UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    requirement_id  UUID NOT NULL REFERENCES requirements(id) ON DELETE CASCADE,
    strength        VARCHAR(20) NOT NULL DEFAULT 'primary'
                    CHECK (strength IN ('primary', 'supporting', 'partial')),
    notes           TEXT,                                  -- mapping justification
    mapped_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One mapping per control-requirement pair per org
    CONSTRAINT uq_control_mapping UNIQUE (org_id, control_id, requirement_id)
);

-- Indexes
CREATE INDEX idx_control_mappings_org ON control_mappings (org_id);
CREATE INDEX idx_control_mappings_control ON control_mappings (control_id);
CREATE INDEX idx_control_mappings_requirement ON control_mappings (requirement_id);
CREATE INDEX idx_control_mappings_strength ON control_mappings (org_id, strength);

-- Composite index for cross-framework matrix query
CREATE INDEX idx_control_mappings_cross_fw ON control_mappings (org_id, requirement_id, control_id);
```

**Design notes:**
- `strength` captures how strongly a control addresses a requirement:
  - `primary` — the control directly and fully addresses the requirement
  - `supporting` — the control contributes but doesn't fully satisfy
  - `partial` — the control partially addresses (gap may remain)
- `mapped_by` tracks who created the mapping for audit.
- No `updated_at` — mappings are created or deleted, not updated. Edit the `notes` by deleting and recreating.
- The composite index on `(org_id, requirement_id, control_id)` powers the cross-framework mapping matrix (spec §3.1.3).
- `org_id` is denormalized (derivable from `control_id`) for efficient org-scoped queries and row-level isolation.

---

## Migration Order

Migrations should be numbered continuing from Sprint 1:

6. `006_sprint2_enums.sql` — New enum types + audit_action extensions
7. `007_frameworks.sql` — Frameworks table + indexes + trigger
8. `008_framework_versions.sql` — Framework versions table + indexes + trigger
9. `009_requirements.sql` — Requirements table + indexes + trigger
10. `010_org_frameworks.sql` — Org frameworks table + indexes + trigger
11. `011_controls.sql` — Controls table + indexes + trigger
12. `012_control_mappings.sql` — Control mappings table + indexes
13. `013_requirement_scopes.sql` — Requirement scopes table + indexes + trigger

---

## Seed Data

### Frameworks

```sql
INSERT INTO frameworks (id, identifier, name, description, category, website_url) VALUES
    ('f0000000-0000-0000-0000-000000000001', 'soc2',     'SOC 2',           'Service Organization Control 2 — Trust Services Criteria for security, availability, processing integrity, confidentiality, and privacy.', 'security_privacy', 'https://www.aicpa.org/soc2'),
    ('f0000000-0000-0000-0000-000000000002', 'iso27001', 'ISO 27001',       'International standard for information security management systems (ISMS).', 'security_privacy', 'https://www.iso.org/standard/27001'),
    ('f0000000-0000-0000-0000-000000000003', 'pci_dss',  'PCI DSS',         'Payment Card Industry Data Security Standard — requirements for protecting cardholder data.', 'payment', 'https://www.pcisecuritystandards.org/'),
    ('f0000000-0000-0000-0000-000000000004', 'gdpr',     'GDPR',            'General Data Protection Regulation — EU data privacy and protection law.', 'data_privacy', 'https://gdpr.eu/'),
    ('f0000000-0000-0000-0000-000000000005', 'ccpa',     'CCPA/CPRA',       'California Consumer Privacy Act / California Privacy Rights Act.', 'data_privacy', 'https://oag.ca.gov/privacy/ccpa');
```

### Framework Versions

```sql
INSERT INTO framework_versions (id, framework_id, version, display_name, status, effective_date, total_requirements) VALUES
    ('v0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000001', '2024',   'SOC 2 (2024 TSC)',      'active', '2024-01-01', 64),
    ('v0000000-0000-0000-0000-000000000002', 'f0000000-0000-0000-0000-000000000002', '2022',   'ISO 27001:2022',        'active', '2022-10-25', 93),
    ('v0000000-0000-0000-0000-000000000003', 'f0000000-0000-0000-0000-000000000003', '4.0.1',  'PCI DSS v4.0.1',       'active', '2024-06-11', 280),
    ('v0000000-0000-0000-0000-000000000004', 'f0000000-0000-0000-0000-000000000004', '2016',   'GDPR (2016/679)',       'active', '2018-05-25', 99),
    ('v0000000-0000-0000-0000-000000000005', 'f0000000-0000-0000-0000-000000000005', '2023',   'CCPA/CPRA (2023)',      'active', '2023-01-01', 42);
```

### Requirements (representative sample — DBE will create full seed)

The DBE should create complete requirement hierarchies for each framework. Here's the structure for PCI DSS as a representative example:

```sql
-- PCI DSS v4.0.1 — Top-level requirements (12 families)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000003', NULL, '1', 'Install and Maintain Network Security Controls', 1, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000003', NULL, '2', 'Apply Secure Configurations to All System Components', 2, 0, FALSE),
    -- ... through requirement 12

-- PCI DSS v4.0.1 — Sub-requirements (example for Req 6)
    ('r0300000-0000-0000-0000-000000000060', 'v0000000-0000-0000-0000-000000000003', NULL, '6', 'Develop and Maintain Secure Systems and Software', 6, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000061', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000060', '6.4', 'Public-Facing Web Applications are Protected Against Attacks', 4, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000062', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000061', '6.4.3', 'All payment page scripts that are loaded and executed in the consumer''s browser are managed', 3, 2, TRUE);

-- SOC 2 — Trust Services Criteria (example)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC1', 'Control Environment', 1, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000001', 'CC1.1', 'The entity demonstrates a commitment to integrity and ethical values', 1, 1, TRUE),
    -- ... etc.

-- ISO 27001:2022 — Annex A controls (example)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0200000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000002', NULL, 'A.5', 'Organizational Controls', 1, 0, FALSE),
    ('r0200000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.1', 'Policies for information security', 1, 1, TRUE),
    -- ... etc.
```

### Control Library Seed (representative — DBE will expand to 300+)

Controls are seeded per-org. The seed should create controls for the demo org (`a0000000-0000-0000-0000-000000000001`) mapped to requirements across all 5 frameworks.

**Control naming convention:** `CTRL-{category_prefix}-{number}`
- `CTRL-AC-xxx` — Access Control
- `CTRL-CM-xxx` — Configuration Management
- `CTRL-DP-xxx` — Data Protection
- `CTRL-IR-xxx` — Incident Response
- `CTRL-LM-xxx` — Logging & Monitoring
- `CTRL-NW-xxx` — Network Security
- `CTRL-PM-xxx` — Policy Management
- `CTRL-PE-xxx` — Physical & Environmental
- `CTRL-RA-xxx` — Risk Assessment
- `CTRL-SA-xxx` — Security Awareness
- `CTRL-SD-xxx` — Secure Development
- `CTRL-VM-xxx` — Vulnerability Management
- `CTRL-VP-xxx` — Vendor & Privacy

**Target: 300+ controls covering all 5 frameworks.**

Example seed structure:

```sql
-- Access Control category
INSERT INTO controls (id, org_id, identifier, title, description, category, status, is_custom, source_template_id) VALUES
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-001',
     'Multi-Factor Authentication',
     'Enforce MFA for all user accounts accessing production systems and sensitive data.',
     'technical', 'active', FALSE, 'TPL-AC-001'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-002',
     'Role-Based Access Control',
     'Implement RBAC with least-privilege access. Users receive only permissions required for their role.',
     'technical', 'active', FALSE, 'TPL-AC-002'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-003',
     'Quarterly Access Reviews',
     'Conduct quarterly reviews of user access across all critical systems. Remove stale/excessive permissions.',
     'administrative', 'active', FALSE, 'TPL-AC-003');
    -- ... 300+ total

-- Then map controls to requirements:
INSERT INTO control_mappings (id, org_id, control_id, requirement_id, strength) VALUES
    -- CTRL-AC-001 (MFA) maps to:
    -- PCI DSS 8.3.x (MFA requirements)
    -- SOC 2 CC6.1 (Logical access security)
    -- ISO 27001 A.8.5 (Secure authentication)
    -- GDPR Art. 32 (Security of processing)
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', '<ctrl_ac_001_id>', '<pci_8.3_id>', 'primary'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', '<ctrl_ac_001_id>', '<soc2_cc6.1_id>', 'primary'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', '<ctrl_ac_001_id>', '<iso_a8.5_id>', 'primary'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', '<ctrl_ac_001_id>', '<gdpr_art32_id>', 'supporting');
    -- ... etc.
```

### Demo Org Framework Activation

```sql
-- Activate all 5 frameworks for the demo org
INSERT INTO org_frameworks (id, org_id, framework_id, active_version_id, status, target_date) VALUES
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000001', 'active', '2026-06-30'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000002', 'active', '2026-09-30'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000003', 'v0000000-0000-0000-0000-000000000003', 'active', '2026-12-31'),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000004', 'v0000000-0000-0000-0000-000000000004', 'active', NULL),
    (gen_random_uuid(), 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000005', 'v0000000-0000-0000-0000-000000000005', 'active', NULL);
```

---

## Query Patterns

### Cross-Framework Mapping Matrix

The signature query — "Show me which controls satisfy which requirements across all frameworks":

```sql
-- For a given org, get the cross-framework mapping matrix
SELECT
    c.identifier AS control_id,
    c.title AS control_title,
    c.category,
    f.identifier AS framework,
    fv.version AS framework_version,
    r.identifier AS requirement_id,
    r.title AS requirement_title,
    cm.strength
FROM control_mappings cm
JOIN controls c ON c.id = cm.control_id
JOIN requirements r ON r.id = cm.requirement_id
JOIN framework_versions fv ON fv.id = r.framework_version_id
JOIN frameworks f ON f.id = fv.framework_id
WHERE cm.org_id = $1
ORDER BY c.identifier, f.identifier, r.identifier;
```

### Coverage Gap Analysis

"Which requirements in framework X have no controls mapped?"

```sql
-- Requirements with no control mappings for a given org + framework version
SELECT r.identifier, r.title, r.depth
FROM requirements r
LEFT JOIN control_mappings cm ON cm.requirement_id = r.id AND cm.org_id = $1
LEFT JOIN requirement_scopes rs ON rs.requirement_id = r.id AND rs.org_id = $1
WHERE r.framework_version_id = $2
  AND r.is_assessable = TRUE
  AND cm.id IS NULL
  AND (rs.id IS NULL OR rs.in_scope = TRUE)  -- exclude out-of-scope
ORDER BY r.section_order;
```

### Requirement Tree (hierarchical)

```sql
-- Recursive CTE to get full requirement tree for a framework version
WITH RECURSIVE req_tree AS (
    SELECT id, parent_id, identifier, title, depth, section_order, is_assessable
    FROM requirements
    WHERE framework_version_id = $1 AND parent_id IS NULL
    UNION ALL
    SELECT r.id, r.parent_id, r.identifier, r.title, r.depth, r.section_order, r.is_assessable
    FROM requirements r
    JOIN req_tree rt ON r.parent_id = rt.id
)
SELECT * FROM req_tree
ORDER BY depth, section_order;
```

---

## Future Considerations

- **Row-Level Security (RLS):** Same as Sprint 1 — org isolation enforced at app layer for now. RLS policies to be added in hardening sprint.
- **Framework version migration workflows:** Spec §3.1.2 calls for guided migration between versions with gap analysis. Schema supports this (multiple versions can be active). Migration workflow is future sprint work.
- **Control templates table:** For scale, a system-level `control_templates` table would avoid duplicating seed data per org. Current approach (seed per-org) is simpler for Sprint 2 and works for MVP.
- **Meilisearch integration:** The full-text GIN index on controls is a stopgap. Sprint 2 frontend includes a "control library browser" which may benefit from Meilisearch for faceted search. Can be added without schema changes.
- **Conflict detection:** Spec §3.1.3 mentions "Conflict detection when framework requirements are contradictory." This is an application-layer concern, not schema — but the mapping data structure supports it.
