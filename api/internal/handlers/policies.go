package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
)

// HTML sanitizer using bluemonday for secure XSS prevention.
// Allows common formatting tags but strips scripts, event handlers, and dangerous elements.
var policySanitizer = bluemonday.UGCPolicy()

func init() {
	// UGCPolicy allows user-generated content formatting (p, h1-h6, ul, ol, li, b, i, a, etc.)
	// but strips scripts, iframes, forms, and event handlers.
	// Customize if needed:
	policySanitizer.AllowAttrs("class").Globally()
	policySanitizer.AllowAttrs("id").Globally()
	policySanitizer.AllowAttrs("style").Globally() // Allow inline styles (bluemonday sanitizes dangerous CSS)
}

func sanitizeHTML(content string) string {
	return policySanitizer.Sanitize(content)
}

func countWords(s string) int {
	// Strip HTML tags for word counting.
	stripped := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, " ")
	words := strings.Fields(stripped)
	return len(words)
}

func computeReviewStatus(nextReviewAt *time.Time) string {
	if nextReviewAt == nil {
		return "no_schedule"
	}
	now := time.Now()
	if nextReviewAt.Before(now) {
		return "overdue"
	}
	if nextReviewAt.Before(now.Add(30 * 24 * time.Hour)) {
		return "due_soon"
	}
	return "on_track"
}

// ListPolicies lists policies with filtering and pagination.
func ListPolicies(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"p.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	// is_template filter (default: false)
	isTemplate := c.DefaultQuery("is_template", "false")
	if isTemplate == "true" {
		where = append(where, "p.is_template = TRUE")
	} else {
		where = append(where, "p.is_template = FALSE")
	}

	if v := c.Query("status"); v != "" {
		where = append(where, fmt.Sprintf("p.status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("category"); v != "" {
		where = append(where, fmt.Sprintf("p.category = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("owner_id"); v != "" {
		where = append(where, fmt.Sprintf("p.owner_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("framework_id"); v != "" {
		where = append(where, fmt.Sprintf("p.template_framework_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("tags"); v != "" {
		tags := strings.Split(v, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				where = append(where, fmt.Sprintf("$%d = ANY(p.tags)", argN))
				args = append(args, tag)
				argN++
			}
		}
	}
	if v := c.Query("review_status"); v != "" {
		switch v {
		case "overdue":
			where = append(where, "p.next_review_at IS NOT NULL AND p.next_review_at < NOW()")
		case "due_soon":
			where = append(where, "p.next_review_at IS NOT NULL AND p.next_review_at >= NOW() AND p.next_review_at < NOW() + INTERVAL '30 days'")
		case "on_track":
			where = append(where, "p.next_review_at IS NOT NULL AND p.next_review_at >= NOW() + INTERVAL '30 days'")
		case "no_schedule":
			where = append(where, "p.next_review_at IS NULL")
		}
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("to_tsvector('english', p.title || ' ' || COALESCE(p.description, '')) @@ plainto_tsquery('english', $%d)", argN))
		args = append(args, v)
		argN++
	}

	// Sorting
	sortCol := c.DefaultQuery("sort", "identifier")
	sortOrder := strings.ToUpper(c.DefaultQuery("order", "ASC"))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}
	allowedSorts := map[string]string{
		"identifier":     "p.identifier",
		"title":          "p.title",
		"category":       "p.category",
		"status":         "p.status",
		"next_review_at": "p.next_review_at",
		"published_at":   "p.published_at",
		"created_at":     "p.created_at",
		"updated_at":     "p.updated_at",
	}
	orderBy := "p.identifier"
	if col, ok := allowedSorts[sortCol]; ok {
		orderBy = col
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM policies p WHERE %s`, whereClause)
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		log.Error().Err(err).Msg("Failed to count policies")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to count policies"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT
			p.id, p.identifier, p.title, p.description, p.category, p.status,
			p.owner_id, p.review_frequency_days, p.next_review_at, p.last_reviewed_at,
			p.approved_at, p.published_at, p.is_template, p.cloned_from_policy_id,
			p.tags, p.created_at, p.updated_at,
			u_owner.id, u_owner.first_name, u_owner.last_name, u_owner.email,
			u_secondary.id, u_secondary.first_name, u_secondary.last_name, u_secondary.email,
			pv.id, pv.version_number, pv.change_summary, pv.word_count, pv.created_at,
			COALESCE((SELECT COUNT(*) FROM policy_controls pc WHERE pc.policy_id = p.id), 0),
			COALESCE((SELECT COUNT(*) FROM policy_signoffs ps WHERE ps.policy_id = p.id AND ps.status = 'pending'), 0)
		FROM policies p
		LEFT JOIN users u_owner ON u_owner.id = p.owner_id
		LEFT JOIN users u_secondary ON u_secondary.id = p.secondary_owner_id
		LEFT JOIN policy_versions pv ON pv.id = p.current_version_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, sortOrder, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list policies")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list policies"))
		return
	}
	defer rows.Close()

	policies := []gin.H{}
	for rows.Next() {
		var (
			id, identifier, title, category, status string
			description                             *string
			ownerID                                 *string
			reviewFreqDays                          *int
			nextReviewAt, lastReviewedAt            *time.Time
			approvedAt, publishedAt                 *time.Time
			isTemplate                              bool
			clonedFromID                            *string
			tags                                    pq.StringArray
			createdAt, updatedAt                    time.Time
			// Owner
			oID, oFirst, oLast, oEmail *string
			// Secondary owner
			sID, sFirst, sLast, sEmail *string
			// Current version
			cvID           *string
			cvVersionNum   *int
			cvChangeSumm   *string
			cvWordCount    *int
			cvCreatedAt    *time.Time
			linkedControls int
			pendingSignoffs int
		)
		if err := rows.Scan(
			&id, &identifier, &title, &description, &category, &status,
			&ownerID, &reviewFreqDays, &nextReviewAt, &lastReviewedAt,
			&approvedAt, &publishedAt, &isTemplate, &clonedFromID,
			&tags, &createdAt, &updatedAt,
			&oID, &oFirst, &oLast, &oEmail,
			&sID, &sFirst, &sLast, &sEmail,
			&cvID, &cvVersionNum, &cvChangeSumm, &cvWordCount, &cvCreatedAt,
			&linkedControls, &pendingSignoffs,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan policy row")
			continue
		}

		policy := gin.H{
			"id":                     id,
			"identifier":             identifier,
			"title":                  title,
			"description":            description,
			"category":               category,
			"status":                 status,
			"is_template":            isTemplate,
			"cloned_from_policy_id":  clonedFromID,
			"review_frequency_days":  reviewFreqDays,
			"next_review_at":         nextReviewAt,
			"last_reviewed_at":       lastReviewedAt,
			"review_status":          computeReviewStatus(nextReviewAt),
			"approved_at":            approvedAt,
			"published_at":           publishedAt,
			"linked_controls_count":  linkedControls,
			"pending_signoffs_count": pendingSignoffs,
			"tags":                   []string(tags),
			"created_at":             createdAt,
			"updated_at":             updatedAt,
		}

		if oID != nil {
			policy["owner"] = gin.H{"id": *oID, "name": *oFirst + " " + *oLast, "email": *oEmail}
		} else {
			policy["owner"] = nil
		}
		if sID != nil {
			policy["secondary_owner"] = gin.H{"id": *sID, "name": *sFirst + " " + *sLast, "email": *sEmail}
		} else {
			policy["secondary_owner"] = nil
		}
		if cvID != nil {
			policy["current_version"] = gin.H{
				"id":             *cvID,
				"version_number": cvVersionNum,
				"change_summary": cvChangeSumm,
				"word_count":     cvWordCount,
				"created_at":     cvCreatedAt,
			}
		} else {
			policy["current_version"] = nil
		}

		policies = append(policies, policy)
	}

	if policies == nil {
		policies = []gin.H{}
	}

	c.JSON(http.StatusOK, listResponse(c, policies, total, page, perPage))
}

// GetPolicy returns a single policy with full details.
func GetPolicy(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	policyID := c.Param("id")

	var (
		id, identifier, title, category, status string
		description                             *string
		ownerID, secondaryOwnerID               *string
		currentVersionID                        *string
		reviewFreqDays                          *int
		nextReviewAt, lastReviewedAt            *time.Time
		approvedAt, publishedAt                 *time.Time
		approvedVersion                         *int
		isTemplate                              bool
		templateFrameworkID, clonedFromID        *string
		tags                                    pq.StringArray
		metadata                                string
		createdAt, updatedAt                    time.Time
	)

	err := database.DB.QueryRow(`
		SELECT id, identifier, title, description, category, status,
			current_version_id, owner_id, secondary_owner_id,
			review_frequency_days, next_review_at, last_reviewed_at,
			approved_at, approved_version, published_at,
			is_template, template_framework_id, cloned_from_policy_id,
			tags, metadata::text, created_at, updated_at
		FROM policies
		WHERE id = $1 AND org_id = $2
	`, policyID, orgID).Scan(
		&id, &identifier, &title, &description, &category, &status,
		&currentVersionID, &ownerID, &secondaryOwnerID,
		&reviewFreqDays, &nextReviewAt, &lastReviewedAt,
		&approvedAt, &approvedVersion, &publishedAt,
		&isTemplate, &templateFrameworkID, &clonedFromID,
		&tags, &metadata, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get policy"))
		return
	}

	// Fetch owner info
	var owner interface{}
	if ownerID != nil {
		var oFirst, oLast, oEmail string
		if err := database.DB.QueryRow(`SELECT first_name, last_name, email FROM users WHERE id = $1`, *ownerID).Scan(&oFirst, &oLast, &oEmail); err == nil {
			owner = gin.H{"id": *ownerID, "name": oFirst + " " + oLast, "email": oEmail}
		}
	}

	var secondaryOwner interface{}
	if secondaryOwnerID != nil {
		var sFirst, sLast, sEmail string
		if err := database.DB.QueryRow(`SELECT first_name, last_name, email FROM users WHERE id = $1`, *secondaryOwnerID).Scan(&sFirst, &sLast, &sEmail); err == nil {
			secondaryOwner = gin.H{"id": *secondaryOwnerID, "name": sFirst + " " + sLast, "email": sEmail}
		}
	}

	// Fetch current version with content
	var currentVersion interface{}
	if currentVersionID != nil {
		var (
			cvID, cvContent, cvContentFormat, cvChangeType string
			cvContentSummary, cvChangeSummary              *string
			cvVersionNum                                   int
			cvWordCount, cvCharCount                       *int
			cvCreatedBy                                    *string
			cvCreatedAt                                    time.Time
		)
		err := database.DB.QueryRow(`
			SELECT pv.id, pv.version_number, pv.content, pv.content_format,
				pv.content_summary, pv.change_summary, pv.change_type,
				pv.word_count, pv.character_count, pv.created_by, pv.created_at
			FROM policy_versions pv
			WHERE pv.id = $1
		`, *currentVersionID).Scan(
			&cvID, &cvVersionNum, &cvContent, &cvContentFormat,
			&cvContentSummary, &cvChangeSummary, &cvChangeType,
			&cvWordCount, &cvCharCount, &cvCreatedBy, &cvCreatedAt,
		)
		if err == nil {
			cv := gin.H{
				"id":              cvID,
				"version_number":  cvVersionNum,
				"content":         cvContent,
				"content_format":  cvContentFormat,
				"content_summary": cvContentSummary,
				"change_summary":  cvChangeSummary,
				"change_type":     cvChangeType,
				"word_count":      cvWordCount,
				"character_count": cvCharCount,
				"created_at":      cvCreatedAt,
			}
			if cvCreatedBy != nil {
				var cbFirst, cbLast string
				if err := database.DB.QueryRow(`SELECT first_name, last_name FROM users WHERE id = $1`, *cvCreatedBy).Scan(&cbFirst, &cbLast); err == nil {
					cv["created_by"] = gin.H{"id": *cvCreatedBy, "name": cbFirst + " " + cbLast}
				}
			}
			currentVersion = cv
		}
	}

	// Fetch linked controls
	linkedControls := []gin.H{}
	ctrlRows, err := database.DB.Query(`
		SELECT pc.id, c.id, c.identifier, c.title, c.category, pc.coverage
		FROM policy_controls pc
		JOIN controls c ON c.id = pc.control_id
		WHERE pc.policy_id = $1 AND pc.org_id = $2
		ORDER BY c.identifier
	`, policyID, orgID)
	if err == nil {
		defer ctrlRows.Close()
		for ctrlRows.Next() {
			var pcID, cID, cIdentifier, cTitle, cCategory, coverage string
			if err := ctrlRows.Scan(&pcID, &cID, &cIdentifier, &cTitle, &cCategory, &coverage); err == nil {
				linkedControls = append(linkedControls, gin.H{
					"id":         cID,
					"identifier": cIdentifier,
					"title":      cTitle,
					"category":   cCategory,
					"coverage":   coverage,
				})
			}
		}
	}

	// Fetch signoff summary for current version
	signoffSummary := gin.H{"total": 0, "approved": 0, "pending": 0, "rejected": 0}
	if currentVersionID != nil {
		var total, approved, pending, rejected int
		database.DB.QueryRow(`
			SELECT
				COUNT(*),
				COUNT(*) FILTER (WHERE status = 'approved'),
				COUNT(*) FILTER (WHERE status = 'pending'),
				COUNT(*) FILTER (WHERE status = 'rejected')
			FROM policy_signoffs
			WHERE policy_version_id = $1
		`, *currentVersionID).Scan(&total, &approved, &pending, &rejected)
		signoffSummary = gin.H{"total": total, "approved": approved, "pending": pending, "rejected": rejected}
	}

	result := gin.H{
		"id":                     id,
		"identifier":             identifier,
		"title":                  title,
		"description":            description,
		"category":               category,
		"status":                 status,
		"owner":                  owner,
		"secondary_owner":        secondaryOwner,
		"current_version":        currentVersion,
		"review_frequency_days":  reviewFreqDays,
		"next_review_at":         nextReviewAt,
		"last_reviewed_at":       lastReviewedAt,
		"review_status":          computeReviewStatus(nextReviewAt),
		"approved_at":            approvedAt,
		"approved_version":       approvedVersion,
		"published_at":           publishedAt,
		"is_template":            isTemplate,
		"template_framework":     templateFrameworkID,
		"cloned_from_policy_id":  clonedFromID,
		"linked_controls":        linkedControls,
		"signoff_summary":        signoffSummary,
		"tags":                   []string(tags),
		"metadata":               metadata,
		"created_at":             createdAt,
		"updated_at":             updatedAt,
	}

	c.JSON(http.StatusOK, successResponse(c, result))
}

// CreatePolicy creates a new policy with its initial version.
func CreatePolicy(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if len(req.Identifier) > 50 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Identifier must be 50 characters or less"))
		return
	}
	if len(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must be 500 characters or less"))
		return
	}
	if !models.IsValidPolicyCategory(req.Category) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid policy category"))
		return
	}
	if len(req.Content) > 1024*1024 {
		c.JSON(http.StatusBadRequest, errorResponse("CONTENT_TOO_LARGE", "Policy content exceeds 1MB limit"))
		return
	}

	contentFormat := models.ContentFormatHTML
	if req.ContentFormat != nil && *req.ContentFormat != "" {
		if !models.IsValidContentFormat(*req.ContentFormat) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid content format"))
			return
		}
		contentFormat = *req.ContentFormat
	}

	// Sanitize content
	content := req.Content
	if contentFormat == models.ContentFormatHTML {
		content = sanitizeHTML(content)
	}

	ownerID := userID
	if req.OwnerID != nil && *req.OwnerID != "" {
		ownerID = *req.OwnerID
	}

	wordCount := countWords(content)
	charCount := utf8.RuneCountInString(content)

	policyID := uuid.New().String()
	versionID := uuid.New().String()

	tx, err := database.DB.Begin()
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create policy"))
		return
	}
	defer tx.Rollback()

	// Check identifier uniqueness
	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM policies WHERE org_id = $1 AND identifier = $2)`, orgID, req.Identifier).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check identifier")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create policy"))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse("DUPLICATE_IDENTIFIER", "Policy identifier already exists"))
		return
	}

	// Create policy
	_, err = tx.Exec(`
		INSERT INTO policies (id, org_id, identifier, title, description, category, status,
			owner_id, secondary_owner_id, review_frequency_days, is_template, tags, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, 'draft', $7, $8, $9, FALSE, $10, '{}')
	`, policyID, orgID, req.Identifier, req.Title, req.Description, req.Category,
		ownerID, req.SecondaryOwnerID, req.ReviewFreqDays, pq.Array(req.Tags))
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create policy"))
		return
	}

	// Create initial version
	_, err = tx.Exec(`
		INSERT INTO policy_versions (id, org_id, policy_id, version_number, is_current,
			content, content_format, content_summary, change_summary, change_type,
			word_count, character_count, created_by)
		VALUES ($1, $2, $3, 1, TRUE, $4, $5, $6, 'Initial version', 'initial', $7, $8, $9)
	`, versionID, orgID, policyID, content, contentFormat, req.ContentSummary, wordCount, charCount, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert policy version")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create policy"))
		return
	}

	// Set current version
	_, err = tx.Exec(`UPDATE policies SET current_version_id = $1 WHERE id = $2`, versionID, policyID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update current version")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create policy"))
		return
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit policy creation")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create policy"))
		return
	}

	// Audit log
	middleware.LogAudit(c, "policy.created", "policy", &policyID, map[string]interface{}{
		"identifier": req.Identifier, "title": req.Title,
	})
	middleware.LogAudit(c, "policy_version.created", "policy_version", &versionID, map[string]interface{}{
		"policy_id": policyID, "version_number": 1,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":         policyID,
		"identifier": req.Identifier,
		"title":      req.Title,
		"status":     "draft",
		"current_version": gin.H{
			"id":             versionID,
			"version_number": 1,
			"change_type":    "initial",
		},
		"created_at": time.Now(),
	}))
}

// UpdatePolicy updates policy metadata (not content).
func UpdatePolicy(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	// Fetch current policy for owner check
	var currentOwnerID *string
	var currentStatus string
	err := database.DB.QueryRow(`SELECT owner_id, status FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&currentOwnerID, &currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update policy"))
		return
	}

	if currentStatus == models.PolicyStatusArchived {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("POLICY_ARCHIVED", "Cannot modify archived policy"))
		return
	}

	// Auth check: must be owner, compliance_manager, ciso, or security_engineer
	isOwner := currentOwnerID != nil && *currentOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.PolicyCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to update this policy"))
		return
	}

	var req models.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	sets := []string{}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		if len(*req.Title) > 500 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must be 500 characters or less"))
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
		if !models.IsValidPolicyCategory(*req.Category) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid policy category"))
			return
		}
		sets = append(sets, fmt.Sprintf("category = $%d", argN))
		args = append(args, *req.Category)
		argN++
	}
	if req.OwnerID != nil {
		sets = append(sets, fmt.Sprintf("owner_id = $%d", argN))
		args = append(args, *req.OwnerID)
		argN++
	}
	if req.SecondaryOwnerID != nil {
		if *req.SecondaryOwnerID == "" {
			sets = append(sets, "secondary_owner_id = NULL")
		} else {
			sets = append(sets, fmt.Sprintf("secondary_owner_id = $%d", argN))
			args = append(args, *req.SecondaryOwnerID)
			argN++
		}
	}
	if req.ReviewFreqDays != nil {
		sets = append(sets, fmt.Sprintf("review_frequency_days = $%d", argN))
		args = append(args, *req.ReviewFreqDays)
		argN++
	}
	if req.NextReviewAt != nil {
		sets = append(sets, fmt.Sprintf("next_review_at = $%d", argN))
		args = append(args, *req.NextReviewAt)
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

	query := fmt.Sprintf(`UPDATE policies SET %s WHERE id = $%d AND org_id = $%d RETURNING id, identifier, title, status, updated_at`,
		strings.Join(sets, ", "), argN, argN+1)
	args = append(args, policyID, orgID)

	var rID, rIdentifier, rTitle, rStatus string
	var rUpdatedAt time.Time
	err = database.DB.QueryRow(query, args...).Scan(&rID, &rIdentifier, &rTitle, &rStatus, &rUpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to update policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update policy"))
		return
	}

	middleware.LogAudit(c, "policy.updated", "policy", &policyID, nil)

	// Check for owner change
	if req.OwnerID != nil && (currentOwnerID == nil || *currentOwnerID != *req.OwnerID) {
		middleware.LogAudit(c, "policy.owner_changed", "policy", &policyID, map[string]interface{}{
			"old_owner_id": currentOwnerID, "new_owner_id": *req.OwnerID,
		})
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         rID,
		"identifier": rIdentifier,
		"title":      rTitle,
		"status":     rStatus,
		"updated_at": rUpdatedAt,
	}))
}

// ArchivePolicy archives a policy.
func ArchivePolicy(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	// RBAC: Only CISO or Compliance Manager can archive policies
	if !models.HasRole(userRole, models.PolicyArchiveRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to archive policies"))
		return
	}

	var currentStatus string
	err := database.DB.QueryRow(`SELECT status FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to archive policy"))
		return
	}

	if currentStatus == models.PolicyStatusArchived {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATUS_TRANSITION", "Policy is already archived"))
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to archive policy"))
		return
	}
	defer tx.Rollback()

	// Archive the policy
	var rIdentifier string
	var rUpdatedAt time.Time
	err = tx.QueryRow(`UPDATE policies SET status = 'archived' WHERE id = $1 AND org_id = $2 RETURNING identifier, updated_at`, policyID, orgID).Scan(&rIdentifier, &rUpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to archive policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to archive policy"))
		return
	}

	// Withdraw pending signoffs
	tx.Exec(`UPDATE policy_signoffs SET status = 'withdrawn', decided_at = NOW() WHERE policy_id = $1 AND status = 'pending'`, policyID)

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit archive")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to archive policy"))
		return
	}

	middleware.LogAudit(c, "policy.archived", "policy", &policyID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         policyID,
		"identifier": rIdentifier,
		"status":     "archived",
		"updated_at": rUpdatedAt,
	}))
}

// SubmitForReview transitions a policy to in_review and creates signoff requests.
func SubmitForReview(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	var req models.SubmitForReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if len(req.SignerIDs) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "At least one signer is required"))
		return
	}
	if len(req.SignerIDs) > 10 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Maximum 10 signers allowed"))
		return
	}

	// Fetch policy
	var currentStatus string
	var currentOwnerID *string
	var currentVersionID *string
	err := database.DB.QueryRow(`SELECT status, owner_id, current_version_id FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&currentStatus, &currentOwnerID, &currentVersionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to submit for review"))
		return
	}

	// Validate status transition
	if currentStatus != models.PolicyStatusDraft && currentStatus != models.PolicyStatusApproved {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATUS_TRANSITION", "Policy must be in draft or approved status to submit for review"))
		return
	}
	if currentVersionID == nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Policy has no version"))
		return
	}

	// Auth: owner, compliance_manager, ciso, security_engineer
	isOwner := currentOwnerID != nil && *currentOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.PolicyCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to submit this policy for review"))
		return
	}

	// Validate signers exist in org
	for _, signerID := range req.SignerIDs {
		var signerExists bool
		database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2 AND status = 'active')`, signerID, orgID).Scan(&signerExists)
		if !signerExists {
			c.JSON(http.StatusUnprocessableEntity, errorResponse("VALIDATION_ERROR", fmt.Sprintf("Signer %s not found or not active in this org", signerID)))
			return
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to submit for review"))
		return
	}
	defer tx.Rollback()

	// Update status to in_review
	tx.Exec(`UPDATE policies SET status = 'in_review' WHERE id = $1`, policyID)

	// Parse due_date if provided
	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			dueDate = &parsed
		}
	}

	// Create signoff requests
	signoffs := []gin.H{}
	for _, signerID := range req.SignerIDs {
		signoffID := uuid.New().String()

		// Get signer role
		var signerRole *string
		var signerFirst, signerLast string
		database.DB.QueryRow(`SELECT role, first_name, last_name FROM users WHERE id = $1`, signerID).Scan(&signerRole, &signerFirst, &signerLast)

		_, err := tx.Exec(`
			INSERT INTO policy_signoffs (id, org_id, policy_id, policy_version_id,
				signer_id, signer_role, requested_by, due_date, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
		`, signoffID, orgID, policyID, *currentVersionID, signerID, signerRole, userID, dueDate)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create signoff")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create signoff requests"))
			return
		}

		signoffs = append(signoffs, gin.H{
			"id":       signoffID,
			"signer":   gin.H{"id": signerID, "name": signerFirst + " " + signerLast},
			"status":   "pending",
			"due_date": dueDate,
		})

		middleware.LogAudit(c, "policy_signoff.requested", "policy_signoff", &signoffID, map[string]interface{}{
			"policy_id": policyID, "signer_id": signerID,
		})
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit submit for review")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to submit for review"))
		return
	}

	middleware.LogAudit(c, "policy.status_changed", "policy", &policyID, map[string]interface{}{
		"from": currentStatus, "to": "in_review",
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":               policyID,
		"status":           "in_review",
		"signoffs_created": len(signoffs),
		"signoffs":         signoffs,
	}))
}

// PublishPolicy publishes an approved policy.
func PublishPolicy(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	// RBAC: Only CISO or Compliance Manager can publish policies
	if !models.HasRole(userRole, models.PolicyPublishRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to publish policies"))
		return
	}

	var currentStatus string
	var reviewFreqDays *int
	var currentVersionID *string
	err := database.DB.QueryRow(`
		SELECT status, review_frequency_days, current_version_id
		FROM policies WHERE id = $1 AND org_id = $2
	`, policyID, orgID).Scan(&currentStatus, &reviewFreqDays, &currentVersionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to publish policy"))
		return
	}

	if currentStatus != models.PolicyStatusApproved {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATUS_TRANSITION", "Policy must be in approved status to publish"))
		return
	}

	now := time.Now()
	sets := "status = 'published', published_at = $3, last_reviewed_at = $4"
	args := []interface{}{policyID, orgID, now, now.Format("2006-01-02")}

	if reviewFreqDays != nil {
		nextReview := now.AddDate(0, 0, *reviewFreqDays)
		sets += ", next_review_at = $5"
		args = append(args, nextReview.Format("2006-01-02"))
	}

	// Get current version number for response
	var versionNum *int
	if currentVersionID != nil {
		database.DB.QueryRow(`SELECT version_number FROM policy_versions WHERE id = $1`, *currentVersionID).Scan(&versionNum)
	}

	query := fmt.Sprintf(`UPDATE policies SET %s WHERE id = $1 AND org_id = $2 RETURNING published_at`, sets)
	var publishedAt time.Time
	err = database.DB.QueryRow(query, args...).Scan(&publishedAt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to publish policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to publish policy"))
		return
	}

	middleware.LogAudit(c, "policy.status_changed", "policy", &policyID, map[string]interface{}{
		"from": "approved", "to": "published",
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":           policyID,
		"status":       "published",
		"published_at": publishedAt,
		"current_version": gin.H{
			"version_number": versionNum,
		},
	}))
}
