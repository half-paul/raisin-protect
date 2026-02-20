-- Seed Data: seed.sql
-- Description: Demo organization, users (one per GRC role), and sample audit entries
-- Created: 2026-02-20
-- Sprint: 1 — Project Scaffolding & Auth
--
-- Password for all demo users: demo123
-- Bcrypt hash (cost 12): $2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS

-- ============================================================================
-- DEMO ORGANIZATION
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
-- DEMO USERS (one per GRC role)
-- ============================================================================

INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role, status)
VALUES
    -- Compliance Manager
    (
        'b0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001',
        'compliance@acme.example.com',
        '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
        'Alice', 'Compliance',
        'compliance_manager', 'active'
    ),
    -- Security Engineer
    (
        'b0000000-0000-0000-0000-000000000002',
        'a0000000-0000-0000-0000-000000000001',
        'security@acme.example.com',
        '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
        'Bob', 'Security',
        'security_engineer', 'active'
    ),
    -- IT Admin
    (
        'b0000000-0000-0000-0000-000000000003',
        'a0000000-0000-0000-0000-000000000001',
        'it@acme.example.com',
        '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
        'Carol', 'IT',
        'it_admin', 'active'
    ),
    -- CISO
    (
        'b0000000-0000-0000-0000-000000000004',
        'a0000000-0000-0000-0000-000000000001',
        'ciso@acme.example.com',
        '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
        'David', 'CISO',
        'ciso', 'active'
    ),
    -- DevOps Engineer
    (
        'b0000000-0000-0000-0000-000000000005',
        'a0000000-0000-0000-0000-000000000001',
        'devops@acme.example.com',
        '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
        'Eve', 'DevOps',
        'devops_engineer', 'active'
    ),
    -- Auditor
    (
        'b0000000-0000-0000-0000-000000000006',
        'a0000000-0000-0000-0000-000000000001',
        'auditor@acme.example.com',
        '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
        'Frank', 'Auditor',
        'auditor', 'active'
    ),
    -- Vendor Manager
    (
        'b0000000-0000-0000-0000-000000000007',
        'a0000000-0000-0000-0000-000000000001',
        'vendor@acme.example.com',
        '$2b$12$nykuT6Xga0gKKNVs0HfJOOSbYCiIEFJKtgI26IXe.zseNbA5k6aNS',
        'Grace', 'Vendor',
        'vendor_manager', 'active'
    )
ON CONFLICT (org_id, email) DO NOTHING;

-- ============================================================================
-- SAMPLE AUDIT LOG ENTRIES
-- ============================================================================

INSERT INTO audit_log (org_id, actor_id, action, resource_type, resource_id, metadata, ip_address)
VALUES
    -- Org creation
    (
        'a0000000-0000-0000-0000-000000000001',
        NULL,
        'org.created',
        'organization',
        'a0000000-0000-0000-0000-000000000001',
        '{"source": "seed", "name": "Acme Corporation"}'::jsonb,
        '127.0.0.1'::inet
    ),
    -- CISO registered
    (
        'a0000000-0000-0000-0000-000000000001',
        'b0000000-0000-0000-0000-000000000004',
        'user.register',
        'user',
        'b0000000-0000-0000-0000-000000000004',
        '{"email": "ciso@acme.example.com", "role": "ciso"}'::jsonb,
        '127.0.0.1'::inet
    ),
    -- CISO assigned roles to other users
    (
        'a0000000-0000-0000-0000-000000000001',
        'b0000000-0000-0000-0000-000000000004',
        'user.role_assigned',
        'user',
        'b0000000-0000-0000-0000-000000000001',
        '{"email": "compliance@acme.example.com", "role": "compliance_manager", "assigned_by": "ciso"}'::jsonb,
        '127.0.0.1'::inet
    ),
    (
        'a0000000-0000-0000-0000-000000000001',
        'b0000000-0000-0000-0000-000000000004',
        'user.role_assigned',
        'user',
        'b0000000-0000-0000-0000-000000000002',
        '{"email": "security@acme.example.com", "role": "security_engineer", "assigned_by": "ciso"}'::jsonb,
        '127.0.0.1'::inet
    ),
    -- Sample login
    (
        'a0000000-0000-0000-0000-000000000001',
        'b0000000-0000-0000-0000-000000000001',
        'user.login',
        'user',
        'b0000000-0000-0000-0000-000000000001',
        '{"email": "compliance@acme.example.com", "method": "password"}'::jsonb,
        '192.168.1.10'::inet
    ),
    -- Sample failed login
    (
        'a0000000-0000-0000-0000-000000000001',
        NULL,
        'user.login_failed',
        'user',
        NULL,
        '{"email": "unknown@acme.example.com", "reason": "user_not_found"}'::jsonb,
        '10.0.0.55'::inet
    )
ON CONFLICT DO NOTHING;
