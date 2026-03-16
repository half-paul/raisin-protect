package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
)

// ListWebhooks returns webhooks for a connection.
func ListWebhooks(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")

	rows, err := database.Query(`
		SELECT id, org_id, connection_id, name, description, signature_header, signature_algo,
			status, event_types, total_received, total_processed, total_errors, last_received_at,
			created_at, updated_at
		FROM integration_webhooks
		WHERE connection_id = $1 AND org_id = $2
		ORDER BY created_at DESC
	`, connID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query webhooks"))
		return
	}
	defer rows.Close()

	items := []gin.H{}
	for rows.Next() {
		var w models.IntegrationWebhook
		err := rows.Scan(
			&w.ID, &w.OrgID, &w.ConnectionID, &w.Name, &w.Description,
			&w.SignatureHeader, &w.SignatureAlgo, &w.Status,
			pq.Array(&w.EventTypes), &w.TotalReceived, &w.TotalProcessed, &w.TotalErrors,
			&w.LastReceivedAt, &w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan webhook"))
			return
		}
		items = append(items, gin.H{
			"id":               w.ID,
			"connection_id":    w.ConnectionID,
			"name":             w.Name,
			"description":      w.Description,
			"signature_header": w.SignatureHeader,
			"signature_algo":   w.SignatureAlgo,
			"status":           w.Status,
			"event_types":      w.EventTypes,
			"total_received":   w.TotalReceived,
			"total_processed":  w.TotalProcessed,
			"total_errors":     w.TotalErrors,
			"last_received_at": w.LastReceivedAt,
			"created_at":       w.CreatedAt,
			"updated_at":       w.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, successResponse(c, items))
}

// CreateWebhook creates a new webhook and returns the secret once.
func CreateWebhook(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")

	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Verify connection exists
	var connExists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM integration_connections WHERE id = $1 AND org_id = $2)", connID, orgID).Scan(&connExists)
	if err != nil || !connExists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	// Generate secure webhook secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to generate webhook secret"))
		return
	}
	secret := "whsec_" + hex.EncodeToString(secretBytes)

	sigHeader := "X-Hub-Signature-256"
	if req.SignatureHeader != nil {
		sigHeader = *req.SignatureHeader
	}
	sigAlgo := "sha256"
	if req.SignatureAlgo != nil {
		sigAlgo = *req.SignatureAlgo
	}
	eventTypes := req.EventTypes
	if eventTypes == nil {
		eventTypes = []string{}
	}

	id := uuid.New()
	_, err = database.Exec(`
		INSERT INTO integration_webhooks (id, org_id, connection_id, name, description, webhook_secret, signature_header, signature_algo, event_types)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, orgID, connID, req.Name, req.Description, secret, sigHeader, sigAlgo, pq.Array(eventTypes))

	if err != nil {
		if strings.Contains(err.Error(), "uq_webhook_connection_name") {
			c.JSON(http.StatusConflict, errorResponse("DUPLICATE", "A webhook with this name already exists for this connection"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create webhook"))
		return
	}

	idStr := id.String()
	middleware.LogAudit(c, "integration_webhook.created", "integration_webhook", &idStr, map[string]interface{}{
		"connection_id": connID,
		"name":          req.Name,
	})

	// Return secret only on creation
	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":             id,
		"webhook_secret": secret,
		"webhook_url":    "/api/v1/webhooks/receive/" + id.String(),
	}))
}

// DeleteWebhook deletes a webhook.
func DeleteWebhook(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")
	webhookID := c.Param("wid")

	res, err := database.Exec("DELETE FROM integration_webhooks WHERE id = $1 AND connection_id = $2 AND org_id = $3", webhookID, connID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to delete webhook"))
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Webhook not found"))
		return
	}

	middleware.LogAudit(c, "integration_webhook.deleted", "integration_webhook", &webhookID, nil)
	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "deleted"}))
}

// RotateWebhookSecret generates a new secret for a webhook.
func RotateWebhookSecret(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")
	webhookID := c.Param("wid")

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to generate webhook secret"))
		return
	}
	newSecret := "whsec_" + hex.EncodeToString(secretBytes)

	res, err := database.Exec(`
		UPDATE integration_webhooks SET webhook_secret = $4, updated_at = NOW()
		WHERE id = $1 AND connection_id = $2 AND org_id = $3
	`, webhookID, connID, orgID, newSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to rotate webhook secret"))
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Webhook not found"))
		return
	}

	middleware.LogAudit(c, "integration_webhook.secret_rotated", "integration_webhook", &webhookID, nil)
	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"webhook_secret": newSecret,
	}))
}
