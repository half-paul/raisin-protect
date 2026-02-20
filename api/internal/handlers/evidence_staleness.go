package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// GetStalenessAlerts returns staleness alerts for evidence artifacts.
func GetStalenessAlerts(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	daysAhead, _ := strconv.Atoi(c.DefaultQuery("days_ahead", "30"))
	if daysAhead < 1 {
		daysAhead = 30
	}

	where := []string{
		"ea.org_id = $1",
		"ea.is_current = TRUE",
		"ea.status NOT IN ('superseded', 'draft')",
		"ea.expires_at IS NOT NULL",
		fmt.Sprintf("ea.expires_at < NOW() + INTERVAL '%d days'", daysAhead),
	}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("alert_level"); v != "" {
		switch v {
		case "expired":
			where = append(where, "ea.expires_at < NOW()")
		case "expiring_soon":
			where = append(where, "ea.expires_at >= NOW()")
		}
	}
	if v := c.Query("evidence_type"); v != "" {
		where = append(where, fmt.Sprintf("ea.evidence_type = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	// Summary
	var totalAlerts, expiredCount, expiringSoonCount int
	database.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*),
			   COUNT(*) FILTER (WHERE ea.expires_at < NOW()),
			   COUNT(*) FILTER (WHERE ea.expires_at >= NOW())
		FROM evidence_artifacts ea WHERE %s
	`, whereClause), args...).Scan(&totalAlerts, &expiredCount, &expiringSoonCount)

	// Count affected controls
	var affectedControls int
	database.QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT el.control_id)
		FROM evidence_artifacts ea
		JOIN evidence_links el ON el.artifact_id = ea.id AND el.target_type = 'control'
		WHERE %s
	`, whereClause), args...).Scan(&affectedControls)

	// Sort
	allowedSort := map[string]string{
		"expires_at":    "ea.expires_at",
		"title":         "ea.title",
		"evidence_type": "ea.evidence_type",
	}
	sortField := c.DefaultQuery("sort", "expires_at")
	sortCol, ok := allowedSort[sortField]
	if !ok {
		sortCol = "ea.expires_at"
	}
	order := c.DefaultQuery("order", "asc")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ea.id, ea.title, ea.evidence_type, ea.status,
			   ea.collection_date, ea.expires_at, ea.freshness_period_days,
			   ea.uploaded_by, COALESCE(u.first_name || ' ' || u.last_name, '')
		FROM evidence_artifacts ea
		LEFT JOIN users u ON u.id = ea.uploaded_by
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get staleness alerts")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	alerts := []gin.H{}
	for rows.Next() {
		var eID, eTitle, eType, eStatus, eCollDate string
		var eExpiresAt *time.Time
		var eFreshDays *int
		var eUploadedBy *string
		var eUploaderName string

		if err := rows.Scan(&eID, &eTitle, &eType, &eStatus,
			&eCollDate, &eExpiresAt, &eFreshDays,
			&eUploadedBy, &eUploaderName); err != nil {
			continue
		}

		alert := gin.H{
			"id":                    eID,
			"title":                 eTitle,
			"evidence_type":         eType,
			"status":                eStatus,
			"collection_date":       eCollDate,
			"expires_at":            eExpiresAt,
			"freshness_period_days": eFreshDays,
		}

		if eExpiresAt != nil && eExpiresAt.Before(time.Now()) {
			alert["alert_level"] = "expired"
			alert["days_overdue"] = int(time.Since(*eExpiresAt).Hours() / 24)
		} else {
			alert["alert_level"] = "expiring_soon"
			alert["days_until_expiry"] = daysUntilExpiry(eExpiresAt)
		}

		if eUploadedBy != nil {
			alert["uploaded_by"] = gin.H{"id": *eUploadedBy, "name": eUploaderName}
		} else {
			alert["uploaded_by"] = nil
		}

		// Get linked controls
		ctrlRows, _ := database.Query(`
			SELECT c.id, c.identifier, c.title FROM evidence_links el
			JOIN controls c ON c.id = el.control_id
			WHERE el.artifact_id = $1 AND el.target_type = 'control'
		`, eID)
		linkedControls := []gin.H{}
		if ctrlRows != nil {
			for ctrlRows.Next() {
				var cID, cIdentifier, cTitle string
				ctrlRows.Scan(&cID, &cIdentifier, &cTitle)
				linkedControls = append(linkedControls, gin.H{
					"id": cID, "identifier": cIdentifier, "title": cTitle,
				})
			}
			ctrlRows.Close()
		}
		alert["linked_controls"] = linkedControls
		alert["linked_controls_count"] = len(linkedControls)

		alerts = append(alerts, alert)
	}

	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"summary": gin.H{
				"total_alerts":      totalAlerts,
				"expired":           expiredCount,
				"expiring_soon":     expiringSoonCount,
				"affected_controls": affectedControls,
			},
			"alerts": alerts,
		},
		"meta": gin.H{
			"total": totalAlerts, "page": page, "per_page": perPage,
			"request_id": reqID,
		},
	})
}

// GetFreshnessSummary returns a high-level freshness overview.
func GetFreshnessSummary(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var totalEvidence int
	database.QueryRow("SELECT COUNT(*) FROM evidence_artifacts WHERE org_id = $1 AND is_current = TRUE", orgID).Scan(&totalEvidence)

	// By freshness
	var freshCount, expiringSoonCount, expiredCount, noExpiryCount int
	database.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at > NOW() + INTERVAL '30 days'),
			COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at <= NOW() + INTERVAL '30 days' AND expires_at > NOW()),
			COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at <= NOW()),
			COUNT(*) FILTER (WHERE expires_at IS NULL)
		FROM evidence_artifacts WHERE org_id = $1 AND is_current = TRUE
	`, orgID).Scan(&freshCount, &expiringSoonCount, &expiredCount, &noExpiryCount)

	// By status
	byStatus := gin.H{}
	statusRows, _ := database.Query(
		"SELECT status, COUNT(*) FROM evidence_artifacts WHERE org_id = $1 AND is_current = TRUE GROUP BY status", orgID)
	if statusRows != nil {
		for statusRows.Next() {
			var s string
			var cnt int
			statusRows.Scan(&s, &cnt)
			byStatus[s] = cnt
		}
		statusRows.Close()
	}

	// By type
	byType := gin.H{}
	typeRows, _ := database.Query(
		"SELECT evidence_type, COUNT(*) FROM evidence_artifacts WHERE org_id = $1 AND is_current = TRUE GROUP BY evidence_type", orgID)
	if typeRows != nil {
		for typeRows.Next() {
			var t string
			var cnt int
			typeRows.Scan(&t, &cnt)
			byType[t] = cnt
		}
		typeRows.Close()
	}

	// Coverage
	var totalActiveControls, controlsWithEvidence int
	database.QueryRow("SELECT COUNT(*) FROM controls WHERE org_id = $1 AND status = 'active'", orgID).Scan(&totalActiveControls)
	database.QueryRow(`
		SELECT COUNT(DISTINCT el.control_id)
		FROM evidence_links el
		JOIN evidence_artifacts ea ON ea.id = el.artifact_id
		WHERE el.org_id = $1 AND el.target_type = 'control'
		  AND ea.is_current = TRUE AND ea.status = 'approved'
		  AND (ea.expires_at IS NULL OR ea.expires_at > NOW())
	`, orgID).Scan(&controlsWithEvidence)

	var coveragePct float64
	if totalActiveControls > 0 {
		coveragePct = float64(controlsWithEvidence) / float64(totalActiveControls) * 100
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"total_evidence": totalEvidence,
		"by_freshness": gin.H{
			"fresh":         freshCount,
			"expiring_soon": expiringSoonCount,
			"expired":       expiredCount,
			"no_expiry":     noExpiryCount,
		},
		"by_status": byStatus,
		"by_type":   byType,
		"coverage": gin.H{
			"total_active_controls":    totalActiveControls,
			"controls_with_evidence":   controlsWithEvidence,
			"controls_without_evidence": totalActiveControls - controlsWithEvidence,
			"evidence_coverage_pct":    coveragePct,
		},
	}))
}
