package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListAlerts lists alerts with filtering, search, and pagination.
func ListAlerts(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"a.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("status"); v != "" {
		statuses := strings.Split(v, ",")
		placeholders := []string{}
		for _, s := range statuses {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
			args = append(args, strings.TrimSpace(s))
			argN++
		}
		where = append(where, fmt.Sprintf("a.status IN (%s)", strings.Join(placeholders, ",")))
	}
	if v := c.Query("severity"); v != "" {
		severities := strings.Split(v, ",")
		placeholders := []string{}
		for _, s := range severities {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
			args = append(args, strings.TrimSpace(s))
			argN++
		}
		where = append(where, fmt.Sprintf("a.severity IN (%s)", strings.Join(placeholders, ",")))
	}
	if v := c.Query("control_id"); v != "" {
		where = append(where, fmt.Sprintf("a.control_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("test_id"); v != "" {
		where = append(where, fmt.Sprintf("a.test_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("assigned_to"); v != "" {
		if v == "unassigned" {
			where = append(where, "a.assigned_to IS NULL")
		} else {
			where = append(where, fmt.Sprintf("a.assigned_to = $%d", argN))
			args = append(args, v)
			argN++
		}
	}
	if v := c.Query("sla_breached"); v == "true" {
		where = append(where, "a.sla_breached = TRUE")
	} else if v == "false" {
		where = append(where, "a.sla_breached = FALSE")
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(a.title ILIKE $%d OR a.description ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}
	if v := c.Query("date_from"); v != "" {
		where = append(where, fmt.Sprintf("a.created_at >= $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_to"); v != "" {
		where = append(where, fmt.Sprintf("a.created_at <= $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts a WHERE %s", whereClause)
	if err := database.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count alerts")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list alerts"))
		return
	}

	sortField := "a.created_at"
	switch c.DefaultQuery("sort", "created_at") {
	case "alert_number":
		sortField = "a.alert_number"
	case "severity":
		sortField = "a.severity"
	case "status":
		sortField = "a.status"
	case "sla_deadline":
		sortField = "a.sla_deadline"
	case "updated_at":
		sortField = "a.updated_at"
	}
	order := "DESC"
	if strings.ToLower(c.DefaultQuery("order", "desc")) == "asc" {
		order = "ASC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT a.id, a.alert_number, a.title, a.description, a.severity, a.status,
			a.control_id, c.identifier, c.title,
			a.test_id, t.identifier, t.title,
			a.assigned_to, a.sla_deadline, a.sla_breached,
			a.created_at, a.updated_at
		FROM alerts a
		LEFT JOIN controls c ON c.id = a.control_id
		LEFT JOIN tests t ON t.id = a.test_id
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortField, order, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query alerts")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list alerts"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			aID, aTitle, aSeverity, aStatus, controlID, controlIdentifier, controlTitle string
			aDescription                                                                 *string
			alertNumber                                                                  int
			testID, testIdentifier, testTitle                                            *string
			assignedTo                                                                   *string
			slaDeadline                                                                  *time.Time
			slaBreached                                                                  bool
			createdAt, updatedAt                                                         time.Time
		)
		if err := rows.Scan(
			&aID, &alertNumber, &aTitle, &aDescription, &aSeverity, &aStatus,
			&controlID, &controlIdentifier, &controlTitle,
			&testID, &testIdentifier, &testTitle,
			&assignedTo, &slaDeadline, &slaBreached,
			&createdAt, &updatedAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan alert row")
			continue
		}

		item := gin.H{
			"id":           aID,
			"alert_number": alertNumber,
			"title":        aTitle,
			"description":  aDescription,
			"severity":     aSeverity,
			"status":       aStatus,
			"control": gin.H{
				"id":         controlID,
				"identifier": controlIdentifier,
				"title":      controlTitle,
			},
			"assigned_to":  assignedTo,
			"sla_deadline":  slaDeadline,
			"sla_breached":  slaBreached,
			"created_at":    createdAt,
			"updated_at":    updatedAt,
		}

		if testID != nil {
			item["test"] = gin.H{
				"id":         *testID,
				"identifier": *testIdentifier,
				"title":      *testTitle,
			}
		}

		// Compute hours_remaining
		if slaDeadline != nil {
			remaining := time.Until(*slaDeadline).Hours()
			item["hours_remaining"] = math.Round(remaining*10) / 10
		}

		results = append(results, item)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// GetAlert gets a single alert with full details.
func GetAlert(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	alertID := c.Param("id")

	var (
		aID, aTitle, aSeverity, aStatus, controlID, controlIdentifier, controlTitle, controlCategory string
		aDescription                                                                                 *string
		alertNumber                                                                                  int
		testID, testIdentifier, testTitle, testType                                                  *string
		testResultID                                                                                 *string
		alertRuleID, alertRuleName                                                                   *string
		assignedTo, assignedBy                                                                       *string
		assignedAt                                                                                   *time.Time
		slaDeadline                                                                                  *time.Time
		slaBreached                                                                                  bool
		resolvedBy                                                                                   *string
		resolvedAt                                                                                   *time.Time
		resolutionNotes                                                                              *string
		suppressedUntil                                                                              *time.Time
		suppressionReason                                                                            *string
		deliveredAt, metadata                                                                        string
		createdAt, updatedAt                                                                         time.Time
	)

	err := database.QueryRow(`
		SELECT a.id, a.alert_number, a.title, a.description, a.severity, a.status,
			a.control_id, c.identifier, c.title, c.category,
			a.test_id, t.identifier, t.title, t.test_type,
			a.test_result_id, a.alert_rule_id, ar.name,
			a.assigned_to, a.assigned_at, a.assigned_by,
			a.sla_deadline, a.sla_breached,
			a.resolved_by, a.resolved_at, a.resolution_notes,
			a.suppressed_until, a.suppression_reason,
			a.delivered_at, a.metadata,
			a.created_at, a.updated_at
		FROM alerts a
		LEFT JOIN controls c ON c.id = a.control_id
		LEFT JOIN tests t ON t.id = a.test_id
		LEFT JOIN alert_rules ar ON ar.id = a.alert_rule_id
		WHERE a.id = $1 AND a.org_id = $2
	`, alertID, orgID).Scan(
		&aID, &alertNumber, &aTitle, &aDescription, &aSeverity, &aStatus,
		&controlID, &controlIdentifier, &controlTitle, &controlCategory,
		&testID, &testIdentifier, &testTitle, &testType,
		&testResultID, &alertRuleID, &alertRuleName,
		&assignedTo, &assignedAt, &assignedBy,
		&slaDeadline, &slaBreached,
		&resolvedBy, &resolvedAt, &resolutionNotes,
		&suppressedUntil, &suppressionReason,
		&deliveredAt, &metadata,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get alert"))
		return
	}

	var deliveredAtObj, metadataObj interface{}
	json.Unmarshal([]byte(deliveredAt), &deliveredAtObj)
	json.Unmarshal([]byte(metadata), &metadataObj)

	result := gin.H{
		"id":           aID,
		"alert_number": alertNumber,
		"title":        aTitle,
		"description":  aDescription,
		"severity":     aSeverity,
		"status":       aStatus,
		"control": gin.H{
			"id":         controlID,
			"identifier": controlIdentifier,
			"title":      controlTitle,
			"category":   controlCategory,
		},
		"test_result_id":     testResultID,
		"assigned_to":        assignedTo,
		"assigned_at":        assignedAt,
		"assigned_by":        assignedBy,
		"sla_deadline":       slaDeadline,
		"sla_breached":       slaBreached,
		"resolved_by":        resolvedBy,
		"resolved_at":        resolvedAt,
		"resolution_notes":   resolutionNotes,
		"suppressed_until":   suppressedUntil,
		"suppression_reason": suppressionReason,
		"delivered_at":       deliveredAtObj,
		"metadata":           metadataObj,
		"created_at":         createdAt,
		"updated_at":         updatedAt,
	}

	if testID != nil {
		result["test"] = gin.H{
			"id":         *testID,
			"identifier": *testIdentifier,
			"title":      *testTitle,
			"test_type":  *testType,
		}
	}
	if alertRuleID != nil {
		result["alert_rule"] = gin.H{
			"id":   *alertRuleID,
			"name": *alertRuleName,
		}
	}

	if slaDeadline != nil {
		remaining := time.Until(*slaDeadline).Hours()
		result["hours_remaining"] = math.Round(remaining*10) / 10
	}

	c.JSON(http.StatusOK, successResponse(c, result))
}

// ChangeAlertStatus changes an alert's lifecycle status.
func ChangeAlertStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	alertID := c.Param("id")

	var req models.ChangeAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Status is required"))
		return
	}

	if !models.IsValidAlertStatus(req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid status"))
		return
	}

	var currentStatus string
	var alertNumber int
	err := database.QueryRow(
		"SELECT status, alert_number FROM alerts WHERE id = $1 AND org_id = $2",
		alertID, orgID,
	).Scan(&currentStatus, &alertNumber)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change status"))
		return
	}

	if !models.IsValidAlertStatusTransition(currentStatus, req.Status) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
			fmt.Sprintf("Cannot transition from '%s' to '%s'", currentStatus, req.Status)))
		return
	}

	_, err = database.Exec(
		"UPDATE alerts SET status = $1, updated_at = NOW() WHERE id = $2 AND org_id = $3",
		req.Status, alertID, orgID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update alert status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change status"))
		return
	}

	middleware.LogAudit(c, "alert.status_changed", "alert", &alertID, map[string]interface{}{
		"alert_number": alertNumber,
		"old_status":   currentStatus,
		"new_status":   req.Status,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":              alertID,
		"alert_number":    alertNumber,
		"status":          req.Status,
		"previous_status": currentStatus,
		"message":         fmt.Sprintf("Alert %s.", req.Status),
	}))
}

// AssignAlert assigns an alert to a user.
func AssignAlert(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	alertID := c.Param("id")

	var req models.AssignAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "assigned_to is required"))
		return
	}

	// Check alert exists and is assignable
	var currentStatus string
	var alertNumber int
	err := database.QueryRow(
		"SELECT status, alert_number FROM alerts WHERE id = $1 AND org_id = $2",
		alertID, orgID,
	).Scan(&currentStatus, &alertNumber)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to assign alert"))
		return
	}

	if currentStatus == "closed" || currentStatus == "resolved" {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
			fmt.Sprintf("Cannot assign alert in '%s' status", currentStatus)))
		return
	}

	// Verify assignee exists in same org
	var assigneeName, assigneeEmail string
	err = database.QueryRow(
		"SELECT COALESCE(first_name || ' ' || last_name, email), email FROM users WHERE id = $1 AND org_id = $2",
		req.AssignedTo, orgID,
	).Scan(&assigneeName, &assigneeEmail)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found in this organization"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check assignee")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to assign alert"))
		return
	}

	// Auto-transition to acknowledged if currently open
	newStatus := currentStatus
	if currentStatus == "open" {
		newStatus = "acknowledged"
	}

	now := time.Now()
	_, err = database.Exec(`
		UPDATE alerts SET assigned_to = $1, assigned_at = $2, assigned_by = $3,
			status = $4, updated_at = NOW()
		WHERE id = $5 AND org_id = $6
	`, req.AssignedTo, now, userID, newStatus, alertID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to assign alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to assign alert"))
		return
	}

	middleware.LogAudit(c, "alert.assigned", "alert", &alertID, map[string]interface{}{
		"alert_number": alertNumber,
		"assigned_to":  req.AssignedTo,
		"assigned_by":  userID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":           alertID,
		"alert_number": alertNumber,
		"assigned_to": gin.H{
			"id":    req.AssignedTo,
			"name":  assigneeName,
			"email": assigneeEmail,
		},
		"assigned_at": now,
		"assigned_by": gin.H{
			"id": userID,
		},
		"message": fmt.Sprintf("Alert assigned to %s.", assigneeName),
	}))
}

// ResolveAlert marks an alert as resolved.
func ResolveAlert(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	alertID := c.Param("id")

	var req models.ResolveAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "resolution_notes is required"))
		return
	}

	if len(req.ResolutionNotes) > 10000 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Resolution notes must be 10000 characters or less"))
		return
	}

	var currentStatus string
	var alertNumber int
	err := database.QueryRow(
		"SELECT status, alert_number FROM alerts WHERE id = $1 AND org_id = $2",
		alertID, orgID,
	).Scan(&currentStatus, &alertNumber)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to resolve alert"))
		return
	}

	if currentStatus != "open" && currentStatus != "acknowledged" && currentStatus != "in_progress" {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
			fmt.Sprintf("Cannot resolve alert in '%s' status", currentStatus)))
		return
	}

	now := time.Now()
	_, err = database.Exec(`
		UPDATE alerts SET status = 'resolved', resolved_by = $1, resolved_at = $2,
			resolution_notes = $3, updated_at = NOW()
		WHERE id = $4 AND org_id = $5
	`, userID, now, req.ResolutionNotes, alertID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to resolve alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to resolve alert"))
		return
	}

	middleware.LogAudit(c, "alert.resolved", "alert", &alertID, map[string]interface{}{
		"alert_number":    alertNumber,
		"resolution_notes": req.ResolutionNotes,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":              alertID,
		"alert_number":    alertNumber,
		"status":          "resolved",
		"previous_status": currentStatus,
		"resolved_by": gin.H{
			"id": userID,
		},
		"resolved_at":      now,
		"resolution_notes":  req.ResolutionNotes,
		"message":          "Alert resolved. Will be verified on next test run.",
	}))
}

// SuppressAlert suppresses (snoozes) an alert.
func SuppressAlert(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	alertID := c.Param("id")

	var req models.SuppressAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "suppressed_until and suppression_reason are required"))
		return
	}

	// Parse suppressed_until
	suppressedUntil, err := time.Parse(time.RFC3339, req.SuppressedUntil)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "suppressed_until must be a valid ISO 8601 datetime"))
		return
	}
	if suppressedUntil.Before(time.Now()) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "suppressed_until must be in the future"))
		return
	}
	if suppressedUntil.After(time.Now().Add(90 * 24 * time.Hour)) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "suppressed_until cannot be more than 90 days in the future"))
		return
	}
	if len(req.SuppressionReason) < 20 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Suppression reason must be at least 20 characters"))
		return
	}
	if len(req.SuppressionReason) > 5000 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Suppression reason must be 5000 characters or less"))
		return
	}

	var currentStatus string
	var alertNumber int
	err = database.QueryRow(
		"SELECT status, alert_number FROM alerts WHERE id = $1 AND org_id = $2",
		alertID, orgID,
	).Scan(&currentStatus, &alertNumber)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to suppress alert"))
		return
	}

	if currentStatus == "closed" {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Cannot suppress a closed alert"))
		return
	}

	_, err = database.Exec(`
		UPDATE alerts SET status = 'suppressed', suppressed_until = $1, suppression_reason = $2,
			updated_at = NOW()
		WHERE id = $3 AND org_id = $4
	`, suppressedUntil, req.SuppressionReason, alertID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to suppress alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to suppress alert"))
		return
	}

	middleware.LogAudit(c, "alert.suppressed", "alert", &alertID, map[string]interface{}{
		"alert_number": alertNumber,
		"until":        suppressedUntil,
		"reason":       req.SuppressionReason,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                 alertID,
		"alert_number":       alertNumber,
		"status":             "suppressed",
		"previous_status":    currentStatus,
		"suppressed_until":   suppressedUntil,
		"suppression_reason": req.SuppressionReason,
		"message":            fmt.Sprintf("Alert suppressed until %s.", suppressedUntil.Format("Jan 2, 2006 3:04 PM UTC")),
	}))
}

// CloseAlert closes an alert.
func CloseAlert(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	alertID := c.Param("id")

	var req models.CloseAlertRequest
	c.ShouldBindJSON(&req) // optional body

	var currentStatus string
	var alertNumber int
	err := database.QueryRow(
		"SELECT status, alert_number FROM alerts WHERE id = $1 AND org_id = $2",
		alertID, orgID,
	).Scan(&currentStatus, &alertNumber)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to close alert"))
		return
	}

	updateQuery := "UPDATE alerts SET status = 'closed', updated_at = NOW()"
	args := []interface{}{}
	argN := 1

	if req.ResolutionNotes != nil {
		updateQuery += fmt.Sprintf(", resolution_notes = $%d", argN)
		args = append(args, *req.ResolutionNotes)
		argN++
	}

	args = append(args, alertID, orgID)
	updateQuery += fmt.Sprintf(" WHERE id = $%d AND org_id = $%d", argN, argN+1)

	_, err = database.Exec(updateQuery, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to close alert")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to close alert"))
		return
	}

	middleware.LogAudit(c, "alert.closed", "alert", &alertID, map[string]interface{}{
		"alert_number": alertNumber,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":              alertID,
		"alert_number":    alertNumber,
		"status":          "closed",
		"previous_status": currentStatus,
		"message":         "Alert closed.",
	}))
}
