-- Migration: 053_pci_quick_wins.sql
-- Description: Quick-win PCI DSS schema additions that unblock downstream sprint work
-- Created: 2026-03-15
-- Sprint: 11 — PCI Quick Wins (LEX-80)
--
-- Changes:
--   1. controls          — add is_compensating + compensating_worksheet (PCI DSS v4.0 §4.3.1)
--   2. requirement_scopes — add customized_approach flag (PCI DSS v4.0 Customized Approach option)
--   3. audit_type enum   — add pci_dss_saq_a, pci_dss_saq_d, pci_dss_aoc values
--   4. audit_request_templates — seed SAQ A, SAQ D, and AOC request templates
--   5. integration_definitions — add FIM (File Integrity Monitoring) webhook integration

-- ============================================================================
-- 1. controls — Compensating control support (PCI DSS v4.0 Appendix B)
-- ============================================================================

ALTER TABLE controls
    ADD COLUMN IF NOT EXISTS is_compensating        BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS compensating_worksheet JSONB;

COMMENT ON COLUMN controls.is_compensating IS
    'TRUE if this is a compensating control per PCI DSS v4.0 Appendix B. '
    'Requires compensating_worksheet to be populated.';
COMMENT ON COLUMN controls.compensating_worksheet IS
    'Structured compensating control worksheet per PCI DSS v4.0 Appendix B. '
    'Expected keys: constraints, objective, identified_risk, definition, validation, maintenance.';

-- Partial index — only compensating controls have a worksheet worth indexing
CREATE INDEX IF NOT EXISTS idx_controls_is_compensating
    ON controls (org_id, is_compensating)
    WHERE is_compensating = TRUE;

-- ============================================================================
-- 2. requirement_scopes — Customized Approach flag (PCI DSS v4.0 §6.6)
-- ============================================================================

ALTER TABLE requirement_scopes
    ADD COLUMN IF NOT EXISTS customized_approach BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN requirement_scopes.customized_approach IS
    'TRUE if the org uses the PCI DSS v4.0 Customized Approach for this requirement '
    'instead of the Defined Approach. Requires additional documentation at the API layer.';

-- Partial index — surface all customized-approach scoping decisions efficiently
CREATE INDEX IF NOT EXISTS idx_requirement_scopes_customized
    ON requirement_scopes (org_id, customized_approach)
    WHERE customized_approach = TRUE;

-- ============================================================================
-- 3. audit_type enum — add SAQ A, SAQ D, and AOC variants
-- ============================================================================

-- pci_dss_saq_a  : SAQ A — card-not-present merchants that fully outsource all CHD functions
-- pci_dss_saq_d  : SAQ D — all other merchants and all service providers
-- pci_dss_aoc    : Attestation of Compliance — the signed compliance declaration document
DO $$ BEGIN ALTER TYPE audit_type ADD VALUE IF NOT EXISTS 'pci_dss_saq_a'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_type ADD VALUE IF NOT EXISTS 'pci_dss_saq_d'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_type ADD VALUE IF NOT EXISTS 'pci_dss_aoc';   EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ============================================================================
-- 4. audit_request_templates — PCI DSS SAQ A, SAQ D, and AOC templates
-- ============================================================================

-- SAQ A (10 templates — simplified scope for fully-outsourced e-commerce)
INSERT INTO audit_request_templates (id, framework, audit_type, reference_number, title, description, priority, tags) VALUES
    ('a6000000-0000-0000-0003-000000000001', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-001',
     'Third-Party Service Provider (TPSP) PCI DSS Compliance Confirmation',
     'Provide current AOC or equivalent compliance documentation from every TPSP that stores, processes, or transmits cardholder data on your behalf.',
     'critical', ARRAY['req12.8', 'tpsp', 'outsourcing']),
    ('a6000000-0000-0000-0003-000000000002', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-002',
     'Payment Page Iframes / Redirect Configuration Evidence',
     'Provide documentation showing that your payment pages only redirect to or embed pages from PCI DSS-compliant TPSPs. Include code samples or architecture diagram.',
     'critical', ARRAY['req6.4', 'payment-page', 'iframe']),
    ('a6000000-0000-0000-0003-000000000003', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-003',
     'Cardholder Data Scope Confirmation',
     'Provide written confirmation that your systems do not store, process, or transmit cardholder data. Include network diagram showing data flow handled exclusively by TPSP.',
     'critical', ARRAY['req3', 'req4', 'scope']),
    ('a6000000-0000-0000-0003-000000000004', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-004',
     'Information Security Policy',
     'Provide your information security policy covering acceptable use, access control, incident response, and responsibilities. Must be reviewed and approved within the last year.',
     'high', ARRAY['req12.1', 'policy']),
    ('a6000000-0000-0000-0003-000000000005', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-005',
     'Security Awareness Training Records',
     'Provide records showing that all personnel handling payment operations completed security awareness training within the last year.',
     'high', ARRAY['req12.6', 'training']),
    ('a6000000-0000-0000-0003-000000000006', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-006',
     'Incident Response Plan',
     'Provide your documented incident response plan for cardholder data breaches, including contact lists, escalation procedures, and evidence of annual review.',
     'high', ARRAY['req12.10', 'incident-response']),
    ('a6000000-0000-0000-0003-000000000007', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-007',
     'System Component Inventory (in-scope systems)',
     'Provide inventory of all systems in scope for SAQ A, including e-commerce servers and any internal systems that interact with the payment page.',
     'medium', ARRAY['req12.5', 'inventory']),
    ('a6000000-0000-0000-0003-000000000008', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-008',
     'User Account and Access Control Evidence',
     'Provide evidence that access to in-scope systems follows least privilege, unique IDs are used, and access is reviewed periodically.',
     'high', ARRAY['req7', 'req8', 'access-control']),
    ('a6000000-0000-0000-0003-000000000009', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-009',
     'Vulnerability Management — Patch Records (Last 6 Months)',
     'Provide patch management records for in-scope systems demonstrating critical patches applied within 30 days of release.',
     'high', ARRAY['req6', 'patching']),
    ('a6000000-0000-0000-0003-000000000010', 'pci_dss', 'pci_dss_saq_a', 'SAQ-A-010',
     'Sub-merchant / Acquirer Agreement Confirming TPSP Responsibility',
     'Provide your agreement or attestation with your acquirer/payment brand confirming which PCI DSS controls are delegated to your TPSP.',
     'critical', ARRAY['req12.9', 'tpsp', 'acquirer'])
ON CONFLICT (framework, audit_type, reference_number) DO NOTHING;

-- SAQ D (15 templates — most comprehensive, covers all merchants not in A/B/C categories)
INSERT INTO audit_request_templates (id, framework, audit_type, reference_number, title, description, priority, tags) VALUES
    ('a6000000-0000-0000-0004-000000000001', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-001',
     'Network Segmentation Architecture and Validation',
     'Provide network diagrams, segmentation architecture documentation, and results from the most recent segmentation penetration test.',
     'critical', ARRAY['req1', 'network', 'segmentation']),
    ('a6000000-0000-0000-0004-000000000002', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-002',
     'Cardholder Data Discovery Scan Results',
     'Provide results from the most recent cardholder data discovery scan across all in-scope systems. Include remediation of any unexpected finds.',
     'critical', ARRAY['req3', 'data-discovery']),
    ('a6000000-0000-0000-0004-000000000003', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-003',
     'Encryption Key Management Procedures and Evidence',
     'Provide key management procedures and evidence of implementation including generation, distribution, storage, rotation, and destruction.',
     'critical', ARRAY['req3', 'req4', 'encryption']),
    ('a6000000-0000-0000-0004-000000000004', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-004',
     'Anti-Malware and FIM Deployment Evidence',
     'Provide evidence of anti-malware deployment and FIM (file integrity monitoring) configuration on all CDE systems.',
     'high', ARRAY['req5', 'req10.3', 'malware', 'fim']),
    ('a6000000-0000-0000-0004-000000000005', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-005',
     'Firewall and ACL Rule Review (Last 6 Months)',
     'Provide documented firewall/router rule review completed within the last 6 months. Include rule set exports and reviewer sign-off.',
     'high', ARRAY['req1', 'firewall', 'rule-review']),
    ('a6000000-0000-0000-0004-000000000006', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-006',
     'MFA Configuration Evidence for All CDE Access',
     'Provide evidence of MFA deployment for all non-console administrative access to CDE systems and all remote network access.',
     'critical', ARRAY['req8', 'mfa', 'access-control']),
    ('a6000000-0000-0000-0004-000000000007', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-007',
     'Log Collection, Retention, and Review Evidence (Req 10)',
     'Provide audit log configuration showing all required events are captured, log retention policy (12 months / 3 months online), and evidence of daily log review.',
     'high', ARRAY['req10', 'logging', 'monitoring']),
    ('a6000000-0000-0000-0004-000000000008', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-008',
     'Quarterly Internal and External Vulnerability Scans',
     'Provide results from the last four quarters of internal and external vulnerability scans. All high/critical findings must show remediation.',
     'critical', ARRAY['req11.3', 'vulnerability-scan']),
    ('a6000000-0000-0000-0004-000000000009', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-009',
     'Annual Penetration Test Report',
     'Provide the most recent annual penetration test report for network and application layers, including remediation status of all findings.',
     'critical', ARRAY['req11.4', 'pentest']),
    ('a6000000-0000-0000-0004-000000000010', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-010',
     'Physical Access Controls — Visitor Logs and Badge Access Records',
     'Provide physical access logs for CDE facilities for the past 90 days, including visitor logs and badge access records.',
     'medium', ARRAY['req9', 'physical-access']),
    ('a6000000-0000-0000-0004-000000000011', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-011',
     'Vendor / TPSP Compliance Tracking Records (Req 12.8)',
     'Provide your third-party service provider register with current PCI DSS compliance status (AOC/SoC) for each entity with access to CHD.',
     'critical', ARRAY['req12.8', 'tpsp', 'vendor']),
    ('a6000000-0000-0000-0004-000000000012', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-012',
     'Risk Assessment Results (Last Annual Cycle)',
     'Provide results from the most recent annual targeted risk assessment including scope, methodology, and risk register updates.',
     'high', ARRAY['req12.3', 'risk-assessment']),
    ('a6000000-0000-0000-0004-000000000013', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-013',
     'Change Management Evidence for CDE Systems (Last 90 Days)',
     'Provide change management records for all changes to CDE systems in the last 90 days, including approval, testing, and back-out procedures.',
     'medium', ARRAY['req6', 'change-management']),
    ('a6000000-0000-0000-0004-000000000014', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-014',
     'PAN Storage — Data-at-Rest Encryption Evidence',
     'Provide evidence that all stored PANs are rendered unreadable (tokenization, truncation, hashing, or strong cryptography). Include configuration and key inventory.',
     'critical', ARRAY['req3.5', 'pan-storage', 'encryption']),
    ('a6000000-0000-0000-0004-000000000015', 'pci_dss', 'pci_dss_saq_d', 'SAQ-D-015',
     'Payment Application Security Scan / Review Evidence',
     'Provide most recent PA-DSS/P2PE validation certificate or internal application security review for custom payment applications.',
     'high', ARRAY['req6', 'application-security', 'pa-dss'])
ON CONFLICT (framework, audit_type, reference_number) DO NOTHING;

-- AOC (6 templates — Attestation of Compliance document preparation)
INSERT INTO audit_request_templates (id, framework, audit_type, reference_number, title, description, priority, tags) VALUES
    ('a6000000-0000-0000-0005-000000000001', 'pci_dss', 'pci_dss_aoc', 'AOC-001',
     'Organization and Contact Information',
     'Provide legal entity name, DBA names, registered address, primary contact for PCI DSS compliance, and all applicable payment brands.',
     'critical', ARRAY['aoc-part1', 'org-info']),
    ('a6000000-0000-0000-0005-000000000002', 'pci_dss', 'pci_dss_aoc', 'AOC-002',
     'Description of Payment Card Business — Merchant/Service Provider Questionnaire',
     'Provide description of payment acceptance channels, transaction volumes, cardholder data storage/processing/transmission scope.',
     'critical', ARRAY['aoc-part2', 'business-description']),
    ('a6000000-0000-0000-0005-000000000003', 'pci_dss', 'pci_dss_aoc', 'AOC-003',
     'Listing of Facilities and Locations in Scope',
     'Provide complete list of all facilities, data centers, and hosted environments in scope for the PCI DSS assessment.',
     'high', ARRAY['aoc-part2', 'scope', 'facilities']),
    ('a6000000-0000-0000-0005-000000000004', 'pci_dss', 'pci_dss_aoc', 'AOC-004',
     'Third-Party Service Providers List with Compliance Status',
     'Provide list of all TPSPs with indication of which PCI DSS requirements each TPSP manages on your behalf and their current compliance status.',
     'critical', ARRAY['aoc-part2g', 'tpsp', 'req12.8']),
    ('a6000000-0000-0000-0005-000000000005', 'pci_dss', 'pci_dss_aoc', 'AOC-005',
     'Executive Officer Attestation — Signature Ready Draft',
     'Prepare the attestation section with completed merchant/service provider declaration for executive officer signature. Include QSA contact information if applicable.',
     'critical', ARRAY['aoc-part4', 'attestation', 'signature']),
    ('a6000000-0000-0000-0005-000000000006', 'pci_dss', 'pci_dss_aoc', 'AOC-006',
     'QSA/ISA Qualification Details (if applicable)',
     'Provide QSA company name, individual assessor name, QSA qualification certificate, and engagement letter. Required for ROC and optional for SAQ D.',
     'medium', ARRAY['aoc-part3', 'qsa', 'assessor'])
ON CONFLICT (framework, audit_type, reference_number) DO NOTHING;

-- ============================================================================
-- 5. integration_definitions — FIM (File Integrity Monitoring) webhook
-- ============================================================================
-- FIM tools (Tripwire, Wazuh/OSSEC, AIDE, Qualys FIM) send webhook events when
-- critical files change. This integration receives those events and auto-creates
-- evidence artifacts or alerts. PCI DSS v4.0 Req 10.3.4 / 11.5.2.

INSERT INTO integration_definitions (
    id, name, slug, provider, category,
    description, short_description, icon_url,
    auth_type, config_schema, capabilities,
    is_active, is_beta, version, tags
)
VALUES (
    'd0000000-0000-0000-0000-000000000006',
    'File Integrity Monitoring (FIM) Webhook',
    'fim-webhook',
    'fim_webhook',
    'monitoring',
    'Receive file integrity monitoring events from any FIM tool (Tripwire, Wazuh/OSSEC, AIDE, Qualys FIM) via webhook. '
    'Automatically creates evidence artifacts on critical file change events and triggers compliance alerts for PCI DSS Req 10.3.4 / 11.5.2.',
    'FIM event ingest for PCI DSS compliance',
    '/icons/integrations/fim.svg',
    'webhook_secret',
    '{
        "type": "object",
        "required": ["fim_tool"],
        "properties": {
            "fim_tool": {
                "type": "string",
                "title": "FIM Tool",
                "description": "The FIM tool sending events to this webhook",
                "enum": ["tripwire", "wazuh", "ossec", "aide", "qualys_fim", "custom"],
                "enumNames": ["Tripwire Enterprise", "Wazuh", "OSSEC", "AIDE", "Qualys FIM", "Custom / Other"]
            },
            "monitored_paths": {
                "type": "array",
                "title": "Monitored Paths (informational)",
                "description": "List of filesystem paths being monitored — for documentation purposes only",
                "items": {"type": "string"}
            },
            "auto_create_evidence": {
                "type": "boolean",
                "title": "Auto-Create Evidence Artifacts",
                "description": "Automatically create an evidence artifact for each FIM change event received",
                "default": true
            },
            "alert_on_critical": {
                "type": "boolean",
                "title": "Alert on Critical File Changes",
                "description": "Trigger a compliance alert when changes are detected in critical paths (e.g. /etc, /bin, /sbin)",
                "default": true
            },
            "pci_req_mapping": {
                "type": "string",
                "title": "PCI DSS Requirement",
                "description": "PCI DSS requirement this integration satisfies",
                "default": "10.3.4",
                "enum": ["10.3.4", "11.5.2", "both"]
            }
        }
    }',
    ARRAY['webhook_inbound', 'evidence_collection', 'compliance_check', 'fim'],
    true,
    false,
    '1.0.0',
    ARRAY['pci-dss', 'fim', 'file-integrity', 'monitoring', 'req10', 'req11']
)
ON CONFLICT (slug) DO NOTHING;
