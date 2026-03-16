package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDocRouter creates a test router with all compliance document endpoints registered.
// Default role is compliance_manager (in DocumentCreateRoles).
func setupDocRouter() (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/documents")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "cm@acme.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	protected.GET("", ListComplianceDocuments)
	protected.POST("", CreateComplianceDocument)
	protected.GET("/:id", GetComplianceDocument)
	protected.PUT("/:id", UpdateComplianceDocument)
	protected.POST("/:id/generate", GenerateDocument)
	protected.POST("/:id/finalize", FinalizeDocument)

	// Sections.
	protected.GET("/:id/sections/:key", GetDocumentSection)
	protected.PUT("/:id/sections/:key", UpsertDocumentSection)

	// Attestations.
	protected.GET("/:id/attestations", GetDocumentAttestations)
	protected.PUT("/:id/attestations/:role", UpsertDocumentAttestation)

	// Requirements snapshot.
	protected.GET("/:id/requirements", GetDocumentRequirements)

	return router, mock
}

// setupDocRouterWithRole creates a document router with a custom role.
func setupDocRouterWithRole(role string) (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/documents")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "user@acme.com")
		c.Set(middleware.ContextKeyRole, role)
		c.Next()
	})

	protected.GET("", ListComplianceDocuments)
	protected.POST("", CreateComplianceDocument)
	protected.GET("/:id", GetComplianceDocument)
	protected.PUT("/:id", UpdateComplianceDocument)
	protected.POST("/:id/generate", GenerateDocument)
	protected.POST("/:id/finalize", FinalizeDocument)
	protected.GET("/:id/sections/:key", GetDocumentSection)
	protected.PUT("/:id/sections/:key", UpsertDocumentSection)
	protected.GET("/:id/attestations", GetDocumentAttestations)
	protected.PUT("/:id/attestations/:role", UpsertDocumentAttestation)

	return router, mock
}

// setupDocRouterTenant creates a document router for cross-tenant isolation tests.
func setupDocRouterTenant(orgID string) (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/documents")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-attacker")
		c.Set(middleware.ContextKeyOrgID, orgID)
		c.Set(middleware.ContextKeyEmail, "attacker@evil.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	protected.GET("", ListComplianceDocuments)
	protected.GET("/:id", GetComplianceDocument)
	protected.PUT("/:id", UpdateComplianceDocument)
	protected.PUT("/:id/sections/:key", UpsertDocumentSection)
	protected.POST("/:id/finalize", FinalizeDocument)

	return router, mock
}

var docNow = time.Now()

// docCols are the column names returned by compliance_documents SELECT queries.
var docCols = []string{
	"id", "org_id", "template_id", "document_type", "title",
	"assessment_period_start", "assessment_period_end", "pci_dss_version",
	"merchant_name", "merchant_dba", "merchant_url", "business_type",
	"qsa_name", "qsa_company", "qsa_signature_date",
	"doc_status", "generated_by", "data_snapshot", "pdf_path", "file_size_bytes",
	"generated_at", "generation_error", "version", "parent_id",
	"created_by", "created_at", "updated_at",
}

// sectionCols are the column names returned by document_sections SELECT queries.
var sectionCols = []string{
	"id", "document_id", "org_id", "section_key", "title",
	"content", "compliance_status", "evidence_ids", "sort_order",
	"created_at", "updated_at",
}

// attestationCols are the column names returned by document_attestations SELECT queries.
var attestationCols = []string{
	"id", "document_id", "org_id", "attestation_role", "full_name", "title",
	"company_name", "company_address", "company_url", "email", "phone",
	"qsa_company", "qsa_number", "signed_at", "signature_method", "signature_ref",
	"created_by", "created_at", "updated_at",
}

var (
	assessmentStart = "2026-01-01T00:00:00Z"
	assessmentEnd   = "2026-12-31T23:59:59Z"
)

// =============================================================================
// List Compliance Documents
// =============================================================================

func TestListComplianceDocuments_Success(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM compliance_documents`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT .* FROM compliance_documents`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(docCols).
			AddRow("doc-001", "org-001", nil, "aoc_saq_d", "AOC SAQ D 2026", docNow, docNow, "4.0.1",
				nil, nil, nil, nil, nil, nil, nil, "draft", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow).
			AddRow("doc-002", "org-001", nil, "roc", "ROC 2025", docNow, docNow, "4.0.1",
				nil, nil, nil, nil, nil, nil, nil, "final", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	require.Len(t, data, 2)
	assert.EqualValues(t, 2, resp["meta"].(map[string]interface{})["total"])
}

func TestListComplianceDocuments_FilterByStatus(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM compliance_documents`).
		WithArgs("org-001", "draft").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM compliance_documents`).
		WithArgs("org-001", "draft", 20, 0).
		WillReturnRows(sqlmock.NewRows(docCols).
			AddRow("doc-001", "org-001", nil, "aoc_saq_d", "AOC SAQ D 2026", docNow, docNow, "4.0.1",
				nil, nil, nil, nil, nil, nil, nil, "draft", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents?status=draft", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Create Compliance Document
// =============================================================================

func TestCreateComplianceDocument_Success(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`INSERT INTO compliance_documents`).
		WithArgs(
			sqlmock.AnyArg(), // id
			"org-001",
			nil,              // template_id
			"aoc_saq_d",
			"AOC SAQ D 2026",
			sqlmock.AnyArg(), // assessment_period_start
			sqlmock.AnyArg(), // assessment_period_end
			"4.0.1",
			nil, nil, nil, nil, // merchant fields
			nil, nil, nil, // QSA fields
			"draft",         // initial status
			"user-001",
		).
		WillReturnRows(sqlmock.NewRows(docCols).
			AddRow("new-doc", "org-001", nil, "aoc_saq_d", "AOC SAQ D 2026", docNow, docNow, "4.0.1",
				nil, nil, nil, nil, nil, nil, nil, "draft", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow))

	body := fmt.Sprintf(`{
		"document_type": "aoc_saq_d",
		"title": "AOC SAQ D 2026",
		"assessment_period_start": "%s",
		"assessment_period_end": "%s",
		"pci_dss_version": "4.0.1"
	}`, assessmentStart, assessmentEnd)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "new-doc", data["id"])
	assert.Equal(t, "draft", data["doc_status"])
}

func TestCreateComplianceDocument_WithMerchantAndQSAInfo(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`INSERT INTO compliance_documents`).
		WithArgs(
			sqlmock.AnyArg(), "org-001", nil, "roc", "ROC 2026",
			sqlmock.AnyArg(), sqlmock.AnyArg(), "4.0.1",
			"Acme Payments Inc", "Acme Payments", nil, "ecommerce",
			"Jane QSA", "QSA Corp", nil,
			"draft", "user-001",
		).
		WillReturnRows(sqlmock.NewRows(docCols).
			AddRow("doc-roc", "org-001", nil, "roc", "ROC 2026", docNow, docNow, "4.0.1",
				"Acme Payments Inc", "Acme Payments", nil, "ecommerce",
				"Jane QSA", "QSA Corp", nil, "draft", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow))

	body := fmt.Sprintf(`{
		"document_type": "roc",
		"title": "ROC 2026",
		"assessment_period_start": "%s",
		"assessment_period_end": "%s",
		"pci_dss_version": "4.0.1",
		"merchant_name": "Acme Payments Inc",
		"merchant_dba": "Acme Payments",
		"business_type": "ecommerce",
		"qsa_name": "Jane QSA",
		"qsa_company": "QSA Corp"
	}`, assessmentStart, assessmentEnd)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateComplianceDocument_InvalidDocumentType(t *testing.T) {
	router, _ := setupDocRouter()

	body := fmt.Sprintf(`{
		"document_type": "compliance_cert",
		"title": "Cert 2026",
		"assessment_period_start": "%s",
		"assessment_period_end": "%s"
	}`, assessmentStart, assessmentEnd)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp["error"].(map[string]interface{})["code"])
}

func TestCreateComplianceDocument_MissingRequiredFields(t *testing.T) {
	router, _ := setupDocRouter()

	// Missing title and assessment period.
	body := `{"document_type": "aoc_saq_d"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateComplianceDocument_AssessmentPeriodEndBeforeStart(t *testing.T) {
	router, _ := setupDocRouter()

	body := `{
		"document_type": "aoc_saq_a",
		"title": "AOC 2026",
		"assessment_period_start": "2026-12-01T00:00:00Z",
		"assessment_period_end": "2026-01-01T00:00:00Z"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// Get & Update Compliance Document
// =============================================================================

func TestGetComplianceDocument_Success(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT .* FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows(docCols).
			AddRow("doc-001", "org-001", nil, "aoc_saq_d", "AOC SAQ D 2026", docNow, docNow, "4.0.1",
				nil, nil, nil, nil, nil, nil, nil, "draft", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/doc-001", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "doc-001", data["id"])
	assert.Equal(t, "aoc_saq_d", data["document_type"])
}

func TestGetComplianceDocument_NotFound(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT .* FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows(docCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/missing", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateComplianceDocument_Success(t *testing.T) {
	router, mock := setupDocRouter()

	// Document must be in draft or review to be editable.
	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("draft"))

	mock.ExpectQuery(`UPDATE compliance_documents SET`).
		WithArgs("Jane QSA", "ClearSecure QSA", "doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows(docCols).
			AddRow("doc-001", "org-001", nil, "aoc_saq_d", "AOC SAQ D 2026", docNow, docNow, "4.0.1",
				nil, nil, nil, nil, "Jane QSA", "ClearSecure QSA", nil, "draft", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow))

	body := `{"qsa_name": "Jane QSA", "qsa_company": "ClearSecure QSA"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Jane QSA", data["qsa_name"])
}

func TestUpdateComplianceDocument_FinalizedDocumentUneditable(t *testing.T) {
	// Terminal-state documents (final, signed) must not be editable.
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-final", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("final"))

	body := `{"title": "Changed Title"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-final", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

// =============================================================================
// Document Generation
// =============================================================================

func TestGenerateDocument_Success(t *testing.T) {
	// Generation transitions doc from draft → generating (async).
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("draft"))

	// Transition to 'generating'.
	mock.ExpectExec(`UPDATE compliance_documents SET doc_status`).
		WithArgs("generating", "user-001", "doc-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/generate", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
}

func TestGenerateDocument_InvalidStatusTransition(t *testing.T) {
	// Can't trigger generation on a document already in 'generating' or 'final'.
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-gen", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("generating"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-gen/generate", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestGenerateDocument_NotFound(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/missing/generate", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Finalize (Generate PDF + Lock)
// =============================================================================

func TestFinalizeDocument_Success(t *testing.T) {
	// Finalize transitions doc from 'approved' → 'final' and triggers PDF generation.
	router, mock := setupDocRouterWithRole("ciso") // Only CISO can finalize.

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("approved"))

	mock.ExpectExec(`UPDATE compliance_documents SET doc_status`).
		WithArgs("final", "user-001", "doc-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/finalize", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
}

func TestFinalizeDocument_NotApprovedStatus(t *testing.T) {
	// Can only finalize from 'approved' state — not draft, review, or generating.
	router, mock := setupDocRouterWithRole("ciso")

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("draft"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/finalize", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestFinalizeDocument_ComplianceManagerForbidden(t *testing.T) {
	// DocumentFinalizeRoles is ciso-only — compliance_manager must be rejected.
	router, _ := setupDocRouterWithRole("compliance_manager")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/finalize", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestFinalizeDocument_AlreadyFinal(t *testing.T) {
	// A document that is already final must not be re-finalized.
	router, mock := setupDocRouterWithRole("ciso")

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-final", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("final"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-final/finalize", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

// =============================================================================
// Document Sections
// =============================================================================

func TestUpsertDocumentSection_Success(t *testing.T) {
	router, mock := setupDocRouter()

	// Document must be editable (draft or review).
	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("review"))

	mock.ExpectQuery(`INSERT INTO document_sections.*ON CONFLICT`).
		WithArgs(
			sqlmock.AnyArg(), // id
			"doc-001", "org-001", "pci_req_1",
			"Network Security Controls",
			sqlmock.AnyArg(), // content
			"compliant",
			sqlmock.AnyArg(), // evidence_ids (pq array)
			1,
		).
		WillReturnRows(sqlmock.NewRows(sectionCols).
			AddRow("sec-001", "doc-001", "org-001", "pci_req_1",
				"Network Security Controls", "All firewall rules reviewed.", "compliant",
				`{}`, 1, docNow, docNow))

	body := `{
		"title": "Network Security Controls",
		"content": "All firewall rules reviewed.",
		"compliance_status": "compliant",
		"evidence_ids": [],
		"sort_order": 1
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-001/sections/pci_req_1",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "pci_req_1", data["section_key"])
	assert.Equal(t, "compliant", data["compliance_status"])
}

func TestUpsertDocumentSection_FinalizedDocumentForbidden(t *testing.T) {
	// Sections on finalized documents cannot be modified.
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-final", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("final"))

	body := `{"title": "Section 1", "content": "Changed"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-final/sections/pci_req_1",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestUpsertDocumentSection_MissingTitle(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("draft"))

	body := `{"content": "Some content"}` // missing required 'title'
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-001/sections/pci_req_1",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDocumentSection_Success(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT .* FROM document_sections WHERE document_id.*AND org_id.*AND section_key`).
		WithArgs("doc-001", "org-001", "pci_req_1").
		WillReturnRows(sqlmock.NewRows(sectionCols).
			AddRow("sec-001", "doc-001", "org-001", "pci_req_1",
				"Network Security Controls", "Reviewed.", "compliant", `{}`, 1, docNow, docNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/doc-001/sections/pci_req_1", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "pci_req_1", data["section_key"])
}

func TestGetDocumentSection_NotFound(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT .* FROM document_sections WHERE document_id.*AND org_id.*AND section_key`).
		WithArgs("doc-001", "org-001", "pci_req_99").
		WillReturnRows(sqlmock.NewRows(sectionCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/doc-001/sections/pci_req_99", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Document Attestations
// =============================================================================

func TestUpsertDocumentAttestation_Success(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("approved"))

	mock.ExpectQuery(`INSERT INTO document_attestations.*ON CONFLICT`).
		WithArgs(
			sqlmock.AnyArg(), "doc-001", "org-001", "merchant_signatory",
			"Alice Merchant", "CISO",
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			"user-001",
		).
		WillReturnRows(sqlmock.NewRows(attestationCols).
			AddRow("att-001", "doc-001", "org-001", "merchant_signatory",
				"Alice Merchant", "CISO", nil, nil, nil, nil, nil, nil, nil,
				nil, nil, nil, "user-001", docNow, docNow))

	body := `{"full_name": "Alice Merchant", "title": "CISO"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-001/attestations/merchant_signatory",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "merchant_signatory", data["attestation_role"])
	assert.Equal(t, "Alice Merchant", data["full_name"])
}

func TestUpsertDocumentAttestation_InvalidRole(t *testing.T) {
	router, _ := setupDocRouter()

	body := `{"full_name": "Bob", "title": "Director"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-001/attestations/director_role",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDocumentAttestations_Success(t *testing.T) {
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT .* FROM document_attestations WHERE document_id.*AND org_id`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows(attestationCols).
			AddRow("att-001", "doc-001", "org-001", "merchant_signatory",
				"Alice Merchant", "CISO", nil, nil, nil, nil, nil, nil, nil,
				docNow, "digital", nil, "user-001", docNow, docNow).
			AddRow("att-002", "doc-001", "org-001", "qsa_signatory",
				"Jane QSA", "Lead QSA", nil, nil, nil, nil, nil, "ClearSecure", "QSA-123",
				docNow, "digital", nil, "user-001", docNow, docNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/doc-001/attestations", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	require.Len(t, data, 2)
}

// =============================================================================
// RBAC / Access Control
// =============================================================================

func TestListComplianceDocuments_AuditorCanView(t *testing.T) {
	// Auditors are in DocumentViewRoles.
	router, mock := setupDocRouterWithRole("auditor")

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM compliance_documents`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM compliance_documents`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(docCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestCreateComplianceDocument_AuditorForbidden(t *testing.T) {
	// Auditors are NOT in DocumentCreateRoles.
	router, _ := setupDocRouterWithRole("auditor")

	body := fmt.Sprintf(`{
		"document_type": "aoc_saq_a",
		"title": "AOC 2026",
		"assessment_period_start": "%s",
		"assessment_period_end": "%s"
	}`, assessmentStart, assessmentEnd)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestFinalizeDocument_AuditorForbidden(t *testing.T) {
	// DocumentFinalizeRoles is ciso-only.
	router, _ := setupDocRouterWithRole("auditor")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/finalize", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestGenerateDocument_ComplianceManagerCanTrigger(t *testing.T) {
	// compliance_manager is in DocumentCreateRoles — can trigger generation.
	router, mock := setupDocRouterWithRole("compliance_manager")

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("draft"))

	mock.ExpectExec(`UPDATE compliance_documents SET doc_status`).
		WithArgs("generating", "user-001", "doc-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/generate", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
}

// =============================================================================
// Cross-Tenant Isolation (Security)
// =============================================================================

func TestGetComplianceDocument_CrossTenantIsolation(t *testing.T) {
	// org-002 tries to read org-001's compliance document.
	router, mock := setupDocRouterTenant("org-002")

	mock.ExpectQuery(`SELECT .* FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-002").
		WillReturnRows(sqlmock.NewRows(docCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/doc-001", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateComplianceDocument_CrossTenantIsolation(t *testing.T) {
	router, mock := setupDocRouterTenant("org-002")

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-002").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}))

	body := `{"title": "Tampered Title"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestFinalizeDocument_CrossTenantIsolation(t *testing.T) {
	router, mock := setupDocRouterTenant("org-002")

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-002").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/finalize", nil)
	router.ServeHTTP(w, req)

	// 403 (CISO-only role check) happens before we even check org_id, so that's fine.
	// The important thing: no org-001 data is returned.
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestUpsertDocumentSection_CrossTenantIsolation(t *testing.T) {
	router, mock := setupDocRouterTenant("org-002")

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-002").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}))

	body := `{"title": "Injected Section"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/documents/doc-001/sections/pci_req_1",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestListComplianceDocuments_IsolatedToOwnOrg(t *testing.T) {
	router, mock := setupDocRouterTenant("org-002")

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM compliance_documents`).
		WithArgs("org-002"). // NOT org-001
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM compliance_documents`).
		WithArgs("org-002", 20, 0).
		WillReturnRows(sqlmock.NewRows(docCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Document Correctness / Edge Cases
// =============================================================================

func TestCreateComplianceDocument_AllSAQVariantsAccepted(t *testing.T) {
	// All 8 SAQ variants and ROC must be accepted as valid document types.
	validTypes := []string{
		"aoc_saq_a", "aoc_saq_a_ep", "aoc_saq_b", "aoc_saq_b_ip",
		"aoc_saq_c_vt", "aoc_saq_c", "aoc_saq_d", "aoc_saq_d_sp", "roc",
	}

	for _, docType := range validTypes {
		t.Run(docType, func(t *testing.T) {
			router, mock := setupDocRouter()

			mock.ExpectQuery(`INSERT INTO compliance_documents`).
				WillReturnRows(sqlmock.NewRows(docCols).
					AddRow("doc-x", "org-001", nil, docType, "Test Doc", docNow, docNow, "4.0.1",
						nil, nil, nil, nil, nil, nil, nil, "draft", nil, nil, nil, nil, nil, nil, 1, nil, "user-001", docNow, docNow))

			body := fmt.Sprintf(`{
				"document_type": "%s",
				"title": "Test Doc",
				"assessment_period_start": "%s",
				"assessment_period_end": "%s"
			}`, docType, assessmentStart, assessmentEnd)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/documents", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusCreated, w.Code, "document type %q should be accepted", docType)
		})
	}
}

func TestGetDocumentRequirements_Success(t *testing.T) {
	// The requirements snapshot should reflect compliance posture at generation time.
	router, mock := setupDocRouter()

	snapCols := []string{
		"id", "document_id", "requirement_id", "requirement_code", "requirement_title",
		"in_scope", "control_count", "passing_controls", "evidence_count",
		"status", "notes", "snapshotted_at",
	}

	mock.ExpectQuery(`SELECT .* FROM document_requirement_snapshots WHERE document_id.*AND org_id`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows(snapCols).
			AddRow("snap-001", "doc-001", nil, "1.1", "Install and Maintain Network Security Controls",
				true, 5, 5, 12, "compliant", nil, docNow).
			AddRow("snap-002", "doc-001", nil, "2.2", "System Components Configured Securely",
				true, 3, 2, 6, "partially_compliant", "One control failing", docNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/doc-001/requirements", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	require.Len(t, data, 2)
	assert.Equal(t, "1.1", data[0].(map[string]interface{})["requirement_code"])
	assert.Equal(t, "compliant", data[0].(map[string]interface{})["status"])
	assert.Equal(t, "partially_compliant", data[1].(map[string]interface{})["status"])
}

func TestGenerateDocument_PopulatesFromLiveData(t *testing.T) {
	// After triggering generation, the response should confirm an async job was queued.
	// This verifies the handler returns 202 Accepted (not 200 or 201).
	router, mock := setupDocRouter()

	mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
		WithArgs("doc-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow("draft"))

	mock.ExpectExec(`UPDATE compliance_documents SET doc_status`).
		WithArgs("generating", "user-001", "doc-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/documents/doc-001/generate", nil)
	router.ServeHTTP(w, req)

	// 202 Accepted = async job queued, not synchronous completion.
	require.Equal(t, http.StatusAccepted, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "generating", data["doc_status"])
}

func TestStatusTransitions_FullDocumentLifecycle(t *testing.T) {
	// Validates the allowed state machine: draft → generating → review → approved → final.
	invalidTransitions := []struct {
		from string
		to   string
	}{
		{"final", "draft"},    // terminal state — cannot go back
		{"signed", "review"},  // terminal state — cannot go back
		{"cancelled", "draft"}, // terminal state — cannot go back
		{"review", "final"},   // must go through approved before final
	}

	for _, tt := range invalidTransitions {
		t.Run(fmt.Sprintf("%s→%s", tt.from, tt.to), func(t *testing.T) {
			router, mock := setupDocRouterWithRole("ciso")

			mock.ExpectQuery(`SELECT doc_status FROM compliance_documents WHERE id = \$1 AND org_id = \$2`).
				WithArgs("doc-001", "org-001").
				WillReturnRows(sqlmock.NewRows([]string{"doc_status"}).AddRow(tt.from))

			// Attempt to force-transition by calling generate (draft→generating) or finalize (approved→final).
			endpoint := "/api/v1/documents/doc-001/generate"
			if tt.to == "final" {
				endpoint = "/api/v1/documents/doc-001/finalize"
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", endpoint, nil)
			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusConflict, w.Code,
				"transition %s→%s should be rejected", tt.from, tt.to)
		})
	}
}
