# Risk Register

The organizational risk inventory with scoring, treatments, and controls.

## Viewing the Risk Register

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

## Creating a Risk

1. On the Risk Register page, click **"New Risk"**
2. Fill in:
   - **Identifier** — A unique code (e.g., `RISK-T-001`)
   - **Title** — A descriptive name
   - **Description** — Full details about the risk
   - **Category** — Select from 8 categories
   - **Source** — Where this risk was identified
   - **Affected Assets** — Systems or data at risk
3. Save the risk — it's created in **Identified** status

!!! note "Permissions"
    CISO, Compliance Manager, and Security Engineer can create risks.

## Linking Controls to Risks

Show which controls mitigate each risk:

1. Open the risk detail page
2. Go to the **Controls** tab
3. Click **"Link Control"**
4. Search for and select a control
5. Set the **effectiveness**: Effective, Partially Effective, Ineffective, or Not Assessed
6. Set the **mitigation percentage** (0–100%)

## Accepting a Risk

Sometimes the right decision is to accept a risk rather than treat it:

1. Open the risk detail page
2. Change the status to **Accepted**
3. You must provide:
   - **Justification** — Why you're accepting this risk
   - **Expiry Date** — When the acceptance should be reviewed

!!! note "Permissions"
    Only CISO and Compliance Manager can accept risks.
