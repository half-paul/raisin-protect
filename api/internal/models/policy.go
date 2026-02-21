package models

import "time"

// Policy category constants.
const (
	PolicyCategoryInformationSecurity  = "information_security"
	PolicyCategoryAcceptableUse        = "acceptable_use"
	PolicyCategoryAccessControl        = "access_control"
	PolicyCategoryDataClassification   = "data_classification"
	PolicyCategoryDataPrivacy          = "data_privacy"
	PolicyCategoryDataRetention        = "data_retention"
	PolicyCategoryIncidentResponse     = "incident_response"
	PolicyCategoryBusinessContinuity   = "business_continuity"
	PolicyCategoryChangeManagement     = "change_management"
	PolicyCategoryVulnerabilityMgmt    = "vulnerability_management"
	PolicyCategoryVendorManagement     = "vendor_management"
	PolicyCategoryPhysicalSecurity     = "physical_security"
	PolicyCategoryEncryption           = "encryption"
	PolicyCategoryNetworkSecurity      = "network_security"
	PolicyCategorySecureDevelopment    = "secure_development"
	PolicyCategoryHumanResources       = "human_resources"
	PolicyCategoryCompliance           = "compliance"
	PolicyCategoryRiskManagement       = "risk_management"
	PolicyCategoryAssetManagement      = "asset_management"
	PolicyCategoryLoggingMonitoring    = "logging_monitoring"
	PolicyCategoryCustom               = "custom"
)

// ValidPolicyCategories is the complete list.
var ValidPolicyCategories = []string{
	PolicyCategoryInformationSecurity, PolicyCategoryAcceptableUse, PolicyCategoryAccessControl,
	PolicyCategoryDataClassification, PolicyCategoryDataPrivacy, PolicyCategoryDataRetention,
	PolicyCategoryIncidentResponse, PolicyCategoryBusinessContinuity, PolicyCategoryChangeManagement,
	PolicyCategoryVulnerabilityMgmt, PolicyCategoryVendorManagement, PolicyCategoryPhysicalSecurity,
	PolicyCategoryEncryption, PolicyCategoryNetworkSecurity, PolicyCategorySecureDevelopment,
	PolicyCategoryHumanResources, PolicyCategoryCompliance, PolicyCategoryRiskManagement,
	PolicyCategoryAssetManagement, PolicyCategoryLoggingMonitoring, PolicyCategoryCustom,
}

// IsValidPolicyCategory checks if a category string is valid.
func IsValidPolicyCategory(cat string) bool {
	for _, c := range ValidPolicyCategories {
		if c == cat {
			return true
		}
	}
	return false
}

// Policy status constants.
const (
	PolicyStatusDraft     = "draft"
	PolicyStatusInReview  = "in_review"
	PolicyStatusApproved  = "approved"
	PolicyStatusPublished = "published"
	PolicyStatusArchived  = "archived"
)

// ValidPolicyStatuses is the complete list.
var ValidPolicyStatuses = []string{
	PolicyStatusDraft, PolicyStatusInReview, PolicyStatusApproved,
	PolicyStatusPublished, PolicyStatusArchived,
}

// IsValidPolicyStatus checks if a status string is valid.
func IsValidPolicyStatus(s string) bool {
	for _, v := range ValidPolicyStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// Content format constants.
const (
	ContentFormatHTML      = "html"
	ContentFormatMarkdown  = "markdown"
	ContentFormatPlainText = "plain_text"
)

// ValidContentFormats lists valid content formats.
var ValidContentFormats = []string{ContentFormatHTML, ContentFormatMarkdown, ContentFormatPlainText}

// IsValidContentFormat checks if a content format is valid.
func IsValidContentFormat(f string) bool {
	for _, v := range ValidContentFormats {
		if v == f {
			return true
		}
	}
	return false
}

// Sign-off status constants.
const (
	SignoffStatusPending   = "pending"
	SignoffStatusApproved  = "approved"
	SignoffStatusRejected  = "rejected"
	SignoffStatusWithdrawn = "withdrawn"
)

// Policy roles.
var PolicyCreateRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}
var PolicyPublishRoles = []string{RoleCISO, RoleComplianceManager}
var PolicyArchiveRoles = []string{RoleCISO, RoleComplianceManager}
var PolicyGapRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleAuditor}

// Policy represents a policy document in the database.
type Policy struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"org_id"`
	Identifier          string     `json:"identifier"`
	Title               string     `json:"title"`
	Description         *string    `json:"description"`
	Category            string     `json:"category"`
	Status              string     `json:"status"`
	CurrentVersionID    *string    `json:"current_version_id"`
	OwnerID             *string    `json:"owner_id"`
	SecondaryOwnerID    *string    `json:"secondary_owner_id"`
	ReviewFrequencyDays *int       `json:"review_frequency_days"`
	NextReviewAt        *time.Time `json:"next_review_at"`
	LastReviewedAt      *time.Time `json:"last_reviewed_at"`
	IsTemplate          bool       `json:"is_template"`
	TemplateFrameworkID *string    `json:"template_framework_id"`
	ClonedFromPolicyID  *string    `json:"cloned_from_policy_id"`
	ApprovedAt          *time.Time `json:"approved_at"`
	ApprovedVersion     *int       `json:"approved_version"`
	PublishedAt         *time.Time `json:"published_at"`
	Tags                []string   `json:"tags"`
	Metadata            string     `json:"metadata"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PolicyVersion represents a version of a policy's content.
type PolicyVersion struct {
	ID             string    `json:"id"`
	OrgID          string    `json:"org_id"`
	PolicyID       string    `json:"policy_id"`
	VersionNumber  int       `json:"version_number"`
	IsCurrent      bool      `json:"is_current"`
	Content        string    `json:"content"`
	ContentFormat  string    `json:"content_format"`
	ContentSummary *string   `json:"content_summary"`
	ChangeSummary  *string   `json:"change_summary"`
	ChangeType     string    `json:"change_type"`
	WordCount      *int      `json:"word_count"`
	CharacterCount *int      `json:"character_count"`
	CreatedBy      *string   `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
}

// PolicySignoff represents a sign-off request for a policy version.
type PolicySignoff struct {
	ID              string     `json:"id"`
	OrgID           string     `json:"org_id"`
	PolicyID        string     `json:"policy_id"`
	PolicyVersionID string     `json:"policy_version_id"`
	SignerID        string     `json:"signer_id"`
	SignerRole      *string    `json:"signer_role"`
	RequestedBy     string     `json:"requested_by"`
	RequestedAt     time.Time  `json:"requested_at"`
	DueDate         *time.Time `json:"due_date"`
	Status          string     `json:"status"`
	DecidedAt       *time.Time `json:"decided_at"`
	Comments        *string    `json:"comments"`
	ReminderSentAt  *time.Time `json:"reminder_sent_at"`
	ReminderCount   int        `json:"reminder_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// PolicyControl represents a policy-to-control mapping.
type PolicyControl struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	PolicyID  string    `json:"policy_id"`
	ControlID string    `json:"control_id"`
	Notes     *string   `json:"notes"`
	Coverage  string    `json:"coverage"`
	LinkedBy  *string   `json:"linked_by"`
	CreatedAt time.Time `json:"created_at"`
}

// Request/response types.

// CreatePolicyRequest is the request for creating a policy.
type CreatePolicyRequest struct {
	Identifier        string   `json:"identifier" binding:"required"`
	Title             string   `json:"title" binding:"required"`
	Description       *string  `json:"description"`
	Category          string   `json:"category" binding:"required"`
	OwnerID           *string  `json:"owner_id"`
	SecondaryOwnerID  *string  `json:"secondary_owner_id"`
	ReviewFreqDays    *int     `json:"review_frequency_days"`
	Tags              []string `json:"tags"`
	Content           string   `json:"content" binding:"required"`
	ContentFormat     *string  `json:"content_format"`
	ContentSummary    *string  `json:"content_summary"`
}

// UpdatePolicyRequest is the request for updating policy metadata.
type UpdatePolicyRequest struct {
	Title             *string    `json:"title"`
	Description       *string    `json:"description"`
	Category          *string    `json:"category"`
	OwnerID           *string    `json:"owner_id"`
	SecondaryOwnerID  *string    `json:"secondary_owner_id"`
	ReviewFreqDays    *int       `json:"review_frequency_days"`
	NextReviewAt      *string    `json:"next_review_at"`
	Tags              []string   `json:"tags"`
}

// CreatePolicyVersionRequest is the request for creating a policy version.
type CreatePolicyVersionRequest struct {
	Content        string  `json:"content" binding:"required"`
	ContentFormat  *string `json:"content_format"`
	ContentSummary *string `json:"content_summary"`
	ChangeSummary  string  `json:"change_summary" binding:"required"`
	ChangeType     *string `json:"change_type"`
}

// SubmitForReviewRequest is the request for submitting a policy for review.
type SubmitForReviewRequest struct {
	SignerIDs []string `json:"signer_ids" binding:"required"`
	DueDate   *string  `json:"due_date"`
	Message   *string  `json:"message"`
}

// SignoffDecisionRequest is the request for approve/reject actions.
type SignoffDecisionRequest struct {
	Comments *string `json:"comments"`
}

// LinkControlRequest is the request for linking a control to a policy.
type LinkControlRequest struct {
	ControlID string  `json:"control_id" binding:"required"`
	Coverage  *string `json:"coverage"`
	Notes     *string `json:"notes"`
}

// BulkLinkControlsRequest is the request for bulk linking controls.
type BulkLinkControlsRequest struct {
	Links []LinkControlRequest `json:"links" binding:"required"`
}

// CloneTemplateRequest is the request for cloning a template.
type CloneTemplateRequest struct {
	Identifier     string   `json:"identifier" binding:"required"`
	Title          *string  `json:"title"`
	Description    *string  `json:"description"`
	OwnerID        *string  `json:"owner_id"`
	ReviewFreqDays *int     `json:"review_frequency_days"`
	Tags           []string `json:"tags"`
}

// RemindSignoffRequest is the request for sending reminders.
type RemindSignoffRequest struct {
	SignoffIDs []string `json:"signoff_ids"`
	Message    *string  `json:"message"`
}
