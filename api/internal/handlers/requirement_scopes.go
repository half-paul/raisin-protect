package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListScoping lists scoping decisions for an org framework.
func ListScoping(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ofID := c.Param("id")

	// Get the version ID for this org framework
	var versionID string
	err := database.QueryRow(`
		SELECT active_version_id FROM org_frameworks WHERE id = $1 AND org_id = $2
	`, ofID, orgID).Scan(&versionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Org framework not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get org framework")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	inScopeFilter := c.Query("in_scope")
	page, perPage := 1, 100
	if p := c.Query("page"); p != "" {
		if v, err := parsePositiveInt(p); err == nil {
			page = v
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if v, err := parsePositiveInt(pp); err == nil && v <= 200 {
			perPage = v
		}
	}

	where := "rs.org_id = $1 AND r.framework_version_id = $2"
	args := []interface{}{orgID, versionID}
	argN := 3

	if inScopeFilter == "true" {
		where += " AND rs.in_scope = TRUE"
	} else if inScopeFilter == "false" {
		where += " AND rs.in_scope = FALSE"
	}

	var total int
	database.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM requirement_scopes rs
		JOIN requirements r ON r.id = rs.requirement_id
		WHERE %s
	`, where), args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT rs.id, r.id, r.identifier, r.title, rs.in_scope, rs.justification,
			   u.id, COALESCE(u.first_name || ' ' || u.last_name, ''), rs.updated_at
		FROM requirement_scopes rs
		JOIN requirements r ON r.id = rs.requirement_id
		LEFT JOIN users u ON u.id = rs.scoped_by
		WHERE %s
		ORDER BY r.section_order
		LIMIT $%d OFFSET $%d
	`, where, argN, argN+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list scoping decisions")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			rsID, rID, rIdentifier, rTitle string
			inScope                        bool
			justification                  *string
			userID                         *string
			userName                       string
			updatedAt                      interface{}
		)
		if err := rows.Scan(&rsID, &rID, &rIdentifier, &rTitle, &inScope, &justification,
			&userID, &userName, &updatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan scoping row")
			continue
		}
		entry := gin.H{
			"id": rsID,
			"requirement": gin.H{
				"id":         rID,
				"identifier": rIdentifier,
				"title":      rTitle,
			},
			"in_scope":      inScope,
			"justification": justification,
			"updated_at":    updatedAt,
		}
		if userID != nil {
			entry["scoped_by"] = gin.H{
				"id":   *userID,
				"name": userName,
			}
		}
		results = append(results, entry)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// SetScope sets or updates the scoping decision for a requirement.
func SetScope(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	ofID := c.Param("id")
	reqID := c.Param("rid")

	// Verify org framework
	var versionID string
	err := database.QueryRow(`
		SELECT active_version_id FROM org_frameworks WHERE id = $1 AND org_id = $2
	`, ofID, orgID).Scan(&versionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Org framework not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get org framework")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Verify requirement belongs to this framework version
	var rIdentifier string
	err = database.QueryRow(`
		SELECT identifier FROM requirements WHERE id = $1 AND framework_version_id = $2
	`, reqID, versionID).Scan(&rIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Requirement doesn't belong to this framework version"))
		return
	}

	var req models.SetScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate: justification required when out-of-scope
	if !req.InScope && (req.Justification == nil || *req.Justification == "") {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Justification required when marking out-of-scope"))
		return
	}
	if req.Justification != nil && len(*req.Justification) > 2000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Justification must be at most 2000 characters"))
		return
	}

	// Upsert
	var scopeID string
	err = database.QueryRow(`
		SELECT id FROM requirement_scopes WHERE org_id = $1 AND requirement_id = $2
	`, orgID, reqID).Scan(&scopeID)

	if err == sql.ErrNoRows {
		scopeID = uuid.New().String()
		_, err = database.Exec(`
			INSERT INTO requirement_scopes (id, org_id, requirement_id, in_scope, justification, scoped_by)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, scopeID, orgID, reqID, req.InScope, req.Justification, userID)
	} else {
		_, err = database.Exec(`
			UPDATE requirement_scopes SET in_scope = $1, justification = $2, scoped_by = $3
			WHERE id = $4
		`, req.InScope, req.Justification, userID, scopeID)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to set scope")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "requirement.scoped", "requirement_scope", &scopeID, map[string]interface{}{
		"requirement": rIdentifier, "in_scope": req.InScope,
	})

	// Get user name
	var userName string
	database.QueryRow("SELECT first_name || ' ' || last_name FROM users WHERE id = $1", userID).Scan(&userName)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                     scopeID,
		"requirement_id":         reqID,
		"requirement_identifier": rIdentifier,
		"in_scope":               req.InScope,
		"justification":          req.Justification,
		"scoped_by": gin.H{
			"id":   userID,
			"name": userName,
		},
	}))
}

// ResetScope removes a scoping decision (requirement goes back to in-scope default).
func ResetScope(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ofID := c.Param("id")
	reqID := c.Param("rid")

	// Verify org framework
	var versionID string
	err := database.QueryRow(`
		SELECT active_version_id FROM org_frameworks WHERE id = $1 AND org_id = $2
	`, ofID, orgID).Scan(&versionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Org framework not found"))
		return
	}

	var rIdentifier string
	database.QueryRow("SELECT identifier FROM requirements WHERE id = $1", reqID).Scan(&rIdentifier)

	result, err := database.Exec(`
		DELETE FROM requirement_scopes WHERE org_id = $1 AND requirement_id = $2
	`, orgID, reqID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reset scope")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "No scoping decision exists for this requirement"))
		return
	}

	middleware.LogAudit(c, "requirement.scoped", "requirement_scope", &reqID, map[string]interface{}{
		"requirement": rIdentifier, "in_scope": true, "action": "reset",
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"message": "Scoping decision removed. Requirement is now implicitly in-scope.",
	}))
}
