-- Migration: 075_aoc_roc_documents.sql
-- Description: AOC/ROC Document Generator — templates, generated documents, attestations,
--              approvals, and requirement snapshots (PCI DSS v4.0.1 Req 12.4)
-- Created: 2026-03-15
-- Sprint: 12 — AOC/ROC Document Generator (LEX-88)
-- ADR: docs/adrs/ADR-002-aoc-roc-document-generator.md

-- ============================================================================
-- TABLE 1: document_templates
-- System-level catalog of document templates (one row per AOC/ROC variant).
-- Templates define the document structure; actual HTML files are embedded Go templates.
-- This table stores metadata + configurable JSONB sections for customization.
-- ============================================================================

CREATE TABLE IF NOT EXISTS document_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID REFERENCES organizations(id) ON DELETE CASCADE,
    -- NULL org_id = system/global template; non-null = org-specific override

    template_type   TEXT NOT NULL
                        CHECK (template_type IN (
                            'aoc',      -- Attestation of Compliance (generic)
                            'roc',      -- Report on Compliance
                            'saq_a',    -- SAQ A — fully-outsourced card-not-present
                            'saq_a_ep', -- SAQ A-EP — e-commerce with redirect
                            'saq_b',    -- SAQ B — imprint machines / standalone terminals
                            'saq_b_ip', -- SAQ B-IP — IP-connected terminals
                            'saq_c_vt', -- SAQ C-VT — virtual payment terminals
                            'saq_c',    -- SAQ C — payment application systems
                            'saq_d',    -- SAQ D — all other merchants and all SPs
                            'saq_d_sp'  -- SAQ D Service Provider variant
                        )),
    name            TEXT NOT NULL,
    description     TEXT,
    pci_dss_version TEXT NOT NULL DEFAULT '4.0.1',
    version         TEXT NOT NULL DEFAULT '1.0',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,

    -- JSONB sections define the configurable structure of this template.
    -- Each section: { "key": "merchant_info", "title": "Merchant Information",
    --                 "required": true, "order": 1, "fields": [...] }
    sections        JSONB NOT NULL DEFAULT '[]',

    -- Audit
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each template_type can have one active version per (org, type, pci_dss_version)
    CONSTRAINT uq_doc_template_version UNIQUE (org_id, template_type, pci_dss_version, version)
);

COMMENT ON TABLE document_templates IS
    'Catalog of AOC/ROC document templates. System templates have org_id = NULL; '
    'org-specific overrides have org_id set. '
    'Satisfies PCI DSS v4.0.1 Req 12.4 (AOC/ROC document management).';
COMMENT ON COLUMN document_templates.sections IS
    'Ordered array of section definitions. Each section: '
    '{"key":"...", "title":"...", "required":bool, "order":int, "fields":[...]}';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_doc_templates_type       ON document_templates (template_type, is_active);
CREATE INDEX IF NOT EXISTS idx_doc_templates_org        ON document_templates (org_id, template_type)
    WHERE org_id IS NOT NULL;

-- RLS (org-specific templates; system templates with NULL org_id always visible)
ALTER TABLE document_templates ENABLE ROW LEVEL SECURITY;
CREATE POLICY doc_templates_visibility ON document_templates
    USING (
        org_id IS NULL                                                       -- system templates visible to all
        OR org_id = current_setting('app.current_org_id', TRUE)::uuid        -- org-specific templates
    );

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_doc_templates_updated_at ON document_templates;
CREATE TRIGGER trg_doc_templates_updated_at
    BEFORE UPDATE ON document_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- SEED: PCI DSS v4.0.1 AOC template structure
-- ============================================================================

INSERT INTO document_templates (id, org_id, template_type, name, description, pci_dss_version, version, sections)
VALUES (
    'e0000000-0000-0000-0001-000000000001',
    NULL,
    'saq_a',
    'PCI DSS v4.0.1 SAQ A — Attestation of Compliance',
    'Standard AOC for SAQ A merchants (card-not-present, fully-outsourced cardholder data functions). '
    'Covers merchants that only use iframes or redirects from PCI DSS-compliant TPSPs.',
    '4.0.1',
    '1.0',
    '[
        {"key": "merchant_info",        "title": "Part 1: Merchant and Qualified Security Assessor Information", "required": true, "order": 1,
         "fields": [
            {"name": "merchant_name",           "label": "Company Name", "type": "text", "required": true},
            {"name": "merchant_dba",            "label": "DBA / Brand Name", "type": "text", "required": false},
            {"name": "contact_name",            "label": "Primary Contact Name", "type": "text", "required": true},
            {"name": "contact_title",           "label": "Contact Title", "type": "text", "required": true},
            {"name": "contact_phone",           "label": "Contact Phone", "type": "text", "required": true},
            {"name": "contact_email",           "label": "Contact Email", "type": "email", "required": true},
            {"name": "merchant_url",            "label": "Merchant URL", "type": "url", "required": false},
            {"name": "business_address",        "label": "Business Address", "type": "textarea", "required": true}
         ]},
        {"key": "assessment_period",    "title": "Part 2a: Assessment Period and Scope", "required": true, "order": 2,
         "fields": [
            {"name": "period_start",            "label": "Assessment Period Start Date", "type": "date", "required": true},
            {"name": "period_end",              "label": "Assessment Period End Date", "type": "date", "required": true},
            {"name": "pci_dss_version",         "label": "PCI DSS Version", "type": "text", "required": true, "default": "4.0.1"},
            {"name": "saq_eligibility",         "label": "SAQ A Eligibility Confirmation", "type": "checkbox", "required": true,
             "description": "Confirm your company does not store, process, or transmit cardholder data"}
         ]},
        {"key": "tpsp_details",         "title": "Part 2b: Third-Party Service Provider Details", "required": true, "order": 3,
         "fields": [
            {"name": "tpsp_name",               "label": "TPSP Company Name", "type": "text", "required": true},
            {"name": "tpsp_pci_status",         "label": "TPSP PCI DSS Compliance Status", "type": "select",
             "options": ["compliant", "compliance_not_validated", "compliance_pending"], "required": true},
            {"name": "tpsp_aoc_date",           "label": "TPSP AOC Date", "type": "date", "required": false}
         ]},
        {"key": "merchant_attestation", "title": "Part 3: Merchant Attestation", "required": true, "order": 4,
         "fields": [
            {"name": "officer_name",            "label": "Officer Name", "type": "text", "required": true},
            {"name": "officer_title",           "label": "Officer Title", "type": "text", "required": true},
            {"name": "signature_date",          "label": "Signature Date", "type": "date", "required": true},
            {"name": "signature_method",        "label": "Signature Method", "type": "select",
             "options": ["manual", "digital_signature", "docusign_ref"], "required": true, "default": "manual"}
         ]},
        {"key": "acquirer_info",        "title": "Part 4: Acquirer / Payment Brand Information", "required": false, "order": 5,
         "fields": [
            {"name": "acquirer_name",           "label": "Acquirer Name", "type": "text", "required": false},
            {"name": "merchant_id",             "label": "Merchant ID", "type": "text", "required": false},
            {"name": "payment_brands",          "label": "Applicable Payment Brands", "type": "multiselect",
             "options": ["Visa", "Mastercard", "American Express", "Discover", "JCB", "UnionPay"], "required": false}
         ]}
    ]'::jsonb
)
ON CONFLICT (org_id, template_type, pci_dss_version, version) DO NOTHING;

-- SAQ D template (condensed — same structure pattern, more fields in practice)
INSERT INTO document_templates (id, org_id, template_type, name, description, pci_dss_version, version, sections)
VALUES (
    'e0000000-0000-0000-0001-000000000002',
    NULL,
    'saq_d',
    'PCI DSS v4.0.1 SAQ D — Attestation of Compliance (Merchants)',
    'Standard AOC for SAQ D merchants. All 12 PCI DSS requirements apply. '
    'For merchants that store, process, or transmit CHD electronically and do not qualify for SAQ A through C.',
    '4.0.1',
    '1.0',
    '[
        {"key": "merchant_info",        "title": "Part 1: Merchant Information", "required": true, "order": 1,
         "fields": [
            {"name": "merchant_name",           "label": "Company Name", "type": "text", "required": true},
            {"name": "merchant_dba",            "label": "DBA / Brand Name", "type": "text", "required": false},
            {"name": "contact_name",            "label": "Primary Contact Name", "type": "text", "required": true},
            {"name": "contact_title",           "label": "Contact Title", "type": "text", "required": true},
            {"name": "contact_phone",           "label": "Contact Phone", "type": "text", "required": true},
            {"name": "contact_email",           "label": "Contact Email", "type": "email", "required": true},
            {"name": "merchant_url",            "label": "Merchant URL", "type": "url", "required": false},
            {"name": "business_address",        "label": "Business Address", "type": "textarea", "required": true},
            {"name": "merchant_type",           "label": "Merchant Type", "type": "select",
             "options": ["ecommerce", "card_present", "mail_telephone", "mixed"], "required": true}
         ]},
        {"key": "assessment_period",    "title": "Part 2a: Assessment Period and PCI DSS Version", "required": true, "order": 2,
         "fields": [
            {"name": "period_start",            "label": "Assessment Period Start Date", "type": "date", "required": true},
            {"name": "period_end",              "label": "Assessment Period End Date", "type": "date", "required": true},
            {"name": "pci_dss_version",         "label": "PCI DSS Version", "type": "text", "required": true, "default": "4.0.1"},
            {"name": "saq_eligibility",         "label": "SAQ D Eligibility Confirmation", "type": "checkbox", "required": true}
         ]},
        {"key": "scope_description",    "title": "Part 2b: Business and Cardholder Data Environment Description", "required": true, "order": 3,
         "fields": [
            {"name": "chd_functions",           "label": "Cardholder Data Functions", "type": "multiselect",
             "options": ["store_pan", "process_transactions", "transmit_chd", "store_sad"], "required": true},
            {"name": "cde_description",         "label": "CDE Description", "type": "textarea", "required": true},
            {"name": "total_in_scope_systems",  "label": "Number of In-Scope Systems", "type": "number", "required": false}
         ]},
        {"key": "compensating_controls","title": "Part 2c: Compensating Controls (if applicable)", "required": false, "order": 4,
         "fields": [
            {"name": "has_compensating",        "label": "Compensating Controls Used", "type": "checkbox", "required": false},
            {"name": "compensating_summary",    "label": "Compensating Controls Summary", "type": "textarea", "required": false}
         ]},
        {"key": "customized_approach",  "title": "Part 2d: Customized Approach (if applicable)", "required": false, "order": 5,
         "fields": [
            {"name": "has_customized",          "label": "Customized Approach Used", "type": "checkbox", "required": false},
            {"name": "customized_summary",      "label": "Customized Approach Summary", "type": "textarea", "required": false}
         ]},
        {"key": "merchant_attestation", "title": "Part 3: Merchant Attestation", "required": true, "order": 6,
         "fields": [
            {"name": "officer_name",            "label": "Officer Name", "type": "text", "required": true},
            {"name": "officer_title",           "label": "Officer Title", "type": "text", "required": true},
            {"name": "signature_date",          "label": "Signature Date", "type": "date", "required": true},
            {"name": "signature_method",        "label": "Signature Method", "type": "select",
             "options": ["manual", "digital_signature", "docusign_ref"], "required": true, "default": "manual"}
         ]},
        {"key": "acquirer_info",        "title": "Part 4: Acquirer / Payment Brand Information", "required": false, "order": 7,
         "fields": [
            {"name": "acquirer_name",           "label": "Acquirer Name", "type": "text", "required": false},
            {"name": "merchant_id",             "label": "Merchant ID", "type": "text", "required": false},
            {"name": "payment_brands",          "label": "Applicable Payment Brands", "type": "multiselect",
             "options": ["Visa", "Mastercard", "American Express", "Discover", "JCB", "UnionPay"], "required": false}
         ]}
    ]'::jsonb
)
ON CONFLICT (org_id, template_type, pci_dss_version, version) DO NOTHING;

-- ROC template (structure only — sections are rendered from requirement-specific Go templates)
INSERT INTO document_templates (id, org_id, template_type, name, description, pci_dss_version, version, sections)
VALUES (
    'e0000000-0000-0000-0001-000000000003',
    NULL,
    'roc',
    'PCI DSS v4.0.1 Report on Compliance',
    'Full ROC for Level 1 merchants (over 6M transactions/year). Produced with QSA involvement. '
    'All 12 PCI DSS requirement domains with testing procedures and findings.',
    '4.0.1',
    '1.0',
    '[
        {"key": "cover_page",           "title": "Cover Page — QSA and Merchant Details", "required": true, "order": 1,
         "fields": [
            {"name": "merchant_name",           "label": "Merchant / SP Name", "type": "text", "required": true},
            {"name": "qsa_company",             "label": "QSA Company Name", "type": "text", "required": true},
            {"name": "qsa_individual",          "label": "QSA Individual Name", "type": "text", "required": true},
            {"name": "qsa_number",              "label": "QSA Certificate Number", "type": "text", "required": true},
            {"name": "assessment_period_start", "label": "Assessment Period Start", "type": "date", "required": true},
            {"name": "assessment_period_end",   "label": "Assessment Period End", "type": "date", "required": true}
         ]},
        {"key": "executive_summary",    "title": "Section 1: Executive Summary", "required": true, "order": 2, "fields": []},
        {"key": "scope",                "title": "Section 2: Scope of Assessment", "required": true, "order": 3, "fields": []},
        {"key": "req1",  "title": "Requirement 1: Install and Maintain Network Security Controls", "required": true, "order": 4,  "fields": []},
        {"key": "req2",  "title": "Requirement 2: Apply Secure Configurations",                   "required": true, "order": 5,  "fields": []},
        {"key": "req3",  "title": "Requirement 3: Protect Stored Account Data",                   "required": true, "order": 6,  "fields": []},
        {"key": "req4",  "title": "Requirement 4: Protect Cardholder Data with Strong Cryptography","required": true, "order": 7, "fields": []},
        {"key": "req5",  "title": "Requirement 5: Protect All Systems Against Malware",            "required": true, "order": 8,  "fields": []},
        {"key": "req6",  "title": "Requirement 6: Develop and Maintain Secure Systems and Software","required": true, "order": 9,  "fields": []},
        {"key": "req7",  "title": "Requirement 7: Restrict Access to System Components",           "required": true, "order": 10, "fields": []},
        {"key": "req8",  "title": "Requirement 8: Identify Users and Authenticate Access",         "required": true, "order": 11, "fields": []},
        {"key": "req9",  "title": "Requirement 9: Restrict Physical Access to Cardholder Data",    "required": true, "order": 12, "fields": []},
        {"key": "req10", "title": "Requirement 10: Log and Monitor All Access to System Components","required": true, "order": 13, "fields": []},
        {"key": "req11", "title": "Requirement 11: Test Security of Systems and Networks Regularly","required": true, "order": 14, "fields": []},
        {"key": "req12", "title": "Requirement 12: Support Information Security with Policies",    "required": true, "order": 15, "fields": []},
        {"key": "appendix_a1","title": "Appendix A1: Additional PCI DSS Requirements (Shared Hosting)", "required": false, "order": 16, "fields": []},
        {"key": "appendix_b", "title": "Appendix B: Compensating Controls", "required": false, "order": 17, "fields": []},
        {"key": "appendix_c", "title": "Appendix C: Customized Approach", "required": false, "order": 18, "fields": []}
    ]'::jsonb
)
ON CONFLICT (org_id, template_type, pci_dss_version, version) DO NOTHING;

-- ============================================================================
-- TABLE 2: compliance_documents (task AC: "generated_documents")
-- Primary record for each generated AOC/ROC compliance document.
-- ============================================================================

CREATE TABLE IF NOT EXISTS compliance_documents (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_id             UUID REFERENCES document_templates(id) ON DELETE SET NULL,

    -- Document classification
    document_type           TEXT NOT NULL
                                CHECK (document_type IN (
                                    'aoc_saq_a', 'aoc_saq_a_ep', 'aoc_saq_b', 'aoc_saq_b_ip',
                                    'aoc_saq_c_vt', 'aoc_saq_c', 'aoc_saq_d', 'aoc_saq_d_sp',
                                    'roc'
                                )),
    title                   TEXT NOT NULL,

    -- Assessment period (denormalized for document integrity)
    assessment_period_start DATE NOT NULL,
    assessment_period_end   DATE NOT NULL,
    pci_dss_version         TEXT NOT NULL DEFAULT '4.0.1',

    -- Merchant/SP details captured at generation time (NOT referenced from org table —
    -- must be stable even if org details change after document finalization)
    merchant_name           TEXT,
    merchant_dba            TEXT,
    merchant_url            TEXT,
    business_type           TEXT,

    -- QSA details (for ROC and QSA-involved SAQs)
    qsa_name                TEXT,
    qsa_company             TEXT,
    qsa_signature_date      DATE,

    -- Document lifecycle
    doc_status              TEXT NOT NULL DEFAULT 'draft'
                                CHECK (doc_status IN (
                                    'draft',        -- Being configured, not yet generated
                                    'generating',   -- PDF generation in progress
                                    'review',       -- Generated, awaiting internal approval
                                    'approved',     -- Internally approved, ready for QSA/signing
                                    'final',        -- Signed, locked, authoritative version
                                    'signed',       -- Alias for final (captures signature received)
                                    'superseded',   -- Replaced by a newer version
                                    'cancelled'
                                )),

    -- Generated by
    generated_by            UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Point-in-time compliance data snapshot (aggregated at generation time)
    data_snapshot           JSONB,  -- Full DocumentData struct serialized at generation time

    -- Storage
    pdf_path                TEXT,           -- MinIO object key / file path for generated PDF
    file_size_bytes         BIGINT,
    generated_at            TIMESTAMPTZ,
    generation_error        TEXT,           -- Set if generation failed; cleared on retry

    -- Versioning
    version                 INTEGER NOT NULL DEFAULT 1,
    parent_id               UUID REFERENCES compliance_documents(id) ON DELETE SET NULL,

    -- Audit
    created_by              UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Version must be positive
    CONSTRAINT chk_doc_version_positive CHECK (version >= 1),
    -- Final/signed documents cannot self-reference as parent
    CONSTRAINT chk_doc_no_self_parent   CHECK (parent_id IS NULL OR parent_id != id)
);

COMMENT ON TABLE compliance_documents IS
    'Primary record for each generated AOC/ROC compliance document. '
    'Satisfies PCI DSS v4.0.1 Req 12.4 (compliance document management and retention). '
    'Final/signed documents are immutable — regeneration creates a new version with parent_id set.';
COMMENT ON COLUMN compliance_documents.data_snapshot IS
    'Full serialized DocumentData captured at generation time. '
    'Ensures the document accurately reflects assessment state even as live data changes.';
COMMENT ON COLUMN compliance_documents.pdf_path IS
    'MinIO object key for the generated PDF (e.g., "orgs/{org_id}/documents/{id}/v{version}.pdf").';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_compliance_docs_org          ON compliance_documents (org_id);
CREATE INDEX IF NOT EXISTS idx_compliance_docs_org_status   ON compliance_documents (org_id, doc_status);
CREATE INDEX IF NOT EXISTS idx_compliance_docs_org_type     ON compliance_documents (org_id, document_type, doc_status);
CREATE INDEX IF NOT EXISTS idx_compliance_docs_generated_by ON compliance_documents (generated_by);
CREATE INDEX IF NOT EXISTS idx_compliance_docs_parent       ON compliance_documents (parent_id)
    WHERE parent_id IS NOT NULL;

-- RLS
ALTER TABLE compliance_documents ENABLE ROW LEVEL SECURITY;
CREATE POLICY compliance_documents_org_isolation ON compliance_documents
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_compliance_docs_updated_at ON compliance_documents;
CREATE TRIGGER trg_compliance_docs_updated_at
    BEFORE UPDATE ON compliance_documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 3: document_sections
-- Captures per-section content within a generated document, aligned with
-- the template's section definitions. Satisfies task AC requirement 3.
-- ============================================================================

CREATE TABLE IF NOT EXISTS document_sections (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id         UUID NOT NULL REFERENCES compliance_documents(id) ON DELETE CASCADE,
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    section_key         TEXT NOT NULL,      -- Matches template section "key" field
    title               TEXT NOT NULL,
    content             TEXT,               -- Rendered/entered content for this section
    compliance_status   TEXT
                            CHECK (compliance_status IS NULL OR compliance_status IN (
                                'compliant', 'non_compliant', 'partially_compliant',
                                'not_applicable', 'compensating_control', 'customized_approach'
                            )),
    evidence_ids        UUID[],             -- References to evidence_artifacts.id
    sort_order          INTEGER NOT NULL DEFAULT 0,

    -- Audit
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_doc_section_key UNIQUE (document_id, section_key)
);

COMMENT ON TABLE document_sections IS
    'Per-section content within a compliance document, keyed to template section definitions. '
    'Evidence IDs link to existing evidence artifacts for QSA traceability.';
COMMENT ON COLUMN document_sections.evidence_ids IS
    'Array of evidence_artifact UUIDs supporting compliance for this section.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_doc_sections_document     ON document_sections (document_id);
CREATE INDEX IF NOT EXISTS idx_doc_sections_org          ON document_sections (org_id);
CREATE INDEX IF NOT EXISTS idx_doc_sections_status       ON document_sections (document_id, compliance_status);

-- RLS
ALTER TABLE document_sections ENABLE ROW LEVEL SECURITY;
CREATE POLICY document_sections_org_isolation ON document_sections
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_doc_sections_updated_at ON document_sections;
CREATE TRIGGER trg_doc_sections_updated_at
    BEFORE UPDATE ON document_sections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 4: document_attestations
-- Structured attestation and signature fields for merchant/QSA signatories.
-- ============================================================================

CREATE TABLE IF NOT EXISTS document_attestations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id         UUID NOT NULL REFERENCES compliance_documents(id) ON DELETE CASCADE,
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    attestation_role    TEXT NOT NULL
                            CHECK (attestation_role IN (
                                'merchant_signatory',  -- Merchant officer signing the AOC
                                'qsa_signatory',       -- QSA signing/validating the ROC
                                'isac_signatory',      -- Internal Security Assessor (SAQ)
                                'sp_signatory'         -- Service provider representative
                            )),

    -- Signatory details
    full_name           TEXT NOT NULL,
    title               TEXT NOT NULL,
    company_name        TEXT,
    company_address     TEXT,
    company_url         TEXT,
    email               TEXT,
    phone               TEXT,

    -- QSA-specific fields
    qsa_company         TEXT,
    qsa_number          TEXT,               -- PCI SSC QSA listing number

    -- Signature
    signed_at           TIMESTAMPTZ,
    signature_method    TEXT
                            CHECK (signature_method IS NULL OR signature_method IN (
                                'manual', 'digital_signature', 'docusign_ref'
                            )),
    signature_ref       TEXT,               -- DocuSign envelope ID or certificate reference

    -- Audit
    created_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each role can only have one attestation record per document
    CONSTRAINT uq_attestation_role UNIQUE (document_id, attestation_role)
);

COMMENT ON TABLE document_attestations IS
    'Signatory information and attestation records for compliance documents. '
    'One row per (document, role). The merchant_signatory and qsa_signatory roles map '
    'to AOC Part 3 and Part 2 respectively per PCI SSC form requirements.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_doc_attestations_document    ON document_attestations (document_id);
CREATE INDEX IF NOT EXISTS idx_doc_attestations_org         ON document_attestations (org_id);

-- RLS
ALTER TABLE document_attestations ENABLE ROW LEVEL SECURITY;
CREATE POLICY document_attestations_org_isolation ON document_attestations
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_doc_attestations_updated_at ON document_attestations;
CREATE TRIGGER trg_doc_attestations_updated_at
    BEFORE UPDATE ON document_attestations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 5: document_approvals
-- Internal approval workflow before document finalization.
-- ============================================================================

CREATE TABLE IF NOT EXISTS document_approvals (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id         UUID NOT NULL REFERENCES compliance_documents(id) ON DELETE CASCADE,
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    approver_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    approval_status     TEXT NOT NULL DEFAULT 'pending'
                            CHECK (approval_status IN ('pending', 'approved', 'rejected', 'withdrawn')),
    comments            TEXT,
    responded_at        TIMESTAMPTZ,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_doc_approver UNIQUE (document_id, approver_id)
);

COMMENT ON TABLE document_approvals IS
    'Internal approval workflow for compliance documents before QSA handoff. '
    'All approvers must respond before the document transitions to approved status.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_doc_approvals_document   ON document_approvals (document_id);
CREATE INDEX IF NOT EXISTS idx_doc_approvals_approver   ON document_approvals (approver_id, approval_status);

-- RLS
ALTER TABLE document_approvals ENABLE ROW LEVEL SECURITY;
CREATE POLICY document_approvals_org_isolation ON document_approvals
    USING (org_id = current_setting('app.current_org_id', TRUE)::uuid);

-- Updated-at trigger
DROP TRIGGER IF EXISTS trg_doc_approvals_updated_at ON document_approvals;
CREATE TRIGGER trg_doc_approvals_updated_at
    BEFORE UPDATE ON document_approvals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- TABLE 6: document_requirement_snapshots
-- Point-in-time compliance posture per requirement, captured at generation time.
-- ============================================================================

CREATE TABLE IF NOT EXISTS document_requirement_snapshots (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id         UUID NOT NULL REFERENCES compliance_documents(id) ON DELETE CASCADE,
    requirement_id      UUID REFERENCES requirements(id) ON DELETE SET NULL,

    -- Denormalized for snapshot integrity (survives requirement record updates)
    requirement_code    TEXT NOT NULL,
    requirement_title   TEXT NOT NULL,

    in_scope            BOOLEAN NOT NULL,
    control_count       INTEGER NOT NULL DEFAULT 0,
    passing_controls    INTEGER NOT NULL DEFAULT 0,
    evidence_count      INTEGER NOT NULL DEFAULT 0,

    status              TEXT NOT NULL
                            CHECK (status IN (
                                'compliant', 'non_compliant', 'partially_compliant',
                                'not_applicable', 'compensating_control', 'customized_approach'
                            )),
    notes               TEXT,
    snapshotted_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE document_requirement_snapshots IS
    'Point-in-time compliance posture for each requirement, captured during document generation. '
    'This table is INSERT-only after generation — rows must never be updated to preserve '
    'document integrity across assessment cycles.';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_doc_req_snapshots_document    ON document_requirement_snapshots (document_id);
CREATE INDEX IF NOT EXISTS idx_doc_req_snapshots_requirement ON document_requirement_snapshots (requirement_id)
    WHERE requirement_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_doc_req_snapshots_status      ON document_requirement_snapshots (document_id, status);
