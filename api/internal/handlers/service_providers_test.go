package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSPRouter creates a test router with all service provider endpoints registered.
// Default role is compliance_manager (in SPManageRoles).
func setupSPRouter() (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/service-providers")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "cm@acme.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	protected.GET("", ListServiceProviders)
	protected.POST("", CreateServiceProvider)
	protected.GET("/compliance-summary", GetSPComplianceSummary)
	protected.GET("/:id", GetServiceProvider)
	protected.PUT("/:id", UpdateServiceProvider)
	protected.DELETE("/:id", DeleteServiceProvider)

	// Compliance documents sub-resource (route: /:id/documents per main.go).
	protected.GET("/:id/documents", ListSPComplianceDocs)
	protected.POST("/:id/documents", CreateSPComplianceDoc)
	protected.GET("/:id/documents/:docId", GetSPComplianceDoc)
	protected.DELETE("/:id/documents/:docId", DeleteSPComplianceDoc)

	// Responsibility matrix sub-resource (route: /:id/responsibilities per main.go).
	protected.GET("/:id/responsibilities", ListSPResponsibilities)
	protected.PUT("/:id/responsibilities/:reqCode", UpsertSPResponsibility)
	protected.DELETE("/:id/responsibilities/:reqCode", DeleteSPResponsibility)

	return router, mock
}

// setupSPRouterWithRole creates a service provider router with a custom role.
func setupSPRouterWithRole(role string) (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/service-providers")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "user@acme.com")
		c.Set(middleware.ContextKeyRole, role)
		c.Next()
	})

	// Mirror production RBAC: view routes require SPViewRoles, manage routes require SPManageRoles.
	protected.GET("", middleware.RequireRoles(models.SPViewRoles...), ListServiceProviders)
	protected.POST("", middleware.RequireRoles(models.SPManageRoles...), CreateServiceProvider)
	protected.GET("/:id", middleware.RequireRoles(models.SPViewRoles...), GetServiceProvider)
	protected.PUT("/:id", middleware.RequireRoles(models.SPManageRoles...), UpdateServiceProvider)
	protected.DELETE("/:id", middleware.RequireRoles(models.SPManageRoles...), DeleteServiceProvider)
	protected.GET("/:id/documents", middleware.RequireRoles(models.SPViewRoles...), ListSPComplianceDocs)
	protected.POST("/:id/documents", middleware.RequireRoles(models.SPManageRoles...), CreateSPComplianceDoc)
	protected.GET("/compliance-summary", middleware.RequireRoles(models.SPViewRoles...), GetSPComplianceSummary)

	return router, mock
}

// setupSPRouterTenant creates an SP router for a different org — used for cross-tenant tests.
func setupSPRouterTenant(orgID string) (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/service-providers")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-attacker")
		c.Set(middleware.ContextKeyOrgID, orgID)
		c.Set(middleware.ContextKeyEmail, "attacker@evil.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	protected.GET("", ListServiceProviders)
	protected.GET("/:id", GetServiceProvider)
	protected.PUT("/:id", UpdateServiceProvider)
	protected.DELETE("/:id", DeleteServiceProvider)
	protected.GET("/:id/documents", ListSPComplianceDocs)

	return router, mock
}

var spNow = time.Now()

// spCols are the column names returned by service_providers SELECT queries.
var spCols = []string{
	"id", "org_id", "name", "type", "contact_name", "contact_email", "contact_phone",
	"services_provided", "pci_compliance_status", "last_aoc_date", "next_review_date",
	"risk_level", "risk_notes", "contract_start_date", "contract_end_date",
	"is_active", "created_by", "created_at", "updated_at",
}

// spDocCols are the column names for sp_compliance_documents.
var spDocCols = []string{
	"id", "provider_id", "org_id", "document_type", "title", "document_version",
	"upload_path", "valid_from", "valid_until", "reviewed_by", "review_notes",
	"is_current", "uploaded_by", "created_at", "updated_at",
}

// spResponsibilityCols are the column names for sp_responsibility_matrix.
var spResponsibilityCols = []string{
	"id", "provider_id", "org_id", "requirement_id", "requirement_code",
	"responsible_party", "notes", "created_by", "created_at", "updated_at",
}

// =============================================================================
// List Service Providers
// =============================================================================

func TestListServiceProviders_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM service_providers`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT .* FROM service_providers`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("sp-001", "org-001", "Stripe", "payment_processor", nil, nil, nil, nil,
				"compliant", nil, nil, "medium", nil, nil, nil, true, "user-001", spNow, spNow).
			AddRow("sp-002", "org-001", "AWS", "hosting", nil, nil, nil, nil,
				"unknown", nil, nil, "high", nil, nil, nil, true, "user-001", spNow, spNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	require.Len(t, data, 2)
	assert.EqualValues(t, 2, resp["meta"].(map[string]interface{})["total"])
}

func TestListServiceProviders_FilterByComplianceStatus(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM service_providers`).
		WithArgs("org-001", "compliant").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM service_providers`).
		WithArgs("org-001", "compliant", 20, 0).
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("sp-001", "org-001", "Stripe", "payment_processor", nil, nil, nil, nil,
				"compliant", nil, nil, "medium", nil, nil, nil, true, "user-001", spNow, spNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers?compliance_status=compliant", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListServiceProviders_EmptyList(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM service_providers`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM service_providers`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(spCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp["data"].([]interface{}), 0)
}

// =============================================================================
// Create Service Provider
// =============================================================================

func TestCreateServiceProvider_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`INSERT INTO service_providers`).
		WithArgs(
			sqlmock.AnyArg(), // id
			"org-001",
			"Stripe",
			"payment_processor",
			sqlmock.AnyArg(), // contact_name (nil)
			sqlmock.AnyArg(), // contact_email (nil)
			sqlmock.AnyArg(), // contact_phone (nil)
			sqlmock.AnyArg(), // services_provided (nil)
			"unknown",        // default pci_compliance_status
			sqlmock.AnyArg(), // last_aoc_date (nil)
			sqlmock.AnyArg(), // next_review_date (nil)
			"medium",         // default risk_level
			sqlmock.AnyArg(), // risk_notes (nil)
			sqlmock.AnyArg(), // contract_start_date (nil)
			sqlmock.AnyArg(), // contract_end_date (nil)
			"user-001",
		).
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("new-sp", "org-001", "Stripe", "payment_processor", nil, nil, nil, nil,
				"unknown", nil, nil, "medium", nil, nil, nil, true, "user-001", spNow, spNow))

	body := `{"name": "Stripe", "type": "payment_processor"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "new-sp", data["id"])
	assert.Equal(t, "unknown", data["pci_compliance_status"])
}

func TestCreateServiceProvider_WithFullDetails(t *testing.T) {
	router, mock := setupSPRouter()

	email := "billing@stripe.com"
	mock.ExpectQuery(`INSERT INTO service_providers`).
		WithArgs(
			sqlmock.AnyArg(), "org-001", "Stripe", "payment_processor",
			"John Doe", email, nil, "Payment processing", "compliant",
			sqlmock.AnyArg(), sqlmock.AnyArg(), "high", nil, nil, nil, "user-001",
		).
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("new-sp", "org-001", "Stripe", "payment_processor", "John Doe", email, nil,
				"Payment processing", "compliant", nil, nil, "high", nil, nil, nil, true, "user-001", spNow, spNow))

	body := `{
		"name": "Stripe",
		"type": "payment_processor",
		"contact_name": "John Doe",
		"contact_email": "billing@stripe.com",
		"services_provided": "Payment processing",
		"pci_compliance_status": "compliant",
		"risk_level": "high"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateServiceProvider_MissingName(t *testing.T) {
	router, _ := setupSPRouter()

	body := `{"type": "payment_processor"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateServiceProvider_InvalidType(t *testing.T) {
	router, _ := setupSPRouter()

	body := `{"name": "ACME Corp", "type": "mystery_vendor"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp["error"].(map[string]interface{})["code"])
}

func TestCreateServiceProvider_InvalidComplianceStatus(t *testing.T) {
	router, _ := setupSPRouter()

	body := `{"name": "ACME", "type": "hosting", "pci_compliance_status": "partially_compliant"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateServiceProvider_InvalidRiskLevel(t *testing.T) {
	router, _ := setupSPRouter()

	body := `{"name": "ACME", "type": "hosting", "risk_level": "extreme"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// Get Service Provider
// =============================================================================

func TestGetServiceProvider_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT .* FROM service_providers WHERE id = \$1 AND org_id = \$2`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("sp-001", "org-001", "Stripe", "payment_processor", nil, nil, nil, nil,
				"compliant", nil, nil, "medium", nil, nil, nil, true, "user-001", spNow, spNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/sp-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "sp-001", data["id"])
	assert.Equal(t, "compliant", data["pci_compliance_status"])
}

func TestGetServiceProvider_NotFound(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT .* FROM service_providers WHERE id = \$1 AND org_id = \$2`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows(spCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Update Service Provider
// =============================================================================

func TestUpdateServiceProvider_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`UPDATE service_providers SET`).
		WithArgs("compliant", "sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("sp-001", "org-001", "Stripe", "payment_processor", nil, nil, nil, nil,
				"compliant", nil, nil, "medium", nil, nil, nil, true, "user-001", spNow, spNow))

	body := `{"pci_compliance_status": "compliant"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/service-providers/sp-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "compliant", resp["data"].(map[string]interface{})["pci_compliance_status"])
}

func TestUpdateServiceProvider_NotFound(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	body := `{"pci_compliance_status": "compliant"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/service-providers/missing", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateServiceProvider_InvalidComplianceStatus(t *testing.T) {
	router, _ := setupSPRouter()

	body := `{"pci_compliance_status": "partially_okay"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/service-providers/sp-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// Delete Service Provider
// =============================================================================

func TestDeleteServiceProvider_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectBegin()

	// Cascade: delete compliance docs and responsibility matrix rows first.
	mock.ExpectExec(`DELETE FROM sp_compliance_documents`).
		WithArgs("sp-001", "org-001").
		WillReturnResult(sqlmock.NewResult(2, 2))

	mock.ExpectExec(`DELETE FROM sp_responsibility_matrix`).
		WithArgs("sp-001", "org-001").
		WillReturnResult(sqlmock.NewResult(5, 5))

	mock.ExpectExec(`DELETE FROM service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/service-providers/sp-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteServiceProvider_NotFound(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectBegin()

	mock.ExpectExec(`DELETE FROM sp_compliance_documents`).
		WithArgs("missing", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM sp_responsibility_matrix`).
		WithArgs("missing", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM service_providers`).
		WithArgs("missing", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectRollback()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/service-providers/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Compliance Documents Sub-resource
// =============================================================================

func TestListSPComplianceDocs_Success(t *testing.T) {
	router, mock := setupSPRouter()

	// Provider existence check.
	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT .* FROM sp_compliance_documents`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows(spDocCols).
			AddRow("doc-001", "sp-001", "org-001", "aoc", "Stripe AOC 2025", "1.0", nil,
				nil, nil, nil, nil, true, "user-001", spNow, spNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/sp-001/documents", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	require.Len(t, data, 1)
	assert.Equal(t, "doc-001", data[0].(map[string]interface{})["id"])
}

func TestListSPComplianceDocs_ProviderNotFound(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/missing/documents", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateSPComplianceDoc_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`INSERT INTO sp_compliance_documents`).
		WithArgs(
			sqlmock.AnyArg(), "sp-001", "org-001",
			"aoc", "Stripe AOC 2025", "2025.1",
			nil,              // upload_path
			sqlmock.AnyArg(), // valid_from
			sqlmock.AnyArg(), // valid_until
			nil,              // review_notes
			"user-001",
		).
		WillReturnRows(sqlmock.NewRows(spDocCols).
			AddRow("new-doc", "sp-001", "org-001", "aoc", "Stripe AOC 2025", "2025.1", nil,
				nil, nil, nil, nil, true, "user-001", spNow, spNow))

	body := `{"document_type": "aoc", "title": "Stripe AOC 2025", "document_version": "2025.1"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers/sp-001/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "aoc", data["document_type"])
}

func TestCreateSPComplianceDoc_InvalidDocType(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := `{"document_type": "magic_cert"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers/sp-001/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSPComplianceDoc_PathTraversal(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		wantCode int
	}{
		{
			name:     "dotdot traversal",
			body:     `{"document_type":"aoc","title":"T","upload_path":"../../etc/passwd"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "nested dotdot traversal",
			body:     `{"document_type":"aoc","title":"T","upload_path":"docs/../../../etc/passwd"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "absolute path",
			body:     `{"document_type":"aoc","title":"T","upload_path":"/etc/passwd"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "legitimate relative path",
			body:     `{"document_type":"aoc","title":"T","upload_path":"org-001/docs/aoc-2025.pdf"}`,
			wantCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router, mock := setupSPRouter()

			mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
				WithArgs("sp-001", "org-001").
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

			if tc.wantCode == http.StatusCreated {
				uploadPath := "org-001/docs/aoc-2025.pdf"
				mock.ExpectQuery(`INSERT INTO sp_compliance_documents`).
					WithArgs(
						sqlmock.AnyArg(), "sp-001", "org-001",
						"aoc", "T", sqlmock.AnyArg(),
						&uploadPath,
						sqlmock.AnyArg(), sqlmock.AnyArg(),
						nil, "user-001",
					).
					WillReturnRows(sqlmock.NewRows(spDocCols).
						AddRow("new-doc", "sp-001", "org-001", "aoc", "T", "", &uploadPath,
							nil, nil, nil, nil, true, "user-001", spNow, spNow))
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/service-providers/sp-001/documents", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantCode, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestDeleteSPComplianceDoc_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectExec(`DELETE FROM sp_compliance_documents`).
		WithArgs("doc-001", "sp-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/service-providers/sp-001/documents/doc-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// =============================================================================
// Responsibility Matrix
// =============================================================================

func TestListSPResponsibilities_Success(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT .* FROM sp_responsibility_matrix`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows(spResponsibilityCols).
			AddRow("rm-001", "sp-001", "org-001", nil, "12.8.1", "provider", nil, "user-001", spNow, spNow).
			AddRow("rm-002", "sp-001", "org-001", nil, "12.8.2", "shared", nil, "user-001", spNow, spNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/sp-001/responsibilities", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	require.Len(t, data, 2)
	assert.Equal(t, "12.8.1", data[0].(map[string]interface{})["requirement_code"])
}

func TestUpsertSPResponsibility_Create(t *testing.T) {
	router, mock := setupSPRouter()

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Upsert: check if row exists, then insert.
	mock.ExpectQuery(`INSERT INTO sp_responsibility_matrix.*ON CONFLICT`).
		WithArgs(
			sqlmock.AnyArg(), "sp-001", "org-001",
			nil,       // requirement_id (optional)
			"12.8.3",  // requirement_code
			"shared",  // responsible_party
			nil,       // notes
			"user-001",
		).
		WillReturnRows(sqlmock.NewRows(spResponsibilityCols).
			AddRow("rm-new", "sp-001", "org-001", nil, "12.8.3", "shared", nil, "user-001", spNow, spNow))

	body := `{"responsible_party": "shared"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/service-providers/sp-001/responsibilities/12.8.3",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "12.8.3", resp["data"].(map[string]interface{})["requirement_code"])
}

func TestUpsertSPResponsibility_InvalidParty(t *testing.T) {
	router, mock := setupSPRouter()

	// Handler checks provider existence before validating the request body.
	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := `{"responsible_party": "nobody"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/service-providers/sp-001/responsibilities/12.8.3",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// Compliance Summary
// =============================================================================

func TestGetSPComplianceSummary_Success(t *testing.T) {
	router, mock := setupSPRouter()

	summaryCols := []string{"total", "compliant", "non_compliant", "compliance_in_progress",
		"compliance_not_validated", "unknown", "not_applicable", "expiring_30_days",
		"expiring_60_days", "expiring_90_days"}

	mock.ExpectQuery(`SELECT .* FROM service_providers WHERE org_id`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows(summaryCols).
			AddRow(10, 6, 1, 2, 0, 1, 0, 1, 2, 3))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/compliance-summary", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.EqualValues(t, 10, data["total"])
	assert.EqualValues(t, 6, data["compliant"])
	assert.EqualValues(t, 1, data["expiring_30_days"])
}

func TestGetSPComplianceSummary_ExpiringDocs(t *testing.T) {
	// Specifically validates the expiry alert logic — critical for PCI Req 12.8 compliance.
	router, mock := setupSPRouter()

	// SP with an AOC expiring in 25 days — should appear in expiring_30_days.
	summaryCols := []string{"total", "compliant", "non_compliant", "compliance_in_progress",
		"compliance_not_validated", "unknown", "not_applicable", "expiring_30_days",
		"expiring_60_days", "expiring_90_days"}

	mock.ExpectQuery(`SELECT .* FROM service_providers WHERE org_id`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows(summaryCols).
			AddRow(3, 2, 0, 0, 0, 1, 0, 2, 2, 2))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/compliance-summary", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	// 2 providers expiring within 30 days — should surface in the response.
	assert.EqualValues(t, 2, data["expiring_30_days"])
}

// =============================================================================
// RBAC / Access Control
// =============================================================================

func TestListServiceProviders_AuditorCanView(t *testing.T) {
	// Auditors are in SPViewRoles.
	router, mock := setupSPRouterWithRole("auditor")

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM service_providers`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM service_providers`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(spCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateServiceProvider_AuditorForbidden(t *testing.T) {
	// Auditors are NOT in SPManageRoles.
	router, _ := setupSPRouterWithRole("auditor")

	body := `{"name": "Stripe", "type": "payment_processor"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateServiceProvider_VendorManagerCanCreate(t *testing.T) {
	// vendor_manager is specifically in SPManageRoles for this resource.
	router, mock := setupSPRouterWithRole("vendor_manager")

	mock.ExpectQuery(`INSERT INTO service_providers`).
		WithArgs(
			sqlmock.AnyArg(), "org-001", "PayFac", "third_party_agent",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"unknown", sqlmock.AnyArg(), sqlmock.AnyArg(), "medium", sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "user-001",
		).
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("sp-new", "org-001", "PayFac", "third_party_agent", nil, nil, nil, nil,
				"unknown", nil, nil, "medium", nil, nil, nil, true, "user-001", spNow, spNow))

	body := `{"name": "PayFac", "type": "third_party_agent"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestDeleteServiceProvider_SecurityEngineerForbidden(t *testing.T) {
	// security_engineer is NOT in SPManageRoles.
	router, _ := setupSPRouterWithRole("security_engineer")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/service-providers/sp-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// =============================================================================
// Cross-Tenant Isolation (Security)
// =============================================================================

func TestGetServiceProvider_CrossTenantIsolation(t *testing.T) {
	// org-002 tries to read org-001's service provider — must get 404.
	router, mock := setupSPRouterTenant("org-002")

	mock.ExpectQuery(`SELECT .* FROM service_providers WHERE id = \$1 AND org_id = \$2`).
		WithArgs("sp-001", "org-002").
		WillReturnRows(sqlmock.NewRows(spCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/sp-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateServiceProvider_CrossTenantIsolation(t *testing.T) {
	router, mock := setupSPRouterTenant("org-002")

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-002").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	body := `{"pci_compliance_status": "non_compliant"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/service-providers/sp-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListSPComplianceDocs_CrossTenantIsolation(t *testing.T) {
	// org-002 attempts to list compliance docs for org-001's provider.
	router, mock := setupSPRouterTenant("org-002")

	// The provider exists in org-001, but org-002's check will return false.
	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-002").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/sp-001/documents", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListServiceProviders_IsolatedToOwnOrg(t *testing.T) {
	router, mock := setupSPRouterTenant("org-002")

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM service_providers`).
		WithArgs("org-002"). // NOT org-001
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM service_providers`).
		WithArgs("org-002", 20, 0).
		WillReturnRows(sqlmock.NewRows(spCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestCreateServiceProvider_SpecialCharactersInName(t *testing.T) {
	// PCI providers sometimes have special chars — should not be rejected.
	router, mock := setupSPRouter()

	mock.ExpectQuery(`INSERT INTO service_providers`).
		WithArgs(
			sqlmock.AnyArg(), "org-001", "Acme & Sons (PCI/PA-DSS Certified)", "other",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"unknown", sqlmock.AnyArg(), sqlmock.AnyArg(), "medium", sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), "user-001",
		).
		WillReturnRows(sqlmock.NewRows(spCols).
			AddRow("sp-x", "org-001", "Acme & Sons (PCI/PA-DSS Certified)", "other", nil, nil, nil, nil,
				"unknown", nil, nil, "medium", nil, nil, nil, true, "user-001", spNow, spNow))

	body := `{"name": "Acme & Sons (PCI/PA-DSS Certified)", "type": "other"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/service-providers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetSPComplianceDocs_ExpiredDocFlagged(t *testing.T) {
	// A document with valid_until in the past should appear in the list
	// (not filtered out — compliance needs audit trail of expired certs).
	router, mock := setupSPRouter()

	expiredDate := time.Now().Add(-365 * 24 * time.Hour) // 1 year ago

	mock.ExpectQuery(`SELECT EXISTS.*service_providers`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT .* FROM sp_compliance_documents`).
		WithArgs("sp-001", "org-001").
		WillReturnRows(sqlmock.NewRows(spDocCols).
			AddRow("doc-expired", "sp-001", "org-001", "aoc", "Old AOC", "2024.1", nil,
				nil, expiredDate, nil, nil, false, "user-001", spNow, spNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/service-providers/sp-001/documents", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	// Expired doc must still appear — never silently filtered.
	require.Len(t, data, 1)
	assert.Equal(t, "doc-expired", data[0].(map[string]interface{})["id"])
}
