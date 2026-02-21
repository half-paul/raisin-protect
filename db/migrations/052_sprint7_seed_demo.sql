-- Migration: 052_sprint7_seed_demo.sql
-- Description: Demo audit engagement — 1 audit, 8 requests, 4 findings, 5 evidence links, 6 comments (Sprint 7)
-- Created: 2026-02-21
-- Sprint: 7 — Audit Hub
-- Idempotent: ON CONFLICT DO NOTHING on all inserts
--
-- Uses existing demo data:
--   Org: a0000000-...-000000000001 (Acme Corporation)
--   Users: b0000000-...-000000000001 (Alice/compliance_manager)
--          b0000000-...-000000000002 (Bob/security_engineer)
--          b0000000-...-000000000004 (David/ciso)
--          b0000000-...-000000000006 (Frank/auditor)
--   Org Framework: d0000000-...-000000000001 (SOC 2 active)
--   Evidence: e0000000-...-000000000001 (Okta MFA Config)
--             e0000000-...-000000000002 (Access Review Report)
--             e0000000-...-000000000004 (InfoSec Policy)

-- ============================================================================
-- ADDITIONAL AUDITOR USERS (2 more for the audit engagement)
-- ============================================================================

INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role, status)
VALUES
    ('b0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001',
     'auditor2@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Hannah', 'Auditor', 'auditor', 'active'),
    ('b0000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000001',
     'auditor3@acme.example.com', '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
     'Ivan', 'Auditor', 'auditor', 'active')
ON CONFLICT (org_id, email) DO NOTHING;

-- ============================================================================
-- DEMO AUDIT ENGAGEMENT: SOC 2 Type II — 2026 Annual
-- ============================================================================

INSERT INTO audits (
    id, org_id, title, description, audit_type, status,
    org_framework_id,
    period_start, period_end,
    planned_start, planned_end,
    actual_start, actual_end,
    audit_firm, lead_auditor_id, auditor_ids,
    internal_lead_id,
    milestones,
    report_type,
    total_requests, open_requests,
    total_findings, open_findings,
    tags
) VALUES (
    'aa000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'SOC 2 Type II — 2026 Annual',
    'Annual SOC 2 Type II audit covering the Trust Services Criteria for security, availability, and confidentiality. Audit period January 1 through December 31, 2026. Deloitte & Touche LLP engaged as the service auditor.',
    'soc2_type2',
    'fieldwork',
    'd0000000-0000-0000-0000-000000000001',
    '2026-01-01', '2026-12-31',
    '2026-02-01', '2026-04-30',
    '2026-02-03', NULL,
    'Deloitte & Touche LLP',
    'b0000000-0000-0000-0000-000000000006',
    ARRAY['b0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000008', 'b0000000-0000-0000-0000-000000000009']::UUID[],
    'b0000000-0000-0000-0000-000000000001',
    '[
        {"name": "Kickoff Meeting", "target_date": "2026-02-03", "completed_at": "2026-02-03T10:00:00Z"},
        {"name": "Fieldwork Start", "target_date": "2026-02-10", "completed_at": "2026-02-10T09:00:00Z"},
        {"name": "Interim Testing", "target_date": "2026-03-01", "completed_at": null},
        {"name": "Final Testing", "target_date": "2026-03-15", "completed_at": null},
        {"name": "Draft Report", "target_date": "2026-04-01", "completed_at": null},
        {"name": "Management Response", "target_date": "2026-04-15", "completed_at": null},
        {"name": "Final Report Issuance", "target_date": "2026-04-30", "completed_at": null}
    ]'::jsonb,
    'SOC 2 Type II',
    8, 5,
    4, 3,
    ARRAY['soc2', 'annual', '2026', 'deloitte']
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- DEMO AUDIT REQUESTS (8)
-- ============================================================================

INSERT INTO audit_requests (
    id, org_id, audit_id, title, description, priority, status,
    control_id, requirement_id,
    requested_by, assigned_to,
    due_date, submitted_at, reviewed_at, reviewer_notes,
    reference_number, tags
) VALUES
    -- 1. Information Security Policy — ACCEPTED
    ('ar000000-0000-0000-0000-000000000001',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Information Security Policy (Current, Approved)',
     'Provide the current board-approved information security policy. Include approval date, version history, and evidence of annual review.',
     'high', 'accepted',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-21', '2026-02-15 14:30:00+00', '2026-02-16 09:15:00+00',
     'Policy document verified. Version 3.1 is current and approved by CISO on 2026-01-15.',
     'PBC-001', ARRAY['policy', 'cc1', 'cc5']),

    -- 2. Access Control Procedures — SUBMITTED (pending review)
    ('ar000000-0000-0000-0000-000000000002',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Access Control Procedures and User Provisioning',
     'Document access control procedures including user provisioning, de-provisioning, RBAC model, and approval workflows for access requests.',
     'high', 'submitted',
     'c0000000-0000-0000-0000-000000000002', NULL,
     'b0000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-25', '2026-02-18 11:00:00+00', NULL,
     NULL,
     'PBC-002', ARRAY['access-control', 'cc6']),

    -- 3. Network Diagram — IN_PROGRESS (assigned)
    ('ar000000-0000-0000-0000-000000000003',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Network Diagram Showing Data Flows',
     'Provide current network architecture diagram showing production environment, data flows, encryption indicators, and security zone boundaries.',
     'high', 'in_progress',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-28', NULL, NULL,
     NULL,
     'PBC-003', ARRAY['network', 'cc6', 'cc7']),

    -- 4. Vulnerability Scan Reports — OPEN (unassigned)
    ('ar000000-0000-0000-0000-000000000004',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Vulnerability Scan Reports Q1-Q2 2026',
     'Provide internal and external vulnerability scan reports covering Q1 and Q2 2026. Include remediation tracking and trend analysis for high/critical findings.',
     'high', 'open',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000008',
     NULL,
     '2026-03-07', NULL, NULL,
     NULL,
     'PBC-004', ARRAY['vulnerability', 'cc7']),

    -- 5. Change Management Logs — ACCEPTED
    ('ar000000-0000-0000-0000-000000000005',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Change Management Procedures and Recent Change Logs',
     'Provide change management policy and change logs from the audit period showing approval, testing, and rollback documentation for production changes.',
     'high', 'accepted',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000005',
     '2026-02-21', '2026-02-14 16:45:00+00', '2026-02-15 10:30:00+00',
     'Change management process well-documented. Change logs reviewed for December-February sample. No unauthorized changes detected.',
     'PBC-005', ARRAY['change-management', 'cc8']),

    -- 6. Incident Response Plan — REJECTED (needs update)
    ('ar000000-0000-0000-0000-000000000006',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Incident Response Plan and Recent Incident Reports',
     'Provide the incident response plan including escalation procedures, communication templates, and reports for any security incidents during the audit period.',
     'high', 'rejected',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-21', '2026-02-17 09:00:00+00', '2026-02-18 14:00:00+00',
     'IR plan provided is version 2.0 from 2024. Needs update to reflect current team structure and contact information. Also missing tabletop exercise results from 2025.',
     'PBC-006', ARRAY['incident-response', 'cc7']),

    -- 7. Employee Training Records — OVERDUE
    ('ar000000-0000-0000-0000-000000000007',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Security Awareness Training Records (Last 12 Months)',
     'Provide evidence of security awareness training completion for all employees. Include completion rates, topics covered, and phishing simulation results.',
     'medium', 'overdue',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000008',
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-14', NULL, NULL,
     NULL,
     'PBC-007', ARRAY['training', 'cc1']),

    -- 8. Vendor Risk Assessments — OPEN
    ('ar000000-0000-0000-0000-000000000008',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Vendor Risk Assessments for Critical Third Parties',
     'Provide vendor risk assessments and due diligence records for top 10 critical third-party vendors. Include SOC report reviews and security questionnaire results.',
     'medium', 'open',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000009',
     NULL,
     '2026-03-01', NULL, NULL,
     NULL,
     'PBC-008', ARRAY['vendor', 'cc9'])
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- DEMO AUDIT FINDINGS (4)
-- ============================================================================

INSERT INTO audit_findings (
    id, org_id, audit_id, title, description, severity, category, status,
    control_id, requirement_id,
    found_by, remediation_owner_id,
    remediation_plan, remediation_due_date,
    remediation_started_at, remediation_completed_at,
    verification_notes, verified_at, verified_by,
    risk_accepted, risk_acceptance_reason, risk_accepted_by, risk_accepted_at,
    reference_number, recommendation, management_response,
    tags
) VALUES
    -- 1. Incomplete access reviews — HIGH, remediation in progress
    ('af000000-0000-0000-0000-000000000001',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Incomplete Quarterly Access Reviews',
     'Access reviews for Q3 and Q4 2025 were incomplete. Three production systems (payment-api, customer-db, analytics-platform) were not included in the quarterly review cycle. 47 user accounts on these systems were not reviewed during the audit period.',
     'high', 'access_control', 'remediation_in_progress',
     'c0000000-0000-0000-0000-000000000003', NULL,
     'b0000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000002',
     'Expand access review scope to include all production systems. Implement automated access review tool (Opal) to ensure complete coverage. Backfill Q3/Q4 reviews for the three missed systems.',
     '2026-03-15',
     '2026-02-18 10:00:00+00', NULL,
     NULL, NULL, NULL,
     FALSE, NULL, NULL, NULL,
     'FINDING-001',
     'Implement an automated access review process that covers all production systems. Consider tools like Opal, ConductorOne, or Zluri for continuous access monitoring.',
     'We acknowledge this finding. The access review tool Opal is being deployed and will cover all production systems. Backfill reviews are in progress.',
     ARRAY['access-review', 'critical-path']),

    -- 2. Missing MFA on admin accounts — CRITICAL, remediation complete
    ('af000000-0000-0000-0000-000000000002',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Missing MFA on Infrastructure Admin Accounts',
     'Two AWS IAM admin accounts and one database admin account were not enrolled in multi-factor authentication. These accounts have full administrative access to production infrastructure.',
     'critical', 'access_control', 'remediation_complete',
     'c0000000-0000-0000-0000-000000000001', NULL,
     'b0000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000002',
     'Immediately enforce MFA on all identified admin accounts. Implement AWS SCP to prevent console access without MFA. Add MFA enrollment check to privileged access provisioning workflow.',
     '2026-02-28',
     '2026-02-11 09:00:00+00', '2026-02-14 16:00:00+00',
     NULL, NULL, NULL,
     FALSE, NULL, NULL, NULL,
     'FINDING-002',
     'Enforce MFA on all administrative and privileged accounts immediately. Implement preventive controls (AWS SCPs, conditional access policies) to block unprotected admin access.',
     'MFA has been enforced on all three accounts. AWS SCP deployed to require MFA for console access. Database admin MFA added via Okta RADIUS integration.',
     ARRAY['mfa', 'admin-access', 'urgent']),

    -- 3. Stale vendor risk assessments — MEDIUM, acknowledged
    ('af000000-0000-0000-0000-000000000003',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Stale Vendor Risk Assessments',
     'Four of ten critical vendors have risk assessments older than 12 months. Vendors affected: Cloudflare (18 months old), Datadog (14 months), Stripe (13 months), and SendGrid (15 months).',
     'medium', 'vendor_risk', 'acknowledged',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000009',
     'b0000000-0000-0000-0000-000000000007',
     NULL, NULL,
     NULL, NULL,
     NULL, NULL, NULL,
     FALSE, NULL, NULL, NULL,
     'FINDING-003',
     'Establish a vendor reassessment calendar with automated reminders. Critical vendors should be reassessed annually at minimum. Consider continuous monitoring for Tier 1 vendors.',
     'Acknowledged. We will schedule reassessments for these four vendors within the next 30 days and implement automated reminders in our vendor management system.',
     ARRAY['vendor', 'reassessment']),

    -- 4. Backup test not documented — LOW, identified
    ('af000000-0000-0000-0000-000000000004',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'Backup Recovery Tests Not Formally Documented',
     'While backup recovery tests are performed quarterly, the test results are not formally documented. Recovery times and success/failure outcomes are communicated via Slack but not captured in a formal test report.',
     'low', 'documentation_gap', 'identified',
     NULL, NULL,
     'b0000000-0000-0000-0000-000000000008',
     NULL,
     NULL, NULL,
     NULL, NULL,
     NULL, NULL, NULL,
     FALSE, NULL, NULL, NULL,
     'FINDING-004',
     'Create a formal backup recovery test template that captures test date, systems tested, recovery time, success/failure, and any issues encountered. Retain test reports per the records retention policy.',
     NULL,
     ARRAY['backup', 'documentation'])
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- DEMO AUDIT EVIDENCE LINKS (5)
-- Links existing evidence artifacts to audit requests, creating chain-of-custody
-- ============================================================================

INSERT INTO audit_evidence_links (
    id, org_id, audit_id, request_id, artifact_id,
    submitted_by, submitted_at, submission_notes,
    status, reviewed_by, reviewed_at, review_notes
) VALUES
    -- Request 1 (InfoSec Policy) — linked to InfoSec Policy artifact, ACCEPTED
    ('al000000-0000-0000-0000-000000000001',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'ar000000-0000-0000-0000-000000000001',
     'e0000000-0000-0000-0000-000000000004',
     'b0000000-0000-0000-0000-000000000001',
     '2026-02-15 14:30:00+00',
     'Current Information Security Policy v3.1, approved by CISO David on 2026-01-15. Annual review completed January 2026.',
     'accepted',
     'b0000000-0000-0000-0000-000000000006',
     '2026-02-16 09:15:00+00',
     'Policy document verified. Version 3.1 is current and meets SOC 2 CC1.1 and CC5.2 requirements.'),

    -- Request 2 (Access Control) — linked to Okta MFA Config, PENDING_REVIEW
    ('al000000-0000-0000-0000-000000000002',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'ar000000-0000-0000-0000-000000000002',
     'e0000000-0000-0000-0000-000000000001',
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18 11:00:00+00',
     'Okta MFA configuration export showing enforcement policy for all user groups in production org. MFA enforced for all users since 2025-06.',
     'pending_review',
     NULL, NULL, NULL),

    -- Request 2 (Access Control) — also linked to Access Review Report, PENDING_REVIEW
    ('al000000-0000-0000-0000-000000000003',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'ar000000-0000-0000-0000-000000000002',
     'e0000000-0000-0000-0000-000000000002',
     'b0000000-0000-0000-0000-000000000002',
     '2026-02-18 11:05:00+00',
     'Q1 2026 access review report covering all production systems. Reviews conducted by department managers with sign-offs.',
     'pending_review',
     NULL, NULL, NULL),

    -- Request 5 (Change Mgmt) — linked to InfoSec Policy (covers change mgmt section), ACCEPTED
    ('al000000-0000-0000-0000-000000000004',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'ar000000-0000-0000-0000-000000000005',
     'e0000000-0000-0000-0000-000000000004',
     'b0000000-0000-0000-0000-000000000005',
     '2026-02-14 16:45:00+00',
     'Section 8 of the InfoSec Policy covers Change Management procedures. Supplemented with Jira change logs exported separately.',
     'accepted',
     'b0000000-0000-0000-0000-000000000006',
     '2026-02-15 10:30:00+00',
     'Change management section reviewed. Jira logs show proper approval workflow for all sampled changes.'),

    -- Request 5 (Change Mgmt) — also linked to Vulnerability Scan (shows change impact), NEEDS_CLARIFICATION
    ('al000000-0000-0000-0000-000000000005',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'ar000000-0000-0000-0000-000000000005',
     'e0000000-0000-0000-0000-000000000003',
     'b0000000-0000-0000-0000-000000000005',
     '2026-02-14 16:50:00+00',
     'Vulnerability scan run post-change to verify no new vulnerabilities introduced by recent production changes.',
     'needs_clarification',
     'b0000000-0000-0000-0000-000000000006',
     '2026-02-15 10:35:00+00',
     'Scan is from Feb 18 but the change log references changes from Jan 15-Feb 10. Please provide scan results that cover the same time period as the sampled changes.')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- DEMO AUDIT COMMENTS (6)
-- Mix of auditor questions, internal responses, and internal-only notes
-- ============================================================================

INSERT INTO audit_comments (
    id, org_id, audit_id,
    target_type, target_id,
    author_id, body,
    parent_comment_id, is_internal
) VALUES
    -- Comment 1: Auditor question on the audit itself
    ('ac000000-0000-0000-0000-000000000001',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'audit', 'aa000000-0000-0000-0000-000000000001',
     'b0000000-0000-0000-0000-000000000006',
     'Can you confirm the list of subservice organizations? We need to verify the system description scope includes all relevant third parties.',
     NULL, FALSE),

    -- Comment 2: Internal response to auditor
    ('ac000000-0000-0000-0000-000000000002',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'audit', 'aa000000-0000-0000-0000-000000000001',
     'b0000000-0000-0000-0000-000000000001',
     'Our subservice organizations are: AWS (IaaS), Okta (IAM), Datadog (monitoring), and Stripe (payment processing). We have current SOC 2 reports from all four. I will upload them to the evidence library.',
     'ac000000-0000-0000-0000-000000000001', FALSE),

    -- Comment 3: Internal-only note (hidden from auditors)
    ('ac000000-0000-0000-0000-000000000003',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'audit', 'aa000000-0000-0000-0000-000000000001',
     'b0000000-0000-0000-0000-000000000004',
     'Internal note: The Stripe SOC 2 report has a qualified opinion on the processing integrity criteria. We should prepare talking points in case the auditor asks about it. Discuss at the Monday leadership sync.',
     NULL, TRUE),

    -- Comment 4: Auditor question on request 6 (rejected IR plan)
    ('ac000000-0000-0000-0000-000000000004',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'request', 'ar000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000006',
     'The IR plan references a "Security Operations Center" team that no longer appears in the org chart. Please update the plan to reflect the current team structure and provide evidence of the most recent tabletop exercise.',
     NULL, FALSE),

    -- Comment 5: Internal response on request 6
    ('ac000000-0000-0000-0000-000000000005',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'request', 'ar000000-0000-0000-0000-000000000006',
     'b0000000-0000-0000-0000-000000000002',
     'Understood. We merged the SOC team into the Security Engineering team in Q3 2025. I am updating the IR plan now — will have a revised version by end of week. The tabletop exercise from November 2025 is being formatted for submission.',
     'ac000000-0000-0000-0000-000000000004', FALSE),

    -- Comment 6: Auditor comment on finding 1 (access reviews)
    ('ac000000-0000-0000-0000-000000000006',
     'a0000000-0000-0000-0000-000000000001',
     'aa000000-0000-0000-0000-000000000001',
     'finding', 'af000000-0000-0000-0000-000000000001',
     'b0000000-0000-0000-0000-000000000008',
     'We will need to see the Opal configuration and at least one complete access review cycle using the new tool before we can consider this finding remediated. Please plan for verification testing during the final fieldwork phase.',
     NULL, FALSE)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- AUDIT LOG ENTRIES
-- ============================================================================

INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'audit.created', 'audit', 'aa000000-0000-0000-0000-000000000001',
     '{"title": "SOC 2 Type II — 2026 Annual", "audit_type": "soc2_type2", "firm": "Deloitte & Touche LLP"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'audit.status_changed', 'audit', 'aa000000-0000-0000-0000-000000000001',
     '{"from": "planning", "to": "fieldwork"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000006',
     'audit_request.created', 'audit_request', 'ar000000-0000-0000-0000-000000000001',
     '{"title": "Information Security Policy", "reference": "PBC-001"}'::jsonb,
     '10.0.0.100'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001',
     'audit_evidence.submitted', 'audit_evidence_link', 'al000000-0000-0000-0000-000000000001',
     '{"request_title": "Information Security Policy", "artifact_title": "Information Security Policy v3.1"}'::jsonb,
     '192.168.1.10'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000006',
     'audit_evidence.reviewed', 'audit_evidence_link', 'al000000-0000-0000-0000-000000000001',
     '{"status": "accepted", "artifact_title": "Information Security Policy v3.1"}'::jsonb,
     '10.0.0.100'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000006',
     'audit_finding.created', 'audit_finding', 'af000000-0000-0000-0000-000000000001',
     '{"title": "Incomplete Quarterly Access Reviews", "severity": "high"}'::jsonb,
     '10.0.0.100'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000006',
     'audit_finding.created', 'audit_finding', 'af000000-0000-0000-0000-000000000002',
     '{"title": "Missing MFA on Infrastructure Admin Accounts", "severity": "critical"}'::jsonb,
     '10.0.0.100'::inet),
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000002',
     'audit_finding.status_changed', 'audit_finding', 'af000000-0000-0000-0000-000000000002',
     '{"from": "identified", "to": "remediation_complete", "title": "Missing MFA on Infrastructure Admin Accounts"}'::jsonb,
     '192.168.1.20'::inet)
ON CONFLICT DO NOTHING;
