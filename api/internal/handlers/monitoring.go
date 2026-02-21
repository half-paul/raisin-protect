package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// GetControlHealthHeatmap returns control health data for the heatmap visualization.
func GetControlHealthHeatmap(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	where := "c.org_id = $1 AND c.status = 'active'"
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("category"); v != "" {
		where += " AND c.category = $2"
		args = append(args, v)
		argN++
	}

	query := `
		SELECT c.id, c.identifier, c.title, c.category,
			lr.status AS latest_status, lr.severity AS latest_severity,
			lr.message AS latest_message, lr.created_at AS latest_tested_at,
			(SELECT COUNT(*) FROM alerts WHERE control_id = c.id AND org_id = c.org_id
				AND status IN ('open', 'acknowledged', 'in_progress')) AS active_alerts,
			(SELECT COUNT(*) FROM tests WHERE control_id = c.id AND org_id = c.org_id
				AND status != 'deprecated') AS tests_count
		FROM controls c
		LEFT JOIN LATERAL (
			SELECT tr.status, tr.severity, tr.message, tr.created_at
			FROM test_results tr
			WHERE tr.control_id = c.id AND tr.org_id = c.org_id
			ORDER BY tr.created_at DESC
			LIMIT 1
		) lr ON TRUE
		WHERE ` + where + `
		ORDER BY
			CASE
				WHEN lr.status = 'fail' THEN 0
				WHEN lr.status = 'error' THEN 1
				WHEN lr.status = 'warning' THEN 2
				WHEN lr.status IS NULL THEN 3
				WHEN lr.status = 'pass' THEN 4
				ELSE 5
			END,
			c.identifier
	`

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query heatmap")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get heatmap"))
		return
	}
	defer rows.Close()

	var (
		controls                                     []gin.H
		totalControls, healthy, failing, errCount, warning, untested int
	)

	for rows.Next() {
		var (
			id, identifier, title, category string
			latestStatus, latestSeverity, latestMessage *string
			latestTestedAt                              *time.Time
			activeAlerts, testsCount                    int
		)
		if err := rows.Scan(
			&id, &identifier, &title, &category,
			&latestStatus, &latestSeverity, &latestMessage, &latestTestedAt,
			&activeAlerts, &testsCount,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan heatmap row")
			continue
		}

		healthStatus := "untested"
		if latestStatus != nil {
			switch *latestStatus {
			case "pass":
				healthStatus = "healthy"
				healthy++
			case "fail":
				healthStatus = "failing"
				failing++
			case "error":
				healthStatus = "error"
				errCount++
			case "warning":
				healthStatus = "warning"
				warning++
			default:
				untested++
			}
		} else {
			untested++
		}
		totalControls++

		item := gin.H{
			"id":            id,
			"identifier":    identifier,
			"title":         title,
			"category":      category,
			"health_status": healthStatus,
			"active_alerts": activeAlerts,
			"tests_count":   testsCount,
		}

		if latestStatus != nil {
			item["latest_result"] = gin.H{
				"status":    *latestStatus,
				"severity":  latestSeverity,
				"message":   latestMessage,
				"tested_at": latestTestedAt,
			}
		} else {
			item["latest_result"] = nil
		}

		controls = append(controls, item)
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"summary": gin.H{
			"total_controls": totalControls,
			"healthy":        healthy,
			"failing":        failing,
			"error":          errCount,
			"warning":        warning,
			"untested":       untested,
		},
		"controls": controls,
	}))
}

// GetCompliancePosture returns compliance posture score per activated framework.
func GetCompliancePosture(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	query := `
		SELECT f.id, f.name, fv.version, of.id AS org_framework_id,
			COUNT(DISTINCT cm.control_id) AS total_mapped_controls,
			COUNT(DISTINCT cm.control_id) FILTER (WHERE lr.status = 'pass') AS passing,
			COUNT(DISTINCT cm.control_id) FILTER (WHERE lr.status = 'fail') AS failing,
			COUNT(DISTINCT cm.control_id) FILTER (WHERE lr.status IS NULL) AS untested
		FROM org_frameworks of
		JOIN framework_versions fv ON fv.id = of.framework_version_id
		JOIN frameworks f ON f.id = fv.framework_id
		JOIN requirements r ON r.framework_version_id = fv.id
		JOIN control_mappings cm ON cm.requirement_id = r.id AND cm.org_id = of.org_id
		LEFT JOIN LATERAL (
			SELECT tr.status
			FROM test_results tr
			WHERE tr.control_id = cm.control_id AND tr.org_id = of.org_id
			ORDER BY tr.created_at DESC
			LIMIT 1
		) lr ON TRUE
		WHERE of.org_id = $1 AND of.status = 'active'
		GROUP BY f.id, f.name, fv.version, of.id
		ORDER BY f.name
	`

	rows, err := database.Query(query, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query posture")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get posture"))
		return
	}
	defer rows.Close()

	var frameworks []gin.H
	var totalPassing, totalMapped int

	for rows.Next() {
		var (
			fID, fName, fVersion, ofID string
			total, passing, failing, untested int
		)
		if err := rows.Scan(&fID, &fName, &fVersion, &ofID, &total, &passing, &failing, &untested); err != nil {
			log.Error().Err(err).Msg("Failed to scan posture row")
			continue
		}

		score := float64(0)
		if total > 0 {
			score = float64(passing) / float64(total) * 100
		}

		totalPassing += passing
		totalMapped += total

		frameworks = append(frameworks, gin.H{
			"framework_id":          fID,
			"framework_name":        fName,
			"framework_version":     fVersion,
			"org_framework_id":      ofID,
			"total_mapped_controls": total,
			"passing":               passing,
			"failing":               failing,
			"untested":              untested,
			"posture_score":         float64(int(score*10)) / 10,
		})
	}

	overallScore := float64(0)
	if totalMapped > 0 {
		overallScore = float64(totalPassing) / float64(totalMapped) * 100
		overallScore = float64(int(overallScore*10)) / 10
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"overall_score": overallScore,
		"frameworks":    frameworks,
	}))
}

// GetMonitoringSummary returns top-level monitoring dashboard summary.
func GetMonitoringSummary(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	result := gin.H{}

	// Controls summary
	var totalActive, controlHealthy, controlFailing, controlUntested int
	err := database.QueryRow(`
		SELECT
			COUNT(*) AS total_active,
			COUNT(*) FILTER (WHERE lr.status = 'pass') AS healthy,
			COUNT(*) FILTER (WHERE lr.status IN ('fail', 'error')) AS failing,
			COUNT(*) FILTER (WHERE lr.status IS NULL) AS untested
		FROM controls c
		LEFT JOIN LATERAL (
			SELECT tr.status
			FROM test_results tr
			WHERE tr.control_id = c.id AND tr.org_id = c.org_id
			ORDER BY tr.created_at DESC
			LIMIT 1
		) lr ON TRUE
		WHERE c.org_id = $1 AND c.status = 'active'
	`, orgID).Scan(&totalActive, &controlHealthy, &controlFailing, &controlUntested)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query control summary")
	}

	healthRate := float64(0)
	if totalActive > 0 {
		healthRate = float64(controlHealthy) / float64(totalActive) * 100
		healthRate = float64(int(healthRate*10)) / 10
	}

	result["controls"] = gin.H{
		"total_active": totalActive,
		"healthy":      controlHealthy,
		"failing":      controlFailing,
		"untested":     controlUntested,
		"health_rate":  healthRate,
	}

	// Tests summary
	var totalActiveTests int
	database.QueryRow("SELECT COUNT(*) FROM tests WHERE org_id = $1 AND status = 'active'", orgID).Scan(&totalActiveTests)

	// Last run
	var lastRunNumber *int
	var lastRunStatus *string
	var lastRunCompleted *time.Time
	var lastRunPassed, lastRunFailed, lastRunErrors *int
	database.QueryRow(`
		SELECT run_number, status, completed_at, passed, failed, errors
		FROM test_runs WHERE org_id = $1 ORDER BY created_at DESC LIMIT 1
	`, orgID).Scan(&lastRunNumber, &lastRunStatus, &lastRunCompleted, &lastRunPassed, &lastRunFailed, &lastRunErrors)

	// Pass rate 24h
	var pass24h, total24h int
	database.QueryRow(`
		SELECT COUNT(*) FILTER (WHERE status = 'pass'), COUNT(*)
		FROM test_results WHERE org_id = $1 AND created_at >= NOW() - INTERVAL '24 hours'
	`, orgID).Scan(&pass24h, &total24h)

	passRate24h := float64(0)
	if total24h > 0 {
		passRate24h = float64(pass24h) / float64(total24h) * 100
		passRate24h = float64(int(passRate24h*10)) / 10
	}

	testsResult := gin.H{
		"total_active":   totalActiveTests,
		"pass_rate_24h":  passRate24h,
	}
	if lastRunNumber != nil {
		testsResult["last_run"] = gin.H{
			"run_number":   *lastRunNumber,
			"status":       lastRunStatus,
			"completed_at": lastRunCompleted,
			"passed":       lastRunPassed,
			"failed":       lastRunFailed,
			"errors":       lastRunErrors,
		}
	}
	result["tests"] = testsResult

	// Alerts summary
	var alertOpen, alertAcknowledged, alertInProgress, alertSlaBreached, alertResolvedToday int
	database.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'open') AS open,
			COUNT(*) FILTER (WHERE status = 'acknowledged') AS acknowledged,
			COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress,
			COUNT(*) FILTER (WHERE sla_breached = TRUE AND status NOT IN ('resolved', 'closed')) AS sla_breached,
			COUNT(*) FILTER (WHERE status = 'resolved' AND resolved_at >= CURRENT_DATE) AS resolved_today
		FROM alerts WHERE org_id = $1
	`, orgID).Scan(&alertOpen, &alertAcknowledged, &alertInProgress, &alertSlaBreached, &alertResolvedToday)

	// Alerts by severity
	var critCount, highCount, medCount, lowCount int
	database.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE severity = 'critical' AND status IN ('open', 'acknowledged', 'in_progress')),
			COUNT(*) FILTER (WHERE severity = 'high' AND status IN ('open', 'acknowledged', 'in_progress')),
			COUNT(*) FILTER (WHERE severity = 'medium' AND status IN ('open', 'acknowledged', 'in_progress')),
			COUNT(*) FILTER (WHERE severity = 'low' AND status IN ('open', 'acknowledged', 'in_progress'))
		FROM alerts WHERE org_id = $1
	`, orgID).Scan(&critCount, &highCount, &medCount, &lowCount)

	result["alerts"] = gin.H{
		"open":           alertOpen,
		"acknowledged":   alertAcknowledged,
		"in_progress":    alertInProgress,
		"sla_breached":   alertSlaBreached,
		"resolved_today": alertResolvedToday,
		"by_severity": gin.H{
			"critical": critCount,
			"high":     highCount,
			"medium":   medCount,
			"low":      lowCount,
		},
	}

	// Recent activity
	recentActivity := []gin.H{}
	activityRows, err := database.Query(`
		(SELECT 'alert_created' AS type, a.alert_number AS number, a.title,
			a.severity, NULL AS resolved_by_name, a.created_at AS ts
		FROM alerts a WHERE a.org_id = $1
		ORDER BY a.created_at DESC LIMIT 5)
		UNION ALL
		(SELECT 'test_run_completed' AS type, tr.run_number, 'Test run' AS title,
			NULL, NULL, tr.completed_at AS ts
		FROM test_runs tr WHERE tr.org_id = $1 AND tr.status = 'completed'
		ORDER BY tr.completed_at DESC LIMIT 5)
		ORDER BY ts DESC LIMIT 10
	`, orgID)
	if err == nil {
		defer activityRows.Close()
		for activityRows.Next() {
			var actType string
			var number int
			var actTitle string
			var actSeverity, resolvedByName *string
			var ts *time.Time
			if err := activityRows.Scan(&actType, &number, &actTitle, &actSeverity, &resolvedByName, &ts); err != nil {
				continue
			}
			item := gin.H{
				"type":      actType,
				"title":     actTitle,
				"timestamp": ts,
			}
			if actType == "alert_created" {
				item["alert_number"] = number
				item["severity"] = actSeverity
			} else {
				item["run_number"] = number
			}
			recentActivity = append(recentActivity, item)
		}
	}
	result["recent_activity"] = recentActivity

	// Overall posture score
	var overallPassing, overallTotal int
	database.QueryRow(`
		SELECT
			COUNT(DISTINCT cm.control_id) FILTER (WHERE lr.status = 'pass'),
			COUNT(DISTINCT cm.control_id)
		FROM org_frameworks of
		JOIN framework_versions fv ON fv.id = of.framework_version_id
		JOIN requirements r ON r.framework_version_id = fv.id
		JOIN control_mappings cm ON cm.requirement_id = r.id AND cm.org_id = of.org_id
		LEFT JOIN LATERAL (
			SELECT tr.status
			FROM test_results tr
			WHERE tr.control_id = cm.control_id AND tr.org_id = of.org_id
			ORDER BY tr.created_at DESC LIMIT 1
		) lr ON TRUE
		WHERE of.org_id = $1 AND of.status = 'active'
	`, orgID).Scan(&overallPassing, &overallTotal)

	overallScore := float64(0)
	if overallTotal > 0 {
		overallScore = float64(overallPassing) / float64(overallTotal) * 100
		overallScore = float64(int(overallScore*10)) / 10
	}
	result["overall_posture_score"] = overallScore

	c.JSON(http.StatusOK, successResponse(c, result))
}

// GetAlertQueue returns the alert queue view for operations.
func GetAlertQueue(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	queue := c.DefaultQuery("queue", "active")

	// Queue summary — always show all counts
	var activeCount, resolvedCount, suppressedCount, closedCount, slaBreachedCount int
	database.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status IN ('open', 'acknowledged', 'in_progress')),
			COUNT(*) FILTER (WHERE status = 'resolved'),
			COUNT(*) FILTER (WHERE status = 'suppressed'),
			COUNT(*) FILTER (WHERE status = 'closed'),
			COUNT(*) FILTER (WHERE sla_breached = TRUE AND status NOT IN ('resolved', 'closed'))
		FROM alerts WHERE org_id = $1
	`, orgID).Scan(&activeCount, &resolvedCount, &suppressedCount, &closedCount, &slaBreachedCount)

	// Build where clause based on queue
	where := "a.org_id = $1"
	switch queue {
	case "active":
		where += " AND a.status IN ('open', 'acknowledged', 'in_progress')"
	case "resolved":
		where += " AND a.status = 'resolved'"
	case "suppressed":
		where += " AND a.status = 'suppressed'"
	case "all":
		// no additional filter
	default:
		where += " AND a.status IN ('open', 'acknowledged', 'in_progress')"
	}

	var total int
	database.QueryRow("SELECT COUNT(*) FROM alerts a WHERE "+where, orgID).Scan(&total)

	offset := (page - 1) * perPage
	query := `
		SELECT a.id, a.alert_number, a.title, a.severity, a.status,
			c.identifier AS control_identifier,
			t.identifier AS test_identifier,
			COALESCE(u.first_name || ' ' || u.last_name, '') AS assigned_to_name,
			a.sla_deadline, a.sla_breached, a.created_at
		FROM alerts a
		LEFT JOIN controls c ON c.id = a.control_id
		LEFT JOIN tests t ON t.id = a.test_id
		LEFT JOIN users u ON u.id = a.assigned_to
		WHERE ` + where + `
		ORDER BY
			CASE a.severity
				WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3
			END,
			COALESCE(a.sla_deadline, '9999-12-31'::TIMESTAMPTZ) ASC,
			a.created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := database.Query(query, orgID, perPage, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query alert queue")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get alert queue"))
		return
	}
	defer rows.Close()

	alerts := []gin.H{}
	for rows.Next() {
		var (
			aID, aTitle, aSeverity, aStatus       string
			controlIdentifier, testIdentifier      *string
			assignedToName                         string
			slaDeadline                            *time.Time
			slaBreached                            bool
			createdAt                              time.Time
		)
		if err := rows.Scan(
			&aID, new(int), &aTitle, &aSeverity, &aStatus,
			&controlIdentifier, &testIdentifier, &assignedToName,
			&slaDeadline, &slaBreached, &createdAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan queue row")
			continue
		}

		item := gin.H{
			"id":                   aID,
			"title":                aTitle,
			"severity":             aSeverity,
			"status":               aStatus,
			"control_identifier":   controlIdentifier,
			"test_identifier":      testIdentifier,
			"assigned_to_name":     nilIfEmpty(assignedToName),
			"sla_deadline":         slaDeadline,
			"sla_breached":         slaBreached,
			"created_at":           createdAt,
		}

		if slaDeadline != nil {
			remaining := time.Until(*slaDeadline).Hours()
			item["hours_remaining"] = json.Number(fmt.Sprintf("%.1f", remaining))
		}

		alerts = append(alerts, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"queue_summary": gin.H{
				"active":      activeCount,
				"resolved":    resolvedCount,
				"suppressed":  suppressedCount,
				"closed":      closedCount,
				"sla_breached": slaBreachedCount,
			},
			"alerts": alerts,
		},
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
