-- Sprint 9 -- Seed: SecurePayments Inc. — Fully Completed PCI DSS v4.0.1 Sample Organization
-- This migration creates a complete sample organization demonstrating full PCI DSS compliance
-- with all policies published, controls active, tests passing, and evidence approved.

DO $$
#variable_conflict use_column
DECLARE
    -- Organization
    org_id UUID := 'a0000000-0000-0000-0000-000000000002';

    -- Users (SecurePayments team)
    u_ciso UUID := 'b2000000-0000-0000-0000-000000000001';
    u_compliance UUID := 'b2000000-0000-0000-0000-000000000002';
    u_security UUID := 'b2000000-0000-0000-0000-000000000003';
    u_itadmin UUID := 'b2000000-0000-0000-0000-000000000004';
    u_devops UUID := 'b2000000-0000-0000-0000-000000000005';
    u_auditor UUID := 'b2000000-0000-0000-0000-000000000006';

    -- Framework references (system-level, already exist)
    fw_pci UUID := 'f0000000-0000-0000-0000-000000000003';
    fv_pci UUID := 'a1000000-0000-0000-0000-000000000003';

    -- Org framework
    of_pci UUID := 'd2000000-0000-0000-0000-000000000001';

    -- Controls (PCI DSS mapped, 24 controls covering all 12 requirements)
    ctrl_nw_001 UUID := 'c2000000-0000-0000-0000-000000000001';  -- Network Firewall Management
    ctrl_nw_002 UUID := 'c2000000-0000-0000-0000-000000000002';  -- Network Segmentation
    ctrl_nw_003 UUID := 'c2000000-0000-0000-0000-000000000003';  -- NSC Configuration Standards
    ctrl_cm_001 UUID := 'c2000000-0000-0000-0000-000000000010';  -- Secure Configuration Baselines
    ctrl_cm_002 UUID := 'c2000000-0000-0000-0000-000000000011';  -- System Hardening
    ctrl_dp_001 UUID := 'c2000000-0000-0000-0000-000000000020';  -- Cardholder Data Protection
    ctrl_dp_002 UUID := 'c2000000-0000-0000-0000-000000000021';  -- Data Encryption at Rest
    ctrl_dp_003 UUID := 'c2000000-0000-0000-0000-000000000022';  -- Encryption in Transit (TLS)
    ctrl_av_001 UUID := 'c2000000-0000-0000-0000-000000000030';  -- Anti-malware Controls
    ctrl_sd_001 UUID := 'c2000000-0000-0000-0000-000000000040';  -- Secure SDLC
    ctrl_sd_002 UUID := 'c2000000-0000-0000-0000-000000000041';  -- Code Review Process
    ctrl_sd_003 UUID := 'c2000000-0000-0000-0000-000000000042';  -- WAF Deployment
    ctrl_vm_001 UUID := 'c2000000-0000-0000-0000-000000000043';  -- Vulnerability Management
    ctrl_cc_001 UUID := 'c2000000-0000-0000-0000-000000000044';  -- Change Control
    ctrl_ac_001 UUID := 'c2000000-0000-0000-0000-000000000050';  -- Role-Based Access Control
    ctrl_ac_002 UUID := 'c2000000-0000-0000-0000-000000000051';  -- Unique User IDs
    ctrl_ac_003 UUID := 'c2000000-0000-0000-0000-000000000052';  -- MFA Enforcement
    ctrl_ac_004 UUID := 'c2000000-0000-0000-0000-000000000053';  -- Password Policy
    ctrl_ac_005 UUID := 'c2000000-0000-0000-0000-000000000054';  -- Service Account Management
    ctrl_ps_001 UUID := 'c2000000-0000-0000-0000-000000000060';  -- Physical Access Controls
    ctrl_lm_001 UUID := 'c2000000-0000-0000-0000-000000000070';  -- Centralized Logging
    ctrl_lm_002 UUID := 'c2000000-0000-0000-0000-000000000071';  -- Log Review Process
    ctrl_st_001 UUID := 'c2000000-0000-0000-0000-000000000080';  -- Vulnerability Scanning
    ctrl_st_002 UUID := 'c2000000-0000-0000-0000-000000000081';  -- Penetration Testing
    ctrl_st_003 UUID := 'c2000000-0000-0000-0000-000000000082';  -- IDS/IPS
    ctrl_st_004 UUID := 'c2000000-0000-0000-0000-000000000083';  -- File Integrity Monitoring
    ctrl_st_005 UUID := 'c2000000-0000-0000-0000-000000000084';  -- File Integrity Monitoring
    ctrl_st_006 UUID := 'c2000000-0000-0000-0000-000000000085';  -- Payment Page Monitoring
    ctrl_po_001 UUID := 'c2000000-0000-0000-0000-000000000090';  -- Information Security Policy
    ctrl_po_002 UUID := 'c2000000-0000-0000-0000-000000000091';  -- Risk Management Program
    ctrl_po_003 UUID := 'c2000000-0000-0000-0000-000000000092';  -- Security Awareness Training
    ctrl_po_004 UUID := 'c2000000-0000-0000-0000-000000000093';  -- Third-Party Management
    ctrl_ir_001 UUID := 'c2000000-0000-0000-0000-000000000094';  -- Incident Response Plan

    -- Policies (12 policies, all published)
    pol_is UUID := 'e2000000-0000-0000-0000-000000000001';   -- Information Security Policy
    pol_ac UUID := 'e2000000-0000-0000-0000-000000000002';   -- Access Control Policy
    pol_dp UUID := 'e2000000-0000-0000-0000-000000000003';   -- Data Protection Policy
    pol_nw UUID := 'e2000000-0000-0000-0000-000000000004';   -- Network Security Policy
    pol_en UUID := 'e2000000-0000-0000-0000-000000000005';   -- Encryption Policy
    pol_vm UUID := 'e2000000-0000-0000-0000-000000000006';   -- Vulnerability Management Policy
    pol_cm UUID := 'e2000000-0000-0000-0000-000000000007';   -- Change Management Policy
    pol_ir UUID := 'e2000000-0000-0000-0000-000000000008';   -- Incident Response Policy
    pol_sa UUID := 'e2000000-0000-0000-0000-000000000009';   -- Security Awareness Policy
    pol_ps UUID := 'e2000000-0000-0000-0000-000000000010';   -- Physical Security Policy
    pol_lm UUID := 'e2000000-0000-0000-0000-000000000011';   -- Logging & Monitoring Policy
    pol_tp UUID := 'e2000000-0000-0000-0000-000000000012';   -- Third-Party Management Policy

    -- Policy versions
    pv_is UUID := 'f2000000-0000-0000-0000-000000000001';
    pv_ac UUID := 'f2000000-0000-0000-0000-000000000002';
    pv_dp UUID := 'f2000000-0000-0000-0000-000000000003';
    pv_nw UUID := 'f2000000-0000-0000-0000-000000000004';
    pv_en UUID := 'f2000000-0000-0000-0000-000000000005';
    pv_vm UUID := 'f2000000-0000-0000-0000-000000000006';
    pv_cm UUID := 'f2000000-0000-0000-0000-000000000007';
    pv_ir UUID := 'f2000000-0000-0000-0000-000000000008';
    pv_sa UUID := 'f2000000-0000-0000-0000-000000000009';
    pv_ps UUID := 'f2000000-0000-0000-0000-000000000010';
    pv_lm UUID := 'f2000000-0000-0000-0000-000000000011';
    pv_tp UUID := 'f2000000-0000-0000-0000-000000000012';

    -- Tests (24 tests covering controls)
    tst_nw_001 UUID := 'a3000000-0000-0000-0000-000000000001';
    tst_nw_002 UUID := 'a3000000-0000-0000-0000-000000000002';
    tst_cm_001 UUID := 'a3000000-0000-0000-0000-000000000003';
    tst_dp_001 UUID := 'a3000000-0000-0000-0000-000000000004';
    tst_dp_002 UUID := 'a3000000-0000-0000-0000-000000000005';
    tst_av_001 UUID := 'a3000000-0000-0000-0000-000000000006';
    tst_sd_001 UUID := 'a3000000-0000-0000-0000-000000000007';
    tst_sd_002 UUID := 'a3000000-0000-0000-0000-000000000008';
    tst_vm_001 UUID := 'a3000000-0000-0000-0000-000000000009';
    tst_ac_001 UUID := 'a3000000-0000-0000-0000-000000000010';
    tst_ac_002 UUID := 'a3000000-0000-0000-0000-000000000011';
    tst_ac_003 UUID := 'a3000000-0000-0000-0000-000000000012';
    tst_ac_004 UUID := 'a3000000-0000-0000-0000-000000000013';
    tst_ps_001 UUID := 'a3000000-0000-0000-0000-000000000014';
    tst_lm_001 UUID := 'a3000000-0000-0000-0000-000000000015';
    tst_lm_002 UUID := 'a3000000-0000-0000-0000-000000000016';
    tst_st_001 UUID := 'a3000000-0000-0000-0000-000000000017';
    tst_st_002 UUID := 'a3000000-0000-0000-0000-000000000018';
    tst_st_003 UUID := 'a3000000-0000-0000-0000-000000000019';
    tst_st_004 UUID := 'a3000000-0000-0000-0000-000000000020';
    tst_cc_001 UUID := 'a3000000-0000-0000-0000-000000000021';
    tst_waf_001 UUID := 'a3000000-0000-0000-0000-000000000022';
    tst_sa_001 UUID := 'a3000000-0000-0000-0000-000000000023';
    tst_ir_001 UUID := 'a3000000-0000-0000-0000-000000000024';

    -- Test run
    run_id UUID := 'a4000000-0000-0000-0000-000000000001';

    -- bcrypt hash for "SecurePay2026!"
    pw_hash VARCHAR := '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS';

BEGIN

-- ================================================================
-- 1. ORGANIZATION
-- ================================================================
INSERT INTO organizations (id, name, slug, domain, status, settings)
VALUES (org_id, 'SecurePayments Inc.', 'securepayments', 'securepayments.example.com', 'active',
    '{"timezone":"America/Chicago","locale":"en-US","industry":"financial_services","pci_level":"1"}')
ON CONFLICT (slug) DO NOTHING;

-- ================================================================
-- 2. USERS
-- ================================================================
INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role, status, mfa_enabled, mfa_secret, last_login_at) VALUES
    (u_ciso, org_id, 'james.chen@securepayments.example.com', pw_hash, 'James', 'Chen', 'ciso', 'active', TRUE, 'JBSWY3DPEHPK3PXP0001', NOW() - INTERVAL '1 hour'),
    (u_compliance, org_id, 'sarah.martinez@securepayments.example.com', pw_hash, 'Sarah', 'Martinez', 'compliance_manager', 'active', TRUE, 'JBSWY3DPEHPK3PXP0002', NOW() - INTERVAL '2 hours'),
    (u_security, org_id, 'alex.kumar@securepayments.example.com', pw_hash, 'Alex', 'Kumar', 'security_engineer', 'active', TRUE, 'JBSWY3DPEHPK3PXP0003', NOW() - INTERVAL '3 hours'),
    (u_itadmin, org_id, 'maria.johnson@securepayments.example.com', pw_hash, 'Maria', 'Johnson', 'it_admin', 'active', TRUE, 'JBSWY3DPEHPK3PXP0004', NOW() - INTERVAL '4 hours'),
    (u_devops, org_id, 'ryan.patel@securepayments.example.com', pw_hash, 'Ryan', 'Patel', 'devops_engineer', 'active', TRUE, 'JBSWY3DPEHPK3PXP0005', NOW() - INTERVAL '5 hours'),
    (u_auditor, org_id, 'lisa.wong@securepayments.example.com', pw_hash, 'Lisa', 'Wong', 'auditor', 'active', TRUE, 'JBSWY3DPEHPK3PXP0006', NOW() - INTERVAL '6 hours')
ON CONFLICT DO NOTHING;

-- ================================================================
-- 3. ORG FRAMEWORK (PCI DSS v4.0.1 activated)
-- ================================================================
INSERT INTO org_frameworks (id, org_id, framework_id, active_version_id, status, target_date, notes)
VALUES (of_pci, org_id, fw_pci, fv_pci, 'active', '2025-03-31', 'PCI DSS Level 1 merchant — annual ROC assessment. Full compliance achieved.')
ON CONFLICT (org_id, framework_id) DO NOTHING;

-- ================================================================
-- 4. CONTROLS (32 controls covering all 12 PCI DSS requirements)
-- ================================================================
INSERT INTO controls (id, org_id, identifier, title, description, implementation_guidance, category, status, owner_id, secondary_owner_id, evidence_requirements, test_criteria, is_custom, source_template_id) VALUES
    -- Requirement 1: Network Security Controls
    (ctrl_nw_001, org_id, 'CTRL-NW-001', 'Network Firewall Management',
     'Maintain and configure firewalls to protect the cardholder data environment. All inbound/outbound traffic is filtered according to documented rulesets.',
     'Deploy next-gen firewalls at all CDE boundaries. Review rulesets quarterly. Document all allow rules with business justification.',
     'technical', 'active', u_security, u_itadmin,
     'Firewall ruleset exports, network diagrams, change logs',
     'Verify firewall rules match approved baseline; confirm no unauthorized rules exist',
     FALSE, 'TPL-NW-001'),

    (ctrl_nw_002, org_id, 'CTRL-NW-002', 'CDE Network Segmentation',
     'Implement network segmentation to isolate the cardholder data environment from other networks. Restrict traffic flows between CDE and non-CDE zones.',
     'Use VLANs and firewall zones. Verify segmentation annually via penetration testing.',
     'technical', 'active', u_security, u_itadmin,
     'Network architecture diagrams, VLAN configurations, segmentation test results',
     'Confirm CDE is isolated; verify no unauthorized paths exist between CDE and other networks',
     FALSE, 'TPL-NW-002'),

    (ctrl_nw_003, org_id, 'CTRL-NW-003', 'NSC Configuration Standards',
     'Establish and maintain configuration standards for all network security controls including firewalls, routers, and switches.',
     'Document standards for each NSC type. Include deny-all default rules, approved protocols, and hardening requirements.',
     'technical', 'active', u_security, NULL,
     'Configuration standard documents, compliance scan results',
     'Verify NSC configs match documented standards; check for deviations',
     FALSE, 'TPL-NW-003'),

    -- Requirement 2: Secure Configurations
    (ctrl_cm_001, org_id, 'CTRL-CM-001', 'Secure Configuration Baselines',
     'Define and enforce secure configuration baselines for all system components in the CDE. Remove vendor defaults.',
     'Use CIS benchmarks as baseline. Automate with configuration management tools. Scan quarterly.',
     'technical', 'active', u_devops, u_security,
     'CIS benchmark compliance reports, configuration management tool outputs',
     'Scan systems against baseline; verify no vendor defaults remain',
     FALSE, 'TPL-CM-001'),

    (ctrl_cm_002, org_id, 'CTRL-CM-002', 'System Hardening Standards',
     'Harden all system components by disabling unnecessary services, protocols, and accounts. Apply security patches within defined SLAs.',
     'Disable all unnecessary services. Remove/disable default accounts. Apply critical patches within 30 days.',
     'technical', 'active', u_devops, u_itadmin,
     'Hardening checklists, service inventories, patch compliance reports',
     'Verify only required services are running; check patch currency',
     FALSE, 'TPL-CM-002'),

    -- Requirement 3: Protect Stored Account Data
    (ctrl_dp_001, org_id, 'CTRL-DP-001', 'Cardholder Data Protection',
     'Protect stored cardholder data by minimizing data retention, masking PAN when displayed, and rendering PAN unreadable in storage.',
     'Implement tokenization for PAN storage. Mask PAN to first 6/last 4 on display. Quarterly data discovery scans.',
     'technical', 'active', u_security, u_devops,
     'Data flow diagrams, tokenization configuration, data discovery scan results',
     'Verify no cleartext PAN in databases or logs; confirm masking rules',
     FALSE, 'TPL-DP-001'),

    (ctrl_dp_002, org_id, 'CTRL-DP-002', 'Data Encryption at Rest',
     'Encrypt all cardholder data at rest using industry-accepted algorithms (AES-256). Manage encryption keys securely.',
     'Use AES-256 for all CDE databases and storage. Implement HSM for key management. Rotate keys annually.',
     'technical', 'active', u_security, NULL,
     'Encryption configuration exports, key management procedures, HSM audit logs',
     'Verify encryption is enabled on all CDE storage; confirm key rotation schedule',
     FALSE, 'TPL-DP-002'),

    -- Requirement 4: Encryption in Transit
    (ctrl_dp_003, org_id, 'CTRL-DP-003', 'Encryption in Transit (TLS)',
     'Use strong cryptography (TLS 1.2+) to protect cardholder data during transmission over open, public networks.',
     'Enforce TLS 1.2 minimum on all endpoints. Disable SSLv3/TLS 1.0/1.1. Use strong cipher suites.',
     'technical', 'active', u_devops, u_security,
     'TLS scan results, cipher suite configuration, certificate inventory',
     'SSL/TLS scan shows no weak protocols; all endpoints use TLS 1.2+',
     FALSE, 'TPL-DP-003'),

    -- Requirement 5: Anti-malware
    (ctrl_av_001, org_id, 'CTRL-AV-001', 'Anti-malware Controls',
     'Deploy anti-malware solutions on all systems commonly affected by malware. Keep signatures current and perform periodic scans.',
     'Deploy EDR on all Windows/Linux endpoints in CDE. Enable real-time scanning. Update signatures at least daily.',
     'technical', 'active', u_itadmin, u_security,
     'EDR deployment reports, signature update logs, scan results',
     'Verify EDR installed on all CDE systems; confirm signatures are current',
     FALSE, 'TPL-AV-001'),

    -- Requirement 6: Secure Development
    (ctrl_sd_001, org_id, 'CTRL-SD-001', 'Secure Software Development Lifecycle',
     'Follow a secure SDLC for all bespoke and custom software. Train developers annually on secure coding practices.',
     'Implement OWASP guidelines. Require security review for all changes. Annual secure coding training.',
     'administrative', 'active', u_devops, u_security,
     'SDLC documentation, training records, code review logs',
     'Verify SDLC includes security gates; confirm developer training completion',
     FALSE, 'TPL-SD-001'),

    (ctrl_sd_002, org_id, 'CTRL-SD-002', 'Code Review Process',
     'All bespoke and custom software changes are reviewed for security vulnerabilities before deployment to production.',
     'Mandatory peer review + automated SAST scanning. Block deployment on critical findings.',
     'technical', 'active', u_devops, u_security,
     'Code review records, SAST scan results, deployment gate logs',
     'Verify all production deployments have code review and SAST results',
     FALSE, 'TPL-SD-002'),

    (ctrl_sd_003, org_id, 'CTRL-SD-003', 'Web Application Firewall',
     'Deploy WAF to protect public-facing web applications from known attacks. Monitor and update rulesets.',
     'Deploy AWS WAF or CloudFlare in front of all payment pages. Enable OWASP Core Rule Set.',
     'technical', 'active', u_devops, u_security,
     'WAF configuration, rule sets, blocked attack logs',
     'Verify WAF is active on all public-facing payment apps; check rule currency',
     FALSE, 'TPL-SD-003'),

    (ctrl_vm_001, org_id, 'CTRL-VM-001', 'Vulnerability Management Program',
     'Identify and address security vulnerabilities through regular scanning and timely remediation per defined SLAs.',
     'Scan internal systems quarterly, external systems quarterly via ASV. Remediate critical within 30 days, high within 60.',
     'technical', 'active', u_security, u_devops,
     'Vulnerability scan reports, remediation tracking, ASV certificates',
     'Verify scan coverage; confirm remediation SLAs are met',
     FALSE, 'TPL-VM-001'),

    (ctrl_cc_001, org_id, 'CTRL-CC-001', 'Change Control Process',
     'Manage all changes to CDE systems through a formal change control process with testing, approval, and rollback procedures.',
     'Use ticketing system for all changes. Require CAB approval for significant changes. Test in staging first.',
     'administrative', 'active', u_devops, u_compliance,
     'Change tickets, approval records, test results, rollback plans',
     'Verify all production changes have tickets, approvals, and test evidence',
     FALSE, 'TPL-CC-001'),

    -- Requirement 7: Access Control
    (ctrl_ac_001, org_id, 'CTRL-AC-001', 'Role-Based Access Control',
     'Restrict access to system components and cardholder data to individuals whose job requires such access. Use RBAC.',
     'Implement least privilege. Define roles per job function. Review access quarterly.',
     'administrative', 'active', u_itadmin, u_compliance,
     'RBAC matrix, access review records, provisioning logs',
     'Verify access permissions match documented RBAC matrix; confirm quarterly reviews',
     FALSE, 'TPL-AC-001'),

    -- Requirement 8: Authentication
    (ctrl_ac_002, org_id, 'CTRL-AC-002', 'Unique User Identification',
     'Assign a unique ID to each person with computer access. Prohibit shared or group accounts.',
     'Every user has a unique ID. Generic accounts are disabled or individually trackable. Service accounts are inventoried.',
     'technical', 'active', u_itadmin, u_security,
     'User account inventories, shared account exception list, audit logs',
     'Verify no shared accounts; all actions traceable to individual users',
     FALSE, 'TPL-AC-002'),

    (ctrl_ac_003, org_id, 'CTRL-AC-003', 'Multi-Factor Authentication',
     'Implement MFA for all non-console administrative access to CDE and for all remote access.',
     'Deploy Okta MFA for all CDE admin access. Require MFA for VPN and remote access.',
     'technical', 'active', u_itadmin, u_security,
     'MFA enrollment reports, authentication logs, policy configuration',
     'Verify MFA enforced for all CDE admin access; check enrollment completion',
     FALSE, 'TPL-AC-003'),

    (ctrl_ac_004, org_id, 'CTRL-AC-004', 'Password Policy Enforcement',
     'Enforce strong password requirements: minimum 12 characters, complexity, 90-day rotation, lockout after failed attempts.',
     'Configure AD/Okta password policies. Minimum 12 chars, upper+lower+number+special. Lockout after 6 failed attempts.',
     'technical', 'active', u_itadmin, NULL,
     'Password policy configuration exports, compliance reports',
     'Verify password policy settings match requirements; check lockout configuration',
     FALSE, 'TPL-AC-004'),

    (ctrl_ac_005, org_id, 'CTRL-AC-005', 'Service Account Management',
     'Strictly manage application and system accounts used for interactive login. Document and review all service accounts.',
     'Inventory all service accounts. Restrict interactive login. Rotate credentials per policy.',
     'technical', 'active', u_devops, u_itadmin,
     'Service account inventory, credential rotation logs, access reviews',
     'Verify service accounts are inventoried; confirm rotation schedule compliance',
     FALSE, 'TPL-AC-005'),

    -- Requirement 9: Physical Security
    (ctrl_ps_001, org_id, 'CTRL-PS-001', 'Physical Access Controls',
     'Restrict physical access to cardholder data and CDE systems through badge access, CCTV, and visitor management.',
     'Deploy card reader access to data center. CCTV with 90-day retention. Visitor logs required.',
     'physical', 'active', u_itadmin, u_compliance,
     'Badge access logs, CCTV footage retention records, visitor logs',
     'Verify access controls on CDE facilities; confirm CCTV coverage and retention',
     FALSE, 'TPL-PS-001'),

    -- Requirement 10: Logging
    (ctrl_lm_001, org_id, 'CTRL-LM-001', 'Centralized Logging & SIEM',
     'Implement centralized audit logging for all CDE system components. Forward logs to SIEM for correlation and alerting.',
     'Deploy Splunk/ELK SIEM. Forward logs from all CDE systems. Retain logs for 12 months (3 months immediately available).',
     'technical', 'active', u_security, u_devops,
     'SIEM configuration, log source inventory, retention policy documentation',
     'Verify all CDE systems forward logs to SIEM; check retention periods',
     FALSE, 'TPL-LM-001'),

    (ctrl_lm_002, org_id, 'CTRL-LM-002', 'Daily Log Review Process',
     'Review security event logs at least daily to identify anomalies or suspicious activity.',
     'SOC team performs daily log review. SIEM alerts for critical events. Document review in security log.',
     'operational', 'active', u_security, u_compliance,
     'Daily review records, SIEM alert tickets, SOC shift logs',
     'Verify daily review process is followed; confirm alert response times',
     FALSE, 'TPL-LM-002'),

    -- Requirement 11: Security Testing
    (ctrl_st_001, org_id, 'CTRL-ST-001', 'Internal Vulnerability Scanning',
     'Perform internal vulnerability scans at least quarterly and after significant changes.',
     'Use Qualys/Nessus for internal scans. Scan all CDE subnets. Rescan after remediation.',
     'technical', 'active', u_security, u_devops,
     'Quarterly scan reports, remediation evidence, rescan results',
     'Verify quarterly scans performed; confirm clean rescans after remediation',
     FALSE, 'TPL-ST-001'),

    (ctrl_st_002, org_id, 'CTRL-ST-002', 'External Vulnerability Scanning (ASV)',
     'Perform external vulnerability scans at least quarterly using a PCI-approved ASV.',
     'Engage approved ASV for quarterly external scans. Address all critical/high findings before rescan.',
     'technical', 'active', u_security, NULL,
     'ASV scan certificates, remediation evidence',
     'Verify quarterly ASV scans; confirm passing scan certificates',
     FALSE, 'TPL-ST-002'),

    (ctrl_st_003, org_id, 'CTRL-ST-003', 'Penetration Testing',
     'Perform internal and external penetration testing at least annually and after significant infrastructure changes.',
     'Annual pentest by qualified third party. Test network and application layers. Retest remediated findings.',
     'technical', 'active', u_security, u_compliance,
     'Pentest reports, remediation evidence, retest results',
     'Verify annual pentest performed; confirm all findings remediated',
     FALSE, 'TPL-ST-003'),

    (ctrl_st_004, org_id, 'CTRL-ST-004', 'Intrusion Detection/Prevention',
     'Deploy IDS/IPS to detect and prevent network intrusions at CDE boundaries and critical network points.',
     'Deploy network IDS/IPS at CDE perimeter. Update signatures daily. Alert on suspicious activity.',
     'technical', 'active', u_security, u_itadmin,
     'IDS/IPS configuration, alert logs, signature update records',
     'Verify IDS/IPS deployed at CDE boundaries; check signature currency',
     FALSE, 'TPL-ST-004'),

    (ctrl_st_005, org_id, 'CTRL-ST-005', 'File Integrity Monitoring',
     'Deploy file integrity monitoring (FIM) on critical system files, configuration files, and content files.',
     'Deploy OSSEC or Tripwire FIM on all CDE servers. Monitor critical OS and application files. Alert on changes.',
     'technical', 'active', u_security, u_devops,
     'FIM configuration, alert logs, baseline comparison reports',
     'Verify FIM deployed on all CDE systems; confirm alerting is functional',
     FALSE, 'TPL-ST-005'),

    (ctrl_st_006, org_id, 'CTRL-ST-006', 'Payment Page Integrity Monitoring',
     'Detect unauthorized changes to HTTP headers and payment page content using change/tamper detection mechanisms.',
     'Deploy CSP headers and SRI on payment pages. Monitor for unauthorized script injection.',
     'technical', 'active', u_devops, u_security,
     'CSP configuration, SRI hashes, monitoring alert logs',
     'Verify CSP deployed on all payment pages; confirm no unauthorized scripts',
     FALSE, NULL),

    -- Requirement 12: Organizational Policies
    (ctrl_po_001, org_id, 'CTRL-PO-001', 'Information Security Policy Program',
     'Establish, publish, maintain, and disseminate a comprehensive information security policy reviewed annually.',
     'CISO owns policy. Annual review cycle. All employees acknowledge. Available on intranet.',
     'administrative', 'active', u_ciso, u_compliance,
     'Published policy documents, acknowledgment records, review history',
     'Verify policy is published and current; confirm annual review completed',
     FALSE, 'TPL-PO-001'),

    (ctrl_po_002, org_id, 'CTRL-PO-002', 'Risk Management Program',
     'Perform targeted risk analysis for each PCI DSS requirement. Maintain a risk register and treatment plans.',
     'Annual risk assessment covering all PCI requirements. Maintain risk register. Review quarterly.',
     'administrative', 'active', u_compliance, u_ciso,
     'Risk assessment reports, risk register, treatment plans',
     'Verify risk assessment completed; confirm all risks have treatment plans',
     FALSE, 'TPL-PO-002'),

    (ctrl_po_003, org_id, 'CTRL-PO-003', 'Security Awareness Training',
     'Implement a formal security awareness program. Train all personnel upon hire and annually. Include PCI-specific content.',
     'Annual training for all employees. Phishing simulations quarterly. Track completion in LMS.',
     'administrative', 'active', u_compliance, u_ciso,
     'Training completion records, program materials, phishing simulation results',
     'Verify all employees completed annual training; check phishing sim results',
     FALSE, 'TPL-PO-003'),

    (ctrl_po_004, org_id, 'CTRL-PO-004', 'Third-Party Service Provider Management',
     'Maintain a list of all TPSPs. Monitor compliance status. Include PCI requirements in contracts.',
     'Inventory all TPSPs. Obtain annual AOCs/SOC reports. Review contracts for PCI clauses.',
     'administrative', 'active', u_compliance, u_ciso,
     'TPSP inventory, AOC/SOC reports, contract excerpts',
     'Verify TPSP list is current; confirm compliance evidence obtained',
     FALSE, 'TPL-PO-004'),

    (ctrl_ir_001, org_id, 'CTRL-IR-001', 'Incident Response Plan',
     'Maintain an incident response plan covering detection, containment, eradication, recovery, and post-incident activities.',
     'Document IRP with roles and escalation paths. Test annually via tabletop exercise. Include card brand notification procedures.',
     'administrative', 'active', u_ciso, u_security,
     'IRP document, tabletop exercise results, incident log',
     'Verify IRP exists and is current; confirm annual testing completed',
     FALSE, 'TPL-IR-001')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ================================================================
-- 5. CONTROL MAPPINGS (controls → PCI DSS v4.0.1 requirements)
-- ================================================================
INSERT INTO control_mappings (org_id, control_id, requirement_id, strength, mapped_by, notes) VALUES
    -- Req 1: Network Security
    (org_id, ctrl_nw_001, 'e0300000-0000-0000-0000-000000000121', 'primary', u_compliance, 'Firewall restricts inbound CDE traffic'),
    (org_id, ctrl_nw_001, 'e0300000-0000-0000-0000-000000000122', 'primary', u_compliance, 'Firewall restricts outbound CDE traffic'),
    (org_id, ctrl_nw_001, 'e0300000-0000-0000-0000-000000000131', 'primary', u_compliance, 'Firewall between trusted/untrusted networks'),
    (org_id, ctrl_nw_001, 'e0300000-0000-0000-0000-000000000132', 'primary', u_compliance, 'Inbound traffic filtering'),
    (org_id, ctrl_nw_002, 'e0300000-0000-0000-0000-000000000121', 'supporting', u_compliance, 'Segmentation restricts CDE access'),
    (org_id, ctrl_nw_002, 'e0300000-0000-0000-0000-000000000151', 'primary', u_compliance, 'Endpoint security for untrusted connections'),
    (org_id, ctrl_nw_003, 'e0300000-0000-0000-0000-000000000102', 'primary', u_compliance, 'NSC policies documented'),
    (org_id, ctrl_nw_003, 'e0300000-0000-0000-0000-000000000103', 'primary', u_compliance, 'NSC roles and responsibilities'),
    (org_id, ctrl_nw_003, 'e0300000-0000-0000-0000-000000000111', 'primary', u_compliance, 'NSC configuration standards'),
    (org_id, ctrl_nw_003, 'e0300000-0000-0000-0000-000000000112', 'primary', u_compliance, 'Approved services/ports documented'),
    -- Req 6: Secure Development
    (org_id, ctrl_sd_001, 'e0300000-0000-0000-0000-000000000602', 'primary', u_compliance, 'Secure dev policies documented'),
    (org_id, ctrl_sd_001, 'e0300000-0000-0000-0000-000000000611', 'primary', u_compliance, 'Secure development practices'),
    (org_id, ctrl_sd_001, 'e0300000-0000-0000-0000-000000000612', 'primary', u_compliance, 'Developer training'),
    (org_id, ctrl_sd_002, 'e0300000-0000-0000-0000-000000000613', 'primary', u_compliance, 'Code review before release'),
    (org_id, ctrl_sd_002, 'e0300000-0000-0000-0000-000000000614', 'primary', u_compliance, 'Prevent common vulnerabilities'),
    (org_id, ctrl_sd_003, 'e0300000-0000-0000-0000-000000000631', 'primary', u_compliance, 'WAF addresses new threats'),
    (org_id, ctrl_sd_003, 'e0300000-0000-0000-0000-000000000632', 'primary', u_compliance, 'WAF detects/prevents web attacks'),
    (org_id, ctrl_sd_003, 'e0300000-0000-0000-0000-000000000633', 'supporting', u_compliance, 'Payment page script management'),
    (org_id, ctrl_vm_001, 'e0300000-0000-0000-0000-000000000621', 'primary', u_compliance, 'Vulnerability identification process'),
    (org_id, ctrl_vm_001, 'e0300000-0000-0000-0000-000000000622', 'primary', u_compliance, 'Software inventory for vuln mgmt'),
    (org_id, ctrl_vm_001, 'e0300000-0000-0000-0000-000000000623', 'primary', u_compliance, 'Patching of known vulnerabilities'),
    (org_id, ctrl_cc_001, 'e0300000-0000-0000-0000-000000000641', 'primary', u_compliance, 'Change control procedures'),
    (org_id, ctrl_cc_001, 'e0300000-0000-0000-0000-000000000642', 'primary', u_compliance, 'PCI requirements verified after changes'),
    -- Req 8: Authentication
    (org_id, ctrl_ac_002, 'e0300000-0000-0000-0000-000000000811', 'primary', u_compliance, 'Unique user IDs assigned'),
    (org_id, ctrl_ac_002, 'e0300000-0000-0000-0000-000000000812', 'primary', u_compliance, 'No shared/group accounts'),
    (org_id, ctrl_ac_002, 'e0300000-0000-0000-0000-000000000802', 'supporting', u_compliance, 'Authentication policies documented'),
    (org_id, ctrl_ac_003, 'e0300000-0000-0000-0000-000000000831', 'primary', u_compliance, 'MFA for CDE admin access'),
    (org_id, ctrl_ac_003, 'e0300000-0000-0000-0000-000000000832', 'primary', u_compliance, 'MFA for all CDE access'),
    (org_id, ctrl_ac_003, 'e0300000-0000-0000-0000-000000000833', 'primary', u_compliance, 'MFA for remote access'),
    (org_id, ctrl_ac_003, 'e0300000-0000-0000-0000-000000000841', 'primary', u_compliance, 'MFA properly configured'),
    (org_id, ctrl_ac_004, 'e0300000-0000-0000-0000-000000000821', 'supporting', u_compliance, 'Password as authentication factor'),
    (org_id, ctrl_ac_004, 'e0300000-0000-0000-0000-000000000822', 'primary', u_compliance, 'Password storage/transmission encryption'),
    (org_id, ctrl_ac_004, 'e0300000-0000-0000-0000-000000000823', 'primary', u_compliance, 'Password complexity requirements'),
    (org_id, ctrl_ac_004, 'e0300000-0000-0000-0000-000000000824', 'primary', u_compliance, 'Password rotation policy'),
    (org_id, ctrl_ac_005, 'e0300000-0000-0000-0000-000000000851', 'primary', u_compliance, 'Service account management'),
    -- Req 10: Logging
    (org_id, ctrl_lm_001, 'e0300000-0000-0000-0000-000000001002', 'primary', u_compliance, 'Logging policies documented'),
    (org_id, ctrl_lm_001, 'e0300000-0000-0000-0000-000000001011', 'primary', u_compliance, 'Audit logs enabled'),
    (org_id, ctrl_lm_001, 'e0300000-0000-0000-0000-000000001012', 'primary', u_compliance, 'Admin actions logged'),
    (org_id, ctrl_lm_001, 'e0300000-0000-0000-0000-000000001021', 'primary', u_compliance, 'Log access restricted'),
    (org_id, ctrl_lm_001, 'e0300000-0000-0000-0000-000000001022', 'primary', u_compliance, 'Logs protected from modification'),
    (org_id, ctrl_lm_001, 'e0300000-0000-0000-0000-000000001023', 'primary', u_compliance, 'Centralized log backup'),
    (org_id, ctrl_lm_002, 'e0300000-0000-0000-0000-000000001031', 'primary', u_compliance, 'Daily log review'),
    -- Req 11: Security Testing
    (org_id, ctrl_st_001, 'e0300000-0000-0000-0000-000000001102', 'supporting', u_compliance, 'Testing policies documented'),
    (org_id, ctrl_st_001, 'e0300000-0000-0000-0000-000000001111', 'primary', u_compliance, 'Internal quarterly vulnerability scans'),
    (org_id, ctrl_st_002, 'e0300000-0000-0000-0000-000000001112', 'primary', u_compliance, 'External quarterly ASV scans'),
    (org_id, ctrl_st_003, 'e0300000-0000-0000-0000-000000001121', 'primary', u_compliance, 'Annual penetration testing'),
    (org_id, ctrl_st_004, 'e0300000-0000-0000-0000-000000001131', 'primary', u_compliance, 'IDS/IPS deployed'),
    (org_id, ctrl_st_005, 'e0300000-0000-0000-0000-000000001132', 'primary', u_compliance, 'FIM deployed for change detection'),
    (org_id, ctrl_st_006, 'e0300000-0000-0000-0000-000000001141', 'primary', u_compliance, 'Payment page tamper detection'),
    -- Req 12: Policies
    (org_id, ctrl_po_001, 'e0300000-0000-0000-0000-000000001202', 'primary', u_compliance, 'Security policy established'),
    (org_id, ctrl_po_001, 'e0300000-0000-0000-0000-000000001203', 'primary', u_compliance, 'Annual policy review'),
    (org_id, ctrl_po_002, 'e0300000-0000-0000-0000-000000001211', 'primary', u_compliance, 'Targeted risk analysis'),
    (org_id, ctrl_po_003, 'e0300000-0000-0000-0000-000000001221', 'primary', u_compliance, 'Security awareness program'),
    (org_id, ctrl_po_003, 'e0300000-0000-0000-0000-000000001222', 'primary', u_compliance, 'Annual program review'),
    (org_id, ctrl_po_004, 'e0300000-0000-0000-0000-000000001231', 'primary', u_compliance, 'TPSP list maintained'),
    (org_id, ctrl_ir_001, 'e0300000-0000-0000-0000-000000001241', 'primary', u_compliance, 'Incident response plan')
ON CONFLICT (org_id, control_id, requirement_id) DO NOTHING;

-- ================================================================
-- 6. POLICIES (12 published policies covering all PCI DSS domains)
-- ================================================================
INSERT INTO policies (id, org_id, identifier, title, description, category, status, owner_id, secondary_owner_id, review_frequency_days, next_review_at, last_reviewed_at, is_template, approved_at, approved_version, published_at, tags) VALUES
    (pol_is, org_id, 'POL-IS-001', 'Information Security Policy', 'Comprehensive information security policy establishing the security framework for SecurePayments Inc. Covers scope, objectives, roles, and overarching security requirements per PCI DSS Req 12.', 'information_security', 'published', u_ciso, u_compliance, 365, '2027-01-15', '2026-01-15', FALSE, '2026-01-15', 1, '2026-01-15', ARRAY['pci-dss', 'foundational']),
    (pol_ac, org_id, 'POL-AC-001', 'Access Control Policy', 'Defines access control requirements for CDE systems including RBAC, unique IDs, MFA, password management, and privileged access per PCI DSS Req 7 & 8.', 'access_control', 'published', u_itadmin, u_compliance, 365, '2027-01-20', '2026-01-20', FALSE, '2026-01-20', 1, '2026-01-20', ARRAY['pci-dss', 'access']),
    (pol_dp, org_id, 'POL-DP-001', 'Data Protection & Privacy Policy', 'Establishes requirements for protecting stored cardholder data including retention, masking, encryption, and key management per PCI DSS Req 3.', 'data_classification', 'published', u_security, u_compliance, 365, '2027-01-25', '2026-01-25', FALSE, '2026-01-25', 1, '2026-01-25', ARRAY['pci-dss', 'data-protection']),
    (pol_nw, org_id, 'POL-NW-001', 'Network Security Policy', 'Defines network security architecture requirements including firewalls, segmentation, NSC configuration, and CDE isolation per PCI DSS Req 1.', 'network_security', 'published', u_security, u_itadmin, 365, '2027-02-01', '2026-02-01', FALSE, '2026-02-01', 1, '2026-02-01', ARRAY['pci-dss', 'network']),
    (pol_en, org_id, 'POL-EN-001', 'Encryption & Key Management Policy', 'Establishes requirements for cryptographic controls including TLS, data-at-rest encryption, key management lifecycle, and certificate management per PCI DSS Req 3 & 4.', 'encryption', 'published', u_security, u_devops, 365, '2027-02-05', '2026-02-05', FALSE, '2026-02-05', 1, '2026-02-05', ARRAY['pci-dss', 'encryption']),
    (pol_vm, org_id, 'POL-VM-001', 'Vulnerability Management Policy', 'Defines vulnerability identification, assessment, and remediation processes including scanning cadence, SLAs, and patching requirements per PCI DSS Req 6 & 11.', 'vulnerability_management', 'published', u_security, u_devops, 365, '2027-02-10', '2026-02-10', FALSE, '2026-02-10', 1, '2026-02-10', ARRAY['pci-dss', 'vulnerability']),
    (pol_cm, org_id, 'POL-CM-001', 'Change Management Policy', 'Establishes change control procedures for all CDE system modifications including approval workflows, testing, documentation, and rollback per PCI DSS Req 6.5.', 'change_management', 'published', u_devops, u_compliance, 365, '2027-02-15', '2026-02-15', FALSE, '2026-02-15', 1, '2026-02-15', ARRAY['pci-dss', 'change-control']),
    (pol_ir, org_id, 'POL-IR-001', 'Incident Response Policy', 'Defines incident detection, response, containment, eradication, recovery, and post-incident procedures including card brand notification requirements per PCI DSS Req 12.10.', 'incident_response', 'published', u_ciso, u_security, 365, '2027-02-20', '2026-02-20', FALSE, '2026-02-20', 1, '2026-02-20', ARRAY['pci-dss', 'incident-response']),
    (pol_sa, org_id, 'POL-SA-001', 'Security Awareness & Training Policy', 'Establishes security awareness program requirements including onboarding training, annual refresher, PCI-specific modules, and phishing simulations per PCI DSS Req 12.6.', 'human_resources', 'published', u_compliance, u_ciso, 365, '2027-02-25', '2026-02-25', FALSE, '2026-02-25', 1, '2026-02-25', ARRAY['pci-dss', 'training']),
    (pol_ps, org_id, 'POL-PS-001', 'Physical Security Policy', 'Defines physical access controls for facilities housing CDE systems including badge access, CCTV monitoring, visitor management, and media handling per PCI DSS Req 9.', 'physical_security', 'published', u_itadmin, u_compliance, 365, '2027-03-01', '2026-03-01', FALSE, '2026-03-01', 1, '2026-03-01', ARRAY['pci-dss', 'physical']),
    (pol_lm, org_id, 'POL-LM-001', 'Logging & Monitoring Policy', 'Establishes audit logging requirements including log generation, protection, retention, review, and SIEM correlation per PCI DSS Req 10.', 'logging_monitoring', 'published', u_security, u_compliance, 365, '2027-03-05', '2026-03-05', FALSE, '2026-03-05', 1, '2026-03-05', ARRAY['pci-dss', 'logging']),
    (pol_tp, org_id, 'POL-TP-001', 'Third-Party Service Provider Policy', 'Defines requirements for managing third-party service providers with access to cardholder data including due diligence, contracts, and ongoing monitoring per PCI DSS Req 12.8.', 'vendor_management', 'published', u_compliance, u_ciso, 365, '2027-03-10', '2026-03-10', FALSE, '2026-03-10', 1, '2026-03-10', ARRAY['pci-dss', 'third-party'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ================================================================
-- 7. POLICY VERSIONS (v1 for each policy — published)
-- ================================================================
INSERT INTO policy_versions (id, org_id, policy_id, version_number, is_current, content, content_format, content_summary, change_type, word_count, created_by) VALUES
    (pv_is, org_id, pol_is, 1, TRUE, '<h1>Information Security Policy</h1><h2>1. Purpose</h2><p>This policy establishes the information security framework for SecurePayments Inc. to protect cardholder data and maintain PCI DSS compliance.</p><h2>2. Scope</h2><p>This policy applies to all employees, contractors, and third parties with access to SecurePayments systems and data.</p><h2>3. Policy Statement</h2><p>SecurePayments is committed to protecting the confidentiality, integrity, and availability of all information assets, with particular focus on cardholder data as defined by PCI DSS.</p><h2>4. Roles and Responsibilities</h2><p>The CISO is responsible for overall security program governance. Department heads are responsible for implementing security controls within their areas.</p><h2>5. Review</h2><p>This policy is reviewed at least annually and updated as needed to reflect changes in business requirements, technology, or regulatory landscape.</p>', 'html', 'Foundational information security policy for PCI DSS compliance', 'initial', 156, u_ciso),
    (pv_ac, org_id, pol_ac, 1, TRUE, '<h1>Access Control Policy</h1><h2>1. Purpose</h2><p>Define access control requirements to restrict access to CDE systems and cardholder data based on business need-to-know.</p><h2>2. Requirements</h2><ul><li>All users must have unique IDs</li><li>Role-based access control (RBAC) must be implemented</li><li>Multi-factor authentication required for all CDE administrative access and remote access</li><li>Passwords must meet complexity requirements: minimum 12 characters, mixed case, numbers, special characters</li><li>Account lockout after 6 failed attempts for 30 minutes</li><li>Password rotation every 90 days</li><li>Quarterly access reviews for all CDE users</li></ul>', 'html', 'Access control requirements for PCI DSS Req 7 & 8', 'initial', 120, u_itadmin),
    (pv_dp, org_id, pol_dp, 1, TRUE, '<h1>Data Protection Policy</h1><h2>1. Purpose</h2><p>Establish requirements for protecting stored cardholder data throughout its lifecycle.</p><h2>2. Data Retention</h2><p>Cardholder data is only retained as long as required for business or legal purposes. Quarterly purge of expired data.</p><h2>3. Data Masking</h2><p>PAN is masked when displayed (first 6, last 4). Full PAN only available to authorized personnel with documented business need.</p><h2>4. Encryption</h2><p>All stored cardholder data is encrypted using AES-256 with keys managed via HSM.</p>', 'html', 'Data protection requirements for PCI DSS Req 3', 'initial', 98, u_security),
    (pv_nw, org_id, pol_nw, 1, TRUE, '<h1>Network Security Policy</h1><h2>1. Purpose</h2><p>Define network security architecture and controls for the cardholder data environment.</p><h2>2. Network Segmentation</h2><p>The CDE is isolated from all other networks via firewalls and VLANs.</p><h2>3. Firewall Rules</h2><p>Default deny-all. Explicit allow rules require documented business justification and quarterly review.</p><h2>4. Configuration Standards</h2><p>All NSCs follow documented configuration standards. Changes require change control approval.</p>', 'html', 'Network security requirements for PCI DSS Req 1', 'initial', 85, u_security),
    (pv_en, org_id, pol_en, 1, TRUE, '<h1>Encryption & Key Management Policy</h1><h2>1. Purpose</h2><p>Establish cryptographic standards for protecting cardholder data in transit and at rest.</p><h2>2. Transport Encryption</h2><p>TLS 1.2 minimum on all public-facing endpoints. TLS 1.3 preferred. Strong cipher suites only.</p><h2>3. Storage Encryption</h2><p>AES-256 for all CDE databases and file storage. Transparent data encryption where supported.</p><h2>4. Key Management</h2><p>Encryption keys managed via HSM. Split knowledge and dual control for key ceremonies. Annual key rotation.</p>', 'html', 'Encryption standards for PCI DSS Req 3 & 4', 'initial', 92, u_security),
    (pv_vm, org_id, pol_vm, 1, TRUE, '<h1>Vulnerability Management Policy</h1><h2>1. Purpose</h2><p>Define vulnerability identification, assessment, and remediation processes.</p><h2>2. Scanning</h2><p>Internal scans quarterly. External ASV scans quarterly. Ad hoc scans after significant changes.</p><h2>3. Remediation SLAs</h2><ul><li>Critical: 30 days</li><li>High: 60 days</li><li>Medium: 90 days</li><li>Low: Next scheduled patch window</li></ul><h2>4. Patching</h2><p>Critical security patches applied within 30 days of release. Emergency patches within 72 hours for actively exploited vulnerabilities.</p>', 'html', 'Vulnerability management for PCI DSS Req 6 & 11', 'initial', 105, u_security),
    (pv_cm, org_id, pol_cm, 1, TRUE, '<h1>Change Management Policy</h1><h2>1. Purpose</h2><p>Establish change control procedures for all CDE system modifications.</p><h2>2. Change Categories</h2><ul><li>Standard: Pre-approved, low-risk (deploy via CI/CD)</li><li>Normal: Requires CAB review and approval</li><li>Emergency: Expedited approval with post-change review</li></ul><h2>3. Requirements</h2><p>All changes require documentation, testing in staging, approval, rollback plan, and PCI compliance verification.</p>', 'html', 'Change control for PCI DSS Req 6.5', 'initial', 78, u_devops),
    (pv_ir, org_id, pol_ir, 1, TRUE, '<h1>Incident Response Policy</h1><h2>1. Purpose</h2><p>Define procedures for detecting, responding to, and recovering from security incidents.</p><h2>2. Incident Classification</h2><ul><li>P1 (Critical): Confirmed data breach, active exploitation</li><li>P2 (High): Potential data exposure, system compromise</li><li>P3 (Medium): Failed attack, policy violation</li><li>P4 (Low): Anomalous activity, informational</li></ul><h2>3. Response Procedures</h2><p>Detection → Triage → Containment → Eradication → Recovery → Post-Incident Review</p><h2>4. Notification</h2><p>Card brands notified within 24 hours of confirmed breach per PCI DSS and card brand requirements.</p>', 'html', 'Incident response for PCI DSS Req 12.10', 'initial', 118, u_ciso),
    (pv_sa, org_id, pol_sa, 1, TRUE, '<h1>Security Awareness & Training Policy</h1><h2>1. Purpose</h2><p>Establish security awareness program for all personnel.</p><h2>2. Training Requirements</h2><ul><li>New hire security orientation within first week</li><li>Annual security awareness refresher</li><li>PCI DSS-specific training for CDE personnel</li><li>Quarterly phishing simulations</li><li>Role-specific security training for developers, admins</li></ul><h2>3. Program Review</h2><p>Program reviewed annually for effectiveness and updated based on threat landscape.</p>', 'html', 'Security awareness for PCI DSS Req 12.6', 'initial', 82, u_compliance),
    (pv_ps, org_id, pol_ps, 1, TRUE, '<h1>Physical Security Policy</h1><h2>1. Purpose</h2><p>Define physical access controls for CDE facilities.</p><h2>2. Access Controls</h2><p>Badge access required for all CDE areas. Biometric for server rooms. Visitor escort required.</p><h2>3. Monitoring</h2><p>CCTV at all entry points and within CDE areas. 90-day video retention.</p><h2>4. Visitor Management</h2><p>All visitors logged, badge issued, escorted at all times, badge returned on exit.</p>', 'html', 'Physical security for PCI DSS Req 9', 'initial', 72, u_itadmin),
    (pv_lm, org_id, pol_lm, 1, TRUE, '<h1>Logging & Monitoring Policy</h1><h2>1. Purpose</h2><p>Establish audit logging requirements for CDE systems.</p><h2>2. Log Requirements</h2><p>All CDE systems must forward logs to central SIEM. Minimum log events: authentication, access to CHD, admin actions, system events.</p><h2>3. Retention</h2><p>12 months total, 3 months immediately accessible for analysis.</p><h2>4. Review</h2><p>Security events reviewed daily by SOC. Weekly summary report to CISO.</p><h2>5. Protection</h2><p>Logs are immutable. Access to log files restricted to authorized personnel.</p>', 'html', 'Logging requirements for PCI DSS Req 10', 'initial', 95, u_security),
    (pv_tp, org_id, pol_tp, 1, TRUE, '<h1>Third-Party Service Provider Policy</h1><h2>1. Purpose</h2><p>Define requirements for managing TPSPs with access to or impact on cardholder data security.</p><h2>2. Due Diligence</h2><p>Security assessment before engagement. Annual review of compliance status.</p><h2>3. Contractual Requirements</h2><p>PCI DSS compliance clause required. Right to audit. Breach notification obligations.</p><h2>4. Ongoing Monitoring</h2><p>Annual AOC/SOC report collection. Quarterly status review meetings for critical providers.</p>', 'html', 'TPSP management for PCI DSS Req 12.8', 'initial', 88, u_compliance)
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Update policies to point to their current version
UPDATE policies SET current_version_id = pv_is WHERE id = pol_is AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_ac WHERE id = pol_ac AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_dp WHERE id = pol_dp AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_nw WHERE id = pol_nw AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_en WHERE id = pol_en AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_vm WHERE id = pol_vm AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_cm WHERE id = pol_cm AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_ir WHERE id = pol_ir AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_sa WHERE id = pol_sa AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_ps WHERE id = pol_ps AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_lm WHERE id = pol_lm AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;
UPDATE policies SET current_version_id = pv_tp WHERE id = pol_tp AND org_id = 'a0000000-0000-0000-0000-000000000002'::uuid;

-- ================================================================
-- 8. POLICY SIGNOFFS (all approved by CISO and Compliance Manager)
-- ================================================================
INSERT INTO policy_signoffs (id, org_id, policy_id, policy_version_id, signer_id, signer_role, requested_by, requested_at, due_date, status, decided_at, comments) VALUES
    (gen_random_uuid(), org_id, pol_is, pv_is, u_ciso, 'ciso', u_compliance, '2026-01-10', '2026-01-15', 'approved', '2026-01-14', 'Approved. Comprehensive policy aligned with PCI DSS v4.0.1.'),
    (gen_random_uuid(), org_id, pol_is, pv_is, u_compliance, 'compliance_manager', u_ciso, '2026-01-10', '2026-01-15', 'approved', '2026-01-13', 'Reviewed and approved.'),
    (gen_random_uuid(), org_id, pol_ac, pv_ac, u_ciso, 'ciso', u_itadmin, '2026-01-15', '2026-01-20', 'approved', '2026-01-19', 'Approved. MFA and password requirements meet PCI v4.0.1.'),
    (gen_random_uuid(), org_id, pol_dp, pv_dp, u_ciso, 'ciso', u_security, '2026-01-20', '2026-01-25', 'approved', '2026-01-24', 'Approved. Tokenization approach is sound.'),
    (gen_random_uuid(), org_id, pol_nw, pv_nw, u_ciso, 'ciso', u_security, '2026-01-25', '2026-02-01', 'approved', '2026-01-30', 'Approved.'),
    (gen_random_uuid(), org_id, pol_en, pv_en, u_ciso, 'ciso', u_security, '2026-01-30', '2026-02-05', 'approved', '2026-02-04', 'Approved. HSM key management exceeds requirements.'),
    (gen_random_uuid(), org_id, pol_vm, pv_vm, u_ciso, 'ciso', u_security, '2026-02-03', '2026-02-10', 'approved', '2026-02-09', 'Approved. Remediation SLAs are appropriate.'),
    (gen_random_uuid(), org_id, pol_cm, pv_cm, u_ciso, 'ciso', u_devops, '2026-02-08', '2026-02-15', 'approved', '2026-02-14', 'Approved.'),
    (gen_random_uuid(), org_id, pol_ir, pv_ir, u_ciso, 'ciso', u_ciso, '2026-02-13', '2026-02-20', 'approved', '2026-02-19', 'Self-approved as policy owner after legal review.'),
    (gen_random_uuid(), org_id, pol_ir, pv_ir, u_compliance, 'compliance_manager', u_ciso, '2026-02-13', '2026-02-20', 'approved', '2026-02-18', 'Approved. Card brand notification timeline is compliant.'),
    (gen_random_uuid(), org_id, pol_sa, pv_sa, u_ciso, 'ciso', u_compliance, '2026-02-18', '2026-02-25', 'approved', '2026-02-24', 'Approved.'),
    (gen_random_uuid(), org_id, pol_ps, pv_ps, u_ciso, 'ciso', u_itadmin, '2026-02-23', '2026-03-01', 'approved', '2026-02-28', 'Approved.'),
    (gen_random_uuid(), org_id, pol_lm, pv_lm, u_ciso, 'ciso', u_security, '2026-02-28', '2026-03-05', 'approved', '2026-03-04', 'Approved. 12-month retention meets requirements.'),
    (gen_random_uuid(), org_id, pol_tp, pv_tp, u_ciso, 'ciso', u_compliance, '2026-03-03', '2026-03-10', 'approved', '2026-03-09', 'Approved.')
ON CONFLICT (policy_version_id, signer_id) DO NOTHING;

-- ================================================================
-- 9. POLICY-CONTROL LINKS
-- ================================================================
INSERT INTO policy_controls (id, org_id, policy_id, control_id, coverage, linked_by) VALUES
    -- Information Security Policy links to policy controls
    (gen_random_uuid(), org_id, pol_is, ctrl_po_001, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_is, ctrl_po_002, 'full', u_compliance),
    -- Access Control Policy
    (gen_random_uuid(), org_id, pol_ac, ctrl_ac_001, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_ac, ctrl_ac_002, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_ac, ctrl_ac_003, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_ac, ctrl_ac_004, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_ac, ctrl_ac_005, 'full', u_compliance),
    -- Data Protection Policy
    (gen_random_uuid(), org_id, pol_dp, ctrl_dp_001, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_dp, ctrl_dp_002, 'full', u_compliance),
    -- Network Security Policy
    (gen_random_uuid(), org_id, pol_nw, ctrl_nw_001, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_nw, ctrl_nw_002, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_nw, ctrl_nw_003, 'full', u_compliance),
    -- Encryption Policy
    (gen_random_uuid(), org_id, pol_en, ctrl_dp_002, 'partial', u_compliance),
    (gen_random_uuid(), org_id, pol_en, ctrl_dp_003, 'full', u_compliance),
    -- Vulnerability Management Policy
    (gen_random_uuid(), org_id, pol_vm, ctrl_vm_001, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_vm, ctrl_st_001, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_vm, ctrl_st_002, 'full', u_compliance),
    -- Change Management Policy
    (gen_random_uuid(), org_id, pol_cm, ctrl_cc_001, 'full', u_compliance),
    -- Incident Response Policy
    (gen_random_uuid(), org_id, pol_ir, ctrl_ir_001, 'full', u_compliance),
    -- Security Awareness Policy
    (gen_random_uuid(), org_id, pol_sa, ctrl_po_003, 'full', u_compliance),
    -- Physical Security Policy
    (gen_random_uuid(), org_id, pol_ps, ctrl_ps_001, 'full', u_compliance),
    -- Logging & Monitoring Policy
    (gen_random_uuid(), org_id, pol_lm, ctrl_lm_001, 'full', u_compliance),
    (gen_random_uuid(), org_id, pol_lm, ctrl_lm_002, 'full', u_compliance),
    -- Third-Party Policy
    (gen_random_uuid(), org_id, pol_tp, ctrl_po_004, 'full', u_compliance)
ON CONFLICT (org_id, policy_id, control_id) DO NOTHING;

-- ================================================================
-- 10. TESTS (24 active tests, all mapped to controls)
-- ================================================================
INSERT INTO tests (id, org_id, identifier, title, description, test_type, severity, status, control_id, schedule_cron, last_run_at, test_script, tags, created_by) VALUES
    (tst_nw_001, org_id, 'TST-NW-001', 'Firewall Rule Baseline Check', 'Verify firewall rules match approved baseline configuration', 'configuration', 'critical', 'active', ctrl_nw_001, '0 6 * * 1', NOW() - INTERVAL '2 days', 'check_firewall_rules()', ARRAY['pci', 'req-1', 'network'], u_security),
    (tst_nw_002, org_id, 'TST-NW-002', 'CDE Segmentation Verification', 'Verify CDE network segmentation by testing cross-segment connectivity', 'network', 'critical', 'active', ctrl_nw_002, '0 3 1 * *', NOW() - INTERVAL '5 days', 'verify_cde_segmentation()', ARRAY['pci', 'req-1', 'segmentation'], u_security),
    (tst_cm_001, org_id, 'TST-CM-001', 'CIS Benchmark Compliance Scan', 'Scan CDE systems against CIS benchmark configuration baselines', 'configuration', 'high', 'active', ctrl_cm_001, '0 2 * * 0', NOW() - INTERVAL '3 days', 'scan_cis_benchmark()', ARRAY['pci', 'req-2', 'hardening'], u_devops),
    (tst_dp_001, org_id, 'TST-DP-001', 'Cleartext PAN Detection', 'Scan databases, files, and logs for unencrypted cardholder data', 'data_protection', 'critical', 'active', ctrl_dp_001, '0 4 * * *', NOW() - INTERVAL '1 day', 'scan_cleartext_pan()', ARRAY['pci', 'req-3', 'data'], u_security),
    (tst_dp_002, org_id, 'TST-DP-002', 'TLS Configuration Scan', 'Verify TLS 1.2+ on all CDE endpoints; detect weak protocols and ciphers', 'network', 'critical', 'active', ctrl_dp_003, '0 5 * * 1', NOW() - INTERVAL '2 days', 'scan_tls_config()', ARRAY['pci', 'req-4', 'encryption'], u_security),
    (tst_av_001, org_id, 'TST-AV-001', 'EDR Deployment & Signature Check', 'Verify EDR is installed on all CDE systems with current signatures', 'endpoint', 'high', 'active', ctrl_av_001, '0 7 * * *', NOW() - INTERVAL '1 day', 'check_edr_deployment()', ARRAY['pci', 'req-5', 'malware'], u_itadmin),
    (tst_sd_001, org_id, 'TST-SD-001', 'SAST Pipeline Verification', 'Verify SAST scanning is enabled in CI/CD pipeline and blocking on critical findings', 'custom', 'high', 'active', ctrl_sd_002, '0 8 * * 1', NOW() - INTERVAL '4 days', 'check_sast_pipeline()', ARRAY['pci', 'req-6', 'development'], u_devops),
    (tst_sd_002, org_id, 'TST-SD-002', 'Code Review Compliance Check', 'Verify all production deployments have mandatory code review completion', 'custom', 'high', 'active', ctrl_sd_001, '0 9 1 * *', NOW() - INTERVAL '10 days', 'check_code_reviews()', ARRAY['pci', 'req-6', 'development'], u_devops),
    (tst_vm_001, org_id, 'TST-VM-001', 'Vulnerability Remediation SLA Check', 'Verify vulnerabilities are remediated within defined SLA timeframes', 'vulnerability', 'high', 'active', ctrl_vm_001, '0 10 * * 1', NOW() - INTERVAL '3 days', 'check_vuln_sla()', ARRAY['pci', 'req-6', 'vulnerability'], u_security),
    (tst_ac_001, org_id, 'TST-AC-001', 'RBAC Permission Audit', 'Verify access permissions match approved RBAC matrix for CDE systems', 'access_control', 'critical', 'active', ctrl_ac_001, '0 6 1 * *', NOW() - INTERVAL '7 days', 'audit_rbac_permissions()', ARRAY['pci', 'req-7', 'access'], u_itadmin),
    (tst_ac_002, org_id, 'TST-AC-002', 'Shared Account Detection', 'Detect shared or generic accounts in CDE systems', 'access_control', 'critical', 'active', ctrl_ac_002, '0 7 * * 1', NOW() - INTERVAL '4 days', 'detect_shared_accounts()', ARRAY['pci', 'req-8', 'authentication'], u_itadmin),
    (tst_ac_003, org_id, 'TST-AC-003', 'MFA Enrollment Verification', 'Verify all CDE users have MFA enrolled and active', 'access_control', 'critical', 'active', ctrl_ac_003, '0 8 * * *', NOW() - INTERVAL '1 day', 'verify_mfa_enrollment()', ARRAY['pci', 'req-8', 'mfa'], u_itadmin),
    (tst_ac_004, org_id, 'TST-AC-004', 'Password Policy Compliance', 'Verify password policy settings meet PCI requirements (length, complexity, rotation)', 'access_control', 'high', 'active', ctrl_ac_004, '0 9 * * 1', NOW() - INTERVAL '3 days', 'check_password_policy()', ARRAY['pci', 'req-8', 'passwords'], u_itadmin),
    (tst_ps_001, org_id, 'TST-PS-001', 'Physical Access Log Review', 'Verify badge access logs for CDE facilities show only authorized access', 'custom', 'medium', 'active', ctrl_ps_001, '0 10 * * 1', NOW() - INTERVAL '4 days', 'review_badge_logs()', ARRAY['pci', 'req-9', 'physical'], u_itadmin),
    (tst_lm_001, org_id, 'TST-LM-001', 'SIEM Log Source Coverage', 'Verify all CDE systems are forwarding logs to central SIEM', 'logging', 'critical', 'active', ctrl_lm_001, '0 5 * * *', NOW() - INTERVAL '1 day', 'check_siem_sources()', ARRAY['pci', 'req-10', 'logging'], u_security),
    (tst_lm_002, org_id, 'TST-LM-002', 'Log Retention Verification', 'Verify log retention meets 12-month requirement with 3-month immediate access', 'logging', 'high', 'active', ctrl_lm_001, '0 6 1 * *', NOW() - INTERVAL '15 days', 'check_log_retention()', ARRAY['pci', 'req-10', 'retention'], u_security),
    (tst_st_001, org_id, 'TST-ST-001', 'Internal Vulnerability Scan Status', 'Verify internal vulnerability scans completed within quarterly schedule', 'vulnerability', 'critical', 'active', ctrl_st_001, '0 7 1 * *', NOW() - INTERVAL '8 days', 'check_internal_scan_status()', ARRAY['pci', 'req-11', 'scanning'], u_security),
    (tst_st_002, org_id, 'TST-ST-002', 'ASV Scan Certificate Verification', 'Verify current ASV scan certificate exists and is passing', 'vulnerability', 'critical', 'active', ctrl_st_002, '0 8 1 * *', NOW() - INTERVAL '12 days', 'check_asv_certificate()', ARRAY['pci', 'req-11', 'asv'], u_security),
    (tst_st_003, org_id, 'TST-ST-003', 'IDS/IPS Signature Currency', 'Verify IDS/IPS signatures are updated within 24 hours', 'network', 'high', 'active', ctrl_st_004, '0 6 * * *', NOW() - INTERVAL '1 day', 'check_ids_signatures()', ARRAY['pci', 'req-11', 'ids'], u_security),
    (tst_st_004, org_id, 'TST-ST-004', 'FIM Alert Verification', 'Verify FIM is operational and generating alerts on file changes', 'custom', 'high', 'active', ctrl_st_005, '0 7 * * 1', NOW() - INTERVAL '3 days', 'verify_fim_alerts()', ARRAY['pci', 'req-11', 'fim'], u_security),
    (tst_cc_001, org_id, 'TST-CC-001', 'Change Control Compliance', 'Verify all recent production changes have proper change tickets and approvals', 'custom', 'high', 'active', ctrl_cc_001, '0 9 * * 1', NOW() - INTERVAL '5 days', 'check_change_tickets()', ARRAY['pci', 'req-6', 'change-control'], u_devops),
    (tst_waf_001, org_id, 'TST-WAF-001', 'WAF Rule Verification', 'Verify WAF is active on all public-facing payment applications with current rulesets', 'network', 'critical', 'active', ctrl_sd_003, '0 4 * * *', NOW() - INTERVAL '1 day', 'check_waf_status()', ARRAY['pci', 'req-6', 'waf'], u_devops),
    (tst_sa_001, org_id, 'TST-SA-001', 'Security Training Completion', 'Verify all employees have completed annual security awareness training', 'custom', 'medium', 'active', ctrl_po_003, '0 10 1 * *', NOW() - INTERVAL '20 days', 'check_training_completion()', ARRAY['pci', 'req-12', 'training'], u_compliance),
    (tst_ir_001, org_id, 'TST-IR-001', 'Incident Response Plan Currency', 'Verify incident response plan is current and tabletop exercise completed within 12 months', 'custom', 'high', 'active', ctrl_ir_001, '0 10 1 * *', NOW() - INTERVAL '25 days', 'check_irp_currency()', ARRAY['pci', 'req-12', 'incident-response'], u_ciso)
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ================================================================
-- 11. TEST RUN (all 24 tests passed)
-- ================================================================
INSERT INTO test_runs (id, org_id, run_number, status, trigger_type, started_at, completed_at, duration_ms, total_tests, passed, failed, errors, skipped, warnings, triggered_by)
VALUES (run_id, org_id, 1, 'completed', 'manual', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '15 minutes', 900000, 24, 24, 0, 0, 0, 0, u_security)
ON CONFLICT DO NOTHING;

-- ================================================================
-- 12. TEST RESULTS (all 24 pass)
-- ================================================================
INSERT INTO test_results (id, org_id, test_run_id, test_id, control_id, status, severity, message, details, started_at, completed_at, duration_ms) VALUES
    (gen_random_uuid(), org_id, run_id, tst_nw_001, ctrl_nw_001, 'pass', 'critical', 'Firewall rules match approved baseline. 0 unauthorized rules found.', '{"rules_checked":156,"unauthorized":0,"last_change":"2026-02-28"}', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '45 seconds', 45000),
    (gen_random_uuid(), org_id, run_id, tst_nw_002, ctrl_nw_002, 'pass', 'critical', 'CDE segmentation verified. No unauthorized cross-segment connectivity detected.', '{"segments_tested":8,"connections_blocked":24,"leaks":0}', NOW() - INTERVAL '1 day' + INTERVAL '1 minute', NOW() - INTERVAL '1 day' + INTERVAL '3 minutes', 120000),
    (gen_random_uuid(), org_id, run_id, tst_cm_001, ctrl_cm_001, 'pass', 'high', 'All 42 CDE systems comply with CIS benchmark. 0 deviations found.', '{"systems_scanned":42,"compliant":42,"deviations":0,"benchmark":"CIS_Level_2"}', NOW() - INTERVAL '1 day' + INTERVAL '3 minutes', NOW() - INTERVAL '1 day' + INTERVAL '5 minutes', 120000),
    (gen_random_uuid(), org_id, run_id, tst_dp_001, ctrl_dp_001, 'pass', 'critical', 'No cleartext PAN found in databases, files, or log systems.', '{"databases_scanned":8,"files_scanned":12400,"logs_scanned":1800000,"pan_found":0}', NOW() - INTERVAL '1 day' + INTERVAL '5 minutes', NOW() - INTERVAL '1 day' + INTERVAL '7 minutes', 120000),
    (gen_random_uuid(), org_id, run_id, tst_dp_002, ctrl_dp_003, 'pass', 'critical', 'All 18 CDE endpoints use TLS 1.2+. No weak protocols or ciphers detected.', '{"endpoints_scanned":18,"tls12":4,"tls13":14,"weak_ciphers":0}', NOW() - INTERVAL '1 day' + INTERVAL '7 minutes', NOW() - INTERVAL '1 day' + INTERVAL '8 minutes', 60000),
    (gen_random_uuid(), org_id, run_id, tst_av_001, ctrl_av_001, 'pass', 'high', 'EDR deployed on all 42 CDE systems. Signatures updated within 4 hours.', '{"systems_checked":42,"edr_installed":42,"signatures_current":42,"avg_sig_age_hours":3.2}', NOW() - INTERVAL '1 day' + INTERVAL '8 minutes', NOW() - INTERVAL '1 day' + INTERVAL '9 minutes', 60000),
    (gen_random_uuid(), org_id, run_id, tst_sd_001, ctrl_sd_002, 'pass', 'high', 'SAST enabled in all 6 CI/CD pipelines. Critical findings block deployment.', '{"pipelines_checked":6,"sast_enabled":6,"critical_findings_blocking":true}', NOW() - INTERVAL '1 day' + INTERVAL '9 minutes', NOW() - INTERVAL '1 day' + INTERVAL '9 minutes 30 seconds', 30000),
    (gen_random_uuid(), org_id, run_id, tst_sd_002, ctrl_sd_001, 'pass', 'high', 'All 45 production deployments (last 30 days) have completed code reviews.', '{"deployments_checked":45,"reviews_completed":45,"missing":0,"period_days":30}', NOW() - INTERVAL '1 day' + INTERVAL '10 minutes', NOW() - INTERVAL '1 day' + INTERVAL '10 minutes 30 seconds', 30000),
    (gen_random_uuid(), org_id, run_id, tst_vm_001, ctrl_vm_001, 'pass', 'high', 'All critical/high vulnerabilities remediated within SLA. 0 overdue.', '{"critical_open":0,"high_open":2,"critical_overdue":0,"high_overdue":0,"avg_remediation_days":18}', NOW() - INTERVAL '1 day' + INTERVAL '11 minutes', NOW() - INTERVAL '1 day' + INTERVAL '11 minutes 30 seconds', 30000),
    (gen_random_uuid(), org_id, run_id, tst_ac_001, ctrl_ac_001, 'pass', 'critical', 'RBAC permissions match approved matrix. Quarterly review completed 2026-02-15.', '{"users_audited":156,"role_mismatches":0,"last_review":"2026-02-15","excessive_permissions":0}', NOW() - INTERVAL '1 day' + INTERVAL '12 minutes', NOW() - INTERVAL '1 day' + INTERVAL '12 minutes 30 seconds', 30000),
    (gen_random_uuid(), org_id, run_id, tst_ac_002, ctrl_ac_002, 'pass', 'critical', 'No shared or generic accounts detected in CDE systems.', '{"accounts_checked":198,"shared_found":0,"generic_found":0,"service_accounts":12,"service_documented":12}', NOW() - INTERVAL '1 day' + INTERVAL '13 minutes', NOW() - INTERVAL '1 day' + INTERVAL '13 minutes 20 seconds', 20000),
    (gen_random_uuid(), org_id, run_id, tst_ac_003, ctrl_ac_003, 'pass', 'critical', 'All 68 CDE users have MFA enrolled and active.', '{"cde_users":68,"mfa_enrolled":68,"mfa_active":68,"enrollment_pct":100}', NOW() - INTERVAL '1 day' + INTERVAL '13 minutes 30 seconds', NOW() - INTERVAL '1 day' + INTERVAL '13 minutes 45 seconds', 15000),
    (gen_random_uuid(), org_id, run_id, tst_ac_004, ctrl_ac_004, 'pass', 'high', 'Password policy compliant: 12 char min, complexity enforced, 90-day rotation, lockout at 6 attempts.', '{"min_length":12,"complexity":true,"rotation_days":90,"lockout_attempts":6,"lockout_duration_min":30}', NOW() - INTERVAL '1 day' + INTERVAL '14 minutes', NOW() - INTERVAL '1 day' + INTERVAL '14 minutes 10 seconds', 10000),
    (gen_random_uuid(), org_id, run_id, tst_ps_001, ctrl_ps_001, 'pass', 'medium', 'Physical access logs reviewed. No unauthorized access to CDE facilities in last 30 days.', '{"entries_reviewed":1240,"unauthorized":0,"visitors_logged":45,"visitors_escorted":45}', NOW() - INTERVAL '1 day' + INTERVAL '14 minutes 15 seconds', NOW() - INTERVAL '1 day' + INTERVAL '14 minutes 25 seconds', 10000),
    (gen_random_uuid(), org_id, run_id, tst_lm_001, ctrl_lm_001, 'pass', 'critical', 'All 42 CDE systems forwarding logs to SIEM. 100% log source coverage.', '{"systems_expected":42,"sources_active":42,"coverage_pct":100,"avg_latency_ms":250}', NOW() - INTERVAL '1 day' + INTERVAL '14 minutes 30 seconds', NOW() - INTERVAL '1 day' + INTERVAL '14 minutes 45 seconds', 15000),
    (gen_random_uuid(), org_id, run_id, tst_lm_002, ctrl_lm_001, 'pass', 'high', 'Log retention verified: 12 months archived, 3 months immediately accessible.', '{"retention_months":12,"immediate_access_months":3,"oldest_accessible_log":"2025-12-06","storage_used_gb":450}', NOW() - INTERVAL '1 day' + INTERVAL '14 minutes 50 seconds', NOW() - INTERVAL '1 day' + INTERVAL '15 minutes', 10000),
    (gen_random_uuid(), org_id, run_id, tst_st_001, ctrl_st_001, 'pass', 'critical', 'Internal vulnerability scan completed 2026-02-28. All critical/high findings remediated.', '{"scan_date":"2026-02-28","critical_found":0,"high_found":3,"high_remediated":3,"medium_found":8}', NOW() - INTERVAL '23 hours 58 minutes', NOW() - INTERVAL '23 hours 57 minutes', 60000),
    (gen_random_uuid(), org_id, run_id, tst_st_002, ctrl_st_002, 'pass', 'critical', 'ASV scan certificate valid. Last scan: 2026-02-25. Status: PASS.', '{"asv_provider":"Qualys","scan_date":"2026-02-25","status":"PASS","certificate_id":"ASV-2026-Q1-SP"}', NOW() - INTERVAL '23 hours 56 minutes', NOW() - INTERVAL '23 hours 55 minutes 30 seconds', 30000),
    (gen_random_uuid(), org_id, run_id, tst_st_003, ctrl_st_004, 'pass', 'high', 'IDS/IPS signatures updated 3 hours ago. All sensors operational.', '{"sensors":6,"operational":6,"signature_age_hours":3,"alerts_24h":12,"blocked_24h":4}', NOW() - INTERVAL '23 hours 55 minutes', NOW() - INTERVAL '23 hours 54 minutes 45 seconds', 15000),
    (gen_random_uuid(), org_id, run_id, tst_st_004, ctrl_st_005, 'pass', 'high', 'FIM operational on all 42 CDE systems. 3 authorized changes detected in last 7 days.', '{"systems_monitored":42,"fim_active":42,"changes_7d":3,"changes_authorized":3,"unauthorized":0}', NOW() - INTERVAL '23 hours 54 minutes', NOW() - INTERVAL '23 hours 53 minutes 45 seconds', 15000),
    (gen_random_uuid(), org_id, run_id, tst_cc_001, ctrl_cc_001, 'pass', 'high', 'All 12 production changes in last 30 days have approved change tickets.', '{"changes_checked":12,"tickets_found":12,"approvals_found":12,"missing":0}', NOW() - INTERVAL '23 hours 53 minutes', NOW() - INTERVAL '23 hours 52 minutes 45 seconds', 15000),
    (gen_random_uuid(), org_id, run_id, tst_waf_001, ctrl_sd_003, 'pass', 'critical', 'WAF active on all 4 public payment endpoints. OWASP CRS v3.3 deployed.', '{"endpoints":4,"waf_active":4,"ruleset":"OWASP_CRS_v3.3","blocked_24h":28}', NOW() - INTERVAL '23 hours 52 minutes', NOW() - INTERVAL '23 hours 51 minutes 45 seconds', 15000),
    (gen_random_uuid(), org_id, run_id, tst_sa_001, ctrl_po_003, 'pass', 'medium', 'Annual security training completed by 98% of employees (2 new hires in grace period).', '{"total_employees":156,"completed":153,"in_grace_period":2,"overdue":1,"completion_pct":98.1}', NOW() - INTERVAL '23 hours 51 minutes', NOW() - INTERVAL '23 hours 50 minutes 45 seconds', 15000),
    (gen_random_uuid(), org_id, run_id, tst_ir_001, ctrl_ir_001, 'pass', 'high', 'IRP current (last updated 2026-02-20). Tabletop exercise completed 2026-01-15.', '{"irp_last_updated":"2026-02-20","tabletop_date":"2026-01-15","months_since_exercise":1.6,"card_brand_contacts_verified":true}', NOW() - INTERVAL '23 hours 50 minutes', NOW() - INTERVAL '23 hours 49 minutes 45 seconds', 15000)
ON CONFLICT (test_run_id, test_id) DO NOTHING;

END $$;
