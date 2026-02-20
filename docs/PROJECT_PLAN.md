# Raisin Protect — GRC Platform Build Plan

## Vision
A Governance, Risk, and Compliance platform that automates evidence collection, monitors controls, manages vendor risk, and maintains audit readiness. Competing with Drata/Vanta.

## Tech Stack (per spec Appendix A, adapted for our tooling)

| Layer | Technology | Port/Config |
|-------|-----------|-------------|
| Frontend | Next.js 14 + TypeScript + Tailwind + shadcn/ui | 3010 |
| Backend API | Go / Gin | 8090 |
| Database | PostgreSQL 16 (row-level security) | 5433 |
| Search | Meilisearch (lightweight alternative to Elastic) | 7700 |
| Cache/Queue | Redis 7 | 6380 |
| Object Storage | MinIO (S3-compatible) | 9000 |
| Worker | Go background jobs | — |
| Docker | docker-compose | — |

## Phase 1 Scope (Foundation — from spec §14)

Focusing on what's buildable as an MVP:

### Sprint 1: Project Scaffolding & Auth
- Go API project structure, config, health checks
- PostgreSQL schema: organizations, users, roles, sessions
- JWT auth: register, login, refresh, logout, change-password
- RBAC middleware with GRC roles (from spec §1.2)
- Docker compose: postgres, redis, api, dashboard
- Next.js dashboard: login, register, auth context
- Basic layout with sidebar navigation

### Sprint 2: Core Entities — Frameworks & Controls
- Schema: frameworks, framework_versions, requirements, controls, control_mappings
- Pre-built framework data: SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, CCPA
- Control library seed (300+ controls mapped to frameworks)
- API: framework CRUD, control CRUD, requirement listing
- API: cross-framework control mapping
- Dashboard: framework list, control library browser, mapping matrix

### Sprint 3: Evidence Management
- Schema: evidence_artifacts, evidence_links, evidence_evaluations
- MinIO integration for file storage
- API: evidence upload, link to controls, version history
- Evidence freshness tracking and staleness alerts
- Dashboard: evidence library, upload interface, linking UI
- Evidence-to-control relationship views

### Sprint 4: Continuous Monitoring Engine
- Schema: tests, test_runs, test_results, alerts, alert_rules
- Test execution worker (background job)
- Alert engine: detect → classify → assign → notify
- Alert delivery: Slack webhook, email (SendGrid/SMTP)
- Dashboard: monitoring view, alert queue, control health heatmap
- Real-time compliance posture score per framework

### Sprint 5: Policy Management
- Schema: policies, policy_versions, policy_signoffs
- Rich text storage for policy content
- Policy templates per framework
- Policy-to-control mapping with gap detection
- Sign-off workflow with approval tracking
- Dashboard: policy editor, template library, sign-off tracking

### Sprint 6: Risk Register
- Schema: risks, risk_assessments, risk_treatments, risk_controls
- Risk scoring (likelihood × impact, configurable formulas)
- Risk heat map visualization
- Risk-to-control linkage
- Treatment plan tracking
- Dashboard: risk register, heat map, treatment progress

### Sprint 7: Audit Hub (Basic)
- Schema: audits, audit_requests, audit_findings, audit_evidence_links
- Auditor workspace with controlled access
- Evidence request/response workflow
- Finding management with remediation tracking
- Dashboard: audit engagement view, request queue

### Sprint 8: User Access Reviews
- Schema: access_reviews, access_review_campaigns, access_entries
- Integration stubs for identity providers
- Review campaign creation and assignment
- Approve/revoke workflow with audit trail
- Dashboard: access review campaigns, decision interface

### Sprint 9: Integration Engine (Foundation)
- Schema: integrations, integration_configs, integration_runs, integration_logs
- Integration framework: base connector class, health monitoring
- 5 starter integrations: AWS Config, GitHub, Okta, Slack, custom webhook
- Connection management UI
- Dashboard: integration status, connection health

### Sprint 10: Reporting & Polish
- Compliance posture reports per framework
- Executive summary dashboard
- PDF report generation
- API documentation (auto-generated OpenAPI spec)
- Security hardening: rate limiting, CORS, CSP headers
- Performance optimization
- End-to-end test suite

---

## Agent Team

### 🏗️ System Architect (SA)
**When:** Start of project, then start of each sprint
**Job:** Design database schemas, write API specs, make architectural decisions
**Output:** `docs/sprints/sprint-N/SCHEMA.md`, `docs/sprints/sprint-N/API_SPEC.md`

### 🗄️ Database Engineer (DBE)  
**When:** After SA delivers sprint schema
**Job:** Write PostgreSQL migrations, seed data, indexes
**Output:** `db/migrations/NNN_*.sql`, `db/seeds/*.sql`

### 👨‍💻 Senior Backend Developer (DEV-BE)
**When:** After DBE delivers migrations (every 2h)
**Job:** Implement Go API handlers, workers, business logic, tests
**Output:** Go code in `api/`, passing tests

### 👩‍💻 Senior Frontend Developer (DEV-FE)
**When:** After DEV-BE has endpoints to consume (every 2h)
**Job:** Build Next.js dashboard pages and components
**Output:** Dashboard code in `dashboard/`, clean build

### 🧪 QA Engineer (QA)
**When:** Every 3h, reviews what's been built
**Job:** Run tests, verify endpoints, file bugs, check for regressions
**Output:** Test results, bug reports in GitHub issues

### 📋 Project Manager (PM)
**When:** Every 6h
**Job:** Track sprint progress, advance to next sprint when done, update STATUS.md
**Output:** Status reports, sprint transitions

---

## Coordination Protocol

### File Structure
```
docs/
├── PROJECT_PLAN.md          — This file
├── STATUS.md                — Current sprint, progress, blockers
├── CHANGELOG.md             — What shipped
└── sprints/
    └── sprint-N/
        ├── SCHEMA.md        — SA's database design
        ├── API_SPEC.md      — SA's endpoint specs
        └── REVIEW.md        — QA's review notes

db/migrations/               — SQL migration files
db/seeds/                    — Seed data

api/                         — Go backend
├── cmd/api/main.go
├── internal/
│   ├── config/
│   ├── db/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   └── workers/
└── Dockerfile

dashboard/                   — Next.js frontend
├── app/
├── components/
├── lib/
└── Dockerfile

docker-compose.yml
.env.example
```

### Agent Rules
1. **Read STATUS.md first** — know the current sprint and what's in progress
2. **Check your dependencies** — SA needs nothing, DBE needs SCHEMA.md, DEV-BE needs migrations, DEV-FE needs API
3. **Do useful work** — If blocked on primary task, do secondary work (refactoring, tests, docs)
4. **Mark progress** — Update TODO items in STATUS.md
5. **Always test** — Run `go test ./...` and `npm run build` before committing
6. **Always commit and push** — Every session should produce a commit
7. **Don't duplicate** — Check git log for recent commits before starting
