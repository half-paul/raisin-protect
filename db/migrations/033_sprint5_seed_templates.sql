-- Migration: 033_sprint5_seed_templates.sql
-- Description: Policy template seed data (templates per framework: SOC 2, ISO 27001, PCI DSS, GDPR, CCPA)
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management
--
-- Templates are org-scoped policies with is_template = TRUE.
-- Organizations clone these to jump-start their policy library.

-- ============================================================================
-- SOC 2 POLICY TEMPLATES
-- ============================================================================

INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('b4000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'TPL-IS-001', 'Information Security Policy',
     'Comprehensive information security policy establishing the organization''s commitment to protecting information assets. Covers scope, objectives, roles, and responsibilities.',
     'information_security', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'pci', 'template', 'mandatory']),

    ('b4000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'TPL-AC-001', 'Access Control Policy',
     'Defines requirements for user access management, authentication, authorization, and periodic access reviews.',
     'access_control', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'pci', 'template']),

    ('b4000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'TPL-IR-001', 'Incident Response Policy',
     'Establishes procedures for detecting, reporting, assessing, responding to, and learning from security incidents.',
     'incident_response', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'pci', 'template']),

    ('b4000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'TPL-CM-001', 'Change Management Policy',
     'Defines the process for requesting, reviewing, approving, implementing, and documenting changes to IT systems and infrastructure.',
     'change_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'template']),

    ('b4000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'TPL-BC-001', 'Business Continuity & Disaster Recovery Policy',
     'Establishes the framework for maintaining business operations during and recovering from disruptive events.',
     'business_continuity', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000001', 365,
     ARRAY['soc2', 'iso27001', 'template'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- PCI DSS POLICY TEMPLATES
-- ============================================================================

INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('b4000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000001',
     'TPL-NW-001', 'Network Security Policy',
     'Defines requirements for network segmentation, firewall configuration, and monitoring of network traffic to protect cardholder data environments.',
     'network_security', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000003', 365,
     ARRAY['pci', 'template', 'network']),

    ('b4000000-0000-0000-0000-000000000011', 'a0000000-0000-0000-0000-000000000001',
     'TPL-EN-001', 'Encryption & Key Management Policy',
     'Establishes standards for cryptographic protection of sensitive data at rest and in transit, and procedures for cryptographic key lifecycle management.',
     'encryption', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000003', 365,
     ARRAY['pci', 'template', 'encryption']),

    ('b4000000-0000-0000-0000-000000000012', 'a0000000-0000-0000-0000-000000000001',
     'TPL-VM-001', 'Vulnerability Management Policy',
     'Defines the process for identifying, assessing, prioritizing, and remediating vulnerabilities in systems and applications.',
     'vulnerability_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000003', 180,
     ARRAY['pci', 'template', 'vulnerability'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- ISO 27001 POLICY TEMPLATES
-- ============================================================================

INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('b4000000-0000-0000-0000-000000000020', 'a0000000-0000-0000-0000-000000000001',
     'TPL-AM-001', 'Asset Management Policy',
     'Defines requirements for identifying, classifying, and managing information assets throughout their lifecycle.',
     'asset_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000002', 365,
     ARRAY['iso27001', 'template', 'asset']),

    ('b4000000-0000-0000-0000-000000000021', 'a0000000-0000-0000-0000-000000000001',
     'TPL-RM-001', 'Risk Management Policy',
     'Establishes the organization''s approach to identifying, assessing, treating, and monitoring information security risks.',
     'risk_management', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000002', 365,
     ARRAY['iso27001', 'template', 'risk'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- GDPR POLICY TEMPLATES
-- ============================================================================

INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('b4000000-0000-0000-0000-000000000030', 'a0000000-0000-0000-0000-000000000001',
     'TPL-DP-001', 'Data Privacy Policy',
     'Establishes the organization''s approach to processing personal data in compliance with GDPR. Covers lawful basis, data subject rights, cross-border transfers, and DPIAs.',
     'data_privacy', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000004', 365,
     ARRAY['gdpr', 'template', 'privacy']),

    ('b4000000-0000-0000-0000-000000000031', 'a0000000-0000-0000-0000-000000000001',
     'TPL-DR-001', 'Data Retention & Disposal Policy',
     'Defines retention periods for different data categories and secure disposal procedures for data that has exceeded its retention period.',
     'data_retention', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000004', 365,
     ARRAY['gdpr', 'ccpa', 'template', 'retention'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- CCPA POLICY TEMPLATES
-- ============================================================================

INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    is_template, template_framework_id, review_frequency_days, tags
) VALUES
    ('b4000000-0000-0000-0000-000000000040', 'a0000000-0000-0000-0000-000000000001',
     'TPL-CP-001', 'Consumer Privacy Rights Policy',
     'Defines procedures for handling consumer privacy rights requests under CCPA/CPRA: right to know, delete, opt-out of sale, and correct.',
     'data_privacy', 'published', TRUE,
     'f0000000-0000-0000-0000-000000000005', 365,
     ARRAY['ccpa', 'template', 'privacy', 'consumer-rights'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- TEMPLATE VERSION CONTENT
-- ============================================================================

-- Information Security Policy template content
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000001', 1, TRUE,
     '<h1>Information Security Policy</h1>
<h2>1. Purpose</h2>
<p>This policy establishes the organization''s commitment to protecting information assets from threats, whether internal or external, deliberate or accidental. It provides the framework for setting objectives and establishing the overall direction and principles for information security.</p>

<h2>2. Scope</h2>
<p>This policy applies to all employees, contractors, consultants, temporaries, and other workers at [Organization Name], including all personnel affiliated with third parties who access organizational information systems.</p>

<h2>3. Policy Statement</h2>
<p>The organization shall:</p>
<ul>
<li>Protect information from unauthorized access, disclosure, modification, or destruction</li>
<li>Ensure business continuity and minimize business damage</li>
<li>Comply with all applicable laws, regulations, and contractual obligations</li>
<li>Provide security awareness training to all personnel</li>
<li>Report and investigate all suspected security incidents</li>
<li>Regularly review and improve the information security management system</li>
</ul>

<h2>4. Roles and Responsibilities</h2>
<h3>4.1 CISO / Security Leader</h3>
<p>Responsible for developing, implementing, and maintaining the information security program. Reports to executive management on the state of information security.</p>

<h3>4.2 IT Administrators</h3>
<p>Responsible for implementing technical security controls, monitoring systems, and responding to security alerts.</p>

<h3>4.3 All Employees</h3>
<p>Responsible for complying with security policies, reporting security incidents, and completing required security awareness training.</p>

<h2>5. Review</h2>
<p>This policy shall be reviewed at least annually or when significant changes occur to the organization, its business processes, or the threat landscape.</p>

<h2>6. Compliance</h2>
<p>Violation of this policy may result in disciplinary action up to and including termination of employment and/or legal action.</p>',
     'html',
     'Comprehensive information security policy template covering purpose, scope, policy statement, roles and responsibilities, review cadence, and compliance.',
     'Initial version',
     'initial',
     280,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Access Control Policy template content
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000002', 1, TRUE,
     '<h1>Access Control Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines the requirements for managing user access to information systems, applications, and data to ensure that access is granted based on the principle of least privilege and business need-to-know.</p>

<h2>2. Scope</h2>
<p>This policy applies to all user accounts, service accounts, and access credentials used to access organizational systems and data.</p>

<h2>3. Access Control Requirements</h2>
<h3>3.1 Authentication</h3>
<ul>
<li>Multi-factor authentication (MFA) shall be required for all user accounts accessing production systems</li>
<li>Passwords shall meet minimum complexity requirements: 12+ characters, mixed case, numbers, and special characters</li>
<li>Service accounts shall use certificate-based or token-based authentication where possible</li>
</ul>

<h3>3.2 Authorization</h3>
<ul>
<li>Access shall be granted based on role-based access control (RBAC)</li>
<li>Users shall receive only the minimum permissions required to perform their job functions</li>
<li>Privileged access shall require additional approval and be time-limited where possible</li>
</ul>

<h3>3.3 Access Reviews</h3>
<ul>
<li>Access reviews shall be conducted quarterly for all critical systems</li>
<li>Reviews shall verify that access levels are appropriate for current job responsibilities</li>
<li>Stale accounts (no login for 90+ days) shall be disabled pending review</li>
</ul>

<h3>3.4 Account Lifecycle</h3>
<ul>
<li>Access shall be provisioned within 24 hours of approved request</li>
<li>Access shall be revoked within 4 hours of employment termination</li>
<li>Role changes shall trigger access review within 5 business days</li>
</ul>

<h2>4. Review</h2>
<p>This policy shall be reviewed annually or when significant changes occur to access control technology or organizational structure.</p>',
     'html',
     'Access control policy template covering authentication, authorization, access reviews, and account lifecycle management.',
     'Initial version',
     'initial',
     250,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Incident Response Policy template content
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000003', 1, TRUE,
     '<h1>Incident Response Policy</h1>
<h2>1. Purpose</h2>
<p>This policy establishes procedures for detecting, reporting, assessing, responding to, and learning from security incidents to minimize damage and recovery time.</p>

<h2>2. Scope</h2>
<p>This policy applies to all information security incidents affecting organizational systems, data, or personnel.</p>

<h2>3. Incident Classification</h2>
<ul>
<li><strong>Critical:</strong> Active data breach, ransomware, system-wide compromise</li>
<li><strong>High:</strong> Unauthorized access to sensitive data, malware on production systems</li>
<li><strong>Medium:</strong> Phishing success (credentials compromised), policy violations</li>
<li><strong>Low:</strong> Suspicious activity, failed attack attempts, minor policy deviations</li>
</ul>

<h2>4. Response Phases</h2>
<h3>4.1 Detection & Reporting</h3>
<p>All personnel must report suspected incidents within 1 hour of discovery.</p>

<h3>4.2 Triage & Assessment</h3>
<p>Security team assesses severity, scope, and potential impact within 2 hours.</p>

<h3>4.3 Containment</h3>
<p>Immediate actions to limit incident scope. Document all containment decisions.</p>

<h3>4.4 Eradication & Recovery</h3>
<p>Remove threat, restore systems, verify integrity before returning to service.</p>

<h3>4.5 Post-Incident Review</h3>
<p>Conduct review within 5 business days. Document lessons learned and update procedures.</p>

<h2>5. Communication</h2>
<p>Regulatory notifications as required (72h GDPR, state breach laws). Customer notifications per contractual obligations.</p>',
     'html',
     'Incident response policy template covering classification, 5-phase response, and communication requirements.',
     'Initial version',
     'initial',
     220,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Change Management Policy template content
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000004', 1, TRUE,
     '<h1>Change Management Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines the process for requesting, reviewing, approving, implementing, and documenting changes to IT systems and infrastructure to minimize service disruption.</p>

<h2>2. Change Categories</h2>
<ul>
<li><strong>Standard:</strong> Pre-approved, low-risk changes (e.g., routine patches)</li>
<li><strong>Normal:</strong> Require CAB review and approval</li>
<li><strong>Emergency:</strong> Critical fixes — expedited approval with post-implementation review</li>
</ul>

<h2>3. Change Process</h2>
<ol>
<li>Submit change request with description, risk assessment, and rollback plan</li>
<li>Review by Change Advisory Board (CAB) for normal/major changes</li>
<li>Approval from designated authority</li>
<li>Implementation during approved maintenance window</li>
<li>Post-implementation verification and documentation</li>
</ol>

<h2>4. Documentation Requirements</h2>
<p>All changes must be documented with: requester, approver, implementation date, description, test results, and rollback status.</p>',
     'html',
     'Change management policy template covering change categories, CAB process, and documentation requirements.',
     'Initial version',
     'initial',
     170,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Business Continuity Policy template content
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000005', 1, TRUE,
     '<h1>Business Continuity &amp; Disaster Recovery Policy</h1>
<h2>1. Purpose</h2>
<p>This policy establishes the framework for maintaining business operations during and recovering from disruptive events including natural disasters, cyber attacks, and infrastructure failures.</p>

<h2>2. Recovery Objectives</h2>
<ul>
<li><strong>RTO (Recovery Time Objective):</strong> Maximum acceptable downtime per system tier</li>
<li><strong>RPO (Recovery Point Objective):</strong> Maximum acceptable data loss per system tier</li>
</ul>

<h2>3. Plan Requirements</h2>
<ul>
<li>Annual business impact analysis (BIA)</li>
<li>Documented recovery procedures for all critical systems</li>
<li>Offsite backup storage with geographic diversity</li>
<li>Semi-annual DR testing (tabletop + technical)</li>
<li>Communication plan for stakeholders during incidents</li>
</ul>

<h2>4. Testing</h2>
<p>DR plans must be tested at least semi-annually. Test results must be documented with identified gaps and remediation timelines.</p>',
     'html',
     'Business continuity and disaster recovery policy template covering RTO/RPO, plan requirements, and testing cadence.',
     'Initial version',
     'initial',
     160,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Network Security Policy template content (PCI DSS)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000010', 1, TRUE,
     '<h1>Network Security Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines requirements for network segmentation, firewall configuration, and monitoring to protect cardholder data environments and critical infrastructure.</p>

<h2>2. Network Segmentation</h2>
<ul>
<li>Cardholder data environment (CDE) must be isolated via network segmentation</li>
<li>Default deny all inbound and outbound traffic; allow only documented business needs</li>
<li>DMZ for public-facing services with no direct connectivity to internal network</li>
</ul>

<h2>3. Firewall Requirements</h2>
<ul>
<li>Review firewall rules semi-annually</li>
<li>Document business justification for every allowed rule</li>
<li>Deny all traffic not explicitly permitted</li>
<li>Log all denied connections for monitoring</li>
</ul>

<h2>4. Monitoring</h2>
<p>Network traffic must be monitored for anomalies. IDS/IPS deployed at CDE boundaries. All security events retained for minimum 12 months.</p>',
     'html',
     'Network security policy for PCI DSS CDE protection: segmentation, firewall rules, and monitoring.',
     'Initial version',
     'initial',
     165,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Encryption & Key Management Policy template content (PCI DSS)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000011', 1, TRUE,
     '<h1>Encryption &amp; Key Management Policy</h1>
<h2>1. Purpose</h2>
<p>This policy establishes standards for cryptographic protection of sensitive data and procedures for key lifecycle management.</p>

<h2>2. Data at Rest</h2>
<ul>
<li>All sensitive data (PII, CHD, credentials) encrypted using AES-256 or equivalent</li>
<li>Database-level TDE or application-level encryption as appropriate</li>
<li>Encryption keys stored separately from encrypted data</li>
</ul>

<h2>3. Data in Transit</h2>
<ul>
<li>TLS 1.2+ required for all external communications</li>
<li>Internal service-to-service communication encrypted via mTLS where feasible</li>
<li>Legacy protocols (SSLv3, TLS 1.0/1.1) prohibited</li>
</ul>

<h2>4. Key Management</h2>
<ul>
<li>Key generation using CSPRNG with sufficient entropy</li>
<li>Key rotation: annually for data keys, quarterly for high-risk keys</li>
<li>Dual control for master key operations</li>
<li>Secure key destruction documented and audited</li>
</ul>',
     'html',
     'Encryption and key management policy covering data at rest/in transit and key lifecycle.',
     'Initial version',
     'initial',
     175,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Vulnerability Management Policy template content (PCI DSS)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000012', 1, TRUE,
     '<h1>Vulnerability Management Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines the process for identifying, assessing, prioritizing, and remediating vulnerabilities in systems and applications.</p>

<h2>2. Scanning Requirements</h2>
<ul>
<li>Internal vulnerability scans: at least quarterly and after significant changes</li>
<li>External vulnerability scans: quarterly by PCI ASV</li>
<li>Penetration testing: annually and after significant infrastructure changes</li>
</ul>

<h2>3. Remediation SLAs</h2>
<ul>
<li><strong>Critical (CVSS 9.0+):</strong> 24 hours</li>
<li><strong>High (CVSS 7.0-8.9):</strong> 7 days</li>
<li><strong>Medium (CVSS 4.0-6.9):</strong> 30 days</li>
<li><strong>Low (CVSS 0.1-3.9):</strong> 90 days</li>
</ul>

<h2>4. Exception Process</h2>
<p>Vulnerabilities that cannot be remediated within SLA require a documented risk acceptance with compensating controls, approved by CISO.</p>',
     'html',
     'Vulnerability management policy covering scanning cadence, remediation SLAs, and exception handling.',
     'Initial version',
     'initial',
     155,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Asset Management Policy template content (ISO 27001)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000020', 1, TRUE,
     '<h1>Asset Management Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines requirements for identifying, classifying, and managing information assets throughout their lifecycle.</p>

<h2>2. Asset Inventory</h2>
<ul>
<li>Maintain a comprehensive inventory of all information assets (hardware, software, data, services)</li>
<li>Each asset must have an assigned owner responsible for its security</li>
<li>Inventory reviewed quarterly; reconciled with discovery tools</li>
</ul>

<h2>3. Asset Classification</h2>
<ul>
<li><strong>Public:</strong> No restrictions on access or disclosure</li>
<li><strong>Internal:</strong> For internal use only; no external distribution without approval</li>
<li><strong>Confidential:</strong> Restricted access on a need-to-know basis</li>
<li><strong>Restricted:</strong> Highest sensitivity; encrypted, access-logged, and audited</li>
</ul>

<h2>4. Asset Lifecycle</h2>
<p>Assets tracked from acquisition through deployment, maintenance, and secure disposal. Disposal procedures include data sanitization per NIST 800-88.</p>',
     'html',
     'Asset management policy covering inventory, classification, and lifecycle management.',
     'Initial version',
     'initial',
     160,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Risk Management Policy template content (ISO 27001)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000010a', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000021', 1, TRUE,
     '<h1>Risk Management Policy</h1>
<h2>1. Purpose</h2>
<p>This policy establishes the organization''s approach to identifying, assessing, treating, and monitoring information security risks.</p>

<h2>2. Risk Assessment</h2>
<ul>
<li>Annual comprehensive risk assessment covering all critical assets and processes</li>
<li>Ad-hoc assessments for significant changes (new systems, acquisitions, regulatory changes)</li>
<li>Methodology: likelihood × impact matrix (5×5 scale)</li>
</ul>

<h2>3. Risk Treatment</h2>
<ul>
<li><strong>Mitigate:</strong> Implement controls to reduce risk to acceptable level</li>
<li><strong>Transfer:</strong> Use insurance or contracts to share risk</li>
<li><strong>Accept:</strong> Document acceptance with CISO approval for risks below threshold</li>
<li><strong>Avoid:</strong> Eliminate the activity creating the risk</li>
</ul>

<h2>4. Monitoring</h2>
<p>Risk register reviewed quarterly. Key risk indicators (KRIs) tracked and reported to management. Escalation for risks exceeding appetite thresholds.</p>',
     'html',
     'Risk management policy covering assessment methodology, treatment options, and monitoring.',
     'Initial version',
     'initial',
     165,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Data Privacy Policy template content (GDPR)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000011a', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000030', 1, TRUE,
     '<h1>Data Privacy Policy</h1>
<h2>1. Purpose</h2>
<p>This policy establishes the organization''s approach to processing personal data in compliance with GDPR and other applicable privacy regulations.</p>

<h2>2. Lawful Basis</h2>
<p>Personal data shall only be processed where a lawful basis exists: consent, contract, legal obligation, vital interests, public task, or legitimate interest. The lawful basis must be documented before processing begins.</p>

<h2>3. Data Subject Rights</h2>
<ul>
<li>Right to access — respond within 30 days</li>
<li>Right to rectification — correct inaccurate data promptly</li>
<li>Right to erasure — honor within 30 days unless legal exception applies</li>
<li>Right to portability — provide data in structured, machine-readable format</li>
<li>Right to object — cease processing unless compelling legitimate grounds</li>
</ul>

<h2>4. Data Protection Impact Assessments</h2>
<p>DPIAs required for high-risk processing: large-scale profiling, sensitive data, systematic monitoring. Conducted before processing begins.</p>

<h2>5. Cross-Border Transfers</h2>
<p>Personal data transfers outside EEA require appropriate safeguards: SCCs, adequacy decisions, or binding corporate rules.</p>',
     'html',
     'GDPR data privacy policy covering lawful basis, data subject rights, DPIAs, and cross-border transfers.',
     'Initial version',
     'initial',
     195,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Data Retention Policy template content (GDPR/CCPA)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000012a', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000031', 1, TRUE,
     '<h1>Data Retention &amp; Disposal Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines retention periods for different data categories and secure disposal procedures for data that has exceeded its retention period.</p>

<h2>2. Retention Schedules</h2>
<ul>
<li><strong>Employee records:</strong> Duration of employment + 7 years</li>
<li><strong>Financial records:</strong> 7 years (tax/audit requirements)</li>
<li><strong>Customer PII:</strong> Duration of relationship + 3 years</li>
<li><strong>Security logs:</strong> Minimum 12 months, recommended 24 months</li>
<li><strong>Marketing data:</strong> Until consent withdrawn or 2 years of inactivity</li>
</ul>

<h2>3. Disposal Procedures</h2>
<ul>
<li>Electronic data: cryptographic erasure or multi-pass overwrite per NIST 800-88</li>
<li>Physical media: cross-cut shredding or certified destruction</li>
<li>Cloud data: verify deletion from all replicas and backups</li>
</ul>

<h2>4. Legal Holds</h2>
<p>Retention schedules suspended during legal holds. Legal team notifies data custodians. Hold releases trigger normal retention schedule resumption.</p>',
     'html',
     'Data retention and disposal policy covering retention schedules, disposal procedures, and legal holds.',
     'Initial version',
     'initial',
     180,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- Consumer Privacy Rights Policy template content (CCPA)
INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    ('b5000000-0000-0000-0000-000000000013a', 'a0000000-0000-0000-0000-000000000001',
     'b4000000-0000-0000-0000-000000000040', 1, TRUE,
     '<h1>Consumer Privacy Rights Policy</h1>
<h2>1. Purpose</h2>
<p>This policy defines procedures for handling consumer privacy rights requests under CCPA/CPRA.</p>

<h2>2. Consumer Rights</h2>
<ul>
<li><strong>Right to Know:</strong> Consumers may request categories and specific pieces of personal information collected</li>
<li><strong>Right to Delete:</strong> Consumers may request deletion of personal information</li>
<li><strong>Right to Opt-Out:</strong> Consumers may opt out of sale/sharing of personal information</li>
<li><strong>Right to Correct:</strong> Consumers may request correction of inaccurate personal information</li>
<li><strong>Right to Limit:</strong> Consumers may limit use of sensitive personal information</li>
</ul>

<h2>3. Request Handling</h2>
<ul>
<li>Verify consumer identity before processing requests</li>
<li>Acknowledge receipt within 10 business days</li>
<li>Fulfill requests within 45 calendar days (90 with extension)</li>
<li>Maintain request log for 24 months</li>
</ul>

<h2>4. Non-Discrimination</h2>
<p>Consumers exercising privacy rights shall not receive discriminatory treatment in pricing, service level, or quality.</p>',
     'html',
     'CCPA/CPRA consumer privacy rights policy covering right to know, delete, opt-out, correct, and limit.',
     'Initial version',
     'initial',
     190,
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- ============================================================================
-- UPDATE current_version_id FOR ALL TEMPLATES
-- ============================================================================

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000001'
WHERE id = 'b4000000-0000-0000-0000-000000000001' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000002'
WHERE id = 'b4000000-0000-0000-0000-000000000002' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000003'
WHERE id = 'b4000000-0000-0000-0000-000000000003' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000004'
WHERE id = 'b4000000-0000-0000-0000-000000000004' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000005'
WHERE id = 'b4000000-0000-0000-0000-000000000005' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000006'
WHERE id = 'b4000000-0000-0000-0000-000000000010' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000007'
WHERE id = 'b4000000-0000-0000-0000-000000000011' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000008'
WHERE id = 'b4000000-0000-0000-0000-000000000012' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000009'
WHERE id = 'b4000000-0000-0000-0000-000000000020' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000010a'
WHERE id = 'b4000000-0000-0000-0000-000000000021' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000011a'
WHERE id = 'b4000000-0000-0000-0000-000000000030' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000012a'
WHERE id = 'b4000000-0000-0000-0000-000000000031' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'b5000000-0000-0000-0000-000000000013a'
WHERE id = 'b4000000-0000-0000-0000-000000000040' AND current_version_id IS NULL;
