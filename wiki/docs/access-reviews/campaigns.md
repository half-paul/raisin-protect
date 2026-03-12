# Review Campaigns

A campaign is a periodic review of access across one or more resources.

## Creating a Campaign

1. **Create** a campaign: `POST /api/v1/access-reviews/campaigns`
   - Define the **scope** (which resources, criticalities, privileged-only flag)
   - Set the **reviewer strategy** (resource owner, department head, explicit, or mixed)
   - Set a **deadline** for completion
   - Choose a **cadence** (monthly, quarterly, semi-annual, annual)

2. **Launch** the campaign: `POST /api/v1/access-reviews/campaigns/:id/launch`
   - This generates individual review records for each active access entry in scope
   - Each review is assigned to a reviewer based on the strategy

3. **Monitor progress**: `GET /api/v1/access-reviews/campaigns/:id/stats`
   - Track total, completed, approved, revoked, and flagged counts

4. **Complete** the campaign: `POST /api/v1/access-reviews/campaigns/:id/complete`
   - Any remaining pending reviews are expired
