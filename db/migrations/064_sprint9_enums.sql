-- Sprint 9 -- Integration Engine Enum Types

-- Integration category
DO $$ BEGIN
    CREATE TYPE integration_category AS ENUM (
        'cloud_infrastructure',
        'identity_provider',
        'version_control',
        'communication',
        'monitoring',
        'ticketing',
        'custom'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Integration auth type
DO $$ BEGIN
    CREATE TYPE integration_auth_type AS ENUM (
        'api_key',
        'oauth2',
        'basic_auth',
        'token',
        'webhook_secret',
        'none'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Connection status
DO $$ BEGIN
    CREATE TYPE connection_status AS ENUM (
        'pending',
        'connected',
        'disconnected',
        'error',
        'disabled'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Connection health
DO $$ BEGIN
    CREATE TYPE connection_health AS ENUM (
        'healthy',
        'degraded',
        'unhealthy',
        'unknown'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Run status
DO $$ BEGIN
    CREATE TYPE run_status AS ENUM (
        'pending',
        'running',
        'completed',
        'partial',
        'failed',
        'cancelled'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Run trigger
DO $$ BEGIN
    CREATE TYPE run_trigger AS ENUM (
        'manual',
        'scheduled',
        'webhook',
        'system'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Log level
DO $$ BEGIN
    CREATE TYPE log_level AS ENUM (
        'debug',
        'info',
        'warn',
        'error'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Webhook status
DO $$ BEGIN
    CREATE TYPE webhook_status AS ENUM (
        'active',
        'inactive',
        'suspended'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Audit action values for integration events
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration.viewed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.updated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.tested';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.enabled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_connection.disabled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_sync.triggered';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_sync.completed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_sync.failed';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_sync.cancelled';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_webhook.created';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_webhook.deleted';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_webhook.secret_rotated';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_webhook.received';
ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'integration_health.checked';
