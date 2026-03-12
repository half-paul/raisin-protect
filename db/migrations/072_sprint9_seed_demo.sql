-- Sprint 9 -- Seed: Demo integration data
DO $$
DECLARE
    demo_org UUID := 'a0000000-0000-0000-0000-000000000001';
    user_ciso UUID := 'a0000000-0000-0000-0000-000000000101';
    user_it UUID := 'a0000000-0000-0000-0000-000000000104';
    conn_okta UUID := 'c9000000-0000-0000-0000-000000000001';
    conn_github UUID := 'c9000000-0000-0000-0000-000000000002';
    conn_slack UUID := 'c9000000-0000-0000-0000-000000000003';
    run1 UUID := 'r9000000-0000-0000-0000-000000000001';
    run2 UUID := 'r9000000-0000-0000-0000-000000000002';
    run3 UUID := 'r9000000-0000-0000-0000-000000000003';
    run4 UUID := 'r9000000-0000-0000-0000-000000000004';
    run5 UUID := 'r9000000-0000-0000-0000-000000000005';
    wh1 UUID := 'w9000000-0000-0000-0000-000000000001';
BEGIN
    -- Demo connections
    INSERT INTO integration_connections (id, org_id, definition_id, name, description, status, health, config, sync_enabled, sync_interval_mins, last_sync_at, last_sync_status, total_runs, successful_runs, failed_runs, created_by)
    VALUES
        (conn_okta, demo_org, 'd0000000-0000-0000-0000-000000000003', 'Okta Production', 'Main Okta tenant for identity sync', 'connected', 'healthy',
         '{"domain":"acme-demo.okta.com","api_token":"****","sync_groups":true,"sync_apps":true}',
         true, 360, NOW() - INTERVAL '2 hours', 'completed', 12, 11, 1, user_it),

        (conn_github, demo_org, 'd0000000-0000-0000-0000-000000000002', 'GitHub Enterprise', 'GitHub org for code security monitoring', 'connected', 'healthy',
         '{"access_token":"****","organization":"acme-corp","webhook_events":["push","pull_request","security_advisory"]}',
         true, 720, NOW() - INTERVAL '4 hours', 'completed', 8, 8, 0, user_it),

        (conn_slack, demo_org, 'd0000000-0000-0000-0000-000000000004', 'Slack Notifications', 'Compliance alert channel', 'connected', 'healthy',
         '{"bot_token":"****","default_channel":"#compliance-alerts","notify_on":["policy_published","risk_escalated","alert_critical"]}',
         false, 360, NULL, NULL, 3, 3, 0, user_ciso)
    ON CONFLICT (org_id, name) DO NOTHING;

    -- Demo runs
    INSERT INTO integration_runs (id, org_id, connection_id, trigger, triggered_by, status, queued_at, started_at, completed_at, duration_ms, stats)
    VALUES
        (run1, demo_org, conn_okta, 'scheduled', NULL, 'completed',
         NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours' + INTERVAL '5 seconds', NOW() - INTERVAL '2 hours' + INTERVAL '45 seconds', 40000,
         '{"users_synced":156,"groups_synced":12,"apps_synced":8,"changes_detected":3}'),

        (run2, demo_org, conn_okta, 'scheduled', NULL, 'completed',
         NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours' + INTERVAL '3 seconds', NOW() - INTERVAL '8 hours' + INTERVAL '38 seconds', 35000,
         '{"users_synced":155,"groups_synced":12,"apps_synced":8,"changes_detected":0}'),

        (run3, demo_org, conn_github, 'scheduled', NULL, 'completed',
         NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours' + INTERVAL '2 seconds', NOW() - INTERVAL '4 hours' + INTERVAL '22 seconds', 20000,
         '{"repos_scanned":24,"vulnerabilities_found":2,"pull_requests_checked":15}'),

        (run4, demo_org, conn_okta, 'manual', user_it, 'partial',
         NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '4 seconds', NOW() - INTERVAL '1 day' + INTERVAL '30 seconds', 26000,
         '{"users_synced":140,"groups_synced":12,"apps_synced":0,"error":"app_sync_timeout"}'),

        (run5, demo_org, conn_slack, 'manual', user_ciso, 'completed',
         NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days' + INTERVAL '1 second', NOW() - INTERVAL '3 days' + INTERVAL '3 seconds', 2000,
         '{"messages_sent":1,"channel_verified":true}')
    ON CONFLICT DO NOTHING;

    -- Demo logs
    INSERT INTO integration_logs (org_id, run_id, connection_id, level, message, source)
    VALUES
        (demo_org, run1, conn_okta, 'info', 'Starting Okta sync', 'okta_connector'),
        (demo_org, run1, conn_okta, 'info', 'Fetched 156 users from Okta', 'okta_connector'),
        (demo_org, run1, conn_okta, 'info', 'Fetched 12 groups from Okta', 'okta_connector'),
        (demo_org, run1, conn_okta, 'info', 'Detected 3 user changes since last sync', 'okta_connector'),
        (demo_org, run1, conn_okta, 'info', 'Sync completed successfully', 'okta_connector'),
        (demo_org, run3, conn_github, 'info', 'Starting GitHub repository scan', 'github_connector'),
        (demo_org, run3, conn_github, 'info', 'Scanned 24 repositories', 'github_connector'),
        (demo_org, run3, conn_github, 'warn', 'Found 2 high-severity vulnerabilities in acme-api', 'github_connector'),
        (demo_org, run3, conn_github, 'info', 'Checked 15 pull requests for compliance', 'github_connector'),
        (demo_org, run3, conn_github, 'info', 'Scan completed successfully', 'github_connector'),
        (demo_org, run4, conn_okta, 'info', 'Starting Okta sync (manual trigger)', 'okta_connector'),
        (demo_org, run4, conn_okta, 'info', 'Fetched 140 users from Okta', 'okta_connector'),
        (demo_org, run4, conn_okta, 'error', 'Timeout while syncing applications: connection reset', 'okta_connector'),
        (demo_org, run4, conn_okta, 'warn', 'Sync completed with partial results', 'okta_connector'),
        (demo_org, run5, conn_slack, 'info', 'Test message sent to #compliance-alerts', 'slack_connector');

    -- Demo webhook
    INSERT INTO integration_webhooks (id, org_id, connection_id, name, description, webhook_secret, signature_header, signature_algo, status, event_types, total_received, total_processed)
    VALUES
        (wh1, demo_org, conn_github, 'GitHub Push Events', 'Receives push events for compliance tracking', 'whsec_demo_secret_not_real', 'X-Hub-Signature-256', 'sha256', 'active',
         ARRAY['push', 'pull_request'], 42, 42)
    ON CONFLICT (connection_id, name) DO NOTHING;

END $$;
