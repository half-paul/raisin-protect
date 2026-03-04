package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
)

// ListIdentityProviders returns a list of identity providers for the organization.
func ListIdentityProviders(c *gin.Context) {
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

	var total int
	err := database.QueryRow("SELECT COUNT(*) FROM identity_providers WHERE org_id = $1", orgID).Scan(&total)
	if err != nil {
		fmt.Printf("COUNT ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", fmt.Sprintf("Failed to count identity providers: %v", err)))
		return
	}

	query := "SELECT id, org_id, name, provider_type, status, config, last_sync_at, last_sync_status, last_sync_error, last_sync_stats, sync_interval_mins, description, created_at, updated_at FROM identity_providers WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"

	rows, err := database.Query(query, orgID, perPage, offset)
	if err != nil {
		fmt.Printf("QUERY ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", fmt.Sprintf("Failed to query identity providers: %v", err)))
		return
	}
	defer rows.Close()

	providers := []models.IdentityProvider{}
	for rows.Next() {
		var p models.IdentityProvider
		var configJSON, statsJSON []byte
		err := rows.Scan(
			&p.ID, &p.OrgID, &p.Name, &p.ProviderType, &p.Status, &configJSON,
			&p.LastSyncAt, &p.LastSyncStatus, &p.LastSyncError, &statsJSON,
			&p.SyncIntervalMins, &p.Description, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan identity provider"))
			return
		}
		json.Unmarshal(configJSON, &p.Config)
		json.Unmarshal(statsJSON, &p.LastSyncStats)
		providers = append(providers, p)
	}

	c.JSON(http.StatusOK, listResponse(c, providers, total, page, perPage))
}

// CreateIdentityProvider creates a new identity provider.
func CreateIdentityProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req struct {
		Name             string                 `json:"name" binding:"required"`
		ProviderType     models.IdentityProviderType `json:"provider_type" binding:"required"`
		Config           map[string]interface{} `json:"config"`
		SyncIntervalMins int                    `json:"sync_interval_mins"`
		Description      *string                `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	if req.SyncIntervalMins == 0 {
		req.SyncIntervalMins = 360 // Default 6 hours
	}

	configJSON, _ := json.Marshal(req.Config)
	if configJSON == nil {
		configJSON = []byte("{}")
	}

	id := uuid.New()
	_, err := database.Exec(`
		INSERT INTO identity_providers (id, org_id, name, provider_type, status, config, sync_interval_mins, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, orgID, req.Name, req.ProviderType, models.IdPStatusPending, configJSON, req.SyncIntervalMins, req.Description)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create identity provider"))
		return
	}

	idStr := id.String()
	middleware.LogAudit(c, "identity_provider.created", "identity_provider", &idStr, map[string]interface{}{
		"name": req.Name,
		"type": req.ProviderType,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{"id": id}))
}

// GetIdentityProvider returns a single identity provider.
func GetIdentityProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var p models.IdentityProvider
	var configJSON, statsJSON []byte
	err := database.QueryRow(`
		SELECT id, org_id, name, provider_type, status, config, last_sync_at, last_sync_status, 
		       last_sync_error, last_sync_stats, sync_interval_mins, description, created_at, updated_at
		FROM identity_providers
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.ProviderType, &p.Status, &configJSON,
		&p.LastSyncAt, &p.LastSyncStatus, &p.LastSyncError, &statsJSON,
		&p.SyncIntervalMins, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Identity provider not found"))
		return
	}

	json.Unmarshal(configJSON, &p.Config)
	json.Unmarshal(statsJSON, &p.LastSyncStats)

	c.JSON(http.StatusOK, successResponse(c, p))
}

// UpdateIdentityProvider updates an existing identity provider.
func UpdateIdentityProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var req struct {
		Name             *string                 `json:"name"`
		Config           map[string]interface{}  `json:"config"`
		SyncIntervalMins *int                    `json:"sync_interval_mins"`
		Description      *string                 `json:"description"`
		Status           *models.IdentityProviderStatus `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Build dynamic query
	query := "UPDATE identity_providers SET updated_at = NOW()"
	args := []interface{}{id, orgID}
	argN := 3

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argN)
		args = append(args, *req.Name)
		argN++
	}
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		query += fmt.Sprintf(", config = $%d", argN)
		args = append(args, configJSON)
		argN++
	}
	if req.SyncIntervalMins != nil {
		query += fmt.Sprintf(", sync_interval_mins = $%d", argN)
		args = append(args, *req.SyncIntervalMins)
		argN++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argN)
		args = append(args, *req.Description)
		argN++
	}
	if req.Status != nil {
		query += fmt.Sprintf(", status = $%d", argN)
		args = append(args, *req.Status)
		argN++
	}

	query += " WHERE id = $1 AND org_id = $2"

	res, err := database.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update identity provider"))
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Identity provider not found"))
		return
	}

	middleware.LogAudit(c, "identity_provider.updated", "identity_provider", &id, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "updated"}))
}

// DeleteIdentityProvider deletes an identity provider.
func DeleteIdentityProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	res, err := database.Exec("DELETE FROM identity_providers WHERE id = $1 AND org_id = $2", id, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to delete identity provider"))
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Identity provider not found"))
		return
	}

	middleware.LogAudit(c, "identity_provider.deleted", "identity_provider", &id, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "deleted"}))
}

// SyncIdentityProvider triggers a sync for an identity provider.
func SyncIdentityProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	// Update status to syncing
	_, err := database.Exec(`
		UPDATE identity_providers 
		SET status = $1, last_sync_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND org_id = $3
	`, models.IdPStatusSyncing, id, orgID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start sync"))
		return
	}

	middleware.LogAudit(c, "identity_provider.sync_started", "identity_provider", &id, nil)

	// In a real app, this would trigger a background worker.
	// For Sprint 8, we just mock it.

	c.JSON(http.StatusAccepted, successResponse(c, gin.H{"status": "sync_started"}))
}

// GetIdentityProviderSyncStats returns sync statistics for an identity provider.
func GetIdentityProviderSyncStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var statsJSON []byte
	var lastSyncAt *time.Time
	var lastSyncStatus *string

	err := database.QueryRow(`
		SELECT last_sync_stats, last_sync_at, last_sync_status
		FROM identity_providers
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&statsJSON, &lastSyncAt, &lastSyncStatus)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Identity provider not found"))
		return
	}

	var stats map[string]interface{}
	json.Unmarshal(statsJSON, &stats)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"stats":            stats,
		"last_sync_at":     lastSyncAt,
		"last_sync_status": lastSyncStatus,
	}))
}
