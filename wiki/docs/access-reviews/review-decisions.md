# Review Decisions

Reviewers decide whether each access grant should continue.

## Making Decisions

1. View your pending reviews: `GET /api/v1/access-reviews/my-reviews`
2. For each review, make a decision:
   - **Approve** — Access is appropriate and should continue
   - **Revoke** — Access should be removed (requires justification)
   - **Flag** — Access needs further investigation (requires justification)
   - **Delegate** — Pass the review to another reviewer
3. Submit your decision: `POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/decide`

## Bulk Decisions

Bulk decisions are supported: `POST /api/v1/access-reviews/campaigns/:id/reviews/bulk-decide`

## Executing Revocations

After revocation decisions, IT Admins mark the actual access removal:
`POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/revocation`
