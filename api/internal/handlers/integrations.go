package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// ListIntegrations returns the system-level integration catalog.
func ListIntegrations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	category := c.Query("category")
	capability := c.Query("capability")
	search := c.Query("search")

	// Build count query
	countQuery := "SELECT COUNT(*) FROM integration_definitions WHERE is_active = true"
	countArgs := []interface{}{}
	argN := 1

	if category != "" {
		countQuery += " AND category = $" + strconv.Itoa(argN)
		countArgs = append(countArgs, category)
		argN++
	}
	if capability != "" {
		countQuery += " AND $" + strconv.Itoa(argN) + " = ANY(capabilities)"
		countArgs = append(countArgs, capability)
		argN++
	}
	if search != "" {
		countQuery += " AND (name ILIKE $" + strconv.Itoa(argN) + " OR description ILIKE $" + strconv.Itoa(argN) + ")"
		countArgs = append(countArgs, "%"+search+"%")
		argN++
	}

	var total int
	err := database.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to count integrations"))
		return
	}

	// Build data query
	dataQuery := "SELECT id, name, slug, provider, category, description, short_description, icon_url, documentation_url, website_url, auth_type, config_schema, capabilities, is_active, is_beta, version, tags, created_at, updated_at FROM integration_definitions WHERE is_active = true"
	dataArgs := []interface{}{}
	argN = 1

	if category != "" {
		dataQuery += " AND category = $" + strconv.Itoa(argN)
		dataArgs = append(dataArgs, category)
		argN++
	}
	if capability != "" {
		dataQuery += " AND $" + strconv.Itoa(argN) + " = ANY(capabilities)"
		dataArgs = append(dataArgs, capability)
		argN++
	}
	if search != "" {
		dataQuery += " AND (name ILIKE $" + strconv.Itoa(argN) + " OR description ILIKE $" + strconv.Itoa(argN) + ")"
		dataArgs = append(dataArgs, "%"+search+"%")
		argN++
	}

	dataQuery += " ORDER BY name ASC LIMIT $" + strconv.Itoa(argN) + " OFFSET $" + strconv.Itoa(argN+1)
	dataArgs = append(dataArgs, perPage, offset)

	rows, err := database.Query(dataQuery, dataArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to query integrations"))
		return
	}
	defer rows.Close()

	items := []gin.H{}
	for rows.Next() {
		var id, name, slug, provider, cat, authType, version string
		var description, shortDesc, iconURL, docURL, webURL *string
		var configSchemaJSON []byte
		var capabilities, tags []string
		var isActive, isBeta bool
		var createdAt, updatedAt interface{}

		err := rows.Scan(&id, &name, &slug, &provider, &cat, &description, &shortDesc, &iconURL, &docURL, &webURL, &authType, &configSchemaJSON, pq.Array(&capabilities), &isActive, &isBeta, &version, pq.Array(&tags), &createdAt, &updatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Failed to scan integration"))
			return
		}

		var configSchema map[string]interface{}
		json.Unmarshal(configSchemaJSON, &configSchema)

		items = append(items, gin.H{
			"id":                id,
			"name":              name,
			"slug":              slug,
			"provider":          provider,
			"category":          cat,
			"description":       description,
			"short_description": shortDesc,
			"icon_url":          iconURL,
			"documentation_url": docURL,
			"website_url":       webURL,
			"auth_type":         authType,
			"config_schema":     configSchema,
			"capabilities":      capabilities,
			"is_active":         isActive,
			"is_beta":           isBeta,
			"version":           version,
			"tags":              tags,
			"created_at":        createdAt,
			"updated_at":        updatedAt,
		})
	}

	c.JSON(http.StatusOK, listResponse(c, items, total, page, perPage))
}

// GetIntegration returns a single integration definition with its config schema.
func GetIntegration(c *gin.Context) {
	id := c.Param("id")

	var name, slug, provider, cat, authType, version string
	var description, shortDesc, iconURL, docURL, webURL *string
	var configSchemaJSON []byte
	var capabilities, tags []string
	var isActive, isBeta bool
	var createdAt, updatedAt interface{}

	err := database.QueryRow(`
		SELECT id, name, slug, provider, category, description, short_description, icon_url, documentation_url, website_url, auth_type, config_schema, capabilities, is_active, is_beta, version, tags, created_at, updated_at
		FROM integration_definitions WHERE id = $1
	`, id).Scan(&id, &name, &slug, &provider, &cat, &description, &shortDesc, &iconURL, &docURL, &webURL, &authType, &configSchemaJSON, pq.Array(&capabilities), &isActive, &isBeta, &version, pq.Array(&tags), &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Integration not found"))
		return
	}

	var configSchema map[string]interface{}
	json.Unmarshal(configSchemaJSON, &configSchema)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":                id,
		"name":              name,
		"slug":              slug,
		"provider":          provider,
		"category":          cat,
		"description":       description,
		"short_description": shortDesc,
		"icon_url":          iconURL,
		"documentation_url": docURL,
		"website_url":       webURL,
		"auth_type":         authType,
		"config_schema":     configSchema,
		"capabilities":      capabilities,
		"is_active":         isActive,
		"is_beta":           isBeta,
		"version":           version,
		"tags":              tags,
		"created_at":        createdAt,
		"updated_at":        updatedAt,
	}))
}
