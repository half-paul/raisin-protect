package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// =============================================================================
// CDE Assets
// =============================================================================

// ListCDEAssets returns a paginated, filtered list of CDE assets for the org.
// GET /api/v1/cde/assets
func ListCDEAssets(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("scope_status"); v != "" {
		where = append(where, fmt.Sprintf("scope_status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("type"); v != "" {
		where = append(where, fmt.Sprintf("type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("environment"); v != "" {
		where = append(where, fmt.Sprintf("environment = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("data_classification"); v != "" {
		where = append(where, fmt.Sprintf("data_classification = $%d", argN))
		args = append(args, v)
		argN++
	}

	allowedSorts := map[string]string{
		"name":                "name",
		"type":                "type",
		"scope_status":        "scope_status",
		"data_classification": "data_classification",
		"environment":         "environment",
		"created_at":          "created_at",
		"updated_at":          "updated_at",
	}
	sortCol := c.DefaultQuery("sort", "name")
	sortOrder := strings.ToUpper(c.DefaultQuery("order", "ASC"))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}
	orderBy := "name"
	if col, ok := allowedSorts[sortCol]; ok {
		orderBy = col
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := database.DB.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM cde_assets WHERE %s`, whereClause), args...,
	).Scan(&total); err != nil {
		log.Error().Err(err).Msg("cde: failed to count assets")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list CDE assets"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, org_id, name, type, ip_address, hostname, environment,
		       scope_status, scope_justification, data_classification, created_at, updated_at
		FROM cde_assets
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, sortOrder, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to query assets")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list CDE assets"))
		return
	}
	defer rows.Close()

	assets := make([]models.CDEAsset, 0, total)
	for rows.Next() {
		var a models.CDEAsset
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.Name, &a.Type, &a.IPAddress, &a.Hostname, &a.Environment,
			&a.ScopeStatus, &a.ScopeJustification, &a.DataClassification, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("cde: failed to scan asset")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read CDE assets"))
			return
		}
		assets = append(assets, a)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("cde: rows iteration error for assets")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list CDE assets"))
		return
	}

	c.JSON(http.StatusOK, listResponse(c, assets, total, page, perPage))
}

// GetCDEAsset returns a single CDE asset by ID.
// GET /api/v1/cde/assets/:id
func GetCDEAsset(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var a models.CDEAsset
	err := database.DB.QueryRow(`
		SELECT id, org_id, name, type, ip_address, hostname, environment,
		       scope_status, scope_justification, data_classification, created_at, updated_at
		FROM cde_assets
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&a.ID, &a.OrgID, &a.Name, &a.Type, &a.IPAddress, &a.Hostname, &a.Environment,
		&a.ScopeStatus, &a.ScopeJustification, &a.DataClassification, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "CDE asset not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to get asset")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get CDE asset"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, a))
}

// CreateCDEAsset creates a new CDE asset record.
// POST /api/v1/cde/assets
func CreateCDEAsset(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.CreateCDEAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if len(req.Name) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "name must be 255 characters or less"))
		return
	}
	if !models.IsValidCDEAssetType(req.Type) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid type; allowed: server, workstation, network_device, database, application, other"))
		return
	}
	if !models.IsValidCDEScopeStatus(req.ScopeStatus) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid scope_status; allowed: in_scope, out_of_scope, unclassified, connected_to_cde"))
		return
	}
	if !models.IsValidCDEDataClass(req.DataClassification) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid data_classification; allowed: pan, chd, sad, confidential, public"))
		return
	}
	if !models.IsValidCDEEnvironment(req.Environment) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid environment; allowed: production, staging, development, test"))
		return
	}

	id := uuid.New().String()
	var a models.CDEAsset
	err := database.DB.QueryRow(`
		INSERT INTO cde_assets (id, org_id, name, type, ip_address, hostname, environment,
		                        scope_status, scope_justification, data_classification)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, org_id, name, type, ip_address, hostname, environment,
		          scope_status, scope_justification, data_classification, created_at, updated_at
	`, id, orgID, req.Name, req.Type, req.IPAddress, req.Hostname, req.Environment,
		req.ScopeStatus, req.ScopeJustification, req.DataClassification,
	).Scan(
		&a.ID, &a.OrgID, &a.Name, &a.Type, &a.IPAddress, &a.Hostname, &a.Environment,
		&a.ScopeStatus, &a.ScopeJustification, &a.DataClassification, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to create asset")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create CDE asset"))
		return
	}

	middleware.LogAudit(c, "cde_asset.created", "cde_asset", &a.ID, map[string]interface{}{
		"name": a.Name, "type": a.Type, "scope_status": a.ScopeStatus,
	})

	c.JSON(http.StatusCreated, successResponse(c, a))
}

// UpdateCDEAsset updates a CDE asset. Only provided (non-nil) fields are changed.
// PUT /api/v1/cde/assets/:id
func UpdateCDEAsset(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	// Verify exists and belongs to org.
	var exists bool
	if err := database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM cde_assets WHERE id = $1 AND org_id = $2)`, id, orgID,
	).Scan(&exists); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to check asset existence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update CDE asset"))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "CDE asset not found"))
		return
	}

	var req models.UpdateCDEAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	// Validate optional enum fields when present.
	if req.Type != nil && !models.IsValidCDEAssetType(*req.Type) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid type"))
		return
	}
	if req.ScopeStatus != nil && !models.IsValidCDEScopeStatus(*req.ScopeStatus) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid scope_status"))
		return
	}
	if req.DataClassification != nil && !models.IsValidCDEDataClass(*req.DataClassification) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid data_classification"))
		return
	}
	if req.Environment != nil && !models.IsValidCDEEnvironment(*req.Environment) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid environment"))
		return
	}
	if req.Name != nil && len(*req.Name) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "name must be 255 characters or less"))
		return
	}

	// Build dynamic SET clause.
	setClauses := []string{"updated_at = NOW()"}
	setArgs := []interface{}{}
	argN := 1

	addSet := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argN))
		setArgs = append(setArgs, val)
		argN++
	}

	if req.Name != nil {
		addSet("name", *req.Name)
	}
	if req.Type != nil {
		addSet("type", *req.Type)
	}
	if req.IPAddress != nil {
		addSet("ip_address", *req.IPAddress)
	}
	if req.Hostname != nil {
		addSet("hostname", *req.Hostname)
	}
	if req.Environment != nil {
		addSet("environment", *req.Environment)
	}
	if req.ScopeStatus != nil {
		addSet("scope_status", *req.ScopeStatus)
	}
	if req.ScopeJustification != nil {
		addSet("scope_justification", *req.ScopeJustification)
	}
	if req.DataClassification != nil {
		addSet("data_classification", *req.DataClassification)
	}

	setArgs = append(setArgs, id, orgID)
	query := fmt.Sprintf(`
		UPDATE cde_assets SET %s
		WHERE id = $%d AND org_id = $%d
		RETURNING id, org_id, name, type, ip_address, hostname, environment,
		          scope_status, scope_justification, data_classification, created_at, updated_at
	`, strings.Join(setClauses, ", "), argN, argN+1)

	var a models.CDEAsset
	err := database.DB.QueryRow(query, setArgs...).Scan(
		&a.ID, &a.OrgID, &a.Name, &a.Type, &a.IPAddress, &a.Hostname, &a.Environment,
		&a.ScopeStatus, &a.ScopeJustification, &a.DataClassification, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to update asset")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update CDE asset"))
		return
	}

	middleware.LogAudit(c, "cde_asset.updated", "cde_asset", &a.ID, map[string]interface{}{
		"name": a.Name,
	})

	c.JSON(http.StatusOK, successResponse(c, a))
}

// DeleteCDEAsset deletes a CDE asset and any data flows that reference it.
// DELETE /api/v1/cde/assets/:id
func DeleteCDEAsset(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	tx, err := database.DB.Begin()
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to begin transaction for asset delete")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete CDE asset"))
		return
	}
	defer tx.Rollback() //nolint:errcheck

	// Verify ownership.
	var exists bool
	if err := tx.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM cde_assets WHERE id = $1 AND org_id = $2)`, id, orgID,
	).Scan(&exists); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to check asset for delete")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete CDE asset"))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "CDE asset not found"))
		return
	}

	// Cascade-delete referencing data flows (org-scoped for safety).
	if _, err := tx.Exec(
		`DELETE FROM cde_data_flows WHERE org_id = $1 AND (source_asset_id = $2 OR dest_asset_id = $2)`,
		orgID, id,
	); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to delete dependent data flows")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete CDE asset"))
		return
	}

	if _, err := tx.Exec(`DELETE FROM cde_assets WHERE id = $1 AND org_id = $2`, id, orgID); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to delete asset")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete CDE asset"))
		return
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to commit asset delete")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete CDE asset"))
		return
	}

	middleware.LogAudit(c, "cde_asset.deleted", "cde_asset", &id, nil)
	c.JSON(http.StatusNoContent, nil)
}

// =============================================================================
// CDE Network Segments
// =============================================================================

// ListCDESegments returns a paginated list of network segments for the org.
// GET /api/v1/cde/segments
func ListCDESegments(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("segment_type"); v != "" {
		where = append(where, fmt.Sprintf("segment_type = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := database.DB.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM cde_network_segments WHERE %s`, whereClause), args...,
	).Scan(&total); err != nil {
		log.Error().Err(err).Msg("cde: failed to count segments")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list network segments"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, org_id, name, vlan, subnet, segment_type, isolation_method, created_at, updated_at
		FROM cde_network_segments
		WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to query segments")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list network segments"))
		return
	}
	defer rows.Close()

	segments := make([]models.CDENetworkSegment, 0)
	for rows.Next() {
		var s models.CDENetworkSegment
		if err := rows.Scan(
			&s.ID, &s.OrgID, &s.Name, &s.VLAN, &s.Subnet, &s.SegmentType, &s.IsolationMethod,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("cde: failed to scan segment")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read network segments"))
			return
		}
		segments = append(segments, s)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("cde: rows error for segments")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list network segments"))
		return
	}

	c.JSON(http.StatusOK, listResponse(c, segments, total, page, perPage))
}

// GetCDESegment returns a single network segment by ID.
// GET /api/v1/cde/segments/:id
func GetCDESegment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var s models.CDENetworkSegment
	err := database.DB.QueryRow(`
		SELECT id, org_id, name, vlan, subnet, segment_type, isolation_method, created_at, updated_at
		FROM cde_network_segments
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&s.ID, &s.OrgID, &s.Name, &s.VLAN, &s.Subnet, &s.SegmentType, &s.IsolationMethod,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Network segment not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to get segment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get network segment"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, s))
}

// CreateCDESegment creates a new network segment.
// POST /api/v1/cde/segments
func CreateCDESegment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.CreateCDESegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if len(req.Name) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "name must be 255 characters or less"))
		return
	}
	if !models.IsValidCDESegmentType(req.SegmentType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid segment_type; allowed: cde, connected, out_of_scope, dmz, management"))
		return
	}

	id := uuid.New().String()
	var s models.CDENetworkSegment
	err := database.DB.QueryRow(`
		INSERT INTO cde_network_segments (id, org_id, name, vlan, subnet, segment_type, isolation_method)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, org_id, name, vlan, subnet, segment_type, isolation_method, created_at, updated_at
	`, id, orgID, req.Name, req.VLAN, req.Subnet, req.SegmentType, req.IsolationMethod,
	).Scan(
		&s.ID, &s.OrgID, &s.Name, &s.VLAN, &s.Subnet, &s.SegmentType, &s.IsolationMethod,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to create segment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create network segment"))
		return
	}

	middleware.LogAudit(c, "cde_segment.created", "cde_network_segment", &s.ID, map[string]interface{}{
		"name": s.Name, "segment_type": s.SegmentType,
	})

	c.JSON(http.StatusCreated, successResponse(c, s))
}

// UpdateCDESegment updates a network segment.
// PUT /api/v1/cde/segments/:id
func UpdateCDESegment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var exists bool
	if err := database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM cde_network_segments WHERE id = $1 AND org_id = $2)`, id, orgID,
	).Scan(&exists); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to check segment existence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update network segment"))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Network segment not found"))
		return
	}

	var req models.UpdateCDESegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if req.SegmentType != nil && !models.IsValidCDESegmentType(*req.SegmentType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid segment_type"))
		return
	}
	if req.Name != nil && len(*req.Name) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "name must be 255 characters or less"))
		return
	}

	setClauses := []string{"updated_at = NOW()"}
	setArgs := []interface{}{}
	argN := 1

	addSet := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argN))
		setArgs = append(setArgs, val)
		argN++
	}

	if req.Name != nil {
		addSet("name", *req.Name)
	}
	if req.VLAN != nil {
		addSet("vlan", *req.VLAN)
	}
	if req.Subnet != nil {
		addSet("subnet", *req.Subnet)
	}
	if req.SegmentType != nil {
		addSet("segment_type", *req.SegmentType)
	}
	if req.IsolationMethod != nil {
		addSet("isolation_method", *req.IsolationMethod)
	}

	setArgs = append(setArgs, id, orgID)
	query := fmt.Sprintf(`
		UPDATE cde_network_segments SET %s
		WHERE id = $%d AND org_id = $%d
		RETURNING id, org_id, name, vlan, subnet, segment_type, isolation_method, created_at, updated_at
	`, strings.Join(setClauses, ", "), argN, argN+1)

	var s models.CDENetworkSegment
	if err := database.DB.QueryRow(query, setArgs...).Scan(
		&s.ID, &s.OrgID, &s.Name, &s.VLAN, &s.Subnet, &s.SegmentType, &s.IsolationMethod,
		&s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to update segment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update network segment"))
		return
	}

	middleware.LogAudit(c, "cde_segment.updated", "cde_network_segment", &s.ID, map[string]interface{}{
		"name": s.Name,
	})

	c.JSON(http.StatusOK, successResponse(c, s))
}

// DeleteCDESegment deletes a network segment.
// DELETE /api/v1/cde/segments/:id
func DeleteCDESegment(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	res, err := database.DB.Exec(
		`DELETE FROM cde_network_segments WHERE id = $1 AND org_id = $2`, id, orgID,
	)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to delete segment")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete network segment"))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Network segment not found"))
		return
	}

	middleware.LogAudit(c, "cde_segment.deleted", "cde_network_segment", &id, nil)
	c.JSON(http.StatusNoContent, nil)
}

// =============================================================================
// CDE Data Flows
// =============================================================================

// ListCDEDataFlows returns a paginated list of data flows for the org.
// GET /api/v1/cde/data-flows
func ListCDEDataFlows(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"f.org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("source_asset_id"); v != "" {
		where = append(where, fmt.Sprintf("f.source_asset_id = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("dest_asset_id"); v != "" {
		where = append(where, fmt.Sprintf("f.dest_asset_id = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := database.DB.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM cde_data_flows f WHERE %s`, whereClause), args...,
	).Scan(&total); err != nil {
		log.Error().Err(err).Msg("cde: failed to count data flows")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list data flows"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT f.id, f.org_id, f.source_asset_id, f.dest_asset_id, f.protocol, f.port,
		       f.data_type, f.encryption_method, f.created_at, f.updated_at
		FROM cde_data_flows f
		WHERE %s
		ORDER BY f.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to query data flows")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list data flows"))
		return
	}
	defer rows.Close()

	flows := make([]models.CDEDataFlow, 0)
	for rows.Next() {
		var f models.CDEDataFlow
		if err := rows.Scan(
			&f.ID, &f.OrgID, &f.SourceAssetID, &f.DestAssetID, &f.Protocol, &f.Port,
			&f.DataType, &f.EncryptionMethod, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("cde: failed to scan data flow")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read data flows"))
			return
		}
		flows = append(flows, f)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("cde: rows error for data flows")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list data flows"))
		return
	}

	c.JSON(http.StatusOK, listResponse(c, flows, total, page, perPage))
}

// GetCDEDataFlow returns a single data flow by ID.
// GET /api/v1/cde/data-flows/:id
func GetCDEDataFlow(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var f models.CDEDataFlow
	err := database.DB.QueryRow(`
		SELECT id, org_id, source_asset_id, dest_asset_id, protocol, port,
		       data_type, encryption_method, created_at, updated_at
		FROM cde_data_flows
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&f.ID, &f.OrgID, &f.SourceAssetID, &f.DestAssetID, &f.Protocol, &f.Port,
		&f.DataType, &f.EncryptionMethod, &f.CreatedAt, &f.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Data flow not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to get data flow")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get data flow"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, f))
}

// CreateCDEDataFlow creates a new data flow record between two CDE assets.
// POST /api/v1/cde/data-flows
func CreateCDEDataFlow(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.CreateCDEDataFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if req.SourceAssetID == req.DestAssetID {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "source_asset_id and dest_asset_id must be different"))
		return
	}
	if req.Port != nil && (*req.Port < 1 || *req.Port > 65535) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "port must be between 1 and 65535"))
		return
	}

	// Verify both assets belong to the org.
	var srcExists, dstExists bool
	if err := database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM cde_assets WHERE id = $1 AND org_id = $2)`,
		req.SourceAssetID, orgID,
	).Scan(&srcExists); err != nil || !srcExists {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "source_asset_id not found in this org"))
		return
	}
	if err := database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM cde_assets WHERE id = $1 AND org_id = $2)`,
		req.DestAssetID, orgID,
	).Scan(&dstExists); err != nil || !dstExists {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "dest_asset_id not found in this org"))
		return
	}

	id := uuid.New().String()
	var f models.CDEDataFlow
	err := database.DB.QueryRow(`
		INSERT INTO cde_data_flows (id, org_id, source_asset_id, dest_asset_id, protocol, port, data_type, encryption_method)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, org_id, source_asset_id, dest_asset_id, protocol, port, data_type, encryption_method, created_at, updated_at
	`, id, orgID, req.SourceAssetID, req.DestAssetID, req.Protocol, req.Port, req.DataType, req.EncryptionMethod,
	).Scan(
		&f.ID, &f.OrgID, &f.SourceAssetID, &f.DestAssetID, &f.Protocol, &f.Port,
		&f.DataType, &f.EncryptionMethod, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to create data flow")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create data flow"))
		return
	}

	middleware.LogAudit(c, "cde_data_flow.created", "cde_data_flow", &f.ID, map[string]interface{}{
		"source_asset_id": f.SourceAssetID, "dest_asset_id": f.DestAssetID,
	})

	c.JSON(http.StatusCreated, successResponse(c, f))
}

// UpdateCDEDataFlow updates a data flow record.
// PUT /api/v1/cde/data-flows/:id
func UpdateCDEDataFlow(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var exists bool
	if err := database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM cde_data_flows WHERE id = $1 AND org_id = $2)`, id, orgID,
	).Scan(&exists); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to check data flow existence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update data flow"))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Data flow not found"))
		return
	}

	var req models.UpdateCDEDataFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if req.Port != nil && (*req.Port < 1 || *req.Port > 65535) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "port must be between 1 and 65535"))
		return
	}

	// If changing asset references, verify they belong to the org.
	if req.SourceAssetID != nil {
		var srcExists bool
		if err := database.DB.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM cde_assets WHERE id = $1 AND org_id = $2)`,
			*req.SourceAssetID, orgID,
		).Scan(&srcExists); err != nil || !srcExists {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "source_asset_id not found in this org"))
			return
		}
	}
	if req.DestAssetID != nil {
		var dstExists bool
		if err := database.DB.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM cde_assets WHERE id = $1 AND org_id = $2)`,
			*req.DestAssetID, orgID,
		).Scan(&dstExists); err != nil || !dstExists {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "dest_asset_id not found in this org"))
			return
		}
	}

	setClauses := []string{"updated_at = NOW()"}
	setArgs := []interface{}{}
	argN := 1

	addSet := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argN))
		setArgs = append(setArgs, val)
		argN++
	}

	if req.SourceAssetID != nil {
		addSet("source_asset_id", *req.SourceAssetID)
	}
	if req.DestAssetID != nil {
		addSet("dest_asset_id", *req.DestAssetID)
	}
	if req.Protocol != nil {
		addSet("protocol", *req.Protocol)
	}
	if req.Port != nil {
		addSet("port", *req.Port)
	}
	if req.DataType != nil {
		addSet("data_type", *req.DataType)
	}
	if req.EncryptionMethod != nil {
		addSet("encryption_method", *req.EncryptionMethod)
	}

	setArgs = append(setArgs, id, orgID)
	query := fmt.Sprintf(`
		UPDATE cde_data_flows SET %s
		WHERE id = $%d AND org_id = $%d
		RETURNING id, org_id, source_asset_id, dest_asset_id, protocol, port,
		          data_type, encryption_method, created_at, updated_at
	`, strings.Join(setClauses, ", "), argN, argN+1)

	var f models.CDEDataFlow
	if err := database.DB.QueryRow(query, setArgs...).Scan(
		&f.ID, &f.OrgID, &f.SourceAssetID, &f.DestAssetID, &f.Protocol, &f.Port,
		&f.DataType, &f.EncryptionMethod, &f.CreatedAt, &f.UpdatedAt,
	); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to update data flow")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update data flow"))
		return
	}

	middleware.LogAudit(c, "cde_data_flow.updated", "cde_data_flow", &f.ID, nil)
	c.JSON(http.StatusOK, successResponse(c, f))
}

// DeleteCDEDataFlow deletes a data flow record.
// DELETE /api/v1/cde/data-flows/:id
func DeleteCDEDataFlow(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	res, err := database.DB.Exec(
		`DELETE FROM cde_data_flows WHERE id = $1 AND org_id = $2`, id, orgID,
	)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to delete data flow")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete data flow"))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Data flow not found"))
		return
	}

	middleware.LogAudit(c, "cde_data_flow.deleted", "cde_data_flow", &id, nil)
	c.JSON(http.StatusNoContent, nil)
}

// =============================================================================
// CDE Segmentation Tests
// =============================================================================

// ListCDESegmentationTests returns a paginated list of segmentation tests.
// GET /api/v1/cde/segmentation-tests
func ListCDESegmentationTests(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	where := []string{"org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if v := c.Query("result"); v != "" {
		where = append(where, fmt.Sprintf("result = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("segment_id"); v != "" {
		where = append(where, fmt.Sprintf("segment_id = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := database.DB.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM segmentation_tests WHERE %s`, whereClause), args...,
	).Scan(&total); err != nil {
		log.Error().Err(err).Msg("cde: failed to count segmentation tests")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list segmentation tests"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, org_id, test_date, tester, methodology, segment_id,
		       result, findings, next_test_date, created_at, updated_at
		FROM segmentation_tests
		WHERE %s
		ORDER BY test_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to query segmentation tests")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list segmentation tests"))
		return
	}
	defer rows.Close()

	tests := make([]models.CDESegmentationTest, 0)
	for rows.Next() {
		var t models.CDESegmentationTest
		if err := rows.Scan(
			&t.ID, &t.OrgID, &t.TestDate, &t.Tester, &t.Methodology, &t.SegmentID,
			&t.Result, &t.Findings, &t.NextTestDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("cde: failed to scan segmentation test")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read segmentation tests"))
			return
		}
		tests = append(tests, t)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("cde: rows error for segmentation tests")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list segmentation tests"))
		return
	}

	c.JSON(http.StatusOK, listResponse(c, tests, total, page, perPage))
}

// GetCDESegmentationTest returns a single segmentation test by ID.
// GET /api/v1/cde/segmentation-tests/:id
func GetCDESegmentationTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var t models.CDESegmentationTest
	err := database.DB.QueryRow(`
		SELECT id, org_id, test_date, tester, methodology, segment_id,
		       result, findings, next_test_date, created_at, updated_at
		FROM segmentation_tests
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&t.ID, &t.OrgID, &t.TestDate, &t.Tester, &t.Methodology, &t.SegmentID,
		&t.Result, &t.Findings, &t.NextTestDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Segmentation test not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to get segmentation test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get segmentation test"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, t))
}

// CreateCDESegmentationTest records a new segmentation test for PCI DSS Req 11.4.
// POST /api/v1/cde/segmentation-tests
func CreateCDESegmentationTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.CreateCDESegTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if !models.IsValidCDESegTestResult(req.Result) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid result; allowed: pass, fail, n/a"))
		return
	}

	testDate, err := time.Parse(time.RFC3339, req.TestDate)
	if err != nil {
		// Try date-only format.
		testDate, err = time.Parse("2006-01-02", req.TestDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "test_date must be RFC3339 or YYYY-MM-DD"))
			return
		}
	}

	var nextTestDate *time.Time
	if req.NextTestDate != nil && *req.NextTestDate != "" {
		t, err := time.Parse(time.RFC3339, *req.NextTestDate)
		if err != nil {
			t, err = time.Parse("2006-01-02", *req.NextTestDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "next_test_date must be RFC3339 or YYYY-MM-DD"))
				return
			}
		}
		nextTestDate = &t
	}

	// If a segment_id is provided, verify it belongs to this org.
	if req.SegmentID != nil {
		var segExists bool
		if err := database.DB.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM cde_network_segments WHERE id = $1 AND org_id = $2)`,
			*req.SegmentID, orgID,
		).Scan(&segExists); err != nil || !segExists {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "segment_id not found in this org"))
			return
		}
	}

	id := uuid.New().String()
	var t models.CDESegmentationTest
	err = database.DB.QueryRow(`
		INSERT INTO segmentation_tests
		    (id, org_id, test_date, tester, methodology, segment_id, result, findings, next_test_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, org_id, test_date, tester, methodology, segment_id,
		          result, findings, next_test_date, created_at, updated_at
	`, id, orgID, testDate, req.Tester, req.Methodology, req.SegmentID, req.Result, req.Findings, nextTestDate,
	).Scan(
		&t.ID, &t.OrgID, &t.TestDate, &t.Tester, &t.Methodology, &t.SegmentID,
		&t.Result, &t.Findings, &t.NextTestDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to create segmentation test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create segmentation test"))
		return
	}

	middleware.LogAudit(c, "cde_seg_test.created", "segmentation_test", &t.ID, map[string]interface{}{
		"result": t.Result, "test_date": t.TestDate,
	})

	c.JSON(http.StatusCreated, successResponse(c, t))
}

// UpdateCDESegmentationTest updates a segmentation test record.
// PUT /api/v1/cde/segmentation-tests/:id
func UpdateCDESegmentationTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var exists bool
	if err := database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM segmentation_tests WHERE id = $1 AND org_id = $2)`, id, orgID,
	).Scan(&exists); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to check seg test existence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update segmentation test"))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Segmentation test not found"))
		return
	}

	var req models.UpdateCDESegTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if req.Result != nil && !models.IsValidCDESegTestResult(*req.Result) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid result; allowed: pass, fail, n/a"))
		return
	}

	setClauses := []string{"updated_at = NOW()"}
	setArgs := []interface{}{}
	argN := 1

	addSet := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argN))
		setArgs = append(setArgs, val)
		argN++
	}

	if req.TestDate != nil {
		t, err := time.Parse(time.RFC3339, *req.TestDate)
		if err != nil {
			t, err = time.Parse("2006-01-02", *req.TestDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "test_date must be RFC3339 or YYYY-MM-DD"))
				return
			}
		}
		addSet("test_date", t)
	}
	if req.Tester != nil {
		addSet("tester", *req.Tester)
	}
	if req.Methodology != nil {
		addSet("methodology", *req.Methodology)
	}
	if req.SegmentID != nil {
		addSet("segment_id", *req.SegmentID)
	}
	if req.Result != nil {
		addSet("result", *req.Result)
	}
	if req.Findings != nil {
		addSet("findings", *req.Findings)
	}
	if req.NextTestDate != nil {
		t, err := time.Parse(time.RFC3339, *req.NextTestDate)
		if err != nil {
			t, err = time.Parse("2006-01-02", *req.NextTestDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "next_test_date must be RFC3339 or YYYY-MM-DD"))
				return
			}
		}
		addSet("next_test_date", t)
	}

	setArgs = append(setArgs, id, orgID)
	query := fmt.Sprintf(`
		UPDATE segmentation_tests SET %s
		WHERE id = $%d AND org_id = $%d
		RETURNING id, org_id, test_date, tester, methodology, segment_id,
		          result, findings, next_test_date, created_at, updated_at
	`, strings.Join(setClauses, ", "), argN, argN+1)

	var t models.CDESegmentationTest
	if err := database.DB.QueryRow(query, setArgs...).Scan(
		&t.ID, &t.OrgID, &t.TestDate, &t.Tester, &t.Methodology, &t.SegmentID,
		&t.Result, &t.Findings, &t.NextTestDate, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to update segmentation test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update segmentation test"))
		return
	}

	middleware.LogAudit(c, "cde_seg_test.updated", "segmentation_test", &t.ID, nil)
	c.JSON(http.StatusOK, successResponse(c, t))
}

// DeleteCDESegmentationTest deletes a segmentation test record.
// DELETE /api/v1/cde/segmentation-tests/:id
func DeleteCDESegmentationTest(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	res, err := database.DB.Exec(
		`DELETE FROM segmentation_tests WHERE id = $1 AND org_id = $2`, id, orgID,
	)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("cde: failed to delete segmentation test")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete segmentation test"))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Segmentation test not found"))
		return
	}

	middleware.LogAudit(c, "cde_seg_test.deleted", "segmentation_test", &id, nil)
	c.JSON(http.StatusNoContent, nil)
}

// =============================================================================
// CDE Scope Summary
// =============================================================================

// GetCDEScopeSummary returns an aggregated view of the CDE scope posture.
// GET /api/v1/cde/scope-summary
func GetCDEScopeSummary(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var summary models.CDEScopeSummary

	// --- Asset counts ---
	rows, err := database.DB.Query(`
		SELECT scope_status, COUNT(*)
		FROM cde_assets
		WHERE org_id = $1
		GROUP BY scope_status
	`, orgID)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to aggregate asset scope")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to compute scope summary"))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			log.Error().Err(err).Msg("cde: failed to scan asset scope row")
			continue
		}
		summary.Assets.Total += count
		switch status {
		case models.CDEScopeStatusInScope:
			summary.Assets.InScope = count
		case models.CDEScopeStatusOutOfScope:
			summary.Assets.OutOfScope = count
		case models.CDEScopeStatusConnectedTo:
			summary.Assets.ConnectedTo = count
		case models.CDEScopeStatusUnclassified:
			summary.Assets.Unclassified = count
		}
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("cde: rows error for asset summary")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to compute scope summary"))
		return
	}

	// --- Segment counts ---
	segRows, err := database.DB.Query(`
		SELECT segment_type, COUNT(*)
		FROM cde_network_segments
		WHERE org_id = $1
		GROUP BY segment_type
	`, orgID)
	if err != nil {
		log.Error().Err(err).Msg("cde: failed to aggregate segment types")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to compute scope summary"))
		return
	}
	defer segRows.Close()
	summary.Segments.ByCDEType = make(map[string]int)
	for segRows.Next() {
		var sType string
		var count int
		if err := segRows.Scan(&sType, &count); err != nil {
			log.Error().Err(err).Msg("cde: failed to scan segment type row")
			continue
		}
		summary.Segments.Total += count
		summary.Segments.ByCDEType[sType] = count
	}
	if err := segRows.Err(); err != nil {
		log.Error().Err(err).Msg("cde: rows error for segment summary")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to compute scope summary"))
		return
	}

	// --- Data flow counts ---
	if err := database.DB.QueryRow(`
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE encryption_method IS NOT NULL AND encryption_method != '')
		FROM cde_data_flows
		WHERE org_id = $1
	`, orgID).Scan(&summary.DataFlows.Total, &summary.DataFlows.Encrypted); err != nil {
		log.Error().Err(err).Msg("cde: failed to aggregate data flows")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to compute scope summary"))
		return
	}

	// --- Segmentation test stats ---
	if err := database.DB.QueryRow(`
		SELECT
			COUNT(*),
			MAX(test_date),
			MIN(next_test_date) FILTER (WHERE next_test_date > NOW()),
			COUNT(*) FILTER (WHERE next_test_date IS NOT NULL AND next_test_date < NOW()),
			COUNT(*) FILTER (WHERE result = 'pass'),
			COUNT(*) FILTER (WHERE result = 'fail')
		FROM segmentation_tests
		WHERE org_id = $1
	`, orgID).Scan(
		&summary.SegmentationTests.Total,
		&summary.SegmentationTests.LastTestDate,
		&summary.SegmentationTests.NextTestDate,
		&summary.SegmentationTests.OverdueCount,
		&summary.SegmentationTests.PassCount,
		&summary.SegmentationTests.FailCount,
	); err != nil && err != sql.ErrNoRows {
		log.Error().Err(err).Msg("cde: failed to aggregate segmentation tests")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to compute scope summary"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, summary))
}
