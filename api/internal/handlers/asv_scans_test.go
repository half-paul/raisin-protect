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

// setupASVRouter creates a test router with all ASV scan endpoints registered.
// Default role is compliance_manager (in ASVManageRoles).
func setupASVRouter() (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/asv-scans")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "sec@acme.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	protected.GET("", ListASVScans)
	protected.POST("", CreateASVScan)
	protected.GET("/quarterly-status", GetASVQuarterlyStatus)
	protected.POST("/import", ImportASVScan)
	protected.GET("/:id", GetASVScan)
	protected.PUT("/:id", UpdateASVScan)
	protected.DELETE("/:id", DeleteASVScan)

	return router, mock
}

// setupASVRouterWithRole creates an ASV router with a custom role — used for RBAC tests.
func setupASVRouterWithRole(role string) (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/asv-scans")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "viewer@acme.com")
		c.Set(middleware.ContextKeyRole, role)
		c.Next()
	})

	protected.GET("", ListASVScans)
	protected.POST("", CreateASVScan)
	protected.GET("/quarterly-status", GetASVQuarterlyStatus)
	protected.POST("/import", ImportASVScan)
	protected.GET("/:id", GetASVScan)
	protected.PUT("/:id", UpdateASVScan)
	protected.DELETE("/:id", DeleteASVScan)

	return router, mock
}

// setupASVRouterTenant creates an ASV router with a different org — used for cross-tenant tests.
func setupASVRouterTenant(orgID string) (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/asv-scans")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-attacker")
		c.Set(middleware.ContextKeyOrgID, orgID)
		c.Set(middleware.ContextKeyEmail, "attacker@evil.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	protected.GET("", ListASVScans)
	protected.GET("/:id", GetASVScan)
	protected.PUT("/:id", UpdateASVScan)
	protected.DELETE("/:id", DeleteASVScan)

	return router, mock
}

// asvCols are the column names returned by asv_scans SELECT queries.
var asvCols = []string{
	"id", "org_id", "asv_vendor", "scan_type", "quarter", "year", "scan_date",
	"status", "findings_count", "critical_count", "high_count", "medium_count",
	"low_count", "informational_count", "remediation_deadline", "report_path",
	"import_format", "raw_findings", "import_notes", "imported_by", "reviewed_by",
	"created_by", "created_at", "updated_at",
}

var asvNow = time.Now()

// =============================================================================
// List ASV Scans
// =============================================================================

func TestListASVScans_Success(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM asv_scans`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT .* FROM asv_scans`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(asvCols).
			AddRow("scan-001", "org-001", "Trustwave", "external", 1, 2026, asvNow,
				"pass", 12, 0, 2, 5, 5, 0, nil, nil, nil, nil, nil, nil, nil, "user-001", asvNow, asvNow).
			AddRow("scan-002", "org-001", "SecurityMetrics", "external", 4, 2025, asvNow,
				"fail", 3, 1, 2, 0, 0, 0, nil, nil, nil, nil, nil, nil, nil, "user-001", asvNow, asvNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	data := resp["data"].([]interface{})
	require.Len(t, data, 2)
	assert.Equal(t, "scan-001", data[0].(map[string]interface{})["id"])
	assert.EqualValues(t, 2, resp["meta"].(map[string]interface{})["total"])
}

func TestListASVScans_FilterByStatus(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM asv_scans`).
		WithArgs("org-001", "fail").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM asv_scans`).
		WithArgs("org-001", "fail", 20, 0).
		WillReturnRows(sqlmock.NewRows(asvCols).
			AddRow("scan-002", "org-001", "SecurityMetrics", "external", 4, 2025, asvNow,
				"fail", 3, 1, 2, 0, 0, 0, nil, nil, nil, nil, nil, nil, nil, "user-001", asvNow, asvNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans?status=fail", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListASVScans_FilterByYear(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM asv_scans`).
		WithArgs("org-001", 2026).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM asv_scans`).
		WithArgs("org-001", 2026, 20, 0).
		WillReturnRows(sqlmock.NewRows(asvCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans?year=2026", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListASVScans_EmptyList(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM asv_scans`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM asv_scans`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(asvCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 0)
	assert.EqualValues(t, 0, resp["meta"].(map[string]interface{})["total"])
}

// =============================================================================
// Create ASV Scan
// =============================================================================

func TestCreateASVScan_Success(t *testing.T) {
	router, mock := setupASVRouter()

	// Uniqueness check: no duplicate scan for same org/type/quarter/year.
	mock.ExpectQuery(`SELECT EXISTS.*asv_scans`).
		WithArgs("org-001", "external", 1, 2026).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO asv_scans`).
		WithArgs(
			sqlmock.AnyArg(), // id
			"org-001",
			"Trustwave",
			"external",
			1,                   // quarter
			2026,                // year
			sqlmock.AnyArg(),    // scan_date (parsed)
			"in_progress",       // default status
			sqlmock.AnyArg(),    // remediation_deadline (nil)
			sqlmock.AnyArg(),    // import_notes (nil)
			"user-001",
		).
		WillReturnRows(sqlmock.NewRows(asvCols).
			AddRow("new-scan", "org-001", "Trustwave", "external", 1, 2026, asvNow,
				"in_progress", 0, 0, 0, 0, 0, 0, nil, nil, nil, nil, nil, nil, nil, "user-001", asvNow, asvNow))

	body := fmt.Sprintf(`{
		"asv_vendor": "Trustwave",
		"scan_type": "external",
		"quarter": 1,
		"year": 2026,
		"scan_date": "%s"
	}`, asvNow.Format(time.RFC3339))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "new-scan", data["id"])
	assert.Equal(t, "in_progress", data["status"])
}

func TestCreateASVScan_MissingRequiredFields(t *testing.T) {
	router, _ := setupASVRouter()

	// Missing scan_date and quarter.
	body := `{"asv_vendor": "Trustwave", "year": 2026}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateASVScan_InvalidStatus(t *testing.T) {
	router, _ := setupASVRouter()

	body := fmt.Sprintf(`{
		"asv_vendor": "Trustwave",
		"quarter": 1,
		"year": 2026,
		"scan_date": "%s",
		"status": "unknown_status"
	}`, asvNow.Format(time.RFC3339))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
}

func TestCreateASVScan_InvalidQuarter_OutOfRange(t *testing.T) {
	// PCI requires quarterly scans — quarter must be 1–4.
	router, _ := setupASVRouter()

	body := fmt.Sprintf(`{
		"asv_vendor": "Trustwave",
		"quarter": 5,
		"year": 2026,
		"scan_date": "%s"
	}`, asvNow.Format(time.RFC3339))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateASVScan_DuplicateQuarterConflict(t *testing.T) {
	// Two external scans for the same org/quarter/year should be rejected (409).
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT EXISTS.*asv_scans`).
		WithArgs("org-001", "external", 1, 2026).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := fmt.Sprintf(`{
		"asv_vendor": "Qualys",
		"scan_type": "external",
		"quarter": 1,
		"year": 2026,
		"scan_date": "%s"
	}`, asvNow.Format(time.RFC3339))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateASVScan_InvalidDate(t *testing.T) {
	router, _ := setupASVRouter()

	body := `{"asv_vendor": "Trustwave", "quarter": 1, "year": 2026, "scan_date": "not-a-date"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// Get ASV Scan
// =============================================================================

func TestGetASVScan_Success(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT .* FROM asv_scans WHERE id = \$1 AND org_id = \$2`).
		WithArgs("scan-001", "org-001").
		WillReturnRows(sqlmock.NewRows(asvCols).
			AddRow("scan-001", "org-001", "Trustwave", "external", 1, 2026, asvNow,
				"pass", 12, 0, 2, 5, 5, 0, nil, nil, nil, nil, nil, nil, nil, "user-001", asvNow, asvNow))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans/scan-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "scan-001", data["id"])
	assert.Equal(t, "pass", data["status"])
}

func TestGetASVScan_NotFound(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT .* FROM asv_scans WHERE id = \$1 AND org_id = \$2`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows(asvCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Update ASV Scan
// =============================================================================

func TestUpdateASVScan_Success(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT EXISTS.*asv_scans`).
		WithArgs("scan-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`UPDATE asv_scans SET`).
		WithArgs("pass", 15, 0, 3, 7, 5, 0, "scan-001", "org-001").
		WillReturnRows(sqlmock.NewRows(asvCols).
			AddRow("scan-001", "org-001", "Trustwave", "external", 1, 2026, asvNow,
				"pass", 15, 0, 3, 7, 5, 0, nil, nil, nil, nil, nil, nil, nil, "user-001", asvNow, asvNow))

	body := `{
		"status": "pass",
		"findings_count": 15,
		"critical_count": 0,
		"high_count": 3,
		"medium_count": 7,
		"low_count": 5,
		"informational_count": 0
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/asv-scans/scan-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "pass", resp["data"].(map[string]interface{})["status"])
}

func TestUpdateASVScan_InvalidStatus(t *testing.T) {
	router, _ := setupASVRouter()

	body := `{"status": "invalid_status"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/asv-scans/scan-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateASVScan_NotFound(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT EXISTS.*asv_scans`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	body := `{"status": "pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/asv-scans/missing", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Delete ASV Scan
// =============================================================================

func TestDeleteASVScan_Success(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectExec(`DELETE FROM asv_scans`).
		WithArgs("scan-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/asv-scans/scan-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteASVScan_NotFound(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectExec(`DELETE FROM asv_scans`).
		WithArgs("missing", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 0))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/asv-scans/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Quarterly Status Report
// =============================================================================

func TestGetASVQuarterlyStatus_Success(t *testing.T) {
	router, mock := setupASVRouter()

	// Returns scans for 2025 and 2026.
	quarterCols := []string{"year", "quarter", "status", "scan_id"}
	mock.ExpectQuery(`SELECT.*year.*quarter.*status.*id.*FROM asv_scans`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows(quarterCols).
			AddRow(2026, 1, "pass", "scan-001").
			AddRow(2025, 4, "fail", "scan-002").
			AddRow(2025, 3, "pass", "scan-003").
			AddRow(2025, 2, "pass", "scan-004").
			AddRow(2025, 1, "pass", "scan-005"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans/quarterly-status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "org-001", data["org_id"])

	periods := data["periods"].([]interface{})
	assert.NotEmpty(t, periods)
}

func TestGetASVQuarterlyStatus_NoScans_ShowsMissingQuarters(t *testing.T) {
	// When no scans exist, all recent quarters should appear as nil status (not missing from response).
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT.*year.*quarter.*status.*id.*FROM asv_scans`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"year", "quarter", "status", "scan_id"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans/quarterly-status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Must still return periods (last 4+ quarters shown as not recorded).
	data := resp["data"].(map[string]interface{})
	periods := data["periods"].([]interface{})
	assert.NotEmpty(t, periods)

	// All periods must have nil status (no scan recorded).
	for _, p := range periods {
		period := p.(map[string]interface{})
		assert.Nil(t, period["status"], "unrecorded quarter must have nil status")
	}
}

func TestGetASVQuarterlyStatus_YearFilter(t *testing.T) {
	router, mock := setupASVRouter()

	mock.ExpectQuery(`SELECT.*year.*quarter.*status.*id.*FROM asv_scans`).
		WithArgs("org-001", 2025).
		WillReturnRows(sqlmock.NewRows([]string{"year", "quarter", "status", "scan_id"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans/quarterly-status?year=2025", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Import ASV Scan (CSV / XML)
// =============================================================================

func TestImportASVScan_CSV_Success(t *testing.T) {
	router, mock := setupASVRouter()

	// Uniqueness check.
	mock.ExpectQuery(`SELECT EXISTS.*asv_scans`).
		WithArgs("org-001", "external", 2, 2026).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO asv_scans`).
		WithArgs(
			sqlmock.AnyArg(), "org-001", "Qualys", "external", 2, 2026,
			sqlmock.AnyArg(), "pass", sqlmock.AnyArg(), nil, "user-001",
		).
		WillReturnRows(sqlmock.NewRows(asvCols).
			AddRow("import-001", "org-001", "Qualys", "external", 2, 2026, asvNow,
				"pass", 8, 0, 1, 3, 4, 0, nil, nil, "csv", nil, nil, "user-001", nil, "user-001", asvNow, asvNow))

	body := fmt.Sprintf(`{
		"asv_vendor": "Qualys",
		"scan_type": "external",
		"quarter": 2,
		"year": 2026,
		"scan_date": "%s",
		"import_format": "csv",
		"raw_findings": "[{\"severity\":\"high\",\"host\":\"10.0.0.1\"}]"
	}`, asvNow.Format(time.RFC3339))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans/import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "import-001", data["id"])
}

func TestImportASVScan_InvalidFormat(t *testing.T) {
	router, _ := setupASVRouter()

	body := fmt.Sprintf(`{
		"asv_vendor": "Qualys",
		"quarter": 2,
		"year": 2026,
		"scan_date": "%s",
		"import_format": "docx"
	}`, asvNow.Format(time.RFC3339))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans/import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// RBAC / Access Control
// =============================================================================

func TestListASVScans_AuditorCanView(t *testing.T) {
	// Auditors are in ASVViewRoles — can list scans.
	router, mock := setupASVRouterWithRole("auditor")

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM asv_scans`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM asv_scans`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(asvCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateASVScan_AuditorForbidden(t *testing.T) {
	// Auditors are NOT in ASVManageRoles — must not create scans.
	router, _ := setupASVRouterWithRole("auditor")

	body := fmt.Sprintf(`{"asv_vendor": "X", "quarter": 1, "year": 2026, "scan_date": "%s"}`,
		asvNow.Format(time.RFC3339))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/asv-scans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteASVScan_AuditorForbidden(t *testing.T) {
	router, _ := setupASVRouterWithRole("auditor")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/asv-scans/scan-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateASVScan_SecurityEngineerCanUpdate(t *testing.T) {
	// security_engineer is in ASVManageRoles.
	router, mock := setupASVRouterWithRole("security_engineer")

	mock.ExpectQuery(`SELECT EXISTS.*asv_scans`).
		WithArgs("scan-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`UPDATE asv_scans SET`).
		WithArgs(sqlmock.AnyArg(), "scan-001", "org-001").
		WillReturnRows(sqlmock.NewRows(asvCols).
			AddRow("scan-001", "org-001", "Trustwave", "external", 1, 2026, asvNow,
				"remediated", 12, 0, 2, 5, 5, 0, nil, nil, nil, nil, nil, nil, nil, "user-001", asvNow, asvNow))

	body := `{"status": "remediated"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/asv-scans/scan-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Cross-Tenant Isolation (Security)
// =============================================================================

func TestGetASVScan_CrossTenantIsolation(t *testing.T) {
	// org-002 tries to read org-001's scan — must get 404, not the actual data.
	router, mock := setupASVRouterTenant("org-002")

	mock.ExpectQuery(`SELECT .* FROM asv_scans WHERE id = \$1 AND org_id = \$2`).
		WithArgs("scan-001", "org-002"). // handler must inject requester's org, not path param
		WillReturnRows(sqlmock.NewRows(asvCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans/scan-001", nil)
	router.ServeHTTP(w, req)

	// 404 — not 200 with org-001 data, and not 403 which would leak existence
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateASVScan_CrossTenantIsolation(t *testing.T) {
	router, mock := setupASVRouterTenant("org-002")

	mock.ExpectQuery(`SELECT EXISTS.*asv_scans`).
		WithArgs("scan-001", "org-002").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	body := `{"status": "pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/asv-scans/scan-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteASVScan_CrossTenantIsolation(t *testing.T) {
	router, mock := setupASVRouterTenant("org-002")

	mock.ExpectExec(`DELETE FROM asv_scans`).
		WithArgs("scan-001", "org-002").
		WillReturnResult(sqlmock.NewResult(0, 0))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/asv-scans/scan-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListASVScans_IsolatedToOwnOrg(t *testing.T) {
	// org-002 must only receive its own scans, never org-001's.
	router, mock := setupASVRouterTenant("org-002")

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM asv_scans`).
		WithArgs("org-002"). // org-002, NOT org-001
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM asv_scans`).
		WithArgs("org-002", 20, 0).
		WillReturnRows(sqlmock.NewRows(asvCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/asv-scans", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// sqlmock enforces the org-002 WHERE clause — if handler passes org-001, mock will fail.
}
