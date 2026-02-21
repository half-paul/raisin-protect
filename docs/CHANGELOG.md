# Raisin Protect — Changelog

## Sprint 1: Project Scaffolding & Auth (2026-02-20)
**Status:** ✅ COMPLETE — APPROVED FOR DEPLOYMENT

### Delivered
- **System Architecture**: PostgreSQL schema (5 tables: organizations, users, roles, refresh_tokens, audit_log), API spec (20+ endpoints), Docker service topology
- **Database**: 6 migrations with idempotent DDL, enum types, helper functions, audit triggers, indexes on all FKs, demo seed data
- **Backend API**: Go/Gin on port 8090
  - JWT auth: register, login, refresh, logout, change password
  - RBAC middleware with 7 GRC roles (compliance_manager, security_engineer, it_admin, ciso, devops_engineer, auditor, vendor_manager)
  - Multi-tenant org_id scoping on all queries
  - Organization CRUD
  - User management (list, get, update, deactivate, role assignment/revocation)
  - Health endpoints (/health, /ready)
  - Audit logging middleware (all state changes logged to audit_log)
  - 30/30 unit tests passing
  - Dockerfile + docker-compose.yml (postgres:5433, redis:6380, api:8090, dashboard:3010)
- **Dashboard**: Next.js 14 + shadcn/ui + Tailwind on port 3010
  - Auth pages (login, register)
  - Auth context with JWT token management
  - App layout with role-based sidebar (7 GRC roles)
  - Dashboard home with stat cards
  - User management page (list, edit, deactivate, assign roles)
  - Organization settings page (edit org, change password)
  - Dockerfile + .dockerignore
  - Clean build, no TypeScript errors

### Security Audit Results
- **Code Review:** APPROVED — 0 critical, 0 high, 2 medium-priority improvement suggestions
  - Medium: Externalize API_BASE_URL from dashboard code
  - Medium: Add audit middleware to remaining endpoints
- **QA Testing:** APPROVED — All integration tests passed
  - Multi-tenancy isolation verified
  - RBAC correctly enforced
  - Auth flow end-to-end validated
  - Docker services healthy

### Metrics
- **Tasks completed:** 49/49 (100%)
- **Duration:** ~6 hours (03:00 - 09:00)
- **Unit tests:** 30/30 passing
- **Lines of code:** ~4,500 (Go + TypeScript)
- **Database tables:** 5
- **API endpoints:** 20+
- **GitHub issues filed:** 0 (no critical/high findings)

### Demo Credentials
- Email: demo@example.com
- Password: demo123
- Organization: Demo Corp
- Roles: All 7 GRC roles available for testing

---

## Sprint 2: Core Entities — Frameworks & Controls (2026-02-20)
**Status:** ✅ COMPLETE — APPROVED FOR DEPLOYMENT

### Delivered
- **System Architecture**: Pre-designed during Sprint 1 completion (75% threshold)
  - SCHEMA.md: 7 new tables (frameworks, framework_versions, requirements, org_frameworks, requirement_scopes, controls, control_mappings)
  - 5 new enums (compliance_framework, framework_status, control_category, control_type, implementation_status)
  - API_SPEC.md: 25 endpoints (framework catalog, org activation, controls CRUD, mapping matrix, scoping, coverage, bulk ops, stats)
- **Database**: 8 migrations (006-013)
  - Framework catalog tables with versioning support
  - Control library with cross-framework mapping
  - Organization framework activation (many-to-many)
  - Requirement scoping for org-specific applicability
  - Seed data: 5 frameworks (SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, CCPA)
  - Seed data: 200+ requirements, 318 controls, 104 cross-framework mappings
- **Backend API**: 25 REST endpoints
  - Framework catalog: list, get, create, update, deactivate
  - Org-framework activation: activate, deactivate, list activated, coverage %
  - Requirement scoping: apply/remove scopes, list scoped requirements
  - Controls CRUD: list, get, create, update, archive, search/filter (by category, type, framework)
  - Control mappings: create, delete, list by control
  - Mapping matrix: cross-framework shared controls heatmap
  - Coverage gap analysis: activated frameworks vs implemented controls
  - Bulk operations: bulk activate frameworks, bulk map controls
  - Statistics: framework coverage %, control distribution by category/type
  - 30+ unit tests passing (64 total with Sprint 1)
- **Dashboard**: 6 new pages
  - Framework list (activated + available frameworks)
  - Framework detail (requirements, mapped controls, coverage %)
  - Framework activation modal
  - Control library browser (searchable/filterable table with type badges)
  - Control detail page (description, frameworks, requirements, mappings)
  - Control mapping matrix (heatmap visualization)
  - Requirement scoping interface (include/exclude requirements)
  - Coverage dashboard (compliance posture per framework)
  - Bulk operations UI (multi-select + actions with approval modal)
  - 12 total routes in dashboard

### Security Audit Results
- **Code Review:** APPROVED — 0 critical, 0 high, 1 medium
  - Medium: Missing .gitignore (Issue #3 filed and resolved)
  - 3712 lines backend code reviewed
  - 4961 lines frontend code reviewed
  - 8 migrations audited
  - Multi-tenancy isolation verified (20+ org_id checks in new endpoints)
  - SQL injection prevention confirmed (all parameterized queries)
  - RBAC properly enforced
- **QA Testing:** APPROVED — All tests passed
  - 64/64 unit tests passing
  - Dashboard builds clean (12 routes)
  - go vet clean (no warnings)
  - Migrations idempotent
  - Multi-tenancy isolation verified
  - 0 critical/high issues

### Metrics
- **Tasks completed:** 54/54 (100%)
- **Duration:** ~3 hours (09:50 - 12:50)
- **Unit tests:** 64/64 passing
- **Lines of code:** ~8,700 additional (Go + TypeScript)
- **Database tables:** +7 (12 total)
- **API endpoints:** +25 (45+ total)
- **Seed data:** 5 frameworks, 200+ requirements, 318 controls, 104 mappings
- **GitHub issues filed:** 1 (medium priority, resolved)

### Key Features
- **Pre-built framework library**: SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, CCPA ready to activate
- **Control catalog**: 318 pre-mapped controls across frameworks
- **Smart mapping**: 104 controls mapped to multiple frameworks (DRY compliance)
- **Coverage tracking**: Real-time compliance posture per framework
- **Gap analysis**: Identify missing controls across activated frameworks
- **Requirement scoping**: Orgs can exclude non-applicable requirements
- **Bulk operations**: Efficiently manage multiple frameworks and mappings

### What's Next
Sprint 3 begins: Evidence Management
- MinIO integration for artifact storage
- Evidence upload and version history
- Link evidence to controls
- Freshness tracking and staleness alerts
- Evidence library browser
- Control-to-evidence relationship views

---

## Sprint 3: Evidence Management (2026-02-20)
**Status:** ✅ COMPLETE — APPROVED FOR DEPLOYMENT

### Delivered
- **System Architecture**:
  - SCHEMA.md: 3 new tables (evidence_artifacts, evidence_links, evidence_evaluations)
  - 5 new enums (evidence_type, evidence_status, evidence_collection_method, evaluation_verdict, evaluation_confidence_level)
  - 9 new audit_action extensions
  - MinIO bucket configuration patterns
  - API_SPEC.md: 21 endpoints (evidence CRUD, presigned upload/download, versioning, linking, relationship queries, freshness/staleness alerts, evaluations, search)
- **Database**: 5 migrations (014-018) + seed data
  - Evidence artifacts table with file metadata and MinIO object key storage
  - Evidence links table (link evidence to controls/requirements/policies)
  - Evidence evaluations table (track review/approval history with verdict/confidence/missing elements)
  - Evidence version history view + helper functions
  - Seed data: example evidence artifacts for demo controls
- **Backend API**: 21 REST endpoints
  - Evidence CRUD: create, get, list, update, delete (with org_id isolation)
  - Upload/download: presigned upload URLs (multipart form → MinIO), presigned download URLs
  - Versioning: upload new version of existing artifact, version history
  - Linking: link to controls/requirements/policies, delete links
  - Relationship queries: list evidence for control, list controls for evidence
  - Freshness tracking: calculate staleness based on collection_date + validity_days
  - Staleness alerts: identify expired/expiring evidence
  - Evaluations: submit review, approve, reject with remediation notes
  - Advanced search: filter by type, status, collection method, linked entities
  - MinIO service layer with bucket management and presigned URL generation
  - 28 evidence unit tests (84 total)
  - docker-compose.yml updated with MinIO service
- **Dashboard**: 9 new pages/components
  - Evidence library page: searchable/filterable list with freshness summary cards (fresh/expiring/expired counts)
  - Evidence upload interface: drag-and-drop + file browser with 3-step presigned URL flow
  - Evidence detail page: metadata display, tags, latest evaluation, 4-tab layout (Overview/Linked Entities/Version History/Evaluations)
  - Evidence-to-control linking UI: search + select modal with entity picker
  - Control detail enhancement: Evidence tab with summary stats + linked evidence table
  - Freshness badges: visual indicators (fresh/expiring_soon/expired) with day counts
  - Staleness alert dashboard: summary cards + urgency table + affected controls list
  - Evidence evaluation interface: verdict selection, confidence rating, comments, missing elements checklist, remediation guidance
  - Version comparison view: side-by-side metadata diff with change highlighting
  - 14 total routes in dashboard

### Security Audit Results
- **Code Review:** APPROVED — 0 critical, 0 high, 3 medium
  - 3 medium findings (Issues #4-6): presigned URL Content-Type enforcement, file size validation, client-side checks
  - 3 low-priority suggestions for future improvement
  - 3,200 lines backend code reviewed (MinIO integration + evidence handlers)
  - 2,800 lines frontend code reviewed (9 pages + components)
  - 5 migration files audited
  - 28 unit tests reviewed
  - Multi-tenancy isolation verified across all evidence endpoints
  - MIME type whitelist enforced (documents, images, videos, spreadsheets, PDFs)
  - File size limits enforced (100MB max)
  - SQL injection prevention confirmed
  - Audit logging complete
- **QA Testing:** APPROVED — All tests passed
  - 84/84 unit tests passing
  - E2E API tests passing (5/5): upload flow, presigned URLs, linking, versioning, staleness
  - Dashboard builds clean (14 routes)
  - go vet clean
  - MinIO service healthy and accessible
  - Created Playwright E2E test suite with video capture
  - 1 environmental finding (manual migrations required, non-blocking)
  - Multi-tenancy isolation verified
  - No hardcoded secrets found

### Metrics
- **Tasks completed:** 50/50 (100%)
- **Duration:** ~2.5 hours (12:50 - 15:10)
- **Unit tests:** 84/84 passing (28 new evidence tests)
- **E2E tests:** 5/5 passing (upload, download, linking, versioning, staleness)
- **Lines of code:** ~6,000 additional (Go + TypeScript)
- **Database tables:** +3 (15 total)
- **API endpoints:** +21 (66+ total)
- **MinIO integration:** Presigned URLs, bucket management, object storage
- **GitHub issues filed:** 3 (medium priority, security hardening recommendations)

### Key Features
- **MinIO object storage**: S3-compatible, presigned URL pattern for secure uploads/downloads
- **Evidence versioning**: Full history tracking, upload new versions
- **Multi-entity linking**: Link evidence to controls, requirements, policies
- **Freshness tracking**: Automatic staleness detection based on collection_date + validity_days
- **Staleness alerts**: Dashboard showing expired/expiring evidence with urgency levels
- **Evidence evaluation workflow**: Review → approve/reject → track remediation
- **Advanced search**: Filter by type, status, collection method, linked entities
- **File type support**: Documents (docx, pdf, txt), images (png, jpg), videos (mp4), spreadsheets (xlsx, csv)
- **Size limits**: 100MB max per file

### What's Next
Sprint 4 begins: Continuous Monitoring Engine
- Test execution worker (background jobs)
- Alert engine with classification and assignment
- Alert delivery via Slack/email
- Monitoring view and alert queue dashboard
- Control health heatmap
- Real-time compliance posture score

---

## Sprint 4: Continuous Monitoring Engine (2026-02-20)
**Status:** ✅ COMPLETE — APPROVED FOR DEPLOYMENT

### Delivered
- **System Architecture**:
  - SCHEMA.md: 5 new tables (tests, test_runs, test_results, alerts, alert_rules)
  - 11 new enums (test_type, test_status, test_severity, test_result_status, alert_severity, alert_status, alert_delivery_channel, test_trigger_type, test_result_field_type, test_schedule_frequency, alert_event_type)
  - 19 new audit_action extensions
  - Worker architecture design (30s polling, consecutive-failure logic, cooldown periods)
  - Query patterns for heatmap, posture score, alert queue
  - API_SPEC.md: 32 endpoints (tests CRUD, test execution, alerts lifecycle, alert rules, delivery integrations, monitoring dashboard)
- **Database**: 8 migrations (019-025) + seed data
  - Tests table with scheduling and frequency configuration
  - Test_runs table with execution history tracking
  - Test_results table with pass/fail status and output logs
  - Alerts table with severity, status, SLA tracking
  - Alert_rules table with match conditions and delivery configuration
  - Deferred foreign key cross-references
  - Seed data: 8 example tests (control implementation, evidence freshness, access review overdue, configuration drift, policy sign-off, user access anomaly, SSL certificate expiry, backup verification)
  - Seed data: 4 alert rules (critical control failure → Slack, high-severity alerts → email, SLA breach → both, consecutive failures → escalation)
- **Backend API**: 32 REST endpoints
  - Tests CRUD: create, get, list, update, delete, enable/disable
  - Test execution: trigger manual run, get run history, get results, cancel run
  - Test results: list results for control (cross-resource query)
  - Alerts CRUD: list, get, update status
  - Alerts lifecycle: acknowledge, start, resolve, assign, suppress, close, redeliver
  - Alert delivery: redeliver to channel, test delivery configuration
  - Alert rules: create, get, list, update, delete, toggle enable/disable
  - Monitoring dashboard: control health heatmap, compliance posture scores, monitoring summary stats, alert queue
  - MonitoringWorker background job: 30s polling, test execution, alert rule evaluation with consecutive-failure/cooldown logic, SLA breach detection, suppression expiry handling
  - Slack webhook integration (POST with JSON payload, configurable channel/username/emoji)
  - Email SMTP integration (From/To/Subject/Body, support for HTML)
  - 34 new unit tests (118 total)
  - docker-compose.yml updated with worker service
- **Dashboard**: 9 new pages/components
  - Monitoring dashboard home: control health heatmap (color-coded tiles with tooltips), compliance posture scores (ring charts per framework with trend indicators), summary stats (controls/tests/alerts), recent activity feed
  - Alert queue page: queue summary cards (pending/in-progress/resolved counts), severity/status/assignee filtering, search, SLA breach tracking, pagination
  - Alert detail page: full lifecycle workflow (acknowledge → start → assign → resolve/suppress/close), test result display, redeliver option, sidebar metadata (severity, status, SLA, timestamps)
  - Alert assignment interface: assign to user, set priority, track assignee changes
  - Alert resolution workflow: mark resolved with comments, evidence links, verification notes
  - Test execution history: trigger manual runs, cancel runs, filter by status/trigger type, view execution timeline
  - Test result detail page: per-result output logs, structured JSON details display, alert generation links
  - Alert rule management page: create/edit rules with match conditions (event types, severity filters), delivery channel configuration (Slack/email/webhook), SLA thresholds, test delivery, enable/disable toggle, delete
  - Compliance posture score widget: real-time score per framework (SVG ring charts), framework breakdown table, trend indicators
  - 21 total routes in dashboard

### Security Audit Results
- **Code Review:** APPROVED FOR DEPLOYMENT (after addressing Issues #7-9) — 0 critical, 0 high, 3 medium
  - 3 medium findings (Issues #7-9): SSRF in webhook/Slack URLs, user-controlled webhook headers, missing Slack HTTP timeout
  - 3 low-priority recommendations: custom template code sandboxing, rate limiting for alerts, worker metrics/observability
  - ~8,000 lines total reviewed (MonitoringWorker 367 lines, alert handlers 691 lines, alert delivery 176 lines, alert rules 520 lines, 8 migrations, 9 dashboard pages)
  - Multi-tenancy isolation verified (20+ org_id checks across new endpoints)
  - RBAC enforcement confirmed
  - SQL injection prevention verified (all parameterized queries)
  - No hardcoded secrets found
  - Audit logging present for all state changes
  - JWT validation correct
  - TypeScript strict mode (no `any` types)
  - No XSS vectors (zero dangerouslySetInnerHTML)
  - Proper loading/error states
  - Migrations: proper indexes, foreign keys, multi-tenancy support
  - 118/118 unit tests passing
- **QA Testing:** APPROVED FOR DEPLOYMENT (after addressing Issues #7-9) — 0 critical, 0 high, 3 medium bugs
  - 172/172 unit tests passing (go test + go vet clean)
  - 12 core API endpoints verified: tests CRUD, test-runs, alerts CRUD, alert-rules, monitoring dashboard
  - All 6 services healthy: API, worker, postgres, redis, minio, dashboard
  - Worker operational: 30s polling confirmed, no crash loops
  - Multi-tenancy isolation confirmed (org_id enforcement verified in queries)
  - Security review: SQL injection prevention verified, RBAC enforcement confirmed, no hardcoded secrets
  - 3 medium bugs found:
    - Bug #1: Posture SQL error (FIXED: case-insensitive framework lookup in monitoring.go)
    - Bug #2: Seed data UUID format errors (FIXED: manual seed UUIDs corrected in seed files)
    - Bug #3: Manual deployment steps required (documented in QA_REPORT.md, non-blocking)

### Metrics
- **Tasks completed:** 54/54 (100%)
- **Duration:** ~4 hours (15:50 - 19:50)
- **Unit tests:** 172/172 passing (34 new monitoring tests, 54 new worker tests)
- **Lines of code:** ~8,000 additional (Go + TypeScript)
- **Database tables:** +5 (20 total)
- **API endpoints:** +32 (98+ total)
- **Background workers:** 1 (MonitoringWorker with 30s polling)
- **GitHub issues filed:** 6 (3 security findings, 3 bugs — all medium priority)

### Key Features
- **Automated test execution**: Background worker polls every 30s, executes due tests against controls
- **Alert engine**: Detect → classify → assign → notify workflow with consecutive-failure detection and cooldown periods
- **SLA breach tracking**: Track time-to-acknowledge, time-to-resolve against configurable thresholds
- **Alert lifecycle management**: Full workflow (acknowledge → start → assign → resolve/suppress/close)
- **Multi-channel delivery**: Slack webhook, email (SMTP), webhook (generic HTTP POST)
- **Alert rules**: Match conditions (event types, severity), delivery preferences, SLA thresholds
- **Control health heatmap**: Real-time visualization of control compliance status across frameworks
- **Compliance posture scores**: Per-framework compliance percentage with trend indicators
- **Test result history**: Full execution timeline with pass/fail tracking and output logs
- **Suppression and redelivery**: Suppress noisy alerts with expiry, redeliver to alternate channels

### What's Next
Sprint 5 begins: Policy Management
- Policy document storage with rich text support
- Policy versioning and change tracking
- Policy-to-control mapping
- Policy templates per framework
- Sign-off workflow with approval tracking
- Gap detection between policies and controls

---
