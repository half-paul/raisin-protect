package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

var identifierRegex = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)

// ListControls lists controls in the org's library with filtering and pagination.
func ListControls(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	statusFilter := c.Query("status")
	categoryFilter := c.Query("category")
	ownerIDFilter := c.Query("owner_id")
	isCustomFilter := c.Query("is_custom")
	frameworkIDFilter := c.Query("framework_id")
	unmappedFilter := c.Query("unmapped")
	search := c.Query("search")
	sortField := c.DefaultQuery("sort", "identifier")
	order := c.DefaultQuery("order", "asc")

	allowedSort := map[string]string{
		"identifier": "c.identifier",
		"title":      "c.title",
		"category":   "c.category",
		"status":     "c.status",
		"created_at": "c.created_at",
		"updated_at": "c.updated_at",
	}
	sortCol, ok := allowedSort[sortField]
	if !ok {
		sortCol = "c.identifier"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	where := []string{"c.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if statusFilter != "" {
		where = append(where, fmt.Sprintf("c.status = $%d", argN))
		args = append(args, statusFilter)
		argN++
	}
	if categoryFilter != "" {
		where = append(where, fmt.Sprintf("c.category = $%d", argN))
		args = append(args, categoryFilter)
		argN++
	}
	if ownerIDFilter != "" {
		where = append(where, fmt.Sprintf("c.owner_id = $%d", argN))
		args = append(args, ownerIDFilter)
		argN++
	}
	if isCustomFilter == "true" {
		where = append(where, "c.is_custom = TRUE")
	} else if isCustomFilter == "false" {
		where = append(where, "c.is_custom = FALSE")
	}
	if frameworkIDFilter != "" {
		where = append(where, fmt.Sprintf(`c.id IN (
			SELECT cm.control_id FROM control_mappings cm
			JOIN requirements r ON r.id = cm.requirement_id
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			WHERE cm.org_id = $1 AND fv.framework_id = $%d
		)`, argN))
		args = append(args, frameworkIDFilter)
		argN++
	}
	if unmappedFilter == "true" {
		where = append(where, `NOT EXISTS (
			SELECT 1 FROM control_mappings cm WHERE cm.control_id = c.id AND cm.org_id = c.org_id
		)`)
	}
	if search != "" {
		where = append(where, fmt.Sprintf(
			"to_tsvector('english', c.title || ' ' || COALESCE(c.description, '')) @@ plainto_tsquery('english', $%d)", argN))
		args = append(args, search)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM controls c WHERE %s", whereClause), countArgs...).Scan(&total)

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT c.id, c.identifier, c.title, c.description, c.category, c.status, c.is_custom,
			   c.owner_id, COALESCE(u.first_name || ' ' || u.last_name, ''), COALESCE(u.email, ''),
			   c.secondary_owner_id,
			   COALESCE((SELECT COUNT(*) FROM control_mappings cm WHERE cm.control_id = c.id), 0) AS mappings_count,
			   c.created_at, c.updated_at
		FROM controls c
		LEFT JOIN users u ON u.id = c.owner_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list controls")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	controls := []gin.H{}
	for rows.Next() {
		var (
			cID, cIdentifier, cTitle, cDescription, cCategory, cStatus string
			cIsCustom                                                   bool
			ownerID, secondaryOwnerID                                  *string
			ownerName, ownerEmail                                      string
			mappingsCount                                              int
			createdAt, updatedAt                                       interface{}
		)
		if err := rows.Scan(&cID, &cIdentifier, &cTitle, &cDescription, &cCategory, &cStatus,
			&cIsCustom, &ownerID, &ownerName, &ownerEmail, &secondaryOwnerID,
			&mappingsCount, &createdAt, &updatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan control row")
			continue
		}

		ctrl := gin.H{
			"id":             cID,
			"identifier":     cIdentifier,
			"title":          cTitle,
			"description":    cDescription,
			"category":       cCategory,
			"status":         cStatus,
			"is_custom":      cIsCustom,
			"mappings_count": mappingsCount,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}

		if ownerID != nil {
			ctrl["owner"] = gin.H{"id": *ownerID, "name": ownerName, "email": ownerEmail}
		} else {
			ctrl["owner"] = nil
		}
		ctrl["secondary_owner"] = nil

		// Get framework names for this control
		fwRows, _ := database.Query(`
			SELECT DISTINCT f.name FROM control_mappings cm
			JOIN requirements r ON r.id = cm.requirement_id
			JOIN framework_versions fv ON fv.id = r.framework_version_id
			JOIN frameworks f ON f.id = fv.framework_id
			WHERE cm.control_id = $1
		`, cID)
		if fwRows != nil {
			fwNames := []string{}
			for fwRows.Next() {
				var name string
				fwRows.Scan(&name)
				fwNames = append(fwNames, name)
			}
			fwRows.Close()
			ctrl["frameworks"] = fwNames
		}

		controls = append(controls, ctrl)
	}

	c.JSON(http.StatusOK, listResponse(c, controls, total, page, perPage))
}

// CreateControl creates a new control in the org's library.
func CreateControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.CreateControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate
	if len(req.Identifier) > 50 || !identifierRegex.MatchString(req.Identifier) {
		c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
			{"field": "identifier", "message": "Must be alphanumeric with hyphens, max 50 chars"},
		}))
		return
	}
	if len(req.Title) > 500 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must be at most 500 characters"))
		return
	}
	if len(req.Description) > 10000 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Description must be at most 10000 characters"))
		return
	}
	if !models.IsValidControlCategory(req.Category) {
		c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
			{"field": "category", "message": "Must be: technical, administrative, physical, operational"},
		}))
		return
	}

	status := "draft"
	if req.Status != nil {
		if !models.IsValidControlStatus(*req.Status) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid control status"))
			return
		}
		status = *req.Status
	}

	// Check identifier uniqueness within org
	var exists bool
	database.QueryRow("SELECT EXISTS(SELECT 1 FROM controls WHERE org_id = $1 AND identifier = $2)", orgID, req.Identifier).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Identifier already exists in this organization"))
		return
	}

	// Validate owners
	if req.OwnerID != nil {
		var ownerExists bool
		database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2)", *req.OwnerID, orgID).Scan(&ownerExists)
		if !ownerExists {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Owner not found in organization"))
			return
		}
	}
	if req.SecondaryOwnerID != nil {
		if req.OwnerID != nil && *req.OwnerID == *req.SecondaryOwnerID {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Secondary owner must be different from primary owner"))
			return
		}
		var soExists bool
		database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2)", *req.SecondaryOwnerID, orgID).Scan(&soExists)
		if !soExists {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Secondary owner not found in organization"))
			return
		}
	}

	metadataJSON := "{}"
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err == nil {
			metadataJSON = string(b)
		}
		if len(metadataJSON) > 10240 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Metadata must be at most 10KB"))
			return
		}
	}

	controlID := uuid.New().String()
	_, err := database.Exec(`
		INSERT INTO controls (id, org_id, identifier, title, description, implementation_guidance,
			category, status, owner_id, secondary_owner_id, evidence_requirements, test_criteria,
			is_custom, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, TRUE, $13)
	`, controlID, orgID, req.Identifier, req.Title, req.Description, req.ImplementationGuidance,
		req.Category, status, req.OwnerID, req.SecondaryOwnerID, req.EvidenceRequirements,
		req.TestCriteria, metadataJSON)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "control.created", "control", &controlID, map[string]interface{}{
		"identifier": req.Identifier, "category": req.Category,
	})

	resp := gin.H{
		"id":                      controlID,
		"identifier":              req.Identifier,
		"title":                   req.Title,
		"description":             req.Description,
		"implementation_guidance": req.ImplementationGuidance,
		"category":                req.Category,
		"status":                  status,
		"is_custom":               true,
		"evidence_requirements":   req.EvidenceRequirements,
		"test_criteria":           req.TestCriteria,
		"metadata":                req.Metadata,
		"mappings_count":          0,
	}

	if req.OwnerID != nil {
		var ownerName string
		database.QueryRow("SELECT first_name || ' ' || last_name FROM users WHERE id = $1", *req.OwnerID).Scan(&ownerName)
		resp["owner"] = gin.H{"id": *req.OwnerID, "name": ownerName}
	} else {
		resp["owner"] = nil
	}

	if req.SecondaryOwnerID != nil {
		var soName string
		database.QueryRow("SELECT first_name || ' ' || last_name FROM users WHERE id = $1", *req.SecondaryOwnerID).Scan(&soName)
		resp["secondary_owner"] = gin.H{"id": *req.SecondaryOwnerID, "name": soName}
	} else {
		resp["secondary_owner"] = nil
	}

	c.JSON(http.StatusCreated, successResponse(c, resp))
}

// GetControl returns a single control with full details and mappings.
func GetControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	var ctrl models.Control
	var ownerName, ownerEmail string
	var secondaryOwnerName *string
	err := database.QueryRow(`
		SELECT c.id, c.identifier, c.title, c.description, c.implementation_guidance,
			   c.category, c.status, c.is_custom, c.source_template_id,
			   c.owner_id, COALESCE(u.first_name || ' ' || u.last_name, ''), COALESCE(u.email, ''),
			   c.secondary_owner_id,
			   c.evidence_requirements, c.test_criteria, c.metadata::text,
			   c.created_at, c.updated_at
		FROM controls c
		LEFT JOIN users u ON u.id = c.owner_id
		WHERE c.id = $1 AND c.org_id = $2
	`, controlID, orgID).Scan(
		&ctrl.ID, &ctrl.Identifier, &ctrl.Title, &ctrl.Description, &ctrl.ImplementationGuidance,
		&ctrl.Category, &ctrl.Status, &ctrl.IsCustom, &ctrl.SourceTemplateID,
		&ctrl.OwnerID, &ownerName, &ownerEmail, &ctrl.SecondaryOwnerID,
		&ctrl.EvidenceRequirements, &ctrl.TestCriteria, &ctrl.Metadata,
		&ctrl.CreatedAt, &ctrl.UpdatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	var metadata interface{}
	json.Unmarshal([]byte(ctrl.Metadata), &metadata)

	resp := gin.H{
		"id":                      ctrl.ID,
		"identifier":              ctrl.Identifier,
		"title":                   ctrl.Title,
		"description":             ctrl.Description,
		"implementation_guidance": ctrl.ImplementationGuidance,
		"category":                ctrl.Category,
		"status":                  ctrl.Status,
		"is_custom":               ctrl.IsCustom,
		"source_template_id":      ctrl.SourceTemplateID,
		"evidence_requirements":   ctrl.EvidenceRequirements,
		"test_criteria":           ctrl.TestCriteria,
		"metadata":                metadata,
		"created_at":              ctrl.CreatedAt,
		"updated_at":              ctrl.UpdatedAt,
	}

	if ctrl.OwnerID != nil {
		resp["owner"] = gin.H{"id": *ctrl.OwnerID, "name": ownerName, "email": ownerEmail}
	} else {
		resp["owner"] = nil
	}

	if ctrl.SecondaryOwnerID != nil {
		database.QueryRow("SELECT first_name || ' ' || last_name FROM users WHERE id = $1", *ctrl.SecondaryOwnerID).Scan(&secondaryOwnerName)
		if secondaryOwnerName != nil {
			resp["secondary_owner"] = gin.H{"id": *ctrl.SecondaryOwnerID, "name": *secondaryOwnerName}
		}
	} else {
		resp["secondary_owner"] = nil
	}

	// Get mappings
	mappingRows, err := database.Query(`
		SELECT cm.id, r.id, r.identifier, r.title, f.name, fv.version, cm.strength, cm.notes
		FROM control_mappings cm
		JOIN requirements r ON r.id = cm.requirement_id
		JOIN framework_versions fv ON fv.id = r.framework_version_id
		JOIN frameworks f ON f.id = fv.framework_id
		WHERE cm.control_id = $1 AND cm.org_id = $2
		ORDER BY f.name, r.identifier
	`, controlID, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get control mappings")
		resp["mappings"] = []gin.H{}
		c.JSON(http.StatusOK, successResponse(c, resp))
		return
	}
	defer mappingRows.Close()

	mappings := []gin.H{}
	for mappingRows.Next() {
		var mID, rID, rIdentifier, rTitle, fName, fvVersion, strength string
		var notes *string
		if err := mappingRows.Scan(&mID, &rID, &rIdentifier, &rTitle, &fName, &fvVersion, &strength, &notes); err != nil {
			continue
		}
		mappings = append(mappings, gin.H{
			"id": mID,
			"requirement": gin.H{
				"id":                rID,
				"identifier":        rIdentifier,
				"title":             rTitle,
				"framework":         fName,
				"framework_version": fvVersion,
			},
			"strength": strength,
			"notes":    notes,
		})
	}
	resp["mappings"] = mappings

	c.JSON(http.StatusOK, successResponse(c, resp))
}

// UpdateControl updates a control's details.
func UpdateControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	callerID := middleware.GetUserID(c)
	callerRole := middleware.GetUserRole(c)
	controlID := c.Param("id")

	// Get current control
	var currentOwnerID *string
	err := database.QueryRow(`
		SELECT owner_id FROM controls WHERE id = $1 AND org_id = $2
	`, controlID, orgID).Scan(&currentOwnerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Check authorization: admin roles or owner
	isAdmin := models.HasRole(callerRole, models.ControlCreateRoles)
	isOwner := currentOwnerID != nil && *currentOwnerID == callerID
	if !isAdmin && !isOwner {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to update this control"))
		return
	}

	var req models.UpdateControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	changes := map[string]interface{}{}

	if req.Title != nil {
		if len(*req.Title) > 500 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Title must be at most 500 characters"))
			return
		}
		database.Exec("UPDATE controls SET title = $1 WHERE id = $2", *req.Title, controlID)
		changes["title"] = *req.Title
	}
	if req.Description != nil {
		if len(*req.Description) > 10000 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Description must be at most 10000 characters"))
			return
		}
		database.Exec("UPDATE controls SET description = $1 WHERE id = $2", *req.Description, controlID)
		changes["description"] = "updated"
	}
	if req.ImplementationGuidance != nil {
		database.Exec("UPDATE controls SET implementation_guidance = $1 WHERE id = $2", *req.ImplementationGuidance, controlID)
		changes["implementation_guidance"] = "updated"
	}
	if req.Category != nil {
		if !models.IsValidControlCategory(*req.Category) {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid control category"))
			return
		}
		database.Exec("UPDATE controls SET category = $1 WHERE id = $2", *req.Category, controlID)
		changes["category"] = *req.Category
	}
	if req.EvidenceRequirements != nil {
		database.Exec("UPDATE controls SET evidence_requirements = $1 WHERE id = $2", *req.EvidenceRequirements, controlID)
	}
	if req.TestCriteria != nil {
		database.Exec("UPDATE controls SET test_criteria = $1 WHERE id = $2", *req.TestCriteria, controlID)
	}
	if req.Metadata != nil {
		// Merge metadata
		var existingMeta string
		database.QueryRow("SELECT metadata::text FROM controls WHERE id = $1", controlID).Scan(&existingMeta)
		existing := map[string]interface{}{}
		json.Unmarshal([]byte(existingMeta), &existing)
		for k, v := range req.Metadata {
			if v == nil {
				delete(existing, k)
			} else {
				existing[k] = v
			}
		}
		merged, _ := json.Marshal(existing)
		database.Exec("UPDATE controls SET metadata = $1 WHERE id = $2", string(merged), controlID)
		changes["metadata"] = "merged"
	}

	if len(changes) > 0 {
		middleware.LogAudit(c, "control.updated", "control", &controlID, changes)
	}

	// Return updated control
	GetControl(c)
}

// ChangeControlOwner changes a control's ownership.
func ChangeControlOwner(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	var currentOwnerID *string
	var cIdentifier string
	err := database.QueryRow(`
		SELECT owner_id, identifier FROM controls WHERE id = $1 AND org_id = $2
	`, controlID, orgID).Scan(&currentOwnerID, &cIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	var req models.ChangeOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if req.OwnerID == nil && req.SecondaryOwnerID == nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "At least one of owner_id or secondary_owner_id is required"))
		return
	}

	if req.OwnerID != nil {
		var exists bool
		database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2)", *req.OwnerID, orgID).Scan(&exists)
		if !exists {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Owner not found in organization"))
			return
		}
		database.Exec("UPDATE controls SET owner_id = $1 WHERE id = $2", *req.OwnerID, controlID)
	}

	if req.SecondaryOwnerID != nil {
		effectiveOwner := req.OwnerID
		if effectiveOwner == nil {
			effectiveOwner = currentOwnerID
		}
		if effectiveOwner != nil && *effectiveOwner == *req.SecondaryOwnerID {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Secondary owner must be different from primary owner"))
			return
		}
		database.Exec("UPDATE controls SET secondary_owner_id = $1 WHERE id = $2", *req.SecondaryOwnerID, controlID)
	}

	var prevOwner *string
	if currentOwnerID != nil {
		prevOwner = currentOwnerID
	}
	middleware.LogAudit(c, "control.owner_changed", "control", &controlID, map[string]interface{}{
		"previous_owner": prevOwner, "new_owner": req.OwnerID,
	})

	// Get updated owner info
	var ownerName string
	newOwnerID := req.OwnerID
	if newOwnerID == nil {
		newOwnerID = currentOwnerID
	}
	if newOwnerID != nil {
		database.QueryRow("SELECT first_name || ' ' || last_name FROM users WHERE id = $1", *newOwnerID).Scan(&ownerName)
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         controlID,
		"identifier": cIdentifier,
		"owner": gin.H{
			"id":   newOwnerID,
			"name": ownerName,
		},
		"secondary_owner": nil,
		"message":         "Ownership updated",
	}))
}

// ChangeControlStatus changes a control's lifecycle status.
func ChangeControlStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	var currentStatus, cIdentifier string
	err := database.QueryRow(`
		SELECT status, identifier FROM controls WHERE id = $1 AND org_id = $2
	`, controlID, orgID).Scan(&currentStatus, &cIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	var req models.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if !models.IsValidControlStatus(req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid control status"))
		return
	}

	if !models.IsValidStatusTransition(currentStatus, req.Status) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE",
			fmt.Sprintf("Invalid status transition from '%s' to '%s'", currentStatus, req.Status)))
		return
	}

	_, err = database.Exec("UPDATE controls SET status = $1 WHERE id = $2", req.Status, controlID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to change control status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "control.status_changed", "control", &controlID, map[string]interface{}{
		"old_status": currentStatus, "new_status": req.Status,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":              controlID,
		"identifier":      cIdentifier,
		"status":          req.Status,
		"previous_status": currentStatus,
		"message":         "Status updated",
	}))
}

// DeprecateControl soft-deletes a control by setting status to deprecated.
func DeprecateControl(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	controlID := c.Param("id")

	var cIdentifier string
	err := database.QueryRow(`
		SELECT identifier FROM controls WHERE id = $1 AND org_id = $2
	`, controlID, orgID).Scan(&cIdentifier)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Control not found"))
		return
	}

	_, err = database.Exec("UPDATE controls SET status = 'deprecated' WHERE id = $1", controlID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to deprecate control")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "control.deprecated", "control", &controlID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":         controlID,
		"identifier": cIdentifier,
		"status":     "deprecated",
		"message":    "Control deprecated. Mappings preserved for audit trail.",
	}))
}

// BulkControlStatus changes status for multiple controls.
func BulkControlStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.BulkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if len(req.ControlIDs) == 0 || len(req.ControlIDs) > 100 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "control_ids must have 1-100 entries"))
		return
	}
	if !models.IsValidControlStatus(req.Status) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid control status"))
		return
	}

	results := []gin.H{}
	updated, failed := 0, 0

	for _, ctrlID := range req.ControlIDs {
		var currentStatus, identifier string
		err := database.QueryRow(`
			SELECT status, identifier FROM controls WHERE id = $1 AND org_id = $2
		`, ctrlID, orgID).Scan(&currentStatus, &identifier)
		if err != nil {
			results = append(results, gin.H{"id": ctrlID, "success": false, "error": "not found"})
			failed++
			continue
		}

		if !models.IsValidStatusTransition(currentStatus, req.Status) {
			results = append(results, gin.H{
				"id": ctrlID, "identifier": identifier, "success": false,
				"error": fmt.Sprintf("Invalid transition from '%s' to '%s'", currentStatus, req.Status),
			})
			failed++
			continue
		}

		_, err = database.Exec("UPDATE controls SET status = $1 WHERE id = $2", req.Status, ctrlID)
		if err != nil {
			results = append(results, gin.H{"id": ctrlID, "identifier": identifier, "success": false, "error": "update failed"})
			failed++
			continue
		}

		middleware.LogAudit(c, "control.status_changed", "control", &ctrlID, map[string]interface{}{
			"old_status": currentStatus, "new_status": req.Status,
		})

		results = append(results, gin.H{"id": ctrlID, "identifier": identifier, "status": req.Status, "success": true})
		updated++
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"updated": updated,
		"failed":  failed,
		"results": results,
	}))
}

// GetControlStats returns aggregate statistics about the org's control library.
func GetControlStats(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var total int
	database.QueryRow("SELECT COUNT(*) FROM controls WHERE org_id = $1", orgID).Scan(&total)

	// By status
	byStatus := gin.H{"draft": 0, "active": 0, "under_review": 0, "deprecated": 0}
	rows, err := database.Query("SELECT status, COUNT(*) FROM controls WHERE org_id = $1 GROUP BY status", orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			rows.Scan(&status, &count)
			byStatus[status] = count
		}
	}

	// By category
	byCategory := gin.H{"technical": 0, "administrative": 0, "physical": 0, "operational": 0}
	rows2, err := database.Query("SELECT category, COUNT(*) FROM controls WHERE org_id = $1 GROUP BY category", orgID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var cat string
			var count int
			rows2.Scan(&cat, &count)
			byCategory[cat] = count
		}
	}

	var customCount, libraryCount int
	database.QueryRow("SELECT COUNT(*) FROM controls WHERE org_id = $1 AND is_custom = TRUE", orgID).Scan(&customCount)
	libraryCount = total - customCount

	var unownedCount, unmappedCount int
	database.QueryRow("SELECT COUNT(*) FROM controls WHERE org_id = $1 AND owner_id IS NULL", orgID).Scan(&unownedCount)
	database.QueryRow(`
		SELECT COUNT(*) FROM controls c
		WHERE c.org_id = $1 AND NOT EXISTS (
			SELECT 1 FROM control_mappings cm WHERE cm.control_id = c.id
		)
	`, orgID).Scan(&unmappedCount)

	// Frameworks coverage
	fwCoverage := []gin.H{}
	fwRows, err := database.Query(`
		SELECT f.name, fv.version, fv.id
		FROM org_frameworks of
		JOIN frameworks f ON f.id = of.framework_id
		JOIN framework_versions fv ON fv.id = of.active_version_id
		WHERE of.org_id = $1 AND of.status = 'active'
		ORDER BY f.name
	`, orgID)
	if err == nil {
		defer fwRows.Close()
		for fwRows.Next() {
			var fName, fvVersion, fvID string
			fwRows.Scan(&fName, &fvVersion, &fvID)

			stats := computeCoverageStats(orgID, fvID)
			fwCoverage = append(fwCoverage, gin.H{
				"framework":    fName,
				"version":      fvVersion,
				"in_scope":     stats["in_scope"],
				"covered":      stats["mapped"],
				"gaps":         stats["unmapped"],
				"coverage_pct": stats["coverage_pct"],
			})
		}
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"total":               total,
		"by_status":           byStatus,
		"by_category":         byCategory,
		"custom_count":        customCount,
		"library_count":       libraryCount,
		"unowned_count":       unownedCount,
		"unmapped_count":      unmappedCount,
		"frameworks_coverage": fwCoverage,
	}))
}
