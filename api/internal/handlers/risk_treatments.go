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
	"github.com/rs/zerolog/log"
)

// ListRiskTreatments lists treatment plans for a risk.
func ListRiskTreatments(c *gin.Context) {
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

	where := []string{"rt.risk_id = $1", "rt.org_id = $2"}
	args := []interface{}{riskID, orgID}
	argN := 3

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("rt.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("treatment_type"); v != "" {
		where = append(where, fmt.Sprintf("rt.treatment_type = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM risk_treatments rt %s", whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT rt.id, rt.risk_id, rt.treatment_type, rt.title, rt.description, rt.status,
		       rt.owner_id, rt.priority, rt.due_date, rt.started_at, rt.completed_at,
		       rt.estimated_effort_hours, rt.actual_effort_hours,
		       rt.effectiveness_rating, rt.effectiveness_notes,
		       rt.expected_residual_likelihood, rt.expected_residual_impact, rt.expected_residual_score,
		       rt.target_control_id, rt.notes,
		       rt.created_by, rt.created_at, rt.updated_at,
		       COALESCE(u.first_name || ' ' || u.last_name, '') AS owner_name,
		       COALESCE(uc.first_name || ' ' || uc.last_name, '') AS creator_name,
		       c.identifier AS ctrl_identifier, c.title AS ctrl_title
		FROM risk_treatments rt
		LEFT JOIN users u ON rt.owner_id = u.id
		LEFT JOIN users uc ON rt.created_by = uc.id
		LEFT JOIN controls c ON rt.target_control_id = c.id
		%s
		ORDER BY rt.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list risk treatments")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list treatments"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, rtRiskID, treatType, title, status string
			description                            *string
			ownerID                                *string
			priority                               string
			dueDate, startedAt, completedAt        *time.Time
			estEffort, actEffort                   *float64
			effRating, effNotes                    *string
			expResLikelihood, expResImpact         *string
			expResScore                            *float64
			targetCtrlID                           *string
			notes                                  *string
			createdBy                              string
			createdAt, updatedAt                   time.Time
			ownerName, creatorName                 string
			ctrlIdentifier, ctrlTitle              sql.NullString
		)

		err := rows.Scan(
			&id, &rtRiskID, &treatType, &title, &description, &status,
			&ownerID, &priority, &dueDate, &startedAt, &completedAt,
			&estEffort, &actEffort,
			&effRating, &effNotes,
			&expResLikelihood, &expResImpact, &expResScore,
			&targetCtrlID, &notes,
			&createdBy, &createdAt, &updatedAt,
			&ownerName, &creatorName,
			&ctrlIdentifier, &ctrlTitle,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan treatment row")
			continue
		}

		var owner interface{}
		if ownerID != nil {
			owner = gin.H{"id": *ownerID, "name": ownerName}
		}

		var expResidual interface{}
		if expResLikelihood != nil && expResImpact != nil && expResScore != nil {
			expResidual = gin.H{
				"likelihood": *expResLikelihood,
				"impact":     *expResImpact,
				"score":      *expResScore,
			}
		}

		var targetCtrl interface{}
		if targetCtrlID != nil && ctrlIdentifier.Valid {
			targetCtrl = gin.H{
				"id":         *targetCtrlID,
				"identifier": ctrlIdentifier.String,
				"title":      ctrlTitle.String,
			}
		}

		result := gin.H{
			"id":                     id,
			"risk_id":               rtRiskID,
			"treatment_type":        treatType,
			"title":                 title,
			"description":           description,
			"status":                status,
			"owner":                 owner,
			"priority":              priority,
			"due_date":              dueDate,
			"started_at":            startedAt,
			"completed_at":          completedAt,
			"estimated_effort_hours": estEffort,
			"actual_effort_hours":   actEffort,
			"effectiveness_rating":  effRating,
			"effectiveness_notes":   effNotes,
			"expected_residual":     expResidual,
			"target_control":        targetCtrl,
			"notes":                 notes,
			"created_by":            gin.H{"id": createdBy, "name": creatorName},
			"created_at":            createdAt,
			"updated_at":            updatedAt,
		}
		results = append(results, result)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// CreateRiskTreatment creates a treatment plan.
func CreateRiskTreatment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")

	var req models.CreateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Validate treatment type
	if !models.IsValidTreatmentType(req.TreatmentType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid treatment_type"))
		return
	}

	// Validate title length
	if utf8.RuneCountInString(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must not exceed 500 characters"))
		return
	}

	// Validate priority
	priority := "medium"
	if req.Priority != nil {
		if !models.IsValidPriority(*req.Priority) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid priority"))
			return
		}
		priority = *req.Priority
	}

	// Check risk exists and isn't archived
	var riskOwnerID *string
	var riskStatus string
	err := database.DB.QueryRow("SELECT owner_id, status FROM risks WHERE id = $1 AND org_id = $2", riskID, orgID).Scan(&riskOwnerID, &riskStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Risk not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get risk")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create treatment"))
		return
	}
	if riskStatus == models.RiskStatusArchived {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Cannot add treatments to an archived risk"))
		return
	}

	// Authorization: owner or authorized roles
	isOwner := riskOwnerID != nil && *riskOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.RiskCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to create treatments for this risk"))
		return
	}

	// Default owner to risk owner
	ownerID := req.OwnerID
	if ownerID == nil {
		ownerID = riskOwnerID
	}

	// Validate owner belongs to org if explicitly provided
	if req.OwnerID != nil {
		var ownerExists bool
		err = database.DB.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2 AND status = 'active')`,
			*req.OwnerID, orgID,
		).Scan(&ownerExists)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate owner_id")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to validate owner"))
			return
		}
		if !ownerExists {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("VALIDATION_ERROR", "owner_id does not exist or does not belong to this organization"))
			return
		}
	}

	// Parse due date
	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid due_date format (use YYYY-MM-DD)"))
			return
		}
		dueDate = &parsed
	}

	// Compute expected residual score
	var expResScore *float64
	if req.ExpectedResidualLikelihood != nil && req.ExpectedResidualImpact != nil {
		if !models.IsValidLikelihood(*req.ExpectedResidualLikelihood) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid expected_residual_likelihood"))
			return
		}
		if !models.IsValidImpact(*req.ExpectedResidualImpact) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid expected_residual_impact"))
			return
		}
		score := float64(models.LikelihoodScore(*req.ExpectedResidualLikelihood) * models.ImpactScore(*req.ExpectedResidualImpact))
		expResScore = &score
	}

	// Validate target_control_id if provided
	if req.TargetControlID != nil {
		var ctrlExists bool
		database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM controls WHERE id = $1 AND org_id = $2)", *req.TargetControlID, orgID).Scan(&ctrlExists)
		if !ctrlExists {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Target control not found"))
			return
		}
	}

	treatmentID := uuid.New().String()
	now := time.Now()

	_, err = database.DB.Exec(`
		INSERT INTO risk_treatments (id, org_id, risk_id, treatment_type, title, description, status,
		                             owner_id, priority, due_date, estimated_effort_hours,
		                             expected_residual_likelihood, expected_residual_impact, expected_residual_score,
		                             target_control_id, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'planned', $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $17)
	`, treatmentID, orgID, riskID, req.TreatmentType, req.Title, req.Description,
		ownerID, priority, dueDate, req.EstimatedEffortHours,
		req.ExpectedResidualLikelihood, req.ExpectedResidualImpact, expResScore,
		req.TargetControlID, req.Notes, userID, now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create treatment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create treatment"))
		return
	}

	// Auto-transition risk to "treating" if it's "identified" or "open"
	if riskStatus == models.RiskStatusIdentified || riskStatus == models.RiskStatusOpen {
		database.DB.Exec("UPDATE risks SET status = 'treating', updated_at = $1 WHERE id = $2", now, riskID)
		middleware.LogAudit(c, "risk.status_changed", "risk", &riskID, map[string]interface{}{
			"from": riskStatus, "to": "treating", "trigger": "treatment_created",
		})
	}

	middleware.LogAudit(c, "risk_treatment.created", "risk_treatment", &treatmentID, map[string]interface{}{
		"risk_id":        riskID,
		"treatment_type": req.TreatmentType,
		"priority":       priority,
	})

	var expResidual interface{}
	if expResScore != nil {
		expResidual = gin.H{
			"likelihood": *req.ExpectedResidualLikelihood,
			"impact":     *req.ExpectedResidualImpact,
			"score":      *expResScore,
		}
	}

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":                treatmentID,
		"risk_id":           riskID,
		"treatment_type":    req.TreatmentType,
		"title":             req.Title,
		"status":            "planned",
		"priority":          priority,
		"due_date":          dueDate,
		"expected_residual": expResidual,
		"created_at":        now,
	}))
}

// UpdateRiskTreatment updates a treatment plan.
func UpdateRiskTreatment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")
	treatmentID := c.Param("treatment_id")

	var req models.UpdateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	// Get current treatment info
	var currentStatus string
	var riskOwnerID, treatOwnerID *string
	err := database.DB.QueryRow(`
		SELECT rt.status, rt.owner_id, r.owner_id
		FROM risk_treatments rt
		JOIN risks r ON rt.risk_id = r.id
		WHERE rt.id = $1 AND rt.risk_id = $2 AND rt.org_id = $3
	`, treatmentID, riskID, orgID).Scan(&currentStatus, &treatOwnerID, &riskOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Treatment not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get treatment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update treatment"))
		return
	}

	// Authorization: risk owner, treatment owner, or authorized roles
	isRiskOwner := riskOwnerID != nil && *riskOwnerID == userID
	isTreatOwner := treatOwnerID != nil && *treatOwnerID == userID
	if !isRiskOwner && !isTreatOwner && !models.HasRole(userRole, models.RiskCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to update this treatment"))
		return
	}

	// Validate status transition if provided
	if req.Status != nil {
		if !models.IsValidTreatmentStatusTransition(currentStatus, *req.Status) {
			c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST",
				fmt.Sprintf("Invalid treatment status transition from '%s' to '%s'", currentStatus, *req.Status)))
			return
		}
	}

	// Build dynamic update
	sets := []string{}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argN))
		args = append(args, *req.Title)
		argN++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argN))
		args = append(args, *req.Description)
		argN++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argN))
		args = append(args, *req.Status)
		argN++
	}
	if req.OwnerID != nil {
		sets = append(sets, fmt.Sprintf("owner_id = $%d", argN))
		args = append(args, *req.OwnerID)
		argN++
	}
	if req.Priority != nil {
		if !models.IsValidPriority(*req.Priority) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid priority"))
			return
		}
		sets = append(sets, fmt.Sprintf("priority = $%d", argN))
		args = append(args, *req.Priority)
		argN++
	}
	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid due_date format"))
			return
		}
		sets = append(sets, fmt.Sprintf("due_date = $%d", argN))
		args = append(args, parsed)
		argN++
	}
	if req.StartedAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.StartedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid started_at format"))
			return
		}
		sets = append(sets, fmt.Sprintf("started_at = $%d", argN))
		args = append(args, parsed)
		argN++
	}
	if req.EstimatedEffortHours != nil {
		sets = append(sets, fmt.Sprintf("estimated_effort_hours = $%d", argN))
		args = append(args, *req.EstimatedEffortHours)
		argN++
	}
	if req.ActualEffortHours != nil {
		sets = append(sets, fmt.Sprintf("actual_effort_hours = $%d", argN))
		args = append(args, *req.ActualEffortHours)
		argN++
	}
	if req.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes = $%d", argN))
		args = append(args, *req.Notes)
		argN++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "No fields to update"))
		return
	}

	now := time.Now()
	sets = append(sets, fmt.Sprintf("updated_at = $%d", argN))
	args = append(args, now)
	argN++

	args = append(args, treatmentID, riskID, orgID)
	query := fmt.Sprintf("UPDATE risk_treatments SET %s WHERE id = $%d AND risk_id = $%d AND org_id = $%d",
		strings.Join(sets, ", "), argN, argN+1, argN+2)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update treatment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update treatment"))
		return
	}

	if req.Status != nil {
		middleware.LogAudit(c, "risk_treatment.status_changed", "risk_treatment", &treatmentID, map[string]interface{}{
			"from": currentStatus, "to": *req.Status,
		})
	} else {
		middleware.LogAudit(c, "risk_treatment.updated", "risk_treatment", &treatmentID, nil)
	}

	// Fetch updated treatment for response
	var updatedTitle, updatedType, updatedStatus string
	var updatedStartedAt *time.Time
	var updatedUpdatedAt time.Time
	database.DB.QueryRow("SELECT treatment_type, title, status, started_at, updated_at FROM risk_treatments WHERE id = $1",
		treatmentID).Scan(&updatedType, &updatedTitle, &updatedStatus, &updatedStartedAt, &updatedUpdatedAt)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":             treatmentID,
		"treatment_type": updatedType,
		"title":          updatedTitle,
		"status":         updatedStatus,
		"started_at":     updatedStartedAt,
		"updated_at":     updatedUpdatedAt,
	}))
}

// CompleteTreatment marks a treatment as complete.
func CompleteTreatment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	riskID := c.Param("id")
	treatmentID := c.Param("treatment_id")

	var req models.CompleteTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = models.CompleteTreatmentRequest{}
	}

	// Get current treatment
	var currentStatus string
	var riskOwnerID, treatOwnerID *string
	err := database.DB.QueryRow(`
		SELECT rt.status, rt.owner_id, r.owner_id
		FROM risk_treatments rt
		JOIN risks r ON rt.risk_id = r.id
		WHERE rt.id = $1 AND rt.risk_id = $2 AND rt.org_id = $3
	`, treatmentID, riskID, orgID).Scan(&currentStatus, &treatOwnerID, &riskOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Treatment not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get treatment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to complete treatment"))
		return
	}

	// Check status allows completion
	if currentStatus == models.TreatmentStatusVerified || currentStatus == models.TreatmentStatusCancelled || currentStatus == models.TreatmentStatusIneffective {
		c.JSON(http.StatusBadRequest, errorResponse("BAD_REQUEST",
			fmt.Sprintf("Treatment is already %s and cannot be completed", currentStatus)))
		return
	}

	// Authorization
	isRiskOwner := riskOwnerID != nil && *riskOwnerID == userID
	isTreatOwner := treatOwnerID != nil && *treatOwnerID == userID
	if !isRiskOwner && !isTreatOwner && !models.HasRole(userRole, models.RiskCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to complete this treatment"))
		return
	}

	// Validate effectiveness rating
	if req.EffectivenessRating != nil && !models.IsValidEffectivenessRating(*req.EffectivenessRating) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid effectiveness_rating"))
		return
	}

	now := time.Now()

	// Determine target status
	targetStatus := models.TreatmentStatusImplemented
	if req.EffectivenessRating != nil {
		targetStatus = models.TreatmentStatusVerified
	}

	_, err = database.DB.Exec(`
		UPDATE risk_treatments SET status = $1, completed_at = $2,
		                           actual_effort_hours = COALESCE($3, actual_effort_hours),
		                           effectiveness_rating = $4, effectiveness_notes = $5,
		                           effectiveness_reviewed_at = $6, effectiveness_reviewed_by = $7,
		                           updated_at = $2
		WHERE id = $8 AND risk_id = $9 AND org_id = $10
	`, targetStatus, now, req.ActualEffortHours, req.EffectivenessRating, req.EffectivenessNotes,
		now, userID, treatmentID, riskID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to complete treatment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to complete treatment"))
		return
	}

	middleware.LogAudit(c, "risk_treatment.completed", "risk_treatment", &treatmentID, map[string]interface{}{
		"status":               targetStatus,
		"effectiveness_rating": req.EffectivenessRating,
	})

	// Check if all treatments are now complete — auto-transition risk to "monitoring"
	var activeCount int
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM risk_treatments
		WHERE risk_id = $1 AND org_id = $2 AND status IN ('planned', 'in_progress', 'implemented')
	`, riskID, orgID).Scan(&activeCount)

	if activeCount == 0 {
		var riskStatus string
		database.DB.QueryRow("SELECT status FROM risks WHERE id = $1", riskID).Scan(&riskStatus)
		if riskStatus == models.RiskStatusTreating {
			database.DB.Exec("UPDATE risks SET status = 'monitoring', updated_at = $1 WHERE id = $2", now, riskID)
			middleware.LogAudit(c, "risk.status_changed", "risk", &riskID, map[string]interface{}{
				"from": "treating", "to": "monitoring", "trigger": "all_treatments_complete",
			})
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                       treatmentID,
		"status":                   targetStatus,
		"completed_at":             now,
		"effectiveness_rating":     req.EffectivenessRating,
		"effectiveness_notes":      req.EffectivenessNotes,
		"effectiveness_reviewed_at": now,
	}))
}
