-- Sprint 9 -- Seed: Integration Definitions (system catalog)
INSERT INTO integration_definitions (id, name, slug, provider, category, description, short_description, icon_url, auth_type, config_schema, capabilities, is_active, is_beta, version, tags)
VALUES
    -- AWS Config
    ('d0000000-0000-0000-0000-000000000001', 'AWS Config', 'aws-config', 'aws_config', 'cloud_infrastructure',
     'Monitor AWS resource configurations and compliance. Automatically collect evidence from AWS Config rules, Security Hub findings, and CloudTrail events.',
     'AWS resource configuration monitoring',
     '/icons/integrations/aws.svg',
     'api_key',
     '{"type":"object","required":["access_key_id","secret_access_key","region"],"properties":{"access_key_id":{"type":"string","title":"Access Key ID","description":"AWS IAM access key ID","pattern":"^AKIA[0-9A-Z]{16}$"},"secret_access_key":{"type":"string","title":"Secret Access Key","description":"AWS IAM secret access key","format":"password"},"region":{"type":"string","title":"AWS Region","description":"Primary AWS region","enum":["us-east-1","us-west-2","eu-west-1","eu-central-1","ap-southeast-1"]},"assume_role_arn":{"type":"string","title":"Assume Role ARN","description":"Optional cross-account role ARN","pattern":"^arn:aws:iam::[0-9]{12}:role/.+$"}}}',
     ARRAY['evidence_collection', 'resource_inventory', 'compliance_check', 'cloud_config'],
     true, false, '1.0.0',
     ARRAY['cloud', 'aws', 'infrastructure', 'compliance']),

    -- GitHub
    ('d0000000-0000-0000-0000-000000000002', 'GitHub', 'github', 'github', 'version_control',
     'Connect to GitHub for repository security scanning, code review compliance, and development workflow monitoring.',
     'Repository and code security monitoring',
     '/icons/integrations/github.svg',
     'oauth2',
     '{"type":"object","required":["access_token"],"properties":{"access_token":{"type":"string","title":"Personal Access Token","description":"GitHub PAT with repo and admin:org scopes","format":"password"},"organization":{"type":"string","title":"Organization","description":"GitHub organization name"},"webhook_events":{"type":"array","title":"Webhook Events","description":"Events to subscribe to","items":{"type":"string","enum":["push","pull_request","security_advisory","repository_vulnerability_alert"]}}}}',
     ARRAY['evidence_collection', 'code_review', 'vulnerability_scan', 'webhook_inbound'],
     true, false, '1.0.0',
     ARRAY['devops', 'github', 'code', 'security']),

    -- Okta
    ('d0000000-0000-0000-0000-000000000003', 'Okta', 'okta', 'okta', 'identity_provider',
     'Sync users, groups, and applications from Okta. Automate access reviews with real-time identity data.',
     'Identity and access management sync',
     '/icons/integrations/okta.svg',
     'token',
     '{"type":"object","required":["domain","api_token"],"properties":{"domain":{"type":"string","title":"Okta Domain","description":"Your Okta domain (e.g., dev-12345.okta.com)","pattern":"^[a-zA-Z0-9-]+\\.okta\\.com$"},"api_token":{"type":"string","title":"API Token","description":"Okta API token with read permissions","format":"password"},"sync_groups":{"type":"boolean","title":"Sync Groups","description":"Include group memberships in sync","default":true},"sync_apps":{"type":"boolean","title":"Sync Applications","description":"Include application assignments in sync","default":true}}}',
     ARRAY['identity_sync', 'user_provisioning', 'group_sync', 'access_review'],
     true, false, '1.0.0',
     ARRAY['identity', 'okta', 'sso', 'access']),

    -- Slack
    ('d0000000-0000-0000-0000-000000000004', 'Slack', 'slack', 'slack', 'communication',
     'Send compliance alerts and notifications to Slack channels. Get real-time updates on policy changes, risk events, and audit activities.',
     'Compliance notifications and alerts',
     '/icons/integrations/slack.svg',
     'oauth2',
     '{"type":"object","required":["bot_token"],"properties":{"bot_token":{"type":"string","title":"Bot Token","description":"Slack Bot OAuth token (xoxb-...)","format":"password","pattern":"^xoxb-"},"default_channel":{"type":"string","title":"Default Channel","description":"Default channel for notifications (e.g., #compliance-alerts)"},"notify_on":{"type":"array","title":"Notification Events","description":"Events that trigger Slack notifications","items":{"type":"string","enum":["policy_published","risk_escalated","audit_finding","evidence_overdue","control_failed","alert_critical"]}}}}',
     ARRAY['notification', 'alert_delivery', 'chat_ops'],
     true, false, '1.0.0',
     ARRAY['communication', 'slack', 'notifications', 'alerts']),

    -- Custom Webhook
    ('d0000000-0000-0000-0000-000000000005', 'Custom Webhook', 'custom-webhook', 'custom_webhook', 'custom',
     'Configure a custom webhook integration for any service. Send and receive events via HTTP webhooks with HMAC signature verification.',
     'Custom HTTP webhook integration',
     '/icons/integrations/webhook.svg',
     'webhook_secret',
     '{"type":"object","required":["endpoint_url"],"properties":{"endpoint_url":{"type":"string","title":"Endpoint URL","description":"URL to send webhook events to","format":"uri"},"secret":{"type":"string","title":"Webhook Secret","description":"Shared secret for HMAC signature verification","format":"password"},"headers":{"type":"object","title":"Custom Headers","description":"Additional HTTP headers to include"},"retry_count":{"type":"integer","title":"Max Retries","description":"Maximum retry attempts for failed deliveries","default":3,"minimum":0,"maximum":10}}}',
     ARRAY['webhook_inbound', 'webhook_outbound', 'custom'],
     true, false, '1.0.0',
     ARRAY['custom', 'webhook', 'api'])
ON CONFLICT (slug) DO NOTHING;
