package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// AuditAccess is middleware that logs all protected endpoint access.
// This creates an automatic audit trail for compliance purposes.
// Place after AuthRequired() in the middleware chain.
func AuditAccess() gin.HandlerFunc {
	// Endpoints that should not be logged (high-frequency, low-value for audit)
	skipEndpoints := map[string]bool{
		"/api/v1/health":        true,
		"/api/v1/health/ready":  true,
		"/api/v1/health/live":   true,
		"/api/v1/me":            true, // Frequent polling
	}

	// Map HTTP method to audit action
	methodToAction := map[string]string{
		"GET":    "read",
		"POST":   "create",
		"PUT":    "update",
		"PATCH":  "update",
		"DELETE": "delete",
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Skip certain endpoints
		if skipEndpoints[path] || method == "OPTIONS" {
			c.Next()
			return
		}

		start := time.Now()

		// Process request
		c.Next()

		// Log after request completes
		status := c.Writer.Status()
		duration := time.Since(start)
		userID := GetUserID(c)
		orgID := GetOrgID(c)

		// Only log authenticated requests that weren't errors
		if userID == "" || status >= 500 {
			return
		}

		// Extract resource type from path
		// e.g., /api/v1/risks/123 -> risks
		resourceType := extractResourceType(path)
		if resourceType == "" {
			return
		}

		// Extract resource ID if present
		resourceID := extractResourceID(path)

		action := methodToAction[method]
		if action == "" {
			action = "access"
		}

		// Build metadata
		metadata := map[string]interface{}{
			"method":      method,
			"path":        path,
			"status_code": status,
			"duration_ms": duration.Milliseconds(),
			"user_agent":  c.GetHeader("User-Agent"),
		}

		// Add query params for GET requests (exclude sensitive ones)
		if method == "GET" && c.Request.URL.RawQuery != "" {
			q := c.Request.URL.Query()
			// Remove sensitive params
			q.Del("access_token")
			q.Del("token")
			q.Del("password")
			if len(q) > 0 {
				metadata["query"] = q.Encode()
			}
		}

		// Log to audit table
		auditAction := "api." + resourceType + "." + action
		LogAudit(c, auditAction, resourceType, resourceID, metadata)

		// Also log to structured logger for monitoring
		log.Debug().
			Str("org_id", orgID).
			Str("user_id", userID).
			Str("action", auditAction).
			Str("resource", resourceType).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("duration", duration).
			Msg("API access")
	}
}

// extractResourceType extracts the primary resource type from an API path.
// e.g., "/api/v1/risks/123/treatments" -> "risks"
func extractResourceType(path string) string {
	// Remove API prefix
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/")

	// Split by slash
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}

	// Skip auth endpoints as they have their own logging
	if parts[0] == "auth" {
		return ""
	}

	return parts[0]
}

// extractResourceID extracts the resource ID from an API path.
// e.g., "/api/v1/risks/abc-123/treatments" -> "abc-123"
func extractResourceID(path string) *string {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/")

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil
	}

	// The ID is typically the second part
	id := parts[1]
	// Basic UUID/ID validation (should be non-empty and not look like an action)
	if id == "" || id == "batch" || id == "search" || id == "export" {
		return nil
	}

	return &id
}
