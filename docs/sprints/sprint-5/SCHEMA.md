# Sprint 5 — Database Schema: Policy Management

## Overview

Sprint 5 introduces the policy management layer — the governance backbone that links written policies to the controls they govern, tracks policy versions with rich text content, and enforces sign-off workflows for policy approvals. This implements spec §6.1 (Policy Management) and §5.1 (AI Policy Agent foundations).

**Key design decisions:**

- **Policies as org-scoped governance documents**: Each policy belongs to one organization, has a lifecycle (Draft → In Review → Approved → Published → Archived), and is owned by a user. Policies can be mapped to controls for gap detection.
- **Version-controlled content**: Policy content is stored in `policy_versions`. Each edit creates a new version. The `policies` table points to the current published version via `current_version_id`. Full version history is preserved for audit and diff views.
- **Rich text stored as HTML**: Content is stored as sanitized HTML (from the rich text editor). The API layer sanitizes on write (strip scripts, event handlers, iframes) and the frontend renders safely. A `content_format` field allows future Markdown support.
- **Sign-off as approval workflow**: Sign-offs are requested per policy version. Multiple signers can be required. A policy version is "fully approved" when all requested sign-offs have `approved` status. Rejections include mandatory comments.
- **Policy-to-control mapping**: `policy_controls` is a many-to-many junction table. One policy can cover many controls; one control can be governed by many policies. Gap detection queries find controls without policy coverage.
- **Template system**: Templates are regular policies with `is_template = TRUE` and an optional `template_framework_id` linking them to a framework. Cloning a template creates a new non-template policy, copying the latest version content.
- **Review reminders built-in**: `review_frequency_days` + `next_review_at` enable automated annual/quarterly review reminders per spec §6.1 ("Automated annual review reminders per policy").
- **Evidence linking ready**: Sprint 3 already defined `evidence_link_target_type = 'policy'`. This sprint adds the FK from `evidence_links` to `policies`, completing the connection.

---

## Entity Relationship Diagram

```
                    ┌─────────────────────────────────────────────────┐
                    │   POLICY DOMAIN (org-scoped)                    │
                    │                                                 │
 organizations ─┬──▶ policies                                        │
                │   │   (policy definitions with ownership)          │
                │   │   ∞──1 users (owner_id)                        │
                │   │   ∞──1 policy_versions (current_version_id)    │
                │   │   ∞──1 frameworks (template_framework_id)      │
                │   │                                                 │
                │   │   1──∞ policy_versions                          │
                │   │         (rich text content, change tracking)    │
                │   │         ∞──1 users (created_by)                 │
                │   │                                                 │
                │   │   1──∞ policy_signoffs                          │
                │   │         (approval workflow tracking)            │
                │   │         ∞──1 policy_versions                    │
                │   │         ∞──1 users (signer_id)                  │
                │   │         ∞──1 users (requested_by)               │
                │   │                                                 │
                │   │   1──∞ policy_controls ──▶ controls             │
                │   │         (many-to-many: policy ↔ control)        │
                │   │                                                 │
                │   │   1──∞ evidence_links (target_type = 'policy')  │
                │   │         (via new FK on evidence_links)          │
                │   │                                                 │
                └──▶ audit_log (extended)                             │
                    └─────────────────────────────────────────────────┘
```

**Relationships:**
```
policies              ∞──1  organizations          (org_id)
policies              ∞──1  users                  (owner_id)
policies              ∞──1  users                  (secondary_owner_id)
policies              ∞──1  policy_versions        (current_version_id, deferred)
policies              ∞──1  frameworks             (template_framework_id, optional)
policy_versions       ∞──1  policies               (policy_id)
policy_versions       ∞──1  users                  (created_by)
policy_signoffs       ∞──1  organizations          (org_id)
policy_signoffs       ∞──1  policies               (policy_id)
policy_signoffs       ∞──1  policy_versions        (policy_version_id)
policy_signoffs       ∞──1  users                  (signer_id)
policy_signoffs       ∞──1  users                  (requested_by)
policy_controls       ∞──1  organizations          (org_id)
policy_controls       ∞──1  policies               (policy_id)
policy_controls       ∞──1  controls               (control_id)
policy_controls       ∞──1  users                  (linked_by)
evidence_links        ∞──1  policies               (policy_id, new FK)
```

---

## New Enums

```sql
-- Categories of policies (from spec §6.1 + industry standard taxonomy)
CREATE TYPE policy_category AS ENUM (
    'information_security',     -- Overall information security policy (ISO 27001 mandatory)
    'acceptable_use',           -- Acceptable use of IT resources
    'access_control',           -- Access management, authentication, authorization
    'data_classification',      -- Data classification and handling procedures
    'data_privacy',             -- Privacy policies (GDPR, CCPA specific)
    'data_retention',           -- Data retention and disposal schedules
    'incident_response',        -- Security incident detection, response, reporting
    'business_continuity',      -- BC/DR planning and testing
    'change_management',        -- Change control procedures
    'vulnerability_management', -- Vulnerability scanning, patching, remediation
    'vendor_management',        -- Third-party risk management
    'physical_security',        -- Physical access, environmental controls
    'encryption',               -- Cryptographic controls and key management
    'network_security',         -- Network segmentation, firewall, monitoring
    'secure_development',       -- SDLC, code review, deployment security
    'human_resources',          -- Background checks, termination, training
    'compliance',               -- Regulatory compliance oversight
    'risk_management',          -- Risk assessment and treatment
    'asset_management',         -- Asset inventory, ownership, lifecycle
    'logging_monitoring',       -- Audit logging, SIEM, alerting
    'custom'                    -- Organization-defined categories
);

-- Policy lifecycle (from spec §6.1)
CREATE TYPE policy_status AS ENUM (
    'draft',                    -- Being written/edited, not visible to general users
    'in_review',                -- Submitted for review, pending sign-offs
    'approved',                 -- All required sign-offs obtained
    'published',                -- Active and visible to all users, in effect
    'archived'                  -- No longer in effect, kept for historical reference
);

-- Sign-off decision status
CREATE TYPE signoff_status AS ENUM (
    'pending',                  -- Awaiting signer's decision
    'approved',                 -- Signer approved
    'rejected',                 -- Signer rejected (must include comments)
    'withdrawn'                 -- Sign-off request was withdrawn/cancelled
);

-- Policy content format
CREATE TYPE policy_content_format AS ENUM (
    'html',                     -- Rich text as sanitized HTML (default from editor)
    'markdown',                 -- Markdown (for code-driven policies)
    'plain_text'                -- Plain text (fallback)
);
```

### Extend Existing Enums

```sql
-- Add Sprint 5 actions to audit_action enum
ALTER TYPE audit_action ADD VALUE 'policy.created';
ALTER TYPE audit_action ADD VALUE 'policy.updated';
ALTER TYPE audit_action ADD VALUE 'policy.status_changed';
ALTER TYPE audit_action ADD VALUE 'policy.archived';
ALTER TYPE audit_action ADD VALUE 'policy.deleted';
ALTER TYPE audit_action ADD VALUE 'policy.owner_changed';
ALTER TYPE audit_action ADD VALUE 'policy.cloned_from_template';
ALTER TYPE audit_action ADD VALUE 'policy_version.created';
ALTER TYPE audit_action ADD VALUE 'policy_version.published';
ALTER TYPE audit_action ADD VALUE 'policy_signoff.requested';
ALTER TYPE audit_action ADD VALUE 'policy_signoff.approved';
ALTER TYPE audit_action ADD VALUE 'policy_signoff.rejected';
ALTER TYPE audit_action ADD VALUE 'policy_signoff.withdrawn';
ALTER TYPE audit_action ADD VALUE 'policy_control.linked';
ALTER TYPE audit_action ADD VALUE 'policy_control.unlinked';
```

---

## Tables

### policies

The core policy definition table. Each row is a policy document with metadata, ownership, lifecycle status, and review scheduling. Templates are also stored here with `is_template = TRUE`.

```sql
CREATE TABLE policies (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    identifier              VARCHAR(50) NOT NULL,            -- human-readable: 'POL-IS-001'
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,                            -- summary/abstract of the policy

    -- Classification
    category                policy_category NOT NULL,
    status                  policy_status NOT NULL DEFAULT 'draft',

    -- Current published version (denormalized for fast lookup)
    -- FK deferred: added after policy_versions table is created
    current_version_id      UUID,

    -- Ownership (spec §3.2.3 pattern — primary + secondary owner)
    owner_id                UUID REFERENCES users(id) ON DELETE SET NULL,
    secondary_owner_id      UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Review scheduling (spec §6.1: "Automated annual review reminders")
    review_frequency_days   INT,                             -- how often policy must be reviewed (e.g., 365 = annual)
    next_review_at          DATE,                            -- when the next review is due
    last_reviewed_at        DATE,                            -- when the policy was last reviewed

    -- Template support
    is_template             BOOLEAN NOT NULL DEFAULT FALSE,  -- TRUE = template policy, not active governance doc
    template_framework_id   UUID REFERENCES frameworks(id) ON DELETE SET NULL,  -- which framework this template is for
    cloned_from_policy_id   UUID REFERENCES policies(id) ON DELETE SET NULL,    -- if cloned from a template, link to source

    -- Approval metadata (denormalized for quick queries)
    approved_at             TIMESTAMPTZ,                     -- when all sign-offs completed
    approved_version        INT,                             -- version number that was approved
    published_at            TIMESTAMPTZ,                     -- when the policy was published

    -- Tags and metadata
    tags                    TEXT[] DEFAULT '{}',              -- free-form tags: ['pci', 'annual', 'mandatory']
    metadata                JSONB NOT NULL DEFAULT '{}',     -- extensible: distribution tracking, custom fields

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_policy_identifier UNIQUE (org_id, identifier),
    CONSTRAINT chk_review_frequency CHECK (review_frequency_days IS NULL OR review_frequency_days > 0),
    CONSTRAINT chk_template_framework CHECK (
        (is_template = FALSE AND template_framework_id IS NULL) OR
        (is_template = TRUE) -- templates may or may not have framework
    )
);

-- Indexes
CREATE INDEX idx_policies_org ON policies (org_id);
CREATE INDEX idx_policies_org_status ON policies (org_id, status);
CREATE INDEX idx_policies_org_category ON policies (org_id, category);
CREATE INDEX idx_policies_owner ON policies (owner_id)
    WHERE owner_id IS NOT NULL;
CREATE INDEX idx_policies_secondary_owner ON policies (secondary_owner_id)
    WHERE secondary_owner_id IS NOT NULL;
CREATE INDEX idx_policies_templates ON policies (org_id, is_template)
    WHERE is_template = TRUE;
CREATE INDEX idx_policies_template_framework ON policies (template_framework_id)
    WHERE template_framework_id IS NOT NULL AND is_template = TRUE;
CREATE INDEX idx_policies_review_due ON policies (org_id, next_review_at)
    WHERE next_review_at IS NOT NULL AND status IN ('published', 'approved') AND is_template = FALSE;
CREATE INDEX idx_policies_identifier ON policies (org_id, identifier);
CREATE INDEX idx_policies_tags ON policies USING gin (tags);
CREATE INDEX idx_policies_cloned_from ON policies (cloned_from_policy_id)
    WHERE cloned_from_policy_id IS NOT NULL;

-- Full-text search on title + description
CREATE INDEX idx_policies_search ON policies
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Trigger
CREATE TRIGGER trg_policies_updated_at
    BEFORE UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `identifier` follows the pattern `POL-{category_prefix}-{number}` — org-scoped unique. E.g., `POL-IS-001` for Information Security Policy.
- `current_version_id` is a denormalized FK to `policy_versions`. Updated when a new version is published. Deferred FK constraint added after `policy_versions` table is created (see below).
- **Template system**: `is_template = TRUE` marks policies as templates. Templates can optionally link to a `framework_id` to indicate which framework they're relevant for. `cloned_from_policy_id` traces the lineage when a template is cloned into a real policy.
- **Review scheduling**: `review_frequency_days` (e.g., 365 for annual, 90 for quarterly) + `next_review_at` computed as `last_reviewed_at + review_frequency_days`. A periodic job or heartbeat check surfaces policies with `next_review_at < NOW()` as overdue.
- `approved_at` and `published_at` are denormalized timestamps for quick dashboard queries — avoids joining through sign-offs to determine approval state.
- **Status transitions**: `draft` → `in_review` (submit for sign-off) → `approved` (all sign-offs complete) → `published` (made active) → `archived` (end of life). The `in_review` → `approved` transition is automatic when all required sign-offs are `approved`.
- `chk_template_framework` ensures non-template policies don't get tagged with a framework (that association goes through `policy_controls` → `control_mappings` → `requirements`).

---

### policy_versions

Stores the actual content of each policy version. Every edit creates a new version. Versions are immutable once created — no editing a published version, only creating new ones.

```sql
CREATE TABLE policy_versions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id               UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,

    -- Version tracking
    version_number          INT NOT NULL,                    -- sequential: 1, 2, 3...
    is_current              BOOLEAN NOT NULL DEFAULT TRUE,   -- TRUE = latest version

    -- Content
    content                 TEXT NOT NULL,                    -- policy body (HTML, Markdown, or plain text)
    content_format          policy_content_format NOT NULL DEFAULT 'html',
    content_summary         TEXT,                            -- brief summary of this version's content

    -- Change tracking
    change_summary          TEXT,                            -- what changed in this version: "Updated section 4.2 — added MFA requirements"
    change_type             VARCHAR(50) NOT NULL DEFAULT 'minor'
                            CHECK (change_type IN ('major', 'minor', 'patch', 'initial')),

    -- Content metrics (denormalized for UI)
    word_count              INT,                             -- word count of content
    character_count         INT,                             -- character count of content

    -- Authorship
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- No updated_at: versions are immutable once created
);

-- Indexes
CREATE INDEX idx_policy_versions_org ON policy_versions (org_id);
CREATE INDEX idx_policy_versions_policy ON policy_versions (policy_id);
CREATE INDEX idx_policy_versions_current ON policy_versions (policy_id, is_current)
    WHERE is_current = TRUE;
CREATE INDEX idx_policy_versions_policy_number ON policy_versions (policy_id, version_number DESC);
CREATE INDEX idx_policy_versions_created_by ON policy_versions (created_by)
    WHERE created_by IS NOT NULL;
CREATE INDEX idx_policy_versions_created ON policy_versions (policy_id, created_at DESC);

-- Each policy can have only one version with a given number
CREATE UNIQUE INDEX uq_policy_version_number ON policy_versions (policy_id, version_number);

-- Each policy can have only one current version
CREATE UNIQUE INDEX uq_policy_version_current ON policy_versions (policy_id)
    WHERE is_current = TRUE;
```

**Design notes:**
- **Immutable**: Policy versions are write-once — no `updated_at`. Once content is published, it becomes a historical record. To change content, create a new version.
- `is_current` is enforced via a unique partial index — only one version per policy can be current. When a new version is created, the previous current version's `is_current` is set to `FALSE`.
- `content` stores the full policy text. For HTML content from the rich text editor, the API layer sanitizes on write: strip `<script>`, `<iframe>`, `on*` event handlers, `javascript:` URLs, and other XSS vectors. Use a library like `bluemonday` (Go) for sanitization.
- `change_type` categorizes the edit:
  - `initial` — first version
  - `major` — significant content changes (new sections, policy rewrite)
  - `minor` — moderate changes (updated procedures, clarified language)
  - `patch` — typo fixes, formatting
- `word_count` and `character_count` are computed at creation time for display ("This policy is approximately 3,200 words").
- `content_summary` is an optional human-written or AI-generated abstract of the version content.
- `change_summary` is critical for version comparison views: "What changed between v2 and v3?"

---

### policy_signoffs

Tracks sign-off/approval requests for policy versions. Each row represents one signer's status for one policy version. A policy version may require multiple signers.

```sql
CREATE TABLE policy_signoffs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id               UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    policy_version_id       UUID NOT NULL REFERENCES policy_versions(id) ON DELETE CASCADE,

    -- Who needs to sign
    signer_id               UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    signer_role             grc_role,                        -- snapshot of signer's role at request time

    -- Request details
    requested_by            UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    requested_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date                DATE,                            -- optional deadline for signing

    -- Sign-off decision
    status                  signoff_status NOT NULL DEFAULT 'pending',
    decided_at              TIMESTAMPTZ,                     -- when signer approved/rejected
    comments                TEXT,                            -- mandatory on rejection, optional on approval

    -- Notification tracking
    reminder_sent_at        TIMESTAMPTZ,                     -- last reminder sent
    reminder_count          INT NOT NULL DEFAULT 0,          -- how many reminders have been sent

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each signer signs each version only once
    CONSTRAINT uq_policy_signoff UNIQUE (policy_version_id, signer_id),
    -- Rejection requires comments
    CONSTRAINT chk_rejection_comments CHECK (
        status != 'rejected' OR comments IS NOT NULL
    ),
    -- Decided_at must be set when status is not pending
    CONSTRAINT chk_decided_at CHECK (
        (status = 'pending' AND decided_at IS NULL) OR
        (status != 'pending' AND decided_at IS NOT NULL) OR
        (status = 'withdrawn')  -- withdrawn can have decided_at or not
    )
);

-- Indexes
CREATE INDEX idx_policy_signoffs_org ON policy_signoffs (org_id);
CREATE INDEX idx_policy_signoffs_policy ON policy_signoffs (policy_id);
CREATE INDEX idx_policy_signoffs_version ON policy_signoffs (policy_version_id);
CREATE INDEX idx_policy_signoffs_signer ON policy_signoffs (signer_id);
CREATE INDEX idx_policy_signoffs_signer_pending ON policy_signoffs (signer_id, status)
    WHERE status = 'pending';
CREATE INDEX idx_policy_signoffs_requested_by ON policy_signoffs (requested_by);
CREATE INDEX idx_policy_signoffs_status ON policy_signoffs (org_id, status);
CREATE INDEX idx_policy_signoffs_pending ON policy_signoffs (org_id, policy_id, status)
    WHERE status = 'pending';
CREATE INDEX idx_policy_signoffs_due ON policy_signoffs (org_id, due_date)
    WHERE status = 'pending' AND due_date IS NOT NULL;

-- Trigger
CREATE TRIGGER trg_policy_signoffs_updated_at
    BEFORE UPDATE ON policy_signoffs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- **Signer uniqueness per version**: `(policy_version_id, signer_id)` ensures each signer only has one sign-off request per version. To re-request after rejection, withdraw the old one and create a new request.
- `signer_role` is snapshotted at request time. If the signer's role changes later, the historical record reflects their role when they signed (audit trail integrity).
- `signer_id` and `requested_by` use `ON DELETE RESTRICT` — cannot delete users who are part of an active sign-off workflow. Ensures audit trail is never broken.
- **Rejection requires comments** (`chk_rejection_comments`): When a signer rejects, they must explain why. This is enforced at the database level.
- `reminder_sent_at` + `reminder_count` support automated reminder workflows: "You have a pending sign-off request that was sent 5 days ago." The worker or API can send periodic reminders and track state.
- `due_date` enables SLA-like tracking: "This sign-off must be completed by March 1st."
- **Approval detection**: A policy version is "fully approved" when `COUNT(*) WHERE status = 'pending' AND policy_version_id = X` returns 0 AND at least one sign-off has `status = 'approved'`. The API layer handles the state transition on `policies.status` when all sign-offs are complete.

---

### policy_controls

Many-to-many junction table linking policies to the controls they govern. This is the foundation for policy gap detection (spec §6.1: "Policy-to-control mapping with gap detection").

```sql
CREATE TABLE policy_controls (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id               UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Link metadata
    notes                   TEXT,                            -- why this policy governs this control
    coverage                VARCHAR(20) NOT NULL DEFAULT 'full'
                            CHECK (coverage IN ('full', 'partial')),

    -- Who created the link
    linked_by               UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each policy links to each control only once
    CONSTRAINT uq_policy_control UNIQUE (org_id, policy_id, control_id)
);

-- Indexes
CREATE INDEX idx_policy_controls_org ON policy_controls (org_id);
CREATE INDEX idx_policy_controls_policy ON policy_controls (policy_id);
CREATE INDEX idx_policy_controls_control ON policy_controls (control_id);
CREATE INDEX idx_policy_controls_linked_by ON policy_controls (linked_by)
    WHERE linked_by IS NOT NULL;
```

**Design notes:**
- No `updated_at` — like `control_mappings` (Sprint 2), links are created or deleted, not updated. To change notes/coverage, delete and recreate.
- `coverage` indicates whether the policy fully addresses the control's governance needs or only partially:
  - `full` — the policy completely governs this control
  - `partial` — the policy addresses some aspects; additional policy coverage may be needed
- `org_id` is denormalized (derivable from `policy_id`) for efficient org-scoped queries and consistent multi-tenancy enforcement.
- **Gap detection query**: Join `controls` LEFT JOIN `policy_controls` WHERE `policy_controls.id IS NULL` to find ungoverned controls.

---

## Deferred Foreign Keys

After all tables are created, add the cross-reference FKs:

```sql
-- policies.current_version_id → policy_versions
ALTER TABLE policies
    ADD CONSTRAINT fk_policies_current_version
    FOREIGN KEY (current_version_id) REFERENCES policy_versions(id) ON DELETE SET NULL;

CREATE INDEX idx_policies_current_version ON policies (current_version_id)
    WHERE current_version_id IS NOT NULL;

-- evidence_links.policy_id → policies (completing Sprint 3's forward declaration)
-- First, add the policy_id column to evidence_links
ALTER TABLE evidence_links
    ADD COLUMN policy_id UUID REFERENCES policies(id) ON DELETE CASCADE;

CREATE INDEX idx_evidence_links_policy ON evidence_links (policy_id)
    WHERE policy_id IS NOT NULL;

-- Update the existing CHECK constraint on evidence_links to include policy FK
-- Drop the old constraint and recreate with policy_id included
ALTER TABLE evidence_links DROP CONSTRAINT chk_evidence_link_target;
ALTER TABLE evidence_links ADD CONSTRAINT chk_evidence_link_target CHECK (
    (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL AND policy_id IS NULL) OR
    (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL AND policy_id IS NULL) OR
    (target_type = 'policy' AND policy_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL)
);

-- Add uniqueness constraint for policy evidence links
ALTER TABLE evidence_links ADD CONSTRAINT uq_evidence_link_policy
    UNIQUE (org_id, artifact_id, policy_id) DEFERRABLE INITIALLY DEFERRED;
```

---

## Migration Order

Migrations continue from Sprint 4:

27. `027_sprint5_enums.sql` — New enum types (policy_category, policy_status, signoff_status, policy_content_format) + audit_action extensions
28. `028_policies.sql` — Policies table + indexes + trigger + constraints (current_version_id column present but FK deferred)
29. `029_policy_versions.sql` — Policy versions table + indexes + unique constraints
30. `030_policy_signoffs.sql` — Policy signoffs table + indexes + trigger + constraints
31. `031_policy_controls.sql` — Policy controls junction table + indexes + unique constraint
32. `032_sprint5_fk_cross_refs.sql` — Deferred foreign keys: policies.current_version_id → policy_versions, evidence_links.policy_id column + FK + updated CHECK constraint
33. `033_sprint5_seed_templates.sql` — Policy template seed data (templates per framework)
34. `034_sprint5_seed_demo.sql` — Demo org example policies with versions, sign-offs, and control mappings

---

## Seed Data

### Policy Templates (per framework)

Templates provide starter policies aligned to framework requirements. Organizations clone these to jump-start their policy library.

```sql
-- SOC 2 Policy Templates
INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('pt000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'TPL-IS-001', 'Information Security Policy',
     'Comprehensive information security policy establishing the organization''s commitment to protecting information assets. Covers scope, objectives, roles, and responsibilities.',
     'information_security', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'pci', 'template', 'mandatory']),

    ('pt000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'TPL-AC-001', 'Access Control Policy',
     'Defines requirements for user access management, authentication, authorization, and periodic access reviews.',
     'access_control', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'pci', 'template']),

    ('pt000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'TPL-IR-001', 'Incident Response Policy',
     'Establishes procedures for detecting, reporting, assessing, responding to, and learning from security incidents.',
     'incident_response', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'pci', 'template']),

    ('pt000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'TPL-CM-001', 'Change Management Policy',
     'Defines the process for requesting, reviewing, approving, implementing, and documenting changes to IT systems and infrastructure.',
     'change_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'template']),

    ('pt000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'TPL-BC-001', 'Business Continuity & Disaster Recovery Policy',
     'Establishes the framework for maintaining business operations during and recovering from disruptive events.',
     'business_continuity', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'template']);

-- PCI DSS Policy Templates
INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('pt000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000001',
     'TPL-NW-001', 'Network Security Policy',
     'Defines requirements for network segmentation, firewall configuration, and monitoring of network traffic to protect cardholder data environments.',
     'network_security', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000003', 365,
     ARRAY['pci', 'template', 'network']),

    ('pt000000-0000-0000-0000-000000000011', 'a0000000-0000-0000-0000-000000000001',
     'TPL-EN-001', 'Encryption & Key Management Policy',
     'Establishes standards for cryptographic protection of sensitive data at rest and in transit, and procedures for cryptographic key lifecycle management.',
     'encryption', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000003', 365,
     ARRAY['pci', 'template', 'encryption']),

    ('pt000000-0000-0000-0000-000000000012', 'a0000000-0000-0000-0000-000000000001',
     'TPL-VM-001', 'Vulnerability Management Policy',
     'Defines the process for identifying, assessing, prioritizing, and remediating vulnerabilities in systems and applications.',
     'vulnerability_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000003', 180,
     ARRAY['pci', 'template', 'vulnerability']);

-- ISO 27001 Policy Templates
INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('pt000000-0000-0000-0000-000000000020', 'a0000000-0000-0000-0000-000000000001',
     'TPL-AM-001', 'Asset Management Policy',
     'Defines requirements for identifying, classifying, and managing information assets throughout their lifecycle.',
     'asset_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000002', 365,
     ARRAY['iso27001', 'template', 'asset']),

    ('pt000000-0000-0000-0000-000000000021', 'a0000000-0000-0000-0000-000000000001',
     'TPL-RM-001', 'Risk Management Policy',
     'Establishes the organization''s approach to identifying, assessing, treating, and monitoring information security risks.',
     'risk_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000002', 365,
     ARRAY['iso27001', 'template', 'risk']);

-- GDPR Policy Templates
INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('pt000000-0000-0000-0000-000000000030', 'a0000000-0000-0000-0000-000000000001',
     'TPL-DP-001', 'Data Privacy Policy',
     'Establishes the organization''s approach to processing personal data in compliance with GDPR. Covers lawful basis, data subject rights, cross-border transfers, and DPIAs.',
     'data_privacy', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000004', 365,
     ARRAY['gdpr', 'template', 'privacy']),

    ('pt000000-0000-0000-0000-000000000031', 'a0000000-0000-0000-0000-000000000001',
     'TPL-DR-001', 'Data Retention & Disposal Policy',
     'Defines retention periods for different data categories and secure disposal procedures for data that has exceeded its retention period.',
     'data_retention', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000004', 365,
     ARRAY['gdpr', 'ccpa', 'template', 'retention']);

-- CCPA Policy Templates
INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('pt000000-0000-0000-0000-000000000040', 'a0000000-0000-0000-0000-000000000001',
     'TPL-CP-001', 'Consumer Privacy Rights Policy',
     'Defines procedures for handling consumer privacy rights requests under CCPA/CPRA: right to know, delete, opt-out of sale, and correct.',
     'data_privacy', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000005', 365,
     ARRAY['ccpa', 'template', 'privacy', 'consumer-rights']);
```

### Policy Template Version Content

```sql
-- Version content for Information Security Policy template
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('pv000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'pt000000-0000-0000-0000-000000000001', 1, TRUE,
     '<h1>Information Security Policy</h1>
<h2>1. Purpose</h2>
<p>This policy establishes the organization''s commitment to protecting information assets from threats, whether internal or external, deliberate or accidental. It provides the framework for setting objectives and establishing the overall direction and principles for information security.</p>

<h2>2. Scope</h2>
<p>This policy applies to all employees, contractors, consultants, temporaries, and other workers at [Organization Name], including all personnel affiliated with third parties who access organizational information systems.</p>

<h2>3. Policy Statement</h2>
<p>The organization shall:</p>
<ul>
<li>Protect information from unauthorized access, disclosure, modification, or destruction</li>
<li>Ensure business continuity and minimize business damage</li>
<li>Comply with all applicable laws, regulations, and contractual obligations</li>
<li>Provide security awareness training to all personnel</li>
<li>Report and investigate all suspected security incidents</li>
<li>Regularly review and improve the information security management system</li>
</ul>

<h2>4. Roles and Responsibilities</h2>
<h3>4.1 CISO / Security Leader</h3>
<p>Responsible for developing, implementing, and maintaining the information security program. Reports to executive management on the state of information security.</p>

<h3>4.2 IT Administrators</h3>
<p>Responsible for implementing technical security controls, monitoring systems, and responding to security alerts.</p>

<h3>4.3 All Employees</h3>
<p>Responsible for complying with security policies, reporting security incidents, and completing required security awareness training.</p>

<h2>5. Review</h2>
<p>This policy shall be reviewed at least annually or when significant changes occur to the organization, its business processes, or the threat landscape.</p>

<h2>6. Compliance</h2>
<p>Violation of this policy may result in disciplinary action up to and including termination of employment and/or legal action.</p>',
     'html',
     'Comprehensive information security policy template covering purpose, scope, policy statement, roles and responsibilities, review cadence, and compliance.',
     'Initial version',
     'initial',
     280,
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1));

-- Version content for Access Control Policy template
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('pv000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'pt000000-0000-0000-0000-000000000002', 1, TRUE,
     '<h1>Access Control Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines the requirements for managing user access to information systems, applications, and data to ensure that access is granted based on the principle of least privilege and business need-to-know.</p>

<h2>2. Scope</h2>
<p>This policy applies to all user accounts, service accounts, and access credentials used to access organizational systems and data.</p>

<h2>3. Access Control Requirements</h2>
<h3>3.1 Authentication</h3>
<ul>
<li>Multi-factor authentication (MFA) shall be required for all user accounts accessing production systems</li>
<li>Passwords shall meet minimum complexity requirements: 12+ characters, mixed case, numbers, and special characters</li>
<li>Service accounts shall use certificate-based or token-based authentication where possible</li>
</ul>

<h3>3.2 Authorization</h3>
<ul>
<li>Access shall be granted based on role-based access control (RBAC)</li>
<li>Users shall receive only the minimum permissions required to perform their job functions</li>
<li>Privileged access shall require additional approval and be time-limited where possible</li>
</ul>

<h3>3.3 Access Reviews</h3>
<ul>
<li>Access reviews shall be conducted quarterly for all critical systems</li>
<li>Reviews shall verify that access levels are appropriate for current job responsibilities</li>
<li>Stale accounts (no login for 90+ days) shall be disabled pending review</li>
</ul>

<h3>3.4 Account Lifecycle</h3>
<ul>
<li>Access shall be provisioned within 24 hours of approved request</li>
<li>Access shall be revoked within 4 hours of employment termination</li>
<li>Role changes shall trigger access review within 5 business days</li>
</ul>

<h2>4. Review</h2>
<p>This policy shall be reviewed annually or when significant changes occur to access control technology or organizational structure.</p>',
     'html',
     'Access control policy template covering authentication, authorization, access reviews, and account lifecycle management.',
     'Initial version',
     'initial',
     250,
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1));

-- Update current_version_id for templates
UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000001'
WHERE id = 'pt000000-0000-0000-0000-000000000001';

UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000002'
WHERE id = 'pt000000-0000-0000-0000-000000000002';
```

### Demo Organization Example Policies

```sql
-- Active policies for the demo org (cloned from templates)
INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    owner_id, is_template, cloned_from_policy_id,
    review_frequency_days, next_review_at, last_reviewed_at,
    approved_at, approved_version, published_at,
    tags
) VALUES
    ('pd000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'POL-IS-001', 'Acme Corp Information Security Policy',
     'Acme Corporation''s information security policy. Establishes the framework for protecting company and customer information assets.',
     'information_security', 'published',
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     FALSE, 'pt000000-0000-0000-0000-000000000001',
     365, '2027-01-15', '2026-01-15',
     '2026-01-15 10:00:00+00', 1, '2026-01-15 10:30:00+00',
     ARRAY['mandatory', 'annual', 'all-employees']),

    ('pd000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'POL-AC-001', 'Acme Corp Access Control Policy',
     'Defines access management requirements for all Acme Corp systems and data.',
     'access_control', 'published',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     FALSE, 'pt000000-0000-0000-0000-000000000002',
     365, '2027-02-01', '2026-02-01',
     '2026-02-01 14:00:00+00', 1, '2026-02-01 14:30:00+00',
     ARRAY['annual', 'access', 'production']),

    ('pd000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'POL-IR-001', 'Acme Corp Incident Response Plan',
     'Procedures for detecting, reporting, and responding to security incidents at Acme Corp.',
     'incident_response', 'in_review',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     FALSE, 'pt000000-0000-0000-0000-000000000003',
     365, NULL, NULL,
     NULL, NULL, NULL,
     ARRAY['annual', 'incident', 'security']);

-- Demo policy versions
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('pv000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001', 1, TRUE,
     '<h1>Acme Corporation — Information Security Policy</h1><p>Version 1.0 — Effective January 15, 2026</p><h2>1. Purpose</h2><p>This policy establishes Acme Corporation''s commitment to protecting information assets belonging to the company, its customers, and its partners.</p><h2>2. Scope</h2><p>This policy applies to all Acme Corp employees, contractors, and third-party service providers.</p><p>[Full content would follow...]</p>',
     'html', 'Acme Corp customized information security policy v1.0', 'Initial version — customized from template', 'initial',
     320,
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)),

    ('pv000000-0000-0000-0000-000000000011', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002', 1, TRUE,
     '<h1>Acme Corporation — Access Control Policy</h1><p>Version 1.0 — Effective February 1, 2026</p><h2>1. Purpose</h2><p>This policy defines requirements for managing user access to Acme Corp information systems.</p><p>[Full content would follow...]</p>',
     'html', 'Acme Corp access control policy v1.0', 'Initial version — customized from template', 'initial',
     280,
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)),

    ('pv000000-0000-0000-0000-000000000012', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000003', 1, TRUE,
     '<h1>Acme Corporation — Incident Response Plan</h1><p>DRAFT — Version 1.0</p><h2>1. Purpose</h2><p>This plan defines the procedures for Acme Corp''s response to information security incidents.</p><p>[Full content — currently in review...]</p>',
     'html', 'Acme Corp incident response plan draft v1.0', 'Initial draft', 'initial',
     150,
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1));

-- Update current_version_id for demo policies
UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000010'
WHERE id = 'pd000000-0000-0000-0000-000000000001';
UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000011'
WHERE id = 'pd000000-0000-0000-0000-000000000002';
UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000012'
WHERE id = 'pd000000-0000-0000-0000-000000000003';

-- Demo sign-offs (approved for published policies, pending for in_review)
INSERT INTO policy_signoffs (
    id, org_id, policy_id, policy_version_id,
    signer_id, signer_role, requested_by, requested_at,
    status, decided_at, comments
) VALUES
    -- Info Security Policy — approved by CISO
    ('ps000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001', 'pv000000-0000-0000-0000-000000000010',
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     'ciso',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     '2026-01-14 09:00:00+00',
     'approved', '2026-01-15 10:00:00+00', 'Reviewed and approved. Meets all framework requirements.'),

    -- Access Control Policy — approved by CISO
    ('ps000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002', 'pv000000-0000-0000-0000-000000000011',
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     'ciso',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-01-30 09:00:00+00',
     'approved', '2026-02-01 14:00:00+00', 'Approved — aligns with PCI DSS 7 and 8 requirements.'),

    -- Incident Response Plan — pending sign-off from CISO
    ('ps000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000003', 'pv000000-0000-0000-0000-000000000012',
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     'ciso',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18 09:00:00+00',
     'pending', NULL, NULL),

    -- Incident Response Plan — pending sign-off from Compliance Manager
    ('ps000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000003', 'pv000000-0000-0000-0000-000000000012',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     'compliance_manager',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18 09:00:00+00',
     'pending', NULL, NULL);

-- Demo policy-control mappings
INSERT INTO policy_controls (
    id, org_id, policy_id, control_id, coverage, notes, linked_by
) VALUES
    -- Info Security Policy → several controls
    ('pc000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-001' LIMIT 1),
     'partial', 'ISP Section 3 references MFA requirements at a high level',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)),

    ('pc000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-SA-001' LIMIT 1),
     'full', 'ISP Section 3 — security awareness training mandate',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)),

    -- Access Control Policy → access controls
    ('pc000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-001' LIMIT 1),
     'full', 'Section 3.1 — MFA enforcement requirements',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)),

    ('pc000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-002' LIMIT 1),
     'full', 'Section 3.2 — RBAC and least privilege',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)),

    ('pc000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-003' LIMIT 1),
     'full', 'Section 3.3 — Quarterly access review requirements',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1));
```

---

## Query Patterns

### Policy Library (list with filters)

"List all non-template policies for the org with category and status filters."

```sql
SELECT
    p.id,
    p.identifier,
    p.title,
    p.description,
    p.category,
    p.status,
    p.review_frequency_days,
    p.next_review_at,
    p.last_reviewed_at,
    p.published_at,
    p.tags,
    u_owner.first_name || ' ' || u_owner.last_name AS owner_name,
    pv.version_number AS current_version,
    pv.word_count,
    pv.created_at AS version_date,
    COUNT(DISTINCT pc.control_id) AS linked_controls_count,
    COUNT(DISTINCT ps.id) FILTER (WHERE ps.status = 'pending') AS pending_signoffs,
    CASE
        WHEN p.next_review_at IS NULL THEN 'no_schedule'
        WHEN p.next_review_at < NOW() THEN 'overdue'
        WHEN p.next_review_at < NOW() + INTERVAL '30 days' THEN 'due_soon'
        ELSE 'on_track'
    END AS review_status
FROM policies p
LEFT JOIN users u_owner ON u_owner.id = p.owner_id
LEFT JOIN policy_versions pv ON pv.id = p.current_version_id
LEFT JOIN policy_controls pc ON pc.policy_id = p.id
LEFT JOIN policy_signoffs ps ON ps.policy_id = p.id AND ps.status = 'pending'
WHERE p.org_id = $1
  AND p.is_template = FALSE
GROUP BY p.id, u_owner.first_name, u_owner.last_name, pv.version_number, pv.word_count, pv.created_at
ORDER BY p.identifier;
```

### Policy Gap Detection

"Which active controls have NO policy coverage?"

```sql
SELECT
    c.id,
    c.identifier,
    c.title,
    c.category,
    c.status AS control_status,
    u.first_name || ' ' || u.last_name AS owner_name,
    COUNT(DISTINCT cm.requirement_id) AS mapped_requirements,
    ARRAY_AGG(DISTINCT f.identifier) AS frameworks
FROM controls c
LEFT JOIN policy_controls pc ON pc.control_id = c.id AND pc.org_id = c.org_id
LEFT JOIN users u ON u.id = c.owner_id
LEFT JOIN control_mappings cm ON cm.control_id = c.id AND cm.org_id = c.org_id
LEFT JOIN requirements r ON r.id = cm.requirement_id
LEFT JOIN framework_versions fv ON fv.id = r.framework_version_id
LEFT JOIN frameworks f ON f.id = fv.framework_id
WHERE c.org_id = $1
  AND c.status = 'active'
  AND pc.id IS NULL  -- no policy covers this control
GROUP BY c.id, u.first_name, u.last_name
ORDER BY
    COUNT(DISTINCT cm.requirement_id) DESC,  -- most-mapped ungoverned controls first
    c.identifier;
```

### Policy Gap by Framework

"For framework X, which requirements have controls that lack policy coverage?"

```sql
SELECT
    r.identifier AS requirement_id,
    r.title AS requirement_title,
    c.identifier AS control_id,
    c.title AS control_title,
    CASE
        WHEN pc.id IS NOT NULL THEN 'covered'
        ELSE 'gap'
    END AS policy_coverage
FROM requirements r
JOIN control_mappings cm ON cm.requirement_id = r.id AND cm.org_id = $1
JOIN controls c ON c.id = cm.control_id
LEFT JOIN policy_controls pc ON pc.control_id = c.id AND pc.org_id = $1
WHERE r.framework_version_id = $2
  AND r.is_assessable = TRUE
  AND c.status = 'active'
ORDER BY r.identifier, c.identifier;
```

### Pending Approvals for a User

"Show me all policy sign-offs I need to complete."

```sql
SELECT
    ps.id AS signoff_id,
    p.identifier AS policy_identifier,
    p.title AS policy_title,
    p.category,
    pv.version_number,
    pv.content_summary,
    ps.requested_at,
    ps.due_date,
    u_req.first_name || ' ' || u_req.last_name AS requested_by_name,
    CASE
        WHEN ps.due_date IS NOT NULL AND ps.due_date < CURRENT_DATE THEN 'overdue'
        WHEN ps.due_date IS NOT NULL AND ps.due_date <= CURRENT_DATE + 3 THEN 'due_soon'
        ELSE 'on_time'
    END AS urgency
FROM policy_signoffs ps
JOIN policies p ON p.id = ps.policy_id
JOIN policy_versions pv ON pv.id = ps.policy_version_id
JOIN users u_req ON u_req.id = ps.requested_by
WHERE ps.signer_id = $1
  AND ps.org_id = $2
  AND ps.status = 'pending'
ORDER BY
    CASE
        WHEN ps.due_date IS NOT NULL AND ps.due_date < CURRENT_DATE THEN 0
        WHEN ps.due_date IS NOT NULL THEN 1
        ELSE 2
    END,
    COALESCE(ps.due_date, '9999-12-31'::DATE),
    ps.requested_at;
```

### Sign-off Status for a Policy Version

"What's the approval status of the latest version of POL-IR-001?"

```sql
SELECT
    ps.id AS signoff_id,
    u.first_name || ' ' || u.last_name AS signer_name,
    u.email AS signer_email,
    ps.signer_role,
    ps.status,
    ps.requested_at,
    ps.decided_at,
    ps.due_date,
    ps.comments,
    ps.reminder_count,
    ps.reminder_sent_at
FROM policy_signoffs ps
JOIN users u ON u.id = ps.signer_id
WHERE ps.policy_version_id = $1
  AND ps.org_id = $2
ORDER BY
    CASE ps.status
        WHEN 'pending' THEN 0
        WHEN 'rejected' THEN 1
        WHEN 'approved' THEN 2
        WHEN 'withdrawn' THEN 3
    END,
    ps.requested_at;
```

### Version Comparison

"Get two versions of the same policy for side-by-side diff."

```sql
SELECT
    pv.version_number,
    pv.content,
    pv.content_format,
    pv.content_summary,
    pv.change_summary,
    pv.change_type,
    pv.word_count,
    u.first_name || ' ' || u.last_name AS author_name,
    pv.created_at
FROM policy_versions pv
LEFT JOIN users u ON u.id = pv.created_by
WHERE pv.policy_id = $1
  AND pv.org_id = $2
  AND pv.version_number IN ($3, $4)
ORDER BY pv.version_number;
```

### Template Library

"List all policy templates, optionally filtered by framework."

```sql
SELECT
    p.id,
    p.identifier,
    p.title,
    p.description,
    p.category,
    p.tags,
    f.name AS framework_name,
    f.identifier AS framework_identifier,
    pv.word_count,
    pv.content_summary,
    p.review_frequency_days
FROM policies p
LEFT JOIN frameworks f ON f.id = p.template_framework_id
LEFT JOIN policy_versions pv ON pv.id = p.current_version_id
WHERE p.org_id = $1
  AND p.is_template = TRUE
  AND ($2::UUID IS NULL OR p.template_framework_id = $2)  -- optional framework filter
ORDER BY f.name NULLS LAST, p.category, p.identifier;
```

### Policies Due for Review

"Which published policies are overdue or due for review within 30 days?"

```sql
SELECT
    p.id,
    p.identifier,
    p.title,
    p.category,
    p.next_review_at,
    p.last_reviewed_at,
    p.review_frequency_days,
    u.first_name || ' ' || u.last_name AS owner_name,
    u.email AS owner_email,
    CASE
        WHEN p.next_review_at < CURRENT_DATE THEN 'overdue'
        ELSE 'due_soon'
    END AS review_urgency,
    CURRENT_DATE - p.next_review_at AS days_overdue
FROM policies p
LEFT JOIN users u ON u.id = p.owner_id
WHERE p.org_id = $1
  AND p.is_template = FALSE
  AND p.status IN ('published', 'approved')
  AND p.next_review_at IS NOT NULL
  AND p.next_review_at <= CURRENT_DATE + INTERVAL '30 days'
ORDER BY p.next_review_at ASC;
```

### Controls Covered by a Policy

"What controls does POL-AC-001 govern?"

```sql
SELECT
    c.id,
    c.identifier,
    c.title,
    c.category,
    c.status AS control_status,
    pc.coverage,
    pc.notes,
    u.first_name || ' ' || u.last_name AS linked_by_name,
    pc.created_at AS linked_at
FROM policy_controls pc
JOIN controls c ON c.id = pc.control_id
LEFT JOIN users u ON u.id = pc.linked_by
WHERE pc.policy_id = $1
  AND pc.org_id = $2
ORDER BY c.identifier;
```

---

## Future Considerations

- **Collaborative editing**: Spec §6.1 mentions "collaborative editing." For Sprint 5, versioning is sequential (save-as-new-version). Real-time collaboration (operational transforms / CRDTs) would be a Phase 2+ feature using a service like Liveblocks or Yjs.
- **Digital signatures**: Current sign-offs are database-tracked approvals. True digital signatures (PKI-based, legal-grade) would require a signing service integration (DocuSign, Adobe Sign).
- **Policy distribution tracking**: Spec §6.1 mentions "Policy distribution tracking — confirm employee acknowledgment." This could be implemented via a `policy_acknowledgments` table (user_id, policy_id, acknowledged_at) in a future sprint.
- **AI policy analysis**: Spec §5.1 (AI Policy Agent) will leverage the `policy_versions.content` for NLP classification, gap analysis, and auto-generated policy updates. The schema supports this — content is stored as text.
- **Content diffing**: Version comparison in the UI will need a diff algorithm. The API can return two versions' content; the frontend handles the diff rendering (e.g., using `diff-match-patch` or a similar library).
- **Policy content search**: Full-text search currently covers `policies.title + description`. For searching within policy content, add a GIN index on `policy_versions.content`:
  ```sql
  CREATE INDEX idx_policy_versions_content_search ON policy_versions
      USING gin(to_tsvector('english', content))
      WHERE content_format != 'html';  -- for HTML, strip tags at search time
  ```
  For HTML content, consider a generated column that strips HTML tags, or index at the application layer via Meilisearch.
