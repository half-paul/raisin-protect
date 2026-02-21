-- Seed Data: seed.sql
-- Description: Demo organization, users, frameworks, requirements, controls, and mappings
-- Created: 2026-02-20
-- Sprint: 1+2 — Full seed for Project Scaffolding, Auth, Frameworks & Controls
--
-- Password for all demo users: demo123
-- Bcrypt hash (cost 12): $2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS

-- ============================================================================
-- SPRINT 1: DEMO ORGANIZATION
-- ============================================================================

INSERT INTO organizations (id, name, slug, domain, status, settings)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Acme Corporation',
    'acme-corp',
    'acme.example.com',
    'active',
    '{"timezone": "America/New_York", "locale": "en-US"}'::jsonb
)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================================
-- SPRINT 1: DEMO USERS (one per GRC role)
-- ============================================================================

INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role, status)
VALUES
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'compliance@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Alice', 'Compliance', 'compliance_manager', 'active'),
    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'security@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Bob', 'Security', 'security_engineer', 'active'),
    ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'it@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Carol', 'IT', 'it_admin', 'active'),
    ('b0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'ciso@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'David', 'CISO', 'ciso', 'active'),
    ('b0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'devops@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Eve', 'DevOps', 'devops_engineer', 'active'),
    ('b0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001',
     'auditor@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Frank', 'Auditor', 'auditor', 'active'),
    ('b0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001',
     'vendor@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Grace', 'Vendor', 'vendor_manager', 'active')
ON CONFLICT (org_id, email) DO NOTHING;

-- ============================================================================
-- SPRINT 1: SAMPLE AUDIT LOG ENTRIES
-- ============================================================================

INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    ('a0000000-0000-0000-0000-000000000001', NULL, 'org.created', 'organization',
     'a0000000-0000-0000-0000-000000000001',
     '{"source": "seed", "name": "Acme Corporation"}'::jsonb, '127.0.0.1'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'user.register', 'user', 'b0000000-0000-0000-0000-000000000004',
     '{"email": "ciso@acme.example.com", "role": "ciso"}'::jsonb, '127.0.0.1'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'user.role_assigned', 'user', 'b0000000-0000-0000-0000-000000000001',
     '{"email": "compliance@acme.example.com", "role": "compliance_manager", "assigned_by": "ciso"}'::jsonb, '127.0.0.1'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'user.login', 'user', 'b0000000-0000-0000-0000-000000000001',
     '{"email": "compliance@acme.example.com", "method": "password"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', NULL, 'user.login_failed', 'user', NULL,
     '{"email": "unknown@acme.example.com", "reason": "user_not_found"}'::jsonb, '10.0.0.55'::inet)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 2: FRAMEWORKS
-- ============================================================================

INSERT INTO frameworks (id, identifier, name, description, category, website_url) VALUES
    ('f0000000-0000-0000-0000-000000000001', 'soc2', 'SOC 2',
     'Service Organization Control 2 — Trust Services Criteria for security, availability, processing integrity, confidentiality, and privacy.',
     'security_privacy', 'https://www.aicpa.org/soc2'),
    ('f0000000-0000-0000-0000-000000000002', 'iso27001', 'ISO 27001',
     'International standard for information security management systems (ISMS).',
     'security_privacy', 'https://www.iso.org/standard/27001'),
    ('f0000000-0000-0000-0000-000000000003', 'pci_dss', 'PCI DSS',
     'Payment Card Industry Data Security Standard — requirements for protecting cardholder data.',
     'payment', 'https://www.pcisecuritystandards.org/'),
    ('f0000000-0000-0000-0000-000000000004', 'gdpr', 'GDPR',
     'General Data Protection Regulation — EU data privacy and protection law.',
     'data_privacy', 'https://gdpr.eu/'),
    ('f0000000-0000-0000-0000-000000000005', 'ccpa', 'CCPA/CPRA',
     'California Consumer Privacy Act / California Privacy Rights Act.',
     'data_privacy', 'https://oag.ca.gov/privacy/ccpa')
ON CONFLICT (identifier) DO NOTHING;

-- ============================================================================
-- SPRINT 2: FRAMEWORK VERSIONS
-- ============================================================================

INSERT INTO framework_versions (id, framework_id, version, display_name, status, effective_date, total_requirements) VALUES
    ('v0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000001', '2024', 'SOC 2 (2024 TSC)', 'active', '2024-01-01', 64),
    ('v0000000-0000-0000-0000-000000000002', 'f0000000-0000-0000-0000-000000000002', '2022', 'ISO 27001:2022', 'active', '2022-10-25', 93),
    ('v0000000-0000-0000-0000-000000000003', 'f0000000-0000-0000-0000-000000000003', '4.0.1', 'PCI DSS v4.0.1', 'active', '2024-06-11', 280),
    ('v0000000-0000-0000-0000-000000000004', 'f0000000-0000-0000-0000-000000000004', '2016', 'GDPR (2016/679)', 'active', '2018-05-25', 99),
    ('v0000000-0000-0000-0000-000000000005', 'f0000000-0000-0000-0000-000000000005', '2023', 'CCPA/CPRA (2023)', 'active', '2023-01-01', 42)
ON CONFLICT (framework_id, version) DO NOTHING;

-- ============================================================================
-- SPRINT 2: REQUIREMENTS — SOC 2 (2024 TSC)
-- 9 CC categories + Availability + Confidentiality + Processing Integrity + Privacy
-- ============================================================================

-- SOC 2 — CC1: Control Environment
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC1', 'Control Environment', 1, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000010', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000001', 'CC1.1', 'The entity demonstrates a commitment to integrity and ethical values', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000011', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000001', 'CC1.2', 'The board of directors demonstrates independence from management and exercises oversight', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000012', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000001', 'CC1.3', 'Management establishes structures, reporting lines, and authorities', 3, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000013', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000001', 'CC1.4', 'The entity demonstrates a commitment to attract, develop, and retain competent individuals', 4, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000014', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000001', 'CC1.5', 'The entity holds individuals accountable for their internal control responsibilities', 5, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC2: Communication and Information
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC2', 'Communication and Information', 2, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000020', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000002', 'CC2.1', 'The entity obtains or generates relevant, quality information to support internal control', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000021', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000002', 'CC2.2', 'The entity internally communicates information to support the functioning of internal control', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000022', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000002', 'CC2.3', 'The entity communicates with external parties regarding matters affecting the functioning of internal control', 3, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC3: Risk Assessment
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000003', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC3', 'Risk Assessment', 3, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000030', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000003', 'CC3.1', 'The entity specifies objectives with sufficient clarity to enable identification of risks', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000031', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000003', 'CC3.2', 'The entity identifies risks to the achievement of its objectives and analyzes risks', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000032', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000003', 'CC3.3', 'The entity considers the potential for fraud in assessing risks', 3, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000033', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000003', 'CC3.4', 'The entity identifies and assesses changes that could significantly impact internal controls', 4, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC4: Monitoring Activities
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000004', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC4', 'Monitoring Activities', 4, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000040', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000004', 'CC4.1', 'The entity selects, develops, and performs ongoing evaluations to ascertain whether controls are present and functioning', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000041', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000004', 'CC4.2', 'The entity evaluates and communicates internal control deficiencies in a timely manner', 2, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC5: Control Activities
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000005', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC5', 'Control Activities', 5, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000050', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000005', 'CC5.1', 'The entity selects and develops control activities that contribute to mitigation of risks', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000051', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000005', 'CC5.2', 'The entity deploys control activities through policies and procedures', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000052', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000005', 'CC5.3', 'The entity selects and develops general controls over technology', 3, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC6: Logical and Physical Access Controls
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000006', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC6', 'Logical and Physical Access Controls', 6, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000060', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.1', 'The entity implements logical access security over protected information assets', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000061', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.2', 'Prior to issuing system credentials, the entity registers and authorizes new users', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000062', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.3', 'The entity authorizes, modifies, or removes access based on authorization and changes', 3, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000063', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.4', 'The entity restricts physical access to facilities and protected information assets', 4, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000064', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.5', 'The entity discontinues logical and physical protections over assets only by authorized personnel', 5, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000065', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.6', 'The entity implements logical access security measures to protect against threats from outside its boundaries', 6, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000066', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.7', 'The entity restricts the transmission, movement, and removal of information to authorized users', 7, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000067', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000006', 'CC6.8', 'The entity implements controls to prevent or detect and act upon unauthorized software', 8, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC7: System Operations
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000007', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC7', 'System Operations', 7, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000070', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000007', 'CC7.1', 'To meet its objectives, the entity uses detection and monitoring procedures to identify changes', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000071', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000007', 'CC7.2', 'The entity monitors system components for anomalies indicative of malicious acts or natural disasters', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000072', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000007', 'CC7.3', 'The entity evaluates security events to determine whether they could or have resulted in incidents', 3, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000073', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000007', 'CC7.4', 'The entity responds to identified security incidents by executing a defined incident response program', 4, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000074', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000007', 'CC7.5', 'The entity identifies, develops, and implements activities to recover from identified security incidents', 5, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC8: Change Management
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000008', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC8', 'Change Management', 8, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000080', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000008', 'CC8.1', 'The entity authorizes, designs, develops, configures, documents, tests, approves, and implements changes', 1, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — CC9: Risk Mitigation
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000009', 'v0000000-0000-0000-0000-000000000001', NULL, 'CC9', 'Risk Mitigation', 9, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000090', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000009', 'CC9.1', 'The entity identifies, selects, and develops risk mitigation activities for risks from business processes', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000091', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000009', 'CC9.2', 'The entity assesses and manages risks associated with vendors and business partners', 2, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — A1: Availability
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000100', 'v0000000-0000-0000-0000-000000000001', NULL, 'A1', 'Additional Criteria for Availability', 10, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000101', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000100', 'A1.1', 'The entity maintains, monitors, and evaluates current processing capacity and availability', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000102', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000100', 'A1.2', 'The entity authorizes, designs, develops, and implements environmental protections and recovery infrastructure', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000103', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000100', 'A1.3', 'The entity tests recovery plan procedures supporting system recovery', 3, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — C1: Confidentiality
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000110', 'v0000000-0000-0000-0000-000000000001', NULL, 'C1', 'Additional Criteria for Confidentiality', 11, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000111', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000110', 'C1.1', 'The entity identifies and maintains confidential information to meet confidentiality commitments', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000112', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000110', 'C1.2', 'The entity disposes of confidential information to meet confidentiality commitments', 2, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — PI1: Processing Integrity
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000120', 'v0000000-0000-0000-0000-000000000001', NULL, 'PI1', 'Additional Criteria for Processing Integrity', 12, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000121', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000120', 'PI1.1', 'The entity obtains or generates, uses, and communicates relevant quality information about processing objectives', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000122', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000120', 'PI1.2', 'The entity implements policies and procedures over system inputs to result in complete, accurate, and timely processing', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000123', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000120', 'PI1.3', 'The entity implements policies and procedures over system processing to result in complete, accurate, and timely processing', 3, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000124', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000120', 'PI1.4', 'The entity implements policies and procedures to make available or deliver output completely, accurately, and timely', 4, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000125', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000120', 'PI1.5', 'The entity implements policies and procedures to store inputs, items in processing, and outputs', 5, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOC 2 — P1: Privacy
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0100000-0000-0000-0000-000000000130', 'v0000000-0000-0000-0000-000000000001', NULL, 'P1', 'Additional Criteria for Privacy', 13, 0, FALSE),
    ('r0100000-0000-0000-0000-000000000131', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.1', 'The entity provides notice about its privacy practices to meet privacy commitments', 1, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000132', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.2', 'The entity communicates choices available regarding the collection, use, and disclosure of personal information', 2, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000133', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.3', 'Personal information is collected consistent with the entity''s objectives and privacy commitments', 3, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000134', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.4', 'The entity limits the use of personal information to the purposes identified in the notice', 4, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000135', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.5', 'The entity retains personal information consistent with commitments and objectives', 5, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000136', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.6', 'The entity disposes of personal information to meet privacy commitments', 6, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000137', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.7', 'The entity discloses personal information to third parties consistent with privacy commitments', 7, 1, TRUE),
    ('r0100000-0000-0000-0000-000000000138', 'v0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000130', 'P1.8', 'The entity provides data subjects the ability to access, correct, amend, or delete their personal information', 8, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- SPRINT 2: REQUIREMENTS — ISO 27001:2022 (Annex A)
-- 4 themes: Organizational (A.5), People (A.6), Physical (A.7), Technological (A.8)
-- ============================================================================

-- ISO 27001 — A.5: Organizational Controls (37 controls)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0200000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000002', NULL, 'A.5', 'Organizational Controls', 1, 0, FALSE),
    ('r0200000-0000-0000-0000-000000000010', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.1', 'Policies for information security', 1, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000011', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.2', 'Information security roles and responsibilities', 2, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000012', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.3', 'Segregation of duties', 3, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000013', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.4', 'Management responsibilities', 4, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000014', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.5', 'Contact with authorities', 5, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000015', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.6', 'Contact with special interest groups', 6, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000016', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.7', 'Threat intelligence', 7, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000017', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.8', 'Information security in project management', 8, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000018', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.9', 'Inventory of information and other associated assets', 9, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000019', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.10', 'Acceptable use of information and other associated assets', 10, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000020', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.11', 'Return of assets', 11, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000021', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.12', 'Classification of information', 12, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000022', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.13', 'Labelling of information', 13, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000023', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.14', 'Information transfer', 14, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000024', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.15', 'Access control', 15, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000025', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.16', 'Identity management', 16, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000026', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.17', 'Authentication information', 17, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000027', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.18', 'Access rights', 18, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000028', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.19', 'Information security in supplier relationships', 19, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000029', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.20', 'Addressing information security within supplier agreements', 20, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000030', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.21', 'Managing information security in the ICT supply chain', 21, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000031', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.22', 'Monitoring, review and change management of supplier services', 22, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000032', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.23', 'Information security for use of cloud services', 23, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000033', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.24', 'Information security incident management planning and preparation', 24, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000034', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.25', 'Assessment and decision on information security events', 25, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000035', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.26', 'Response to information security incidents', 26, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000036', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.27', 'Learning from information security incidents', 27, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000037', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.28', 'Collection of evidence', 28, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000038', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.29', 'Information security during disruption', 29, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000039', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.30', 'ICT readiness for business continuity', 30, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000040', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.31', 'Legal, statutory, regulatory and contractual requirements', 31, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000041', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.32', 'Intellectual property rights', 32, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000042', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.33', 'Protection of records', 33, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000043', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.34', 'Privacy and protection of PII', 34, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000044', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.35', 'Independent review of information security', 35, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000045', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.36', 'Compliance with policies, rules and standards for information security', 36, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000046', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000001', 'A.5.37', 'Documented operating procedures', 37, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ISO 27001 — A.6: People Controls (8 controls)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0200000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000002', NULL, 'A.6', 'People Controls', 2, 0, FALSE),
    ('r0200000-0000-0000-0000-000000000060', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.1', 'Screening', 1, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000061', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.2', 'Terms and conditions of employment', 2, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000062', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.3', 'Information security awareness, education and training', 3, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000063', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.4', 'Disciplinary process', 4, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000064', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.5', 'Responsibilities after termination or change of employment', 5, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000065', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.6', 'Confidentiality or non-disclosure agreements', 6, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000066', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.7', 'Remote working', 7, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000067', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000002', 'A.6.8', 'Information security event reporting', 8, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ISO 27001 — A.7: Physical Controls (14 controls)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0200000-0000-0000-0000-000000000003', 'v0000000-0000-0000-0000-000000000002', NULL, 'A.7', 'Physical Controls', 3, 0, FALSE),
    ('r0200000-0000-0000-0000-000000000070', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.1', 'Physical security perimeters', 1, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000071', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.2', 'Physical entry', 2, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000072', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.3', 'Securing offices, rooms and facilities', 3, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000073', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.4', 'Physical security monitoring', 4, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000074', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.5', 'Protecting against physical and environmental threats', 5, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000075', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.6', 'Working in secure areas', 6, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000076', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.7', 'Clear desk and clear screen', 7, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000077', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.8', 'Equipment siting and protection', 8, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000078', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.9', 'Security of assets off-premises', 9, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000079', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.10', 'Storage media', 10, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000080', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.11', 'Supporting utilities', 11, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000081', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.12', 'Cabling security', 12, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000082', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.13', 'Equipment maintenance', 13, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000083', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000003', 'A.7.14', 'Secure disposal or re-use of equipment', 14, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ISO 27001 — A.8: Technological Controls (34 controls)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0200000-0000-0000-0000-000000000004', 'v0000000-0000-0000-0000-000000000002', NULL, 'A.8', 'Technological Controls', 4, 0, FALSE),
    ('r0200000-0000-0000-0000-000000000084', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.1', 'User endpoint devices', 1, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000085', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.2', 'Privileged access rights', 2, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000086', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.3', 'Information access restriction', 3, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000087', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.4', 'Access to source code', 4, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000088', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.5', 'Secure authentication', 5, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000089', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.6', 'Capacity management', 6, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000090', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.7', 'Protection against malware', 7, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000091', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.8', 'Management of technical vulnerabilities', 8, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000092', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.9', 'Configuration management', 9, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000093', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.10', 'Information deletion', 10, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000094', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.11', 'Data masking', 11, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000095', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.12', 'Data leakage prevention', 12, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000096', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.13', 'Information backup', 13, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000097', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.14', 'Redundancy of information processing facilities', 14, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000098', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.15', 'Logging', 15, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000099', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.16', 'Monitoring activities', 16, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000100', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.17', 'Clock synchronization', 17, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000101', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.18', 'Use of privileged utility programs', 18, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000102', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.19', 'Installation of software on operational systems', 19, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000103', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.20', 'Networks security', 20, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000104', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.21', 'Security of network services', 21, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000105', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.22', 'Segregation of networks', 22, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000106', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.23', 'Web filtering', 23, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000107', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.24', 'Use of cryptography', 24, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000108', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.25', 'Secure development life cycle', 25, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000109', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.26', 'Application security requirements', 26, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000110', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.27', 'Secure system architecture and engineering principles', 27, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000111', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.28', 'Secure coding', 28, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000112', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.29', 'Security testing in development and acceptance', 29, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000113', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.30', 'Outsourced development', 30, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000114', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.31', 'Separation of development, test and production environments', 31, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000115', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.32', 'Change management', 32, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000116', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.33', 'Test information', 33, 1, TRUE),
    ('r0200000-0000-0000-0000-000000000117', 'v0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000004', 'A.8.34', 'Protection of information systems during audit testing', 34, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- SPRINT 2: REQUIREMENTS — PCI DSS v4.0.1
-- 12 top-level requirements with key sub-requirements (hierarchical)
-- ============================================================================

-- PCI DSS — Requirement families (top-level)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000003', NULL, '1', 'Install and Maintain Network Security Controls', 1, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000003', NULL, '2', 'Apply Secure Configurations to All System Components', 2, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000003', 'v0000000-0000-0000-0000-000000000003', NULL, '3', 'Protect Stored Account Data', 3, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000004', 'v0000000-0000-0000-0000-000000000003', NULL, '4', 'Protect Cardholder Data with Strong Cryptography During Transmission', 4, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000005', 'v0000000-0000-0000-0000-000000000003', NULL, '5', 'Protect All Systems and Networks from Malicious Software', 5, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000006', 'v0000000-0000-0000-0000-000000000003', NULL, '6', 'Develop and Maintain Secure Systems and Software', 6, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000007', 'v0000000-0000-0000-0000-000000000003', NULL, '7', 'Restrict Access to System Components and Cardholder Data by Business Need to Know', 7, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000008', 'v0000000-0000-0000-0000-000000000003', NULL, '8', 'Identify Users and Authenticate Access to System Components', 8, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000009', 'v0000000-0000-0000-0000-000000000003', NULL, '9', 'Restrict Physical Access to Cardholder Data', 9, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000010', 'v0000000-0000-0000-0000-000000000003', NULL, '10', 'Log and Monitor All Access to System Components and Cardholder Data', 10, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000011', 'v0000000-0000-0000-0000-000000000003', NULL, '11', 'Test Security of Systems and Networks Regularly', 11, 0, FALSE),
    ('r0300000-0000-0000-0000-000000000012', 'v0000000-0000-0000-0000-000000000003', NULL, '12', 'Support Information Security with Organizational Policies and Programs', 12, 0, FALSE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- PCI DSS — Requirement 1 sub-requirements
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000000101', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000001', '1.1', 'Processes and mechanisms for network security controls are defined and understood', 1, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000102', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000101', '1.1.1', 'All security policies and operational procedures identified in Req 1 are documented and kept up to date', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000103', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000101', '1.1.2', 'Roles and responsibilities for performing activities in Req 1 are documented, assigned, and understood', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000110', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000001', '1.2', 'Network security controls are configured and maintained', 2, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000111', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000110', '1.2.1', 'Configuration standards for NSC rulesets are defined, implemented, and maintained', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000112', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000110', '1.2.5', 'All services, protocols, and ports allowed are identified, approved, and have a defined business need', 5, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000120', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000001', '1.3', 'Network access to and from the cardholder data environment is restricted', 3, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000121', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000120', '1.3.1', 'Inbound traffic to the CDE is restricted to only necessary traffic', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000122', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000120', '1.3.2', 'Outbound traffic from the CDE is restricted to only necessary traffic', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000130', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000001', '1.4', 'Network connections between trusted and untrusted networks are controlled', 4, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000131', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000130', '1.4.1', 'NSCs are implemented between trusted and untrusted networks', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000132', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000130', '1.4.2', 'Inbound traffic from untrusted networks to trusted networks is restricted to authorized communications', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000150', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000001', '1.5', 'Risks to the CDE from computing devices connecting via untrusted networks are mitigated', 5, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000151', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000150', '1.5.1', 'Security controls are implemented on any computing device that connects to both untrusted and trusted networks', 1, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- PCI DSS — Requirement 6 sub-requirements (critical: secure development)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000000601', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000006', '6.1', 'Processes and mechanisms for developing secure systems are defined and understood', 1, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000602', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000601', '6.1.1', 'All security policies and operational procedures in Req 6 are documented and kept up to date', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000610', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000006', '6.2', 'Bespoke and custom software are developed securely', 2, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000611', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000610', '6.2.1', 'Bespoke and custom software is developed securely following industry best practices', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000612', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000610', '6.2.2', 'Software development personnel are trained at least once every 12 months', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000613', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000610', '6.2.3', 'Bespoke and custom software is reviewed prior to release to identify and correct potential coding vulnerabilities', 3, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000614', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000610', '6.2.4', 'Software engineering techniques prevent or mitigate common software attacks and related vulnerabilities', 4, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000620', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000006', '6.3', 'Security vulnerabilities are identified and addressed', 3, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000621', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000620', '6.3.1', 'Security vulnerabilities are identified and managed via established process', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000622', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000620', '6.3.2', 'An inventory of bespoke and custom software is maintained to facilitate vulnerability management', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000623', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000620', '6.3.3', 'All system components are protected from known vulnerabilities by installing applicable patches', 3, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000630', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000006', '6.4', 'Public-facing web applications are protected against attacks', 4, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000631', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000630', '6.4.1', 'For public-facing web applications, new threats and vulnerabilities are addressed on an ongoing basis', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000632', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000630', '6.4.2', 'For public-facing web applications, an automated technical solution is deployed that detects and prevents web-based attacks', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000633', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000630', '6.4.3', 'All payment page scripts that are loaded and executed in the consumer browser are managed', 3, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000640', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000006', '6.5', 'Changes to all system components are managed securely', 5, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000641', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000640', '6.5.1', 'Changes are managed using established change control procedures', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000642', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000640', '6.5.2', 'Upon completion of a significant change, all applicable PCI DSS requirements are confirmed to be in place', 2, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- PCI DSS — Requirement 8 sub-requirements (authentication)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000000801', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000008', '8.1', 'Processes and mechanisms for identification and authentication are defined and understood', 1, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000802', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000801', '8.1.1', 'All security policies and procedures in Req 8 are documented and kept up to date', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000810', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000008', '8.2', 'User identification and related accounts are strictly managed', 2, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000811', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000810', '8.2.1', 'All users are assigned a unique ID before access to system components', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000812', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000810', '8.2.2', 'Group, shared, or generic accounts are not used except where specifically allowed', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000820', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000008', '8.3', 'Strong authentication is established and managed', 3, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000821', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000820', '8.3.1', 'All user access to system components is authenticated via at least one factor', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000822', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000820', '8.3.2', 'Strong cryptography is used to render all authentication factors unreadable during transmission and storage', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000823', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000820', '8.3.6', 'If passwords/passphrases are used, they meet minimum complexity requirements', 6, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000824', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000820', '8.3.9', 'If passwords/passphrases are the only authentication factor, they are changed at least once every 90 days', 9, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000830', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000008', '8.4', 'Multi-factor authentication is implemented to secure access', 4, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000831', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000830', '8.4.1', 'MFA is implemented for all non-console access into the CDE for personnel with administrative access', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000832', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000830', '8.4.2', 'MFA is implemented for all access into the CDE', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000833', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000830', '8.4.3', 'MFA is implemented for all remote network access originating from outside the entity network', 3, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000840', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000008', '8.5', 'Multi-factor authentication systems are configured to prevent misuse', 5, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000841', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000840', '8.5.1', 'MFA systems are implemented with all authentication factors required', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000000850', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000008', '8.6', 'Use of application and system accounts is strictly managed', 6, 1, FALSE),
    ('r0300000-0000-0000-0000-000000000851', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000850', '8.6.1', 'If accounts used by systems or applications can be used for interactive login, they are managed as follows', 1, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- PCI DSS — Requirement 10 sub-requirements (logging)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000001001', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000010', '10.1', 'Processes and mechanisms for logging and monitoring are defined and understood', 1, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001002', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001001', '10.1.1', 'All security policies and procedures in Req 10 are documented and kept up to date', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001010', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000010', '10.2', 'Audit logs are implemented to support detection of anomalies', 2, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001011', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001010', '10.2.1', 'Audit logs are enabled and active for all system components and cardholder data', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001012', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001010', '10.2.2', 'Audit logs record all actions taken by any individual with administrative access', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001020', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000010', '10.3', 'Audit logs are protected from destruction and unauthorized modifications', 3, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001021', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001020', '10.3.1', 'Read access to audit logs files is limited to those with a job-related need', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001022', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001020', '10.3.2', 'Audit log files are protected to prevent modifications by individuals', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001023', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001020', '10.3.3', 'Audit log files are promptly backed up to a secure, central, internal log server', 3, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001030', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000010', '10.4', 'Audit logs are reviewed to identify anomalies or suspicious activity', 4, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001031', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001030', '10.4.1', 'Security events are reviewed at least once daily', 1, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- PCI DSS — Requirement 11 sub-requirements (testing)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000001101', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000011', '11.1', 'Processes for security testing are defined and understood', 1, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001102', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001101', '11.1.1', 'All security policies and procedures in Req 11 are documented and kept up to date', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001110', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000011', '11.3', 'External and internal vulnerabilities are regularly identified, prioritized, and addressed', 3, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001111', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001110', '11.3.1', 'Internal vulnerability scans are performed at least once every three months', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001112', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001110', '11.3.2', 'External vulnerability scans are performed at least once every three months', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001120', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000011', '11.4', 'External and internal penetration testing is regularly performed', 4, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001121', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001120', '11.4.1', 'External penetration testing is performed at least once every 12 months', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001130', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000011', '11.5', 'Network intrusions and unexpected file changes are detected and responded to', 5, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001131', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001130', '11.5.1', 'Intrusion-detection/prevention techniques are used to detect and/or prevent intrusions into the network', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001132', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001130', '11.5.2', 'A change-detection mechanism is deployed to alert on unauthorized modification of critical files', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001140', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000011', '11.6', 'Unauthorized changes on payment pages are detected and responded to', 6, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001141', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001140', '11.6.1', 'A change- and tamper-detection mechanism is deployed to alert on unauthorized changes to HTTP headers and payment page content', 1, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- PCI DSS — Requirement 12 sub-requirements (policies)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0300000-0000-0000-0000-000000001201', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000012', '12.1', 'A comprehensive information security policy is established and maintained', 1, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001202', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001201', '12.1.1', 'An overall information security policy is established, published, and disseminated', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001203', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001201', '12.1.2', 'The information security policy is reviewed at least once every 12 months', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001210', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000012', '12.3', 'Risks to the cardholder data environment are formally identified, evaluated, and managed', 3, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001211', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001210', '12.3.1', 'A targeted risk analysis is performed for each PCI DSS requirement', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001220', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000012', '12.6', 'Security awareness education is an ongoing activity', 6, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001221', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001220', '12.6.1', 'A formal security awareness program is implemented', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001222', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001220', '12.6.2', 'The security awareness program is reviewed at least once every 12 months', 2, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001230', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000012', '12.8', 'Risk to information assets from relationships with third parties is managed', 8, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001231', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001230', '12.8.1', 'A list of all third-party service providers is maintained', 1, 2, TRUE),
    ('r0300000-0000-0000-0000-000000001240', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000000012', '12.10', 'Security incidents and suspected security incidents are responded to immediately', 10, 1, FALSE),
    ('r0300000-0000-0000-0000-000000001241', 'v0000000-0000-0000-0000-000000000003', 'r0300000-0000-0000-0000-000000001240', '12.10.1', 'An incident response plan exists to be initiated in the event of a suspected or confirmed security incident', 1, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- SPRINT 2: REQUIREMENTS — GDPR
-- Key articles relevant to compliance management
-- ============================================================================

-- GDPR — Chapter groups
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0400000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000004', NULL, 'Ch.II', 'Principles', 1, 0, FALSE),
    ('r0400000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000004', NULL, 'Ch.III', 'Rights of the Data Subject', 2, 0, FALSE),
    ('r0400000-0000-0000-0000-000000000003', 'v0000000-0000-0000-0000-000000000004', NULL, 'Ch.IV', 'Controller and Processor', 3, 0, FALSE),
    ('r0400000-0000-0000-0000-000000000004', 'v0000000-0000-0000-0000-000000000004', NULL, 'Ch.V', 'Transfers of Personal Data to Third Countries', 4, 0, FALSE),
    ('r0400000-0000-0000-0000-000000000005', 'v0000000-0000-0000-0000-000000000004', NULL, 'Ch.IX', 'Specific Processing Situations', 5, 0, FALSE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- GDPR — Chapter II: Principles
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0400000-0000-0000-0000-000000000010', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000001', 'Art.5', 'Principles relating to processing of personal data', 1, 1, FALSE),
    ('r0400000-0000-0000-0000-000000000011', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000010', 'Art.5(1)(a)', 'Lawfulness, fairness and transparency', 1, 2, TRUE),
    ('r0400000-0000-0000-0000-000000000012', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000010', 'Art.5(1)(b)', 'Purpose limitation', 2, 2, TRUE),
    ('r0400000-0000-0000-0000-000000000013', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000010', 'Art.5(1)(c)', 'Data minimisation', 3, 2, TRUE),
    ('r0400000-0000-0000-0000-000000000014', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000010', 'Art.5(1)(d)', 'Accuracy', 4, 2, TRUE),
    ('r0400000-0000-0000-0000-000000000015', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000010', 'Art.5(1)(e)', 'Storage limitation', 5, 2, TRUE),
    ('r0400000-0000-0000-0000-000000000016', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000010', 'Art.5(1)(f)', 'Integrity and confidentiality', 6, 2, TRUE),
    ('r0400000-0000-0000-0000-000000000017', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000010', 'Art.5(2)', 'Accountability', 7, 2, TRUE),
    ('r0400000-0000-0000-0000-000000000020', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000001', 'Art.6', 'Lawfulness of processing', 2, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000021', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000001', 'Art.7', 'Conditions for consent', 3, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000022', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000001', 'Art.8', 'Conditions applicable to child''s consent', 4, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000023', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000001', 'Art.9', 'Processing of special categories of personal data', 5, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000024', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000001', 'Art.10', 'Processing of personal data relating to criminal convictions', 6, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000025', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000001', 'Art.11', 'Processing which does not require identification', 7, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- GDPR — Chapter III: Rights of Data Subject
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0400000-0000-0000-0000-000000000030', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.12', 'Transparent information, communication and modalities', 1, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000031', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.13', 'Information to be provided where personal data are collected from the data subject', 2, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000032', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.14', 'Information to be provided where personal data have not been obtained from the data subject', 3, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000033', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.15', 'Right of access by the data subject', 4, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000034', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.16', 'Right to rectification', 5, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000035', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.17', 'Right to erasure (right to be forgotten)', 6, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000036', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.18', 'Right to restriction of processing', 7, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000037', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.19', 'Notification obligation regarding rectification or erasure', 8, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000038', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.20', 'Right to data portability', 9, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000039', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.21', 'Right to object', 10, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000040', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.22', 'Automated individual decision-making, including profiling', 11, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000041', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000002', 'Art.23', 'Restrictions on rights and obligations', 12, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- GDPR — Chapter IV: Controller and Processor
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0400000-0000-0000-0000-000000000050', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.24', 'Responsibility of the controller', 1, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000051', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.25', 'Data protection by design and by default', 2, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000052', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.26', 'Joint controllers', 3, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000053', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.27', 'Representatives of controllers not established in the Union', 4, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000054', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.28', 'Processor', 5, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000055', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.29', 'Processing under the authority of the controller or processor', 6, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000056', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.30', 'Records of processing activities', 7, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000057', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.31', 'Cooperation with the supervisory authority', 8, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000058', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.32', 'Security of processing', 9, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000059', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.33', 'Notification of a personal data breach to the supervisory authority', 10, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000060', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.34', 'Communication of a personal data breach to the data subject', 11, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000061', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.35', 'Data protection impact assessment', 12, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000062', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.36', 'Prior consultation', 13, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000063', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.37', 'Designation of the data protection officer', 14, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000064', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.38', 'Position of the data protection officer', 15, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000065', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.39', 'Tasks of the data protection officer', 16, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000066', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.40', 'Codes of conduct', 17, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000067', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.41', 'Monitoring of approved codes of conduct', 18, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000068', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.42', 'Certification', 19, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000069', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000003', 'Art.43', 'Certification bodies', 20, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- GDPR — Chapter V: Transfers
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0400000-0000-0000-0000-000000000070', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000004', 'Art.44', 'General principle for transfers', 1, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000071', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000004', 'Art.45', 'Transfers on the basis of an adequacy decision', 2, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000072', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000004', 'Art.46', 'Transfers subject to appropriate safeguards', 3, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000073', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000004', 'Art.47', 'Binding corporate rules', 4, 1, TRUE),
    ('r0400000-0000-0000-0000-000000000074', 'v0000000-0000-0000-0000-000000000004', 'r0400000-0000-0000-0000-000000000004', 'Art.49', 'Derogations for specific situations', 5, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;
-- ============================================================================
-- SPRINT 2: REQUIREMENTS — CCPA/CPRA (2023)
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('r0500000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.100', 'Consumer Rights', 1, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000010', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000001', '1798.100(a)', 'Right to know what personal information is collected', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000011', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000001', '1798.100(b)', 'Right to know what personal information is sold or disclosed', 2, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000012', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000001', '1798.100(d)', 'Right to delete personal information', 3, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000013', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000001', '1798.100(e)', 'Right to correct inaccurate personal information', 4, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.105', 'Right to Deletion', 2, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000020', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000002', '1798.105(a)', 'Consumer right to request deletion of personal information', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000021', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000002', '1798.105(b)', 'Business obligation to comply with deletion requests', 2, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000003', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.106', 'Right to Correction', 3, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000030', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000003', '1798.106(a)', 'Consumer right to request correction of inaccurate personal information', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000004', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.110', 'Right to Know', 4, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000040', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000004', '1798.110(a)', 'Consumer right to request categories of personal information collected', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000041', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000004', '1798.110(b)', 'Consumer right to request specific pieces of personal information collected', 2, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000005', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.115', 'Right to Know - Disclosure', 5, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000050', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000005', '1798.115(a)', 'Consumer right to know about disclosure for business purposes', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000006', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.120', 'Right to Opt-Out', 6, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000060', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000006', '1798.120(a)', 'Consumer right to opt-out of sale or sharing of personal information', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000061', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000006', '1798.120(b)', 'Business obligation to respect opt-out requests', 2, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000007', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.121', 'Right to Limit Use of Sensitive Personal Information', 7, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000070', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000007', '1798.121(a)', 'Consumer right to limit use and disclosure of sensitive personal information', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000008', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.125', 'Non-Discrimination', 8, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000080', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000008', '1798.125(a)', 'Business shall not discriminate against consumer for exercising rights', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000009', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.130', 'Notice and Process Requirements', 9, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000090', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000009', '1798.130(a)(1)', 'Methods for submitting requests to know, delete, and correct', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000091', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000009', '1798.130(a)(2)', 'Verification of consumer requests', 2, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000092', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000009', '1798.130(a)(5)', 'Disclosure of personal information within 45 days', 3, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000100', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.135', 'Opt-Out and Opt-In', 10, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000101', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000100', '1798.135(a)', 'Clear and conspicuous link on homepage titled Do Not Sell or Share My Personal Information', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000102', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000100', '1798.135(b)', 'Clear and conspicuous link titled Limit the Use of My Sensitive Personal Information', 2, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000110', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.140', 'Definitions and Categories', 11, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000111', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000110', '1798.140(v)', 'Categories of sensitive personal information defined and tracked', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000120', 'v0000000-0000-0000-0000-000000000005', NULL, 'Sec.1798.185', 'Regulations', 12, 0, FALSE),
    ('r0500000-0000-0000-0000-000000000121', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000120', '1798.185(a)(15)', 'Cybersecurity audit regulations for businesses with significant risk', 1, 1, TRUE),
    ('r0500000-0000-0000-0000-000000000122', 'v0000000-0000-0000-0000-000000000005', 'r0500000-0000-0000-0000-000000000120', '1798.185(a)(16)', 'Risk assessment regulations for processing that presents significant risk', 2, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- SPRINT 2: ORG FRAMEWORK ACTIVATIONS (demo org activates all 5)
-- ============================================================================

INSERT INTO org_frameworks (id, org_id, framework_id, active_version_id, status, target_date, notes) VALUES
    ('d0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000001', 'v0000000-0000-0000-0000-000000000001', 'active', '2026-06-30', 'Primary compliance target for Q2 2026'),
    ('d0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000002', 'v0000000-0000-0000-0000-000000000002', 'active', '2026-09-30', 'ISO certification planned for Q3 2026'),
    ('d0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000003', 'v0000000-0000-0000-0000-000000000003', 'active', '2026-12-31', 'Payment processing compliance — must be compliant by EOY'),
    ('d0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000004', 'v0000000-0000-0000-0000-000000000004', 'active', NULL, 'GDPR compliance required for EU customer base'),
    ('d0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000005', 'v0000000-0000-0000-0000-000000000005', 'active', NULL, 'CCPA compliance required for California consumer data')
ON CONFLICT (org_id, framework_id) DO NOTHING;

-- ============================================================================
-- SPRINT 2: CONTROL LIBRARY (300+ controls for demo org)
-- Owner assignments by role:
--   technical → Bob (security engineer) b0000000-...-000000000002
--   administrative → Alice (compliance manager) b0000000-...-000000000001
--   physical → Carol (IT admin) b0000000-...-000000000003
--   operational → Eve (DevOps) b0000000-...-000000000005
-- ============================================================================

-- ---------------------------------------------------------------------------
-- ACCESS CONTROL (AC) — 30 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-001', 'Multi-Factor Authentication', 'Enforce MFA for all user accounts accessing production systems and sensitive data. Supports TOTP, WebAuthn, and push-based authentication.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-001'),
    ('c0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-002', 'Role-Based Access Control', 'Implement RBAC with least-privilege access. Users receive only permissions required for their role. Roles are defined per system and reviewed quarterly.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-002'),
    ('c0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-003', 'Quarterly Access Reviews', 'Conduct quarterly reviews of user access across all critical systems. Remove stale or excessive permissions. Document findings and remediation.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-AC-003'),
    ('c0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-004', 'Privileged Access Management', 'Privileged accounts are managed through a PAM solution with session recording, just-in-time access, and automatic credential rotation.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-004'),
    ('c0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-005', 'Account Provisioning and Deprovisioning', 'Automated provisioning of user accounts upon hire and deprovisioning within 24 hours of termination. Tied to HR system events.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-AC-005'),
    ('c0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-006', 'Password Policy Enforcement', 'Enforce minimum password complexity (12+ chars, mixed case, numbers, symbols). Prevent password reuse for last 24 passwords. Lock accounts after 5 failed attempts.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-006'),
    ('c0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-007', 'Single Sign-On (SSO)', 'Centralized authentication via SSO for all SaaS and internal applications using SAML 2.0 or OIDC protocols.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-007'),
    ('c0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-008', 'Session Management', 'Enforce session timeouts (15 min idle, 8 hour absolute). Implement secure session tokens with anti-fixation protections.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-008'),
    ('c0000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-009', 'Service Account Management', 'All service accounts are inventoried, have defined owners, use non-interactive authentication, and have credentials rotated every 90 days.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-AC-009'),
    ('c0000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-010', 'Remote Access Policy', 'Remote access to corporate resources requires VPN with MFA. Split tunneling is disabled. All remote sessions are logged.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-010'),
    ('c0000000-0000-0000-0000-000000000011', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-011', 'Access Control Policy', 'Formal access control policy defining principles of least privilege, need-to-know, and separation of duties. Reviewed annually.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-AC-011'),
    ('c0000000-0000-0000-0000-000000000012', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-012', 'Unique User Identification', 'All users are assigned a unique identifier. Shared or group accounts are prohibited except where technically unavoidable and documented.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-012'),
    ('c0000000-0000-0000-0000-000000000013', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-013', 'API Authentication and Authorization', 'All API endpoints require authentication via OAuth2/JWT tokens. Authorization checks enforce resource-level access control.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-013'),
    ('c0000000-0000-0000-0000-000000000014', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-014', 'Database Access Restrictions', 'Direct database access is restricted to DBAs via bastion hosts. Application access uses dedicated service accounts with minimal permissions.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-014'),
    ('c0000000-0000-0000-0000-000000000015', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-015', 'Emergency Access Procedure', 'Break-glass procedures exist for emergency access to critical systems. All emergency access is logged, reviewed within 24h, and requires justification.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-AC-015'),
    ('c0000000-0000-0000-0000-000000000016', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-016', 'Identity Lifecycle Management', 'User identities are managed from creation through modification to deletion with formal approval workflows at each stage.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-AC-016'),
    ('c0000000-0000-0000-0000-000000000017', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-017', 'Conditional Access Policies', 'Access decisions factor in device compliance, location, risk score, and time-of-day. Non-compliant devices are restricted to limited access.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-017'),
    ('c0000000-0000-0000-0000-000000000018', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-018', 'Third-Party Access Management', 'Third-party access is granted via dedicated accounts with expiration dates, restricted scope, and enhanced logging.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-AC-018'),
    ('c0000000-0000-0000-0000-000000000019', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-019', 'Admin Console Access Control', 'Access to cloud admin consoles (AWS, GCP, Azure) requires MFA, IP whitelisting, and is limited to designated cloud administrators.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-019'),
    ('c0000000-0000-0000-0000-000000000020', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-020', 'Network Access Control (NAC)', 'Network access control verifies device identity and posture before allowing network connectivity. Non-compliant devices are quarantined.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-020'),
    ('c0000000-0000-0000-0000-000000000021', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-021', 'Data Access Classification', 'Access to data is controlled based on data classification level. Highly sensitive data requires additional authorization and logging.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-AC-021'),
    ('c0000000-0000-0000-0000-000000000022', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-022', 'Segregation of Duties', 'Critical business processes enforce separation of duties. No single individual can initiate, approve, and execute sensitive transactions.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-AC-022'),
    ('c0000000-0000-0000-0000-000000000023', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-023', 'Endpoint Device Management', 'All endpoint devices accessing corporate resources are enrolled in MDM/UEM. Device policies enforce encryption, screen lock, and patch compliance.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-AC-023'),
    ('c0000000-0000-0000-0000-000000000024', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-024', 'Wireless Network Security', 'Enterprise Wi-Fi uses WPA3-Enterprise with certificate-based authentication. Guest networks are isolated from corporate resources.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-024'),
    ('c0000000-0000-0000-0000-000000000025', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-025', 'Source Code Access Control', 'Access to source code repositories is restricted based on team membership. Code changes require peer review before merge.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-AC-025'),
    ('c0000000-0000-0000-0000-000000000026', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-026', 'Production Environment Access', 'Production access is limited to on-call engineers via time-boxed escalation. All production changes go through CI/CD pipelines.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-AC-026'),
    ('c0000000-0000-0000-0000-000000000027', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-027', 'Contractor Access Offboarding', 'Contractor access is automatically revoked upon contract end date. Access reviews verify no residual permissions remain.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-AC-027'),
    ('c0000000-0000-0000-0000-000000000028', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-028', 'Privileged Session Monitoring', 'All privileged sessions are recorded and stored for 12 months. Anomalous privileged activity triggers automated alerts.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-028'),
    ('c0000000-0000-0000-0000-000000000029', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-029', 'Cross-Account Access Controls', 'Cross-account and cross-tenant access in cloud environments uses assume-role patterns with external ID verification.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-AC-029'),
    ('c0000000-0000-0000-0000-000000000030', 'a0000000-0000-0000-0000-000000000001', 'CTRL-AC-030', 'Access Request Workflow', 'All access requests follow a formal workflow with manager approval, security review for sensitive systems, and automated provisioning.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-AC-030')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- CONFIGURATION MANAGEMENT (CM) — 25 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000031', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-001', 'Baseline Configuration Standards', 'Hardened baseline configurations are defined for all system types (servers, databases, network devices). CIS Benchmarks are used as the baseline.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-CM-001'),
    ('c0000000-0000-0000-0000-000000000032', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-002', 'Configuration Drift Detection', 'Automated configuration drift detection runs daily against approved baselines. Drift is reported and must be remediated within SLA.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-002'),
    ('c0000000-0000-0000-0000-000000000033', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-003', 'Infrastructure as Code', 'All infrastructure is defined as code (Terraform/Pulumi). Manual changes to production are prohibited. IaC is version controlled.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-003'),
    ('c0000000-0000-0000-0000-000000000034', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-004', 'Asset Inventory Management', 'Automated discovery and inventory of all hardware and software assets. Inventory is reconciled monthly. Unauthorized assets are flagged.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-CM-004'),
    ('c0000000-0000-0000-0000-000000000035', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-005', 'Default Credential Removal', 'Default credentials are changed or disabled on all systems before deployment. Automated scans verify no default credentials exist.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-CM-005'),
    ('c0000000-0000-0000-0000-000000000036', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-006', 'Unnecessary Services Disabled', 'All unnecessary services, protocols, and ports are disabled. Only services required for business function are enabled and documented.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-CM-006'),
    ('c0000000-0000-0000-0000-000000000037', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-007', 'System Hardening Procedures', 'Formal hardening procedures exist for each system type. Procedures are reviewed and updated with each major OS/application release.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-007'),
    ('c0000000-0000-0000-0000-000000000038', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-008', 'Container Image Security', 'Container images are built from approved base images, scanned for vulnerabilities, and signed. Only signed images can be deployed.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-008'),
    ('c0000000-0000-0000-0000-000000000039', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-009', 'Software Inventory and Licensing', 'All installed software is inventoried with license status. Unauthorized or unlicensed software is removed within 48 hours of detection.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-CM-009'),
    ('c0000000-0000-0000-0000-000000000040', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-010', 'Configuration Change Control', 'All configuration changes follow the change management process with documented approval, testing, rollback plans, and post-change verification.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-CM-010'),
    ('c0000000-0000-0000-0000-000000000041', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-011', 'Cloud Configuration Standards', 'Cloud environments follow provider-specific security benchmarks (AWS Well-Architected, GCP Security, Azure Security Benchmark).', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-011'),
    ('c0000000-0000-0000-0000-000000000042', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-012', 'NTP Synchronization', 'All systems synchronize time via NTP from authorized time sources. Clock drift tolerance is configured at under 1 second.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-012'),
    ('c0000000-0000-0000-0000-000000000043', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-013', 'Environment Separation', 'Development, testing, staging, and production environments are logically or physically separated. Production data is not used in non-production.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-013'),
    ('c0000000-0000-0000-0000-000000000044', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-014', 'Secure DNS Configuration', 'DNSSEC is enabled where possible. Internal DNS resolvers are hardened. DNS queries from production are logged.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-CM-014'),
    ('c0000000-0000-0000-0000-000000000045', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-015', 'TLS Configuration Standards', 'TLS 1.2 minimum is enforced across all services. Weak cipher suites are disabled. Certificates are managed via automated renewal.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-CM-015'),
    ('c0000000-0000-0000-0000-000000000046', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-016', 'Secrets Management', 'Secrets (API keys, passwords, certificates) are stored in a dedicated vault service. No secrets in code repositories or configuration files.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-016'),
    ('c0000000-0000-0000-0000-000000000047', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-017', 'Browser Security Configuration', 'Managed browsers enforce security settings including safe browsing, extension whitelisting, and automatic updates.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-CM-017'),
    ('c0000000-0000-0000-0000-000000000048', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-018', 'Database Configuration Standards', 'Database instances follow hardened configurations: remote root login disabled, TLS required, audit logging enabled, default schemas removed.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-CM-018'),
    ('c0000000-0000-0000-0000-000000000049', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-019', 'Firmware Update Management', 'Firmware on network devices, servers, and IoT devices is tracked and updated according to vendor advisories and risk assessment.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-CM-019'),
    ('c0000000-0000-0000-0000-000000000050', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-020', 'Acceptable Use of Assets', 'Acceptable use policy defines permitted use of company assets including devices, networks, email, and cloud services. Reviewed annually.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-CM-020'),
    ('c0000000-0000-0000-0000-000000000051', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-021', 'Cloud Resource Tagging', 'All cloud resources are tagged with owner, cost center, environment, and data classification. Untagged resources are flagged for remediation.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-CM-021'),
    ('c0000000-0000-0000-0000-000000000052', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-022', 'Mobile Device Management', 'Mobile devices accessing corporate data are enrolled in MDM with remote wipe capability, encryption enforcement, and jailbreak detection.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-CM-022'),
    ('c0000000-0000-0000-0000-000000000053', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-023', 'Application Whitelisting', 'Application whitelisting is enforced on critical servers. Only approved executables can run in production environments.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-CM-023'),
    ('c0000000-0000-0000-0000-000000000054', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-024', 'Configuration Audit', 'Quarterly configuration audits verify systems match approved baselines. Findings are tracked to remediation with assigned owners.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-CM-024'),
    ('c0000000-0000-0000-0000-000000000055', 'a0000000-0000-0000-0000-000000000001', 'CTRL-CM-025', 'Return of Assets Procedure', 'Upon termination, all company assets (laptops, badges, tokens) are returned within 3 business days. Remote wipe is performed on any unreturned devices.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-CM-025')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- DATA PROTECTION (DP) — 25 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000056', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-001', 'Data Classification Policy', 'Four-tier data classification scheme: Public, Internal, Confidential, Restricted. Classification drives handling, storage, and access requirements.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-001'),
    ('c0000000-0000-0000-0000-000000000057', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-002', 'Encryption at Rest', 'All data at rest is encrypted using AES-256. Encryption keys are managed through a dedicated KMS with automatic rotation.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-002'),
    ('c0000000-0000-0000-0000-000000000058', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-003', 'Encryption in Transit', 'All data in transit is encrypted using TLS 1.2 or higher. Internal service-to-service communication uses mTLS.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-003'),
    ('c0000000-0000-0000-0000-000000000059', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-004', 'Data Loss Prevention (DLP)', 'DLP controls monitor and prevent unauthorized transmission of sensitive data via email, web, and removable media.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-004'),
    ('c0000000-0000-0000-0000-000000000060', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-005', 'Backup and Recovery', 'Automated daily backups with 30-day retention. Backups are encrypted and stored in a separate region. Recovery tested quarterly.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-DP-005'),
    ('c0000000-0000-0000-0000-000000000061', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-006', 'Data Retention and Disposal', 'Data retention schedules are defined per data type and regulatory requirement. Expired data is securely deleted with verification.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-006'),
    ('c0000000-0000-0000-0000-000000000062', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-007', 'PII Handling Procedures', 'Specific handling procedures for personally identifiable information including collection minimization, purpose limitation, and consent tracking.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-007'),
    ('c0000000-0000-0000-0000-000000000063', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-008', 'Data Masking and Anonymization', 'PII and sensitive data is masked or anonymized in non-production environments. Techniques include tokenization, pseudonymization, and k-anonymity.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-008'),
    ('c0000000-0000-0000-0000-000000000064', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-009', 'Key Management', 'Cryptographic key lifecycle management covering generation, distribution, storage, rotation, and destruction. HSMs used for root keys.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-009'),
    ('c0000000-0000-0000-0000-000000000065', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-010', 'Secure File Transfer', 'Sensitive data transfers use SFTP, SCP, or HTTPS with mutual authentication. Transfer logs are retained for 12 months.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-010'),
    ('c0000000-0000-0000-0000-000000000066', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-011', 'Media Sanitization', 'Storage media is sanitized before disposal or reuse using NIST SP 800-88 guidelines. Destruction certificates are maintained.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-DP-011'),
    ('c0000000-0000-0000-0000-000000000067', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-012', 'Data Processing Agreements', 'Data processing agreements with all third parties who process personal data define permitted processing, security requirements, and breach notification.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-012'),
    ('c0000000-0000-0000-0000-000000000068', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-013', 'Cross-Border Data Transfer Controls', 'Personal data transfers outside the EEA use approved mechanisms: adequacy decisions, SCCs, or binding corporate rules.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-013'),
    ('c0000000-0000-0000-0000-000000000069', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-014', 'Privacy Notice Management', 'Privacy notices are maintained for all data collection points. Notices are clear, accessible, and updated when processing changes.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-014'),
    ('c0000000-0000-0000-0000-000000000070', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-015', 'Consent Management', 'Consent for personal data processing is obtained, recorded, and can be withdrawn. Consent records are immutable and auditable.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-015'),
    ('c0000000-0000-0000-0000-000000000071', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-016', 'Data Subject Rights Process', 'Processes exist for handling data subject requests (access, rectification, erasure, portability) within regulatory timelines (30 days GDPR, 45 days CCPA).', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-016'),
    ('c0000000-0000-0000-0000-000000000072', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-017', 'Records of Processing Activities', 'Maintain a register of all processing activities including purpose, categories of data, recipients, retention periods, and legal basis.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-017'),
    ('c0000000-0000-0000-0000-000000000073', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-018', 'Data Protection Impact Assessment', 'DPIAs are conducted for high-risk processing activities before implementation. Residual risks are documented and accepted by management.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-018'),
    ('c0000000-0000-0000-0000-000000000074', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-019', 'Data Labeling', 'Sensitive data stores and documents are labeled according to classification. Automated classification tools scan for unlabeled sensitive data.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-DP-019'),
    ('c0000000-0000-0000-0000-000000000075', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-020', 'Database Encryption', 'Transparent data encryption (TDE) is enabled on all production databases. Column-level encryption for credit card numbers and SSNs.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-020'),
    ('c0000000-0000-0000-0000-000000000076', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-021', 'Data Minimization', 'Only necessary personal data is collected. Collection forms and APIs are reviewed to ensure no excessive data collection.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-021'),
    ('c0000000-0000-0000-0000-000000000077', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-022', 'Opt-Out Mechanism', 'Consumers can opt out of data sale/sharing via a clearly visible link. Opt-out requests are processed within 15 business days.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-022'),
    ('c0000000-0000-0000-0000-000000000078', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-023', 'Non-Discrimination for Rights Exercise', 'Consumers who exercise privacy rights are not discriminated against in pricing, service level, or quality of goods/services.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-023'),
    ('c0000000-0000-0000-0000-000000000079', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-024', 'Data Protection by Design', 'Privacy and data protection requirements are embedded into system design from the earliest stages. Privacy review is part of the SDLC.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-DP-024'),
    ('c0000000-0000-0000-0000-000000000080', 'a0000000-0000-0000-0000-000000000001', 'CTRL-DP-025', 'Cardholder Data Protection', 'Cardholder data is stored only when absolutely necessary, masked in display (first 6/last 4), and never stored post-authorization (CVV, full track data).', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-DP-025')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- INCIDENT RESPONSE (IR) — 20 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000081', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-001', 'Incident Response Plan', 'Documented incident response plan covering detection, analysis, containment, eradication, recovery, and post-incident review.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-001'),
    ('c0000000-0000-0000-0000-000000000082', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-002', 'Incident Response Team', 'Defined incident response team with clear roles (incident commander, technical lead, communications, legal). On-call rotation maintained 24/7.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-002'),
    ('c0000000-0000-0000-0000-000000000083', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-003', 'Incident Classification', 'Incidents are classified by severity (P1-P4) with defined response timelines, escalation paths, and communication requirements.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-003'),
    ('c0000000-0000-0000-0000-000000000084', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-004', 'Security Event Triage', 'Security events from SIEM and detection tools are triaged by the SOC team. Events are classified, prioritized, and escalated within defined SLAs.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-004'),
    ('c0000000-0000-0000-0000-000000000085', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-005', 'Breach Notification Process', 'Data breach notification procedures comply with GDPR (72h to DPA), CCPA (expedient notification), and PCI DSS requirements. Templates pre-approved.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-IR-005'),
    ('c0000000-0000-0000-0000-000000000086', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-006', 'Incident Tabletop Exercises', 'Tabletop exercises conducted quarterly simulating different threat scenarios. Lessons learned are incorporated into the IRP.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-006'),
    ('c0000000-0000-0000-0000-000000000087', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-007', 'Digital Forensics Capability', 'Forensic investigation tools and procedures are maintained. Evidence handling follows chain-of-custody requirements for potential legal proceedings.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-007'),
    ('c0000000-0000-0000-0000-000000000088', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-008', 'Post-Incident Review', 'Blameless post-incident reviews are conducted within 5 business days of incident resolution. Action items are tracked to completion.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-008'),
    ('c0000000-0000-0000-0000-000000000089', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-009', 'Incident Communication Plan', 'Pre-defined communication templates for customers, regulators, media, and internal stakeholders during security incidents.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-IR-009'),
    ('c0000000-0000-0000-0000-000000000090', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-010', 'Automated Incident Response', 'SOAR playbooks automate initial containment actions for common incident types (compromised credentials, malware, data exfiltration).', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-010'),
    ('c0000000-0000-0000-0000-000000000091', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-011', 'Incident Tracking System', 'All incidents are tracked in a dedicated system with lifecycle management, SLA tracking, and management reporting.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-IR-011'),
    ('c0000000-0000-0000-0000-000000000092', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-012', 'Regulatory Reporting Procedures', 'Procedures for reporting security incidents to relevant regulators (ICO, CNIL, FTC, PCI Council) within required timelines.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-IR-012'),
    ('c0000000-0000-0000-0000-000000000093', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-013', 'Business Continuity Plan', 'BCP covers critical business functions with RTOs and RPOs. Plan is tested annually through simulation exercises.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-IR-013'),
    ('c0000000-0000-0000-0000-000000000094', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-014', 'Disaster Recovery Plan', 'DR plan covers infrastructure failover with documented procedures, automated failover where possible, and tested quarterly.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-IR-014'),
    ('c0000000-0000-0000-0000-000000000095', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-015', 'Incident Metrics and Reporting', 'Monthly incident metrics: MTTD, MTTR, incident volume by type, repeat incidents. Reported to CISO and executive team.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-IR-015'),
    ('c0000000-0000-0000-0000-000000000096', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-016', 'External IR Retainer', 'Retainer agreement with a specialized IR firm for surge support during major incidents. Engagement procedures are pre-defined.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-IR-016'),
    ('c0000000-0000-0000-0000-000000000097', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-017', 'ICT Continuity', 'ICT systems supporting critical business processes have redundancy and failover. RPO ≤1h, RTO ≤4h for tier-1 systems.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-IR-017'),
    ('c0000000-0000-0000-0000-000000000098', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-018', 'Threat Intelligence Integration', 'Threat intelligence feeds are integrated with SIEM and IDS. IoCs are automatically correlated with observed activity.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-018'),
    ('c0000000-0000-0000-0000-000000000099', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-019', 'Recovery Testing', 'Backup recovery and DR procedures are tested quarterly. Recovery success metrics are tracked and reported.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-IR-019'),
    ('c0000000-0000-0000-0000-000000000100', 'a0000000-0000-0000-0000-000000000001', 'CTRL-IR-020', 'Security Operations Center', 'SOC provides 24/7 monitoring capability through combination of in-house and managed security services.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-IR-020')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- LOGGING & MONITORING (LM) — 25 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000101', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-001', 'Centralized Log Management', 'All system, application, and security logs are collected in a centralized SIEM. Log retention is 12 months online, 7 years archived.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-001'),
    ('c0000000-0000-0000-0000-000000000102', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-002', 'Audit Log Integrity', 'Audit logs are write-once and cannot be modified or deleted by administrators. Integrity is verified through log checksums.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-002'),
    ('c0000000-0000-0000-0000-000000000103', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-003', 'Security Event Monitoring', 'Real-time monitoring of security events with correlation rules for detecting attack patterns, anomalies, and policy violations.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-003'),
    ('c0000000-0000-0000-0000-000000000104', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-004', 'Privileged Activity Monitoring', 'All actions by privileged users are logged with enhanced detail. Privileged activity is reviewed daily by the security team.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-004'),
    ('c0000000-0000-0000-0000-000000000105', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-005', 'Authentication Event Logging', 'All authentication events (success, failure, lockout, MFA) are logged with source IP, timestamp, user agent, and outcome.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-005'),
    ('c0000000-0000-0000-0000-000000000106', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-006', 'Network Traffic Monitoring', 'Network traffic is monitored for anomalies including unusual data volumes, connections to known-bad destinations, and lateral movement.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-006'),
    ('c0000000-0000-0000-0000-000000000107', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-007', 'Database Activity Monitoring', 'Database queries are monitored for suspicious patterns: bulk exports, schema changes, access outside business hours, and direct access bypassing application tier.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-007'),
    ('c0000000-0000-0000-0000-000000000108', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-008', 'Cloud Infrastructure Monitoring', 'Cloud provider audit logs (CloudTrail, Activity Log, Audit Log) are enabled for all accounts and forwarded to central SIEM.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-LM-008'),
    ('c0000000-0000-0000-0000-000000000109', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-009', 'Alerting and Escalation', 'Critical security alerts are delivered to on-call personnel within 5 minutes. Escalation procedures ensure response within SLA.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-009'),
    ('c0000000-0000-0000-0000-000000000110', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-010', 'Log Access Control', 'Access to audit logs is restricted to security team and auditors. All log access is itself logged. No users can delete logs.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-010'),
    ('c0000000-0000-0000-0000-000000000111', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-011', 'Application Performance Monitoring', 'APM tools monitor application health, error rates, latency, and throughput. Anomaly detection alerts on deviations from baseline.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-LM-011'),
    ('c0000000-0000-0000-0000-000000000112', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-012', 'Capacity Monitoring', 'Infrastructure capacity (CPU, memory, disk, network) is monitored with alerts at 70% and 90% thresholds. Capacity planning is reviewed monthly.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-LM-012'),
    ('c0000000-0000-0000-0000-000000000113', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-013', 'Daily Log Review', 'Security logs are reviewed daily for indicators of compromise. Review scope includes authentication failures, privilege escalation, and data access anomalies.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-013'),
    ('c0000000-0000-0000-0000-000000000114', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-014', 'File Integrity Monitoring', 'FIM monitors critical system files, configuration files, and binaries for unauthorized changes. Changes trigger immediate alerts.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-014'),
    ('c0000000-0000-0000-0000-000000000115', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-015', 'IDS/IPS Deployment', 'Network IDS/IPS deployed at network perimeter and internal segment boundaries. Signatures updated daily. Custom rules for application-specific threats.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-015'),
    ('c0000000-0000-0000-0000-000000000116', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-016', 'Web Application Firewall', 'WAF protects all public-facing web applications. OWASP Core Rule Set is enabled with tuning for false positive reduction.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-016'),
    ('c0000000-0000-0000-0000-000000000117', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-017', 'Endpoint Detection and Response', 'EDR agents deployed on all endpoints providing behavioral detection, threat hunting, and automated containment capabilities.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-017'),
    ('c0000000-0000-0000-0000-000000000118', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-018', 'Anti-Malware Protection', 'Anti-malware with real-time scanning deployed on all endpoints and servers. Definitions updated hourly. Full scans weekly.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-018'),
    ('c0000000-0000-0000-0000-000000000119', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-019', 'Security Dashboard', 'Executive security dashboard showing real-time metrics: open incidents, vulnerability status, compliance posture, and trend analysis.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-019'),
    ('c0000000-0000-0000-0000-000000000120', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-020', 'Log Backup and Archival', 'Logs are backed up daily and archived for 7 years to meet regulatory requirements. Archived logs are retrievable within 48 hours.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-LM-020'),
    ('c0000000-0000-0000-0000-000000000121', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-021', 'Payment Page Monitoring', 'Payment page scripts are inventoried, authorized, and monitored for unauthorized changes. Tamper detection alerts on any modification.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-021'),
    ('c0000000-0000-0000-0000-000000000122', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-022', 'HTTP Header Security Monitoring', 'Security headers (CSP, HSTS, X-Frame-Options) are monitored for unauthorized modifications on all web-facing services.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-022'),
    ('c0000000-0000-0000-0000-000000000123', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-023', 'Change Detection for Critical Files', 'Change detection mechanism deployed for critical system files and payment page content. Alerts within 5 minutes of unauthorized change.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-023'),
    ('c0000000-0000-0000-0000-000000000124', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-024', 'User Behavior Analytics', 'UEBA solutions baseline normal user behavior and alert on anomalies: unusual access patterns, impossible travel, data hoarding.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-024'),
    ('c0000000-0000-0000-0000-000000000125', 'a0000000-0000-0000-0000-000000000001', 'CTRL-LM-025', 'Email Security Monitoring', 'Email gateway monitors inbound/outbound for phishing, malware, BEC, and data exfiltration. Suspicious emails are quarantined.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-LM-025')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- NETWORK SECURITY (NW) — 20 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000126', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-001', 'Network Segmentation', 'Network is segmented into security zones (DMZ, internal, CDE, management). Traffic between zones is controlled by firewall rules.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-001'),
    ('c0000000-0000-0000-0000-000000000127', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-002', 'Firewall Management', 'Firewall rules follow deny-by-default. Rules are reviewed quarterly and unused rules are removed. Changes require formal approval.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-002'),
    ('c0000000-0000-0000-0000-000000000128', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-003', 'VPN Security', 'Remote access VPN enforces MFA, uses IPsec or WireGuard, and restricts access to authorized resources only.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-003'),
    ('c0000000-0000-0000-0000-000000000129', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-004', 'DDoS Protection', 'DDoS mitigation in place for all public-facing services through cloud-based protection (CDN, WAF, rate limiting).', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-004'),
    ('c0000000-0000-0000-0000-000000000130', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-005', 'Network Documentation', 'Current network diagrams document all connections, data flows, and security boundaries. Updated within 5 days of any change.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-005'),
    ('c0000000-0000-0000-0000-000000000131', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-006', 'CDE Network Isolation', 'Cardholder data environment is isolated from all other networks. Access restricted to authorized personnel and systems only.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-006'),
    ('c0000000-0000-0000-0000-000000000132', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-007', 'Inbound Traffic Restriction', 'Inbound traffic to the CDE is restricted to only necessary and documented traffic. All other inbound traffic is denied.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-007'),
    ('c0000000-0000-0000-0000-000000000133', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-008', 'Outbound Traffic Restriction', 'Outbound traffic from the CDE is restricted to documented and approved destinations. DNS, HTTP, and HTTPS are proxied through inspection points.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-008'),
    ('c0000000-0000-0000-0000-000000000134', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-009', 'Micro-Segmentation', 'Workload-level network segmentation in cloud environments using security groups and network policies. Zero-trust east-west traffic.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-NW-009'),
    ('c0000000-0000-0000-0000-000000000135', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-010', 'Web Filtering', 'Web content filtering blocks known-malicious domains and enforces acceptable use categories. Filtering policies are updated daily.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-010'),
    ('c0000000-0000-0000-0000-000000000136', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-011', 'Network Device Hardening', 'Network devices (routers, switches, load balancers) follow vendor hardening guides. Default management credentials changed. SNMP v3 required.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-011'),
    ('c0000000-0000-0000-0000-000000000137', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-012', 'Service Port Management', 'All services, protocols, and ports allowed through firewalls are inventoried, approved by security, and have documented business justification.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-012'),
    ('c0000000-0000-0000-0000-000000000138', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-013', 'Network Security Policy', 'Network security policy defines segmentation requirements, traffic filtering, and secure network design principles. Reviewed annually.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-NW-013'),
    ('c0000000-0000-0000-0000-000000000139', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-014', 'Trusted/Untrusted Network Controls', 'NSCs are implemented at all junctions between trusted and untrusted networks. All traffic from untrusted networks is inspected.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-014'),
    ('c0000000-0000-0000-0000-000000000140', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-015', 'Remote Device Security', 'Devices connecting via untrusted networks have host-based firewall, antivirus, and cannot bridge trusted and untrusted networks simultaneously.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-015'),
    ('c0000000-0000-0000-0000-000000000141', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-016', 'API Gateway Security', 'All external API traffic passes through an API gateway with rate limiting, authentication, input validation, and threat protection.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-016'),
    ('c0000000-0000-0000-0000-000000000142', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-017', 'Network Service Security', 'Network services (DNS, DHCP, NTP) are secured, monitored, and restricted to authorized servers. Rogue service detection is active.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-017'),
    ('c0000000-0000-0000-0000-000000000143', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-018', 'Network Redundancy', 'Critical network paths have redundancy. Automatic failover for internet, WAN, and internal backbone links.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-NW-018'),
    ('c0000000-0000-0000-0000-000000000144', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-019', 'Network Access Logging', 'All network access events including firewall accepts/denies, VPN connections, and NAC decisions are logged for 12 months.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-NW-019'),
    ('c0000000-0000-0000-0000-000000000145', 'a0000000-0000-0000-0000-000000000001', 'CTRL-NW-020', 'Cloud Network Security', 'Cloud VPCs use private subnets for workloads, NAT gateways for outbound, and VPC flow logs enabled for all interfaces.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-NW-020')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- POLICY MANAGEMENT (PM) — 25 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000146', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-001', 'Information Security Policy', 'Overarching information security policy approved by executive management. Reviewed and updated annually. Communicated to all employees.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-001'),
    ('c0000000-0000-0000-0000-000000000147', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-002', 'Policy Management Framework', 'Framework for creating, reviewing, approving, distributing, and retiring security policies. Version control and review schedule maintained.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-002'),
    ('c0000000-0000-0000-0000-000000000148', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-003', 'Acceptable Use Policy', 'Defines acceptable use of information systems, network, email, internet, and portable devices. All employees acknowledge annually.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-003'),
    ('c0000000-0000-0000-0000-000000000149', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-004', 'Data Privacy Policy', 'Privacy policy covering personal data collection, use, storage, sharing, and deletion. Compliant with GDPR, CCPA, and applicable laws.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-004'),
    ('c0000000-0000-0000-0000-000000000150', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-005', 'Change Management Policy', 'Change management policy covering all changes to production systems. Includes risk assessment, approval workflow, and rollback procedures.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-005'),
    ('c0000000-0000-0000-0000-000000000151', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-006', 'Vendor Management Policy', 'Policy for assessing, onboarding, monitoring, and offboarding third-party vendors. Risk-based vendor tiering drives assessment frequency.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-006'),
    ('c0000000-0000-0000-0000-000000000152', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-007', 'Encryption Policy', 'Encryption standards for data at rest and in transit. Minimum key lengths, approved algorithms, and key management procedures.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-007'),
    ('c0000000-0000-0000-0000-000000000153', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-008', 'Security Roles and Responsibilities', 'Security roles and responsibilities documented for all positions. RACI matrix for key security processes maintained and reviewed.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-008'),
    ('c0000000-0000-0000-0000-000000000154', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-009', 'Policy Exception Process', 'Formal process for requesting, reviewing, approving, and tracking exceptions to security policies. Exceptions have time limits and compensating controls.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-009'),
    ('c0000000-0000-0000-0000-000000000155', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-010', 'Regulatory Compliance Tracking', 'Active tracking of applicable regulations and their requirements. Gap assessments performed when new regulations are identified.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-010'),
    ('c0000000-0000-0000-0000-000000000156', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-011', 'Code of Conduct', 'Code of conduct covering ethics, integrity, confidentiality, and reporting obligations. Signed by all employees at hire and annually.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-011'),
    ('c0000000-0000-0000-0000-000000000157', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-012', 'Disciplinary Process', 'Disciplinary process for security policy violations. Graduated response from warning to termination depending on severity.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-012'),
    ('c0000000-0000-0000-0000-000000000158', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-013', 'Board-Level Security Reporting', 'CISO reports to the board quarterly on security posture, risk trends, compliance status, and key initiatives.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000004', FALSE, 'TPL-PM-013'),
    ('c0000000-0000-0000-0000-000000000159', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-014', 'Independent Security Review', 'Annual independent security review (internal audit or third-party assessment) evaluates the effectiveness of the security program.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000006', FALSE, 'TPL-PM-014'),
    ('c0000000-0000-0000-0000-000000000160', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-015', 'Operating Procedures', 'Documented operating procedures for key security processes (backup, patching, monitoring, access provisioning). Reviewed semi-annually.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-PM-015'),
    ('c0000000-0000-0000-0000-000000000161', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-016', 'Legal and Regulatory Requirements', 'Inventory of all applicable legal, statutory, regulatory, and contractual requirements. Compliance status tracked per requirement.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-016'),
    ('c0000000-0000-0000-0000-000000000162', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-017', 'IP Rights Protection', 'Procedures to protect intellectual property rights including software licensing, patent tracking, and open-source compliance.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-017'),
    ('c0000000-0000-0000-0000-000000000163', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-018', 'Records Management', 'Records management policy covering retention, protection, and disposal of business and compliance records.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-018'),
    ('c0000000-0000-0000-0000-000000000164', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-019', 'Data Protection Officer', 'DPO is appointed with documented responsibilities, adequate resources, and reports directly to highest level of management.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-019'),
    ('c0000000-0000-0000-0000-000000000165', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-020', 'Security in Project Management', 'Security requirements are integrated into project management methodology. Security review gate required before production deployment.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-020'),
    ('c0000000-0000-0000-0000-000000000166', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-021', 'Contact with Authorities', 'Procedures for contacting relevant authorities (law enforcement, regulators, CERT teams) during security incidents.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-021'),
    ('c0000000-0000-0000-0000-000000000167', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-022', 'Risk Management Framework', 'Enterprise risk management framework with defined risk appetite, assessment methodology, treatment options, and risk register.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-022'),
    ('c0000000-0000-0000-0000-000000000168', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-023', 'Management Commitment', 'Executive management demonstrates commitment to security through policy sponsorship, resource allocation, and regular engagement.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000004', FALSE, 'TPL-PM-023'),
    ('c0000000-0000-0000-0000-000000000169', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-024', 'Policy Compliance Monitoring', 'Automated and manual checks verify ongoing compliance with security policies. Non-compliance is tracked and remediated within SLA.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-024'),
    ('c0000000-0000-0000-0000-000000000170', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PM-025', 'Risk Assessment Process', 'Annual risk assessments identify, evaluate, and prioritize risks. Targeted risk analyses performed for specific PCI DSS requirements.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-PM-025')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- PHYSICAL & ENVIRONMENTAL (PE) — 15 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000171', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-001', 'Physical Security Perimeters', 'Physical security perimeters protect data center and office spaces. Multi-layer access controls from building to server cage.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-001'),
    ('c0000000-0000-0000-0000-000000000172', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-002', 'Physical Access Control', 'Badge-based access control system for all secure areas. Access logs retained for 12 months. Visitor escort required in secure areas.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-002'),
    ('c0000000-0000-0000-0000-000000000173', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-003', 'CCTV Monitoring', 'Video surveillance covers all entry/exit points, server rooms, and secure areas. Footage retained for 90 days minimum.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-003'),
    ('c0000000-0000-0000-0000-000000000174', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-004', 'Environmental Controls', 'Data center environmental controls include HVAC, fire suppression (FM-200), water detection, and temperature/humidity monitoring.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-004'),
    ('c0000000-0000-0000-0000-000000000175', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-005', 'UPS and Power Protection', 'Uninterruptible power supply provides minimum 30 minutes runtime. Generator backup for extended outages tested monthly.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-005'),
    ('c0000000-0000-0000-0000-000000000176', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-006', 'Secure Area Procedures', 'Documented procedures for working in secure areas. Clean desk policy enforced. No photography or recording devices without authorization.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-006'),
    ('c0000000-0000-0000-0000-000000000177', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-007', 'Equipment Protection', 'Server and network equipment is secured in locked cabinets or cages. Cable management prevents tampering. Tamper-evident seals used.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-007'),
    ('c0000000-0000-0000-0000-000000000178', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-008', 'Visitor Management', 'Visitors are pre-registered, verified at reception, issued temporary badges, escorted at all times, and logged in visitor register.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-008'),
    ('c0000000-0000-0000-0000-000000000179', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-009', 'Clear Desk and Clear Screen', 'Clear desk policy enforced in all work areas. Screens auto-lock after 5 minutes of inactivity. Sensitive documents stored in locked cabinets.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-009'),
    ('c0000000-0000-0000-0000-000000000180', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-010', 'Off-Site Asset Security', 'Procedures for protecting organizational assets used outside the office. Laptops are encrypted, trackable, and remotely wipeable.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-010'),
    ('c0000000-0000-0000-0000-000000000181', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-011', 'Cabling Security', 'Network and power cabling is protected from interception, interference, and damage. Fiber optic for sensitive data paths.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-011'),
    ('c0000000-0000-0000-0000-000000000182', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-012', 'Equipment Maintenance', 'Equipment maintenance follows vendor schedules. Maintenance is performed by authorized personnel only. Maintenance logs are maintained.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-012'),
    ('c0000000-0000-0000-0000-000000000183', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-013', 'Secure Equipment Disposal', 'Equipment containing storage media is securely sanitized or physically destroyed before disposal. Destruction certificates maintained.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-013'),
    ('c0000000-0000-0000-0000-000000000184', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-014', 'Physical Access Review', 'Physical access lists are reviewed quarterly. Access revoked for terminated employees within 24 hours. Badge collection upon departure.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-014'),
    ('c0000000-0000-0000-0000-000000000185', 'a0000000-0000-0000-0000-000000000001', 'CTRL-PE-015', 'Delivery and Loading Security', 'Delivery areas are separated from secure processing areas. Incoming materials are inspected and registered. Delivery personnel do not access secure areas.', 'physical', 'active', 'b0000000-0000-0000-0000-000000000003', FALSE, 'TPL-PE-015')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- RISK ASSESSMENT (RA) — 20 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000186', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-001', 'Annual Risk Assessment', 'Comprehensive risk assessment performed annually covering information security, operational, and compliance risks.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-001'),
    ('c0000000-0000-0000-0000-000000000187', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-002', 'Risk Register', 'Central risk register maintained with risk identification, assessment (likelihood x impact), owner, treatment plan, and status tracking.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-002'),
    ('c0000000-0000-0000-0000-000000000188', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-003', 'Threat Modeling', 'Threat modeling performed for all new applications and major changes using STRIDE or PASTA methodology.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-RA-003'),
    ('c0000000-0000-0000-0000-000000000189', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-004', 'Vendor Risk Assessment', 'Third-party vendors are risk-assessed before onboarding and periodically based on risk tier. Assessment covers security, privacy, and financial stability.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000007', FALSE, 'TPL-RA-004'),
    ('c0000000-0000-0000-0000-000000000190', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-005', 'Risk Appetite Statement', 'Executive-approved risk appetite statement defines acceptable risk levels across security, compliance, and operational domains.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000004', FALSE, 'TPL-RA-005'),
    ('c0000000-0000-0000-0000-000000000191', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-006', 'Fraud Risk Assessment', 'Annual fraud risk assessment evaluates potential for fraud including financial fraud, data theft, and social engineering.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-006'),
    ('c0000000-0000-0000-0000-000000000192', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-007', 'Change Risk Assessment', 'All significant changes undergo risk assessment as part of the change management process. High-risk changes require additional review.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-007'),
    ('c0000000-0000-0000-0000-000000000193', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-008', 'Supply Chain Risk Assessment', 'Annual assessment of risks in the ICT supply chain including software dependencies, hardware procurement, and cloud providers.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000007', FALSE, 'TPL-RA-008'),
    ('c0000000-0000-0000-0000-000000000194', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-009', 'Risk Treatment Plans', 'All risks above appetite have documented treatment plans with timelines, owners, and compensating controls where full remediation is not feasible.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-009'),
    ('c0000000-0000-0000-0000-000000000195', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-010', 'Vulnerability Risk Scoring', 'Vulnerabilities are scored using CVSS with environmental factors. Remediation SLAs: Critical=24h, High=7d, Medium=30d, Low=90d.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-RA-010'),
    ('c0000000-0000-0000-0000-000000000196', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-011', 'Business Impact Analysis', 'BIA identifies critical business processes, their dependencies, and acceptable downtime. Updated annually and after major changes.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-011'),
    ('c0000000-0000-0000-0000-000000000197', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-012', 'Risk Monitoring and Review', 'Key risk indicators (KRIs) are monitored continuously. Risk posture is reported to the risk committee monthly.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-012'),
    ('c0000000-0000-0000-0000-000000000198', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-013', 'Cloud Risk Assessment', 'Risk assessment specific to cloud services covering data residency, shared responsibility, vendor lock-in, and multi-tenancy risks.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-RA-013'),
    ('c0000000-0000-0000-0000-000000000199', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-014', 'Targeted Risk Analysis (PCI)', 'Targeted risk analyses performed for each PCI DSS requirement that allows entity-defined frequencies per Req 12.3.1.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-014'),
    ('c0000000-0000-0000-0000-000000000200', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-015', 'Privacy Risk Assessment', 'Privacy risk assessments conducted for processing activities that involve personal data. Risk-based approach to DPIA requirements.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-015'),
    ('c0000000-0000-0000-0000-000000000201', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-016', 'Vendor SLA Monitoring', 'Third-party vendor SLAs are monitored for security-relevant metrics. SLA breaches trigger vendor review escalation.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000007', FALSE, 'TPL-RA-016'),
    ('c0000000-0000-0000-0000-000000000202', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-017', 'Vendor Contract Requirements', 'Third-party contracts include security requirements: incident notification, audit rights, data protection, and termination provisions.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000007', FALSE, 'TPL-RA-017'),
    ('c0000000-0000-0000-0000-000000000203', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-018', 'Vendor Inventory', 'Maintained list of all third-party service providers with documented services, data shared, risk tier, and last assessment date.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000007', FALSE, 'TPL-RA-018'),
    ('c0000000-0000-0000-0000-000000000204', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-019', 'NDA and Confidentiality', 'Non-disclosure agreements required for all employees and contractors before accessing confidential information.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-RA-019'),
    ('c0000000-0000-0000-0000-000000000205', 'a0000000-0000-0000-0000-000000000001', 'CTRL-RA-020', 'Vendor Offboarding', 'Formal vendor offboarding process including access revocation, data return/destruction, and final security review.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000007', FALSE, 'TPL-RA-020')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- SECURITY AWARENESS (SA) — 20 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000206', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-001', 'Security Awareness Training', 'Annual security awareness training for all employees. Covers phishing, social engineering, password hygiene, data handling, and incident reporting.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-001'),
    ('c0000000-0000-0000-0000-000000000207', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-002', 'Phishing Simulation', 'Monthly phishing simulation campaigns. Users who fail receive targeted training. Results reported to management quarterly.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-002'),
    ('c0000000-0000-0000-0000-000000000208', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-003', 'New Hire Security Orientation', 'Security orientation within first week of employment covering policies, acceptable use, incident reporting, and data handling.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-003'),
    ('c0000000-0000-0000-0000-000000000209', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-004', 'Role-Based Security Training', 'Additional security training based on role: developers (OWASP Top 10), admins (hardening), management (risk), HR (data handling).', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-004'),
    ('c0000000-0000-0000-0000-000000000210', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-005', 'Developer Security Training', 'Software development personnel receive security training at least once every 12 months covering secure coding and common vulnerabilities.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-005'),
    ('c0000000-0000-0000-0000-000000000211', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-006', 'Security Champions Program', 'Designated security champions in each engineering team promote security best practices and serve as first point of contact for security questions.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-006'),
    ('c0000000-0000-0000-0000-000000000212', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-007', 'Privacy Training', 'Annual privacy-specific training covering GDPR, CCPA, data subject rights, data handling, and breach reporting.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-007'),
    ('c0000000-0000-0000-0000-000000000213', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-008', 'Background Checks', 'Background checks performed for all employees before starting. Enhanced checks for roles with access to sensitive data or privileged systems.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-008'),
    ('c0000000-0000-0000-0000-000000000214', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-009', 'Employment Terms and Conditions', 'Employment agreements include security and confidentiality obligations. Employees acknowledge information security responsibilities.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-009'),
    ('c0000000-0000-0000-0000-000000000215', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-010', 'Termination Responsibilities', 'Offboarding procedures include access revocation, asset return, knowledge transfer, and reminder of ongoing confidentiality obligations.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-010'),
    ('c0000000-0000-0000-0000-000000000216', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-011', 'Security Event Reporting', 'All employees are trained on how and when to report security events. Multiple reporting channels available (email, phone, anonymous).', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-011'),
    ('c0000000-0000-0000-0000-000000000217', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-012', 'Remote Work Security Training', 'Training specific to remote work security: home network security, public Wi-Fi risks, physical security of devices and documents.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-012'),
    ('c0000000-0000-0000-0000-000000000218', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-013', 'Training Effectiveness Measurement', 'Training effectiveness measured through quizzes, phishing simulation results, and incident metrics. Program updated based on findings.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-013'),
    ('c0000000-0000-0000-0000-000000000219', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-014', 'Security Communications', 'Regular security communications (newsletter, Slack channel, intranet) keep employees informed of threats, tips, and policy updates.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-014'),
    ('c0000000-0000-0000-0000-000000000220', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-015', 'Social Engineering Defense', 'Training and procedures specifically addressing social engineering attacks: pretexting, tailgating, vishing, and whaling.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-015'),
    ('c0000000-0000-0000-0000-000000000221', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-016', 'Special Interest Group Engagement', 'Active participation in information security special interest groups and information sharing communities (ISACs, OWASP, local chapters).', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-016'),
    ('c0000000-0000-0000-0000-000000000222', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-017', 'Security Awareness Program Review', 'Security awareness program is reviewed at least once every 12 months and updated based on new threats and organizational changes.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-017'),
    ('c0000000-0000-0000-0000-000000000223', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-018', 'Insider Threat Awareness', 'Training covers insider threat indicators and reporting procedures. Anonymous reporting mechanism available.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SA-018'),
    ('c0000000-0000-0000-0000-000000000224', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-019', 'Data Handling Training', 'Training on proper handling of classified data (labeling, storage, transmission, disposal) per data classification level.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-019'),
    ('c0000000-0000-0000-0000-000000000225', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SA-020', 'Competency Assessment', 'Annual competency assessments for personnel in security-critical roles. Training plans address identified gaps.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-SA-020')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- SECURE DEVELOPMENT (SD) — 25 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000226', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-001', 'Secure SDLC', 'Secure development lifecycle with security gates at design, development, testing, and deployment phases. OWASP SAMM model.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-001'),
    ('c0000000-0000-0000-0000-000000000227', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-002', 'Code Review Process', 'All code changes require peer review before merge. Security-sensitive changes require security team review.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-002'),
    ('c0000000-0000-0000-0000-000000000228', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-003', 'Static Application Security Testing', 'SAST tools run in CI/CD pipeline on every commit. Critical and high findings block deployment.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-003'),
    ('c0000000-0000-0000-0000-000000000229', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-004', 'Dynamic Application Security Testing', 'DAST scans run against staging environment before production deployment. Covers OWASP Top 10 and CWE Top 25.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-004'),
    ('c0000000-0000-0000-0000-000000000230', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-005', 'Software Composition Analysis', 'SCA tools analyze open-source dependencies for known vulnerabilities and license compliance. Updated daily.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-005'),
    ('c0000000-0000-0000-0000-000000000231', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-006', 'Secure Coding Standards', 'Documented secure coding standards covering input validation, output encoding, authentication, session management, and error handling.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-006'),
    ('c0000000-0000-0000-0000-000000000232', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-007', 'Input Validation', 'All user input is validated on both client and server side. Parameterized queries for all database operations. Content-type enforcement.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-007'),
    ('c0000000-0000-0000-0000-000000000233', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-008', 'CI/CD Pipeline Security', 'CI/CD pipelines enforce security checks: SAST, SCA, secrets scanning, container scanning, and compliance gates before deployment.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-008'),
    ('c0000000-0000-0000-0000-000000000234', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-009', 'Outsourced Development Security', 'Third-party developed code undergoes same security review and testing as internal code. Contractual security requirements defined.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-009'),
    ('c0000000-0000-0000-0000-000000000235', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-010', 'Security Testing in Acceptance', 'Security testing is a required component of user acceptance testing. Penetration testing for major releases.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-010'),
    ('c0000000-0000-0000-0000-000000000236', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-011', 'Application Security Requirements', 'Security requirements documented for each application covering authentication, authorization, input validation, and data protection.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-011'),
    ('c0000000-0000-0000-0000-000000000237', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-012', 'Test Data Management', 'Test environments use synthetic or anonymized data. Production data is not used in development or testing environments.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-012'),
    ('c0000000-0000-0000-0000-000000000238', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-013', 'Change Management Process', 'Formal change management for all production changes including documentation, approval, testing, and post-implementation review.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-013'),
    ('c0000000-0000-0000-0000-000000000239', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-014', 'Software Installation Control', 'Installation of software on production systems is controlled through automated deployment pipelines. Manual installs are prohibited.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-014'),
    ('c0000000-0000-0000-0000-000000000240', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-015', 'Secure Architecture Principles', 'System architecture follows defense-in-depth, least privilege, fail-secure, and zero-trust principles. Architecture review for major changes.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-015'),
    ('c0000000-0000-0000-0000-000000000241', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-016', 'API Security Standards', 'API security standards covering authentication (OAuth2), authorization, rate limiting, input validation, and output encoding.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-016'),
    ('c0000000-0000-0000-0000-000000000242', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-017', 'Secrets Scanning', 'Pre-commit hooks and CI pipeline scan for secrets (API keys, passwords, tokens) in code. Detected secrets are rotated immediately.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-017'),
    ('c0000000-0000-0000-0000-000000000243', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-018', 'Dependency Management', 'Third-party dependencies are pinned to specific versions. Automated alerts for new vulnerabilities in dependencies.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-018'),
    ('c0000000-0000-0000-0000-000000000244', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-019', 'Post-Change Verification', 'All significant changes are verified after deployment. PCI DSS compliance confirmed post-change for CDE-affecting modifications.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-019'),
    ('c0000000-0000-0000-0000-000000000245', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-020', 'WAF/RASP Protection', 'Public-facing web applications protected by WAF. Runtime Application Self-Protection deployed for critical applications.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-020'),
    ('c0000000-0000-0000-0000-000000000246', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-021', 'Software Bill of Materials', 'SBOM generated for all applications tracking all components, versions, and licenses. Updated with each release.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-021'),
    ('c0000000-0000-0000-0000-000000000247', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-022', 'Payment Page Script Management', 'All scripts on payment pages are authorized, inventoried, and integrity-checked. Unauthorized scripts are blocked immediately.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-022'),
    ('c0000000-0000-0000-0000-000000000248', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-023', 'Vulnerability Disclosure Program', 'Responsible vulnerability disclosure program with documented procedures, response SLAs, and safe harbor provisions.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-023'),
    ('c0000000-0000-0000-0000-000000000249', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-024', 'Software Inventory', 'Complete inventory of bespoke and custom software maintained to facilitate vulnerability and patch management.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-SD-024'),
    ('c0000000-0000-0000-0000-000000000250', 'a0000000-0000-0000-0000-000000000001', 'CTRL-SD-025', 'Threat Vulnerability Management', 'Ongoing process to identify new threats and vulnerabilities for public-facing web applications and address them proactively.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-SD-025')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ---------------------------------------------------------------------------
-- VULNERABILITY MANAGEMENT (VM) — 25 controls
-- ---------------------------------------------------------------------------
INSERT INTO controls (id, org_id, identifier, title, description, category, status, owner_id, is_custom, source_template_id) VALUES
    ('c0000000-0000-0000-0000-000000000251', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-001', 'Vulnerability Scanning - Internal', 'Internal vulnerability scans performed at least quarterly and after significant changes. All critical/high findings remediated per SLA.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-001'),
    ('c0000000-0000-0000-0000-000000000252', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-002', 'Vulnerability Scanning - External', 'External vulnerability scans performed quarterly by ASV. Clean scans required for PCI compliance. Rescans for failures.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-002'),
    ('c0000000-0000-0000-0000-000000000253', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-003', 'Penetration Testing - External', 'Annual external penetration testing by qualified third party. Findings remediated and retested before considered resolved.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-003'),
    ('c0000000-0000-0000-0000-000000000254', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-004', 'Penetration Testing - Internal', 'Annual internal penetration testing covering network segmentation validation and lateral movement detection.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-004'),
    ('c0000000-0000-0000-0000-000000000255', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-005', 'Patch Management', 'Automated patch management for OS and applications. Critical patches within 24h, high within 7d, medium within 30d, low within 90d.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-VM-005'),
    ('c0000000-0000-0000-0000-000000000256', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-006', 'Vulnerability Management Program', 'Formal vulnerability management program covering identification, assessment, prioritization, remediation, and verification.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-006'),
    ('c0000000-0000-0000-0000-000000000257', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-007', 'Cloud Security Posture Management', 'CSPM tools continuously assess cloud infrastructure for misconfigurations, compliance violations, and security risks.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-VM-007'),
    ('c0000000-0000-0000-0000-000000000258', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-008', 'Container Vulnerability Scanning', 'Container images scanned for vulnerabilities before deployment and continuously in runtime. Images with critical vulns are blocked.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-VM-008'),
    ('c0000000-0000-0000-0000-000000000259', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-009', 'Vulnerability Prioritization', 'Vulnerabilities prioritized using CVSS + environmental factors (asset criticality, exposure, exploitability). Risk-based remediation.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-009'),
    ('c0000000-0000-0000-0000-000000000260', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-010', 'Red Team Exercises', 'Annual red team exercises simulating real-world attack scenarios. Scope includes social engineering, physical, and digital attack vectors.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-010'),
    ('c0000000-0000-0000-0000-000000000261', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-011', 'Bug Bounty Program', 'Public bug bounty program incentivizes external researchers to find and responsibly disclose vulnerabilities.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-011'),
    ('c0000000-0000-0000-0000-000000000262', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-012', 'Vulnerability Metrics', 'Monthly vulnerability metrics: open count by severity, MTTR, aging, scan coverage. Trends reported to CISO.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-012'),
    ('c0000000-0000-0000-0000-000000000263', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-013', 'Segmentation Validation', 'Network segmentation effectiveness validated at least annually through penetration testing. CDE isolation specifically verified.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-013'),
    ('c0000000-0000-0000-0000-000000000264', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-014', 'Wireless Network Scanning', 'Quarterly wireless scanning to detect rogue access points and unauthorized wireless devices near cardholder data environment.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-014'),
    ('c0000000-0000-0000-0000-000000000265', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-015', 'Remediation Verification', 'Remediated vulnerabilities are verified through rescan. Compensating controls documented when direct remediation is not feasible.', 'operational', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-015'),
    ('c0000000-0000-0000-0000-000000000266', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-016', 'Threat Intelligence Feeds', 'Multiple threat intelligence feeds provide indicators of compromise. Automated correlation with vulnerability data and asset inventory.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-016'),
    ('c0000000-0000-0000-0000-000000000267', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-017', 'API Security Testing', 'Dedicated API security testing covering authentication bypass, BOLA, injection, and excessive data exposure per OWASP API Top 10.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-017'),
    ('c0000000-0000-0000-0000-000000000268', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-018', 'Vulnerability Exception Process', 'Formal process for vulnerability exceptions with risk acceptance, compensating controls, and time-limited waivers.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-018'),
    ('c0000000-0000-0000-0000-000000000269', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-019', 'Asset Criticality Rating', 'All assets are rated for business criticality (Tier 1-4). Rating drives vulnerability remediation priority and SLAs.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000002', FALSE, 'TPL-VM-019'),
    ('c0000000-0000-0000-0000-000000000270', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-020', 'Configuration Compliance Scanning', 'Automated scans verify systems comply with CIS Benchmarks. Non-compliant configurations are remediated within 30 days.', 'technical', 'active', 'b0000000-0000-0000-0000-000000000005', FALSE, 'TPL-VM-020'),
    ('c0000000-0000-0000-0000-000000000271', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-021', 'Audit Testing Protection', 'Security controls protect information systems during audit testing. Audit tools are controlled and their use is authorized and logged.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000006', FALSE, 'TPL-VM-021'),
    ('c0000000-0000-0000-0000-000000000272', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-022', 'Cyber Risk Insurance', 'Cyber risk insurance policy covers data breaches, ransomware, business interruption, and third-party liability.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-VM-022'),
    ('c0000000-0000-0000-0000-000000000273', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-023', 'Security Benchmarking', 'Annual security benchmarking against industry peers and maturity frameworks (NIST CSF, CIS Controls).', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000001', FALSE, 'TPL-VM-023'),
    ('c0000000-0000-0000-0000-000000000274', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-024', 'Security Certification', 'Organization maintains relevant security certifications (SOC 2 Type II, ISO 27001) with annual renewal audits.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000006', FALSE, 'TPL-VM-024'),
    ('c0000000-0000-0000-0000-000000000275', 'a0000000-0000-0000-0000-000000000001', 'CTRL-VM-025', 'Cybersecurity Audit', 'Comprehensive cybersecurity audit performed for businesses processing significant personal data per CCPA/CPRA requirements.', 'administrative', 'active', 'b0000000-0000-0000-0000-000000000006', FALSE, 'TPL-VM-025')
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- SPRINT 2: CROSS-FRAMEWORK CONTROL MAPPINGS
-- This is the key value prop: one control → multiple framework requirements
-- ============================================================================

INSERT INTO control_mappings (org_id, control_id, requirement_id, strength, mapped_by, notes) VALUES
-- CTRL-AC-001 (MFA) → maps to 4 frameworks
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'r0100000-0000-0000-0000-000000000060', 'primary', 'b0000000-0000-0000-0000-000000000001', 'MFA directly addresses logical access security'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'r0200000-0000-0000-0000-000000000088', 'primary', 'b0000000-0000-0000-0000-000000000001', 'MFA implements secure authentication'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'r0300000-0000-0000-0000-000000000831', 'primary', 'b0000000-0000-0000-0000-000000000001', 'MFA for CDE administrative access'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'r0300000-0000-0000-0000-000000000832', 'primary', 'b0000000-0000-0000-0000-000000000001', 'MFA for all CDE access'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'r0300000-0000-0000-0000-000000000833', 'primary', 'b0000000-0000-0000-0000-000000000001', 'MFA for remote network access'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'r0400000-0000-0000-0000-000000000058', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'MFA supports security of processing'),

-- CTRL-AC-002 (RBAC) → maps to 4 frameworks
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002', 'r0100000-0000-0000-0000-000000000060', 'primary', 'b0000000-0000-0000-0000-000000000001', 'RBAC implements logical access security'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000024', 'primary', 'b0000000-0000-0000-0000-000000000001', 'RBAC implements access control'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002', 'r0200000-0000-0000-0000-000000000086', 'primary', 'b0000000-0000-0000-0000-000000000001', 'RBAC enforces information access restriction'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002', 'r0400000-0000-0000-0000-000000000058', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'RBAC supports security of processing'),

-- CTRL-AC-003 (Access Reviews) → SOC 2, ISO, PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000003', 'r0100000-0000-0000-0000-000000000062', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Reviews ensure access modification/removal'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000003', 'r0200000-0000-0000-0000-000000000027', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Reviews manage access rights lifecycle'),

-- CTRL-AC-006 (Password Policy) → PCI, SOC 2
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000006', 'r0300000-0000-0000-0000-000000000823', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Password complexity meets PCI requirements'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000006', 'r0300000-0000-0000-0000-000000000824', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Password rotation meets PCI requirements'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000006', 'r0200000-0000-0000-0000-000000000026', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Password policy implements authentication information control'),

-- CTRL-AC-012 (Unique User ID) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000012', 'r0300000-0000-0000-0000-000000000811', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Unique IDs assigned to all users'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000012', 'r0300000-0000-0000-0000-000000000812', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Prohibits shared/group accounts'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000012', 'r0200000-0000-0000-0000-000000000025', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Identity management'),

-- CTRL-AC-022 (Segregation of Duties) → ISO, SOC 2
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000022', 'r0200000-0000-0000-0000-000000000012', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Segregation of duties per ISO'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000022', 'r0100000-0000-0000-0000-000000000050', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Control activities for risk mitigation'),

-- CTRL-DP-002 (Encryption at Rest) → SOC 2, ISO, PCI, GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000057', 'r0100000-0000-0000-0000-000000000111', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Encryption maintains confidential information'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000057', 'r0200000-0000-0000-0000-000000000107', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Use of cryptography per ISO'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000057', 'r0400000-0000-0000-0000-000000000016', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Encryption ensures integrity and confidentiality of data'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000057', 'r0400000-0000-0000-0000-000000000058', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Encryption for security of processing'),

-- CTRL-DP-003 (Encryption in Transit) → same frameworks
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000058', 'r0100000-0000-0000-0000-000000000066', 'primary', 'b0000000-0000-0000-0000-000000000001', 'TLS protects data in transit'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000058', 'r0200000-0000-0000-0000-000000000107', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'TLS implements cryptography controls'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000058', 'r0300000-0000-0000-0000-000000000822', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Cryptography renders auth factors unreadable'),

-- CTRL-DP-016 (Data Subject Rights) → GDPR, CCPA
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000071', 'r0400000-0000-0000-0000-000000000033', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right of access'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000071', 'r0400000-0000-0000-0000-000000000034', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right to rectification'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000071', 'r0400000-0000-0000-0000-000000000035', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right to erasure'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000071', 'r0400000-0000-0000-0000-000000000038', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right to data portability'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000071', 'r0500000-0000-0000-0000-000000000010', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right to know (CCPA)'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000071', 'r0500000-0000-0000-0000-000000000012', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right to delete (CCPA)'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000071', 'r0500000-0000-0000-0000-000000000013', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right to correct (CCPA)'),

-- CTRL-DP-014 (Privacy Notice) → GDPR, CCPA
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000069', 'r0400000-0000-0000-0000-000000000030', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Transparent information and communication'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000069', 'r0400000-0000-0000-0000-000000000031', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Information provided at collection'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000069', 'r0100000-0000-0000-0000-000000000131', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'Privacy notice practices (SOC 2)'),

-- CTRL-DP-015 (Consent Management) → GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000070', 'r0400000-0000-0000-0000-000000000021', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Conditions for consent'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000070', 'r0400000-0000-0000-0000-000000000011', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'Lawfulness, fairness and transparency'),

-- CTRL-DP-017 (ROPA) → GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000072', 'r0400000-0000-0000-0000-000000000056', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Records of processing activities (Art 30)'),

-- CTRL-DP-018 (DPIA) → GDPR, CCPA
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000073', 'r0400000-0000-0000-0000-000000000061', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Data protection impact assessment'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000073', 'r0500000-0000-0000-0000-000000000122', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Risk assessment for significant risk processing'),

-- CTRL-DP-022 (Opt-Out) → CCPA
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000077', 'r0500000-0000-0000-0000-000000000060', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Right to opt-out of sale/sharing'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000077', 'r0500000-0000-0000-0000-000000000101', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Do Not Sell link on homepage'),

-- CTRL-IR-001 (IRP) → SOC 2, ISO, PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000081', 'r0100000-0000-0000-0000-000000000073', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 incident response program'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000081', 'r0200000-0000-0000-0000-000000000033', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO incident management planning'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000081', 'r0300000-0000-0000-0000-000000001241', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI incident response plan'),

-- CTRL-IR-005 (Breach Notification) → GDPR, CCPA
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000085', 'r0400000-0000-0000-0000-000000000059', 'primary', 'b0000000-0000-0000-0000-000000000001', 'GDPR breach notification to DPA'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000085', 'r0400000-0000-0000-0000-000000000060', 'primary', 'b0000000-0000-0000-0000-000000000001', 'GDPR breach notification to data subject'),

-- CTRL-LM-001 (Centralized Logging) → SOC 2, ISO, PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000101', 'r0100000-0000-0000-0000-000000000070', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 detection and monitoring'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000101', 'r0200000-0000-0000-0000-000000000098', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO logging control'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000101', 'r0300000-0000-0000-0000-000000001011', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI audit logs enabled and active'),

-- CTRL-LM-002 (Audit Log Integrity) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000102', 'r0300000-0000-0000-0000-000000001022', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Logs protected from modification'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000102', 'r0300000-0000-0000-0000-000000001023', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Logs backed up to central server'),

-- CTRL-LM-010 (Log Access Control) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000110', 'r0300000-0000-0000-0000-000000001021', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Read access to logs limited to need'),

-- CTRL-LM-013 (Daily Log Review) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000113', 'r0300000-0000-0000-0000-000000001031', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Security events reviewed daily'),

-- CTRL-LM-015 (IDS/IPS) → PCI, SOC 2
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000115', 'r0300000-0000-0000-0000-000000001131', 'primary', 'b0000000-0000-0000-0000-000000000001', 'IDS/IPS for network intrusion detection'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000115', 'r0100000-0000-0000-0000-000000000071', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 monitoring for anomalies'),

-- CTRL-LM-014 (FIM) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000114', 'r0300000-0000-0000-0000-000000001132', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Change detection for critical files'),

-- CTRL-LM-021 (Payment Page Monitoring) → PCI 11.6.1
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000121', 'r0300000-0000-0000-0000-000000001141', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Payment page tamper detection'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000121', 'r0300000-0000-0000-0000-000000000633', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Payment page script management'),

-- CTRL-NW-001 (Segmentation) → PCI, SOC 2, ISO
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000126', 'r0300000-0000-0000-0000-000000000131', 'primary', 'b0000000-0000-0000-0000-000000000001', 'NSCs between trusted/untrusted'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000126', 'r0100000-0000-0000-0000-000000000065', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 logical access at boundaries'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000126', 'r0200000-0000-0000-0000-000000000105', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO network segregation'),

-- CTRL-NW-007 (Inbound CDE) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000132', 'r0300000-0000-0000-0000-000000000121', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Inbound CDE traffic restriction'),

-- CTRL-NW-008 (Outbound CDE) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000133', 'r0300000-0000-0000-0000-000000000122', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Outbound CDE traffic restriction'),

-- CTRL-PM-001 (InfoSec Policy) → SOC 2, ISO, PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000146', 'r0100000-0000-0000-0000-000000000051', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 control activities via policies'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000146', 'r0200000-0000-0000-0000-000000000010', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO information security policies'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000146', 'r0300000-0000-0000-0000-000000001202', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI security policy established'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000146', 'r0300000-0000-0000-0000-000000001203', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI policy reviewed annually'),

-- CTRL-PM-019 (DPO) → GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000164', 'r0400000-0000-0000-0000-000000000063', 'primary', 'b0000000-0000-0000-0000-000000000001', 'DPO designation'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000164', 'r0400000-0000-0000-0000-000000000064', 'primary', 'b0000000-0000-0000-0000-000000000001', 'DPO position and resources'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000164', 'r0400000-0000-0000-0000-000000000065', 'primary', 'b0000000-0000-0000-0000-000000000001', 'DPO tasks defined'),

-- CTRL-PM-025 (Risk Assessment) → SOC 2, PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000170', 'r0100000-0000-0000-0000-000000000031', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 risk identification and analysis'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000170', 'r0300000-0000-0000-0000-000000001211', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI targeted risk analysis'),

-- CTRL-SA-001 (Awareness Training) → SOC 2, ISO, PCI, GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000206', 'r0100000-0000-0000-0000-000000000013', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 commitment to competent individuals'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000206', 'r0200000-0000-0000-0000-000000000062', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO security awareness training'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000206', 'r0300000-0000-0000-0000-000000001221', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI security awareness program'),

-- CTRL-SD-003 (SAST) + CTRL-SD-004 (DAST) → PCI, ISO
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000228', 'r0300000-0000-0000-0000-000000000613', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Code review for vulnerabilities'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000228', 'r0200000-0000-0000-0000-000000000112', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Security testing in dev and acceptance'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000229', 'r0300000-0000-0000-0000-000000000632', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Technical solution to detect web attacks'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000229', 'r0200000-0000-0000-0000-000000000112', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'Security testing in dev and acceptance'),

-- CTRL-SD-013 (Change Management) → PCI, SOC 2
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000238', 'r0100000-0000-0000-0000-000000000080', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 change authorization and management'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000238', 'r0300000-0000-0000-0000-000000000641', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI change control procedures'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000238', 'r0200000-0000-0000-0000-000000000115', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO change management'),

-- CTRL-SD-022 (Payment Page Scripts) → PCI 6.4.3
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000247', 'r0300000-0000-0000-0000-000000000633', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Payment page script management'),

-- CTRL-VM-001 (Internal Vuln Scan) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000251', 'r0300000-0000-0000-0000-000000001111', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Internal vulnerability scans quarterly'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000251', 'r0200000-0000-0000-0000-000000000091', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO tech vulnerability management'),

-- CTRL-VM-002 (External Vuln Scan) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000252', 'r0300000-0000-0000-0000-000000001112', 'primary', 'b0000000-0000-0000-0000-000000000001', 'External vulnerability scans quarterly'),

-- CTRL-VM-003 (External Pen Test) → PCI
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000253', 'r0300000-0000-0000-0000-000000001121', 'primary', 'b0000000-0000-0000-0000-000000000001', 'External pen testing annually'),

-- CTRL-VM-005 (Patching) → PCI, ISO, SOC 2
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000255', 'r0300000-0000-0000-0000-000000000623', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Install applicable patches'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000255', 'r0200000-0000-0000-0000-000000000091', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'ISO tech vulnerability management'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000255', 'r0100000-0000-0000-0000-000000000067', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 unauthorized software controls'),

-- CTRL-DP-013 (Cross-Border Transfers) → GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000068', 'r0400000-0000-0000-0000-000000000070', 'primary', 'b0000000-0000-0000-0000-000000000001', 'General transfer principles'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000068', 'r0400000-0000-0000-0000-000000000072', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Appropriate safeguards for transfers'),

-- CTRL-DP-024 (Privacy by Design) → GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000079', 'r0400000-0000-0000-0000-000000000051', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Data protection by design and default'),

-- CTRL-DP-021 (Data Minimization) → GDPR
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000076', 'r0400000-0000-0000-0000-000000000013', 'primary', 'b0000000-0000-0000-0000-000000000001', 'Data minimisation principle'),

-- CTRL-RA-004 (Vendor Risk Assessment) → ISO, SOC 2
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000189', 'r0200000-0000-0000-0000-000000000028', 'primary', 'b0000000-0000-0000-0000-000000000001', 'ISO supplier security'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000189', 'r0100000-0000-0000-0000-000000000091', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 vendor risk assessment'),

-- CTRL-RA-018 (Vendor Inventory) → PCI, ISO
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000203', 'r0300000-0000-0000-0000-000000001231', 'primary', 'b0000000-0000-0000-0000-000000000001', 'PCI third-party service provider list'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000203', 'r0200000-0000-0000-0000-000000000028', 'supporting', 'b0000000-0000-0000-0000-000000000001', 'ISO supplier relationships'),

-- CTRL-DP-006 (Retention) → GDPR, SOC 2
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000061', 'r0400000-0000-0000-0000-000000000015', 'primary', 'b0000000-0000-0000-0000-000000000001', 'GDPR storage limitation'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000061', 'r0100000-0000-0000-0000-000000000135', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 PI retention'),
('a0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000061', 'r0100000-0000-0000-0000-000000000136', 'primary', 'b0000000-0000-0000-0000-000000000001', 'SOC 2 PI disposal')
ON CONFLICT (org_id, control_id, requirement_id) DO NOTHING;

-- ============================================================================
-- SPRINT 2: SAMPLE REQUIREMENT SCOPING (demo: some PCI requirements out of scope)
-- ============================================================================

INSERT INTO requirement_scopes (org_id, requirement_id, in_scope, justification, scoped_by) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'r0300000-0000-0000-0000-000000000151',
     FALSE, 'Cloud-only company — no computing devices bridge trusted and untrusted networks simultaneously',
     'b0000000-0000-0000-0000-000000000001')
ON CONFLICT (org_id, requirement_id) DO NOTHING;

-- ============================================================================
-- SPRINT 2: AUDIT LOG ENTRIES (framework activations)
-- ============================================================================

INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'framework.activated', 'org_framework', 'd0000000-0000-0000-0000-000000000001',
     '{"framework": "soc2", "version": "2024", "target_date": "2026-06-30"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'framework.activated', 'org_framework', 'd0000000-0000-0000-0000-000000000002',
     '{"framework": "iso27001", "version": "2022", "target_date": "2026-09-30"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'framework.activated', 'org_framework', 'd0000000-0000-0000-0000-000000000003',
     '{"framework": "pci_dss", "version": "4.0.1", "target_date": "2026-12-31"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'framework.activated', 'org_framework', 'd0000000-0000-0000-0000-000000000004',
     '{"framework": "gdpr", "version": "2016"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'framework.activated', 'org_framework', 'd0000000-0000-0000-0000-000000000005',
     '{"framework": "ccpa", "version": "2023"}'::jsonb, '192.168.1.10'::inet)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 3: EVIDENCE ARTIFACTS
-- ============================================================================

INSERT INTO evidence_artifacts (
    id, org_id, title, description, evidence_type, status, collection_method,
    file_name, file_size, mime_type, object_key, version, is_current,
    collection_date, expires_at, freshness_period_days, source_system,
    uploaded_by, tags
) VALUES
    -- 1. Okta MFA Configuration Export
    (
        'e0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'Okta MFA Configuration Export',
        'Export of MFA policy settings from Okta showing enforcement for all users in the production org.',
        'configuration_export', 'approved', 'system_export',
        'okta-mfa-config-2026-02.json', 15234, 'application/json',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000001/1/okta-mfa-config-2026-02.json',
        1, TRUE,
        '2026-02-15', '2026-05-15 00:00:00+00', 90, 'okta',
        'b0000000-0000-0000-0000-000000000002',
        ARRAY['mfa', 'okta', 'access-control', 'q1-2026']
    ),
    -- 2. Quarterly Access Review Report
    (
        'e0000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'Quarterly Access Review Report - Q1 2026',
        'Complete access review report covering all production systems. Reviews conducted by department managers.',
        'access_list', 'approved', 'manual_upload',
        'access-review-q1-2026.pdf', 2456789, 'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000002/1/access-review-q1-2026.pdf',
        1, TRUE,
        '2026-01-31', '2026-04-30 00:00:00+00', 90, 'manual',
        'b0000000-0000-0000-0000-000000000001',
        ARRAY['access-review', 'quarterly', 'q1-2026']
    ),
    -- 3. Vulnerability Scan Results (pending review)
    (
        'e0000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'Vulnerability Scan Results - February 2026',
        'Qualys vulnerability scan of production environment. No critical or high findings.',
        'vulnerability_report', 'pending_review', 'automated_pull',
        'qualys-scan-feb-2026.pdf', 892345, 'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000003/1/qualys-scan-feb-2026.pdf',
        1, TRUE,
        '2026-02-18', '2026-03-18 00:00:00+00', 30, 'qualys',
        'b0000000-0000-0000-0000-000000000002',
        ARRAY['vulnerability', 'qualys', 'scan', 'feb-2026']
    ),
    -- 4. Information Security Policy
    (
        'e0000000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'Information Security Policy v3.1',
        'Current information security policy document. Approved by CISO, signed by all employees.',
        'policy_document', 'approved', 'manual_upload',
        'infosec-policy-v3.1.pdf', 345678, 'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000004/1/infosec-policy-v3.1.pdf',
        1, TRUE,
        '2026-01-15', '2027-01-15 00:00:00+00', 365, 'manual',
        'b0000000-0000-0000-0000-000000000001',
        ARRAY['policy', 'infosec', 'annual']
    ),
    -- 5. AWS CloudTrail Logging Config
    (
        'e0000000-0000-0000-0000-000000000005',
        'a0000000-0000-0000-0000-000000000001',
        'AWS CloudTrail Logging Configuration',
        'Screenshot and config export showing CloudTrail enabled in all regions with S3 logging.',
        'screenshot', 'approved', 'screenshot_capture',
        'aws-cloudtrail-config.png', 456789, 'image/png',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000005/1/aws-cloudtrail-config.png',
        1, TRUE,
        '2026-02-10', '2026-05-10 00:00:00+00', 90, 'aws',
        'b0000000-0000-0000-0000-000000000005',
        ARRAY['aws', 'cloudtrail', 'logging', 'monitoring']
    ),
    -- 6. Security Awareness Training Completion
    (
        'e0000000-0000-0000-0000-000000000006',
        'a0000000-0000-0000-0000-000000000001',
        'Security Awareness Training Completion - 2026',
        'KnowBe4 training completion report showing 100% employee participation.',
        'training_record', 'approved', 'automated_pull',
        'knowbe4-training-2026.csv', 89012, 'text/csv',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000006/1/knowbe4-training-2026.csv',
        1, TRUE,
        '2026-02-01', '2027-02-01 00:00:00+00', 365, 'knowbe4',
        'b0000000-0000-0000-0000-000000000005',
        ARRAY['training', 'knowbe4', 'annual', '2026']
    ),
    -- 7. Penetration Test Report (draft)
    (
        'e0000000-0000-0000-0000-000000000007',
        'a0000000-0000-0000-0000-000000000001',
        'Annual Penetration Test Report - 2026',
        'External penetration test conducted by SecureTech Inc. 2 medium findings, 0 critical/high.',
        'penetration_test', 'draft', 'manual_upload',
        'pentest-report-2026.pdf', 1234567, 'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000007/1/pentest-report-2026.pdf',
        1, TRUE,
        '2026-02-12', '2027-02-12 00:00:00+00', 365, 'manual',
        'b0000000-0000-0000-0000-000000000002',
        ARRAY['pentest', 'annual', '2026']
    ),
    -- 8. SOC 2 Audit Report (previous year — for reference)
    (
        'e0000000-0000-0000-0000-000000000008',
        'a0000000-0000-0000-0000-000000000001',
        'SOC 2 Type II Audit Report - 2025',
        'Previous year SOC 2 Type II audit report from Deloitte. Clean opinion.',
        'audit_report', 'approved', 'manual_upload',
        'soc2-audit-2025.pdf', 3456789, 'application/pdf',
        'a0000000-0000-0000-0000-000000000001/e0000000-0000-0000-0000-000000000008/1/soc2-audit-2025.pdf',
        1, TRUE,
        '2025-12-15', '2026-12-15 00:00:00+00', 365, 'manual',
        'b0000000-0000-0000-0000-000000000006',
        ARRAY['soc2', 'audit', 'annual', '2025']
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 3: EVIDENCE LINKS (link evidence to controls)
-- ============================================================================

INSERT INTO evidence_links (id, org_id, artifact_id, target_type, control_id, strength, notes, linked_by) VALUES
    -- MFA config → CTRL-AC-001 (Access Control: MFA)
    (
        'l0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000001',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-001' LIMIT 1),
        'primary',
        'Okta MFA policy export directly demonstrates enforcement of multi-factor authentication for all users.',
        'b0000000-0000-0000-0000-000000000002'
    ),
    -- Access review → CTRL-AC-003 (Access Review)
    (
        'l0000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000002',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-003' LIMIT 1),
        'primary',
        'Quarterly access review report covers all production systems with department manager sign-off.',
        'b0000000-0000-0000-0000-000000000001'
    ),
    -- CloudTrail config → CTRL-LM-001 (Logging & Monitoring)
    (
        'l0000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000005',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-LM-001' LIMIT 1),
        'primary',
        'CloudTrail configuration screenshot shows all-region logging enabled with S3 archival.',
        'b0000000-0000-0000-0000-000000000005'
    ),
    -- Training completion → CTRL-SA-001 (Security Awareness Training)
    (
        'l0000000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000006',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-SA-001' LIMIT 1),
        'primary',
        'KnowBe4 training completion report shows 100% participation for annual security awareness training.',
        'b0000000-0000-0000-0000-000000000005'
    ),
    -- InfoSec policy → CTRL-PM-001 (Policy Management) as supporting evidence
    (
        'l0000000-0000-0000-0000-000000000005',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000004',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-PM-001' LIMIT 1),
        'primary',
        'Current information security policy document, CISO-approved, covering all organizational security requirements.',
        'b0000000-0000-0000-0000-000000000001'
    ),
    -- Vuln scan → CTRL-VM-001 (Vulnerability Management)
    (
        'l0000000-0000-0000-0000-000000000006',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000003',
        'control',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-VM-001' LIMIT 1),
        'primary',
        'Qualys vulnerability scan results for February 2026 production environment assessment.',
        'b0000000-0000-0000-0000-000000000002'
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 3: EVIDENCE EVALUATIONS
-- ============================================================================

INSERT INTO evidence_evaluations (id, org_id, artifact_id, evidence_link_id, verdict, confidence, comments, evaluated_by) VALUES
    -- MFA config evaluation: sufficient
    (
        'ev000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000001',
        'l0000000-0000-0000-0000-000000000001',
        'sufficient', 'high',
        'MFA is enforced for all user types. Configuration export shows Okta MFA policy is set to "Always" with no exceptions. Meets PCI DSS 8.3 and SOC 2 CC6.1 requirements.',
        'b0000000-0000-0000-0000-000000000001'
    ),
    -- Access review evaluation: sufficient
    (
        'ev000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000002',
        'l0000000-0000-0000-0000-000000000002',
        'sufficient', 'high',
        'Access review covers all production systems. Department managers signed off on all entries. Stale accounts identified and removed. Meets quarterly cadence requirement.',
        'b0000000-0000-0000-0000-000000000004'
    ),
    -- Training completion evaluation: sufficient
    (
        'ev000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000006',
        'l0000000-0000-0000-0000-000000000004',
        'sufficient', 'high',
        '100% employee participation confirmed. Training content covers phishing, social engineering, data handling, and incident reporting. Annual cadence meets SOC 2 and ISO 27001 requirements.',
        'b0000000-0000-0000-0000-000000000001'
    ),
    -- CloudTrail evaluation: sufficient
    (
        'ev000000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000005',
        'l0000000-0000-0000-0000-000000000003',
        'sufficient', 'medium',
        'CloudTrail is enabled in all regions. S3 logging confirmed. However, screenshot alone does not show log retention policy — recommend supplementing with AWS Config export.',
        'b0000000-0000-0000-0000-000000000001'
    ),
    -- Pentest report evaluation: needs_update (still in draft)
    (
        'ev000000-0000-0000-0000-000000000005',
        'a0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000007',
        NULL,
        'needs_update', 'low',
        'Report is still in draft status. Needs final version with remediation evidence for the 2 medium findings before it can be submitted for compliance.',
        'b0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 3: AUDIT LOG ENTRIES (evidence activity)
-- ============================================================================

INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'evidence.uploaded', 'evidence_artifact', 'e0000000-0000-0000-0000-000000000001',
     '{"title": "Okta MFA Configuration Export", "type": "configuration_export", "file_size": 15234}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'evidence.uploaded', 'evidence_artifact', 'e0000000-0000-0000-0000-000000000002',
     '{"title": "Quarterly Access Review Report - Q1 2026", "type": "access_list", "file_size": 2456789}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'evidence.linked', 'evidence_link', 'l0000000-0000-0000-0000-000000000001',
     '{"artifact_title": "Okta MFA Configuration Export", "target_type": "control", "control": "CTRL-AC-001"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'evidence.evaluated', 'evidence_evaluation', 'ev000000-0000-0000-0000-000000000001',
     '{"artifact_title": "Okta MFA Configuration Export", "verdict": "sufficient", "confidence": "high"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'evidence.evaluated', 'evidence_evaluation', 'ev000000-0000-0000-0000-000000000002',
     '{"artifact_title": "Quarterly Access Review Report", "verdict": "sufficient", "confidence": "high"}'::jsonb, '192.168.1.10'::inet)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 4: TEST DEFINITIONS
-- ============================================================================

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
        '0 * * * *',
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
        '0 8 * * 1',
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
        '0 */2 * * *',
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
        '0 */4 * * *',
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
        '0 9 * * *',
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
        '0 * * * *',
        120,
        '{"provider": "aws", "service": "ec2", "check": "no_open_ingress", "restricted_ports": [22, 3389, 3306, 5432]}'::JSONB,
        ARRAY['network', 'aws', 'firewall', 'pci'],
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1)
    ),
    -- Endpoint compliance check (mapped to CTRL-SA-001 since CTRL-EP doesn't exist)
    (
        't0000000-0000-0000-0000-000000000007',
        'a0000000-0000-0000-0000-000000000001',
        'TST-EP-001',
        'Endpoint Disk Encryption',
        'Verifies that all managed endpoints have full-disk encryption enabled.',
        'endpoint',
        'high',
        'active',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-SA-001' LIMIT 1),
        '0 */6 * * *',
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
        '0 8 * * *',
        60,
        '{"provider": "okta", "check": "password_policy", "min_length": 12, "require_uppercase": true, "require_number": true, "require_special": true}'::JSONB,
        ARRAY['configuration', 'password', 'okta', 'pci'],
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1)
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 4: SAMPLE TEST RUN (completed sweep)
-- ============================================================================

INSERT INTO test_runs (
    id, org_id, status, trigger_type, started_at, completed_at, duration_ms,
    total_tests, passed, failed, errors, skipped, warnings,
    triggered_by, worker_id
) VALUES
    (
        'tr000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'completed',
        'scheduled',
        '2026-02-20 16:00:00+00',
        '2026-02-20 16:03:42+00',
        222000,
        8, 5, 2, 1, 0, 0,
        NULL,
        'worker-1'
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 4: SAMPLE TEST RESULTS
-- ============================================================================

INSERT INTO test_results (
    id, org_id, test_run_id, test_id, control_id, status, severity,
    message, details, started_at, completed_at, duration_ms, alert_generated
) VALUES
    -- TST-AC-001: PASS
    (
        'res00000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000001',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-001' LIMIT 1),
        'pass', 'critical',
        'MFA is enforced for all 47 users. No exceptions found.',
        '{"total_users": 47, "mfa_enabled": 47, "exceptions": 0}'::JSONB,
        '2026-02-20 16:00:00+00', '2026-02-20 16:00:12+00', 12000, FALSE
    ),
    -- TST-AC-002: PASS
    (
        'res00000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000002',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-003' LIMIT 1),
        'pass', 'high',
        'Quarterly access review completed 12 days ago. Within 90-day cadence.',
        '{"last_review_date": "2026-02-08", "days_since_review": 12, "cadence_days": 90}'::JSONB,
        '2026-02-20 16:00:12+00', '2026-02-20 16:00:18+00', 6000, FALSE
    ),
    -- TST-DP-001: FAIL — 2 unencrypted S3 buckets
    (
        'res00000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000003',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-DP-001' LIMIT 1),
        'fail', 'critical',
        '2 of 15 S3 buckets lack default encryption. Non-compliant with PCI DSS 3.4.',
        '{"total_buckets": 15, "encrypted": 13, "unencrypted": ["staging-logs-2026", "temp-upload-buffer"], "expected_algorithms": ["AES256", "aws:kms"]}'::JSONB,
        '2026-02-20 16:00:18+00', '2026-02-20 16:00:45+00', 27000, TRUE
    ),
    -- TST-LM-001: PASS
    (
        'res00000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000004',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-LM-001' LIMIT 1),
        'pass', 'high',
        'CloudTrail enabled in all 4 active regions. S3 log delivery confirmed.',
        '{"regions_checked": 4, "regions_enabled": 4, "s3_delivery": true, "trail_name": "org-main-trail"}'::JSONB,
        '2026-02-20 16:00:45+00', '2026-02-20 16:01:02+00', 17000, FALSE
    ),
    -- TST-VM-001: PASS
    (
        'res00000-0000-0000-0000-000000000005',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000005',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-VM-001' LIMIT 1),
        'pass', 'high',
        'Last vulnerability scan completed 5 days ago. Within 30-day cadence.',
        '{"last_scan_date": "2026-02-15", "days_since_scan": 5, "cadence_days": 30, "scanner": "qualys"}'::JSONB,
        '2026-02-20 16:01:02+00', '2026-02-20 16:01:10+00', 8000, FALSE
    ),
    -- TST-NW-001: FAIL — open security group found
    (
        'res00000-0000-0000-0000-000000000006',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000006',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-NW-001' LIMIT 1),
        'fail', 'critical',
        '1 security group allows inbound 0.0.0.0/0 on port 22 (SSH). Immediate remediation required.',
        '{"total_security_groups": 23, "violations": [{"sg_id": "sg-0abc123def456", "sg_name": "dev-bastion-sg", "port": 22, "source": "0.0.0.0/0"}]}'::JSONB,
        '2026-02-20 16:01:10+00', '2026-02-20 16:01:35+00', 25000, TRUE
    ),
    -- TST-EP-001: ERROR — integration timeout
    (
        'res00000-0000-0000-0000-000000000007',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000007',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-SA-001' LIMIT 1),
        'error', 'high',
        'Jamf API connection timed out after 180 seconds.',
        '{"provider": "jamf", "error_type": "timeout", "timeout_seconds": 180}'::JSONB,
        '2026-02-20 16:01:35+00', '2026-02-20 16:04:35+00', 180000, FALSE
    ),
    -- TST-CFG-001: PASS
    (
        'res00000-0000-0000-0000-000000000008',
        'a0000000-0000-0000-0000-000000000001',
        'tr000000-0000-0000-0000-000000000001',
        't0000000-0000-0000-0000-000000000008',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-AC-002' LIMIT 1),
        'pass', 'medium',
        'Password policy meets all minimum complexity requirements.',
        '{"min_length": 12, "actual_min_length": 14, "uppercase": true, "numbers": true, "special_chars": true}'::JSONB,
        '2026-02-20 16:04:35+00', '2026-02-20 16:04:42+00', 7000, FALSE
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 4: ALERT RULES
-- ============================================================================

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
        NULL,
        ARRAY['critical']::test_severity[],
        ARRAY['fail']::test_result_status[],
        1, 60,
        'critical', 4,
        ARRAY['slack', 'email', 'in_app']::alert_delivery_channel[],
        ARRAY['security@acme.example.com', 'ciso@acme.example.com'],
        10,
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
        2, 120,
        'high', 24,
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
        3, 360,
        'medium', 72,
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
        NULL,
        ARRAY['error']::test_result_status[],
        3, 240,
        'high', 24,
        ARRAY['email', 'in_app']::alert_delivery_channel[],
        ARRAY['devops@acme.example.com'],
        15,
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1)
    )
ON CONFLICT DO NOTHING;

-- ============================================================================
-- SPRINT 4: SAMPLE ALERTS (from the test run above)
-- ============================================================================

INSERT INTO alerts (
    id, org_id, title, description, severity, status,
    test_id, test_result_id, control_id, alert_rule_id,
    assigned_to, assigned_at, assigned_by,
    sla_deadline, sla_breached,
    delivery_channels, tags
) VALUES
    -- Alert from S3 encryption failure (TST-DP-001)
    (
        'alt00000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'Encryption at Rest — S3 Buckets FAILED',
        '2 of 15 S3 buckets lack default encryption. Non-compliant with PCI DSS 3.4. Buckets: staging-logs-2026, temp-upload-buffer.',
        'critical',
        'open',
        't0000000-0000-0000-0000-000000000003',
        'res00000-0000-0000-0000-000000000003',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-DP-001' LIMIT 1),
        'ar000000-0000-0000-0000-000000000001',
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1),
        '2026-02-20 16:00:45+00',
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
        '2026-02-20 20:00:45+00',
        FALSE,
        ARRAY['slack', 'email', 'in_app']::alert_delivery_channel[],
        ARRAY['encryption', 'aws', 's3', 'pci', 'critical']
    ),
    -- Alert from open security group (TST-NW-001)
    (
        'alt00000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'Security Group — No Open Inbound 0.0.0.0/0 FAILED',
        '1 security group (dev-bastion-sg) allows unrestricted SSH access from 0.0.0.0/0. Immediate remediation required.',
        'critical',
        'acknowledged',
        't0000000-0000-0000-0000-000000000006',
        'res00000-0000-0000-0000-000000000006',
        (SELECT id FROM controls WHERE org_id = 'a0000000-0000-0000-0000-000000000001' AND identifier = 'CTRL-NW-001' LIMIT 1),
        'ar000000-0000-0000-0000-000000000001',
        (SELECT id FROM users WHERE email = 'devops@acme.example.com' LIMIT 1),
        '2026-02-20 16:01:35+00',
        (SELECT id FROM users WHERE email = 'security@acme.example.com' LIMIT 1),
        '2026-02-20 20:01:35+00',
        FALSE,
        ARRAY['slack', 'email', 'in_app']::alert_delivery_channel[],
        ARRAY['network', 'aws', 'firewall', 'ssh', 'critical']
    )
ON CONFLICT DO NOTHING;

-- Link test results back to their alerts
UPDATE test_results SET alert_id = 'alt00000-0000-0000-0000-000000000001'
WHERE id = 'res00000-0000-0000-0000-000000000003';

UPDATE test_results SET alert_id = 'alt00000-0000-0000-0000-000000000002'
WHERE id = 'res00000-0000-0000-0000-000000000006';

-- ============================================================================
-- SPRINT 4: AUDIT LOG ENTRIES (test and alert activity)
-- ============================================================================

INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'test.created', 'test', 't0000000-0000-0000-0000-000000000001',
     '{"identifier": "TST-AC-001", "title": "MFA Enforcement Verification", "type": "access_control", "severity": "critical"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', NULL,
     'test_run.started', 'test_run', 'tr000000-0000-0000-0000-000000000001',
     '{"trigger": "scheduled", "total_tests": 8, "worker": "worker-1"}'::jsonb, '10.0.1.50'::inet),
    ('a0000000-0000-0000-0000-000000000001', NULL,
     'test_run.completed', 'test_run', 'tr000000-0000-0000-0000-000000000001',
     '{"passed": 5, "failed": 2, "errors": 1, "duration_ms": 222000}'::jsonb, '10.0.1.50'::inet),
    ('a0000000-0000-0000-0000-000000000001', NULL,
     'alert.created', 'alert', 'alt00000-0000-0000-0000-000000000001',
     '{"title": "Encryption at Rest — S3 Buckets FAILED", "severity": "critical", "rule": "Critical Test Failures"}'::jsonb, '10.0.1.50'::inet),
    ('a0000000-0000-0000-0000-000000000001', NULL,
     'alert.created', 'alert', 'alt00000-0000-0000-0000-000000000002',
     '{"title": "Security Group — No Open Inbound 0.0.0.0/0 FAILED", "severity": "critical", "rule": "Critical Test Failures"}'::jsonb, '10.0.1.50'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'alert.assigned', 'alert', 'alt00000-0000-0000-0000-000000000001',
     '{"assigned_to": "devops@acme.example.com", "severity": "critical"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000005',
     'alert.acknowledged', 'alert', 'alt00000-0000-0000-0000-000000000002',
     '{"title": "Security Group — No Open Inbound 0.0.0.0/0 FAILED", "previous_status": "open"}'::jsonb, '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'alert_rule.created', 'alert_rule', 'ar000000-0000-0000-0000-000000000001',
     '{"name": "Critical Test Failures", "severity": "critical", "sla_hours": 4}'::jsonb, '192.168.1.10'::inet)
ON CONFLICT DO NOTHING;
