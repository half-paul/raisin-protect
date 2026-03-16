package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
)

// isDocTerminalStatus returns true when the document cannot be edited.
func isDocTerminalStatus(status string) bool {
	return status == models.DocStatusFinal ||
		status == models.DocStatusSigned ||
		status == models.DocStatusCancelled
}

// roleAllowed reports whether role is contained in the allowed slice.
func roleAllowed(role string, allowed []string) bool {
	for _, r := range allowed {
		if r == role {
			return true
		}
	}
	return false
}

// getDocPaginationParams extracts page / per_page from query params.
func getDocPaginationParams(c *gin.Context) (page, perPage, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset = (page - 1) * perPage
	return
}

// scanComplianceDoc scans all 27 compliance_documents columns into doc.
func scanComplianceDoc(scanner interface {
	Scan(dest ...interface{}) error
}, doc *models.ComplianceDocument) error {
	return scanner.Scan(
		&doc.ID, &doc.OrgID, &doc.TemplateID, &doc.DocumentType, &doc.Title,
		&doc.AssessmentPeriodStart, &doc.AssessmentPeriodEnd, &doc.PCIDSSVersion,
		&doc.MerchantName, &doc.MerchantDBA, &doc.MerchantURL, &doc.BusinessType,
		&doc.QSAName, &doc.QSACompany, &doc.QSASignatureDate,
		&doc.DocStatus, &doc.GeneratedBy, &doc.DataSnapshot, &doc.PDFPath, &doc.FileSizeBytes,
		&doc.GeneratedAt, &doc.GenerationError, &doc.Version, &doc.ParentID,
		&doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
}

const selectDocCols = `id, org_id, template_id, document_type, title,
		assessment_period_start, assessment_period_end, pci_dss_version,
		merchant_name, merchant_dba, merchant_url, business_type,
		qsa_name, qsa_company, qsa_signature_date,
		doc_status, generated_by, data_snapshot, pdf_path, file_size_bytes,
		generated_at, generation_error, version, parent_id,
		created_by, created_at, updated_at`

// =============================================================================
// List Compliance Documents
// =============================================================================

// ListComplianceDocuments handles GET /api/v1/documents.
// Returns paginated AOC/ROC compliance documents for the org.
func ListComplianceDocuments(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	page, perPage, offset := getDocPaginationParams(c)

	countArgs := []interface{}{orgID}
	listArgs := []interface{}{orgID}
	where := "WHERE org_id = $1"
	argN := 2

	if status := c.Query("status"); status != "" {
		where += fmt.Sprintf(" AND doc_status = $%d", argN)
		countArgs = append(countArgs, status)
		listArgs = append(listArgs, status)
		argN++
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM compliance_documents " + where
	if err := database.DB.QueryRow(countSQL, countArgs...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	listArgs = append(listArgs, perPage, offset)
	listSQL := fmt.Sprintf(
		"SELECT %s FROM compliance_documents %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		selectDocCols, where, argN, argN+1,
	)

	rows, err := database.DB.Query(listSQL, listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	docs := []models.ComplianceDocument{}
	for rows.Next() {
		var doc models.ComplianceDocument
		if err := scanComplianceDoc(rows, &doc); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
			return
		}
		docs = append(docs, doc)
	}

	c.JSON(http.StatusOK, listResponse(c, docs, total, page, perPage))
}

// =============================================================================
// Create Compliance Document
// =============================================================================

// CreateComplianceDocument handles POST /api/v1/documents.
// Creates a new AOC/ROC document in draft status.
// Requires DocumentCreateRoles (compliance_manager, ciso).
func CreateComplianceDocument(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

	if !roleAllowed(userRole, models.DocumentCreateRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Insufficient permissions"))
		return
	}

	var req models.CreateComplianceDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	if !models.IsValidDocumentType(req.DocumentType) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid document_type"))
		return
	}

	start, err := time.Parse(time.RFC3339, req.AssessmentPeriodStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid assessment_period_start"))
		return
	}
	end, err := time.Parse(time.RFC3339, req.AssessmentPeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid assessment_period_end"))
		return
	}
	if !end.After(start) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "assessment_period_end must be after assessment_period_start"))
		return
	}

	pciDSSVersion := "4.0.1"
	if req.PCIDSSVersion != nil && *req.PCIDSSVersion != "" {
		pciDSSVersion = *req.PCIDSSVersion
	}

	var qsaSignatureDate *time.Time
	if req.QSASignatureDate != nil {
		t, err := time.Parse(time.RFC3339, *req.QSASignatureDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid qsa_signature_date"))
			return
		}
		qsaSignatureDate = &t
	}

	id := uuid.New().String()

	const insertSQL = `
		INSERT INTO compliance_documents (
			id, org_id, template_id, document_type, title,
			assessment_period_start, assessment_period_end, pci_dss_version,
			merchant_name, merchant_dba, merchant_url, business_type,
			qsa_name, qsa_company, qsa_signature_date,
			doc_status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING ` + selectDocCols

	row := database.DB.QueryRow(insertSQL,
		id, orgID, req.TemplateID, req.DocumentType, req.Title,
		start, end, pciDSSVersion,
		req.MerchantName, req.MerchantDBA, req.MerchantURL, req.BusinessType,
		req.QSAName, req.QSACompany, qsaSignatureDate,
		models.DocStatusDraft, userID,
	)

	var doc models.ComplianceDocument
	if err := scanComplianceDoc(row, &doc); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(c, doc))
}

// =============================================================================
// Get Compliance Document
// =============================================================================

// GetComplianceDocument handles GET /api/v1/documents/:id.
// Returns a single compliance document scoped to the requesting org.
func GetComplianceDocument(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	docID := c.Param("id")

	row := database.DB.QueryRow(
		"SELECT "+selectDocCols+" FROM compliance_documents WHERE id = $1 AND org_id = $2",
		docID, orgID,
	)

	var doc models.ComplianceDocument
	if err := scanComplianceDoc(row, &doc); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "compliance document not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, doc))
}

// =============================================================================
// Update Compliance Document
// =============================================================================

// UpdateComplianceDocument handles PUT /api/v1/documents/:id.
// Updates editable fields. Blocked for terminal states (final, signed, cancelled).
func UpdateComplianceDocument(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	docID := c.Param("id")

	// Check current status first.
	var currentStatus string
	err := database.DB.QueryRow(
		"SELECT doc_status FROM compliance_documents WHERE id = $1 AND org_id = $2",
		docID, orgID,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "compliance document not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}
	if isDocTerminalStatus(currentStatus) {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "document is in a terminal state and cannot be edited"))
		return
	}

	var req models.UpdateComplianceDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argN))
		args = append(args, *req.Title)
		argN++
	}
	if req.MerchantName != nil {
		setClauses = append(setClauses, fmt.Sprintf("merchant_name = $%d", argN))
		args = append(args, *req.MerchantName)
		argN++
	}
	if req.MerchantDBA != nil {
		setClauses = append(setClauses, fmt.Sprintf("merchant_dba = $%d", argN))
		args = append(args, *req.MerchantDBA)
		argN++
	}
	if req.MerchantURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("merchant_url = $%d", argN))
		args = append(args, *req.MerchantURL)
		argN++
	}
	if req.BusinessType != nil {
		setClauses = append(setClauses, fmt.Sprintf("business_type = $%d", argN))
		args = append(args, *req.BusinessType)
		argN++
	}
	if req.QSAName != nil {
		setClauses = append(setClauses, fmt.Sprintf("qsa_name = $%d", argN))
		args = append(args, *req.QSAName)
		argN++
	}
	if req.QSACompany != nil {
		setClauses = append(setClauses, fmt.Sprintf("qsa_company = $%d", argN))
		args = append(args, *req.QSACompany)
		argN++
	}
	if req.QSASignatureDate != nil {
		t, err := time.Parse(time.RFC3339, *req.QSASignatureDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid qsa_signature_date"))
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("qsa_signature_date = $%d", argN))
		args = append(args, t)
		argN++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "no fields to update"))
		return
	}

	// Append WHERE args.
	args = append(args, docID, orgID)
	whereN := argN
	updateSQL := fmt.Sprintf(
		"UPDATE compliance_documents SET %s WHERE id = $%d AND org_id = $%d RETURNING %s",
		join(setClauses, ", "), whereN, whereN+1, selectDocCols,
	)

	row := database.DB.QueryRow(updateSQL, args...)
	var doc models.ComplianceDocument
	if err := scanComplianceDoc(row, &doc); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "compliance document not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, doc))
}

// =============================================================================
// Generate Document
// =============================================================================

// GenerateDocument handles POST /api/v1/documents/:id/generate.
// Transitions draft → generating and queues an async generation job.
// Returns 202 Accepted.
func GenerateDocument(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	docID := c.Param("id")

	var currentStatus string
	err := database.DB.QueryRow(
		"SELECT doc_status FROM compliance_documents WHERE id = $1 AND org_id = $2",
		docID, orgID,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "compliance document not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	if currentStatus != models.DocStatusDraft {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "document must be in draft status to trigger generation"))
		return
	}

	_, err = database.DB.Exec(
		"UPDATE compliance_documents SET doc_status = $1, generated_by = $2 WHERE id = $3 AND org_id = $4",
		models.DocStatusGenerating, userID, docID, orgID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusAccepted, successResponse(c, gin.H{"doc_status": models.DocStatusGenerating}))
}

// =============================================================================
// Finalize Document
// =============================================================================

// FinalizeDocument handles POST /api/v1/documents/:id/finalize.
// Transitions approved → final and triggers PDF rendering.
// CISO-only (DocumentFinalizeRoles). Returns 202 Accepted.
func FinalizeDocument(c *gin.Context) {
	userRole := middleware.GetUserRole(c)

	// Internal CISO-only check — not enforced at the route level.
	if !roleAllowed(userRole, models.DocumentFinalizeRoles) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "only CISO may finalize documents"))
		return
	}

	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	docID := c.Param("id")

	var currentStatus string
	err := database.DB.QueryRow(
		"SELECT doc_status FROM compliance_documents WHERE id = $1 AND org_id = $2",
		docID, orgID,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "compliance document not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	if currentStatus != models.DocStatusApproved {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "document must be in approved status to finalize"))
		return
	}

	_, err = database.DB.Exec(
		"UPDATE compliance_documents SET doc_status = $1, generated_by = $2 WHERE id = $3 AND org_id = $4",
		models.DocStatusFinal, userID, docID, orgID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusAccepted, successResponse(c, gin.H{"doc_status": models.DocStatusFinal}))
}

// =============================================================================
// Document Sections
// =============================================================================

// GetDocumentSection handles GET /api/v1/documents/:id/sections/:key.
// Returns a single section within a compliance document.
func GetDocumentSection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	docID := c.Param("id")
	sectionKey := c.Param("key")

	row := database.DB.QueryRow(`
		SELECT id, document_id, org_id, section_key, title,
			content, compliance_status, evidence_ids, sort_order,
			created_at, updated_at
		FROM document_sections
		WHERE document_id = $1 AND org_id = $2 AND section_key = $3`,
		docID, orgID, sectionKey,
	)

	var sec models.DocumentSection
	if err := row.Scan(
		&sec.ID, &sec.DocumentID, &sec.OrgID, &sec.SectionKey, &sec.Title,
		&sec.Content, &sec.ComplianceStatus, pq.Array(&sec.EvidenceIDs), &sec.SortOrder,
		&sec.CreatedAt, &sec.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "section not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, sec))
}

// UpsertDocumentSection handles PUT /api/v1/documents/:id/sections/:key.
// Creates or updates a section. Blocked for finalized/signed/cancelled documents.
func UpsertDocumentSection(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	docID := c.Param("id")
	sectionKey := c.Param("key")

	// Verify document exists and is editable.
	var currentStatus string
	err := database.DB.QueryRow(
		"SELECT doc_status FROM compliance_documents WHERE id = $1 AND org_id = $2",
		docID, orgID,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "compliance document not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}
	if isDocTerminalStatus(currentStatus) {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "document is in a terminal state and cannot be edited"))
		return
	}

	var req models.UpsertDocumentSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	evidenceIDs := req.EvidenceIDs
	if evidenceIDs == nil {
		evidenceIDs = []string{}
	}

	id := uuid.New().String()

	const upsertSQL = `
		INSERT INTO document_sections (id, document_id, org_id, section_key, title,
			content, compliance_status, evidence_ids, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (document_id, section_key) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			compliance_status = EXCLUDED.compliance_status,
			evidence_ids = EXCLUDED.evidence_ids,
			sort_order = EXCLUDED.sort_order,
			updated_at = NOW()
		RETURNING id, document_id, org_id, section_key, title,
			content, compliance_status, evidence_ids, sort_order,
			created_at, updated_at`

	row := database.DB.QueryRow(upsertSQL,
		id, docID, orgID, sectionKey, req.Title,
		req.Content, req.ComplianceStatus, pq.Array(evidenceIDs), sortOrder,
	)

	var sec models.DocumentSection
	if err := row.Scan(
		&sec.ID, &sec.DocumentID, &sec.OrgID, &sec.SectionKey, &sec.Title,
		&sec.Content, &sec.ComplianceStatus, pq.Array(&sec.EvidenceIDs), &sec.SortOrder,
		&sec.CreatedAt, &sec.UpdatedAt,
	); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, sec))
}

// =============================================================================
// Document Attestations
// =============================================================================

// GetDocumentAttestations handles GET /api/v1/documents/:id/attestations.
// Returns all signatory attestations for a document.
func GetDocumentAttestations(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	docID := c.Param("id")

	rows, err := database.DB.Query(`
		SELECT id, document_id, org_id, attestation_role, full_name, title,
			company_name, company_address, company_url, email, phone,
			qsa_company, qsa_number, signed_at, signature_method, signature_ref,
			created_by, created_at, updated_at
		FROM document_attestations
		WHERE document_id = $1 AND org_id = $2
		ORDER BY attestation_role`,
		docID, orgID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	attestations := []models.DocumentAttestation{}
	for rows.Next() {
		var a models.DocumentAttestation
		if err := rows.Scan(
			&a.ID, &a.DocumentID, &a.OrgID, &a.AttestationRole, &a.FullName, &a.Title,
			&a.CompanyName, &a.CompanyAddress, &a.CompanyURL, &a.Email, &a.Phone,
			&a.QSACompany, &a.QSANumber, &a.SignedAt, &a.SignatureMethod, &a.SignatureRef,
			&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
			return
		}
		attestations = append(attestations, a)
	}

	c.JSON(http.StatusOK, successResponse(c, attestations))
}

// UpsertDocumentAttestation handles PUT /api/v1/documents/:id/attestations/:role.
// Creates or updates an attestation. Validates attestation_role.
func UpsertDocumentAttestation(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	docID := c.Param("id")
	role := c.Param("role")

	if !attestationRoleAllowed(role) {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid attestation_role"))
		return
	}

	// Verify document exists and is editable.
	var currentStatus string
	err := database.DB.QueryRow(
		"SELECT doc_status FROM compliance_documents WHERE id = $1 AND org_id = $2",
		docID, orgID,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "compliance document not found"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}
	if isDocTerminalStatus(currentStatus) {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "document is in a terminal state"))
		return
	}

	var req models.UpsertAttestationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	var signedAt *time.Time
	if req.SignedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.SignedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid signed_at"))
			return
		}
		signedAt = &t
	}

	id := uuid.New().String()

	const upsertSQL = `
		INSERT INTO document_attestations (
			id, document_id, org_id, attestation_role, full_name, title,
			company_name, company_address, company_url, email, phone,
			qsa_company, qsa_number, signed_at, signature_method, signature_ref,
			created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (document_id, attestation_role) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			title = EXCLUDED.title,
			company_name = EXCLUDED.company_name,
			company_address = EXCLUDED.company_address,
			company_url = EXCLUDED.company_url,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			qsa_company = EXCLUDED.qsa_company,
			qsa_number = EXCLUDED.qsa_number,
			signed_at = EXCLUDED.signed_at,
			signature_method = EXCLUDED.signature_method,
			signature_ref = EXCLUDED.signature_ref,
			updated_at = NOW()
		RETURNING id, document_id, org_id, attestation_role, full_name, title,
			company_name, company_address, company_url, email, phone,
			qsa_company, qsa_number, signed_at, signature_method, signature_ref,
			created_by, created_at, updated_at`

	row := database.DB.QueryRow(upsertSQL,
		id, docID, orgID, role, req.FullName, req.Title,
		req.CompanyName, req.CompanyAddress, req.CompanyURL, req.Email, req.Phone,
		req.QSACompany, req.QSANumber, signedAt, req.SignatureMethod, req.SignatureRef,
		userID,
	)

	var a models.DocumentAttestation
	if err := row.Scan(
		&a.ID, &a.DocumentID, &a.OrgID, &a.AttestationRole, &a.FullName, &a.Title,
		&a.CompanyName, &a.CompanyAddress, &a.CompanyURL, &a.Email, &a.Phone,
		&a.QSACompany, &a.QSANumber, &a.SignedAt, &a.SignatureMethod, &a.SignatureRef,
		&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, successResponse(c, a))
}

// attestationRoleAllowed returns true if role is a valid attestation role.
func attestationRoleAllowed(role string) bool {
	for _, r := range models.ValidAttestationRoles {
		if r == role {
			return true
		}
	}
	return false
}

// =============================================================================
// Document Requirements Snapshot
// =============================================================================

// GetDocumentRequirements handles GET /api/v1/documents/:id/requirements.
// Returns the requirement compliance snapshot captured at document generation time.
func GetDocumentRequirements(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	docID := c.Param("id")

	rows, err := database.DB.Query(`
		SELECT id, document_id, requirement_id, requirement_code, requirement_title,
			in_scope, control_count, passing_controls, evidence_count,
			status, notes, snapshotted_at
		FROM document_requirement_snapshots
		WHERE document_id = $1 AND org_id = $2
		ORDER BY requirement_code`,
		docID, orgID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	snapshots := []models.DocumentRequirementSnapshot{}
	for rows.Next() {
		var s models.DocumentRequirementSnapshot
		if err := rows.Scan(
			&s.ID, &s.DocumentID, &s.RequirementID, &s.RequirementCode, &s.RequirementTitle,
			&s.InScope, &s.ControlCount, &s.PassingControls, &s.EvidenceCount,
			&s.Status, &s.Notes, &s.SnapshottedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "Internal server error"))
			return
		}
		snapshots = append(snapshots, s)
	}

	c.JSON(http.StatusOK, successResponse(c, snapshots))
}
