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
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListPolicySignoffs lists signoffs for a policy.
func ListPolicySignoffs(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	policyID := c.Param("id")

	// Verify policy exists
	var exists bool
	database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policies WHERE id = $1 AND org_id = $2)`, policyID, orgID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}

	where := []string{"ps.policy_id = $1", "ps.org_id = $2"}
	args := []interface{}{policyID, orgID}
	argN := 3

	if v := c.Query("version_number"); v != "" {
		vNum, err := strconv.Atoi(v)
		if err == nil {
			where = append(where, fmt.Sprintf(`ps.policy_version_id = (
				SELECT id FROM policy_versions WHERE policy_id = $1 AND version_number = $%d
			)`, argN))
			args = append(args, vNum)
			argN++
		}
	}
	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("ps.status = $%d", argN))
		args = append(args, v)
		argN++
	}

	query := fmt.Sprintf(`
		SELECT ps.id, ps.policy_id,
			pv.id, pv.version_number,
			ps.signer_id, u_signer.first_name, u_signer.last_name, u_signer.email, ps.signer_role,
			ps.requested_by, u_req.first_name, u_req.last_name,
			ps.requested_at, ps.due_date, ps.status, ps.decided_at, ps.comments,
			ps.reminder_count, ps.reminder_sent_at
		FROM policy_signoffs ps
		LEFT JOIN policy_versions pv ON pv.id = ps.policy_version_id
		LEFT JOIN users u_signer ON u_signer.id = ps.signer_id
		LEFT JOIN users u_req ON u_req.id = ps.requested_by
		WHERE %s
		ORDER BY ps.created_at
	`, joinWhere(where))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list signoffs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list signoffs"))
		return
	}
	defer rows.Close()

	signoffs := []gin.H{}
	for rows.Next() {
		var (
			id, policyIDr                       string
			pvID                                *string
			pvVersionNum                        *int
			signerID, sFirst, sLast, sEmail     string
			signerRole                          *string
			requestedByID, rFirst, rLast        string
			requestedAt                         time.Time
			dueDate                             *time.Time
			status                              string
			decidedAt                           *time.Time
			comments                            *string
			reminderCount                       int
			reminderSentAt                      *time.Time
		)
		if err := rows.Scan(
			&id, &policyIDr,
			&pvID, &pvVersionNum,
			&signerID, &sFirst, &sLast, &sEmail, &signerRole,
			&requestedByID, &rFirst, &rLast,
			&requestedAt, &dueDate, &status, &decidedAt, &comments,
			&reminderCount, &reminderSentAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan signoff row")
			continue
		}

		s := gin.H{
			"id":        id,
			"policy_id": policyIDr,
			"signer": gin.H{
				"id":    signerID,
				"name":  sFirst + " " + sLast,
				"email": sEmail,
				"role":  signerRole,
			},
			"signer_role":  signerRole,
			"requested_by": gin.H{"id": requestedByID, "name": rFirst + " " + rLast},
			"requested_at":    requestedAt,
			"due_date":        dueDate,
			"status":          status,
			"decided_at":      decidedAt,
			"comments":        comments,
			"reminder_count":  reminderCount,
			"reminder_sent_at": reminderSentAt,
		}

		if pvID != nil {
			s["policy_version"] = gin.H{"id": *pvID, "version_number": pvVersionNum}
		}

		signoffs = append(signoffs, s)
	}

	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": signoffs,
		"meta": gin.H{
			"total":      len(signoffs),
			"page":       1,
			"per_page":   20,
			"request_id": reqID,
		},
	})
}

// ApproveSignoff approves a sign-off request.
func ApproveSignoff(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	policyID := c.Param("id")
	signoffID := c.Param("signoff_id")

	var req models.SignoffDecisionRequest
	c.ShouldBindJSON(&req)

	// Get signoff
	var signerID, status, pvID string
	err := database.DB.QueryRow(`
		SELECT signer_id, status, policy_version_id
		FROM policy_signoffs
		WHERE id = $1 AND policy_id = $2 AND org_id = $3
	`, signoffID, policyID, orgID).Scan(&signerID, &status, &pvID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Sign-off not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to approve sign-off"))
		return
	}

	if signerID != userID {
		c.JSON(http.StatusForbidden, errorResponse("NOT_SIGNER", "You are not the designated signer"))
		return
	}
	if status != models.SignoffStatusPending {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATUS_TRANSITION", "Sign-off is not in pending status"))
		return
	}

	now := time.Now()
	_, err = database.DB.Exec(`
		UPDATE policy_signoffs SET status = 'approved', decided_at = $1, comments = $2
		WHERE id = $3
	`, now, req.Comments, signoffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to approve sign-off"))
		return
	}

	middleware.LogAudit(c, "policy_signoff.approved", "policy_signoff", &signoffID, map[string]interface{}{
		"policy_id": policyID,
	})

	// Check if all signoffs for this version are now approved
	var pendingCount int
	database.DB.QueryRow(`
		SELECT COUNT(*) FROM policy_signoffs
		WHERE policy_version_id = $1 AND status = 'pending'
	`, pvID).Scan(&pendingCount)

	allComplete := pendingCount == 0
	policyStatus := "in_review"

	if allComplete {
		// Get version number for approved_version
		var versionNum int
		database.DB.QueryRow(`SELECT version_number FROM policy_versions WHERE id = $1`, pvID).Scan(&versionNum)

		database.DB.Exec(`
			UPDATE policies SET status = 'approved', approved_at = $1, approved_version = $2
			WHERE id = $3
		`, now, versionNum, policyID)
		policyStatus = "approved"

		middleware.LogAudit(c, "policy.status_changed", "policy", &policyID, map[string]interface{}{
			"from": "in_review", "to": "approved",
		})
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                    signoffID,
		"status":                "approved",
		"decided_at":            now,
		"comments":              req.Comments,
		"policy_status":         policyStatus,
		"all_signoffs_complete": allComplete,
	}))
}

// RejectSignoff rejects a sign-off request.
func RejectSignoff(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	policyID := c.Param("id")
	signoffID := c.Param("signoff_id")

	var req models.SignoffDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if req.Comments == nil || *req.Comments == "" {
		c.JSON(http.StatusBadRequest, errorResponse("REJECTION_REQUIRES_COMMENTS", "Comments are required for rejection"))
		return
	}

	// Get signoff
	var signerID, status string
	err := database.DB.QueryRow(`
		SELECT signer_id, status
		FROM policy_signoffs
		WHERE id = $1 AND policy_id = $2 AND org_id = $3
	`, signoffID, policyID, orgID).Scan(&signerID, &status)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Sign-off not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to reject sign-off"))
		return
	}

	if signerID != userID {
		c.JSON(http.StatusForbidden, errorResponse("NOT_SIGNER", "You are not the designated signer"))
		return
	}
	if status != models.SignoffStatusPending {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATUS_TRANSITION", "Sign-off is not in pending status"))
		return
	}

	now := time.Now()
	_, err = database.DB.Exec(`
		UPDATE policy_signoffs SET status = 'rejected', decided_at = $1, comments = $2
		WHERE id = $3
	`, now, *req.Comments, signoffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to reject sign-off"))
		return
	}

	middleware.LogAudit(c, "policy_signoff.rejected", "policy_signoff", &signoffID, map[string]interface{}{
		"policy_id": policyID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":            signoffID,
		"status":        "rejected",
		"decided_at":    now,
		"comments":      *req.Comments,
		"policy_status": "in_review",
	}))
}

// WithdrawSignoff withdraws a pending sign-off request.
func WithdrawSignoff(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")
	signoffID := c.Param("signoff_id")

	// Get signoff
	var requestedByID, status string
	err := database.DB.QueryRow(`
		SELECT requested_by, status
		FROM policy_signoffs
		WHERE id = $1 AND policy_id = $2 AND org_id = $3
	`, signoffID, policyID, orgID).Scan(&requestedByID, &status)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Sign-off not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to withdraw sign-off"))
		return
	}

	if status != models.SignoffStatusPending {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATUS_TRANSITION", "Sign-off is not in pending status"))
		return
	}

	// Auth: original requester, compliance_manager, ciso
	isRequester := requestedByID == userID
	if !isRequester && !models.HasRole(userRole, models.PolicyPublishRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to withdraw this sign-off"))
		return
	}

	now := time.Now()
	_, err = database.DB.Exec(`
		UPDATE policy_signoffs SET status = 'withdrawn', decided_at = $1
		WHERE id = $2
	`, now, signoffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to withdraw sign-off"))
		return
	}

	middleware.LogAudit(c, "policy_signoff.withdrawn", "policy_signoff", &signoffID, map[string]interface{}{
		"policy_id": policyID,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         signoffID,
		"status":     "withdrawn",
		"decided_at": now,
	}))
}

// ListPendingSignoffs lists pending signoffs for the authenticated user.
func ListPendingSignoffs(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"ps.signer_id = $1", "ps.org_id = $2", "ps.status = 'pending'"}
	args := []interface{}{userID, orgID}
	argN := 3

	if v := c.Query("urgency"); v != "" {
		switch v {
		case "overdue":
			where = append(where, "ps.due_date IS NOT NULL AND ps.due_date < NOW()")
		case "due_soon":
			where = append(where, "ps.due_date IS NOT NULL AND ps.due_date >= NOW() AND ps.due_date < NOW() + INTERVAL '3 days'")
		case "on_time":
			where = append(where, "(ps.due_date IS NULL OR ps.due_date >= NOW() + INTERVAL '3 days')")
		}
	}

	whereClause := joinWhere(where)

	var total int
	database.DB.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM policy_signoffs ps WHERE %s`, whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ps.id,
			p.id, p.identifier, p.title, p.category,
			pv.id, pv.version_number, pv.content_summary, pv.word_count,
			ps.requested_by, u_req.first_name, u_req.last_name,
			ps.requested_at, ps.due_date, ps.reminder_count
		FROM policy_signoffs ps
		JOIN policies p ON p.id = ps.policy_id
		LEFT JOIN policy_versions pv ON pv.id = ps.policy_version_id
		LEFT JOIN users u_req ON u_req.id = ps.requested_by
		WHERE %s
		ORDER BY COALESCE(ps.due_date, '2999-12-31'::date), ps.requested_at
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list pending signoffs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list pending sign-offs"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id                                     string
			pID, pIdentifier, pTitle, pCategory    string
			pvID                                   *string
			pvVersionNum                           *int
			pvSummary                              *string
			pvWordCount                            *int
			reqByID, reqFirst, reqLast             string
			requestedAt                            time.Time
			dueDate                                *time.Time
			reminderCount                          int
		)
		if err := rows.Scan(
			&id,
			&pID, &pIdentifier, &pTitle, &pCategory,
			&pvID, &pvVersionNum, &pvSummary, &pvWordCount,
			&reqByID, &reqFirst, &reqLast,
			&requestedAt, &dueDate, &reminderCount,
		); err != nil {
			continue
		}

		urgency := "on_time"
		if dueDate != nil {
			if dueDate.Before(time.Now()) {
				urgency = "overdue"
			} else if dueDate.Before(time.Now().Add(3 * 24 * time.Hour)) {
				urgency = "due_soon"
			}
		}

		r := gin.H{
			"id": id,
			"policy": gin.H{
				"id":         pID,
				"identifier": pIdentifier,
				"title":      pTitle,
				"category":   pCategory,
			},
			"requested_by":  gin.H{"id": reqByID, "name": reqFirst + " " + reqLast},
			"requested_at":  requestedAt,
			"due_date":       dueDate,
			"urgency":        urgency,
			"reminder_count": reminderCount,
		}

		if pvID != nil {
			r["policy_version"] = gin.H{
				"id":              *pvID,
				"version_number":  pvVersionNum,
				"content_summary": pvSummary,
				"word_count":      pvWordCount,
			}
		}

		results = append(results, r)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}

// RemindSignoffs sends reminders for pending signoffs.
func RemindSignoffs(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	// Verify policy and get owner
	var policyOwnerID *string
	err := database.DB.QueryRow(`SELECT owner_id FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&policyOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to send reminders"))
		return
	}

	// Auth: owner, compliance_manager, ciso, or original requester
	isOwner := policyOwnerID != nil && *policyOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.PolicyPublishRoles) {
		// Check if user is a requester
		var isRequester bool
		database.DB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM policy_signoffs WHERE policy_id = $1 AND requested_by = $2 AND status = 'pending')
		`, policyID, userID).Scan(&isRequester)
		if !isRequester {
			c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to send reminders"))
			return
		}
	}

	var req models.RemindSignoffRequest
	c.ShouldBindJSON(&req)

	// Find pending signoffs
	where := "ps.policy_id = $1 AND ps.org_id = $2 AND ps.status = 'pending'"
	args := []interface{}{policyID, orgID}

	if len(req.SignoffIDs) > 0 {
		placeholders := make([]string, len(req.SignoffIDs))
		for i, sid := range req.SignoffIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+3)
			args = append(args, sid)
		}
		where += fmt.Sprintf(" AND ps.id IN (%s)", joinStrings(placeholders))
	}

	rows, err := database.DB.Query(fmt.Sprintf(`
		SELECT ps.id, ps.signer_id, u.first_name, u.last_name, ps.reminder_sent_at, ps.reminder_count
		FROM policy_signoffs ps
		LEFT JOIN users u ON u.id = ps.signer_id
		WHERE %s
	`, where), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to find pending signoffs"))
		return
	}
	defer rows.Close()

	type pendingSignoff struct {
		ID             string
		SignerID       string
		SignerName     string
		ReminderSentAt *time.Time
		ReminderCount  int
	}

	var pending []pendingSignoff
	for rows.Next() {
		var ps pendingSignoff
		var first, last string
		if err := rows.Scan(&ps.ID, &ps.SignerID, &first, &last, &ps.ReminderSentAt, &ps.ReminderCount); err == nil {
			ps.SignerName = first + " " + last
			pending = append(pending, ps)
		}
	}

	if len(pending) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("NO_PENDING_SIGNOFFS", "No pending sign-offs found"))
		return
	}

	now := time.Now()
	signers := []gin.H{}
	sent := 0

	for _, ps := range pending {
		// Rate limit: 1 reminder per 24h
		if ps.ReminderSentAt != nil && now.Sub(*ps.ReminderSentAt) < 24*time.Hour {
			continue // skip, already sent recently
		}

		database.DB.Exec(`
			UPDATE policy_signoffs SET reminder_sent_at = $1, reminder_count = reminder_count + 1
			WHERE id = $2
		`, now, ps.ID)

		signers = append(signers, gin.H{
			"id":             ps.SignerID,
			"name":           ps.SignerName,
			"reminder_count": ps.ReminderCount + 1,
		})
		sent++
	}

	if sent == 0 {
		c.JSON(http.StatusTooManyRequests, errorResponse("REMINDER_RATE_LIMITED", "Reminder already sent within the last 24 hours"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"reminders_sent": sent,
		"signers":        signers,
	}))
}

// joinWhere joins where conditions with AND.
func joinWhere(conditions []string) string {
	return strings.Join(conditions, " AND ")
}

// joinStrings joins strings with commas.
func joinStrings(s []string) string {
	return strings.Join(s, ", ")
}
