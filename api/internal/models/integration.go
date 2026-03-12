package models

import "time"

// Integration category constants.
const (
	IntegrationCategoryCloudInfra   = "cloud_infrastructure"
	IntegrationCategoryIdentity     = "identity_provider"
	IntegrationCategoryVersionCtrl  = "version_control"
	IntegrationCategoryCommunication = "communication"
	IntegrationCategoryMonitoring   = "monitoring"
	IntegrationCategoryTicketing    = "ticketing"
	IntegrationCategoryCustom       = "custom"
)

// Integration auth type constants.
const (
	IntegrationAuthAPIKey       = "api_key"
	IntegrationAuthOAuth2       = "oauth2"
	IntegrationAuthBasicAuth    = "basic_auth"
	IntegrationAuthToken        = "token"
	IntegrationAuthWebhookSecret = "webhook_secret"
	IntegrationAuthNone         = "none"
)

// Connection status constants.
const (
	ConnectionStatusPending      = "pending"
	ConnectionStatusConnected    = "connected"
	ConnectionStatusDisconnected = "disconnected"
	ConnectionStatusError        = "error"
	ConnectionStatusDisabled     = "disabled"
)

// Connection health constants.
const (
	ConnectionHealthHealthy   = "healthy"
	ConnectionHealthDegraded  = "degraded"
	ConnectionHealthUnhealthy = "unhealthy"
	ConnectionHealthUnknown   = "unknown"
)

// Run status constants.
const (
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusPartial   = "partial"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"
)

// Run trigger constants.
const (
	RunTriggerManual    = "manual"
	RunTriggerScheduled = "scheduled"
	RunTriggerWebhook   = "webhook"
	RunTriggerSystem    = "system"
)

// Log level constants.
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Webhook status constants.
const (
	WebhookStatusActive    = "active"
	WebhookStatusInactive  = "inactive"
	WebhookStatusSuspended = "suspended"
)

// Validation helpers.

var validConnectionStatuses = []string{
	ConnectionStatusPending, ConnectionStatusConnected, ConnectionStatusDisconnected,
	ConnectionStatusError, ConnectionStatusDisabled,
}

func IsValidConnectionStatus(s string) bool {
	for _, v := range validConnectionStatuses {
		if v == s {
			return true
		}
	}
	return false
}

var validRunStatuses = []string{
	RunStatusPending, RunStatusRunning, RunStatusCompleted,
	RunStatusPartial, RunStatusFailed, RunStatusCancelled,
}

func IsValidRunStatus(s string) bool {
	for _, v := range validRunStatuses {
		if v == s {
			return true
		}
	}
	return false
}

var validWebhookStatuses = []string{WebhookStatusActive, WebhookStatusInactive, WebhookStatusSuspended}

func IsValidWebhookStatus(s string) bool {
	for _, v := range validWebhookStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// Role permission lists for integration operations.
var IntegrationViewRoles = []string{
	RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleITAdmin, RoleDevOpsEngineer,
}

var IntegrationManageRoles = []string{
	RoleCISO, RoleComplianceManager, RoleITAdmin, RoleDevOpsEngineer,
}

var IntegrationDeleteRoles = []string{
	RoleCISO, RoleITAdmin,
}

// IntegrationDefinition represents a system-level integration catalog entry.
type IntegrationDefinition struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Provider         string                 `json:"provider"`
	Category         string                 `json:"category"`
	Description      *string                `json:"description,omitempty"`
	ShortDescription *string                `json:"short_description,omitempty"`
	IconURL          *string                `json:"icon_url,omitempty"`
	DocumentationURL *string                `json:"documentation_url,omitempty"`
	WebsiteURL       *string                `json:"website_url,omitempty"`
	AuthType         string                 `json:"auth_type"`
	ConfigSchema     map[string]interface{} `json:"config_schema"`
	Capabilities     []string               `json:"capabilities"`
	IsActive         bool                   `json:"is_active"`
	IsBeta           bool                   `json:"is_beta"`
	Version          string                 `json:"version"`
	Tags             []string               `json:"tags"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// IntegrationConnection represents a per-org integration connection instance.
type IntegrationConnection struct {
	ID                  string                 `json:"id"`
	OrgID               string                 `json:"org_id"`
	DefinitionID        string                 `json:"definition_id"`
	Name                string                 `json:"name"`
	Description         *string                `json:"description,omitempty"`
	InstanceLabel       *string                `json:"instance_label,omitempty"`
	Status              string                 `json:"status"`
	Health              string                 `json:"health"`
	Config              map[string]interface{} `json:"config"`
	SyncEnabled         bool                   `json:"sync_enabled"`
	SyncIntervalMins    int                    `json:"sync_interval_mins"`
	SyncCron            *string                `json:"sync_cron,omitempty"`
	NextSyncAt          *time.Time             `json:"next_sync_at,omitempty"`
	LastSyncAt          *time.Time             `json:"last_sync_at,omitempty"`
	LastSyncStatus      *string                `json:"last_sync_status,omitempty"`
	LastSyncError       *string                `json:"last_sync_error,omitempty"`
	LastHealthCheckAt   *time.Time             `json:"last_health_check_at,omitempty"`
	LastHealthStatus    *string                `json:"last_health_status,omitempty"`
	ConsecutiveFailures int                    `json:"consecutive_failures"`
	TotalRuns           int                    `json:"total_runs"`
	SuccessfulRuns      int                    `json:"successful_runs"`
	FailedRuns          int                    `json:"failed_runs"`
	CreatedBy           *string                `json:"created_by,omitempty"`
	Tags                []string               `json:"tags"`
	Metadata            map[string]interface{} `json:"metadata"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	// Joined fields
	DefinitionName     string `json:"definition_name,omitempty"`
	DefinitionSlug     string `json:"definition_slug,omitempty"`
	DefinitionCategory string `json:"definition_category,omitempty"`
	DefinitionIconURL  string `json:"definition_icon_url,omitempty"`
}

// IntegrationRun represents a sync execution record.
type IntegrationRun struct {
	ID            string                 `json:"id"`
	OrgID         string                 `json:"org_id"`
	ConnectionID  string                 `json:"connection_id"`
	Trigger       string                 `json:"trigger"`
	TriggeredBy   *string                `json:"triggered_by,omitempty"`
	Status        string                 `json:"status"`
	QueuedAt      time.Time              `json:"queued_at"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	DurationMs    *int                   `json:"duration_ms,omitempty"`
	Stats         map[string]interface{} `json:"stats"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	ErrorDetails  map[string]interface{} `json:"error_details,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	MaxRetries    int                    `json:"max_retries"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// IntegrationLog represents a log entry for an integration run.
type IntegrationLog struct {
	ID           string                 `json:"id"`
	OrgID        string                 `json:"org_id"`
	RunID        string                 `json:"run_id"`
	ConnectionID string                 `json:"connection_id"`
	Level        string                 `json:"level"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Source       *string                `json:"source,omitempty"`
	ItemRef      *string                `json:"item_ref,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// IntegrationWebhook represents a webhook endpoint for an integration connection.
type IntegrationWebhook struct {
	ID              string     `json:"id"`
	OrgID           string     `json:"org_id"`
	ConnectionID    string     `json:"connection_id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	WebhookSecret   string     `json:"-"`
	SignatureHeader string     `json:"signature_header"`
	SignatureAlgo   string     `json:"signature_algo"`
	Status          string     `json:"status"`
	EventTypes      []string   `json:"event_types"`
	TotalReceived   int        `json:"total_received"`
	TotalProcessed  int        `json:"total_processed"`
	TotalErrors     int        `json:"total_errors"`
	LastReceivedAt  *time.Time `json:"last_received_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateConnectionRequest is the request body for creating an integration connection.
type CreateConnectionRequest struct {
	DefinitionID     string                 `json:"definition_id" binding:"required"`
	Name             string                 `json:"name" binding:"required"`
	Description      *string                `json:"description"`
	InstanceLabel    *string                `json:"instance_label"`
	Config           map[string]interface{} `json:"config"`
	SyncEnabled      *bool                  `json:"sync_enabled"`
	SyncIntervalMins *int                   `json:"sync_interval_mins"`
	SyncCron         *string                `json:"sync_cron"`
	Tags             []string               `json:"tags"`
}

// UpdateConnectionRequest is the request body for updating an integration connection.
type UpdateConnectionRequest struct {
	Name             *string                `json:"name"`
	Description      *string                `json:"description"`
	InstanceLabel    *string                `json:"instance_label"`
	Config           map[string]interface{} `json:"config"`
	SyncEnabled      *bool                  `json:"sync_enabled"`
	SyncIntervalMins *int                   `json:"sync_interval_mins"`
	SyncCron         *string                `json:"sync_cron"`
	Tags             []string               `json:"tags"`
}

// CreateWebhookRequest is the request body for creating a webhook.
type CreateWebhookRequest struct {
	Name            string   `json:"name" binding:"required"`
	Description     *string  `json:"description"`
	SignatureHeader *string  `json:"signature_header"`
	SignatureAlgo   *string  `json:"signature_algo"`
	EventTypes      []string `json:"event_types"`
}

// TriggerSyncRequest is the request body for triggering a sync.
type TriggerSyncRequest struct {
	Force bool `json:"force"`
}
