-- Migration: 034_sprint5_seed_demo.sql
-- Description: Demo organization example policies with versions, sign-offs, and control mappings
-- Created: 2026-02-20
-- Sprint: 5 — Policy Management
--
-- Creates 3 demo policies for Acme Corp:
--   POL-IS-001 (published) — Information Security Policy
--   POL-AC-001 (published) — Access Control Policy
--   POL-IR-001 (in_review) — Incident Response Plan

-- ============================================================================
-- DEMO POLICIES (cloned from templates)
-- ============================================================================

INSERT INTO policies (
    id, org_id, identifier, title, description, category, status,
    owner_id, is_template, cloned_from_policy_id,
    review_frequency_days, next_review_at, last_reviewed_at,
    approved_at, approved_version, published_at,
    tags
) VALUES
    -- Published: Information Security Policy
    ('pd000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'POL-IS-001', 'Acme Corp Information Security Policy',
     'Acme Corporation''s information security policy. Establishes the framework for protecting company and customer information assets.',
     'information_security', 'published',
     'b0000000-0000-0000-0000-000000000004',  -- CISO (David)
     FALSE, 'pt000000-0000-0000-0000-000000000001',
     365, '2027-01-15', '2026-01-15',
     '2026-01-15 10:00:00+00', 1, '2026-01-15 10:30:00+00',
     ARRAY['mandatory', 'annual', 'all-employees']),

    -- Published: Access Control Policy
    ('pd000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'POL-AC-001', 'Acme Corp Access Control Policy',
     'Defines access management requirements for all Acme Corp systems and data.',
     'access_control', 'published',
     'b0000000-0000-0000-0000-000000000002',  -- Security Engineer (Bob)
     FALSE, 'pt000000-0000-0000-0000-000000000002',
     365, '2027-02-01', '2026-02-01',
     '2026-02-01 14:00:00+00', 1, '2026-02-01 14:30:00+00',
     ARRAY['annual', 'access', 'production']),

    -- In Review: Incident Response Plan
    ('pd000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'POL-IR-001', 'Acme Corp Incident Response Plan',
     'Procedures for detecting, reporting, and responding to security incidents at Acme Corp.',
     'incident_response', 'in_review',
     'b0000000-0000-0000-0000-000000000002',  -- Security Engineer (Bob)
     FALSE, 'pt000000-0000-0000-0000-000000000003',
     365, NULL, NULL,
     NULL, NULL, NULL,
     ARRAY['annual', 'incident', 'security'])
ON CONFLICT (org_id, identifier) DO NOTHING;

-- ============================================================================
-- DEMO POLICY VERSIONS
-- ============================================================================

INSERT INTO policy_versions (
    id, org_id, policy_id, version_number, is_current,
    content, content_format, content_summary, change_summary, change_type,
    word_count, created_by
) VALUES
    -- Info Security Policy v1
    ('pv000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001', 1, TRUE,
     '<h1>Acme Corporation — Information Security Policy</h1>
<p>Version 1.0 — Effective January 15, 2026</p>

<h2>1. Purpose</h2>
<p>This policy establishes Acme Corporation''s commitment to protecting information assets belonging to the company, its customers, and its partners. It provides the overarching framework for all information security activities.</p>

<h2>2. Scope</h2>
<p>This policy applies to all Acme Corp employees, contractors, consultants, and third-party service providers who access, process, store, or transmit Acme Corp information.</p>

<h2>3. Policy Statement</h2>
<p>Acme Corporation shall:</p>
<ul>
<li>Maintain an Information Security Management System (ISMS) aligned with ISO 27001</li>
<li>Protect information from unauthorized access, disclosure, modification, or destruction</li>
<li>Ensure compliance with SOC 2, PCI DSS, GDPR, and CCPA requirements</li>
<li>Provide annual security awareness training to all personnel</li>
<li>Conduct annual risk assessments and maintain a risk register</li>
<li>Report and investigate all security incidents within established timeframes</li>
</ul>

<h2>4. Roles and Responsibilities</h2>
<p><strong>CISO (David):</strong> Oversees the security program, reports to executive leadership, approves policies.</p>
<p><strong>Security Team (Bob):</strong> Implements technical controls, monitors threats, responds to incidents.</p>
<p><strong>Compliance (Alice):</strong> Manages framework alignment, audit preparation, policy governance.</p>
<p><strong>All Employees:</strong> Comply with policies, complete training, report incidents.</p>

<h2>5. Review</h2>
<p>This policy is reviewed annually by the CISO and Compliance Manager. Next review: January 15, 2027.</p>

<h2>6. Compliance</h2>
<p>Violations may result in disciplinary action. Intentional violations may result in termination and/or legal action.</p>',
     'html',
     'Acme Corp customized information security policy v1.0 — covers ISMS framework, multi-framework compliance, roles.',
     'Initial version — customized from template TPL-IS-001',
     'initial',
     320,
     'b0000000-0000-0000-0000-000000000001'),  -- Compliance Manager (Alice)

    -- Access Control Policy v1
    ('pv000000-0000-0000-0000-000000000011', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002', 1, TRUE,
     '<h1>Acme Corporation — Access Control Policy</h1>
<p>Version 1.0 — Effective February 1, 2026</p>

<h2>1. Purpose</h2>
<p>This policy defines requirements for managing user access to Acme Corp information systems, applications, and data.</p>

<h2>2. Authentication</h2>
<ul>
<li>MFA required for all production system access and VPN connections</li>
<li>Password policy: minimum 14 characters, complexity required, 24 password history</li>
<li>Service accounts use certificate-based authentication; rotated every 90 days</li>
</ul>

<h2>3. Authorization</h2>
<ul>
<li>RBAC enforced across all systems with 7 defined GRC roles</li>
<li>Least privilege principle — no standing admin access</li>
<li>Privileged access via PAM with session recording and JIT approval</li>
</ul>

<h2>4. Access Reviews</h2>
<ul>
<li>Quarterly access reviews for all critical systems</li>
<li>Stale accounts (90+ days inactive) automatically disabled</li>
<li>Termination access revocation within 4 hours</li>
</ul>

<h2>5. Review</h2>
<p>This policy is reviewed annually. Next review: February 1, 2027.</p>',
     'html',
     'Acme Corp access control policy v1.0 — MFA, RBAC, PAM, quarterly access reviews.',
     'Initial version — customized from template TPL-AC-001',
     'initial',
     280,
     'b0000000-0000-0000-0000-000000000002'),  -- Security Engineer (Bob)

    -- Incident Response Plan v1 (draft/in-review)
    ('pv000000-0000-0000-0000-000000000012', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000003', 1, TRUE,
     '<h1>Acme Corporation — Incident Response Plan</h1>
<p>DRAFT — Version 1.0 — Pending Approval</p>

<h2>1. Purpose</h2>
<p>This plan defines procedures for Acme Corp''s detection, reporting, and response to information security incidents.</p>

<h2>2. Incident Classification</h2>
<ul>
<li><strong>P1 Critical:</strong> Active breach, ransomware, system-wide compromise — 15 min response</li>
<li><strong>P2 High:</strong> Unauthorized data access, production malware — 1 hour response</li>
<li><strong>P3 Medium:</strong> Phishing success, policy violation — 4 hour response</li>
<li><strong>P4 Low:</strong> Suspicious activity, failed attempts — 24 hour response</li>
</ul>

<h2>3. Response Team</h2>
<p><strong>Incident Commander:</strong> CISO (David) — or delegate</p>
<p><strong>Technical Lead:</strong> Security Engineer (Bob)</p>
<p><strong>Communications:</strong> Compliance Manager (Alice)</p>
<p><strong>IT Support:</strong> IT Admin (Carol)</p>

<h2>4. Response Phases</h2>
<p>Detection → Triage → Containment → Eradication → Recovery → Post-Incident Review</p>

<h2>5. Notifications</h2>
<ul>
<li>GDPR breach notification: 72 hours to supervisory authority</li>
<li>PCI DSS: immediate notification to acquirer and card brands</li>
<li>State breach laws: per jurisdiction requirements (varies 30-60 days)</li>
</ul>

<p><em>This plan is currently in review. Awaiting sign-off from CISO and Compliance Manager.</em></p>',
     'html',
     'Acme Corp incident response plan draft v1.0 — classification, response team, phases, notification timelines.',
     'Initial draft — customized from template TPL-IR-001',
     'initial',
     240,
     'b0000000-0000-0000-0000-000000000002')  -- Security Engineer (Bob)
ON CONFLICT (policy_id, version_number) DO NOTHING;

-- ============================================================================
-- UPDATE current_version_id FOR DEMO POLICIES
-- ============================================================================

UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000010'
WHERE id = 'pd000000-0000-0000-0000-000000000001' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000011'
WHERE id = 'pd000000-0000-0000-0000-000000000002' AND current_version_id IS NULL;

UPDATE policies SET current_version_id = 'pv000000-0000-0000-0000-000000000012'
WHERE id = 'pd000000-0000-0000-0000-000000000003' AND current_version_id IS NULL;

-- ============================================================================
-- DEMO SIGN-OFFS
-- ============================================================================

INSERT INTO policy_signoffs (
    id, org_id, policy_id, policy_version_id,
    signer_id, signer_role, requested_by, requested_at,
    status, decided_at, comments
) VALUES
    -- Info Security Policy — approved by CISO
    ('ps000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001', 'pv000000-0000-0000-0000-000000000010',
     'b0000000-0000-0000-0000-000000000004',  -- CISO (David)
     'ciso',
     'b0000000-0000-0000-0000-000000000001',  -- Compliance Manager (Alice)
     '2026-01-14 09:00:00+00',
     'approved', '2026-01-15 10:00:00+00',
     'Reviewed and approved. Meets all framework requirements. Aligns with our SOC 2 and ISO 27001 obligations.'),

    -- Access Control Policy — approved by CISO
    ('ps000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002', 'pv000000-0000-0000-0000-000000000011',
     'b0000000-0000-0000-0000-000000000004',  -- CISO (David)
     'ciso',
     'b0000000-0000-0000-0000-000000000002',  -- Security Engineer (Bob)
     '2026-01-30 09:00:00+00',
     'approved', '2026-02-01 14:00:00+00',
     'Approved — aligns with PCI DSS Requirements 7 and 8. MFA and access review cadences meet compliance needs.'),

    -- Incident Response Plan — pending sign-off from CISO
    ('ps000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000003', 'pv000000-0000-0000-0000-000000000012',
     'b0000000-0000-0000-0000-000000000004',  -- CISO (David)
     'ciso',
     'b0000000-0000-0000-0000-000000000002',  -- Security Engineer (Bob)
     '2026-02-18 09:00:00+00',
     'pending', NULL, NULL),

    -- Incident Response Plan — pending sign-off from Compliance Manager
    ('ps000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000003', 'pv000000-0000-0000-0000-000000000012',
     'b0000000-0000-0000-0000-000000000001',  -- Compliance Manager (Alice)
     'compliance_manager',
     'b0000000-0000-0000-0000-000000000002',  -- Security Engineer (Bob)
     '2026-02-18 09:00:00+00',
     'pending', NULL, NULL)
ON CONFLICT (policy_version_id, signer_id) DO NOTHING;

-- ============================================================================
-- DEMO POLICY-CONTROL MAPPINGS
-- ============================================================================

INSERT INTO policy_controls (
    id, org_id, policy_id, control_id, coverage, notes, linked_by
) VALUES
    -- Info Security Policy → high-level controls
    ('pc000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001',
     'c0000000-0000-0000-0000-000000000001',  -- CTRL-AC-001 (MFA)
     'partial', 'ISP Section 3 references MFA requirements at a high level — detailed coverage in Access Control Policy',
     'b0000000-0000-0000-0000-000000000001'),  -- Alice

    ('pc000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000001',
     'c0000000-0000-0000-0000-000000000206',  -- CTRL-SA-001 (Security Awareness Training)
     'full', 'ISP Section 3 — security awareness training mandate: annual training for all personnel',
     'b0000000-0000-0000-0000-000000000001'),  -- Alice

    -- Access Control Policy → access control domain
    ('pc000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000001',  -- CTRL-AC-001 (MFA)
     'full', 'Section 2 — MFA enforcement requirements for production and VPN',
     'b0000000-0000-0000-0000-000000000002'),  -- Bob

    ('pc000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000002',  -- CTRL-AC-002 (RBAC)
     'full', 'Section 3 — RBAC and least privilege across all systems with 7 GRC roles',
     'b0000000-0000-0000-0000-000000000002'),  -- Bob

    ('pc000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000003',  -- CTRL-AC-003 (Quarterly Access Reviews)
     'full', 'Section 4 — Quarterly access review requirements with stale account cleanup',
     'b0000000-0000-0000-0000-000000000002'),  -- Bob

    ('pc000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000004',  -- CTRL-AC-004 (PAM)
     'full', 'Section 3 — Privileged access via PAM with session recording and JIT approval',
     'b0000000-0000-0000-0000-000000000002'),  -- Bob

    ('pc000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000005',  -- CTRL-AC-005 (Provisioning/Deprovisioning)
     'full', 'Section 4 — Termination access revocation within 4 hours',
     'b0000000-0000-0000-0000-000000000002'),  -- Bob

    ('pc000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001',
     'pd000000-0000-0000-0000-000000000002',
     'c0000000-0000-0000-0000-000000000006',  -- CTRL-AC-006 (Password Policy)
     'full', 'Section 2 — Password policy: 14+ chars, complexity, 24 password history',
     'b0000000-0000-0000-0000-000000000002')  -- Bob
ON CONFLICT (org_id, policy_id, control_id) DO NOTHING;

-- ============================================================================
-- AUDIT LOG ENTRIES FOR DEMO POLICIES
-- ============================================================================

INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'policy.created', 'policy', 'pd000000-0000-0000-0000-000000000001',
     '{"identifier": "POL-IS-001", "title": "Acme Corp Information Security Policy", "cloned_from": "TPL-IS-001"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'policy_version.created', 'policy_version', 'pv000000-0000-0000-0000-000000000010',
     '{"policy_id": "pd000000-0000-0000-0000-000000000001", "version": 1, "change_type": "initial"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000004',
     'policy_signoff.approved', 'policy_signoff', 'ps000000-0000-0000-0000-000000000001',
     '{"policy_id": "pd000000-0000-0000-0000-000000000001", "policy": "POL-IS-001", "signer": "ciso@acme.example.com"}'::jsonb,
     '192.168.1.20'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'policy.status_changed', 'policy', 'pd000000-0000-0000-0000-000000000001',
     '{"identifier": "POL-IS-001", "from": "in_review", "to": "published"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'policy.created', 'policy', 'pd000000-0000-0000-0000-000000000003',
     '{"identifier": "POL-IR-001", "title": "Acme Corp Incident Response Plan", "cloned_from": "TPL-IR-001"}'::jsonb,
     '192.168.1.15'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'policy_signoff.requested', 'policy_signoff', 'ps000000-0000-0000-0000-000000000003',
     '{"policy_id": "pd000000-0000-0000-0000-000000000003", "policy": "POL-IR-001", "signer": "ciso@acme.example.com"}'::jsonb,
     '192.168.1.15'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'policy_control.linked', 'policy_control', 'pc000000-0000-0000-0000-000000000003',
     '{"policy": "POL-AC-001", "control": "CTRL-AC-001", "coverage": "full"}'::jsonb,
     '192.168.1.15'::inet)
ON CONFLICT DO NOTHING;
