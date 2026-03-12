# Evidence Management

Evidence artifacts are files that prove your controls are working — screenshots, configuration exports, audit reports, certificates, and more.

## Uploading Evidence

1. Go to **Compliance > Evidence** in the sidebar
2. Click **"Upload Evidence"**
3. **Drag and drop** a file onto the upload zone, or click to browse:
   - Supported formats: PDF, JSON, CSV, TXT, PNG, JPG, GIF, XLSX, DOCX, XML, ZIP
   - Maximum file size: 100 MB
4. The form auto-fills based on your file:
   - PDF → "Policy Document" type
   - Images → "Screenshot" type
   - JSON → "Configuration Export" type
   - Title is generated from the filename
5. Review and complete the metadata:
   - **Title** — A descriptive name
   - **Description** — What this evidence shows (optional)
   - **Evidence Type** — Screenshot, Configuration Export, Audit Report, etc.
   - **Collection Method** — How it was gathered (Manual Upload, API Pull, etc.)
   - **Collection Date** — When the evidence was collected (defaults to today)
   - **Freshness Period** — How many days this evidence stays "fresh" (e.g., 90 days)
   - **Source System** — Where it came from (e.g., "Okta", "AWS", "Jira")
   - **Tags** — Comma-separated labels (e.g., "mfa, okta, q1-2026")
6. Click **"Upload"** — you'll see a three-step progress: Creating record → Uploading file → Confirming

!!! note "Permissions"
    CISO, Compliance Manager, Security Engineer, IT Admin, and DevOps Engineer can upload evidence.

## Linking Evidence to Controls

After uploading, link your evidence to the controls it supports:

1. Open the evidence detail page (click its title from the evidence library)
2. Click **"Link to Controls"**
3. In the dialog:
   - **Search** for a control by identifier or title
   - **Select** the control from the list
   - Choose the **link strength**:
     - **Primary** — This evidence is the main proof for this control
     - **Supporting** — It contributes supporting proof
     - **Supplementary** — It provides additional context
   - Add optional **notes** explaining the connection
4. Click **"Link"**

You can link the same evidence to multiple controls.

## Managing Evidence Versions

When evidence needs to be refreshed (e.g., a new quarterly screenshot), create a new version rather than uploading a separate artifact:

1. Open the evidence detail page
2. Click **"Add Version"**
3. Upload the new file with updated metadata
4. The new version becomes the **current** version; previous versions are preserved in the version history

## Tracking Freshness & Staleness

Raisin Protect automatically tracks whether your evidence is fresh or stale based on the **freshness period** you set during upload.

- **Fresh** (green badge) — The evidence was collected within its freshness period
- **Expiring Soon** (amber badge) — The evidence will expire within 30 days
- **Expired** (red badge) — The evidence is past its freshness period and needs to be refreshed

### Staleness Alerts Page

1. Go to **Compliance > Staleness Alerts** in the sidebar
2. See summary cards: Total Alerts, Expired count, Expiring Soon count, Affected Controls count
3. The table lists all stale or expiring evidence, sorted by urgency (most urgent first)
4. Use filters:
   - **Alert Level** — Show only Expired or only Expiring Soon
   - **Evidence Type** — Filter by type
   - **Look Ahead** — Change how far ahead to look for expiring evidence (default: 30 days)
5. Click any evidence title to go to its detail page and upload a new version

!!! tip
    Set up a weekly routine to check the staleness alerts page and refresh any expired evidence.

## Evaluating Evidence

Evidence evaluations are quality assessments — typically performed by auditors or compliance managers.

1. Open the evidence detail page
2. Go to the **Evaluations** tab
3. Click **"Add Evaluation"** (if available for your role)
4. Fill in:
   - **Verdict** — Sufficient, Partially Sufficient, Insufficient, or Not Applicable
   - **Confidence** — Low, Medium, or High
   - **Comments** — Explain your assessment
   - **Missing Elements** — What's lacking (optional)
   - **Remediation Notes** — What needs to be done to fix it (optional)
5. Click **"Submit"**

Evaluations are immutable — they create a permanent record and cannot be edited after submission.

!!! note "Permissions"
    CISO, Compliance Manager, and Auditor can evaluate evidence.
