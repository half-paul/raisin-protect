package models

import "time"

// Alert represents a generated alert from a test failure.
type Alert struct {
	ID                string     `json:"id"`
	OrgID             string     `json:"org_id"`
	AlertNumber       int        `json:"alert_number"`
	Title             string     `json:"title"`
	Description       *string    `json:"description"`
	Severity          string     `json:"severity"`
	Status            string     `json:"status"`
	TestID            *string    `json:"test_id"`
	TestResultID      *string    `json:"test_result_id"`
	ControlID         string     `json:"control_id"`
	AlertRuleID       *string    `json:"alert_rule_id"`
	AssignedTo        *string    `json:"assigned_to"`
	AssignedAt        *time.Time `json:"assigned_at"`
	AssignedBy        *string    `json:"assigned_by"`
	SLADeadline       *time.Time `json:"sla_deadline"`
	SLABreached       bool       `json:"sla_breached"`
	ResolvedBy        *string    `json:"resolved_by"`
	ResolvedAt        *time.Time `json:"resolved_at"`
	ResolutionNotes   *string    `json:"resolution_notes"`
	SuppressedUntil   *time.Time `json:"suppressed_until"`
	SuppressionReason *string    `json:"suppression_reason"`
	DeliveryChannels  []string   `json:"delivery_channels"`
	DeliveredAt       string     `json:"delivered_at"` // raw JSON
	Tags              []string   `json:"tags"`
	Metadata          string     `json:"metadata"` // raw JSON
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Valid alert severities.
var ValidAlertSeverities = []string{"critical", "high", "medium", "low"}

// Valid alert statuses.
var ValidAlertStatuses = []string{"open", "acknowledged", "in_progress", "resolved", "suppressed", "closed"}

// Valid alert delivery channels.
var ValidAlertDeliveryChannels = []string{"slack", "email", "webhook", "in_app"}

// Valid alert status transitions.
var ValidAlertStatusTransitions = map[string][]string{
	"open":         {"acknowledged", "in_progress", "suppressed", "closed"},
	"acknowledged": {"in_progress", "suppressed", "closed"},
	"in_progress":  {"resolved", "suppressed", "closed"},
	"resolved":     {"closed", "open"},
	"suppressed":   {"open", "closed"},
	"closed":       {"open"},
}

// AlertStatusRoles can change alert status.
var AlertStatusRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleITAdmin, RoleDevOpsEngineer}

// AlertAssignRoles can assign alerts.
var AlertAssignRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// AlertResolveRoles can resolve alerts.
var AlertResolveRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleITAdmin, RoleDevOpsEngineer}

// AlertSuppressRoles can suppress/close alerts.
var AlertSuppressRoles = []string{RoleCISO, RoleComplianceManager}

// AlertDeliveryRoles can re-deliver/test-deliver alerts.
var AlertDeliveryRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// IsValidAlertSeverity checks if the given severity is valid.
func IsValidAlertSeverity(s string) bool {
	for _, v := range ValidAlertSeverities {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidAlertStatus checks if the given status is valid.
func IsValidAlertStatus(s string) bool {
	for _, v := range ValidAlertStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidAlertDeliveryChannel checks if the given channel is valid.
func IsValidAlertDeliveryChannel(ch string) bool {
	for _, v := range ValidAlertDeliveryChannels {
		if v == ch {
			return true
		}
	}
	return false
}

// IsValidAlertStatusTransition checks if an alert status transition is allowed.
func IsValidAlertStatusTransition(from, to string) bool {
	allowed, ok := ValidAlertStatusTransitions[from]
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

// ChangeAlertStatusRequest is the request body for changing alert status.
type ChangeAlertStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// AssignAlertRequest is the request body for assigning an alert.
type AssignAlertRequest struct {
	AssignedTo string `json:"assigned_to" binding:"required"`
}

// ResolveAlertRequest is the request body for resolving an alert.
type ResolveAlertRequest struct {
	ResolutionNotes string `json:"resolution_notes" binding:"required"`
}

// SuppressAlertRequest is the request body for suppressing an alert.
type SuppressAlertRequest struct {
	SuppressedUntil   string `json:"suppressed_until" binding:"required"`
	SuppressionReason string `json:"suppression_reason" binding:"required"`
}

// CloseAlertRequest is the request body for closing an alert.
type CloseAlertRequest struct {
	ResolutionNotes *string `json:"resolution_notes"`
}
