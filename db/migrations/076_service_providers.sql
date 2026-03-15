-- Migration: 076_service_providers.sql
-- Description: Service Provider / Vendor Management (PCI DSS v4.0.1 Req 12.8, 12.9)
-- Created: 2026-03-15
-- Sprint: 12 — Service Provider Management (LEX-91)
--
-- PCI DSS Req 12.8: Manage service providers with whom account data is shared,
--   or that could affect the security of account data.
-- PCI DSS Req 12.9: Service providers acknowledge their responsibility to protect
--   cardholder data.
--
-- Tables:
--   1. service_providers       — vendor/SP registry with risk classification
--   2. sp_compliance_documents — AOC/SOC 2/ISO 27001 document tracking per SP
--   3. sp_responsibility_matrix — which PCI DSS requirements each SP owns/shares

-- ============================================================================
-- TABLE 1: service_providers
-- Registry of all vendors and service providers that store, process, or transmit
-- cardholder data, or could affect its security.
-- ============================================================================

CREATE TABLE IF NOT EXISTS service_providers (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Identity
    name                    TEXT NOT NULL,
    type                    TEXT NOT NULL
                                CHECK (type IN (
                                    'payment_processor',    -- Processes card transactions
                                    'payment_gateway',      -- Routes payment authorizations
                                    'acquirer',             -- Acquiring bank / merchant account
                                    'tokenization',         -- Tokenizes/de-tokenizes PANs
                                    'hosting',              -- Hosts CDE systems (IaaS/PaaS)
                                    'managed_security',     -- MSSP / SOC-as-a-service
                                    'software',             -- Payment application vendor
                                    'network',              -- Network connectivity / SD-WAN
                                    'cloud_storage',        -- Cloud storage with CHD
                                    'third_party_agent',    -- Contracted field agent with CHD access
                                    'other'
                                )),

    -- Primary contact
    contact_name            TEXT,
    contact_email           TEXT,
    contact_phone           TEXT,

    -- Services provided to the org
    services_provided       TEXT,       -- Free-text description for audit evidence

    -- PCI DSS compliance status
    pci_compliance_status   TEXT NOT NULL DEFAULT 'unknown'
                                CHECK (pci_compliance_status IN (
                                    'compliant',                -- AOC on file, within validity
                                    'compliance_in_progress',   -- Undergoing assessment
                                    'compliance_not_validated', -- No current AOC / PCI status unknown
                                    'non_compliant',            -- Known gaps or expired without renewal
                                    'not_applicable',           -- SP scope does not include CHD
                                    'unknown'                   -- Status not yet determined
                                )),
    last_aoc_date           DATE,           -- Date of most recent AOC received
    next_review_date        DATE,           -- Scheduled next compliance review

    -- Risk classification
    risk_level              TEXT NOT NULL DEFAULT 'medium'
                                CHECK (risk_level IN ('critical', 'high', 'medium', 'low')),
    risk_notes              TEXT,

    -- Relationship metadata
    contract_start_date     DATE,
    contract_end_date       DATE,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,

    -- Audit
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_service_provider_name UNIQUE (org_id, name)
);

COMMENT ON TABLE service_providers IS
    'Registry of vendors and service providers per PCI DSS v4.0.1 Req 12.8. '
    'Covers all entities that store, process, or transmit CHD, or could affect its security.';
COMMENT ON COLUMN service_providers.pci_compliance_status IS
    'Current PCI DSS compliance posture. Must be reviewed at least annually per Req 12.8.4.';
COMMENT ON COLUMN service_providers.next_review_date IS
    'Next scheduled compliance review date. Should be set to 1 year from last_aoc_date at minimum.';
COMMENT ON COLUMN service_providers.risk_level IS
    'Risk classification: critical=direct CHD handler, high=adjacent/significant access, '
    'medium=indirect exposure, low=minimal/no CHD access.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sp_org                   ON service_providers (org_id);
CREATE INDEX IF NOT EXISTS idx_sp_org_type              ON service_providers (org_id, type);
CREATE INDEX IF NOT EXISTS idx_sp_org_status            ON service_providers (org_id, pci_compliance_status);
CREATE INDEX IF NOT EXISTS idx_sp_org_risk              ON service_providers (org_id, risk_level);
CREATE INDEX IF NOT EXISTS idx_sp_org_active            ON service_providers (org_id, is_active);
CREATE INDEX IF NOT EXISTS idx_sp_next_review           ON service_providers (org_id, next_review_date)
    WHERE next_review_date IS NOT NULL AND is_active = TRUE;

-- RLS
ALTER TABLE service_providers ENABLE ROW LEVEL SECURITY;
CREATE POLICY service_providers_org_isolation ON service_providers
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_service_providers_updated_at ON service_providers;
CREATE TRIGGER trg_service_providers_updated_at
    BEFORE UPDATE ON service_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 2: sp_compliance_documents
-- Tracks compliance documentation (AOC, SOC 2, ISO 27001) received from each SP.
-- PCI DSS Req 12.8.3: Maintain a program to engage service providers,
-- including a policy for managing (monitoring) SP compliance.
-- ============================================================================

CREATE TABLE IF NOT EXISTS sp_compliance_documents (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id         UUID NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    document_type       TEXT NOT NULL
                            CHECK (document_type IN (
                                'aoc',          -- PCI DSS Attestation of Compliance
                                'soc2_type1',   -- SOC 2 Type I report
                                'soc2_type2',   -- SOC 2 Type II report
                                'iso27001',     -- ISO/IEC 27001 certification
                                'csa_star',     -- CSA STAR certification
                                'pentest',      -- Penetration test report
                                'questionnaire',-- Security questionnaire (CAIQ, SIG, etc.)
                                'other'
                            )),
    title               TEXT,               -- e.g., "Stripe Inc AOC 2025"
    document_version    TEXT,               -- e.g., "PCI DSS 4.0.1"

    -- File storage
    upload_path         TEXT,               -- MinIO object key or file path

    -- Validity period
    valid_from          DATE,
    valid_until         DATE,

    -- Review metadata
    reviewed_by         UUID REFERENCES users(id) ON DELETE SET NULL,
    review_notes        TEXT,
    is_current          BOOLEAN NOT NULL DEFAULT TRUE,  -- Latest valid doc for this type

    -- Audit
    uploaded_by         UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Validity date range must be logical if both are specified
    CONSTRAINT chk_sp_doc_validity_range CHECK (
        valid_from IS NULL OR valid_until IS NULL OR valid_from <= valid_until
    )
);

COMMENT ON TABLE sp_compliance_documents IS
    'Compliance document library per service provider. Tracks AOC, SOC 2, ISO 27001, '
    'and other evidence received from vendors. Required by PCI DSS Req 12.8.3 / 12.8.4.';
COMMENT ON COLUMN sp_compliance_documents.is_current IS
    'TRUE if this is the active/current document for this provider+type combination. '
    'Previous documents should be set to is_current=FALSE when a newer version is uploaded.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sp_comp_docs_provider    ON sp_compliance_documents (provider_id);
CREATE INDEX IF NOT EXISTS idx_sp_comp_docs_org         ON sp_compliance_documents (org_id);
CREATE INDEX IF NOT EXISTS idx_sp_comp_docs_type        ON sp_compliance_documents (provider_id, document_type);
CREATE INDEX IF NOT EXISTS idx_sp_comp_docs_current     ON sp_compliance_documents (provider_id, document_type, is_current)
    WHERE is_current = TRUE;
CREATE INDEX IF NOT EXISTS idx_sp_comp_docs_expiry      ON sp_compliance_documents (org_id, valid_until)
    WHERE valid_until IS NOT NULL AND is_current = TRUE;

-- RLS
ALTER TABLE sp_compliance_documents ENABLE ROW LEVEL SECURITY;
CREATE POLICY sp_compliance_documents_org_isolation ON sp_compliance_documents
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_sp_comp_docs_updated_at ON sp_compliance_documents;
CREATE TRIGGER trg_sp_comp_docs_updated_at
    BEFORE UPDATE ON sp_compliance_documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 3: sp_responsibility_matrix
-- Documents which party (merchant/provider/shared) is responsible for each
-- PCI DSS requirement when working with a given service provider.
-- PCI DSS Req 12.9.2: Service providers and their customers maintain a
-- documented list of PCI DSS requirements managed by the SP vs. the entity.
-- ============================================================================

CREATE TABLE IF NOT EXISTS sp_responsibility_matrix (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id         UUID NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requirement_id      UUID REFERENCES requirements(id) ON DELETE SET NULL,

    -- Denormalized requirement code for display even if requirement is later removed
    requirement_code    TEXT NOT NULL,

    responsible_party   TEXT NOT NULL
                            CHECK (responsible_party IN (
                                'merchant',     -- Merchant organization is fully responsible
                                'provider',     -- Service provider is fully responsible
                                'shared'        -- Both parties share responsibility
                            )),
    notes               TEXT,   -- Implementation notes, how shared responsibility is divided, etc.

    -- Audit
    created_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each (provider, requirement) pair has one responsibility assignment
    CONSTRAINT uq_sp_responsibility UNIQUE (provider_id, requirement_id),
    -- Also unique by provider + code for cases where requirement_id is NULL
    CONSTRAINT uq_sp_responsibility_code UNIQUE (provider_id, requirement_code)
);

COMMENT ON TABLE sp_responsibility_matrix IS
    'Documents PCI DSS responsibility allocation between merchant and service provider. '
    'Required by PCI DSS v4.0.1 Req 12.9.2. '
    'Each row maps one PCI DSS requirement to the responsible party for a specific SP relationship.';
COMMENT ON COLUMN sp_responsibility_matrix.responsible_party IS
    'merchant=org handles this req, provider=SP handles it, shared=both parties have controls. '
    '"shared" requires explanation in notes field for QSA review.';
COMMENT ON COLUMN sp_responsibility_matrix.requirement_code IS
    'Denormalized PCI DSS requirement code (e.g., "1.3.2") — stable even if requirement record changes.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sp_matrix_provider       ON sp_responsibility_matrix (provider_id);
CREATE INDEX IF NOT EXISTS idx_sp_matrix_org            ON sp_responsibility_matrix (org_id);
CREATE INDEX IF NOT EXISTS idx_sp_matrix_requirement    ON sp_responsibility_matrix (requirement_id)
    WHERE requirement_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sp_matrix_party          ON sp_responsibility_matrix (provider_id, responsible_party);

-- RLS
ALTER TABLE sp_responsibility_matrix ENABLE ROW LEVEL SECURITY;
CREATE POLICY sp_responsibility_matrix_org_isolation ON sp_responsibility_matrix
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_sp_matrix_updated_at ON sp_responsibility_matrix;
CREATE TRIGGER trg_sp_matrix_updated_at
    BEFORE UPDATE ON sp_responsibility_matrix
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- Extend audit_action enum for service provider events
-- ============================================================================

DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'service_provider.created';     EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'service_provider.updated';     EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'service_provider.deleted';     EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'service_provider.reviewed';    EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'sp_compliance_doc.uploaded';   EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'sp_compliance_doc.expired';    EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'sp_responsibility.updated';    EXCEPTION WHEN duplicate_object THEN NULL; END $$;
