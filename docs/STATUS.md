# Raisin Protect — Status

## Current Sprint: 1 — Project Scaffolding & Auth
**Started:** 2026-02-20

## Sprint 1 Tasks

### System Architect
- [ ] Design Sprint 1 schema (organizations, users, roles, sessions)
- [ ] Write Sprint 1 API spec (health, auth, users, org endpoints)
- [ ] Define Docker service topology
- [ ] Write docs/sprints/sprint-1/SCHEMA.md
- [ ] Write docs/sprints/sprint-1/API_SPEC.md

### Database Engineer
- [ ] Write migration: enum types and helper functions
- [ ] Write migration: organizations table
- [ ] Write migration: users table with GRC roles
- [ ] Write migration: refresh_tokens table
- [ ] Write migration: audit_log table
- [ ] Write seed data (demo org, users for each GRC role)

### Backend Developer
- [ ] Go module init, directory structure, Gin setup
- [ ] Config package (env-based)
- [ ] Database connection package
- [ ] Redis connection package
- [ ] JWT auth: register, login, refresh, logout
- [ ] Change password endpoint
- [ ] RBAC middleware (GRC roles from spec §1.2)
- [ ] Organization CRUD
- [ ] User management (list, get, update, deactivate)
- [ ] User role assignment/revocation
- [ ] Health endpoints (/health, /ready)
- [ ] Audit logging middleware
- [ ] Dockerfile
- [ ] docker-compose.yml (postgres, redis, api, dashboard)
- [ ] Unit tests for auth handlers
- [ ] Unit tests for user handlers

### Frontend Developer
- [ ] Next.js 14 project with shadcn/ui + Tailwind
- [ ] Auth pages (login, register)
- [ ] Auth context + token management (lib/auth.ts)
- [ ] App layout with sidebar
- [ ] Role-based navigation
- [ ] Dashboard home (placeholder cards)
- [ ] User management page
- [ ] Organization settings page
- [ ] Dockerfile
- [ ] .dockerignore

### QA Engineer
- [ ] Verify all API tests pass
- [ ] Verify Docker compose starts all services healthy
- [ ] Test auth flow (register → login → refresh → logout)
- [ ] Test RBAC (each role sees correct data)
- [ ] Test dashboard builds and renders

## Blockers
_(none)_

## Agent Activity Log
| Timestamp | Agent | Action |
|-----------|-------|--------|
| 2026-02-20 03:22 | PM | Project plan and status created |
