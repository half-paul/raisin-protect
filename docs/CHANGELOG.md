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

### What's Next
Sprint 2 begins immediately: Core Entities (Frameworks & Controls)
- 7 new tables (frameworks, framework_versions, requirements, controls, control_mappings, org_frameworks, requirement_scopes)
- 300+ control library seeded
- Framework catalog with SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, CCPA
- Cross-framework control mapping matrix
- Coverage gap analysis

---
