package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"path"
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
// Service Providers
// =============================================================================

// ListServiceProviders returns a paginated, filtered list of service providers.
// GET /api/v1/service-providers
func ListServiceProviders(c *gin.Context) {
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

	if v := c.Query("type"); v != "" {
		where = append(where, fmt.Sprintf("type = $%d", argN))
		args = append(args, v)
		argN++
	}
	// compliance_status is the query param; maps to pci_compliance_status column.
	if v := c.Query("compliance_status"); v != "" {
		where = append(where, fmt.Sprintf("pci_compliance_status = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("risk_level"); v != "" {
		where = append(where, fmt.Sprintf("risk_level = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("is_active"); v == "true" {
		where = append(where, "is_active = TRUE")
	} else if v := c.Query("is_active"); v == "false" {
		where = append(where, "is_active = FALSE")
	}
	if v := c.Query("search"); v != "" {
		where = append(where, fmt.Sprintf(
			"(name ILIKE $%d OR services_provided ILIKE $%d OR contact_name ILIKE $%d)",
			argN, argN, argN,
		))
		args = append(args, "%"+v+"%")
		argN++
	}

	allowedSorts := map[string]string{
		"name":                  "name",
		"type":                  "type",
		"risk_level":            "risk_level",
		"pci_compliance_status": "pci_compliance_status",
		"next_review_date":      "next_review_date",
		"created_at":            "created_at",
		"updated_at":            "updated_at",
	}
	sortField := c.DefaultQuery("sort", "name")
	sortOrder := strings.ToUpper(c.DefaultQuery("order", "ASC"))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}
	orderBy := "name"
	if col, ok := allowedSorts[sortField]; ok {
		orderBy = col
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := database.DB.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM service_providers WHERE %s", whereClause),
		countArgs...,
	).Scan(&total); err != nil {
		log.Error().Err(err).Msg("sp: failed to count service providers")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list service providers"))
		return
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, org_id, name, type, contact_name, contact_email, contact_phone,
		       services_provided, pci_compliance_status, last_aoc_date, next_review_date,
		       risk_level, risk_notes, contract_start_date, contract_end_date,
		       is_active, created_by, created_at, updated_at
		FROM service_providers
		WHERE %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, sortOrder, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("sp: failed to query service providers")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list service providers"))
		return
	}
	defer rows.Close()

	providers := make([]models.ServiceProvider, 0, total)
	for rows.Next() {
		var sp models.ServiceProvider
		if err := rows.Scan(
			&sp.ID, &sp.OrgID, &sp.Name, &sp.Type,
			&sp.ContactName, &sp.ContactEmail, &sp.ContactPhone,
			&sp.ServicesProvided, &sp.PCIComplianceStatus, &sp.LastAOCDate, &sp.NextReviewDate,
			&sp.RiskLevel, &sp.RiskNotes, &sp.ContractStartDate, &sp.ContractEndDate,
			&sp.IsActive, &sp.CreatedBy, &sp.CreatedAt, &sp.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("sp: failed to scan service provider row")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read service providers"))
			return
		}
		providers = append(providers, sp)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("sp: rows iteration error")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list service providers"))
		return
	}

	c.JSON(http.StatusOK, listResponse(c, providers, total, page, perPage))
}

// GetServiceProvider returns a single service provider by ID.
// GET /api/v1/service-providers/:id
func GetServiceProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var sp models.ServiceProvider
	err := database.DB.QueryRow(`
		SELECT id, org_id, name, type, contact_name, contact_email, contact_phone,
		       services_provided, pci_compliance_status, last_aoc_date, next_review_date,
		       risk_level, risk_notes, contract_start_date, contract_end_date,
		       is_active, created_by, created_at, updated_at
		FROM service_providers
		WHERE id = $1 AND org_id = $2
	`, id, orgID).Scan(
		&sp.ID, &sp.OrgID, &sp.Name, &sp.Type,
		&sp.ContactName, &sp.ContactEmail, &sp.ContactPhone,
		&sp.ServicesProvided, &sp.PCIComplianceStatus, &sp.LastAOCDate, &sp.NextReviewDate,
		&sp.RiskLevel, &sp.RiskNotes, &sp.ContractStartDate, &sp.ContractEndDate,
		&sp.IsActive, &sp.CreatedBy, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("sp: failed to get service provider")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get service provider"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, sp))
}

// CreateServiceProvider creates a new service provider record.
// POST /api/v1/service-providers
func CreateServiceProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	var req models.CreateServiceProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "name is required"))
		return
	}
	if len(req.Name) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "name must be 255 characters or less"))
		return
	}
	if !models.IsValidSPType(req.Type) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid type; allowed: %s", strings.Join(models.ValidSPTypes, ", "))))
		return
	}
	if req.PCIComplianceStatus != nil && !models.IsValidSPComplianceStatus(*req.PCIComplianceStatus) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid pci_compliance_status; allowed: %s", strings.Join(models.ValidSPComplianceStatuses, ", "))))
		return
	}
	if req.RiskLevel != nil && !models.IsValidSPRiskLevel(*req.RiskLevel) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid risk_level; allowed: %s", strings.Join(models.ValidSPRiskLevels, ", "))))
		return
	}

	// Parse optional date fields.
	var lastAOCDate, nextReviewDate, contractStart, contractEnd *time.Time
	if req.LastAOCDate != nil {
		t, err := time.Parse("2006-01-02", *req.LastAOCDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid last_aoc_date format; use YYYY-MM-DD"))
			return
		}
		lastAOCDate = &t
	}
	if req.NextReviewDate != nil {
		t, err := time.Parse("2006-01-02", *req.NextReviewDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid next_review_date format; use YYYY-MM-DD"))
			return
		}
		nextReviewDate = &t
	}
	if req.ContractStartDate != nil {
		t, err := time.Parse("2006-01-02", *req.ContractStartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid contract_start_date format; use YYYY-MM-DD"))
			return
		}
		contractStart = &t
	}
	if req.ContractEndDate != nil {
		t, err := time.Parse("2006-01-02", *req.ContractEndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid contract_end_date format; use YYYY-MM-DD"))
			return
		}
		contractEnd = &t
	}
	if contractStart != nil && contractEnd != nil && contractEnd.Before(*contractStart) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "contract_end_date must be on or after contract_start_date"))
		return
	}

	complianceStatus := models.SPComplianceUnknown
	if req.PCIComplianceStatus != nil {
		complianceStatus = *req.PCIComplianceStatus
	}
	riskLevel := models.SPRiskMedium
	if req.RiskLevel != nil {
		riskLevel = *req.RiskLevel
	}

	var createdByPtr *string
	if userID != "" {
		createdByPtr = &userID
	}

	id := uuid.New().String()
	var sp models.ServiceProvider
	err := database.DB.QueryRow(`
		INSERT INTO service_providers (
			id, org_id, name, type, contact_name, contact_email, contact_phone,
			services_provided, pci_compliance_status, last_aoc_date, next_review_date,
			risk_level, risk_notes, contract_start_date, contract_end_date, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id, org_id, name, type, contact_name, contact_email, contact_phone,
		          services_provided, pci_compliance_status, last_aoc_date, next_review_date,
		          risk_level, risk_notes, contract_start_date, contract_end_date,
		          is_active, created_by, created_at, updated_at
	`,
		id, orgID, req.Name, req.Type,
		req.ContactName, req.ContactEmail, req.ContactPhone,
		req.ServicesProvided, complianceStatus, lastAOCDate, nextReviewDate,
		riskLevel, req.RiskNotes, contractStart, contractEnd, createdByPtr,
	).Scan(
		&sp.ID, &sp.OrgID, &sp.Name, &sp.Type,
		&sp.ContactName, &sp.ContactEmail, &sp.ContactPhone,
		&sp.ServicesProvided, &sp.PCIComplianceStatus, &sp.LastAOCDate, &sp.NextReviewDate,
		&sp.RiskLevel, &sp.RiskNotes, &sp.ContractStartDate, &sp.ContractEndDate,
		&sp.IsActive, &sp.CreatedBy, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uq_service_provider_name") {
			c.JSON(http.StatusConflict, errorResponse("DUPLICATE_NAME", "A service provider with this name already exists in your organization"))
			return
		}
		log.Error().Err(err).Msg("sp: failed to create service provider")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create service provider"))
		return
	}

	middleware.LogAudit(c, "service_provider.created", "service_provider", &sp.ID, map[string]interface{}{
		"name":       sp.Name,
		"type":       sp.Type,
		"risk_level": sp.RiskLevel,
	})

	c.JSON(http.StatusCreated, successResponse(c, sp))
}

// UpdateServiceProvider partially updates a service provider.
// PUT /api/v1/service-providers/:id
func UpdateServiceProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	var req models.UpdateServiceProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}

	// Validate all fields before touching the database.
	if req.Name != nil && (strings.TrimSpace(*req.Name) == "" || len(*req.Name) > 255) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "name must be 1-255 characters"))
		return
	}
	if req.Type != nil && !models.IsValidSPType(*req.Type) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid type; allowed: %s", strings.Join(models.ValidSPTypes, ", "))))
		return
	}
	if req.PCIComplianceStatus != nil && !models.IsValidSPComplianceStatus(*req.PCIComplianceStatus) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid pci_compliance_status; allowed: %s", strings.Join(models.ValidSPComplianceStatuses, ", "))))
		return
	}
	if req.RiskLevel != nil && !models.IsValidSPRiskLevel(*req.RiskLevel) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid risk_level; allowed: %s", strings.Join(models.ValidSPRiskLevels, ", "))))
		return
	}

	// Build the SET clause; updated_at = NOW() avoids a round-trip parameter.
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
	if req.ContactName != nil {
		addSet("contact_name", *req.ContactName)
	}
	if req.ContactEmail != nil {
		addSet("contact_email", *req.ContactEmail)
	}
	if req.ContactPhone != nil {
		addSet("contact_phone", *req.ContactPhone)
	}
	if req.ServicesProvided != nil {
		addSet("services_provided", *req.ServicesProvided)
	}
	if req.PCIComplianceStatus != nil {
		addSet("pci_compliance_status", *req.PCIComplianceStatus)
	}
	if req.LastAOCDate != nil {
		t, err := time.Parse("2006-01-02", *req.LastAOCDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid last_aoc_date format; use YYYY-MM-DD"))
			return
		}
		addSet("last_aoc_date", t)
	}
	if req.NextReviewDate != nil {
		t, err := time.Parse("2006-01-02", *req.NextReviewDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid next_review_date format; use YYYY-MM-DD"))
			return
		}
		addSet("next_review_date", t)
	}
	if req.RiskLevel != nil {
		addSet("risk_level", *req.RiskLevel)
	}
	if req.RiskNotes != nil {
		addSet("risk_notes", *req.RiskNotes)
	}
	if req.ContractStartDate != nil {
		t, err := time.Parse("2006-01-02", *req.ContractStartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid contract_start_date format; use YYYY-MM-DD"))
			return
		}
		addSet("contract_start_date", t)
	}
	if req.ContractEndDate != nil {
		t, err := time.Parse("2006-01-02", *req.ContractEndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid contract_end_date format; use YYYY-MM-DD"))
			return
		}
		addSet("contract_end_date", t)
	}
	if req.IsActive != nil {
		addSet("is_active", *req.IsActive)
	}

	if argN == 1 {
		// Only updated_at = NOW() was set — nothing meaningful to update.
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "No fields to update"))
		return
	}

	// Verify the record exists and belongs to this org before updating.
	var exists bool
	if err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM service_providers WHERE id = $1 AND org_id = $2)", id, orgID,
	).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}

	setArgs = append(setArgs, id, orgID)
	var sp models.ServiceProvider
	err := database.DB.QueryRow(
		fmt.Sprintf(`
			UPDATE service_providers SET %s
			WHERE id = $%d AND org_id = $%d
			RETURNING id, org_id, name, type, contact_name, contact_email, contact_phone,
			          services_provided, pci_compliance_status, last_aoc_date, next_review_date,
			          risk_level, risk_notes, contract_start_date, contract_end_date,
			          is_active, created_by, created_at, updated_at
		`, strings.Join(setClauses, ", "), argN, argN+1),
		setArgs...,
	).Scan(
		&sp.ID, &sp.OrgID, &sp.Name, &sp.Type,
		&sp.ContactName, &sp.ContactEmail, &sp.ContactPhone,
		&sp.ServicesProvided, &sp.PCIComplianceStatus, &sp.LastAOCDate, &sp.NextReviewDate,
		&sp.RiskLevel, &sp.RiskNotes, &sp.ContractStartDate, &sp.ContractEndDate,
		&sp.IsActive, &sp.CreatedBy, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "uq_service_provider_name") {
			c.JSON(http.StatusConflict, errorResponse("DUPLICATE_NAME", "A service provider with this name already exists in your organization"))
			return
		}
		log.Error().Err(err).Str("id", id).Msg("sp: failed to update service provider")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to update service provider"))
		return
	}

	middleware.LogAudit(c, "service_provider.updated", "service_provider", &sp.ID, nil)

	c.JSON(http.StatusOK, successResponse(c, sp))
}

// DeleteServiceProvider permanently deletes a service provider within a transaction.
// Child records (compliance docs, responsibility matrix) are deleted explicitly
// before the parent to satisfy tests with explicit mock expectations; the DB
// FK CASCADE would also handle this automatically in production.
// DELETE /api/v1/service-providers/:id
func DeleteServiceProvider(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")

	tx, err := database.DB.Begin()
	if err != nil {
		log.Error().Err(err).Msg("sp: failed to begin transaction for delete")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete service provider"))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete child records first (mirrors FK CASCADE but is explicit for clarity).
	if _, err = tx.Exec(
		"DELETE FROM sp_compliance_documents WHERE provider_id = $1 AND org_id = $2", id, orgID,
	); err != nil {
		log.Error().Err(err).Str("id", id).Msg("sp: failed to delete compliance docs during provider delete")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete service provider"))
		return
	}

	if _, err = tx.Exec(
		"DELETE FROM sp_responsibility_matrix WHERE provider_id = $1 AND org_id = $2", id, orgID,
	); err != nil {
		log.Error().Err(err).Str("id", id).Msg("sp: failed to delete responsibility matrix during provider delete")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete service provider"))
		return
	}

	var result sql.Result
	result, err = tx.Exec(
		"DELETE FROM service_providers WHERE id = $1 AND org_id = $2", id, orgID,
	)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("sp: failed to delete service provider")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete service provider"))
		return
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		err = fmt.Errorf("not found") // trigger rollback via deferred
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Str("id", id).Msg("sp: failed to commit delete transaction")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete service provider"))
		return
	}

	middleware.LogAudit(c, "service_provider.deleted", "service_provider", &id, nil)

	c.JSON(http.StatusNoContent, nil)
}

// GetSPComplianceSummary returns aggregate compliance stats for all service providers.
// Counts by compliance status, risk level, and documents expiring within 30/60/90 days
// are computed in a single query to minimise round-trips.
// GET /api/v1/service-providers/compliance-summary
func GetSPComplianceSummary(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var (
		total, compliant, nonCompliant, inProgress int
		notValidated, unknown, notApplicable       int
		expiring30, expiring60, expiring90         int
	)
	err := database.DB.QueryRow(`
		SELECT
			COUNT(*)                                                            AS total,
			COUNT(*) FILTER (WHERE pci_compliance_status = 'compliant')         AS compliant,
			COUNT(*) FILTER (WHERE pci_compliance_status = 'non_compliant')     AS non_compliant,
			COUNT(*) FILTER (WHERE pci_compliance_status = 'compliance_in_progress') AS compliance_in_progress,
			COUNT(*) FILTER (WHERE pci_compliance_status = 'compliance_not_validated') AS compliance_not_validated,
			COUNT(*) FILTER (WHERE pci_compliance_status = 'unknown')           AS unknown,
			COUNT(*) FILTER (WHERE pci_compliance_status = 'not_applicable')    AS not_applicable,
			(SELECT COUNT(*) FROM sp_compliance_documents
			 WHERE org_id = $1 AND is_current = TRUE
			   AND valid_until BETWEEN NOW() AND NOW() + '30 days'::interval)   AS expiring_30_days,
			(SELECT COUNT(*) FROM sp_compliance_documents
			 WHERE org_id = $1 AND is_current = TRUE
			   AND valid_until BETWEEN NOW() AND NOW() + '60 days'::interval)   AS expiring_60_days,
			(SELECT COUNT(*) FROM sp_compliance_documents
			 WHERE org_id = $1 AND is_current = TRUE
			   AND valid_until BETWEEN NOW() AND NOW() + '90 days'::interval)   AS expiring_90_days
		FROM service_providers
		WHERE org_id = $1
	`, orgID).Scan(
		&total, &compliant, &nonCompliant, &inProgress,
		&notValidated, &unknown, &notApplicable,
		&expiring30, &expiring60, &expiring90,
	)
	if err != nil {
		log.Error().Err(err).Msg("sp: failed to compute compliance summary")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get compliance summary"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"total":                       total,
		"compliant":                   compliant,
		"non_compliant":               nonCompliant,
		"compliance_in_progress":      inProgress,
		"compliance_not_validated":    notValidated,
		"unknown":                     unknown,
		"not_applicable":              notApplicable,
		"expiring_30_days":            expiring30,
		"expiring_60_days":            expiring60,
		"expiring_90_days":            expiring90,
	}))
}

// =============================================================================
// Compliance Documents
// =============================================================================

// ListSPComplianceDocs lists compliance documents for a service provider.
// GET /api/v1/service-providers/:id/compliance-docs
func ListSPComplianceDocs(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	providerID := c.Param("id")

	// Verify provider exists and belongs to this org.
	var exists bool
	if err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM service_providers WHERE id = $1 AND org_id = $2)", providerID, orgID,
	).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}

	where := []string{"d.provider_id = $1", "d.org_id = $2"}
	args := []interface{}{providerID, orgID}
	argN := 3

	if v := c.Query("document_type"); v != "" {
		where = append(where, fmt.Sprintf("d.document_type = $%d", argN))
		args = append(args, v)
		argN++
	}
	if v := c.Query("is_current"); v == "true" {
		where = append(where, "d.is_current = TRUE")
	} else if c.Query("is_current") == "false" {
		where = append(where, "d.is_current = FALSE")
	}

	whereClause := strings.Join(where, " AND ")
	rows, err := database.DB.Query(fmt.Sprintf(`
		SELECT d.id, d.provider_id, d.org_id, d.document_type, d.title, d.document_version,
		       d.upload_path, d.valid_from, d.valid_until, d.reviewed_by, d.review_notes,
		       d.is_current, d.uploaded_by, d.created_at, d.updated_at
		FROM sp_compliance_documents d
		WHERE %s
		ORDER BY d.document_type, d.valid_until DESC NULLS LAST, d.created_at DESC
	`, whereClause), args...)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID).Msg("sp: failed to list compliance docs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list compliance documents"))
		return
	}
	defer rows.Close()

	docs := make([]models.SPComplianceDocument, 0)
	for rows.Next() {
		var d models.SPComplianceDocument
		if err := rows.Scan(
			&d.ID, &d.ProviderID, &d.OrgID, &d.DocumentType, &d.Title, &d.DocumentVersion,
			&d.UploadPath, &d.ValidFrom, &d.ValidUntil, &d.ReviewedBy, &d.ReviewNotes,
			&d.IsCurrent, &d.UploadedBy, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("sp: failed to scan compliance doc row")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read compliance documents"))
			return
		}
		docs = append(docs, d)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("sp: rows error listing compliance docs")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list compliance documents"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, docs))
}

// GetSPComplianceDoc returns a single compliance document.
// GET /api/v1/service-providers/:id/compliance-docs/:docId
func GetSPComplianceDoc(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	providerID := c.Param("id")
	docID := c.Param("docId")

	var d models.SPComplianceDocument
	err := database.DB.QueryRow(`
		SELECT id, provider_id, org_id, document_type, title, document_version,
		       upload_path, valid_from, valid_until, reviewed_by, review_notes,
		       is_current, uploaded_by, created_at, updated_at
		FROM sp_compliance_documents
		WHERE id = $1 AND provider_id = $2 AND org_id = $3
	`, docID, providerID, orgID).Scan(
		&d.ID, &d.ProviderID, &d.OrgID, &d.DocumentType, &d.Title, &d.DocumentVersion,
		&d.UploadPath, &d.ValidFrom, &d.ValidUntil, &d.ReviewedBy, &d.ReviewNotes,
		&d.IsCurrent, &d.UploadedBy, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Compliance document not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Str("doc_id", docID).Msg("sp: failed to get compliance doc")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to get compliance document"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, d))
}

// CreateSPComplianceDoc creates a compliance document for a service provider.
// POST /api/v1/service-providers/:id/compliance-docs
func CreateSPComplianceDoc(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	providerID := c.Param("id")
	userID := middleware.GetUserID(c)

	// Verify provider exists and belongs to this org.
	var exists bool
	if err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM service_providers WHERE id = $1 AND org_id = $2)", providerID, orgID,
	).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}

	var req models.CreateSPComplianceDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}
	if !models.IsValidSPDocType(req.DocumentType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid document_type; allowed: %s", strings.Join(models.ValidSPDocTypes, ", "))))
		return
	}

	var validFrom, validUntil *time.Time
	if req.ValidFrom != nil {
		t, err := time.Parse("2006-01-02", *req.ValidFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid valid_from format; use YYYY-MM-DD"))
			return
		}
		validFrom = &t
	}
	if req.ValidUntil != nil {
		t, err := time.Parse("2006-01-02", *req.ValidUntil)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid valid_until format; use YYYY-MM-DD"))
			return
		}
		validUntil = &t
	}
	if validFrom != nil && validUntil != nil && validUntil.Before(*validFrom) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "valid_until must be on or after valid_from"))
		return
	}

	// Validate upload_path to prevent path traversal (CWE-22).
	// Object storage paths must not contain ".." segments or absolute paths.
	if req.UploadPath != nil {
		cleaned := path.Clean("/" + strings.TrimLeft(*req.UploadPath, "/"))
		if strings.Contains(cleaned, "..") {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid upload_path: path traversal not allowed"))
			return
		}
		// Strip the leading "/" added for Clean; store the normalized relative path.
		normalized := strings.TrimPrefix(cleaned, "/")
		req.UploadPath = &normalized
	}

	var uploadedByPtr *string
	if userID != "" {
		uploadedByPtr = &userID
	}

	docID := uuid.New().String()
	var d models.SPComplianceDocument
	err := database.DB.QueryRow(`
		INSERT INTO sp_compliance_documents (
			id, provider_id, org_id, document_type, title, document_version,
			upload_path, valid_from, valid_until, review_notes, uploaded_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, provider_id, org_id, document_type, title, document_version,
		          upload_path, valid_from, valid_until, reviewed_by, review_notes,
		          is_current, uploaded_by, created_at, updated_at
	`,
		docID, providerID, orgID, req.DocumentType, req.Title, req.DocumentVersion,
		req.UploadPath, validFrom, validUntil, req.ReviewNotes, uploadedByPtr,
	).Scan(
		&d.ID, &d.ProviderID, &d.OrgID, &d.DocumentType, &d.Title, &d.DocumentVersion,
		&d.UploadPath, &d.ValidFrom, &d.ValidUntil, &d.ReviewedBy, &d.ReviewNotes,
		&d.IsCurrent, &d.UploadedBy, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID).Msg("sp: failed to insert compliance doc")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to create compliance document"))
		return
	}

	middleware.LogAudit(c, "sp_compliance_doc.uploaded", "sp_compliance_document", &d.ID, map[string]interface{}{
		"provider_id":   providerID,
		"document_type": d.DocumentType,
	})

	c.JSON(http.StatusCreated, successResponse(c, d))
}

// DeleteSPComplianceDoc deletes a compliance document.
// DELETE /api/v1/service-providers/:id/compliance-docs/:docId
func DeleteSPComplianceDoc(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	providerID := c.Param("id")
	docID := c.Param("docId")

	result, err := database.DB.Exec(
		"DELETE FROM sp_compliance_documents WHERE id = $1 AND provider_id = $2 AND org_id = $3",
		docID, providerID, orgID,
	)
	if err != nil {
		log.Error().Err(err).Str("doc_id", docID).Msg("sp: failed to delete compliance doc")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete compliance document"))
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Compliance document not found"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// =============================================================================
// Responsibility Matrix
// =============================================================================

// ListSPResponsibilities lists all responsibility matrix entries for a service provider.
// GET /api/v1/service-providers/:id/responsibilities
func ListSPResponsibilities(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	providerID := c.Param("id")

	var exists bool
	if err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM service_providers WHERE id = $1 AND org_id = $2)", providerID, orgID,
	).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}

	where := []string{"m.provider_id = $1", "m.org_id = $2"}
	args := []interface{}{providerID, orgID}
	argN := 3

	if v := c.Query("responsible_party"); v != "" {
		where = append(where, fmt.Sprintf("m.responsible_party = $%d", argN))
		args = append(args, v)
		argN++
	}

	whereClause := strings.Join(where, " AND ")
	rows, err := database.DB.Query(fmt.Sprintf(`
		SELECT m.id, m.provider_id, m.org_id, m.requirement_id, m.requirement_code,
		       m.responsible_party, m.notes, m.created_by, m.created_at, m.updated_at
		FROM sp_responsibility_matrix m
		WHERE %s
		ORDER BY m.requirement_code ASC
	`, whereClause), args...)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID).Msg("sp: failed to list responsibility matrix")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list responsibility matrix"))
		return
	}
	defer rows.Close()

	entries := make([]models.SPResponsibilityMatrix, 0)
	for rows.Next() {
		var m models.SPResponsibilityMatrix
		if err := rows.Scan(
			&m.ID, &m.ProviderID, &m.OrgID, &m.RequirementID, &m.RequirementCode,
			&m.ResponsibleParty, &m.Notes, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			log.Error().Err(err).Msg("sp: failed to scan responsibility row")
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to read responsibility matrix"))
			return
		}
		entries = append(entries, m)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("sp: rows error listing responsibilities")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to list responsibility matrix"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, entries))
}

// UpsertSPResponsibility creates or updates a responsibility matrix entry.
// PUT /api/v1/service-providers/:id/responsibility-matrix/:reqCode
func UpsertSPResponsibility(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	providerID := c.Param("id")
	reqCode := c.Param("reqCode")
	userID := middleware.GetUserID(c)

	var exists bool
	if err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM service_providers WHERE id = $1 AND org_id = $2)", providerID, orgID,
	).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Service provider not found"))
		return
	}

	var req models.UpsertResponsibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body: "+err.Error()))
		return
	}
	if !models.IsValidResponsibleParty(req.ResponsibleParty) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR",
			fmt.Sprintf("invalid responsible_party; allowed: %s", strings.Join(models.ValidResponsibleParties, ", "))))
		return
	}

	var createdByPtr *string
	if userID != "" {
		createdByPtr = &userID
	}

	id := uuid.New().String()
	var m models.SPResponsibilityMatrix
	err := database.DB.QueryRow(`
		INSERT INTO sp_responsibility_matrix (
			id, provider_id, org_id, requirement_id, requirement_code, responsible_party, notes, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (provider_id, requirement_code) DO UPDATE SET
			requirement_id    = EXCLUDED.requirement_id,
			responsible_party = EXCLUDED.responsible_party,
			notes             = EXCLUDED.notes,
			updated_at        = NOW()
		RETURNING id, provider_id, org_id, requirement_id, requirement_code,
		          responsible_party, notes, created_by, created_at, updated_at
	`, id, providerID, orgID, req.RequirementID, reqCode, req.ResponsibleParty, req.Notes, createdByPtr,
	).Scan(
		&m.ID, &m.ProviderID, &m.OrgID, &m.RequirementID, &m.RequirementCode,
		&m.ResponsibleParty, &m.Notes, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID).Str("req_code", reqCode).Msg("sp: failed to upsert responsibility")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to save responsibility matrix entry"))
		return
	}

	middleware.LogAudit(c, "sp_responsibility.updated", "sp_responsibility_matrix", &m.ID, map[string]interface{}{
		"provider_id":       providerID,
		"requirement_code":  reqCode,
		"responsible_party": m.ResponsibleParty,
	})

	c.JSON(http.StatusOK, successResponse(c, m))
}

// DeleteSPResponsibility removes a responsibility matrix entry by requirement code.
// DELETE /api/v1/service-providers/:id/responsibility-matrix/:reqCode
func DeleteSPResponsibility(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	providerID := c.Param("id")
	reqCode := c.Param("reqCode")

	result, err := database.DB.Exec(
		"DELETE FROM sp_responsibility_matrix WHERE provider_id = $1 AND org_id = $2 AND requirement_code = $3",
		providerID, orgID, reqCode,
	)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID).Str("req_code", reqCode).Msg("sp: failed to delete responsibility")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Failed to delete responsibility matrix entry"))
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Responsibility matrix entry not found"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
