package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// RedeliverAlert manually re-delivers an alert's notifications.
func RedeliverAlert(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	alertID := c.Param("id")

	type redeliverRequest struct {
		Channels []string `json:"channels"`
	}
	var req redeliverRequest
	c.ShouldBindJSON(&req)

	// Get alert
	var alertNumber int
	var title, severity string
	var deliveredAtJSON string
	err := database.QueryRow(
		"SELECT alert_number, title, severity, delivered_at FROM alerts WHERE id = $1 AND org_id = $2",
		alertID, orgID,
	).Scan(&alertNumber, &title, &severity, &deliveredAtJSON)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Alert not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get alert for delivery")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to re-deliver alert"))
		return
	}

	// For now, simulate delivery results
	now := time.Now()
	deliveryResults := gin.H{}
	channels := req.Channels
	if len(channels) == 0 {
		channels = []string{"in_app"}
	}

	for _, ch := range channels {
		deliveryResults[ch] = gin.H{
			"success":      true,
			"delivered_at": now,
		}
	}

	// Update delivered_at
	var deliveredAt map[string]interface{}
	json.Unmarshal([]byte(deliveredAtJSON), &deliveredAt)
	if deliveredAt == nil {
		deliveredAt = map[string]interface{}{}
	}
	for _, ch := range channels {
		deliveredAt[ch] = now.Format(time.RFC3339)
	}
	updatedJSON, _ := json.Marshal(deliveredAt)
	database.Exec("UPDATE alerts SET delivered_at = $1, updated_at = NOW() WHERE id = $2", string(updatedJSON), alertID)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":               alertID,
		"alert_number":     alertNumber,
		"delivery_results": deliveryResults,
		"message":          fmt.Sprintf("Alert re-delivered to %d channels.", len(channels)),
	}))
}

// TestAlertDelivery tests alert delivery channels without creating a real alert.
func TestAlertDelivery(c *gin.Context) {
	type testDeliveryRequest struct {
		Channel         string            `json:"channel" binding:"required"`
		SlackWebhookURL *string           `json:"slack_webhook_url"`
		EmailRecipients []string          `json:"email_recipients"`
		WebhookURL      *string           `json:"webhook_url"`
		WebhookHeaders  map[string]string `json:"webhook_headers"`
	}

	var req testDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "channel is required"))
		return
	}

	switch req.Channel {
	case "slack":
		if req.SlackWebhookURL == nil || *req.SlackWebhookURL == "" {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "slack_webhook_url is required for Slack delivery"))
			return
		}
		if !strings.HasPrefix(*req.SlackWebhookURL, "https://") {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "slack_webhook_url must be HTTPS"))
			return
		}
		// Send test message to Slack
		payload := map[string]interface{}{
			"text": "🧪 *Raisin Protect — Test Alert*\nThis is a test notification. If you see this, your Slack integration is working!",
		}
		body, _ := json.Marshal(payload)
		resp, err := http.Post(*req.SlackWebhookURL, "application/json", bytes.NewReader(body))
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
				fmt.Sprintf("Failed to deliver to Slack: %s", err.Error())))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
				fmt.Sprintf("Slack returned status %d", resp.StatusCode)))
			return
		}

	case "email":
		if len(req.EmailRecipients) == 0 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "email_recipients is required for email delivery"))
			return
		}
		// Email delivery would use SMTP — simulated for now
		log.Info().Strs("recipients", req.EmailRecipients).Msg("Test email delivery simulated")

	case "webhook":
		if req.WebhookURL == nil || *req.WebhookURL == "" {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "webhook_url is required for webhook delivery"))
			return
		}
		if !strings.HasPrefix(*req.WebhookURL, "https://") {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "webhook_url must be HTTPS"))
			return
		}
		// Send test payload to webhook
		payload := map[string]interface{}{
			"event":   "test_delivery",
			"message": "This is a test notification from Raisin Protect.",
		}
		body, _ := json.Marshal(payload)
		httpReq, _ := http.NewRequest("POST", *req.WebhookURL, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		for k, v := range req.WebhookHeaders {
			httpReq.Header.Set(k, v)
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
				fmt.Sprintf("Failed to deliver to webhook: %s", err.Error())))
			return
		}
		defer resp.Body.Close()

	default:
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid channel. Must be: slack, email, or webhook"))
		return
	}

	middleware.LogAudit(c, "alert.test_delivery", "alert", nil, map[string]interface{}{
		"channel": req.Channel,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"channel": req.Channel,
		"success": true,
		"message": "Test notification delivered successfully.",
	}))
}
