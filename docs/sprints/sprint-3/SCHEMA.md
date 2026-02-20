# Sprint 3 — Database Schema: Evidence Management

## Overview

Sprint 3 introduces the evidence management layer — the mechanism by which organizations upload, link, version, evaluate, and track freshness of compliance artifacts. Evidence bridges the gap between "we have a control" and "we can prove it works" (spec §3.4).

**Key design decisions:**
- **MinIO object storage for files**: Evidence files (PDFs, screenshots, configs, exports) are stored in MinIO (S3-compatible). The database stores metadata and the MinIO object key — never the file bytes.
- **Versioning via parent chain**: New versions of an artifact point to the same `parent_artifact_id`. The latest version is `is_current = TRUE`. This preserves full history while keeping queries simple.
- **Many-to-many linking**: One evidence artifact can satisfy requirements across multiple controls (cross-framework reuse per spec §3.4.3). One control can have multiple evidence artifacts.
- **Evaluation workflow**: Evidence is reviewed/evaluated before being considered sufficient. Evaluations track the confidence of evidence against a specific control requirement.
- **Freshness tracking**: Every artifact has a `collection_date` and `expires_at`. Staleness is computed from these fields — no separate tracking table needed.

---

## Entity Relationship Diagram

```
                    ┌───────────────────────────────────┐
                    │   EVIDENCE DOMAIN (org-scoped)     │
                    │                                   │
 organizations ─┬──▶ evidence_artifacts                 │
                │   │   (file metadata, MinIO key)      │
                │   │   ↻ parent_artifact_id (versions) │
                │   │                                   │
                │   │   1──∞ evidence_links ──▶ controls │
                │   │          (also links to            │
                │   │           requirements, policies)  │
                │   │                                   │
                │   │   1──∞ evidence_evaluations        │
                │   │          (review/approve/reject)   │
                │   │                                   │
                └──▶ audit_log (extended)                │
                    └───────────────────────────────────┘
```

**Relationships:**
```
evidence_artifacts      ∞──1  organizations        (org_id)
evidence_artifacts      ∞──1  users                (uploaded_by)
evidence_artifacts      ∞──1  evidence_artifacts   (parent_artifact_id, version chain)
evidence_links          ∞──1  evidence_artifacts   (artifact_id)
evidence_links          ∞──1  controls             (control_id, optional)
evidence_links          ∞──1  requirements         (requirement_id, optional)
evidence_evaluations    ∞──1  evidence_artifacts   (artifact_id)
evidence_evaluations    ∞──1  evidence_links       (evidence_link_id, optional)
evidence_evaluations    ∞──1  users                (evaluated_by)
```

---

## New Enums

```sql
-- Types of evidence artifacts (from spec §3.4.1)
CREATE TYPE evidence_type AS ENUM (
    'screenshot',           -- UI screenshots, dashboards
    'api_response',         -- API output captures
    'configuration_export', -- System config exports (JSON, YAML, etc.)
    'log_sample',           -- Audit log extracts, SIEM exports
    'policy_document',      -- Policy PDFs, signed documents
    'access_list',          -- User access reports, permission exports
    'vulnerability_report', -- Scan results (Qualys, Nessus, etc.)
    'certificate',          -- SSL certs, ISO certs, SOC reports
    'training_record',      -- Security awareness training completion
    'penetration_test',     -- Pen test reports
    'audit_report',         -- Internal/external audit reports
    'other'                 -- Catch-all for uncategorized evidence
);

-- Evidence lifecycle status
CREATE TYPE evidence_status AS ENUM (
    'draft',                -- Uploaded but not yet submitted for review
    'pending_review',       -- Submitted, awaiting evaluation
    'approved',             -- Evaluated and deemed sufficient
    'rejected',             -- Evaluated and deemed insufficient
    'expired',              -- Past its expiration date (stale)
    'superseded'            -- Replaced by a newer version
);

-- How evidence was collected (from spec §3.4.1)
CREATE TYPE evidence_collection_method AS ENUM (
    'manual_upload',        -- Human uploaded the file
    'automated_pull',       -- Pulled from an integration automatically
    'api_ingestion',        -- Pushed via API by external system
    'screenshot_capture',   -- Automated screenshot tool
    'system_export'         -- Exported from a connected system
);

-- Evaluation verdict
CREATE TYPE evidence_evaluation_verdict AS ENUM (
    'sufficient',           -- Evidence fully satisfies the requirement
    'partial',              -- Evidence partially satisfies (gaps remain)
    'insufficient',         -- Evidence does not satisfy the requirement
    'needs_update'          -- Evidence is outdated or needs refresh
);

-- What entity type an evidence link targets
CREATE TYPE evidence_link_target_type AS ENUM (
    'control',              -- Linked to a control
    'requirement',          -- Linked directly to a requirement
    'policy'                -- Linked to a policy (future Sprint 5, but define now)
);
```

### Extend Existing Enums

```sql
-- Add Sprint 3 actions to audit_action enum
ALTER TYPE audit_action ADD VALUE 'evidence.uploaded';
ALTER TYPE audit_action ADD VALUE 'evidence.updated';
ALTER TYPE audit_action ADD VALUE 'evidence.deleted';
ALTER TYPE audit_action ADD VALUE 'evidence.version_created';
ALTER TYPE audit_action ADD VALUE 'evidence.status_changed';
ALTER TYPE audit_action ADD VALUE 'evidence.linked';
ALTER TYPE audit_action ADD VALUE 'evidence.unlinked';
ALTER TYPE audit_action ADD VALUE 'evidence.evaluated';
ALTER TYPE audit_action ADD VALUE 'evidence.expired';
```

---

## Tables

### evidence_artifacts

The core evidence table. Each row represents one version of an evidence artifact — the file metadata, storage location, and freshness tracking fields.

```sql
CREATE TABLE evidence_artifacts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Identity & display
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,
    evidence_type           evidence_type NOT NULL,
    status                  evidence_status NOT NULL DEFAULT 'draft',
    collection_method       evidence_collection_method NOT NULL DEFAULT 'manual_upload',
    
    -- File storage (MinIO)
    file_name               VARCHAR(500) NOT NULL,          -- original filename
    file_size               BIGINT NOT NULL,                -- size in bytes
    mime_type               VARCHAR(255) NOT NULL,           -- e.g., 'application/pdf', 'image/png'
    object_key              VARCHAR(1000) NOT NULL UNIQUE,   -- MinIO object key: '{org_id}/{artifact_id}/{version}/{filename}'
    checksum_sha256         VARCHAR(64),                     -- SHA-256 of file contents for integrity
    
    -- Versioning
    parent_artifact_id      UUID REFERENCES evidence_artifacts(id) ON DELETE SET NULL,
    version                 INT NOT NULL DEFAULT 1,          -- version number (1, 2, 3...)
    is_current              BOOLEAN NOT NULL DEFAULT TRUE,   -- TRUE = latest version
    
    -- Freshness tracking (spec §3.4.1)
    collection_date         DATE NOT NULL,                   -- when the evidence was collected/generated
    expires_at              TIMESTAMPTZ,                     -- when evidence becomes stale (NULL = never expires)
    freshness_period_days   INT,                             -- how often evidence should be refreshed (e.g., 90)
    
    -- Source attribution
    source_system           VARCHAR(255),                    -- e.g., 'aws-config', 'okta', 'manual'
    source_integration_id   UUID,                            -- FK to integrations table (Sprint 9, nullable for now)
    
    -- Ownership
    uploaded_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Tags for search/filter
    tags                    TEXT[] DEFAULT '{}',              -- free-form tags: ['pci', 'network', 'q1-2026']
    
    -- Metadata
    metadata                JSONB NOT NULL DEFAULT '{}',     -- extensible: redaction info, page count, etc.
    
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_evidence_artifacts_org ON evidence_artifacts (org_id);
CREATE INDEX idx_evidence_artifacts_status ON evidence_artifacts (org_id, status);
CREATE INDEX idx_evidence_artifacts_type ON evidence_artifacts (org_id, evidence_type);
CREATE INDEX idx_evidence_artifacts_collection_method ON evidence_artifacts (org_id, collection_method);
CREATE INDEX idx_evidence_artifacts_parent ON evidence_artifacts (parent_artifact_id) 
    WHERE parent_artifact_id IS NOT NULL;
CREATE INDEX idx_evidence_artifacts_current ON evidence_artifacts (org_id, is_current) 
    WHERE is_current = TRUE;
CREATE INDEX idx_evidence_artifacts_uploaded_by ON evidence_artifacts (uploaded_by);
CREATE INDEX idx_evidence_artifacts_collection_date ON evidence_artifacts (org_id, collection_date DESC);
CREATE INDEX idx_evidence_artifacts_expires ON evidence_artifacts (org_id, expires_at) 
    WHERE expires_at IS NOT NULL AND status NOT IN ('expired', 'superseded');
CREATE INDEX idx_evidence_artifacts_tags ON evidence_artifacts USING gin (tags);

-- Full-text search on title + description
CREATE INDEX idx_evidence_artifacts_search ON evidence_artifacts
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Trigger
CREATE TRIGGER trg_evidence_artifacts_updated_at
    BEFORE UPDATE ON evidence_artifacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `object_key` is the MinIO path. Pattern: `{org_id}/{artifact_id}/{version}/{filename}`. This ensures org isolation at the storage level.
- `parent_artifact_id` creates a version chain. All versions of the same logical artifact share the same parent (the original). When uploading v2, the v1 row gets `is_current = FALSE` and `status = 'superseded'`.
- `freshness_period_days` is how often this evidence needs to be refreshed (e.g., "access review evidence must be refreshed every 90 days"). `expires_at` is computed from `collection_date + freshness_period_days` (or set manually).
- `checksum_sha256` supports chain-of-custody integrity verification (spec §3.4.3).
- `tags` uses PostgreSQL array type with GIN index for flexible tagging without a separate junction table.
- `source_integration_id` is a forward-looking FK for Sprint 9 (Integration Engine). Nullable for now.
- No hard file size limit in schema — enforced at API/MinIO layer (recommended: 100MB per spec §13.1).

---

### evidence_links

Many-to-many relationship between evidence artifacts and the entities they support (controls, requirements, policies). This is where "upload once, apply everywhere" lives (spec §3.4.3).

```sql
CREATE TABLE evidence_links (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    artifact_id         UUID NOT NULL REFERENCES evidence_artifacts(id) ON DELETE CASCADE,
    
    -- Polymorphic target (what this evidence is linked to)
    target_type         evidence_link_target_type NOT NULL,
    control_id          UUID REFERENCES controls(id) ON DELETE CASCADE,
    requirement_id      UUID REFERENCES requirements(id) ON DELETE CASCADE,
    -- policy_id        UUID REFERENCES policies(id) ON DELETE CASCADE,  -- Sprint 5
    
    -- Link metadata
    notes               TEXT,                              -- why this evidence supports this control/requirement
    strength            VARCHAR(20) NOT NULL DEFAULT 'primary'
                        CHECK (strength IN ('primary', 'supporting', 'supplementary')),
    linked_by           UUID REFERENCES users(id) ON DELETE SET NULL,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure each artifact links to each target only once
    CONSTRAINT uq_evidence_link_control UNIQUE (org_id, artifact_id, control_id) 
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT uq_evidence_link_requirement UNIQUE (org_id, artifact_id, requirement_id) 
        DEFERRABLE INITIALLY DEFERRED,
    
    -- Enforce that the correct FK is populated for the target_type
    CONSTRAINT chk_evidence_link_target CHECK (
        (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL) OR
        (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL) OR
        (target_type = 'policy' AND control_id IS NULL AND requirement_id IS NULL)
    )
);

-- Indexes
CREATE INDEX idx_evidence_links_org ON evidence_links (org_id);
CREATE INDEX idx_evidence_links_artifact ON evidence_links (artifact_id);
CREATE INDEX idx_evidence_links_control ON evidence_links (control_id) 
    WHERE control_id IS NOT NULL;
CREATE INDEX idx_evidence_links_requirement ON evidence_links (requirement_id) 
    WHERE requirement_id IS NOT NULL;
CREATE INDEX idx_evidence_links_target_type ON evidence_links (org_id, target_type);

-- Trigger
CREATE TRIGGER trg_evidence_links_updated_at
    BEFORE UPDATE ON evidence_links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- **Polymorphic linking**: `target_type` determines which FK column is populated. The `CHECK` constraint enforces correctness at the DB level.
- `strength` describes how strongly the evidence supports the target:
  - `primary` — directly and fully demonstrates compliance
  - `supporting` — provides additional context or corroboration
  - `supplementary` — nice-to-have, not essential
- Deferrable unique constraints allow transactional bulk inserts while still preventing duplicates.
- When evidence is linked to a **control**, it implicitly covers all requirements that control is mapped to (cross-framework reuse). Linking directly to a **requirement** is for cases where evidence doesn't map through a control.
- `policy` target_type is defined now but the FK won't be added until Sprint 5 (Policy Management).

---

### evidence_evaluations

Tracks review/approval decisions on evidence artifacts. Each evaluation assesses whether a piece of evidence sufficiently satisfies a particular control or requirement link.

```sql
CREATE TABLE evidence_evaluations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    artifact_id         UUID NOT NULL REFERENCES evidence_artifacts(id) ON DELETE CASCADE,
    evidence_link_id    UUID REFERENCES evidence_links(id) ON DELETE SET NULL,  -- optional: evaluate against specific link
    
    -- Evaluation
    verdict             evidence_evaluation_verdict NOT NULL,
    confidence          VARCHAR(20) NOT NULL DEFAULT 'medium'
                        CHECK (confidence IN ('high', 'medium', 'low')),
    comments            TEXT NOT NULL,                      -- required: explain the verdict
    
    -- What was found lacking (if not sufficient)
    missing_elements    TEXT[],                             -- e.g., ['date', 'signature', 'scope statement']
    remediation_notes   TEXT,                               -- suggested actions to fix
    
    -- Who evaluated
    evaluated_by        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- No updated_at: evaluations are immutable (create a new one to re-evaluate)
);

-- Indexes
CREATE INDEX idx_evidence_evaluations_org ON evidence_evaluations (org_id);
CREATE INDEX idx_evidence_evaluations_artifact ON evidence_evaluations (artifact_id);
CREATE INDEX idx_evidence_evaluations_link ON evidence_evaluations (evidence_link_id) 
    WHERE evidence_link_id IS NOT NULL;
CREATE INDEX idx_evidence_evaluations_verdict ON evidence_evaluations (org_id, verdict);
CREATE INDEX idx_evidence_evaluations_evaluator ON evidence_evaluations (evaluated_by);
CREATE INDEX idx_evidence_evaluations_created ON evidence_evaluations (artifact_id, created_at DESC);
```

**Design notes:**
- **Immutable**: Evaluations are append-only — no `updated_at`. To re-evaluate, create a new evaluation. This preserves the full history of review decisions.
- `evidence_link_id` is optional. When provided, the evaluation is specifically about "does this evidence satisfy THIS control/requirement link?" When NULL, it's a general quality evaluation of the artifact itself.
- `confidence` aligns with spec §3.4.2 AI-Powered Evidence Evaluation: "Confidence scoring: High / Medium / Low / Insufficient". Human evaluators also use this scale.
- `missing_elements` is a text array — enables structured tracking of what's missing (spec §3.4.2: "Auto-flag missing elements: dates, roles, signatures, required sections").
- `evaluated_by` uses `ON DELETE RESTRICT` — can't delete a user who has made evaluations (audit trail integrity).
- The latest evaluation for an artifact determines its effective status (query: `ORDER BY created_at DESC LIMIT 1`).

---

## Migration Order

Migrations continue from Sprint 2:

14. `014_sprint3_enums.sql` — New enum types (evidence_type, evidence_status, evidence_collection_method, evidence_evaluation_verdict, evidence_link_target_type) + audit_action extensions
15. `015_evidence_artifacts.sql` — Evidence artifacts table + all indexes + trigger
16. `016_evidence_links.sql` — Evidence links table + indexes + constraints + trigger
17. `017_evidence_evaluations.sql` — Evidence evaluations table + indexes

---

## Seed Data

### Example Evidence Artifacts for Demo Org

```sql
-- Example: MFA configuration screenshot
INSERT INTO evidence_artifacts (
    id, org_id, title, description, evidence_type, status, collection_method,
    file_name, file_size, mime_type, object_key, version, is_current,
    collection_date, expires_at, freshness_period_days, source_system,
    uploaded_by, tags
) VALUES
    (
        'e0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'Okta MFA Configuration Export',
        'Export of MFA policy settings from Okta showing enforcement for all users in the production org.',
        'configuration_export',
        'approved',
        'system_export',
        'okta-mfa-config-2026-02.json',
        15234,
        'application/json',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000001/1/okta-mfa-config-2026-02.json',
        1, TRUE,
        '2026-02-15',
        '2026-05-15 00:00:00+00',
        90,
        'okta',
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
        ARRAY['mfa', 'okta', 'access-control', 'q1-2026']
    ),
    (
        'e0000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'Quarterly Access Review Report - Q1 2026',
        'Complete access review report covering all production systems. Reviews conducted by department managers.',
        'access_list',
        'approved',
        'manual_upload',
        'access-review-q1-2026.pdf',
        2456789,
        'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000002/1/access-review-q1-2026.pdf',
        1, TRUE,
        '2026-01-31',
        '2026-04-30 00:00:00+00',
        90,
        'manual',
        (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
        ARRAY['access-review', 'quarterly', 'q1-2026']
    ),
    (
        'e0000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'Vulnerability Scan Results - February 2026',
        'Qualys vulnerability scan of production environment. No critical or high findings.',
        'vulnerability_report',
        'pending_review',
        'automated_pull',
        'qualys-scan-feb-2026.pdf',
        892345,
        'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000003/1/qualys-scan-feb-2026.pdf',
        1, TRUE,
        '2026-02-18',
        '2026-03-18 00:00:00+00',
        30,
        'qualys',
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
        ARRAY['vulnerability', 'qualys', 'scan', 'feb-2026']
    ),
    (
        'e0000000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'Information Security Policy v3.1',
        'Current information security policy document. Approved by CISO, signed by all employees.',
        'policy_document',
        'approved',
        'manual_upload',
        'infosec-policy-v3.1.pdf',
        345678,
        'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000004/1/infosec-policy-v3.1.pdf',
        1, TRUE,
        '2026-01-15',
        '2027-01-15 00:00:00+00',
        365,
        'manual',
        (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
        ARRAY['policy', 'infosec', 'annual']
    ),
    (
        'e0000000-0000-0000-0000-000000000005',
        'a0000000-0000-0000-0000-000000000001',
        'AWS CloudTrail Logging Configuration',
        'Screenshot and config export showing CloudTrail enabled in all regions with S3 logging.',
        'screenshot',
        'approved',
        'screenshot_capture',
        'aws-cloudtrail-config.png',
        456789,
        'image/png',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000005/1/aws-cloudtrail-config.png',
        1, TRUE,
        '2026-02-10',
        '2026-05-10 00:00:00+00',
        90,
        'aws',
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1),
        ARRAY['aws', 'cloudtrail', 'logging', 'monitoring']
    ),
    (
        'e0000000-0000-0000-0000-000000000006',
        'a0000000-0000-0000-0000-000000000001',
        'Security Awareness Training Completion - 2026',
        'KnowBe4 training completion report showing 100% employee participation.',
        'training_record',
        'approved',
        'automated_pull',
        'knowbe4-training-2026.csv',
        89012,
        'text/csv',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000006/1/knowbe4-training-2026.csv',
        1, TRUE,
        '2026-02-01',
        '2027-02-01 00:00:00+00',
        365,
        'knowbe4',
        (SELECT id FROM users WHERE email = 'it@acme.example.com' LIMIT 1),
        ARRAY['training', 'knowbe4', 'annual', '2026']
    );
```

### Example Evidence Links

```sql
-- Link MFA config to MFA control
INSERT INTO evidence_links (id, org_id, artifact_id, target_type, control_id, strength, linked_by) VALUES
    (
        'l0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000001',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-001' LIMIT 1),
        'primary',
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)
    ),
    -- Link access review to access review control
    (
        'l0000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000002',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-003' LIMIT 1),
        'primary',
        (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)
    );
```

### Example Evidence Evaluations

```sql
-- Evaluate the MFA config as sufficient
INSERT INTO evidence_evaluations (id, org_id, artifact_id, evidence_link_id, verdict, confidence, comments, evaluated_by) VALUES
    (
        'ev000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000001',
        'l0000000-0000-0000-0000-000000000001',
        'sufficient',
        'high',
        'MFA is enforced for all user types. Configuration export shows Okta MFA policy is set to "Always" with no exceptions. Meets PCI DSS 8.3 and SOC 2 CC6.1 requirements.',
        (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)
    ),
    -- Evaluate access review as sufficient
    (
        'ev000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000002',
        'l0000000-0000-0000-0000-000000000002',
        'sufficient',
        'high',
        'Access review covers all production systems. Department managers signed off on all entries. Stale accounts identified and removed. Meets quarterly cadence requirement.',
        (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1)
    );
```

---

## Query Patterns

### Evidence for a Control

"What evidence supports CTRL-AC-001?"

```sql
SELECT
    ea.id,
    ea.title,
    ea.evidence_type,
    ea.status,
    ea.collection_date,
    ea.expires_at,
    ea.is_current,
    el.strength,
    el.notes AS link_notes,
    CASE
        WHEN ea.expires_at IS NULL THEN 'fresh'
        WHEN ea.expires_at < NOW() THEN 'expired'
        WHEN ea.expires_at < NOW() + INTERVAL '30 days' THEN 'expiring_soon'
        ELSE 'fresh'
    END AS freshness_status
FROM evidence_links el
JOIN evidence_artifacts ea ON ea.id = el.artifact_id
WHERE el.control_id = $1
  AND el.org_id = $2
  AND ea.is_current = TRUE
  AND ea.status NOT IN ('superseded', 'expired')
ORDER BY el.strength, ea.collection_date DESC;
```

### Controls for an Evidence Artifact

"Which controls does this evidence artifact support?"

```sql
SELECT
    c.id,
    c.identifier,
    c.title,
    c.category,
    c.status AS control_status,
    el.strength,
    el.notes
FROM evidence_links el
JOIN controls c ON c.id = el.control_id
WHERE el.artifact_id = $1
  AND el.org_id = $2
  AND el.target_type = 'control'
ORDER BY c.identifier;
```

### Staleness Alerts

"Which evidence artifacts are expired or expiring within 30 days?"

```sql
SELECT
    ea.id,
    ea.title,
    ea.evidence_type,
    ea.status,
    ea.collection_date,
    ea.expires_at,
    ea.freshness_period_days,
    CASE
        WHEN ea.expires_at < NOW() THEN 'expired'
        ELSE 'expiring_soon'
    END AS alert_level,
    ea.expires_at - NOW() AS time_remaining,
    COUNT(el.id) AS linked_controls_count
FROM evidence_artifacts ea
LEFT JOIN evidence_links el ON el.artifact_id = ea.id AND el.target_type = 'control'
WHERE ea.org_id = $1
  AND ea.is_current = TRUE
  AND ea.status NOT IN ('superseded', 'draft')
  AND ea.expires_at IS NOT NULL
  AND ea.expires_at < NOW() + INTERVAL '30 days'
GROUP BY ea.id
ORDER BY ea.expires_at ASC;
```

### Version History

"Get all versions of an evidence artifact"

```sql
-- Given an artifact ID, find all versions (including self)
WITH RECURSIVE version_chain AS (
    -- Find the root (original artifact)
    SELECT id, parent_artifact_id, version, title, status, collection_date, is_current, created_at
    FROM evidence_artifacts
    WHERE id = $1 AND org_id = $2
    
    UNION ALL
    
    -- Find siblings with same parent
    SELECT ea.id, ea.parent_artifact_id, ea.version, ea.title, ea.status, ea.collection_date, ea.is_current, ea.created_at
    FROM evidence_artifacts ea
    JOIN version_chain vc ON ea.parent_artifact_id = vc.parent_artifact_id
    WHERE ea.org_id = $2 AND ea.id != vc.id
)
SELECT * FROM version_chain
ORDER BY version DESC;

-- Simpler alternative: query by parent_artifact_id directly
SELECT id, version, title, status, collection_date, is_current, file_name, file_size, created_at
FROM evidence_artifacts
WHERE (parent_artifact_id = $1 OR id = $1)
  AND org_id = $2
ORDER BY version DESC;
```

### Evidence Coverage per Control

"For each of my controls, how much evidence do I have?"

```sql
SELECT
    c.id,
    c.identifier,
    c.title,
    c.category,
    COUNT(DISTINCT el.artifact_id) FILTER (WHERE ea.is_current AND ea.status = 'approved') AS approved_evidence_count,
    COUNT(DISTINCT el.artifact_id) FILTER (WHERE ea.is_current AND ea.status = 'pending_review') AS pending_evidence_count,
    COUNT(DISTINCT el.artifact_id) FILTER (WHERE ea.is_current AND ea.expires_at < NOW()) AS expired_evidence_count,
    CASE
        WHEN COUNT(DISTINCT el.artifact_id) FILTER (WHERE ea.is_current AND ea.status = 'approved' AND (ea.expires_at IS NULL OR ea.expires_at > NOW())) > 0 THEN 'covered'
        WHEN COUNT(DISTINCT el.artifact_id) FILTER (WHERE ea.is_current AND ea.status = 'pending_review') > 0 THEN 'pending'
        ELSE 'gap'
    END AS evidence_status
FROM controls c
LEFT JOIN evidence_links el ON el.control_id = c.id AND el.org_id = c.org_id
LEFT JOIN evidence_artifacts ea ON ea.id = el.artifact_id
WHERE c.org_id = $1
  AND c.status = 'active'
GROUP BY c.id
ORDER BY c.identifier;
```

---

## MinIO Configuration

### Bucket Structure

```
rp-evidence/                          -- Single bucket for all orgs
├── {org_id}/                         -- Org isolation at prefix level
│   ├── {artifact_id}/               -- One directory per logical artifact
│   │   ├── 1/                       -- Version 1
│   │   │   └── original-filename.pdf
│   │   └── 2/                       -- Version 2
│   │       └── updated-filename.pdf
│   └── ...
└── ...
```

### Bucket Policy

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Deny",
            "Principal": "*",
            "Action": "s3:*",
            "Resource": ["arn:aws:s3:::rp-evidence/*"],
            "Condition": {
                "StringNotEquals": {
                    "aws:PrincipalType": "Service"
                }
            }
        }
    ]
}
```

**Access model:** All file access goes through the API (presigned URLs). No direct bucket access from the browser. Presigned upload URLs expire in 15 minutes. Presigned download URLs expire in 1 hour.

### File Validation Rules (API Layer)

| Rule | Value |
|------|-------|
| Max file size | 100 MB |
| Allowed MIME types | application/pdf, image/png, image/jpeg, image/gif, application/json, text/csv, text/plain, application/xml, application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, application/vnd.openxmlformats-officedocument.wordprocessingml.document |
| Max filename length | 255 characters |
| Filename sanitization | Strip path separators, control characters, null bytes |

---

## Future Considerations

- **AI-powered evaluation** (Sprint 3+/Phase 2): Spec §3.4.2 calls for AI evidence analysis. The `evidence_evaluations` table supports both human and AI evaluators (AI evaluations would use a system user or `evaluated_by = NULL` with a `source = 'ai'` metadata field).
- **Evidence redaction** (spec §3.4.3): For sharing sensitive evidence with auditors. Can be added via `metadata` field or a future `evidence_redactions` table.
- **Chain-of-custody tracking** (spec §3.4.3): The combination of `checksum_sha256`, immutable evaluations, `uploaded_by`, and audit log entries provides basic chain-of-custody. A dedicated `evidence_custody_events` table could be added for strict regulatory needs.
- **Policy linking**: The `evidence_link_target_type` enum includes `'policy'` but the FK column is commented out until Sprint 5 (Policy Management).
- **Automated collection scheduling**: Sprint 9 (Integration Engine) will add scheduled pulls. The `source_integration_id` FK is pre-positioned for this.
- **MinIO lifecycle rules**: Configure automatic deletion of superseded versions after a retention period (e.g., 1 year) to manage storage costs.
