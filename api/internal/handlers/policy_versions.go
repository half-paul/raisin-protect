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

// ListPolicyVersions lists all versions of a policy.
func ListPolicyVersions(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	policyID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Verify policy exists and belongs to org
	var exists bool
	if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policies WHERE id = $1 AND org_id = $2)`, policyID, orgID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}

	var total int
	database.DB.QueryRow(`SELECT COUNT(*) FROM policy_versions WHERE policy_id = $1 AND org_id = $2`, policyID, orgID).Scan(&total)

	offset := (page - 1) * perPage
	rows, err := database.DB.Query(`
		SELECT pv.id, pv.version_number, pv.is_current, pv.content_format,
			pv.content_summary, pv.change_summary, pv.change_type,
			pv.word_count, pv.character_count, pv.created_by, pv.created_at,
			u.id, u.first_name, u.last_name
		FROM policy_versions pv
		LEFT JOIN users u ON u.id = pv.created_by
		WHERE pv.policy_id = $1 AND pv.org_id = $2
		ORDER BY pv.version_number DESC
		LIMIT $3 OFFSET $4
	`, policyID, orgID, perPage, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list policy versions")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list versions"))
		return
	}
	defer rows.Close()

	versions := []gin.H{}
	for rows.Next() {
		var (
			id, contentFormat, changeType    string
			versionNum                       int
			isCurrent                        bool
			contentSummary, changeSummary    *string
			wordCount, charCount             *int
			createdByID                      *string
			createdAt                        time.Time
			uID                              *string
			uFirst, uLast                    *string
		)
		if err := rows.Scan(
			&id, &versionNum, &isCurrent, &contentFormat,
			&contentSummary, &changeSummary, &changeType,
			&wordCount, &charCount, &createdByID, &createdAt,
			&uID, &uFirst, &uLast,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan version row")
			continue
		}

		v := gin.H{
			"id":              id,
			"version_number":  versionNum,
			"is_current":      isCurrent,
			"content_format":  contentFormat,
			"content_summary": contentSummary,
			"change_summary":  changeSummary,
			"change_type":     changeType,
			"word_count":      wordCount,
			"character_count": charCount,
			"created_at":      createdAt,
		}

		if uID != nil {
			v["created_by"] = gin.H{"id": *uID, "name": *uFirst + " " + *uLast}
		} else {
			v["created_by"] = nil
		}

		// Signoff summary per version
		var sTotal, sApproved, sPending, sRejected int
		database.DB.QueryRow(`
			SELECT COUNT(*),
				COUNT(*) FILTER (WHERE status = 'approved'),
				COUNT(*) FILTER (WHERE status = 'pending'),
				COUNT(*) FILTER (WHERE status = 'rejected')
			FROM policy_signoffs WHERE policy_version_id = $1
		`, id).Scan(&sTotal, &sApproved, &sPending, &sRejected)
		v["signoff_summary"] = gin.H{"total": sTotal, "approved": sApproved, "pending": sPending, "rejected": sRejected}

		versions = append(versions, v)
	}

	c.JSON(http.StatusOK, listResponse(c, versions, total, page, perPage))
}

// GetPolicyVersion gets a specific version with full content.
func GetPolicyVersion(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	policyID := c.Param("id")
	versionNumStr := c.Param("version_number")

	versionNum, err := strconv.Atoi(versionNumStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid version number"))
		return
	}

	// Verify policy exists
	var policyExists bool
	database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policies WHERE id = $1 AND org_id = $2)`, policyID, orgID).Scan(&policyExists)
	if !policyExists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}

	var (
		id, content, contentFormat, changeType string
		contentSummary, changeSummary          *string
		isCurrent                              bool
		wordCount, charCount                   *int
		createdBy                              *string
		createdAt                              time.Time
	)
	err = database.DB.QueryRow(`
		SELECT id, is_current, content, content_format, content_summary,
			change_summary, change_type, word_count, character_count,
			created_by, created_at
		FROM policy_versions
		WHERE policy_id = $1 AND org_id = $2 AND version_number = $3
	`, policyID, orgID, versionNum).Scan(
		&id, &isCurrent, &content, &contentFormat, &contentSummary,
		&changeSummary, &changeType, &wordCount, &charCount,
		&createdBy, &createdAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Version not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get policy version")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get version"))
		return
	}

	result := gin.H{
		"id":              id,
		"policy_id":       policyID,
		"version_number":  versionNum,
		"is_current":      isCurrent,
		"content":         content,
		"content_format":  contentFormat,
		"content_summary": contentSummary,
		"change_summary":  changeSummary,
		"change_type":     changeType,
		"word_count":      wordCount,
		"character_count": charCount,
		"created_at":      createdAt,
	}

	if createdBy != nil {
		var cbFirst, cbLast, cbEmail string
		if err := database.DB.QueryRow(`SELECT first_name, last_name, email FROM users WHERE id = $1`, *createdBy).Scan(&cbFirst, &cbLast, &cbEmail); err == nil {
			result["created_by"] = gin.H{"id": *createdBy, "name": cbFirst + " " + cbLast, "email": cbEmail}
		}
	}

	// Signoffs for this version
	signoffs := []gin.H{}
	sRows, err := database.DB.Query(`
		SELECT ps.id, ps.signer_id, u.first_name, u.last_name,
			ps.status, ps.decided_at, ps.comments
		FROM policy_signoffs ps
		LEFT JOIN users u ON u.id = ps.signer_id
		WHERE ps.policy_version_id = $1
		ORDER BY ps.created_at
	`, id)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var sID, signerID, sFirst, sLast, sStatus string
			var sDecidedAt *time.Time
			var sComments *string
			if err := sRows.Scan(&sID, &signerID, &sFirst, &sLast, &sStatus, &sDecidedAt, &sComments); err == nil {
				signoffs = append(signoffs, gin.H{
					"id":         sID,
					"signer":     gin.H{"id": signerID, "name": sFirst + " " + sLast},
					"status":     sStatus,
					"decided_at": sDecidedAt,
					"comments":   sComments,
				})
			}
		}
	}
	result["signoffs"] = signoffs

	c.JSON(http.StatusOK, successResponse(c, result))
}

// CreatePolicyVersion creates a new version of a policy.
func CreatePolicyVersion(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	policyID := c.Param("id")

	var req models.CreatePolicyVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if len(req.Content) > 1024*1024 {
		c.JSON(http.StatusBadRequest, errorResponse("CONTENT_TOO_LARGE", "Policy content exceeds 1MB limit"))
		return
	}

	// Fetch policy
	var currentStatus string
	var currentOwnerID *string
	err := database.DB.QueryRow(`SELECT status, owner_id FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&currentStatus, &currentOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create version"))
		return
	}

	if currentStatus == models.PolicyStatusArchived {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("POLICY_ARCHIVED", "Cannot create new versions for archived policy"))
		return
	}

	// Auth: owner, compliance_manager, ciso, security_engineer
	isOwner := currentOwnerID != nil && *currentOwnerID == userID
	if !isOwner && !models.HasRole(userRole, models.PolicyCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to create versions for this policy"))
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

	changeType := "minor"
	if req.ChangeType != nil && *req.ChangeType != "" {
		validTypes := []string{"major", "minor", "patch"}
		valid := false
		for _, t := range validTypes {
			if *req.ChangeType == t {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid change type"))
			return
		}
		changeType = *req.ChangeType
	}

	// Sanitize content
	content := req.Content
	if contentFormat == models.ContentFormatHTML {
		content = sanitizeHTML(content)
	}

	wordCount := countWords(content)
	charCount := utf8.RuneCountInString(content)

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create version"))
		return
	}
	defer tx.Rollback()

	// Get next version number
	var maxVersion int
	tx.QueryRow(`SELECT COALESCE(MAX(version_number), 0) FROM policy_versions WHERE policy_id = $1`, policyID).Scan(&maxVersion)
	newVersion := maxVersion + 1

	// Mark previous current as not current
	tx.Exec(`UPDATE policy_versions SET is_current = FALSE WHERE policy_id = $1 AND is_current = TRUE`, policyID)

	// Insert new version
	versionID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO policy_versions (id, org_id, policy_id, version_number, is_current,
			content, content_format, content_summary, change_summary, change_type,
			word_count, character_count, created_by)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6, $7, $8, $9, $10, $11, $12)
	`, versionID, orgID, policyID, newVersion, content, contentFormat,
		req.ContentSummary, req.ChangeSummary, changeType, wordCount, charCount, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert policy version")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create version"))
		return
	}

	// Update policy current_version_id
	tx.Exec(`UPDATE policies SET current_version_id = $1 WHERE id = $2`, versionID, policyID)

	// If policy was approved or published, revert to draft (content changed)
	if currentStatus == models.PolicyStatusApproved || currentStatus == models.PolicyStatusPublished {
		tx.Exec(`UPDATE policies SET status = 'draft' WHERE id = $1`, policyID)
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit version creation")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create version"))
		return
	}

	middleware.LogAudit(c, "policy_version.created", "policy_version", &versionID, map[string]interface{}{
		"policy_id": policyID, "version_number": newVersion,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":              versionID,
		"policy_id":       policyID,
		"version_number":  newVersion,
		"is_current":      true,
		"content_format":  contentFormat,
		"change_summary":  req.ChangeSummary,
		"change_type":     changeType,
		"word_count":      wordCount,
		"created_at":      time.Now(),
	}))
}

// CompareVersions compares two versions side-by-side.
func CompareVersions(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	policyID := c.Param("id")

	v1Str := c.Query("v1")
	v2Str := c.Query("v2")
	if v1Str == "" || v2Str == "" {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Both v1 and v2 query parameters are required"))
		return
	}

	v1, err1 := strconv.Atoi(v1Str)
	v2, err2 := strconv.Atoi(v2Str)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid version numbers"))
		return
	}
	if v1 == v2 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Cannot compare a version to itself"))
		return
	}

	// Verify policy exists
	var policyIdentifier string
	err := database.DB.QueryRow(`SELECT identifier FROM policies WHERE id = $1 AND org_id = $2`, policyID, orgID).Scan(&policyIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Policy not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to compare versions"))
		return
	}

	type versionData struct {
		VersionNumber  int
		Content        string
		ContentFormat  string
		ContentSummary *string
		ChangeSummary  *string
		ChangeType     string
		WordCount      *int
		CreatedByID    *string
		CreatedByName  string
		CreatedAt      time.Time
	}

	fetchVersion := func(vNum int) (*versionData, error) {
		var vd versionData
		var createdByID *string
		err := database.DB.QueryRow(`
			SELECT pv.version_number, pv.content, pv.content_format,
				pv.content_summary, pv.change_summary, pv.change_type,
				pv.word_count, pv.created_by, pv.created_at
			FROM policy_versions pv
			WHERE pv.policy_id = $1 AND pv.org_id = $2 AND pv.version_number = $3
		`, policyID, orgID, vNum).Scan(
			&vd.VersionNumber, &vd.Content, &vd.ContentFormat,
			&vd.ContentSummary, &vd.ChangeSummary, &vd.ChangeType,
			&vd.WordCount, &createdByID, &vd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		vd.CreatedByID = createdByID
		if createdByID != nil {
			var first, last string
			database.DB.QueryRow(`SELECT first_name, last_name FROM users WHERE id = $1`, *createdByID).Scan(&first, &last)
			vd.CreatedByName = first + " " + last
		}
		return &vd, nil
	}

	ver1, err := fetchVersion(v1)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", fmt.Sprintf("Version %d not found", v1)))
		return
	}
	ver2, err := fetchVersion(v2)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", fmt.Sprintf("Version %d not found", v2)))
		return
	}

	buildVersionJSON := func(vd *versionData) gin.H {
		v := gin.H{
			"version_number":  vd.VersionNumber,
			"content":         vd.Content,
			"content_format":  vd.ContentFormat,
			"content_summary": vd.ContentSummary,
			"change_summary":  vd.ChangeSummary,
			"change_type":     vd.ChangeType,
			"word_count":      vd.WordCount,
			"created_at":      vd.CreatedAt,
		}
		if vd.CreatedByID != nil {
			v["created_by"] = gin.H{"id": *vd.CreatedByID, "name": vd.CreatedByName}
		}
		return v
	}

	var wordDelta int
	if ver1.WordCount != nil && ver2.WordCount != nil {
		wordDelta = *ver2.WordCount - *ver1.WordCount
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"policy_id":         policyID,
		"policy_identifier": policyIdentifier,
		"versions": []gin.H{
			buildVersionJSON(ver1),
			buildVersionJSON(ver2),
		},
		"word_count_delta": wordDelta,
	}))
}

// SearchPolicies provides advanced search across policies and content.
func SearchPolicies(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Query parameter 'q' is required"))
		return
	}

	scope := c.DefaultQuery("scope", "metadata")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	where := []string{"p.org_id = $1", "p.is_template = FALSE"}
	args := []interface{}{orgID}
	argN := 2

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

	var searchCondition string
	switch scope {
	case "content", "all":
		searchCondition = fmt.Sprintf(`(
			to_tsvector('english', p.title || ' ' || COALESCE(p.description, '')) @@ plainto_tsquery('english', $%d) OR
			EXISTS (
				SELECT 1 FROM policy_versions pv2
				WHERE pv2.id = p.current_version_id
				AND to_tsvector('english', pv2.content) @@ plainto_tsquery('english', $%d)
			)
		)`, argN, argN)
	default: // "metadata"
		searchCondition = fmt.Sprintf(`to_tsvector('english', p.title || ' ' || COALESCE(p.description, '')) @@ plainto_tsquery('english', $%d)`, argN)
	}
	where = append(where, searchCondition)
	args = append(args, q)
	argN++

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM policies p WHERE %s`, whereClause)
	database.DB.QueryRow(countQuery, args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT p.id, p.identifier, p.title, p.description, p.category, p.status,
			p.owner_id, u.first_name, u.last_name,
			pv.version_number
		FROM policies p
		LEFT JOIN users u ON u.id = p.owner_id
		LEFT JOIN policy_versions pv ON pv.id = p.current_version_id
		WHERE %s
		ORDER BY p.identifier
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search policies")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to search policies"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, identifier, title, category, status string
			description                             *string
			ownerID                                 *string
			uFirst, uLast                           *string
			cvVersionNum                            *int
		)
		if err := rows.Scan(&id, &identifier, &title, &description, &category, &status, &ownerID, &uFirst, &uLast, &cvVersionNum); err != nil {
			continue
		}

		r := gin.H{
			"id":                     id,
			"identifier":             identifier,
			"title":                  title,
			"description":            description,
			"category":               category,
			"status":                 status,
			"current_version_number": cvVersionNum,
			"match_source":           "title",
		}

		if ownerID != nil && uFirst != nil {
			r["owner"] = gin.H{"id": *ownerID, "name": *uFirst + " " + *uLast}
		}

		// Try to get match context from content if scope includes content
		if scope == "content" || scope == "all" {
			var contentSnippet string
			database.DB.QueryRow(`
				SELECT ts_headline('english', pv.content, plainto_tsquery('english', $3),
					'MaxWords=25, MinWords=10, StartSel=, StopSel=')
				FROM policy_versions pv
				WHERE pv.id = (SELECT current_version_id FROM policies WHERE id = $1 AND org_id = $2)
				AND to_tsvector('english', pv.content) @@ plainto_tsquery('english', $3)
			`, id, orgID, q).Scan(&contentSnippet)
			if contentSnippet != "" {
				r["match_context"] = contentSnippet
				r["match_source"] = "content"
			}
		}

		results = append(results, r)
	}

	c.JSON(http.StatusOK, listResponse(c, results, total, page, perPage))
}
