package models

import "time"

// AlertRule represents an org-configurable alert generation rule.
type AlertRule struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"org_id"`
	Name                string     `json:"name"`
	Description         *string    `json:"description"`
	Enabled             bool       `json:"enabled"`
	MatchTestTypes      []string   `json:"match_test_types"`
	MatchSeverities     []string   `json:"match_severities"`
	MatchResultStatuses []string   `json:"match_result_statuses"`
	MatchControlIDs     []string   `json:"match_control_ids"`
	MatchTags           []string   `json:"match_tags"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	CooldownMinutes     int        `json:"cooldown_minutes"`
	AlertSeverity       string     `json:"alert_severity"`
	AlertTitleTemplate  *string    `json:"alert_title_template"`
	AutoAssignTo        *string    `json:"auto_assign_to"`
	SLAHours            *int       `json:"sla_hours"`
	DeliveryChannels    []string   `json:"delivery_channels"`
	SlackWebhookURL     *string    `json:"slack_webhook_url,omitempty"`
	EmailRecipients     []string   `json:"email_recipients,omitempty"`
	WebhookURL          *string    `json:"webhook_url,omitempty"`
	WebhookHeaders      string     `json:"webhook_headers,omitempty"` // raw JSON
	Priority            int        `json:"priority"`
	AlertsGenerated     int        `json:"alerts_generated,omitempty"`
	CreatedBy           *string    `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// AlertRuleViewRoles can view alert rules.
var AlertRuleViewRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// AlertRuleCreateRoles can create/update/delete alert rules.
var AlertRuleCreateRoles = []string{RoleCISO, RoleComplianceManager}

// CreateAlertRuleRequest is the request body for creating an alert rule.
type CreateAlertRuleRequest struct {
	Name                string                 `json:"name" binding:"required"`
	Description         *string                `json:"description"`
	Enabled             *bool                  `json:"enabled"`
	MatchTestTypes      []string               `json:"match_test_types"`
	MatchSeverities     []string               `json:"match_severities"`
	MatchResultStatuses []string               `json:"match_result_statuses"`
	MatchControlIDs     []string               `json:"match_control_ids"`
	MatchTags           []string               `json:"match_tags"`
	ConsecutiveFailures *int                   `json:"consecutive_failures"`
	CooldownMinutes     *int                   `json:"cooldown_minutes"`
	AlertSeverity       string                 `json:"alert_severity" binding:"required"`
	AlertTitleTemplate  *string                `json:"alert_title_template"`
	AutoAssignTo        *string                `json:"auto_assign_to"`
	SLAHours            *int                   `json:"sla_hours"`
	DeliveryChannels    []string               `json:"delivery_channels" binding:"required"`
	SlackWebhookURL     *string                `json:"slack_webhook_url"`
	EmailRecipients     []string               `json:"email_recipients"`
	WebhookURL          *string                `json:"webhook_url"`
	WebhookHeaders      map[string]interface{} `json:"webhook_headers"`
	Priority            *int                   `json:"priority"`
}

// UpdateAlertRuleRequest is the request body for updating an alert rule.
type UpdateAlertRuleRequest struct {
	Name                *string                `json:"name"`
	Description         *string                `json:"description"`
	Enabled             *bool                  `json:"enabled"`
	MatchTestTypes      []string               `json:"match_test_types"`
	MatchSeverities     []string               `json:"match_severities"`
	MatchResultStatuses []string               `json:"match_result_statuses"`
	MatchControlIDs     []string               `json:"match_control_ids"`
	MatchTags           []string               `json:"match_tags"`
	ConsecutiveFailures *int                   `json:"consecutive_failures"`
	CooldownMinutes     *int                   `json:"cooldown_minutes"`
	AlertSeverity       *string                `json:"alert_severity"`
	AlertTitleTemplate  *string                `json:"alert_title_template"`
	AutoAssignTo        *string                `json:"auto_assign_to"`
	SLAHours            *int                   `json:"sla_hours"`
	DeliveryChannels    []string               `json:"delivery_channels"`
	SlackWebhookURL     *string                `json:"slack_webhook_url"`
	EmailRecipients     []string               `json:"email_recipients"`
	WebhookURL          *string                `json:"webhook_url"`
	WebhookHeaders      map[string]interface{} `json:"webhook_headers"`
	Priority            *int                   `json:"priority"`
}
