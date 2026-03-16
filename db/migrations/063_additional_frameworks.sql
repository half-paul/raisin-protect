-- Migration 063: Additional Frameworks — HIPAA, NIST CSF 2.0, CIS Controls v8, SOX
-- Idempotent: uses ON CONFLICT DO NOTHING

-- ============================================================================
-- FRAMEWORKS
-- ============================================================================

INSERT INTO frameworks (id, identifier, name, description, category, website_url) VALUES
    ('f0000000-0000-0000-0000-000000000006', 'hipaa', 'HIPAA',
     'Health Insurance Portability and Accountability Act — US federal law for protecting sensitive patient health information.',
     'industry', 'https://www.hhs.gov/hipaa/'),
    ('f0000000-0000-0000-0000-000000000007', 'nist_csf', 'NIST CSF',
     'NIST Cybersecurity Framework 2.0 — voluntary guidance for managing cybersecurity risk based on existing standards.',
     'security_privacy', 'https://www.nist.gov/cyberframework'),
    ('f0000000-0000-0000-0000-000000000008', 'cis_controls', 'CIS Controls',
     'Center for Internet Security Controls v8 — prioritized set of safeguards to mitigate the most prevalent cyber attacks.',
     'security_privacy', 'https://www.cisecurity.org/controls'),
    ('f0000000-0000-0000-0000-000000000009', 'sox', 'SOX',
     'Sarbanes-Oxley Act — US federal law mandating financial reporting controls and IT controls supporting financial systems.',
     'industry', 'https://www.sec.gov/spotlight/sarbanes-oxley.htm')
ON CONFLICT (identifier) DO NOTHING;

-- ============================================================================
-- FRAMEWORK VERSIONS
-- ============================================================================

INSERT INTO framework_versions (id, framework_id, version, display_name, status, effective_date, total_requirements) VALUES
    ('a1000000-0000-0000-0000-000000000007', 'f0000000-0000-0000-0000-000000000006', '2013', 'HIPAA Security Rule (2013)', 'active', '2013-01-25', 54),
    ('a1000000-0000-0000-0000-000000000008', 'f0000000-0000-0000-0000-000000000007', '2.0', 'NIST CSF 2.0', 'active', '2024-02-26', 106),
    ('a1000000-0000-0000-0000-000000000009', 'f0000000-0000-0000-0000-000000000008', '8', 'CIS Controls v8', 'active', '2021-05-18', 153),
    ('a1000000-0000-0000-0000-000000000010', 'f0000000-0000-0000-0000-000000000009', '2002', 'SOX (2002)', 'active', '2002-07-30', 40)
ON CONFLICT (framework_id, version) DO NOTHING;

-- ============================================================================
-- HIPAA Security Rule Requirements
-- ============================================================================

-- HIPAA — Top-level safeguard categories
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0700000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000007', NULL, '164.308', 'Administrative Safeguards', 1, 0, FALSE),
    ('e0700000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000007', NULL, '164.310', 'Physical Safeguards', 2, 0, FALSE),
    ('e0700000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000007', NULL, '164.312', 'Technical Safeguards', 3, 0, FALSE),
    ('e0700000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000007', NULL, '164.314', 'Organizational Requirements', 4, 0, FALSE),
    ('e0700000-0000-0000-0000-000000000005', 'a1000000-0000-0000-0000-000000000007', NULL, '164.316', 'Policies and Procedures and Documentation', 5, 0, FALSE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- HIPAA — Administrative Safeguards (164.308)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0700000-0000-0000-0000-000000000010', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(1)', 'Security Management Process', 1, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000011', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000010', '164.308(a)(1)(i)', 'Risk Analysis (Required)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000012', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000010', '164.308(a)(1)(ii)(A)', 'Risk Management (Required)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000013', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000010', '164.308(a)(1)(ii)(B)', 'Sanction Policy (Required)', 3, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000014', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000010', '164.308(a)(1)(ii)(D)', 'Information System Activity Review (Required)', 4, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000020', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(2)', 'Assigned Security Responsibility (Required)', 2, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000030', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(3)', 'Workforce Security', 3, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000031', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000030', '164.308(a)(3)(ii)(A)', 'Authorization and/or Supervision (Addressable)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000032', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000030', '164.308(a)(3)(ii)(B)', 'Workforce Clearance Procedure (Addressable)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000033', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000030', '164.308(a)(3)(ii)(C)', 'Termination Procedures (Addressable)', 3, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000040', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(4)', 'Information Access Management', 4, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000041', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000040', '164.308(a)(4)(ii)(B)', 'Access Authorization (Addressable)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000042', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000040', '164.308(a)(4)(ii)(C)', 'Access Establishment and Modification (Addressable)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000050', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(5)', 'Security Awareness and Training', 5, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000051', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000050', '164.308(a)(5)(ii)(A)', 'Security Reminders (Addressable)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000052', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000050', '164.308(a)(5)(ii)(B)', 'Protection from Malicious Software (Addressable)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000053', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000050', '164.308(a)(5)(ii)(C)', 'Log-in Monitoring (Addressable)', 3, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000054', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000050', '164.308(a)(5)(ii)(D)', 'Password Management (Addressable)', 4, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000060', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(6)', 'Security Incident Procedures', 6, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000061', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000060', '164.308(a)(6)(ii)', 'Response and Reporting (Required)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000070', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(7)', 'Contingency Plan', 7, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000071', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000070', '164.308(a)(7)(ii)(A)', 'Data Backup Plan (Required)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000072', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000070', '164.308(a)(7)(ii)(B)', 'Disaster Recovery Plan (Required)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000073', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000070', '164.308(a)(7)(ii)(C)', 'Emergency Mode Operation Plan (Required)', 3, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000074', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000070', '164.308(a)(7)(ii)(D)', 'Testing and Revision Procedures (Addressable)', 4, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000080', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(a)(8)', 'Evaluation (Required)', 8, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000090', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000001', '164.308(b)(1)', 'Business Associate Contracts and Other Arrangements', 9, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- HIPAA — Physical Safeguards (164.310)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0700000-0000-0000-0000-000000000100', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000002', '164.310(a)', 'Facility Access Controls', 1, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000101', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000100', '164.310(a)(2)(i)', 'Contingency Operations (Addressable)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000102', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000100', '164.310(a)(2)(ii)', 'Facility Security Plan (Addressable)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000103', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000100', '164.310(a)(2)(iii)', 'Access Control and Validation Procedures (Addressable)', 3, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000104', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000100', '164.310(a)(2)(iv)', 'Maintenance Records (Addressable)', 4, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000110', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000002', '164.310(b)', 'Workstation Use (Required)', 2, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000120', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000002', '164.310(c)', 'Workstation Security (Required)', 3, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000130', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000002', '164.310(d)', 'Device and Media Controls', 4, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000131', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000130', '164.310(d)(2)(i)', 'Disposal (Required)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000132', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000130', '164.310(d)(2)(ii)', 'Media Re-use (Required)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000133', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000130', '164.310(d)(2)(iii)', 'Accountability (Addressable)', 3, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000134', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000130', '164.310(d)(2)(iv)', 'Data Backup and Storage (Addressable)', 4, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- HIPAA — Technical Safeguards (164.312)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0700000-0000-0000-0000-000000000200', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000003', '164.312(a)', 'Access Control', 1, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000201', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000200', '164.312(a)(2)(i)', 'Unique User Identification (Required)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000202', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000200', '164.312(a)(2)(ii)', 'Emergency Access Procedure (Required)', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000203', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000200', '164.312(a)(2)(iii)', 'Automatic Logoff (Addressable)', 3, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000204', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000200', '164.312(a)(2)(iv)', 'Encryption and Decryption (Addressable)', 4, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000210', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000003', '164.312(b)', 'Audit Controls (Required)', 2, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000220', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000003', '164.312(c)', 'Integrity', 3, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000221', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000220', '164.312(c)(2)', 'Mechanism to Authenticate Electronic PHI (Addressable)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000230', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000003', '164.312(d)', 'Person or Entity Authentication (Required)', 4, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000240', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000003', '164.312(e)', 'Transmission Security', 5, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000241', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000240', '164.312(e)(2)(i)', 'Integrity Controls (Addressable)', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000242', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000240', '164.312(e)(2)(ii)', 'Encryption (Addressable)', 2, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- HIPAA — Organizational Requirements (164.314) and Documentation (164.316)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0700000-0000-0000-0000-000000000300', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000004', '164.314(a)', 'Business Associate Contracts or Other Arrangements (Required)', 1, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000310', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000004', '164.314(b)', 'Requirements for Group Health Plans', 2, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000400', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000005', '164.316(a)', 'Policies and Procedures (Required)', 1, 1, TRUE),
    ('e0700000-0000-0000-0000-000000000410', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000005', '164.316(b)', 'Documentation', 2, 1, FALSE),
    ('e0700000-0000-0000-0000-000000000411', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000410', '164.316(b)(2)(i)', 'Time Limit (Required) — Retain documentation for 6 years', 1, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000412', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000410', '164.316(b)(2)(ii)', 'Availability (Required) — Make documentation available to those responsible', 2, 2, TRUE),
    ('e0700000-0000-0000-0000-000000000413', 'a1000000-0000-0000-0000-000000000007', 'e0700000-0000-0000-0000-000000000410', '164.316(b)(2)(iii)', 'Updates (Required) — Review and update documentation periodically', 3, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- NIST CSF 2.0 Requirements
-- ============================================================================

-- NIST CSF — 6 Core Functions
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0800000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000008', NULL, 'GV', 'Govern', 1, 0, FALSE),
    ('e0800000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000008', NULL, 'ID', 'Identify', 2, 0, FALSE),
    ('e0800000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000008', NULL, 'PR', 'Protect', 3, 0, FALSE),
    ('e0800000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000008', NULL, 'DE', 'Detect', 4, 0, FALSE),
    ('e0800000-0000-0000-0000-000000000005', 'a1000000-0000-0000-0000-000000000008', NULL, 'RS', 'Respond', 5, 0, FALSE),
    ('e0800000-0000-0000-0000-000000000006', 'a1000000-0000-0000-0000-000000000008', NULL, 'RC', 'Recover', 6, 0, FALSE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- NIST CSF — Govern Categories
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0800000-0000-0000-0000-000000000010', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000001', 'GV.OC', 'Organizational Context', 1, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000011', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000010', 'GV.OC-01', 'The organizational mission is understood and informs cybersecurity risk management', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000012', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000010', 'GV.OC-02', 'Internal and external stakeholders are understood, and their needs and expectations are addressed', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000013', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000010', 'GV.OC-03', 'Legal, regulatory, and contractual requirements regarding cybersecurity are understood and managed', 3, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000020', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000001', 'GV.RM', 'Risk Management Strategy', 2, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000021', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000020', 'GV.RM-01', 'Risk management objectives are established and agreed to by organizational stakeholders', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000022', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000020', 'GV.RM-02', 'Risk appetite and risk tolerance statements are established', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000023', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000020', 'GV.RM-03', 'Cybersecurity risk management activities and outcomes are included in enterprise risk management', 3, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000030', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000001', 'GV.RR', 'Roles, Responsibilities, and Authorities', 3, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000031', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000030', 'GV.RR-01', 'Organizational leadership is responsible and accountable for cybersecurity risk', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000032', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000030', 'GV.RR-02', 'Roles, responsibilities, and authorities related to cybersecurity are established and communicated', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000040', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000001', 'GV.PO', 'Policy', 4, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000041', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000040', 'GV.PO-01', 'Policy for managing cybersecurity risks is established based on organizational context and strategy', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000042', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000040', 'GV.PO-02', 'Policy is reviewed, updated, communicated, and enforced', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000050', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000001', 'GV.SC', 'Cybersecurity Supply Chain Risk Management', 5, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000051', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000050', 'GV.SC-01', 'A cybersecurity supply chain risk management program is established', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000052', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000050', 'GV.SC-03', 'Cybersecurity supply chain risk management is integrated into risk management processes', 3, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- NIST CSF — Identify Categories
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0800000-0000-0000-0000-000000000100', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000002', 'ID.AM', 'Asset Management', 1, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000101', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000100', 'ID.AM-01', 'Inventories of hardware managed by the organization are maintained', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000102', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000100', 'ID.AM-02', 'Inventories of software, services, and systems managed by the organization are maintained', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000103', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000100', 'ID.AM-07', 'Inventories of data and corresponding metadata are maintained', 7, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000110', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000002', 'ID.RA', 'Risk Assessment', 2, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000111', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000110', 'ID.RA-01', 'Vulnerabilities in assets are identified, validated, and recorded', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000112', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000110', 'ID.RA-02', 'Cyber threat intelligence is received from information sharing forums and sources', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000113', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000110', 'ID.RA-03', 'Internal and external threats to the organization are identified and recorded', 3, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000120', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000002', 'ID.IM', 'Improvement', 3, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000121', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000120', 'ID.IM-01', 'Improvements are identified from evaluations', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000122', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000120', 'ID.IM-02', 'Improvements are identified from security tests and exercises', 2, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- NIST CSF — Protect Categories
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0800000-0000-0000-0000-000000000200', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000003', 'PR.AA', 'Identity Management, Authentication, and Access Control', 1, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000201', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000200', 'PR.AA-01', 'Identities and credentials for authorized users, services, and hardware are managed', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000202', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000200', 'PR.AA-03', 'Users, services, and hardware are authenticated', 3, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000203', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000200', 'PR.AA-05', 'Access permissions, entitlements, and authorizations are defined and managed', 5, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000210', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000003', 'PR.AT', 'Awareness and Training', 2, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000211', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000210', 'PR.AT-01', 'Personnel are provided cybersecurity awareness and training', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000220', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000003', 'PR.DS', 'Data Security', 3, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000221', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000220', 'PR.DS-01', 'The confidentiality, integrity, and availability of data-at-rest are protected', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000222', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000220', 'PR.DS-02', 'The confidentiality, integrity, and availability of data-in-transit are protected', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000230', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000003', 'PR.PS', 'Platform Security', 4, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000231', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000230', 'PR.PS-01', 'Configuration management practices are established and applied', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000232', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000230', 'PR.PS-02', 'Software is maintained, replaced, and removed commensurate with risk', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000240', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000003', 'PR.IR', 'Technology Infrastructure Resilience', 5, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000241', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000240', 'PR.IR-01', 'Networks and environments are protected from unauthorized logical access', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000242', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000240', 'PR.IR-04', 'Adequate resource capacity to ensure availability is maintained', 4, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- NIST CSF — Detect, Respond, Recover Categories
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0800000-0000-0000-0000-000000000300', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000004', 'DE.CM', 'Continuous Monitoring', 1, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000301', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000300', 'DE.CM-01', 'Networks and network services are monitored to find potentially adverse events', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000302', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000300', 'DE.CM-02', 'The physical environment is monitored to find potentially adverse events', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000303', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000300', 'DE.CM-03', 'Personnel activity and technology usage are monitored to find potentially adverse events', 3, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000310', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000004', 'DE.AE', 'Adverse Event Analysis', 2, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000311', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000310', 'DE.AE-02', 'Potentially adverse events are analyzed to better understand associated activities', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000312', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000310', 'DE.AE-06', 'Information on adverse events is provided to authorized staff and tools', 6, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000400', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000005', 'RS.MA', 'Incident Management', 1, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000401', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000400', 'RS.MA-01', 'The incident response plan is executed in coordination with relevant third parties', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000402', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000400', 'RS.MA-02', 'Incident reports are triaged and validated', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000403', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000400', 'RS.MA-03', 'Incidents are categorized and prioritized', 3, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000410', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000005', 'RS.CO', 'Incident Reporting and Communication', 2, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000411', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000410', 'RS.CO-02', 'Internal and external stakeholders are notified of incidents', 2, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000420', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000005', 'RS.AN', 'Incident Analysis', 3, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000421', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000420', 'RS.AN-03', 'Analysis is performed to establish what has taken place during an incident', 3, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000500', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000006', 'RC.RP', 'Incident Recovery Plan Execution', 1, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000501', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000500', 'RC.RP-01', 'The recovery portion of the incident response plan is executed', 1, 2, TRUE),
    ('e0800000-0000-0000-0000-000000000510', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000006', 'RC.CO', 'Incident Recovery Communication', 2, 1, FALSE),
    ('e0800000-0000-0000-0000-000000000511', 'a1000000-0000-0000-0000-000000000008', 'e0800000-0000-0000-0000-000000000510', 'RC.CO-03', 'Recovery activities and progress in restoring operational capabilities are communicated', 3, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- CIS Controls v8 Requirements
-- ============================================================================

-- CIS Controls — 18 Top-Level Controls
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0900000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.01', 'Inventory and Control of Enterprise Assets', 1, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.02', 'Inventory and Control of Software Assets', 2, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.03', 'Data Protection', 3, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.04', 'Secure Configuration of Enterprise Assets and Software', 4, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000005', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.05', 'Account Management', 5, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000006', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.06', 'Access Control Management', 6, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000007', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.07', 'Continuous Vulnerability Management', 7, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000008', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.08', 'Audit Log Management', 8, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000009', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.09', 'Email and Web Browser Protections', 9, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000010', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.10', 'Malware Defenses', 10, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000011', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.11', 'Data Recovery', 11, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000012', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.12', 'Network Infrastructure Management', 12, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000013', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.13', 'Network Monitoring and Defense', 13, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000014', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.14', 'Security Awareness and Skills Training', 14, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000015', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.15', 'Service Provider Management', 15, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000016', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.16', 'Application Software Security', 16, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000017', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.17', 'Incident Response Management', 17, 0, FALSE),
    ('e0900000-0000-0000-0000-000000000018', 'a1000000-0000-0000-0000-000000000009', NULL, 'CIS.18', 'Penetration Testing', 18, 0, FALSE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- CIS Controls — Key Sub-Controls (Safeguards)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    -- Control 1: Enterprise Assets
    ('e0900000-0000-0000-0000-000000000101', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000001', 'CIS.01.1', 'Establish and Maintain Detailed Enterprise Asset Inventory', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000102', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000001', 'CIS.01.2', 'Address Unauthorized Assets', 2, 1, TRUE),
    -- Control 2: Software Assets
    ('e0900000-0000-0000-0000-000000000201', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000002', 'CIS.02.1', 'Establish and Maintain a Software Inventory', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000202', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000002', 'CIS.02.2', 'Ensure Authorized Software is Currently Supported', 2, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000203', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000002', 'CIS.02.3', 'Address Unauthorized Software', 3, 1, TRUE),
    -- Control 3: Data Protection
    ('e0900000-0000-0000-0000-000000000301', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000003', 'CIS.03.1', 'Establish and Maintain a Data Management Process', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000302', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000003', 'CIS.03.3', 'Configure Data Access Control Lists', 3, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000303', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000003', 'CIS.03.6', 'Encrypt Data on End-User Devices', 6, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000304', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000003', 'CIS.03.9', 'Encrypt Data on Removable Media', 9, 1, TRUE),
    -- Control 4: Secure Configuration
    ('e0900000-0000-0000-0000-000000000401', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000004', 'CIS.04.1', 'Establish and Maintain a Secure Configuration Process', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000402', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000004', 'CIS.04.2', 'Establish and Maintain a Secure Configuration Process for Network Infrastructure', 2, 1, TRUE),
    -- Control 5: Account Management
    ('e0900000-0000-0000-0000-000000000501', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000005', 'CIS.05.1', 'Establish and Maintain an Inventory of Accounts', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000502', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000005', 'CIS.05.2', 'Use Unique Passwords', 2, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000503', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000005', 'CIS.05.3', 'Disable Dormant Accounts', 3, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000504', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000005', 'CIS.05.4', 'Restrict Administrator Privileges to Dedicated Administrator Accounts', 4, 1, TRUE),
    -- Control 6: Access Control
    ('e0900000-0000-0000-0000-000000000601', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000006', 'CIS.06.1', 'Establish an Access Granting Process', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000602', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000006', 'CIS.06.2', 'Establish an Access Revoking Process', 2, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000603', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000006', 'CIS.06.3', 'Require MFA for Externally-Exposed Applications', 3, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000604', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000006', 'CIS.06.4', 'Require MFA for Remote Network Access', 4, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000605', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000006', 'CIS.06.5', 'Require MFA for Administrative Access', 5, 1, TRUE),
    -- Control 7: Vulnerability Management
    ('e0900000-0000-0000-0000-000000000701', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000007', 'CIS.07.1', 'Establish and Maintain a Vulnerability Management Process', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000702', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000007', 'CIS.07.4', 'Perform Automated Application Patch Management', 4, 1, TRUE),
    -- Control 8: Audit Log
    ('e0900000-0000-0000-0000-000000000801', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000008', 'CIS.08.1', 'Establish and Maintain an Audit Log Management Process', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000802', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000008', 'CIS.08.2', 'Collect Audit Logs', 2, 1, TRUE),
    ('e0900000-0000-0000-0000-000000000803', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000008', 'CIS.08.5', 'Collect Detailed Audit Logs', 5, 1, TRUE),
    -- Control 10-18 key safeguards
    ('e0900000-0000-0000-0000-000000001001', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000010', 'CIS.10.1', 'Deploy and Maintain Anti-Malware Software', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001101', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000011', 'CIS.11.1', 'Establish and Maintain a Data Recovery Process', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001201', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000012', 'CIS.12.1', 'Ensure Network Infrastructure is Up-to-Date', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001301', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000013', 'CIS.13.1', 'Centralize Security Event Alerting', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001401', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000014', 'CIS.14.1', 'Establish and Maintain a Security Awareness Program', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001501', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000015', 'CIS.15.1', 'Establish and Maintain an Inventory of Service Providers', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001601', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000016', 'CIS.16.1', 'Establish and Maintain a Secure Application Development Process', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001701', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000017', 'CIS.17.1', 'Designate Personnel to Manage Incident Handling', 1, 1, TRUE),
    ('e0900000-0000-0000-0000-000000001801', 'a1000000-0000-0000-0000-000000000009', 'e0900000-0000-0000-0000-000000000018', 'CIS.18.1', 'Establish and Maintain a Penetration Testing Program', 1, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- SOX IT Controls Requirements
-- ============================================================================

-- SOX — Top-level sections
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000010', NULL, 'SOX.302', 'Section 302: Corporate Responsibility for Financial Reports', 1, 0, FALSE),
    ('e1000000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000010', NULL, 'SOX.404', 'Section 404: Management Assessment of Internal Controls', 2, 0, FALSE),
    ('e1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000010', NULL, 'SOX.ITGC', 'IT General Controls', 3, 0, FALSE),
    ('e1000000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000010', NULL, 'SOX.ITAC', 'IT Application Controls', 4, 0, FALSE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOX — Section 302 Requirements
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e1000000-0000-0000-0000-000000000010', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000001', 'SOX.302.1', 'CEO/CFO certify financial statements are accurate and complete', 1, 1, TRUE),
    ('e1000000-0000-0000-0000-000000000011', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000001', 'SOX.302.2', 'Officers have designed and maintained internal controls over financial reporting', 2, 1, TRUE),
    ('e1000000-0000-0000-0000-000000000012', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000001', 'SOX.302.3', 'Officers have disclosed significant deficiencies and material weaknesses to the auditor', 3, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOX — Section 404 Requirements
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e1000000-0000-0000-0000-000000000020', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000002', 'SOX.404.1', 'Management assessment of the effectiveness of internal controls over financial reporting', 1, 1, TRUE),
    ('e1000000-0000-0000-0000-000000000021', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000002', 'SOX.404.2', 'External auditor attestation on management assessment of internal controls', 2, 1, TRUE),
    ('e1000000-0000-0000-0000-000000000022', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000002', 'SOX.404.3', 'Identification and testing of key controls', 3, 1, TRUE),
    ('e1000000-0000-0000-0000-000000000023', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000002', 'SOX.404.4', 'Documentation of control design and operating effectiveness', 4, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOX — IT General Controls (ITGC)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e1000000-0000-0000-0000-000000000100', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000003', 'SOX.ITGC.AC', 'Access to Programs and Data', 1, 1, FALSE),
    ('e1000000-0000-0000-0000-000000000101', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000100', 'SOX.ITGC.AC.1', 'Logical access to financial systems is appropriately restricted', 1, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000102', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000100', 'SOX.ITGC.AC.2', 'New access is appropriately authorized and provisioned', 2, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000103', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000100', 'SOX.ITGC.AC.3', 'Access is removed upon termination or role change', 3, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000104', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000100', 'SOX.ITGC.AC.4', 'Periodic user access reviews are performed', 4, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000105', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000100', 'SOX.ITGC.AC.5', 'Privileged access is appropriately restricted and monitored', 5, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000106', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000100', 'SOX.ITGC.AC.6', 'Authentication mechanisms are appropriately configured', 6, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000110', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000003', 'SOX.ITGC.CM', 'Change Management', 2, 1, FALSE),
    ('e1000000-0000-0000-0000-000000000111', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000110', 'SOX.ITGC.CM.1', 'Changes to applications and systems follow a defined change management process', 1, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000112', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000110', 'SOX.ITGC.CM.2', 'Changes are tested before implementation in production', 2, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000113', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000110', 'SOX.ITGC.CM.3', 'Changes are appropriately approved before deployment', 3, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000114', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000110', 'SOX.ITGC.CM.4', 'Separation of duties exists between development and production environments', 4, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000120', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000003', 'SOX.ITGC.CO', 'Computer Operations', 3, 1, FALSE),
    ('e1000000-0000-0000-0000-000000000121', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000120', 'SOX.ITGC.CO.1', 'Job scheduling and batch processing are appropriately managed', 1, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000122', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000120', 'SOX.ITGC.CO.2', 'Backup and recovery procedures are in place and tested', 2, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000123', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000120', 'SOX.ITGC.CO.3', 'Problem and incident management procedures are defined and followed', 3, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- SOX — IT Application Controls (ITAC)
INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e1000000-0000-0000-0000-000000000200', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000004', 'SOX.ITAC.IP', 'Input Processing Controls', 1, 1, FALSE),
    ('e1000000-0000-0000-0000-000000000201', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000200', 'SOX.ITAC.IP.1', 'Completeness and accuracy of input data is validated', 1, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000202', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000200', 'SOX.ITAC.IP.2', 'Data processing is complete, accurate, and timely', 2, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000210', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000004', 'SOX.ITAC.IF', 'Interface Controls', 2, 1, FALSE),
    ('e1000000-0000-0000-0000-000000000211', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000210', 'SOX.ITAC.IF.1', 'Data transferred between systems is complete and accurate', 1, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000220', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000004', 'SOX.ITAC.RP', 'Reporting Controls', 3, 1, FALSE),
    ('e1000000-0000-0000-0000-000000000221', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000220', 'SOX.ITAC.RP.1', 'Financial reports generated from IT systems are accurate and complete', 1, 2, TRUE),
    ('e1000000-0000-0000-0000-000000000222', 'a1000000-0000-0000-0000-000000000010', 'e1000000-0000-0000-0000-000000000220', 'SOX.ITAC.RP.2', 'Report access is restricted to authorized personnel', 2, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- ORG FRAMEWORK ACTIVATIONS (demo org activates new frameworks)
-- ============================================================================

INSERT INTO org_frameworks (id, org_id, framework_id, active_version_id, status, target_date, notes) VALUES
    ('d0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000006', 'a1000000-0000-0000-0000-000000000007', 'active', NULL, 'HIPAA compliance for healthcare data handling'),
    ('d0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000007', 'a1000000-0000-0000-0000-000000000008', 'active', NULL, 'NIST CSF 2.0 as cybersecurity baseline'),
    ('d0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000008', 'a1000000-0000-0000-0000-000000000009', 'active', NULL, 'CIS Controls v8 for operational security'),
    ('d0000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000009', 'a1000000-0000-0000-0000-000000000010', 'active', '2026-12-31', 'SOX IT controls for financial reporting systems')
ON CONFLICT (org_id, framework_id) DO NOTHING;
