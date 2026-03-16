package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
)

// ListConnections returns integration connections for the org.
func ListConnections(c *gin.Context) {
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

	status := c.Query("status")
	health := c.Query("health")
	category := c.Query("category")

	// Count
	countQuery := "SELECT COUNT(*) FROM integration_connections ic JOIN integration_definitions id ON ic.definition_id = id.id WHERE ic.org_id = $1"
	countArgs := []interface{}{orgID}
	argN := 2

	if status != "" {
		countQuery += fmt.Sprintf(" AND ic.status = $%d", argN)
		countArgs = append(countArgs, status)
		argN++
	}
	if health != "" {
		countQuery += fmt.Sprintf(" AND ic.health = $%d", argN)
		countArgs = append(countArgs, health)
		argN++
	}
	if category != "" {
		countQuery += fmt.Sprintf(" AND id.category = $%d", argN)
		countArgs = append(countArgs, category)
		argN++
	}

	var total int
	err := database.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count connections"))
		return
	}

	// Data
	dataQuery := `SELECT ic.id, ic.org_id, ic.definition_id, ic.name, ic.description, ic.instance_label,
		ic.status, ic.health, ic.config, ic.sync_enabled, ic.sync_interval_mins, ic.sync_cron,
		ic.next_sync_at, ic.last_sync_at, ic.last_sync_status, ic.last_sync_error,
		ic.last_health_check_at, ic.last_health_status, ic.consecutive_failures,
		ic.total_runs, ic.successful_runs, ic.failed_runs, ic.created_by,
		ic.tags, ic.metadata, ic.created_at, ic.updated_at,
		id.name, id.slug, id.category, COALESCE(id.icon_url, '')
		FROM integration_connections ic
		JOIN integration_definitions id ON ic.definition_id = id.id
		WHERE ic.org_id = $1`
	dataArgs := []interface{}{orgID}
	argN = 2

	if status != "" {
		dataQuery += fmt.Sprintf(" AND ic.status = $%d", argN)
		dataArgs = append(dataArgs, status)
		argN++
	}
	if health != "" {
		dataQuery += fmt.Sprintf(" AND ic.health = $%d", argN)
		dataArgs = append(dataArgs, health)
		argN++
	}
	if category != "" {
		dataQuery += fmt.Sprintf(" AND id.category = $%d", argN)
		dataArgs = append(dataArgs, category)
		argN++
	}

	dataQuery += fmt.Sprintf(" ORDER BY ic.created_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)
	dataArgs = append(dataArgs, perPage, offset)

	rows, err := database.Query(dataQuery, dataArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query connections"))
		return
	}
	defer rows.Close()

	items := []gin.H{}
	for rows.Next() {
		var conn models.IntegrationConnection
		var configJSON, metadataJSON []byte
		err := rows.Scan(
			&conn.ID, &conn.OrgID, &conn.DefinitionID, &conn.Name, &conn.Description, &conn.InstanceLabel,
			&conn.Status, &conn.Health, &configJSON, &conn.SyncEnabled, &conn.SyncIntervalMins, &conn.SyncCron,
			&conn.NextSyncAt, &conn.LastSyncAt, &conn.LastSyncStatus, &conn.LastSyncError,
			&conn.LastHealthCheckAt, &conn.LastHealthStatus, &conn.ConsecutiveFailures,
			&conn.TotalRuns, &conn.SuccessfulRuns, &conn.FailedRuns, &conn.CreatedBy,
			pq.Array(&conn.Tags), &metadataJSON, &conn.CreatedAt, &conn.UpdatedAt,
			&conn.DefinitionName, &conn.DefinitionSlug, &conn.DefinitionCategory, &conn.DefinitionIconURL,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan connection"))
			return
		}
		json.Unmarshal(configJSON, &conn.Config)
		json.Unmarshal(metadataJSON, &conn.Metadata)
		maskSecrets(conn.Config)
		items = append(items, connectionToH(conn))
	}

	c.JSON(http.StatusOK, listResponse(c, items, total, page, perPage))
}

// GetConnection returns a single integration connection with masked config.
func GetConnection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	conn, err := getConnectionByID(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}
	maskSecrets(conn.Config)

	c.JSON(http.StatusOK, successResponse(c, connectionToH(*conn)))
}

// CreateConnection creates a new integration connection.
func CreateConnection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Verify definition exists
	var defExists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM integration_definitions WHERE id = $1 AND is_active = true)", req.DefinitionID).Scan(&defExists)
	if err != nil || !defExists {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", "Integration definition not found or inactive"))
		return
	}

	configJSON, _ := json.Marshal(req.Config)
	if configJSON == nil {
		configJSON = []byte("{}")
	}

	syncEnabled := false
	if req.SyncEnabled != nil {
		syncEnabled = *req.SyncEnabled
	}
	syncInterval := 360
	if req.SyncIntervalMins != nil {
		syncInterval = *req.SyncIntervalMins
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	id := uuid.New()
	_, err = database.Exec(`
		INSERT INTO integration_connections (id, org_id, definition_id, name, description, instance_label, config, sync_enabled, sync_interval_mins, sync_cron, tags, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, id, orgID, req.DefinitionID, req.Name, req.Description, req.InstanceLabel, configJSON, syncEnabled, syncInterval, req.SyncCron, pq.Array(tags), userID)

	if err != nil {
		if strings.Contains(err.Error(), "uq_connection_org_name") {
			c.JSON(http.StatusConflict, errorResponse("DUPLICATE", "A connection with this name already exists"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create connection"))
		return
	}

	idStr := id.String()
	middleware.LogAudit(c, "integration_connection.created", "integration_connection", &idStr, map[string]interface{}{
		"name":          req.Name,
		"definition_id": req.DefinitionID,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{"id": id}))
}

// UpdateConnection updates an existing integration connection.
func UpdateConnection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var req models.UpdateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	query := "UPDATE integration_connections SET updated_at = NOW()"
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
	if req.InstanceLabel != nil {
		query += fmt.Sprintf(", instance_label = $%d", argN)
		args = append(args, *req.InstanceLabel)
		argN++
	}
	if req.Config != nil {
		// Merge config: fetch existing, merge new keys (skip masked values)
		existing, err := getConnectionByID(id, orgID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
			return
		}
		merged := existing.Config
		if merged == nil {
			merged = map[string]interface{}{}
		}
		for k, v := range req.Config {
			// Skip values that look like masked secrets (e.g., "****" or "****abcd")
			if s, ok := v.(string); ok && strings.HasPrefix(s, "****") {
				continue
			}
			merged[k] = v
		}
		configJSON, _ := json.Marshal(merged)
		query += fmt.Sprintf(", config = $%d", argN)
		args = append(args, configJSON)
		argN++
	}
	if req.SyncEnabled != nil {
		query += fmt.Sprintf(", sync_enabled = $%d", argN)
		args = append(args, *req.SyncEnabled)
		argN++
	}
	if req.SyncIntervalMins != nil {
		query += fmt.Sprintf(", sync_interval_mins = $%d", argN)
		args = append(args, *req.SyncIntervalMins)
		argN++
	}
	if req.SyncCron != nil {
		query += fmt.Sprintf(", sync_cron = $%d", argN)
		args = append(args, *req.SyncCron)
		argN++
	}
	if req.Tags != nil {
		query += fmt.Sprintf(", tags = $%d", argN)
		args = append(args, pq.Array(req.Tags))
		argN++
	}

	query += " WHERE id = $1 AND org_id = $2"

	res, err := database.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "uq_connection_org_name") {
			c.JSON(http.StatusConflict, errorResponse("DUPLICATE", "A connection with this name already exists"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update connection"))
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	middleware.LogAudit(c, "integration_connection.updated", "integration_connection", &id, nil)
	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "updated"}))
}

// DeleteConnection deletes an integration connection.
func DeleteConnection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	// Block deletion if there's an active sync running
	var runningCount int
	err := database.QueryRow("SELECT COUNT(*) FROM integration_runs WHERE connection_id = $1 AND org_id = $2 AND status IN ('pending', 'running')", id, orgID).Scan(&runningCount)
	if err == nil && runningCount > 0 {
		c.JSON(http.StatusConflict, errorResponse("SYNC_IN_PROGRESS", "Cannot delete connection while sync is in progress"))
		return
	}

	res, err := database.Exec("DELETE FROM integration_connections WHERE id = $1 AND org_id = $2", id, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to delete connection"))
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	middleware.LogAudit(c, "integration_connection.deleted", "integration_connection", &id, nil)
	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "deleted"}))
}

// TestConnection simulates testing an integration connection.
func TestConnection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	conn, err := getConnectionByID(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	// Simulate a connection test — in production this would call the provider
	now := time.Now()
	_, err = database.Exec(`
		UPDATE integration_connections SET
			last_health_check_at = $3, last_health_status = $4, health = $4, updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, id, orgID, now, models.ConnectionHealthHealthy)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update health check"))
		return
	}

	middleware.LogAudit(c, "integration_connection.tested", "integration_connection", &id, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"status":    "success",
		"message":   "Connection test successful",
		"provider":  conn.DefinitionSlug,
		"tested_at": now,
	}))
}

// EnableConnection sets a connection to connected status.
func EnableConnection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	res, err := database.Exec(`
		UPDATE integration_connections SET status = $3, updated_at = NOW()
		WHERE id = $1 AND org_id = $2 AND status != 'connected'
	`, id, orgID, models.ConnectionStatusConnected)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to enable connection"))
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found or already enabled"))
		return
	}

	middleware.LogAudit(c, "integration_connection.enabled", "integration_connection", &id, nil)
	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "enabled"}))
}

// DisableConnection sets a connection to disabled status.
func DisableConnection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	res, err := database.Exec(`
		UPDATE integration_connections SET status = $3, sync_enabled = false, updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, id, orgID, models.ConnectionStatusDisabled)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to disable connection"))
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	middleware.LogAudit(c, "integration_connection.disabled", "integration_connection", &id, nil)
	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "disabled"}))
}

// Helper: get connection by ID and org
func getConnectionByID(id, orgID string) (*models.IntegrationConnection, error) {
	var conn models.IntegrationConnection
	var configJSON, metadataJSON []byte

	err := database.QueryRow(`
		SELECT ic.id, ic.org_id, ic.definition_id, ic.name, ic.description, ic.instance_label,
			ic.status, ic.health, ic.config, ic.sync_enabled, ic.sync_interval_mins, ic.sync_cron,
			ic.next_sync_at, ic.last_sync_at, ic.last_sync_status, ic.last_sync_error,
			ic.last_health_check_at, ic.last_health_status, ic.consecutive_failures,
			ic.total_runs, ic.successful_runs, ic.failed_runs, ic.created_by,
			ic.tags, ic.metadata, ic.created_at, ic.updated_at,
			id.name, id.slug, id.category, COALESCE(id.icon_url, '')
		FROM integration_connections ic
		JOIN integration_definitions id ON ic.definition_id = id.id
		WHERE ic.id = $1 AND ic.org_id = $2
	`, id, orgID).Scan(
		&conn.ID, &conn.OrgID, &conn.DefinitionID, &conn.Name, &conn.Description, &conn.InstanceLabel,
		&conn.Status, &conn.Health, &configJSON, &conn.SyncEnabled, &conn.SyncIntervalMins, &conn.SyncCron,
		&conn.NextSyncAt, &conn.LastSyncAt, &conn.LastSyncStatus, &conn.LastSyncError,
		&conn.LastHealthCheckAt, &conn.LastHealthStatus, &conn.ConsecutiveFailures,
		&conn.TotalRuns, &conn.SuccessfulRuns, &conn.FailedRuns, &conn.CreatedBy,
		pq.Array(&conn.Tags), &metadataJSON, &conn.CreatedAt, &conn.UpdatedAt,
		&conn.DefinitionName, &conn.DefinitionSlug, &conn.DefinitionCategory, &conn.DefinitionIconURL,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configJSON, &conn.Config)
	json.Unmarshal(metadataJSON, &conn.Metadata)
	return &conn, nil
}

// Helper: mask secret values in config (recurses into nested objects)
func maskSecrets(config map[string]interface{}) {
	secretKeys := []string{"secret", "token", "password", "api_key", "access_key", "secret_access_key", "api_token", "bot_token", "access_token"}
	for k, v := range config {
		// Recurse into nested objects
		if nested, ok := v.(map[string]interface{}); ok {
			maskSecrets(nested)
			continue
		}
		for _, sk := range secretKeys {
			if strings.Contains(strings.ToLower(k), sk) {
				if s, ok := v.(string); ok && len(s) > 20 {
					config[k] = "****" + s[len(s)-4:]
				} else {
					config[k] = "****"
				}
			}
		}
	}
}

// Helper: convert connection to gin.H
func connectionToH(conn models.IntegrationConnection) gin.H {
	return gin.H{
		"id":                    conn.ID,
		"org_id":                conn.OrgID,
		"definition_id":         conn.DefinitionID,
		"name":                  conn.Name,
		"description":           conn.Description,
		"instance_label":        conn.InstanceLabel,
		"status":                conn.Status,
		"health":                conn.Health,
		"config":                conn.Config,
		"sync_enabled":          conn.SyncEnabled,
		"sync_interval_mins":    conn.SyncIntervalMins,
		"sync_cron":             conn.SyncCron,
		"next_sync_at":          conn.NextSyncAt,
		"last_sync_at":          conn.LastSyncAt,
		"last_sync_status":      conn.LastSyncStatus,
		"last_sync_error":       conn.LastSyncError,
		"last_health_check_at":  conn.LastHealthCheckAt,
		"last_health_status":    conn.LastHealthStatus,
		"consecutive_failures":  conn.ConsecutiveFailures,
		"total_runs":            conn.TotalRuns,
		"successful_runs":       conn.SuccessfulRuns,
		"failed_runs":           conn.FailedRuns,
		"created_by":            conn.CreatedBy,
		"tags":                  conn.Tags,
		"metadata":              conn.Metadata,
		"created_at":            conn.CreatedAt,
		"updated_at":            conn.UpdatedAt,
		"definition_name":       conn.DefinitionName,
		"definition_slug":       conn.DefinitionSlug,
		"definition_category":   conn.DefinitionCategory,
		"definition_icon_url":   conn.DefinitionIconURL,
	}
}
