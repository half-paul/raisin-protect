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

// ListRiskControls lists controls linked to a risk.
func ListRiskControls(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	riskID := c.Param("id")

	// Verify risk exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM risks WHERE id = $1 AND org_id = $2)", riskID, orgID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}

	rows, err := database.DB.Query(`
		SELECT rc.id, c.id, c.identifier, c.title, c.description, c.category, c.status,
		       rc.effectiveness, rc.mitigation_percentage, rc.notes,
		       rc.last_effectiveness_review, rc.reviewed_by,
		       rc.linked_by, rc.created_at,
		       COALESCE(u_linked.first_name || ' ' || u_linked.last_name, '') AS linked_by_name,
		       COALESCE(u_review.first_name || ' ' || u_review.last_name, '') AS reviewed_by_name
		FROM risk_controls rc
		JOIN controls c ON rc.control_id = c.id
		LEFT JOIN users u_linked ON rc.linked_by = u_linked.id
		LEFT JOIN users u_review ON rc.reviewed_by = u_review.id
		WHERE rc.risk_id = $1 AND rc.org_id = $2
		ORDER BY rc.created_at ASC
	`, riskID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list risk controls")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list controls"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			rcID, ctrlID, ctrlIdentifier, ctrlTitle string
			ctrlDesc                                *string
			ctrlCat, ctrlStatus                     string
			effectiveness                           string
			mitPct                                  *int
			notes                                   *string
			lastReview                              *time.Time
			reviewedBy                              *string
			linkedBy                                string
			createdAt                               time.Time
			linkedByName, reviewedByName            string
		)

		err := rows.Scan(
			&rcID, &ctrlID, &ctrlIdentifier, &ctrlTitle, &ctrlDesc, &ctrlCat, &ctrlStatus,
			&effectiveness, &mitPct, &notes,
			&lastReview, &reviewedBy,
			&linkedBy, &createdAt,
			&linkedByName, &reviewedByName,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan risk-control row")
			continue
		}

		var reviewedByObj interface{}
		if reviewedBy != nil {
			reviewedByObj = gin.H{"id": *reviewedBy, "name": reviewedByName}
		}

		result := gin.H{
			"id":                        ctrlID,
			"risk_control_id":           rcID,
			"identifier":                ctrlIdentifier,
			"title":                     ctrlTitle,
			"description":               ctrlDesc,
			"category":                  ctrlCat,
			"status":                    ctrlStatus,
			"effectiveness":             effectiveness,
			"mitigation_percentage":     mitPct,
			"notes":                     notes,
			"last_effectiveness_review": lastReview,
			"reviewed_by":              reviewedByObj,
			"linked_by":                gin.H{"id": linkedBy, "name": linkedByName},
			"linked_at":                createdAt,
		}
		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{
			"total": len(results),
		},
	})
}

// LinkRiskControl links a control to a risk.
func LinkRiskControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")

	var req models.LinkRiskControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Check risk exists and get owner
	var riskOwnerID *string
	err := database.DB.QueryRow("SELECT owner_id FROM risks WHERE id = $1 AND org_id = $2", riskID, orgID).Scan(&riskOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to link control"))
		return
	}

	// Authorization
	isOwner := riskOwnerID != nil && *riskOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.RiskControlLinkRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to link controls"))
		return
	}

	// Check control exists and belongs to same org
	var ctrlExists bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM controls WHERE id = $1 AND org_id = $2)", req.ControlID, orgID).Scan(&ctrlExists)
	if !ctrlExists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	// Check not already linked
	var alreadyLinked bool
	database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM risk_controls WHERE risk_id = $1 AND control_id = $2 AND org_id = $3)",
		riskID, req.ControlID, orgID).Scan(&alreadyLinked)
	if alreadyLinked {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Control is already linked to this risk"))
		return
	}

	// Validate effectiveness
	effectiveness := models.EffectivenessNotAssessed
	if req.Effectiveness != nil {
		if !models.IsValidEffectiveness(*req.Effectiveness) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid effectiveness value"))
			return
		}
		effectiveness = *req.Effectiveness
	}

	// Validate mitigation percentage
	if req.MitigationPercentage != nil && (*req.MitigationPercentage < 0 || *req.MitigationPercentage > 100) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Mitigation percentage must be between 0 and 100"))
		return
	}

	rcID := uuid.New().String()
	now := time.Now()

	_, err = database.DB.Exec(`
		INSERT INTO risk_controls (id, org_id, risk_id, control_id, effectiveness, mitigation_percentage,
		                           notes, linked_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
	`, rcID, orgID, riskID, req.ControlID, effectiveness, req.MitigationPercentage,
		req.Notes, userID, now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to link control to risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to link control"))
		return
	}

	middleware.LogAudit(c, "risk_control.linked", "risk_control", &rcID, map[string]interface{}{
		"risk_id":    riskID,
		"control_id": req.ControlID,
	})

	// Get control details for response
	var ctrlIdentifier, ctrlTitle string
	database.DB.QueryRow("SELECT identifier, title FROM controls WHERE id = $1", req.ControlID).Scan(&ctrlIdentifier, &ctrlTitle)

	var linkedByName string
	database.DB.QueryRow("SELECT COALESCE(first_name || ' ' || last_name, '') FROM users WHERE id = $1", userID).Scan(&linkedByName)

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":      rcID,
		"risk_id": riskID,
		"control": gin.H{
			"id":         req.ControlID,
			"identifier": ctrlIdentifier,
			"title":      ctrlTitle,
		},
		"effectiveness":        effectiveness,
		"mitigation_percentage": req.MitigationPercentage,
		"notes":                req.Notes,
		"linked_by":            gin.H{"id": userID, "name": linkedByName},
		"created_at":           now,
	}))
}

// UpdateRiskControl updates effectiveness for a risk-control linkage.
func UpdateRiskControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")
	controlID := c.Param("control_id")

	var req models.UpdateRiskControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Get risk-control record
	var rcID string
	var riskOwnerID *string
	err := database.DB.QueryRow(`
		SELECT rc.id, r.owner_id FROM risk_controls rc
		JOIN risks r ON rc.risk_id = r.id
		WHERE rc.risk_id = $1 AND rc.control_id = $2 AND rc.org_id = $3
	`, riskID, controlID, orgID).Scan(&rcID, &riskOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk-control link not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk-control link")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update"))
		return
	}

	// Authorization
	isOwner := riskOwnerID != nil && *riskOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.RiskControlLinkRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to update risk-control effectiveness"))
		return
	}

	now := time.Now()

	if req.Effectiveness != nil {
		if !models.IsValidEffectiveness(*req.Effectiveness) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid effectiveness value"))
			return
		}
	}
	if req.MitigationPercentage != nil && (*req.MitigationPercentage < 0 || *req.MitigationPercentage > 100) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Mitigation percentage must be between 0 and 100"))
		return
	}

	_, err = database.DB.Exec(`
		UPDATE risk_controls SET
			effectiveness = COALESCE($1, effectiveness),
			mitigation_percentage = COALESCE($2, mitigation_percentage),
			notes = COALESCE($3, notes),
			last_effectiveness_review = $4,
			reviewed_by = $5,
			updated_at = $4
		WHERE id = $6
	`, req.Effectiveness, req.MitigationPercentage, req.Notes, now, userID, rcID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update risk-control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update"))
		return
	}

	middleware.LogAudit(c, "risk_control.effectiveness_updated", "risk_control", &rcID, map[string]interface{}{
		"risk_id":    riskID,
		"control_id": controlID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                        rcID,
		"risk_id":                   riskID,
		"control_id":                controlID,
		"effectiveness":             req.Effectiveness,
		"mitigation_percentage":     req.MitigationPercentage,
		"last_effectiveness_review": now.Format("2006-01-02"),
		"reviewed_by":              gin.H{"id": userID},
		"updated_at":               now,
	}))
}

// UnlinkRiskControl unlinks a control from a risk.
func UnlinkRiskControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")
	controlID := c.Param("control_id")

	// Get record
	var rcID string
	var riskOwnerID *string
	err := database.DB.QueryRow(`
		SELECT rc.id, r.owner_id FROM risk_controls rc
		JOIN risks r ON rc.risk_id = r.id
		WHERE rc.risk_id = $1 AND rc.control_id = $2 AND rc.org_id = $3
	`, riskID, controlID, orgID).Scan(&rcID, &riskOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk-control link not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to find risk-control link")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to unlink"))
		return
	}

	// Authorization
	isOwner := riskOwnerID != nil && *riskOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.RiskControlLinkRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to unlink controls"))
		return
	}

	_, err = database.DB.Exec("DELETE FROM risk_controls WHERE id = $1", rcID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete risk-control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to unlink"))
		return
	}

	middleware.LogAudit(c, "risk_control.unlinked", "risk_control", &rcID, map[string]interface{}{
		"risk_id":    riskID,
		"control_id": controlID,
	})

	c.Status(http.StatusNoContent)
}
