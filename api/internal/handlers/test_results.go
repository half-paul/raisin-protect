package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// ListTestResultsByTest lists execution history for a specific test.
func ListTestResultsByTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	testID := c.Param("id")

	// Check test exists
	var testIdentifier, testTitle string
	err := database.QueryRow(
		"SELECT identifier, title FROM tests WHERE id = $1 AND org_id = $2",
		testID, orgID,
	).Scan(&testIdentifier, &testTitle)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Test not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"r.test_id = $1", "r.org_id = $2"}
	args := []interface{}{testID, orgID}
	argN := 3

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_from"); v != "" {
		where = append(where, fmt.Sprintf("r.created_at >= $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_to"); v != "" {
		where = append(where, fmt.Sprintf("r.created_at <= $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM test_results r WHERE %s", whereClause)
	if err := database.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count test results")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}

	sortField := "r.created_at"
	switch c.DefaultQuery("sort", "created_at") {
	case "status":
		sortField = "r.status"
	case "duration_ms":
		sortField = "r.duration_ms"
	}
	order := "DESC"
	if strings.ToLower(c.DefaultQuery("order", "desc")) == "asc" {
		order = "ASC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.test_run_id, tr.run_number, r.status, r.severity, r.message,
			r.duration_ms, r.alert_generated, r.started_at, r.created_at
		FROM test_results r
		LEFT JOIN test_runs tr ON tr.id = r.test_run_id
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, whereClause, sortField, order, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query test results")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			rID, testRunID, rStatus, rSeverity string
			runNumber                           int
			message                             *string
			durationMs                          *int
			alertGenerated                      bool
			startedAt, createdAt                time.Time
		)
		if err := rows.Scan(
			&rID, &testRunID, &runNumber, &rStatus, &rSeverity, &message,
			&durationMs, &alertGenerated, &startedAt, &createdAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan result row")
			continue
		}

		results = append(results, gin.H{
			"id":              rID,
			"test_run_id":     testRunID,
			"run_number":      runNumber,
			"status":          rStatus,
			"severity":        rSeverity,
			"message":         message,
			"duration_ms":     durationMs,
			"alert_generated": alertGenerated,
			"started_at":      startedAt,
			"created_at":      createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"test": gin.H{
				"id":         testID,
				"identifier": testIdentifier,
				"title":      testTitle,
			},
			"results": results,
		},
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}

// ListControlTestResults lists test execution history for a specific control.
func ListControlTestResults(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	// Check control exists
	var controlIdentifier, controlTitle string
	err := database.QueryRow(
		"SELECT identifier, title FROM controls WHERE id = $1 AND org_id = $2",
		controlID, orgID,
	).Scan(&controlIdentifier, &controlTitle)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to check control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"r.control_id = $1", "r.org_id = $2"}
	args := []interface{}{controlID, orgID}
	argN := 3

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_from"); v != "" {
		where = append(where, fmt.Sprintf("r.created_at >= $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("date_to"); v != "" {
		where = append(where, fmt.Sprintf("r.created_at <= $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM test_results r WHERE %s", whereClause)
	if err := database.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count control test results")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}

	// Get health status and test count
	var healthStatus string
	var testsCount int
	err = database.QueryRow(
		"SELECT COUNT(*) FROM tests WHERE control_id = $1 AND org_id = $2 AND status != 'deprecated'",
		controlID, orgID,
	).Scan(&testsCount)
	if err != nil {
		testsCount = 0
	}

	// Get latest result for health status
	var latestStatus *string
	database.QueryRow(`
		SELECT r.status FROM test_results r
		WHERE r.control_id = $1 AND r.org_id = $2
		ORDER BY r.created_at DESC LIMIT 1
	`, controlID, orgID).Scan(&latestStatus)

	if latestStatus == nil {
		healthStatus = "untested"
	} else {
		switch *latestStatus {
		case "pass":
			healthStatus = "healthy"
		case "fail":
			healthStatus = "failing"
		case "error":
			healthStatus = "error"
		case "warning":
			healthStatus = "warning"
		default:
			healthStatus = "untested"
		}
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.test_id, t.identifier, t.title,
			tr.run_number, r.status, r.severity, r.message,
			r.duration_ms, r.started_at, r.created_at
		FROM test_results r
		LEFT JOIN tests t ON t.id = r.test_id
		LEFT JOIN test_runs tr ON tr.id = r.test_run_id
		WHERE %s
		ORDER BY r.created_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query control results")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list results"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			rID, testID, testIdentifier, testTitle, rStatus, rSeverity string
			runNumber                                                   int
			message                                                     *string
			durationMs                                                  *int
			startedAt, createdAt                                        time.Time
		)
		if err := rows.Scan(
			&rID, &testID, &testIdentifier, &testTitle,
			&runNumber, &rStatus, &rSeverity, &message,
			&durationMs, &startedAt, &createdAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan result row")
			continue
		}

		results = append(results, gin.H{
			"id": rID,
			"test": gin.H{
				"id":         testID,
				"identifier": testIdentifier,
				"title":      testTitle,
			},
			"run_number":  runNumber,
			"status":      rStatus,
			"severity":    rSeverity,
			"message":     message,
			"duration_ms": durationMs,
			"started_at":  startedAt,
			"created_at":  createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"control": gin.H{
				"id":         controlID,
				"identifier": controlIdentifier,
				"title":      controlTitle,
			},
			"health_status": healthStatus,
			"tests_count":   testsCount,
			"results":       results,
		},
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}
