package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck is a liveness probe — returns 200 if the process is running.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"version":   "0.1.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadyCheck is a readiness probe — returns 200 only if all dependencies are reachable.
func ReadyCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := gin.H{}
	allReady := true

	// Check PostgreSQL
	if database != nil {
		if err := database.PingContext(ctx); err != nil {
			checks["postgres"] = "error: " + err.Error()
			allReady = false
		} else {
			checks["postgres"] = "ok"
		}
	} else {
		checks["postgres"] = "error: not configured"
		allReady = false
	}

	// Check Redis
	if redisDB != nil {
		if err := redisDB.Ping(ctx).Err(); err != nil {
			checks["redis"] = "error: " + err.Error()
			allReady = false
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "error: not configured"
		allReady = false
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allReady {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":    status,
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
