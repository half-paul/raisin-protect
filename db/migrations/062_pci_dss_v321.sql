-- Migration 062: PCI DSS v3.2.1 Framework Version + Requirements
-- Adds the legacy PCI DSS v3.2.1 alongside existing v4.0.1
-- Idempotent: uses ON CONFLICT DO NOTHING

-- ============================================================================
-- PCI DSS v3.2.1 — Framework Version
-- ============================================================================

INSERT INTO framework_versions (id, framework_id, version, display_name, status, effective_date, total_requirements) VALUES
    ('a1000000-0000-0000-0000-000000000006', 'f0000000-0000-0000-0000-000000000003', '3.2.1', 'PCI DSS v3.2.1', 'sunset', '2018-05-01', 247)
ON CONFLICT (framework_id, version) DO NOTHING;

-- ============================================================================
-- PCI DSS v3.2.1 — 12 Top-Level Requirements
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000006', NULL, '1', 'Install and maintain a firewall configuration to protect cardholder data', 1, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000006', NULL, '2', 'Do not use vendor-supplied defaults for system passwords and other security parameters', 2, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000006', NULL, '3', 'Protect stored cardholder data', 3, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000006', NULL, '4', 'Encrypt transmission of cardholder data across open, public networks', 4, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000005', 'a1000000-0000-0000-0000-000000000006', NULL, '5', 'Protect all systems against malware and regularly update anti-virus software or programs', 5, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000006', 'a1000000-0000-0000-0000-000000000006', NULL, '6', 'Develop and maintain secure systems and applications', 6, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000007', 'a1000000-0000-0000-0000-000000000006', NULL, '7', 'Restrict access to cardholder data by business need to know', 7, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000008', 'a1000000-0000-0000-0000-000000000006', NULL, '8', 'Identify and authenticate access to system components', 8, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000009', 'a1000000-0000-0000-0000-000000000006', NULL, '9', 'Restrict physical access to cardholder data', 9, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000010', 'a1000000-0000-0000-0000-000000000006', NULL, '10', 'Track and monitor all access to network resources and cardholder data', 10, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000011', 'a1000000-0000-0000-0000-000000000006', NULL, '11', 'Regularly test security systems and processes', 11, 0, FALSE),
    ('e0600000-0000-0000-0000-000000000012', 'a1000000-0000-0000-0000-000000000006', NULL, '12', 'Maintain a policy that addresses information security for all personnel', 12, 0, FALSE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 1: Firewall Configuration
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000101', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000001', '1.1', 'Establish and implement firewall and router configuration standards', 1, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000102', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000101', '1.1.1', 'A formal process for approving and testing all network connections and changes to firewall/router configurations', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000103', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000101', '1.1.2', 'Current network diagram that identifies all connections between CDE and other networks', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000104', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000101', '1.1.3', 'Current diagram that shows all cardholder data flows across systems and networks', 3, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000105', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000101', '1.1.4', 'Requirements for a firewall at each Internet connection and between any DMZ and internal network', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000106', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000101', '1.1.6', 'Documentation of business justification for use of all services, protocols, and ports allowed', 6, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000107', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000101', '1.1.7', 'Requirement to review firewall and router rule sets at least every six months', 7, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000110', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000001', '1.2', 'Build firewall and router configurations that restrict connections between untrusted networks and CDE', 2, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000111', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000110', '1.2.1', 'Restrict inbound and outbound traffic to that which is necessary for the CDE', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000112', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000110', '1.2.3', 'Install perimeter firewalls between all wireless networks and the CDE', 3, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000120', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000001', '1.3', 'Prohibit direct public access between the Internet and any system component in the CDE', 3, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000121', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000120', '1.3.1', 'Implement a DMZ to limit inbound traffic to only system components that provide authorized services', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000122', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000120', '1.3.4', 'Do not allow unauthorized outbound traffic from the CDE to the Internet', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000130', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000001', '1.4', 'Install personal firewall software on any portable computing devices that connect to the Internet outside the network', 4, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 2: Vendor Defaults
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000201', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000002', '2.1', 'Always change vendor-supplied defaults and remove or disable unnecessary default accounts', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000202', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000002', '2.2', 'Develop configuration standards for all system components consistent with industry-accepted hardening standards', 2, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000203', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000202', '2.2.1', 'Implement only one primary function per server', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000204', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000202', '2.2.2', 'Enable only necessary services, protocols, daemons as required for the function of the system', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000205', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000202', '2.2.4', 'Configure system security parameters to prevent misuse', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000206', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000002', '2.3', 'Encrypt all non-console administrative access using strong cryptography', 3, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000210', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000002', '2.4', 'Maintain an inventory of system components that are in scope for PCI DSS', 4, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000211', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000002', '2.6', 'Shared hosting providers must protect each entity hosted environment and cardholder data', 6, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 3: Protect Stored Data
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000301', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000003', '3.1', 'Keep cardholder data storage to a minimum with data retention and disposal policies', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000302', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000003', '3.2', 'Do not store sensitive authentication data after authorization', 2, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000303', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000003', '3.3', 'Mask PAN when displayed so only authorized personnel see full PAN', 3, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000304', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000003', '3.4', 'Render PAN unreadable anywhere it is stored using strong cryptography', 4, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000305', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000003', '3.5', 'Document and implement procedures to protect keys used to secure stored cardholder data', 5, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000306', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000305', '3.5.1', 'Restrict access to cryptographic keys to the fewest number of custodians necessary', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000310', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000003', '3.6', 'Fully document and implement all key-management processes and procedures', 6, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000311', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000310', '3.6.7', 'Prevention of unauthorized substitution of cryptographic keys', 7, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 4: Encrypt Transmission
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000401', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000004', '4.1', 'Use strong cryptography and security protocols to safeguard sensitive cardholder data during transmission', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000402', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000004', '4.2', 'Never send unprotected PANs by end-user messaging technologies', 2, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 5: Anti-Malware
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000501', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000005', '5.1', 'Deploy anti-virus software on all systems commonly affected by malicious software', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000502', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000005', '5.2', 'Ensure all anti-virus mechanisms are kept current, perform periodic scans, and generate audit logs', 2, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000503', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000005', '5.3', 'Ensure anti-virus mechanisms are actively running and cannot be disabled or altered by users', 3, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000504', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000005', '5.4', 'Ensure security policies and operational procedures for protecting against malware are documented and known', 4, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 6: Secure Systems
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000601', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000006', '6.1', 'Establish a process to identify security vulnerabilities using reputable outside sources', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000602', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000006', '6.2', 'Ensure all system components and software are protected from known vulnerabilities by installing applicable vendor-supplied security patches', 2, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000603', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000006', '6.3', 'Develop internal and external software applications securely', 3, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000604', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000603', '6.3.1', 'Remove development, test, and custom application accounts/IDs before applications become active', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000605', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000603', '6.3.2', 'Review custom code prior to release to production to identify any potential coding vulnerability', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000610', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000006', '6.4', 'Follow change control processes and procedures for all changes to system components', 4, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000611', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000610', '6.4.1', 'Separate development/test environments from production environments', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000612', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000610', '6.4.2', 'Separation of duties between development/test and production environments', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000613', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000610', '6.4.5', 'Change control procedures for implementation of security patches and software modifications', 5, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000620', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000006', '6.5', 'Address common coding vulnerabilities in software development processes', 5, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000621', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000620', '6.5.1', 'Injection flaws, particularly SQL injection', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000622', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000620', '6.5.3', 'Insecure cryptographic storage', 3, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000623', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000620', '6.5.7', 'Cross-site scripting (XSS)', 7, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000630', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000006', '6.6', 'For public-facing web applications, address new threats and vulnerabilities on an ongoing basis', 6, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000640', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000006', '6.7', 'Ensure security policies and procedures for developing secure systems are documented and known', 7, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 7: Restrict Access
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000701', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000007', '7.1', 'Limit access to system components and cardholder data to only those individuals whose job requires such access', 1, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000702', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000701', '7.1.1', 'Define access needs for each role including system components and data resources', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000703', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000701', '7.1.2', 'Restrict access to privileged user IDs to least privileges necessary', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000704', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000701', '7.1.4', 'Require documented approval by authorized parties specifying required privileges', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000710', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000007', '7.2', 'Establish an access control system for systems components that restricts access based on user need to know', 2, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000711', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000710', '7.2.1', 'Coverage of all system components', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000712', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000710', '7.2.3', 'Default deny-all setting', 3, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000720', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000007', '7.3', 'Ensure security policies and procedures for restricting access are documented and known', 3, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 8: Authentication
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000801', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000008', '8.1', 'Define and implement policies and procedures to ensure proper user identification management', 1, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000802', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000801', '8.1.1', 'Assign all users a unique ID before allowing them to access system components or cardholder data', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000803', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000801', '8.1.4', 'Remove/disable inactive user accounts within 90 days', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000804', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000801', '8.1.6', 'Limit repeated access attempts by locking out the user ID after not more than six attempts', 6, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000805', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000801', '8.1.8', 'If a session has been idle for more than 15 minutes, require re-authentication', 8, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000810', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000008', '8.2', 'In addition to assigning a unique ID, ensure proper user-authentication management', 2, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000811', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000810', '8.2.1', 'Using strong cryptography, render all authentication credentials unreadable during transmission and storage', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000812', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000810', '8.2.3', 'Passwords/passphrases must require a minimum length of at least seven characters and contain both numeric and alphabetic characters', 3, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000813', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000810', '8.2.4', 'Change user passwords/passphrases at least once every 90 days', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000814', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000810', '8.2.5', 'Do not allow an individual to submit a new password that is the same as any of the last four passwords', 5, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000820', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000008', '8.3', 'Secure all individual non-console administrative access and all remote access to the CDE using multi-factor authentication', 3, 1, FALSE),
    ('e0600000-0000-0000-0000-000000000821', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000820', '8.3.1', 'Incorporate multi-factor authentication for all non-console access into the CDE for personnel with administrative access', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000822', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000820', '8.3.2', 'Incorporate multi-factor authentication for all remote network access originating from outside the entity network', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000000830', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000008', '8.5', 'Do not use group, shared, or generic IDs, passwords, or other authentication methods', 5, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000840', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000008', '8.7', 'All access to any database containing cardholder data is restricted', 7, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000850', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000008', '8.8', 'Ensure security policies and procedures for identification and authentication are documented and known', 8, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 9: Physical Access
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000000901', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000009', '9.1', 'Use appropriate facility entry controls to limit and monitor physical access to systems in the CDE', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000902', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000009', '9.2', 'Develop procedures to easily distinguish between onsite personnel and visitors', 2, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000903', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000009', '9.3', 'Control physical access for onsite personnel to sensitive areas', 3, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000905', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000009', '9.5', 'Physically secure all media', 5, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000906', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000009', '9.6', 'Maintain strict control over the internal or external distribution of any kind of media', 6, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000908', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000009', '9.8', 'Destroy media when it is no longer needed for business or legal reasons', 8, 1, TRUE),
    ('e0600000-0000-0000-0000-000000000909', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000009', '9.9', 'Protect devices that capture payment card data via direct physical interaction with the card from tampering and substitution', 9, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 10: Logging and Monitoring
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000001001', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000010', '10.1', 'Implement audit trails to link all access to system components to each individual user', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001002', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000010', '10.2', 'Implement automated audit trails for all system components to reconstruct events', 2, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001003', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001002', '10.2.1', 'All individual user accesses to cardholder data', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001004', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001002', '10.2.2', 'All actions taken by any individual with root or administrative privileges', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001005', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001002', '10.2.4', 'Invalid logical access attempts', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001006', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001002', '10.2.7', 'Creation and deletion of system-level objects', 7, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001010', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000010', '10.3', 'Record audit trail entries for all system components for each event', 3, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001020', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000010', '10.5', 'Secure audit trails so they cannot be altered', 5, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001021', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001020', '10.5.1', 'Limit viewing of audit trails to those with a job-related need', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001022', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001020', '10.5.2', 'Protect audit trail files from unauthorized modifications', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001023', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001020', '10.5.4', 'Write logs for external-facing technologies onto a secure, centralized, internal log server', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001030', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000010', '10.6', 'Review logs and security events for all system components to identify anomalies', 6, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001031', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001030', '10.6.1', 'Review security events, logs of all system components, critical system components, and servers with security functions at least daily', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001040', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000010', '10.7', 'Retain audit trail history for at least one year, with minimum three months immediately available', 7, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001050', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000010', '10.8', 'Ensure security policies and procedures for monitoring all access are documented and known', 8, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 11: Testing
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000001101', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000011', '11.1', 'Implement processes to test for the presence of wireless access points and detect and identify all authorized and unauthorized wireless access points on a quarterly basis', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001102', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000011', '11.2', 'Run internal and external network vulnerability scans at least quarterly and after any significant change', 2, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001103', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001102', '11.2.1', 'Perform quarterly internal vulnerability scans and address vulnerabilities and perform rescans', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001104', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001102', '11.2.2', 'Perform quarterly external vulnerability scans via an Approved Scanning Vendor (ASV)', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001110', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000011', '11.3', 'Implement a methodology for penetration testing', 3, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001111', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001110', '11.3.1', 'Perform external penetration testing at least annually and after any significant infrastructure or application upgrade', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001112', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001110', '11.3.2', 'Perform internal penetration testing at least annually and after any significant infrastructure or application upgrade', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001113', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001110', '11.3.4', 'If segmentation is used to isolate the CDE, perform penetration tests to verify segmentation methods are operational and effective', 4, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001120', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000011', '11.4', 'Use intrusion-detection and/or intrusion-prevention techniques to detect and/or prevent intrusions into the network', 4, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001130', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000011', '11.5', 'Deploy a change-detection mechanism to alert personnel to unauthorized modification of critical system files', 5, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001140', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000011', '11.6', 'Ensure security policies and procedures for security monitoring and testing are documented and known', 6, 1, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;

-- ============================================================================
-- Requirement 12: Information Security Policy
-- ============================================================================

INSERT INTO requirements (id, framework_version_id, parent_id, identifier, title, section_order, depth, is_assessable) VALUES
    ('e0600000-0000-0000-0000-000000001201', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.1', 'Establish, publish, maintain, and disseminate a security policy', 1, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001202', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.2', 'Implement a risk-assessment process that is performed at least annually', 2, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001203', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.3', 'Develop usage policies for critical technologies and define proper use', 3, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001204', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.4', 'Ensure security policy and procedures clearly define information security responsibilities for all personnel', 4, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001205', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.5', 'Assign information security management responsibilities to an individual or team', 5, 1, TRUE),
    ('e0600000-0000-0000-0000-000000001206', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.6', 'Implement a formal security awareness program to make all personnel aware of the cardholder data security policy', 6, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001207', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001206', '12.6.1', 'Educate personnel upon hire and at least annually', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001210', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.8', 'Maintain and implement policies and procedures to manage service providers with whom cardholder data is shared', 8, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001211', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001210', '12.8.1', 'Maintain a list of service providers including a description of the service provided', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001212', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001210', '12.8.2', 'Maintain a written agreement that includes an acknowledgement that the service providers are responsible for the security of cardholder data', 2, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001220', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000000012', '12.10', 'Implement an incident response plan. Be prepared to respond immediately to a system breach', 10, 1, FALSE),
    ('e0600000-0000-0000-0000-000000001221', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001220', '12.10.1', 'Create the incident response plan to be implemented in the event of system breach', 1, 2, TRUE),
    ('e0600000-0000-0000-0000-000000001222', 'a1000000-0000-0000-0000-000000000006', 'e0600000-0000-0000-0000-000000001220', '12.10.2', 'Review and test the plan at least annually', 2, 2, TRUE)
ON CONFLICT (framework_version_id, identifier) DO NOTHING;
