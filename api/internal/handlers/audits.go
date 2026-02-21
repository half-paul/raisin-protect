package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// checkAuditAccess verifies org_id isolation and auditor access.
// Returns the audit status or empty string if not found/no access.
// If denied, it writes the error response to c.
func checkAuditAccess(c *gin.Context, auditID, orgID, userID, userRole string) (status string, ok bool) {
	var auditStatus string
	var auditorIDs pq.StringArray
	err := database.DB.QueryRow(
		"SELECT status, auditor_ids FROM audits WHERE id = $1 AND org_id = $2",
		auditID, orgID,
	).Scan(&auditStatus, &auditorIDs)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_NOT_FOUND", "Audit not found"))
		return "", false
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check audit access")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to check audit access"))
		return "", false
	}

	// Auditor isolation: auditors can only see audits they're assigned to
	if userRole == models.RoleAuditor {
		found := false
		for _, id := range auditorIDs {
			if id == userID {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, errorResponse("AUDIT_NOT_FOUND", "Audit not found"))
			return "", false
		}
	}

	return auditStatus, true
}

// checkAuditNotTerminal verifies the audit isn't completed or cancelled.
func checkAuditNotTerminal(c *gin.Context, status string) bool {
	if status == models.AuditStatusCompleted {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_COMPLETED", "Cannot modify a completed audit"))
		return false
	}
	if status == models.AuditStatusCancelled {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_CANCELLED", "Cannot modify a cancelled audit"))
		return false
	}
	return true
}

// ListAudits lists audits with filtering and pagination.
func ListAudits(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

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

	// Auditor isolation
	if userRole == models.RoleAuditor {
		where = append(where, fmt.Sprintf("$%d = ANY(a.auditor_ids)", argN))
		args = append(args, userID)
		argN++
	}

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("a.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("audit_type"); v != "" {
		where = append(where, fmt.Sprintf("a.audit_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("framework_id"); v != "" {
		where = append(where, fmt.Sprintf("a.org_framework_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(a.title ILIKE $%d OR a.description ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	sortCol := "a.created_at"
	switch c.DefaultQuery("sort", "created_at") {
	case "planned_end":
		sortCol = "a.planned_end"
	case "title":
		sortCol = "a.title"
	case "status":
		sortCol = "a.status"
	}
	order := "DESC"
	if strings.ToLower(c.DefaultQuery("order", "desc")) == "asc" {
		order = "ASC"
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM audits a WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count audits")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to count audits"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT a.id, a.title, a.description, a.audit_type, a.status,
		       a.org_framework_id, COALESCE(of.display_name, ''),
		       a.period_start, a.period_end, a.planned_start, a.planned_end,
		       a.actual_start, a.actual_end,
		       a.audit_firm, a.lead_auditor_id,
		       COALESCE(la.first_name || ' ' || la.last_name, ''),
		       a.internal_lead_id,
		       COALESCE(il.first_name || ' ' || il.last_name, ''),
		       a.total_requests, a.open_requests, a.total_findings, a.open_findings,
		       a.tags, a.created_at, a.updated_at
		FROM audits a
		LEFT JOIN org_frameworks of ON a.org_framework_id = of.id
		LEFT JOIN users la ON a.lead_auditor_id = la.id
		LEFT JOIN users il ON a.internal_lead_id = il.id
		WHERE %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list audits")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list audits"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, title, auditType, status                        string
			description, orgFrameworkID, frameworkName           *string
			periodStart, periodEnd, plannedStart, plannedEnd    *time.Time
			actualStart, actualEnd                              *time.Time
			auditFirm, leadAuditorID, leadAuditorName           *string
			internalLeadID, internalLeadName                    *string
			totalRequests, openRequests, totalFindings, openFindings int
			tags                                                 pq.StringArray
			createdAt, updatedAt                                 time.Time
		)

		if err := rows.Scan(
			&id, &title, &description, &auditType, &status,
			&orgFrameworkID, &frameworkName,
			&periodStart, &periodEnd, &plannedStart, &plannedEnd,
			&actualStart, &actualEnd,
			&auditFirm, &leadAuditorID, &leadAuditorName,
			&internalLeadID, &internalLeadName,
			&totalRequests, &openRequests, &totalFindings, &openFindings,
			&tags, &createdAt, &updatedAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan audit row")
			continue
		}

		result := gin.H{
			"id":                id,
			"title":             title,
			"description":       description,
			"audit_type":        auditType,
			"status":            status,
			"org_framework_id":  orgFrameworkID,
			"framework_name":    frameworkName,
			"period_start":      periodStart,
			"period_end":        periodEnd,
			"planned_start":     plannedStart,
			"planned_end":       plannedEnd,
			"actual_start":      actualStart,
			"actual_end":        actualEnd,
			"audit_firm":        auditFirm,
			"lead_auditor_id":   leadAuditorID,
			"lead_auditor_name": leadAuditorName,
			"internal_lead_id":  internalLeadID,
			"internal_lead_name": internalLeadName,
			"total_requests":    totalRequests,
			"open_requests":     openRequests,
			"total_findings":    totalFindings,
			"open_findings":     openFindings,
			"tags":              []string(tags),
			"created_at":        createdAt,
			"updated_at":        updatedAt,
		}
		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"pagination": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": (total + perPage - 1) / perPage,
		},
	})
}

// GetAudit gets a single audit with full detail.
func GetAudit(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}

	var (
		id, title, auditType                                  string
		description, orgFrameworkID, frameworkName             *string
		periodStart, periodEnd, plannedStart, plannedEnd      *time.Time
		actualStart, actualEnd                                *time.Time
		auditFirm, leadAuditorID, leadAuditorName             *string
		internalLeadID, internalLeadName                      *string
		auditorIDs                                             pq.StringArray
		milestonesJSON                                         string
		reportType, reportURL                                  *string
		reportIssuedAt                                         *time.Time
		totalRequests, openRequests, totalFindings, openFindings int
		tags                                                   pq.StringArray
		metadataJSON                                           string
		createdAt, updatedAt                                   time.Time
	)

	err := database.DB.QueryRow(`
		SELECT a.id, a.title, a.description, a.audit_type, a.status,
		       a.org_framework_id, COALESCE(of.display_name, ''),
		       a.period_start, a.period_end, a.planned_start, a.planned_end,
		       a.actual_start, a.actual_end,
		       a.audit_firm, a.lead_auditor_id,
		       COALESCE(la.first_name || ' ' || la.last_name, ''),
		       a.internal_lead_id,
		       COALESCE(il.first_name || ' ' || il.last_name, ''),
		       a.auditor_ids, COALESCE(a.milestones::text, '[]'),
		       a.report_type, a.report_url, a.report_issued_at,
		       a.total_requests, a.open_requests, a.total_findings, a.open_findings,
		       a.tags, COALESCE(a.metadata::text, '{}'),
		       a.created_at, a.updated_at
		FROM audits a
		LEFT JOIN org_frameworks of ON a.org_framework_id = of.id
		LEFT JOIN users la ON a.lead_auditor_id = la.id
		LEFT JOIN users il ON a.internal_lead_id = il.id
		WHERE a.id = $1 AND a.org_id = $2
	`, auditID, orgID).Scan(
		&id, &title, &description, &auditType, &auditStatus,
		&orgFrameworkID, &frameworkName,
		&periodStart, &periodEnd, &plannedStart, &plannedEnd,
		&actualStart, &actualEnd,
		&auditFirm, &leadAuditorID, &leadAuditorName,
		&internalLeadID, &internalLeadName,
		&auditorIDs, &milestonesJSON,
		&reportType, &reportURL, &reportIssuedAt,
		&totalRequests, &openRequests, &totalFindings, &openFindings,
		&tags, &metadataJSON,
		&createdAt, &updatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get audit detail")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get audit"))
		return
	}

	var milestones interface{}
	json.Unmarshal([]byte(milestonesJSON), &milestones)
	var metadata interface{}
	json.Unmarshal([]byte(metadataJSON), &metadata)

	result := gin.H{
		"id":                 id,
		"title":              title,
		"description":        description,
		"audit_type":         auditType,
		"status":             auditStatus,
		"org_framework_id":   orgFrameworkID,
		"framework_name":     frameworkName,
		"period_start":       periodStart,
		"period_end":         periodEnd,
		"planned_start":      plannedStart,
		"planned_end":        plannedEnd,
		"actual_start":       actualStart,
		"actual_end":         actualEnd,
		"audit_firm":         auditFirm,
		"lead_auditor_id":    leadAuditorID,
		"lead_auditor_name":  leadAuditorName,
		"internal_lead_id":   internalLeadID,
		"internal_lead_name": internalLeadName,
		"auditor_ids":        []string(auditorIDs),
		"milestones":         milestones,
		"report_type":        reportType,
		"report_url":         reportURL,
		"report_issued_at":   reportIssuedAt,
		"total_requests":     totalRequests,
		"open_requests":      openRequests,
		"total_findings":     totalFindings,
		"open_findings":      openFindings,
		"tags":               []string(tags),
		"metadata":           metadata,
		"created_at":         createdAt,
		"updated_at":         updatedAt,
	}

	c.JSON(http.StatusOK, successResponse(c, result))
}

// CreateAudit creates a new audit engagement.
func CreateAudit(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.CreateAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if utf8.RuneCountInString(req.Title) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must not exceed 255 characters"))
		return
	}
	if !models.IsValidAuditType(req.AuditType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid audit_type"))
		return
	}

	// Serialize milestones
	milestonesJSON := "[]"
	if len(req.Milestones) > 0 {
		b, _ := json.Marshal(req.Milestones)
		milestonesJSON = string(b)
	}

	auditID := uuid.New().String()
	now := time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO audits (id, org_id, title, description, audit_type, status,
		                     org_framework_id, period_start, period_end,
		                     planned_start, planned_end,
		                     audit_firm, lead_auditor_id, auditor_ids, internal_lead_id,
		                     milestones, report_type, tags,
		                     created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'planning',
		        $6, $7::date, $8::date, $9::date, $10::date,
		        $11, $12, $13, $14, $15::jsonb, $16, $17, $18, $18)
	`, auditID, orgID, req.Title, req.Description, req.AuditType,
		req.OrgFrameworkID, req.PeriodStart, req.PeriodEnd,
		req.PlannedStart, req.PlannedEnd,
		req.AuditFirm, req.LeadAuditorID, pq.Array(req.AuditorIDs), req.InternalLeadID,
		milestonesJSON, req.ReportType, pq.Array(req.Tags), now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create audit")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create audit"))
		return
	}

	middleware.LogAudit(c, "audit.created", "audit", &auditID, map[string]interface{}{
		"title":      req.Title,
		"audit_type": req.AuditType,
	})

	// Fetch full record for response
	c.Set("redirect_get_audit", true)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: auditID})
	// Return created
	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":         auditID,
		"title":      req.Title,
		"audit_type": req.AuditType,
		"status":     "planning",
		"created_at": now,
	}))
}

// UpdateAudit updates an audit engagement.
func UpdateAudit(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.UpdateAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	sets := []string{}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		if utf8.RuneCountInString(*req.Title) > 255 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must not exceed 255 characters"))
			return
		}
		sets = append(sets, fmt.Sprintf("title = $%d", argN))
		args = append(args, *req.Title)
		argN++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argN))
		args = append(args, *req.Description)
		argN++
	}
	if req.OrgFrameworkID != nil {
		sets = append(sets, fmt.Sprintf("org_framework_id = $%d", argN))
		args = append(args, *req.OrgFrameworkID)
		argN++
	}
	if req.PeriodStart != nil {
		sets = append(sets, fmt.Sprintf("period_start = $%d::date", argN))
		args = append(args, *req.PeriodStart)
		argN++
	}
	if req.PeriodEnd != nil {
		sets = append(sets, fmt.Sprintf("period_end = $%d::date", argN))
		args = append(args, *req.PeriodEnd)
		argN++
	}
	if req.PlannedStart != nil {
		sets = append(sets, fmt.Sprintf("planned_start = $%d::date", argN))
		args = append(args, *req.PlannedStart)
		argN++
	}
	if req.PlannedEnd != nil {
		sets = append(sets, fmt.Sprintf("planned_end = $%d::date", argN))
		args = append(args, *req.PlannedEnd)
		argN++
	}
	if req.AuditFirm != nil {
		sets = append(sets, fmt.Sprintf("audit_firm = $%d", argN))
		args = append(args, *req.AuditFirm)
		argN++
	}
	if req.LeadAuditorID != nil {
		sets = append(sets, fmt.Sprintf("lead_auditor_id = $%d", argN))
		args = append(args, *req.LeadAuditorID)
		argN++
	}
	if req.InternalLeadID != nil {
		sets = append(sets, fmt.Sprintf("internal_lead_id = $%d", argN))
		args = append(args, *req.InternalLeadID)
		argN++
	}
	if req.Milestones != nil {
		b, _ := json.Marshal(req.Milestones)
		sets = append(sets, fmt.Sprintf("milestones = $%d::jsonb", argN))
		args = append(args, string(b))
		argN++
	}
	if req.ReportType != nil {
		sets = append(sets, fmt.Sprintf("report_type = $%d", argN))
		args = append(args, *req.ReportType)
		argN++
	}
	if req.ReportURL != nil {
		sets = append(sets, fmt.Sprintf("report_url = $%d", argN))
		args = append(args, *req.ReportURL)
		argN++
	}
	if req.Tags != nil {
		sets = append(sets, fmt.Sprintf("tags = $%d", argN))
		args = append(args, pq.Array(req.Tags))
		argN++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "No fields to update"))
		return
	}

	now := time.Now()
	sets = append(sets, fmt.Sprintf("updated_at = $%d", argN))
	args = append(args, now)
	argN++

	args = append(args, auditID, orgID)
	query := fmt.Sprintf("UPDATE audits SET %s WHERE id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1)

	_, err := database.DB.Exec(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update audit")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update audit"))
		return
	}

	middleware.LogAudit(c, "audit.updated", "audit", &auditID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         auditID,
		"updated_at": now,
	}))
}

// ChangeAuditStatus transitions an audit's status.
func ChangeAuditStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}

	var req models.ChangeAuditStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if !models.IsValidAuditStatusTransition(auditStatus, req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_INVALID_TRANSITION",
			fmt.Sprintf("Invalid status transition from '%s' to '%s'", auditStatus, req.Status)))
		return
	}

	now := time.Now()
	sets := "status = $1, updated_at = $2"
	args := []interface{}{req.Status, now}
	argN := 3

	// Set actual_start when moving to fieldwork
	if req.Status == models.AuditStatusFieldwork && auditStatus == models.AuditStatusPlanning {
		sets += fmt.Sprintf(", actual_start = $%d", argN)
		args = append(args, now)
		argN++
	}
	// Set actual_end when completing
	if req.Status == models.AuditStatusCompleted {
		sets += fmt.Sprintf(", actual_end = $%d", argN)
		args = append(args, now)
		argN++
	}

	args = append(args, auditID, orgID)
	query := fmt.Sprintf("UPDATE audits SET %s WHERE id = $%d AND org_id = $%d", sets, argN, argN+1)

	_, err := database.DB.Exec(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to change audit status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change audit status"))
		return
	}

	middleware.LogAudit(c, "audit.status_changed", "audit", &auditID, map[string]interface{}{
		"old_status": auditStatus,
		"new_status": req.Status,
		"notes":      req.Notes,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         auditID,
		"status":     req.Status,
		"updated_at": now,
	}))
}

// AddAuditAuditor adds an auditor to the engagement.
func AddAuditAuditor(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.AddAuditorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Verify user exists and is an auditor
	var userExists bool
	var targetRole string
	err := database.DB.QueryRow(
		"SELECT TRUE, role FROM users WHERE id = $1 AND org_id = $2 AND status = 'active'",
		req.UserID, orgID,
	).Scan(&userExists, &targetRole)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check user")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to add auditor"))
		return
	}
	if targetRole != models.RoleAuditor {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "User is not an auditor"))
		return
	}

	// Check not already in list
	var currentAuditorIDs pq.StringArray
	database.DB.QueryRow("SELECT auditor_ids FROM audits WHERE id = $1", auditID).Scan(&currentAuditorIDs)
	for _, id := range currentAuditorIDs {
		if id == req.UserID {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "User is already an auditor for this audit"))
			return
		}
	}

	now := time.Now()
	_, err = database.DB.Exec(
		"UPDATE audits SET auditor_ids = array_append(auditor_ids, $1), updated_at = $2 WHERE id = $3 AND org_id = $4",
		req.UserID, now, auditID, orgID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add auditor")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to add auditor"))
		return
	}

	middleware.LogAudit(c, "audit.auditor_added", "audit", &auditID, map[string]interface{}{
		"user_id": req.UserID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         auditID,
		"added_user": req.UserID,
		"updated_at": now,
	}))
}

// RemoveAuditAuditor removes an auditor from the engagement.
func RemoveAuditAuditor(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	targetUserID := c.Param("user_id")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	now := time.Now()
	_, err := database.DB.Exec(
		"UPDATE audits SET auditor_ids = array_remove(auditor_ids, $1), updated_at = $2 WHERE id = $3 AND org_id = $4",
		targetUserID, now, auditID, orgID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove auditor")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to remove auditor"))
		return
	}

	middleware.LogAudit(c, "audit.auditor_removed", "audit", &auditID, map[string]interface{}{
		"user_id": targetUserID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":           auditID,
		"removed_user": targetUserID,
		"updated_at":   now,
	}))
}
