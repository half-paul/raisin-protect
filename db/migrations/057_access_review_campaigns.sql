-- Sprint 8 — Table 4: access_review_campaigns

CREATE TABLE IF NOT EXISTS access_review_campaigns (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Campaign identity
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    status              campaign_status NOT NULL DEFAULT 'draft',
    cadence             campaign_cadence NOT NULL DEFAULT 'quarterly',

    -- Scope configuration (which access entries to include)
    scope               JSONB NOT NULL DEFAULT '{}',

    -- Reviewer assignment strategy
    reviewer_strategy   VARCHAR(50) NOT NULL DEFAULT 'resource_owner',
    default_reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,  -- Fallback reviewer

    -- Timeline
    started_at          TIMESTAMPTZ,
    deadline            TIMESTAMPTZ NOT NULL,          -- All reviews must be completed by this date
    completed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,

    -- Escalation settings
    escalation_config   JSONB NOT NULL DEFAULT '{}',

    -- Denormalized counts
    total_reviews       INTEGER NOT NULL DEFAULT 0,
    completed_reviews   INTEGER NOT NULL DEFAULT 0,
    approved_count      INTEGER NOT NULL DEFAULT 0,
    revoked_count       INTEGER NOT NULL DEFAULT 0,
    flagged_count       INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    created_by          UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    tags                TEXT[] NOT NULL DEFAULT '{}',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_campaigns_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_campaigns_total_reviews CHECK (total_reviews >= 0),
    CONSTRAINT chk_campaigns_completed_reviews CHECK (completed_reviews >= 0 AND completed_reviews <= total_reviews),
    CONSTRAINT chk_campaigns_approved CHECK (approved_count >= 0),
    CONSTRAINT chk_campaigns_revoked CHECK (revoked_count >= 0),
    CONSTRAINT chk_campaigns_flagged CHECK (flagged_count >= 0),
    CONSTRAINT chk_campaigns_deadline_future CHECK (
        status = 'draft' OR deadline IS NOT NULL
    ),
    CONSTRAINT chk_campaigns_completed_after_started CHECK (
        completed_at IS NULL OR (started_at IS NOT NULL AND completed_at >= started_at)
    )
);

COMMENT ON TABLE access_review_campaigns IS 'Governed access review cycles. Scope rules determine which access entries are reviewed.';
COMMENT ON COLUMN access_review_campaigns.scope IS 'JSONB scope config: resource_ids, criticalities, types, departments, privileged_only, service_accounts, stale_threshold.';
COMMENT ON COLUMN access_review_campaigns.reviewer_strategy IS 'How reviewers are assigned: resource_owner, department_head, explicit, or mixed.';
COMMENT ON COLUMN access_review_campaigns.escalation_config IS 'Escalation settings: reminder schedule, escalation delay, recipient, auto-expire.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_campaigns_org_id ON access_review_campaigns (org_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_org_status ON access_review_campaigns (org_id, status);
CREATE INDEX IF NOT EXISTS idx_campaigns_org_cadence ON access_review_campaigns (org_id, cadence);
CREATE INDEX IF NOT EXISTS idx_campaigns_deadline ON access_review_campaigns (deadline) WHERE status IN ('active', 'in_review');
CREATE INDEX IF NOT EXISTS idx_campaigns_created_by ON access_review_campaigns (created_by);
CREATE INDEX IF NOT EXISTS idx_campaigns_tags ON access_review_campaigns USING GIN (tags);

-- Trigger
DROP TRIGGER IF EXISTS trg_campaigns_updated_at ON access_review_campaigns;
CREATE TRIGGER trg_campaigns_updated_at
    BEFORE UPDATE ON access_review_campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
