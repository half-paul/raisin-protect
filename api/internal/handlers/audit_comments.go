package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListAuditComments lists comments for an audit with optional target filtering.
func ListAuditComments(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	where := []string{"ac.audit_id = $1", "ac.org_id = $2", "ac.parent_comment_id IS NULL"}
	args := []interface{}{auditID, orgID}
	argN := 3

	// Auditor cannot see internal comments
	if userRole == models.RoleAuditor {
		where = append(where, "ac.is_internal = FALSE")
	}

	if v := c.Query("target_type"); v != "" {
		where = append(where, fmt.Sprintf("ac.target_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("target_id"); v != "" {
		where = append(where, fmt.Sprintf("ac.target_id = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := ""
	for i, w := range where {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += w
	}

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM audit_comments ac WHERE %s", whereClause), args...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT ac.id, ac.audit_id, ac.target_type, ac.target_id,
		       ac.author_id, COALESCE(u.first_name || ' ' || u.last_name, ''), COALESCE(u.role, ''),
		       ac.body, ac.is_internal, ac.edited_at,
		       ac.created_at, ac.updated_at
		FROM audit_comments ac
		LEFT JOIN users u ON ac.author_id = u.id
		WHERE %s
		ORDER BY ac.created_at ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list audit comments")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list comments"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, aID, targetType, targetID    string
			authorID, authorName, authorRole string
			body                             string
			isInternal                       bool
			editedAt                         *time.Time
			createdAt, updatedAt             time.Time
		)
		if err := rows.Scan(
			&id, &aID, &targetType, &targetID,
			&authorID, &authorName, &authorRole,
			&body, &isInternal, &editedAt,
			&createdAt, &updatedAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan comment row")
			continue
		}

		// Fetch replies
		replies := fetchCommentReplies(id, orgID, userRole)

		results = append(results, gin.H{
			"id": id, "audit_id": aID,
			"target_type": targetType, "target_id": targetID,
			"author_id": authorID, "author_name": authorName, "author_role": authorRole,
			"body": body, "parent_comment_id": nil,
			"is_internal": isInternal, "edited_at": editedAt,
			"replies":    replies,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       results,
		"pagination": gin.H{"page": page, "per_page": perPage, "total": total, "total_pages": (total + perPage - 1) / perPage},
	})
}

// fetchCommentReplies returns replies for a parent comment.
func fetchCommentReplies(parentID, orgID, userRole string) []gin.H {
	internalFilter := ""
	if userRole == models.RoleAuditor {
		internalFilter = " AND ac.is_internal = FALSE"
	}

	rows, err := database.DB.Query(fmt.Sprintf(`
		SELECT ac.id, ac.author_id, COALESCE(u.first_name || ' ' || u.last_name, ''), COALESCE(u.role, ''),
		       ac.body, ac.parent_comment_id, ac.is_internal, ac.edited_at, ac.created_at
		FROM audit_comments ac
		LEFT JOIN users u ON ac.author_id = u.id
		WHERE ac.parent_comment_id = $1 AND ac.org_id = $2%s
		ORDER BY ac.created_at ASC
	`, internalFilter), parentID, orgID)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()

	replies := []gin.H{}
	for rows.Next() {
		var (
			id, authorID, authorName, authorRole string
			body                                 string
			parentCommentID                      *string
			isInternal                           bool
			editedAt                             *time.Time
			createdAt                            time.Time
		)
		rows.Scan(&id, &authorID, &authorName, &authorRole,
			&body, &parentCommentID, &isInternal, &editedAt, &createdAt)
		replies = append(replies, gin.H{
			"id": id, "author_id": authorID, "author_name": authorName, "author_role": authorRole,
			"body": body, "parent_comment_id": parentCommentID,
			"is_internal": isInternal, "edited_at": editedAt, "created_at": createdAt,
		})
	}
	return replies
}

// CreateAuditComment creates a comment on an audit, request, or finding.
func CreateAuditComment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var req models.CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if !models.IsValidCommentTargetType(req.TargetType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid target_type"))
		return
	}
	if utf8.RuneCountInString(req.Body) > 10000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Body must not exceed 10000 characters"))
		return
	}

	// Auditors cannot create internal comments
	isInternal := false
	if req.IsInternal != nil {
		isInternal = *req.IsInternal
	}
	if isInternal && userRole == models.RoleAuditor {
		c.JSON(http.StatusBadRequest, errorResponse("AUDIT_INTERNAL_COMMENT_DENIED", "Auditors cannot create internal comments"))
		return
	}

	// Verify target exists within this audit
	var targetExists bool
	switch req.TargetType {
	case models.CommentTargetAudit:
		targetExists = req.TargetID == auditID
	case models.CommentTargetRequest:
		database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_requests WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
			req.TargetID, auditID, orgID).Scan(&targetExists)
	case models.CommentTargetFinding:
		database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_findings WHERE id = $1 AND audit_id = $2 AND org_id = $3)",
			req.TargetID, auditID, orgID).Scan(&targetExists)
	}
	if !targetExists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Target entity not found"))
		return
	}

	// Verify parent comment if provided
	if req.ParentCommentID != nil {
		var parentExists bool
		database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM audit_comments WHERE id = $1 AND target_type = $2 AND target_id = $3 AND org_id = $4)",
			*req.ParentCommentID, req.TargetType, req.TargetID, orgID).Scan(&parentExists)
		if !parentExists {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Parent comment not found on same target"))
			return
		}
	}

	commentID := uuid.New().String()
	now := time.Now()

	_, err := database.DB.Exec(`
		INSERT INTO audit_comments (id, org_id, audit_id, target_type, target_id,
		                            author_id, body, parent_comment_id, is_internal,
		                            created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
	`, commentID, orgID, auditID, req.TargetType, req.TargetID,
		userID, req.Body, req.ParentCommentID, isInternal, now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create comment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create comment"))
		return
	}

	// Get author info
	var authorName, authorRoleStr string
	database.DB.QueryRow("SELECT COALESCE(first_name || ' ' || last_name, ''), role FROM users WHERE id = $1", userID).Scan(&authorName, &authorRoleStr)

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id": commentID, "audit_id": auditID,
		"target_type": req.TargetType, "target_id": req.TargetID,
		"author_id": userID, "author_name": authorName, "author_role": authorRoleStr,
		"body": req.Body, "parent_comment_id": req.ParentCommentID,
		"is_internal": isInternal, "created_at": now,
	}))
}

// UpdateAuditComment edits a comment (only author can edit).
func UpdateAuditComment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	commentID := c.Param("cid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var req models.UpdateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if utf8.RuneCountInString(req.Body) > 10000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Body must not exceed 10000 characters"))
		return
	}

	// Check comment exists and user is author
	var authorID string
	err := database.DB.QueryRow("SELECT author_id FROM audit_comments WHERE id = $1 AND audit_id = $2 AND org_id = $3",
		commentID, auditID, orgID).Scan(&authorID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_COMMENT_NOT_FOUND", "Comment not found"))
		return
	}
	if authorID != userID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Only the author can edit a comment"))
		return
	}

	now := time.Now()
	database.DB.Exec(
		"UPDATE audit_comments SET body = $1, edited_at = $2, updated_at = $2 WHERE id = $3 AND org_id = $4",
		req.Body, now, commentID, orgID,
	)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id": commentID, "body": req.Body, "edited_at": now, "updated_at": now,
	}))
}

// DeleteAuditComment deletes a comment (author or compliance_manager/ciso).
func DeleteAuditComment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")
	commentID := c.Param("cid")

	if _, ok := checkAuditAccess(c, auditID, orgID, userID, userRole); !ok {
		return
	}

	var authorID string
	err := database.DB.QueryRow("SELECT author_id FROM audit_comments WHERE id = $1 AND audit_id = $2 AND org_id = $3",
		commentID, auditID, orgID).Scan(&authorID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("AUDIT_COMMENT_NOT_FOUND", "Comment not found"))
		return
	}

	// Only author or admin roles
	isAuthor := authorID == userID
	if !isAuthor && !models.HasRole(userRole, models.AuditCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to delete this comment"))
		return
	}

	database.DB.Exec("DELETE FROM audit_comments WHERE id = $1 AND org_id = $2", commentID, orgID)

	c.Status(http.StatusNoContent)
}
