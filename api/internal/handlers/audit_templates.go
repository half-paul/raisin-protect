package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// ListAuditRequestTemplates lists available PBC request templates.
func ListAuditRequestTemplates(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	_ = orgID // templates are global, but auth is still required

	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	if v := c.Query("audit_type"); v != "" {
		where = append(where, fmt.Sprintf("audit_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("framework"); v != "" {
		where = append(where, fmt.Sprintf("framework = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argN, argN))
		args = append(args, "%"+v+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, title, description, audit_type, framework, category, default_priority, tags FROM audit_request_templates WHERE %s ORDER BY framework, category, title", whereClause),
		args...,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list audit request templates")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list templates"))
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var (
			id, title, description, auditType, framework, category, defaultPriority string
			tags                                                                     pq.StringArray
		)
		if err := rows.Scan(&id, &title, &description, &auditType, &framework, &category, &defaultPriority, &tags); err != nil {
			log.Error().Err(err).Msg("Failed to scan template row")
			continue
		}
		results = append(results, gin.H{
			"id": id, "title": title, "description": description,
			"audit_type": auditType, "framework": framework,
			"category": category, "default_priority": defaultPriority,
			"tags": []string(tags),
		})
	}

	c.JSON(http.StatusOK, successResponse(c, results))
}

// CreateFromTemplate creates requests from one or more templates.
func CreateFromTemplate(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	auditID := c.Param("id")

	auditStatus, ok := checkAuditAccess(c, auditID, orgID, userID, userRole)
	if !ok {
		return
	}
	if !checkAuditNotTerminal(c, auditStatus) {
		return
	}

	var req models.CreateFromTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request: "+err.Error()))
		return
	}

	if len(req.TemplateIDs) == 0 || len(req.TemplateIDs) > 100 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Must provide 1-100 template IDs"))
		return
	}

	autoNumber := true
	if req.AutoNumber != nil {
		autoNumber = *req.AutoNumber
	}
	prefix := "PBC"
	if req.NumberPrefix != nil {
		prefix = *req.NumberPrefix
	}

	now := time.Now()
	created := []gin.H{}

	// Get current highest PBC number for auto-numbering
	var maxNum int
	if autoNumber {
		database.DB.QueryRow(`
			SELECT COALESCE(MAX(
				CASE WHEN reference_number ~ ('^' || $1 || '-[0-9]+$')
				     THEN CAST(SUBSTRING(reference_number FROM '[0-9]+$') AS INTEGER)
				     ELSE 0
				END
			), 0)
			FROM audit_requests WHERE audit_id = $2 AND org_id = $3
		`, prefix, auditID, orgID).Scan(&maxNum)
	}

	for i, templateID := range req.TemplateIDs {
		var title, description, defaultPriority string
		var tags pq.StringArray
		err := database.DB.QueryRow(
			"SELECT title, description, default_priority, tags FROM audit_request_templates WHERE id = $1",
			templateID,
		).Scan(&title, &description, &defaultPriority, &tags)
		if err != nil {
			log.Error().Err(err).Str("template_id", templateID).Msg("Template not found, skipping")
			continue
		}

		reqID := uuid.New().String()
		var refNumber *string
		if autoNumber {
			num := maxNum + i + 1
			ref := fmt.Sprintf("%s-%03d", prefix, num)
			refNumber = &ref
		}

		_, err = database.DB.Exec(`
			INSERT INTO audit_requests (id, org_id, audit_id, title, description, priority, status,
			                            requested_by, due_date, reference_number, tags, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, $8::date, $9, $10, $11, $11)
		`, reqID, orgID, auditID, title, description, defaultPriority,
			userID, req.DefaultDueDate, refNumber, pq.Array(tags), now)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create request from template")
			continue
		}

		created = append(created, gin.H{
			"id": reqID, "title": title, "priority": defaultPriority,
			"reference_number": refNumber, "status": "open",
		})
	}

	updateAuditRequestCounts(auditID, orgID)

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"created":  len(created),
		"requests": created,
	}))
}
