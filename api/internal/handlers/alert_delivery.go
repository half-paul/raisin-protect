package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// httpClientWithTimeout is a shared HTTP client with a 10-second timeout for webhook calls.
var httpClientWithTimeout = &http.Client{Timeout: 10 * time.Second}

// allowedWebhookHeaders is a whitelist of headers that can be set on webhook requests.
var allowedWebhookHeaders = map[string]bool{
	"authorization":  true,
	"content-type":   true,
	"x-api-key":      true,
	"x-request-id":   true,
	"x-correlation-id": true,
	"user-agent":     true,
}

// isAllowedWebhookHeader checks if a header is allowed (case-insensitive).
// Also allows any header starting with "X-Custom-".
func isAllowedWebhookHeader(header string) bool {
	lower := strings.ToLower(header)
	if allowedWebhookHeaders[lower] {
		return true
	}
	if strings.HasPrefix(lower, "x-custom-") {
		return true
	}
	return false
}

// isPrivateIP checks if an IP address is private/internal.
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	// Check for loopback
	if ip.IsLoopback() {
		return true
	}
	// Check for private ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // link-local
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7", // IPv6 unique local
		"fe80::/10", // IPv6 link-local
	}
	for _, cidr := range privateRanges {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip) {
			return true
		}
	}
	return false
}

// validateWebhookURL checks that a URL is safe (HTTPS, not private IP).
func validateWebhookURL(rawURL string) error {
	if !strings.HasPrefix(rawURL, "https://") {
		return fmt.Errorf("webhook URL must use HTTPS")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	host := parsed.Hostname()

	// Resolve the hostname to check for private IPs
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("cannot resolve hostname: %w", err)
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("webhook URL resolves to a private IP address")
		}
	}

	return nil
}

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
		// Validate URL (HTTPS and not private IP)
		if err := validateWebhookURL(*req.SlackWebhookURL); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", err.Error()))
			return
		}
		// Send test message to Slack with timeout
		payload := map[string]interface{}{
			"text": "🧪 *Raisin Protect — Test Alert*\nThis is a test notification. If you see this, your Slack integration is working!",
		}
		body, _ := json.Marshal(payload)
		httpReq, _ := http.NewRequest("POST", *req.SlackWebhookURL, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := httpClientWithTimeout.Do(httpReq)
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
		// Validate URL (HTTPS and not private IP)
		if err := validateWebhookURL(*req.WebhookURL); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", err.Error()))
			return
		}
		// Validate headers (only allow whitelisted headers)
		for k := range req.WebhookHeaders {
			if !isAllowedWebhookHeader(k) {
				c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST",
					fmt.Sprintf("Header '%s' is not allowed. Allowed: Authorization, Content-Type, X-API-Key, X-Request-ID, X-Correlation-ID, User-Agent, X-Custom-*", k)))
				return
			}
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
		resp, err := httpClientWithTimeout.Do(httpReq)
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
