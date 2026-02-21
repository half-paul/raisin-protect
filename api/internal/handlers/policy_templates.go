package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// ListPolicyTemplates lists available policy templates.
func ListPolicyTemplates(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	where := "p.org_id = $1 AND p.is_template = TRUE"
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("framework_id"); v != "" {
		where += " AND p.template_framework_id = $" + itoa(argN)
		args = append(args, v)
		argN++
	}
	if v := c.Query("category"); v != "" {
		where += " AND p.category = $" + itoa(argN)
		args = append(args, v)
		argN++
	}
	if v := c.Query("search"); v != "" {
		where += " AND to_tsvector('english', p.title || ' ' || COALESCE(p.description, '')) @@ plainto_tsquery('english', $" + itoa(argN) + ")"
		args = append(args, v)
		argN++
	}

	rows, err := database.DB.Query(`
		SELECT p.id, p.identifier, p.title, p.description, p.category,
			p.template_framework_id, f.identifier, f.name,
			pv.id, pv.version_number, pv.word_count, pv.content_summary,
			p.review_frequency_days, p.tags
		FROM policies p
		LEFT JOIN frameworks f ON f.id = p.template_framework_id
		LEFT JOIN policy_versions pv ON pv.id = p.current_version_id
		WHERE `+where+`
		ORDER BY p.identifier
	`, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list policy templates")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list templates"))
		return
	}
	defer rows.Close()

	templates := []gin.H{}
	for rows.Next() {
		var (
			id, identifier, title, category string
			description                     *string
			fwID, fwIdentifier, fwName      *string
			pvID                            *string
			pvVersionNum                    *int
			pvWordCount                     *int
			pvSummary                       *string
			reviewFreqDays                  *int
			tags                            pq.StringArray
		)
		if err := rows.Scan(
			&id, &identifier, &title, &description, &category,
			&fwID, &fwIdentifier, &fwName,
			&pvID, &pvVersionNum, &pvWordCount, &pvSummary,
			&reviewFreqDays, &tags,
		); err != nil {
			continue
		}

		t := gin.H{
			"id":                    id,
			"identifier":            identifier,
			"title":                 title,
			"description":           description,
			"category":              category,
			"review_frequency_days": reviewFreqDays,
			"tags":                  []string(tags),
		}

		if fwID != nil {
			t["framework"] = gin.H{"id": *fwID, "identifier": *fwIdentifier, "name": *fwName}
		} else {
			t["framework"] = nil
		}

		if pvID != nil {
			t["current_version"] = gin.H{
				"id":              *pvID,
				"version_number":  pvVersionNum,
				"word_count":      pvWordCount,
				"content_summary": pvSummary,
			}
		}

		templates = append(templates, t)
	}

	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	c.JSON(http.StatusOK, gin.H{
		"data": templates,
		"meta": gin.H{
			"total":      len(templates),
			"request_id": reqID,
		},
	})
}

// ClonePolicyTemplate clones a template into a new policy.
func ClonePolicyTemplate(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	templateID := c.Param("id")

	var req models.CloneTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	// Fetch template
	var (
		tTitle, tCategory      string
		tDescription           *string
		tReviewFreqDays        *int
		tTags                  pq.StringArray
		tCurrentVersionID      *string
		tIdentifier            string
	)
	err := database.DB.QueryRow(`
		SELECT identifier, title, description, category, review_frequency_days, tags, current_version_id
		FROM policies
		WHERE id = $1 AND org_id = $2 AND is_template = TRUE
	`, templateID, orgID).Scan(&tIdentifier, &tTitle, &tDescription, &tCategory, &tReviewFreqDays, &tTags, &tCurrentVersionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Template not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get template")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to clone template"))
		return
	}

	// Check identifier uniqueness
	var exists bool
	database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policies WHERE org_id = $1 AND identifier = $2)`, orgID, req.Identifier).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, errorResponse("DUPLICATE_IDENTIFIER", "Policy identifier already exists"))
		return
	}

	// Use template defaults for unspecified fields
	title := tTitle
	if req.Title != nil {
		title = *req.Title
	}
	description := tDescription
	if req.Description != nil {
		description = req.Description
	}
	reviewFreqDays := tReviewFreqDays
	if req.ReviewFreqDays != nil {
		reviewFreqDays = req.ReviewFreqDays
	}
	ownerID := userID
	if req.OwnerID != nil && *req.OwnerID != "" {
		ownerID = *req.OwnerID
	}

	// Filter out 'template' from tags
	tags := []string{}
	if req.Tags != nil {
		tags = req.Tags
	} else {
		for _, t := range tTags {
			if t != "template" {
				tags = append(tags, t)
			}
		}
	}

	// Fetch template version content before starting the transaction
	var content, contentFormat string
	var wordCount *int
	if tCurrentVersionID != nil {
		database.DB.QueryRow(`
			SELECT content, content_format, word_count FROM policy_versions WHERE id = $1
		`, *tCurrentVersionID).Scan(&content, &contentFormat, &wordCount)
	}
	if content == "" {
		content = "<p>Policy content — cloned from template.</p>"
		contentFormat = "html"
	}

	policyID := uuid.New().String()
	versionID := uuid.New().String()

	tx, err := database.DB.Begin()
	if err != nil {
		log.Error().Err(err).Str("step", "begin_tx").Msg("Failed to clone template")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to clone template"))
		return
	}
	defer tx.Rollback()

	// Create new policy
	_, err = tx.Exec(`
		INSERT INTO policies (id, org_id, identifier, title, description, category, status,
			owner_id, review_frequency_days, is_template, cloned_from_policy_id, tags, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, 'draft', $7, $8, FALSE, $9, $10, '{}')
	`, policyID, orgID, req.Identifier, title, description, tCategory,
		ownerID, reviewFreqDays, templateID, pq.Array(tags))
	if err != nil {
		log.Error().Err(err).Str("step", "insert_policy").Msg("Failed to insert cloned policy")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to clone template"))
		return
	}

	charCount := utf8.RuneCountInString(content)
	changeSummary := "Cloned from template: " + tIdentifier

	_, err = tx.Exec(`
		INSERT INTO policy_versions (id, org_id, policy_id, version_number, is_current,
			content, content_format, content_summary, change_summary, change_type,
			word_count, character_count, created_by)
		VALUES ($1, $2, $3, 1, TRUE, $4, $5, $6, $7, 'initial', $8, $9, $10)
	`, versionID, orgID, policyID, content, contentFormat, description, changeSummary, wordCount, charCount, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert cloned version")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to clone template"))
		return
	}

	// Set current version
	tx.Exec(`UPDATE policies SET current_version_id = $1 WHERE id = $2`, versionID, policyID)

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit clone")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to clone template"))
		return
	}

	middleware.LogAudit(c, "policy.created", "policy", &policyID, map[string]interface{}{
		"identifier": req.Identifier, "cloned_from": templateID,
	})
	middleware.LogAudit(c, "policy.cloned_from_template", "policy", &policyID, map[string]interface{}{
		"template_id": templateID, "template_identifier": tIdentifier,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":                    policyID,
		"identifier":            req.Identifier,
		"title":                 title,
		"status":                "draft",
		"cloned_from_policy_id": templateID,
		"current_version": gin.H{
			"id":             versionID,
			"version_number": 1,
			"word_count":     wordCount,
		},
		"created_at": time.Now(),
	}))
}

// itoa converts int to string (simple helper for query building).
func itoa(n int) string {
	return strconv.Itoa(n)
}
