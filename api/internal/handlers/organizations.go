package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// GetCurrentOrganization returns the authenticated user's organization.
func GetCurrentOrganization(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var id, name, slug, status, settingsStr string
	var domain sql.NullString
	var createdAt, updatedAt string

	err := database.QueryRow(`
		SELECT id, name, slug, domain, status, settings::text, created_at, updated_at
		FROM organizations WHERE id = $1
	`, orgID).Scan(&id, &name, &slug, &domain, &status, &settingsStr, &createdAt, &updatedAt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	var settings interface{}
	_ = json.Unmarshal([]byte(settingsStr), &settings)

	resp := gin.H{
		"id":         id,
		"name":       name,
		"slug":       slug,
		"status":     status,
		"settings":   settings,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}
	if domain.Valid {
		resp["domain"] = domain.String
	}

	c.JSON(http.StatusOK, successResponse(c, resp))
}

// UpdateCurrentOrganization updates the authenticated user's organization.
func UpdateCurrentOrganization(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req struct {
		Name     *string     `json:"name"`
		Domain   *string     `json:"domain"`
		Settings interface{} `json:"settings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Build dynamic update
	changes := map[string]interface{}{}

	if req.Name != nil {
		if len(*req.Name) > 255 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Name must be at most 255 characters"))
			return
		}
		_, err := database.Exec("UPDATE organizations SET name = $1 WHERE id = $2", *req.Name, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update org name")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
			return
		}
		changes["name"] = *req.Name
	}

	if req.Domain != nil {
		if len(*req.Domain) > 255 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Domain must be at most 255 characters"))
			return
		}
		_, err := database.Exec("UPDATE organizations SET domain = $1 WHERE id = $2", *req.Domain, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update org domain")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
			return
		}
		changes["domain"] = *req.Domain
	}

	if req.Settings != nil {
		// Merge settings (not replace) by using jsonb || operator
		settingsJSON, err := json.Marshal(req.Settings)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid settings JSON"))
			return
		}
		_, err = database.Exec("UPDATE organizations SET settings = settings || $1::jsonb WHERE id = $2", string(settingsJSON), orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update org settings")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
			return
		}
		changes["settings"] = req.Settings
	}

	if len(changes) > 0 {
		middleware.LogAudit(c, "org.updated", "organization", &orgID, changes)
	}

	// Return updated org
	GetCurrentOrganization(c)
}
