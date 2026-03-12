# Evidence Requests (PBC Items)

PBC (Prepared By Client) requests are how auditors ask for specific evidence.

## Creating Evidence Requests

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

!!! note "Permissions"
    CISO, Compliance Manager, and Auditor can create evidence requests.

## Submitting Evidence to a Request

1. Open the request detail page (from the audit's Requests tab)
2. Click **"Link Evidence"**
3. Enter the **Artifact ID** of existing evidence from the Evidence Library
4. Add **Submission Notes** explaining relevance (optional)
5. Click **"Submit Evidence"**
6. Once you've attached all relevant evidence, click **"Submit for Review"** to notify the auditor

The request status progresses: **Open → In Progress → Submitted**

## Reviewing Evidence (Auditors)

As an auditor, when evidence has been submitted:

1. Open the request detail page
2. In the **Chain of Custody** table, find evidence with **"Pending Review"** status
3. Click **"Review"** on each piece of evidence
4. Choose your decision:
   - **Accept** — Evidence is satisfactory
   - **Reject** — Evidence doesn't meet requirements (provide notes)
   - **Needs Clarification** — More information needed
5. Once all evidence is reviewed, click **"Review Request"** to accept or reject the entire request
