-- Sprint 8 — Seed Data: Demo Data (Acme Corp)

-- 1. Create specific access resources linked to the providers
DO $$
DECLARE
    org_id UUID := 'a0000000-0000-0000-0000-000000000001';
    okta_id UUID;
    azure_id UUID;
    bob_id UUID := 'b0000000-0000-0000-0000-000000000002'; -- CISO
    carol_id UUID := 'b0000000-0000-0000-0000-000000000003'; -- Compliance Mgr
    eve_id UUID := 'b0000000-0000-0000-0000-000000000005'; -- Dev Lead
BEGIN
    SELECT id INTO okta_id FROM identity_providers WHERE org_id = org_id AND name = 'Okta Workforce Identity';
    SELECT id INTO azure_id FROM identity_providers WHERE org_id = org_id AND name = 'Azure AD / Entra ID';

    -- Slack
    INSERT INTO access_resources (org_id, identity_provider_id, external_id, name, description, resource_type, criticality, department, category, owner_id)
    VALUES (org_id, okta_id, 'app_slack_001', 'Slack (Enterprise)', 'Corporate messaging and collaboration.', 'application', 'medium', 'IT', 'Collaboration', bob_id)
    ON CONFLICT (org_id, identity_provider_id, external_id) DO NOTHING;

    -- AWS Production
    INSERT INTO access_resources (org_id, identity_provider_id, external_id, name, description, resource_type, criticality, department, category, owner_id)
    VALUES (org_id, azure_id, 'aws_prod_001', 'AWS Production Console', 'Cloud infrastructure for production services.', 'infrastructure', 'critical', 'Engineering', 'Cloud Infrastructure', eve_id)
    ON CONFLICT (org_id, identity_provider_id, external_id) DO NOTHING;

    -- GitHub (Org)
    INSERT INTO access_resources (org_id, identity_provider_id, external_id, name, description, resource_type, criticality, department, category, owner_id)
    VALUES (org_id, okta_id, 'app_github_001', 'GitHub (Acme-Corp)', 'Source code management and CI/CD.', 'application', 'high', 'Engineering', 'Developer Tools', eve_id)
    ON CONFLICT (org_id, identity_provider_id, external_id) DO NOTHING;
END $$;
