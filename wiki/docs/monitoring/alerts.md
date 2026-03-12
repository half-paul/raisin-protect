# Alerts

When a test fails and matches an alert rule, an alert is created. Alerts track issues from detection to resolution.

## Alert Queue

1. Go to **Monitoring > Alert Queue** in the sidebar
2. The page shows summary cards you can click to filter:
   - **Active** — Alerts needing attention (open, acknowledged, in progress)
   - **Resolved** — Fixed but not yet closed
   - **Suppressed** — Temporarily silenced
   - **Closed** — Fully resolved and archived
   - **SLA Breached** — Past their response deadline (highlighted in red)
3. Use the **Severity** filter (Critical, High, Medium, Low)

## Alert Lifecycle

Alerts progress through these statuses:

```
Open → Acknowledged → In Progress → Resolved → Closed
```

Additional transitions:

- Any active status → **Suppressed** (temporarily silence an alert)
- Suppressed → **Open** (un-suppress)
- Resolved or Closed → **Open** (reopen if the issue recurs)

## Working with an Alert

1. Click an alert title to open its detail page
2. **Assign** the alert to a team member (CISO, Compliance Manager, Security Engineer)
3. **Acknowledge** it to show you're aware of the issue
4. **Update status** as you work through the resolution
5. **Resolve** with resolution notes when the issue is fixed
6. **Close** when the resolution is confirmed

!!! info "SLA Tracking"
    If an alert rule sets SLA hours, you'll see a countdown (e.g., "4h left") in the alert queue. Rows with breached SLAs are highlighted red.
