package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// GetPolicyGap identifies controls without policy coverage.
func GetPolicyGap(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	includePartial := c.Query("include_partial") == "true"

	where := "c.org_id = $1 AND c.status = 'active'"
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("framework_id"); v != "" {
		where += fmt.Sprintf(` AND c.id IN (
			SELECT cm.control_id FROM control_mappings cm
			JOIN requirements r ON r.id = cm.requirement_id
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			WHERE fv.framework_id = $%d
		)`, argN)
		args = append(args, v)
		argN++
	}

	if v := c.Query("category"); v != "" {
		where += fmt.Sprintf(` AND c.category = $%d`, argN)
		args = append(args, v)
		argN++
	}

	// Gap condition: no policy_controls link OR (include_partial and only partial coverage)
	gapCondition := `NOT EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id)`
	if includePartial {
		gapCondition = `(
			NOT EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id)
			OR (
				NOT EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id AND pc.coverage = 'full')
				AND EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id AND pc.coverage = 'partial')
			)
		)`
	}

	// Summary stats
	var totalActive, withFull, withPartial, withoutCoverage int
	database.DB.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id AND pc.coverage = 'full')),
			COUNT(*) FILTER (WHERE NOT EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id AND pc.coverage = 'full')
				AND EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id AND pc.coverage = 'partial')),
			COUNT(*) FILTER (WHERE NOT EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id))
		FROM controls c WHERE `+where, args...).Scan(&totalActive, &withFull, &withPartial, &withoutCoverage)

	coveragePct := float64(0)
	if totalActive > 0 {
		coveragePct = float64(withFull+withPartial) / float64(totalActive) * 100
	}

	// Count gaps
	var totalGaps int
	gapCountQuery := fmt.Sprintf(`SELECT COUNT(*) FROM controls c WHERE %s AND %s`, where, gapCondition)
	database.DB.QueryRow(gapCountQuery, args...).Scan(&totalGaps)

	// Fetch gap records
	offset := (page - 1) * perPage
	gapQuery := fmt.Sprintf(`
		SELECT c.id, c.identifier, c.title, c.category, c.status,
			c.owner_id, u.first_name, u.last_name,
			(SELECT COUNT(DISTINCT cm.requirement_id) FROM control_mappings cm WHERE cm.control_id = c.id) AS req_count,
			CASE
				WHEN NOT EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id) THEN 'none'
				WHEN NOT EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id AND pc.coverage = 'full') THEN 'partial'
				ELSE 'full'
			END AS policy_coverage
		FROM controls c
		LEFT JOIN users u ON u.id = c.owner_id
		WHERE %s AND %s
		ORDER BY (SELECT COUNT(DISTINCT cm.requirement_id) FROM control_mappings cm WHERE cm.control_id = c.id) DESC
		LIMIT $%d OFFSET $%d
	`, where, gapCondition, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(gapQuery, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query policy gaps")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to query policy gaps"))
		return
	}
	defer rows.Close()

	gaps := []gin.H{}
	for rows.Next() {
		var (
			cID, cIdentifier, cTitle, cCategory, cStatus string
			ownerID                                       *string
			oFirst, oLast                                 *string
			reqCount                                      int
			policyCoverage                                string
		)
		if err := rows.Scan(&cID, &cIdentifier, &cTitle, &cCategory, &cStatus, &ownerID, &oFirst, &oLast, &reqCount, &policyCoverage); err != nil {
			continue
		}

		ctrl := gin.H{
			"id":         cID,
			"identifier": cIdentifier,
			"title":      cTitle,
			"category":   cCategory,
			"status":     cStatus,
		}
		if ownerID != nil && oFirst != nil {
			ctrl["owner"] = gin.H{"id": *ownerID, "name": *oFirst + " " + *oLast}
		}

		// Get frameworks mapped to this control
		frameworks := []string{}
		fRows, _ := database.DB.Query(`
			SELECT DISTINCT f.name
			FROM control_mappings cm
			JOIN requirements r ON r.id = cm.requirement_id
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			JOIN frameworks f ON f.id = fv.framework_id
			WHERE cm.control_id = $1
		`, cID)
		if fRows != nil {
			for fRows.Next() {
				var fName string
				if fRows.Scan(&fName) == nil {
					frameworks = append(frameworks, fName)
				}
			}
			fRows.Close()
		}

		// Suggest categories based on control category
		suggestedCats := suggestPolicyCategories(cCategory)

		gaps = append(gaps, gin.H{
			"control":                  ctrl,
			"mapped_frameworks":        frameworks,
			"mapped_requirements_count": reqCount,
			"policy_coverage":          policyCoverage,
			"suggested_categories":     suggestedCats,
		})
	}

	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"summary": gin.H{
				"total_active_controls":          totalActive,
				"controls_with_full_coverage":    withFull,
				"controls_with_partial_coverage": withPartial,
				"controls_without_coverage":      withoutCoverage,
				"coverage_percentage":            coveragePct,
			},
			"gaps": gaps,
		},
		"meta": gin.H{
			"total_gaps": totalGaps,
			"page":       page,
			"per_page":   perPage,
			"request_id": reqID,
		},
	})
}

// GetPolicyGapByFramework returns gap analysis grouped by framework.
func GetPolicyGapByFramework(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	rows, err := database.DB.Query(`
		SELECT f.id, f.identifier, f.name, fv.version_identifier,
			COUNT(DISTINCT r.id) AS total_reqs,
			COUNT(DISTINCT cm.control_id) AS reqs_with_controls,
			COUNT(DISTINCT CASE WHEN EXISTS (
				SELECT 1 FROM policy_controls pc WHERE pc.control_id = cm.control_id AND pc.org_id = $1 AND pc.coverage = 'full'
			) THEN cm.control_id END) AS controls_with_policy,
			COUNT(DISTINCT CASE WHEN NOT EXISTS (
				SELECT 1 FROM policy_controls pc WHERE pc.control_id = cm.control_id AND pc.org_id = $1
			) THEN cm.control_id END) AS controls_without_policy
		FROM org_frameworks of2
		JOIN framework_versions fv ON fv.id = of2.framework_version_id
		JOIN frameworks f ON f.id = fv.framework_id
		LEFT JOIN requirements r ON r.framework_version_id = fv.id
		LEFT JOIN control_mappings cm ON cm.requirement_id = r.id
		WHERE of2.org_id = $1 AND of2.status = 'active'
		GROUP BY f.id, f.identifier, f.name, fv.version_identifier
		ORDER BY f.name
	`, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query gap by framework")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to query gap analysis"))
		return
	}
	defer rows.Close()

	frameworks := []gin.H{}
	for rows.Next() {
		var (
			fID, fIdentifier, fName, fVersion           string
			totalReqs, reqsWithControls                  int
			controlsWithPolicy, controlsWithoutPolicy    int
		)
		if err := rows.Scan(&fID, &fIdentifier, &fName, &fVersion, &totalReqs, &reqsWithControls, &controlsWithPolicy, &controlsWithoutPolicy); err != nil {
			continue
		}

		pct := float64(0)
		if reqsWithControls > 0 {
			pct = float64(controlsWithPolicy) / float64(reqsWithControls) * 100
		}

		frameworks = append(frameworks, gin.H{
			"framework": gin.H{
				"id":         fID,
				"identifier": fIdentifier,
				"name":       fName,
				"version":    fVersion,
			},
			"total_requirements":                totalReqs,
			"requirements_with_controls":        reqsWithControls,
			"controls_with_policy_coverage":     controlsWithPolicy,
			"controls_without_policy_coverage":  controlsWithoutPolicy,
			"policy_coverage_percentage":        pct,
			"gap_count":                         controlsWithoutPolicy,
		})
	}

	c.JSON(http.StatusOK, successResponse(c, frameworks))
}

// GetPolicyStats returns policy management statistics.
func GetPolicyStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	// By status
	var draft, inReview, approved, published, archived int
	database.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'draft'),
			COUNT(*) FILTER (WHERE status = 'in_review'),
			COUNT(*) FILTER (WHERE status = 'approved'),
			COUNT(*) FILTER (WHERE status = 'published'),
			COUNT(*) FILTER (WHERE status = 'archived')
		FROM policies WHERE org_id = $1 AND is_template = FALSE
	`, orgID).Scan(&draft, &inReview, &approved, &published, &archived)

	totalPolicies := draft + inReview + approved + published + archived

	// By category
	catRows, err := database.DB.Query(`
		SELECT category, COUNT(*) FROM policies
		WHERE org_id = $1 AND is_template = FALSE
		GROUP BY category ORDER BY category
	`, orgID)
	byCategory := gin.H{}
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var cat string
			var cnt int
			if catRows.Scan(&cat, &cnt) == nil {
				byCategory[cat] = cnt
			}
		}
	}

	// Review status
	var overdue, dueSoon, onTrack, noSchedule int
	database.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE next_review_at IS NOT NULL AND next_review_at < NOW()),
			COUNT(*) FILTER (WHERE next_review_at IS NOT NULL AND next_review_at >= NOW() AND next_review_at < NOW() + INTERVAL '30 days'),
			COUNT(*) FILTER (WHERE next_review_at IS NOT NULL AND next_review_at >= NOW() + INTERVAL '30 days'),
			COUNT(*) FILTER (WHERE next_review_at IS NULL)
		FROM policies WHERE org_id = $1 AND is_template = FALSE AND status != 'archived'
	`, orgID).Scan(&overdue, &dueSoon, &onTrack, &noSchedule)

	// Signoff summary
	var pendingSignoffs, overdueSignoffs int
	database.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'pending' AND due_date IS NOT NULL AND due_date < NOW())
		FROM policy_signoffs WHERE org_id = $1
	`, orgID).Scan(&pendingSignoffs, &overdueSignoffs)

	// Gap summary
	var totalActiveControls, controlsWithPolicy int
	database.DB.QueryRow(`
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM policy_controls pc WHERE pc.control_id = c.id AND pc.org_id = c.org_id))
		FROM controls c WHERE c.org_id = $1 AND c.status = 'active'
	`, orgID).Scan(&totalActiveControls, &controlsWithPolicy)

	gapPct := float64(0)
	if totalActiveControls > 0 {
		gapPct = float64(controlsWithPolicy) / float64(totalActiveControls) * 100
	}

	// Templates count
	var templatesAvailable int
	database.DB.QueryRow(`SELECT COUNT(*) FROM policies WHERE org_id = $1 AND is_template = TRUE`, orgID).Scan(&templatesAvailable)

	// Recent activity from audit log
	recentActivity := []gin.H{}
	actRows, _ := database.DB.Query(`
		SELECT al.action, al.resource_id, al.created_at,
			u.first_name || ' ' || u.last_name,
			p.identifier
		FROM audit_log al
		LEFT JOIN users u ON u.id = al.actor_id
		LEFT JOIN policies p ON p.id::text = al.resource_id
		WHERE al.org_id = $1 AND al.resource_type IN ('policy', 'policy_version', 'policy_signoff', 'policy_control')
		ORDER BY al.created_at DESC
		LIMIT 5
	`, orgID)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var action string
			var resourceID *string
			var timestamp interface{}
			var actor *string
			var policyIdentifier *string
			if actRows.Scan(&action, &resourceID, &timestamp, &actor, &policyIdentifier) == nil {
				entry := gin.H{
					"action":    action,
					"timestamp": timestamp,
				}
				if actor != nil {
					entry["actor"] = *actor
				}
				if policyIdentifier != nil {
					entry["policy_identifier"] = *policyIdentifier
				}
				recentActivity = append(recentActivity, entry)
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"total_policies": totalPolicies,
		"by_status": gin.H{
			"draft":     draft,
			"in_review": inReview,
			"approved":  approved,
			"published": published,
			"archived":  archived,
		},
		"by_category": byCategory,
		"review_status": gin.H{
			"overdue":             overdue,
			"due_within_30_days":  dueSoon,
			"on_track":            onTrack,
			"no_schedule":         noSchedule,
		},
		"signoff_summary": gin.H{
			"total_pending":     pendingSignoffs,
			"overdue_signoffs":  overdueSignoffs,
		},
		"gap_summary": gin.H{
			"total_active_controls":          totalActiveControls,
			"controls_with_policy_coverage":  controlsWithPolicy,
			"coverage_percentage":            gapPct,
		},
		"templates_available": templatesAvailable,
		"recent_activity":     recentActivity,
	}))
}

// suggestPolicyCategories suggests policy categories based on control category.
func suggestPolicyCategories(controlCategory string) []string {
	mapping := map[string][]string{
		"technical":      {"encryption", "network_security", "access_control", "secure_development"},
		"administrative": {"compliance", "risk_management", "human_resources", "change_management"},
		"physical":       {"physical_security", "asset_management"},
		"operational":    {"incident_response", "business_continuity", "vulnerability_management", "logging_monitoring"},
	}
	if cats, ok := mapping[controlCategory]; ok {
		return cats
	}
	return []string{"information_security"}
}
