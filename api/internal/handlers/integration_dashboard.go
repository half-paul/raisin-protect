package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
)

// IntegrationDashboard returns a summary of integration status across the org.
func IntegrationDashboard(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	// Connection counts by status
	var totalConns, connected, disconnected, errored, disabled int
	database.QueryRow(`
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'connected'),
			COUNT(*) FILTER (WHERE status = 'disconnected'),
			COUNT(*) FILTER (WHERE status = 'error'),
			COUNT(*) FILTER (WHERE status = 'disabled')
		FROM integration_connections WHERE org_id = $1
	`, orgID).Scan(&totalConns, &connected, &disconnected, &errored, &disabled)

	// Health counts
	var healthy, degraded, unhealthy int
	database.QueryRow(`
		SELECT COUNT(*) FILTER (WHERE health = 'healthy'),
			COUNT(*) FILTER (WHERE health = 'degraded'),
			COUNT(*) FILTER (WHERE health = 'unhealthy')
		FROM integration_connections WHERE org_id = $1 AND status = 'connected'
	`, orgID).Scan(&healthy, &degraded, &unhealthy)

	// Recent failures (last 24h)
	var recentFailures int
	database.QueryRow(`
		SELECT COUNT(*) FROM integration_runs
		WHERE org_id = $1 AND status = 'failed' AND created_at > NOW() - INTERVAL '24 hours'
	`, orgID).Scan(&recentFailures)

	// Active syncs
	var activeSyncs int
	database.QueryRow(`
		SELECT COUNT(*) FROM integration_runs
		WHERE org_id = $1 AND status IN ('pending', 'running')
	`, orgID).Scan(&activeSyncs)

	// Stale connections (no sync in 24+ hours while sync enabled)
	var staleConns int
	database.QueryRow(`
		SELECT COUNT(*) FROM integration_connections
		WHERE org_id = $1 AND sync_enabled = true AND status = 'connected'
		AND (last_sync_at IS NULL OR last_sync_at < NOW() - INTERVAL '24 hours')
	`, orgID).Scan(&staleConns)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"connections": gin.H{
			"total":        totalConns,
			"connected":    connected,
			"disconnected": disconnected,
			"error":        errored,
			"disabled":     disabled,
		},
		"health": gin.H{
			"healthy":   healthy,
			"degraded":  degraded,
			"unhealthy": unhealthy,
		},
		"recent_failures": recentFailures,
		"active_syncs":    activeSyncs,
		"stale_connections": staleConns,
	}))
}

// IntegrationSyncActivity returns sync activity grouped by hourly buckets.
func IntegrationSyncActivity(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	hours := 24 // default last 24 hours

	rows, err := database.Query(`
		SELECT date_trunc('hour', created_at) AS bucket,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed,
			COUNT(*) FILTER (WHERE status = 'partial') AS partial
		FROM integration_runs
		WHERE org_id = $1 AND created_at > NOW() - INTERVAL '1 hour' * $2
		GROUP BY bucket
		ORDER BY bucket ASC
	`, orgID, hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query sync activity"))
		return
	}
	defer rows.Close()

	buckets := []gin.H{}
	for rows.Next() {
		var bucket time.Time
		var total, completed, failed, partial int
		rows.Scan(&bucket, &total, &completed, &failed, &partial)
		buckets = append(buckets, gin.H{
			"timestamp": bucket,
			"total":     total,
			"completed": completed,
			"failed":    failed,
			"partial":   partial,
		})
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"hours":    hours,
		"buckets":  buckets,
	}))
}

// IntegrationPreview returns provider-specific preview data for a connection.
func IntegrationPreview(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")

	// Get the connection and its latest run stats
	var connName, defSlug string
	var lastSyncAt *time.Time
	var statsJSON []byte
	err := database.QueryRow(`
		SELECT ic.name, id.slug, ic.last_sync_at,
			COALESCE((SELECT stats FROM integration_runs WHERE connection_id = ic.id AND org_id = ic.org_id ORDER BY created_at DESC LIMIT 1), '{}')
		FROM integration_connections ic
		JOIN integration_definitions id ON ic.definition_id = id.id
		WHERE ic.id = $1 AND ic.org_id = $2
	`, connID, orgID).Scan(&connName, &defSlug, &lastSyncAt, &statsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	var stats map[string]interface{}
	json.Unmarshal(statsJSON, &stats)

	// Get recent log summary
	var infoCount, warnCount, errorCount int
	database.QueryRow(`
		SELECT COUNT(*) FILTER (WHERE level = 'info'),
			COUNT(*) FILTER (WHERE level = 'warn'),
			COUNT(*) FILTER (WHERE level = 'error')
		FROM integration_logs
		WHERE connection_id = $1 AND org_id = $2 AND created_at > NOW() - INTERVAL '24 hours'
	`, connID, orgID).Scan(&infoCount, &warnCount, &errorCount)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"connection_name": connName,
		"provider":        defSlug,
		"last_sync_at":    lastSyncAt,
		"latest_stats":    stats,
		"recent_logs": gin.H{
			"info":  infoCount,
			"warn":  warnCount,
			"error": errorCount,
		},
	}))
}
