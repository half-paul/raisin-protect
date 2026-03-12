# Policy Management

Manage organizational policies and governance documents through their full lifecycle.

## Creating a Policy

1. Go to **Policy Management > Policies** in the sidebar
2. Click **"New Policy"**
3. Fill in the form:
   - **Identifier** — A unique code (e.g., `POL-IS-001`)
   - **Category** — Information Security, Access Control, Incident Response, etc. (21 categories available)
   - **Title** — A descriptive name
   - **Description** — Brief summary of the policy's purpose (optional)
   - **Content** — The full policy text. A three-section template is pre-filled:
     - Purpose
     - Scope
     - Policy Statement
   - **Review Frequency** — How often this policy should be reviewed, in days (default: 365)
   - **Tags** — Comma-separated labels (optional)
4. Click **"Create"**

The policy is created in **Draft** status.

!!! note "Permissions"
    CISO, Compliance Manager, and Security Engineer can create policies.

## Editing Policy Content

1. Open the policy detail page
2. Click **"Edit Content"** to go to the rich text editor
3. Edit the HTML content
4. Save your changes — this creates a new version automatically

Each save creates an immutable version record with word count, author, and timestamp.

## Policy Lifecycle

```
Draft → In Review → Approved → Published
                                    ↓
                              Archived (from any status)
```

## Linking Policies to Controls

Show which controls implement your policies:

1. Open the policy detail page
2. Go to the **Controls** tab
3. Click **"Link Control"**
4. Search for a control by identifier or title
5. Select the control and choose coverage:
   - **Full** — This policy fully addresses the control
   - **Partial** — This policy partially addresses the control
6. Add optional notes and click **"Link"**
