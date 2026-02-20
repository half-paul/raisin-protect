# Sprint 1 — Database Schema

## Overview

Sprint 1 establishes the foundational data model: multi-tenant organizations, users with GRC roles, authentication (JWT + refresh tokens), and an immutable audit log. All tables use UUIDs for primary keys and include `created_at`/`updated_at` timestamps. Multi-tenancy is enforced via `org_id` foreign keys on all tenant-scoped tables.

---

## Entity Relationship Diagram (Text)

```
organizations 1──∞ users
users         1──∞ refresh_tokens
users         1──∞ audit_log
organizations 1──∞ audit_log
```

---

## Enums

```sql
-- GRC roles from spec §1.2
CREATE TYPE grc_role AS ENUM (
    'compliance_manager',
    'security_engineer',
    'it_admin',
    'ciso',
    'devops_engineer',
    'auditor',
    'vendor_manager'
);

-- Organization-level status
CREATE TYPE org_status AS ENUM (
    'active',
    'suspended',
    'deactivated'
);

-- User account status
CREATE TYPE user_status AS ENUM (
    'active',
    'invited',
    'deactivated',
    'locked'
);

-- Audit log action categories
CREATE TYPE audit_action AS ENUM (
    'user.register',
    'user.login',
    'user.logout',
    'user.login_failed',
    'user.password_changed',
    'user.updated',
    'user.deactivated',
    'user.reactivated',
    'user.role_assigned',
    'user.role_revoked',
    'org.created',
    'org.updated',
    'org.suspended',
    'org.deactivated',
    'token.refreshed',
    'token.revoked'
);
```

---

## Helper Functions

```sql
-- Auto-update updated_at on row modification
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Generate a short, human-readable org slug from name
CREATE OR REPLACE FUNCTION generate_slug(input TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN lower(regexp_replace(trim(input), '[^a-zA-Z0-9]+', '-', 'g'));
END;
$$ LANGUAGE plpgsql;
```

---

## Tables

### organizations

The root tenant entity. Every other tenant-scoped table references `organizations.id`.

```sql
CREATE TABLE organizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) NOT NULL UNIQUE,
    domain          VARCHAR(255),                          -- optional: for SSO domain matching
    status          org_status NOT NULL DEFAULT 'active',
    settings        JSONB NOT NULL DEFAULT '{}',           -- org-level config (timezone, locale, etc.)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_organizations_slug ON organizations (slug);
CREATE INDEX idx_organizations_status ON organizations (status);
CREATE INDEX idx_organizations_domain ON organizations (domain) WHERE domain IS NOT NULL;

-- Trigger
CREATE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `slug` provides URL-friendly org identifiers (e.g., `/orgs/acme-corp`)
- `domain` supports future SSO auto-provisioning (match email domain → org)
- `settings` JSONB is extensible without migrations for org-level preferences

---

### users

Users belong to exactly one organization. Each user has a single GRC role (simplifying RBAC for Sprint 1; can be extended to many-to-many in future sprints).

```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    role            grc_role NOT NULL DEFAULT 'compliance_manager',
    status          user_status NOT NULL DEFAULT 'active',
    mfa_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret      VARCHAR(255),                          -- TOTP secret (encrypted at app level)
    last_login_at   TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique email per org (same person can exist in different orgs)
    CONSTRAINT uq_users_org_email UNIQUE (org_id, email)
);

-- Indexes
CREATE INDEX idx_users_org_id ON users (org_id);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (org_id, role);
CREATE INDEX idx_users_status ON users (org_id, status);

-- Trigger
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- Email uniqueness is scoped to org (`uq_users_org_email`), not global — same person may use same email across multiple tenants
- `password_hash` stores bcrypt output (app layer handles hashing)
- `mfa_secret` is stored encrypted at the application level; Sprint 1 includes the column but MFA enforcement comes later
- Single `role` column for Sprint 1 simplicity; future sprints may introduce a `user_roles` junction table for multi-role support

---

### refresh_tokens

Stores active refresh tokens for JWT rotation. Tokens are single-use: on refresh, the old token is revoked and a new one issued.

```sql
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,          -- SHA-256 of the actual token
    user_agent      VARCHAR(500),                          -- browser/client identifier
    ip_address      INET,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,                           -- null = active, set = revoked
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at)
    WHERE revoked_at IS NULL;

-- Cleanup: expired/revoked tokens older than 30 days (run via scheduled job)
-- DELETE FROM refresh_tokens WHERE (revoked_at IS NOT NULL OR expires_at < NOW()) AND created_at < NOW() - INTERVAL '30 days';
```

**Design notes:**
- We store `token_hash` (SHA-256), never the raw token — the raw token is only sent to the client
- `revoked_at` pattern: NULL means active, timestamp means revoked (supports both explicit logout and token rotation)
- `user_agent` and `ip_address` enable session management UI and security auditing
- `org_id` is denormalized (could derive from user_id) for efficient org-scoped queries and cleanup

---

### audit_log

Immutable, append-only log of all security-relevant actions. This is the backbone of compliance evidence for the platform itself.

```sql
CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id        UUID REFERENCES users(id) ON DELETE SET NULL,  -- null for system actions
    action          audit_action NOT NULL,
    resource_type   VARCHAR(50) NOT NULL,                  -- 'user', 'organization', 'token'
    resource_id     UUID,                                  -- ID of affected resource
    metadata        JSONB NOT NULL DEFAULT '{}',           -- action-specific details
    ip_address      INET,
    user_agent      VARCHAR(500),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- No updated_at: audit logs are immutable
);

-- Indexes for common query patterns
CREATE INDEX idx_audit_log_org_id ON audit_log (org_id);
CREATE INDEX idx_audit_log_actor_id ON audit_log (actor_id);
CREATE INDEX idx_audit_log_action ON audit_log (org_id, action);
CREATE INDEX idx_audit_log_resource ON audit_log (org_id, resource_type, resource_id);
CREATE INDEX idx_audit_log_created_at ON audit_log (org_id, created_at DESC);

-- Partitioning hint: for production, partition by created_at (monthly) for performance
-- CREATE TABLE audit_log (...) PARTITION BY RANGE (created_at);
```

**Design notes:**
- `actor_id` is nullable with `ON DELETE SET NULL` — preserves audit history even if user is deleted
- `metadata` JSONB captures action-specific context (e.g., `{"old_role": "it_admin", "new_role": "ciso"}`, `{"ip": "...", "reason": "..."}`)
- No `updated_at` — audit logs must be immutable per spec §11.3
- `resource_type` + `resource_id` enable generic "what was affected" queries
- Partitioning by `created_at` recommended for production (spec §11.3: 7-year retention default)

---

## Docker Service Topology

Sprint 1 establishes 4 core services plus the network topology for the entire project:

```
┌─────────────────────────────────────────────────────┐
│                  Docker Network: rp-net              │
│                                                     │
│  ┌───────────┐  ┌───────────┐  ┌────────────────┐  │
│  │ PostgreSQL │  │   Redis   │  │    MinIO        │  │
│  │  :5433     │  │   :6380   │  │  :9000 (S3)    │  │
│  │  (db)      │  │  (cache)  │  │  :9001 (console│  │
│  └─────┬─────┘  └─────┬─────┘  └───────┬────────┘  │
│        │               │                │            │
│  ┌─────┴───────────────┴────────────────┴──────┐    │
│  │                  API (Go/Gin)                │    │
│  │                  :8090                       │    │
│  │  Depends: postgres, redis                    │    │
│  └──────────────────────┬──────────────────────┘    │
│                          │                           │
│  ┌──────────────────────┴──────────────────────┐    │
│  │             Dashboard (Next.js)              │    │
│  │             :3010                            │    │
│  │  Depends: api                                │    │
│  └─────────────────────────────────────────────┘    │
│                                                     │
│  ┌─────────────────────────────────────────────┐    │
│  │           Meilisearch (Sprint 2+)           │    │
│  │           :7700                              │    │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

### Service Details

| Service | Image | Host Port | Container Port | Volume |
|---------|-------|-----------|----------------|--------|
| postgres | postgres:16-alpine | 5433 | 5432 | rp_postgres_data |
| redis | redis:7-alpine | 6380 | 6379 | rp_redis_data |
| minio | minio/minio:latest | 9000, 9001 | 9000, 9001 | rp_minio_data |
| api | Custom Dockerfile | 8090 | 8090 | — |
| dashboard | Custom Dockerfile | 3010 | 3000 | — |

**Port selection rationale:** Non-default ports (5433, 6380, 8090, 3010) avoid conflicts with Raisin Shield which uses 5432, 6379, 8080, 3000.

### Health Checks

- **postgres:** `pg_isready -U rp -d raisin_protect`
- **redis:** `redis-cli ping`
- **minio:** `mc ready local` or HTTP GET on `:9000/minio/health/live`
- **api:** `GET /health` returns 200
- **dashboard:** HTTP GET on `:3000` returns 200

### Environment Variables (.env.example)

```env
# PostgreSQL
POSTGRES_USER=rp
POSTGRES_PASSWORD=<generate-strong-password>
POSTGRES_DB=raisin_protect
POSTGRES_PORT=5433

# Redis
REDIS_PORT=6380

# MinIO
MINIO_ROOT_USER=rp-admin
MINIO_ROOT_PASSWORD=<generate-strong-password>
MINIO_BUCKET=rp-evidence

# API
API_PORT=8090
JWT_SECRET=<generate-32-char-secret>
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=7d
BCRYPT_COST=12
CORS_ORIGINS=http://localhost:3010

# Dashboard
NEXT_PUBLIC_API_URL=http://localhost:8090
DASHBOARD_PORT=3010
```

---

## Migration Order

Migrations should be numbered and applied in this order:

1. `001_enums_and_functions.sql` — All enum types + helper functions
2. `002_organizations.sql` — Organizations table + indexes + trigger
3. `003_users.sql` — Users table + indexes + trigger
4. `004_refresh_tokens.sql` — Refresh tokens table + indexes
5. `005_audit_log.sql` — Audit log table + indexes

---

## Seed Data

### Demo Organization

```sql
INSERT INTO organizations (id, name, slug, domain, status) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'Acme Corporation', 'acme-corp', 'acme.example.com', 'active');
```

### Demo Users (one per GRC role)

All demo users use password: `demo123` (bcrypt hash generated at seed time)

| Email | Role | Name |
|-------|------|------|
| compliance@acme.example.com | compliance_manager | Alice Compliance |
| security@acme.example.com | security_engineer | Bob Security |
| it@acme.example.com | it_admin | Carol IT |
| ciso@acme.example.com | ciso | David CISO |
| devops@acme.example.com | devops_engineer | Eve DevOps |
| auditor@acme.example.com | auditor | Frank Auditor |
| vendor@acme.example.com | vendor_manager | Grace Vendor |

---

## Future Considerations

- **Row-Level Security (RLS):** Spec §11.1 calls for RLS. For Sprint 1, org isolation is enforced at the application layer (middleware injects `org_id`). RLS policies will be added in a hardening sprint.
- **Multi-role users:** Current schema uses a single `role` column. If users need multiple roles, migrate to a `user_roles` junction table.
- **Audit log partitioning:** For production with 7-year retention, partition `audit_log` by month.
- **Password history:** A `password_history` table can be added to prevent password reuse.
