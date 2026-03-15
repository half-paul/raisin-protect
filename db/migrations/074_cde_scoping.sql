-- Migration: 074_cde_scoping.sql
-- Description: CDE Scoping Module — asset inventory, network segments, data flows,
--              segmentation tests, and scope history (PCI DSS v4.0.1 Req 1.2.4, 11.4.1)
-- Created: 2026-03-15
-- Sprint: 11 — CDE Scoping Module (LEX-83)
-- ADR: docs/adrs/ADR-001-cde-scoping-module.md

-- ============================================================================
-- TABLE 1: cde_assets
-- Represents any system component relevant to CDE scoping (servers, applications,
-- network devices, cloud services, etc.)
-- ============================================================================

CREATE TABLE IF NOT EXISTS cde_assets (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    name                    TEXT NOT NULL,
    type                    TEXT NOT NULL
                                CHECK (type IN (
                                    'server', 'workstation', 'network_device', 'database',
                                    'application', 'virtual_machine', 'container_cluster',
                                    'cloud_service', 'network_segment', 'storage_system',
                                    'terminal', 'other'
                                )),
    -- Location / addressing
    ip_address              INET,
    hostname                TEXT,
    environment             TEXT NOT NULL DEFAULT 'production'
                                CHECK (environment IN ('production', 'staging', 'development', 'test')),

    -- CDE scope classification
    scope_status            TEXT NOT NULL DEFAULT 'unclassified'
                                CHECK (scope_status IN (
                                    'in_scope',       -- Stores, processes, or transmits CHD/SAD
                                    'connected_to_cde', -- Connected to CDE, does not handle CHD/SAD
                                    'out_of_scope',   -- Isolated from CDE with verified segmentation
                                    'unclassified'    -- Awaiting classification decision
                                )),
    scope_justification     TEXT,       -- Required for out_of_scope and connected_to_cde
    scope_decided_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    scope_decided_at        TIMESTAMPTZ,

    -- Data classification (what sensitive data this asset handles)
    data_classification     TEXT NOT NULL DEFAULT 'public'
                                CHECK (data_classification IN ('pan', 'chd', 'sad', 'confidential', 'public')),

    -- Ownership and lifecycle
    owner_id                UUID REFERENCES users(id) ON DELETE SET NULL,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    tags                    TEXT[] NOT NULL DEFAULT '{}',

    -- Audit
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE cde_assets IS
    'Inventory of all system components relevant to CDE scope classification. '
    'Satisfies PCI DSS v4.0.1 Req 1.2.4 (asset inventory) and Req 11.4.1 (scope boundary documentation).';
COMMENT ON COLUMN cde_assets.scope_status IS
    'CDE classification: in_scope=handles CHD/SAD, connected_to_cde=adjacent but no CHD, '
    'out_of_scope=isolated (requires segmentation test), unclassified=pending decision';
COMMENT ON COLUMN cde_assets.data_classification IS
    'Highest sensitivity of data processed by this asset: pan/chd/sad/confidential/public';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cde_assets_org           ON cde_assets (org_id);
CREATE INDEX IF NOT EXISTS idx_cde_assets_org_scope     ON cde_assets (org_id, scope_status);
CREATE INDEX IF NOT EXISTS idx_cde_assets_org_type      ON cde_assets (org_id, type);
CREATE INDEX IF NOT EXISTS idx_cde_assets_org_active    ON cde_assets (org_id, is_active);
CREATE INDEX IF NOT EXISTS idx_cde_assets_org_class     ON cde_assets (org_id, data_classification);
CREATE INDEX IF NOT EXISTS idx_cde_assets_owner         ON cde_assets (owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_cde_assets_tags          ON cde_assets USING GIN (tags);

-- RLS — org isolation at database layer (defense in depth)
ALTER TABLE cde_assets ENABLE ROW LEVEL SECURITY;
CREATE POLICY cde_assets_org_isolation ON cde_assets
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_cde_assets_updated_at ON cde_assets;
CREATE TRIGGER trg_cde_assets_updated_at
    BEFORE UPDATE ON cde_assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 2: cde_network_segments
-- Tracks distinct network segments/VLANs and their relationship to the CDE.
-- Referenced by cde_segmentation_tests.
-- ============================================================================

CREATE TABLE IF NOT EXISTS cde_network_segments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    name                TEXT NOT NULL,
    vlan                TEXT,                   -- VLAN ID or tag (e.g., "101", "PCI-VLAN")
    subnet              TEXT,                   -- CIDR notation (e.g., "10.1.5.0/24")
    segment_type        TEXT NOT NULL
                            CHECK (segment_type IN (
                                'cde',          -- Inside the CDE
                                'connected',    -- Connected to CDE, out-of-scope systems
                                'out_of_scope', -- Fully isolated from CDE
                                'dmz',          -- Demilitarized zone
                                'management'    -- Management/admin network
                            )),
    isolation_method    TEXT,   -- e.g., "VLAN + ACL", "physical firewall", "SDN policy"
    description         TEXT,

    -- Audit
    created_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE cde_network_segments IS
    'Network segments/VLANs and their CDE scope classification. '
    'Used to document network topology for QSA review (PCI DSS v4.0.1 Req 1.2.3).';
COMMENT ON COLUMN cde_network_segments.isolation_method IS
    'Technical method used to isolate this segment from in-scope networks '
    '(e.g., "stateful firewall + deny-all ACL", "physical separation", "SDN micro-segmentation")';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cde_net_seg_org          ON cde_network_segments (org_id);
CREATE INDEX IF NOT EXISTS idx_cde_net_seg_org_type     ON cde_network_segments (org_id, segment_type);

-- RLS
ALTER TABLE cde_network_segments ENABLE ROW LEVEL SECURITY;
CREATE POLICY cde_network_segments_org_isolation ON cde_network_segments
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_cde_network_segments_updated_at ON cde_network_segments;
CREATE TRIGGER trg_cde_network_segments_updated_at
    BEFORE UPDATE ON cde_network_segments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 3: cde_data_flows
-- Documents how CHD/SAD flows between assets. Required by PCI DSS Req 1.2.4.
-- ============================================================================

CREATE TABLE IF NOT EXISTS cde_data_flows (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Graph edge: source → destination
    source_asset_id     UUID NOT NULL REFERENCES cde_assets(id) ON DELETE CASCADE,
    dest_asset_id       UUID NOT NULL REFERENCES cde_assets(id) ON DELETE CASCADE,

    -- Flow metadata
    protocol            TEXT,               -- e.g., "TLS 1.3", "HTTPS", "SSH", "SFTP"
    port                INTEGER,            -- Destination port (1-65535)
    data_type           TEXT,               -- e.g., "pan", "sad", "authorization_request"
    encryption_method   TEXT,               -- e.g., "TLS 1.3", "AES-256-GCM", "none"

    -- Additional context
    description         TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,

    -- Audit
    created_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Self-loops are nonsensical for data flow documentation
    CONSTRAINT cde_data_flows_no_self_loop CHECK (source_asset_id != dest_asset_id),
    -- Port range validation
    CONSTRAINT cde_data_flows_port_range   CHECK (port IS NULL OR (port >= 1 AND port <= 65535))
);

COMMENT ON TABLE cde_data_flows IS
    'Data flow registry documenting how CHD/SAD moves between CDE assets. '
    'Provides evidence for PCI DSS v4.0.1 Req 1.2.4 (accurate data-flow diagrams).';
COMMENT ON COLUMN cde_data_flows.encryption_method IS
    'Transport encryption in use for this flow. "none" indicates an unencrypted flow '
    'requiring remediation per PCI DSS Req 4.2.1.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cde_data_flows_org           ON cde_data_flows (org_id);
CREATE INDEX IF NOT EXISTS idx_cde_data_flows_source        ON cde_data_flows (source_asset_id);
CREATE INDEX IF NOT EXISTS idx_cde_data_flows_dest          ON cde_data_flows (dest_asset_id);
CREATE INDEX IF NOT EXISTS idx_cde_data_flows_org_active    ON cde_data_flows (org_id, is_active);

-- RLS
ALTER TABLE cde_data_flows ENABLE ROW LEVEL SECURITY;
CREATE POLICY cde_data_flows_org_isolation ON cde_data_flows
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_cde_data_flows_updated_at ON cde_data_flows;
CREATE TRIGGER trg_cde_data_flows_updated_at
    BEFORE UPDATE ON cde_data_flows
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 4: cde_segmentation_tests
-- Tracks network segmentation test results. PCI DSS Req 11.4.1 requires testing
-- at least every 6 months and after any significant infrastructure changes.
-- ============================================================================

CREATE TABLE IF NOT EXISTS cde_segmentation_tests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Test subject
    segment_id              UUID REFERENCES cde_network_segments(id) ON DELETE SET NULL,
    title                   TEXT NOT NULL,
    description             TEXT,

    -- Test details
    methodology             TEXT
                                CHECK (methodology IS NULL OR methodology IN (
                                    'internal_scan',
                                    'external_penetration',
                                    'firewall_rule_review',
                                    'manual_verification',
                                    'automated_tool'
                                )),
    tester                  TEXT NOT NULL,          -- Tester name or company
    tester_user_id          UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Dates
    test_date               DATE NOT NULL,
    next_test_date          DATE,                   -- Computed as test_date + 6 months by API

    -- Results
    result                  TEXT NOT NULL DEFAULT 'n/a'
                                CHECK (result IN ('pass', 'fail', 'n/a')),
    findings                TEXT,                   -- Narrative summary of findings

    -- Audit
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE cde_segmentation_tests IS
    'Network segmentation test records. PCI DSS v4.0.1 Req 11.4.1 mandates testing '
    'at least every 6 months and after any significant infrastructure changes. '
    'next_test_date is set to test_date + 6 months by the API layer.';
COMMENT ON COLUMN cde_segmentation_tests.segment_id IS
    'The network segment tested. NULL if the test covers multiple segments or the full perimeter.';
COMMENT ON COLUMN cde_segmentation_tests.result IS
    'pass=segmentation verified, fail=breach detected (remediation required), n/a=informational scan';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cde_seg_tests_org        ON cde_segmentation_tests (org_id);
CREATE INDEX IF NOT EXISTS idx_cde_seg_tests_segment    ON cde_segmentation_tests (segment_id);
CREATE INDEX IF NOT EXISTS idx_cde_seg_tests_org_date   ON cde_segmentation_tests (org_id, test_date DESC);
CREATE INDEX IF NOT EXISTS idx_cde_seg_tests_due        ON cde_segmentation_tests (org_id, next_test_date)
    WHERE next_test_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_cde_seg_tests_result     ON cde_segmentation_tests (org_id, result);

-- RLS
ALTER TABLE cde_segmentation_tests ENABLE ROW LEVEL SECURITY;
CREATE POLICY cde_segmentation_tests_org_isolation ON cde_segmentation_tests
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_cde_seg_tests_updated_at ON cde_segmentation_tests;
CREATE TRIGGER trg_cde_seg_tests_updated_at
    BEFORE UPDATE ON cde_segmentation_tests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 5: cde_scope_history
-- Immutable audit trail for scope classification changes (PCI DSS audit trail req).
-- ============================================================================

CREATE TABLE IF NOT EXISTS cde_scope_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    asset_id        UUID NOT NULL REFERENCES cde_assets(id) ON DELETE CASCADE,
    previous_status TEXT,
    new_status      TEXT NOT NULL,
    justification   TEXT,
    changed_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE cde_scope_history IS
    'Immutable audit trail of scope classification changes. '
    'Rows are NEVER updated or deleted — INSERT only. '
    'Provides evidence that scoping decisions are controlled and reviewable by QSAs.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cde_scope_history_asset  ON cde_scope_history (asset_id, changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_cde_scope_history_org    ON cde_scope_history (org_id, changed_at DESC);

-- RLS (read-only for non-privileged connections; no updates ever expected)
ALTER TABLE cde_scope_history ENABLE ROW LEVEL SECURITY;
CREATE POLICY cde_scope_history_org_isolation ON cde_scope_history
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- ============================================================================
-- Cross-reference: extend evidence_link_target_type to support CDE assets
-- ============================================================================
DO $$ BEGIN
    ALTER TYPE evidence_link_target_type ADD VALUE IF NOT EXISTS 'cde_asset';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TYPE evidence_link_target_type ADD VALUE IF NOT EXISTS 'cde_segmentation_test';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
