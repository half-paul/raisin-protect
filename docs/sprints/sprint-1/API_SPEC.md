# Sprint 1 — API Specification

## Overview

Sprint 1 API provides: health checks, authentication (register/login/refresh/logout/change-password), organization management, user management, and role assignment. All endpoints return JSON. Auth uses JWT access tokens (15min) + refresh tokens (7 days).

**Base URL:** `http://localhost:8090/api/v1`

---

## Authentication Model

### JWT Access Token

- **Location:** `Authorization: Bearer <token>` header
- **Lifetime:** 15 minutes
- **Payload claims:**

```json
{
  "sub": "<user_id>",
  "org": "<org_id>",
  "role": "compliance_manager",
  "email": "user@example.com",
  "exp": 1708500000,
  "iat": 1708499100
}
```

### Refresh Token

- **Location:** Response body (stored client-side, sent in request body)
- **Lifetime:** 7 days
- **Single-use:** Each refresh issues a new token pair and revokes the old refresh token
- **Storage:** SHA-256 hash stored in `refresh_tokens` table

### Role-Based Access Control (RBAC)

Sprint 1 RBAC is role-checked middleware. Endpoints declare which roles can access them.

| Role | Scope |
|------|-------|
| `ciso` | Full access to all resources within their org |
| `compliance_manager` | Full access to all resources within their org |
| `security_engineer` | Read all; write controls, monitoring, evidence |
| `it_admin` | Read all; write integrations, users, access reviews |
| `devops_engineer` | Read all; write integrations, monitoring |
| `auditor` | Read-only access to all resources (external auditor role) |
| `vendor_manager` | Read all; write vendor assessments |

For Sprint 1, `ciso` and `compliance_manager` are admin-level roles with full CRUD on all resources. Other roles have restricted write access as listed.

---

## Common Response Patterns

### Success Response

```json
{
  "data": { ... },
  "meta": {
    "request_id": "uuid"
  }
}
```

### List Response (Paginated)

```json
{
  "data": [ ... ],
  "meta": {
    "total": 42,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": [
      { "field": "email", "message": "must be a valid email address" }
    ]
  },
  "meta": {
    "request_id": "uuid"
  }
}
```

### Standard Error Codes

| HTTP Status | Code | When |
|-------------|------|------|
| 400 | `VALIDATION_ERROR` | Invalid request body or params |
| 401 | `UNAUTHORIZED` | Missing or invalid token |
| 403 | `FORBIDDEN` | Valid token but insufficient role |
| 404 | `NOT_FOUND` | Resource doesn't exist (or not in caller's org) |
| 409 | `CONFLICT` | Duplicate resource (e.g., email already exists) |
| 422 | `UNPROCESSABLE` | Business logic rejection |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |

---

## Endpoints

---

### 1. Health

#### `GET /health`

Liveness check. Always returns 200 if the API process is running.

- **Auth:** None
- **Roles:** Public

**Response 200:**

```json
{
  "status": "ok",
  "version": "0.1.0",
  "timestamp": "2026-02-20T10:00:00Z"
}
```

---

#### `GET /ready`

Readiness check. Returns 200 only if all dependencies (PostgreSQL, Redis) are reachable.

- **Auth:** None
- **Roles:** Public

**Response 200:**

```json
{
  "status": "ready",
  "checks": {
    "postgres": "ok",
    "redis": "ok"
  },
  "timestamp": "2026-02-20T10:00:00Z"
}
```

**Response 503:**

```json
{
  "status": "not_ready",
  "checks": {
    "postgres": "ok",
    "redis": "error: connection refused"
  },
  "timestamp": "2026-02-20T10:00:00Z"
}
```

---

### 2. Authentication

#### `POST /api/v1/auth/register`

Register a new user and create their organization. This is the onboarding entry point.

- **Auth:** None
- **Roles:** Public

**Request:**

```json
{
  "email": "alice@acme.com",
  "password": "SecureP@ss123",
  "first_name": "Alice",
  "last_name": "Compliance",
  "org_name": "Acme Corporation"
}
```

**Validation:**
- `email`: required, valid email format, max 255 chars
- `password`: required, min 8 chars, must contain uppercase, lowercase, number, and special character
- `first_name`: required, max 100 chars
- `last_name`: required, max 100 chars
- `org_name`: required, max 255 chars

**Response 201:**

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "alice@acme.com",
      "first_name": "Alice",
      "last_name": "Compliance",
      "role": "compliance_manager",
      "status": "active",
      "created_at": "2026-02-20T10:00:00Z"
    },
    "organization": {
      "id": "uuid",
      "name": "Acme Corporation",
      "slug": "acme-corporation",
      "status": "active",
      "created_at": "2026-02-20T10:00:00Z"
    },
    "access_token": "eyJhbG...",
    "refresh_token": "dGhpcyBpcyBhIHJl...",
    "expires_in": 900
  }
}
```

**Errors:**
- `409 CONFLICT` — Email already registered in an existing org

**Audit log:** `user.register`

**Notes:**
- The registering user becomes `compliance_manager` (admin-level) of the new org
- Org slug is auto-generated from `org_name`
- Both access and refresh tokens are returned immediately (user is logged in)

---

#### `POST /api/v1/auth/login`

Authenticate with email/password. Returns token pair.

- **Auth:** None
- **Roles:** Public

**Request:**

```json
{
  "email": "alice@acme.com",
  "password": "SecureP@ss123"
}
```

**Validation:**
- `email`: required
- `password`: required

**Response 200:**

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "alice@acme.com",
      "first_name": "Alice",
      "last_name": "Compliance",
      "role": "compliance_manager",
      "org_id": "uuid",
      "status": "active"
    },
    "access_token": "eyJhbG...",
    "refresh_token": "dGhpcyBpcyBhIHJl...",
    "expires_in": 900
  }
}
```

**Errors:**
- `401 UNAUTHORIZED` — Invalid email or password (do not reveal which)

**Audit log:** `user.login` on success, `user.login_failed` on failure

**Notes:**
- Updates `users.last_login_at`
- Logs IP address and user agent in audit metadata
- Locked/deactivated users get 401 with message "Account is not active"

---

#### `POST /api/v1/auth/refresh`

Exchange a valid refresh token for a new access token + refresh token pair.

- **Auth:** None (refresh token in body)
- **Roles:** Public

**Request:**

```json
{
  "refresh_token": "dGhpcyBpcyBhIHJl..."
}
```

**Response 200:**

```json
{
  "data": {
    "access_token": "eyJhbG...",
    "refresh_token": "bmV3IHJlZnJlc2g...",
    "expires_in": 900
  }
}
```

**Errors:**
- `401 UNAUTHORIZED` — Token invalid, expired, or already revoked

**Audit log:** `token.refreshed`

**Notes:**
- Old refresh token is immediately revoked (`revoked_at` set)
- New refresh token inherits the original expiry window (not extended)
- If a revoked token is used, revoke ALL tokens for that user (theft detection)

---

#### `POST /api/v1/auth/logout`

Revoke the current refresh token.

- **Auth:** Bearer token required
- **Roles:** Any authenticated user

**Request:**

```json
{
  "refresh_token": "dGhpcyBpcyBhIHJl..."
}
```

**Response 200:**

```json
{
  "data": {
    "message": "Logged out successfully"
  }
}
```

**Errors:**
- `401 UNAUTHORIZED` — Invalid access token

**Audit log:** `user.logout`

**Notes:**
- Revokes the provided refresh token
- Access token remains valid until natural expiry (15min max)
- Client should discard both tokens

---

#### `POST /api/v1/auth/change-password`

Change the authenticated user's password.

- **Auth:** Bearer token required
- **Roles:** Any authenticated user

**Request:**

```json
{
  "current_password": "OldP@ss123",
  "new_password": "NewSecureP@ss456"
}
```

**Validation:**
- `current_password`: required, must match current hash
- `new_password`: required, min 8 chars, complexity requirements, must differ from current

**Response 200:**

```json
{
  "data": {
    "message": "Password changed successfully"
  }
}
```

**Errors:**
- `401 UNAUTHORIZED` — Current password incorrect
- `422 UNPROCESSABLE` — New password doesn't meet requirements or same as current

**Audit log:** `user.password_changed`

**Notes:**
- Revokes ALL refresh tokens for the user (forces re-login on all devices)
- Updates `users.password_changed_at`

---

### 3. Organizations

#### `GET /api/v1/organizations/current`

Get the authenticated user's organization.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "domain": "acme.example.com",
    "status": "active",
    "settings": {},
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

---

#### `PUT /api/v1/organizations/current`

Update the authenticated user's organization.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "name": "Acme Corp International",
  "domain": "acme-intl.com",
  "settings": {
    "timezone": "America/New_York",
    "locale": "en-US"
  }
}
```

**Validation:**
- `name`: optional, max 255 chars
- `domain`: optional, valid domain format, max 255 chars
- `settings`: optional, valid JSON object

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "name": "Acme Corp International",
    "slug": "acme-corp",
    "domain": "acme-intl.com",
    "status": "active",
    "settings": {
      "timezone": "America/New_York",
      "locale": "en-US"
    },
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:30:00Z"
  }
}
```

**Errors:**
- `403 FORBIDDEN` — Role not authorized

**Audit log:** `org.updated` with changed fields in metadata

**Notes:**
- Slug is NOT updatable (used in URLs, integrations)
- `settings` is merged (not replaced) — send only the keys to update

---

### 4. Users

#### `GET /api/v1/users`

List all users in the authenticated user's organization.

- **Auth:** Bearer token required
- **Roles:** All roles

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Items per page (max 100) |
| `status` | string | — | Filter by status: `active`, `invited`, `deactivated`, `locked` |
| `role` | string | — | Filter by GRC role |
| `search` | string | — | Search by name or email (case-insensitive, partial match) |
| `sort` | string | `created_at` | Sort field: `created_at`, `email`, `last_name`, `role`, `last_login_at` |
| `order` | string | `desc` | Sort order: `asc`, `desc` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "email": "alice@acme.com",
      "first_name": "Alice",
      "last_name": "Compliance",
      "role": "compliance_manager",
      "status": "active",
      "mfa_enabled": false,
      "last_login_at": "2026-02-20T09:00:00Z",
      "created_at": "2026-02-20T08:00:00Z",
      "updated_at": "2026-02-20T09:00:00Z"
    }
  ],
  "meta": {
    "total": 7,
    "page": 1,
    "per_page": 20,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Never returns `password_hash` or `mfa_secret`
- Results are automatically scoped to the caller's `org_id`

---

#### `GET /api/v1/users/:id`

Get a specific user by ID.

- **Auth:** Bearer token required
- **Roles:** All roles

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "email": "bob@acme.com",
    "first_name": "Bob",
    "last_name": "Security",
    "role": "security_engineer",
    "status": "active",
    "mfa_enabled": false,
    "last_login_at": "2026-02-20T09:00:00Z",
    "created_at": "2026-02-20T08:00:00Z",
    "updated_at": "2026-02-20T09:00:00Z"
  }
}
```

**Errors:**
- `404 NOT_FOUND` — User doesn't exist or belongs to a different org

---

#### `POST /api/v1/users`

Create (invite) a new user in the organization.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `it_admin`

**Request:**

```json
{
  "email": "newuser@acme.com",
  "first_name": "New",
  "last_name": "User",
  "role": "security_engineer",
  "password": "TempP@ss123"
}
```

**Validation:**
- `email`: required, valid email, unique within org
- `first_name`: required, max 100 chars
- `last_name`: required, max 100 chars
- `role`: required, must be a valid `grc_role`
- `password`: required, min 8 chars, complexity requirements

**Response 201:**

```json
{
  "data": {
    "id": "uuid",
    "email": "newuser@acme.com",
    "first_name": "New",
    "last_name": "User",
    "role": "security_engineer",
    "status": "invited",
    "mfa_enabled": false,
    "created_at": "2026-02-20T10:00:00Z",
    "updated_at": "2026-02-20T10:00:00Z"
  }
}
```

**Errors:**
- `409 CONFLICT` — Email already exists in this org
- `403 FORBIDDEN` — Caller's role cannot create users

**Audit log:** `user.register` with `invited_by` in metadata

**Notes:**
- New user status is `invited` (transitions to `active` on first login)
- Sprint 1 uses a temporary password; email invite flow can be added later

---

#### `PUT /api/v1/users/:id`

Update a user's profile information.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `it_admin` (or the user themselves for own profile)

**Request:**

```json
{
  "first_name": "Robert",
  "last_name": "Security-Lead",
  "email": "bob.lead@acme.com"
}
```

**Validation:**
- `first_name`: optional, max 100 chars
- `last_name`: optional, max 100 chars
- `email`: optional, valid email, unique within org

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "email": "bob.lead@acme.com",
    "first_name": "Robert",
    "last_name": "Security-Lead",
    "role": "security_engineer",
    "status": "active",
    "mfa_enabled": false,
    "last_login_at": "2026-02-20T09:00:00Z",
    "created_at": "2026-02-20T08:00:00Z",
    "updated_at": "2026-02-20T10:30:00Z"
  }
}
```

**Errors:**
- `403 FORBIDDEN` — Not authorized to update this user
- `404 NOT_FOUND` — User not found in this org
- `409 CONFLICT` — New email already in use

**Audit log:** `user.updated` with changed fields in metadata

**Notes:**
- Users can update their own `first_name`, `last_name`, `email`
- Role and status changes use dedicated endpoints (below)
- Admins can update any user in their org

---

#### `POST /api/v1/users/:id/deactivate`

Deactivate a user account.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "deactivated",
    "message": "User deactivated successfully"
  }
}
```

**Errors:**
- `403 FORBIDDEN` — Not authorized
- `404 NOT_FOUND` — User not found
- `422 UNPROCESSABLE` — Cannot deactivate yourself

**Audit log:** `user.deactivated`

**Notes:**
- Revokes ALL refresh tokens for the deactivated user
- Deactivated users cannot log in
- Does NOT delete the user (preserves audit trail)

---

#### `POST /api/v1/users/:id/reactivate`

Reactivate a previously deactivated user.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "status": "active",
    "message": "User reactivated successfully"
  }
}
```

**Errors:**
- `403 FORBIDDEN` — Not authorized
- `404 NOT_FOUND` — User not found
- `422 UNPROCESSABLE` — User is not deactivated

**Audit log:** `user.reactivated`

---

### 5. Role Management

#### `PUT /api/v1/users/:id/role`

Change a user's GRC role.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`

**Request:**

```json
{
  "role": "ciso"
}
```

**Validation:**
- `role`: required, must be a valid `grc_role`

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "email": "david@acme.com",
    "role": "ciso",
    "previous_role": "security_engineer",
    "message": "Role updated successfully"
  }
}
```

**Errors:**
- `403 FORBIDDEN` — Not authorized
- `404 NOT_FOUND` — User not found
- `422 UNPROCESSABLE` — Cannot change your own role / same role as current

**Audit log:** `user.role_assigned` with `{"old_role": "...", "new_role": "..."}`

**Notes:**
- Only `ciso` and `compliance_manager` can change roles
- Users cannot change their own role (prevents privilege escalation)
- Single role per user in Sprint 1

---

### 6. Audit Log

#### `GET /api/v1/audit-log`

Query audit log entries for the organization.

- **Auth:** Bearer token required
- **Roles:** `ciso`, `compliance_manager`, `auditor`

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 50 | Items per page (max 200) |
| `action` | string | — | Filter by action type (e.g., `user.login`) |
| `actor_id` | uuid | — | Filter by actor |
| `resource_type` | string | — | Filter by resource type |
| `resource_id` | uuid | — | Filter by specific resource |
| `from` | ISO 8601 | — | Start date (inclusive) |
| `to` | ISO 8601 | — | End date (inclusive) |
| `sort` | string | `created_at` | Sort field |
| `order` | string | `desc` | Sort order |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "actor_id": "uuid",
      "actor_email": "alice@acme.com",
      "actor_name": "Alice Compliance",
      "action": "user.role_assigned",
      "resource_type": "user",
      "resource_id": "uuid",
      "metadata": {
        "old_role": "security_engineer",
        "new_role": "ciso"
      },
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2026-02-20T10:00:00Z"
    }
  ],
  "meta": {
    "total": 156,
    "page": 1,
    "per_page": 50,
    "request_id": "uuid"
  }
}
```

**Notes:**
- Read-only endpoint — audit logs cannot be modified or deleted via API
- `actor_email` and `actor_name` are joined from users table for display convenience
- Results are always scoped to caller's `org_id`

---

## Middleware Stack

Request processing order:

```
Request
  → RequestID middleware (generates UUID, adds to response headers + context)
  → Logger middleware (structured JSON logging)
  → CORS middleware (configured origins)
  → Rate limiter middleware (Redis-backed, per IP for public, per user for auth'd)
  → Recovery middleware (panic recovery)
  → [Auth middleware] (JWT validation, user context injection)
  → [RBAC middleware] (role check against endpoint requirements)
  → [Org middleware] (inject org_id from token, scope all queries)
  → Handler
  → [Audit middleware] (log action on success)
  → Response
```

### Rate Limits

| Scope | Limit | Window |
|-------|-------|--------|
| Public endpoints (login, register) | 10 requests | per minute per IP |
| Authenticated endpoints | 100 requests | per minute per user |
| Refresh token | 5 requests | per minute per user |

---

## Go Project Structure

```
api/
├── cmd/
│   └── api/
│       └── main.go                  # Entry point, server setup
├── internal/
│   ├── config/
│   │   └── config.go               # Env-based configuration
│   ├── db/
│   │   ├── postgres.go             # PostgreSQL connection pool
│   │   └── redis.go                # Redis connection
│   ├── handlers/
│   │   ├── health.go               # /health, /ready
│   │   ├── auth.go                 # register, login, refresh, logout, change-password
│   │   ├── organizations.go        # org CRUD
│   │   ├── users.go                # user CRUD, deactivate, reactivate
│   │   └── audit.go                # audit log query
│   ├── middleware/
│   │   ├── auth.go                 # JWT validation
│   │   ├── rbac.go                 # Role-based access check
│   │   ├── org.go                  # Org context injection
│   │   ├── audit.go                # Audit logging
│   │   ├── ratelimit.go            # Rate limiting
│   │   ├── requestid.go            # Request ID generation
│   │   └── cors.go                 # CORS configuration
│   ├── models/
│   │   ├── organization.go         # Org struct + DB methods
│   │   ├── user.go                 # User struct + DB methods
│   │   ├── refresh_token.go        # Token struct + DB methods
│   │   └── audit_log.go            # AuditLog struct + DB methods
│   └── services/
│       ├── auth.go                 # Auth business logic
│       └── user.go                 # User business logic
├── go.mod
├── go.sum
└── Dockerfile
```

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gin-gonic/gin` | HTTP framework |
| `github.com/jackc/pgx/v5` | PostgreSQL driver |
| `github.com/redis/go-redis/v9` | Redis client |
| `github.com/golang-jwt/jwt/v5` | JWT creation/validation |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `github.com/google/uuid` | UUID generation |
| `github.com/rs/zerolog` | Structured logging |
