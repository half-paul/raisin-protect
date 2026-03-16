# Controls

The control library is the backbone of your compliance program. Controls represent the security measures your organization has in place.

## Browsing Controls

1. Go to **Compliance > Controls** in the sidebar
2. The control library shows summary stats at the top: total, active, draft, unmapped, and custom counts
3. Use the filters to narrow down:
   - **Search** — Type a control identifier or title
   - **Status** — Filter by Draft, Active, Under Review, or Deprecated
   - **Category** — Filter by Technical, Administrative, Physical, or Operational
4. Click any control's identifier or title to see its full details

## Creating a Control

1. On the Controls page, click **"New Control"**
2. Fill in the form:
   - **Identifier** — A unique code (e.g., `CTRL-AC-001`)
   - **Title** — A descriptive name (e.g., "Multi-Factor Authentication for Admin Access")
   - **Description** — What this control does and how it's implemented
   - **Category** — Technical, Administrative, Physical, or Operational
   - **Initial Status** — Draft (recommended) or Active
3. Click **"Create"**

!!! note "Permissions"
    CISO, Compliance Manager, and Security Engineer can create controls.

## Mapping Controls to Requirements

Mapping connects your controls to framework requirements, showing which controls satisfy which compliance criteria.

1. Open a control's detail page (click its identifier from the control library)
2. Go to the **Mappings** tab
3. Add a mapping with:
   - **Requirement** — Which framework requirement this control satisfies
   - **Strength** — How strongly it satisfies the requirement:
     - **Primary** — Directly and fully satisfies the requirement
     - **Supporting** — Partially addresses the requirement
     - **Partial** — Contributes to but does not fully cover the requirement

!!! tip
    A single control can map to multiple requirements across different frameworks. This is how Raisin Protect tracks cross-framework coverage.

## Managing Control Status

Controls follow a lifecycle:

```
Draft → Active → Under Review → Active (back)
                              → Deprecated
Deprecated → Draft (reactivate)
```

Available transitions depend on the current status:

- **Draft** → Activate or Deprecate
- **Active** → Mark for Review or Deprecate
- **Under Review** → Reactivate or Deprecate
- **Deprecated** → Revert to Draft

For bulk changes, select multiple controls using checkboxes and click the **"Bulk Status"** button (CISO/Compliance Manager only).
