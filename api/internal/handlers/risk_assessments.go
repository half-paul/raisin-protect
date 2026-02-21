package handlers

import (
	"database/sql"
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

// ListRiskAssessments lists assessments for a risk.
func ListRiskAssessments(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	riskID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Verify risk exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM risks WHERE id = $1 AND org_id = $2)", riskID, orgID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}

	where := []string{"ra.risk_id = $1", "ra.org_id = $2"}
	args := []interface{}{riskID, orgID}
	argN := 3

	if v := c.Query("assessment_type"); v != "" {
		where = append(where, fmt.Sprintf("ra.assessment_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if c.Query("current_only") == "true" {
		where = append(where, "ra.is_current = TRUE")
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	// Count
	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM risk_assessments ra %s", whereClause), args...).Scan(&total)

	// Query
	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ra.id, ra.risk_id, ra.assessment_type, ra.likelihood, ra.impact,
		       ra.likelihood_score, ra.impact_score, ra.overall_score, ra.scoring_formula,
		       ra.justification, ra.assumptions, ra.data_sources,
		       ra.assessed_by, ra.assessment_date, ra.valid_until, ra.is_current, ra.superseded_by,
		       ra.created_at,
		       COALESCE(u.first_name || ' ' || u.last_name, '') AS assessor_name
		FROM risk_assessments ra
		LEFT JOIN users u ON ra.assessed_by = u.id
		%s
		ORDER BY ra.assessment_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list risk assessments")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list assessments"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, raRiskID, assessType, likelihood, impact string
			lScore, iScore                               int
			overallScore                                 float64
			formula                                      string
			justification, assumptions                   *string
			dataSources                                  pq.StringArray
			assessedBy                                   string
			assessmentDate                               time.Time
			validUntil                                   *time.Time
			isCurrent                                    bool
			supersededBy                                 *string
			createdAt                                    time.Time
			assessorName                                 string
		)

		err := rows.Scan(
			&id, &raRiskID, &assessType, &likelihood, &impact,
			&lScore, &iScore, &overallScore, &formula,
			&justification, &assumptions, &dataSources,
			&assessedBy, &assessmentDate, &validUntil, &isCurrent, &supersededBy,
			&createdAt,
			&assessorName,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan assessment row")
			continue
		}

		severity := models.ScoreSeverity(overallScore)

		result := gin.H{
			"id":               id,
			"risk_id":          raRiskID,
			"assessment_type":  assessType,
			"likelihood":       likelihood,
			"impact":           impact,
			"likelihood_score": lScore,
			"impact_score":     iScore,
			"overall_score":    overallScore,
			"scoring_formula":  formula,
			"severity":         severity,
			"justification":    justification,
			"assumptions":      assumptions,
			"data_sources":     []string(dataSources),
			"assessed_by":      gin.H{"id": assessedBy, "name": assessorName},
			"assessment_date":  assessmentDate.Format("2006-01-02"),
			"valid_until":      nil,
			"is_current":       isCurrent,
			"superseded_by":    supersededBy,
			"created_at":       createdAt,
		}
		if validUntil != nil {
			result["valid_until"] = validUntil.Format("2006-01-02")
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// CreateRiskAssessment creates a new assessment.
func CreateRiskAssessment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")

	var req models.CreateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Validate
	if !models.IsValidAssessmentType(req.AssessmentType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid assessment_type"))
		return
	}
	if !models.IsValidLikelihood(req.Likelihood) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid likelihood value"))
		return
	}
	if !models.IsValidImpact(req.Impact) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid impact value"))
		return
	}

	formula := "likelihood_x_impact"
	if req.ScoringFormula != nil {
		formula = *req.ScoringFormula
	}
	if formula != "likelihood_x_impact" {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Only 'likelihood_x_impact' formula is currently supported"))
		return
	}

	// Check risk exists and get owner
	var ownerID *string
	err := database.DB.QueryRow("SELECT owner_id FROM risks WHERE id = $1 AND org_id = $2", riskID, orgID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create assessment"))
		return
	}

	// Authorization: owner or authorized roles
	isOwner := ownerID != nil && *ownerID == userID
	if !isOwner && !models.HasRole(userRole, models.RiskAssessRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to assess this risk"))
		return
	}

	// Compute scores
	lScore := models.LikelihoodScore(req.Likelihood)
	iScore := models.ImpactScore(req.Impact)
	overallScore := float64(lScore * iScore)
	severity := models.ScoreSeverity(overallScore)

	assessID := uuid.New().String()
	now := time.Now()

	var validUntil *time.Time
	if req.ValidUntil != nil {
		parsed, err := time.Parse("2006-01-02", *req.ValidUntil)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid valid_until format (use YYYY-MM-DD)"))
			return
		}
		validUntil = &parsed
	}

	// Mark previous current assessment of same type as not current
	var prevID *string
	database.DB.QueryRow(`
		SELECT id FROM risk_assessments
		WHERE risk_id = $1 AND org_id = $2 AND assessment_type = $3 AND is_current = TRUE
		ORDER BY assessment_date DESC LIMIT 1
	`, riskID, orgID, req.AssessmentType).Scan(&prevID)

	if prevID != nil {
		database.DB.Exec("UPDATE risk_assessments SET is_current = FALSE, superseded_by = $1 WHERE id = $2", assessID, *prevID)
	}

	// Insert assessment
	_, err = database.DB.Exec(`
		INSERT INTO risk_assessments (id, org_id, risk_id, assessment_type, likelihood, impact,
		                              likelihood_score, impact_score, overall_score, scoring_formula,
		                              justification, assumptions, data_sources,
		                              assessed_by, assessment_date, valid_until, is_current, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, TRUE, $15)
	`, assessID, orgID, riskID, req.AssessmentType, req.Likelihood, req.Impact,
		lScore, iScore, overallScore, formula,
		req.Justification, req.Assumptions, pq.Array(req.DataSources),
		userID, now, validUntil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create assessment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create assessment"))
		return
	}

	// Denormalize scores onto risks table
	var riskUpdated gin.H
	switch req.AssessmentType {
	case models.AssessmentTypeInherent:
		database.DB.Exec(`
			UPDATE risks SET inherent_likelihood = $1, inherent_impact = $2, inherent_score = $3,
			                 last_assessed_at = $4, updated_at = $4
			WHERE id = $5 AND org_id = $6
		`, req.Likelihood, req.Impact, overallScore, now, riskID, orgID)
	case models.AssessmentTypeResidual:
		database.DB.Exec(`
			UPDATE risks SET residual_likelihood = $1, residual_impact = $2, residual_score = $3,
			                 last_assessed_at = $4, updated_at = $4
			WHERE id = $5 AND org_id = $6
		`, req.Likelihood, req.Impact, overallScore, now, riskID, orgID)

		// Check appetite breach for response
		var threshold *float64
		database.DB.QueryRow("SELECT risk_appetite_threshold FROM risks WHERE id = $1", riskID).Scan(&threshold)
		breached := false
		if threshold != nil {
			breached = overallScore > *threshold
		}
		riskUpdated = gin.H{
			"residual_likelihood": req.Likelihood,
			"residual_impact":     req.Impact,
			"residual_score":      overallScore,
			"appetite_breached":   breached,
		}
	}

	// Update next_assessment_at
	var assessFreq *int
	database.DB.QueryRow("SELECT assessment_frequency_days FROM risks WHERE id = $1", riskID).Scan(&assessFreq)
	if assessFreq != nil && *assessFreq > 0 {
		nextAssess := now.AddDate(0, 0, *assessFreq)
		database.DB.Exec("UPDATE risks SET next_assessment_at = $1, last_assessed_at = $2 WHERE id = $3", nextAssess, now, riskID)
	}

	middleware.LogAudit(c, "risk_assessment.created", "risk_assessment", &assessID, map[string]interface{}{
		"risk_id":         riskID,
		"assessment_type": req.AssessmentType,
		"score":           overallScore,
		"severity":        severity,
	})

	// Get assessor name
	var assessorName string
	database.DB.QueryRow("SELECT COALESCE(first_name || ' ' || last_name, '') FROM users WHERE id = $1", userID).Scan(&assessorName)

	result := gin.H{
		"id":               assessID,
		"risk_id":          riskID,
		"assessment_type":  req.AssessmentType,
		"likelihood":       req.Likelihood,
		"impact":           req.Impact,
		"likelihood_score": lScore,
		"impact_score":     iScore,
		"overall_score":    overallScore,
		"severity":         severity,
		"scoring_formula":  formula,
		"is_current":       true,
		"assessed_by":      gin.H{"id": userID, "name": assessorName},
		"assessment_date":  now.Format("2006-01-02"),
		"created_at":       now,
	}
	if riskUpdated != nil {
		result["risk_updated"] = riskUpdated
	}

	c.JSON(http.StatusCreated, successResponse(c, result))
}

// RecalculateRiskScores recalculates denormalized scores from assessments.
func RecalculateRiskScores(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")

	// RBAC: Only authorized roles can recalculate scores
	if !models.HasRole(userRole, models.RiskRecalcRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to recalculate risk scores"))
		return
	}

	// Check risk exists
	var identifier string
	err := database.DB.QueryRow("SELECT identifier FROM risks WHERE id = $1 AND org_id = $2", riskID, orgID).Scan(&identifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk for recalculation")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to recalculate"))
		return
	}

	now := time.Now()

	// Read current inherent assessment
	var inhL, inhI *string
	var inhScore *float64
	database.DB.QueryRow(`
		SELECT likelihood, impact, overall_score FROM risk_assessments
		WHERE risk_id = $1 AND org_id = $2 AND assessment_type = 'inherent' AND is_current = TRUE
		ORDER BY assessment_date DESC LIMIT 1
	`, riskID, orgID).Scan(&inhL, &inhI, &inhScore)

	// Read current residual assessment
	var resL, resI *string
	var resScore *float64
	database.DB.QueryRow(`
		SELECT likelihood, impact, overall_score FROM risk_assessments
		WHERE risk_id = $1 AND org_id = $2 AND assessment_type = 'residual' AND is_current = TRUE
		ORDER BY assessment_date DESC LIMIT 1
	`, riskID, orgID).Scan(&resL, &resI, &resScore)

	// Update denormalized fields
	database.DB.Exec(`
		UPDATE risks SET
			inherent_likelihood = $1, inherent_impact = $2, inherent_score = $3,
			residual_likelihood = $4, residual_impact = $5, residual_score = $6,
			updated_at = $7
		WHERE id = $8 AND org_id = $9
	`, inhL, inhI, inhScore, resL, resI, resScore, now, riskID, orgID)

	middleware.LogAudit(c, "risk.score_recalculated", "risk", &riskID, nil)

	// Build response
	var inhScoreObj, resScoreObj interface{}
	if inhL != nil && inhI != nil && inhScore != nil {
		inhScoreObj = gin.H{
			"likelihood": *inhL,
			"impact":     *inhI,
			"score":      *inhScore,
			"severity":   models.ScoreSeverity(*inhScore),
		}
	}
	if resL != nil && resI != nil && resScore != nil {
		resScoreObj = gin.H{
			"likelihood": *resL,
			"impact":     *resI,
			"score":      *resScore,
			"severity":   models.ScoreSeverity(*resScore),
		}
	}

	var threshold *float64
	database.DB.QueryRow("SELECT risk_appetite_threshold FROM risks WHERE id = $1", riskID).Scan(&threshold)
	breached := false
	if resScore != nil && threshold != nil {
		breached = *resScore > *threshold
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                identifier,
		"identifier":        identifier,
		"inherent_score":    inhScoreObj,
		"residual_score":    resScoreObj,
		"appetite_breached": breached,
		"recalculated_at":   now,
	}))
}
// end of file
