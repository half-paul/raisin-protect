-- Sprint 8 — Cross-Reference FK Extension

-- Add FK column for access_review_campaigns to evidence_links
ALTER TABLE evidence_links
    ADD COLUMN IF NOT EXISTS access_review_campaign_id UUID REFERENCES access_review_campaigns(id) ON DELETE SET NULL;

-- Index for the new FK
CREATE INDEX IF NOT EXISTS idx_evidence_links_access_review_campaign
    ON evidence_links (access_review_campaign_id)
    WHERE access_review_campaign_id IS NOT NULL;
