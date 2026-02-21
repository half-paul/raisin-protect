-- Migration: 028_policies.sql
-- Description: Policies table (policy definitions with ownership, lifecycle, templates)
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management

-- ============================================================================
-- POLICIES TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS policies (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    identifier              VARCHAR(50) NOT NULL,
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,

    -- Classification
    category                policy_category NOT NULL,
    status                  policy_status NOT NULL DEFAULT 'draft',

    -- Current published version (FK deferred — added after policy_versions table)
    current_version_id      UUID,

    -- Ownership (primary + secondary owner)
    owner_id                UUID REFERENCES users(id) ON DELETE SET NULL,
    secondary_owner_id      UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Review scheduling (spec §6.1: "Automated annual review reminders")
    review_frequency_days   INT,
    next_review_at          DATE,
    last_reviewed_at        DATE,

    -- Template support
    is_template             BOOLEAN NOT NULL DEFAULT FALSE,
    template_framework_id   UUID REFERENCES frameworks(id) ON DELETE SET NULL,
    cloned_from_policy_id   UUID REFERENCES policies(id) ON DELETE SET NULL,

    -- Approval metadata (denormalized for quick queries)
    approved_at             TIMESTAMPTZ,
    approved_version        INT,
    published_at            TIMESTAMPTZ,

    -- Tags and metadata
    tags                    TEXT[] DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_policy_identifier UNIQUE (org_id, identifier),
    CONSTRAINT chk_review_frequency CHECK (review_frequency_days IS NULL OR review_frequency_days > 0),
    CONSTRAINT chk_template_framework CHECK (
        (is_template = FALSE AND template_framework_id IS NULL) OR
        (is_template = TRUE)
    )
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_policies_org
    ON policies (org_id);

CREATE INDEX IF NOT EXISTS idx_policies_org_status
    ON policies (org_id, status);

CREATE INDEX IF NOT EXISTS idx_policies_org_category
    ON policies (org_id, category);

CREATE INDEX IF NOT EXISTS idx_policies_owner
    ON policies (owner_id)
    WHERE owner_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_policies_secondary_owner
    ON policies (secondary_owner_id)
    WHERE secondary_owner_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_policies_templates
    ON policies (org_id, is_template)
    WHERE is_template = TRUE;

CREATE INDEX IF NOT EXISTS idx_policies_template_framework
    ON policies (template_framework_id)
    WHERE template_framework_id IS NOT NULL AND is_template = TRUE;

CREATE INDEX IF NOT EXISTS idx_policies_review_due
    ON policies (org_id, next_review_at)
    WHERE next_review_at IS NOT NULL AND status IN ('published', 'approved') AND is_template = FALSE;

CREATE INDEX IF NOT EXISTS idx_policies_identifier
    ON policies (org_id, identifier);

CREATE INDEX IF NOT EXISTS idx_policies_tags
    ON policies USING gin (tags);

CREATE INDEX IF NOT EXISTS idx_policies_cloned_from
    ON policies (cloned_from_policy_id)
    WHERE cloned_from_policy_id IS NOT NULL;

-- Full-text search on title + description
CREATE INDEX IF NOT EXISTS idx_policies_search
    ON policies USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- ============================================================================
-- TRIGGER
-- ============================================================================

DROP TRIGGER IF EXISTS trg_policies_updated_at ON policies;
CREATE TRIGGER trg_policies_updated_at
    BEFORE UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE policies IS 'Policy definitions with ownership, lifecycle, review scheduling, and template support (spec §6.1)';
COMMENT ON COLUMN policies.identifier IS 'Human-readable identifier: POL-IS-001, POL-AC-001, etc. Org-scoped unique.';
COMMENT ON COLUMN policies.current_version_id IS 'Denormalized FK to the current published policy_version. Deferred FK added in 032.';
COMMENT ON COLUMN policies.is_template IS 'TRUE = template policy for cloning, not an active governance document';
COMMENT ON COLUMN policies.review_frequency_days IS 'How often policy must be reviewed (e.g., 365 = annual, 90 = quarterly)';
COMMENT ON COLUMN policies.next_review_at IS 'When the next review is due — computed as last_reviewed_at + review_frequency_days';
COMMENT ON COLUMN policies.tags IS 'Free-form tags: [''pci'', ''annual'', ''mandatory'']';
