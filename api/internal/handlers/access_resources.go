package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
)

// ListAccessResources returns a list of access resources.
func ListAccessResources(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Filters
	criticality := c.Query("criticality")
	resourceType := c.Query("type")
	department := c.Query("department")
	ownerID := c.Query("owner_id")
	isActive := c.Query("is_active")

	where := []string{"org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if criticality != "" {
		where = append(where, fmt.Sprintf("criticality = $%d", argN))
		args = append(args, criticality)
		argN++
	}
	if resourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", argN))
		args = append(args, resourceType)
		argN++
	}
	if department != "" {
		where = append(where, fmt.Sprintf("department = $%d", argN))
		args = append(args, department)
		argN++
	}
	if ownerID != "" {
		where = append(where, fmt.Sprintf("owner_id = $%d", argN))
		args = append(args, ownerID)
		argN++
	}
	if isActive != "" {
		where = append(where, fmt.Sprintf("is_active = $%d", argN))
		args = append(args, isActive == "true")
		argN++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + fmt.Sprintf("%s", join(where, " AND "))
	}

	var total int
	err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM access_resources %s", whereClause), args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count resources"))
		return
	}

	queryArgs := append(args, perPage, offset)
	rows, err := database.Query(fmt.Sprintf(`
		SELECT id, org_id, identity_provider_id, external_id, name, description, resource_type,
		       criticality, department, category, tags, owner_id, total_users, total_roles,
		       last_sync_at, url, last_reviewed_at, review_cadence, is_active, created_at, updated_at
		FROM access_resources
		%s
		ORDER BY criticality DESC, name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1), queryArgs...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query resources"))
		return
	}
	defer rows.Close()

	resources := []models.AccessResource{}
	for rows.Next() {
		var r models.AccessResource
		err := rows.Scan(
			&r.ID, &r.OrgID, &r.IdentityProviderID, &r.ExternalID, &r.Name, &r.Description, &r.ResourceType,
			&r.Criticality, &r.Department, &r.Category, pq.Array(&r.Tags), &r.OwnerID, &r.TotalUsers, &r.TotalRoles,
			&r.LastSyncAt, &r.URL, &r.LastReviewedAt, &r.ReviewCadence, &r.IsActive, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan resource"))
			return
		}
		resources = append(resources, r)
	}

	c.JSON(http.StatusOK, listResponse(c, resources, total, page, perPage))
}

// CreateAccessResource creates a new access resource.
func CreateAccessResource(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req struct {
		Name               string                     `json:"name" binding:"required"`
		Description        *string                    `json:"description"`
		ResourceType       models.ResourceType        `json:"resource_type" binding:"required"`
		Criticality        models.ResourceCriticality `json:"criticality" binding:"required"`
		Department         *string                    `json:"department"`
		Category           *string                    `json:"category"`
		Tags               []string                   `json:"tags"`
		OwnerID            *uuid.UUID                 `json:"owner_id"`
		URL                *string                    `json:"url"`
		ReviewCadence      *models.CampaignCadence    `json:"review_cadence"`
		IdentityProviderID *uuid.UUID                 `json:"identity_provider_id"`
		ExternalID         *string                    `json:"external_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	id := uuid.New()
	_, err := database.Exec(`
		INSERT INTO access_resources (
			id, org_id, name, description, resource_type, criticality, department, 
			category, tags, owner_id, url, review_cadence, identity_provider_id, external_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, id, orgID, req.Name, req.Description, req.ResourceType, req.Criticality, req.Department,
		req.Category, pq.Array(req.Tags), req.OwnerID, req.URL, req.ReviewCadence, req.IdentityProviderID, req.ExternalID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create resource"))
		return
	}

	idStr := id.String()
	middleware.LogAudit(c, "access_resource.created", "access_resource", &idStr, map[string]interface{}{
		"name": req.Name,
		"type": req.ResourceType,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{"id": id}))
}

// GetAccessResource returns a single access resource.
func GetAccessResource(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var r models.AccessResource
	err := database.QueryRow(`
		SELECT id, org_id, identity_provider_id, external_id, name, description, resource_type,
		       criticality, department, category, tags, owner_id, total_users, total_roles,
		       last_sync_at, url, last_reviewed_at, review_cadence, is_active, created_at, updated_at
		FROM access_resources
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&r.ID, &r.OrgID, &r.IdentityProviderID, &r.ExternalID, &r.Name, &r.Description, &r.ResourceType,
		&r.Criticality, &r.Department, &r.Category, pq.Array(&r.Tags), &r.OwnerID, &r.TotalUsers, &r.TotalRoles,
		&r.LastSyncAt, &r.URL, &r.LastReviewedAt, &r.ReviewCadence, &r.IsActive, &r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Access resource not found"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, r))
}

// UpdateAccessResource updates an access resource.
func UpdateAccessResource(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var req struct {
		Name          *string                     `json:"name"`
		Description   *string                     `json:"description"`
		ResourceType  *models.ResourceType        `json:"resource_type"`
		Criticality   *models.ResourceCriticality `json:"criticality"`
		Department    *string                     `json:"department"`
		Category      *string                     `json:"category"`
		Tags          []string                    `json:"tags"`
		OwnerID       *uuid.UUID                  `json:"owner_id"`
		URL           *string                     `json:"url"`
		ReviewCadence *models.CampaignCadence     `json:"review_cadence"`
		IsActive      *bool                       `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Build dynamic query
	query := "UPDATE access_resources SET updated_at = NOW()"
	args := []interface{}{id, orgID}
	argN := 3

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argN)
		args = append(args, *req.Name)
		argN++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argN)
		args = append(args, *req.Description)
		argN++
	}
	if req.ResourceType != nil {
		query += fmt.Sprintf(", resource_type = $%d", argN)
		args = append(args, *req.ResourceType)
		argN++
	}
	if req.Criticality != nil {
		query += fmt.Sprintf(", criticality = $%d", argN)
		args = append(args, *req.Criticality)
		argN++
	}
	if req.Department != nil {
		query += fmt.Sprintf(", department = $%d", argN)
		args = append(args, *req.Department)
		argN++
	}
	if req.Category != nil {
		query += fmt.Sprintf(", category = $%d", argN)
		args = append(args, *req.Category)
		argN++
	}
	if req.Tags != nil {
		query += fmt.Sprintf(", tags = $%d", argN)
		args = append(args, pq.Array(req.Tags))
		argN++
	}
	if req.OwnerID != nil {
		query += fmt.Sprintf(", owner_id = $%d", argN)
		args = append(args, *req.OwnerID)
		argN++
	}
	if req.URL != nil {
		query += fmt.Sprintf(", url = $%d", argN)
		args = append(args, *req.URL)
		argN++
	}
	if req.ReviewCadence != nil {
		query += fmt.Sprintf(", review_cadence = $%d", argN)
		args = append(args, *req.ReviewCadence)
		argN++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argN)
		args = append(args, *req.IsActive)
		argN++
	}

	query += " WHERE id = $1 AND org_id = $2"

	res, err := database.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update resource"))
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Access resource not found"))
		return
	}

	middleware.LogAudit(c, "access_resource.updated", "access_resource", &id, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "updated"}))
}

// DeleteAccessResource deletes an access resource.
func DeleteAccessResource(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	res, err := database.Exec("DELETE FROM access_resources WHERE id = $1 AND org_id = $2", id, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to delete resource"))
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Access resource not found"))
		return
	}

	middleware.LogAudit(c, "access_resource.deleted", "access_resource", &id, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "deleted"}))
}

// ListResourceUsers returns access entries for a specific resource.
func ListResourceUsers(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	resourceID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	err := database.QueryRow(`
		SELECT COUNT(*) FROM access_entries 
		WHERE org_id = $1 AND resource_id = $2
	`, orgID, resourceID).Scan(&total)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count resource users"))
		return
	}

	rows, err := database.Query(`
		SELECT id, org_id, identity_provider_id, resource_id, external_user_id, user_email,
		       user_display_name, user_department, user_title, user_manager_email, internal_user_id,
		       role_name, access_level, permissions, is_privileged, expected_role, expected_access_level,
		       has_role_drift, granted_at, last_used_at, last_login_at, status, mfa_enabled,
		       anomalies, last_sync_at, is_service_account, notes, created_at, updated_at
		FROM access_entries
		WHERE org_id = $1 AND resource_id = $2
		ORDER BY user_display_name ASC
		LIMIT $3 OFFSET $4
	`, orgID, resourceID, perPage, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query resource users"))
		return
	}
	defer rows.Close()

	entries := []models.AccessEntry{}
	for rows.Next() {
		var e models.AccessEntry
		var permissionsJSON, anomaliesJSON []byte
		err := rows.Scan(
			&e.ID, &e.OrgID, &e.IdentityProviderID, &e.ResourceID, &e.ExternalUserID, &e.UserEmail,
			&e.UserDisplayName, &e.UserDepartment, &e.UserTitle, &e.UserManagerEmail, &e.InternalUserID,
			&e.RoleName, &e.AccessLevel, &permissionsJSON, &e.IsPrivileged, &e.ExpectedRole, &e.ExpectedLevel,
			&e.HasRoleDrift, &e.GrantedAt, &e.LastUsedAt, &e.LastLoginAt, &e.Status, &e.MFAEnabled,
			&anomaliesJSON, &e.LastSyncAt, &e.IsServiceAccount, &e.Notes, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan entry"))
			return
		}
		json.Unmarshal(permissionsJSON, &e.Permissions)
		json.Unmarshal(anomaliesJSON, &e.Anomalies)
		entries = append(entries, e)
	}

	c.JSON(http.StatusOK, listResponse(c, entries, total, page, perPage))
}

// GetAccessResourceStats returns summary statistics for access resources.
func GetAccessResourceStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var stats struct {
		TotalResources      int            `json:"total_resources"`
		CriticalResources   int            `json:"critical_resources"`
		TotalAccessEntries  int            `json:"total_access_entries"`
		PrivilegedEntries   int            `json:"privileged_entries"`
		OrphanedEntries     int            `json:"orphaned_entries"`
		ResourcesByCritical map[string]int `json:"resources_by_critical"`
	}

	err := database.QueryRow(`
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE criticality = 'critical'),
			(SELECT COUNT(*) FROM access_entries WHERE org_id = $1),
			(SELECT COUNT(*) FROM access_entries WHERE org_id = $1 AND is_privileged = TRUE),
			(SELECT COUNT(*) FROM access_entries WHERE org_id = $1 AND status = 'orphaned')
		FROM access_resources
		WHERE org_id = $1
	`, orgID).Scan(&stats.TotalResources, &stats.CriticalResources, &stats.TotalAccessEntries, &stats.PrivilegedEntries, &stats.OrphanedEntries)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query stats"))
		return
	}

	rows, err2 := database.Query(`
		SELECT criticality, COUNT(*)
		FROM access_resources
		WHERE org_id = $1
		GROUP BY criticality
	`, orgID)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query resource breakdown"))
		return
	}
	defer rows.Close()

	stats.ResourcesByCritical = make(map[string]int)
	for rows.Next() {
		var crit string
		var count int
		rows.Scan(&crit, &count)
		stats.ResourcesByCritical[crit] = count
	}

	c.JSON(http.StatusOK, successResponse(c, stats))
}
