package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// ListEvidenceEvaluations lists all evaluations for an evidence artifact.
func ListEvidenceEvaluations(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	artifactID := c.Param("id")

	// Verify artifact exists
	var exists bool
	database.QueryRow("SELECT EXISTS(SELECT 1 FROM evidence_artifacts WHERE id = $1 AND org_id = $2)",
		artifactID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	var total int
	database.QueryRow("SELECT COUNT(*) FROM evidence_evaluations WHERE artifact_id = $1 AND org_id = $2",
		artifactID, orgID).Scan(&total)

	offset := (page - 1) * perPage
	rows, err := database.Query(`
		SELECT ee.id, ee.verdict, ee.confidence, ee.comments,
			   ee.missing_elements, ee.remediation_notes,
			   ee.evidence_link_id,
			   ee.evaluated_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   COALESCE(u.role, ''),
			   ee.created_at
		FROM evidence_evaluations ee
		LEFT JOIN users u ON u.id = ee.evaluated_by
		WHERE ee.artifact_id = $1 AND ee.org_id = $2
		ORDER BY ee.created_at DESC
		LIMIT $3 OFFSET $4
	`, artifactID, orgID, perPage, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list evaluations")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	evaluations := []gin.H{}
	for rows.Next() {
		var eID, eVerdict, eConfidence, eComments string
		var eMissing pq.StringArray
		var eRemediation *string
		var eLinkID *string
		var eEvaluatedBy, eEvaluatorName, eEvaluatorRole string
		var eCreatedAt time.Time

		if err := rows.Scan(&eID, &eVerdict, &eConfidence, &eComments,
			&eMissing, &eRemediation, &eLinkID,
			&eEvaluatedBy, &eEvaluatorName, &eEvaluatorRole,
			&eCreatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan evaluation row")
			continue
		}

		eval := gin.H{
			"id":                eID,
			"verdict":           eVerdict,
			"confidence":        eConfidence,
			"comments":          eComments,
			"missing_elements":  []string(eMissing),
			"remediation_notes": eRemediation,
			"evaluated_by": gin.H{
				"id":   eEvaluatedBy,
				"name": eEvaluatorName,
				"role": eEvaluatorRole,
			},
			"created_at": eCreatedAt,
		}

		// Get evidence link details if present
		if eLinkID != nil {
			var lTargetType string
			var lControlID *string
			database.QueryRow("SELECT target_type, control_id FROM evidence_links WHERE id = $1", *eLinkID).
				Scan(&lTargetType, &lControlID)
			linkInfo := gin.H{"id": *eLinkID, "target_type": lTargetType}
			if lControlID != nil {
				var cIdentifier string
				database.QueryRow("SELECT identifier FROM controls WHERE id = $1", *lControlID).Scan(&cIdentifier)
				linkInfo["control_identifier"] = cIdentifier
			}
			eval["evidence_link"] = linkInfo
		} else {
			eval["evidence_link"] = nil
		}

		evaluations = append(evaluations, eval)
	}

	c.JSON(http.StatusOK, listResponse(c, evaluations, total, page, perPage))
}

// CreateEvidenceEvaluation submits an evaluation for an evidence artifact.
func CreateEvidenceEvaluation(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	artifactID := c.Param("id")

	// Verify artifact exists
	var artifactStatus string
	err := database.QueryRow("SELECT status FROM evidence_artifacts WHERE id = $1 AND org_id = $2",
		artifactID, orgID).Scan(&artifactStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	var req models.CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate
	if !models.IsValidEvalVerdict(req.Verdict) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid verdict: must be sufficient, partial, insufficient, or needs_update"))
		return
	}
	confidence := "medium"
	if req.Confidence != nil {
		if !models.IsValidConfidenceLevel(*req.Confidence) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid confidence: must be high, medium, or low"))
			return
		}
		confidence = *req.Confidence
	}
	if len(req.Comments) > 5000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Comments must be at most 5000 characters"))
		return
	}
	if len(req.MissingElements) > 20 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Maximum 20 missing elements"))
		return
	}
	for _, elem := range req.MissingElements {
		if len(elem) > 200 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Each missing element must be at most 200 characters"))
			return
		}
	}
	if req.RemediationNotes != nil && len(*req.RemediationNotes) > 5000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Remediation notes must be at most 5000 characters"))
		return
	}

	// Validate evidence_link_id if provided
	if req.EvidenceLinkID != nil {
		var linkExists bool
		database.QueryRow("SELECT EXISTS(SELECT 1 FROM evidence_links WHERE id = $1 AND artifact_id = $2 AND org_id = $3)",
			*req.EvidenceLinkID, artifactID, orgID).Scan(&linkExists)
		if !linkExists {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "evidence_link_id does not belong to this artifact"))
			return
		}
	}

	evalID := uuid.New().String()

	missingElems := req.MissingElements
	if missingElems == nil {
		missingElems = []string{}
	}

	_, err = database.Exec(`
		INSERT INTO evidence_evaluations (id, org_id, artifact_id, evidence_link_id,
			verdict, confidence, comments, missing_elements, remediation_notes, evaluated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, evalID, orgID, artifactID, req.EvidenceLinkID,
		req.Verdict, confidence, req.Comments, pq.Array(missingElems),
		req.RemediationNotes, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create evaluation")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Auto-status change based on verdict
	if artifactStatus == "pending_review" {
		var newStatus string
		if req.Verdict == "sufficient" {
			newStatus = "approved"
		} else if req.Verdict == "insufficient" {
			newStatus = "rejected"
		}
		if newStatus != "" {
			database.Exec("UPDATE evidence_artifacts SET status = $1 WHERE id = $2", newStatus, artifactID)
			middleware.LogAudit(c, "evidence.status_changed", "evidence", &artifactID, map[string]interface{}{
				"old_status": artifactStatus, "new_status": newStatus, "trigger": "evaluation",
			})
		}
	}

	middleware.LogAudit(c, "evidence.evaluated", "evidence", &artifactID, map[string]interface{}{
		"verdict": req.Verdict, "confidence": confidence, "link_id": req.EvidenceLinkID,
	})

	// Get evaluator name
	var evaluatorName string
	database.QueryRow("SELECT first_name || ' ' || last_name FROM users WHERE id = $1", userID).Scan(&evaluatorName)

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":                evalID,
		"artifact_id":       artifactID,
		"evidence_link_id":  req.EvidenceLinkID,
		"verdict":           req.Verdict,
		"confidence":        confidence,
		"comments":          req.Comments,
		"missing_elements":  missingElems,
		"remediation_notes": req.RemediationNotes,
		"evaluated_by": gin.H{
			"id":   userID,
			"name": evaluatorName,
		},
		"created_at": time.Now(),
	}))
}
