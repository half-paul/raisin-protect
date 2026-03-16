# Identity Providers

Connect your identity systems to Raisin Protect for access reviews.

## Setting Up an Identity Provider

1. Create an identity provider via the API: `POST /api/v1/access-reviews/identity-providers`
2. Configure the connection with provider-specific settings
3. Trigger a sync: `POST /api/v1/access-reviews/identity-providers/:id/sync`
4. The sync pulls in users, roles, and access records

## Supported Providers

- Okta
- Azure AD
- Google Workspace
- JumpCloud
- OneLogin
- Custom

## Managing Access Resources

Access resources represent the applications, databases, servers, and services that users access:

- List resources: `GET /api/v1/access-reviews/resources`
- Create resources manually: `POST /api/v1/access-reviews/resources`
- Resources synced from identity providers are created automatically

Each resource has a **criticality** level (Low, Medium, High, Critical) that determines review priority.
