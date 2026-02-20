package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListControlMappings lists all requirement mappings for a control.
func ListControlMappings(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	// Verify control exists
	var exists bool
	database.QueryRow("SELECT EXISTS(SELECT 1 FROM controls WHERE id = $1 AND org_id = $2)", controlID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	rows, err := database.Query(`
		SELECT cm.id, r.id, r.identifier, r.title,
			   f.identifier, f.name, fv.version,
			   cm.strength, cm.notes,
			   cm.mapped_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   cm.created_at
		FROM control_mappings cm
		JOIN requirements r ON r.id = cm.requirement_id
		JOIN framework_versions fv ON fv.id = r.framework_version_id
		JOIN frameworks f ON f.id = fv.framework_id
		LEFT JOIN users u ON u.id = cm.mapped_by
		WHERE cm.control_id = $1 AND cm.org_id = $2
		ORDER BY f.name, r.identifier
	`, controlID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list control mappings")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	mappings := []gin.H{}
	for rows.Next() {
		var (
			cmID, rID, rIdentifier, rTitle  string
			fIdentifier, fName, fvVersion   string
			strength                        string
			notes                           *string
			mappedByID                      *string
			mappedByName                    string
			createdAt                       interface{}
		)
		if err := rows.Scan(&cmID, &rID, &rIdentifier, &rTitle,
			&fIdentifier, &fName, &fvVersion,
			&strength, &notes, &mappedByID, &mappedByName, &createdAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan mapping row")
			continue
		}

		entry := gin.H{
			"id": cmID,
			"requirement": gin.H{
				"id":         rID,
				"identifier": rIdentifier,
				"title":      rTitle,
				"framework": gin.H{
					"identifier": fIdentifier,
					"name":       fName,
				},
				"version": fvVersion,
			},
			"strength":   strength,
			"notes":      notes,
			"created_at": createdAt,
		}
		if mappedByID != nil {
			entry["mapped_by"] = gin.H{"id": *mappedByID, "name": mappedByName}
		}
		mappings = append(mappings, entry)
	}

	reqID, _ := c.Get("request_id")
	c.JSON(http.StatusOK, gin.H{
		"data": mappings,
		"meta": gin.H{
			"total":      len(mappings),
			"request_id": reqID,
		},
	})
}

// CreateControlMappings creates one or more control-requirement mappings.
func CreateControlMappings(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	controlID := c.Param("id")

	// Verify control
	var cIdentifier string
	err := database.QueryRow("SELECT identifier FROM controls WHERE id = $1 AND org_id = $2", controlID, orgID).Scan(&cIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	var req models.BulkCreateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Normalize: single mapping → bulk
	mappings := req.Mappings
	if len(mappings) == 0 && req.RequirementID != "" {
		mappings = []models.CreateMappingRequest{{
			RequirementID: req.RequirementID,
			Strength:      req.Strength,
			Notes:         req.Notes,
		}}
	}

	if len(mappings) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "At least one mapping required"))
		return
	}
	if len(mappings) > 50 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Maximum 50 mappings per request"))
		return
	}

	created := []gin.H{}
	for _, m := range mappings {
		if m.RequirementID == "" {
			continue
		}

		strength := "primary"
		if m.Strength != nil && *m.Strength != "" {
			if !models.IsValidMappingStrength(*m.Strength) {
				continue // skip invalid
			}
			strength = *m.Strength
		}

		// Verify requirement exists and is assessable
		var rIdentifier string
		var isAssessable bool
		err := database.QueryRow(`
			SELECT r.identifier, r.is_assessable FROM requirements r
			WHERE r.id = $1
		`, m.RequirementID).Scan(&rIdentifier, &isAssessable)
		if err == sql.ErrNoRows {
			continue
		}
		if !isAssessable {
			continue
		}

		// Verify requirement belongs to an activated framework
		var fwActivated bool
		database.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM org_frameworks of
				JOIN framework_versions fv ON fv.id = of.active_version_id
				JOIN requirements r ON r.framework_version_id = fv.id
				WHERE of.org_id = $1 AND r.id = $2 AND of.status = 'active'
			)
		`, orgID, m.RequirementID).Scan(&fwActivated)
		if !fwActivated {
			continue
		}

		// Check for duplicate
		var dupExists bool
		database.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM control_mappings WHERE org_id = $1 AND control_id = $2 AND requirement_id = $3)
		`, orgID, controlID, m.RequirementID).Scan(&dupExists)
		if dupExists {
			continue
		}

		mappingID := uuid.New().String()
		_, err = database.Exec(`
			INSERT INTO control_mappings (id, org_id, control_id, requirement_id, strength, notes, mapped_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, mappingID, orgID, controlID, m.RequirementID, strength, m.Notes, userID)
		if err != nil {
			log.Error().Err(err).Str("mapping_id", mappingID).Msg("Failed to create mapping")
			continue
		}

		// Get framework info for audit
		var fIdentifier string
		database.QueryRow(`
			SELECT f.identifier FROM requirements r
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			JOIN frameworks f ON f.id = fv.framework_id
			WHERE r.id = $1
		`, m.RequirementID).Scan(&fIdentifier)

		middleware.LogAudit(c, "control_mapping.created", "control_mapping", &mappingID, map[string]interface{}{
			"control": cIdentifier, "requirement": rIdentifier, "framework": fIdentifier,
		})

		created = append(created, gin.H{
			"id":             mappingID,
			"control_id":     controlID,
			"requirement_id": m.RequirementID,
			"strength":       strength,
			"created_at":     "now",
		})
	}

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"created":  len(created),
		"mappings": created,
	}))
}

// DeleteControlMapping removes a single control-requirement mapping.
func DeleteControlMapping(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")
	mappingID := c.Param("mid")

	// Get mapping details for audit
	var cIdentifier, rIdentifier, fIdentifier string
	err := database.QueryRow(`
		SELECT c.identifier, r.identifier, f.identifier
		FROM control_mappings cm
		JOIN controls c ON c.id = cm.control_id
		JOIN requirements r ON r.id = cm.requirement_id
		JOIN framework_versions fv ON fv.id = r.framework_version_id
		JOIN frameworks f ON f.id = fv.framework_id
		WHERE cm.id = $1 AND cm.control_id = $2 AND cm.org_id = $3
	`, mappingID, controlID, orgID).Scan(&cIdentifier, &rIdentifier, &fIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Mapping not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get mapping")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	_, err = database.Exec("DELETE FROM control_mappings WHERE id = $1", mappingID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete mapping")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "control_mapping.deleted", "control_mapping", &mappingID, map[string]interface{}{
		"control": cIdentifier, "requirement": rIdentifier, "framework": fIdentifier,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"message": "Mapping removed",
	}))
}
