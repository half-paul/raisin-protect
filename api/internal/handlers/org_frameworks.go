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

// ListOrgFrameworks lists frameworks activated by the org.
func ListOrgFrameworks(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	statusFilter := c.Query("status")

	query := `
		SELECT of.id, of.status, of.target_date::text, of.notes, of.activated_at, of.created_at,
			   f.id, f.identifier, f.name, f.category,
			   fv.id, fv.version, fv.display_name, fv.total_requirements
		FROM org_frameworks of
		JOIN frameworks f ON f.id = of.framework_id
		JOIN framework_versions fv ON fv.id = of.active_version_id
		WHERE of.org_id = $1
	`
	args := []interface{}{orgID}

	if statusFilter != "" {
		query += " AND of.status = $2"
		args = append(args, statusFilter)
	}
	query += " ORDER BY f.name"

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list org frameworks")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			ofID, ofStatus, fID, fIdentifier, fName, fCategory string
			fvID, fvVersion, fvDisplayName                     string
			ofTargetDate, ofNotes                              *string
			ofActivatedAt, ofCreatedAt                         interface{}
			fvTotalReqs                                        int
		)
		if err := rows.Scan(&ofID, &ofStatus, &ofTargetDate, &ofNotes, &ofActivatedAt, &ofCreatedAt,
			&fID, &fIdentifier, &fName, &fCategory,
			&fvID, &fvVersion, &fvDisplayName, &fvTotalReqs); err != nil {
			log.Error().Err(err).Msg("Failed to scan org framework row")
			continue
		}

		// Compute coverage stats for this org framework
		stats := computeCoverageStats(orgID, fvID)

		results = append(results, gin.H{
			"id": ofID,
			"framework": gin.H{
				"id":         fID,
				"identifier": fIdentifier,
				"name":       fName,
				"category":   fCategory,
			},
			"active_version": gin.H{
				"id":                 fvID,
				"version":            fvVersion,
				"display_name":       fvDisplayName,
				"total_requirements": fvTotalReqs,
			},
			"status":       ofStatus,
			"target_date":  ofTargetDate,
			"notes":        ofNotes,
			"stats":        stats,
			"activated_at": ofActivatedAt,
			"created_at":   ofCreatedAt,
		})
	}

	reqID, _ := c.Get("request_id")
	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{
			"total":      len(results),
			"request_id": reqID,
		},
	})
}

// computeCoverageStats calculates coverage stats for a framework version in an org.
func computeCoverageStats(orgID, versionID string) gin.H {
	var totalReqs, assessableReqs int
	database.QueryRow(`
		SELECT COUNT(*), COUNT(*) FILTER (WHERE is_assessable = TRUE)
		FROM requirements WHERE framework_version_id = $1
	`, versionID).Scan(&totalReqs, &assessableReqs)

	var outOfScope int
	database.QueryRow(`
		SELECT COUNT(*) FROM requirement_scopes rs
		JOIN requirements r ON r.id = rs.requirement_id
		WHERE rs.org_id = $1 AND r.framework_version_id = $2 AND rs.in_scope = FALSE AND r.is_assessable = TRUE
	`, orgID, versionID).Scan(&outOfScope)

	inScope := assessableReqs - outOfScope

	var mapped int
	database.QueryRow(`
		SELECT COUNT(DISTINCT r.id) FROM requirements r
		JOIN control_mappings cm ON cm.requirement_id = r.id AND cm.org_id = $1
		LEFT JOIN requirement_scopes rs ON rs.requirement_id = r.id AND rs.org_id = $1
		WHERE r.framework_version_id = $2 AND r.is_assessable = TRUE
		  AND (rs.id IS NULL OR rs.in_scope = TRUE)
	`, orgID, versionID).Scan(&mapped)

	unmapped := inScope - mapped
	if unmapped < 0 {
		unmapped = 0
	}

	coveragePct := 0.0
	if inScope > 0 {
		coveragePct = float64(mapped) / float64(inScope) * 100
	}

	return gin.H{
		"total_requirements": totalReqs,
		"in_scope":           inScope,
		"out_of_scope":       outOfScope,
		"mapped":             mapped,
		"unmapped":           unmapped,
		"coverage_pct":       coveragePct,
	}
}

// ActivateFramework activates a framework for the org.
func ActivateFramework(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.ActivateFrameworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate notes length
	if req.Notes != nil && len(*req.Notes) > 2000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Notes must be at most 2000 characters"))
		return
	}

	// Verify framework exists
	var fwExists bool
	database.QueryRow("SELECT EXISTS(SELECT 1 FROM frameworks WHERE id = $1)", req.FrameworkID).Scan(&fwExists)
	if !fwExists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Framework not found"))
		return
	}

	// Verify version exists and is active for this framework
	var versionStatus string
	err := database.QueryRow(`
		SELECT status FROM framework_versions WHERE id = $1 AND framework_id = $2
	`, req.VersionID, req.FrameworkID).Scan(&versionStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Framework version not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check framework version")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	if versionStatus != "active" {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Version is not in active status"))
		return
	}

	// Check for existing activation
	var alreadyExists bool
	database.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM org_frameworks WHERE org_id = $1 AND framework_id = $2)
	`, orgID, req.FrameworkID).Scan(&alreadyExists)
	if alreadyExists {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Framework already activated for this organization"))
		return
	}

	ofID := uuid.New().String()
	_, err = database.Exec(`
		INSERT INTO org_frameworks (id, org_id, framework_id, active_version_id, status, target_date, notes)
		VALUES ($1, $2, $3, $4, 'active', $5, $6)
	`, ofID, orgID, req.FrameworkID, req.VersionID, req.TargetDate, req.Notes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to activate framework")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Seed controls if requested (default true)
	seedControls := req.SeedControls == nil || *req.SeedControls
	controlsSeeded, mappingsSeeded := 0, 0
	if seedControls {
		controlsSeeded, mappingsSeeded = seedFrameworkControls(orgID, req.VersionID)
	}

	// Get framework info for response
	var fIdentifier, fName string
	database.QueryRow("SELECT identifier, name FROM frameworks WHERE id = $1", req.FrameworkID).Scan(&fIdentifier, &fName)
	var vVersion, vDisplayName string
	database.QueryRow("SELECT version, display_name FROM framework_versions WHERE id = $1", req.VersionID).Scan(&vVersion, &vDisplayName)

	middleware.LogAudit(c, "framework.activated", "org_framework", &ofID, map[string]interface{}{
		"framework": fIdentifier, "version": vVersion, "seed_controls": seedControls,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id": ofID,
		"framework": gin.H{
			"id":         req.FrameworkID,
			"identifier": fIdentifier,
			"name":       fName,
		},
		"active_version": gin.H{
			"id":           req.VersionID,
			"version":      vVersion,
			"display_name": vDisplayName,
		},
		"status":           "active",
		"target_date":      req.TargetDate,
		"notes":            req.Notes,
		"controls_seeded":  controlsSeeded,
		"mappings_seeded":  mappingsSeeded,
		"activated_at":     "now",
	}))
}

// seedFrameworkControls seeds controls from the library template for the given framework version.
func seedFrameworkControls(orgID, versionID string) (int, int) {
	// For now, seeding is handled by the DBE seed data — controls already exist for demo org.
	// In a production system, this would clone control templates into the org.
	// Return 0,0 since the demo org already has controls seeded.
	return 0, 0
}

// UpdateOrgFramework updates an org's framework activation.
func UpdateOrgFramework(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ofID := c.Param("id")

	// Verify org framework exists
	var currentVersionID, currentStatus, frameworkID string
	err := database.QueryRow(`
		SELECT active_version_id, status, framework_id FROM org_frameworks WHERE id = $1 AND org_id = $2
	`, ofID, orgID).Scan(&currentVersionID, &currentStatus, &frameworkID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Org framework not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get org framework")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	var req models.UpdateOrgFrameworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if req.Notes != nil && len(*req.Notes) > 2000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Notes must be at most 2000 characters"))
		return
	}

	// Version change validation
	if req.VersionID != nil {
		var vStatus string
		err := database.QueryRow(`
			SELECT status FROM framework_versions WHERE id = $1 AND framework_id = $2
		`, *req.VersionID, frameworkID).Scan(&vStatus)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Version doesn't belong to this framework"))
			return
		}
		if vStatus != "active" {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Version is not in active status"))
			return
		}
		database.Exec("UPDATE org_frameworks SET active_version_id = $1 WHERE id = $2", *req.VersionID, ofID)
		middleware.LogAudit(c, "framework.version_changed", "org_framework", &ofID, map[string]interface{}{
			"old_version": currentVersionID, "new_version": *req.VersionID,
		})
	}

	if req.TargetDate != nil {
		database.Exec("UPDATE org_frameworks SET target_date = $1 WHERE id = $2", *req.TargetDate, ofID)
	}
	if req.Notes != nil {
		database.Exec("UPDATE org_frameworks SET notes = $1 WHERE id = $2", *req.Notes, ofID)
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "inactive" {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Status must be active or inactive"))
			return
		}
		database.Exec("UPDATE org_frameworks SET status = $1 WHERE id = $2", *req.Status, ofID)
		if *req.Status == "inactive" {
			database.Exec("UPDATE org_frameworks SET deactivated_at = NOW() WHERE id = $1", ofID)
			middleware.LogAudit(c, "framework.deactivated", "org_framework", &ofID, nil)
		} else {
			middleware.LogAudit(c, "framework.activated", "org_framework", &ofID, nil)
		}
	}

	// Return updated record
	var of struct {
		Status, TargetDate, Notes                    *string
		ActivatedAt, UpdatedAt                       interface{}
		FID, FIdentifier, FName                      string
		FVID, FVVersion, FVDisplayName               string
	}
	database.QueryRow(`
		SELECT of.status, of.target_date::text, of.notes, of.activated_at, of.updated_at,
			   f.id, f.identifier, f.name,
			   fv.id, fv.version, fv.display_name
		FROM org_frameworks of
		JOIN frameworks f ON f.id = of.framework_id
		JOIN framework_versions fv ON fv.id = of.active_version_id
		WHERE of.id = $1
	`, ofID).Scan(&of.Status, &of.TargetDate, &of.Notes, &of.ActivatedAt, &of.UpdatedAt,
		&of.FID, &of.FIdentifier, &of.FName,
		&of.FVID, &of.FVVersion, &of.FVDisplayName)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": ofID,
		"framework": gin.H{
			"id":         of.FID,
			"identifier": of.FIdentifier,
			"name":       of.FName,
		},
		"active_version": gin.H{
			"id":           of.FVID,
			"version":      of.FVVersion,
			"display_name": of.FVDisplayName,
		},
		"status":      of.Status,
		"target_date": of.TargetDate,
		"notes":       of.Notes,
		"updated_at":  of.UpdatedAt,
	}))
}

// DeactivateFramework deactivates a framework for the org (soft delete).
func DeactivateFramework(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ofID := c.Param("id")

	var exists bool
	database.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM org_frameworks WHERE id = $1 AND org_id = $2)
	`, ofID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Org framework not found"))
		return
	}

	_, err := database.Exec(`
		UPDATE org_frameworks SET status = 'inactive', deactivated_at = NOW() WHERE id = $1
	`, ofID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to deactivate framework")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "framework.deactivated", "org_framework", &ofID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":      ofID,
		"status":  "inactive",
		"message": "Framework deactivated. Controls and mappings preserved.",
	}))
}

// GetCoverage returns coverage gap analysis for an activated framework.
func GetCoverage(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	ofID := c.Param("id")
	statusFilter := c.Query("status")

	// Get the framework version
	var fIdentifier, fName, fvVersion, versionID string
	err := database.QueryRow(`
		SELECT f.identifier, f.name, fv.version, fv.id
		FROM org_frameworks of
		JOIN frameworks f ON f.id = of.framework_id
		JOIN framework_versions fv ON fv.id = of.active_version_id
		WHERE of.id = $1 AND of.org_id = $2
	`, ofID, orgID).Scan(&fIdentifier, &fName, &fvVersion, &versionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Org framework not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get coverage")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Compute summary
	stats := computeCoverageStats(orgID, versionID)

	// Get page params
	page := 1
	perPage := 50
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

	// Get requirements with coverage info
	query := `
		SELECT r.id, r.identifier, r.title, r.depth,
			   CASE WHEN rs.in_scope = FALSE THEN FALSE ELSE TRUE END AS in_scope,
			   CASE WHEN cm_count.cnt > 0 THEN 'covered' ELSE 'gap' END AS status
		FROM requirements r
		LEFT JOIN requirement_scopes rs ON rs.requirement_id = r.id AND rs.org_id = $1
		LEFT JOIN (
			SELECT requirement_id, COUNT(*) AS cnt
			FROM control_mappings WHERE org_id = $1
			GROUP BY requirement_id
		) cm_count ON cm_count.requirement_id = r.id
		WHERE r.framework_version_id = $2 AND r.is_assessable = TRUE
		  AND (rs.id IS NULL OR rs.in_scope = TRUE)
	`
	args := []interface{}{orgID, versionID}
	argN := 3

	if statusFilter == "covered" {
		query += " AND cm_count.cnt > 0"
	} else if statusFilter == "gap" {
		query += " AND (cm_count.cnt IS NULL OR cm_count.cnt = 0)"
	}

	// Count total for pagination
	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") sub"
	database.QueryRow(countQuery, args...).Scan(&total)

	query += " ORDER BY r.section_order"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query coverage requirements")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	requirements := []gin.H{}
	for rows.Next() {
		var rID, rIdentifier, rTitle, rStatus string
		var rDepth int
		var rInScope bool
		if err := rows.Scan(&rID, &rIdentifier, &rTitle, &rDepth, &rInScope, &rStatus); err != nil {
			log.Error().Err(err).Msg("Failed to scan coverage row")
			continue
		}

		// Get mapped controls for this requirement
		controls := getControlsForRequirement(orgID, rID)

		requirements = append(requirements, gin.H{
			"id":         rID,
			"identifier": rIdentifier,
			"title":      rTitle,
			"depth":      rDepth,
			"in_scope":   rInScope,
			"status":     rStatus,
			"controls":   controls,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"framework": gin.H{
				"identifier": fIdentifier,
				"name":       fName,
				"version":    fvVersion,
			},
			"summary":      stats,
			"requirements": requirements,
		},
		"meta": listMeta(c, total, page, perPage),
	})
}

// getControlsForRequirement returns the controls mapped to a requirement.
func getControlsForRequirement(orgID, requirementID string) []gin.H {
	rows, err := database.Query(`
		SELECT c.id, c.identifier, c.title, cm.strength, c.status
		FROM control_mappings cm
		JOIN controls c ON c.id = cm.control_id
		WHERE cm.org_id = $1 AND cm.requirement_id = $2
	`, orgID, requirementID)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()

	controls := []gin.H{}
	for rows.Next() {
		var cID, cIdentifier, cTitle, cStrength, cStatus string
		if err := rows.Scan(&cID, &cIdentifier, &cTitle, &cStrength, &cStatus); err != nil {
			continue
		}
		controls = append(controls, gin.H{
			"id":         cID,
			"identifier": cIdentifier,
			"title":      cTitle,
			"strength":   cStrength,
			"status":     cStatus,
		})
	}
	return controls
}

// listMeta creates a meta object for paginated responses.
func listMeta(c *gin.Context, total, page, perPage int) gin.H {
	reqID, _ := c.Get("request_id")
	return gin.H{
		"total":      total,
		"page":       page,
		"per_page":   perPage,
		"request_id": reqID,
	}
}

// parsePositiveInt parses a positive integer from a string.
func parsePositiveInt(s string) (int, error) {
	v := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number")
		}
		v = v*10 + int(c-'0')
	}
	if v < 1 {
		return 0, fmt.Errorf("must be positive")
	}
	return v, nil
}
