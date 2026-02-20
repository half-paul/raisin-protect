package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/rs/zerolog/log"
)

// ListAuditLogs returns paginated audit log entries for the organization.
func ListAuditLogs(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	actionFilter := c.Query("action")
	actorIDFilter := c.Query("actor_id")
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	sortField := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	if sortField != "created_at" {
		sortField = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	where := []string{"al.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if actionFilter != "" {
		where = append(where, fmt.Sprintf("al.action = $%d", argN))
		args = append(args, actionFilter)
		argN++
	}
	if actorIDFilter != "" {
		where = append(where, fmt.Sprintf("al.actor_id = $%d", argN))
		args = append(args, actorIDFilter)
		argN++
	}
	if resourceType != "" {
		where = append(where, fmt.Sprintf("al.resource_type = $%d", argN))
		args = append(args, resourceType)
		argN++
	}
	if resourceID != "" {
		where = append(where, fmt.Sprintf("al.resource_id = $%d", argN))
		args = append(args, resourceID)
		argN++
	}
	if fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			where = append(where, fmt.Sprintf("al.created_at >= $%d", argN))
			args = append(args, t)
			argN++
		}
	}
	if toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			where = append(where, fmt.Sprintf("al.created_at <= $%d", argN))
			args = append(args, t)
			argN++
		}
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM audit_log al WHERE %s", whereClause), countArgs...).Scan(&total)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count audit logs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT al.id, al.actor_id, COALESCE(u.email, ''), COALESCE(u.first_name || ' ' || u.last_name, ''),
			   al.action, al.resource_type, al.resource_id, al.metadata::text,
			   al.ip_address, al.user_agent, al.created_at
		FROM audit_log al
		LEFT JOIN users u ON u.id = al.actor_id
		WHERE %s
		ORDER BY al.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query audit logs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	entries := []gin.H{}
	for rows.Next() {
		var (
			id, actorEmail, actorName, action, resType, metadataStr string
			actorID, resourceIDVal, ipAddress, userAgent            *string
			createdAt                                                time.Time
		)
		if err := rows.Scan(&id, &actorID, &actorEmail, &actorName, &action, &resType, &resourceIDVal,
			&metadataStr, &ipAddress, &userAgent, &createdAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan audit log row")
			continue
		}

		var metadata interface{}
		_ = json.Unmarshal([]byte(metadataStr), &metadata)

		entry := gin.H{
			"id":            id,
			"actor_id":      actorID,
			"actor_email":   actorEmail,
			"actor_name":    actorName,
			"action":        action,
			"resource_type": resType,
			"resource_id":   resourceIDVal,
			"metadata":      metadata,
			"ip_address":    ipAddress,
			"user_agent":    userAgent,
			"created_at":    createdAt,
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, listResponse(c, entries, total, page, perPage))
}
