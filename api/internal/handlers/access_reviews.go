package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
)

// ListCampaigns returns a list of access review campaigns.
func ListCampaigns(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	status := c.Query("status")
	cadence := c.Query("cadence")

	where := []string{"org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, status)
		argN++
	}
	if cadence != "" {
		where = append(where, fmt.Sprintf("cadence = $%d", argN))
		args = append(args, cadence)
		argN++
	}

	whereClause := "WHERE " + join(where, " AND ")

	var total int
	err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM access_review_campaigns %s", whereClause), args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count campaigns"))
		return
	}

	queryArgs := append(args, perPage, offset)
	rows, err := database.Query(fmt.Sprintf(`
		SELECT id, org_id, name, description, status, cadence, scope, reviewer_strategy,
		       default_reviewer_id, started_at, deadline, completed_at, cancelled_at,
		       escalation_config, total_reviews, completed_reviews, approved_count,
		       revoked_count, flagged_count, created_by, tags, notes, created_at, updated_at
		FROM access_review_campaigns
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1), queryArgs...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query campaigns"))
		return
	}
	defer rows.Close()

	campaigns := []models.AccessReviewCampaign{}
	for rows.Next() {
		var cmp models.AccessReviewCampaign
		var scopeJSON, escJSON []byte
		err := rows.Scan(
			&cmp.ID, &cmp.OrgID, &cmp.Name, &cmp.Description, &cmp.Status, &cmp.Cadence, &scopeJSON,
			&cmp.ReviewerStrategy, &cmp.DefaultReviewerID, &cmp.StartedAt, &cmp.Deadline,
			&cmp.CompletedAt, &cmp.CancelledAt, &escJSON, &cmp.TotalReviews, &cmp.CompletedReviews,
			&cmp.ApprovedCount, &cmp.RevokedCount, &cmp.FlaggedCount, &cmp.CreatedBy,
			pq.Array(&cmp.Tags), &cmp.Notes, &cmp.CreatedAt, &cmp.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan campaign"))
			return
		}
		json.Unmarshal(scopeJSON, &cmp.Scope)
		json.Unmarshal(escJSON, &cmp.EscalationConfig)
		campaigns = append(campaigns, cmp)
	}

	c.JSON(http.StatusOK, listResponse(c, campaigns, total, page, perPage))
}

// CreateCampaign creates a new access review campaign.
func CreateCampaign(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req struct {
		Name              string                 `json:"name" binding:"required"`
		Description       *string                `json:"description"`
		Cadence           models.CampaignCadence `json:"cadence" binding:"required"`
		Scope             map[string]interface{} `json:"scope"`
		ReviewerStrategy  string                 `json:"reviewer_strategy" binding:"required"`
		DefaultReviewerID *uuid.UUID             `json:"default_reviewer_id"`
		Deadline          time.Time              `json:"deadline" binding:"required"`
		EscalationConfig  map[string]interface{} `json:"escalation_config"`
		Tags              []string               `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	scopeJSON, _ := json.Marshal(req.Scope)
	escJSON, _ := json.Marshal(req.EscalationConfig)
	if scopeJSON == nil {
		scopeJSON = []byte("{}")
	}
	if escJSON == nil {
		escJSON = []byte("{}")
	}

	id := uuid.New()
	query := "INSERT INTO access_review_campaigns ( id, org_id, name, description, status, cadence, scope, reviewer_strategy, default_reviewer_id, deadline, escalation_config, created_by, tags ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"
	_, err := database.Exec(query, id, orgID, req.Name, req.Description, models.CampaignStatusDraft, req.Cadence, scopeJSON,
		req.ReviewerStrategy, req.DefaultReviewerID, req.Deadline, escJSON, userID, pq.Array(req.Tags))

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create campaign"))
		return
	}

	idStr := id.String()
	middleware.LogAudit(c, "campaign.created", "access_review_campaign", &idStr, map[string]interface{}{
		"name": req.Name,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{"id": id}))
}

// GetCampaign returns a single campaign.
func GetCampaign(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var cmp models.AccessReviewCampaign
	var scopeJSON, escJSON []byte
	err := database.QueryRow(`
		SELECT id, org_id, name, description, status, cadence, scope, reviewer_strategy,
		       default_reviewer_id, started_at, deadline, completed_at, cancelled_at,
		       escalation_config, total_reviews, completed_reviews, approved_count,
		       revoked_count, flagged_count, created_by, tags, notes, created_at, updated_at
		FROM access_review_campaigns
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&cmp.ID, &cmp.OrgID, &cmp.Name, &cmp.Description, &cmp.Status, &cmp.Cadence, &scopeJSON,
		&cmp.ReviewerStrategy, &cmp.DefaultReviewerID, &cmp.StartedAt, &cmp.Deadline,
		&cmp.CompletedAt, &cmp.CancelledAt, &escJSON, &cmp.TotalReviews, &cmp.CompletedReviews,
		&cmp.ApprovedCount, &cmp.RevokedCount, &cmp.FlaggedCount, &cmp.CreatedBy,
		pq.Array(&cmp.Tags), &cmp.Notes, &cmp.CreatedAt, &cmp.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}

	json.Unmarshal(scopeJSON, &cmp.Scope)
	json.Unmarshal(escJSON, &cmp.EscalationConfig)

	c.JSON(http.StatusOK, successResponse(c, cmp))
}

// LaunchCampaign transitions a campaign to active and generates reviews.
func LaunchCampaign(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	// 1. Get campaign and check status
	var status models.CampaignStatus
	var scopeJSON []byte
	var reviewerStrategy string
	var defaultReviewerID *uuid.UUID

	err := database.QueryRow(`
		SELECT status, scope, reviewer_strategy, default_reviewer_id
		FROM access_review_campaigns
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&status, &scopeJSON, &reviewerStrategy, &defaultReviewerID)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}

	if status != models.CampaignStatusDraft {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Only draft campaigns can be launched"))
		return
	}

	// 2. Start transaction
	tx, err := database.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start transaction"))
		return
	}
	defer tx.Rollback()

	// 3. Select access entries based on scope
	// In a real implementation, we would parse scopeJSON.
	// For Sprint 8, we include all active entries for simplicity.
	rows, err := tx.Query(`
		SELECT e.id, r.owner_id
		FROM access_entries e
		JOIN access_resources r ON e.resource_id = r.id
		WHERE e.org_id = $1 AND e.status = 'active'
	`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to select entries for review"))
		return
	}
	defer rows.Close()

	var entryIDs []uuid.UUID
	var owners []uuid.UUID
	for rows.Next() {
		var eid, oid uuid.UUID
		rows.Scan(&eid, &oid)
		entryIDs = append(entryIDs, eid)
		owners = append(owners, oid)
	}

	// 4. Create reviews
	total := 0
	for i, eid := range entryIDs {
		reviewerID := defaultReviewerID
		if reviewerStrategy == "resource_owner" {
			reviewerID = &owners[i]
		}

		_, err = tx.Exec(`
			INSERT INTO access_reviews (org_id, campaign_id, entry_id, reviewer_id, assigned_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT DO NOTHING
		`, orgID, id, eid, reviewerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create reviews"))
			return
		}
		total++
	}

	// 5. Update campaign status
	_, err = tx.Exec(`
		UPDATE access_review_campaigns
		SET status = $1, started_at = NOW(), total_reviews = $2, updated_at = NOW()
		WHERE id = $3
	`, models.CampaignStatusActive, total, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update campaign status"))
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to commit transaction"))
		return
	}

	middleware.LogAudit(c, "campaign.launched", "access_review_campaign", &id, map[string]interface{}{
		"total_reviews": total,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "launched", "total_reviews": total}))
}

// DecideReview records a decision for a single access review.
func DecideReview(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	reviewID := c.Param("id")

	var req struct {
		Decision      models.ReviewDecision `json:"decision" binding:"required"`
		Justification *string               `json:"justification"`
		Notes         *string               `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	if (req.Decision == models.DecisionRevoked || req.Decision == models.DecisionFlagged) && req.Justification == nil {
		c.JSON(http.StatusBadRequest, errorResponse("MISSING_JUSTIFICATION", "Justification is required for revoke/flag decisions"))
		return
	}

	// 1. Get review and check if it belongs to the org
	var campaignID uuid.UUID
	var currentDecision models.ReviewDecision
	err := database.QueryRow(`
		SELECT campaign_id, decision FROM access_reviews WHERE id = $1 AND org_id = $2
	`, reviewID, orgID).Scan(&campaignID, &currentDecision)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Review not found"))
		return
	}

	if currentDecision != models.DecisionPending {
		c.JSON(http.StatusConflict, errorResponse("ALREADY_DECIDED", "Review has already been decided"))
		return
	}

	// 1b. Check campaign is still active
	var campStatus models.CampaignStatus
	err = database.QueryRow(`
		SELECT status FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, campaignID, orgID).Scan(&campStatus)
	if err != nil || (campStatus != models.CampaignStatusActive && campStatus != models.CampaignStatusInReview) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("INVALID_STATE", "Campaign is not active"))
		return
	}

	// 2. Start transaction
	tx, err := database.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start transaction"))
		return
	}
	defer tx.Rollback()

	// 3. Update review
	_, err = tx.Exec(`
		UPDATE access_reviews
		SET decision = $1, justification = $2, decided_by = $3, decided_at = NOW(), notes = $4, updated_at = NOW()
		WHERE id = $5
	`, req.Decision, req.Justification, userID, req.Notes, reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update review"))
		return
	}

	// 4. Update campaign counts
	// If it was already decided, we'd need to adjust counts, but for Sprint 8 we assume first decision.
	col := ""
	switch req.Decision {
	case models.DecisionApproved:
		col = "approved_count"
	case models.DecisionRevoked:
		col = "revoked_count"
	case models.DecisionFlagged:
		col = "flagged_count"
	}

	if col != "" {
		_, err = tx.Exec(fmt.Sprintf(`
			UPDATE access_review_campaigns
			SET completed_reviews = completed_reviews + 1, %s = %s + 1, updated_at = NOW()
			WHERE id = $1
		`, col, col), campaignID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update campaign counts"))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to commit transaction"))
		return
	}

	middleware.LogAudit(c, "access_review.decided", "access_review", &reviewID, map[string]interface{}{
		"decision": req.Decision,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "decided"}))
}

// ListMyReviews returns the reviewer's personal queue.
func ListMyReviews(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	rows, err := database.Query(`
		SELECT r.id, r.campaign_id, r.entry_id, r.decision, r.assigned_at,
		       c.name as campaign_name, e.user_display_name, e.user_email, e.role_name, res.name as resource_name
		FROM access_reviews r
		JOIN access_review_campaigns c ON r.campaign_id = c.id
		JOIN access_entries e ON r.entry_id = e.id
		JOIN access_resources res ON e.resource_id = res.id
		WHERE r.org_id = $1 AND r.reviewer_id = $2 AND r.decision = 'pending'
		ORDER BY r.assigned_at ASC
	`, orgID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query my reviews"))
		return
	}
	defer rows.Close()

	type ReviewItem struct {
		ID           uuid.UUID `json:"id"`
		CampaignID   uuid.UUID `json:"campaign_id"`
		CampaignName string    `json:"campaign_name"`
		EntryID      uuid.UUID `json:"entry_id"`
		User         string    `json:"user_display_name"`
		Email        string    `json:"user_email"`
		Role         string    `json:"role_name"`
		Resource     string    `json:"resource_name"`
		Decision     string    `json:"decision"`
		AssignedAt   time.Time `json:"assigned_at"`
	}

	items := []ReviewItem{}
	for rows.Next() {
		var item ReviewItem
		rows.Scan(&item.ID, &item.CampaignID, &item.EntryID, &item.Decision, &item.AssignedAt,
			&item.CampaignName, &item.User, &item.Email, &item.Role, &item.Resource)
		items = append(items, item)
	}

	c.JSON(http.StatusOK, successResponse(c, items))
}

// UpdateCampaign updates an access review campaign.
func UpdateCampaign(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	// Get current campaign status
	var status models.CampaignStatus
	err := database.QueryRow(`
		SELECT status FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&status)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}

	if status == models.CampaignStatusCompleted || status == models.CampaignStatusCancelled {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Cannot update completed or cancelled campaigns"))
		return
	}

	var req struct {
		Name              *string                 `json:"name"`
		Description       *string                 `json:"description"`
		Cadence           *models.CampaignCadence `json:"cadence"`
		Scope             map[string]interface{}   `json:"scope"`
		ReviewerStrategy  *string                 `json:"reviewer_strategy"`
		DefaultReviewerID *uuid.UUID              `json:"default_reviewer_id"`
		Deadline          *time.Time              `json:"deadline"`
		EscalationConfig  map[string]interface{}   `json:"escalation_config"`
		Tags              []string                `json:"tags"`
		Notes             *string                 `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Active/in_review campaigns: limited fields only
	isActive := status == models.CampaignStatusActive || status == models.CampaignStatusInReview
	if isActive && (req.Name != nil || req.Cadence != nil || req.Scope != nil || req.ReviewerStrategy != nil) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("RESTRICTED_FIELDS", "Cannot change name, cadence, scope, or reviewer_strategy on active campaigns"))
		return
	}

	query := "UPDATE access_review_campaigns SET updated_at = NOW()"
	args := []interface{}{id, orgID}
	argN := 3

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argN)
		args = append(args, *req.Name)
		argN++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argN)
		args = append(args, *req.Description)
		argN++
	}
	if req.Cadence != nil {
		query += fmt.Sprintf(", cadence = $%d", argN)
		args = append(args, *req.Cadence)
		argN++
	}
	if req.Scope != nil {
		scopeJSON, _ := json.Marshal(req.Scope)
		query += fmt.Sprintf(", scope = $%d", argN)
		args = append(args, scopeJSON)
		argN++
	}
	if req.ReviewerStrategy != nil {
		query += fmt.Sprintf(", reviewer_strategy = $%d", argN)
		args = append(args, *req.ReviewerStrategy)
		argN++
	}
	if req.DefaultReviewerID != nil {
		query += fmt.Sprintf(", default_reviewer_id = $%d", argN)
		args = append(args, *req.DefaultReviewerID)
		argN++
	}
	if req.Deadline != nil {
		query += fmt.Sprintf(", deadline = $%d", argN)
		args = append(args, *req.Deadline)
		argN++
	}
	if req.EscalationConfig != nil {
		escJSON, _ := json.Marshal(req.EscalationConfig)
		query += fmt.Sprintf(", escalation_config = $%d", argN)
		args = append(args, escJSON)
		argN++
	}
	if req.Tags != nil {
		query += fmt.Sprintf(", tags = $%d", argN)
		args = append(args, pq.Array(req.Tags))
		argN++
	}
	if req.Notes != nil {
		query += fmt.Sprintf(", notes = $%d", argN)
		args = append(args, *req.Notes)
		argN++
	}

	query += " WHERE id = $1 AND org_id = $2"

	_, err = database.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update campaign"))
		return
	}

	middleware.LogAudit(c, "campaign.updated", "access_review_campaign", &id, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{"status": "updated"}))
}

// CompleteCampaign completes an active/in_review campaign.
func CompleteCampaign(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var req struct {
		ExpirePending   *bool   `json:"expire_pending"`
		CompletionNotes *string `json:"completion_notes"`
	}
	c.ShouldBindJSON(&req)

	expirePending := true
	if req.ExpirePending != nil {
		expirePending = *req.ExpirePending
	}

	// Get campaign
	var status models.CampaignStatus
	err := database.QueryRow(`
		SELECT status FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&status)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}

	if status != models.CampaignStatusActive && status != models.CampaignStatusInReview {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Only active or in_review campaigns can be completed"))
		return
	}

	// Check pending reviews
	var pendingCount int
	database.QueryRow(`
		SELECT COUNT(*) FROM access_reviews WHERE campaign_id = $1 AND org_id = $2 AND decision = 'pending'
	`, id, orgID).Scan(&pendingCount)

	if !expirePending && pendingCount > 0 {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("PENDING_REVIEWS",
			fmt.Sprintf("Campaign has %d pending reviews. Set expire_pending=true to auto-expire them.", pendingCount)))
		return
	}

	tx, err := database.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start transaction"))
		return
	}
	defer tx.Rollback()

	// Expire pending reviews
	expiredCount := 0
	if expirePending && pendingCount > 0 {
		res, err := tx.Exec(`
			UPDATE access_reviews SET decision = 'expired', updated_at = NOW()
			WHERE campaign_id = $1 AND org_id = $2 AND decision = 'pending'
		`, id, orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to expire pending reviews"))
			return
		}
		rows, _ := res.RowsAffected()
		expiredCount = int(rows)
	}

	// Update campaign
	notesVal := interface{}(nil)
	if req.CompletionNotes != nil {
		notesVal = *req.CompletionNotes
	}

	_, err = tx.Exec(`
		UPDATE access_review_campaigns
		SET status = 'completed', completed_at = NOW(), notes = COALESCE($3, notes), updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, id, orgID, notesVal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to complete campaign"))
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to commit transaction"))
		return
	}

	// Get final counts
	var totalReviews, completedReviews, approvedCount, revokedCount, flaggedCount int
	database.QueryRow(`
		SELECT total_reviews, completed_reviews, approved_count, revoked_count, flagged_count
		FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&totalReviews, &completedReviews, &approvedCount, &revokedCount, &flaggedCount)

	middleware.LogAudit(c, "campaign.completed", "access_review_campaign", &id, map[string]interface{}{
		"expired_reviews": expiredCount,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                id,
		"status":            "completed",
		"completed_at":      time.Now(),
		"total_reviews":     totalReviews,
		"completed_reviews": completedReviews + expiredCount,
		"expired_reviews":   expiredCount,
		"approved_count":    approvedCount,
		"revoked_count":     revokedCount,
		"flagged_count":     flaggedCount,
		"message":           fmt.Sprintf("Campaign completed. %d pending reviews expired.", expiredCount),
	}))
}

// CancelCampaign cancels a campaign with a reason.
func CancelCampaign(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", "Reason is required"))
		return
	}

	var status models.CampaignStatus
	err := database.QueryRow(`
		SELECT status FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&status)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}

	if status == models.CampaignStatusCompleted || status == models.CampaignStatusCancelled {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Campaign is already completed or cancelled"))
		return
	}

	tx, err := database.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start transaction"))
		return
	}
	defer tx.Rollback()

	// Expire all pending reviews
	tx.Exec(`
		UPDATE access_reviews SET decision = 'expired', updated_at = NOW()
		WHERE campaign_id = $1 AND org_id = $2 AND decision = 'pending'
	`, id, orgID)

	// Cancel campaign
	_, err = tx.Exec(`
		UPDATE access_review_campaigns
		SET status = 'cancelled', cancelled_at = NOW(), notes = $3, updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, id, orgID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to cancel campaign"))
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to commit transaction"))
		return
	}

	middleware.LogAudit(c, "campaign.cancelled", "access_review_campaign", &id, map[string]interface{}{
		"reason": req.Reason,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":           id,
		"status":       "cancelled",
		"cancelled_at": time.Now(),
		"message":      "Campaign cancelled. All pending reviews discarded.",
	}))
}

// GetCampaignStats returns detailed statistics for a campaign.
func GetCampaignStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	// Get campaign info
	var name string
	var status models.CampaignStatus
	var startedAt *time.Time
	var deadline time.Time
	var totalReviews, completedReviews, approvedCount, revokedCount, flaggedCount int

	err := database.QueryRow(`
		SELECT name, status, started_at, deadline, total_reviews, completed_reviews,
		       approved_count, revoked_count, flagged_count
		FROM access_review_campaigns
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(&name, &status, &startedAt, &deadline, &totalReviews, &completedReviews,
		&approvedCount, &revokedCount, &flaggedCount)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}

	// Decision counts
	var delegatedCount, expiredCount, pendingCount int
	database.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE decision = 'delegated'),
			COUNT(*) FILTER (WHERE decision = 'expired'),
			COUNT(*) FILTER (WHERE decision = 'pending')
		FROM access_reviews WHERE campaign_id = $1 AND org_id = $2
	`, id, orgID).Scan(&delegatedCount, &expiredCount, &pendingCount)

	// Timeline
	completionPct := float64(0)
	if totalReviews > 0 {
		completionPct = float64(completedReviews) / float64(totalReviews) * 100
	}

	daysElapsed := 0
	if startedAt != nil {
		daysElapsed = int(time.Since(*startedAt).Hours() / 24)
	}
	daysRemaining := int(time.Until(deadline).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// By resource
	type ResourceStats struct {
		ResourceID          string `json:"resource_id"`
		ResourceName        string `json:"resource_name"`
		ResourceCriticality string `json:"resource_criticality"`
		Total               int    `json:"total"`
		Approved            int    `json:"approved"`
		Revoked             int    `json:"revoked"`
		Pending             int    `json:"pending"`
		Anomalies           int    `json:"anomalies"`
	}

	resRows, _ := database.Query(`
		SELECT e.resource_id, r.name, r.criticality,
		       COUNT(*),
		       COUNT(*) FILTER (WHERE ar.decision = 'approved'),
		       COUNT(*) FILTER (WHERE ar.decision = 'revoked'),
		       COUNT(*) FILTER (WHERE ar.decision = 'pending'),
		       COUNT(*) FILTER (WHERE jsonb_array_length(COALESCE(e.anomalies, '[]'::jsonb)) > 0)
		FROM access_reviews ar
		JOIN access_entries e ON ar.entry_id = e.id
		JOIN access_resources r ON e.resource_id = r.id
		WHERE ar.campaign_id = $1 AND ar.org_id = $2
		GROUP BY e.resource_id, r.name, r.criticality
	`, id, orgID)

	byResource := []ResourceStats{}
	if resRows != nil {
		defer resRows.Close()
		for resRows.Next() {
			var rs ResourceStats
			resRows.Scan(&rs.ResourceID, &rs.ResourceName, &rs.ResourceCriticality,
				&rs.Total, &rs.Approved, &rs.Revoked, &rs.Pending, &rs.Anomalies)
			byResource = append(byResource, rs)
		}
	}

	// By reviewer
	type ReviewerStats struct {
		ReviewerID string `json:"reviewer_id"`
		ReviewerName string `json:"reviewer_name"`
		Assigned    int    `json:"assigned"`
		Completed   int    `json:"completed"`
		Pending     int    `json:"pending"`
		IsEscalated bool   `json:"is_escalated"`
	}

	revRows, _ := database.Query(`
		SELECT ar.reviewer_id, COALESCE(u.full_name, 'Unknown'),
		       COUNT(*),
		       COUNT(*) FILTER (WHERE ar.decision NOT IN ('pending', 'delegated')),
		       COUNT(*) FILTER (WHERE ar.decision = 'pending'),
		       BOOL_OR(ar.is_escalated)
		FROM access_reviews ar
		LEFT JOIN users u ON ar.reviewer_id = u.id
		WHERE ar.campaign_id = $1 AND ar.org_id = $2 AND ar.reviewer_id IS NOT NULL
		GROUP BY ar.reviewer_id, u.full_name
	`, id, orgID)

	byReviewer := []ReviewerStats{}
	if revRows != nil {
		defer revRows.Close()
		for revRows.Next() {
			var rs ReviewerStats
			revRows.Scan(&rs.ReviewerID, &rs.ReviewerName, &rs.Assigned, &rs.Completed, &rs.Pending, &rs.IsEscalated)
			byReviewer = append(byReviewer, rs)
		}
	}

	// Privileged access stats
	var privTotal, privApproved, privRevoked, privFlagged, privPending int
	database.QueryRow(`
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE ar.decision = 'approved'),
		       COUNT(*) FILTER (WHERE ar.decision = 'revoked'),
		       COUNT(*) FILTER (WHERE ar.decision = 'flagged'),
		       COUNT(*) FILTER (WHERE ar.decision = 'pending')
		FROM access_reviews ar
		JOIN access_entries e ON ar.entry_id = e.id
		WHERE ar.campaign_id = $1 AND ar.org_id = $2 AND e.is_privileged = TRUE
	`, id, orgID).Scan(&privTotal, &privApproved, &privRevoked, &privFlagged, &privPending)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"campaign_id":   id,
		"campaign_name": name,
		"status":        status,
		"completion_pct": completionPct,
		"timeline": gin.H{
			"started_at":     startedAt,
			"deadline":       deadline,
			"days_elapsed":   daysElapsed,
			"days_remaining": daysRemaining,
			"is_overdue":     time.Now().After(deadline),
		},
		"decisions": gin.H{
			"total":     totalReviews,
			"approved":  approvedCount,
			"revoked":   revokedCount,
			"flagged":   flaggedCount,
			"delegated": delegatedCount,
			"expired":   expiredCount,
			"pending":   pendingCount,
		},
		"by_resource": byResource,
		"by_reviewer": byReviewer,
		"privileged_access": gin.H{
			"total_privileged_reviews": privTotal,
			"approved":                 privApproved,
			"revoked":                  privRevoked,
			"flagged":                  privFlagged,
			"pending":                  privPending,
		},
	}))
}

// ListCampaignReviews returns reviews in a campaign with filters.
func ListCampaignReviews(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	campaignID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	where := []string{"ar.org_id = $1", "ar.campaign_id = $2"}
	args := []interface{}{orgID, campaignID}
	argN := 3

	if v := c.Query("decision"); v != "" {
		where = append(where, fmt.Sprintf("ar.decision = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("reviewer_id"); v != "" {
		where = append(where, fmt.Sprintf("ar.reviewer_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("is_escalated"); v != "" {
		where = append(where, fmt.Sprintf("ar.is_escalated = $%d", argN))
		args = append(args, v == "true")
		argN++
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(e.user_email ILIKE $%d OR r.name ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	whereClause := "WHERE " + join(where, " AND ")

	var total int
	err := database.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM access_reviews ar
		JOIN access_entries e ON ar.entry_id = e.id
		JOIN access_resources r ON e.resource_id = r.id
		%s
	`, whereClause), args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count reviews"))
		return
	}

	queryArgs := append(args, perPage, offset)
	rows, err := database.Query(fmt.Sprintf(`
		SELECT ar.id, ar.campaign_id, ar.entry_id, ar.reviewer_id, ar.assigned_at,
		       ar.decision, ar.justification, ar.decided_by, ar.decided_at,
		       ar.delegated_to_id, ar.is_escalated, ar.access_snapshot,
		       ar.revocation_executed, ar.notes, ar.created_at, ar.updated_at,
		       COALESCE(u.full_name, '') as reviewer_name
		FROM access_reviews ar
		JOIN access_entries e ON ar.entry_id = e.id
		JOIN access_resources r ON e.resource_id = r.id
		LEFT JOIN users u ON ar.reviewer_id = u.id
		%s
		ORDER BY ar.created_at ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1), queryArgs...)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query reviews"))
		return
	}
	defer rows.Close()

	type ReviewListItem struct {
		ID                uuid.UUID       `json:"id"`
		CampaignID        uuid.UUID       `json:"campaign_id"`
		EntryID           uuid.UUID       `json:"entry_id"`
		ReviewerID        *uuid.UUID      `json:"reviewer_id"`
		ReviewerName      string          `json:"reviewer_name"`
		AssignedAt        *time.Time      `json:"assigned_at"`
		Decision          string          `json:"decision"`
		Justification     *string         `json:"justification"`
		DecidedBy         *uuid.UUID      `json:"decided_by"`
		DecidedAt         *time.Time      `json:"decided_at"`
		DelegatedToID     *uuid.UUID      `json:"delegated_to_id"`
		IsEscalated       bool            `json:"is_escalated"`
		AccessSnapshot    json.RawMessage `json:"access_snapshot"`
		RevocationExecuted bool           `json:"revocation_executed"`
		Notes             *string         `json:"notes"`
		CreatedAt         time.Time       `json:"created_at"`
		UpdatedAt         time.Time       `json:"updated_at"`
	}

	items := []ReviewListItem{}
	for rows.Next() {
		var item ReviewListItem
		var snapshot []byte
		err := rows.Scan(
			&item.ID, &item.CampaignID, &item.EntryID, &item.ReviewerID, &item.AssignedAt,
			&item.Decision, &item.Justification, &item.DecidedBy, &item.DecidedAt,
			&item.DelegatedToID, &item.IsEscalated, &snapshot,
			&item.RevocationExecuted, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
			&item.ReviewerName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan review"))
			return
		}
		item.AccessSnapshot = snapshot
		items = append(items, item)
	}

	c.JSON(http.StatusOK, listResponse(c, items, total, page, perPage))
}

// GetReviewDetail returns full detail for a single review.
func GetReviewDetail(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	campaignID := c.Param("id")
	reviewID := c.Param("rid")

	var r models.AccessReview
	var snapshot []byte
	var campaignName, reviewerName string
	var decidedByName *string

	err := database.QueryRow(`
		SELECT ar.id, ar.org_id, ar.campaign_id, ar.entry_id, ar.reviewer_id, ar.assigned_at,
		       ar.decision, ar.justification, ar.decided_by, ar.decided_at,
		       ar.delegated_to_id, ar.delegated_at, ar.delegation_reason,
		       ar.is_escalated, ar.escalated_at, ar.escalated_to_id,
		       ar.access_snapshot, ar.revocation_executed, ar.revocation_executed_at, ar.revocation_notes,
		       ar.notes, ar.created_at, ar.updated_at,
		       c.name,
		       COALESCE(u.full_name, ''),
		       u2.full_name
		FROM access_reviews ar
		JOIN access_review_campaigns c ON ar.campaign_id = c.id
		LEFT JOIN users u ON ar.reviewer_id = u.id
		LEFT JOIN users u2 ON ar.decided_by = u2.id
		WHERE ar.id = $1 AND ar.campaign_id = $2 AND ar.org_id = $3
	`, reviewID, campaignID, orgID).Scan(
		&r.ID, &r.OrgID, &r.CampaignID, &r.EntryID, &r.ReviewerID, &r.AssignedAt,
		&r.Decision, &r.Justification, &r.DecidedBy, &r.DecidedAt,
		&r.DelegatedToID, &r.DelegatedAt, &r.DelegationReason,
		&r.IsEscalated, &r.EscalatedAt, &r.EscalatedToID,
		&snapshot, &r.RevocationExecuted, &r.RevocationExecutedAt, &r.RevocationNotes,
		&r.Notes, &r.CreatedAt, &r.UpdatedAt,
		&campaignName,
		&reviewerName,
		&decidedByName,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Review not found"))
		return
	}

	var accessSnap interface{}
	json.Unmarshal(snapshot, &accessSnap)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                    r.ID,
		"campaign_id":           r.CampaignID,
		"campaign_name":         campaignName,
		"entry_id":              r.EntryID,
		"reviewer_id":           r.ReviewerID,
		"reviewer_name":         reviewerName,
		"assigned_at":           r.AssignedAt,
		"decision":              r.Decision,
		"justification":         r.Justification,
		"decided_by":            r.DecidedBy,
		"decided_by_name":       decidedByName,
		"decided_at":            r.DecidedAt,
		"delegated_to_id":       r.DelegatedToID,
		"delegated_at":          r.DelegatedAt,
		"delegation_reason":     r.DelegationReason,
		"is_escalated":          r.IsEscalated,
		"escalated_at":          r.EscalatedAt,
		"escalated_to_id":       r.EscalatedToID,
		"access_snapshot":       accessSnap,
		"revocation_executed":   r.RevocationExecuted,
		"revocation_executed_at": r.RevocationExecutedAt,
		"revocation_notes":      r.RevocationNotes,
		"notes":                 r.Notes,
		"created_at":            r.CreatedAt,
		"updated_at":            r.UpdatedAt,
	}))
}

// DecideReviewNested handles POST campaigns/:id/reviews/:rid/decide.
func DecideReviewNested(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	campaignID := c.Param("id")
	reviewID := c.Param("rid")

	var req struct {
		Decision      models.ReviewDecision `json:"decision" binding:"required"`
		Justification *string               `json:"justification"`
		Notes         *string               `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	if (req.Decision == models.DecisionRevoked || req.Decision == models.DecisionFlagged) && req.Justification == nil {
		c.JSON(http.StatusBadRequest, errorResponse("MISSING_JUSTIFICATION", "Justification is required for revoke/flag decisions"))
		return
	}

	// Check campaign status
	var campStatus models.CampaignStatus
	err := database.QueryRow(`
		SELECT status FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, campaignID, orgID).Scan(&campStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}
	if campStatus != models.CampaignStatusActive && campStatus != models.CampaignStatusInReview {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("INVALID_STATE", "Campaign is not active"))
		return
	}

	// Check review
	var currentDecision models.ReviewDecision
	err = database.QueryRow(`
		SELECT decision FROM access_reviews WHERE id = $1 AND campaign_id = $2 AND org_id = $3
	`, reviewID, campaignID, orgID).Scan(&currentDecision)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Review not found"))
		return
	}
	if currentDecision != models.DecisionPending {
		c.JSON(http.StatusConflict, errorResponse("ALREADY_DECIDED", "Review has already been decided"))
		return
	}

	tx, err := database.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start transaction"))
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE access_reviews
		SET decision = $1, justification = $2, decided_by = $3, decided_at = NOW(), notes = $4, updated_at = NOW()
		WHERE id = $5
	`, req.Decision, req.Justification, userID, req.Notes, reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update review"))
		return
	}

	col := ""
	switch req.Decision {
	case models.DecisionApproved:
		col = "approved_count"
	case models.DecisionRevoked:
		col = "revoked_count"
	case models.DecisionFlagged:
		col = "flagged_count"
	}
	if col != "" {
		_, err = tx.Exec(fmt.Sprintf(`
			UPDATE access_review_campaigns
			SET completed_reviews = completed_reviews + 1, %s = %s + 1, updated_at = NOW()
			WHERE id = $1
		`, col, col), campaignID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to update campaign counts"))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to commit transaction"))
		return
	}

	// Get campaign progress
	var totalR, completedR int
	database.QueryRow(`
		SELECT total_reviews, completed_reviews FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, campaignID, orgID).Scan(&totalR, &completedR)

	pct := float64(0)
	if totalR > 0 {
		pct = float64(completedR) / float64(totalR) * 100
	}

	middleware.LogAudit(c, "access_review.decided", "access_review", &reviewID, map[string]interface{}{
		"decision":    req.Decision,
		"campaign_id": campaignID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":            reviewID,
		"decision":      req.Decision,
		"justification": req.Justification,
		"decided_by":    userID,
		"decided_at":    time.Now(),
		"campaign_progress": gin.H{
			"total_reviews":     totalR,
			"completed_reviews": completedR,
			"completion_pct":    pct,
		},
	}))
}

// BulkDecideReviews handles bulk decisions for reviews in a campaign.
func BulkDecideReviews(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	campaignID := c.Param("id")

	var req struct {
		Reviews []struct {
			ReviewID      string                `json:"review_id" binding:"required"`
			Decision      models.ReviewDecision `json:"decision" binding:"required"`
			Justification *string               `json:"justification"`
		} `json:"reviews" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	if len(req.Reviews) == 0 || len(req.Reviews) > 100 {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", "Reviews array must contain 1-100 items"))
		return
	}

	// Check campaign status
	var campStatus models.CampaignStatus
	err := database.QueryRow(`
		SELECT status FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, campaignID, orgID).Scan(&campStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}
	if campStatus != models.CampaignStatusActive && campStatus != models.CampaignStatusInReview {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("INVALID_STATE", "Campaign is not active"))
		return
	}

	tx, err := database.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to start transaction"))
		return
	}
	defer tx.Rollback()

	type ResultItem struct {
		ReviewID string `json:"review_id"`
		Status   string `json:"status"`
		Decision string `json:"decision"`
		Error    string `json:"error,omitempty"`
	}

	results := []ResultItem{}
	succeeded := 0
	failed := 0

	for _, rev := range req.Reviews {
		if (rev.Decision == models.DecisionRevoked || rev.Decision == models.DecisionFlagged) && rev.Justification == nil {
			results = append(results, ResultItem{ReviewID: rev.ReviewID, Status: "failed", Decision: string(rev.Decision), Error: "Justification required"})
			failed++
			continue
		}

		// Check review exists and is pending
		var currentDecision models.ReviewDecision
		err := tx.QueryRow(`
			SELECT decision FROM access_reviews WHERE id = $1 AND campaign_id = $2 AND org_id = $3
		`, rev.ReviewID, campaignID, orgID).Scan(&currentDecision)
		if err != nil {
			results = append(results, ResultItem{ReviewID: rev.ReviewID, Status: "failed", Decision: string(rev.Decision), Error: "Review not found"})
			failed++
			continue
		}
		if currentDecision != models.DecisionPending {
			results = append(results, ResultItem{ReviewID: rev.ReviewID, Status: "failed", Decision: string(rev.Decision), Error: "Already decided"})
			failed++
			continue
		}

		_, err = tx.Exec(`
			UPDATE access_reviews
			SET decision = $1, justification = $2, decided_by = $3, decided_at = NOW(), updated_at = NOW()
			WHERE id = $4
		`, rev.Decision, rev.Justification, userID, rev.ReviewID)
		if err != nil {
			results = append(results, ResultItem{ReviewID: rev.ReviewID, Status: "failed", Decision: string(rev.Decision), Error: "Update failed"})
			failed++
			continue
		}

		col := ""
		switch rev.Decision {
		case models.DecisionApproved:
			col = "approved_count"
		case models.DecisionRevoked:
			col = "revoked_count"
		case models.DecisionFlagged:
			col = "flagged_count"
		}
		if col != "" {
			tx.Exec(fmt.Sprintf(`
				UPDATE access_review_campaigns
				SET completed_reviews = completed_reviews + 1, %s = %s + 1, updated_at = NOW()
				WHERE id = $1
			`, col, col), campaignID)
		}

		results = append(results, ResultItem{ReviewID: rev.ReviewID, Status: "success", Decision: string(rev.Decision)})
		succeeded++
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to commit transaction"))
		return
	}

	// Get campaign progress
	var totalR, completedR int
	database.QueryRow(`
		SELECT total_reviews, completed_reviews FROM access_review_campaigns WHERE id = $1 AND org_id = $2
	`, campaignID, orgID).Scan(&totalR, &completedR)

	pct := float64(0)
	if totalR > 0 {
		pct = float64(completedR) / float64(totalR) * 100
	}

	middleware.LogAudit(c, "access_review.bulk_decided", "access_review_campaign", &campaignID, map[string]interface{}{
		"succeeded": succeeded,
		"failed":    failed,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"processed": len(req.Reviews),
		"succeeded": succeeded,
		"failed":    failed,
		"results":   results,
		"campaign_progress": gin.H{
			"total_reviews":     totalR,
			"completed_reviews": completedR,
			"completion_pct":    pct,
		},
	}))
}

// DelegateReview delegates a review to another user.
func DelegateReview(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	campaignID := c.Param("id")
	reviewID := c.Param("rid")

	var req struct {
		DelegateToID string `json:"delegate_to_id" binding:"required"`
		Reason       string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Check review exists and is pending
	var currentDecision models.ReviewDecision
	err := database.QueryRow(`
		SELECT decision FROM access_reviews WHERE id = $1 AND campaign_id = $2 AND org_id = $3
	`, reviewID, campaignID, orgID).Scan(&currentDecision)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Review not found"))
		return
	}
	if currentDecision != models.DecisionPending {
		c.JSON(http.StatusConflict, errorResponse("ALREADY_DECIDED", "Cannot delegate a decided review"))
		return
	}

	// Verify delegate user exists
	var delegateName string
	err = database.QueryRow(`
		SELECT full_name FROM users WHERE id = $1 AND org_id = $2
	`, req.DelegateToID, orgID).Scan(&delegateName)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Delegate user not found"))
		return
	}

	_, err = database.Exec(`
		UPDATE access_reviews
		SET decision = 'delegated', delegated_to_id = $1, delegated_at = NOW(),
		    delegation_reason = $2, reviewer_id = $1, updated_at = NOW()
		WHERE id = $3
	`, req.DelegateToID, req.Reason, reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to delegate review"))
		return
	}

	// Create a new pending review for the delegate
	_, err = database.Exec(`
		INSERT INTO access_reviews (org_id, campaign_id, entry_id, reviewer_id, assigned_at, access_snapshot)
		SELECT org_id, campaign_id, entry_id, $1, NOW(), access_snapshot
		FROM access_reviews WHERE id = $2 AND org_id = $3
	`, req.DelegateToID, reviewID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to create delegated review"))
		return
	}

	// Increment total_reviews to account for the new review
	database.Exec(`
		UPDATE access_review_campaigns SET total_reviews = total_reviews + 1, updated_at = NOW()
		WHERE id = $1 AND org_id = $2
	`, campaignID, orgID)

	middleware.LogAudit(c, "access_review.delegated", "access_review", &reviewID, map[string]interface{}{
		"delegated_to": req.DelegateToID,
		"reason":       req.Reason,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                reviewID,
		"decision":          "delegated",
		"delegated_to_id":   req.DelegateToID,
		"delegated_to_name": delegateName,
		"delegated_at":      time.Now(),
		"delegation_reason": req.Reason,
		"message":           fmt.Sprintf("Review delegated to %s.", delegateName),
	}))
}

// EscalateReview escalates a review to another user.
func EscalateReview(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	campaignID := c.Param("id")
	reviewID := c.Param("rid")

	var req struct {
		EscalateToID string `json:"escalate_to_id" binding:"required"`
		Reason       string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Check review exists and is pending
	var currentDecision models.ReviewDecision
	err := database.QueryRow(`
		SELECT decision FROM access_reviews WHERE id = $1 AND campaign_id = $2 AND org_id = $3
	`, reviewID, campaignID, orgID).Scan(&currentDecision)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Review not found"))
		return
	}
	if currentDecision != models.DecisionPending {
		c.JSON(http.StatusConflict, errorResponse("ALREADY_DECIDED", "Cannot escalate a decided review"))
		return
	}

	// Verify escalate-to user exists
	var escalateName string
	err = database.QueryRow(`
		SELECT full_name FROM users WHERE id = $1 AND org_id = $2
	`, req.EscalateToID, orgID).Scan(&escalateName)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Escalation target user not found"))
		return
	}

	_, err = database.Exec(`
		UPDATE access_reviews
		SET is_escalated = TRUE, escalated_at = NOW(), escalated_to_id = $1,
		    reviewer_id = $1, updated_at = NOW()
		WHERE id = $2
	`, req.EscalateToID, reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to escalate review"))
		return
	}

	middleware.LogAudit(c, "access_review.escalated", "access_review", &reviewID, map[string]interface{}{
		"escalated_to": req.EscalateToID,
		"reason":       req.Reason,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                reviewID,
		"is_escalated":      true,
		"escalated_at":      time.Now(),
		"escalated_to_id":   req.EscalateToID,
		"escalated_to_name": escalateName,
		"reviewer_id":       req.EscalateToID,
		"reviewer_name":     escalateName,
		"message":           fmt.Sprintf("Review escalated to %s.", escalateName),
	}))
}

// MarkRevocation marks a revoked review's revocation as executed.
func MarkRevocation(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	campaignID := c.Param("id")
	reviewID := c.Param("rid")

	var req struct {
		Executed bool    `json:"executed" binding:"required"`
		Notes    *string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_INPUT", err.Error()))
		return
	}

	// Check review exists, is revoked, and not already executed
	var decision models.ReviewDecision
	var alreadyExecuted bool
	err := database.QueryRow(`
		SELECT decision, revocation_executed
		FROM access_reviews WHERE id = $1 AND campaign_id = $2 AND org_id = $3
	`, reviewID, campaignID, orgID).Scan(&decision, &alreadyExecuted)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Review not found"))
		return
	}
	if decision != models.DecisionRevoked {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Review decision is not revoked"))
		return
	}
	if alreadyExecuted {
		c.JSON(http.StatusConflict, errorResponse("ALREADY_EXECUTED", "Revocation already executed"))
		return
	}

	_, err = database.Exec(`
		UPDATE access_reviews
		SET revocation_executed = TRUE, revocation_executed_at = NOW(), revocation_notes = $1, updated_at = NOW()
		WHERE id = $2
	`, req.Notes, reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to mark revocation"))
		return
	}

	middleware.LogAudit(c, "access_review.revocation_executed", "access_review", &reviewID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                     reviewID,
		"decision":               "revoked",
		"revocation_executed":    true,
		"revocation_executed_at": time.Now(),
		"revocation_notes":       req.Notes,
	}))
}

// GetAccessReviewDashboard returns comprehensive dashboard statistics.
func GetAccessReviewDashboard(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	// Summary counts
	var totalIdPs, connectedIdPs, totalResources, totalEntries, entriesWithAnomalies, privilegedEntries int
	database.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM identity_providers WHERE org_id = $1),
			(SELECT COUNT(*) FROM identity_providers WHERE org_id = $1 AND status = 'connected'),
			(SELECT COUNT(*) FROM access_resources WHERE org_id = $1),
			(SELECT COUNT(*) FROM access_entries WHERE org_id = $1),
			(SELECT COUNT(*) FROM access_entries WHERE org_id = $1 AND jsonb_array_length(COALESCE(anomalies, '[]'::jsonb)) > 0),
			(SELECT COUNT(*) FROM access_entries WHERE org_id = $1 AND is_privileged = TRUE)
	`, orgID).Scan(&totalIdPs, &connectedIdPs, &totalResources, &totalEntries, &entriesWithAnomalies, &privilegedEntries)

	var activeCampaigns, pendingReviews, pendingRevocations int
	database.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM access_review_campaigns WHERE org_id = $1 AND status IN ('active', 'in_review')),
			(SELECT COUNT(*) FROM access_reviews WHERE org_id = $1 AND decision = 'pending'),
			(SELECT COUNT(*) FROM access_reviews WHERE org_id = $1 AND decision = 'revoked' AND revocation_executed = FALSE)
	`, orgID).Scan(&activeCampaigns, &pendingReviews, &pendingRevocations)

	// Active campaigns list
	type ActiveCampaignItem struct {
		ID              string     `json:"id"`
		Name            string     `json:"name"`
		Status          string     `json:"status"`
		CompletionPct   float64    `json:"completion_pct"`
		Deadline        time.Time  `json:"deadline"`
		DaysRemaining   int        `json:"days_remaining"`
		PendingReviews  int        `json:"pending_reviews"`
		EscalatedReviews int       `json:"escalated_reviews"`
	}

	campRows, _ := database.Query(`
		SELECT c.id, c.name, c.status, c.deadline, c.total_reviews, c.completed_reviews,
		       (SELECT COUNT(*) FROM access_reviews WHERE campaign_id = c.id AND decision = 'pending'),
		       (SELECT COUNT(*) FROM access_reviews WHERE campaign_id = c.id AND is_escalated = TRUE AND decision = 'pending')
		FROM access_review_campaigns c
		WHERE c.org_id = $1 AND c.status IN ('active', 'in_review')
		ORDER BY c.deadline ASC
	`, orgID)

	activeCampaignList := []ActiveCampaignItem{}
	if campRows != nil {
		defer campRows.Close()
		for campRows.Next() {
			var item ActiveCampaignItem
			var totalR, completedR int
			campRows.Scan(&item.ID, &item.Name, &item.Status, &item.Deadline, &totalR, &completedR,
				&item.PendingReviews, &item.EscalatedReviews)
			if totalR > 0 {
				item.CompletionPct = float64(completedR) / float64(totalR) * 100
			}
			item.DaysRemaining = int(time.Until(item.Deadline).Hours() / 24)
			if item.DaysRemaining < 0 {
				item.DaysRemaining = 0
			}
			activeCampaignList = append(activeCampaignList, item)
		}
	}

	// Recent decisions
	type RecentDecision struct {
		ReviewID      string    `json:"review_id"`
		CampaignName  string    `json:"campaign_name"`
		UserEmail     string    `json:"user_email"`
		ResourceName  string    `json:"resource_name"`
		Decision      string    `json:"decision"`
		DecidedByName string    `json:"decided_by_name"`
		DecidedAt     time.Time `json:"decided_at"`
	}

	decRows, _ := database.Query(`
		SELECT ar.id, c.name, e.user_email, r.name, ar.decision, COALESCE(u.full_name, ''), ar.decided_at
		FROM access_reviews ar
		JOIN access_review_campaigns c ON ar.campaign_id = c.id
		JOIN access_entries e ON ar.entry_id = e.id
		JOIN access_resources r ON e.resource_id = r.id
		LEFT JOIN users u ON ar.decided_by = u.id
		WHERE ar.org_id = $1 AND ar.decision NOT IN ('pending', 'expired')
		ORDER BY ar.decided_at DESC
		LIMIT 10
	`, orgID)

	recentDecisions := []RecentDecision{}
	if decRows != nil {
		defer decRows.Close()
		for decRows.Next() {
			var d RecentDecision
			decRows.Scan(&d.ReviewID, &d.CampaignName, &d.UserEmail, &d.ResourceName,
				&d.Decision, &d.DecidedByName, &d.DecidedAt)
			recentDecisions = append(recentDecisions, d)
		}
	}

	// Revocation tracking
	var totalRevocations, executedRevocations, pendingExec int
	database.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE revocation_executed = TRUE),
			COUNT(*) FILTER (WHERE revocation_executed = FALSE)
		FROM access_reviews
		WHERE org_id = $1 AND decision = 'revoked'
	`, orgID).Scan(&totalRevocations, &executedRevocations, &pendingExec)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"summary": gin.H{
			"total_identity_providers": totalIdPs,
			"connected_providers":      connectedIdPs,
			"total_resources":          totalResources,
			"total_entries":            totalEntries,
			"entries_with_anomalies":   entriesWithAnomalies,
			"privileged_entries":       privilegedEntries,
			"active_campaigns":         activeCampaigns,
			"pending_reviews":          pendingReviews,
			"pending_revocations":      pendingRevocations,
		},
		"active_campaigns":  activeCampaignList,
		"recent_decisions":  recentDecisions,
		"revocation_tracking": gin.H{
			"total_revocations":    totalRevocations,
			"executed":             executedRevocations,
			"pending_execution":    pendingExec,
		},
	}))
}

// GetCertificationReport generates a certification report for a completed campaign.
func GetCertificationReport(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	campaignID := c.Param("id")

	// Get campaign details
	var name string
	var status models.CampaignStatus
	var cadence models.CampaignCadence
	var scopeJSON []byte
	var startedAt, completedAt *time.Time
	var deadline time.Time
	var totalReviews, approvedCount, revokedCount, flaggedCount int

	err := database.QueryRow(`
		SELECT name, status, cadence, scope, started_at, completed_at, deadline,
		       total_reviews, approved_count, revoked_count, flagged_count
		FROM access_review_campaigns
		WHERE id = $1 AND org_id = $2
	`, campaignID, orgID).Scan(&name, &status, &cadence, &scopeJSON, &startedAt, &completedAt,
		&deadline, &totalReviews, &approvedCount, &revokedCount, &flaggedCount)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Campaign not found"))
		return
	}

	if status != models.CampaignStatusCompleted {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Certification reports are only available for completed campaigns"))
		return
	}

	var scope interface{}
	json.Unmarshal(scopeJSON, &scope)

	// Get expired count
	var expiredCount int
	database.QueryRow(`
		SELECT COUNT(*) FROM access_reviews WHERE campaign_id = $1 AND org_id = $2 AND decision = 'expired'
	`, campaignID, orgID).Scan(&expiredCount)

	// Get all decisions
	type DecisionItem struct {
		ResourceName        string      `json:"resource_name"`
		ResourceCriticality string      `json:"resource_criticality"`
		UserEmail           string      `json:"user_email"`
		UserDepartment      *string     `json:"user_department"`
		UserTitle           *string     `json:"user_title"`
		RoleName            string      `json:"role_name"`
		IsPrivileged        bool        `json:"is_privileged"`
		Decision            string      `json:"decision"`
		Justification       *string     `json:"justification"`
		ReviewerName        string      `json:"reviewer_name"`
		ReviewerEmail       string      `json:"reviewer_email"`
		DecidedAt           *time.Time  `json:"decided_at"`
		RevocationExecuted  *bool       `json:"revocation_executed"`
		RevocationAt        *time.Time  `json:"revocation_executed_at"`
	}

	rows, err := database.Query(`
		SELECT r.name, r.criticality, e.user_email, e.user_department, e.user_title,
		       e.role_name, e.is_privileged, ar.decision, ar.justification,
		       COALESCE(u.full_name, ''), COALESCE(u.email, ''), ar.decided_at,
		       ar.revocation_executed, ar.revocation_executed_at
		FROM access_reviews ar
		JOIN access_entries e ON ar.entry_id = e.id
		JOIN access_resources r ON e.resource_id = r.id
		LEFT JOIN users u ON ar.decided_by = u.id
		WHERE ar.campaign_id = $1 AND ar.org_id = $2
		ORDER BY r.name, e.user_email
	`, campaignID, orgID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query decisions"))
		return
	}
	defer rows.Close()

	decisions := []DecisionItem{}
	for rows.Next() {
		var d DecisionItem
		rows.Scan(&d.ResourceName, &d.ResourceCriticality, &d.UserEmail, &d.UserDepartment, &d.UserTitle,
			&d.RoleName, &d.IsPrivileged, &d.Decision, &d.Justification,
			&d.ReviewerName, &d.ReviewerEmail, &d.DecidedAt,
			&d.RevocationExecuted, &d.RevocationAt)
		decisions = append(decisions, d)
	}

	// Revocation summary
	var totalRev, executedRev, pendingRev int
	database.QueryRow(`
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE revocation_executed = TRUE),
		       COUNT(*) FILTER (WHERE revocation_executed = FALSE)
		FROM access_reviews
		WHERE campaign_id = $1 AND org_id = $2 AND decision = 'revoked'
	`, campaignID, orgID).Scan(&totalRev, &executedRev, &pendingRev)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"report": gin.H{
			"campaign_id":   campaignID,
			"campaign_name": name,
			"cadence":       cadence,
			"period": gin.H{
				"started_at":   startedAt,
				"completed_at": completedAt,
				"deadline":     deadline,
			},
			"scope": scope,
			"summary": gin.H{
				"total_reviews":   totalReviews,
				"approved":        approvedCount,
				"revoked":         revokedCount,
				"flagged":         flaggedCount,
				"expired":         expiredCount,
				"completion_rate": func() float64 {
					if totalReviews == 0 {
						return 0
					}
					return float64(totalReviews-expiredCount) / float64(totalReviews) * 100
				}(),
			},
			"decisions": decisions,
			"revocation_summary": gin.H{
				"total_revocations": totalRev,
				"executed":          executedRev,
				"pending":           pendingRev,
			},
		},
		"generated_at": time.Now(),
	}))
}
