# Raisin Protect User Guide

> Your step-by-step guide to using Raisin Protect — the AI-native GRC platform for managing compliance, risk, and security operations.

---

## Table of Contents

- [Getting Started](#getting-started)
  - [Creating Your Account](#creating-your-account)
  - [Signing In](#signing-in)
  - [The Dashboard](#the-dashboard)
  - [Navigating the Sidebar](#navigating-the-sidebar)
  - [Understanding Your Role](#understanding-your-role)
- [Compliance Management](#compliance-management)
  - [Activating a Framework](#activating-a-framework)
  - [Browsing Requirements](#browsing-requirements)
  - [Scoping Requirements](#scoping-requirements)
  - [Viewing Coverage](#viewing-coverage)
- [Control Library](#control-library)
  - [Browsing Controls](#browsing-controls)
  - [Creating a Control](#creating-a-control)
  - [Mapping Controls to Requirements](#mapping-controls-to-requirements)
  - [Managing Control Status](#managing-control-status)
  - [Using the Mapping Matrix](#using-the-mapping-matrix)
- [Evidence Management](#evidence-management)
  - [Uploading Evidence](#uploading-evidence)
  - [Linking Evidence to Controls](#linking-evidence-to-controls)
  - [Managing Evidence Versions](#managing-evidence-versions)
  - [Tracking Freshness & Staleness](#tracking-freshness--staleness)
  - [Evaluating Evidence](#evaluating-evidence)
- [Continuous Monitoring](#continuous-monitoring)
  - [Monitoring Dashboard](#monitoring-dashboard)
  - [Understanding the Control Health Heatmap](#understanding-the-control-health-heatmap)
  - [Running Tests](#running-tests)
  - [Viewing Test Results](#viewing-test-results)
  - [Managing Alerts](#managing-alerts)
  - [Configuring Alert Rules](#configuring-alert-rules)
- [Policy Management](#policy-management)
  - [Creating a Policy](#creating-a-policy)
  - [Editing Policy Content](#editing-policy-content)
  - [Submitting for Review](#submitting-for-review)
  - [Approving or Rejecting a Policy](#approving-or-rejecting-a-policy)
  - [Publishing a Policy](#publishing-a-policy)
  - [Using Policy Templates](#using-policy-templates)
  - [Comparing Versions](#comparing-versions)
  - [Linking Policies to Controls](#linking-policies-to-controls)
  - [Policy Gap Analysis](#policy-gap-analysis)
- [Risk Management](#risk-management)
  - [Viewing the Risk Register](#viewing-the-risk-register)
  - [Creating a Risk](#creating-a-risk)
  - [Assessing a Risk](#assessing-a-risk)
  - [Understanding the Heat Map](#understanding-the-heat-map)
  - [Creating Treatment Plans](#creating-treatment-plans)
  - [Completing a Treatment](#completing-a-treatment)
  - [Linking Controls to Risks](#linking-controls-to-risks)
  - [Accepting a Risk](#accepting-a-risk)
  - [Risk Gap Analysis](#risk-gap-analysis)
- [Audit Hub](#audit-hub)
  - [Creating an Audit Engagement](#creating-an-audit-engagement)
  - [Managing Audit Status](#managing-audit-status)
  - [Creating Evidence Requests (PBC Items)](#creating-evidence-requests-pbc-items)
  - [Submitting Evidence to a Request](#submitting-evidence-to-a-request)
  - [Reviewing Evidence (Auditors)](#reviewing-evidence-auditors)
  - [Creating and Tracking Findings](#creating-and-tracking-findings)
  - [Submitting a Management Response](#submitting-a-management-response)
  - [Using PBC Templates](#using-pbc-templates)
  - [Checking Audit Readiness](#checking-audit-readiness)
  - [The Auditor Workspace](#the-auditor-workspace)
- [Access Reviews](#access-reviews)
  - [Setting Up Identity Providers](#setting-up-identity-providers)
  - [Managing Access Resources](#managing-access-resources)
  - [Running a Review Campaign](#running-a-review-campaign)
  - [Making Review Decisions](#making-review-decisions)
- [Administration](#administration)
  - [Managing Users](#managing-users)
  - [Organization Settings](#organization-settings)
  - [Changing Your Password](#changing-your-password)
- [Compliance Posture](#compliance-posture)
- [Tips & Best Practices](#tips--best-practices)

---

## Getting Started

### Creating Your Account

1. Navigate to the Raisin Protect URL (default: `http://localhost:3010`)
2. Click **"Create one"** on the login page
3. Fill in the registration form:
   - **Organization Name** — Your company name (e.g., "Acme Corporation")
   - **First Name** and **Last Name**
   - **Email** — Your work email
   - **Password** — Must be at least 8 characters with one uppercase letter, one lowercase letter, one number, and one special character
   - **Confirm Password** — Re-enter your password
4. Click **"Create account"**

You'll be signed in automatically as the first user of your organization with the **Compliance Manager** role.

> **Tip:** After setup, have your CISO sign in and invite additional team members with appropriate roles.

---

### Signing In

1. Go to the login page
2. Enter your **Email** and **Password**
3. Click **"Sign in"**

If you see a red error message, check your email and password. After too many failed attempts, you may be temporarily rate-limited (10 attempts per minute).

---

### The Dashboard

After signing in, you land on the main **Dashboard** — your compliance command center.

The dashboard shows:

| Card | What It Tells You |
|------|-------------------|
| **Compliance Score** | Your overall coverage percentage across all active frameworks |
| **Active Frameworks** | How many compliance frameworks you're tracking |
| **Controls** | Total controls with a breakdown of active vs. draft |
| **Coverage Gaps** | Controls not yet mapped to any framework requirement |

Below the stats you'll find:
- **Recent Audit Activity** — Latest actions taken in the platform
- **System Health** — Status of backend services
- **Framework Coverage** — A progress bar for each active framework (click any bar to see that framework's details)

---

### Navigating the Sidebar

The left sidebar is your main navigation. It's organized into sections:

| Section | What You'll Find |
|---------|------------------|
| **Overview** | Dashboard (home page) |
| **Risk & Posture** | Risk dashboard, compliance posture scores |
| **Compliance** | Frameworks, controls, mapping matrix, coverage, evidence, staleness alerts |
| **Risk Management** | Risk register, heat map, gap analysis, treatment plans |
| **Policy Management** | Policies, templates, approval queue, gap analysis |
| **Monitoring** | Real-time monitoring, alert queue, test runs, alert rules |
| **Audit Hub** | Audit engagements, PBC templates, readiness dashboard, auditor workspace |
| **Administration** | User management, organization settings |

> **Note:** You'll only see sidebar items that your role has permission to access. For example, only auditors see the "Auditor Workspace" item.

At the bottom of the sidebar you'll see your name, role badge, and a **Sign Out** button.

---

### Understanding Your Role

Your role determines what you can see and do in Raisin Protect. Here's a practical guide:

| Role | What You Can Do |
|------|----------------|
| **CISO** | Everything. Full platform access. Can accept risks, publish policies, manage users, and approve audit findings. |
| **Compliance Manager** | Nearly everything. Same as CISO except risk acceptance requires CISO. Primary driver of compliance programs. |
| **Security Engineer** | Create and manage controls, evidence, policies, and risks. Run monitoring tests. Cannot publish policies or manage users. |
| **IT Admin** | Upload evidence, manage identity providers, handle alerts. Focused on infrastructure compliance tasks. |
| **DevOps Engineer** | Upload evidence, run monitoring tests, manage alerts. Focused on automated compliance. |
| **Auditor** | View compliance data, create audit findings, review evidence, evaluate artifacts. Cannot modify controls or policies. Has a dedicated workspace. |
| **Vendor Manager** | Manage vendor relationships (vendor management module). |

> **Key concept:** Buttons and actions you don't have permission for are simply hidden — you won't see them. If a section is missing from your sidebar, it means your role doesn't have access to it.

---

## Compliance Management

### Activating a Framework

Before you can track compliance, you need to activate one or more frameworks.

1. Go to **Compliance > Frameworks** in the sidebar
2. Click the **"Available"** tab to see frameworks you haven't activated yet
3. Find your framework (SOC 2, ISO 27001, PCI DSS, GDPR, or CCPA)
4. Click **"Activate Framework"**
5. In the dialog:
   - **Version** — Select the framework version (e.g., "SOC 2 — 2024 TSC")
   - **Target Compliance Date** — When you aim to be compliant (optional)
   - **Notes** — Why you're activating this framework (optional)
   - **Seed pre-built controls** — Leave checked to get starter controls from the template library (recommended for new setups)
6. Click **"Activate"**

The framework now appears in your **"Activated"** tab with a coverage progress bar.

> **Who can do this:** CISO and Compliance Manager only.

---

### Browsing Requirements

Each framework has a tree of requirements — the specific criteria you need to satisfy.

1. Go to **Compliance > Frameworks**
2. Click on an activated framework card
3. The framework detail page shows all requirements organized hierarchically (e.g., SOC 2's CC1 through CC9 sections, each with sub-requirements like CC6.1, CC6.2)

Requirements that are **assessable** (leaf-level items) are the ones you need to map controls to.

---

### Scoping Requirements

Not every requirement may apply to your organization. You can mark requirements as in-scope or out-of-scope.

1. Open a framework's detail page
2. Navigate to the **Scoping** view via the org-framework's scoping endpoint
3. For each requirement, set whether it's **In Scope** or **Out of Scope**
4. Provide a **justification** for out-of-scope decisions (important for audit evidence)

> **Who can do this:** CISO and Compliance Manager only.

---

### Viewing Coverage

The **Coverage** page shows how well your controls cover each framework's requirements.

1. Go to **Compliance > Coverage** in the sidebar
2. At the top, you'll see:
   - **Overall Coverage %** with a progress bar
   - **Active Frameworks** count
   - **Requirements Covered** count
   - **Coverage Gaps** count
3. Below, each framework shows:
   - A coverage percentage with a color-coded progress bar (green 80%+, amber 50%+, red below 50%)
   - Covered, gap, and out-of-scope counts
   - Target compliance date

This page is read-only — it automatically reflects your control mappings.

---

## Control Library

### Browsing Controls

1. Go to **Compliance > Controls** in the sidebar
2. The control library shows summary stats at the top: total, active, draft, unmapped, and custom counts
3. Use the filters to narrow down:
   - **Search** — Type a control identifier or title
   - **Status** — Filter by Draft, Active, Under Review, or Deprecated
   - **Category** — Filter by Technical, Administrative, Physical, or Operational
4. Click any control's identifier or title to see its full details

---

### Creating a Control

1. On the Controls page, click **"New Control"**
2. Fill in the form:
   - **Identifier** — A unique code (e.g., `CTRL-AC-001`)
   - **Title** — A descriptive name (e.g., "Multi-Factor Authentication for Admin Access")
   - **Description** — What this control does and how it's implemented
   - **Category** — Technical, Administrative, Physical, or Operational
   - **Initial Status** — Draft (recommended) or Active
3. Click **"Create"**

> **Who can do this:** CISO, Compliance Manager, and Security Engineer.

---

### Mapping Controls to Requirements

Mapping connects your controls to framework requirements, showing which controls satisfy which compliance criteria.

1. Open a control's detail page (click its identifier from the control library)
2. Go to the **Mappings** tab
3. To add a mapping, use the control mappings endpoint with:
   - **Requirement** — Which framework requirement this control satisfies
   - **Strength** — How strongly it satisfies the requirement:
     - **Primary** — Directly and fully satisfies the requirement
     - **Supporting** — Partially addresses the requirement
     - **Partial** — Contributes to but does not fully cover the requirement

> **Tip:** A single control can map to multiple requirements across different frameworks. This is how Raisin Protect tracks cross-framework coverage.

---

### Managing Control Status

Controls follow a lifecycle:

```
Draft → Active → Under Review → Active (back)
                              → Deprecated
Deprecated → Draft (reactivate)
```

To change a control's status:
1. Open the control's detail page
2. Use the status action from the row's action menu (hover over the row to see the three-dot menu)
3. Available transitions depend on the current status:
   - **Draft** → Activate or Deprecate
   - **Active** → Mark for Review or Deprecate
   - **Under Review** → Reactivate or Deprecate
   - **Deprecated** → Revert to Draft

For bulk changes, select multiple controls using checkboxes and click the **"Bulk Status"** button (CISO/Compliance Manager only).

---

### Using the Mapping Matrix

The **Mapping Matrix** gives you a bird's-eye view of how your controls map across all active frameworks.

1. Go to **Compliance > Mapping Matrix** in the sidebar
2. The matrix shows controls as rows and frameworks as columns
3. Each cell shows a colored badge:
   - **Green** — Primary mapping
   - **Blue** — Supporting mapping
   - **Amber** — Partial mapping
   - **Dash (–)** — No mapping
4. Hover over any cell to see the specific requirements mapped
5. Use the search bar and category filter to narrow down the view

The first column (control identifier) is sticky — it stays visible as you scroll horizontally across frameworks.

---

## Evidence Management

### Uploading Evidence

Evidence artifacts are files that prove your controls are working — screenshots, configuration exports, audit reports, certificates, and more.

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

> **Who can do this:** CISO, Compliance Manager, Security Engineer, IT Admin, and DevOps Engineer.

---

### Linking Evidence to Controls

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

---

### Managing Evidence Versions

When evidence needs to be refreshed (e.g., a new quarterly screenshot), create a new version rather than uploading a separate artifact:

1. Open the evidence detail page
2. Click **"Add Version"**
3. Upload the new file with updated metadata
4. The new version becomes the **current** version; previous versions are preserved in the version history

The **Versions** tab on the evidence detail page shows the full version chain.

---

### Tracking Freshness & Staleness

Raisin Protect automatically tracks whether your evidence is fresh or stale based on the **freshness period** you set during upload.

- **Fresh** (green badge) — The evidence was collected within its freshness period
- **Expiring Soon** (amber badge) — The evidence will expire within 30 days
- **Expired** (red badge) — The evidence is past its freshness period and needs to be refreshed

#### Staleness Alerts Page

1. Go to **Compliance > Staleness Alerts** in the sidebar
2. See summary cards: Total Alerts, Expired count, Expiring Soon count, Affected Controls count
3. The table lists all stale or expiring evidence, sorted by urgency (most urgent first)
4. Use filters:
   - **Alert Level** — Show only Expired or only Expiring Soon
   - **Evidence Type** — Filter by type
   - **Look Ahead** — Change how far ahead to look for expiring evidence (default: 30 days)
5. Click any evidence title to go to its detail page and upload a new version

> **Tip:** Set up a weekly routine to check the staleness alerts page and refresh any expired evidence.

---

### Evaluating Evidence

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

> **Who can do this:** CISO, Compliance Manager, and Auditor.

---

## Continuous Monitoring

### Monitoring Dashboard

The monitoring dashboard gives you a real-time view of your compliance posture.

1. Go to **Monitoring > Monitoring** in the sidebar
2. At the top you'll see:
   - **Posture Score** — Overall compliance health percentage
   - **Test Pass Rate (24h)** — How many tests passed in the last 24 hours
   - **Open Alerts** — Active alerts requiring attention
   - **SLA Breached** — Alerts that have exceeded their response deadline

---

### Understanding the Control Health Heatmap

The heatmap is the central visualization on the monitoring dashboard. It shows every control as a colored square:

| Color | Meaning |
|-------|---------|
| **Green** | Healthy — All tests passing |
| **Red** | Failing — One or more tests failing |
| **Orange** | Error — Tests encountering errors |
| **Amber** | Warning — Tests producing warnings |
| **Gray** | Untested — No tests configured for this control |

**How to use it:**
- **Hover** over any square to see the control's identifier, title, health status, test count, active alerts, and when it was last tested
- **Click** any square to jump to that control's detail page
- **Filter by category** using the dropdown above the heatmap (All, Technical, Administrative, Physical, Operational)

Below the heatmap, you'll see:
- **Compliance Posture** — Per-framework scores with trend indicators (Improving, Declining, or Stable)
- **Recent Activity** — Latest alerts created, resolved, and test runs completed
- **Quick Actions** — Links to the alert queue, test history, and alert rules

---

### Running Tests

Tests check whether your controls are working correctly. You can run them manually or let them run on a schedule.

#### Running All Tests

1. Go to **Monitoring > Test Runs** in the sidebar
2. Click **"Run All Tests"**
3. Confirm in the dialog by clicking **"Run Tests"**
4. A new test run appears in the list with status **Pending**, then **Running**
5. When complete, the status changes to **Completed** with pass/fail counts

#### Viewing Test Run History

The Test Runs page shows all previous runs with:
- **Run number** and status badge
- **Trigger type** — Manual, Scheduled, On Change, or Webhook
- **Results** — Pass, Fail, and Error counts
- **Duration** — How long the run took
- **Triggered By** — Who or what started it

Use the **Status** and **Trigger** filters to narrow down the list.

> **Who can run tests:** CISO, Compliance Manager, Security Engineer, and DevOps Engineer.

---

### Viewing Test Results

1. Click the link icon on any test run row, or navigate to a test run's detail page
2. See summary cards: Total Tests, Passed (green), Failed (red), Errors (orange), Skipped/Warnings
3. Use the filter buttons to show only failed, errored, passing, or all results
4. Each result shows:
   - **Status** — Pass, Fail, Error, Skip, or Warning
   - **Severity** — How critical the test is
   - **Test ID and title** — Which test ran
   - **Control** — Which control this test checks (clickable)
   - **Message** — Brief result summary
   - **Alert** — If a failure generated an alert (clickable link to the alert)
5. Click the detail icon on any result to see the full output log, error messages, and JSON details

---

### Managing Alerts

When a test fails and matches an alert rule, an alert is created. Alerts track issues from detection to resolution.

1. Go to **Monitoring > Alert Queue** in the sidebar
2. The page shows summary cards you can click to filter:
   - **Active** — Alerts needing attention (open, acknowledged, in progress)
   - **Resolved** — Fixed but not yet closed
   - **Suppressed** — Temporarily silenced
   - **Closed** — Fully resolved and archived
   - **SLA Breached** — Past their response deadline (highlighted in red)
3. Use the **Severity** filter (Critical, High, Medium, Low)

#### Alert Lifecycle

Alerts progress through these statuses:

```
Open → Acknowledged → In Progress → Resolved → Closed
```

Additional transitions:
- Any active status → **Suppressed** (temporarily silence an alert)
- Suppressed → **Open** (un-suppress)
- Resolved or Closed → **Open** (reopen if the issue recurs)

#### Working with an Alert

1. Click an alert title to open its detail page
2. **Assign** the alert to a team member (CISO, Compliance Manager, Security Engineer)
3. **Acknowledge** it to show you're aware of the issue
4. **Update status** as you work through the resolution
5. **Resolve** with resolution notes when the issue is fixed
6. **Close** when the resolution is confirmed

> **SLA Tracking:** If an alert rule sets SLA hours, you'll see a countdown (e.g., "4h left") in the alert queue. Rows with breached SLAs are highlighted red.

---

### Configuring Alert Rules

Alert rules define when and how alerts are generated from test failures.

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

#### Testing Delivery

Before relying on a delivery channel, test it:
1. Click **"Test Delivery"** in the alert rules page header
2. Select a channel (Slack, Email, or Webhook)
3. Enter the target URL or email
4. Click **"Send Test"** to verify it works

#### Enabling/Disabling Rules

Use the power toggle icon on each rule row, or the action dropdown menu. Disabled rules appear dimmed and won't generate alerts.

---

## Policy Management

### Creating a Policy

1. Go to **Policy Management > Policies** in the sidebar
2. Click **"New Policy"**
3. Fill in the form:
   - **Identifier** — A unique code (e.g., `POL-IS-001`)
   - **Category** — Information Security, Access Control, Incident Response, etc. (21 categories available)
   - **Title** — A descriptive name
   - **Description** — Brief summary of the policy's purpose (optional)
   - **Content** — The full policy text. A three-section template is pre-filled for you:
     - Purpose
     - Scope
     - Policy Statement
   - **Review Frequency** — How often this policy should be reviewed, in days (default: 365)
   - **Tags** — Comma-separated labels (optional)
4. Click **"Create"**

The policy is created in **Draft** status.

> **Who can do this:** CISO, Compliance Manager, and Security Engineer.

---

### Editing Policy Content

1. Open the policy detail page
2. Click **"Edit Content"** to go to the rich text editor
3. Edit the HTML content
4. Save your changes — this creates a new version automatically

Each save creates an immutable version record with word count, author, and timestamp.

---

### Submitting for Review

When your policy is ready, submit it for approval:

1. Open the policy detail page (must be in **Draft** or **Approved** status)
2. Click **"Submit for Review"**
3. In the dialog:
   - Enter the **Signer IDs** — the UUIDs of users who need to approve (you can find user IDs in the user management page)
   - Set a **Due Date** for the review (optional)
   - Add a **Message** for the reviewers (optional)
4. Click **"Submit"**

The policy moves to **In Review** status, and sign-off records are created for each signer.

---

### Approving or Rejecting a Policy

If you've been asked to sign off on a policy:

1. Go to **Policy Management > Approvals** in the sidebar
2. The **"My Pending"** tab shows policies waiting for your approval
3. For each pending sign-off, you can:
   - Click **"View"** to read the full policy first
   - Click **"Approve"** — add optional comments and confirm
   - Click **"Reject"** — you must provide comments explaining why (required)

**What happens after all signers approve:**
The policy automatically transitions to **Approved** status. No manual action needed — the system detects when all pending sign-offs are complete.

**What happens if someone rejects:**
The policy stays in **In Review**. The owner can address the feedback, update the content, and re-submit.

---

### Publishing a Policy

After a policy is approved, it can be published to make it official:

1. Open the approved policy's detail page
2. Click **"Publish"**
3. The policy moves to **Published** status

> **Who can do this:** CISO and Compliance Manager only.

#### Policy Lifecycle Summary

```
Draft → In Review → Approved → Published
                                    ↓
                              Archived (from any status)
```

---

### Using Policy Templates

Save time by starting from pre-built templates:

1. Go to **Policy Management > Templates** in the sidebar
2. Browse templates organized by framework (SOC 2, ISO 27001, General, etc.)
3. Use search and filters to find what you need
4. Click **"Clone to My Policies"** on the template you want
5. In the clone dialog:
   - Review the pre-filled **Identifier**, **Title**, and **Description**
   - Adjust the **Review Frequency** if needed
   - Modify **Tags** as needed
6. Click **"Clone"**

You'll be taken directly to your new policy where you can customize the content.

---

### Comparing Versions

1. Open a policy detail page
2. Go to the **Versions** tab (or click the link to the versions page)
3. The version table shows every version with change type, word count, and author
4. In the **Compare Versions** section:
   - Select a **From (older)** version and a **To (newer)** version
   - Click **"Compare"**
5. The two versions appear side-by-side:
   - **Left** (red tint) — The older version
   - **Right** (green tint) — The newer version
   - Word count difference shown as "+N" or "-N"

---

### Linking Policies to Controls

Show which controls implement your policies:

1. Open the policy detail page
2. Go to the **Controls** tab
3. Click **"Link Control"**
4. Search for a control by identifier or title
5. Select the control and choose coverage:
   - **Full** — This policy fully addresses the control
   - **Partial** — This policy partially addresses the control
6. Add optional notes and click **"Link"**

---

### Policy Gap Analysis

Find controls that lack policy coverage:

1. Go to **Policy Management > Policy Gaps** in the sidebar
2. See summary stats: total controls, full/partial/no coverage counts, and overall coverage percentage
3. **By Control tab** — Lists controls with no or partial policy coverage, sorted by impact (controls mapped to the most framework requirements appear first)
4. **By Framework tab** — Shows coverage statistics per framework with gap counts and progress bars

Use the filters to focus on specific frameworks or categories.

---

## Risk Management

### Viewing the Risk Register

1. Go to **Risk Management > Risk Register** in the sidebar
2. See summary stats: Total Risks, Critical count, High count, Appetite Breaches, and Average Risk Reduction
3. A mini heat map preview shows risk distribution (click **"View Full"** for the full heat map)
4. The table lists all risks with:
   - **Identifier** and **Title**
   - **Category** — Operational, Financial, Compliance, Strategic, Technology, Third Party, Physical, Reputational
   - **Status** — Identified, Open, Assessing, Treating, Monitoring, Accepted, Closed
   - **Inherent Score** — Risk level before controls (color-coded by severity)
   - **Residual Score** — Risk level after controls (color-coded by severity)
   - **Owner** and **Linked Controls**

---

### Creating a Risk

1. On the Risk Register page, click **"New Risk"**
2. You'll be taken to the risk editor page where you can fill in:
   - **Identifier** — A unique code (e.g., `RISK-T-001`)
   - **Title** — A descriptive name
   - **Description** — Full details about the risk
   - **Category** — Select from 8 categories
   - **Source** — Where this risk was identified
   - **Affected Assets** — Systems or data at risk
3. Save the risk — it's created in **Identified** status

> **Who can do this:** CISO, Compliance Manager, and Security Engineer.

---

### Assessing a Risk

Risk assessment calculates a score using a 5x5 matrix of Likelihood and Impact.

1. Open a risk's detail page
2. Go to the **Assessments** tab
3. Click **"New Assessment"**
4. Fill in the form:
   - **Assessment Type**:
     - *Inherent* — Risk level without any controls
     - *Residual* — Risk level with current controls in place
   - **Likelihood** — How likely is this risk to occur?
     - 1 = Rare, 2 = Unlikely, 3 = Possible, 4 = Likely, 5 = Almost Certain
   - **Impact** — How severe would the consequences be?
     - 1 = Negligible, 2 = Minor, 3 = Moderate, 4 = Major, 5 = Severe
   - The dialog shows a **live preview** of the calculated score (Likelihood x Impact) and severity
   - **Justification** — Explain your reasoning
   - **Assumptions** — Any assumptions made
   - **Valid Until** — When this assessment should be reviewed (optional)
5. Click **"Save"**

#### Scoring Guide

| Score Range | Severity | Color |
|-------------|----------|-------|
| 20–25 | Critical | Red |
| 12–19 | High | Orange |
| 6–11 | Medium | Amber |
| 1–5 | Low | Green |

---

### Understanding the Heat Map

The heat map visualizes all your risks on a 5x5 grid:

1. Go to **Risk Management > Heat Map** in the sidebar
2. The Y-axis shows **Likelihood** (Rare at bottom, Almost Certain at top)
3. The X-axis shows **Impact** (Negligible at left, Severe at right)
4. Each cell shows a count of risks at that intersection, color-coded by severity
5. **Hover** over any cell to see which specific risks are there
6. Use the filters:
   - **Score Type** — Switch between Residual (default) and Inherent risk scores
   - **Category** — Filter by risk category

The goal is to move risks from the upper-right (high likelihood, high impact) toward the lower-left (low likelihood, low impact) through controls and treatments.

---

### Creating Treatment Plans

Treatment plans describe how you'll address a risk.

1. Open the risk detail page
2. Go to the **Treatments** tab
3. Click **"New Treatment"**
4. Fill in:
   - **Treatment Type**:
     - *Mitigate* — Reduce the likelihood or impact
     - *Transfer* — Shift the risk to a third party (e.g., insurance)
     - *Accept* — Acknowledge and accept the risk
     - *Avoid* — Eliminate the risk by removing its source
   - **Priority** — Low, Medium, High, or Critical
   - **Title** — What action you'll take
   - **Description** — Detailed plan
   - **Due Date** — Target completion date
   - **Estimated Hours** — Expected effort
   - **Expected Residual Scores** — Where you expect likelihood and impact to be after treatment (optional)
5. Click **"Create"**

---

### Completing a Treatment

When you've finished implementing a treatment:

1. Open the risk detail page
2. Go to the **Treatments** tab
3. On the treatment card, click the **"Complete"** button
4. Fill in:
   - **Actual Effort Hours** — How long it actually took
   - **Effectiveness Rating** — How well the treatment worked:
     - Highly Effective, Effective, Partially Effective, or Ineffective
   - **Effectiveness Notes** — Describe the outcome
5. Click **"Submit"**

Treatment lifecycle:
```
Planned → In Progress → Implemented → Verified
                                    → Ineffective
```

---

### Linking Controls to Risks

Show which controls mitigate each risk:

1. Open the risk detail page
2. Go to the **Controls** tab
3. Click **"Link Control"**
4. Search for and select a control
5. Set the **effectiveness**:
   - Effective, Partially Effective, Ineffective, or Not Assessed
6. Set the **mitigation percentage** (0–100%)

---

### Accepting a Risk

Sometimes the right decision is to accept a risk rather than treat it:

1. Open the risk detail page
2. Change the status to **Accepted**
3. You must provide:
   - **Justification** — Why you're accepting this risk
   - **Expiry Date** — When the acceptance should be reviewed

> **Who can do this:** CISO and Compliance Manager only.

---

### Risk Gap Analysis

Find risks that are missing treatments, controls, or have overdue assessments:

1. Go to **Risk Management > Risk Gaps** in the sidebar
2. See summary cards showing counts of active risks, those without treatments, without controls, with overdue assessments, and expired acceptances
3. The gap table shows each risk with its gap types highlighted:
   - **No Treatments** — Risk has no treatment plans
   - **No Controls** — No linked controls
   - **High Risk (No Controls)** — High or critical risk without controls (most urgent)
   - **Overdue Assessment** — Assessment is past its review date
   - **Expired Acceptance** — Risk acceptance has expired
4. Use filters to focus on specific gap types or severity levels

---

## Audit Hub

### Creating an Audit Engagement

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

> **Who can do this:** CISO and Compliance Manager only.

---

### Managing Audit Status

Audits follow a defined workflow. Status transitions are managed by CISO and Compliance Manager.

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

---

### Creating Evidence Requests (PBC Items)

PBC (Prepared By Client) requests are how auditors ask for specific evidence.

1. Open an audit detail page
2. Go to the **Requests** tab
3. Click **"New Request"**
4. Fill in:
   - **Title** — What evidence is needed (e.g., "MFA configuration screenshot for admin access")
   - **Description** — Detailed requirements
   - **Priority** — Critical, High, Medium, or Low
   - **Due Date** — When the evidence is needed
   - **Reference Number** — e.g., "PBC-001"
5. Click **"Create"**

#### Bulk Creating from Templates

For faster setup:
1. Go to **Audit Hub > PBC Templates** in the sidebar
2. Select relevant templates using checkboxes
3. Click **"Create N Requests"**
4. Choose the target audit, set a default due date, and a reference prefix
5. The system creates individual requests from each selected template

> **Who can create requests:** CISO, Compliance Manager, and Auditor.

---

### Submitting Evidence to a Request

When you've been assigned a PBC request:

1. Open the request detail page (from the audit's Requests tab)
2. Click **"Link Evidence"**
3. Enter the **Artifact ID** of existing evidence from the Evidence Library
4. Add **Submission Notes** explaining relevance (optional)
5. Click **"Submit Evidence"**
6. Once you've attached all relevant evidence, click **"Submit for Review"** to notify the auditor

The request status progresses: **Open → In Progress → Submitted**

> **Who can submit:** CISO, Compliance Manager, Security Engineer, and IT Admin.

---

### Reviewing Evidence (Auditors)

As an auditor, when evidence has been submitted:

1. Open the request detail page
2. In the **Chain of Custody** table, find evidence with **"Pending Review"** status
3. Click **"Review"** on each piece of evidence
4. Choose your decision:
   - **Accept** — Evidence is satisfactory
   - **Reject** — Evidence doesn't meet requirements (provide notes)
   - **Needs Clarification** — More information needed
5. Once all evidence is reviewed, click **"Review Request"** to accept or reject the entire request

> **Who can review:** Auditor only.

---

### Creating and Tracking Findings

Auditors create findings when they discover deficiencies:

1. Open an audit's **Findings** tab
2. Click **"New Finding"** (Auditor only)
3. Fill in:
   - **Title** — Brief description of the deficiency
   - **Description** — Detailed explanation
   - **Severity** — Critical, High, Medium, Low, or Informational
   - **Category** — Control Deficiency, Documentation Gap, Process Gap, Configuration Issue, etc.
   - **Recommendation** — What should be done to fix it
   - **Remediation Due Date** — When the fix is expected
4. Click **"Create"**

#### Finding Lifecycle

```
Identified → Acknowledged → Remediation Planned → Remediation In Progress →
Remediation Complete → Verified → Closed
```

At certain stages, specific information is required:
- **Remediation Planned** — Requires a remediation plan text
- **Verified** — Typically done by the auditor after confirming the fix
- **Risk Accepted** — Requires CISO approval and a justification reason

To advance a finding:
1. Use the status dropdown in the findings table or detail page
2. Select the next status
3. Fill in any required fields (remediation plan, notes, etc.)
4. Confirm the transition

---

### Submitting a Management Response

When an audit finding needs a formal response from management:

1. Open the finding detail page
2. Click **"Mgmt Response"** (visible to CISO/Compliance Manager)
3. Enter your organization's formal response to the finding
4. Click **"Submit"**

The management response appears in the finding detail alongside the auditor's recommendation.

---

### Using PBC Templates

PBC templates pre-define common evidence requests for different audit types:

1. Go to **Audit Hub > PBC Templates** in the sidebar
2. Browse templates organized by framework
3. Use filters:
   - **Search** — Find templates by title
   - **Audit Type** — Filter by SOC 2, ISO 27001, PCI DSS, etc.
   - **Framework** — Filter by framework
4. Select templates using checkboxes (or use "Select All")
5. Click **"Create N Requests"**
6. Choose which audit to create them in, set defaults, and confirm

---

### Checking Audit Readiness

Before an audit begins, check your preparation:

1. Go to **Audit Hub > Audit Readiness** in the sidebar
2. Select an active audit from the dropdown
3. Review:
   - **Overall Readiness %** — How many evidence requests are accepted
   - **Accepted Requests** — Count of completed items
   - **Overdue Requests** — Items past their due date
   - **Critical/High Findings** — Outstanding findings count
4. Scroll down to see readiness by requirement and by control, with individual progress bars
5. The **Coverage Gaps** table highlights requirements that still need evidence
6. The **Audit Timeline** shows planned dates, days elapsed, and upcoming milestones

---

### The Auditor Workspace

Auditors have a dedicated workspace designed for their specific needs:

1. Go to **Audit Hub > Auditor Workspace** in the sidebar (Auditor role only)
2. See at a glance:
   - **Active Engagements** — Cards for each audit you're assigned to
   - **Open Requests** — Total evidence requests awaiting your review
   - **Open Findings** — Findings you've created that are still being remediated
   - **Overdue Requests** — Requests past their due dates
3. Below, you'll find:
   - **Overdue Evidence Requests** table with days-overdue badges
   - **Critical/High Findings** table showing the most serious issues
   - **Recent Activity** feed with the latest audit actions
   - **Completed Engagements** archive

> **Auditor isolation:** You only see audits where you've been added as an auditor. The system automatically filters based on your assignment — you don't need to apply any filters.

---

## Access Reviews

Access reviews ensure that user access to systems is appropriate and follows the principle of least privilege. This feature is currently available via the API.

### Setting Up Identity Providers

Connect your identity systems (Okta, Azure AD, Google Workspace, etc.):

1. Create an identity provider via the API: `POST /api/v1/access-reviews/identity-providers`
2. Configure the connection with provider-specific settings
3. Trigger a sync: `POST /api/v1/access-reviews/identity-providers/:id/sync`
4. The sync pulls in users, roles, and access records

Supported providers: Okta, Azure AD, Google Workspace, JumpCloud, OneLogin, and Custom.

---

### Managing Access Resources

Access resources represent the applications, databases, servers, and services that users access:

1. List resources: `GET /api/v1/access-reviews/resources`
2. Create resources manually: `POST /api/v1/access-reviews/resources`
3. Resources synced from identity providers are created automatically

Each resource has a **criticality** level (Low, Medium, High, Critical) that determines review priority.

---

### Running a Review Campaign

A campaign is a periodic review of access across one or more resources:

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

---

### Making Review Decisions

Reviewers decide whether each access grant should continue:

1. View your pending reviews: `GET /api/v1/access-reviews/my-reviews`
2. For each review, make a decision:
   - **Approve** — Access is appropriate and should continue
   - **Revoke** — Access should be removed (requires justification)
   - **Flag** — Access needs further investigation (requires justification)
   - **Delegate** — Pass the review to another reviewer
3. Submit your decision: `POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/decide`

Bulk decisions are supported: `POST /api/v1/access-reviews/campaigns/:id/reviews/bulk-decide`

After revocation decisions, IT Admins mark the actual access removal:
`POST /api/v1/access-reviews/campaigns/:id/reviews/:rid/revocation`

---

## Administration

### Managing Users

1. Go to **Administration > Users** in the sidebar
2. The page shows all users in your organization with their name, email, role, and status

#### Inviting a New User

1. Click **"Invite User"** (CISO, Compliance Manager, or IT Admin)
2. Fill in:
   - **First Name** and **Last Name**
   - **Email**
   - **Role** — Select from the 7 available roles
   - **Temporary Password** — Must be at least 8 characters with complexity requirements
3. Click **"Create User"**

Share the temporary credentials with the new user and ask them to change their password after first login.

#### Deactivating a User

1. Find the user in the list
2. Click the deactivate icon (person with minus)
3. Confirm the action

Deactivated users cannot sign in but their data and audit history are preserved.

#### Reactivating a User

1. Find the deactivated user (they'll have a red "deactivated" badge)
2. Click the reactivate icon (person with checkmark)

> **Who can manage users:** CISO and Compliance Manager can deactivate/reactivate/change roles. IT Admin can create users.

---

### Organization Settings

1. Go to **Administration > Organization** in the sidebar
2. View and (if admin) edit:
   - **Organization Name**
   - **Domain** — Your company's domain (used for SSO matching)
   - **Slug** — URL-safe identifier (read-only)
   - **Status** — Active, Suspended, or Deactivated (read-only)
3. Click **"Save Changes"** after editing

> **Who can edit:** CISO and Compliance Manager only. Other roles see read-only values.

---

### Changing Your Password

1. Go to **Administration > Organization** in the sidebar
2. Scroll down to the **Change Password** card
3. Enter your **Current Password**
4. Enter your **New Password** (8+ characters, must include uppercase, lowercase, number, and special character)
5. **Confirm** the new password
6. Click **"Change Password"**

After a successful change, you'll see a confirmation message. You'll need to sign in again with your new password.

---

## Compliance Posture

The Posture page shows your real-time compliance health across all frameworks:

1. Go to **Risk & Posture > Posture Overview** in the sidebar
2. See the **overall posture score** as a large animated ring gauge:
   - Green (80%+) — Strong compliance posture
   - Amber (60%+) — Moderate compliance posture
   - Red (below 60%) — Weak compliance posture
3. Below, each framework shows:
   - Its own score ring
   - A stacked bar: green (passing tests), red (failing), gray (untested)
   - Counts of passing, failing, and untested controls
   - **Trend indicators**: Improving, Declining, or Stable (comparing to 7-day and 30-day ago scores)

Click **"Refresh"** to get the latest data.

---

## Tips & Best Practices

### Getting Started Checklist

1. **Activate your frameworks** — Start with your primary compliance target (e.g., SOC 2)
2. **Seed controls** — Use the template library when activating frameworks
3. **Map controls to requirements** — Build your compliance coverage
4. **Upload evidence** — Start with existing documentation you already have
5. **Set freshness periods** — So the system alerts you when evidence gets stale
6. **Create policies** — Use templates to get started quickly
7. **Assess your risks** — Build your risk register with initial inherent assessments
8. **Link controls to risks** — Show how controls mitigate your identified risks

### Weekly Routine

- Check **Staleness Alerts** and refresh expired evidence
- Review the **Alert Queue** and address any open alerts
- Check **Policy Approvals** for pending sign-offs
- Review the **Monitoring Dashboard** heatmap for any red squares

### Before an Audit

1. Run the **Audit Readiness** dashboard to see your preparation status
2. Address any **Coverage Gaps** in frameworks
3. Run a full **Test Sweep** to get current monitoring data
4. Review the **Policy Gap Analysis** to ensure all controls have policies
5. Check the **Risk Gap Dashboard** for unmitigated risks
6. Refresh any **stale evidence** that auditors will review

### Role-Based Workflows

**For CISOs and Compliance Managers:**
- Focus on the dashboard, posture overview, and gap analyses
- Manage audit engagements and publish policies
- Review and accept risks that can't be mitigated
- Monitor the compliance posture trends

**For Security Engineers:**
- Create and maintain controls with evidence
- Build and run monitoring tests
- Write and update security policies
- Assess risks and create treatment plans

**For IT Admins:**
- Upload infrastructure evidence (configs, access lists)
- Manage identity providers for access reviews
- Handle alerts related to infrastructure controls
- Submit evidence for audit requests

**For Auditors:**
- Use the Auditor Workspace as your home base
- Create PBC requests using templates
- Review submitted evidence and make accept/reject decisions
- Document findings with severity and recommendations
- Verify remediations before closing findings

---

## Keyboard Shortcuts & UI Tips

- **Search fields** — Press Enter to trigger the search, or wait for the 300ms auto-search
- **Table actions** — Hover over table rows to reveal the action menu (three-dot icon)
- **Status badges** — Color-coded across the app: green = good, amber = attention needed, red = urgent
- **Clickable identifiers** — Monospace-styled identifiers (like `CTRL-AC-001` or `POL-IS-001`) are clickable links to detail pages
- **Back navigation** — Use the arrow icon in page headers to go back to the parent list

---

*This guide covers Raisin Protect through Sprint 8. New features including the Integration Engine (Sprint 9) and Reporting & Executive Dashboard (Sprint 10) are coming soon.*
