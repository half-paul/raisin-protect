package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// GetRiskHeatMap returns aggregated risk data for the 5×5 heat map.
func GetRiskHeatMap(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	scoreType := c.DefaultQuery("score_type", "residual")
	likelihoodCol := "residual_likelihood"
	impactCol := "residual_impact"
	if scoreType == "inherent" {
		likelihoodCol = "inherent_likelihood"
		impactCol = "inherent_impact"
	}

	where := []string{
		"org_id = $1",
		"is_template = FALSE",
		"status NOT IN ('closed', 'archived')",
	}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("category"); v != "" {
		where = append(where, fmt.Sprintf("category = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("status"); v != "" {
		statuses := strings.Split(v, ",")
		placeholders := []string{}
		for _, s := range statuses {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
			args = append(args, s)
			argN++
		}
		where = append(where, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	whereClause := strings.Join(where, " AND ")

	// Build the grid: query risks grouped by likelihood × impact
	query := fmt.Sprintf(`
		SELECT %s, %s, id, identifier, title, status
		FROM risks
		WHERE %s AND %s IS NOT NULL AND %s IS NOT NULL
		ORDER BY %s DESC, %s DESC
	`, likelihoodCol, impactCol, whereClause, likelihoodCol, impactCol, likelihoodCol, impactCol)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query heat map data")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get heat map data"))
		return
	}
	defer rows.Close()

	// Collect risks into cells
	type cellKey struct {
		likelihood, impact string
	}
	cellRisks := map[cellKey][]gin.H{}

	totalRisks := 0
	bySeverity := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}
	var totalScore float64
	appetiteBreaches := 0

	for rows.Next() {
		var likelihood, impact, id, identifier, title, status string
		if err := rows.Scan(&likelihood, &impact, &id, &identifier, &title, &status); err != nil {
			continue
		}
		key := cellKey{likelihood, impact}
		cellRisks[key] = append(cellRisks[key], gin.H{
			"id": id, "identifier": identifier, "title": title, "status": status,
		})

		score := float64(models.LikelihoodScore(likelihood) * models.ImpactScore(impact))
		severity := models.ScoreSeverity(score)
		bySeverity[severity]++
		totalScore += score
		totalRisks++
	}

	// Check appetite breaches separately
	database.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM risks
		WHERE %s AND residual_score IS NOT NULL AND risk_appetite_threshold IS NOT NULL AND residual_score > risk_appetite_threshold
	`, whereClause), args[:argN-1]...).Scan(&appetiteBreaches)

	// Build full 25-cell grid
	likelihoods := models.ValidLikelihoodLevels
	impacts := models.ValidImpactLevels
	grid := []gin.H{}

	for li := len(likelihoods) - 1; li >= 0; li-- {
		for ii := len(impacts) - 1; ii >= 0; ii-- {
			l := likelihoods[li]
			i := impacts[ii]
			lScore := models.LikelihoodScore(l)
			iScore := models.ImpactScore(i)
			score := lScore * iScore
			severity := models.ScoreSeverity(float64(score))

			key := cellKey{l, i}
			risks := cellRisks[key]
			if risks == nil {
				risks = []gin.H{}
			}

			grid = append(grid, gin.H{
				"likelihood":       l,
				"likelihood_score": lScore,
				"impact":           i,
				"impact_score":     iScore,
				"score":            score,
				"severity":         severity,
				"count":            len(risks),
				"risks":            risks,
			})
		}
	}

	avgScore := 0.0
	if totalRisks > 0 {
		avgScore = math.Round(totalScore/float64(totalRisks)*100) / 100
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"score_type": scoreType,
		"grid":       grid,
		"summary": gin.H{
			"total_risks":       totalRisks,
			"by_severity":       bySeverity,
			"average_score":     avgScore,
			"appetite_breaches": appetiteBreaches,
		},
	}))
}

// GetRiskGaps identifies risks with missing treatments or controls.
func GetRiskGaps(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	gapType := c.DefaultQuery("gap_type", "all")
	minSeverity := c.Query("min_severity")

	// Summary counts
	var totalActive, noTreatments, noControls, highNoControls, overdueAssess, expiredAccept int

	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE AND status NOT IN ('closed','archived')
	`, orgID).Scan(&totalActive)

	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks r WHERE r.org_id = $1 AND r.is_template = FALSE
		AND r.status NOT IN ('closed','archived')
		AND (SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id) = 0
	`, orgID).Scan(&noTreatments)

	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks r WHERE r.org_id = $1 AND r.is_template = FALSE
		AND r.status NOT IN ('closed','archived')
		AND (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) = 0
	`, orgID).Scan(&noControls)

	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks r WHERE r.org_id = $1 AND r.is_template = FALSE
		AND r.status NOT IN ('closed','archived')
		AND r.residual_score >= 12
		AND (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) = 0
	`, orgID).Scan(&highNoControls)

	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE
		AND status NOT IN ('closed','archived')
		AND next_assessment_at IS NOT NULL AND next_assessment_at < NOW()
	`, orgID).Scan(&overdueAssess)

	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND status = 'accepted'
		AND acceptance_expiry IS NOT NULL AND acceptance_expiry < NOW()
	`, orgID).Scan(&expiredAccept)

	// Build gap query
	gapWhere := []string{
		"r.org_id = $1",
		"r.is_template = FALSE",
		"r.status NOT IN ('closed','archived')",
	}
	gapArgs := []interface{}{orgID}
	gapArgN := 2

	// Filter by min severity
	if minSeverity != "" {
		switch minSeverity {
		case "critical":
			gapWhere = append(gapWhere, "r.residual_score >= 20")
		case "high":
			gapWhere = append(gapWhere, "r.residual_score >= 12")
		case "medium":
			gapWhere = append(gapWhere, "r.residual_score >= 6")
		}
	}

	// Filter by gap type
	gapConditions := []string{}
	switch gapType {
	case "no_treatments":
		gapConditions = append(gapConditions, "(SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id) = 0")
	case "no_controls":
		gapConditions = append(gapConditions, "(SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) = 0")
	case "high_without_controls":
		gapConditions = append(gapConditions, "r.residual_score >= 12 AND (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) = 0")
	case "overdue_assessment":
		gapConditions = append(gapConditions, "r.next_assessment_at IS NOT NULL AND r.next_assessment_at < NOW()")
	case "expired_acceptance":
		gapConditions = append(gapConditions, "r.status = 'accepted' AND r.acceptance_expiry IS NOT NULL AND r.acceptance_expiry < NOW()")
	default: // "all"
		gapConditions = append(gapConditions, `(
			(SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id) = 0
			OR (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) = 0
			OR (r.next_assessment_at IS NOT NULL AND r.next_assessment_at < NOW())
			OR (r.status = 'accepted' AND r.acceptance_expiry IS NOT NULL AND r.acceptance_expiry < NOW())
		)`)
	}

	allWhere := append(gapWhere, gapConditions...)
	whereClause := strings.Join(allWhere, " AND ")

	// Count total gaps
	var totalGaps int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM risks r WHERE %s", whereClause), gapArgs...).Scan(&totalGaps)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.identifier, r.title, r.category, r.status,
		       r.residual_score, r.owner_id, r.created_at,
		       r.next_assessment_at, r.acceptance_expiry,
		       COALESCE(u.first_name || ' ' || u.last_name, '') AS owner_name,
		       (SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id) AS treat_count,
		       (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) AS ctrl_count
		FROM risks r
		LEFT JOIN users u ON r.owner_id = u.id
		WHERE %s
		ORDER BY r.residual_score DESC NULLS LAST, r.created_at ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, gapArgN, gapArgN+1)
	gapArgs = append(gapArgs, perPage, offset)

	rows, err := database.DB.Query(query, gapArgs...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query risk gaps")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get gaps"))
		return
	}
	defer rows.Close()

	gaps := []gin.H{}
	for rows.Next() {
		var (
			id, identifier, title, category, status string
			residualScore                           *float64
			ownerID                                 *string
			createdAt                               time.Time
			nextAssess, acceptExpiry                *time.Time
			ownerName                               string
			treatCount, ctrlCount                   int
		)

		if err := rows.Scan(&id, &identifier, &title, &category, &status,
			&residualScore, &ownerID, &createdAt,
			&nextAssess, &acceptExpiry,
			&ownerName, &treatCount, &ctrlCount); err != nil {
			continue
		}

		severity := "low"
		if residualScore != nil {
			severity = models.ScoreSeverity(*residualScore)
		}

		// Determine gap types
		gapTypes := []string{}
		if treatCount == 0 {
			gapTypes = append(gapTypes, "no_treatments")
		}
		if ctrlCount == 0 {
			gapTypes = append(gapTypes, "no_controls")
		}
		if residualScore != nil && *residualScore >= 12 && ctrlCount == 0 {
			gapTypes = append(gapTypes, "high_without_controls")
		}
		if nextAssess != nil && nextAssess.Before(time.Now()) {
			gapTypes = append(gapTypes, "overdue_assessment")
		}
		if status == "accepted" && acceptExpiry != nil && acceptExpiry.Before(time.Now()) {
			gapTypes = append(gapTypes, "expired_acceptance")
		}

		daysOpen := int(time.Since(createdAt).Hours() / 24)

		// Generate recommendation
		recommendation := generateRecommendation(gapTypes, severity)

		var owner interface{}
		if ownerID != nil {
			owner = gin.H{"id": *ownerID, "name": ownerName}
		}

		gap := gin.H{
			"risk": gin.H{
				"id":             id,
				"identifier":    identifier,
				"title":         title,
				"category":      category,
				"status":        status,
				"residual_score": residualScore,
				"severity":      severity,
				"owner":         owner,
			},
			"gap_types":      gapTypes,
			"days_open":      daysOpen,
			"recommendation": recommendation,
		}

		if acceptExpiry != nil {
			daysUntilExpiry := int(time.Until(*acceptExpiry).Hours() / 24)
			gap["acceptance_expiry"] = acceptExpiry.Format("2006-01-02")
			gap["days_until_expiry"] = daysUntilExpiry
		}

		gaps = append(gaps, gap)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"summary": gin.H{
				"total_active_risks":          totalActive,
				"risks_without_treatments":    noTreatments,
				"risks_without_controls":      noControls,
				"high_risks_without_controls": highNoControls,
				"overdue_assessments":         overdueAssess,
				"expired_acceptances":         expiredAccept,
			},
			"gaps": gaps,
		},
		"meta": gin.H{
			"total_gaps": totalGaps,
			"page":       page,
			"per_page":   perPage,
		},
	})
}

func generateRecommendation(gapTypes []string, severity string) string {
	parts := []string{}
	hasNoTreatments := false
	hasNoControls := false

	for _, g := range gapTypes {
		switch g {
		case "no_treatments":
			hasNoTreatments = true
		case "no_controls":
			hasNoControls = true
		case "overdue_assessment":
			parts = append(parts, "Assessment is overdue — schedule reassessment.")
		case "expired_acceptance":
			parts = append(parts, "Risk acceptance has expired. Re-assess and either renew acceptance or create treatment plan.")
		}
	}

	if hasNoTreatments && hasNoControls {
		prefix := ""
		if severity == "high" || severity == "critical" {
			prefix = strings.Title(severity) + "-severity risk with "
		} else {
			prefix = "Risk with "
		}
		parts = append([]string{prefix + "no treatments or controls. Immediate action needed: create mitigation plan and link relevant controls."}, parts...)
	} else if hasNoTreatments {
		parts = append([]string{"No treatment plans exist. Create a mitigation plan."}, parts...)
	} else if hasNoControls {
		parts = append([]string{"No controls linked. Link relevant controls to track mitigation."}, parts...)
	}

	if len(parts) == 0 {
		return "Review recommended."
	}
	return strings.Join(parts, " ")
}

// SearchRisks provides full-text search across risks.
func SearchRisks(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Search query 'q' is required"))
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

	where := []string{"r.org_id = $1", "r.is_template = FALSE"}
	args := []interface{}{orgID}
	argN := 2

	// Full-text search
	where = append(where, fmt.Sprintf("(r.title ILIKE $%d OR r.description ILIKE $%d)", argN, argN))
	args = append(args, "%"+q+"%")
	argN++

	if v := c.Query("category"); v != "" {
		where = append(where, fmt.Sprintf("r.category = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("severity"); v != "" {
		switch v {
		case "critical":
			where = append(where, "r.residual_score >= 20")
		case "high":
			where = append(where, "r.residual_score >= 12 AND r.residual_score < 20")
		case "medium":
			where = append(where, "r.residual_score >= 6 AND r.residual_score < 12")
		case "low":
			where = append(where, "r.residual_score < 6")
		}
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM risks r WHERE %s", whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.identifier, r.title, r.description, r.category, r.status,
		       r.residual_score, r.owner_id,
		       COALESCE(u.first_name || ' ' || u.last_name, '') AS owner_name
		FROM risks r
		LEFT JOIN users u ON r.owner_id = u.id
		WHERE %s
		ORDER BY r.residual_score DESC NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search risks")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to search risks"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, identifier, title, category, status string
			description                             *string
			residualScore                           *float64
			ownerID                                 *string
			ownerName                               string
		)

		if err := rows.Scan(&id, &identifier, &title, &description, &category, &status,
			&residualScore, &ownerID, &ownerName); err != nil {
			continue
		}

		severity := "low"
		if residualScore != nil {
			severity = models.ScoreSeverity(*residualScore)
		}

		// Generate match context
		matchCtx := ""
		if description != nil {
			matchCtx = generateMatchContext(*description, q)
		}

		var owner interface{}
		if ownerID != nil {
			owner = gin.H{"id": *ownerID, "name": ownerName}
		}

		results = append(results, gin.H{
			"id":             id,
			"identifier":    identifier,
			"title":         title,
			"description":   description,
			"category":      category,
			"status":        status,
			"residual_score": residualScore,
			"severity":      severity,
			"match_context": matchCtx,
			"owner":         owner,
		})
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

func generateMatchContext(text, query string) string {
	lower := strings.ToLower(text)
	qLower := strings.ToLower(query)
	idx := strings.Index(lower, qLower)
	if idx == -1 {
		if len(text) > 200 {
			return text[:200] + "..."
		}
		return text
	}

	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 50
	if end > len(text) {
		end = len(text)
	}

	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(text) {
		suffix = "..."
	}

	return prefix + text[start:end] + suffix
}

// GetRiskStats returns risk management statistics.
func GetRiskStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	// Total risks (non-template)
	var totalRisks int
	database.DB.QueryRow("SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE", orgID).Scan(&totalRisks)

	// By status
	byStatus := map[string]int{}
	for _, s := range models.ValidRiskStatuses {
		byStatus[s] = 0
	}
	statusRows, _ := database.DB.Query("SELECT status, COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE GROUP BY status", orgID)
	if statusRows != nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var s string
			var cnt int
			statusRows.Scan(&s, &cnt)
			byStatus[s] = cnt
		}
	}

	// By category
	byCategory := map[string]int{}
	catRows, _ := database.DB.Query("SELECT category, COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE GROUP BY category ORDER BY COUNT(*) DESC", orgID)
	if catRows != nil {
		defer catRows.Close()
		for catRows.Next() {
			var cat string
			var cnt int
			catRows.Scan(&cat, &cnt)
			byCategory[cat] = cnt
		}
	}

	// By severity (from residual score)
	bySeverity := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}
	sevRows, _ := database.DB.Query(`
		SELECT
			CASE
				WHEN residual_score >= 20 THEN 'critical'
				WHEN residual_score >= 12 THEN 'high'
				WHEN residual_score >= 6 THEN 'medium'
				ELSE 'low'
			END as severity, COUNT(*)
		FROM risks WHERE org_id = $1 AND is_template = FALSE AND residual_score IS NOT NULL
		GROUP BY severity
	`, orgID)
	if sevRows != nil {
		defer sevRows.Close()
		for sevRows.Next() {
			var sev string
			var cnt int
			sevRows.Scan(&sev, &cnt)
			bySeverity[sev] = cnt
		}
	}

	// Scoring summary
	var avgInherent, avgResidual sql.NullFloat64
	database.DB.QueryRow(`
		SELECT AVG(inherent_score), AVG(residual_score)
		FROM risks WHERE org_id = $1 AND is_template = FALSE AND inherent_score IS NOT NULL
	`, orgID).Scan(&avgInherent, &avgResidual)

	avgReduction := 0.0
	if avgInherent.Valid && avgResidual.Valid && avgInherent.Float64 > 0 {
		avgReduction = math.Round((1-avgResidual.Float64/avgInherent.Float64)*10000) / 100
	}

	// Highest residual
	var highestRisk interface{}
	var hID, hIdentifier, hTitle string
	var hScore float64
	err := database.DB.QueryRow(`
		SELECT id, identifier, title, residual_score FROM risks
		WHERE org_id = $1 AND is_template = FALSE AND residual_score IS NOT NULL
		ORDER BY residual_score DESC LIMIT 1
	`, orgID).Scan(&hID, &hIdentifier, &hTitle, &hScore)
	if err == nil {
		highestRisk = gin.H{
			"id": hID, "identifier": hIdentifier, "title": hTitle,
			"score": hScore, "severity": models.ScoreSeverity(hScore),
		}
	}

	// Treatment summary
	treatSummary := gin.H{
		"total": 0, "planned": 0, "in_progress": 0, "implemented": 0,
		"verified": 0, "ineffective": 0, "cancelled": 0, "overdue": 0,
	}
	totalTreatments := 0
	treatRows, _ := database.DB.Query(`
		SELECT status, COUNT(*) FROM risk_treatments
		WHERE org_id = $1 GROUP BY status
	`, orgID)
	if treatRows != nil {
		defer treatRows.Close()
		for treatRows.Next() {
			var ts string
			var cnt int
			treatRows.Scan(&ts, &cnt)
			treatSummary[ts] = cnt
			totalTreatments += cnt
		}
	}
	treatSummary["total"] = totalTreatments

	// Overdue treatments
	var overdueTreatments int
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risk_treatments
		WHERE org_id = $1 AND due_date < NOW() AND status IN ('planned','in_progress')
	`, orgID).Scan(&overdueTreatments)
	treatSummary["overdue"] = overdueTreatments

	// Control coverage
	var withControls, withoutControls int
	var avgControlsPerRisk float64
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks r
		WHERE r.org_id = $1 AND r.is_template = FALSE AND r.status NOT IN ('closed','archived')
		AND (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) > 0
	`, orgID).Scan(&withControls)
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks r
		WHERE r.org_id = $1 AND r.is_template = FALSE AND r.status NOT IN ('closed','archived')
		AND (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) = 0
	`, orgID).Scan(&withoutControls)
	database.DB.QueryRow(`
		SELECT COALESCE(AVG(cnt), 0) FROM (
			SELECT COUNT(*) as cnt FROM risk_controls rc
			JOIN risks r ON rc.risk_id = r.id
			WHERE rc.org_id = $1 AND r.is_template = FALSE
			GROUP BY rc.risk_id
		) sub
	`, orgID).Scan(&avgControlsPerRisk)

	// Assessment health
	var overdueAssess, dueSoon, expiredAccept int
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE
		AND status NOT IN ('closed','archived') AND next_assessment_at < NOW()
	`, orgID).Scan(&overdueAssess)
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE
		AND status NOT IN ('closed','archived')
		AND next_assessment_at >= NOW() AND next_assessment_at < NOW() + INTERVAL '30 days'
	`, orgID).Scan(&dueSoon)
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND status = 'accepted'
		AND acceptance_expiry IS NOT NULL AND acceptance_expiry < NOW()
	`, orgID).Scan(&expiredAccept)

	// Appetite summary
	var withinAppetite, breaching, noThreshold int
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE
		AND status NOT IN ('closed','archived') AND risk_appetite_threshold IS NOT NULL
		AND (residual_score IS NULL OR residual_score <= risk_appetite_threshold)
	`, orgID).Scan(&withinAppetite)
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE
		AND status NOT IN ('closed','archived') AND risk_appetite_threshold IS NOT NULL
		AND residual_score IS NOT NULL AND residual_score > risk_appetite_threshold
	`, orgID).Scan(&breaching)
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = FALSE
		AND status NOT IN ('closed','archived') AND risk_appetite_threshold IS NULL
	`, orgID).Scan(&noThreshold)

	// Templates count
	var templatesAvailable int
	database.DB.QueryRow("SELECT COUNT(*) FROM risks WHERE org_id = $1 AND is_template = TRUE", orgID).Scan(&templatesAvailable)

	// Recent activity
	recentActivity := []gin.H{}
	actRows, _ := database.DB.Query(`
		SELECT al.action, al.resource_id,
		       COALESCE(u.first_name || ' ' || u.last_name, 'System'),
		       al.created_at
		FROM audit_log al
		LEFT JOIN users u ON al.actor_id = u.id
		WHERE al.org_id = $1 AND al.resource_type IN ('risk', 'risk_assessment', 'risk_treatment', 'risk_control')
		ORDER BY al.created_at DESC LIMIT 5
	`, orgID)
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var action string
			var resID *string
			var actor string
			var ts time.Time
			actRows.Scan(&action, &resID, &actor, &ts)

			// Try to get risk identifier
			riskIdent := ""
			if resID != nil {
				database.DB.QueryRow("SELECT identifier FROM risks WHERE id = $1", *resID).Scan(&riskIdent)
			}

			recentActivity = append(recentActivity, gin.H{
				"risk_identifier": riskIdent,
				"action":          action,
				"actor":           actor,
				"timestamp":       ts,
			})
		}
	}

	scoringSummary := gin.H{
		"average_inherent_score":  0.0,
		"average_residual_score":  0.0,
		"average_risk_reduction":  avgReduction,
		"highest_residual":        highestRisk,
	}
	if avgInherent.Valid {
		scoringSummary["average_inherent_score"] = math.Round(avgInherent.Float64*100) / 100
	}
	if avgResidual.Valid {
		scoringSummary["average_residual_score"] = math.Round(avgResidual.Float64*100) / 100
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"total_risks":      totalRisks,
		"by_status":        byStatus,
		"by_category":      byCategory,
		"by_severity":      bySeverity,
		"scoring_summary":  scoringSummary,
		"treatment_summary": treatSummary,
		"control_coverage": gin.H{
			"risks_with_controls":      withControls,
			"risks_without_controls":   withoutControls,
			"average_controls_per_risk": math.Round(avgControlsPerRisk*10) / 10,
		},
		"assessment_health": gin.H{
			"overdue_assessments":  overdueAssess,
			"due_within_30_days":   dueSoon,
			"expired_acceptances":  expiredAccept,
		},
		"appetite_summary": gin.H{
			"within_appetite":    withinAppetite,
			"breaching_appetite": breaching,
			"no_threshold_set":   noThreshold,
		},
		"templates_available": templatesAvailable,
		"recent_activity":     recentActivity,
	}))
}
