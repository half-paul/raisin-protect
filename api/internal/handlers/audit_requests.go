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

// updateAuditRequestCounts recalculates denormalized request counts on the audit.
func updateAuditRequestCounts(auditID, orgID string) {
	database.DB.Exec(`
		UPDATE audits SET
			total_requests = (SELECT COUNT(*) FROM audit_requests WHERE audit_id = $1 AND org_id = $2),
			open_requests = (SELECT COUNT(*) FROM audit_requests WHERE audit_id = $1 AND org_id = $2 AND status NOT IN ('accepted', 'closed'))
		WHERE id = $1 AND org_id = $2
	`, auditID, orgID)
}

// ListAuditRequests lists evidence requests for an audit.
func ListAuditRequests(c *gin.Context) {
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

	where := []string{"ar.audit_id = $1", "ar.org_id = $2"}
	args := []interface{}{auditID, orgID}
	argN := 3

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("ar.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("priority"); v != "" {
		where = append(where, fmt.Sprintf("ar.priority = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("assigned_to"); v != "" {
		where = append(where, fmt.Sprintf("ar.assigned_to = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("control_id"); v != "" {
		where = append(where, fmt.Sprintf("ar.control_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("requirement_id"); v != "" {
		where = append(where, fmt.Sprintf("ar.requirement_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("overdue"); v == "true" {
		where = append(where, "ar.due_date < CURRENT_DATE AND ar.status NOT IN ('accepted', 'closed')")
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(ar.title ILIKE $%d OR ar.description ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	sortCol := "ar.due_date"
	switch c.DefaultQuery("sort", "due_date") {
	case "created_at":
		sortCol = "ar.created_at"
	case "priority":
		sortCol = "ar.priority"
	case "status":
		sortCol = "ar.status"
	}
	order := "ASC"
	if strings.ToLower(c.DefaultQuery("order", "asc")) == "desc" {
		order = "DESC"
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM audit_requests ar WHERE %s", whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ar.id, ar.audit_id, ar.title, ar.description, ar.priority, ar.status,
		       ar.control_id, COALESCE(ctrl.title, ''),
		       ar.requirement_id, COALESCE(req.title, ''),
		       ar.requested_by, COALESCE(rb.first_name || ' ' || rb.last_name, ''),
		       ar.assigned_to, COALESCE(at.first_name || ' ' || at.last_name, ''),
		       ar.due_date, ar.submitted_at, ar.reviewed_at, ar.reviewer_notes,
		       ar.reference_number,
		       (SELECT COUNT(*) FROM audit_evidence_links ael WHERE ael.request_id = ar.id),
		       ar.tags, ar.created_at, ar.updated_at
		FROM audit_requests ar
		LEFT JOIN controls ctrl ON ar.control_id = ctrl.id
		LEFT JOIN requirements req ON ar.requirement_id = req.id
		LEFT JOIN users rb ON ar.requested_by = rb.id
		LEFT JOIN users at ON ar.assigned_to = at.id
		WHERE %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list audit requests")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list requests"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, aID, title, description, priority, status string
			controlID, controlTitle                        *string
			requirementID, requirementTitle                 *string
			requestedBy, requestedByName                   *string
			assignedTo, assignedToName                     *string
			dueDate                                        *time.Time
			submittedAt, reviewedAt                         *time.Time
			reviewerNotes, referenceNumber                  *string
			evidenceCount                                   int
			tags                                            pq.StringArray
			createdAt, updatedAt                            time.Time
		)
		if err := rows.Scan(
			&id, &aID, &title, &description, &priority, &status,
			&controlID, &controlTitle,
			&requirementID, &requirementTitle,
			&requestedBy, &requestedByName,
			&assignedTo, &assignedToName,
			&dueDate, &submittedAt, &reviewedAt, &reviewerNotes,
			&referenceNumber, &evidenceCount,
			&tags, &createdAt, &updatedAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan audit request row")
			continue
		}

		results = append(results, gin.H{
			"id": id, "audit_id": aID,
			"title": title, "description": description,
			"priority": priority, "status": status,
			"control_id": controlID, "control_title": controlTitle,
			"requirement_id": requirementID, "requirement_title": requirementTitle,
			"requested_by": requestedBy, "requested_by_name": requestedByName,
			"assigned_to": assignedTo, "assigned_to_name": assignedToName,
			"due_date": dueDate, "submitted_at": submittedAt,
			"reviewed_at": reviewedAt, "reviewer_notes": reviewerNotes,
			"reference_number": referenceNumber, "evidence_count": evidenceCount,
			"tags": []string(tags), "created_at": createdAt, "updated_at": updatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       results,
		"pagination": gin.H{"page": page, "per_page": perPage, "total": total, "total_pages": (total + perPage - 1) / perPage},
	})
}

// GetAuditRequest gets a single request with linked evidence.
func GetAuditRequest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	requestID := c.Param("rid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var (
		id, aID, title, description, priority, status string
		controlID, controlTitle                        *string
		requirementID, requirementTitle                 *string
		requestedBy, requestedByName                   *string
		assignedTo, assignedToName                     *string
		dueDate                                        *time.Time
		submittedAt, reviewedAt                         *time.Time
		reviewerNotes, referenceNumber                  *string
		tags                                            pq.StringArray
		createdAt, updatedAt                            time.Time
	)

	err := database.DB.QueryRow(`
		SELECT ar.id, ar.audit_id, ar.title, ar.description, ar.priority, ar.status,
		       ar.control_id, COALESCE(ctrl.title, ''),
		       ar.requirement_id, COALESCE(req.title, ''),
		       ar.requested_by, COALESCE(rb.first_name || ' ' || rb.last_name, ''),
		       ar.assigned_to, COALESCE(at.first_name || ' ' || at.last_name, ''),
		       ar.due_date, ar.submitted_at, ar.reviewed_at, ar.reviewer_notes,
		       ar.reference_number, ar.tags, ar.created_at, ar.updated_at
		FROM audit_requests ar
		LEFT JOIN controls ctrl ON ar.control_id = ctrl.id
		LEFT JOIN requirements req ON ar.requirement_id = req.id
		LEFT JOIN users rb ON ar.requested_by = rb.id
		LEFT JOIN users at ON ar.assigned_to = at.id
		WHERE ar.id = $1 AND ar.audit_id = $2 AND ar.org_id = $3
	`, requestID, auditID, orgID).Scan(
		&id, &aID, &title, &description, &priority, &status,
		&controlID, &controlTitle,
		&requirementID, &requirementTitle,
		&requestedBy, &requestedByName,
		&assignedTo, &assignedToName,
		&dueDate, &submittedAt, &reviewedAt, &reviewerNotes,
		&referenceNumber, &tags, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_REQUEST_NOT_FOUND", "Request not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get audit request")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get request"))
		return
	}

	// Fetch linked evidence
	evidence := []gin.H{}
	eRows, err := database.DB.Query(`
		SELECT ael.id, ael.artifact_id, ea.title, ea.file_name, ea.file_size, ea.mime_type,
		       ael.submitted_by, COALESCE(sub.first_name || ' ' || sub.last_name, ''),
		       ael.submitted_at, ael.submission_notes, ael.status,
		       ael.reviewed_by, COALESCE(rev.first_name || ' ' || rev.last_name, ''),
		       ael.reviewed_at, ael.review_notes
		FROM audit_evidence_links ael
		JOIN evidence_artifacts ea ON ael.artifact_id = ea.id
		LEFT JOIN users sub ON ael.submitted_by = sub.id
		LEFT JOIN users rev ON ael.reviewed_by = rev.id
		WHERE ael.request_id = $1 AND ael.org_id = $2
		ORDER BY ael.submitted_at
	`, requestID, orgID)
	if err == nil {
		defer eRows.Close()
		for eRows.Next() {
			var (
				linkID, artifactID, artifactTitle string
				fileName                          *string
				fileSize                          *int64
				mimeType                          *string
				subBy, subByName                  *string
				subAt                             time.Time
				subNotes                          *string
				linkStatus                        string
				revBy, revByName                  *string
				revAt                             *time.Time
				revNotes                          *string
			)
			eRows.Scan(&linkID, &artifactID, &artifactTitle, &fileName, &fileSize, &mimeType,
				&subBy, &subByName, &subAt, &subNotes, &linkStatus,
				&revBy, &revByName, &revAt, &revNotes)
			evidence = append(evidence, gin.H{
				"link_id": linkID, "artifact_id": artifactID, "artifact_title": artifactTitle,
				"file_name": fileName, "file_size": fileSize, "mime_type": mimeType,
				"submitted_by": subBy, "submitted_by_name": subByName,
				"submitted_at": subAt, "submission_notes": subNotes,
				"status": linkStatus,
				"reviewed_by": revBy, "reviewed_by_name": revByName,
				"reviewed_at": revAt, "review_notes": revNotes,
			})
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": id, "audit_id": aID,
		"title": title, "description": description,
		"priority": priority, "status": status,
		"control_id": controlID, "control_title": controlTitle,
		"requirement_id": requirementID, "requirement_title": requirementTitle,
		"requested_by": requestedBy, "requested_by_name": requestedByName,
		"assigned_to": assignedTo, "assigned_to_name": assignedToName,
		"due_date": dueDate, "submitted_at": submittedAt,
		"reviewed_at": reviewedAt, "reviewer_notes": reviewerNotes,
		"reference_number": referenceNumber,
		"evidence": evidence,
		"tags": []string(tags), "created_at": createdAt, "updated_at": updatedAt,
	}))
}

// CreateAuditRequest creates a new evidence request.
func CreateAuditRequest(c *gin.Context) {
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

	var req models.CreateAuditRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	if utf8.RuneCountInString(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must not exceed 500 characters"))
		return
	}

	priority := "medium"
	if req.Priority != nil && models.IsValidAuditRequestPriority(*req.Priority) {
		priority = *req.Priority
	}

	// If assigned, start as in_progress
	status := models.AuditRequestStatusOpen
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		status = models.AuditRequestStatusInProgress
	}

	reqID := uuid.New().String()
	now := time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO audit_requests (id, org_id, audit_id, title, description, priority, status,
		                            control_id, requirement_id, requested_by, assigned_to,
		                            due_date, reference_number, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::date, $13, $14, $15, $15)
	`, reqID, orgID, auditID, req.Title, req.Description, priority, status,
		req.ControlID, req.RequirementID, userID, req.AssignedTo,
		req.DueDate, req.ReferenceNumber, pq.Array(req.Tags), now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create audit request")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create request"))
		return
	}

	updateAuditRequestCounts(auditID, orgID)
	middleware.LogAudit(c, "audit_request.created", "audit_request", &reqID, map[string]interface{}{
		"audit_id": auditID, "title": req.Title,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id": reqID, "audit_id": auditID,
		"title": req.Title, "priority": priority, "status": status,
		"created_at": now,
	}))
}

// UpdateAuditRequest updates request metadata.
func UpdateAuditRequest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	requestID := c.Param("rid")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.UpdateAuditRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Verify request exists
	var exists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_requests WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
		requestID, auditID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_REQUEST_NOT_FOUND", "Request not found"))
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
	if req.Priority != nil {
		if !models.IsValidAuditRequestPriority(*req.Priority) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid priority"))
			return
		}
		sets = append(sets, fmt.Sprintf("priority = $%d", argN))
		args = append(args, *req.Priority)
		argN++
	}
	if req.DueDate != nil {
		sets = append(sets, fmt.Sprintf("due_date = $%d::date", argN))
		args = append(args, *req.DueDate)
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
	args = append(args, requestID, auditID, orgID)

	query := fmt.Sprintf("UPDATE audit_requests SET %s WHERE id = $%d AND audit_id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1, argN+2)
	database.DB.Exec(query, args...)

	c.JSON(http.StatusOK, successResponse(c, gin.H{"id": requestID, "updated_at": now}))
}

// AssignAuditRequest assigns or reassigns a request.
func AssignAuditRequest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	requestID := c.Param("rid")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.AssignAuditRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	var currentStatus string
	err := database.DB.QueryRow("SELECT status FROM audit_requests WHERE id = $1 AND audit_id = $2 AND org_id = $3",
		requestID, auditID, orgID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_REQUEST_NOT_FOUND", "Request not found"))
		return
	}

	now := time.Now()
	newStatus := currentStatus
	if currentStatus == models.AuditRequestStatusOpen {
		newStatus = models.AuditRequestStatusInProgress
	}

	database.DB.Exec(
		"UPDATE audit_requests SET assigned_to = $1, status = $2, updated_at = $3 WHERE id = $4 AND org_id = $5",
		req.AssignedTo, newStatus, now, requestID, orgID,
	)

	middleware.LogAudit(c, "audit_request.assigned", "audit_request", &requestID, map[string]interface{}{
		"assigned_to": req.AssignedTo,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": requestID, "status": newStatus, "assigned_to": req.AssignedTo, "updated_at": now,
	}))
}

// SubmitAuditRequest marks a request as submitted.
func SubmitAuditRequest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	requestID := c.Param("rid")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.SubmitAuditRequestReq
	c.ShouldBindJSON(&req)

	var currentStatus string
	err := database.DB.QueryRow("SELECT status FROM audit_requests WHERE id = $1 AND audit_id = $2 AND org_id = $3",
		requestID, auditID, orgID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_REQUEST_NOT_FOUND", "Request not found"))
		return
	}

	if currentStatus != models.AuditRequestStatusOpen && currentStatus != models.AuditRequestStatusInProgress && currentStatus != models.AuditRequestStatusRejected {
		c.JSON(http.StatusConflict, errorResponse("AUDIT_INVALID_TRANSITION",
			fmt.Sprintf("Cannot submit request in '%s' status", currentStatus)))
		return
	}

	// Check evidence exists
	var evidenceCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM audit_evidence_links WHERE request_id = $1 AND org_id = $2",
		requestID, orgID).Scan(&evidenceCount)
	if evidenceCount == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_NO_EVIDENCE", "Cannot submit request with no evidence attached"))
		return
	}

	now := time.Now()
	database.DB.Exec(
		"UPDATE audit_requests SET status = 'submitted', submitted_at = $1, updated_at = $1 WHERE id = $2 AND org_id = $3",
		now, requestID, orgID,
	)

	middleware.LogAudit(c, "audit_request.submitted", "audit_request", &requestID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": requestID, "status": "submitted", "submitted_at": now, "updated_at": now,
	}))
}

// ReviewAuditRequest lets an auditor accept or reject a request.
func ReviewAuditRequest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	requestID := c.Param("rid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var req models.ReviewAuditRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if req.Decision != "accepted" && req.Decision != "rejected" {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Decision must be 'accepted' or 'rejected'"))
		return
	}
	if req.Decision == "rejected" && (req.Notes == nil || *req.Notes == "") {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_REJECTION_REQUIRES_NOTES", "Rejection must include notes"))
		return
	}

	var currentStatus string
	err := database.DB.QueryRow("SELECT status FROM audit_requests WHERE id = $1 AND audit_id = $2 AND org_id = $3",
		requestID, auditID, orgID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_REQUEST_NOT_FOUND", "Request not found"))
		return
	}

	if currentStatus != models.AuditRequestStatusSubmitted {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_INVALID_TRANSITION", "Request must be in 'submitted' status to review"))
		return
	}

	now := time.Now()
	database.DB.Exec(
		"UPDATE audit_requests SET status = $1, reviewed_at = $2, reviewer_notes = $3, updated_at = $2 WHERE id = $4 AND org_id = $5",
		req.Decision, now, req.Notes, requestID, orgID,
	)

	updateAuditRequestCounts(auditID, orgID)

	action := "audit_request.accepted"
	if req.Decision == "rejected" {
		action = "audit_request.rejected"
	}
	middleware.LogAudit(c, action, "audit_request", &requestID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": requestID, "status": req.Decision, "reviewed_at": now, "updated_at": now,
	}))
}

// CloseAuditRequest closes a request.
func CloseAuditRequest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	requestID := c.Param("rid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var req models.CloseAuditRequestReq
	c.ShouldBindJSON(&req)

	var exists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_requests WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
		requestID, auditID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_REQUEST_NOT_FOUND", "Request not found"))
		return
	}

	now := time.Now()
	database.DB.Exec(
		"UPDATE audit_requests SET status = 'closed', updated_at = $1 WHERE id = $2 AND org_id = $3",
		now, requestID, orgID,
	)

	updateAuditRequestCounts(auditID, orgID)
	middleware.LogAudit(c, "audit_request.closed", "audit_request", &requestID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": requestID, "status": "closed", "updated_at": now,
	}))
}

// BulkCreateAuditRequests creates multiple requests at once.
func BulkCreateAuditRequests(c *gin.Context) {
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

	var req models.BulkCreateRequestsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if len(req.Requests) == 0 || len(req.Requests) > 100 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Must provide 1-100 requests"))
		return
	}

	now := time.Now()
	created := []gin.H{}

	for _, r := range req.Requests {
		priority := "medium"
		if r.Priority != nil && models.IsValidAuditRequestPriority(*r.Priority) {
			priority = *r.Priority
		}
		status := models.AuditRequestStatusOpen
		if r.AssignedTo != nil && *r.AssignedTo != "" {
			status = models.AuditRequestStatusInProgress
		}

		reqID := uuid.New().String()
		_, err := database.DB.Exec(`
			INSERT INTO audit_requests (id, org_id, audit_id, title, description, priority, status,
			                            control_id, requirement_id, requested_by, assigned_to,
			                            due_date, reference_number, tags, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::date, $13, $14, $15, $15)
		`, reqID, orgID, auditID, r.Title, r.Description, priority, status,
			r.ControlID, r.RequirementID, userID, r.AssignedTo,
			r.DueDate, r.ReferenceNumber, pq.Array(r.Tags), now)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create bulk request item")
			continue
		}
		created = append(created, gin.H{"id": reqID, "title": r.Title, "priority": priority, "status": status})
	}

	updateAuditRequestCounts(auditID, orgID)

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"created":  len(created),
		"requests": created,
	}))
}
