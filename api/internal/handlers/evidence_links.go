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

// ListEvidenceLinks lists all links for an evidence artifact.
func ListEvidenceLinks(c *gin.Context) {
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

	rows, err := database.Query(`
		SELECT el.id, el.target_type, el.control_id, el.requirement_id,
			   el.strength, el.notes,
			   el.linked_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			   el.created_at
		FROM evidence_links el
		LEFT JOIN users u ON u.id = el.linked_by
		WHERE el.artifact_id = $1 AND el.org_id = $2
		ORDER BY el.created_at
	`, artifactID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list evidence links")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	links := []gin.H{}
	for rows.Next() {
		var lID, lType, lStrength string
		var lControlID, lReqID, lNotes, lLinkedByID *string
		var lLinkerName string
		var lCreatedAt time.Time

		if err := rows.Scan(&lID, &lType, &lControlID, &lReqID,
			&lStrength, &lNotes, &lLinkedByID, &lLinkerName, &lCreatedAt); err != nil {
			continue
		}

		link := gin.H{
			"id":          lID,
			"target_type": lType,
			"strength":    lStrength,
			"notes":       lNotes,
			"created_at":  lCreatedAt,
		}

		if lLinkedByID != nil {
			link["linked_by"] = gin.H{"id": *lLinkedByID, "name": lLinkerName}
		} else {
			link["linked_by"] = nil
		}

		if lType == "control" && lControlID != nil {
			var cID, cIdentifier, cTitle, cCategory, cStatus string
			database.QueryRow("SELECT id, identifier, title, category, status FROM controls WHERE id = $1",
				*lControlID).Scan(&cID, &cIdentifier, &cTitle, &cCategory, &cStatus)
			link["control"] = gin.H{
				"id": cID, "identifier": cIdentifier, "title": cTitle,
				"category": cCategory, "status": cStatus,
			}
			link["requirement"] = nil
		} else if lType == "requirement" && lReqID != nil {
			var rIdentifier, rTitle, fName, fvVersion string
			database.QueryRow(`
				SELECT r.identifier, r.title, f.name, fv.version
				FROM requirements r
				JOIN framework_versions fv ON fv.id = r.framework_version_id
				JOIN frameworks f ON f.id = fv.framework_id
				WHERE r.id = $1
			`, *lReqID).Scan(&rIdentifier, &rTitle, &fName, &fvVersion)
			link["control"] = nil
			link["requirement"] = gin.H{
				"id": *lReqID, "identifier": rIdentifier, "title": rTitle,
				"framework": fName, "framework_version": fvVersion,
			}
		}

		links = append(links, link)
	}

	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": links,
		"meta": gin.H{
			"total":      len(links),
			"request_id": reqID,
		},
	})
}

// CreateEvidenceLinks creates evidence links (single or bulk).
func CreateEvidenceLinks(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	artifactID := c.Param("id")

	// Verify artifact exists
	var exists bool
	database.QueryRow("SELECT EXISTS(SELECT 1 FROM evidence_artifacts WHERE id = $1 AND org_id = $2)",
		artifactID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence artifact not found"))
		return
	}

	var req models.BulkCreateLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Normalize: if single link, wrap in array
	links := req.Links
	if len(links) == 0 && req.TargetType != "" {
		links = []models.CreateLinkRequest{{
			TargetType:    req.TargetType,
			ControlID:     req.ControlID,
			RequirementID: req.RequirementID,
			Strength:      req.Strength,
			Notes:         req.Notes,
		}}
	}

	if len(links) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "At least one link is required"))
		return
	}
	if len(links) > 50 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Maximum 50 links per request"))
		return
	}

	createdLinks := []gin.H{}

	for _, link := range links {
		if !models.IsValidLinkTargetType(link.TargetType) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid target_type: must be 'control' or 'requirement'"))
			return
		}

		strength := "primary"
		if link.Strength != nil {
			if !models.IsValidLinkStrength(*link.Strength) {
				c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid strength: must be primary, supporting, or supplementary"))
				return
			}
			strength = *link.Strength
		}

		if link.Notes != nil && len(*link.Notes) > 2000 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Notes must be at most 2000 characters"))
			return
		}

		linkID := uuid.New().String()

		if link.TargetType == "control" {
			if link.ControlID == nil {
				c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "control_id is required for target_type 'control'"))
				return
			}
			// Verify control exists in org
			var ctrlExists bool
			database.QueryRow("SELECT EXISTS(SELECT 1 FROM controls WHERE id = $1 AND org_id = $2)",
				*link.ControlID, orgID).Scan(&ctrlExists)
			if !ctrlExists {
				c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found in organization"))
				return
			}
			// Check duplicate
			var dupExists bool
			database.QueryRow("SELECT EXISTS(SELECT 1 FROM evidence_links WHERE artifact_id = $1 AND control_id = $2 AND org_id = $3)",
				artifactID, *link.ControlID, orgID).Scan(&dupExists)
			if dupExists {
				c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Link already exists for this artifact-control pair"))
				return
			}

			_, err := database.Exec(`
				INSERT INTO evidence_links (id, org_id, artifact_id, target_type, control_id, strength, notes, linked_by)
				VALUES ($1, $2, $3, 'control', $4, $5, $6, $7)
			`, linkID, orgID, artifactID, *link.ControlID, strength, link.Notes, userID)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create evidence link")
				c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
				return
			}

			middleware.LogAudit(c, "evidence.linked", "evidence_link", &linkID, map[string]interface{}{
				"artifact_id": artifactID, "target_type": "control", "target_id": *link.ControlID,
			})

			createdLinks = append(createdLinks, gin.H{
				"id": linkID, "target_type": "control", "control_id": *link.ControlID,
				"strength": strength, "created_at": time.Now(),
			})

		} else if link.TargetType == "requirement" {
			if link.RequirementID == nil {
				c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "requirement_id is required for target_type 'requirement'"))
				return
			}
			// Verify requirement exists
			var reqExists bool
			database.QueryRow("SELECT EXISTS(SELECT 1 FROM requirements r WHERE r.id = $1)",
				*link.RequirementID).Scan(&reqExists)
			if !reqExists {
				c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Requirement not found"))
				return
			}
			// Check duplicate
			var dupExists bool
			database.QueryRow("SELECT EXISTS(SELECT 1 FROM evidence_links WHERE artifact_id = $1 AND requirement_id = $2 AND org_id = $3)",
				artifactID, *link.RequirementID, orgID).Scan(&dupExists)
			if dupExists {
				c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Link already exists for this artifact-requirement pair"))
				return
			}

			_, err := database.Exec(`
				INSERT INTO evidence_links (id, org_id, artifact_id, target_type, requirement_id, strength, notes, linked_by)
				VALUES ($1, $2, $3, 'requirement', $4, $5, $6, $7)
			`, linkID, orgID, artifactID, *link.RequirementID, strength, link.Notes, userID)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create evidence link")
				c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
				return
			}

			middleware.LogAudit(c, "evidence.linked", "evidence_link", &linkID, map[string]interface{}{
				"artifact_id": artifactID, "target_type": "requirement", "target_id": *link.RequirementID,
			})

			createdLinks = append(createdLinks, gin.H{
				"id": linkID, "target_type": "requirement", "requirement_id": *link.RequirementID,
				"strength": strength, "created_at": time.Now(),
			})
		}
	}

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"created": len(createdLinks),
		"links":   createdLinks,
	}))
}

// DeleteEvidenceLink removes an evidence link.
func DeleteEvidenceLink(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	artifactID := c.Param("id")
	linkID := c.Param("lid")

	var targetType string
	var controlID, requirementID *string
	err := database.QueryRow(`
		SELECT target_type, control_id, requirement_id FROM evidence_links
		WHERE id = $1 AND artifact_id = $2 AND org_id = $3
	`, linkID, artifactID, orgID).Scan(&targetType, &controlID, &requirementID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Evidence link not found"))
		return
	}

	_, err = database.Exec("DELETE FROM evidence_links WHERE id = $1", linkID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete evidence link")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	targetID := ""
	if controlID != nil {
		targetID = *controlID
	} else if requirementID != nil {
		targetID = *requirementID
	}

	middleware.LogAudit(c, "evidence.unlinked", "evidence_link", &linkID, map[string]interface{}{
		"artifact_id": artifactID, "target_type": targetType, "target_id": targetID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"message": "Evidence link removed",
	}))
}
