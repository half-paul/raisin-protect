# Audit Engagements

Manage audit engagements from planning through completion.

## Creating an Audit Engagement

1. Go to **Audit Hub > Audit Hub** in the sidebar
2. Click **"New Audit"**
3. Fill in the form:
   - **Title** — e.g., "SOC 2 Type II — 2026 Annual"
   - **Audit Type** — SOC 2 Type I, SOC 2 Type II, ISO 27001 Certification, PCI DSS ROC, GDPR DPIA, Internal, Custom, etc.
   - **Description** — Scope and objectives (optional)
   - **Audit Firm** — e.g., "Deloitte & Touche LLP"
   - **Audit Period** — Start and End dates (the period being examined)
   - **Planned Start/End** — When the audit work will happen
   - **Report Type** — Type of report expected (optional)
4. Click **"Create Engagement"**

The audit is created in **Planning** status.

!!! note "Permissions"
    Only CISO and Compliance Manager can create audit engagements.

## Managing Audit Status

Audits follow a defined workflow:

```
Planning → Fieldwork → Review → Draft Report → Management Response → Final Report → Completed
```

At any non-terminal stage, an audit can also be **Cancelled**.

To advance an audit:

1. Open the audit detail page
2. In the header, you'll see buttons for each valid next status
3. Click the desired transition button
4. In the confirmation dialog, add optional **notes** and confirm

**Key automatic behaviors:**

- Moving to **Fieldwork** sets the actual start date
- Moving to **Completed** sets the actual end date
- **Completed** and **Cancelled** are terminal — no further changes allowed
