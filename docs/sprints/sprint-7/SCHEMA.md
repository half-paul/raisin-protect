# Sprint 7 — Database Schema: Audit Hub

## Overview

Sprint 7 introduces the Audit Hub — the collaboration layer between organizations and their external auditors. This implements spec §6.4 (Audit Hub): dedicated auditor workspaces with controlled access, engagement management with timelines and milestones, evidence request/response workflows, finding management with remediation tracking, and multi-audit support for concurrent engagements.

**Key design decisions:**

- **Audits are org-scoped engagements**: Each audit represents a single audit engagement (e.g., "SOC 2 Type II 2026" or "PCI DSS v4.0.1 ROC Q1"). An org can have multiple concurrent audits (spec §6.4: "Multi-audit management for organizations running concurrent audits"). Each audit tracks its framework, auditor firm, timeline, milestones, and overall status.
- **Audit requests are the evidence exchange protocol**: Auditors create requests for specific evidence artifacts. Internal teams respond by linking evidence, uploading new artifacts, or providing written responses. Requests have their own lifecycle (open → in_progress → submitted → accepted → rejected → closed) and SLA tracking via `due_date`. This implements spec §6.4: "Evidence request/response workflow between auditor and internal teams."
- **Audit findings capture deficiencies**: When auditors identify issues during the engagement, they create findings with severity, category, and remediation requirements. Findings link to specific controls and requirements, and internal teams create remediation plans with progress tracking. This implements spec §6.4: "Finding management with remediation tracking."
- **Audit evidence links create a chain-of-custody**: Rather than duplicating evidence, `audit_evidence_links` creates a bridge between audit requests and existing evidence artifacts. Each link records who submitted it, when, and optional auditor notes/status. This supports spec §3.4.3: "Chain-of-custody tracking for regulatory requirements."
- **Role-based access via the existing `auditor` GRC role**: External auditors are granted the `auditor` role (from Sprint 1 enums). The audit engagement's `auditor_ids` array controls which auditor users can see which audits. All queries enforce org_id isolation AND auditor access checks.
- **Comments enable real-time collaboration**: `audit_comments` supports threaded comments on audits, requests, and findings. This implements spec §6.4: "Real-time commenting and annotation on evidence artifacts."
- **Milestones track engagement progress**: Embedded in the audit record via JSONB `milestones` array rather than a separate table — milestones are lightweight (name + target_date + completed_at) and always queried with their parent audit.

---

## Entity Relationship Diagram

```
                    ┌──────────────────────────────────────────────────────┐
                    │   AUDIT HUB DOMAIN (org-scoped)                      │
                    │                                                      │
 organizations ─┬──▶ audits                                                │
                │   │   (engagement: timeline, framework, firm, status)    │
                │   │   ∞──1 org_frameworks (org_framework_id)             │
                │   │   ∞──1 users (lead_auditor_id)                       │
                │   │   ∞──1 users (internal_lead_id)                      │
                │   │                                                      │
                │   │   1──∞ audit_requests                                │
                │   │         (evidence requests from auditor)             │
                │   │         ∞──1 users (requested_by)                    │
                │   │         ∞──1 users (assigned_to)                     │
                │   │         ∞──1 controls (control_id, optional)         │
                │   │         ∞──1 requirements (requirement_id, optional) │
                │   │                                                      │
                │   │   1──∞ audit_findings                                │
                │   │         (deficiencies found during audit)            │
                │   │         ∞──1 users (found_by)                        │
                │   │         ∞──1 users (remediation_owner_id)            │
                │   │         ∞──1 controls (control_id, optional)         │
                │   │         ∞──1 requirements (requirement_id, optional) │
                │   │                                                      │
                │   │   1──∞ audit_evidence_links                          │
                │   │         (evidence submitted for requests)            │
                │   │         ∞──1 audit_requests (request_id)             │
                │   │         ∞──1 evidence_artifacts (artifact_id)        │
                │   │         ∞──1 users (submitted_by)                    │
                │   │                                                      │
                │   │   1──∞ audit_comments                                │
                │   │         (threaded discussion on audits/requests/     │
                │   │          findings)                                   │
                │   │         ∞──1 users (author_id)                       │
                │   │                                                      │
                └──▶ audit_log (extended with Sprint 7 actions)            │
                    └──────────────────────────────────────────────────────┘
```

**Cross-domain relationships:**
```
audits.org_framework_id ──▶ org_frameworks (Sprint 2)
audit_requests.control_id ──▶ controls (Sprint 2)
audit_requests.requirement_id ──▶ requirements (Sprint 2)
audit_findings.control_id ──▶ controls (Sprint 2)
audit_findings.requirement_id ──▶ requirements (Sprint 2)
audit_evidence_links.artifact_id ──▶ evidence_artifacts (Sprint 3)
evidence_links.audit_id ──▶ audits (new FK, via evidence_link_target_type extension)
```

---

## New Enum Types

### `audit_status`
Audit engagement lifecycle.

```sql
CREATE TYPE audit_status AS ENUM (
    'planning',        -- Engagement scoped, not yet started
    'fieldwork',       -- Active evidence collection and testing
    'review',          -- Auditor reviewing collected evidence
    'draft_report',    -- Auditor preparing draft report
    'management_response', -- Internal team responding to draft findings
    'final_report',    -- Final report being prepared
    'completed',       -- Engagement closed, report issued
    'cancelled'        -- Engagement cancelled before completion
);
```

### `audit_type`
Type of audit engagement.

```sql
CREATE TYPE audit_type AS ENUM (
    'soc2_type1',
    'soc2_type2',
    'iso27001_certification',
    'iso27001_surveillance',
    'pci_dss_roc',
    'pci_dss_saq',
    'gdpr_dpia',
    'internal',
    'custom'
);
```

### `audit_request_status`
Evidence request lifecycle.

```sql
CREATE TYPE audit_request_status AS ENUM (
    'open',            -- Request created, awaiting assignment
    'in_progress',     -- Assigned, internal team working on it
    'submitted',       -- Evidence submitted, awaiting auditor review
    'accepted',        -- Auditor accepted the evidence
    'rejected',        -- Auditor rejected, needs rework
    'closed',          -- Request closed (withdrawn or superseded)
    'overdue'          -- Past due_date without submission
);
```

### `audit_request_priority`
Priority level for evidence requests.

```sql
CREATE TYPE audit_request_priority AS ENUM (
    'critical',
    'high',
    'medium',
    'low'
);
```

### `audit_finding_severity`
Severity of audit findings.

```sql
CREATE TYPE audit_finding_severity AS ENUM (
    'critical',        -- Significant deficiency / material weakness
    'high',            -- Deficiency requiring immediate remediation
    'medium',          -- Observation requiring remediation plan
    'low',             -- Minor observation / improvement opportunity
    'informational'    -- Advisory only, no remediation required
);
```

### `audit_finding_status`
Finding lifecycle with remediation tracking.

```sql
CREATE TYPE audit_finding_status AS ENUM (
    'identified',      -- Finding documented by auditor
    'acknowledged',    -- Internal team acknowledges the finding
    'remediation_planned', -- Remediation plan created
    'remediation_in_progress', -- Remediation underway
    'remediation_complete', -- Internal team claims fix is done
    'verified',        -- Auditor verified remediation is effective
    'risk_accepted',   -- Organization accepted residual risk
    'closed'           -- Finding closed
);
```

### `audit_finding_category`
Category of finding aligned with audit standards.

```sql
CREATE TYPE audit_finding_category AS ENUM (
    'control_deficiency',      -- Control not operating effectively
    'control_gap',             -- Required control not implemented
    'documentation_gap',       -- Missing or incomplete documentation
    'process_gap',             -- Process not followed consistently
    'configuration_issue',     -- System misconfiguration
    'access_control',          -- Inappropriate access or permissions
    'monitoring_gap',          -- Insufficient logging or monitoring
    'policy_violation',        -- Action contrary to established policy
    'vendor_risk',             -- Third-party risk issue
    'data_handling',           -- Data classification or handling issue
    'other'
);
```

### `audit_comment_target_type`
Polymorphic target for comments (which entity the comment is attached to).

```sql
CREATE TYPE audit_comment_target_type AS ENUM (
    'audit',
    'request',
    'finding'
);
```

### `audit_evidence_link_status`
Auditor's evaluation of submitted evidence.

```sql
CREATE TYPE audit_evidence_link_status AS ENUM (
    'pending_review',  -- Submitted, not yet reviewed by auditor
    'accepted',        -- Auditor accepts evidence as sufficient
    'rejected',        -- Auditor rejects evidence (insufficient/irrelevant)
    'needs_clarification' -- Auditor needs additional context
);
```

---

## Extended Enums

### `audit_action` Extensions (20 new values)

```sql
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.status_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.cancelled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.auditor_added';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit.auditor_removed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.assigned';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.submitted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.accepted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.rejected';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_request.closed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.status_changed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.remediation_planned';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_finding.verified';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_evidence.submitted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'audit_evidence.reviewed';
```

### `evidence_link_target_type` Extension

```sql
ALTER TYPE evidence_link_target_type ADD VALUE IF NOT EXISTS 'audit';
```

---

## Tables

### Table 1: `audits`

The core audit engagement record. Tracks the full lifecycle of an audit from planning through final report.

```sql
CREATE TABLE IF NOT EXISTS audits (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Engagement identity
    title               VARCHAR(255) NOT NULL,
    description         TEXT,
    audit_type          audit_type NOT NULL,
    status              audit_status NOT NULL DEFAULT 'planning',

    -- Framework linkage (which framework is being audited)
    org_framework_id    UUID REFERENCES org_frameworks(id) ON DELETE SET NULL,

    -- Audit period (what time range is being audited)
    period_start        DATE,
    period_end          DATE,

    -- Engagement timeline
    planned_start       DATE,
    planned_end         DATE,
    actual_start        DATE,
    actual_end          DATE,

    -- Auditor information
    audit_firm          VARCHAR(255),
    lead_auditor_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    auditor_ids         UUID[] NOT NULL DEFAULT '{}',  -- All auditors with access

    -- Internal team
    internal_lead_id    UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Milestones (lightweight, embedded JSONB)
    -- Array of: { "name": "Kickoff", "target_date": "2026-03-01", "completed_at": null }
    milestones          JSONB NOT NULL DEFAULT '[]',

    -- Report information
    report_type         VARCHAR(100),    -- e.g., "SOC 2 Type II", "ISO 27001 SoA"
    report_url          TEXT,            -- MinIO presigned URL or external link
    report_issued_at    TIMESTAMPTZ,

    -- Summary statistics (denormalized for dashboard performance)
    total_requests      INTEGER NOT NULL DEFAULT 0,
    open_requests       INTEGER NOT NULL DEFAULT 0,
    total_findings      INTEGER NOT NULL DEFAULT 0,
    open_findings       INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    tags                TEXT[] NOT NULL DEFAULT '{}',
    metadata            JSONB NOT NULL DEFAULT '{}',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_audit_period CHECK (
        period_start IS NULL OR period_end IS NULL OR period_start <= period_end
    ),
    CONSTRAINT chk_audit_timeline CHECK (
        planned_start IS NULL OR planned_end IS NULL OR planned_start <= planned_end
    ),
    CONSTRAINT chk_audit_actual CHECK (
        actual_start IS NULL OR actual_end IS NULL OR actual_start <= actual_end
    )
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audits_org ON audits (org_id);
CREATE INDEX IF NOT EXISTS idx_audits_org_status ON audits (org_id, status);
CREATE INDEX IF NOT EXISTS idx_audits_org_framework ON audits (org_id, org_framework_id) WHERE org_framework_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audits_lead_auditor ON audits (lead_auditor_id) WHERE lead_auditor_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audits_internal_lead ON audits (internal_lead_id) WHERE internal_lead_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audits_org_type ON audits (org_id, audit_type);
CREATE INDEX IF NOT EXISTS idx_audits_org_planned_end ON audits (org_id, planned_end) WHERE planned_end IS NOT NULL;

-- Trigger
DROP TRIGGER IF EXISTS trg_audits_updated_at ON audits;
CREATE TRIGGER trg_audits_updated_at
    BEFORE UPDATE ON audits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audits IS 'Audit engagements: SOC 2, ISO 27001, PCI DSS, etc. (spec §6.4)';
COMMENT ON COLUMN audits.org_framework_id IS 'Which activated framework this audit covers';
COMMENT ON COLUMN audits.period_start IS 'Start of the audit observation period';
COMMENT ON COLUMN audits.period_end IS 'End of the audit observation period';
COMMENT ON COLUMN audits.auditor_ids IS 'Array of user IDs with auditor role who can access this engagement';
COMMENT ON COLUMN audits.milestones IS 'JSONB array of {name, target_date, completed_at} milestone objects';
COMMENT ON COLUMN audits.total_requests IS 'Denormalized count of evidence requests (updated by triggers/app)';
COMMENT ON COLUMN audits.open_requests IS 'Denormalized count of non-closed requests (updated by triggers/app)';
```

---

### Table 2: `audit_requests`

Evidence requests created by auditors and fulfilled by internal teams. The primary collaboration mechanism.

```sql
CREATE TABLE IF NOT EXISTS audit_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id            UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,

    -- Request content
    title               VARCHAR(500) NOT NULL,
    description         TEXT NOT NULL,
    priority            audit_request_priority NOT NULL DEFAULT 'medium',
    status              audit_request_status NOT NULL DEFAULT 'open',

    -- What this request relates to
    control_id          UUID REFERENCES controls(id) ON DELETE SET NULL,
    requirement_id      UUID REFERENCES requirements(id) ON DELETE SET NULL,

    -- People
    requested_by        UUID REFERENCES users(id) ON DELETE SET NULL,   -- Auditor who created
    assigned_to         UUID REFERENCES users(id) ON DELETE SET NULL,   -- Internal person responsible

    -- Timing
    due_date            DATE,
    submitted_at        TIMESTAMPTZ,
    reviewed_at         TIMESTAMPTZ,

    -- Auditor feedback (when rejected or needing clarification)
    reviewer_notes      TEXT,

    -- Request metadata
    reference_number    VARCHAR(50),   -- Auditor's own reference (e.g., "PBC-042")
    tags                TEXT[] NOT NULL DEFAULT '{}',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_requests_org ON audit_requests (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_requests_audit ON audit_requests (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_requests_audit_status ON audit_requests (audit_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_requests_assigned ON audit_requests (assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_requested_by ON audit_requests (requested_by) WHERE requested_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_control ON audit_requests (control_id) WHERE control_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_requirement ON audit_requests (requirement_id) WHERE requirement_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_requests_due_date ON audit_requests (audit_id, due_date) WHERE due_date IS NOT NULL AND status NOT IN ('accepted', 'closed');
CREATE INDEX IF NOT EXISTS idx_audit_requests_org_status ON audit_requests (org_id, status);

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_requests_updated_at ON audit_requests;
CREATE TRIGGER trg_audit_requests_updated_at
    BEFORE UPDATE ON audit_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_requests IS 'Evidence requests from auditors to internal teams (spec §6.4)';
COMMENT ON COLUMN audit_requests.reference_number IS 'Auditor''s own reference number (e.g., PBC list item)';
COMMENT ON COLUMN audit_requests.reviewer_notes IS 'Auditor feedback when rejecting or requesting clarification';
COMMENT ON COLUMN audit_requests.due_date IS 'Deadline for evidence submission; used for SLA/overdue tracking';
```

---

### Table 3: `audit_findings`

Deficiencies, observations, and recommendations identified during the audit.

```sql
CREATE TABLE IF NOT EXISTS audit_findings (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id                UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,

    -- Finding content
    title                   VARCHAR(500) NOT NULL,
    description             TEXT NOT NULL,
    severity                audit_finding_severity NOT NULL,
    category                audit_finding_category NOT NULL,
    status                  audit_finding_status NOT NULL DEFAULT 'identified',

    -- What this finding relates to
    control_id              UUID REFERENCES controls(id) ON DELETE SET NULL,
    requirement_id          UUID REFERENCES requirements(id) ON DELETE SET NULL,

    -- People
    found_by                UUID REFERENCES users(id) ON DELETE SET NULL,   -- Auditor who identified
    remediation_owner_id    UUID REFERENCES users(id) ON DELETE SET NULL,   -- Internal owner of fix

    -- Remediation tracking
    remediation_plan        TEXT,
    remediation_due_date    DATE,
    remediation_started_at  TIMESTAMPTZ,
    remediation_completed_at TIMESTAMPTZ,
    verification_notes      TEXT,      -- Auditor notes when verifying remediation
    verified_at             TIMESTAMPTZ,
    verified_by             UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Risk acceptance (if org accepts instead of remediating)
    risk_accepted           BOOLEAN NOT NULL DEFAULT FALSE,
    risk_acceptance_reason  TEXT,
    risk_accepted_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    risk_accepted_at        TIMESTAMPTZ,

    -- Finding metadata
    reference_number        VARCHAR(50),   -- Auditor's finding reference
    recommendation          TEXT,          -- Auditor's recommended fix
    management_response     TEXT,          -- Org's formal response
    tags                    TEXT[] NOT NULL DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_finding_risk_accepted CHECK (
        (risk_accepted = FALSE) OR
        (risk_accepted = TRUE AND risk_acceptance_reason IS NOT NULL)
    ),
    CONSTRAINT chk_finding_verification CHECK (
        verified_at IS NULL OR verified_by IS NOT NULL
    )
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_findings_org ON audit_findings (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_findings_audit ON audit_findings (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_findings_audit_status ON audit_findings (audit_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_findings_audit_severity ON audit_findings (audit_id, severity);
CREATE INDEX IF NOT EXISTS idx_audit_findings_remediation_owner ON audit_findings (remediation_owner_id) WHERE remediation_owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_findings_control ON audit_findings (control_id) WHERE control_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_findings_requirement ON audit_findings (requirement_id) WHERE requirement_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_findings_org_status ON audit_findings (org_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_findings_due_date ON audit_findings (audit_id, remediation_due_date) WHERE remediation_due_date IS NOT NULL AND status NOT IN ('verified', 'closed');

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_findings_updated_at ON audit_findings;
CREATE TRIGGER trg_audit_findings_updated_at
    BEFORE UPDATE ON audit_findings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_findings IS 'Audit deficiencies and observations with remediation tracking (spec §6.4)';
COMMENT ON COLUMN audit_findings.severity IS 'critical=material weakness, high=significant deficiency, medium=observation, low=minor, informational=advisory';
COMMENT ON COLUMN audit_findings.management_response IS 'Organization''s formal response to the finding (spec §6.4: management response phase)';
COMMENT ON COLUMN audit_findings.recommendation IS 'Auditor''s recommended remediation action';
```

---

### Table 4: `audit_evidence_links`

Bridge table connecting evidence artifacts to audit requests. Creates the chain-of-custody trail.

```sql
CREATE TABLE IF NOT EXISTS audit_evidence_links (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id            UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,
    request_id          UUID NOT NULL REFERENCES audit_requests(id) ON DELETE CASCADE,
    artifact_id         UUID NOT NULL REFERENCES evidence_artifacts(id) ON DELETE CASCADE,

    -- Submission info
    submitted_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    submitted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submission_notes    TEXT,

    -- Auditor review
    status              audit_evidence_link_status NOT NULL DEFAULT 'pending_review',
    reviewed_by         UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    review_notes        TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Prevent duplicate submissions of same artifact to same request
    CONSTRAINT uq_audit_evidence_request_artifact UNIQUE (request_id, artifact_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_org ON audit_evidence_links (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_audit ON audit_evidence_links (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_request ON audit_evidence_links (request_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_artifact ON audit_evidence_links (artifact_id);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_status ON audit_evidence_links (request_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_evidence_links_submitted_by ON audit_evidence_links (submitted_by) WHERE submitted_by IS NOT NULL;

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_evidence_links_updated_at ON audit_evidence_links;
CREATE TRIGGER trg_audit_evidence_links_updated_at
    BEFORE UPDATE ON audit_evidence_links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_evidence_links IS 'Evidence artifacts submitted for audit requests — chain-of-custody (spec §3.4.3, §6.4)';
COMMENT ON COLUMN audit_evidence_links.status IS 'Auditor review status: pending_review → accepted/rejected/needs_clarification';
COMMENT ON COLUMN audit_evidence_links.submission_notes IS 'Internal team notes explaining what this evidence demonstrates';
```

---

### Table 5: `audit_comments`

Threaded comments on audits, requests, and findings. Supports real-time collaboration per spec §6.4.

```sql
CREATE TABLE IF NOT EXISTS audit_comments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    audit_id            UUID NOT NULL REFERENCES audits(id) ON DELETE CASCADE,

    -- Polymorphic target
    target_type         audit_comment_target_type NOT NULL,
    target_id           UUID NOT NULL,  -- ID of audit, request, or finding

    -- Comment content
    author_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body                TEXT NOT NULL,

    -- Threading
    parent_comment_id   UUID REFERENCES audit_comments(id) ON DELETE CASCADE,

    -- Metadata
    is_internal         BOOLEAN NOT NULL DEFAULT FALSE,  -- If true, visible only to internal team (not auditors)
    edited_at           TIMESTAMPTZ,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_comments_org ON audit_comments (org_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_audit ON audit_comments (audit_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_target ON audit_comments (target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_author ON audit_comments (author_id);
CREATE INDEX IF NOT EXISTS idx_audit_comments_parent ON audit_comments (parent_comment_id) WHERE parent_comment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_comments_target_created ON audit_comments (target_type, target_id, created_at);

-- Trigger
DROP TRIGGER IF EXISTS trg_audit_comments_updated_at ON audit_comments;
CREATE TRIGGER trg_audit_comments_updated_at
    BEFORE UPDATE ON audit_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE audit_comments IS 'Threaded comments on audit engagements, requests, and findings (spec §6.4)';
COMMENT ON COLUMN audit_comments.target_type IS 'Which entity this comment is on: audit, request, or finding';
COMMENT ON COLUMN audit_comments.target_id IS 'UUID of the target entity (audit, audit_request, or audit_finding)';
COMMENT ON COLUMN audit_comments.is_internal IS 'If true, only visible to internal team (not auditor role users)';
COMMENT ON COLUMN audit_comments.parent_comment_id IS 'Self-referencing FK for threaded replies';
```

---

## FK Extensions to Existing Tables

### `evidence_links` — Add `audit_id` FK

```sql
-- Add audit_id column to evidence_links (Sprint 3 table)
ALTER TABLE evidence_links
    ADD COLUMN IF NOT EXISTS audit_id UUID REFERENCES audits(id) ON DELETE CASCADE;

-- Add index
CREATE INDEX IF NOT EXISTS idx_evidence_links_audit
    ON evidence_links (audit_id) WHERE audit_id IS NOT NULL;

-- Update unique constraint for audit links
ALTER TABLE evidence_links
    ADD CONSTRAINT uq_evidence_link_audit UNIQUE (org_id, artifact_id, audit_id)
    DEFERRABLE INITIALLY DEFERRED;

-- Update CHECK constraint to include 'audit' target_type
-- Drop and recreate since CHECK constraints can't be ALTERed
ALTER TABLE evidence_links DROP CONSTRAINT IF EXISTS chk_evidence_link_target;
ALTER TABLE evidence_links ADD CONSTRAINT chk_evidence_link_target CHECK (
    (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'policy' AND control_id IS NULL AND requirement_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'risk' AND control_id IS NULL AND requirement_id IS NULL AND audit_id IS NULL) OR
    (target_type = 'audit' AND audit_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL)
);
```

---

## Migration Files

Sprint 7 produces **9 migration files** (044–052):

| # | File | Description |
|---|------|-------------|
| 044 | `044_sprint7_enums.sql` | 9 new enum types + 20 audit_action extensions + evidence_link_target_type extension |
| 045 | `045_audits.sql` | `audits` table with milestone JSONB, denormalized counts, timeline constraints |
| 046 | `046_audit_requests.sql` | `audit_requests` table with SLA tracking, control/requirement linkage |
| 047 | `047_audit_findings.sql` | `audit_findings` table with remediation lifecycle, risk acceptance, verification |
| 048 | `048_audit_evidence_links.sql` | `audit_evidence_links` bridge table with auditor review status |
| 049 | `049_audit_comments.sql` | `audit_comments` threaded comments with internal visibility flag |
| 050 | `050_sprint7_fk_cross_refs.sql` | `evidence_links.audit_id` FK + updated CHECK constraint |
| 051 | `051_sprint7_seed_templates.sql` | Audit type templates, common PBC (Prepared by Client) request templates |
| 052 | `052_sprint7_seed_demo.sql` | Demo audit engagement with requests, findings, evidence links, comments |

---

## Seed Data

### Audit Request Templates (PBC List Templates)

Pre-built evidence request templates organized by framework and audit type. ~80 templates:

**SOC 2 Type II PBC templates (25):**
- Organization chart and reporting structure
- Information security policy (current, approved)
- Access control procedures and user provisioning
- Change management procedures and recent change logs
- Incident response plan and recent incident reports
- Business continuity and disaster recovery plans
- Vendor management policy and vendor risk assessments
- Employee onboarding/offboarding procedures
- Security awareness training records (last 12 months)
- Network diagram showing data flows
- Encryption standards and key management procedures
- Logical access reviews (last 4 quarters)
- System monitoring and alerting configuration
- Vulnerability scan reports (last 12 months)
- Penetration test report (most recent)
- Board/management meeting minutes re: security governance
- Risk assessment documentation (most recent)
- Data classification policy and procedures
- Physical security controls documentation
- Backup and recovery test results
- SLA/uptime reports for critical systems
- Third-party SOC reports for subservice providers
- Employee background check procedures
- Firewall and IDS/IPS configuration standards
- Exception/waiver documentation

**PCI DSS v4.0.1 ROC templates (25):**
- Network segmentation documentation and validation results
- Cardholder data flow diagram
- CDE asset inventory
- Firewall/router configuration standards
- Wireless network scanning results
- Encryption key management procedures
- Anti-malware deployment evidence
- Critical security patch installation records (last 6 months)
- Secure SDLC documentation
- Payment page script inventory (Req 6.4.3)
- Role-based access control matrix for CDE
- Multi-factor authentication configuration evidence
- Physical access logs for data center (last 90 days)
- Audit log configuration and samples (Req 10)
- Internal vulnerability scan results (last 4 quarters)
- External ASV scan results (last 4 quarters)
- Penetration test results (internal and external)
- Service provider responsibility matrix
- Incident response plan and test results
- Security awareness training completion records
- PCI DSS policy suite (12 requirement families)
- Targeted Risk Analysis documentation
- Compensating control documentation
- File integrity monitoring configuration and alerts
- Change detection mechanism evidence (Req 11.6.1)

**ISO 27001 templates (15):**
- ISMS scope statement
- Statement of Applicability (SoA)
- Risk assessment methodology
- Risk treatment plan
- Information security policy
- Asset inventory and classification
- Access control policy and procedures
- Cryptography policy
- Physical security procedures
- Operations security procedures
- Communications security procedures
- Supplier relationship procedures
- Incident management procedures
- Business continuity procedures
- Compliance procedures and legal register

**GDPR DPIA templates (15):**
- Data processing register (Article 30)
- Privacy impact assessment methodology
- Data subject rights procedures
- Consent management procedures
- Data breach notification procedures
- Data protection officer appointment documentation
- International data transfer mechanisms
- Data retention schedule
- Privacy by design evidence
- Processor agreements (Article 28)
- Legitimate interest assessments
- Privacy notices and transparency documentation
- Cookie consent implementation evidence
- Employee privacy training records
- Supervisory authority correspondence

### Demo Data

One complete demo audit engagement:

**Audit:** "SOC 2 Type II — 2026 Annual" (status: fieldwork)
- Framework: SOC 2 (linked to org_framework)
- Period: 2026-01-01 to 2026-12-31
- Planned: 2026-02-01 to 2026-04-30
- Audit firm: "Deloitte & Touche LLP"
- Milestones: Kickoff (completed), Fieldwork Start (completed), Interim Testing (in progress), Final Testing, Draft Report, Management Response, Final Report
- 3 auditor users, 1 internal lead

**Requests (8):**
1. "Information Security Policy" — accepted (evidence linked)
2. "Access Control Procedures" — submitted (pending review)
3. "Network Diagram" — in_progress (assigned)
4. "Vulnerability Scan Reports Q1-Q2" — open (unassigned)
5. "Change Management Logs" — accepted (evidence linked)
6. "Incident Response Plan" — rejected (needs update)
7. "Employee Training Records" — overdue (past due date)
8. "Vendor Risk Assessments" — open

**Findings (4):**
1. "Incomplete access reviews" — severity: high, status: remediation_in_progress
2. "Missing MFA on admin accounts" — severity: critical, status: remediation_complete
3. "Stale vendor risk assessments" — severity: medium, status: acknowledged
4. "Backup test not documented" — severity: low, status: identified

**Evidence Links (5):** Connecting existing evidence artifacts to requests 1, 2, 5

**Comments (6):** Mix of auditor questions, internal responses, and internal-only notes

---

## Query Patterns

### Audit Dashboard Statistics

```sql
-- Org-level audit summary
SELECT
    COUNT(*) AS total_audits,
    COUNT(*) FILTER (WHERE status IN ('planning', 'fieldwork', 'review', 'draft_report', 'management_response')) AS active_audits,
    COUNT(*) FILTER (WHERE status = 'completed') AS completed_audits,
    SUM(open_requests) AS total_open_requests,
    SUM(open_findings) AS total_open_findings
FROM audits
WHERE org_id = $1;
```

### Request Queue (Internal Team View)

```sql
-- Open requests assigned to current user or unassigned
SELECT ar.*, a.title AS audit_title, a.audit_type
FROM audit_requests ar
JOIN audits a ON a.id = ar.audit_id
WHERE ar.org_id = $1
  AND ar.status NOT IN ('accepted', 'closed')
  AND (ar.assigned_to = $2 OR ar.assigned_to IS NULL)
ORDER BY
    CASE ar.priority
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    ar.due_date NULLS LAST;
```

### Overdue Request Detection

```sql
-- Flag overdue requests
SELECT ar.*, a.title AS audit_title
FROM audit_requests ar
JOIN audits a ON a.id = ar.audit_id
WHERE ar.org_id = $1
  AND ar.due_date < CURRENT_DATE
  AND ar.status NOT IN ('accepted', 'closed', 'overdue');
```

### Finding Remediation Progress

```sql
-- Finding remediation progress per audit
SELECT
    a.id AS audit_id,
    a.title,
    COUNT(*) AS total_findings,
    COUNT(*) FILTER (WHERE af.status IN ('verified', 'closed')) AS resolved,
    COUNT(*) FILTER (WHERE af.status = 'risk_accepted') AS accepted,
    COUNT(*) FILTER (WHERE af.status NOT IN ('verified', 'closed', 'risk_accepted')) AS open,
    COUNT(*) FILTER (WHERE af.severity = 'critical' AND af.status NOT IN ('verified', 'closed', 'risk_accepted')) AS critical_open
FROM audits a
LEFT JOIN audit_findings af ON af.audit_id = a.id
WHERE a.org_id = $1
GROUP BY a.id, a.title;
```

### Audit Readiness Score

```sql
-- Per-audit readiness: % of requests fulfilled
SELECT
    a.id,
    a.title,
    a.total_requests,
    COUNT(*) FILTER (WHERE ar.status = 'accepted') AS accepted_requests,
    CASE
        WHEN a.total_requests = 0 THEN 0
        ELSE ROUND(
            100.0 * COUNT(*) FILTER (WHERE ar.status = 'accepted') / a.total_requests
        )
    END AS readiness_pct
FROM audits a
LEFT JOIN audit_requests ar ON ar.audit_id = a.id
WHERE a.org_id = $1
GROUP BY a.id, a.title, a.total_requests;
```

### Evidence Chain-of-Custody

```sql
-- Full chain of custody for an evidence artifact in an audit
SELECT
    ael.id AS link_id,
    ea.title AS artifact_title,
    ea.file_name,
    ar.title AS request_title,
    ar.reference_number,
    u_sub.email AS submitted_by_email,
    ael.submitted_at,
    ael.status AS review_status,
    u_rev.email AS reviewed_by_email,
    ael.reviewed_at,
    ael.review_notes
FROM audit_evidence_links ael
JOIN evidence_artifacts ea ON ea.id = ael.artifact_id
JOIN audit_requests ar ON ar.id = ael.request_id
LEFT JOIN users u_sub ON u_sub.id = ael.submitted_by
LEFT JOIN users u_rev ON u_rev.id = ael.reviewed_by
WHERE ael.audit_id = $1 AND ael.org_id = $2
ORDER BY ael.submitted_at;
```

---

## Auditor Access Model

Auditors see ONLY audits where their user ID is in `audits.auditor_ids`:

```sql
-- Middleware filter for auditor role
-- Applied to ALL audit hub queries when user.role = 'auditor'
AND (
    $user_role != 'auditor'
    OR a.id IN (
        SELECT id FROM audits
        WHERE org_id = $org_id AND $user_id = ANY(auditor_ids)
    )
)
```

Internal comments (`is_internal = TRUE`) are filtered out for auditor role users:

```sql
-- Comment visibility filter
WHERE audit_comments.audit_id = $1
  AND ($user_role != 'auditor' OR audit_comments.is_internal = FALSE)
```

---

## Notes for DBE

1. **Migration numbering:** Continue from 044 (last Sprint 6 migration was 043)
2. **Enum pattern:** Use `DO $$ BEGIN ... EXCEPTION WHEN duplicate_object THEN NULL; END $$;` for new enums, `ALTER TYPE ... ADD VALUE IF NOT EXISTS` for extensions
3. **FK cross-refs:** Migration 050 handles the `evidence_links.audit_id` FK and CHECK constraint update — must run after `audits` table exists
4. **Seed templates:** Store PBC request templates in a JSONB format suitable for cloning: `{ "title": "...", "description": "...", "framework": "soc2", "audit_type": "soc2_type2", "tags": [...] }`
5. **Demo data:** Use existing demo org, framework, and users. Create 2-3 auditor-role users in seed. Link to existing evidence artifacts from Sprint 3 demo data.
6. **Denormalized counts:** `audits.total_requests` / `open_requests` / `total_findings` / `open_findings` should be maintained by application code (not triggers), consistent with how other sprints handle denormalization.
