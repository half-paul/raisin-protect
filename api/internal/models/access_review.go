package models

import (
	"time"

	"github.com/google/uuid"
)

// --- Enums as constants ---

type IdentityProviderType string
const (
	IdPOkta            IdentityProviderType = "okta"
	IdPAzureAD         IdentityProviderType = "azure_ad"
	IdPGoogleWorkspace IdentityProviderType = "google_workspace"
	IdPJumpCloud       IdentityProviderType = "jumpcloud"
	IdPOneLogin        IdentityProviderType = "onelogin"
	IdPCustom          IdentityProviderType = "custom"
)

type IdentityProviderStatus string
const (
	IdPStatusPending      IdentityProviderStatus = "pending_setup"
	IdPStatusConnected    IdentityProviderStatus = "connected"
	IdPStatusSyncing      IdentityProviderStatus = "syncing"
	IdPStatusError        IdentityProviderStatus = "error"
	IdPStatusDisconnected IdentityProviderStatus = "disconnected"
)

type ResourceCriticality string
const (
	CriticalityCritical ResourceCriticality = "critical"
	CriticalityHigh     ResourceCriticality = "high"
	CriticalityMedium   ResourceCriticality = "medium"
	CriticalityLow      ResourceCriticality = "low"
)

type ResourceType string
const (
	ResourceTypeApp    ResourceType = "application"
	ResourceTypeInfra  ResourceType = "infrastructure"
	ResourceTypeDir    ResourceType = "directory"
	ResourceTypeCustom ResourceType = "custom"
)

type CampaignStatus string
const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusInReview  CampaignStatus = "in_review"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type CampaignCadence string
const (
	CadenceMonthly    CampaignCadence = "monthly"
	CadenceQuarterly  CampaignCadence = "quarterly"
	CadenceSemiAnnual CampaignCadence = "semi_annual"
	CadenceAnnual     CampaignCadence = "annual"
	CadenceCustom     CampaignCadence = "custom"
)

type ReviewDecision string
const (
	DecisionPending   ReviewDecision = "pending"
	DecisionApproved  ReviewDecision = "approved"
	DecisionRevoked   ReviewDecision = "revoked"
	DecisionFlagged   ReviewDecision = "flagged"
	DecisionDelegated ReviewDecision = "delegated"
	DecisionExpired   ReviewDecision = "expired"
)

type AccessEntryStatus string
const (
	EntryStatusActive            AccessEntryStatus = "active"
	EntryStatusInactive          AccessEntryStatus = "inactive"
	EntryStatusOrphaned          AccessEntryStatus = "orphaned"
	EntryStatusSuspended         AccessEntryStatus = "suspended"
	EntryStatusPendingRevocation AccessEntryStatus = "pending_revocation"
)

// --- Models ---

type IdentityProvider struct {
	ID               uuid.UUID              `json:"id"`
	OrgID            uuid.UUID              `json:"org_id"`
	Name             string                 `json:"name"`
	ProviderType     IdentityProviderType   `json:"provider_type"`
	Status           IdentityProviderStatus `json:"status"`
	Config           map[string]interface{} `json:"config"`
	LastSyncAt       *time.Time             `json:"last_sync_at"`
	LastSyncStatus   *string                `json:"last_sync_status"`
	LastSyncError    *string                `json:"last_sync_error"`
	LastSyncStats    map[string]interface{} `json:"last_sync_stats"`
	SyncIntervalMins int                    `json:"sync_interval_mins"`
	Description      *string                `json:"description"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type AccessResource struct {
	ID                 uuid.UUID           `json:"id"`
	OrgID              uuid.UUID           `json:"org_id"`
	IdentityProviderID *uuid.UUID          `json:"identity_provider_id"`
	ExternalID         *string             `json:"external_id"`
	Name               string              `json:"name"`
	Description        *string             `json:"description"`
	ResourceType       ResourceType        `json:"resource_type"`
	Criticality        ResourceCriticality `json:"criticality"`
	Department         *string             `json:"department"`
	Category           *string             `json:"category"`
	Tags               []string            `json:"tags"`
	OwnerID            *uuid.UUID          `json:"owner_id"`
	TotalUsers         int                 `json:"total_users"`
	TotalRoles         int                 `json:"total_roles"`
	LastSyncAt         *time.Time          `json:"last_sync_at"`
	URL                *string             `json:"url"`
	LastReviewedAt     *time.Time          `json:"last_reviewed_at"`
	ReviewCadence      *CampaignCadence    `json:"review_cadence"`
	IsActive           bool                `json:"is_active"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

type AccessEntry struct {
	ID                 uuid.UUID              `json:"id"`
	OrgID              uuid.UUID              `json:"org_id"`
	IdentityProviderID *uuid.UUID             `json:"identity_provider_id"`
	ResourceID         uuid.UUID              `json:"resource_id"`
	ExternalUserID     *string                `json:"external_user_id"`
	UserEmail          string                 `json:"user_email"`
	UserDisplayName    string                 `json:"user_display_name"`
	UserDepartment     *string                `json:"user_department"`
	UserTitle          *string                `json:"user_title"`
	UserManagerEmail   *string                `json:"user_manager_email"`
	InternalUserID     *uuid.UUID             `json:"internal_user_id"`
	RoleName           string                 `json:"role_name"`
	AccessLevel        *string                `json:"access_level"`
	Permissions        map[string]interface{} `json:"permissions"`
	IsPrivileged       bool                   `json:"is_privileged"`
	ExpectedRole       *string                `json:"expected_role"`
	ExpectedLevel      *string                `json:"expected_access_level"`
	HasRoleDrift       bool                   `json:"has_role_drift"`
	GrantedAt          *time.Time             `json:"granted_at"`
	LastUsedAt         *time.Time             `json:"last_used_at"`
	LastLoginAt        *time.Time             `json:"last_login_at"`
	Status             AccessEntryStatus      `json:"status"`
	MFAEnabled         *bool                  `json:"mfa_enabled"`
	Anomalies          []interface{}          `json:"anomalies"`
	LastSyncAt         *time.Time             `json:"last_sync_at"`
	IsServiceAccount   bool                   `json:"is_service_account"`
	Notes              *string                `json:"notes"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type AccessReviewCampaign struct {
	ID                uuid.UUID              `json:"id"`
	OrgID             uuid.UUID              `json:"org_id"`
	Name              string                 `json:"name"`
	Description       *string                `json:"description"`
	Status            CampaignStatus         `json:"status"`
	Cadence           CampaignCadence        `json:"cadence"`
	Scope             map[string]interface{} `json:"scope"`
	ReviewerStrategy  string                 `json:"reviewer_strategy"`
	DefaultReviewerID *uuid.UUID             `json:"default_reviewer_id"`
	StartedAt         *time.Time             `json:"started_at"`
	Deadline          time.Time              `json:"deadline"`
	CompletedAt       *time.Time             `json:"completed_at"`
	CancelledAt       *time.Time             `json:"cancelled_at"`
	EscalationConfig  map[string]interface{} `json:"escalation_config"`
	TotalReviews      int                    `json:"total_reviews"`
	CompletedReviews  int                    `json:"completed_reviews"`
	ApprovedCount     int                    `json:"approved_count"`
	RevokedCount      int                    `json:"revoked_count"`
	FlaggedCount      int                    `json:"flagged_count"`
	CreatedBy         uuid.UUID              `json:"created_by"`
	Tags              []string               `json:"tags"`
	Notes             *string                `json:"notes"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type AccessReview struct {
	ID                   uuid.UUID      `json:"id"`
	OrgID                uuid.UUID      `json:"org_id"`
	CampaignID           uuid.UUID      `json:"campaign_id"`
	EntryID              uuid.UUID      `json:"entry_id"`
	ReviewerID           *uuid.UUID     `json:"reviewer_id"`
	AssignedAt           *time.Time     `json:"assigned_at"`
	Decision             ReviewDecision `json:"decision"`
	Justification        *string        `json:"justification"`
	DecidedBy            *uuid.UUID     `json:"decided_by"`
	DecidedAt            *time.Time     `json:"decided_at"`
	DelegatedToID        *uuid.UUID     `json:"delegated_to_id"`
	DelegatedAt          *time.Time     `json:"delegated_at"`
	DelegationReason     *string        `json:"delegation_reason"`
	IsEscalated          bool           `json:"is_escalated"`
	EscalatedAt          *time.Time     `json:"escalated_at"`
	EscalatedToID        *uuid.UUID     `json:"escalated_to_id"`
	AccessSnapshot       interface{}    `json:"access_snapshot"`
	RevocationExecuted   bool           `json:"revocation_executed"`
	RevocationExecutedAt *time.Time     `json:"revocation_executed_at"`
	RevocationNotes      *string        `json:"revocation_notes"`
	Notes                *string        `json:"notes"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

// --- RBAC Role Slices ---

var (
	AccessReviewViewRoles = []string{"compliance_manager", "security_engineer", "ciso", "auditor"}
	AccessReviewAdminRoles = []string{"compliance_manager", "ciso"}
	AccessReviewReviewerRoles = []string{"compliance_manager", "security_engineer", "ciso", "auditor", "it_admin", "dev_lead", "department_head"}

	IdPManageRoles = []string{"compliance_manager", "ciso", "it_admin"}
	ResourceManageRoles = []string{"compliance_manager", "ciso", "it_admin", "security_engineer"}
	CampaignManageRoles = []string{"compliance_manager", "ciso"}
)
