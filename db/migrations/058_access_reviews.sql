-- Sprint 8 — Table 5: access_reviews

CREATE TABLE IF NOT EXISTS access_reviews (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    campaign_id         UUID NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    entry_id            UUID NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,

    -- Assignment
    reviewer_id         UUID REFERENCES users(id) ON DELETE SET NULL,  -- Who is assigned to review
    assigned_at         TIMESTAMPTZ,

    -- Decision
    decision            review_decision NOT NULL DEFAULT 'pending',
    justification       TEXT,                          -- Required for revoke/flag decisions
    decided_by          UUID REFERENCES users(id) ON DELETE SET NULL,  -- Who actually made the decision
    decided_at          TIMESTAMPTZ,

    -- Delegation
    delegated_to_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    delegated_at        TIMESTAMPTZ,
    delegation_reason   TEXT,

    -- Escalation
    is_escalated        BOOLEAN NOT NULL DEFAULT FALSE,
    escalated_at        TIMESTAMPTZ,
    escalated_to_id     UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Context snapshot (frozen at review time for audit trail)
    access_snapshot     JSONB NOT NULL DEFAULT '{}',

    -- Revocation tracking
    revocation_executed BOOLEAN NOT NULL DEFAULT FALSE,
    revocation_executed_at TIMESTAMPTZ,
    revocation_notes    TEXT,

    -- Metadata
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One review per entry per campaign
    CONSTRAINT uq_access_reviews_campaign_entry UNIQUE (campaign_id, entry_id),
    CONSTRAINT chk_reviews_justification_on_revoke CHECK (
        decision NOT IN ('revoked', 'flagged') OR justification IS NOT NULL
    ),
    CONSTRAINT chk_reviews_decided_at CHECK (
        decision = 'pending' OR decided_at IS NOT NULL
    ),
    CONSTRAINT chk_reviews_decided_by CHECK (
        decision = 'pending' OR decided_by IS NOT NULL
    ),
    CONSTRAINT chk_reviews_revocation CHECK (
        revocation_executed = FALSE OR (decision = 'revoked' AND revocation_executed_at IS NOT NULL)
    )
);

COMMENT ON TABLE access_reviews IS 'Individual access review decisions. One review per access entry per campaign. Decisions are immutable once made.';
COMMENT ON COLUMN access_reviews.access_snapshot IS 'Frozen snapshot of access state at review time. Ensures audit trail survives access entry changes.';
COMMENT ON COLUMN access_reviews.decided_by IS 'May differ from reviewer_id when review was delegated.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_access_reviews_org_id ON access_reviews (org_id);
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign ON access_reviews (campaign_id);
CREATE INDEX IF NOT EXISTS idx_access_reviews_entry ON access_reviews (entry_id);
CREATE INDEX IF NOT EXISTS idx_access_reviews_reviewer ON access_reviews (reviewer_id) WHERE reviewer_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign_decision ON access_reviews (campaign_id, decision);
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign_pending ON access_reviews (campaign_id) WHERE decision = 'pending';
CREATE INDEX IF NOT EXISTS idx_access_reviews_campaign_revoked ON access_reviews (campaign_id) WHERE decision = 'revoked';
CREATE INDEX IF NOT EXISTS idx_access_reviews_escalated ON access_reviews (campaign_id) WHERE is_escalated = TRUE;
CREATE INDEX IF NOT EXISTS idx_access_reviews_revocation_pending ON access_reviews (org_id) WHERE decision = 'revoked' AND revocation_executed = FALSE;
CREATE INDEX IF NOT EXISTS idx_access_reviews_decided_by ON access_reviews (decided_by) WHERE decided_by IS NOT NULL;

-- Trigger
DROP TRIGGER IF EXISTS trg_access_reviews_updated_at ON access_reviews;
CREATE TRIGGER trg_access_reviews_updated_at
    BEFORE UPDATE ON access_reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
