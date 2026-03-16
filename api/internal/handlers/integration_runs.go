package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
)

// TriggerSync creates a new pending sync run for a connection.
func TriggerSync(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	connID := c.Param("id")

	// Verify connection exists and is connected
	var connStatus string
	err := database.QueryRow("SELECT status FROM integration_connections WHERE id = $1 AND org_id = $2", connID, orgID).Scan(&connStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Connection not found"))
		return
	}
	if connStatus != models.ConnectionStatusConnected {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Connection must be in connected state to trigger sync"))
		return
	}

	// Use transaction to atomically check for active runs and insert
	tx, err := database.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start transaction"))
		return
	}
	defer tx.Rollback()

	var activeCount int
	tx.QueryRow("SELECT COUNT(*) FROM integration_runs WHERE connection_id = $1 AND org_id = $2 AND status IN ('pending', 'running') FOR UPDATE", connID, orgID).Scan(&activeCount)
	if activeCount > 0 {
		c.JSON(http.StatusConflict, errorResponse("SYNC_IN_PROGRESS", "A sync is already in progress for this connection"))
		return
	}

	runID := uuid.New()
	_, err = tx.Exec(`
		INSERT INTO integration_runs (id, org_id, connection_id, trigger, triggered_by, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, runID, orgID, connID, models.RunTriggerManual, userID, models.RunStatusPending)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create sync run"))
		return
	}

	// Update connection run counter
	tx.Exec("UPDATE integration_connections SET total_runs = total_runs + 1, updated_at = NOW() WHERE id = $1 AND org_id = $2", connID, orgID)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to commit sync run"))
		return
	}

	idStr := connID
	middleware.LogAudit(c, "integration_sync.triggered", "integration_connection", &idStr, map[string]interface{}{
		"run_id": runID.String(),
	})

	c.JSON(http.StatusAccepted, successResponse(c, gin.H{
		"run_id": runID,
		"status": "pending",
	}))
}

// ListRuns returns sync runs for a connection.
func ListRuns(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")

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
	err := database.QueryRow("SELECT COUNT(*) FROM integration_runs WHERE connection_id = $1 AND org_id = $2", connID, orgID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count runs"))
		return
	}

	rows, err := database.Query(`
		SELECT id, org_id, connection_id, trigger, triggered_by, status, queued_at, started_at, completed_at,
			duration_ms, stats, error_message, error_details, retry_count, max_retries, created_at, updated_at
		FROM integration_runs
		WHERE connection_id = $1 AND org_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4
	`, connID, orgID, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query runs"))
		return
	}
	defer rows.Close()

	items := []models.IntegrationRun{}
	for rows.Next() {
		var r models.IntegrationRun
		var statsJSON, errDetailsJSON []byte
		err := rows.Scan(
			&r.ID, &r.OrgID, &r.ConnectionID, &r.Trigger, &r.TriggeredBy, &r.Status,
			&r.QueuedAt, &r.StartedAt, &r.CompletedAt, &r.DurationMs,
			&statsJSON, &r.ErrorMessage, &errDetailsJSON,
			&r.RetryCount, &r.MaxRetries, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan run"))
			return
		}
		json.Unmarshal(statsJSON, &r.Stats)
		json.Unmarshal(errDetailsJSON, &r.ErrorDetails)
		if r.Stats == nil {
			r.Stats = map[string]interface{}{}
		}
		items = append(items, r)
	}

	c.JSON(http.StatusOK, listResponse(c, items, total, page, perPage))
}

// GetRun returns a single run.
func GetRun(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")
	runID := c.Param("rid")

	var r models.IntegrationRun
	var statsJSON, errDetailsJSON []byte
	err := database.QueryRow(`
		SELECT id, org_id, connection_id, trigger, triggered_by, status, queued_at, started_at, completed_at,
			duration_ms, stats, error_message, error_details, retry_count, max_retries, created_at, updated_at
		FROM integration_runs
		WHERE id = $1 AND connection_id = $2 AND org_id = $3
	`, runID, connID, orgID).Scan(
		&r.ID, &r.OrgID, &r.ConnectionID, &r.Trigger, &r.TriggeredBy, &r.Status,
		&r.QueuedAt, &r.StartedAt, &r.CompletedAt, &r.DurationMs,
		&statsJSON, &r.ErrorMessage, &errDetailsJSON,
		&r.RetryCount, &r.MaxRetries, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Run not found"))
		return
	}
	json.Unmarshal(statsJSON, &r.Stats)
	json.Unmarshal(errDetailsJSON, &r.ErrorDetails)
	if r.Stats == nil {
		r.Stats = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, successResponse(c, r))
}

// CancelRun cancels a pending or running sync run.
func CancelRun(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")
	runID := c.Param("rid")

	res, err := database.Exec(`
		UPDATE integration_runs SET status = $4, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND connection_id = $2 AND org_id = $3 AND status IN ('pending', 'running')
	`, runID, connID, orgID, models.RunStatusCancelled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to cancel run"))
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Run not found or not cancellable"))
		return
	}

	middleware.LogAudit(c, "integration_sync.cancelled", "integration_run", &runID, nil)
	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "cancelled"}))
}

// GetRunLogs returns logs for a specific run.
func GetRunLogs(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	connID := c.Param("id")
	runID := c.Param("rid")

	level := c.Query("level")

	query := `SELECT id, org_id, run_id, connection_id, level, message, details, source, item_ref, created_at
		FROM integration_logs WHERE run_id = $1 AND connection_id = $2 AND org_id = $3`
	args := []interface{}{runID, connID, orgID}

	if level != "" {
		query += " AND level = $4"
		args = append(args, level)
	}
	query += " ORDER BY created_at ASC"

	rows, err := database.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query logs"))
		return
	}
	defer rows.Close()

	items := []models.IntegrationLog{}
	for rows.Next() {
		var l models.IntegrationLog
		var detailsJSON []byte
		err := rows.Scan(&l.ID, &l.OrgID, &l.RunID, &l.ConnectionID, &l.Level, &l.Message, &detailsJSON, &l.Source, &l.ItemRef, &l.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan log"))
			return
		}
		json.Unmarshal(detailsJSON, &l.Details)
		items = append(items, l)
	}

	c.JSON(http.StatusOK, successResponse(c, items))
}
