package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// ListAlertRules lists the org's alert rules.
func ListAlertRules(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	where := []string{"r.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("enabled"); v != "" {
		where = append(where, fmt.Sprintf("r.enabled = $%d", argN))
		args = append(args, v == "true")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alert_rules r WHERE %s", whereClause)
	if err := database.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count alert rules")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list alert rules"))
		return
	}

	sortField := "r.priority"
	switch c.DefaultQuery("sort", "priority") {
	case "name":
		sortField = "r.name"
	case "alert_severity":
		sortField = "r.alert_severity"
	case "created_at":
		sortField = "r.created_at"
	}
	order := "ASC"
	if strings.ToLower(c.DefaultQuery("order", "asc")) == "desc" {
		order = "DESC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.name, r.description, r.enabled,
			r.match_test_types, r.match_severities, r.match_result_statuses,
			r.match_control_ids, r.match_tags,
			r.consecutive_failures, r.cooldown_minutes,
			r.alert_severity, r.alert_title_template, r.auto_assign_to,
			r.sla_hours, r.delivery_channels,
			r.slack_webhook_url, r.email_recipients, r.webhook_url,
			r.priority, r.created_at, r.updated_at,
			(SELECT COUNT(*) FROM alerts WHERE alert_rule_id = r.id) AS alerts_generated
		FROM alert_rules r
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortField, order, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query alert rules")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list alert rules"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, name, alertSeverity                           string
			description, alertTitleTemplate, autoAssignTo      *string
			slackWebhookURL, webhookURL                       *string
			enabled                                           bool
			matchTestTypes, matchSeverities, matchResultStatuses pq.StringArray
			matchControlIDs, matchTags                        pq.StringArray
			deliveryChannels, emailRecipients                 pq.StringArray
			consecutiveFailures, cooldownMinutes, priority    int
			slaHours                                          *int
			alertsGenerated                                   int
			createdAt, updatedAt                              time.Time
		)
		if err := rows.Scan(
			&id, &name, &description, &enabled,
			&matchTestTypes, &matchSeverities, &matchResultStatuses,
			&matchControlIDs, &matchTags,
			&consecutiveFailures, &cooldownMinutes,
			&alertSeverity, &alertTitleTemplate, &autoAssignTo,
			&slaHours, &deliveryChannels,
			&slackWebhookURL, &emailRecipients, &webhookURL,
			&priority, &createdAt, &updatedAt,
			&alertsGenerated,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan alert rule row")
			continue
		}

		results = append(results, gin.H{
			"id":                    id,
			"name":                  name,
			"description":           description,
			"enabled":               enabled,
			"match_test_types":      matchTestTypes,
			"match_severities":      matchSeverities,
			"match_result_statuses": matchResultStatuses,
			"match_control_ids":     matchControlIDs,
			"match_tags":            matchTags,
			"consecutive_failures":  consecutiveFailures,
			"cooldown_minutes":      cooldownMinutes,
			"alert_severity":        alertSeverity,
			"alert_title_template":  alertTitleTemplate,
			"auto_assign_to":        autoAssignTo,
			"sla_hours":             slaHours,
			"delivery_channels":     deliveryChannels,
			"slack_webhook_url":     slackWebhookURL,
			"email_recipients":      emailRecipients,
			"webhook_url":           webhookURL,
			"priority":              priority,
			"alerts_generated":      alertsGenerated,
			"created_at":            createdAt,
			"updated_at":            updatedAt,
		})
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// CreateAlertRule creates a new alert rule.
func CreateAlertRule(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "name, alert_severity, and delivery_channels are required"))
		return
	}

	if len(req.Name) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Name must be 255 characters or less"))
		return
	}
	if !models.IsValidAlertSeverity(req.AlertSeverity) {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid alert_severity"))
		return
	}
	if len(req.DeliveryChannels) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "At least one delivery channel is required"))
		return
	}
	for _, ch := range req.DeliveryChannels {
		if !models.IsValidAlertDeliveryChannel(ch) {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", fmt.Sprintf("Invalid delivery channel: %s", ch)))
			return
		}
	}

	// Check name uniqueness
	var exists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM alert_rules WHERE org_id = $1 AND name = $2)", orgID, req.Name).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check rule name")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create alert rule"))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Alert rule name already exists in this organization"))
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	consecutiveFailures := 1
	if req.ConsecutiveFailures != nil {
		consecutiveFailures = *req.ConsecutiveFailures
	}
	cooldownMinutes := 0
	if req.CooldownMinutes != nil {
		cooldownMinutes = *req.CooldownMinutes
	}
	priority := 100
	if req.Priority != nil {
		priority = *req.Priority
	}

	webhookHeadersJSON := "{}"
	if req.WebhookHeaders != nil {
		b, _ := json.Marshal(req.WebhookHeaders)
		webhookHeadersJSON = string(b)
	}

	// Default match_result_statuses
	matchResultStatuses := req.MatchResultStatuses
	if len(matchResultStatuses) == 0 {
		matchResultStatuses = []string{"fail"}
	}

	id := uuid.New().String()
	now := time.Now()

	_, err = database.Exec(`
		INSERT INTO alert_rules (id, org_id, name, description, enabled,
			match_test_types, match_severities, match_result_statuses, match_control_ids, match_tags,
			consecutive_failures, cooldown_minutes,
			alert_severity, alert_title_template, auto_assign_to, sla_hours,
			delivery_channels, slack_webhook_url, email_recipients, webhook_url, webhook_headers,
			priority, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $24)
	`, id, orgID, req.Name, req.Description, enabled,
		pq.Array(req.MatchTestTypes), pq.Array(req.MatchSeverities), pq.Array(matchResultStatuses),
		pq.Array(req.MatchControlIDs), pq.Array(req.MatchTags),
		consecutiveFailures, cooldownMinutes,
		req.AlertSeverity, req.AlertTitleTemplate, req.AutoAssignTo, req.SLAHours,
		pq.Array(req.DeliveryChannels), req.SlackWebhookURL, pq.Array(req.EmailRecipients),
		req.WebhookURL, webhookHeadersJSON,
		priority, userID, now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert alert rule")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create alert rule"))
		return
	}

	middleware.LogAudit(c, "alert_rule.created", "alert_rule", &id, map[string]interface{}{
		"name":     req.Name,
		"severity": req.AlertSeverity,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":             id,
		"name":           req.Name,
		"enabled":        enabled,
		"alert_severity": req.AlertSeverity,
		"priority":       priority,
		"created_at":     now,
	}))
}

// GetAlertRule gets a single alert rule with full details.
func GetAlertRule(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ruleID := c.Param("id")

	var (
		id, name, alertSeverity                           string
		description, alertTitleTemplate, autoAssignTo      *string
		slackWebhookURL, webhookURL                       *string
		webhookHeadersJSON                                string
		enabled                                           bool
		matchTestTypes, matchSeverities, matchResultStatuses pq.StringArray
		matchControlIDs, matchTags                        pq.StringArray
		deliveryChannels, emailRecipients                 pq.StringArray
		consecutiveFailures, cooldownMinutes, priority    int
		slaHours                                          *int
		createdBy                                         *string
		createdAt, updatedAt                              time.Time
	)

	err := database.QueryRow(`
		SELECT r.id, r.name, r.description, r.enabled,
			r.match_test_types, r.match_severities, r.match_result_statuses,
			r.match_control_ids, r.match_tags,
			r.consecutive_failures, r.cooldown_minutes,
			r.alert_severity, r.alert_title_template, r.auto_assign_to,
			r.sla_hours, r.delivery_channels,
			r.slack_webhook_url, r.email_recipients, r.webhook_url, r.webhook_headers,
			r.priority, r.created_by, r.created_at, r.updated_at
		FROM alert_rules r
		WHERE r.id = $1 AND r.org_id = $2
	`, ruleID, orgID).Scan(
		&id, &name, &description, &enabled,
		&matchTestTypes, &matchSeverities, &matchResultStatuses,
		&matchControlIDs, &matchTags,
		&consecutiveFailures, &cooldownMinutes,
		&alertSeverity, &alertTitleTemplate, &autoAssignTo,
		&slaHours, &deliveryChannels,
		&slackWebhookURL, &emailRecipients, &webhookURL, &webhookHeadersJSON,
		&priority, &createdBy, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert rule not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert rule")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get alert rule"))
		return
	}

	var webhookHeaders interface{}
	json.Unmarshal([]byte(webhookHeadersJSON), &webhookHeaders)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                    id,
		"name":                  name,
		"description":           description,
		"enabled":               enabled,
		"match_test_types":      matchTestTypes,
		"match_severities":      matchSeverities,
		"match_result_statuses": matchResultStatuses,
		"match_control_ids":     matchControlIDs,
		"match_tags":            matchTags,
		"consecutive_failures":  consecutiveFailures,
		"cooldown_minutes":      cooldownMinutes,
		"alert_severity":        alertSeverity,
		"alert_title_template":  alertTitleTemplate,
		"auto_assign_to":        autoAssignTo,
		"sla_hours":             slaHours,
		"delivery_channels":     deliveryChannels,
		"slack_webhook_url":     slackWebhookURL,
		"email_recipients":      emailRecipients,
		"webhook_url":           webhookURL,
		"webhook_headers":       webhookHeaders,
		"priority":              priority,
		"created_by":            createdBy,
		"created_at":            createdAt,
		"updated_at":            updatedAt,
	}))
}

// UpdateAlertRule updates an alert rule.
func UpdateAlertRule(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ruleID := c.Param("id")

	var req models.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid request body"))
		return
	}

	// Check exists
	var currentName string
	err := database.QueryRow("SELECT name FROM alert_rules WHERE id = $1 AND org_id = $2", ruleID, orgID).Scan(&currentName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert rule not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check rule")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update alert rule"))
		return
	}

	sets := []string{}
	args := []interface{}{}
	argN := 1

	if req.Name != nil {
		if len(*req.Name) > 255 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Name must be 255 characters or less"))
			return
		}
		// Check uniqueness if name changed
		if *req.Name != currentName {
			var exists bool
			database.QueryRow("SELECT EXISTS(SELECT 1 FROM alert_rules WHERE org_id = $1 AND name = $2 AND id != $3)", orgID, *req.Name, ruleID).Scan(&exists)
			if exists {
				c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Alert rule name already exists"))
				return
			}
		}
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, *req.Name)
		argN++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argN))
		args = append(args, *req.Description)
		argN++
	}
	if req.Enabled != nil {
		sets = append(sets, fmt.Sprintf("enabled = $%d", argN))
		args = append(args, *req.Enabled)
		argN++
	}
	if req.MatchTestTypes != nil {
		sets = append(sets, fmt.Sprintf("match_test_types = $%d", argN))
		args = append(args, pq.Array(req.MatchTestTypes))
		argN++
	}
	if req.MatchSeverities != nil {
		sets = append(sets, fmt.Sprintf("match_severities = $%d", argN))
		args = append(args, pq.Array(req.MatchSeverities))
		argN++
	}
	if req.MatchResultStatuses != nil {
		sets = append(sets, fmt.Sprintf("match_result_statuses = $%d", argN))
		args = append(args, pq.Array(req.MatchResultStatuses))
		argN++
	}
	if req.AlertSeverity != nil {
		if !models.IsValidAlertSeverity(*req.AlertSeverity) {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid alert_severity"))
			return
		}
		sets = append(sets, fmt.Sprintf("alert_severity = $%d", argN))
		args = append(args, *req.AlertSeverity)
		argN++
	}
	if req.ConsecutiveFailures != nil {
		sets = append(sets, fmt.Sprintf("consecutive_failures = $%d", argN))
		args = append(args, *req.ConsecutiveFailures)
		argN++
	}
	if req.CooldownMinutes != nil {
		sets = append(sets, fmt.Sprintf("cooldown_minutes = $%d", argN))
		args = append(args, *req.CooldownMinutes)
		argN++
	}
	if req.SLAHours != nil {
		sets = append(sets, fmt.Sprintf("sla_hours = $%d", argN))
		args = append(args, *req.SLAHours)
		argN++
	}
	if req.DeliveryChannels != nil {
		sets = append(sets, fmt.Sprintf("delivery_channels = $%d", argN))
		args = append(args, pq.Array(req.DeliveryChannels))
		argN++
	}
	if req.SlackWebhookURL != nil {
		sets = append(sets, fmt.Sprintf("slack_webhook_url = $%d", argN))
		args = append(args, *req.SlackWebhookURL)
		argN++
	}
	if req.EmailRecipients != nil {
		sets = append(sets, fmt.Sprintf("email_recipients = $%d", argN))
		args = append(args, pq.Array(req.EmailRecipients))
		argN++
	}
	if req.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", argN))
		args = append(args, *req.Priority)
		argN++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "No fields to update"))
		return
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, ruleID, orgID)
	query := fmt.Sprintf("UPDATE alert_rules SET %s WHERE id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1)

	_, err = database.Exec(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update alert rule")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update alert rule"))
		return
	}

	middleware.LogAudit(c, "alert_rule.updated", "alert_rule", &ruleID, map[string]interface{}{
		"name": currentName,
	})

	// Return updated rule
	GetAlertRule(c)
}

// DeleteAlertRule deletes an alert rule.
func DeleteAlertRule(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ruleID := c.Param("id")

	var name string
	err := database.QueryRow("SELECT name FROM alert_rules WHERE id = $1 AND org_id = $2", ruleID, orgID).Scan(&name)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert rule not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check rule")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete alert rule"))
		return
	}

	_, err = database.Exec("DELETE FROM alert_rules WHERE id = $1 AND org_id = $2", ruleID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete alert rule")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete alert rule"))
		return
	}

	middleware.LogAudit(c, "alert_rule.deleted", "alert_rule", &ruleID, map[string]interface{}{
		"name":    name,
		"rule_id": ruleID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":      ruleID,
		"message": "Alert rule deleted. Existing alerts preserved.",
	}))
}
