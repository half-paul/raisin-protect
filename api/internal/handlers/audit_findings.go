package handlers

import (
	"database/sql"
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

// updateAuditFindingCounts recalculates denormalized finding counts on the audit.
func updateAuditFindingCounts(auditID, orgID string) {
	database.DB.Exec(`
		UPDATE audits SET
			total_findings = (SELECT COUNT(*) FROM audit_findings WHERE audit_id = $1 AND org_id = $2),
			open_findings = (SELECT COUNT(*) FROM audit_findings WHERE audit_id = $1 AND org_id = $2 AND status NOT IN ('verified', 'closed', 'risk_accepted'))
		WHERE id = $1 AND org_id = $2
	`, auditID, orgID)
}

// ListAuditFindings lists findings for an audit.
func ListAuditFindings(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"af.audit_id = $1", "af.org_id = $2"}
	args := []interface{}{auditID, orgID}
	argN := 3

	if v := c.Query("severity"); v != "" {
		where = append(where, fmt.Sprintf("af.severity = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("af.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("category"); v != "" {
		where = append(where, fmt.Sprintf("af.category = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("remediation_owner_id"); v != "" {
		where = append(where, fmt.Sprintf("af.remediation_owner_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(af.title ILIKE $%d OR af.description ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	sortCol := "af.created_at"
	switch c.DefaultQuery("sort", "created_at") {
	case "severity":
		sortCol = "af.severity"
	case "status":
		sortCol = "af.status"
	case "remediation_due_date":
		sortCol = "af.remediation_due_date"
	}
	order := "DESC"
	if strings.ToLower(c.DefaultQuery("order", "desc")) == "asc" {
		order = "ASC"
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM audit_findings af WHERE %s", whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT af.id, af.audit_id, af.title, af.description, af.severity, af.category, af.status,
		       af.control_id, COALESCE(ctrl.title, ''),
		       af.requirement_id, COALESCE(req.title, ''),
		       af.found_by, COALESCE(fb.first_name || ' ' || fb.last_name, ''),
		       af.remediation_owner_id, COALESCE(ro.first_name || ' ' || ro.last_name, ''),
		       af.remediation_plan, af.remediation_due_date,
		       af.remediation_started_at, af.remediation_completed_at,
		       af.verified_at, af.reference_number, af.recommendation,
		       af.management_response, af.risk_accepted,
		       (SELECT COUNT(*) FROM audit_comments ac WHERE ac.target_type = 'finding' AND ac.target_id = af.id),
		       af.tags, af.created_at, af.updated_at
		FROM audit_findings af
		LEFT JOIN controls ctrl ON af.control_id = ctrl.id
		LEFT JOIN requirements req ON af.requirement_id = req.id
		LEFT JOIN users fb ON af.found_by = fb.id
		LEFT JOIN users ro ON af.remediation_owner_id = ro.id
		WHERE %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list audit findings")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list findings"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, aID, title, description, severity, category, status string
			controlID, controlTitle                                  *string
			requirementID, requirementTitle                          *string
			foundBy, foundByName                                    *string
			remOwnerID, remOwnerName                                *string
			remPlan                                                 *string
			remDueDate                                              *time.Time
			remStartedAt, remCompletedAt, verifiedAt                *time.Time
			refNumber, recommendation, mgmtResponse                 *string
			riskAccepted                                            bool
			commentCount                                            int
			tags                                                    pq.StringArray
			createdAt, updatedAt                                    time.Time
		)
		if err := rows.Scan(
			&id, &aID, &title, &description, &severity, &category, &status,
			&controlID, &controlTitle,
			&requirementID, &requirementTitle,
			&foundBy, &foundByName,
			&remOwnerID, &remOwnerName,
			&remPlan, &remDueDate,
			&remStartedAt, &remCompletedAt, &verifiedAt,
			&refNumber, &recommendation, &mgmtResponse, &riskAccepted,
			&commentCount,
			&tags, &createdAt, &updatedAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan finding row")
			continue
		}

		results = append(results, gin.H{
			"id": id, "audit_id": aID,
			"title": title, "description": description,
			"severity": severity, "category": category, "status": status,
			"control_id": controlID, "control_title": controlTitle,
			"requirement_id": requirementID, "requirement_title": requirementTitle,
			"found_by": foundBy, "found_by_name": foundByName,
			"remediation_owner_id": remOwnerID, "remediation_owner_name": remOwnerName,
			"remediation_plan": remPlan, "remediation_due_date": remDueDate,
			"remediation_started_at": remStartedAt, "remediation_completed_at": remCompletedAt,
			"verified_at": verifiedAt,
			"reference_number": refNumber, "recommendation": recommendation,
			"management_response": mgmtResponse, "risk_accepted": riskAccepted,
			"comment_count": commentCount,
			"tags": []string(tags), "created_at": createdAt, "updated_at": updatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       results,
		"pagination": gin.H{"page": page, "per_page": perPage, "total": total, "total_pages": (total + perPage - 1) / perPage},
	})
}

// GetAuditFinding gets a single finding with full detail.
func GetAuditFinding(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	findingID := c.Param("fid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var (
		id, aID, title, description, severity, category, status string
		controlID, controlTitle                                  *string
		requirementID, requirementTitle                          *string
		foundBy, foundByName                                    *string
		remOwnerID, remOwnerName                                *string
		remPlan                                                 *string
		remDueDate                                              *time.Time
		remStartedAt, remCompletedAt                            *time.Time
		verificationNotes                                       *string
		verifiedAt                                              *time.Time
		verifiedBy                                              *string
		riskAccepted                                            bool
		riskAcceptReason                                        *string
		riskAcceptedBy                                          *string
		riskAcceptedAt                                          *time.Time
		refNumber, recommendation, mgmtResponse                 *string
		tags                                                    pq.StringArray
		metadataJSON                                            string
		createdAt, updatedAt                                    time.Time
	)

	err := database.DB.QueryRow(`
		SELECT af.id, af.audit_id, af.title, af.description, af.severity, af.category, af.status,
		       af.control_id, COALESCE(ctrl.title, ''),
		       af.requirement_id, COALESCE(req.title, ''),
		       af.found_by, COALESCE(fb.first_name || ' ' || fb.last_name, ''),
		       af.remediation_owner_id, COALESCE(ro.first_name || ' ' || ro.last_name, ''),
		       af.remediation_plan, af.remediation_due_date,
		       af.remediation_started_at, af.remediation_completed_at,
		       af.verification_notes, af.verified_at, af.verified_by,
		       af.risk_accepted, af.risk_acceptance_reason, af.risk_accepted_by, af.risk_accepted_at,
		       af.reference_number, af.recommendation, af.management_response,
		       af.tags, COALESCE(af.metadata::text, '{}'),
		       af.created_at, af.updated_at
		FROM audit_findings af
		LEFT JOIN controls ctrl ON af.control_id = ctrl.id
		LEFT JOIN requirements req ON af.requirement_id = req.id
		LEFT JOIN users fb ON af.found_by = fb.id
		LEFT JOIN users ro ON af.remediation_owner_id = ro.id
		WHERE af.id = $1 AND af.audit_id = $2 AND af.org_id = $3
	`, findingID, auditID, orgID).Scan(
		&id, &aID, &title, &description, &severity, &category, &status,
		&controlID, &controlTitle,
		&requirementID, &requirementTitle,
		&foundBy, &foundByName,
		&remOwnerID, &remOwnerName,
		&remPlan, &remDueDate,
		&remStartedAt, &remCompletedAt,
		&verificationNotes, &verifiedAt, &verifiedBy,
		&riskAccepted, &riskAcceptReason, &riskAcceptedBy, &riskAcceptedAt,
		&refNumber, &recommendation, &mgmtResponse,
		&tags, &metadataJSON,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_FINDING_NOT_FOUND", "Finding not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get finding")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get finding"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": id, "audit_id": aID,
		"title": title, "description": description,
		"severity": severity, "category": category, "status": status,
		"control_id": controlID, "control_title": controlTitle,
		"requirement_id": requirementID, "requirement_title": requirementTitle,
		"found_by": foundBy, "found_by_name": foundByName,
		"remediation_owner_id": remOwnerID, "remediation_owner_name": remOwnerName,
		"remediation_plan": remPlan, "remediation_due_date": remDueDate,
		"remediation_started_at": remStartedAt, "remediation_completed_at": remCompletedAt,
		"verification_notes": verificationNotes, "verified_at": verifiedAt, "verified_by": verifiedBy,
		"risk_accepted": riskAccepted, "risk_acceptance_reason": riskAcceptReason,
		"risk_accepted_by": riskAcceptedBy, "risk_accepted_at": riskAcceptedAt,
		"reference_number": refNumber, "recommendation": recommendation,
		"management_response": mgmtResponse,
		"tags": []string(tags), "metadata": metadataJSON,
		"created_at": createdAt, "updated_at": updatedAt,
	}))
}

// CreateAuditFinding creates a new finding.
func CreateAuditFinding(c *gin.Context) {
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

	var req models.CreateFindingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if utf8.RuneCountInString(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must not exceed 500 characters"))
		return
	}
	if !models.IsValidFindingSeverity(req.Severity) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid severity"))
		return
	}
	if !models.IsValidFindingCategory(req.Category) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid category"))
		return
	}

	findingID := uuid.New().String()
	now := time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO audit_findings (id, org_id, audit_id, title, description, severity, category, status,
		                            control_id, requirement_id, found_by,
		                            remediation_owner_id, remediation_due_date,
		                            reference_number, recommendation, tags,
		                            created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'identified',
		        $8, $9, $10, $11, $12::date, $13, $14, $15, $16, $16)
	`, findingID, orgID, auditID, req.Title, req.Description, req.Severity, req.Category,
		req.ControlID, req.RequirementID, userID,
		req.RemediationOwnerID, req.RemediationDueDate,
		req.ReferenceNumber, req.Recommendation, pq.Array(req.Tags), now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create finding")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create finding"))
		return
	}

	updateAuditFindingCounts(auditID, orgID)
	middleware.LogAudit(c, "audit_finding.created", "audit_finding", &findingID, map[string]interface{}{
		"audit_id": auditID, "severity": req.Severity, "category": req.Category,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id": findingID, "audit_id": auditID,
		"title": req.Title, "severity": req.Severity,
		"category": req.Category, "status": "identified",
		"created_at": now,
	}))
}

// UpdateAuditFinding updates finding metadata.
func UpdateAuditFinding(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	findingID := c.Param("fid")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.UpdateFindingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	var exists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_findings WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
		findingID, auditID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_FINDING_NOT_FOUND", "Finding not found"))
		return
	}

	sets := []string{}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argN))
		args = append(args, *req.Title)
		argN++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argN))
		args = append(args, *req.Description)
		argN++
	}
	if req.Severity != nil {
		if !models.IsValidFindingSeverity(*req.Severity) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid severity"))
			return
		}
		sets = append(sets, fmt.Sprintf("severity = $%d", argN))
		args = append(args, *req.Severity)
		argN++
	}
	if req.Category != nil {
		if !models.IsValidFindingCategory(*req.Category) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid category"))
			return
		}
		sets = append(sets, fmt.Sprintf("category = $%d", argN))
		args = append(args, *req.Category)
		argN++
	}
	if req.Recommendation != nil {
		sets = append(sets, fmt.Sprintf("recommendation = $%d", argN))
		args = append(args, *req.Recommendation)
		argN++
	}
	if req.ReferenceNumber != nil {
		sets = append(sets, fmt.Sprintf("reference_number = $%d", argN))
		args = append(args, *req.ReferenceNumber)
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
	args = append(args, findingID, auditID, orgID)

	query := fmt.Sprintf("UPDATE audit_findings SET %s WHERE id = $%d AND audit_id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1, argN+2)
	database.DB.Exec(query, args...)

	middleware.LogAudit(c, "audit_finding.updated", "audit_finding", &findingID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{"id": findingID, "updated_at": now}))
}

// ChangeFindingStatus transitions a finding's status through the remediation lifecycle.
func ChangeFindingStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	findingID := c.Param("fid")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.ChangeFindingStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	var currentStatus string
	err := database.DB.QueryRow("SELECT status FROM audit_findings WHERE id = $1 AND audit_id = $2 AND org_id = $3",
		findingID, auditID, orgID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_FINDING_NOT_FOUND", "Finding not found"))
		return
	}

	if !models.IsValidFindingStatusTransition(currentStatus, req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_INVALID_TRANSITION",
			fmt.Sprintf("Invalid status transition from '%s' to '%s'", currentStatus, req.Status)))
		return
	}

	now := time.Now()
	sets := []string{"status = $1", "updated_at = $2"}
	args := []interface{}{req.Status, now}
	argN := 3

	switch req.Status {
	case models.FindingStatusRemediationPlanned:
		if req.RemediationPlan == nil || *req.RemediationPlan == "" {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Remediation plan is required"))
			return
		}
		sets = append(sets, fmt.Sprintf("remediation_plan = $%d", argN))
		args = append(args, *req.RemediationPlan)
		argN++
		if req.RemediationDueDate != nil {
			sets = append(sets, fmt.Sprintf("remediation_due_date = $%d::date", argN))
			args = append(args, *req.RemediationDueDate)
			argN++
		}
		if req.RemediationOwnerID != nil {
			sets = append(sets, fmt.Sprintf("remediation_owner_id = $%d", argN))
			args = append(args, *req.RemediationOwnerID)
			argN++
		}

	case models.FindingStatusRemediationInProgress:
		if currentStatus == models.FindingStatusRemediationPlanned {
			sets = append(sets, fmt.Sprintf("remediation_started_at = $%d", argN))
			args = append(args, now)
			argN++
		}
		// Reopen from remediation_complete by auditor requires notes
		if currentStatus == models.FindingStatusRemediationComplete {
			if req.Notes == nil || *req.Notes == "" {
				c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Notes required when reopening a finding"))
				return
			}
		}

	case models.FindingStatusRemediationComplete:
		sets = append(sets, fmt.Sprintf("remediation_completed_at = $%d", argN))
		args = append(args, now)
		argN++
		if req.ManagementResponse != nil {
			sets = append(sets, fmt.Sprintf("management_response = $%d", argN))
			args = append(args, *req.ManagementResponse)
			argN++
		}

	case models.FindingStatusVerified:
		sets = append(sets, fmt.Sprintf("verified_at = $%d", argN))
		args = append(args, now)
		argN++
		sets = append(sets, fmt.Sprintf("verified_by = $%d", argN))
		args = append(args, userID)
		argN++
		if req.VerificationNotes != nil {
			sets = append(sets, fmt.Sprintf("verification_notes = $%d", argN))
			args = append(args, *req.VerificationNotes)
			argN++
		}

	case models.FindingStatusRiskAccepted:
		// Only CISO can accept risk
		if userRole != models.RoleCISO {
			c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Only CISO can accept risk on findings"))
			return
		}
		if req.RiskAcceptReason == nil || *req.RiskAcceptReason == "" {
			c.JSON(http.StatusBadRequest, errorResponse("AUDIT_RISK_ACCEPT_REQUIRES_REASON", "Risk acceptance requires a justification"))
			return
		}
		sets = append(sets, "risk_accepted = TRUE")
		sets = append(sets, fmt.Sprintf("risk_acceptance_reason = $%d", argN))
		args = append(args, *req.RiskAcceptReason)
		argN++
		sets = append(sets, fmt.Sprintf("risk_accepted_by = $%d", argN))
		args = append(args, userID)
		argN++
		sets = append(sets, fmt.Sprintf("risk_accepted_at = $%d", argN))
		args = append(args, now)
		argN++
	}

	args = append(args, findingID, auditID, orgID)
	query := fmt.Sprintf("UPDATE audit_findings SET %s WHERE id = $%d AND audit_id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1, argN+2)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to change finding status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change finding status"))
		return
	}

	updateAuditFindingCounts(auditID, orgID)
	middleware.LogAudit(c, "audit_finding.status_changed", "audit_finding", &findingID, map[string]interface{}{
		"old_status": currentStatus, "new_status": req.Status,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": findingID, "status": req.Status, "updated_at": now,
	}))
}

// SubmitManagementResponse submits the org's formal management response.
func SubmitManagementResponse(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	findingID := c.Param("fid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var req models.ManagementResponseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	var exists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_findings WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
		findingID, auditID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_FINDING_NOT_FOUND", "Finding not found"))
		return
	}

	now := time.Now()
	database.DB.Exec(
		"UPDATE audit_findings SET management_response = $1, updated_at = $2 WHERE id = $3 AND org_id = $4",
		req.ManagementResponse, now, findingID, orgID,
	)

	middleware.LogAudit(c, "audit_finding.updated", "audit_finding", &findingID, map[string]interface{}{
		"field": "management_response",
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": findingID, "management_response": req.ManagementResponse, "updated_at": now,
	}))
}
