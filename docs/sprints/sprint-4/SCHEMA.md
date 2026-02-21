# Sprint 4 — Database Schema: Continuous Monitoring Engine

## Overview

Sprint 4 introduces the continuous monitoring engine — the heartbeat of the GRC platform. This layer defines **what to test** (test definitions), **when it ran** (test runs), **what happened** (test results), **what went wrong** (alerts), and **how to respond** (alert rules). Together, these enable the spec §3.3 vision: 1,200+ automated tests executed hourly, with severity-based alert workflows and real-time compliance posture scoring.

**Key design decisions:**
- **Tests are linked to controls**: Each test validates one or more controls. A failing test directly degrades the control's health status, which cascades to framework compliance posture.
- **Two-tier execution model**: `test_runs` are batch execution envelopes (a "sweep"), and `test_results` are per-test outcomes within that sweep. This separates "when did we run?" from "what happened for each test?"
- **Alert generation is rule-driven**: `alert_rules` define thresholds and conditions. When a test result matches a rule, an alert is auto-generated. Rules are org-configurable, not hardcoded.
- **Alerts have full lifecycle**: Open → Acknowledged → In Progress → Resolved → Closed, with optional suppression. Each transition is audited.
- **SLA tracking built-in**: Alerts carry `sla_deadline` computed from severity + org-configurable SLA durations. Overdue alerts are surfaced in the monitoring dashboard.
- **Worker-friendly**: `test_runs` use `started_at`/`completed_at` + status to support background job execution. The worker picks up scheduled tests, creates a run, executes tests, writes results, and triggers alert evaluation.

---

## Entity Relationship Diagram

```
                    ┌─────────────────────────────────────────────┐
                    │   MONITORING DOMAIN (org-scoped)             │
                    │                                             │
 organizations ─┬──▶ tests                                       │
                │   │   (test definitions with schedules)         │
                │   │   ∞──1 controls (control being tested)      │
                │   │                                             │
                │   │   1──∞ test_results ──∞──1 test_runs        │
                │   │          (per-test outcomes in a sweep)      │
                │   │                                             │
                │   ├──▶ test_runs                                │
                │   │   (batch execution sweeps)                  │
                │   │                                             │
                │   ├──▶ alerts                                   │
                │   │   (generated from test failures)            │
                │   │   ∞──1 tests                                │
                │   │   ∞──1 controls                             │
                │   │   ∞──1 test_results (triggering result)     │
                │   │   ∞──1 users (assigned_to)                  │
                │   │                                             │
                │   ├──▶ alert_rules                              │
                │   │   (org-configurable alert generation rules) │
                │   │                                             │
                └──▶ audit_log (extended)                         │
                    └─────────────────────────────────────────────┘
```

**Relationships:**
```
tests               ∞──1  organizations        (org_id)
tests               ∞──1  controls             (control_id)
tests               ∞──1  users                (created_by)
test_runs           ∞──1  organizations        (org_id)
test_runs           ∞──1  users                (triggered_by, optional)
test_results        ∞──1  test_runs            (test_run_id)
test_results        ∞──1  tests                (test_id)
test_results        ∞──1  controls             (control_id, denormalized)
alerts              ∞──1  organizations        (org_id)
alerts              ∞──1  tests                (test_id, optional)
alerts              ∞──1  controls             (control_id)
alerts              ∞──1  test_results         (test_result_id, optional)
alerts              ∞──1  users                (assigned_to, optional)
alerts              ∞──1  users                (resolved_by, optional)
alert_rules         ∞──1  organizations        (org_id)
alert_rules         ∞──1  users                (created_by)
```

---

## New Enums

```sql
-- Categories of automated tests (from spec §3.3.1)
CREATE TYPE test_type AS ENUM (
    'configuration',       -- Cloud resource settings match security baselines
    'access_control',      -- Permissions align with least-privilege policies
    'endpoint',            -- Devices meet encryption, firewall, patch requirements
    'vulnerability',       -- Known vulnerabilities tracked against SLA deadlines
    'data_protection',     -- Sensitive data handling and encryption verification
    'network',             -- Firewall rules, segmentation, monitoring validation
    'logging',             -- Audit log completeness and retention verification
    'custom'               -- User-defined test logic
);

-- Test definition lifecycle
CREATE TYPE test_status AS ENUM (
    'draft',               -- Being configured, not yet active
    'active',              -- Running on schedule
    'paused',              -- Temporarily disabled (manual pause)
    'deprecated'           -- No longer relevant, kept for history
);

-- Severity of a test (determines alert severity on failure)
CREATE TYPE test_severity AS ENUM (
    'critical',            -- System-down, data-breach-level
    'high',                -- Major compliance gap
    'medium',              -- Moderate risk, needs attention
    'low',                 -- Minor finding, best-practice
    'informational'        -- No risk, awareness only
);

-- Outcome of an individual test execution
CREATE TYPE test_result_status AS ENUM (
    'pass',                -- Test passed, control is healthy
    'fail',                -- Test failed, control is non-compliant
    'error',               -- Test could not execute (infra issue, timeout, etc.)
    'skip',                -- Test was skipped (e.g., dependency not met)
    'warning'              -- Passed with concerns (approaching threshold)
);

-- Lifecycle of a test run (batch sweep)
CREATE TYPE test_run_status AS ENUM (
    'pending',             -- Queued, waiting to start
    'running',             -- Currently executing tests
    'completed',           -- All tests in the sweep finished
    'failed',              -- Sweep aborted (infrastructure failure)
    'cancelled'            -- Manually cancelled before completion
);

-- What triggered a test run
CREATE TYPE test_run_trigger AS ENUM (
    'scheduled',           -- Automatic, from cron/interval schedule
    'manual',              -- Human triggered via UI or API
    'on_change',           -- Triggered by a detected change in the system
    'webhook'              -- Triggered by an external webhook event
);

-- Alert severity (from spec §3.3.2)
CREATE TYPE alert_severity AS ENUM (
    'critical',            -- Immediate action required
    'high',                -- Action required within SLA
    'medium',              -- Needs attention, plan remediation
    'low'                  -- Best-practice improvement
);

-- Alert lifecycle (from spec §3.3.2 workflow)
CREATE TYPE alert_status AS ENUM (
    'open',                -- Newly created, unacknowledged
    'acknowledged',        -- Someone has seen it, not yet working
    'in_progress',         -- Actively being remediated
    'resolved',            -- Fix applied, awaiting verification
    'suppressed',          -- Intentionally snoozed with justification
    'closed'               -- Verified fixed OR accepted as risk
);

-- How alerts are delivered
CREATE TYPE alert_delivery_channel AS ENUM (
    'slack',               -- Slack webhook
    'email',               -- Email via SMTP/SendGrid
    'webhook',             -- Custom outbound webhook
    'in_app'               -- In-app notification only (always generated)
);
```

### Extend Existing Enums

```sql
-- Add Sprint 4 actions to audit_action enum
ALTER TYPE audit_action ADD VALUE 'test.created';
ALTER TYPE audit_action ADD VALUE 'test.updated';
ALTER TYPE audit_action ADD VALUE 'test.status_changed';
ALTER TYPE audit_action ADD VALUE 'test.deleted';
ALTER TYPE audit_action ADD VALUE 'test_run.started';
ALTER TYPE audit_action ADD VALUE 'test_run.completed';
ALTER TYPE audit_action ADD VALUE 'test_run.failed';
ALTER TYPE audit_action ADD VALUE 'test_run.cancelled';
ALTER TYPE audit_action ADD VALUE 'alert.created';
ALTER TYPE audit_action ADD VALUE 'alert.acknowledged';
ALTER TYPE audit_action ADD VALUE 'alert.assigned';
ALTER TYPE audit_action ADD VALUE 'alert.status_changed';
ALTER TYPE audit_action ADD VALUE 'alert.resolved';
ALTER TYPE audit_action ADD VALUE 'alert.suppressed';
ALTER TYPE audit_action ADD VALUE 'alert.closed';
ALTER TYPE audit_action ADD VALUE 'alert.reopened';
ALTER TYPE audit_action ADD VALUE 'alert_rule.created';
ALTER TYPE audit_action ADD VALUE 'alert_rule.updated';
ALTER TYPE audit_action ADD VALUE 'alert_rule.deleted';
```

---

## Tables

### tests

Test definitions — the "what to check" catalog. Each test is linked to exactly one control and defines how to validate that control is working. Tests have schedules, severity, and optionally contain the test logic itself (for custom tests).

```sql
CREATE TABLE tests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    identifier              VARCHAR(50) NOT NULL,            -- human-readable: 'TST-CFG-001'
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,

    -- Classification
    test_type               test_type NOT NULL,
    severity                test_severity NOT NULL DEFAULT 'medium',
    status                  test_status NOT NULL DEFAULT 'draft',

    -- What it validates
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Schedule (cron expression or interval)
    schedule_cron           VARCHAR(100),                    -- cron expression: '0 * * * *' (every hour)
    schedule_interval_min   INT,                             -- alternative: run every N minutes
    next_run_at             TIMESTAMPTZ,                     -- precomputed next execution time (for worker polling)
    last_run_at             TIMESTAMPTZ,                     -- when it last ran (for staleness detection)

    -- Test logic (for custom tests — spec §3.2.2 Compliance as Code)
    test_script             TEXT,                            -- script content (shell, Python, JS)
    test_script_language    VARCHAR(30),                     -- 'shell', 'python', 'javascript'
    test_config             JSONB NOT NULL DEFAULT '{}',     -- test-specific config (thresholds, params, endpoints)

    -- Execution parameters
    timeout_seconds         INT NOT NULL DEFAULT 300,        -- max execution time (5 min default, per spec §13.1)
    retry_count             INT NOT NULL DEFAULT 0,          -- retries on error (0 = no retry)
    retry_delay_seconds     INT NOT NULL DEFAULT 60,         -- delay between retries

    -- Metadata
    tags                    TEXT[] DEFAULT '{}',              -- free-form tags: ['pci', 'network', 'automated']
    source_integration_id   UUID,                            -- FK to integrations (Sprint 9, nullable for now)

    -- Ownership
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_test_identifier UNIQUE (org_id, identifier),
    CONSTRAINT chk_test_schedule CHECK (
        (schedule_cron IS NOT NULL AND schedule_interval_min IS NULL) OR
        (schedule_cron IS NULL AND schedule_interval_min IS NOT NULL) OR
        (schedule_cron IS NULL AND schedule_interval_min IS NULL)  -- manual-only tests
    ),
    CONSTRAINT chk_test_interval_positive CHECK (schedule_interval_min IS NULL OR schedule_interval_min > 0),
    CONSTRAINT chk_test_timeout_positive CHECK (timeout_seconds > 0 AND timeout_seconds <= 3600),
    CONSTRAINT chk_test_retry CHECK (retry_count >= 0 AND retry_count <= 5)
);

-- Indexes
CREATE INDEX idx_tests_org ON tests (org_id);
CREATE INDEX idx_tests_org_status ON tests (org_id, status);
CREATE INDEX idx_tests_control ON tests (control_id);
CREATE INDEX idx_tests_type ON tests (org_id, test_type);
CREATE INDEX idx_tests_severity ON tests (org_id, severity);
CREATE INDEX idx_tests_next_run ON tests (next_run_at)
    WHERE status = 'active' AND next_run_at IS NOT NULL;
CREATE INDEX idx_tests_identifier ON tests (org_id, identifier);
CREATE INDEX idx_tests_tags ON tests USING gin (tags);
CREATE INDEX idx_tests_created_by ON tests (created_by)
    WHERE created_by IS NOT NULL;

-- Trigger
CREATE TRIGGER trg_tests_updated_at
    BEFORE UPDATE ON tests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `identifier` is a human-readable code like `TST-CFG-001` (org-scoped unique).
- **Schedule flexibility**: Either `schedule_cron` (full cron expression) or `schedule_interval_min` (simpler "every N minutes"). Both NULL means manual-only. The CHECK constraint ensures they're mutually exclusive.
- `next_run_at` is precomputed by the worker after each run. The worker polls `WHERE status = 'active' AND next_run_at <= NOW()` to find tests due for execution. This avoids parsing cron expressions at query time.
- `test_script` + `test_script_language` support Compliance as Code (spec §3.2.2). For built-in test types, the worker uses the `test_type` + `test_config` to determine what to check. For `custom` type, it executes the script.
- `test_config` is a JSONB catch-all for test-specific parameters: thresholds, API endpoints, expected values, regex patterns, etc.
- `timeout_seconds` caps at 3600 (1 hour). Spec §13.1 says full sweep should complete in <5 minutes, but individual tests could run longer for complex checks.
- `source_integration_id` is forward-looking for Sprint 9 (Integration Engine), indicating which integration this test pulls data from.

---

### test_runs

Batch execution records — a "sweep" that contains multiple test results. Each run represents one cycle of the monitoring engine.

```sql
CREATE TABLE test_runs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Run metadata
    run_number              SERIAL,                          -- auto-incrementing run counter per org
    status                  test_run_status NOT NULL DEFAULT 'pending',
    trigger_type            test_run_trigger NOT NULL DEFAULT 'scheduled',

    -- Timing
    started_at              TIMESTAMPTZ,                     -- when execution began
    completed_at            TIMESTAMPTZ,                     -- when all tests finished
    duration_ms             INT,                             -- total wall-clock time in milliseconds

    -- Results summary (denormalized for fast dashboard queries)
    total_tests             INT NOT NULL DEFAULT 0,
    passed                  INT NOT NULL DEFAULT 0,
    failed                  INT NOT NULL DEFAULT 0,
    errors                  INT NOT NULL DEFAULT 0,
    skipped                 INT NOT NULL DEFAULT 0,
    warnings                INT NOT NULL DEFAULT 0,

    -- Who/what triggered it
    triggered_by            UUID REFERENCES users(id) ON DELETE SET NULL,  -- NULL for scheduled runs
    trigger_metadata        JSONB NOT NULL DEFAULT '{}',     -- e.g., {"source": "webhook", "webhook_id": "..."}

    -- Worker tracking
    worker_id               VARCHAR(100),                    -- which worker instance processed this run
    error_message           TEXT,                            -- if status = 'failed', what went wrong

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_test_runs_org ON test_runs (org_id);
CREATE INDEX idx_test_runs_org_status ON test_runs (org_id, status);
CREATE INDEX idx_test_runs_org_created ON test_runs (org_id, created_at DESC);
CREATE INDEX idx_test_runs_trigger ON test_runs (org_id, trigger_type);
CREATE INDEX idx_test_runs_started ON test_runs (org_id, started_at DESC)
    WHERE started_at IS NOT NULL;
CREATE INDEX idx_test_runs_pending ON test_runs (status, created_at)
    WHERE status = 'pending';

-- Trigger
CREATE TRIGGER trg_test_runs_updated_at
    BEFORE UPDATE ON test_runs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `run_number` is a SERIAL for human-readable run identification ("Run #42"). It's not globally unique — it auto-increments but isn't used as a key.
- **Denormalized counters** (`total_tests`, `passed`, `failed`, `errors`, `skipped`, `warnings`) are updated when the run completes. This avoids expensive COUNT queries on `test_results` for dashboard display.
- `worker_id` tracks which worker instance picked up the run. Useful for debugging in multi-worker deployments.
- `duration_ms` is computed as `completed_at - started_at` when the run finishes.
- `trigger_metadata` stores additional context: for webhook triggers, which webhook; for on_change, what changed; etc.
- `pending` status index supports the worker's polling query: "give me the next queued run."

---

### test_results

Individual test outcomes within a test run. Each row is "test X ran during sweep Y and the result was Z."

```sql
CREATE TABLE test_results (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    test_run_id             UUID NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
    test_id                 UUID NOT NULL REFERENCES tests(id) ON DELETE CASCADE,

    -- Denormalized for query performance (avoids joining through tests for every dashboard query)
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- Result
    status                  test_result_status NOT NULL,
    severity                test_severity NOT NULL,          -- copied from test at execution time (snapshot)

    -- Details
    message                 TEXT,                            -- human-readable summary: "MFA is enabled for all users"
    details                 JSONB NOT NULL DEFAULT '{}',     -- structured output: actual vs expected values
    output_log              TEXT,                            -- raw execution log/output (capped at 64KB)
    error_message           TEXT,                            -- if status = 'error', what went wrong

    -- Timing
    started_at              TIMESTAMPTZ NOT NULL,
    completed_at            TIMESTAMPTZ,
    duration_ms             INT,                             -- execution time for this specific test

    -- Alert linkage
    alert_generated         BOOLEAN NOT NULL DEFAULT FALSE,  -- TRUE if this result triggered an alert
    alert_id                UUID,                            -- FK added after alerts table is created (see below)

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- No updated_at: test results are immutable once written
);

-- Indexes
CREATE INDEX idx_test_results_org ON test_results (org_id);
CREATE INDEX idx_test_results_run ON test_results (test_run_id);
CREATE INDEX idx_test_results_test ON test_results (test_id);
CREATE INDEX idx_test_results_control ON test_results (control_id);
CREATE INDEX idx_test_results_status ON test_results (org_id, status);
CREATE INDEX idx_test_results_severity ON test_results (org_id, severity);
CREATE INDEX idx_test_results_run_status ON test_results (test_run_id, status);
CREATE INDEX idx_test_results_control_latest ON test_results (control_id, created_at DESC);
CREATE INDEX idx_test_results_test_latest ON test_results (test_id, created_at DESC);
CREATE INDEX idx_test_results_failures ON test_results (org_id, created_at DESC)
    WHERE status IN ('fail', 'error');

-- Unique constraint: each test runs once per test_run
CREATE UNIQUE INDEX uq_test_results_run_test ON test_results (test_run_id, test_id);
```

**Design notes:**
- **Immutable**: Test results are write-once — no `updated_at`. Once a test produces a result, it's a historical fact.
- `control_id` is **denormalized** from `tests.control_id`. This avoids joining through `tests` for every control health query. The value is copied at execution time.
- `severity` is also **snapshotted** from the test definition at execution time. If the test's severity is later changed, historical results retain their original severity.
- `details` JSONB stores structured output: `{"expected": "enabled", "actual": "disabled", "resource": "arn:aws:iam::123456:policy/MFA"}`. The schema is test-type-specific.
- `output_log` captures raw script output for custom tests. Capped at 64KB in the API layer.
- `alert_generated` and `alert_id` provide forward/backward linkage between results and alerts. `alert_id` FK is deferred (see below).
- The unique index `(test_run_id, test_id)` ensures each test appears at most once per sweep.
- `idx_test_results_control_latest` supports the control health heatmap: "what's the latest result for each control?"
- `idx_test_results_failures` supports the alert generation engine: quickly find recent failures.

---

### alerts

Generated alerts from test failures or rule matches. Alerts follow the lifecycle from spec §3.3.2: Detect → Classify → Assign → Remediate → Verify → Close.

```sql
CREATE TABLE alerts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    alert_number            SERIAL,                          -- human-readable: "ALT-42"
    title                   VARCHAR(500) NOT NULL,
    description             TEXT,

    -- Classification
    severity                alert_severity NOT NULL,
    status                  alert_status NOT NULL DEFAULT 'open',

    -- Source (what triggered this alert)
    test_id                 UUID REFERENCES tests(id) ON DELETE SET NULL,
    test_result_id          UUID REFERENCES test_results(id) ON DELETE SET NULL,
    control_id              UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    alert_rule_id           UUID,                            -- FK added after alert_rules table (see below)

    -- Assignment
    assigned_to             UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at             TIMESTAMPTZ,
    assigned_by             UUID REFERENCES users(id) ON DELETE SET NULL,

    -- SLA tracking
    sla_deadline            TIMESTAMPTZ,                     -- when this alert must be resolved by
    sla_breached            BOOLEAN NOT NULL DEFAULT FALSE,  -- TRUE if sla_deadline has passed without resolution

    -- Resolution
    resolved_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_at             TIMESTAMPTZ,
    resolution_notes        TEXT,                            -- what was done to fix it

    -- Suppression (spec §3.3.2: snooze/suppress with justification)
    suppressed_until        TIMESTAMPTZ,                     -- when suppression expires
    suppression_reason      TEXT,                            -- mandatory justification

    -- Delivery tracking
    delivery_channels       alert_delivery_channel[] DEFAULT '{in_app}',
    delivered_at            JSONB NOT NULL DEFAULT '{}',     -- {"slack": "2026-02-20T10:00:00Z", "email": "2026-02-20T10:00:05Z"}

    -- Metadata
    tags                    TEXT[] DEFAULT '{}',
    metadata                JSONB NOT NULL DEFAULT '{}',     -- extensible: ticket links, external IDs, etc.

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_alerts_org ON alerts (org_id);
CREATE INDEX idx_alerts_org_status ON alerts (org_id, status);
CREATE INDEX idx_alerts_org_severity ON alerts (org_id, severity);
CREATE INDEX idx_alerts_org_status_severity ON alerts (org_id, status, severity);
CREATE INDEX idx_alerts_control ON alerts (control_id);
CREATE INDEX idx_alerts_test ON alerts (test_id)
    WHERE test_id IS NOT NULL;
CREATE INDEX idx_alerts_test_result ON alerts (test_result_id)
    WHERE test_result_id IS NOT NULL;
CREATE INDEX idx_alerts_assigned_to ON alerts (assigned_to)
    WHERE assigned_to IS NOT NULL;
CREATE INDEX idx_alerts_sla_deadline ON alerts (org_id, sla_deadline)
    WHERE status NOT IN ('resolved', 'closed', 'suppressed') AND sla_deadline IS NOT NULL;
CREATE INDEX idx_alerts_sla_breached ON alerts (org_id, sla_breached)
    WHERE sla_breached = TRUE AND status NOT IN ('resolved', 'closed');
CREATE INDEX idx_alerts_created ON alerts (org_id, created_at DESC);
CREATE INDEX idx_alerts_open ON alerts (org_id, created_at DESC)
    WHERE status IN ('open', 'acknowledged', 'in_progress');
CREATE INDEX idx_alerts_suppressed ON alerts (org_id, suppressed_until)
    WHERE status = 'suppressed' AND suppressed_until IS NOT NULL;
CREATE INDEX idx_alerts_tags ON alerts USING gin (tags);

-- Trigger
CREATE TRIGGER trg_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- `alert_number` is SERIAL for human-readable display ("ALT-42"). Combined with org, provides an easy reference.
- **Source tracing**: `test_id` + `test_result_id` trace exactly which test and which execution produced this alert. `control_id` is denormalized for direct control-level queries. All three can be NULL-ish except `control_id` — alerts always relate to a control.
- **SLA tracking**: `sla_deadline` is computed at creation time based on alert severity and org-configurable SLA durations (stored in `alert_rules` or org settings). `sla_breached` is a boolean flag updated by a periodic check (or on access).
- **Suppression**: Spec §3.3.2 requires snooze/suppress with mandatory justification and expiration. `suppressed_until` + `suppression_reason` implement this. When `suppressed_until` passes, the alert auto-reopens.
- **Delivery tracking**: `delivery_channels` is an array of how this alert should be delivered. `delivered_at` JSONB tracks when each channel was successfully notified (for retry logic).
- **Resolution workflow**: `resolved_by`, `resolved_at`, `resolution_notes` capture who fixed it and how. Status `resolved` means fix applied; `closed` means verified.
- `metadata` JSONB for extensibility: external ticket URLs (Jira, Linear), webhook response IDs, etc.
- Partial indexes on `status` filter active alerts efficiently — most queries are about unresolved alerts.

---

### alert_rules

Organization-configurable rules that define when and how alerts are generated. Rules match conditions (test type, severity, control category, etc.) and specify alert behavior (severity mapping, SLA, delivery channels, auto-assignment).

```sql
CREATE TABLE alert_rules (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    enabled                 BOOLEAN NOT NULL DEFAULT TRUE,

    -- Matching conditions (all must match — AND logic)
    match_test_types        test_type[],                     -- NULL = match all types
    match_severities        test_severity[],                 -- NULL = match all severities
    match_result_statuses   test_result_status[],            -- default: ['fail'] — what results trigger this rule
    match_control_ids       UUID[],                          -- NULL = match all controls; specific control IDs if targeted
    match_tags              TEXT[],                           -- NULL = match all; specific tags to filter

    -- Threshold (suppress noise)
    consecutive_failures    INT NOT NULL DEFAULT 1,          -- require N consecutive failures before alerting
    cooldown_minutes        INT NOT NULL DEFAULT 0,          -- don't re-alert for same test within N minutes

    -- Alert generation
    alert_severity          alert_severity NOT NULL,         -- severity to assign to generated alerts
    alert_title_template    VARCHAR(500),                    -- template: "{{test.title}} failed on {{control.identifier}}"
    auto_assign_to          UUID REFERENCES users(id) ON DELETE SET NULL,  -- auto-assign alerts to this user

    -- SLA
    sla_hours               INT,                             -- hours until SLA breach (NULL = no SLA)

    -- Delivery
    delivery_channels       alert_delivery_channel[] NOT NULL DEFAULT '{in_app}',
    slack_webhook_url       TEXT,                            -- org-specific Slack webhook (encrypted at rest)
    email_recipients        TEXT[],                          -- email addresses for alert delivery
    webhook_url             TEXT,                            -- custom outbound webhook URL
    webhook_headers         JSONB NOT NULL DEFAULT '{}',     -- custom headers for webhook delivery

    -- Priority / ordering
    priority                INT NOT NULL DEFAULT 100,        -- lower = higher priority; first matching rule wins

    -- Ownership
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_alert_rule_name UNIQUE (org_id, name),
    CONSTRAINT chk_sla_hours_positive CHECK (sla_hours IS NULL OR sla_hours > 0),
    CONSTRAINT chk_consecutive_positive CHECK (consecutive_failures > 0 AND consecutive_failures <= 100),
    CONSTRAINT chk_cooldown_nonneg CHECK (cooldown_minutes >= 0 AND cooldown_minutes <= 10080),  -- max 1 week
    CONSTRAINT chk_priority_range CHECK (priority >= 0 AND priority <= 1000)
);

-- Indexes
CREATE INDEX idx_alert_rules_org ON alert_rules (org_id);
CREATE INDEX idx_alert_rules_org_enabled ON alert_rules (org_id, enabled, priority)
    WHERE enabled = TRUE;
CREATE INDEX idx_alert_rules_created_by ON alert_rules (created_by)
    WHERE created_by IS NOT NULL;

-- Trigger
CREATE TRIGGER trg_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Design notes:**
- **Rule matching**: All `match_*` conditions use AND logic. Within each array field, OR logic applies. Example: `match_test_types = ['configuration', 'access_control']` AND `match_severities = ['critical', 'high']` means "configuration OR access_control tests" AND "critical OR high severity."
- **NULL = wildcard**: A NULL match field means "match everything." This allows simple rules like "alert on all critical failures" without specifying every test type.
- `consecutive_failures` implements threshold alerting — don't alert on a single flap. Requires N consecutive fail/error results for the same test before generating an alert. The worker tracks this from `test_results` history.
- `cooldown_minutes` prevents duplicate alerts for the same test within a time window. If test X fails at 10:00 and generates an alert, another failure at 10:15 won't generate a second alert (with a 60-minute cooldown).
- `alert_title_template` supports basic mustache-style templating. If NULL, a default title is generated from the test name.
- `priority` determines rule evaluation order — lower number = checked first. First matching rule wins (short-circuit).
- **Delivery configuration** is per-rule, allowing different notification channels for different alert types. Critical alerts might go to Slack + email, while low-severity ones are in-app only.
- `slack_webhook_url` should be encrypted at rest (application-level encryption). The schema stores it as TEXT; the API layer handles encryption/decryption.
- `email_recipients` is a text array of email addresses. Max enforced at API layer (max 20 recipients).

---

## Deferred Foreign Keys

After all tables are created, add the cross-reference FKs:

```sql
-- test_results → alerts (back-reference)
ALTER TABLE test_results
    ADD CONSTRAINT fk_test_results_alert
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE SET NULL;

CREATE INDEX idx_test_results_alert ON test_results (alert_id)
    WHERE alert_id IS NOT NULL;

-- alerts → alert_rules (which rule generated this alert)
ALTER TABLE alerts
    ADD CONSTRAINT fk_alerts_alert_rule
    FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE SET NULL;

CREATE INDEX idx_alerts_rule ON alerts (alert_rule_id)
    WHERE alert_rule_id IS NOT NULL;
```

---

## Migration Order

Migrations continue from Sprint 3:

19. `019_sprint4_enums.sql` — New enum types (test_type, test_status, test_severity, test_result_status, test_run_status, test_run_trigger, alert_severity, alert_status, alert_delivery_channel) + audit_action extensions
20. `020_tests.sql` — Tests table + indexes + trigger + constraints
21. `021_test_runs.sql` — Test runs table + indexes + trigger
22. `022_test_results.sql` — Test results table + indexes + unique constraint
23. `023_alerts.sql` — Alerts table + indexes + trigger
24. `024_alert_rules.sql` — Alert rules table + indexes + trigger + constraints
25. `025_sprint4_fk_cross_refs.sql` — Deferred foreign keys (test_results.alert_id → alerts, alerts.alert_rule_id → alert_rules)
26. `026_sprint4_seed.sql` — Seed data (example tests, alert rules)

---

## Seed Data

### Example Test Definitions

```sql
-- Example tests for the demo org's controls (from Sprint 2 control library seed)
INSERT INTO tests (
    id, org_id, identifier, title, description, test_type, severity, status,
    control_id, schedule_cron, timeout_seconds, test_config, tags, created_by
) VALUES
    -- MFA enforcement check
    (
        't0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'TST-AC-001',
        'MFA Enforcement Verification',
        'Verifies that multi-factor authentication is enforced for all users in the identity provider.',
        'access_control',
        'critical',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-001' LIMIT 1),
        '0 * * * *',  -- every hour
        120,
        '{"provider": "okta", "check": "mfa_enforced", "expected": true}'::JSONB,
        ARRAY['mfa', 'access-control', 'pci', 'soc2'],
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)
    ),
    -- Access review freshness check
    (
        't0000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'TST-AC-002',
        'Quarterly Access Review Completeness',
        'Verifies that access reviews have been completed within the required quarterly cadence.',
        'access_control',
        'high',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-003' LIMIT 1),
        '0 8 * * 1',  -- every Monday at 8am
        60,
        '{"cadence_days": 90, "check": "review_completed_within_cadence"}'::JSONB,
        ARRAY['access-review', 'quarterly', 'soc2'],
        (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)
    ),
    -- Encryption at rest check
    (
        't0000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'TST-DP-001',
        'Encryption at Rest — S3 Buckets',
        'Checks all S3 buckets have default encryption enabled (AES-256 or KMS).',
        'data_protection',
        'critical',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-DP-001' LIMIT 1),
        '0 */2 * * *',  -- every 2 hours
        180,
        '{"provider": "aws", "service": "s3", "check": "default_encryption", "expected_algorithms": ["AES256", "aws:kms"]}'::JSONB,
        ARRAY['encryption', 'aws', 's3', 'pci', 'data-protection'],
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1)
    ),
    -- CloudTrail logging enabled check
    (
        't0000000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'TST-LM-001',
        'CloudTrail Multi-Region Logging',
        'Verifies AWS CloudTrail is enabled in all regions with S3 log delivery.',
        'logging',
        'high',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-LM-001' LIMIT 1),
        '0 */4 * * *',  -- every 4 hours
        120,
        '{"provider": "aws", "service": "cloudtrail", "check": "multi_region_enabled", "expected": true}'::JSONB,
        ARRAY['logging', 'aws', 'cloudtrail', 'pci', 'soc2'],
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1)
    ),
    -- Vulnerability scan age check
    (
        't0000000-0000-0000-0000-000000000005',
        'a0000000-0000-0000-0000-000000000001',
        'TST-VM-001',
        'Monthly Vulnerability Scan Freshness',
        'Checks that a vulnerability scan was completed within the last 30 days.',
        'vulnerability',
        'high',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-VM-001' LIMIT 1),
        '0 9 * * *',  -- daily at 9am
        60,
        '{"cadence_days": 30, "check": "scan_completed_within_cadence", "scanner": "qualys"}'::JSONB,
        ARRAY['vulnerability', 'qualys', 'pci', 'monthly'],
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)
    ),
    -- Firewall rules audit
    (
        't0000000-0000-0000-0000-000000000006',
        'a0000000-0000-0000-0000-000000000001',
        'TST-NW-001',
        'Security Group — No Open Inbound 0.0.0.0/0',
        'Checks that no security groups allow unrestricted inbound access on sensitive ports.',
        'network',
        'critical',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-NW-001' LIMIT 1),
        '0 * * * *',  -- every hour
        120,
        '{"provider": "aws", "service": "ec2", "check": "no_open_ingress", "restricted_ports": [22, 3389, 3306, 5432]}'::JSONB,
        ARRAY['network', 'aws', 'firewall', 'pci'],
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1)
    ),
    -- Endpoint compliance check
    (
        't0000000-0000-0000-0000-000000000007',
        'a0000000-0000-0000-0000-000000000001',
        'TST-EP-001',
        'Endpoint Disk Encryption',
        'Verifies that all managed endpoints have full-disk encryption enabled.',
        'endpoint',
        'high',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-EP-001' LIMIT 1),
        '0 */6 * * *',  -- every 6 hours
        180,
        '{"provider": "jamf", "check": "filevault_enabled", "expected": true}'::JSONB,
        ARRAY['endpoint', 'encryption', 'jamf', 'soc2'],
        (SELECT id FROM users WHERE email = 'it@acme.example.com' LIMIT 1)
    ),
    -- Configuration baseline check
    (
        't0000000-0000-0000-0000-000000000008',
        'a0000000-0000-0000-0000-000000000001',
        'TST-CFG-001',
        'Password Policy — Minimum Complexity',
        'Verifies that password policy meets minimum complexity requirements (length, character types).',
        'configuration',
        'medium',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-002' LIMIT 1),
        '0 8 * * *',  -- daily at 8am
        60,
        '{"provider": "okta", "check": "password_policy", "min_length": 12, "require_uppercase": true, "require_number": true, "require_special": true}'::JSONB,
        ARRAY['configuration', 'password', 'okta', 'pci'],
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)
    );
```

### Example Alert Rules

```sql
INSERT INTO alert_rules (
    id, org_id, name, description, enabled,
    match_test_types, match_severities, match_result_statuses,
    consecutive_failures, cooldown_minutes,
    alert_severity, sla_hours,
    delivery_channels, email_recipients,
    priority, created_by
) VALUES
    -- Critical failures → immediate alert, Slack + email
    (
        'ar000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'Critical Test Failures',
        'Alert immediately on any critical test failure. Delivered via Slack and email to security team.',
        TRUE,
        NULL,                                          -- match all test types
        ARRAY['critical']::test_severity[],            -- only critical severity tests
        ARRAY['fail']::test_result_status[],           -- only on failure
        1,                                             -- alert on first failure
        60,                                            -- don't re-alert within 60 min
        'critical',                                    -- critical alert
        4,                                             -- 4-hour SLA
        ARRAY['slack', 'email', 'in_app']::alert_delivery_channel[],
        ARRAY['security@acme.example.com', 'ciso@acme.example.com'],
        10,                                            -- highest priority rule
        (SELECT id FROM users WHERE email = 'ciso@acme.example.com' LIMIT 1)
    ),
    -- High failures → alert after 2 consecutive, email
    (
        'ar000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'High Severity Failures',
        'Alert on high-severity test failures after 2 consecutive failures. Email notification to compliance team.',
        TRUE,
        NULL,
        ARRAY['high']::test_severity[],
        ARRAY['fail']::test_result_status[],
        2,                                             -- require 2 consecutive failures
        120,                                           -- 2-hour cooldown
        'high',
        24,                                            -- 24-hour SLA
        ARRAY['email', 'in_app']::alert_delivery_channel[],
        ARRAY['compliance@acme.example.com'],
        20,
        (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)
    ),
    -- Medium failures → alert after 3 consecutive, in-app only
    (
        'ar000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'Medium Severity Findings',
        'Alert on medium-severity findings after 3 consecutive failures. In-app notification only.',
        TRUE,
        NULL,
        ARRAY['medium']::test_severity[],
        ARRAY['fail']::test_result_status[],
        3,
        360,                                           -- 6-hour cooldown
        'medium',
        72,                                            -- 72-hour SLA
        ARRAY['in_app']::alert_delivery_channel[],
        NULL,
        30,
        (SELECT id FROM users WHERE email = 'compliance@acme.example.com' LIMIT 1)
    ),
    -- Test execution errors → alert on infra issues
    (
        'ar000000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'Test Execution Errors',
        'Alert when tests cannot execute (infrastructure issues, timeouts, connection failures).',
        TRUE,
        NULL,
        NULL,                                          -- all severities
        ARRAY['error']::test_result_status[],          -- only on errors (not failures)
        3,                                             -- 3 consecutive errors before alerting
        240,                                           -- 4-hour cooldown
        'high',
        24,
        ARRAY['email', 'in_app']::alert_delivery_channel[],
        ARRAY['devops@acme.example.com'],
        15,
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1)
    );
```

---

## Query Patterns

### Control Health Heatmap

"For each active control, what's its current health status (last test result)?"

```sql
SELECT
    c.id AS control_id,
    c.identifier,
    c.title,
    c.category,
    tr.status AS latest_result,
    tr.severity,
    tr.created_at AS last_tested_at,
    tr.message,
    t.identifier AS test_identifier,
    t.title AS test_title,
    CASE
        WHEN tr.status = 'pass' THEN 'healthy'
        WHEN tr.status = 'fail' THEN 'failing'
        WHEN tr.status = 'error' THEN 'error'
        WHEN tr.status = 'warning' THEN 'warning'
        WHEN tr.status IS NULL THEN 'untested'
        ELSE 'unknown'
    END AS health_status
FROM controls c
LEFT JOIN LATERAL (
    SELECT trs.status, trs.severity, trs.created_at, trs.message, trs.test_id
    FROM test_results trs
    WHERE trs.control_id = c.id
      AND trs.org_id = c.org_id
    ORDER BY trs.created_at DESC
    LIMIT 1
) tr ON TRUE
LEFT JOIN tests t ON t.id = tr.test_id
WHERE c.org_id = $1
  AND c.status = 'active'
ORDER BY
    CASE
        WHEN tr.status = 'fail' THEN 0
        WHEN tr.status = 'error' THEN 1
        WHEN tr.status = 'warning' THEN 2
        WHEN tr.status IS NULL THEN 3
        WHEN tr.status = 'pass' THEN 4
        ELSE 5
    END,
    c.identifier;
```

### Alert Queue (Active Alerts)

"Show me all active alerts, grouped by status, ordered by severity and SLA urgency."

```sql
SELECT
    a.id,
    a.alert_number,
    a.title,
    a.severity,
    a.status,
    a.control_id,
    c.identifier AS control_identifier,
    c.title AS control_title,
    a.assigned_to,
    u.name AS assigned_to_name,
    a.sla_deadline,
    a.sla_breached,
    CASE
        WHEN a.sla_deadline IS NULL THEN NULL
        WHEN a.sla_deadline < NOW() THEN EXTRACT(EPOCH FROM (NOW() - a.sla_deadline)) / 3600
        ELSE NULL
    END AS hours_overdue,
    CASE
        WHEN a.sla_deadline IS NULL THEN NULL
        WHEN a.sla_deadline > NOW() THEN EXTRACT(EPOCH FROM (a.sla_deadline - NOW())) / 3600
        ELSE NULL
    END AS hours_remaining,
    a.created_at
FROM alerts a
JOIN controls c ON c.id = a.control_id
LEFT JOIN users u ON u.id = a.assigned_to
WHERE a.org_id = $1
  AND a.status IN ('open', 'acknowledged', 'in_progress')
ORDER BY
    CASE a.severity
        WHEN 'critical' THEN 0
        WHEN 'high' THEN 1
        WHEN 'medium' THEN 2
        WHEN 'low' THEN 3
    END,
    COALESCE(a.sla_deadline, '9999-12-31'::TIMESTAMPTZ) ASC,
    a.created_at ASC;
```

### Compliance Posture Score per Framework

"For each activated framework, what percentage of controls are passing?"

```sql
SELECT
    f.id AS framework_id,
    f.name AS framework_name,
    fv.version AS framework_version,
    of.id AS org_framework_id,
    COUNT(DISTINCT cm.control_id) AS total_mapped_controls,
    COUNT(DISTINCT cm.control_id) FILTER (
        WHERE latest_result.status = 'pass'
    ) AS passing_controls,
    COUNT(DISTINCT cm.control_id) FILTER (
        WHERE latest_result.status = 'fail'
    ) AS failing_controls,
    COUNT(DISTINCT cm.control_id) FILTER (
        WHERE latest_result.status IS NULL
    ) AS untested_controls,
    CASE
        WHEN COUNT(DISTINCT cm.control_id) = 0 THEN 0
        ELSE ROUND(
            100.0 * COUNT(DISTINCT cm.control_id) FILTER (WHERE latest_result.status = 'pass')
            / COUNT(DISTINCT cm.control_id),
            1
        )
    END AS posture_score_pct
FROM org_frameworks of
JOIN framework_versions fv ON fv.id = of.framework_version_id
JOIN frameworks f ON f.id = fv.framework_id
JOIN requirements r ON r.framework_version_id = fv.id
JOIN control_mappings cm ON cm.requirement_id = r.id AND cm.org_id = of.org_id
LEFT JOIN LATERAL (
    SELECT tr.status
    FROM test_results tr
    WHERE tr.control_id = cm.control_id
      AND tr.org_id = of.org_id
    ORDER BY tr.created_at DESC
    LIMIT 1
) latest_result ON TRUE
WHERE of.org_id = $1
  AND of.status = 'active'
GROUP BY f.id, f.name, fv.version, of.id
ORDER BY f.name;
```

### Test Execution History for a Control

"Show me the last 20 test runs for CTRL-AC-001."

```sql
SELECT
    tr.id,
    tr.test_id,
    t.identifier AS test_identifier,
    t.title AS test_title,
    tr.status,
    tr.severity,
    tr.message,
    tr.duration_ms,
    tr.alert_generated,
    tr.started_at,
    tr.completed_at,
    run.run_number,
    run.trigger_type
FROM test_results tr
JOIN tests t ON t.id = tr.test_id
JOIN test_runs run ON run.id = tr.test_run_id
WHERE tr.control_id = $1
  AND tr.org_id = $2
ORDER BY tr.created_at DESC
LIMIT 20;
```

### Consecutive Failure Count (for Alert Rule Evaluation)

"How many consecutive failures does test X have?"

```sql
WITH recent_results AS (
    SELECT
        status,
        created_at,
        ROW_NUMBER() OVER (ORDER BY created_at DESC) AS rn
    FROM test_results
    WHERE test_id = $1
      AND org_id = $2
    ORDER BY created_at DESC
    LIMIT 100  -- look back max 100 results
)
SELECT COUNT(*) AS consecutive_failures
FROM recent_results
WHERE rn <= (
    -- Find the position of the first non-fail result
    SELECT COALESCE(MIN(rn) - 1, (SELECT COUNT(*) FROM recent_results))
    FROM recent_results
    WHERE status NOT IN ('fail', 'error')
);
```

### SLA Breach Check (periodic worker job)

"Find all alerts where SLA has been breached but flag hasn't been set."

```sql
UPDATE alerts
SET sla_breached = TRUE, updated_at = NOW()
WHERE sla_deadline < NOW()
  AND sla_breached = FALSE
  AND status NOT IN ('resolved', 'closed', 'suppressed')
RETURNING id, org_id, alert_number, title, severity, sla_deadline;
```

---

## Worker Architecture Notes

The continuous monitoring worker is a background Go process (separate goroutine pool or separate binary) that:

1. **Polls for due tests**: `SELECT * FROM tests WHERE status = 'active' AND next_run_at <= NOW()` (using the `idx_tests_next_run` index).
2. **Creates a test run**: Inserts a `test_runs` row with `status = 'pending'`.
3. **Executes tests**: For each test, runs the appropriate check based on `test_type` + `test_config`. Writes `test_results` rows.
4. **Updates run summary**: Sets `status = 'completed'`, fills in denormalized counters.
5. **Evaluates alert rules**: For each `fail` or `error` result, checks `alert_rules` (ordered by `priority`). If a rule matches and thresholds are met (consecutive failures, cooldown), generates an alert.
6. **Delivers alerts**: For each new alert, sends notifications via configured delivery channels (Slack webhook, email, custom webhook).
7. **Updates next_run_at**: Computes the next execution time from `schedule_cron` or `schedule_interval_min`.

The worker should be idempotent — if it crashes mid-sweep, the `test_run` status will be `running`, and a recovery process can clean it up.

---

## Future Considerations

- **Integration-driven tests** (Sprint 9): When integrations are built, tests will pull data from connected systems. The `source_integration_id` FK is pre-positioned.
- **Custom test builder UI** (spec §3.2.2): The `test_script` + `test_script_language` fields support Compliance as Code. A visual Test Builder could generate these scripts.
- **Alert grouping** (spec §3.3.2): "Alert grouping to prevent notification fatigue." The `cooldown_minutes` field is a start; more sophisticated grouping (by control, by category) could be added via a future `alert_groups` table.
- **Auto-remediation**: Alerts could trigger automated remediation actions. The `metadata` JSONB field on alerts can store remediation playbook references. A future `alert_actions` table could define automated responses.
- **Jira/Linear/GitHub ticket creation** (spec §3.3.2): Alert rules could include `auto_create_ticket` config. The `metadata` JSONB on alerts stores the external ticket URL after creation.
- **Real-time websocket updates**: The monitoring dashboard could use WebSocket connections to push live test results and alert status changes. The event-driven architecture (audit log) provides the event source.
