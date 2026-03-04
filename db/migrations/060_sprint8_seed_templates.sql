-- Sprint 8 — Seed Data: Templates

-- Identity Provider Templates
-- Note: Configs are stubs, in production these would be encrypted.
INSERT INTO identity_providers (org_id, name, provider_type, status, config, description, sync_interval_mins)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'Okta Workforce Identity', 'okta', 'pending_setup',
     '{"domain": "acme.okta.com", "api_token": "stubs_token_okta"}'::jsonb,
     'Primary workforce identity provider for all employees.', 360),
    ('a0000000-0000-0000-0000-000000000001', 'Azure AD / Entra ID', 'azure_ad', 'pending_setup',
     '{"tenant_id": "azure-tenant-uuid", "client_id": "azure-client-uuid", "client_secret": "stubs_secret_azure"}'::jsonb,
     'Microsoft 365 and Azure infrastructure identity.', 360),
    ('a0000000-0000-0000-0000-000000000001', 'Google Workspace', 'google_workspace', 'pending_setup',
     '{"domain": "acme.com", "service_account_key": "{}"}'::jsonb,
     'Corporate email and productivity suite identity.', 360)
ON CONFLICT (org_id, name) DO NOTHING;

-- Access Resource Templates (Sample Resources)
-- These will be used by the demo seed to create specific resources.
-- Here we just ensure we have some owners assigned.
-- (Owners are from seed.sql: Bob b00...02, Carol b00...03, Eve b00...05)
