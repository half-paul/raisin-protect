# Raisin Protect — GRC Platform

## Project Overview
Raisin Protect is an AI-native Governance, Risk, and Compliance (GRC) platform designed to automate evidence collection, monitor security controls, manage vendor risk, and maintain audit readiness. It is built as a multi-tenant solution supporting frameworks like SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, and more.

The project is structured as a monorepo containing a high-performance Go backend and a modern React/Next.js dashboard.

### Core Tech Stack
- **Frontend**: Next.js 14 (App Router), TypeScript, Tailwind CSS, shadcn/ui, Lucide Icons.
- **Backend**: Go 1.24, Gin (Web Framework), JWT Authentication, RBAC.
- **Database**: PostgreSQL 16 with Row-Level Security (RLS) for tenant isolation.
- **Caching/Queue**: Redis 7.
- **Object Storage**: MinIO (S3-compatible) for evidence artifact storage.
- **Testing**: Go `testing` package (backend), Playwright (E2E), `sqlmock` for DB tests.
- **Infrastructure**: Docker Compose, GitHub Actions (CI/CD).

## Building and Running

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development)
- Node.js 20+ and npm (for local development)

### Quick Start (Docker)
The entire stack can be launched using Docker Compose. Note that it uses non-standard ports to avoid conflicts:
- **Frontend**: http://localhost:3010
- **API**: http://localhost:8090
- **Postgres**: localhost:5434
- **Redis**: localhost:6380
- **MinIO**: http://localhost:9000 (API) / http://localhost:9001 (Console)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

### Local Development
Each component can be run independently for faster development cycles.

#### Backend (API)
```bash
cd api
# Copy .env.example to .env and adjust if needed
go run cmd/api/main.go
```

#### Frontend (Dashboard)
```bash
cd dashboard
npm install
npm run dev
```

### Testing
- **Backend Unit Tests**: `cd api && go test ./...`
- **E2E Tests**: `cd tests/e2e && npm install && npx playwright test`

## Development Conventions

### Architecture & Design
- **Multi-tenancy**: All data must be scoped by `org_id`. Ensure the `org_id` is extracted from the JWT context and used in all database queries.
- **RBAC**: Enforce Role-Based Access Control using the defined GRC roles (e.g., `compliance_manager`, `auditor`, `ciso`).
- **Security**:
  - Always sanitize HTML content (using `bluemonday` on backend or `DOMPurify` on frontend).
  - Use parameterized SQL queries to prevent injection.
  - Implement terminal state guards for workflow transitions (e.g., cannot edit a published policy).

### Coding Standards
- **Go**: Follow standard Go idioms. Use the `internal/` directory for private logic. Handlers are located in `api/internal/handlers/`.
- **TypeScript**: Strict typing is mandatory. Avoid using `any`. Use functional components with React 18+ hooks.
- **API**: All state-changing operations must be logged to the `audit_log` via middleware.
- **Database**: All schema changes must be implemented via SQL migrations in `db/migrations/`.

### Documentation
The project maintains a rigorous documentation standard in the `docs/` directory:
- `PROJECT_PLAN.md`: The 10-sprint delivery roadmap.
- `STATUS.md`: Current sprint progress and blockers.
- `CHANGELOG.md`: Detailed history of delivered features and security audits.
- `sprints/`: Sprint-specific designs (`SCHEMA.md`, `API_SPEC.md`) and review reports.

## Current Project Status
As of March 2026, the project has successfully completed **Sprint 7 (Audit Hub)**. All core modules including Frameworks, Controls, Evidence, Monitoring, Policies, Risks, and Audit Hub are functional. Sprint 8 (User Access Reviews) is in the design phase.
