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
		&cmp.ApprovedCount, &cmp.ReviewerStrategy, &cmp.FlaggedCount, &cmp.CreatedBy,
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
