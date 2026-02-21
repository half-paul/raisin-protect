package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListRequestEvidence lists evidence submitted for a specific request.
func ListRequestEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	requestID := c.Param("rid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
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

	rows, err := database.DB.Query(`
		SELECT ael.id, ael.artifact_id, ea.title, ea.file_name, ea.file_size, ea.mime_type,
		       ea.evidence_type, ea.status AS evidence_status,
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
	if err != nil {
		log.Error().Err(err).Msg("Failed to list request evidence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list evidence"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			linkID, artifactID, artifactTitle string
			fileName                          *string
			fileSize                          *int64
			mimeType, evidenceType, evStatus  *string
			subBy, subByName                  *string
			subAt                             time.Time
			subNotes                          *string
			linkStatus                        string
			revBy, revByName                  *string
			revAt                             *time.Time
			revNotes                          *string
		)
		if err := rows.Scan(
			&linkID, &artifactID, &artifactTitle, &fileName, &fileSize, &mimeType,
			&evidenceType, &evStatus,
			&subBy, &subByName, &subAt, &subNotes, &linkStatus,
			&revBy, &revByName, &revAt, &revNotes,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan evidence row")
			continue
		}
		results = append(results, gin.H{
			"link_id": linkID, "artifact_id": artifactID, "artifact_title": artifactTitle,
			"file_name": fileName, "file_size": fileSize, "mime_type": mimeType,
			"evidence_type": evidenceType, "evidence_status": evStatus,
			"submitted_by": subBy, "submitted_by_name": subByName,
			"submitted_at": subAt, "submission_notes": subNotes,
			"status": linkStatus,
			"reviewed_by": revBy, "reviewed_by_name": revByName,
			"reviewed_at": revAt, "review_notes": revNotes,
		})
	}

	c.JSON(http.StatusOK, successResponse(c, results))
}

// SubmitRequestEvidence links an evidence artifact to a request.
func SubmitRequestEvidence(c *gin.Context) {
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

	var req models.SubmitEvidenceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Verify request exists
	var reqExists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_requests WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
		requestID, auditID, orgID).Scan(&reqExists)
	if !reqExists {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_REQUEST_NOT_FOUND", "Request not found"))
		return
	}

	// Verify artifact exists in same org
	var artifactExists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM evidence_artifacts WHERE id = $1 AND org_id = $2)",
		req.ArtifactID, orgID).Scan(&artifactExists)
	if !artifactExists {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Evidence artifact not found in this organization"))
		return
	}

	// Check for duplicate
	var duplicate bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_evidence_links WHERE request_id = $1 AND artifact_id = $2)",
		requestID, req.ArtifactID).Scan(&duplicate)
	if duplicate {
		c.JSON(http.StatusConflict, errorResponse("AUDIT_DUPLICATE_EVIDENCE", "Evidence artifact already linked to this request"))
		return
	}

	linkID := uuid.New().String()
	now := time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO audit_evidence_links (id, org_id, audit_id, request_id, artifact_id,
		                                  submitted_by, submitted_at, submission_notes, status,
		                                  created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending_review', $7, $7)
	`, linkID, orgID, auditID, requestID, req.ArtifactID, userID, now, req.Notes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to submit evidence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to submit evidence"))
		return
	}

	middleware.LogAudit(c, "audit_evidence.submitted", "audit_evidence_link", &linkID, map[string]interface{}{
		"request_id": requestID, "artifact_id": req.ArtifactID,
	})

	// Get artifact title for response
	var artifactTitle string
	database.DB.QueryRow("SELECT title FROM evidence_artifacts WHERE id = $1", req.ArtifactID).Scan(&artifactTitle)

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"link_id":          linkID,
		"artifact_id":      req.ArtifactID,
		"artifact_title":   artifactTitle,
		"submitted_by":     userID,
		"submitted_at":     now,
		"submission_notes": req.Notes,
		"status":           "pending_review",
	}))
}

// ReviewRequestEvidence lets an auditor review a specific evidence submission.
func ReviewRequestEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	linkID := c.Param("lid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var req models.ReviewEvidenceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if !models.IsValidEvidenceLinkStatus(req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Status must be 'accepted', 'rejected', or 'needs_clarification'"))
		return
	}
	if (req.Status == "rejected" || req.Status == "needs_clarification") && (req.Notes == nil || *req.Notes == "") {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_REJECTION_REQUIRES_NOTES", "Notes required for rejection or clarification"))
		return
	}

	var exists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_evidence_links WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
		linkID, auditID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence link not found"))
		return
	}

	now := time.Now()
	_, err := database.DB.Exec(`
		UPDATE audit_evidence_links SET status = $1, reviewed_by = $2, reviewed_at = $3, review_notes = $4, updated_at = $3
		WHERE id = $5 AND org_id = $6
	`, req.Status, userID, now, req.Notes, linkID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to review evidence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to review evidence"))
		return
	}

	middleware.LogAudit(c, "audit_evidence.reviewed", "audit_evidence_link", &linkID, map[string]interface{}{
		"status": req.Status,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"link_id": linkID, "status": req.Status, "reviewed_at": now, "updated_at": now,
	}))
}

// RemoveRequestEvidence removes an evidence submission.
func RemoveRequestEvidence(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	linkID := c.Param("lid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	// Check link exists and get submitter
	var submittedBy *string
	err := database.DB.QueryRow(
		"SELECT submitted_by FROM audit_evidence_links WHERE id = $1 AND audit_id = $2 AND org_id = $3",
		linkID, auditID, orgID,
	).Scan(&submittedBy)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence link not found"))
		return
	}

	// Auth: compliance_manager, ciso, or the submitter
	isSubmitter := submittedBy != nil && *submittedBy == userID
	if !isSubmitter && !models.HasRole(userRole, models.AuditCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to remove this evidence"))
		return
	}

	database.DB.Exec("DELETE FROM audit_evidence_links WHERE id = $1 AND org_id = $2", linkID, orgID)

	c.Status(http.StatusNoContent)
}
