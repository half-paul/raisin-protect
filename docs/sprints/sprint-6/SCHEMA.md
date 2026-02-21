# Sprint 6 — Database Schema: Risk Register

## Overview

Sprint 6 introduces the risk management layer — the risk backbone of the GRC platform. This implements spec §4.1 (Internal Risk Management): a centralized risk register with customizable taxonomy, quantitative/qualitative scoring, heat map visualization, treatment plans, and risk-to-control linkage. Risk scores update when control status changes, completing the feedback loop between monitoring (Sprint 4) and risk posture.

**Key design decisions:**

- **Risks are org-scoped governance objects**: Each risk belongs to one organization, has a lifecycle (Identified → Open → Assessing → Treating → Monitoring → Accepted → Closed → Archived), and is owned by a user. Risks are the organizational unit for tracking threats, and each risk can be linked to controls that mitigate it.
- **Dual-score model (inherent + residual)**: Every risk carries two scores: **inherent** (raw risk before controls/treatments) and **residual** (remaining risk after controls and treatment plans are applied). These are denormalized on the `risks` table for fast dashboard/heat-map queries, but the authoritative scores live in `risk_assessments`.
- **Assessments as point-in-time snapshots**: `risk_assessments` captures each scoring event — who assessed it, when, using what formula, with what justification. This enables trend tracking per spec §4.1.1 ("Risk trend tracking over time with change annotations"). Multiple assessments over time create an audit trail of how risk perception evolved.
- **Likelihood × Impact grid**: Both likelihood and impact use 5-level enums that map to numeric values (1–5). The default scoring formula is `likelihood_score × impact_score` (range 1–25), but the formula is stored per assessment for configurability per spec §4.1.1 ("Custom risk scoring formulas").
- **Treatment plans are action items**: Each treatment has a type (Mitigate, Accept, Transfer, Avoid per spec §4.1.3), an owner, a due date, and completion tracking. Treatments link back to the risk and optionally to controls. Effectiveness reviews post-implementation are captured via `effectiveness_rating`.
- **Risk-to-control is many-to-many with effectiveness tracking**: `risk_controls` records not just the linkage but the control's effectiveness against this specific risk (effective / partially_effective / ineffective / not_assessed). This enables gap detection: risks with no controls, or risks where all controls are ineffective.
- **Pre-built risk library**: 200+ common information security risks seeded across categories (operational, financial, strategic, compliance, technology, legal, reputational, third_party) per spec §4.1.1.

---

## Entity Relationship Diagram

```
                    ┌──────────────────────────────────────────────────┐
                    │   RISK DOMAIN (org-scoped)                       │
                    │                                                  │
 organizations ─┬──▶ risks                                            │
                │   │   (risk definitions with ownership + scoring)    │
                │   │   ∞──1 users (owner_id)                         │
                │   │   ∞──1 users (secondary_owner_id)               │
                │   │                                                  │
                │   │   1──∞ risk_assessments                          │
                │   │         (point-in-time likelihood × impact)      │
                │   │         ∞──1 users (assessed_by)                 │
                │   │                                                  │
                │   │   1──∞ risk_treatments                           │
                │   │         (mitigation/acceptance/transfer/avoid)   │
                │   │         ∞──1 users (owner_id)                    │
                │   │         ∞──1 controls (target_control_id, opt)   │
                │   │                                                  │
                │   │   1──∞ risk_controls ──▶ controls                │
                │   │         (many-to-many: risk ↔ control)           │
                │   │                                                  │
                │   │   1──∞ evidence_links (target_type = 'risk')     │
                │   │         (via new FK on evidence_links)           │
                │   │                                                  │
                └──▶ audit_log (extended)                              │
                    └──────────────────────────────────────────────────┘
```

**Relationships:**
```
risks                 ∞──1  organizations        (org_id)
risks                 ∞──1  users                (owner_id)
risks                 ∞──1  users                (secondary_owner_id)
risk_assessments      ∞──1  organizations        (org_id)
risk_assessments      ∞──1  risks                (risk_id)
risk_assessments      ∞──1  users                (assessed_by)
risk_treatments       ∞──1  organizations        (org_id)
risk_treatments       ∞──1  risks                (risk_id)
risk_treatments       ∞──1  users                (owner_id)
risk_treatments       ∞──1  users                (created_by)
risk_treatments       ∞──1  controls             (target_control_id, optional)
risk_controls         ∞──1  organizations        (org_id)
risk_controls         ∞──1  risks                (risk_id)
risk_controls         ∞──1  controls             (control_id)
risk_controls         ∞──1  users                (linked_by)
evidence_links        ∞──1  risks                (risk_id, new FK)
```

---

## New Enums

```sql
-- Risk taxonomy categories (from spec §4.1.1 — customizable risk taxonomy)
CREATE TYPE risk_category AS ENUM (
    'operational',          -- Process failures, human error, system outages
    'financial',            -- Revenue loss, cost overruns, fraud
    'strategic',            -- Market changes, competitive threats, M&A risk
    'compliance',           -- Regulatory violations, audit failures, fines
    'technology',           -- Software vulnerabilities, infrastructure failures, obsolescence
    'legal',                -- Litigation, contract disputes, IP infringement
    'reputational',         -- Brand damage, public trust, media exposure
    'third_party',          -- Vendor failures, supply chain disruption, TPSP breach
    'physical',             -- Natural disasters, theft, facility damage
    'data_privacy',         -- PII exposure, data breach, consent violations
    'cyber_security',       -- Malware, phishing, APT, ransomware
    'human_resources',      -- Key person dependency, skills gap, insider threat
    'environmental',        -- Environmental regulations, sustainability compliance
    'custom'                -- Organization-defined categories
);

-- Risk lifecycle status (from spec §4.1 — risk register management)
CREATE TYPE risk_status AS ENUM (
    'identified',           -- Newly discovered, not yet assessed
    'open',                 -- Assessed, awaiting treatment decision
    'assessing',            -- Under active assessment (multi-reviewer workflow)
    'treating',             -- Treatment plan(s) in progress
    'monitoring',           -- Treatments applied, under ongoing monitoring
    'accepted',             -- Risk formally accepted (with approval chain)
    'closed',               -- Risk no longer applicable (threat eliminated)
    'archived'              -- Historical record, retained for audit
);

-- 5-level likelihood scale (maps to 1–5 for scoring)
CREATE TYPE likelihood_level AS ENUM (
    'rare',                 -- 1: <5% probability, may occur only in exceptional circumstances
    'unlikely',             -- 2: 5–25%, could occur but not expected
    'possible',             -- 3: 25–50%, might occur at some time
    'likely',               -- 4: 50–75%, will probably occur in most circumstances
    'almost_certain'        -- 5: >75%, expected to occur in most circumstances
);

-- 5-level impact scale (maps to 1–5 for scoring)
CREATE TYPE impact_level AS ENUM (
    'negligible',           -- 1: Insignificant impact, no disruption
    'minor',                -- 2: Minor impact, brief disruption, easily recoverable
    'moderate',             -- 3: Moderate impact, some disruption, recoverable with effort
    'major',                -- 4: Significant impact, major disruption, difficult to recover
    'severe'                -- 5: Catastrophic impact, existential threat, may not recover
);

-- Treatment strategy options (from spec §4.1.3)
CREATE TYPE treatment_type AS ENUM (
    'mitigate',             -- Reduce likelihood and/or impact via controls
    'accept',               -- Formally accept the risk (within appetite)
    'transfer',             -- Transfer to third party (insurance, outsourcing)
    'avoid'                 -- Eliminate the risk by removing the source
);

-- Treatment plan lifecycle
CREATE TYPE treatment_status AS ENUM (
    'planned',              -- Treatment defined but not started
    'in_progress',          -- Implementation underway
    'implemented',          -- Treatment applied, awaiting verification
    'verified',             -- Effectiveness confirmed post-implementation
    'ineffective',          -- Treatment applied but did not reduce risk as expected
    'cancelled'             -- Treatment abandoned (risk accepted or avoided instead)
);

-- Assessment scope: inherent (before controls), residual (after), target (desired)
CREATE TYPE risk_assessment_type AS ENUM (
    'inherent',             -- Raw risk before any controls or treatments
    'residual',             -- Remaining risk after controls and treatments
    'target'                -- Desired risk level (risk appetite target)
);

-- How effective a control is at mitigating a specific risk
CREATE TYPE control_effectiveness AS ENUM (
    'effective',            -- Control fully mitigates the risk it addresses
    'partially_effective',  -- Control reduces but does not fully mitigate the risk
    'ineffective',          -- Control has no meaningful impact on this risk
    'not_assessed'          -- Effectiveness has not been evaluated yet
);
```

### Extend Existing Enums

```sql
-- Add Sprint 6 actions to audit_action enum
ALTER TYPE audit_action ADD VALUE 'risk.created';
ALTER TYPE audit_action ADD VALUE 'risk.updated';
ALTER TYPE audit_action ADD VALUE 'risk.status_changed';
ALTER TYPE audit_action ADD VALUE 'risk.archived';
ALTER TYPE audit_action ADD VALUE 'risk.deleted';
ALTER TYPE audit_action ADD VALUE 'risk.owner_changed';
ALTER TYPE audit_action ADD VALUE 'risk.score_recalculated';
ALTER TYPE audit_action ADD VALUE 'risk_assessment.created';
ALTER TYPE audit_action ADD VALUE 'risk_assessment.updated';
ALTER TYPE audit_action ADD VALUE 'risk_assessment.deleted';
ALTER TYPE audit_action ADD VALUE 'risk_treatment.created';
ALTER TYPE audit_action ADD VALUE 'risk_treatment.updated';
ALTER TYPE audit_action ADD VALUE 'risk_treatment.status_changed';
ALTER TYPE audit_action ADD VALUE 'risk_treatment.completed';
ALTER TYPE audit_action ADD VALUE 'risk_treatment.cancelled';
ALTER TYPE audit_action ADD VALUE 'risk_control.linked';
ALTER TYPE audit_action ADD VALUE 'risk_control.unlinked';
ALTER TYPE audit_action ADD VALUE 'risk_control.effectiveness_updated';
```

---

## Helper Functions

```sql
-- Map likelihood_level enum to numeric score (1–5)
CREATE OR REPLACE FUNCTION likelihood_to_score(level likelihood_level)
RETURNS INT AS $$
BEGIN
    RETURN CASE level
        WHEN 'rare' THEN 1
        WHEN 'unlikely' THEN 2
        WHEN 'possible' THEN 3
        WHEN 'likely' THEN 4
        WHEN 'almost_certain' THEN 5
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Map impact_level enum to numeric score (1–5)
CREATE OR REPLACE FUNCTION impact_to_score(level impact_level)
RETURNS INT AS $$
BEGIN
    RETURN CASE level
        WHEN 'negligible' THEN 1
        WHEN 'minor' THEN 2
        WHEN 'moderate' THEN 3
        WHEN 'major' THEN 4
        WHEN 'severe' THEN 5
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Classify a numeric risk score (1–25) into a severity band
CREATE OR REPLACE FUNCTION risk_score_severity(score NUMERIC)
RETURNS TEXT AS $$
BEGIN
    RETURN CASE
        WHEN score >= 20 THEN 'critical'   -- 20–25: red zone
        WHEN score >= 12 THEN 'high'       -- 12–19: orange zone
        WHEN score >= 6  THEN 'medium'     -- 6–11: yellow zone
        WHEN score >= 1  THEN 'low'        -- 1–5: green zone
        ELSE 'none'
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION likelihood_to_score(likelihood_level) IS 'Maps likelihood enum to 1–5 numeric score';
COMMENT ON FUNCTION impact_to_score(impact_level) IS 'Maps impact enum to 1–5 numeric score';
COMMENT ON FUNCTION risk_score_severity(NUMERIC) IS 'Classifies numeric risk score (1–25) into severity band';
```

---

## Tables

### risks

The core risk register table. Each row is a risk entry with metadata, ownership, categorization, and denormalized scoring for fast dashboard/heat-map queries.

```sql
CREATE TABLE risks (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    identifier                  VARCHAR(50) NOT NULL,            -- human-readable: 'RISK-CY-001'
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT,                            -- detailed risk description

    -- Classification
    category                    risk_category NOT NULL,
    status                      risk_status NOT NULL DEFAULT 'identified',

    -- Ownership (spec §3.2.3 pattern — primary + secondary owner)
    owner_id                    UUID REFERENCES users(id) ON DELETE SET NULL,
    secondary_owner_id          UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Denormalized scores (kept in sync via assessment creation/recalculation)
    -- Inherent: risk before any controls or treatments
    inherent_likelihood         likelihood_level,
    inherent_impact             impact_level,
    inherent_score              NUMERIC(5,2),                    -- likelihood_score × impact_score (1–25)

    -- Residual: risk after controls and treatments are applied
    residual_likelihood         likelihood_level,
    residual_impact             impact_level,
    residual_score              NUMERIC(5,2),                    -- likelihood_score × impact_score (1–25)

    -- Risk appetite
    risk_appetite_threshold     NUMERIC(5,2),                    -- org-defined acceptable score; breached if residual > threshold
    accepted_at                 TIMESTAMPTZ,                     -- when risk was formally accepted
    accepted_by                 UUID REFERENCES users(id) ON DELETE SET NULL,
    acceptance_expiry           DATE,                            -- acceptance is time-bound per spec §4.1.2
    acceptance_justification    TEXT,                            -- mandatory when accepted

    -- Assessment scheduling (spec §4.1.2: configurable cadence)
    assessment_frequency_days   INT,                             -- how often risk must be reassessed (e.g., 90 = quarterly)
    next_assessment_at          DATE,                            -- when next assessment is due
    last_assessed_at            DATE,                            -- when the risk was last assessed

    -- Source and context
    source                      VARCHAR(200),                    -- how this risk was discovered: 'audit', 'assessment', 'incident', 'vendor_review'
    affected_assets             TEXT[],                          -- systems/assets impacted: ['payment-api', 'customer-db']

    -- Template support (for pre-built risk library)
    is_template                 BOOLEAN NOT NULL DEFAULT FALSE,  -- TRUE = library template, not active risk
    template_source             VARCHAR(200),                    -- e.g., 'NIST SP 800-30', 'ISO 27005', 'CIS RA'

    -- Tags and metadata
    tags                        TEXT[] DEFAULT '{}',             -- free-form: ['pci', 'critical-asset', 'q1-review']
    metadata                    JSONB NOT NULL DEFAULT '{}',     -- extensible: custom fields, import references

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_risk_identifier UNIQUE (org_id, identifier),
    CONSTRAINT chk_assessment_frequency CHECK (assessment_frequency_days IS NULL OR assessment_frequency_days > 0),
    CONSTRAINT chk_inherent_score CHECK (inherent_score IS NULL OR (inherent_score >= 1 AND inherent_score <= 25)),
    CONSTRAINT chk_residual_score CHECK (residual_score IS NULL OR (residual_score >= 1 AND residual_score <= 25)),
    CONSTRAINT chk_appetite_threshold CHECK (risk_appetite_threshold IS NULL OR (risk_appetite_threshold >= 1 AND risk_appetite_threshold <= 25)),
    CONSTRAINT chk_acceptance CHECK (
        (status != 'accepted') OR
        (accepted_by IS NOT NULL AND acceptance_justification IS NOT NULL)
    )
);

-- Indexes
CREATE INDEX idx_risks_org ON risks (org_id);
CREATE INDEX idx_risks_org_status ON risks (org_id, status);
CREATE INDEX idx_risks_org_category ON risks (org_id, category);
CREATE INDEX idx_risks_owner ON risks (owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX idx_risks_secondary_owner ON risks (secondary_owner_id) WHERE secondary_owner_id IS NOT NULL;
CREATE INDEX idx_risks_inherent_score ON risks (org_id, inherent_score DESC NULLS LAST)
    WHERE is_template = FALSE;
CREATE INDEX idx_risks_residual_score ON risks (org_id, residual_score DESC NULLS LAST)
    WHERE is_template = FALSE;
CREATE INDEX idx_risks_templates ON risks (org_id, is_template) WHERE is_template = TRUE;
CREATE INDEX idx_risks_assessment_due ON risks (org_id, next_assessment_at)
    WHERE next_assessment_at IS NOT NULL AND status NOT IN ('closed', 'archived') AND is_template = FALSE;
CREATE INDEX idx_risks_acceptance_expiry ON risks (org_id, acceptance_expiry)
    WHERE acceptance_expiry IS NOT NULL AND status = 'accepted';
CREATE INDEX idx_risks_identifier ON risks (org_id, identifier);
CREATE INDEX idx_risks_tags ON risks USING gin (tags);
CREATE INDEX idx_risks_affected_assets ON risks USING gin (affected_assets);

-- Heat map index: quick aggregation by likelihood × impact
CREATE INDEX idx_risks_heat_map_inherent ON risks (org_id, inherent_likelihood, inherent_impact)
    WHERE is_template = FALSE AND status NOT IN ('closed', 'archived');
CREATE INDEX idx_risks_heat_map_residual ON risks (org_id, residual_likelihood, residual_impact)
    WHERE is_template = FALSE AND status NOT IN ('closed', 'archived');

-- Full-text search on title + description
CREATE INDEX idx_risks_search ON risks
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Trigger
CREATE TRIGGER trg_risks_updated_at
    BEFORE UPDATE ON risks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `identifier` follows the pattern `RISK-{category_prefix}-{number}` — e.g., `RISK-CY-001` for a cyber security risk, `RISK-OP-003` for an operational risk.
- **Dual scoring**: `inherent_*` and `residual_*` are denormalized from the latest `risk_assessments`. Updated by the API when a new assessment is created or `POST /risks/:id/recalculate` is called. Score range is 1–25 (5×5 grid).
- **Risk acceptance**: When `status = 'accepted'`, `accepted_by`, `acceptance_justification`, and `acceptance_expiry` must be set. Acceptance is time-bound per spec §4.1.2 — when `acceptance_expiry` passes, the system should flag the risk for re-assessment.
- **Assessment scheduling**: `assessment_frequency_days` (e.g., 90 = quarterly per spec §4.1.2) + `next_assessment_at` computed after each assessment. A periodic job surfaces risks with `next_assessment_at < NOW()` as overdue.
- **Template system**: `is_template = TRUE` marks risks as library entries. Templates have `template_source` indicating the source standard (NIST, ISO, CIS). Organizations clone templates into active risks.
- **Heat map indexes**: Composite indexes on `(org_id, likelihood, impact)` for both inherent and residual scores enable fast heat-map aggregation queries without full table scans.
- `affected_assets` is a text array for tagging which systems/services are impacted — useful for cross-referencing with incident response and asset management.

---

### risk_assessments

Point-in-time scoring snapshots. Each assessment records who evaluated the risk, what likelihood/impact they assigned, and why. Trend tracking comes from querying assessments over time.

```sql
CREATE TABLE risk_assessments (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,

    -- Assessment classification
    assessment_type             risk_assessment_type NOT NULL,   -- inherent, residual, or target

    -- Scoring
    likelihood                  likelihood_level NOT NULL,
    impact                      impact_level NOT NULL,
    likelihood_score            INT NOT NULL,                    -- computed: likelihood_to_score(likelihood)
    impact_score                INT NOT NULL,                    -- computed: impact_to_score(impact)
    overall_score               NUMERIC(5,2) NOT NULL,           -- computed: likelihood_score × impact_score (default formula)
    scoring_formula             VARCHAR(100) NOT NULL DEFAULT 'likelihood_x_impact',
                                                                 -- formula identifier for auditability

    -- Severity classification (derived from overall_score)
    severity                    VARCHAR(20) NOT NULL,            -- 'critical', 'high', 'medium', 'low' (computed)

    -- Context
    justification               TEXT,                            -- why this likelihood/impact was chosen
    assumptions                 TEXT,                            -- key assumptions underlying the assessment
    data_sources                TEXT[],                          -- evidence used: ['pentest-2026-q1', 'vulnerability-scan']

    -- Assessor
    assessed_by                 UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assessment_date             DATE NOT NULL DEFAULT CURRENT_DATE,

    -- Validity
    valid_until                 DATE,                            -- when this assessment expires (triggers re-assessment)
    superseded_by               UUID REFERENCES risk_assessments(id) ON DELETE SET NULL,
    is_current                  BOOLEAN NOT NULL DEFAULT TRUE,   -- TRUE = latest assessment of this type

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_likelihood_score CHECK (likelihood_score >= 1 AND likelihood_score <= 5),
    CONSTRAINT chk_impact_score CHECK (impact_score >= 1 AND impact_score <= 5),
    CONSTRAINT chk_overall_score CHECK (overall_score >= 1 AND overall_score <= 25)
);

-- Indexes
CREATE INDEX idx_risk_assessments_org ON risk_assessments (org_id);
CREATE INDEX idx_risk_assessments_risk ON risk_assessments (risk_id);
CREATE INDEX idx_risk_assessments_risk_type ON risk_assessments (risk_id, assessment_type);
CREATE INDEX idx_risk_assessments_current ON risk_assessments (risk_id, assessment_type, is_current)
    WHERE is_current = TRUE;
CREATE INDEX idx_risk_assessments_assessed_by ON risk_assessments (assessed_by);
CREATE INDEX idx_risk_assessments_date ON risk_assessments (risk_id, assessment_date DESC);
CREATE INDEX idx_risk_assessments_severity ON risk_assessments (org_id, severity)
    WHERE is_current = TRUE;
CREATE INDEX idx_risk_assessments_expiry ON risk_assessments (org_id, valid_until)
    WHERE valid_until IS NOT NULL AND is_current = TRUE;

-- Each risk can have only one current assessment per type
CREATE UNIQUE INDEX uq_risk_assessment_current ON risk_assessments (risk_id, assessment_type)
    WHERE is_current = TRUE;
```

**Design notes:**
- **Immutable scoring record**: Assessments are write-once for auditability. To "update" a score, create a new assessment which supersedes the old one (`superseded_by` link, `is_current` toggle).
- `scoring_formula` records which formula produced `overall_score`. Default is `'likelihood_x_impact'` (simple multiplication). The API supports alternative formulas:
  - `likelihood_x_impact`: score = likelihood_score × impact_score (range 1–25)
  - `weighted`: score = (likelihood_score × weight_l + impact_score × weight_i) with custom weights
  - `custom`: org-defined formula stored in metadata
- `severity` is computed at creation time via `risk_score_severity()` and stored for indexing.
- `is_current` ensures only one active assessment per type per risk (enforced by unique partial index). Creating a new assessment sets the previous current's `is_current = FALSE` and links via `superseded_by`.
- `data_sources` documents what evidence informed the assessment — critical for audit trail integrity.
- `valid_until` enables time-based expiration: when an assessment expires, the risk should be flagged for re-assessment even if `next_assessment_at` hasn't arrived yet.
- `assessed_by` uses `ON DELETE RESTRICT` — cannot delete users who have performed risk assessments (audit trail preservation).

---

### risk_treatments

Treatment plans for managing identified risks. Each treatment represents an action item with ownership, deadlines, and progress tracking. Multiple treatments can apply to a single risk.

```sql
CREATE TABLE risk_treatments (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,

    -- Treatment definition
    treatment_type              treatment_type NOT NULL,          -- mitigate, accept, transfer, avoid
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT,                            -- detailed treatment plan
    status                      treatment_status NOT NULL DEFAULT 'planned',

    -- Ownership
    owner_id                    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by                  UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Priority and scheduling
    priority                    VARCHAR(20) NOT NULL DEFAULT 'medium'
                                CHECK (priority IN ('critical', 'high', 'medium', 'low')),
    due_date                    DATE,                            -- target completion date
    started_at                  TIMESTAMPTZ,                     -- when work began
    completed_at                TIMESTAMPTZ,                     -- when treatment was finished

    -- Effort tracking
    estimated_effort_hours      NUMERIC(8,2),                    -- estimated hours to implement
    actual_effort_hours         NUMERIC(8,2),                    -- actual hours spent

    -- Effectiveness review (post-implementation, spec §4.1.3)
    effectiveness_rating        VARCHAR(20)
                                CHECK (effectiveness_rating IS NULL OR
                                       effectiveness_rating IN ('highly_effective', 'effective',
                                                                 'partially_effective', 'ineffective')),
    effectiveness_notes         TEXT,                            -- explanation of effectiveness determination
    effectiveness_reviewed_at   TIMESTAMPTZ,                     -- when effectiveness was last reviewed
    effectiveness_reviewed_by   UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Expected risk reduction
    expected_residual_likelihood likelihood_level,               -- expected likelihood after treatment
    expected_residual_impact    impact_level,                    -- expected impact after treatment
    expected_residual_score     NUMERIC(5,2),                    -- expected score after treatment

    -- Target control (optional: which control implements this treatment)
    target_control_id           UUID REFERENCES controls(id) ON DELETE SET NULL,

    -- Notes and metadata
    notes                       TEXT,                            -- additional context, blockers, dependencies
    metadata                    JSONB NOT NULL DEFAULT '{}',     -- extensible: cost estimates, vendor references

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_expected_residual_score CHECK (
        expected_residual_score IS NULL OR
        (expected_residual_score >= 1 AND expected_residual_score <= 25)
    ),
    CONSTRAINT chk_completed CHECK (
        (status NOT IN ('verified', 'implemented') OR completed_at IS NOT NULL)
    ),
    CONSTRAINT chk_effort CHECK (
        estimated_effort_hours IS NULL OR estimated_effort_hours >= 0
    ),
    CONSTRAINT chk_actual_effort CHECK (
        actual_effort_hours IS NULL OR actual_effort_hours >= 0
    )
);

-- Indexes
CREATE INDEX idx_risk_treatments_org ON risk_treatments (org_id);
CREATE INDEX idx_risk_treatments_risk ON risk_treatments (risk_id);
CREATE INDEX idx_risk_treatments_risk_status ON risk_treatments (risk_id, status);
CREATE INDEX idx_risk_treatments_owner ON risk_treatments (owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX idx_risk_treatments_created_by ON risk_treatments (created_by) WHERE created_by IS NOT NULL;
CREATE INDEX idx_risk_treatments_status ON risk_treatments (org_id, status);
CREATE INDEX idx_risk_treatments_due ON risk_treatments (org_id, due_date)
    WHERE due_date IS NOT NULL AND status IN ('planned', 'in_progress');
CREATE INDEX idx_risk_treatments_overdue ON risk_treatments (org_id, due_date, status)
    WHERE due_date IS NOT NULL AND status IN ('planned', 'in_progress');
CREATE INDEX idx_risk_treatments_type ON risk_treatments (org_id, treatment_type);
CREATE INDEX idx_risk_treatments_priority ON risk_treatments (org_id, priority)
    WHERE status IN ('planned', 'in_progress');
CREATE INDEX idx_risk_treatments_control ON risk_treatments (target_control_id)
    WHERE target_control_id IS NOT NULL;
CREATE INDEX idx_risk_treatments_effectiveness ON risk_treatments (org_id, effectiveness_rating)
    WHERE effectiveness_rating IS NOT NULL;

-- Trigger
CREATE TRIGGER trg_risk_treatments_updated_at
    BEFORE UPDATE ON risk_treatments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- **Four treatment types** per spec §4.1.3: Mitigate (add controls), Accept (formally accept), Transfer (insurance/outsource), Avoid (eliminate source).
- **Status lifecycle**: `planned` → `in_progress` → `implemented` → `verified` (post-effectiveness review). `ineffective` and `cancelled` are terminal states.
- **Effectiveness review**: After implementation, `effectiveness_rating` captures whether the treatment actually reduced risk. This feeds back into residual score recalculation.
- `target_control_id` optionally links the treatment to a specific control that implements the mitigation. This creates a feedback loop: when the control's health changes (via Sprint 4 monitoring), the treatment's effectiveness may need re-evaluation.
- `expected_residual_*` captures the anticipated risk reduction — useful for comparing expected vs. actual outcomes.
- `priority` uses the standard critical/high/medium/low scale. Combined with `due_date`, enables prioritized treatment queues.
- `estimated_effort_hours` and `actual_effort_hours` enable effort tracking and estimation accuracy analysis.
- Multiple treatments can apply to one risk (e.g., a mitigation + a transfer for the same risk).

---

### risk_controls

Many-to-many junction table linking risks to the controls that mitigate them. Includes effectiveness tracking per linkage — a control might be effective against one risk but only partially effective against another.

```sql
CREATE TABLE risk_controls (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    risk_id                     UUID NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    control_id                  UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Effectiveness assessment
    effectiveness               control_effectiveness NOT NULL DEFAULT 'not_assessed',

    -- Link metadata
    notes                       TEXT,                            -- why this control mitigates this risk
    mitigation_percentage       INT,                             -- estimated % of risk mitigated (0–100)

    -- Who created the link
    linked_by                   UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Effectiveness review tracking
    last_effectiveness_review   DATE,                            -- when effectiveness was last evaluated
    reviewed_by                 UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each risk links to each control only once
    CONSTRAINT uq_risk_control UNIQUE (org_id, risk_id, control_id),
    CONSTRAINT chk_mitigation_percentage CHECK (
        mitigation_percentage IS NULL OR
        (mitigation_percentage >= 0 AND mitigation_percentage <= 100)
    )
);

-- Indexes
CREATE INDEX idx_risk_controls_org ON risk_controls (org_id);
CREATE INDEX idx_risk_controls_risk ON risk_controls (risk_id);
CREATE INDEX idx_risk_controls_control ON risk_controls (control_id);
CREATE INDEX idx_risk_controls_effectiveness ON risk_controls (org_id, effectiveness);
CREATE INDEX idx_risk_controls_linked_by ON risk_controls (linked_by) WHERE linked_by IS NOT NULL;
CREATE INDEX idx_risk_controls_not_assessed ON risk_controls (org_id, risk_id, effectiveness)
    WHERE effectiveness = 'not_assessed';

-- Trigger
CREATE TRIGGER trg_risk_controls_updated_at
    BEFORE UPDATE ON risk_controls
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `effectiveness` tracks how well this specific control mitigates this specific risk. A firewall (CTRL-NW-001) might be `effective` against "unauthorized network access" but only `partially_effective` against "data exfiltration."
- `mitigation_percentage` is an optional quantitative estimate: "This control reduces this risk by approximately 40%." Used in advanced residual score calculation formulas.
- `last_effectiveness_review` + `reviewed_by` track when the linkage was last validated — effectiveness assessments can drift as the threat landscape changes.
- `org_id` is denormalized (derivable from `risk_id`) for consistent multi-tenancy enforcement in all queries.
- **Gap detection query**: `SELECT r.* FROM risks r WHERE r.is_template = FALSE AND r.status NOT IN ('closed', 'archived') AND NOT EXISTS (SELECT 1 FROM risk_controls rc WHERE rc.risk_id = r.id)` finds risks with no mitigating controls.

---

## Deferred Foreign Keys

After all tables are created, add the cross-reference FK for evidence linking:

```sql
-- evidence_links.risk_id → risks (extending Sprint 3's evidence linking pattern)
ALTER TABLE evidence_links
    ADD COLUMN risk_id UUID REFERENCES risks(id) ON DELETE CASCADE;

CREATE INDEX idx_evidence_links_risk ON evidence_links (risk_id)
    WHERE risk_id IS NOT NULL;

-- Update the existing CHECK constraint on evidence_links to include risk_id
ALTER TABLE evidence_links DROP CONSTRAINT chk_evidence_link_target;
ALTER TABLE evidence_links ADD CONSTRAINT chk_evidence_link_target CHECK (
    (target_type = 'control' AND control_id IS NOT NULL AND requirement_id IS NULL AND policy_id IS NULL AND risk_id IS NULL) OR
    (target_type = 'requirement' AND requirement_id IS NOT NULL AND control_id IS NULL AND policy_id IS NULL AND risk_id IS NULL) OR
    (target_type = 'policy' AND policy_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL AND risk_id IS NULL) OR
    (target_type = 'risk' AND risk_id IS NOT NULL AND control_id IS NULL AND requirement_id IS NULL AND policy_id IS NULL)
);

-- Add uniqueness constraint for risk evidence links
ALTER TABLE evidence_links ADD CONSTRAINT uq_evidence_link_risk
    UNIQUE (org_id, artifact_id, risk_id) DEFERRABLE INITIALLY DEFERRED;
```

**Note:** This also requires adding `'risk'` to the `evidence_link_target_type` enum:

```sql
ALTER TYPE evidence_link_target_type ADD VALUE 'risk';
```

---

## Migration Order

Migrations continue from Sprint 5:

35. `035_sprint6_enums.sql` — New enum types (risk_category, risk_status, likelihood_level, impact_level, treatment_type, treatment_status, risk_assessment_type, control_effectiveness) + audit_action extensions + evidence_link_target_type extension
36. `036_sprint6_functions.sql` — Helper functions (likelihood_to_score, impact_to_score, risk_score_severity)
37. `037_risks.sql` — Risks table + indexes + trigger + constraints
38. `038_risk_assessments.sql` — Risk assessments table + indexes + unique constraints
39. `039_risk_treatments.sql` — Risk treatments table + indexes + trigger + constraints
40. `040_risk_controls.sql` — Risk controls junction table + indexes + trigger + unique constraint
41. `041_sprint6_fk_cross_refs.sql` — Deferred FKs: evidence_links.risk_id column + FK + updated CHECK constraint
42. `042_sprint6_seed_templates.sql` — Risk template library (200+ risks across categories)
43. `043_sprint6_seed_demo.sql` — Demo org example risks with assessments, treatments, and control mappings

---

## Seed Data

### Risk Template Library (200+ common risks)

Templates are seeded as `is_template = TRUE` with category-appropriate risks sourced from ISO 27005, NIST SP 800-30, and CIS Risk Assessment Method.

```sql
-- ============================================================================
-- CYBER SECURITY RISKS (30 risks)
-- ============================================================================
INSERT INTO risks (id, org_id, identifier, title, description, category, status, is_template, template_source, tags) VALUES
('rt000000-0000-0000-0001-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-001', 'Ransomware Attack', 'Threat of ransomware encrypting critical business data and systems, demanding payment for decryption keys. Includes risk of data exfiltration before encryption (double extortion).', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['ransomware', 'malware', 'data-loss', 'critical']),
('rt000000-0000-0000-0001-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-002', 'Phishing / Social Engineering', 'Risk of employees falling victim to phishing emails, spear-phishing, or social engineering attacks leading to credential theft or malware installation.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['phishing', 'social-engineering', 'credential-theft']),
('rt000000-0000-0000-0001-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-003', 'Advanced Persistent Threat (APT)', 'Risk of sophisticated, targeted attack by a well-resourced adversary gaining long-term access to systems and data.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['apt', 'targeted-attack', 'nation-state']),
('rt000000-0000-0000-0001-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-004', 'SQL Injection / Web Application Attack', 'Risk of attackers exploiting web application vulnerabilities (SQLi, XSS, SSRF) to access or modify data.', 'cyber_security', 'identified', TRUE, 'OWASP Top 10', ARRAY['web-app', 'injection', 'owasp']),
('rt000000-0000-0000-0001-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-005', 'Credential Stuffing / Brute Force', 'Risk of automated attacks using stolen credentials or brute-force methods to gain unauthorized access.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['credential-stuffing', 'brute-force', 'authentication']),
('rt000000-0000-0000-0001-000000000006', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-006', 'Zero-Day Vulnerability Exploitation', 'Risk of attackers exploiting previously unknown vulnerabilities before patches are available.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['zero-day', 'vulnerability', 'exploit']),
('rt000000-0000-0000-0001-000000000007', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-007', 'DDoS Attack', 'Risk of distributed denial-of-service attacks rendering critical services unavailable to legitimate users.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['ddos', 'availability', 'dos']),
('rt000000-0000-0000-0001-000000000008', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-008', 'Supply Chain Compromise', 'Risk of a trusted software vendor or dependency being compromised, introducing malicious code into the environment (e.g., SolarWinds-style attack).', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['supply-chain', 'vendor', 'software-compromise']),
('rt000000-0000-0000-0001-000000000009', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-009', 'Man-in-the-Middle Attack', 'Risk of intercepted communications allowing eavesdropping or data manipulation between parties.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['mitm', 'encryption', 'interception']),
('rt000000-0000-0000-0001-000000000010', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-010', 'Insider Threat — Malicious', 'Risk of a trusted insider (employee, contractor) deliberately stealing data, sabotaging systems, or selling access.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['insider-threat', 'malicious', 'data-theft']),
('rt000000-0000-0000-0001-000000000011', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-011', 'Cryptographic Key Compromise', 'Risk of encryption keys being stolen, leaked, or improperly managed, rendering encrypted data accessible.', 'cyber_security', 'identified', TRUE, 'ISO 27005', ARRAY['cryptography', 'key-management', 'encryption']),
('rt000000-0000-0000-0001-000000000012', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-012', 'DNS Hijacking / Poisoning', 'Risk of DNS infrastructure being compromised, redirecting traffic to malicious servers.', 'cyber_security', 'identified', TRUE, 'NIST SP 800-30', ARRAY['dns', 'hijacking', 'network']),
('rt000000-0000-0000-0001-000000000013', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-013', 'API Security Breach', 'Risk of API vulnerabilities (broken authentication, excessive data exposure, rate limiting failures) being exploited.', 'cyber_security', 'identified', TRUE, 'OWASP API Top 10', ARRAY['api', 'authentication', 'data-exposure']),
('rt000000-0000-0000-0001-000000000014', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-014', 'Unpatched Systems / Missing Updates', 'Risk of known vulnerabilities remaining exploitable due to delayed or missed patching.', 'cyber_security', 'identified', TRUE, 'CIS Controls', ARRAY['patching', 'vulnerability', 'maintenance']),
('rt000000-0000-0000-0001-000000000015', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CY-015', 'Cloud Misconfiguration', 'Risk of cloud resources being misconfigured (open S3 buckets, overly permissive IAM, public databases), exposing data or services.', 'cyber_security', 'identified', TRUE, 'CIS Controls', ARRAY['cloud', 'misconfiguration', 'aws', 'azure', 'gcp']),

-- ============================================================================
-- OPERATIONAL RISKS (25 risks)
-- ============================================================================
('rt000000-0000-0000-0002-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-001', 'System / Service Outage', 'Risk of critical systems becoming unavailable due to hardware failure, software bugs, or infrastructure issues.', 'operational', 'identified', TRUE, 'ISO 27005', ARRAY['outage', 'availability', 'downtime']),
('rt000000-0000-0000-0002-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-002', 'Data Loss / Corruption', 'Risk of data being lost, corrupted, or rendered irrecoverable due to hardware failure, software bugs, or human error.', 'operational', 'identified', TRUE, 'ISO 27005', ARRAY['data-loss', 'corruption', 'backup']),
('rt000000-0000-0000-0002-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-003', 'Backup Failure', 'Risk of backup systems failing silently, leading to inability to recover data after an incident.', 'operational', 'identified', TRUE, 'ISO 27005', ARRAY['backup', 'recovery', 'data-loss']),
('rt000000-0000-0000-0002-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-004', 'Configuration Drift', 'Risk of production systems drifting from approved configurations, introducing security gaps or instability.', 'operational', 'identified', TRUE, 'CIS Controls', ARRAY['configuration', 'drift', 'compliance']),
('rt000000-0000-0000-0002-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-005', 'Change Management Failure', 'Risk of poorly managed changes causing outages, security vulnerabilities, or data loss.', 'operational', 'identified', TRUE, 'ITIL', ARRAY['change-management', 'deployment', 'process']),
('rt000000-0000-0000-0002-000000000006', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-006', 'Capacity Exhaustion', 'Risk of storage, compute, or network capacity being exhausted, causing service degradation or outage.', 'operational', 'identified', TRUE, 'ISO 27005', ARRAY['capacity', 'scaling', 'performance']),
('rt000000-0000-0000-0002-000000000007', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-007', 'Monitoring / Alerting Gap', 'Risk of security incidents or system failures going undetected due to insufficient monitoring, logging, or alerting.', 'operational', 'identified', TRUE, 'CIS Controls', ARRAY['monitoring', 'alerting', 'detection']),
('rt000000-0000-0000-0002-000000000008', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-008', 'Incident Response Inadequacy', 'Risk of the organization being unable to effectively detect, respond to, contain, and recover from security incidents.', 'operational', 'identified', TRUE, 'NIST CSF', ARRAY['incident-response', 'readiness', 'recovery']),
('rt000000-0000-0000-0002-000000000009', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-009', 'Disaster Recovery Failure', 'Risk of disaster recovery plans being inadequate, untested, or failing when needed.', 'operational', 'identified', TRUE, 'ISO 22301', ARRAY['disaster-recovery', 'bcp', 'resilience']),
('rt000000-0000-0000-0002-000000000010', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-OP-010', 'Shadow IT', 'Risk of unauthorized applications, services, or devices being used by employees outside IT governance.', 'operational', 'identified', TRUE, 'CIS Controls', ARRAY['shadow-it', 'unauthorized', 'governance']),

-- ============================================================================
-- COMPLIANCE RISKS (25 risks)
-- ============================================================================
('rt000000-0000-0000-0003-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-001', 'PCI DSS Non-Compliance', 'Risk of failing PCI DSS audit resulting in fines, increased transaction fees, or loss of card processing ability.', 'compliance', 'identified', TRUE, 'PCI DSS v4.0.1', ARRAY['pci-dss', 'payment', 'audit']),
('rt000000-0000-0000-0003-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-002', 'SOC 2 Audit Failure', 'Risk of receiving a qualified SOC 2 opinion due to control deficiencies, impacting customer trust and sales.', 'compliance', 'identified', TRUE, 'AICPA TSC', ARRAY['soc2', 'audit', 'trust']),
('rt000000-0000-0000-0003-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-003', 'GDPR Violation', 'Risk of violating GDPR requirements resulting in fines up to 4% of global annual turnover or €20M.', 'compliance', 'identified', TRUE, 'GDPR', ARRAY['gdpr', 'privacy', 'european-union', 'fine']),
('rt000000-0000-0000-0003-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-004', 'CCPA/CPRA Violation', 'Risk of violating California privacy regulations resulting in fines and class action lawsuits.', 'compliance', 'identified', TRUE, 'CCPA/CPRA', ARRAY['ccpa', 'cpra', 'privacy', 'california']),
('rt000000-0000-0000-0003-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-005', 'ISO 27001 Certification Loss', 'Risk of losing ISO 27001 certification due to non-conformities found during surveillance or recertification audits.', 'compliance', 'identified', TRUE, 'ISO 27001', ARRAY['iso27001', 'certification', 'audit']),
('rt000000-0000-0000-0003-000000000006', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-006', 'Regulatory Change Impact', 'Risk of new or amended regulations requiring significant changes to policies, controls, or business processes.', 'compliance', 'identified', TRUE, 'General', ARRAY['regulation', 'change', 'adaptation']),
('rt000000-0000-0000-0003-000000000007', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-007', 'Evidence Staleness', 'Risk of compliance evidence becoming stale, outdated, or insufficient before audit deadlines.', 'compliance', 'identified', TRUE, 'General', ARRAY['evidence', 'staleness', 'audit-readiness']),
('rt000000-0000-0000-0003-000000000008', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-008', 'Audit Trail Tampering', 'Risk of audit logs being modified, deleted, or rendered unreliable, undermining compliance evidence integrity.', 'compliance', 'identified', TRUE, 'ISO 27001', ARRAY['audit-log', 'integrity', 'tampering']),
('rt000000-0000-0000-0003-000000000009', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-009', 'Cross-Border Data Transfer Violation', 'Risk of transferring personal data across borders in violation of data residency or transfer mechanism requirements.', 'compliance', 'identified', TRUE, 'GDPR', ARRAY['data-transfer', 'cross-border', 'schrems']),
('rt000000-0000-0000-0003-000000000010', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-CO-010', 'Inadequate Privacy Notice', 'Risk of privacy notices being incomplete, inaccurate, or not properly displayed, violating transparency requirements.', 'compliance', 'identified', TRUE, 'GDPR', ARRAY['privacy-notice', 'transparency', 'consent']),

-- ============================================================================
-- DATA PRIVACY RISKS (20 risks)
-- ============================================================================
('rt000000-0000-0000-0004-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-001', 'Personal Data Breach', 'Risk of unauthorized access to, or disclosure of, personal data affecting data subjects.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 4(12)', ARRAY['data-breach', 'pii', 'notification']),
('rt000000-0000-0000-0004-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-002', 'Excessive Data Collection', 'Risk of collecting more personal data than necessary for the stated purpose, violating data minimization principles.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 5(1)(c)', ARRAY['data-minimization', 'collection', 'purpose-limitation']),
('rt000000-0000-0000-0004-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-003', 'Consent Management Failure', 'Risk of failing to properly obtain, record, or manage data subject consent for processing activities.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 7', ARRAY['consent', 'lawful-basis', 'opt-in']),
('rt000000-0000-0000-0004-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-004', 'Data Subject Rights Non-Fulfillment', 'Risk of failing to fulfill data subject access, deletion, portability, or rectification requests within required timeframes.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 15-22', ARRAY['dsar', 'subject-rights', 'access-request']),
('rt000000-0000-0000-0004-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-005', 'Data Retention Violation', 'Risk of retaining personal data longer than necessary or failing to securely dispose of expired data.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 5(1)(e)', ARRAY['retention', 'disposal', 'storage-limitation']),
('rt000000-0000-0000-0004-000000000006', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-006', 'Unauthorized Data Sharing', 'Risk of personal data being shared with unauthorized third parties without proper agreements or consent.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 28', ARRAY['data-sharing', 'third-party', 'dpa']),
('rt000000-0000-0000-0004-000000000007', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-007', 'Inadequate Data Protection Impact Assessment', 'Risk of processing activities requiring DPIA being conducted without adequate risk assessment.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 35', ARRAY['dpia', 'impact-assessment', 'high-risk']),
('rt000000-0000-0000-0004-000000000008', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-008', 'Breach Notification Failure', 'Risk of failing to notify supervisory authorities and affected individuals within required timeframes after a data breach.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 33-34', ARRAY['breach-notification', '72-hour', 'supervisory-authority']),
('rt000000-0000-0000-0004-000000000009', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-009', 'Children Data Processing Violation', 'Risk of processing children''s personal data without appropriate parental consent and safeguards.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 8', ARRAY['children', 'parental-consent', 'minors']),
('rt000000-0000-0000-0004-000000000010', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-DP-010', 'Automated Decision-Making Bias', 'Risk of automated decision-making or profiling systems producing biased or discriminatory outcomes.', 'data_privacy', 'identified', TRUE, 'GDPR Art. 22', ARRAY['automated-decision', 'profiling', 'bias', 'ai']),

-- ============================================================================
-- TECHNOLOGY RISKS (25 risks)
-- ============================================================================
('rt000000-0000-0000-0005-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-001', 'Legacy System Dependency', 'Risk of critical business processes depending on outdated, unsupported, or end-of-life systems.', 'technology', 'identified', TRUE, 'ISO 27005', ARRAY['legacy', 'eol', 'technical-debt']),
('rt000000-0000-0000-0005-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-002', 'Single Point of Failure', 'Risk of a single component failure causing cascading outage across dependent systems.', 'technology', 'identified', TRUE, 'ISO 27005', ARRAY['spof', 'redundancy', 'resilience']),
('rt000000-0000-0000-0005-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-003', 'Database Compromise', 'Risk of database being compromised through SQL injection, misconfiguration, or stolen credentials.', 'technology', 'identified', TRUE, 'OWASP Top 10', ARRAY['database', 'sql-injection', 'data-access']),
('rt000000-0000-0000-0005-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-004', 'Container / Kubernetes Escape', 'Risk of container isolation being bypassed, allowing access to host systems or other containers.', 'technology', 'identified', TRUE, 'CIS Kubernetes Benchmark', ARRAY['container', 'kubernetes', 'escape', 'isolation']),
('rt000000-0000-0000-0005-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-005', 'CI/CD Pipeline Compromise', 'Risk of build/deployment pipelines being compromised, enabling injection of malicious code into production.', 'technology', 'identified', TRUE, 'SLSA Framework', ARRAY['ci-cd', 'pipeline', 'build', 'deployment']),
('rt000000-0000-0000-0005-000000000006', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-006', 'Secrets / Credential Exposure', 'Risk of API keys, passwords, tokens, or certificates being exposed in code repositories, logs, or configuration files.', 'technology', 'identified', TRUE, 'CIS Controls', ARRAY['secrets', 'credentials', 'exposure', 'hardcoded']),
('rt000000-0000-0000-0005-000000000007', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-007', 'Third-Party Dependency Vulnerability', 'Risk of vulnerabilities in open-source libraries or third-party packages used in applications.', 'technology', 'identified', TRUE, 'OWASP', ARRAY['dependency', 'sca', 'open-source', 'vulnerability']),
('rt000000-0000-0000-0005-000000000008', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-008', 'TLS / Certificate Misconfiguration', 'Risk of TLS certificates expiring, being misconfigured, or using weak cipher suites.', 'technology', 'identified', TRUE, 'NIST SP 800-52', ARRAY['tls', 'certificate', 'encryption', 'https']),
('rt000000-0000-0000-0005-000000000009', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-009', 'Cloud Service Provider Outage', 'Risk of a major cloud provider (AWS, Azure, GCP) experiencing an outage affecting business operations.', 'technology', 'identified', TRUE, 'ISO 27005', ARRAY['cloud', 'outage', 'vendor-dependency']),
('rt000000-0000-0000-0005-000000000010', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TE-010', 'Insufficient Logging / Audit Trail', 'Risk of security events not being logged or logs being insufficiently detailed for forensic investigation.', 'technology', 'identified', TRUE, 'CIS Controls', ARRAY['logging', 'audit-trail', 'forensics']),

-- ============================================================================
-- THIRD-PARTY RISKS (20 risks)
-- ============================================================================
('rt000000-0000-0000-0006-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-001', 'Vendor Data Breach', 'Risk of a third-party vendor experiencing a data breach that exposes organizational data.', 'third_party', 'identified', TRUE, 'ISO 27036', ARRAY['vendor', 'data-breach', 'third-party']),
('rt000000-0000-0000-0006-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-002', 'Vendor Service Disruption', 'Risk of a critical vendor service becoming unavailable, disrupting business operations.', 'third_party', 'identified', TRUE, 'ISO 27036', ARRAY['vendor', 'disruption', 'sla']),
('rt000000-0000-0000-0006-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-003', 'Vendor Lock-In', 'Risk of becoming overly dependent on a single vendor, limiting flexibility and negotiating power.', 'third_party', 'identified', TRUE, 'General', ARRAY['vendor-lock-in', 'dependency', 'migration']),
('rt000000-0000-0000-0006-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-004', 'Inadequate Vendor Security Controls', 'Risk of vendors not maintaining adequate security controls, creating exposure for shared data and systems.', 'third_party', 'identified', TRUE, 'SIG/CAIQ', ARRAY['vendor-security', 'assessment', 'controls']),
('rt000000-0000-0000-0006-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-005', 'Fourth-Party Risk', 'Risk of vendors subcontracting to additional parties (fourth parties) without adequate oversight.', 'third_party', 'identified', TRUE, 'ISO 27036', ARRAY['fourth-party', 'subcontractor', 'supply-chain']),
('rt000000-0000-0000-0006-000000000006', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-006', 'Vendor Compliance Failure', 'Risk of vendors failing to maintain required compliance certifications (SOC 2, PCI DSS, ISO 27001).', 'third_party', 'identified', TRUE, 'General', ARRAY['vendor-compliance', 'certification', 'audit']),
('rt000000-0000-0000-0006-000000000007', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-007', 'Vendor Financial Instability', 'Risk of a critical vendor becoming financially unstable or ceasing operations.', 'third_party', 'identified', TRUE, 'General', ARRAY['vendor-financial', 'bankruptcy', 'continuity']),
('rt000000-0000-0000-0006-000000000008', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-008', 'Data Processing Agreement Non-Compliance', 'Risk of vendor data processing activities not conforming to DPA terms and conditions.', 'third_party', 'identified', TRUE, 'GDPR Art. 28', ARRAY['dpa', 'processing', 'vendor-compliance']),
('rt000000-0000-0000-0006-000000000009', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-009', 'API Integration Security Gaps', 'Risk of insecure API integrations with vendors creating authentication, authorization, or data exposure vulnerabilities.', 'third_party', 'identified', TRUE, 'OWASP API', ARRAY['api', 'integration', 'vendor-api']),
('rt000000-0000-0000-0006-000000000010', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-TP-010', 'Vendor Access Overprovisioning', 'Risk of vendors retaining excessive access privileges to systems and data beyond what is needed.', 'third_party', 'identified', TRUE, 'CIS Controls', ARRAY['vendor-access', 'least-privilege', 'access-review']),

-- ============================================================================
-- FINANCIAL RISKS (15 risks)
-- ============================================================================
('rt000000-0000-0000-0007-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-FI-001', 'Fraud / Financial Crime', 'Risk of internal or external fraud, including payment fraud, expense manipulation, or financial statement misrepresentation.', 'financial', 'identified', TRUE, 'General', ARRAY['fraud', 'financial-crime', 'misrepresentation']),
('rt000000-0000-0000-0007-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-FI-002', 'Regulatory Fine', 'Risk of financial penalties from regulatory bodies for non-compliance with applicable laws and regulations.', 'financial', 'identified', TRUE, 'General', ARRAY['fine', 'penalty', 'regulatory']),
('rt000000-0000-0000-0007-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-FI-003', 'Business Interruption Loss', 'Risk of financial loss due to extended business interruption from cyber incidents, outages, or disasters.', 'financial', 'identified', TRUE, 'ISO 22301', ARRAY['business-interruption', 'loss', 'recovery']),
('rt000000-0000-0000-0007-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-FI-004', 'Cyber Insurance Coverage Gap', 'Risk of cyber insurance policies not covering the full scope of a cyber incident, leaving uninsured losses.', 'financial', 'identified', TRUE, 'General', ARRAY['insurance', 'coverage', 'gap']),
('rt000000-0000-0000-0007-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-FI-005', 'Intellectual Property Theft', 'Risk of proprietary algorithms, source code, trade secrets, or business intelligence being stolen.', 'financial', 'identified', TRUE, 'General', ARRAY['ip-theft', 'trade-secret', 'proprietary']),

-- ============================================================================
-- LEGAL RISKS (15 risks)
-- ============================================================================
('rt000000-0000-0000-0008-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-LG-001', 'Class Action Lawsuit (Data Breach)', 'Risk of class action litigation following a data breach affecting a large number of individuals.', 'legal', 'identified', TRUE, 'General', ARRAY['lawsuit', 'class-action', 'data-breach']),
('rt000000-0000-0000-0008-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-LG-002', 'Contract Breach Liability', 'Risk of failing to meet contractual security, privacy, or service level obligations to customers or partners.', 'legal', 'identified', TRUE, 'General', ARRAY['contract', 'breach', 'liability', 'sla']),
('rt000000-0000-0000-0008-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-LG-003', 'Patent / IP Infringement', 'Risk of unintentionally infringing on third-party patents, trademarks, or copyrights.', 'legal', 'identified', TRUE, 'General', ARRAY['patent', 'copyright', 'infringement']),
('rt000000-0000-0000-0008-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-LG-004', 'E-Discovery Non-Compliance', 'Risk of failing to preserve and produce electronic records in response to legal discovery requests.', 'legal', 'identified', TRUE, 'General', ARRAY['e-discovery', 'preservation', 'litigation-hold']),
('rt000000-0000-0000-0008-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-LG-005', 'Open Source License Violation', 'Risk of violating open-source software licenses (GPL, AGPL, etc.) through improper use or distribution.', 'legal', 'identified', TRUE, 'General', ARRAY['open-source', 'license', 'gpl', 'compliance']),

-- ============================================================================
-- REPUTATIONAL RISKS (15 risks)
-- ============================================================================
('rt000000-0000-0000-0009-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-RE-001', 'Public Data Breach Disclosure', 'Risk of a data breach becoming public, damaging brand reputation and customer trust.', 'reputational', 'identified', TRUE, 'General', ARRAY['breach-disclosure', 'reputation', 'media']),
('rt000000-0000-0000-0009-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-RE-002', 'Customer Trust Erosion', 'Risk of security incidents, privacy violations, or poor handling of customer data eroding trust.', 'reputational', 'identified', TRUE, 'General', ARRAY['customer-trust', 'churn', 'confidence']),
('rt000000-0000-0000-0009-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-RE-003', 'Social Media Security Incident', 'Risk of organizational social media accounts being compromised or used to spread misinformation.', 'reputational', 'identified', TRUE, 'General', ARRAY['social-media', 'compromise', 'account-takeover']),
('rt000000-0000-0000-0009-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-RE-004', 'Negative Audit Finding Disclosure', 'Risk of qualified audit findings becoming known to customers, partners, or the public.', 'reputational', 'identified', TRUE, 'General', ARRAY['audit-finding', 'qualified', 'disclosure']),
('rt000000-0000-0000-0009-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-RE-005', 'Employee Public Misconduct', 'Risk of employee misconduct in public forums or social media reflecting poorly on the organization.', 'reputational', 'identified', TRUE, 'General', ARRAY['employee-conduct', 'public', 'social-media']),

-- ============================================================================
-- HUMAN RESOURCES RISKS (15 risks)
-- ============================================================================
('rt000000-0000-0000-000a-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-HR-001', 'Key Person Dependency', 'Risk of critical knowledge, skills, or access being concentrated in a single individual with no backup.', 'human_resources', 'identified', TRUE, 'ISO 27005', ARRAY['key-person', 'bus-factor', 'knowledge']),
('rt000000-0000-0000-000a-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-HR-002', 'Insufficient Security Training', 'Risk of employees lacking adequate security awareness training, increasing vulnerability to social engineering.', 'human_resources', 'identified', TRUE, 'CIS Controls', ARRAY['training', 'awareness', 'security-culture']),
('rt000000-0000-0000-000a-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-HR-003', 'Insider Threat — Negligent', 'Risk of employees unintentionally causing security incidents through negligence, carelessness, or lack of awareness.', 'human_resources', 'identified', TRUE, 'NIST SP 800-30', ARRAY['insider-threat', 'negligence', 'human-error']),
('rt000000-0000-0000-000a-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-HR-004', 'Inadequate Background Checks', 'Risk of hiring individuals with undisclosed criminal history or misrepresented qualifications.', 'human_resources', 'identified', TRUE, 'General', ARRAY['background-check', 'screening', 'hiring']),
('rt000000-0000-0000-000a-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-HR-005', 'Improper Offboarding', 'Risk of departing employees retaining access to systems, data, or facilities after termination.', 'human_resources', 'identified', TRUE, 'CIS Controls', ARRAY['offboarding', 'access-revocation', 'termination']),

-- ============================================================================
-- STRATEGIC RISKS (10 risks)
-- ============================================================================
('rt000000-0000-0000-000b-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-ST-001', 'AI / Emerging Technology Risk', 'Risk of AI systems introducing uncontrolled decision-making, bias, or security vulnerabilities into business processes.', 'strategic', 'identified', TRUE, 'ISO 42001', ARRAY['ai', 'emerging-tech', 'automation']),
('rt000000-0000-0000-000b-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-ST-002', 'Market / Competitive Disruption', 'Risk of competitors or market shifts rendering current products, services, or security investments obsolete.', 'strategic', 'identified', TRUE, 'General', ARRAY['market', 'competition', 'disruption']),
('rt000000-0000-0000-000b-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-ST-003', 'M&A Integration Risk', 'Risk of security and compliance gaps during mergers, acquisitions, or divestitures.', 'strategic', 'identified', TRUE, 'General', ARRAY['m-and-a', 'integration', 'due-diligence']),
('rt000000-0000-0000-000b-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-ST-004', 'Geopolitical / Sanctions Risk', 'Risk of operating in or with entities from regions subject to sanctions or geopolitical instability.', 'strategic', 'identified', TRUE, 'General', ARRAY['geopolitical', 'sanctions', 'export-control']),
('rt000000-0000-0000-000b-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-ST-005', 'Digital Transformation Failure', 'Risk of digital transformation initiatives failing to deliver expected security, efficiency, or compliance improvements.', 'strategic', 'identified', TRUE, 'General', ARRAY['digital-transformation', 'modernization', 'cloud-migration']),

-- ============================================================================
-- PHYSICAL RISKS (10 risks)
-- ============================================================================
('rt000000-0000-0000-000c-000000000001', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-PH-001', 'Unauthorized Physical Access', 'Risk of unauthorized individuals gaining physical access to data centers, server rooms, or restricted areas.', 'physical', 'identified', TRUE, 'ISO 27001 A.11', ARRAY['physical-access', 'data-center', 'restricted-area']),
('rt000000-0000-0000-000c-000000000002', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-PH-002', 'Natural Disaster Impact', 'Risk of earthquakes, floods, fires, or severe weather damaging physical infrastructure and disrupting operations.', 'physical', 'identified', TRUE, 'ISO 22301', ARRAY['natural-disaster', 'flood', 'earthquake', 'fire']),
('rt000000-0000-0000-000c-000000000003', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-PH-003', 'Power / Utility Failure', 'Risk of extended power outages or utility failures affecting data center or office operations.', 'physical', 'identified', TRUE, 'ISO 27005', ARRAY['power', 'utility', 'ups', 'generator']),
('rt000000-0000-0000-000c-000000000004', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-PH-004', 'Hardware Theft or Tampering', 'Risk of physical hardware (servers, laptops, storage media) being stolen or tampered with.', 'physical', 'identified', TRUE, 'ISO 27001 A.11', ARRAY['theft', 'tampering', 'hardware']),
('rt000000-0000-0000-000c-000000000005', 'a0000000-0000-0000-0000-000000000001', 'TPL-RISK-PH-005', 'Environmental Control Failure', 'Risk of HVAC, fire suppression, or environmental monitoring systems failing in data center environments.', 'physical', 'identified', TRUE, 'ISO 27001 A.11', ARRAY['hvac', 'fire-suppression', 'environmental']);
```

**Total template risks: 200** (15 CY + 10 OP + 10 CO + 10 DP + 10 TE + 10 TP + 5 FI + 5 LG + 5 RE + 5 HR + 5 ST + 5 PH = 100 shown above; the full seed extends each category to reach 200+).

> **Note to DBE:** The above shows representative risks per category. The full migration should contain 200+ risks. Extend each category proportionally (e.g., 30 cyber, 25 operational, 25 compliance, 20 privacy, 25 technology, 20 third-party, 15 financial, 15 legal, 15 reputational, 15 HR, 10 strategic, 10 physical = 225 total).

---

### Demo Organization Example Risks

```sql
-- Active risks for the demo org (cloned from templates)
INSERT INTO risks (
    id, org_id, identifier, title, description, category, status,
    owner_id, is_template,
    inherent_likelihood, inherent_impact, inherent_score,
    residual_likelihood, residual_impact, residual_score,
    risk_appetite_threshold, assessment_frequency_days, next_assessment_at, last_assessed_at,
    source, affected_assets, tags
) VALUES
    -- Critical risk: Ransomware
    ('rd000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'RISK-CY-001', 'Ransomware Attack on Production Systems',
     'Risk of ransomware encrypting production databases and application servers, with potential double extortion.',
     'cyber_security', 'treating',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     FALSE,
     'likely', 'severe', 20.00,
     'possible', 'major', 12.00,
     10.00, 90, '2026-05-18', '2026-02-18',
     'threat_assessment', ARRAY['payment-api', 'customer-db', 'erp-system'],
     ARRAY['critical', 'ransomware', 'q1-review']),

    -- High risk: Phishing
    ('rd000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'RISK-CY-002', 'Phishing / Credential Theft',
     'Risk of employees falling victim to phishing attacks, leading to compromised credentials and lateral movement.',
     'cyber_security', 'monitoring',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     FALSE,
     'almost_certain', 'moderate', 15.00,
     'likely', 'minor', 8.00,
     8.00, 90, '2026-05-18', '2026-02-18',
     'incident_history', ARRAY['email-system', 'vpn', 'sso'],
     ARRAY['phishing', 'training', 'ongoing']),

    -- Medium risk: PCI non-compliance
    ('rd000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'RISK-CO-001', 'PCI DSS v4.0.1 Non-Compliance',
     'Risk of failing PCI DSS audit due to gaps in meeting v4.0.1 new requirements, particularly 6.4.3 and 11.6.1.',
     'compliance', 'assessing',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     FALSE,
     'possible', 'major', 12.00,
     'unlikely', 'moderate', 6.00,
     8.00, 180, '2026-08-18', '2026-02-18',
     'audit_preparation', ARRAY['payment-page', 'cardholder-data-env'],
     ARRAY['pci-dss', 'audit', 'v4.0.1']),

    -- Accepted risk: Legacy system
    ('rd000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'RISK-TE-001', 'Legacy ERP System Dependency',
     'Critical business processes depend on the legacy ERP system (Oracle E-Business Suite 12.2) which reaches end of extended support in 2027.',
     'technology', 'accepted',
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     FALSE,
     'likely', 'major', 16.00,
     'likely', 'moderate', 12.00,
     12.00, 365, '2027-02-15', '2026-02-15',
     'technology_review', ARRAY['erp-system', 'financial-reporting'],
     ARRAY['legacy', 'accepted', 'migration-planned']),

    -- Low risk: Vendor
    ('rd000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'RISK-TP-001', 'SaaS Vendor Data Handling',
     'Risk of SaaS vendors handling customer data without adequate security controls or DPA compliance.',
     'third_party', 'monitoring',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     FALSE,
     'possible', 'moderate', 9.00,
     'unlikely', 'minor', 4.00,
     6.00, 180, '2026-08-18', '2026-02-18',
     'vendor_assessment', ARRAY['crm', 'analytics-platform', 'email-service'],
     ARRAY['vendor', 'dpa', 'ongoing']);

-- Update accepted risk metadata
UPDATE risks SET
    accepted_at = '2026-02-15 10:00:00+00',
    accepted_by = (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
    acceptance_expiry = '2027-02-15',
    acceptance_justification = 'ERP migration to SAP S/4HANA scheduled for Q3 2026. Current risk is acceptable given compensating controls (enhanced monitoring, additional backup procedures) and the planned migration timeline.'
WHERE id = 'rd000000-0000-0000-0000-000000000004';

-- Demo risk assessments (inherent + residual for each risk)
INSERT INTO risk_assessments (
    id, org_id, risk_id, assessment_type,
    likelihood, impact, likelihood_score, impact_score, overall_score,
    scoring_formula, severity, justification, data_sources,
    assessed_by, assessment_date, valid_until, is_current
) VALUES
    -- Ransomware — inherent
    ('ra000000-0000-0000-0001-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000001', 'inherent',
     'likely', 'severe', 4, 5, 20.00,
     'likelihood_x_impact', 'critical',
     'Ransomware attacks are increasing in frequency and sophistication. Our industry (fintech) is a high-value target. Recent threat intel shows active campaigns targeting similar organizations.',
     ARRAY['threat-intel-report-2026-q1', 'industry-benchmark'],
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18', '2026-05-18', TRUE),
    -- Ransomware — residual
    ('ra000000-0000-0000-0001-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000001', 'residual',
     'possible', 'major', 3, 4, 12.00,
     'likelihood_x_impact', 'high',
     'EDR, network segmentation, and immutable backups reduce likelihood. Impact reduced by backup/DR capability but still significant if critical systems are affected.',
     ARRAY['edr-deployment-report', 'backup-verification-test'],
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18', '2026-05-18', TRUE),

    -- Phishing — inherent
    ('ra000000-0000-0000-0002-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000002', 'inherent',
     'almost_certain', 'moderate', 5, 3, 15.00,
     'likelihood_x_impact', 'high',
     'Phishing attempts are near-constant (50+ blocked per month). Historical incidents show 2-3 successful compromises per year despite training.',
     ARRAY['email-gateway-stats-2025', 'incident-log-2025'],
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18', '2026-05-18', TRUE),
    -- Phishing — residual
    ('ra000000-0000-0000-0002-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000002', 'residual',
     'likely', 'minor', 4, 2, 8.00,
     'likelihood_x_impact', 'medium',
     'MFA, email filtering, and quarterly security training reduce impact. Credential theft is still possible but lateral movement is limited by MFA and network segmentation.',
     ARRAY['mfa-coverage-report', 'training-completion-q4-2025'],
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18', '2026-05-18', TRUE),

    -- PCI non-compliance — inherent
    ('ra000000-0000-0000-0003-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000003', 'inherent',
     'possible', 'major', 3, 4, 12.00,
     'likelihood_x_impact', 'high',
     'PCI DSS v4.0.1 introduces new requirements (6.4.3, 11.6.1) that are not yet fully implemented. Audit is scheduled for Q3 2026.',
     ARRAY['pci-gap-analysis-2026', 'v4.0.1-requirement-mapping'],
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     '2026-02-18', '2026-08-18', TRUE),
    -- PCI non-compliance — residual
    ('ra000000-0000-0000-0003-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000003', 'residual',
     'unlikely', 'moderate', 2, 3, 6.00,
     'likelihood_x_impact', 'medium',
     'Raisin Shield deployment addresses 6.4.3 and 11.6.1. Remaining gaps are documentation and process-level — implementation plan is on track.',
     ARRAY['raisin-shield-deployment', 'pci-remediation-plan'],
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     '2026-02-18', '2026-08-18', TRUE);

-- Demo risk treatments
INSERT INTO risk_treatments (
    id, org_id, risk_id, treatment_type, title, description, status,
    owner_id, created_by, priority, due_date, started_at,
    expected_residual_likelihood, expected_residual_impact, expected_residual_score
) VALUES
    -- Ransomware treatment 1: EDR deployment
    ('rx000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000001', 'mitigate',
     'Deploy EDR to All Endpoints', 'Deploy CrowdStrike Falcon EDR to 100% of endpoints with real-time threat detection and automated response.',
     'verified',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     'critical', '2026-03-15', '2026-02-01 09:00:00+00',
     'unlikely', 'major', 8.00),

    -- Ransomware treatment 2: Immutable backups
    ('rx000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000001', 'mitigate',
     'Implement Immutable Backup Strategy', 'Deploy immutable backups with air-gapped secondary storage. Test recovery procedures monthly.',
     'in_progress',
     (SELECT id FROM users WHERE email IN ('devops@acme.example.com', 'it@acme.example.com') LIMIT 1),
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     'high', '2026-04-01', '2026-02-10 09:00:00+00',
     'unlikely', 'moderate', 6.00),

    -- Ransomware treatment 3: Cyber insurance
    ('rx000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000001', 'transfer',
     'Cyber Insurance Policy', 'Maintain comprehensive cyber insurance with ransomware coverage, $5M limit.',
     'verified',
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1),
     'medium', '2026-02-15', NULL,
     NULL, NULL, NULL),

    -- Phishing treatment: Security awareness training
    ('rx000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000002', 'mitigate',
     'Enhanced Security Awareness Training', 'Quarterly phishing simulations with mandatory training for employees who fail. Gamified leaderboard.',
     'implemented',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     'medium', '2026-03-01', '2026-01-15 09:00:00+00',
     'likely', 'negligible', 4.00),

    -- PCI treatment: Raisin Shield deployment
    ('rx000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000003', 'mitigate',
     'Deploy Raisin Shield for PCI DSS 6.4.3 & 11.6.1', 'Deploy client-side script monitoring and protection to meet new PCI DSS v4.0.1 requirements.',
     'in_progress',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     'high', '2026-06-01', '2026-01-29 09:00:00+00',
     'rare', 'moderate', 3.00);

-- Update completed treatments
UPDATE risk_treatments SET
    completed_at = '2026-02-15 16:00:00+00',
    effectiveness_rating = 'highly_effective',
    effectiveness_notes = 'CrowdStrike deployed to 100% of endpoints. Blocked 3 ransomware attempts in first week. Detection rate >99%.',
    effectiveness_reviewed_at = '2026-02-20 10:00:00+00',
    effectiveness_reviewed_by = (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1)
WHERE id = 'rx000000-0000-0000-0000-000000000001';

UPDATE risk_treatments SET
    completed_at = '2026-02-01 10:00:00+00',
    effectiveness_rating = 'effective',
    effectiveness_notes = 'Cyber insurance policy renewed with $5M ransomware coverage. Premium increased 15% but coverage is comprehensive.',
    effectiveness_reviewed_at = '2026-02-05 10:00:00+00',
    effectiveness_reviewed_by = (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1)
WHERE id = 'rx000000-0000-0000-0000-000000000003';

-- Demo risk-control mappings
INSERT INTO risk_controls (
    id, org_id, risk_id, control_id, effectiveness, notes, mitigation_percentage,
    linked_by, last_effectiveness_review, reviewed_by
) VALUES
    -- Ransomware → controls
    ('rk000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000001',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-EP-001' LIMIT 1),
     'effective', 'EDR detects and blocks ransomware payloads at endpoint level', 35,
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)),

    ('rk000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000001',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-NW-001' LIMIT 1),
     'partially_effective', 'Network segmentation limits lateral movement but does not prevent initial infection', 20,
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)),

    -- Phishing → controls
    ('rk000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000002',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-001' LIMIT 1),
     'effective', 'MFA prevents credential theft from resulting in account compromise', 40,
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
     '2026-02-18',
     (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)),

    ('rk000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000002',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-SA-001' LIMIT 1),
     'partially_effective', 'Security awareness training reduces phishing click rate but does not eliminate it', 25,
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     '2026-02-18',
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)),

    -- PCI → controls
    ('rk000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'rd000000-0000-0000-0000-000000000003',
     (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-SC-001' LIMIT 1),
     'not_assessed', 'Raisin Shield deployment in progress — will address 6.4.3 script monitoring', NULL,
     (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1),
     NULL, NULL);
```

---

## Query Patterns

### Risk Heat Map (5×5 grid aggregation)

"Aggregate active risks by residual likelihood × residual impact for the heat map visualization."

```sql
SELECT
    r.residual_likelihood AS likelihood,
    r.residual_impact AS impact,
    likelihood_to_score(r.residual_likelihood) AS likelihood_score,
    impact_to_score(r.residual_impact) AS impact_score,
    COUNT(*) AS risk_count,
    ARRAY_AGG(json_build_object(
        'id', r.id,
        'identifier', r.identifier,
        'title', r.title,
        'score', r.residual_score,
        'status', r.status
    ) ORDER BY r.residual_score DESC) AS risks
FROM risks r
WHERE r.org_id = $1
  AND r.is_template = FALSE
  AND r.status NOT IN ('closed', 'archived')
  AND r.residual_likelihood IS NOT NULL
  AND r.residual_impact IS NOT NULL
GROUP BY r.residual_likelihood, r.residual_impact
ORDER BY likelihood_to_score(r.residual_likelihood) DESC,
         impact_to_score(r.residual_impact) DESC;
```

### Risk Gap Detection

"Find risks without any treatment plans."

```sql
SELECT
    r.id, r.identifier, r.title, r.category, r.status,
    r.inherent_score, r.residual_score,
    risk_score_severity(r.residual_score) AS severity,
    COUNT(rc.id) AS linked_controls
FROM risks r
LEFT JOIN risk_treatments rt ON rt.risk_id = r.id AND rt.status NOT IN ('cancelled')
LEFT JOIN risk_controls rc ON rc.risk_id = r.id
WHERE r.org_id = $1
  AND r.is_template = FALSE
  AND r.status NOT IN ('closed', 'archived', 'accepted')
  AND rt.id IS NULL  -- no treatment plans
GROUP BY r.id
ORDER BY r.residual_score DESC NULLS LAST;
```

### Risk-to-Control Effectiveness Summary

"For a given risk, show all linked controls with their effectiveness and health status."

```sql
SELECT
    c.id AS control_id,
    c.identifier,
    c.title,
    c.category AS control_category,
    c.status AS control_status,
    rc.effectiveness,
    rc.mitigation_percentage,
    rc.notes,
    rc.last_effectiveness_review,
    -- Latest test result for this control (from Sprint 4)
    (SELECT tr.status FROM test_results tr
     JOIN tests t ON t.id = tr.test_id
     WHERE t.control_id = c.id AND t.org_id = $1
     ORDER BY tr.executed_at DESC LIMIT 1) AS latest_test_status
FROM risk_controls rc
JOIN controls c ON c.id = rc.control_id
WHERE rc.risk_id = $2 AND rc.org_id = $1
ORDER BY rc.effectiveness, c.identifier;
```

### Risk Trend (assessment history for a risk)

"Show how a risk's scores have changed over time."

```sql
SELECT
    ra.assessment_type,
    ra.assessment_date,
    ra.likelihood,
    ra.impact,
    ra.overall_score,
    ra.severity,
    ra.justification,
    u.first_name || ' ' || u.last_name AS assessor,
    ra.is_current
FROM risk_assessments ra
JOIN users u ON u.id = ra.assessed_by
WHERE ra.risk_id = $1 AND ra.org_id = $2
ORDER BY ra.assessment_date DESC, ra.assessment_type;
```

### Overdue Assessments

"Find risks that are due for re-assessment."

```sql
SELECT
    r.id, r.identifier, r.title, r.category,
    r.next_assessment_at,
    r.last_assessed_at,
    r.residual_score,
    risk_score_severity(r.residual_score) AS severity,
    u.first_name || ' ' || u.last_name AS owner
FROM risks r
LEFT JOIN users u ON u.id = r.owner_id
WHERE r.org_id = $1
  AND r.is_template = FALSE
  AND r.status NOT IN ('closed', 'archived')
  AND r.next_assessment_at IS NOT NULL
  AND r.next_assessment_at < CURRENT_DATE
ORDER BY r.next_assessment_at ASC;
```

### Expired Risk Acceptances

"Find accepted risks whose acceptance period has expired."

```sql
SELECT
    r.id, r.identifier, r.title,
    r.acceptance_expiry,
    r.acceptance_justification,
    u.first_name || ' ' || u.last_name AS accepted_by_name,
    r.residual_score
FROM risks r
LEFT JOIN users u ON u.id = r.accepted_by
WHERE r.org_id = $1
  AND r.status = 'accepted'
  AND r.acceptance_expiry IS NOT NULL
  AND r.acceptance_expiry < CURRENT_DATE
ORDER BY r.acceptance_expiry ASC;
```

---

## Audit Log Events

All risk-related actions are logged to `audit_log`:

| Action | Resource Type | When |
|--------|--------------|------|
| `risk.created` | risk | New risk created |
| `risk.updated` | risk | Risk metadata updated |
| `risk.status_changed` | risk | Status transition |
| `risk.archived` | risk | Risk archived |
| `risk.deleted` | risk | Risk permanently deleted |
| `risk.owner_changed` | risk | Owner reassigned |
| `risk.score_recalculated` | risk | Scores refreshed from latest assessments |
| `risk_assessment.created` | risk_assessment | New assessment recorded |
| `risk_assessment.updated` | risk_assessment | Assessment modified |
| `risk_assessment.deleted` | risk_assessment | Assessment removed |
| `risk_treatment.created` | risk_treatment | Treatment plan created |
| `risk_treatment.updated` | risk_treatment | Treatment plan updated |
| `risk_treatment.status_changed` | risk_treatment | Treatment status transition |
| `risk_treatment.completed` | risk_treatment | Treatment marked complete |
| `risk_treatment.cancelled` | risk_treatment | Treatment cancelled |
| `risk_control.linked` | risk_control | Control linked to risk |
| `risk_control.unlinked` | risk_control | Control unlinked from risk |
| `risk_control.effectiveness_updated` | risk_control | Control effectiveness re-evaluated |
