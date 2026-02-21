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
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// ListTests lists test definitions with filtering and pagination.
func ListTests(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"t.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("t.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("test_type"); v != "" {
		where = append(where, fmt.Sprintf("t.test_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("severity"); v != "" {
		where = append(where, fmt.Sprintf("t.severity = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("control_id"); v != "" {
		where = append(where, fmt.Sprintf("t.control_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("tags"); v != "" {
		tags := strings.Split(v, ",")
		where = append(where, fmt.Sprintf("t.tags @> $%d", argN))
		args = append(args, pq.Array(tags))
		argN++
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(t.title ILIKE $%d OR t.description ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tests t WHERE %s", whereClause)
	if err := database.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count tests")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to count tests"))
		return
	}

	// Sort
	sortField := "t.identifier"
	switch c.DefaultQuery("sort", "identifier") {
	case "title":
		sortField = "t.title"
	case "test_type":
		sortField = "t.test_type"
	case "severity":
		sortField = "t.severity"
	case "status":
		sortField = "t.status"
	case "last_run_at":
		sortField = "t.last_run_at"
	case "created_at":
		sortField = "t.created_at"
	}
	order := "ASC"
	if strings.ToLower(c.DefaultQuery("order", "asc")) == "desc" {
		order = "DESC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT t.id, t.identifier, t.title, t.description, t.test_type, t.severity, t.status,
			t.control_id, c.identifier, c.title,
			t.schedule_cron, t.schedule_interval_min, t.next_run_at, t.last_run_at,
			t.tags, t.created_at, t.updated_at
		FROM tests t
		LEFT JOIN controls c ON c.id = t.control_id
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortField, order, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query tests")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list tests"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, identifier, title, testType, severity, status, controlID string
			description                                                   *string
			controlIdentifier, controlTitle                               string
			scheduleCron                                                  *string
			scheduleIntervalMin                                           *int
			nextRunAt, lastRunAt                                          *time.Time
			tags                                                          pq.StringArray
			createdAt, updatedAt                                          time.Time
		)
		if err := rows.Scan(
			&id, &identifier, &title, &description, &testType, &severity, &status,
			&controlID, &controlIdentifier, &controlTitle,
			&scheduleCron, &scheduleIntervalMin, &nextRunAt, &lastRunAt,
			&tags, &createdAt, &updatedAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan test row")
			continue
		}

		item := gin.H{
			"id":                    id,
			"identifier":            identifier,
			"title":                 title,
			"description":           description,
			"test_type":             testType,
			"severity":              severity,
			"status":                status,
			"control": gin.H{
				"id":         controlID,
				"identifier": controlIdentifier,
				"title":      controlTitle,
			},
			"schedule_cron":         scheduleCron,
			"schedule_interval_min": scheduleIntervalMin,
			"next_run_at":           nextRunAt,
			"last_run_at":           lastRunAt,
			"tags":                  tags,
			"created_at":            createdAt,
			"updated_at":            updatedAt,
		}
		results = append(results, item)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// CreateTest creates a new test definition.
func CreateTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid request body"))
		return
	}

	// Validate test_type
	if !models.IsValidTestType(req.TestType) {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid test_type"))
		return
	}

	// Validate identifier length
	if len(req.Identifier) > 50 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Identifier must be 50 characters or less"))
		return
	}
	if len(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Title must be 500 characters or less"))
		return
	}

	// Validate severity
	severity := "medium"
	if req.Severity != nil {
		if !models.IsValidTestSeverity(*req.Severity) {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid severity"))
			return
		}
		severity = *req.Severity
	}

	// Validate schedule mutual exclusivity
	if req.ScheduleCron != nil && req.ScheduleIntervalMin != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "schedule_cron and schedule_interval_min are mutually exclusive"))
		return
	}
	if req.ScheduleIntervalMin != nil && (*req.ScheduleIntervalMin < 1 || *req.ScheduleIntervalMin > 10080) {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "schedule_interval_min must be between 1 and 10080"))
		return
	}

	// Validate timeout
	timeoutSeconds := 300
	if req.TimeoutSeconds != nil {
		if *req.TimeoutSeconds < 1 || *req.TimeoutSeconds > 3600 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "timeout_seconds must be between 1 and 3600"))
			return
		}
		timeoutSeconds = *req.TimeoutSeconds
	}

	retryCount := 0
	if req.RetryCount != nil {
		if *req.RetryCount < 0 || *req.RetryCount > 5 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "retry_count must be between 0 and 5"))
			return
		}
		retryCount = *req.RetryCount
	}

	retryDelay := 60
	if req.RetryDelaySeconds != nil {
		if *req.RetryDelaySeconds < 1 || *req.RetryDelaySeconds > 3600 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "retry_delay_seconds must be between 1 and 3600"))
			return
		}
		retryDelay = *req.RetryDelaySeconds
	}

	if len(req.Tags) > 20 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Maximum 20 tags allowed"))
		return
	}

	// Check identifier uniqueness
	var exists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM tests WHERE org_id = $1 AND identifier = $2)", orgID, req.Identifier).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check test identifier")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create test"))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Test identifier already exists in this organization"))
		return
	}

	// Check control exists and is active
	var controlIdentifier, controlTitle, controlStatus string
	err = database.QueryRow(
		"SELECT identifier, title, status FROM controls WHERE id = $1 AND org_id = $2",
		req.ControlID, orgID,
	).Scan(&controlIdentifier, &controlTitle, &controlStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Control not found in this organization"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create test"))
		return
	}
	if controlStatus != "active" {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Control must be in active status"))
		return
	}

	// Marshal test_config
	configJSON := "{}"
	if req.TestConfig != nil {
		b, err := json.Marshal(req.TestConfig)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid test_config"))
			return
		}
		configJSON = string(b)
	}

	id := uuid.New().String()
	now := time.Now()

	_, err = database.Exec(`
		INSERT INTO tests (id, org_id, identifier, title, description, test_type, severity, status,
			control_id, schedule_cron, schedule_interval_min, test_config, test_script, test_script_language,
			timeout_seconds, retry_count, retry_delay_seconds, tags, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'draft', $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $19)
	`, id, orgID, req.Identifier, req.Title, req.Description, req.TestType, severity,
		req.ControlID, req.ScheduleCron, req.ScheduleIntervalMin, configJSON,
		req.TestScript, req.TestScriptLanguage,
		timeoutSeconds, retryCount, retryDelay, pq.Array(req.Tags), userID, now)

	if err != nil {
		log.Error().Err(err).Msg("Failed to insert test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create test"))
		return
	}

	middleware.LogAudit(c, "test.created", "test", &id, map[string]interface{}{
		"identifier": req.Identifier,
		"test_type":  req.TestType,
		"control_id": req.ControlID,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":            id,
		"identifier":    req.Identifier,
		"title":         req.Title,
		"test_type":     req.TestType,
		"severity":      severity,
		"status":        "draft",
		"control": gin.H{
			"id":         req.ControlID,
			"identifier": controlIdentifier,
			"title":      controlTitle,
		},
		"schedule_cron": req.ScheduleCron,
		"next_run_at":   nil,
		"created_at":    now,
	}))
}

// GetTest gets a single test definition with full details.
func GetTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	testID := c.Param("id")

	var (
		id, identifier, title, testType, severity, status, controlID string
		description                                                   *string
		controlIdentifier, controlTitle, controlCategory, controlStatus string
		scheduleCron                                                   *string
		scheduleIntervalMin                                            *int
		nextRunAt, lastRunAt                                           *time.Time
		testConfig                                                     string
		testScript, testScriptLanguage                                 *string
		timeoutSeconds, retryCount, retryDelay                         int
		tags                                                           pq.StringArray
		createdBy                                                      *string
		createdByName                                                  string
		createdAt, updatedAt                                           time.Time
	)

	err := database.QueryRow(`
		SELECT t.id, t.identifier, t.title, t.description, t.test_type, t.severity, t.status,
			t.control_id, c.identifier, c.title, c.category, c.status,
			t.schedule_cron, t.schedule_interval_min, t.next_run_at, t.last_run_at,
			t.test_config, t.test_script, t.test_script_language,
			t.timeout_seconds, t.retry_count, t.retry_delay_seconds,
			t.tags, t.created_by, COALESCE(u.first_name || ' ' || u.last_name, ''),
			t.created_at, t.updated_at
		FROM tests t
		LEFT JOIN controls c ON c.id = t.control_id
		LEFT JOIN users u ON u.id = t.created_by
		WHERE t.id = $1 AND t.org_id = $2
	`, testID, orgID).Scan(
		&id, &identifier, &title, &description, &testType, &severity, &status,
		&controlID, &controlIdentifier, &controlTitle, &controlCategory, &controlStatus,
		&scheduleCron, &scheduleIntervalMin, &nextRunAt, &lastRunAt,
		&testConfig, &testScript, &testScriptLanguage,
		&timeoutSeconds, &retryCount, &retryDelay,
		&tags, &createdBy, &createdByName,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get test"))
		return
	}

	// Parse test_config as raw JSON
	var configObj interface{}
	json.Unmarshal([]byte(testConfig), &configObj)

	result := gin.H{
		"id":                    id,
		"identifier":            identifier,
		"title":                 title,
		"description":           description,
		"test_type":             testType,
		"severity":              severity,
		"status":                status,
		"control": gin.H{
			"id":         controlID,
			"identifier": controlIdentifier,
			"title":      controlTitle,
			"category":   controlCategory,
			"status":     controlStatus,
		},
		"schedule_cron":         scheduleCron,
		"schedule_interval_min": scheduleIntervalMin,
		"next_run_at":           nextRunAt,
		"last_run_at":           lastRunAt,
		"timeout_seconds":       timeoutSeconds,
		"retry_count":           retryCount,
		"retry_delay_seconds":   retryDelay,
		"test_config":           configObj,
		"test_script":           testScript,
		"test_script_language":  testScriptLanguage,
		"tags":                  tags,
		"created_at":            createdAt,
		"updated_at":            updatedAt,
	}

	if createdBy != nil {
		result["created_by"] = gin.H{
			"id":   *createdBy,
			"name": createdByName,
		}
	}

	c.JSON(http.StatusOK, successResponse(c, result))
}

// UpdateTest updates a test definition.
func UpdateTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	testID := c.Param("id")

	var req models.UpdateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid request body"))
		return
	}

	// Check test exists
	var currentIdentifier string
	err := database.QueryRow("SELECT identifier FROM tests WHERE id = $1 AND org_id = $2", testID, orgID).Scan(&currentIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update test"))
		return
	}

	sets := []string{}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		if len(*req.Title) > 500 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Title must be 500 characters or less"))
			return
		}
		sets = append(sets, fmt.Sprintf("title = $%d", argN))
		args = append(args, *req.Title)
		argN++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argN))
		args = append(args, *req.Description)
		argN++
	}
	if req.Severity != nil {
		if !models.IsValidTestSeverity(*req.Severity) {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid severity"))
			return
		}
		sets = append(sets, fmt.Sprintf("severity = $%d", argN))
		args = append(args, *req.Severity)
		argN++
	}
	if req.ScheduleCron != nil {
		sets = append(sets, fmt.Sprintf("schedule_cron = $%d", argN))
		args = append(args, *req.ScheduleCron)
		argN++
		sets = append(sets, "schedule_interval_min = NULL")
	}
	if req.ScheduleIntervalMin != nil {
		if *req.ScheduleIntervalMin < 1 || *req.ScheduleIntervalMin > 10080 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "schedule_interval_min must be between 1 and 10080"))
			return
		}
		sets = append(sets, fmt.Sprintf("schedule_interval_min = $%d", argN))
		args = append(args, *req.ScheduleIntervalMin)
		argN++
		sets = append(sets, "schedule_cron = NULL")
	}
	if req.TimeoutSeconds != nil {
		if *req.TimeoutSeconds < 1 || *req.TimeoutSeconds > 3600 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "timeout_seconds must be between 1 and 3600"))
			return
		}
		sets = append(sets, fmt.Sprintf("timeout_seconds = $%d", argN))
		args = append(args, *req.TimeoutSeconds)
		argN++
	}
	if req.RetryCount != nil {
		sets = append(sets, fmt.Sprintf("retry_count = $%d", argN))
		args = append(args, *req.RetryCount)
		argN++
	}
	if req.RetryDelaySeconds != nil {
		sets = append(sets, fmt.Sprintf("retry_delay_seconds = $%d", argN))
		args = append(args, *req.RetryDelaySeconds)
		argN++
	}
	if req.TestConfig != nil {
		b, _ := json.Marshal(req.TestConfig)
		sets = append(sets, fmt.Sprintf("test_config = $%d", argN))
		args = append(args, string(b))
		argN++
	}
	if req.TestScript != nil {
		sets = append(sets, fmt.Sprintf("test_script = $%d", argN))
		args = append(args, *req.TestScript)
		argN++
	}
	if req.TestScriptLanguage != nil {
		sets = append(sets, fmt.Sprintf("test_script_language = $%d", argN))
		args = append(args, *req.TestScriptLanguage)
		argN++
	}
	if req.Tags != nil {
		if len(req.Tags) > 20 {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Maximum 20 tags allowed"))
			return
		}
		sets = append(sets, fmt.Sprintf("tags = $%d", argN))
		args = append(args, pq.Array(req.Tags))
		argN++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "No fields to update"))
		return
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, testID, orgID)
	query := fmt.Sprintf("UPDATE tests SET %s WHERE id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1)

	_, err = database.Exec(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update test"))
		return
	}

	middleware.LogAudit(c, "test.updated", "test", &testID, map[string]interface{}{
		"identifier": currentIdentifier,
	})

	// Return updated test
	GetTest(c)
}

// ChangeTestStatus changes a test's lifecycle status.
func ChangeTestStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	testID := c.Param("id")

	var req models.ChangeTestStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Status is required"))
		return
	}

	if !models.IsValidTestStatus(req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Invalid status"))
		return
	}

	// Get current status
	var currentStatus, identifier string
	err := database.QueryRow("SELECT status, identifier FROM tests WHERE id = $1 AND org_id = $2", testID, orgID).Scan(&currentStatus, &identifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get test status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change status"))
		return
	}

	if !models.IsValidTestStatusTransition(currentStatus, req.Status) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
			fmt.Sprintf("Cannot transition from '%s' to '%s'", currentStatus, req.Status)))
		return
	}

	// Compute next_run_at on activation
	var nextRunAt *time.Time
	updateFields := "status = $1, updated_at = NOW()"
	updateArgs := []interface{}{req.Status}
	argN := 2

	if req.Status == "active" {
		// Set next_run_at to now for first run
		now := time.Now().Add(time.Minute)
		nextRunAt = &now
		updateFields += fmt.Sprintf(", next_run_at = $%d", argN)
		updateArgs = append(updateArgs, now)
		argN++
	} else if req.Status == "paused" || req.Status == "deprecated" {
		updateFields += ", next_run_at = NULL"
	}

	updateArgs = append(updateArgs, testID, orgID)
	query := fmt.Sprintf("UPDATE tests SET %s WHERE id = $%d AND org_id = $%d", updateFields, argN, argN+1)
	_, err = database.Exec(query, updateArgs...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update test status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change status"))
		return
	}

	middleware.LogAudit(c, "test.status_changed", "test", &testID, map[string]interface{}{
		"old_status": currentStatus,
		"new_status": req.Status,
	})

	message := fmt.Sprintf("Test status changed to %s.", req.Status)
	if req.Status == "active" {
		message = "Test activated. First run scheduled."
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":              testID,
		"status":          req.Status,
		"previous_status": currentStatus,
		"next_run_at":     nextRunAt,
		"message":         message,
	}))
}

// DeleteTest soft-deletes a test by marking it deprecated.
func DeleteTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	testID := c.Param("id")

	var identifier string
	err := database.QueryRow("SELECT identifier FROM tests WHERE id = $1 AND org_id = $2", testID, orgID).Scan(&identifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete test"))
		return
	}

	_, err = database.Exec(
		"UPDATE tests SET status = 'deprecated', next_run_at = NULL, updated_at = NOW() WHERE id = $1 AND org_id = $2",
		testID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to deprecate test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete test"))
		return
	}

	middleware.LogAudit(c, "test.deleted", "test", &testID, map[string]interface{}{
		"identifier": identifier,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":      testID,
		"status":  "deprecated",
		"message": "Test deprecated. Historical results preserved.",
	}))
}
