package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// CreateTestRun triggers a manual test run.
func CreateTestRun(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreateTestRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body for full sweep
		req = models.CreateTestRunRequest{}
	}

	if len(req.TestIDs) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Maximum 500 test IDs per request"))
		return
	}

	// Check no run already in progress
	var existingCount int
	err := database.QueryRow(
		"SELECT COUNT(*) FROM test_runs WHERE org_id = $1 AND status IN ('pending', 'running')",
		orgID,
	).Scan(&existingCount)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing runs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create test run"))
		return
	}
	if existingCount > 0 {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "A test run is already in progress for this organization"))
		return
	}

	// Count tests to run
	var totalTests int
	if len(req.TestIDs) > 0 {
		// Validate all test IDs exist and belong to org
		for _, tid := range req.TestIDs {
			var testStatus string
			err := database.QueryRow(
				"SELECT status FROM tests WHERE id = $1 AND org_id = $2",
				tid, orgID,
			).Scan(&testStatus)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
					fmt.Sprintf("Test %s not found", tid)))
				return
			}
			if err != nil {
				log.Error().Err(err).Msg("Failed to validate test ID")
				c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create test run"))
				return
			}
		}
		totalTests = len(req.TestIDs)
	} else {
		err := database.QueryRow(
			"SELECT COUNT(*) FROM tests WHERE org_id = $1 AND status = 'active'",
			orgID,
		).Scan(&totalTests)
		if err != nil {
			log.Error().Err(err).Msg("Failed to count active tests")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create test run"))
			return
		}
	}

	if totalTests == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "No tests to run"))
		return
	}

	// Marshal trigger metadata
	metadataJSON := "{}"
	if req.TriggerMetadata != nil {
		b, _ := json.Marshal(req.TriggerMetadata)
		metadataJSON = string(b)
	}

	id := uuid.New().String()
	now := time.Now()

	var runNumber int
	err = database.QueryRow(`
		INSERT INTO test_runs (id, org_id, status, trigger_type, total_tests, triggered_by, trigger_metadata, created_at, updated_at)
		VALUES ($1, $2, 'pending', 'manual', $3, $4, $5, $6, $6)
		RETURNING run_number
	`, id, orgID, totalTests, userID, metadataJSON, now).Scan(&runNumber)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert test run")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create test run"))
		return
	}

	middleware.LogAudit(c, "test_run.started", "test_run", &id, map[string]interface{}{
		"run_number": runNumber,
		"trigger":    "manual",
		"test_count": totalTests,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":          id,
		"run_number":  runNumber,
		"status":      "pending",
		"trigger_type": "manual",
		"total_tests": totalTests,
		"triggered_by": gin.H{
			"id": userID,
		},
		"created_at": now,
	}))
}

// ListTestRuns lists test runs with filtering and pagination.
func ListTestRuns(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"tr.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("tr.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("trigger_type"); v != "" {
		where = append(where, fmt.Sprintf("tr.trigger_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_from"); v != "" {
		where = append(where, fmt.Sprintf("tr.created_at >= $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_to"); v != "" {
		where = append(where, fmt.Sprintf("tr.created_at <= $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM test_runs tr WHERE %s", whereClause)
	if err := database.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count test runs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list test runs"))
		return
	}

	sortField := "tr.created_at"
	switch c.DefaultQuery("sort", "created_at") {
	case "run_number":
		sortField = "tr.run_number"
	case "status":
		sortField = "tr.status"
	case "trigger_type":
		sortField = "tr.trigger_type"
	case "started_at":
		sortField = "tr.started_at"
	case "total_tests":
		sortField = "tr.total_tests"
	}
	order := "DESC"
	if strings.ToLower(c.DefaultQuery("order", "desc")) == "asc" {
		order = "ASC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT tr.id, tr.run_number, tr.status, tr.trigger_type,
			tr.started_at, tr.completed_at, tr.duration_ms,
			tr.total_tests, tr.passed, tr.failed, tr.errors, tr.skipped, tr.warnings,
			tr.triggered_by, tr.created_at
		FROM test_runs tr
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortField, order, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query test runs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list test runs"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, status, triggerType string
			runNumber, totalTests, passed, failed, errors, skipped, warnings int
			startedAt, completedAt *time.Time
			durationMs             *int
			triggeredBy            *string
			createdAt              time.Time
		)
		if err := rows.Scan(
			&id, &runNumber, &status, &triggerType,
			&startedAt, &completedAt, &durationMs,
			&totalTests, &passed, &failed, &errors, &skipped, &warnings,
			&triggeredBy, &createdAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan test run row")
			continue
		}

		results = append(results, gin.H{
			"id":           id,
			"run_number":   runNumber,
			"status":       status,
			"trigger_type": triggerType,
			"started_at":   startedAt,
			"completed_at": completedAt,
			"duration_ms":  durationMs,
			"total_tests":  totalTests,
			"passed":       passed,
			"failed":       failed,
			"errors":       errors,
			"skipped":      skipped,
			"warnings":     warnings,
			"triggered_by": triggeredBy,
			"created_at":   createdAt,
		})
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// GetTestRun gets a single test run with full details.
func GetTestRun(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	runID := c.Param("id")

	var (
		id, status, triggerType string
		runNumber, totalTests, passed, failed, errorsCount, skipped, warnings int
		startedAt, completedAt *time.Time
		durationMs             *int
		triggeredBy            *string
		triggerMetadata        string
		workerID               *string
		errorMessage           *string
		createdAt, updatedAt   time.Time
	)

	err := database.QueryRow(`
		SELECT tr.id, tr.run_number, tr.status, tr.trigger_type,
			tr.started_at, tr.completed_at, tr.duration_ms,
			tr.total_tests, tr.passed, tr.failed, tr.errors, tr.skipped, tr.warnings,
			tr.triggered_by, tr.trigger_metadata, tr.worker_id, tr.error_message,
			tr.created_at, tr.updated_at
		FROM test_runs tr
		WHERE tr.id = $1 AND tr.org_id = $2
	`, runID, orgID).Scan(
		&id, &runNumber, &status, &triggerType,
		&startedAt, &completedAt, &durationMs,
		&totalTests, &passed, &failed, &errorsCount, &skipped, &warnings,
		&triggeredBy, &triggerMetadata, &workerID, &errorMessage,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test run not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get test run")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get test run"))
		return
	}

	var metaObj interface{}
	json.Unmarshal([]byte(triggerMetadata), &metaObj)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":               id,
		"run_number":       runNumber,
		"status":           status,
		"trigger_type":     triggerType,
		"started_at":       startedAt,
		"completed_at":     completedAt,
		"duration_ms":      durationMs,
		"total_tests":      totalTests,
		"passed":           passed,
		"failed":           failed,
		"errors":           errorsCount,
		"skipped":          skipped,
		"warnings":         warnings,
		"triggered_by":     triggeredBy,
		"trigger_metadata": metaObj,
		"worker_id":        workerID,
		"error_message":    errorMessage,
		"created_at":       createdAt,
		"updated_at":       updatedAt,
	}))
}

// CancelTestRun cancels a pending or running test run.
func CancelTestRun(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	runID := c.Param("id")

	var currentStatus string
	var runNumber int
	err := database.QueryRow(
		"SELECT status, run_number FROM test_runs WHERE id = $1 AND org_id = $2",
		runID, orgID,
	).Scan(&currentStatus, &runNumber)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test run not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get test run")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to cancel test run"))
		return
	}

	if currentStatus != "pending" && currentStatus != "running" {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
			fmt.Sprintf("Cannot cancel run in '%s' status", currentStatus)))
		return
	}

	_, err = database.Exec(
		"UPDATE test_runs SET status = 'cancelled', completed_at = NOW(), updated_at = NOW() WHERE id = $1",
		runID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cancel test run")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to cancel test run"))
		return
	}

	middleware.LogAudit(c, "test_run.cancelled", "test_run", &runID, map[string]interface{}{
		"run_number": runNumber,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":              runID,
		"status":          "cancelled",
		"previous_status": currentStatus,
		"message":         "Test run cancelled.",
	}))
}

// ListTestRunResults lists individual test results within a test run.
func ListTestRunResults(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	runID := c.Param("id")

	// Verify run exists
	var exists bool
	if err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM test_runs WHERE id = $1 AND org_id = $2)", runID, orgID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test run not found"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	where := []string{"r.test_run_id = $1", "r.org_id = $2"}
	args := []interface{}{runID, orgID}
	argN := 3

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("severity"); v != "" {
		where = append(where, fmt.Sprintf("r.severity = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM test_results r WHERE %s", whereClause)
	if err := database.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count results")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}

	sortField := "r.status"
	switch c.DefaultQuery("sort", "status") {
	case "severity":
		sortField = "r.severity"
	case "test_identifier":
		sortField = "t.identifier"
	case "duration_ms":
		sortField = "r.duration_ms"
	case "started_at":
		sortField = "r.started_at"
	}
	order := "ASC"
	if strings.ToLower(c.DefaultQuery("order", "asc")) == "desc" {
		order = "DESC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.test_id, t.identifier, t.title, t.test_type,
			r.control_id, c.identifier, c.title,
			r.status, r.severity, r.message, r.details, r.duration_ms,
			r.alert_generated, r.alert_id,
			r.started_at, r.completed_at, r.created_at
		FROM test_results r
		LEFT JOIN tests t ON t.id = r.test_id
		LEFT JOIN controls c ON c.id = r.control_id
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortField, order, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query results")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			rID, testID, testIdentifier, testTitle, testType string
			controlID, controlIdentifier, controlTitle       string
			rStatus, rSeverity                               string
			message                                          *string
			details                                          string
			durationMs                                       *int
			alertGenerated                                   bool
			alertID                                          *string
			startedAt                                        time.Time
			completedAt                                      *time.Time
			createdAt                                        time.Time
		)
		if err := rows.Scan(
			&rID, &testID, &testIdentifier, &testTitle, &testType,
			&controlID, &controlIdentifier, &controlTitle,
			&rStatus, &rSeverity, &message, &details, &durationMs,
			&alertGenerated, &alertID,
			&startedAt, &completedAt, &createdAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan result row")
			continue
		}

		var detailsObj interface{}
		json.Unmarshal([]byte(details), &detailsObj)

		results = append(results, gin.H{
			"id": rID,
			"test": gin.H{
				"id":         testID,
				"identifier": testIdentifier,
				"title":      testTitle,
				"test_type":  testType,
			},
			"control": gin.H{
				"id":         controlID,
				"identifier": controlIdentifier,
				"title":      controlTitle,
			},
			"status":          rStatus,
			"severity":        rSeverity,
			"message":         message,
			"details":         detailsObj,
			"duration_ms":     durationMs,
			"alert_generated": alertGenerated,
			"alert_id":        alertID,
			"started_at":      startedAt,
			"completed_at":    completedAt,
			"created_at":      createdAt,
		})
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// GetTestRunResult gets a single test result with full details.
func GetTestRunResult(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	runID := c.Param("id")
	resultID := c.Param("rid")

	var (
		rID, testRunID, testID, testIdentifier, testTitle, testType, testSeverity string
		controlID, controlIdentifier, controlTitle                                 string
		rStatus, rSeverity                                                         string
		message                                                                    *string
		details                                                                    string
		outputLog, errorMessage                                                    *string
		durationMs                                                                 *int
		alertGenerated                                                             bool
		alertID                                                                    *string
		startedAt                                                                  time.Time
		completedAt                                                                *time.Time
		createdAt                                                                  time.Time
	)

	err := database.QueryRow(`
		SELECT r.id, r.test_run_id, r.test_id, t.identifier, t.title, t.test_type, t.severity,
			r.control_id, c.identifier, c.title,
			r.status, r.severity, r.message, r.details, r.output_log, r.error_message,
			r.duration_ms, r.alert_generated, r.alert_id,
			r.started_at, r.completed_at, r.created_at
		FROM test_results r
		LEFT JOIN tests t ON t.id = r.test_id
		LEFT JOIN controls c ON c.id = r.control_id
		WHERE r.id = $1 AND r.test_run_id = $2 AND r.org_id = $3
	`, resultID, runID, orgID).Scan(
		&rID, &testRunID, &testID, &testIdentifier, &testTitle, &testType, &testSeverity,
		&controlID, &controlIdentifier, &controlTitle,
		&rStatus, &rSeverity, &message, &details, &outputLog, &errorMessage,
		&durationMs, &alertGenerated, &alertID,
		&startedAt, &completedAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test result not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get test result")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get test result"))
		return
	}

	var detailsObj interface{}
	json.Unmarshal([]byte(details), &detailsObj)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":          rID,
		"test_run_id": testRunID,
		"test": gin.H{
			"id":         testID,
			"identifier": testIdentifier,
			"title":      testTitle,
			"test_type":  testType,
			"severity":   testSeverity,
		},
		"control": gin.H{
			"id":         controlID,
			"identifier": controlIdentifier,
			"title":      controlTitle,
		},
		"status":          rStatus,
		"severity":        rSeverity,
		"message":         message,
		"details":         detailsObj,
		"output_log":      outputLog,
		"error_message":   errorMessage,
		"duration_ms":     durationMs,
		"alert_generated": alertGenerated,
		"alert_id":        alertID,
		"started_at":      startedAt,
		"completed_at":    completedAt,
		"created_at":      createdAt,
	}))
}
