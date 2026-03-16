package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// GetCompensatingWorksheet handles GET /api/v1/controls/:id/compensating-worksheet.
// Returns the PCI DSS Appendix B worksheet for a compensating control.
func GetCompensatingWorksheet(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	var isCompensating bool
	var worksheetJSON *string

	err := database.QueryRow(`
		SELECT is_compensating, compensating_worksheet::text
		FROM controls
		WHERE id = $1 AND org_id = $2`,
		controlID, orgID,
	).Scan(&isCompensating, &worksheetJSON)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "control not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get compensating worksheet")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	var worksheet *models.CompensatingWorksheetData
	if worksheetJSON != nil && *worksheetJSON != "" && *worksheetJSON != "null" {
		worksheet = &models.CompensatingWorksheetData{}
		if err := json.Unmarshal([]byte(*worksheetJSON), worksheet); err != nil {
			log.Error().Err(err).Msg("Failed to parse compensating worksheet JSON")
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"control_id":      controlID,
		"is_compensating": isCompensating,
		"worksheet":       worksheet,
	}))
}

// UpdateCompensatingWorksheet handles PUT /api/v1/controls/:id/compensating-worksheet.
// Creates or replaces the PCI DSS Appendix B worksheet. Sets is_compensating = TRUE.
func UpdateCompensatingWorksheet(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	var req models.CompensatingWorksheetData
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Sanitize all free-text fields to prevent stored XSS.
	sanitizeStringPtr := func(s *string) *string {
		if s == nil {
			return nil
		}
		clean := sanitizeHTML(*s)
		return &clean
	}
	req.OriginalRequirement = sanitizeStringPtr(req.OriginalRequirement)
	req.Constraint = sanitizeStringPtr(req.Constraint)
	req.Objective = sanitizeStringPtr(req.Objective)
	req.CompensatingControl = sanitizeStringPtr(req.CompensatingControl)
	req.Validation = sanitizeStringPtr(req.Validation)
	req.RiskAssessment = sanitizeStringPtr(req.RiskAssessment)
	req.MaintenancePlan = sanitizeStringPtr(req.MaintenancePlan)

	worksheetJSON, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "failed to serialize worksheet"))
		return
	}

	var updatedIsCompensating bool
	var updatedWorksheetJSON string

	err = database.QueryRow(`
		UPDATE controls
		SET is_compensating = TRUE,
		    compensating_worksheet = $1::jsonb,
		    updated_at = NOW()
		WHERE id = $2 AND org_id = $3
		RETURNING is_compensating, compensating_worksheet::text`,
		string(worksheetJSON), controlID, orgID,
	).Scan(&updatedIsCompensating, &updatedWorksheetJSON)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "control not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to update compensating worksheet")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	var worksheet models.CompensatingWorksheetData
	json.Unmarshal([]byte(updatedWorksheetJSON), &worksheet)

	middleware.LogAudit(c, "control.compensating_worksheet.updated", "control", &controlID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"control_id":      controlID,
		"is_compensating": updatedIsCompensating,
		"worksheet":       worksheet,
	}))
}
