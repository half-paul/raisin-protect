package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
)

// GetConnectionHealth returns health info for a connection including recent runs and uptime stats.
func GetConnectionHealth(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")

	// Get connection basic health info
	var status, health string
	var lastHealthCheckAt *time.Time
	var lastHealthStatus *string
	var consecutiveFailures, totalRuns, successfulRuns, failedRuns int

	err := database.QueryRow(`
		SELECT status, health, last_health_check_at, last_health_status, consecutive_failures,
			total_runs, successful_runs, failed_runs
		FROM integration_connections
		WHERE id = $1 AND org_id = $2
	`, connID, orgID).Scan(&status, &health, &lastHealthCheckAt, &lastHealthStatus,
		&consecutiveFailures, &totalRuns, &successfulRuns, &failedRuns)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	// Get recent runs (last 5)
	rows, err := database.Query(`
		SELECT id, status, trigger, started_at, completed_at, duration_ms, stats, error_message
		FROM integration_runs
		WHERE connection_id = $1 AND org_id = $2
		ORDER BY created_at DESC LIMIT 5
	`, connID, orgID)

	recentRuns := []gin.H{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, runStatus, trigger string
			var startedAt, completedAt *time.Time
			var durationMs *int
			var statsJSON []byte
			var errorMsg *string
			rows.Scan(&id, &runStatus, &trigger, &startedAt, &completedAt, &durationMs, &statsJSON, &errorMsg)
			var stats map[string]interface{}
			json.Unmarshal(statsJSON, &stats)
			recentRuns = append(recentRuns, gin.H{
				"id":           id,
				"status":       runStatus,
				"trigger":      trigger,
				"started_at":   startedAt,
				"completed_at": completedAt,
				"duration_ms":  durationMs,
				"stats":        stats,
				"error_message": errorMsg,
			})
		}
	}

	// Calculate uptime percentage
	var uptimePct float64
	if totalRuns > 0 {
		uptimePct = float64(successfulRuns) / float64(totalRuns) * 100
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"status":               status,
		"health":               health,
		"last_health_check_at": lastHealthCheckAt,
		"last_health_status":   lastHealthStatus,
		"consecutive_failures": consecutiveFailures,
		"total_runs":           totalRuns,
		"successful_runs":      successfulRuns,
		"failed_runs":          failedRuns,
		"uptime_pct":           uptimePct,
		"recent_runs":          recentRuns,
	}))
}

// TriggerHealthCheck performs a health check on a connection.
func TriggerHealthCheck(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")

	// Verify connection exists
	var exists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM integration_connections WHERE id = $1 AND org_id = $2)", connID, orgID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}

	// Simulate health check — in production this would call the provider
	now := time.Now()
	newHealth := models.ConnectionHealthHealthy

	_, err = database.Exec(`
		UPDATE integration_connections SET
			last_health_check_at = $3, last_health_status = $4, health = $4,
			consecutive_failures = 0, updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, connID, orgID, now, newHealth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update health check"))
		return
	}

	middleware.LogAudit(c, "integration_health.checked", "integration_connection", &connID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"health":     newHealth,
		"checked_at": now,
		"message":    "Health check passed",
	}))
}
