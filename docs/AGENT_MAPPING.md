# Agent Mapping — Raisin Protect Project

This document maps the agent tags used in `STATUS.md` to the actual agent team members.

## Agent Tag Mappings

| Tag in STATUS.md | Agent Name | Emoji | Role | agentId (for sessions_send) |
|------------------|------------|-------|------|----------------------------|
| **SA** | David | 🏗️ | System Architect | `david` |
| **DBE** | Dana | 🗄️ | Database Engineer & Data Architect | `dana` |
| **DEV-BE** | Logan | 👨‍💻 | Backend Developer (Go/Node.js) | `logan` |
| **DEV-FE** | Alex | ⚛️ | Frontend Developer (React/Next.js) | `alex` |
| **CR** | Rex | 🔍 | Code Reviewer (Security Audits) | `rex` |
| **QA** | Tom | ⚙️ | QA Engineer (Testing Strategy) | `tom` |
| **PM** | Bruce | 🦞 | PM/Product Owner | `bruce` |
| **DEV** | Logan | 👨‍💻 | General Developer (bugfixes, patches) | `logan` |

## Usage in STATUS.md

When you see entries like:
- `| SA | 4/4 (100%) | ✅ DONE |` → **David** completed System Architect tasks
- `| DBE | 8/8 (100%) | ✅ DONE |` → **Dana** completed Database Engineer tasks
- `| DEV-BE | 14/14 (100%) | ✅ DONE |` → **Logan** completed Backend Developer tasks
- `| DEV-FE | 9/9 (100%) | ✅ DONE |` → **Alex** completed Frontend Developer tasks
- `| CR | 10/10 (100%) | ✅ DONE |` → **Rex** completed Code Reviewer tasks
- `| QA | 9/9 (100%) | ✅ DONE |` → **Tom** completed QA Engineer tasks
- `| PM | Agent lifecycle update |` → **Bruce** performed PM/orchestration tasks

## Agent Communication Examples

To coordinate with agents on Raisin Protect work:

**David (System Architect):**
```
sessions_send(agentId="david", message="Sprint 8 pre-design needed: User Access Reviews. Review the STATUS.md Sprint 8 requirements and create SCHEMA.md + API_SPEC.md.")
```

**Dana (Database Engineer):**
```
sessions_send(agentId="dana", message="Sprint 8 DBE tasks ready: 9 migrations needed (identity_providers, access_resources, access_entries, access_review_campaigns, access_reviews). See docs/sprints/sprint-8/SCHEMA.md for DDL.")
```

**Logan (Backend Developer):**
```
sessions_send(agentId="logan", message="Sprint 8 DEV-BE: Implement 34 REST endpoints for User Access Reviews (IdP CRUD, resource management, review workflow). See API_SPEC.md for full endpoint list.")
```

**Alex (Frontend Developer):**
```
sessions_send(agentId="alex", message="Sprint 8 DEV-FE: Build 9 dashboard pages for User Access Reviews (IdP connections, resource catalog, review campaigns, reviewer workspace). See API_SPEC.md for data contracts.")
```

**Rex (Code Reviewer):**
```
sessions_send(agentId="rex", message="Sprint 8 CR: Review User Access Review implementation. Focus on: multi-tenant isolation in access_entries, RBAC for IdP management, SQL injection prevention, reviewer assignment validation.")
```

**Tom (QA Engineer):**
```
sessions_send(agentId="tom", message="Sprint 8 QA: Test User Access Review workflow end-to-end. Verify: IdP sync, resource listing, campaign creation, review submission, bulk decisions, anomaly detection, certification report generation.")
```

**Bruce (PM):**
```
sessions_send(agentId="bruce", message="Sprint 8 status update: SA/DBE/DEV-BE/DEV-FE/CR all complete. QA in progress (6/9 tasks). Sprint at 89% completion. ETA: tonight.")
```

## Sprint Workflow Pattern

Typical sprint progression in Raisin Protect:

1. **SA (David)** → Designs sprint (SCHEMA.md + API_SPEC.md)
2. **DBE (Dana)** → Writes migrations and seed data
3. **DEV-BE (Logan)** → Implements REST API endpoints
4. **DEV-FE (Alex) + CR (Rex)** → Run in parallel (Alex builds UI, Rex reviews code)
5. **QA (Tom)** → Tests everything (unit, functional, E2E, security)
6. **PM (Bruce)** → Orchestrates agent lifecycle (enables/disables agents per dependency rules)

Once sprint hits 100%, PM transitions to next sprint and the cycle repeats.

## Current Sprint Status

**Sprint 7 (Audit Hub):** 100% complete ✅
- All agents DISABLED
- Ready for Sprint 8 transition

**Sprint 8 (User Access Reviews):** Pre-design complete
- SA (David) completed SCHEMA.md + API_SPEC.md @ 08:04
- Ready for DBE (Dana) to start migrations

**Sprint 9 (Integration Engine):** Pre-design complete
- SA (David) completed SCHEMA.md + API_SPEC.md @ 15:40
- Waiting for Sprint 8 to finish

---

**Last updated:** 2026-02-26 by Mike 🔧
