# Raisin Protect Wiki

> **Raisin Protect** is an AI-native Governance, Risk, and Compliance (GRC) platform. Multi-tenant SaaS supporting SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, and CCPA. Built with a Go backend, Next.js dashboard, PostgreSQL, Redis, and MinIO.

---

## Table of Contents

1. [Platform Overview](#1-platform-overview)
2. [Architecture](#2-architecture)
3. [Authentication & Authorization](#3-authentication--authorization)
4. [Multi-Tenancy](#4-multi-tenancy)
5. [Feature Modules](#5-feature-modules)
   - [Frameworks & Controls](#51-frameworks--controls)
   - [Evidence Management](#52-evidence-management)
   - [Continuous Monitoring](#53-continuous-monitoring)
   - [Policy Management](#54-policy-management)
   - [Risk Management](#55-risk-management)
   - [Audit Hub](#56-audit-hub)
   - [Access Reviews](#57-access-reviews)
6. [Database Schema](#6-database-schema)
7. [API Structure](#7-api-structure)
8. [Frontend Navigation](#8-frontend-navigation)
9. [Infrastructure & Deployment](#9-infrastructure--deployment)
10. [Development Workflow](#10-development-workflow)
11. [Security Model](#11-security-model)

---

## 1. Platform Overview

Raisin Protect consolidates compliance frameworks, controls, evidence, monitoring, policies, risks, audits, and access reviews into a single platform. It replaces spreadsheets, manual tracking, and fragmented tools with an integrated GRC workflow.

### Key Value Propositions

- **Multi-framework support** — SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, CCPA with cross-framework control mapping
- **Continuous monitoring** — Automated test execution with alert rules, SLA tracking, and delivery channels
- **Evidence lifecycle** — Upload, version, link to controls/requirements, track freshness, detect staleness
- **Policy governance** — Version-controlled policies with multi-signer approval workflows
- **Risk quantification** — 5x5 scoring matrix, heat maps, treatment plans, control effectiveness tracking
- **Audit readiness** — Audit engagements, PBC request management, finding remediation, evidence chain of custody
- **Access reviews** — Identity provider sync, review campaigns, approve/revoke decisions, anomaly detection
- **Multi-tenant isolation** — Row-Level Security at the database level, org-scoped JWT claims, middleware enforcement

### Supported Compliance Frameworks

| Framework | Version | Requirements |
|-----------|---------|-------------|
| SOC 2 | 2024 TSC | 64 |
| ISO 27001 | 2022 | 93 |
| PCI DSS | v4.0.1 | 280 |
| GDPR | 2016/679 | 99 |
| CCPA/CPRA | 2023 | 42 |

### GRC Roles

| Role | Code | Description |
|------|------|-------------|
| CISO | `ciso` | Full platform access, risk acceptance authority |
| Compliance Manager | `compliance_manager` | Full platform access, policy/audit management |
| Security Engineer | `security_engineer` | Controls, evidence, monitoring, risk assessment |
| IT Admin | `it_admin` | Evidence upload, alert handling, identity provider management |
| DevOps Engineer | `devops_engineer` | Evidence upload, monitoring, test execution |
| Auditor | `auditor` | Read-only + audit workspace, finding creation, evidence review |
| Vendor Manager | `vendor_manager` | Vendor management (future) |

---

## 2. Architecture

### System Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        Browser["Browser<br/>localhost:3010"]
    end

    subgraph "Frontend"
        NextJS["Next.js 14<br/>App Router<br/>TypeScript + Tailwind<br/>Port 3010"]
    end

    subgraph "Backend"
        API["Go API Server<br/>Gin Framework<br/>Port 8090"]
        Worker["Background Worker<br/>30s Poll Interval<br/>Port 8091"]
    end

    subgraph "Data Layer"
        Postgres["PostgreSQL 16<br/>Row-Level Security<br/>Port 5434 → 5432"]
        Redis["Redis 7<br/>Caching & Sessions<br/>Port 6380 → 6379"]
        MinIO["MinIO<br/>S3-Compatible Storage<br/>Port 9000"]
    end

    Browser --> NextJS
    NextJS -->|"/api/* proxy"| API
    API --> Postgres
    API --> Redis
    API --> MinIO
    Worker --> Postgres
    Worker --> Redis
```

### Service Responsibilities

| Service | Technology | Purpose |
|---------|-----------|---------|
| **Dashboard** | Next.js 14, TypeScript, Tailwind CSS, shadcn/ui | SPA with App Router, proxies `/api/*` to backend |
| **API** | Go 1.24, Gin framework | REST API, JWT auth, RBAC, all business logic |
| **Worker** | Go (same binary, different entry) | Background monitoring — polls every 30s for scheduled tests |
| **PostgreSQL** | PostgreSQL 16 Alpine | Primary data store with RLS for tenant isolation |
| **Redis** | Redis 7 Alpine | Caching and session storage |
| **MinIO** | MinIO (latest) | S3-compatible object storage for evidence files |

### Request Flow

```mermaid
sequenceDiagram
    participant B as Browser
    participant N as Next.js
    participant G as Go API
    participant P as PostgreSQL
    participant M as MinIO

    B->>N: GET /evidence
    N->>N: Client-side render (SPA)
    B->>N: GET /api/v1/evidence
    N->>G: Proxy /api/v1/evidence
    G->>G: RequestID → CORS → Auth → RateLimit
    G->>P: SELECT * FROM evidence_artifacts WHERE org_id = $1
    P-->>G: Rows
    G-->>N: JSON response
    N-->>B: Display data

    B->>N: POST /api/v1/evidence/:id/upload
    N->>G: Proxy upload request
    G->>M: Generate presigned upload URL
    M-->>G: Presigned URL
    G-->>B: Return upload URL
    B->>M: Direct upload to MinIO
```

---

## 3. Authentication & Authorization

### JWT Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API
    participant DB as PostgreSQL

    Note over C,DB: Registration
    C->>A: POST /api/v1/auth/register {email, password, org_name}
    A->>A: ValidatePassword (8+ chars, upper, lower, digit, special)
    A->>A: HashPassword (bcrypt cost 12)
    A->>DB: INSERT organization + user
    A->>A: GenerateTokenPair (HMAC-SHA256)
    A-->>C: {access_token, refresh_token, expires_in}

    Note over C,DB: Login
    C->>A: POST /api/v1/auth/login {email, password}
    A->>DB: SELECT user WHERE email = $1
    A->>A: CheckPassword (bcrypt compare)
    A->>A: GenerateTokenPair
    A-->>C: {access_token, refresh_token, expires_in}

    Note over C,DB: API Request
    C->>A: GET /api/v1/controls (Authorization: Bearer <access_token>)
    A->>A: ValidateAccessToken → extract Claims
    A->>A: Set context: user_id, org_id, email, role
    A->>DB: SELECT * FROM controls WHERE org_id = $1
    A-->>C: JSON response

    Note over C,DB: Token Refresh
    C->>A: POST /api/v1/auth/refresh {refresh_token}
    A->>A: ValidateRefreshToken
    A->>A: GenerateAccessToken (new access token only)
    A-->>C: {access_token, expires_in}
```

### JWT Token Details

| Property | Access Token | Refresh Token |
|----------|-------------|---------------|
| Algorithm | HMAC-SHA256 | HMAC-SHA256 |
| Default TTL | 15 minutes | 7 days (168h) |
| Claims | sub, org, email, role, type="access" | sub, org, email, role, type="refresh" |
| Issuer | `raisin-protect` | `raisin-protect` |
| Secret | Shared symmetric key (`RP_JWT_SECRET`) | Same key |

### Middleware Pipeline

```mermaid
flowchart LR
    A[Request] --> B[Recovery]
    B --> C[RequestID<br/>UUID v4 → X-Request-ID]
    C --> D[CORS<br/>Origin check + preflight]
    D --> E{Route Type?}

    E -->|Public /auth/*| F[RateLimitPublic<br/>10/min per IP]
    F --> G[Handler]

    E -->|Protected| H[AuthRequired<br/>JWT validation]
    H --> I[RateLimitAuth<br/>100/min per user]
    I --> J{Role Check?}
    J -->|No restriction| G
    J -->|RequireRoles| K[RBAC Check<br/>Role ∈ allowed list]
    K -->|Pass| G
    K -->|Fail| L[403 Forbidden]
    H -->|Invalid JWT| M[401 Unauthorized]
```

### RBAC Role-Permission Matrix

> Roles listed for each feature area. Unlisted roles have no access to that feature's write operations.

| Feature | CISO | Compliance Mgr | Security Eng | IT Admin | DevOps Eng | Auditor | Vendor Mgr |
|---------|:----:|:--------------:|:------------:|:--------:|:----------:|:-------:|:----------:|
| **User Management** | Create, Deactivate, Role Change | Create, Deactivate, Role Change | — | Create | — | — | — |
| **Framework Activation** | Activate, Deactivate | Activate, Deactivate | — | — | — | — | — |
| **Control Management** | Full CRUD, Bulk Ops | Full CRUD, Bulk Ops | Create, Update, Status | — | — | — | — |
| **Evidence Upload** | Upload, Status | Upload, Status | Upload, Link | Upload | Upload | — | — |
| **Evidence Evaluation** | Evaluate | Evaluate | — | — | — | Evaluate | — |
| **Test Management** | Full CRUD | Full CRUD | Create, Update, Status | — | Create, Update | — | — |
| **Alert Management** | Full (incl. Suppress) | Full (incl. Suppress) | Status, Assign, Resolve | Status, Resolve | Status, Resolve | — | — |
| **Policy Management** | Create, Publish, Archive | Create, Publish, Archive | Create, Update | — | — | View, Gap | — |
| **Risk Management** | Full, Accept | Full, Accept | Create, Assess | — | — | View, Gap | — |
| **Audit Management** | Create, Status | Create, Status | View, Evidence Submit | Evidence Submit | — | View, Requests, Findings | — |
| **Access Reviews** | Full Admin | Full Admin | View, Review | IdP Manage, Review | — | View, Review | — |
| **Audit Log** | View | View | — | — | — | View | — |

---

## 4. Multi-Tenancy

Raisin Protect enforces tenant isolation at three independent levels. A breach at any single level cannot expose data across organizations.

```mermaid
flowchart TB
    subgraph "Level 1: Middleware"
        JWT["JWT Token<br/>Contains org_id claim"]
        Auth["AuthRequired()<br/>Extracts org_id → context"]
    end

    subgraph "Level 2: Handler"
        Handler["Every Handler<br/>orgID := middleware.GetOrgID(c)<br/>All queries: WHERE org_id = $orgID"]
    end

    subgraph "Level 3: Database"
        RLS["Row-Level Security (RLS)<br/>PostgreSQL policies on every table<br/>SET app.current_org = $org_id"]
    end

    JWT --> Auth
    Auth --> Handler
    Handler --> RLS
```

### How It Works

1. **JWT Claim** — Every token contains an `org` claim with the user's organization UUID. Set at login/registration time.

2. **Middleware Extraction** — `AuthRequired()` middleware validates the JWT and stores `org_id` in Gin's request context. All subsequent handler code accesses it via `middleware.GetOrgID(c)`.

3. **Handler Enforcement** — Every handler includes `WHERE org_id = $orgID` in all SQL queries. This is a code convention enforced through code review and testing.

4. **Database RLS** — PostgreSQL Row-Level Security policies on every table ensure that even if a handler has a bug, the database itself prevents cross-tenant data access.

### MinIO Object Isolation

Evidence files are stored in MinIO with the key pattern:
```
{org_id}/{artifact_id}/{version}/{filename}
```

This ensures objects are namespaced by organization. The API generates presigned URLs scoped to the correct path.

---

## 5. Feature Modules

### 5.1 Frameworks & Controls

#### Concept

Frameworks (SOC 2, ISO 27001, etc.) are **system-level** resources — they exist globally and are not org-scoped. Organizations **activate** a framework version, which creates an `org_framework` record. Controls are **org-scoped** and are mapped to framework requirements.

#### Framework Hierarchy

```mermaid
graph TB
    F["Framework<br/>(e.g., SOC 2)"]
    FV["Framework Version<br/>(e.g., 2024 TSC)"]
    R1["Requirement Section<br/>(e.g., CC6: Access Controls)"]
    R2["Requirement<br/>(e.g., CC6.1)"]
    R3["Requirement<br/>(e.g., CC6.2)"]

    OF["Org Framework<br/>(activation record)"]
    C1["Control<br/>(org-scoped)"]
    C2["Control<br/>(org-scoped)"]
    CM["Control Mapping<br/>(primary/supporting/partial)"]
    RS["Requirement Scope<br/>(in_scope/out_of_scope)"]

    F --> FV
    FV --> R1
    R1 --> R2
    R1 --> R3
    OF --> FV
    C1 --> CM
    CM --> R2
    C2 --> CM
    RS --> R2
```

#### Control Status State Machine

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> active
    draft --> deprecated
    active --> under_review
    active --> deprecated
    under_review --> active
    under_review --> deprecated
    deprecated --> draft
```

#### Control Categories
`technical`, `administrative`, `physical`, `operational`

#### Mapping Strength Levels
- **Primary** — The control directly satisfies the requirement
- **Supporting** — The control partially addresses the requirement
- **Partial** — The control contributes to but does not fully cover the requirement

#### Key Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/frameworks` | List all frameworks (system-level) |
| POST | `/api/v1/org-frameworks` | Activate a framework for the org |
| GET | `/api/v1/org-frameworks/:id/coverage` | Coverage analysis |
| GET | `/api/v1/controls` | List org controls |
| POST | `/api/v1/controls/:id/mappings` | Map control to requirement(s) |
| GET | `/api/v1/mapping-matrix` | Cross-framework control mapping matrix |

---

### 5.2 Evidence Management

#### Concept

Evidence artifacts are files (screenshots, configs, reports, etc.) uploaded to MinIO and linked to controls and/or requirements. Each artifact has a freshness period, enabling automatic staleness detection.

#### Evidence Upload & Version Flow

```mermaid
sequenceDiagram
    participant U as User
    participant API as Go API
    participant DB as PostgreSQL
    participant S3 as MinIO

    U->>API: POST /api/v1/evidence<br/>{title, type, file_name, file_size}
    API->>DB: INSERT evidence_artifact (status=draft)
    API-->>U: {id, status: "draft"}

    U->>API: POST /api/v1/evidence/:id/upload
    API->>S3: GeneratePresignedUploadURL<br/>key: {org_id}/{id}/{version}/{filename}
    API-->>U: {upload_url, expires_in}

    U->>S3: PUT file directly to presigned URL
    S3-->>U: 200 OK

    U->>API: POST /api/v1/evidence/:id/confirm<br/>{checksum_sha256}
    API->>DB: UPDATE status = pending_review
    API-->>U: {status: "pending_review"}

    Note over U,S3: New Version
    U->>API: POST /api/v1/evidence/:id/versions<br/>{title, file_name, file_size}
    API->>DB: INSERT new artifact<br/>parent_artifact_id = original<br/>version = N+1, is_current = true
    API->>DB: UPDATE old artifact is_current = false
    API-->>U: {id: new_artifact_id, version: N+1}
```

#### Evidence Status State Machine

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> pending_review
    pending_review --> approved
    pending_review --> rejected
    rejected --> pending_review
    approved --> expired
    expired --> pending_review
```

#### Staleness Lifecycle

Evidence staleness is determined by `freshness_period_days` on each artifact:
- **Fresh** — `collection_date + freshness_period_days > now`
- **Stale** — `collection_date + freshness_period_days <= now`
- **No Schedule** — `freshness_period_days` is null

The API provides dedicated staleness endpoints:
- `GET /api/v1/evidence/staleness` — Lists stale artifacts
- `GET /api/v1/evidence/freshness-summary` — Aggregate freshness statistics

#### Evidence Types
`screenshot`, `api_response`, `configuration_export`, `log_sample`, `policy_document`, `access_list`, `vulnerability_report`, `certificate`, `training_record`, `penetration_test`, `audit_report`, `other`

#### Evidence Link Target Types
Evidence can be linked to:
- **Controls** — `target_type = "control"`
- **Requirements** — `target_type = "requirement"`
- **Policies** — `target_type = "policy"`

Each link has a **strength**: `primary`, `supporting`, `supplementary`

---

### 5.3 Continuous Monitoring

#### Concept

Monitoring tests are definitions that can be executed manually or on a schedule. Test runs group multiple test results. Failed results generate alerts via configurable alert rules.

#### Test Execution Pipeline

```mermaid
flowchart LR
    subgraph "Test Definition"
        T["Test<br/>(cron/interval schedule)"]
    end

    subgraph "Execution"
        TR["Test Run<br/>(manual or scheduled)"]
        W["Background Worker<br/>(30s poll)"]
    end

    subgraph "Results"
        P["Pass"]
        F["Fail"]
        E["Error"]
    end

    subgraph "Alerting"
        AR["Alert Rule<br/>(match conditions)"]
        A["Alert<br/>(with SLA deadline)"]
        D["Delivery<br/>(Slack, Email, Webhook)"]
    end

    T --> TR
    W -->|Scheduled trigger| TR
    TR --> P
    TR --> F
    TR --> E
    F --> AR
    E --> AR
    AR -->|Conditions match| A
    A --> D
```

#### Test Status State Machine

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> active
    active --> paused
    active --> deprecated
    paused --> active
    paused --> deprecated
    deprecated --> [*]
```

#### Alert Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> open
    open --> acknowledged
    open --> in_progress
    open --> suppressed
    open --> closed
    acknowledged --> in_progress
    acknowledged --> suppressed
    acknowledged --> closed
    in_progress --> resolved
    in_progress --> suppressed
    in_progress --> closed
    resolved --> closed
    resolved --> open : Reopen
    suppressed --> open : Unsuppress
    suppressed --> closed
    closed --> open : Reopen
```

#### Alert Rule Configuration

Alert rules define conditions for generating alerts from test results:

| Field | Description |
|-------|-------------|
| `match_test_types` | Filter by test type (configuration, access_control, etc.) |
| `match_severities` | Filter by test severity |
| `match_result_statuses` | Filter by result status (fail, error, etc.) |
| `match_control_ids` | Scope to specific controls |
| `match_tags` | Tag-based matching |
| `consecutive_failures` | Noise suppression — only alert after N consecutive failures |
| `cooldown_minutes` | Minimum time between alerts for the same condition |
| `sla_hours` | Automatically set SLA deadline on generated alerts |
| `delivery_channels` | `slack`, `email`, `webhook`, `in_app` |
| `priority` | Rule evaluation order (0–1000, lower = evaluated first) |

#### Test Types
`configuration`, `access_control`, `endpoint`, `vulnerability`, `data_protection`, `network`, `logging`, `custom`

---

### 5.4 Policy Management

#### Concept

Policies are version-controlled documents with a multi-signer approval workflow. Policies can be created from templates, linked to controls, and tracked for review schedule compliance.

#### Policy Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> in_review : Submit for Review<br/>(creates signoff records)
    in_review --> approved : All Signers Approve<br/>(automatic transition)
    in_review --> in_review : Signer Rejects<br/>(stays in review)
    approved --> in_review : Re-submit for Review
    approved --> published : Publish<br/>(CISO/Compliance Mgr)
    draft --> archived : Archive
    in_review --> archived : Archive<br/>(withdraws pending signoffs)
    approved --> archived : Archive
    published --> archived : Archive
```

#### Approval Workflow

```mermaid
sequenceDiagram
    participant O as Policy Owner
    participant API as API
    participant S1 as Signer 1
    participant S2 as Signer 2
    participant DB as Database

    O->>API: POST /policies/:id/submit-for-review<br/>{signer_ids: [s1, s2]}
    API->>DB: UPDATE policy status = in_review
    API->>DB: INSERT signoff (s1, pending)
    API->>DB: INSERT signoff (s2, pending)
    API-->>O: Policy in review

    S1->>API: POST /policies/:id/signoffs/:sid/approve
    API->>DB: UPDATE signoff status = approved
    API->>DB: COUNT pending signoffs = 1
    API-->>S1: Approved (1 pending remaining)

    S2->>API: POST /policies/:id/signoffs/:sid/approve
    API->>DB: UPDATE signoff status = approved
    API->>DB: COUNT pending signoffs = 0
    API->>DB: UPDATE policy status = approved ← automatic!
    API-->>S2: Approved (policy now approved)

    O->>API: POST /policies/:id/publish
    API->>DB: UPDATE policy status = published
    API-->>O: Policy published
```

#### Policy Categories (21)
`information_security`, `acceptable_use`, `access_control`, `data_classification`, `data_privacy`, `data_retention`, `incident_response`, `business_continuity`, `change_management`, `vulnerability_management`, `vendor_management`, `physical_security`, `encryption`, `network_security`, `secure_development`, `human_resources`, `compliance`, `risk_management`, `asset_management`, `logging_monitoring`, `custom`

#### Policy Version Tracking

Each policy edit creates a new `policy_version` record:
- `version_number` — Auto-incrementing per policy
- `content` — Full text (HTML, Markdown, or plain text)
- `change_type` — `initial`, `major`, `minor`, `patch`
- `is_current` — Only one version is current at a time
- `word_count`, `character_count` — Auto-calculated

The API supports version comparison: `GET /policies/:id/versions/compare?v1=1&v2=3`

#### Review Schedule
- `review_frequency_days` — How often the policy should be reviewed
- `next_review_at` — Computed deadline
- Review status is computed at query time: `overdue`, `due_soon` (within 30 days), `on_track`, `no_schedule`

---

### 5.5 Risk Management

#### Concept

The risk register tracks organizational risks with a 5x5 likelihood-impact scoring matrix. Risks go through assessment and treatment workflows, with controls linked to demonstrate mitigation.

#### Risk Scoring Formula

```mermaid
flowchart LR
    L["Likelihood<br/>1=Rare<br/>2=Unlikely<br/>3=Possible<br/>4=Likely<br/>5=Almost Certain"]
    I["Impact<br/>1=Negligible<br/>2=Minor<br/>3=Moderate<br/>4=Major<br/>5=Severe"]
    S["Score = L × I<br/>(Range: 1–25)"]
    SEV["Severity Band<br/>≥20 Critical<br/>≥12 High<br/>≥6 Medium<br/>&lt;6 Low"]

    L --> S
    I --> S
    S --> SEV
```

#### Risk Heat Map (5x5 Matrix)

```
                        Impact →
                 Negligible  Minor  Moderate  Major  Severe
                    (1)       (2)     (3)     (4)     (5)
Almost Certain (5)   5        10      15       20      25
Likely         (4)   4         8      12       16      20
Possible       (3)   3         6       9       12      15
Unlikely       (2)   2         4       6        8      10
Rare           (1)   1         2       3        4       5

Severity: ■ Critical (≥20)  ■ High (≥12)  ■ Medium (≥6)  ■ Low (<6)
```

The API provides the heat map data via `GET /api/v1/risks/heat-map`.

#### Risk Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> identified
    identified --> open
    identified --> assessing
    open --> assessing
    open --> treating
    open --> accepted : CISO/CM only<br/>Requires justification + expiry
    open --> closed
    assessing --> open
    assessing --> treating
    assessing --> accepted
    treating --> monitoring
    treating --> accepted
    treating --> closed
    monitoring --> treating
    monitoring --> accepted
    monitoring --> closed
    accepted --> open : Reopen
    accepted --> assessing
    accepted --> treating
    closed --> open : Reopen
    state archived {
        [*] --> terminal
    }
```

#### Treatment Workflow

```mermaid
stateDiagram-v2
    [*] --> planned
    planned --> in_progress
    planned --> cancelled
    in_progress --> implemented
    in_progress --> cancelled
    implemented --> verified : With effectiveness rating
    implemented --> ineffective
    verified --> ineffective
    cancelled --> [*]
    verified --> [*]
    ineffective --> [*]
```

Treatment types: `mitigate`, `accept`, `transfer`, `avoid`

#### Risk-Control Linkage

Controls can be linked to risks with effectiveness tracking:
- **Effectiveness**: `effective`, `partially_effective`, `ineffective`, `not_assessed`
- **Mitigation Percentage**: 0–100%
- **Periodic Review**: Last reviewed date and reviewer

#### Risk Assessments

Each risk has two tracked assessment types:
- **Inherent** — Risk level before controls are applied
- **Residual** — Risk level after controls are in place

Assessments are immutable snapshots with `is_current` flag. New assessments supersede previous ones.

---

### 5.6 Audit Hub

#### Concept

The Audit Hub manages the full lifecycle of audit engagements — from planning through fieldwork, evidence collection (PBC requests), finding tracking, and report issuance. Auditors have an isolated workspace with restricted visibility.

#### Audit Engagement Lifecycle

```mermaid
stateDiagram-v2
    [*] --> planning
    planning --> fieldwork : Sets actual_start date
    planning --> cancelled
    fieldwork --> review
    fieldwork --> cancelled
    review --> draft_report
    review --> fieldwork : Return to fieldwork
    review --> cancelled
    draft_report --> management_response
    draft_report --> cancelled
    management_response --> final_report
    management_response --> draft_report : Revise report
    final_report --> completed : Sets actual_end date
    completed --> [*]
    cancelled --> [*]
```

#### Evidence Request (PBC) Workflow

```mermaid
sequenceDiagram
    participant Aud as Auditor
    participant API as API
    participant CM as Compliance Manager
    participant SE as Security Engineer

    Aud->>API: POST /audits/:id/requests<br/>{title, description, priority, due_date}
    API-->>Aud: Request created (status: open)

    CM->>API: PUT /audits/:id/requests/:rid/assign<br/>{assigned_to: SE}
    API-->>CM: Assigned to Security Engineer

    SE->>API: POST /audits/:id/requests/:rid/evidence<br/>{artifact_id: existing_evidence}
    API-->>SE: Evidence submitted

    SE->>API: PUT /audits/:id/requests/:rid/submit
    API-->>SE: Request submitted

    Aud->>API: PUT /audits/:id/requests/:rid/evidence/:lid/review<br/>{status: "accepted"}
    API-->>Aud: Evidence accepted

    Aud->>API: PUT /audits/:id/requests/:rid/review<br/>{status: "accepted"}
    API-->>Aud: Request accepted
```

#### Finding Remediation Flow

```mermaid
stateDiagram-v2
    [*] --> identified : Auditor creates finding
    identified --> acknowledged
    identified --> risk_accepted : CISO only
    acknowledged --> remediation_planned
    acknowledged --> risk_accepted
    remediation_planned --> remediation_in_progress
    remediation_planned --> risk_accepted
    remediation_in_progress --> remediation_complete
    remediation_in_progress --> risk_accepted
    remediation_complete --> verified : Auditor verifies
    remediation_complete --> remediation_in_progress : Reopen
    verified --> closed
    risk_accepted --> closed
```

**Special Rules:**
- `remediation_planned` requires a `remediation_plan` text
- `verified` sets `verified_at` and `verified_by` automatically
- `risk_accepted` requires CISO role and `risk_acceptance_reason`
- Reopening from `remediation_complete` requires notes explaining why

#### Audit Types
`soc2_type1`, `soc2_type2`, `iso27001_certification`, `iso27001_surveillance`, `pci_dss_roc`, `pci_dss_saq`, `gdpr_dpia`, `internal`, `custom`

#### Finding Severity Levels
`critical`, `high`, `medium`, `low`, `informational`

#### Auditor Isolation

Auditors have restricted visibility — they can only see audits where their user ID appears in the `auditor_ids` array. This is enforced:
1. In `ListAudits` via SQL `WHERE $userID = ANY(auditor_ids)` for auditor role
2. In individual audit access via `checkAuditAccess()` helper
3. Internal comments (`is_internal = true`) are hidden from auditors

---

### 5.7 Access Reviews

#### Concept

Access reviews enable organizations to periodically review who has access to what systems, ensuring least-privilege and compliance with SOC 2 CC6, ISO 27001 A.9, and similar controls.

#### Identity Provider Sync Flow

```mermaid
sequenceDiagram
    participant Admin as IT Admin
    participant API as API
    participant IdP as Identity Provider<br/>(Okta, Azure AD, etc.)
    participant DB as Database

    Admin->>API: POST /access-reviews/identity-providers<br/>{name, type: "okta", config: {...}}
    API->>DB: INSERT identity_provider (status: pending_setup)

    Admin->>API: POST /identity-providers/:id/sync
    API->>IdP: Fetch users and roles
    IdP-->>API: User/role data
    API->>DB: UPSERT access_resources
    API->>DB: UPSERT access_entries
    API->>DB: UPDATE identity_provider<br/>last_sync_at, sync_stats
    API-->>Admin: Sync complete<br/>{users_synced, resources_synced}
```

#### Identity Provider Types
`okta`, `azure_ad`, `google_workspace`, `jumpcloud`, `onelogin`, `custom`

#### Campaign Lifecycle

```mermaid
stateDiagram-v2
    [*] --> draft : Create campaign<br/>Define scope & reviewers
    draft --> active : Launch Campaign<br/>Generates individual reviews
    draft --> cancelled
    active --> completed : Complete Campaign<br/>Expires remaining pending reviews
    active --> cancelled : Cancel Campaign<br/>Expires all pending reviews
    completed --> [*]
    cancelled --> [*]
```

#### Review Decision Flow

```mermaid
flowchart TB
    R["Access Review<br/>(pending)"]
    D{Reviewer<br/>Decision}
    A["Approved<br/>Access confirmed"]
    V["Revoked<br/>Requires justification"]
    F["Flagged<br/>Requires justification"]
    DEL["Delegated<br/>Reassigned to another reviewer"]
    ESC["Escalated<br/>Raised to admin"]
    REV["Revocation Executed<br/>IT Admin marks as done"]

    R --> D
    D -->|Approve| A
    D -->|Revoke| V
    D -->|Flag| F
    D -->|Delegate| DEL
    D -->|Escalate| ESC
    V --> REV
```

#### Campaign Scope Configuration (JSONB)
```json
{
  "resource_ids": ["uuid1", "uuid2"],
  "criticalities": ["critical", "high"],
  "privileged_only": true,
  "departments": ["Engineering"]
}
```

#### Reviewer Strategy
- `resource_owner` — The resource's owner is the reviewer
- `department_head` — Reviewer by department
- `explicit` — Manually specified reviewer
- `mixed` — Combination approach

#### Anomaly Detection
The API supports detecting access anomalies:
- `GET /access-reviews/entries/anomalies` — List entries with detected anomalies
- `POST /access-reviews/entries/detect-anomalies` — Trigger anomaly detection scan
- **Role drift**: `has_role_drift = true` when `role_name != expected_role`

---

## 6. Database Schema

### Entity-Relationship Diagram (Core Tables)

```mermaid
erDiagram
    organizations ||--o{ users : "has many"
    organizations ||--o{ controls : "has many"
    organizations ||--o{ evidence_artifacts : "has many"
    organizations ||--o{ policies : "has many"
    organizations ||--o{ risks : "has many"
    organizations ||--o{ audits : "has many"
    organizations ||--o{ org_frameworks : "has many"
    organizations ||--o{ tests : "has many"
    organizations ||--o{ identity_providers : "has many"
    organizations ||--o{ access_review_campaigns : "has many"

    frameworks ||--o{ framework_versions : "has versions"
    framework_versions ||--o{ requirements : "contains"
    requirements ||--o{ requirements : "parent-child"
    org_frameworks }o--|| framework_versions : "activates"

    controls ||--o{ control_mappings : "mapped to"
    control_mappings }o--|| requirements : "satisfies"

    evidence_artifacts ||--o{ evidence_links : "linked via"
    evidence_links }o--|| controls : "linked to"
    evidence_artifacts ||--o{ evidence_artifacts : "version chain"
    evidence_artifacts ||--o{ evidence_evaluations : "evaluated"

    tests }o--|| controls : "tests"
    test_runs ||--o{ test_results : "contains"
    test_results }o--|| tests : "for test"
    test_results ||--o| alerts : "generates"

    policies ||--o{ policy_versions : "has versions"
    policies ||--o{ policy_signoffs : "requires"
    policies ||--o{ policy_controls : "linked to"
    policy_controls }o--|| controls : "references"

    risks ||--o{ risk_assessments : "assessed by"
    risks ||--o{ risk_treatments : "treated by"
    risks ||--o{ risk_controls : "mitigated by"
    risk_controls }o--|| controls : "references"

    audits ||--o{ audit_requests : "contains"
    audits ||--o{ audit_findings : "contains"
    audits ||--o{ audit_comments : "has"
    audit_requests ||--o{ audit_evidence_links : "has evidence"
    audit_evidence_links }o--|| evidence_artifacts : "references"

    identity_providers ||--o{ access_resources : "syncs"
    access_resources ||--o{ access_entries : "has"
    access_review_campaigns ||--o{ access_reviews : "contains"
    access_reviews }o--|| access_entries : "reviews"

    organizations {
        uuid id PK
        varchar name
        varchar slug UK
        varchar domain
        org_status status
        jsonb settings
    }

    users {
        uuid id PK
        uuid org_id FK
        varchar email
        varchar password_hash
        varchar first_name
        varchar last_name
        grc_role role
        user_status status
        boolean mfa_enabled
    }

    frameworks {
        uuid id PK
        varchar identifier UK
        varchar name
        framework_category category
        boolean is_custom
    }

    controls {
        uuid id PK
        uuid org_id FK
        varchar identifier
        varchar title
        control_category category
        control_status status
        uuid owner_id FK
        boolean is_custom
    }

    evidence_artifacts {
        uuid id PK
        uuid org_id FK
        varchar title
        evidence_type evidence_type
        evidence_status status
        varchar object_key UK
        varchar checksum_sha256
        integer version
        boolean is_current
        integer freshness_period_days
    }

    policies {
        uuid id PK
        uuid org_id FK
        varchar identifier
        varchar title
        policy_category category
        policy_status status
        uuid owner_id FK
        uuid current_version_id FK
    }

    risks {
        uuid id PK
        uuid org_id FK
        varchar identifier
        varchar title
        risk_category category
        risk_status status
        uuid owner_id FK
        numeric inherent_score
        numeric residual_score
    }

    audits {
        uuid id PK
        uuid org_id FK
        varchar title
        audit_type audit_type
        audit_status status
        uuid org_framework_id FK
        date period_start
        date period_end
        uuid lead_auditor_id FK
        uuid_array auditor_ids
    }

    tests {
        uuid id PK
        uuid org_id FK
        varchar identifier
        varchar title
        test_type test_type
        test_status status
        uuid control_id FK
        varchar schedule_cron
    }

    alerts {
        uuid id PK
        uuid org_id FK
        integer alert_number
        varchar title
        alert_severity severity
        alert_status status
        uuid control_id FK
        uuid assigned_to FK
        boolean sla_breached
    }

    access_review_campaigns {
        uuid id PK
        uuid org_id FK
        varchar name
        campaign_status status
        jsonb scope
        timestamptz deadline
        integer total_reviews
        integer completed_reviews
    }

    access_reviews {
        uuid id PK
        uuid campaign_id FK
        uuid entry_id FK
        uuid reviewer_id FK
        review_decision decision
        text justification
    }
```

### Migration File Organization

63 migration files in `db/migrations/`, numbered `001` through `063`. Each sprint follows a pattern:

1. **Enum definitions** — PL/pgSQL functions and custom types
2. **Core tables** — Main entities with foreign keys
3. **Junction tables** — Many-to-many relationships
4. **Cross-references** — FK constraints added in later migration
5. **Seed templates** — System-provided read-only data (`is_custom = false`)
6. **Seed demo** — Development-only sample data

### Key Design Decisions

| Pattern | Description |
|---------|-------------|
| **Immutable tables** | `audit_log`, `test_results`, `policy_versions`, `risk_assessments`, `evidence_evaluations` — no `updated_at`, triggers prevent UPDATE/DELETE |
| **Polymorphic FKs** | `evidence_links.target_type` (control/requirement/policy), `audit_comments.target_type` (audit/request/finding) with check constraints |
| **Denormalized counts** | `test_runs` (pass/fail counts), `audits` (request/finding counts), `campaigns` (decision counts) for dashboard performance |
| **Self-referencing** | `requirements.parent_id` (hierarchical tree), `evidence_artifacts.parent_artifact_id` (version chain), `audit_comments.parent_comment_id` (threaded) |
| **Forward references** | `source_integration_id` fields on evidence, tests, and access entries point to Sprint 9's `integrations` table |

---

## 7. API Structure

### Base URL and Conventions

- **Base path**: `/api/v1`
- **Content type**: `application/json`
- **Auth**: Bearer token in `Authorization` header
- **Success response wrapper**: `successResponse(c, data)` with metadata
- **Error response wrapper**: `errorResponse(c, statusCode, errorCode, message)`
- **Audit logging**: All mutations log to `audit_log` via `middleware.LogAudit()`

### Route Organization

| Route Group | Sprint | Endpoints | Description |
|-------------|--------|-----------|-------------|
| `/auth/*` | 1 | 5 | Register, login, refresh, logout, change-password |
| `/organizations/*` | 1 | 2 | Get/update current org |
| `/users/*` | 1 | 7 | User CRUD, role change, activate/deactivate |
| `/audit-log` | 1 | 1 | Immutable audit trail |
| `/frameworks/*` | 2 | 4 | Framework catalog (read-only) |
| `/org-frameworks/*` | 2 | 8 | Activate frameworks, scoping, coverage |
| `/controls/*` | 2 | 14 | Control CRUD, mappings, evidence, test results |
| `/mapping-matrix` | 2 | 1 | Cross-framework control mapping |
| `/evidence/*` | 3 | 18 | Evidence CRUD, upload, versions, links, evaluations |
| `/tests/*` | 4 | 7 | Test definitions CRUD |
| `/test-runs/*` | 4 | 6 | Test execution and results |
| `/alerts/*` | 4 | 10 | Alert lifecycle management |
| `/alert-rules/*` | 4 | 5 | Alert rule CRUD |
| `/monitoring/*` | 4 | 4 | Dashboard views (heatmap, posture, summary) |
| `/policies/*` | 5 | 22 | Policy CRUD, versions, signoffs, controls |
| `/policy-templates/*` | 5 | 2 | Template library |
| `/policy-gap/*` | 5 | 2 | Gap analysis |
| `/risks/*` | 6 | 18 | Risk CRUD, assessments, treatments, controls |
| `/audits/*` | 7 | 30 | Audit engagements, requests, findings, evidence, comments |
| `/access-reviews/*` | 8 | 35 | IdPs, resources, entries, campaigns, reviews |
| **Total** | | **~200** | |

### Complete Endpoint Reference

<details>
<summary>Click to expand full endpoint list</summary>

#### Authentication (Public — 10/min rate limit)
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
```

#### Authentication (Protected)
```
POST   /api/v1/auth/logout
POST   /api/v1/auth/change-password
```

#### Organizations
```
GET    /api/v1/organizations/current
PUT    /api/v1/organizations/current              [Admin]
```

#### Users
```
GET    /api/v1/users
GET    /api/v1/users/:id
POST   /api/v1/users                              [CISO, CM, IT Admin]
PUT    /api/v1/users/:id                          [Self or Admin]
POST   /api/v1/users/:id/deactivate               [Admin]
POST   /api/v1/users/:id/reactivate               [Admin]
PUT    /api/v1/users/:id/role                      [Admin]
```

#### Frameworks & Controls
```
GET    /api/v1/frameworks
GET    /api/v1/frameworks/:id
GET    /api/v1/frameworks/:id/versions/:vid
GET    /api/v1/frameworks/:id/versions/:vid/requirements
GET    /api/v1/org-frameworks
POST   /api/v1/org-frameworks                     [CISO, CM]
PUT    /api/v1/org-frameworks/:id                  [CISO, CM]
DELETE /api/v1/org-frameworks/:id                  [CISO, CM]
GET    /api/v1/org-frameworks/:id/coverage
GET    /api/v1/org-frameworks/:id/scoping
PUT    /api/v1/org-frameworks/:id/requirements/:rid/scope    [CISO, CM]
DELETE /api/v1/org-frameworks/:id/requirements/:rid/scope    [CISO, CM]
GET    /api/v1/controls
POST   /api/v1/controls                           [CISO, CM, SE]
GET    /api/v1/controls/stats
POST   /api/v1/controls/bulk-status                [Admin]
GET    /api/v1/controls/:id
PUT    /api/v1/controls/:id                        [Owner check]
PUT    /api/v1/controls/:id/owner                  [Admin]
PUT    /api/v1/controls/:id/status                 [CISO, CM, SE]
DELETE /api/v1/controls/:id                        [Admin]
GET    /api/v1/controls/:id/mappings
POST   /api/v1/controls/:id/mappings               [CISO, CM, SE]
DELETE /api/v1/controls/:id/mappings/:mid           [CISO, CM, SE]
GET    /api/v1/controls/:id/evidence
GET    /api/v1/controls/:id/test-results
GET    /api/v1/mapping-matrix
GET    /api/v1/requirements/:id/evidence
```

#### Evidence Management
```
GET    /api/v1/evidence
POST   /api/v1/evidence                           [CISO, CM, SE, IT, DevOps]
GET    /api/v1/evidence/staleness
GET    /api/v1/evidence/freshness-summary
GET    /api/v1/evidence/search
GET    /api/v1/evidence/:id
PUT    /api/v1/evidence/:id                        [Uploader check]
DELETE /api/v1/evidence/:id                        [CISO, CM]
PUT    /api/v1/evidence/:id/status                 [CISO, CM]
POST   /api/v1/evidence/:id/confirm                [Uploader check]
POST   /api/v1/evidence/:id/upload                 [Uploader check]
GET    /api/v1/evidence/:id/download
POST   /api/v1/evidence/:id/versions               [CISO, CM, SE, IT, DevOps]
GET    /api/v1/evidence/:id/versions
GET    /api/v1/evidence/:id/links
POST   /api/v1/evidence/:id/links                  [CISO, CM, SE]
DELETE /api/v1/evidence/:id/links/:lid              [CISO, CM, SE]
GET    /api/v1/evidence/:id/evaluations
POST   /api/v1/evidence/:id/evaluations             [CISO, CM, Auditor]
```

#### Continuous Monitoring
```
GET    /api/v1/tests
POST   /api/v1/tests                              [CISO, CM, SE, DevOps]
GET    /api/v1/tests/:id
PUT    /api/v1/tests/:id                           [CISO, CM, SE, DevOps]
PUT    /api/v1/tests/:id/status                    [CISO, CM, SE]
DELETE /api/v1/tests/:id                           [CISO, CM]
GET    /api/v1/tests/:id/results
POST   /api/v1/test-runs                          [CISO, CM, SE, DevOps]
GET    /api/v1/test-runs
GET    /api/v1/test-runs/:id
POST   /api/v1/test-runs/:id/cancel                [CISO, CM, SE]
GET    /api/v1/test-runs/:id/results
GET    /api/v1/test-runs/:id/results/:rid
GET    /api/v1/alerts
GET    /api/v1/alerts/:id
PUT    /api/v1/alerts/:id/status                   [CISO, CM, SE, IT, DevOps]
PUT    /api/v1/alerts/:id/assign                   [CISO, CM, SE]
PUT    /api/v1/alerts/:id/resolve                  [CISO, CM, SE, IT, DevOps]
PUT    /api/v1/alerts/:id/suppress                 [CISO, CM]
PUT    /api/v1/alerts/:id/close                    [CISO, CM]
POST   /api/v1/alerts/:id/deliver                  [CISO, CM, SE]
POST   /api/v1/alerts/test-delivery                [CISO, CM]
GET    /api/v1/alert-rules                         [CISO, CM, SE]
POST   /api/v1/alert-rules                         [CISO, CM]
GET    /api/v1/alert-rules/:id                     [CISO, CM, SE]
PUT    /api/v1/alert-rules/:id                     [CISO, CM]
DELETE /api/v1/alert-rules/:id                     [CISO, CM]
GET    /api/v1/monitoring/heatmap
GET    /api/v1/monitoring/posture
GET    /api/v1/monitoring/summary
GET    /api/v1/monitoring/alert-queue
```

#### Policy Management
```
GET    /api/v1/policies
POST   /api/v1/policies                           [CISO, CM, SE]
GET    /api/v1/policies/search
GET    /api/v1/policies/stats
GET    /api/v1/policies/:id
PUT    /api/v1/policies/:id                        [Owner check]
POST   /api/v1/policies/:id/archive                [CISO, CM]
POST   /api/v1/policies/:id/submit-for-review      [Owner check]
POST   /api/v1/policies/:id/publish                [CISO, CM]
GET    /api/v1/policies/:id/versions
GET    /api/v1/policies/:id/versions/compare
GET    /api/v1/policies/:id/versions/:vnum
POST   /api/v1/policies/:id/versions               [Owner check]
GET    /api/v1/policies/:id/signoffs
POST   /api/v1/policies/:id/signoffs/remind         [Owner check]
POST   /api/v1/policies/:id/signoffs/:sid/approve   [Signer check]
POST   /api/v1/policies/:id/signoffs/:sid/reject    [Signer check]
POST   /api/v1/policies/:id/signoffs/:sid/withdraw  [Requester check]
GET    /api/v1/policies/:id/controls
POST   /api/v1/policies/:id/controls               [Owner check]
POST   /api/v1/policies/:id/controls/bulk           [Owner check]
DELETE /api/v1/policies/:id/controls/:cid           [Owner check]
GET    /api/v1/signoffs/pending
GET    /api/v1/policy-templates
POST   /api/v1/policy-templates/:id/clone           [CISO, CM, SE]
GET    /api/v1/policy-gap                          [CISO, CM, SE, Auditor]
GET    /api/v1/policy-gap/by-framework              [CISO, CM, SE, Auditor]
```

#### Risk Management
```
GET    /api/v1/risks
POST   /api/v1/risks                              [CISO, CM, SE]
GET    /api/v1/risks/heat-map
GET    /api/v1/risks/gaps                          [CISO, CM, SE, Auditor]
GET    /api/v1/risks/search
GET    /api/v1/risks/stats
GET    /api/v1/risks/:id
PUT    /api/v1/risks/:id                           [Owner check]
POST   /api/v1/risks/:id/archive                   [CISO, CM]
PUT    /api/v1/risks/:id/status                    [Owner + role check]
POST   /api/v1/risks/:id/recalculate               [CISO, CM, SE]
GET    /api/v1/risks/:id/assessments
POST   /api/v1/risks/:id/assessments               [Owner + role check]
GET    /api/v1/risks/:id/treatments
POST   /api/v1/risks/:id/treatments                [Owner + role check]
PUT    /api/v1/risks/:id/treatments/:tid            [Owner check]
POST   /api/v1/risks/:id/treatments/:tid/complete   [Owner check]
GET    /api/v1/risks/:id/controls
POST   /api/v1/risks/:id/controls                  [Owner + role check]
PUT    /api/v1/risks/:id/controls/:cid              [Owner + role check]
DELETE /api/v1/risks/:id/controls/:cid              [Owner + role check]
```

#### Audit Hub
```
GET    /api/v1/audit-request-templates             [CISO, CM, Auditor]
GET    /api/v1/audits                              [CISO, CM, SE, IT, Auditor]
POST   /api/v1/audits                              [CISO, CM]
GET    /api/v1/audits/dashboard                    [CISO, CM, SE]
GET    /api/v1/audits/:id                          [CISO, CM, SE, IT, Auditor]
PUT    /api/v1/audits/:id                          [CISO, CM]
PUT    /api/v1/audits/:id/status                   [CISO, CM]
POST   /api/v1/audits/:id/auditors                 [CISO, CM]
DELETE /api/v1/audits/:id/auditors/:uid             [CISO, CM]
GET    /api/v1/audits/:id/stats                    [CISO, CM, SE, IT, Auditor]
GET    /api/v1/audits/:id/readiness                [CISO, CM, SE, IT, Auditor]
GET    /api/v1/audits/:id/requests                 [CISO, CM, SE, IT, Auditor]
GET    /api/v1/audits/:id/requests/:rid            [CISO, CM, SE, IT, Auditor]
POST   /api/v1/audits/:id/requests                 [CISO, CM, Auditor]
PUT    /api/v1/audits/:id/requests/:rid            [CISO, CM, Auditor]
PUT    /api/v1/audits/:id/requests/:rid/assign     [CISO, CM]
PUT    /api/v1/audits/:id/requests/:rid/submit     [CISO, CM, SE, IT]
PUT    /api/v1/audits/:id/requests/:rid/review     [Auditor]
PUT    /api/v1/audits/:id/requests/:rid/close      [CISO, CM, Auditor]
POST   /api/v1/audits/:id/requests/bulk            [CISO, CM, Auditor]
POST   /api/v1/audits/:id/requests/from-template   [CISO, CM, Auditor]
GET    /api/v1/audits/:id/requests/:rid/evidence   [CISO, CM, SE, IT, Auditor]
POST   /api/v1/audits/:id/requests/:rid/evidence   [CISO, CM, SE, IT]
PUT    /api/v1/audits/:id/requests/:rid/evidence/:lid/review [Auditor]
DELETE /api/v1/audits/:id/requests/:rid/evidence/:lid [Auth check]
GET    /api/v1/audits/:id/findings                 [CISO, CM, SE, IT, Auditor]
GET    /api/v1/audits/:id/findings/:fid            [CISO, CM, SE, IT, Auditor]
POST   /api/v1/audits/:id/findings                 [Auditor]
PUT    /api/v1/audits/:id/findings/:fid            [Auditor]
PUT    /api/v1/audits/:id/findings/:fid/status     [Role check per transition]
PUT    /api/v1/audits/:id/findings/:fid/management-response [CISO, CM]
GET    /api/v1/audits/:id/comments                 [CISO, CM, SE, IT, Auditor]
POST   /api/v1/audits/:id/comments                 [CISO, CM, SE, IT, Auditor]
PUT    /api/v1/audits/:id/comments/:cid            [Author check]
DELETE /api/v1/audits/:id/comments/:cid            [Author + Admin check]
```

#### Access Reviews
```
GET    /api/v1/access-reviews/identity-providers         [CM, SE, CISO, Auditor]
POST   /api/v1/access-reviews/identity-providers         [CM, CISO, IT]
GET    /api/v1/access-reviews/identity-providers/:id     [CM, SE, CISO, Auditor]
PUT    /api/v1/access-reviews/identity-providers/:id     [CM, CISO, IT]
DELETE /api/v1/access-reviews/identity-providers/:id     [CM, CISO, IT]
POST   /api/v1/access-reviews/identity-providers/:id/sync [CM, CISO, IT]
GET    /api/v1/access-reviews/identity-providers/:id/sync-stats [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/resources                  [CM, SE, CISO, Auditor]
POST   /api/v1/access-reviews/resources                  [CM, CISO, IT, SE]
GET    /api/v1/access-reviews/resources/stats            [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/resources/:id              [CM, SE, CISO, Auditor]
PUT    /api/v1/access-reviews/resources/:id              [CM, CISO, IT, SE]
DELETE /api/v1/access-reviews/resources/:id              [CM, CISO, IT, SE]
GET    /api/v1/access-reviews/resources/:id/users        [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/entries                    [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/entries/anomalies          [CM, SE, CISO, Auditor]
POST   /api/v1/access-reviews/entries/detect-anomalies   [CM, CISO, IT, SE]
GET    /api/v1/access-reviews/entries/:id                [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/campaigns                  [CM, SE, CISO, Auditor]
POST   /api/v1/access-reviews/campaigns                  [CM, CISO]
GET    /api/v1/access-reviews/campaigns/:id              [CM, SE, CISO, Auditor]
PUT    /api/v1/access-reviews/campaigns/:id              [CM, CISO]
POST   /api/v1/access-reviews/campaigns/:id/launch       [CM, CISO]
POST   /api/v1/access-reviews/campaigns/:id/complete     [CM, CISO]
POST   /api/v1/access-reviews/campaigns/:id/cancel       [CM, CISO]
GET    /api/v1/access-reviews/campaigns/:id/stats        [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/campaigns/:id/certification-report [CM, CISO, Auditor]
GET    /api/v1/access-reviews/campaigns/:id/reviews      [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/campaigns/:id/reviews/:rid [CM, SE, CISO, Auditor]
POST   /api/v1/access-reviews/campaigns/:id/reviews/:rid/decide [Reviewer roles]
POST   /api/v1/access-reviews/campaigns/:id/reviews/bulk-decide [Reviewer roles]
POST   /api/v1/access-reviews/campaigns/:id/reviews/:rid/delegate [Reviewer roles]
POST   /api/v1/access-reviews/campaigns/:id/reviews/:rid/escalate [CM, CISO]
POST   /api/v1/access-reviews/campaigns/:id/reviews/:rid/revocation [IT, CISO]
PUT    /api/v1/access-reviews/reviews/:id                [Reviewer roles]
GET    /api/v1/access-reviews/dashboard                  [CM, SE, CISO, Auditor]
GET    /api/v1/access-reviews/my-reviews                 [Any authenticated]
```

#### Health Checks (No auth)
```
GET    /health
GET    /ready
```

</details>

---

## 8. Frontend Navigation

### Technology Stack
- **Next.js 14** with App Router
- **TypeScript** in strict mode
- **Tailwind CSS** for styling
- **shadcn/ui + Radix UI** for component primitives
- Path alias: `@/*` maps to project root

### Route Protection

The `AuthGuard` component (`components/auth-guard.tsx`) provides client-side route protection:
- **Public paths**: `/login`, `/register`
- **Protected paths**: Everything else — redirects unauthenticated users to `/login`
- **Authenticated on public path**: Redirects to `/` (dashboard)

Note: This is a UX-level guard. Security is enforced at the API/JWT level.

### Sidebar Navigation Structure

The sidebar (`components/sidebar.tsx`) renders navigation sections with role-based filtering. Items with a `roles` array are only visible to users with a matching role.

```
Overview
├── Dashboard                          [All roles]

Risk & Posture
├── Risk Dashboard                     [CISO, CM, SE]
└── Posture Overview                   [CISO, CM, SE]

Compliance
├── Frameworks                         [CISO, CM, SE, Auditor]
├── Controls                           [CISO, CM, SE, Auditor]
├── Mapping Matrix                     [CISO, CM, SE, Auditor]
├── Coverage                           [CISO, CM, SE, Auditor]
├── Evidence                           [CISO, CM, SE, Auditor]
└── Staleness Alerts                   [CISO, CM, SE, Auditor]

Risk Management
├── Risk Register                      [CISO, CM, SE, Auditor]
├── Heat Map                           [CISO, CM, SE, Auditor]
├── Risk Gaps                          [CISO, CM, SE, Auditor]
└── Treatments                         [CISO, CM, SE]

Policy Management
├── Policies                           [CISO, CM, SE, Auditor]
├── Templates                          [CISO, CM, SE]
├── Approvals                          [All roles]
└── Policy Gaps                        [CISO, CM, SE, Auditor]

Monitoring
├── Monitoring                         [CISO, CM, SE, DevOps]
├── Alert Queue                        [CISO, CM, SE, DevOps, IT]
├── Test Runs                          [CISO, CM, SE, DevOps]
└── Alert Rules                        [CISO, CM, SE]

Audit Hub
├── Audit Hub                          [CISO, CM, SE, IT, Auditor]
├── PBC Templates                      [CISO, CM, Auditor]
├── Audit Readiness                    [CISO, CM, SE]
└── Auditor Workspace                  [Auditor only]

Vendor Management
└── Vendors                            [CISO, CM, Vendor Mgr]

Integration
├── Integrations                       [CISO, CM, IT, DevOps]
└── API                                [DevOps only]

Administration
├── Users                              [CISO, CM, IT]
└── Organization                       [CISO, CM]
```

### Page Route Map

```
/                              Dashboard (home)
/login                         Login form
/register                      Registration form

Compliance
/frameworks                    Framework list
/frameworks/[id]               Framework detail + requirements tree
/controls                      Control library
/controls/[id]                 Control detail (mappings, evidence, tests)
/mapping-matrix                Cross-framework mapping view
/coverage                      Coverage analytics
/evidence                      Evidence artifact list
/evidence/[id]                 Evidence detail (versions, links, evaluations)
/staleness                     Stale evidence alerts

Monitoring
/monitoring                    Monitoring overview (heatmap, posture)
/alerts                        Alert queue
/alerts/[id]                   Alert detail
/test-runs                     Test run list
/test-runs/[id]                Test run results
/alert-rules                   Alert rule management

Policies
/policies                      Policy list
/policies/[id]                 Policy detail
/policies/[id]/edit            Rich text policy editor
/policies/[id]/versions        Version history
/policy-approvals              Pending signoff queue
/policy-templates              Template library
/policy-gap                    Policy-to-control gap analysis

Risk Management
/risk                          Risk dashboard
/risks                         Risk register
/risks/[id]                    Risk detail (assessments, treatments, controls)
/risks/[id]/edit               Risk editor
/risk-heatmap                  5x5 heat map visualization
/risk-gaps                     Risk gap analysis
/risk-treatments               Treatment plan list
/posture                       Compliance posture overview

Audit Hub
/audit                         Audit engagement list
/audit/[id]                    Audit detail
/audit/[id]/requests           PBC requests
/audit/[id]/requests/[rid]     Request detail
/audit/[id]/findings           Finding list
/audit/[id]/findings/[fid]     Finding detail
/audit-templates               PBC template library
/audit-readiness               Audit readiness dashboard
/audit/workspace               Auditor workspace (auditor only)

Administration
/users                         User management
/settings                      Organization settings
```

---

## 9. Infrastructure & Deployment

### Docker Compose Topology

```mermaid
graph TB
    subgraph "Docker Network: rp-network"
        subgraph "Frontend"
            D["dashboard<br/>Next.js<br/>:3010"]
        end

        subgraph "Backend"
            A["api<br/>Go/Gin<br/>:8090"]
            W["worker<br/>Go<br/>:8091<br/>(not exposed)"]
        end

        subgraph "Data"
            PG["postgres<br/>PostgreSQL 16<br/>:5434→5432"]
            R["redis<br/>Redis 7<br/>:6380→6379"]
            M["minio<br/>MinIO<br/>:9000, :9001"]
        end
    end

    D -->|"depends_on"| A
    A -->|"depends_on"| PG
    A -->|"depends_on"| R
    A -->|"depends_on"| M
    W -->|"depends_on"| PG
    W -->|"depends_on"| R
```

### Port Mappings

| Service | Host Port | Container Port | Notes |
|---------|-----------|----------------|-------|
| PostgreSQL | 5434 | 5432 | Non-standard to avoid conflicts |
| Redis | 6380 | 6379 | Non-standard to avoid conflicts |
| MinIO API | 9000 | 9000 | S3-compatible API |
| MinIO Console | 9001 | 9001 | Web UI for bucket management |
| Go API | 8090 | 8090 | REST API server |
| Go Worker | — | 8091 | Not exposed externally |
| Next.js | 3010 | 3010 | Dashboard UI |

### Memory Limits

| Service | Memory |
|---------|--------|
| PostgreSQL | 512 MB |
| Dashboard | 512 MB |
| API | 256 MB |
| Worker | 256 MB |
| MinIO | 256 MB |
| Redis | 128 MB |

### Persistent Volumes

| Volume | Service | Purpose |
|--------|---------|---------|
| `rp_postgres_data` | PostgreSQL | Database files |
| `rp_redis_data` | Redis | AOF persistence |
| `rp_minio_data` | MinIO | Object storage |

### Environment Configuration

All environment variables use the `RP_` prefix. See `api/.env.example` for the full list.

| Variable | Default | Description |
|----------|---------|-------------|
| `RP_PORT` | `8090` | HTTP listen port |
| `RP_ENV` | `development` | Environment (`development`, `staging`, `production`) |
| `RP_DB_URL` | `postgres://rp:...@localhost:5434/...` | PostgreSQL connection string |
| `RP_REDIS_URL` | `redis://localhost:6380` | Redis connection URL |
| `RP_JWT_SECRET` | Auto-generated in dev | HMAC signing key (32+ chars required in prod) |
| `RP_JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `RP_JWT_REFRESH_TTL` | `168h` | Refresh token lifetime (7 days) |
| `RP_JWT_ISSUER` | `raisin-protect` | JWT issuer claim |
| `RP_BCRYPT_COST` | `12` | Password hashing work factor (10–15) |
| `RP_CORS_ORIGINS` | `http://localhost:3010` | Comma-separated allowed origins |
| `RP_MINIO_ENDPOINT` | `localhost:9000` | MinIO/S3 endpoint |
| `RP_MINIO_ACCESS_KEY` | `rp-admin` | MinIO access key |
| `RP_MINIO_SECRET_KEY` | `changeme-minio` | MinIO secret key |
| `RP_MINIO_BUCKET` | `rp-evidence` | Default evidence bucket |
| `RP_MINIO_USE_SSL` | `false` | Enable TLS for MinIO |
| `RP_LOG_LEVEL` | `info` | Log level (`debug`, `info`) |

### Production Guards
- `RP_JWT_SECRET` **must** be set in production/staging (not auto-generated)
- `RP_JWT_SECRET` **must** be 32+ characters in production/staging
- A random secret is generated in development (tokens won't survive restarts)

---

## 10. Development Workflow

### Prerequisites
- Go 1.24+
- Node.js 18+ / npm
- Docker & Docker Compose
- PostgreSQL 16 (via Docker or local)

### Quick Start

```bash
# 1. Start infrastructure
docker-compose up -d

# 2. Run backend locally (optional — or use the Docker container)
cd api && go run cmd/api/main.go

# 3. Run frontend locally (optional — or use the Docker container)
cd dashboard && npm install && npm run dev

# Access:
# Dashboard: http://localhost:3010
# API:       http://localhost:8090
# MinIO UI:  http://localhost:9001
```

### Database Migrations

Migrations run automatically on API startup. They are applied in order from `db/migrations/001_*.sql` through `db/migrations/063_*.sql`.

To add a new migration:
1. Create `db/migrations/064_description.sql`
2. Restart the API — migrations run on boot

### Seed Data

The seed file (`db/seeds/seed.sql`) creates:
- **Acme Corporation** — Demo organization
- **9 demo users** — One per role (password: `demo123`)
- **5 frameworks** — SOC 2, ISO 27001, PCI DSS v4.0.1, GDPR, CCPA/CPRA
- **Full requirement trees** — Hierarchical requirements for each framework
- **Sample audit log entries**

### Running Tests

```bash
# All backend tests (195+)
cd api && go test ./...

# Single package with verbose output
cd api && go test -v ./internal/handlers/

# With coverage report
cd api && go test -v -cover ./...

# E2E tests (Playwright)
cd tests/e2e && npx playwright test
```

### Testing Approach
- Handler tests use `sqlmock` for database mocking
- Test files sit alongside their handler files (e.g., `handlers/policies.go` → `handlers/policies_test.go`)
- Tests validate HTTP status codes, response shapes, and error conditions
- No integration tests against a real database (by design — fast, isolated tests)

---

## 11. Security Model

### Defense in Depth

```mermaid
flowchart TB
    subgraph "Network Layer"
        CORS["CORS<br/>Origin whitelist"]
        RL["Rate Limiting<br/>10/min public, 100/min auth"]
    end

    subgraph "Authentication Layer"
        JWT["JWT Validation<br/>HMAC-SHA256"]
        PW["Password Security<br/>bcrypt cost 12"]
        PWVAL["Password Complexity<br/>8+ chars, upper, lower, digit, special"]
    end

    subgraph "Authorization Layer"
        RBAC["RBAC Middleware<br/>7 roles, per-endpoint"]
        OWN["Ownership Checks<br/>In-handler validation"]
        AUD["Auditor Isolation<br/>auditor_ids array filter"]
    end

    subgraph "Data Layer"
        TENANT["Multi-Tenant Isolation<br/>org_id on every query"]
        RLS["Row-Level Security<br/>PostgreSQL policies"]
        AUDIT["Audit Logging<br/>Immutable audit_log table"]
    end

    subgraph "Input/Output"
        SANBE["Backend Sanitization<br/>bluemonday (HTML)"]
        SANFE["Frontend Sanitization<br/>DOMPurify (HTML)"]
        PARAM["Parameterized Queries<br/>No string concatenation"]
    end

    CORS --> JWT
    RL --> JWT
    JWT --> RBAC
    PW --> JWT
    PWVAL --> PW
    RBAC --> OWN
    RBAC --> AUD
    OWN --> TENANT
    AUD --> TENANT
    TENANT --> RLS
    TENANT --> AUDIT
    SANBE --> PARAM
    SANFE --> PARAM
```

### Password Security

| Property | Value |
|----------|-------|
| Algorithm | bcrypt |
| Default cost | 12 (configurable 10–15 via `RP_BCRYPT_COST`) |
| Minimum length | 8 characters |
| Complexity | Must include: uppercase, lowercase, digit, special character |
| Storage | Only hash stored, never plaintext |

### JWT Token Security

| Property | Value |
|----------|-------|
| Algorithm | HMAC-SHA256 |
| Library | `github.com/golang-jwt/jwt/v5` |
| Access token TTL | 15 minutes (configurable) |
| Refresh token TTL | 7 days (configurable) |
| Secret requirements | 32+ characters in production |
| Token types | Separate `access` and `refresh` types validated independently |

### CORS Configuration

- Allowed origins configurable via `RP_CORS_ORIGINS` (comma-separated)
- Default: `http://localhost:3010`
- Allowed methods: `GET, POST, PUT, PATCH, DELETE, OPTIONS`
- Credentials: Allowed
- Preflight cache: 86400 seconds (24 hours)
- OPTIONS requests return 204 immediately

### Rate Limiting

| Scope | Limit | Key |
|-------|-------|-----|
| Public endpoints (`/auth/*`) | 10 requests/minute | Client IP |
| Authenticated endpoints | 100 requests/minute | User ID (fallback: IP) |

Implementation: In-memory sliding window (no Redis dependency). Protected by `sync.Mutex`.

### Audit Logging

Every mutation is logged to the immutable `audit_log` table:
- **Trigger prevention** — Database triggers prevent UPDATE and DELETE on `audit_log`
- **Automatic fields** — `org_id`, `actor_id`, `ip_address`, `user_agent` extracted from context
- **Structured metadata** — JSONB field with operation-specific details
- **Pre-auth events** — Registration and login logged via `LogAuditWithOrg()` (before JWT context exists)

### HTML Sanitization

- **Backend**: `bluemonday` library sanitizes HTML content before storage
- **Frontend**: `DOMPurify` sanitizes HTML before rendering
- Prevents XSS via stored content (policies, descriptions, comments)

### SQL Injection Prevention

All database queries use parameterized statements (`$1`, `$2`, etc.). No string concatenation for SQL construction.

### HTTP Server Hardening

| Setting | Value |
|---------|-------|
| Read timeout | 30 seconds |
| Write timeout | 30 seconds |
| Idle timeout | 120 seconds |
| Request ID | UUID v4 on every request (`X-Request-ID` header) |
| Panic recovery | Gin's built-in `Recovery()` middleware |

---

## Sprint Roadmap

| Sprint | Status | Focus |
|--------|--------|-------|
| 1 | Complete | Scaffolding, Auth, RBAC, Docker |
| 2 | Complete | Frameworks, Controls, Mapping Matrix |
| 3 | Complete | Evidence Management, MinIO, Freshness Tracking |
| 4 | Complete | Continuous Monitoring, Tests, Alerts, Background Worker |
| 5 | Complete | Policy Management, Versioning, Approval Workflow |
| 6 | Complete | Risk Register, Scoring Matrix, Treatments |
| 7 | Complete | Audit Hub, PBC Requests, Findings, Comments |
| 8 | Complete | Access Reviews, Identity Providers, Campaigns |
| 9 | Next | Integration Engine (AWS Config, GitHub, Okta, Slack) |
| 10 | Planned | Reporting, Executive Dashboard, OpenAPI, E2E Tests |

---

## Demo Credentials

For local development with seed data:

| Role | Email | Password |
|------|-------|----------|
| CISO | `ciso@acme.example.com` | `demo123` |
| Compliance Manager | `compliance@acme.example.com` | `demo123` |
| Security Engineer | `security@acme.example.com` | `demo123` |
| IT Admin | `it@acme.example.com` | `demo123` |
| DevOps Engineer | `devops@acme.example.com` | `demo123` |
| Auditor | `auditor@acme.example.com` | `demo123` |
| Vendor Manager | `vendor@acme.example.com` | `demo123` |

Organization: **Acme Corporation** (`acme-corp`)
