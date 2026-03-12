# Monitoring Dashboard

The monitoring dashboard gives you a real-time view of your compliance posture.

## Overview

1. Go to **Monitoring > Monitoring** in the sidebar
2. At the top you'll see:
   - **Posture Score** — Overall compliance health percentage
   - **Test Pass Rate (24h)** — How many tests passed in the last 24 hours
   - **Open Alerts** — Active alerts requiring attention
   - **SLA Breached** — Alerts that have exceeded their response deadline

## Control Health Heatmap

The heatmap is the central visualization on the monitoring dashboard. It shows every control as a colored square:

| Color | Meaning |
|-------|---------|
| **Green** | Healthy — All tests passing |
| **Red** | Failing — One or more tests failing |
| **Orange** | Error — Tests encountering errors |
| **Amber** | Warning — Tests producing warnings |
| **Gray** | Untested — No tests configured for this control |

### How to use it

- **Hover** over any square to see the control's identifier, title, health status, test count, active alerts, and when it was last tested
- **Click** any square to jump to that control's detail page
- **Filter by category** using the dropdown above the heatmap (All, Technical, Administrative, Physical, Operational)

Below the heatmap, you'll see:

- **Compliance Posture** — Per-framework scores with trend indicators (Improving, Declining, or Stable)
- **Recent Activity** — Latest alerts created, resolved, and test runs completed
- **Quick Actions** — Links to the alert queue, test history, and alert rules
