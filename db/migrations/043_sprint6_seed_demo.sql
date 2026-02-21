-- Migration: 043_sprint6_seed_demo.sql
-- Description: Demo organization risk data — active risks, assessments, treatments, and control mappings (Sprint 6)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register
-- Idempotent: ON CONFLICT DO NOTHING on all inserts

-- ============================================================================
-- DEMO ACTIVE RISKS (5 risks cloned from templates for the demo org)
-- ============================================================================
INSERT INTO risks (
    id, org_id, identifier, title, description, category, status,
    owner_id, is_template,
    inherent_likelihood, inherent_impact, inherent_score,
    residual_likelihood, residual_impact, residual_score,
    risk_appetite_threshold, assessment_frequency_days, next_assessment_at, last_assessed_at,
    source, affected_assets, tags
) VALUES
    -- Critical risk: Ransomware (score 20 → 12)
    ('c2000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'RISK-CY-001', 'Ransomware Attack on Production Systems',
     'Risk of ransomware encrypting production databases and application servers, with potential double extortion.',
     'cyber_security', 'treating',
     'b0000000-0000-0000-0000-000000000002',
     FALSE,
     'likely', 'severe', 20.00,
     'possible', 'major', 12.00,
     10.00, 90, '2026-05-18', '2026-02-18',
     'threat_assessment', ARRAY['payment-api', 'customer-db', 'erp-system'],
     ARRAY['critical', 'ransomware', 'q1-review']),

    -- High risk: Phishing (score 15 → 8)
    ('c2000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'RISK-CY-002', 'Phishing / Credential Theft',
     'Risk of employees falling victim to phishing attacks, leading to compromised credentials and lateral movement.',
     'cyber_security', 'monitoring',
     'b0000000-0000-0000-0000-000000000002',
     FALSE,
     'almost_certain', 'moderate', 15.00,
     'likely', 'minor', 8.00,
     8.00, 90, '2026-05-18', '2026-02-18',
     'incident_history', ARRAY['email-system', 'vpn', 'sso'],
     ARRAY['phishing', 'training', 'ongoing']),

    -- Medium risk: PCI non-compliance (score 12 → 6)
    ('c2000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'RISK-CO-001', 'PCI DSS v4.0.1 Non-Compliance',
     'Risk of failing PCI DSS audit due to gaps in meeting v4.0.1 new requirements, particularly 6.4.3 and 11.6.1.',
     'compliance', 'assessing',
     'b0000000-0000-0000-0000-000000000001',
     FALSE,
     'possible', 'major', 12.00,
     'unlikely', 'moderate', 6.00,
     8.00, 180, '2026-08-18', '2026-02-18',
     'audit_preparation', ARRAY['payment-page', 'cardholder-data-env'],
     ARRAY['pci-dss', 'audit', 'v4.0.1']),

    -- Accepted risk: Legacy system (score 16 → 12)
    ('c2000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'RISK-TE-001', 'Legacy ERP System Dependency',
     'Critical business processes depend on the legacy ERP system (Oracle E-Business Suite 12.2) which reaches end of extended support in 2027.',
     'technology', 'accepted',
     'b0000000-0000-0000-0000-000000000004',
     FALSE,
     'likely', 'major', 16.00,
     'likely', 'moderate', 12.00,
     12.00, 365, '2027-02-15', '2026-02-15',
     'technology_review', ARRAY['erp-system', 'financial-reporting'],
     ARRAY['legacy', 'accepted', 'migration-planned']),

    -- Low risk: Vendor data handling (score 9 → 4)
    ('c2000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'RISK-TP-001', 'SaaS Vendor Data Handling',
     'Risk of SaaS vendors handling customer data without adequate security controls or DPA compliance.',
     'third_party', 'monitoring',
     'b0000000-0000-0000-0000-000000000001',
     FALSE,
     'possible', 'moderate', 9.00,
     'unlikely', 'minor', 4.00,
     6.00, 180, '2026-08-18', '2026-02-18',
     'vendor_assessment', ARRAY['crm', 'analytics-platform', 'email-service'],
     ARRAY['vendor', 'dpa', 'ongoing'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- UPDATE ACCEPTED RISK METADATA (rd..004)
-- ============================================================================
UPDATE risks SET
    accepted_at = '2026-02-15 10:00:00+00',
    accepted_by = 'b0000000-0000-0000-0000-000000000004',
    acceptance_expiry = '2027-02-15',
    acceptance_justification = 'ERP migration to SAP S/4HANA scheduled for Q3 2026. Current risk is acceptable given compensating controls (enhanced monitoring, additional backup procedures) and the planned migration timeline.'
WHERE id = 'c2000000-0000-0000-0000-000000000004'
  AND accepted_at IS NULL;

-- ============================================================================
-- RISK ASSESSMENTS (inherent + residual for each of the 5 demo risks)
-- ============================================================================
INSERT INTO risk_assessments (
    id, org_id, risk_id, assessment_type,
    likelihood, impact, likelihood_score, impact_score, overall_score,
    scoring_formula, severity, justification, data_sources,
    assessed_by, assessment_date, valid_until, is_current
) VALUES
    -- Ransomware — inherent
    ('c1000000-0000-0000-0001-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000001', 'inherent',
     'likely', 'severe', 4, 5, 20.00,
     'likelihood_x_impact', 'critical',
     'Ransomware attacks are increasing in frequency and sophistication. Our industry (fintech) is a high-value target. Recent threat intel shows active campaigns targeting similar organizations.',
     ARRAY['threat-intel-report-2026-q1', 'industry-benchmark'],
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18', '2026-05-18', TRUE),
    -- Ransomware — residual
    ('c1000000-0000-0000-0001-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000001', 'residual',
     'possible', 'major', 3, 4, 12.00,
     'likelihood_x_impact', 'high',
     'EDR, network segmentation, and immutable backups reduce likelihood. Impact reduced by backup/DR capability but still significant if critical systems are affected.',
     ARRAY['edr-deployment-report', 'backup-verification-test'],
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18', '2026-05-18', TRUE),

    -- Phishing — inherent
    ('c1000000-0000-0000-0002-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000002', 'inherent',
     'almost_certain', 'moderate', 5, 3, 15.00,
     'likelihood_x_impact', 'high',
     'Phishing attempts are near-constant (50+ blocked per month). Historical incidents show 2-3 successful compromises per year despite training.',
     ARRAY['email-gateway-stats-2025', 'incident-log-2025'],
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18', '2026-05-18', TRUE),
    -- Phishing — residual
    ('c1000000-0000-0000-0002-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000002', 'residual',
     'likely', 'minor', 4, 2, 8.00,
     'likelihood_x_impact', 'medium',
     'MFA, email filtering, and quarterly security training reduce impact. Credential theft is still possible but lateral movement is limited by MFA and network segmentation.',
     ARRAY['mfa-coverage-report', 'training-completion-q4-2025'],
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18', '2026-05-18', TRUE),

    -- PCI non-compliance — inherent
    ('c1000000-0000-0000-0003-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000003', 'inherent',
     'possible', 'major', 3, 4, 12.00,
     'likelihood_x_impact', 'high',
     'PCI DSS v4.0.1 introduces new requirements (6.4.3, 11.6.1) that are not yet fully implemented. Audit is scheduled for Q3 2026.',
     ARRAY['pci-gap-analysis-2026', 'v4.0.1-requirement-mapping'],
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-18', '2026-08-18', TRUE),
    -- PCI non-compliance — residual
    ('c1000000-0000-0000-0003-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000003', 'residual',
     'unlikely', 'moderate', 2, 3, 6.00,
     'likelihood_x_impact', 'medium',
     'Raisin Shield deployment addresses 6.4.3 and 11.6.1. Remaining gaps are documentation and process-level — implementation plan is on track.',
     ARRAY['raisin-shield-deployment', 'pci-remediation-plan'],
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-18', '2026-08-18', TRUE),

    -- Legacy ERP — inherent
    ('c1000000-0000-0000-0004-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000004', 'inherent',
     'likely', 'major', 4, 4, 16.00,
     'likelihood_x_impact', 'high',
     'ERP reaches end of extended support in 2027. Known CVEs accumulating. Vendor patch cadence declining.',
     ARRAY['vendor-eol-notice', 'cve-tracking-q1-2026'],
     'b0000000-0000-0000-0000-000000000004',
     '2026-02-15', '2027-02-15', TRUE),
    -- Legacy ERP — residual
    ('c1000000-0000-0000-0004-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000004', 'residual',
     'likely', 'moderate', 4, 3, 12.00,
     'likelihood_x_impact', 'high',
     'Compensating controls (network isolation, enhanced monitoring, additional backup) reduce impact but likelihood remains high due to aging infrastructure.',
     ARRAY['compensating-controls-review', 'network-isolation-config'],
     'b0000000-0000-0000-0000-000000000004',
     '2026-02-15', '2027-02-15', TRUE),

    -- Vendor data handling — inherent
    ('c1000000-0000-0000-0005-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000005', 'inherent',
     'possible', 'moderate', 3, 3, 9.00,
     'likelihood_x_impact', 'medium',
     'Multiple SaaS vendors handle customer data. Not all have completed SOC 2 audits. DPA compliance is inconsistent.',
     ARRAY['vendor-inventory-2026', 'dpa-tracker'],
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-18', '2026-08-18', TRUE),
    -- Vendor data handling — residual
    ('c1000000-0000-0000-0005-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000005', 'residual',
     'unlikely', 'minor', 2, 2, 4.00,
     'likelihood_x_impact', 'low',
     'Vendor assessment program, DPA requirements, and annual SOC 2 review reduce exposure. Data minimization policy limits sensitive data shared with vendors.',
     ARRAY['vendor-assessment-report-q4-2025', 'data-minimization-policy'],
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-18', '2026-08-18', TRUE)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- RISK TREATMENTS (5 treatment plans across demo risks)
-- ============================================================================
INSERT INTO risk_treatments (
    id, org_id, risk_id, treatment_type, title, description, status,
    owner_id, created_by, priority, due_date, started_at,
    expected_residual_likelihood, expected_residual_impact, expected_residual_score
) VALUES
    -- Ransomware treatment 1: EDR deployment (verified)
    ('c5000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000001', 'mitigate',
     'Deploy EDR to All Endpoints',
     'Deploy CrowdStrike Falcon EDR to 100% of endpoints with real-time threat detection and automated response.',
     'verified',
     'b0000000-0000-0000-0000-000000000002',
     'b0000000-0000-0000-0000-000000000004',
     'critical', '2026-03-15', '2026-02-01 09:00:00+00',
     'unlikely', 'major', 8.00),

    -- Ransomware treatment 2: Immutable backups (in progress)
    ('c5000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000001', 'mitigate',
     'Implement Immutable Backup Strategy',
     'Deploy immutable backups with air-gapped secondary storage. Test recovery procedures monthly.',
     'in_progress',
     'b0000000-0000-0000-0000-000000000005',
     'b0000000-0000-0000-0000-000000000002',
     'high', '2026-04-01', '2026-02-10 09:00:00+00',
     'unlikely', 'moderate', 6.00),

    -- Ransomware treatment 3: Cyber insurance (verified — transfer)
    ('c5000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000001', 'transfer',
     'Cyber Insurance Policy',
     'Maintain comprehensive cyber insurance with ransomware coverage, $5M limit.',
     'verified',
     'b0000000-0000-0000-0000-000000000004',
     'b0000000-0000-0000-0000-000000000004',
     'medium', '2026-02-15', NULL,
     NULL, NULL, NULL),

    -- Phishing treatment: Security awareness training (implemented)
    ('c5000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000002', 'mitigate',
     'Enhanced Security Awareness Training',
     'Quarterly phishing simulations with mandatory training for employees who fail. Gamified leaderboard.',
     'implemented',
     'b0000000-0000-0000-0000-000000000001',
     'b0000000-0000-0000-0000-000000000002',
     'medium', '2026-03-01', '2026-01-15 09:00:00+00',
     'likely', 'negligible', 4.00),

    -- PCI treatment: Raisin Shield deployment (in progress)
    ('c5000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000003', 'mitigate',
     'Deploy Raisin Shield for PCI DSS 6.4.3 & 11.6.1',
     'Deploy client-side script monitoring and protection to meet new PCI DSS v4.0.1 requirements.',
     'in_progress',
     'b0000000-0000-0000-0000-000000000002',
     'b0000000-0000-0000-0000-000000000001',
     'high', '2026-06-01', '2026-01-29 09:00:00+00',
     'rare', 'moderate', 3.00)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- UPDATE COMPLETED TREATMENTS WITH EFFECTIVENESS DATA
-- ============================================================================

-- EDR deployment — verified, highly effective
UPDATE risk_treatments SET
    completed_at = '2026-02-15 16:00:00+00',
    actual_effort_hours = 120.00,
    effectiveness_rating = 'highly_effective',
    effectiveness_notes = 'CrowdStrike deployed to 100% of endpoints. Blocked 3 ransomware attempts in first week. Detection rate >99%.',
    effectiveness_reviewed_at = '2026-02-20 10:00:00+00',
    effectiveness_reviewed_by = 'b0000000-0000-0000-0000-000000000004'
WHERE id = 'c5000000-0000-0000-0000-000000000001'
  AND completed_at IS NULL;

-- Cyber insurance — verified, effective
UPDATE risk_treatments SET
    completed_at = '2026-02-01 10:00:00+00',
    actual_effort_hours = 16.00,
    effectiveness_rating = 'effective',
    effectiveness_notes = 'Cyber insurance policy renewed with $5M ransomware coverage. Premium increased 15% but coverage is comprehensive.',
    effectiveness_reviewed_at = '2026-02-05 10:00:00+00',
    effectiveness_reviewed_by = 'b0000000-0000-0000-0000-000000000004'
WHERE id = 'c5000000-0000-0000-0000-000000000003'
  AND completed_at IS NULL;

-- Security awareness training — implemented, completed
UPDATE risk_treatments SET
    completed_at = '2026-02-20 14:00:00+00',
    actual_effort_hours = 40.00
WHERE id = 'c5000000-0000-0000-0000-000000000004'
  AND completed_at IS NULL;

-- ============================================================================
-- RISK-CONTROL MAPPINGS (link risks to existing controls with effectiveness)
-- ============================================================================
INSERT INTO risk_controls (
    id, org_id, risk_id, control_id, effectiveness, notes, mitigation_percentage,
    linked_by, last_effectiveness_review, reviewed_by
) VALUES
    -- Ransomware → CTRL-VM-001 (Vulnerability Scanning)
    ('c3000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000001',
     'c0000000-0000-0000-0000-000000000251',
     'effective', 'Vulnerability scanning detects exploitable weaknesses that ransomware uses for initial access', 35,
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18',
     'b0000000-0000-0000-0000-000000000002'),

    -- Ransomware → CTRL-NW-001 (Network Segmentation)
    ('c3000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000001',
     'c0000000-0000-0000-0000-000000000126',
     'partially_effective', 'Network segmentation limits lateral movement but does not prevent initial infection', 20,
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18',
     'b0000000-0000-0000-0000-000000000002'),

    -- Phishing → CTRL-AC-001 (MFA)
    ('c3000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000001',
     'effective', 'MFA prevents credential theft from resulting in account compromise', 40,
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18',
     'b0000000-0000-0000-0000-000000000002'),

    -- Phishing → CTRL-SA-001 (Security Awareness Training)
    ('c3000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000206',
     'partially_effective', 'Security awareness training reduces phishing click rate but does not eliminate it', 25,
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-18',
     'b0000000-0000-0000-0000-000000000001'),

    -- PCI non-compliance → CTRL-SD-001 (Secure SDLC)
    ('c3000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000003',
     'c0000000-0000-0000-0000-000000000226',
     'not_assessed', 'Raisin Shield deployment in progress — will address 6.4.3 script monitoring via secure SDLC', NULL,
     'b0000000-0000-0000-0000-000000000001',
     NULL, NULL),

    -- PCI non-compliance → CTRL-PM-001 (Information Security Policy)
    ('c3000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000003',
     'c0000000-0000-0000-0000-000000000146',
     'partially_effective', 'Overarching security policy covers PCI scope but specific v4.0.1 requirements still need policy updates', 15,
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-18',
     'b0000000-0000-0000-0000-000000000001'),

    -- Vendor data handling → CTRL-DP-001 (Data Classification)
    ('c3000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001',
     'c2000000-0000-0000-0000-000000000005',
     'c0000000-0000-0000-0000-000000000056',
     'effective', 'Data classification policy ensures vendors only receive data classified as Internal or lower', 30,
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-18',
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (org_id, risk_id, control_id) DO NOTHING;

-- ============================================================================
-- AUDIT LOG ENTRIES FOR RISK OPERATIONS
-- ============================================================================
INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'risk.created', 'risk', 'c2000000-0000-0000-0000-000000000001',
     '{"identifier": "RISK-CY-001", "title": "Ransomware Attack on Production Systems", "category": "cyber_security"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'risk.created', 'risk', 'c2000000-0000-0000-0000-000000000002',
     '{"identifier": "RISK-CY-002", "title": "Phishing / Credential Theft", "category": "cyber_security"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'risk.created', 'risk', 'c2000000-0000-0000-0000-000000000003',
     '{"identifier": "RISK-CO-001", "title": "PCI DSS v4.0.1 Non-Compliance", "category": "compliance"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'risk.status_changed', 'risk', 'c2000000-0000-0000-0000-000000000004',
     '{"identifier": "RISK-TE-001", "old_status": "assessing", "new_status": "accepted", "justification": "ERP migration planned Q3 2026"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'risk_assessment.created', 'risk_assessment', 'c1000000-0000-0000-0001-000000000001',
     '{"risk_identifier": "RISK-CY-001", "type": "inherent", "score": 20.0, "severity": "critical"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'risk_treatment.created', 'risk_treatment', 'c5000000-0000-0000-0000-000000000001',
     '{"risk_identifier": "RISK-CY-001", "treatment": "Deploy EDR to All Endpoints", "type": "mitigate"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'risk_treatment.completed', 'risk_treatment', 'c5000000-0000-0000-0000-000000000001',
     '{"risk_identifier": "RISK-CY-001", "treatment": "Deploy EDR to All Endpoints", "effectiveness": "highly_effective"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'risk_control.linked', 'risk_control', 'c3000000-0000-0000-0000-000000000001',
     '{"risk_identifier": "RISK-CY-001", "control_identifier": "CTRL-VM-001", "effectiveness": "effective"}'::jsonb,
     '192.168.1.10'::inet)
ON CONFLICT DO NOTHING;
