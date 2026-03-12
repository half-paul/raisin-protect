package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ReceiveWebhook handles inbound webhook events (public, no JWT — HMAC auth).
func ReceiveWebhook(c *gin.Context) {
	webhookID := c.Param("id")

	// Look up webhook to get secret and connection info
	var secret, sigHeader, sigAlgo, status, connID, orgID string
	err := database.QueryRow(`
		SELECT webhook_secret, signature_header, signature_algo, status, connection_id, org_id
		FROM integration_webhooks WHERE id = $1
	`, webhookID).Scan(&secret, &sigHeader, &sigAlgo, &status, &connID, &orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	if status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "webhook is not active"})
		return
	}

	// Read body (limit to 1MB to prevent DoS)
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Verify HMAC signature
	providedSig := c.GetHeader(sigHeader)
	if providedSig == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature header"})
		return
	}

	// Strip algorithm prefix if present (e.g., "sha256=abc123")
	providedSig = strings.TrimPrefix(providedSig, sigAlgo+"=")

	if !verifyHMAC(body, providedSig, secret, sigAlgo) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Update webhook counters
	now := time.Now()
	database.Exec(`
		UPDATE integration_webhooks SET
			total_received = total_received + 1, total_processed = total_processed + 1,
			last_received_at = $2, updated_at = NOW()
		WHERE id = $1 AND org_id = $3
	`, webhookID, now, orgID)

	c.JSON(http.StatusOK, gin.H{
		"status":      "received",
		"webhook_id":  webhookID,
		"received_at": now,
	})
}

func verifyHMAC(body []byte, providedSig, secret, algo string) bool {
	var mac []byte
	switch algo {
	case "sha256":
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(body)
		mac = h.Sum(nil)
	default:
		return false
	}

	expectedSig := hex.EncodeToString(mac)
	return hmac.Equal([]byte(expectedSig), []byte(providedSig))
}
