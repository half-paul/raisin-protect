package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListFrameworks lists all compliance frameworks in the catalog.
func ListFrameworks(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("search")

	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	if category != "" {
		where = append(where, fmt.Sprintf("f.category = $%d", argN))
		args = append(args, category)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(f.name ILIKE $%d OR f.identifier ILIKE $%d)", argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	query := fmt.Sprintf(`
		SELECT f.id, f.identifier, f.name, f.description, f.category, f.website_url, f.logo_url,
			   COALESCE((SELECT COUNT(*) FROM framework_versions fv WHERE fv.framework_id = f.id), 0) AS versions_count,
			   f.created_at
		FROM frameworks f
		WHERE %s
		ORDER BY f.name
	`, whereClause)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list frameworks")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	frameworks := []gin.H{}
	for rows.Next() {
		var f models.Framework
		if err := rows.Scan(&f.ID, &f.Identifier, &f.Name, &f.Description, &f.Category,
			&f.WebsiteURL, &f.LogoURL, &f.VersionsCount, &f.CreatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan framework row")
			continue
		}
		frameworks = append(frameworks, gin.H{
			"id":             f.ID,
			"identifier":     f.Identifier,
			"name":           f.Name,
			"description":    f.Description,
			"category":       f.Category,
			"website_url":    f.WebsiteURL,
			"logo_url":       f.LogoURL,
			"versions_count": f.VersionsCount,
			"created_at":     f.CreatedAt,
		})
	}

	reqID, _ := c.Get("request_id")
	c.JSON(http.StatusOK, gin.H{
		"data": frameworks,
		"meta": gin.H{
			"total":      len(frameworks),
			"request_id": reqID,
		},
	})
}

// GetFramework returns a single framework with all its versions.
func GetFramework(c *gin.Context) {
	frameworkID := c.Param("id")

	var f models.Framework
	err := database.QueryRow(`
		SELECT id, identifier, name, description, category, website_url, logo_url, created_at
		FROM frameworks WHERE id = $1
	`, frameworkID).Scan(&f.ID, &f.Identifier, &f.Name, &f.Description, &f.Category,
		&f.WebsiteURL, &f.LogoURL, &f.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Framework not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get framework")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Get versions
	rows, err := database.Query(`
		SELECT id, version, display_name, status, effective_date::text, sunset_date::text,
			   total_requirements, created_at
		FROM framework_versions WHERE framework_id = $1
		ORDER BY effective_date DESC NULLS LAST
	`, frameworkID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list framework versions")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	versions := []gin.H{}
	for rows.Next() {
		var v models.FrameworkVersion
		if err := rows.Scan(&v.ID, &v.Version, &v.DisplayName, &v.Status,
			&v.EffectiveDate, &v.SunsetDate, &v.TotalRequirements, &v.CreatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan version row")
			continue
		}
		versions = append(versions, gin.H{
			"id":                 v.ID,
			"version":            v.Version,
			"display_name":       v.DisplayName,
			"status":             v.Status,
			"effective_date":     v.EffectiveDate,
			"sunset_date":        v.SunsetDate,
			"total_requirements": v.TotalRequirements,
			"created_at":         v.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":          f.ID,
		"identifier":  f.Identifier,
		"name":        f.Name,
		"description": f.Description,
		"category":    f.Category,
		"website_url": f.WebsiteURL,
		"logo_url":    f.LogoURL,
		"versions":    versions,
		"created_at":  f.CreatedAt,
	}))
}

// GetFrameworkVersion returns a specific version of a framework.
func GetFrameworkVersion(c *gin.Context) {
	frameworkID := c.Param("id")
	versionID := c.Param("vid")

	var v models.FrameworkVersion
	err := database.QueryRow(`
		SELECT fv.id, fv.framework_id, f.identifier, f.name, fv.version, fv.display_name,
			   fv.status, fv.effective_date::text, fv.sunset_date::text, fv.changelog,
			   fv.total_requirements, fv.created_at
		FROM framework_versions fv
		JOIN frameworks f ON f.id = fv.framework_id
		WHERE fv.id = $1 AND fv.framework_id = $2
	`, versionID, frameworkID).Scan(&v.ID, &v.FrameworkID, &v.FrameworkIdentifier, &v.FrameworkName,
		&v.Version, &v.DisplayName, &v.Status, &v.EffectiveDate, &v.SunsetDate, &v.Changelog,
		&v.TotalRequirements, &v.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Framework version not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get framework version")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                   v.ID,
		"framework_id":         v.FrameworkID,
		"framework_identifier": v.FrameworkIdentifier,
		"framework_name":       v.FrameworkName,
		"version":              v.Version,
		"display_name":         v.DisplayName,
		"status":               v.Status,
		"effective_date":       v.EffectiveDate,
		"sunset_date":          v.SunsetDate,
		"changelog":            v.Changelog,
		"total_requirements":   v.TotalRequirements,
		"created_at":           v.CreatedAt,
	}))
}

// ListRequirements lists requirements for a framework version (flat or tree).
func ListRequirements(c *gin.Context) {
	frameworkID := c.Param("id")
	versionID := c.Param("vid")

	// Verify version exists
	var exists bool
	err := database.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM framework_versions WHERE id = $1 AND framework_id = $2)
	`, versionID, frameworkID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Framework version not found"))
		return
	}

	format := c.DefaultQuery("format", "flat")
	assessableOnly := c.Query("assessable_only") == "true"
	parentID := c.Query("parent_id")
	search := c.Query("search")

	if format == "tree" {
		listRequirementsTree(c, versionID, assessableOnly)
		return
	}

	// Flat mode with pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	where := []string{"r.framework_version_id = $1"}
	args := []interface{}{versionID}
	argN := 2

	if assessableOnly {
		where = append(where, "r.is_assessable = TRUE")
	}
	if parentID != "" {
		where = append(where, fmt.Sprintf("r.parent_id = $%d", argN))
		args = append(args, parentID)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(r.identifier ILIKE $%d OR r.title ILIKE $%d)", argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM requirements r WHERE %s", whereClause), countArgs...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.identifier, r.title, r.description, r.guidance, r.parent_id,
			   r.depth, r.section_order, r.is_assessable, r.created_at
		FROM requirements r
		WHERE %s
		ORDER BY r.depth, r.section_order
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list requirements")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	reqs := []gin.H{}
	for rows.Next() {
		var r models.Requirement
		if err := rows.Scan(&r.ID, &r.Identifier, &r.Title, &r.Description, &r.Guidance,
			&r.ParentID, &r.Depth, &r.SectionOrder, &r.IsAssessable, &r.CreatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan requirement row")
			continue
		}
		reqs = append(reqs, gin.H{
			"id":            r.ID,
			"identifier":    r.Identifier,
			"title":         r.Title,
			"description":   r.Description,
			"guidance":      r.Guidance,
			"parent_id":     r.ParentID,
			"depth":         r.Depth,
			"section_order": r.SectionOrder,
			"is_assessable": r.IsAssessable,
			"created_at":    r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, listResponse(c, reqs, total, page, perPage))
}

// listRequirementsTree returns requirements as a nested tree.
func listRequirementsTree(c *gin.Context, versionID string, assessableOnly bool) {
	query := `
		WITH RECURSIVE req_tree AS (
			SELECT id, parent_id, identifier, title, depth, section_order, is_assessable
			FROM requirements
			WHERE framework_version_id = $1 AND parent_id IS NULL
			UNION ALL
			SELECT r.id, r.parent_id, r.identifier, r.title, r.depth, r.section_order, r.is_assessable
			FROM requirements r
			JOIN req_tree rt ON r.parent_id = rt.id
		)
		SELECT * FROM req_tree ORDER BY depth, section_order
	`

	rows, err := database.Query(query, versionID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query requirement tree")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	var reqs []models.Requirement
	for rows.Next() {
		var r models.Requirement
		if err := rows.Scan(&r.ID, &r.ParentID, &r.Identifier, &r.Title,
			&r.Depth, &r.SectionOrder, &r.IsAssessable); err != nil {
			log.Error().Err(err).Msg("Failed to scan tree row")
			continue
		}
		if assessableOnly && !r.IsAssessable {
			// Still include non-assessable parents for tree structure
		}
		reqs = append(reqs, r)
	}

	tree := models.BuildRequirementTree(reqs)

	reqID, _ := c.Get("request_id")
	c.JSON(http.StatusOK, gin.H{
		"data": tree,
		"meta": gin.H{
			"request_id": reqID,
		},
	})
}
