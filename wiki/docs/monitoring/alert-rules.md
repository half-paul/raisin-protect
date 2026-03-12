# Alert Rules

Alert rules define when and how alerts are generated from test failures.

## Creating a Rule

1. Go to **Monitoring > Alert Rules** in the sidebar
2. Click **"New Rule"** (CISO/Compliance Manager only)
3. Configure the rule:

| Field | What It Does |
|-------|-------------|
| **Name** | A descriptive name for the rule |
| **Alert Severity** | What severity level the generated alerts will have |
| **Priority** | Rule evaluation order (lower = evaluated first, 0–1000) |
| **Match Test Severities** | Only trigger for tests at these severity levels (leave empty = match all) |
| **Consecutive Failures** | How many times a test must fail before triggering (noise reduction) |
| **Cooldown (minutes)** | Minimum time between alerts for the same condition |
| **SLA Hours** | Auto-set an SLA deadline on generated alerts |
| **Delivery Channels** | How to notify: Slack, Email, Webhook, or In-App |

4. If you selected **Slack**, enter your webhook URL
5. If you selected **Email**, enter recipient email addresses (comma-separated)
6. Click **"Create"**

## Testing Delivery

Before relying on a delivery channel, test it:

1. Click **"Test Delivery"** in the alert rules page header
2. Select a channel (Slack, Email, or Webhook)
3. Enter the target URL or email
4. Click **"Send Test"** to verify it works

## Enabling/Disabling Rules

Use the power toggle icon on each rule row, or the action dropdown menu. Disabled rules appear dimmed and won't generate alerts.
