package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// validateOwnerInOrg checks that the given userID exists and belongs to the specified org.
func validateOwnerInOrg(userID, orgID string) (bool, error) {
	var exists bool
	err := database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2 AND status = 'active')`,
		userID, orgID,
	).Scan(&exists)
	return exists, err
}

// computeAssessmentStatus returns the assessment status based on next_assessment_at.
func computeAssessmentStatus(nextAssessmentAt *time.Time) string {
	if nextAssessmentAt == nil {
		return "no_schedule"
	}
	now := time.Now()
	if nextAssessmentAt.Before(now) {
		return "overdue"
	}
	if nextAssessmentAt.Before(now.Add(30 * 24 * time.Hour)) {
		return "due_soon"
	}
	return "on_track"
}

// ListRisks lists risks with filtering and pagination.
func ListRisks(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"r.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	// is_template filter (default: false)
	isTemplate := c.DefaultQuery("is_template", "false")
	if isTemplate == "true" {
		where = append(where, "r.is_template = TRUE")
	} else {
		where = append(where, "r.is_template = FALSE")
	}

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("category"); v != "" {
		where = append(where, fmt.Sprintf("r.category = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("owner_id"); v != "" {
		where = append(where, fmt.Sprintf("r.owner_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("severity"); v != "" {
		// Severity is computed from residual_score
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
	if v := c.Query("score_min"); v != "" {
		scoreType := c.DefaultQuery("score_type", "residual")
		col := "r.residual_score"
		if scoreType == "inherent" {
			col = "r.inherent_score"
		}
		where = append(where, fmt.Sprintf("%s >= $%d", col, argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("score_max"); v != "" {
		scoreType := c.DefaultQuery("score_type", "residual")
		col := "r.residual_score"
		if scoreType == "inherent" {
			col = "r.inherent_score"
		}
		where = append(where, fmt.Sprintf("%s <= $%d", col, argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("has_treatments"); v != "" {
		if v == "true" {
			where = append(where, "(SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id) > 0")
		} else {
			where = append(where, "(SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id) = 0")
		}
	}
	if v := c.Query("has_controls"); v != "" {
		if v == "true" {
			where = append(where, "(SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) > 0")
		} else {
			where = append(where, "(SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) = 0")
		}
	}
	if v := c.Query("overdue_assessment"); v == "true" {
		where = append(where, "r.next_assessment_at < NOW()")
	}
	if v := c.Query("tags"); v != "" {
		tags := strings.Split(v, ",")
		where = append(where, fmt.Sprintf("r.tags @> $%d", argN))
		args = append(args, pq.Array(tags))
		argN++
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(r.title ILIKE $%d OR r.description ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	// Sort
	sortCol := "r.residual_score"
	switch c.DefaultQuery("sort", "residual_score") {
	case "identifier":
		sortCol = "r.identifier"
	case "title":
		sortCol = "r.title"
	case "category":
		sortCol = "r.category"
	case "status":
		sortCol = "r.status"
	case "inherent_score":
		sortCol = "r.inherent_score"
	case "residual_score":
		sortCol = "r.residual_score"
	case "next_assessment_at":
		sortCol = "r.next_assessment_at"
	case "created_at":
		sortCol = "r.created_at"
	case "updated_at":
		sortCol = "r.updated_at"
	}

	order := "DESC"
	if strings.ToLower(c.DefaultQuery("order", "desc")) == "asc" {
		order = "ASC"
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM risks r WHERE %s", whereClause)
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count risks")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to count risks"))
		return
	}

	// Query
	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT r.id, r.identifier, r.title, r.description, r.category, r.status,
		       r.owner_id, r.secondary_owner_id,
		       r.inherent_likelihood, r.inherent_impact, r.inherent_score,
		       r.residual_likelihood, r.residual_impact, r.residual_score,
		       r.risk_appetite_threshold,
		       r.assessment_frequency_days, r.next_assessment_at, r.last_assessed_at,
		       r.source, r.affected_assets, r.tags,
		       r.created_at, r.updated_at,
		       COALESCE(u.first_name || ' ' || u.last_name, '') AS owner_name,
		       COALESCE(u.email, '') AS owner_email,
		       (SELECT COUNT(*) FROM risk_controls rc WHERE rc.risk_id = r.id) AS linked_controls_count,
		       (SELECT COUNT(*) FROM risk_treatments rt WHERE rt.risk_id = r.id AND rt.status NOT IN ('cancelled','verified','ineffective')) AS active_treatments_count
		FROM risks r
		LEFT JOIN users u ON r.owner_id = u.id
		WHERE %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list risks")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list risks"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, identifier, title, category, status string
			description, ownerID, secondaryOwnerID  *string
			inhLikelihood, inhImpact                *string
			resLikelihood, resImpact                *string
			inhScore, resScore, appetiteThreshold   *float64
			assessFreq                              *int
			nextAssessAt, lastAssessAt              *time.Time
			source                                  *string
			affectedAssets, tags                     pq.StringArray
			createdAt, updatedAt                    time.Time
			ownerName, ownerEmail                   string
			linkedControlsCount, activeTreatments   int
		)

		err := rows.Scan(
			&id, &identifier, &title, &description, &category, &status,
			&ownerID, &secondaryOwnerID,
			&inhLikelihood, &inhImpact, &inhScore,
			&resLikelihood, &resImpact, &resScore,
			&appetiteThreshold,
			&assessFreq, &nextAssessAt, &lastAssessAt,
			&source, &affectedAssets, &tags,
			&createdAt, &updatedAt,
			&ownerName, &ownerEmail,
			&linkedControlsCount, &activeTreatments,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan risk row")
			continue
		}

		// Build owner object
		var owner interface{}
		if ownerID != nil {
			owner = gin.H{"id": *ownerID, "name": ownerName, "email": ownerEmail}
		}

		// Build score objects
		var inherentScoreObj, residualScoreObj interface{}
		if inhLikelihood != nil && inhImpact != nil && inhScore != nil {
			inherentScoreObj = gin.H{
				"likelihood": *inhLikelihood,
				"impact":     *inhImpact,
				"score":      *inhScore,
				"severity":   models.ScoreSeverity(*inhScore),
			}
		}
		if resLikelihood != nil && resImpact != nil && resScore != nil {
			residualScoreObj = gin.H{
				"likelihood": *resLikelihood,
				"impact":     *resImpact,
				"score":      *resScore,
				"severity":   models.ScoreSeverity(*resScore),
			}
		}

		appetiteBreached := false
		if resScore != nil && appetiteThreshold != nil {
			appetiteBreached = *resScore > *appetiteThreshold
		}

		result := gin.H{
			"id":                     id,
			"identifier":             identifier,
			"title":                  title,
			"description":            description,
			"category":               category,
			"status":                 status,
			"owner":                  owner,
			"inherent_score":         inherentScoreObj,
			"residual_score":         residualScoreObj,
			"risk_appetite_threshold": appetiteThreshold,
			"appetite_breached":      appetiteBreached,
			"assessment_frequency_days": assessFreq,
			"next_assessment_at":     nextAssessAt,
			"last_assessed_at":       lastAssessAt,
			"assessment_status":      computeAssessmentStatus(nextAssessAt),
			"source":                 source,
			"affected_assets":        []string(affectedAssets),
			"linked_controls_count":  linkedControlsCount,
			"active_treatments_count": activeTreatments,
			"tags":                   []string(tags),
			"created_at":             createdAt,
			"updated_at":             updatedAt,
		}
		results = append(results, result)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// GetRisk gets a single risk with full details.
func GetRisk(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	riskID := c.Param("id")

	var (
		id, identifier, title, category, status string
		description, ownerID, secondaryOwnerID  *string
		inhLikelihood, inhImpact                *string
		resLikelihood, resImpact                *string
		inhScore, resScore, appetiteThreshold   *float64
		acceptedAt                              *time.Time
		acceptedBy, acceptJustification         *string
		acceptExpiry                            *time.Time
		assessFreq                              *int
		nextAssessAt, lastAssessAt              *time.Time
		source                                  *string
		affectedAssets, tags                     pq.StringArray
		isTemplate                              bool
		templateSourceID                        *string
		metadata                                string
		createdAt, updatedAt                    time.Time
		ownerName, ownerEmail                   string
		secondaryName, secondaryEmail           sql.NullString
	)

	err := database.DB.QueryRow(`
		SELECT r.id, r.identifier, r.title, r.description, r.category, r.status,
		       r.owner_id, r.secondary_owner_id,
		       r.inherent_likelihood, r.inherent_impact, r.inherent_score,
		       r.residual_likelihood, r.residual_impact, r.residual_score,
		       r.risk_appetite_threshold,
		       r.accepted_at, r.accepted_by, r.acceptance_justification, r.acceptance_expiry,
		       r.assessment_frequency_days, r.next_assessment_at, r.last_assessed_at,
		       r.source, r.affected_assets, r.is_template, r.template_source_id,
		       r.tags, COALESCE(r.metadata::text, '{}'),
		       r.created_at, r.updated_at,
		       COALESCE(u.first_name || ' ' || u.last_name, ''), COALESCE(u.email, ''),
		       u2.first_name || ' ' || u2.last_name, u2.email
		FROM risks r
		LEFT JOIN users u ON r.owner_id = u.id
		LEFT JOIN users u2 ON r.secondary_owner_id = u2.id
		WHERE r.id = $1 AND r.org_id = $2
	`, riskID, orgID).Scan(
		&id, &identifier, &title, &description, &category, &status,
		&ownerID, &secondaryOwnerID,
		&inhLikelihood, &inhImpact, &inhScore,
		&resLikelihood, &resImpact, &resScore,
		&appetiteThreshold,
		&acceptedAt, &acceptedBy, &acceptJustification, &acceptExpiry,
		&assessFreq, &nextAssessAt, &lastAssessAt,
		&source, &affectedAssets, &isTemplate, &templateSourceID,
		&tags, &metadata,
		&createdAt, &updatedAt,
		&ownerName, &ownerEmail,
		&secondaryName, &secondaryEmail,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get risk"))
		return
	}

	// Build owner
	var owner, secondaryOwner interface{}
	if ownerID != nil {
		owner = gin.H{"id": *ownerID, "name": ownerName, "email": ownerEmail}
	}
	if secondaryOwnerID != nil && secondaryName.Valid {
		secondaryOwner = gin.H{"id": *secondaryOwnerID, "name": secondaryName.String, "email": secondaryEmail.String}
	}

	// Build scores
	var inherentScoreObj, residualScoreObj interface{}
	if inhLikelihood != nil && inhImpact != nil && inhScore != nil {
		inherentScoreObj = gin.H{
			"likelihood":       *inhLikelihood,
			"likelihood_score": models.LikelihoodScore(*inhLikelihood),
			"impact":           *inhImpact,
			"impact_score":     models.ImpactScore(*inhImpact),
			"score":            *inhScore,
			"severity":         models.ScoreSeverity(*inhScore),
		}
	}
	if resLikelihood != nil && resImpact != nil && resScore != nil {
		residualScoreObj = gin.H{
			"likelihood":       *resLikelihood,
			"likelihood_score": models.LikelihoodScore(*resLikelihood),
			"impact":           *resImpact,
			"impact_score":     models.ImpactScore(*resImpact),
			"score":            *resScore,
			"severity":         models.ScoreSeverity(*resScore),
		}
	}

	appetiteBreached := false
	if resScore != nil && appetiteThreshold != nil {
		appetiteBreached = *resScore > *appetiteThreshold
	}

	// Acceptance
	var acceptance interface{}
	if status == models.RiskStatusAccepted && acceptedAt != nil {
		acceptanceObj := gin.H{
			"accepted_at":   acceptedAt,
			"justification": acceptJustification,
		}
		if acceptedBy != nil {
			// Fetch acceptor name
			var acceptorName string
			database.DB.QueryRow("SELECT COALESCE(first_name || ' ' || last_name, '') FROM users WHERE id = $1", *acceptedBy).Scan(&acceptorName)
			acceptanceObj["accepted_by"] = gin.H{"id": *acceptedBy, "name": acceptorName}
		}
		if acceptExpiry != nil {
			acceptanceObj["expiry"] = acceptExpiry.Format("2006-01-02")
		}
		acceptance = acceptanceObj
	}

	// Linked controls
	linkedControls := []gin.H{}
	ctrlRows, err := database.DB.Query(`
		SELECT rc.id, c.identifier, c.title, rc.effectiveness, rc.mitigation_percentage
		FROM risk_controls rc
		JOIN controls c ON rc.control_id = c.id
		WHERE rc.risk_id = $1 AND rc.org_id = $2
	`, riskID, orgID)
	if err == nil {
		defer ctrlRows.Close()
		for ctrlRows.Next() {
			var rcID, cIdentifier, cTitle string
			var eff string
			var mitPct *int
			ctrlRows.Scan(&rcID, &cIdentifier, &cTitle, &eff, &mitPct)
			linkedControls = append(linkedControls, gin.H{
				"id":                    rcID,
				"identifier":           cIdentifier,
				"title":                cTitle,
				"effectiveness":        eff,
				"mitigation_percentage": mitPct,
			})
		}
	}

	// Treatment summary
	treatmentSummary := gin.H{
		"total": 0, "planned": 0, "in_progress": 0,
		"implemented": 0, "verified": 0, "ineffective": 0, "cancelled": 0,
	}
	tRows, err := database.DB.Query(`
		SELECT status, COUNT(*) FROM risk_treatments
		WHERE risk_id = $1 AND org_id = $2 GROUP BY status
	`, riskID, orgID)
	if err == nil {
		defer tRows.Close()
		totalTreatments := 0
		for tRows.Next() {
			var tStatus string
			var cnt int
			tRows.Scan(&tStatus, &cnt)
			treatmentSummary[tStatus] = cnt
			totalTreatments += cnt
		}
		treatmentSummary["total"] = totalTreatments
	}

	// Latest assessments
	latestAssessments := gin.H{}
	for _, aType := range []string{"inherent", "residual"} {
		var aID, aJust string
		var aDate time.Time
		var aValidUntil *time.Time
		var assessorName string
		err := database.DB.QueryRow(`
			SELECT ra.id, COALESCE(ra.justification, ''), ra.assessment_date, ra.valid_until,
			       COALESCE(u.first_name || ' ' || u.last_name, '')
			FROM risk_assessments ra
			LEFT JOIN users u ON ra.assessed_by = u.id
			WHERE ra.risk_id = $1 AND ra.org_id = $2 AND ra.assessment_type = $3 AND ra.is_current = TRUE
			ORDER BY ra.assessment_date DESC LIMIT 1
		`, riskID, orgID, aType).Scan(&aID, &aJust, &aDate, &aValidUntil, &assessorName)
		if err == nil {
			entry := gin.H{
				"id":              aID,
				"assessment_date": aDate.Format("2006-01-02"),
				"assessor":        assessorName,
				"justification":   aJust,
			}
			if aValidUntil != nil {
				entry["valid_until"] = aValidUntil.Format("2006-01-02")
			}
			latestAssessments[aType] = entry
		}
	}

	result := gin.H{
		"id":                        id,
		"identifier":                identifier,
		"title":                     title,
		"description":               description,
		"category":                  category,
		"status":                    status,
		"owner":                     owner,
		"secondary_owner":           secondaryOwner,
		"inherent_score":            inherentScoreObj,
		"residual_score":            residualScoreObj,
		"risk_appetite_threshold":   appetiteThreshold,
		"appetite_breached":         appetiteBreached,
		"acceptance":                acceptance,
		"assessment_frequency_days": assessFreq,
		"next_assessment_at":        nextAssessAt,
		"last_assessed_at":          lastAssessAt,
		"assessment_status":         computeAssessmentStatus(nextAssessAt),
		"source":                    source,
		"affected_assets":           []string(affectedAssets),
		"is_template":               isTemplate,
		"template_source":           templateSourceID,
		"linked_controls":           linkedControls,
		"treatment_summary":         treatmentSummary,
		"latest_assessments":        latestAssessments,
		"tags":                      []string(tags),
		"metadata":                  metadata,
		"created_at":                createdAt,
		"updated_at":                updatedAt,
	}

	c.JSON(http.StatusOK, successResponse(c, result))
}

// CreateRisk creates a new risk.
func CreateRisk(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreateRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Validate category
	if !models.IsValidRiskCategory(req.Category) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid risk category"))
		return
	}

	// Validate title length
	if utf8.RuneCountInString(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must not exceed 500 characters"))
		return
	}

	// Validate appetite threshold
	if req.RiskAppetiteThreshold != nil && (*req.RiskAppetiteThreshold < 1 || *req.RiskAppetiteThreshold > 25) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Risk appetite threshold must be between 1 and 25"))
		return
	}

	// Validate initial assessment if provided
	if req.InitialAssessment != nil {
		if !models.IsValidLikelihood(req.InitialAssessment.InherentLikelihood) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid inherent_likelihood value"))
			return
		}
		if !models.IsValidImpact(req.InitialAssessment.InherentImpact) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid inherent_impact value"))
			return
		}
		if req.InitialAssessment.ResidualLikelihood != nil && !models.IsValidLikelihood(*req.InitialAssessment.ResidualLikelihood) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid residual_likelihood value"))
			return
		}
		if req.InitialAssessment.ResidualImpact != nil && !models.IsValidImpact(*req.InitialAssessment.ResidualImpact) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid residual_impact value"))
			return
		}
	}

	// Default owner to creator
	ownerID := req.OwnerID
	if ownerID == nil {
		ownerID = &userID
	}

	// Validate owner belongs to org
	if ownerID != nil && *ownerID != userID {
		exists, err := validateOwnerInOrg(*ownerID, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate owner_id")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to validate owner"))
			return
		}
		if !exists {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("VALIDATION_ERROR", "owner_id does not exist or does not belong to this organization"))
			return
		}
	}

	// Validate secondary owner belongs to org if provided
	if req.SecondaryOwnerID != nil {
		exists, err := validateOwnerInOrg(*req.SecondaryOwnerID, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate secondary_owner_id")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to validate secondary owner"))
			return
		}
		if !exists {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("VALIDATION_ERROR", "secondary_owner_id does not exist or does not belong to this organization"))
			return
		}
	}

	riskID := uuid.New().String()
	now := time.Now()

	// Check identifier uniqueness
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM risks WHERE org_id = $1 AND identifier = $2)", orgID, req.Identifier).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check risk identifier")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create risk"))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Risk identifier already exists in this organization"))
		return
	}

	// Insert risk
	_, err = database.DB.Exec(`
		INSERT INTO risks (id, org_id, identifier, title, description, category, status,
		                    owner_id, secondary_owner_id, risk_appetite_threshold,
		                    assessment_frequency_days, source, affected_assets, tags, metadata,
		                    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'identified',
		        $7, $8, $9, $10, $11, $12, $13, '{}', $14, $14)
	`, riskID, orgID, req.Identifier, req.Title, req.Description, req.Category,
		ownerID, req.SecondaryOwnerID, req.RiskAppetiteThreshold,
		req.AssessmentFrequencyDays, req.Source, pq.Array(req.AffectedAssets), pq.Array(req.Tags), now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create risk"))
		return
	}

	// Handle initial assessment
	var inhScoreObj, resScoreObj interface{}
	if req.InitialAssessment != nil {
		ia := req.InitialAssessment

		// Create inherent assessment
		inhLScore := models.LikelihoodScore(ia.InherentLikelihood)
		inhIScore := models.ImpactScore(ia.InherentImpact)
		inhOverall := float64(inhLScore * inhIScore)
		inhSeverity := models.ScoreSeverity(inhOverall)
		inhAssessID := uuid.New().String()

		_, err = database.DB.Exec(`
			INSERT INTO risk_assessments (id, org_id, risk_id, assessment_type, likelihood, impact,
			                              likelihood_score, impact_score, overall_score, scoring_formula,
			                              justification, assessed_by, assessment_date, is_current, created_at)
			VALUES ($1, $2, $3, 'inherent', $4, $5, $6, $7, $8, 'likelihood_x_impact', $9, $10, $11, TRUE, $11)
		`, inhAssessID, orgID, riskID, ia.InherentLikelihood, ia.InherentImpact,
			inhLScore, inhIScore, inhOverall, ia.Justification, userID, now)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create inherent assessment")
		}

		// Update risk with inherent scores
		database.DB.Exec(`
			UPDATE risks SET inherent_likelihood = $1, inherent_impact = $2, inherent_score = $3,
			                 last_assessed_at = $4, updated_at = $4
			WHERE id = $5
		`, ia.InherentLikelihood, ia.InherentImpact, inhOverall, now, riskID)

		inhScoreObj = gin.H{
			"likelihood": ia.InherentLikelihood,
			"impact":     ia.InherentImpact,
			"score":      inhOverall,
			"severity":   inhSeverity,
		}

		// Compute next_assessment_at
		if req.AssessmentFrequencyDays != nil {
			nextAssess := now.AddDate(0, 0, *req.AssessmentFrequencyDays)
			database.DB.Exec("UPDATE risks SET next_assessment_at = $1 WHERE id = $2", nextAssess, riskID)
		}

		// Create residual assessment if provided
		if ia.ResidualLikelihood != nil && ia.ResidualImpact != nil {
			resLScore := models.LikelihoodScore(*ia.ResidualLikelihood)
			resIScore := models.ImpactScore(*ia.ResidualImpact)
			resOverall := float64(resLScore * resIScore)
			resSeverity := models.ScoreSeverity(resOverall)
			resAssessID := uuid.New().String()

			_, err = database.DB.Exec(`
				INSERT INTO risk_assessments (id, org_id, risk_id, assessment_type, likelihood, impact,
				                              likelihood_score, impact_score, overall_score, scoring_formula,
				                              justification, assessed_by, assessment_date, is_current, created_at)
				VALUES ($1, $2, $3, 'residual', $4, $5, $6, $7, $8, 'likelihood_x_impact', $9, $10, $11, TRUE, $11)
			`, resAssessID, orgID, riskID, *ia.ResidualLikelihood, *ia.ResidualImpact,
				resLScore, resIScore, resOverall, ia.Justification, userID, now)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create residual assessment")
			}

			database.DB.Exec(`
				UPDATE risks SET residual_likelihood = $1, residual_impact = $2, residual_score = $3, updated_at = $4
				WHERE id = $5
			`, *ia.ResidualLikelihood, *ia.ResidualImpact, resOverall, now, riskID)

			resScoreObj = gin.H{
				"likelihood": *ia.ResidualLikelihood,
				"impact":     *ia.ResidualImpact,
				"score":      resOverall,
				"severity":   resSeverity,
			}
		}
	}

	// Audit log
	middleware.LogAudit(c, "risk.created", "risk", &riskID, map[string]interface{}{
		"identifier": req.Identifier,
		"category":   req.Category,
	})

	result := gin.H{
		"id":             riskID,
		"identifier":     req.Identifier,
		"title":          req.Title,
		"status":         "identified",
		"inherent_score": inhScoreObj,
		"residual_score": resScoreObj,
		"created_at":     now,
	}

	c.JSON(http.StatusCreated, successResponse(c, result))
}

// UpdateRisk updates risk metadata.
func UpdateRisk(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")

	var req models.UpdateRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Check risk exists and get current owner
	var currentOwnerID *string
	var currentStatus string
	err := database.DB.QueryRow("SELECT owner_id, status FROM risks WHERE id = $1 AND org_id = $2", riskID, orgID).Scan(&currentOwnerID, &currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk for update")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update risk"))
		return
	}

	// Authorization: compliance_manager, ciso, security_engineer, or owner
	isOwner := currentOwnerID != nil && *currentOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.RiskCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to update this risk"))
		return
	}

	// Build update query dynamically
	sets := []string{}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		if utf8.RuneCountInString(*req.Title) > 500 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must not exceed 500 characters"))
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
	if req.Category != nil {
		if !models.IsValidRiskCategory(*req.Category) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid risk category"))
			return
		}
		sets = append(sets, fmt.Sprintf("category = $%d", argN))
		args = append(args, *req.Category)
		argN++
	}
	if req.OwnerID != nil {
		// Validate owner belongs to org
		exists, err := validateOwnerInOrg(*req.OwnerID, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate owner_id")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to validate owner"))
			return
		}
		if !exists {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("VALIDATION_ERROR", "owner_id does not exist or does not belong to this organization"))
			return
		}

		oldOwner := currentOwnerID
		sets = append(sets, fmt.Sprintf("owner_id = $%d", argN))
		args = append(args, *req.OwnerID)
		argN++
		// Log owner change
		if oldOwner == nil || *oldOwner != *req.OwnerID {
			middleware.LogAudit(c, "risk.owner_changed", "risk", &riskID, map[string]interface{}{
				"old_owner_id": oldOwner,
				"new_owner_id": *req.OwnerID,
			})
		}
	}
	if req.SecondaryOwnerID != nil {
		// Validate secondary owner belongs to org
		exists, err := validateOwnerInOrg(*req.SecondaryOwnerID, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate secondary_owner_id")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to validate secondary owner"))
			return
		}
		if !exists {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("VALIDATION_ERROR", "secondary_owner_id does not exist or does not belong to this organization"))
			return
		}

		sets = append(sets, fmt.Sprintf("secondary_owner_id = $%d", argN))
		args = append(args, *req.SecondaryOwnerID)
		argN++
	}
	if req.RiskAppetiteThreshold != nil {
		if *req.RiskAppetiteThreshold < 1 || *req.RiskAppetiteThreshold > 25 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Risk appetite threshold must be between 1 and 25"))
			return
		}
		sets = append(sets, fmt.Sprintf("risk_appetite_threshold = $%d", argN))
		args = append(args, *req.RiskAppetiteThreshold)
		argN++
	}
	if req.AssessmentFrequencyDays != nil {
		sets = append(sets, fmt.Sprintf("assessment_frequency_days = $%d", argN))
		args = append(args, *req.AssessmentFrequencyDays)
		argN++
	}
	if req.NextAssessmentAt != nil {
		parsed, err := time.Parse("2006-01-02", *req.NextAssessmentAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid next_assessment_at format (use YYYY-MM-DD)"))
			return
		}
		sets = append(sets, fmt.Sprintf("next_assessment_at = $%d", argN))
		args = append(args, parsed)
		argN++
	}
	if req.Source != nil {
		sets = append(sets, fmt.Sprintf("source = $%d", argN))
		args = append(args, *req.Source)
		argN++
	}
	if req.AffectedAssets != nil {
		sets = append(sets, fmt.Sprintf("affected_assets = $%d", argN))
		args = append(args, pq.Array(req.AffectedAssets))
		argN++
	}
	if req.Tags != nil {
		sets = append(sets, fmt.Sprintf("tags = $%d", argN))
		args = append(args, pq.Array(req.Tags))
		argN++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "No fields to update"))
		return
	}

	sets = append(sets, fmt.Sprintf("updated_at = $%d", argN))
	args = append(args, time.Now())
	argN++

	args = append(args, riskID, orgID)
	query := fmt.Sprintf("UPDATE risks SET %s WHERE id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update risk"))
		return
	}

	middleware.LogAudit(c, "risk.updated", "risk", &riskID, nil)

	// Fetch updated risk for response
	var updatedTitle, updatedIdentifier, updatedStatus string
	var updatedAt time.Time
	database.DB.QueryRow("SELECT identifier, title, status, updated_at FROM risks WHERE id = $1", riskID).Scan(
		&updatedIdentifier, &updatedTitle, &updatedStatus, &updatedAt)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         riskID,
		"identifier": updatedIdentifier,
		"title":      updatedTitle,
		"status":     updatedStatus,
		"updated_at": updatedAt,
	}))
}

// ArchiveRisk archives a risk (soft-delete).
func ArchiveRisk(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")

	// RBAC: Only CISO or Compliance Manager can archive risks
	if !models.HasRole(userRole, models.RiskArchiveRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to archive risks"))
		return
	}

	var currentStatus string
	err := database.DB.QueryRow("SELECT status FROM risks WHERE id = $1 AND org_id = $2", riskID, orgID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk for archiving")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to archive risk"))
		return
	}

	if currentStatus == models.RiskStatusArchived {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST", "Risk is already archived"))
		return
	}

	now := time.Now()

	// Archive the risk
	_, err = database.DB.Exec("UPDATE risks SET status = 'archived', updated_at = $1 WHERE id = $2 AND org_id = $3", now, riskID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to archive risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to archive risk"))
		return
	}

	// Cancel active treatments
	database.DB.Exec(`
		UPDATE risk_treatments SET status = 'cancelled', updated_at = $1
		WHERE risk_id = $2 AND org_id = $3 AND status IN ('planned', 'in_progress')
	`, now, riskID, orgID)

	middleware.LogAudit(c, "risk.archived", "risk", &riskID, map[string]interface{}{"previous_status": currentStatus})

	var identifier string
	database.DB.QueryRow("SELECT identifier FROM risks WHERE id = $1", riskID).Scan(&identifier)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         riskID,
		"identifier": identifier,
		"status":     "archived",
		"updated_at": now,
	}))
}

// ChangeRiskStatus transitions risk status.
func ChangeRiskStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")

	var req models.ChangeRiskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Get current status and owner
	var currentStatus string
	var ownerID *string
	err := database.DB.QueryRow("SELECT status, owner_id FROM risks WHERE id = $1 AND org_id = $2", riskID, orgID).Scan(&currentStatus, &ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk for status change")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change risk status"))
		return
	}

	// Validate transition
	if !models.IsValidRiskStatusTransition(currentStatus, req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST",
			fmt.Sprintf("Invalid status transition from '%s' to '%s'", currentStatus, req.Status)))
		return
	}

	// Authorization: owner or authorized roles
	isOwner := ownerID != nil && *ownerID == userID
	if !isOwner && !models.HasRole(userRole, models.RiskCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to change risk status"))
		return
	}

	// Special handling for "accepted" status
	if req.Status == models.RiskStatusAccepted {
		// Only CISO or compliance_manager
		if !models.HasRole(userRole, models.RiskAcceptRoles) {
			c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Only CISO or compliance manager can accept risks"))
			return
		}
		if req.Justification == nil || *req.Justification == "" {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Justification is required when accepting a risk"))
			return
		}
		if req.AcceptanceExpiry == nil || *req.AcceptanceExpiry == "" {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Acceptance expiry date is required when accepting a risk"))
			return
		}

		expiry, err := time.Parse("2006-01-02", *req.AcceptanceExpiry)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid acceptance_expiry format (use YYYY-MM-DD)"))
			return
		}

		now := time.Now()
		_, err = database.DB.Exec(`
			UPDATE risks SET status = 'accepted', accepted_at = $1, accepted_by = $2,
			                 acceptance_justification = $3, acceptance_expiry = $4, updated_at = $1
			WHERE id = $5 AND org_id = $6
		`, now, userID, *req.Justification, expiry, riskID, orgID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to accept risk")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change risk status"))
			return
		}

		middleware.LogAudit(c, "risk.status_changed", "risk", &riskID, map[string]interface{}{
			"from": currentStatus, "to": "accepted",
		})

		var userName string
		database.DB.QueryRow("SELECT COALESCE(first_name || ' ' || last_name, '') FROM users WHERE id = $1", userID).Scan(&userName)

		c.JSON(http.StatusOK, successResponse(c, gin.H{
			"id":         riskID,
			"status":     "accepted",
			"acceptance": gin.H{
				"accepted_at":   now,
				"accepted_by":   gin.H{"id": userID, "name": userName},
				"expiry":        expiry.Format("2006-01-02"),
				"justification": *req.Justification,
			},
			"updated_at": now,
		}))
		return
	}

	// Normal status transition
	now := time.Now()
	_, err = database.DB.Exec(`
		UPDATE risks SET status = $1, updated_at = $2 WHERE id = $3 AND org_id = $4
	`, req.Status, now, riskID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to change risk status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to change risk status"))
		return
	}

	// Clear acceptance fields when moving from accepted to active status
	if currentStatus == models.RiskStatusAccepted {
		database.DB.Exec(`
			UPDATE risks SET accepted_at = NULL, accepted_by = NULL,
			                 acceptance_justification = NULL, acceptance_expiry = NULL
			WHERE id = $1
		`, riskID)
	}

	middleware.LogAudit(c, "risk.status_changed", "risk", &riskID, map[string]interface{}{
		"from": currentStatus, "to": req.Status,
	})

	var identifier string
	database.DB.QueryRow("SELECT identifier FROM risks WHERE id = $1", riskID).Scan(&identifier)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         riskID,
		"identifier": identifier,
		"status":     req.Status,
		"updated_at": now,
	}))
}
