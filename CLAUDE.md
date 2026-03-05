# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Raisin Protect is an AI-native GRC (Governance, Risk, and Compliance) platform. Multi-tenant, supporting SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, CCPA. Go backend + Next.js dashboard + PostgreSQL + Redis + MinIO.

## Build & Run

```bash
# Full stack via Docker (non-standard ports to avoid conflicts)
docker-compose up -d
# Frontend: http://localhost:3010 | API: http://localhost:8090
# Postgres: localhost:5434 | Redis: localhost:6380 | MinIO: localhost:9000

# Local backend
cd api && go run cmd/api/main.go

# Local frontend
cd dashboard && npm install && npm run dev
```

## Testing

```bash
# Backend unit tests (195+ tests)
cd api && go test ./...

# Single package
cd api && go test -v ./internal/handlers/

# With coverage
cd api && go test -v -cover ./...

# E2E (Playwright)
cd tests/e2e && npx playwright test
```

## Architecture

### Backend (`api/`)
- **Go 1.24 + Gin framework**, entry point at `cmd/api/main.go`
- `internal/handlers/` — 50+ HTTP handlers (stateless, use package-level DB/JWT clients)
- `internal/middleware/` — Auth (JWT), RBAC, audit logging, rate limiting, CORS, request ID
- `internal/models/` — 22+ domain models with GRC role constants
- `internal/auth/` — JWT generation/validation, bcrypt password hashing
- `internal/config/` — Environment config (all vars prefixed `RP_*`)
- `internal/services/minio.go` — S3-compatible object storage
- `internal/workers/` — Background monitoring worker (30s poll interval)

### Frontend (`dashboard/`)
- **Next.js 14 App Router**, TypeScript strict mode, Tailwind CSS, shadcn/ui + Radix UI
- `app/(dashboard)/` — Protected route group (26+ pages)
- `components/app-shell.tsx` — Main layout with sidebar
- `components/auth-guard.tsx` — Route protection, redirects to login
- `components/ui/` — shadcn/ui primitives
- Path alias: `@/*` maps to project root
- Next.js proxies `/api/*` to backend via `next.config.js`

### Database (`db/`)
- **PostgreSQL 16** with Row-Level Security (RLS) for tenant isolation
- `db/migrations/` — 60+ numbered SQL files, run in order on startup
- `db/seeds/seed.sql` — Demo org, users, frameworks, controls
- Enums defined via PL/pgSQL functions
- All tables scoped by `org_id` foreign key

### API Structure
- Base path: `/api/v1`
- Public: `/auth/register`, `/auth/login`, `/auth/refresh`
- Protected routes require JWT Bearer token
- All responses use `successResponse()` wrapper with metadata
- All mutations logged to `audit_log` table via middleware

## Key Design Patterns

**Multi-tenancy:** Every query must filter by `org_id` extracted from JWT context via `middleware.GetOrgID(c)`. This is enforced at middleware + handler + database (RLS) levels.

**RBAC roles** (defined in `models/user.go`): `ciso`, `compliance_manager`, `security_engineer`, `it_admin`, `devops_engineer`, `auditor`, `vendor_manager`. Middleware enforces via `middleware.RequireRoles(...)`. Frontend role checks are UX-only, not security boundaries.

**System vs user data:** `is_custom=false` means system-provided (read-only). `is_custom=true` or `is_template=false` means user-created. Seed files: `*_seed_templates.sql` (shipped), `*_seed_demo.sql` (dev only).

**Workflow state machines:** Domain objects (policies, risks, audits) have status enums with terminal state guards preventing invalid transitions (e.g., can't edit a published policy).

**Handler pattern:** Handlers extract user/org from context, query DB with parameterized SQL, return via `successResponse(c, data)` or `errorResponse(...)`.

**Auditor isolation:** Auditors only see audits where they're in the `auditor_ids` array, enforced by `audit_access` middleware.

## Conventions

- Go handlers in `internal/handlers/` are tested with `sqlmock` — test files sit alongside handlers
- Schema changes go in new numbered migration files in `db/migrations/`
- HTML content must be sanitized: `bluemonday` (backend), `DOMPurify` (frontend)
- Sprint documentation lives in `docs/sprints/<sprint-N>/` with `SCHEMA.md` and `API_SPEC.md`
- Environment variables use `RP_` prefix (see `api/.env.example`)

## Sprint Roadmap

10-sprint plan documented in `docs/PROJECT_PLAN.md`. Current status tracked in `docs/STATUS.md`. Sprints 1-8 complete (Auth, Frameworks, Controls, Evidence, Monitoring, Policies, Risks, Audit Hub, Access Reviews). Sprint 9 (Integration Engine) is next.
