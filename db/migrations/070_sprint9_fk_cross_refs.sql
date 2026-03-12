-- Sprint 9 -- Cross-reference: link identity_providers to integration_connections
ALTER TABLE identity_providers
    ADD COLUMN IF NOT EXISTS integration_connection_id UUID REFERENCES integration_connections(id);

CREATE INDEX IF NOT EXISTS idx_identity_providers_integration_connection
    ON identity_providers (integration_connection_id)
    WHERE integration_connection_id IS NOT NULL;
