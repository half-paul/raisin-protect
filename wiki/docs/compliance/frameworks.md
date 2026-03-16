# Compliance Frameworks

Manage your organization's compliance framework activations.

## Activating a Framework

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

!!! note "Permissions"
    Only CISO and Compliance Manager can activate frameworks.

## Browsing Requirements

Each framework has a tree of requirements — the specific criteria you need to satisfy.

1. Go to **Compliance > Frameworks**
2. Click on an activated framework card
3. The framework detail page shows all requirements organized hierarchically (e.g., SOC 2's CC1 through CC9 sections, each with sub-requirements like CC6.1, CC6.2)

Requirements that are **assessable** (leaf-level items) are the ones you need to map controls to.

## Scoping Requirements

Not every requirement may apply to your organization. You can mark requirements as in-scope or out-of-scope.

1. Open a framework's detail page
2. Navigate to the **Scoping** view
3. For each requirement, set whether it's **In Scope** or **Out of Scope**
4. Provide a **justification** for out-of-scope decisions (important for audit evidence)

!!! note "Permissions"
    Only CISO and Compliance Manager can scope requirements.
