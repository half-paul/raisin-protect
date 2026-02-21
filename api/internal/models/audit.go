package models

// --- Audit status constants (from audit_status enum) ---
const (
	AuditStatusPlanning           = "planning"
	AuditStatusFieldwork          = "fieldwork"
	AuditStatusReview             = "review"
	AuditStatusDraftReport        = "draft_report"
	AuditStatusManagementResponse = "management_response"
	AuditStatusFinalReport        = "final_report"
	AuditStatusCompleted          = "completed"
	AuditStatusCancelled          = "cancelled"
)

// ValidAuditStatuses lists all audit statuses.
var ValidAuditStatuses = []string{
	AuditStatusPlanning, AuditStatusFieldwork, AuditStatusReview,
	AuditStatusDraftReport, AuditStatusManagementResponse,
	AuditStatusFinalReport, AuditStatusCompleted, AuditStatusCancelled,
}

// IsValidAuditStatus checks if a status is valid.
func IsValidAuditStatus(s string) bool {
	for _, v := range ValidAuditStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// AuditStatusTransitions defines valid audit status transitions.
var AuditStatusTransitions = map[string][]string{
	AuditStatusPlanning:           {AuditStatusFieldwork, AuditStatusCancelled},
	AuditStatusFieldwork:          {AuditStatusReview, AuditStatusCancelled},
	AuditStatusReview:             {AuditStatusDraftReport, AuditStatusFieldwork, AuditStatusCancelled},
	AuditStatusDraftReport:        {AuditStatusManagementResponse, AuditStatusCancelled},
	AuditStatusManagementResponse: {AuditStatusFinalReport, AuditStatusDraftReport},
	AuditStatusFinalReport:        {AuditStatusCompleted},
	// completed and cancelled are terminal
}

// IsValidAuditStatusTransition checks if a transition is allowed.
func IsValidAuditStatusTransition(from, to string) bool {
	allowed, ok := AuditStatusTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

// IsAuditTerminal checks if an audit status is terminal.
func IsAuditTerminal(status string) bool {
	return status == AuditStatusCompleted || status == AuditStatusCancelled
}

// --- Audit type constants ---
const (
	AuditTypeSOC2Type1           = "soc2_type1"
	AuditTypeSOC2Type2           = "soc2_type2"
	AuditTypeISO27001Cert        = "iso27001_certification"
	AuditTypeISO27001Surveillance = "iso27001_surveillance"
	AuditTypePCIDSSROC           = "pci_dss_roc"
	AuditTypePCIDSSSAQ           = "pci_dss_saq"
	AuditTypeGDPRDPIA            = "gdpr_dpia"
	AuditTypeInternal            = "internal"
	AuditTypeCustom              = "custom"
)

var ValidAuditTypes = []string{
	AuditTypeSOC2Type1, AuditTypeSOC2Type2, AuditTypeISO27001Cert,
	AuditTypeISO27001Surveillance, AuditTypePCIDSSROC, AuditTypePCIDSSSAQ,
	AuditTypeGDPRDPIA, AuditTypeInternal, AuditTypeCustom,
}

func IsValidAuditType(t string) bool {
	for _, v := range ValidAuditTypes {
		if v == t {
			return true
		}
	}
	return false
}

// --- Audit request status constants ---
const (
	AuditRequestStatusOpen       = "open"
	AuditRequestStatusInProgress = "in_progress"
	AuditRequestStatusSubmitted  = "submitted"
	AuditRequestStatusAccepted   = "accepted"
	AuditRequestStatusRejected   = "rejected"
	AuditRequestStatusClosed     = "closed"
)

var ValidAuditRequestStatuses = []string{
	AuditRequestStatusOpen, AuditRequestStatusInProgress,
	AuditRequestStatusSubmitted, AuditRequestStatusAccepted,
	AuditRequestStatusRejected, AuditRequestStatusClosed,
}

func IsValidAuditRequestStatus(s string) bool {
	for _, v := range ValidAuditRequestStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// --- Audit request priority ---
const (
	AuditRequestPriorityCritical = "critical"
	AuditRequestPriorityHigh     = "high"
	AuditRequestPriorityMedium   = "medium"
	AuditRequestPriorityLow      = "low"
)

var ValidAuditRequestPriorities = []string{
	AuditRequestPriorityCritical, AuditRequestPriorityHigh,
	AuditRequestPriorityMedium, AuditRequestPriorityLow,
}

func IsValidAuditRequestPriority(p string) bool {
	for _, v := range ValidAuditRequestPriorities {
		if v == p {
			return true
		}
	}
	return false
}

// --- Finding severity ---
const (
	FindingSeverityCritical      = "critical"
	FindingSeverityHigh          = "high"
	FindingSeverityMedium        = "medium"
	FindingSeverityLow           = "low"
	FindingSeverityInformational = "informational"
)

var ValidFindingSeverities = []string{
	FindingSeverityCritical, FindingSeverityHigh, FindingSeverityMedium,
	FindingSeverityLow, FindingSeverityInformational,
}

func IsValidFindingSeverity(s string) bool {
	for _, v := range ValidFindingSeverities {
		if v == s {
			return true
		}
	}
	return false
}

// --- Finding status ---
const (
	FindingStatusIdentified           = "identified"
	FindingStatusAcknowledged         = "acknowledged"
	FindingStatusRemediationPlanned   = "remediation_planned"
	FindingStatusRemediationInProgress = "remediation_in_progress"
	FindingStatusRemediationComplete  = "remediation_complete"
	FindingStatusVerified             = "verified"
	FindingStatusRiskAccepted         = "risk_accepted"
	FindingStatusClosed               = "closed"
)

var ValidFindingStatuses = []string{
	FindingStatusIdentified, FindingStatusAcknowledged,
	FindingStatusRemediationPlanned, FindingStatusRemediationInProgress,
	FindingStatusRemediationComplete, FindingStatusVerified,
	FindingStatusRiskAccepted, FindingStatusClosed,
}

func IsValidFindingStatus(s string) bool {
	for _, v := range ValidFindingStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// FindingStatusTransitions defines valid finding transitions.
var FindingStatusTransitions = map[string][]string{
	FindingStatusIdentified:            {FindingStatusAcknowledged, FindingStatusRiskAccepted},
	FindingStatusAcknowledged:          {FindingStatusRemediationPlanned, FindingStatusRiskAccepted},
	FindingStatusRemediationPlanned:    {FindingStatusRemediationInProgress, FindingStatusRiskAccepted},
	FindingStatusRemediationInProgress: {FindingStatusRemediationComplete, FindingStatusRiskAccepted},
	FindingStatusRemediationComplete:   {FindingStatusVerified, FindingStatusRemediationInProgress},
	FindingStatusVerified:              {FindingStatusClosed},
	FindingStatusRiskAccepted:          {FindingStatusClosed},
}

func IsValidFindingStatusTransition(from, to string) bool {
	allowed, ok := FindingStatusTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

// --- Finding category ---
const (
	FindingCategoryControlDeficiency = "control_deficiency"
	FindingCategoryControlGap        = "control_gap"
	FindingCategoryDocumentationGap  = "documentation_gap"
	FindingCategoryProcessGap        = "process_gap"
	FindingCategoryConfigIssue       = "configuration_issue"
	FindingCategoryAccessControl     = "access_control"
	FindingCategoryMonitoringGap     = "monitoring_gap"
	FindingCategoryPolicyViolation   = "policy_violation"
	FindingCategoryVendorRisk        = "vendor_risk"
	FindingCategoryDataHandling      = "data_handling"
	FindingCategoryOther             = "other"
)

var ValidFindingCategories = []string{
	FindingCategoryControlDeficiency, FindingCategoryControlGap,
	FindingCategoryDocumentationGap, FindingCategoryProcessGap,
	FindingCategoryConfigIssue, FindingCategoryAccessControl,
	FindingCategoryMonitoringGap, FindingCategoryPolicyViolation,
	FindingCategoryVendorRisk, FindingCategoryDataHandling, FindingCategoryOther,
}

func IsValidFindingCategory(c string) bool {
	for _, v := range ValidFindingCategories {
		if v == c {
			return true
		}
	}
	return false
}

// --- Evidence link status ---
const (
	EvidenceLinkStatusPendingReview      = "pending_review"
	EvidenceLinkStatusAccepted           = "accepted"
	EvidenceLinkStatusRejected           = "rejected"
	EvidenceLinkStatusNeedsClarification = "needs_clarification"
)

var ValidEvidenceLinkStatuses = []string{
	EvidenceLinkStatusPendingReview, EvidenceLinkStatusAccepted,
	EvidenceLinkStatusRejected, EvidenceLinkStatusNeedsClarification,
}

func IsValidEvidenceLinkStatus(s string) bool {
	for _, v := range ValidEvidenceLinkStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// --- Comment target types ---
const (
	CommentTargetAudit   = "audit"
	CommentTargetRequest = "request"
	CommentTargetFinding = "finding"
)

var ValidCommentTargetTypes = []string{
	CommentTargetAudit, CommentTargetRequest, CommentTargetFinding,
}

func IsValidCommentTargetType(t string) bool {
	for _, v := range ValidCommentTargetTypes {
		if v == t {
			return true
		}
	}
	return false
}

// --- Role sets for audit hub ---

// AuditCreateRoles can create/update audits.
var AuditCreateRoles = []string{RoleCISO, RoleComplianceManager}

// AuditViewRoles can view audits (auditor gets filtered view).
var AuditHubViewRoles = []string{
	RoleCISO, RoleComplianceManager, RoleSecurityEngineer,
	RoleITAdmin, RoleAuditor,
}

// AuditDashboardRoles can view the audit hub dashboard.
var AuditDashboardRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// AuditRequestCreateRoles can create requests.
var AuditRequestCreateRoles = []string{RoleCISO, RoleComplianceManager, RoleAuditor}

// AuditRequestAssignRoles can assign requests.
var AuditRequestAssignRoles = []string{RoleCISO, RoleComplianceManager}

// AuditEvidenceSubmitRoles can submit evidence.
var AuditEvidenceSubmitRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleITAdmin}

// AuditEvidenceReviewRoles can review evidence.
var AuditEvidenceReviewRoles = []string{RoleAuditor}

// AuditFindingCreateRoles can create findings.
var AuditFindingCreateRoles = []string{RoleAuditor}

// AuditFindingStatusInternalRoles internal roles that can transition finding status.
var AuditFindingStatusInternalRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// AuditCommentCreateRoles can create comments.
var AuditCommentCreateRoles = []string{
	RoleCISO, RoleComplianceManager, RoleSecurityEngineer,
	RoleITAdmin, RoleAuditor,
}

// AuditManagementResponseRoles can submit management response.
var AuditManagementResponseRoles = []string{RoleCISO, RoleComplianceManager}

// --- Request types ---

// CreateAuditRequest is the request body for creating an audit.
type CreateAuditRequest struct {
	Title           string            `json:"title" binding:"required"`
	Description     *string           `json:"description"`
	AuditType       string            `json:"audit_type" binding:"required"`
	OrgFrameworkID  *string           `json:"org_framework_id"`
	PeriodStart     *string           `json:"period_start"`
	PeriodEnd       *string           `json:"period_end"`
	PlannedStart    *string           `json:"planned_start"`
	PlannedEnd      *string           `json:"planned_end"`
	AuditFirm       *string           `json:"audit_firm"`
	LeadAuditorID   *string           `json:"lead_auditor_id"`
	AuditorIDs      []string          `json:"auditor_ids"`
	InternalLeadID  *string           `json:"internal_lead_id"`
	Milestones      []MilestoneInput  `json:"milestones"`
	ReportType      *string           `json:"report_type"`
	Tags            []string          `json:"tags"`
}

// MilestoneInput represents a milestone in create/update requests.
type MilestoneInput struct {
	Name       string  `json:"name"`
	TargetDate string  `json:"target_date"`
}

// UpdateAuditRequest is the request body for updating an audit.
type UpdateAuditRequest struct {
	Title          *string           `json:"title"`
	Description    *string           `json:"description"`
	OrgFrameworkID *string           `json:"org_framework_id"`
	PeriodStart    *string           `json:"period_start"`
	PeriodEnd      *string           `json:"period_end"`
	PlannedStart   *string           `json:"planned_start"`
	PlannedEnd     *string           `json:"planned_end"`
	AuditFirm      *string           `json:"audit_firm"`
	LeadAuditorID  *string           `json:"lead_auditor_id"`
	InternalLeadID *string           `json:"internal_lead_id"`
	Milestones     []MilestoneInput  `json:"milestones"`
	ReportType     *string           `json:"report_type"`
	ReportURL      *string           `json:"report_url"`
	Tags           []string          `json:"tags"`
}

// ChangeAuditStatusRequest is for status transitions.
type ChangeAuditStatusRequest struct {
	Status string  `json:"status" binding:"required"`
	Notes  *string `json:"notes"`
}

// AddAuditorRequest adds an auditor to an audit.
type AddAuditorRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// CreateAuditRequestReq is the request body for creating an evidence request.
type CreateAuditRequestReq struct {
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	Priority        *string  `json:"priority"`
	ControlID       *string  `json:"control_id"`
	RequirementID   *string  `json:"requirement_id"`
	AssignedTo      *string  `json:"assigned_to"`
	DueDate         *string  `json:"due_date"`
	ReferenceNumber *string  `json:"reference_number"`
	Tags            []string `json:"tags"`
}

// UpdateAuditRequestReq is for updating request metadata.
type UpdateAuditRequestReq struct {
	Title           *string  `json:"title"`
	Description     *string  `json:"description"`
	Priority        *string  `json:"priority"`
	DueDate         *string  `json:"due_date"`
	ReferenceNumber *string  `json:"reference_number"`
	Tags            []string `json:"tags"`
}

// AssignAuditRequestReq assigns a request.
type AssignAuditRequestReq struct {
	AssignedTo string `json:"assigned_to" binding:"required"`
}

// SubmitAuditRequestReq marks request as submitted.
type SubmitAuditRequestReq struct {
	Notes *string `json:"notes"`
}

// ReviewAuditRequestReq for auditor review.
type ReviewAuditRequestReq struct {
	Decision string  `json:"decision" binding:"required"`
	Notes    *string `json:"notes"`
}

// CloseAuditRequestReq closes a request.
type CloseAuditRequestReq struct {
	Reason *string `json:"reason"`
}

// BulkCreateRequestsReq creates multiple requests.
type BulkCreateRequestsReq struct {
	Requests []CreateAuditRequestReq `json:"requests" binding:"required"`
}

// SubmitEvidenceReq links evidence to a request.
type SubmitEvidenceReq struct {
	ArtifactID string  `json:"artifact_id" binding:"required"`
	Notes      *string `json:"notes"`
}

// ReviewEvidenceReq for auditor evidence review.
type ReviewEvidenceReq struct {
	Status string  `json:"status" binding:"required"`
	Notes  *string `json:"notes"`
}

// CreateFindingReq creates a finding.
type CreateFindingReq struct {
	Title              string   `json:"title" binding:"required"`
	Description        string   `json:"description" binding:"required"`
	Severity           string   `json:"severity" binding:"required"`
	Category           string   `json:"category" binding:"required"`
	ControlID          *string  `json:"control_id"`
	RequirementID      *string  `json:"requirement_id"`
	RemediationOwnerID *string  `json:"remediation_owner_id"`
	RemediationDueDate *string  `json:"remediation_due_date"`
	ReferenceNumber    *string  `json:"reference_number"`
	Recommendation     *string  `json:"recommendation"`
	Tags               []string `json:"tags"`
}

// UpdateFindingReq updates finding metadata.
type UpdateFindingReq struct {
	Title           *string  `json:"title"`
	Description     *string  `json:"description"`
	Severity        *string  `json:"severity"`
	Category        *string  `json:"category"`
	Recommendation  *string  `json:"recommendation"`
	ReferenceNumber *string  `json:"reference_number"`
	Tags            []string `json:"tags"`
}

// ChangeFindingStatusReq transitions finding status.
type ChangeFindingStatusReq struct {
	Status             string  `json:"status" binding:"required"`
	RemediationPlan    *string `json:"remediation_plan"`
	RemediationDueDate *string `json:"remediation_due_date"`
	RemediationOwnerID *string `json:"remediation_owner_id"`
	ManagementResponse *string `json:"management_response"`
	VerificationNotes  *string `json:"verification_notes"`
	RiskAcceptReason   *string `json:"risk_acceptance_reason"`
	Notes              *string `json:"notes"`
}

// ManagementResponseReq submits management response.
type ManagementResponseReq struct {
	ManagementResponse string `json:"management_response" binding:"required"`
}

// CreateCommentReq creates a comment.
type CreateCommentReq struct {
	TargetType      string  `json:"target_type" binding:"required"`
	TargetID        string  `json:"target_id" binding:"required"`
	Body            string  `json:"body" binding:"required"`
	ParentCommentID *string `json:"parent_comment_id"`
	IsInternal      *bool   `json:"is_internal"`
}

// UpdateCommentReq updates a comment.
type UpdateCommentReq struct {
	Body string `json:"body" binding:"required"`
}

// CreateFromTemplateReq creates requests from templates.
type CreateFromTemplateReq struct {
	TemplateIDs    []string `json:"template_ids" binding:"required"`
	DefaultDueDate *string  `json:"default_due_date"`
	AutoNumber     *bool    `json:"auto_number"`
	NumberPrefix   *string  `json:"number_prefix"`
}
