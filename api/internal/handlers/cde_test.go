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

// setupCDERouter creates a test router with all CDE endpoints registered and
// the org context pre-seeded from JWT middleware.
func setupCDERouter() (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1/cde")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "sec@acme.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	// Assets
	protected.GET("/assets", ListCDEAssets)
	protected.POST("/assets", CreateCDEAsset)
	protected.GET("/assets/:id", GetCDEAsset)
	protected.PUT("/assets/:id", UpdateCDEAsset)
	protected.DELETE("/assets/:id", DeleteCDEAsset)

	// Segments
	protected.GET("/segments", ListCDESegments)
	protected.POST("/segments", CreateCDESegment)
	protected.GET("/segments/:id", GetCDESegment)
	protected.PUT("/segments/:id", UpdateCDESegment)
	protected.DELETE("/segments/:id", DeleteCDESegment)

	// Data Flows
	protected.GET("/data-flows", ListCDEDataFlows)
	protected.POST("/data-flows", CreateCDEDataFlow)
	protected.GET("/data-flows/:id", GetCDEDataFlow)
	protected.PUT("/data-flows/:id", UpdateCDEDataFlow)
	protected.DELETE("/data-flows/:id", DeleteCDEDataFlow)

	// Segmentation Tests
	protected.GET("/segmentation-tests", ListCDESegmentationTests)
	protected.POST("/segmentation-tests", CreateCDESegmentationTest)
	protected.GET("/segmentation-tests/:id", GetCDESegmentationTest)
	protected.PUT("/segmentation-tests/:id", UpdateCDESegmentationTest)
	protected.DELETE("/segmentation-tests/:id", DeleteCDESegmentationTest)

	// Summary
	protected.GET("/scope-summary", GetCDEScopeSummary)

	return router, mock
}

// assetCols are the column names returned by cde_assets SELECT queries.
var assetCols = []string{
	"id", "org_id", "name", "type", "ip_address", "hostname", "environment",
	"scope_status", "scope_justification", "data_classification", "created_at", "updated_at",
}

// segCols are the column names returned by cde_network_segments SELECT queries.
var segCols = []string{
	"id", "org_id", "name", "vlan", "subnet", "segment_type", "isolation_method",
	"created_at", "updated_at",
}

// flowCols are the column names returned by cde_data_flows SELECT queries.
var flowCols = []string{
	"id", "org_id", "source_asset_id", "dest_asset_id", "protocol", "port",
	"data_type", "encryption_method", "created_at", "updated_at",
}

// segTestCols are the column names returned by segmentation_tests SELECT queries.
var segTestCols = []string{
	"id", "org_id", "test_date", "tester", "methodology", "segment_id",
	"result", "findings", "next_test_date", "created_at", "updated_at",
}

var now = time.Now()

// =============================================================================
// CDE Assets
// =============================================================================

func TestListCDEAssets_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cde_assets`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM cde_assets`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(assetCols).AddRow(
			"asset-001", "org-001", "DB Server", "database", "10.0.0.1", "db01.internal",
			"production", "in_scope", "Processes CHD", "pan", now, now,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/assets", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	data := resp["data"].([]interface{})
	require.Len(t, data, 1)
	assert.Equal(t, "asset-001", data[0].(map[string]interface{})["id"])
}

func TestListCDEAssets_WithFilters(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cde_assets`).
		WithArgs("org-001", "in_scope").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM cde_assets`).
		WithArgs("org-001", "in_scope", 20, 0).
		WillReturnRows(sqlmock.NewRows(assetCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/assets?scope_status=in_scope", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCDEAsset_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT .* FROM cde_assets WHERE id = \$1 AND org_id = \$2`).
		WithArgs("asset-001", "org-001").
		WillReturnRows(sqlmock.NewRows(assetCols).AddRow(
			"asset-001", "org-001", "DB Server", "database", "10.0.0.1", "db01.internal",
			"production", "in_scope", "Processes CHD", "pan", now, now,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/assets/asset-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "asset-001", resp["data"].(map[string]interface{})["id"])
}

func TestGetCDEAsset_NotFound(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT .* FROM cde_assets WHERE id = \$1 AND org_id = \$2`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows(assetCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/assets/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateCDEAsset_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`INSERT INTO cde_assets`).
		WithArgs(
			sqlmock.AnyArg(), // id (uuid)
			"org-001",
			"Web App Server",
			"server",
			nil,    // ip_address
			nil,    // hostname
			"production",
			"in_scope",
			nil,    // scope_justification
			"chd",
		).
		WillReturnRows(sqlmock.NewRows(assetCols).AddRow(
			"new-asset", "org-001", "Web App Server", "server", nil, nil,
			"production", "in_scope", nil, "chd", now, now,
		))

	body := `{
		"name": "Web App Server",
		"type": "server",
		"environment": "production",
		"scope_status": "in_scope",
		"data_classification": "chd"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/assets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Web App Server", data["name"])
	assert.Equal(t, "in_scope", data["scope_status"])
}

func TestCreateCDEAsset_ValidationError_InvalidType(t *testing.T) {
	router, _ := setupCDERouter()

	body := `{
		"name": "Asset",
		"type": "invalid_type",
		"environment": "production",
		"scope_status": "in_scope",
		"data_classification": "pan"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/assets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
}

func TestCreateCDEAsset_ValidationError_MissingRequired(t *testing.T) {
	router, _ := setupCDERouter()

	// Missing required 'name' field.
	body := `{"type": "server", "environment": "production", "scope_status": "in_scope", "data_classification": "pan"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/assets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCDEAsset_Success(t *testing.T) {
	router, mock := setupCDERouter()

	// Existence check.
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("asset-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Update query.
	mock.ExpectQuery(`UPDATE cde_assets SET`).
		WithArgs("out_of_scope", "asset-001", "org-001").
		WillReturnRows(sqlmock.NewRows(assetCols).AddRow(
			"asset-001", "org-001", "DB Server", "database", nil, nil,
			"production", "out_of_scope", nil, "pan", now, now,
		))

	body := `{"scope_status": "out_of_scope"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/cde/assets/asset-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "out_of_scope", resp["data"].(map[string]interface{})["scope_status"])
}

func TestUpdateCDEAsset_NotFound(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	body := `{"scope_status": "out_of_scope"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/cde/assets/missing", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteCDEAsset_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("asset-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Cascade delete data flows.
	mock.ExpectExec(`DELETE FROM cde_data_flows`).
		WithArgs("org-001", "asset-001").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Delete asset.
	mock.ExpectExec(`DELETE FROM cde_assets`).
		WithArgs("asset-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/cde/assets/asset-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteCDEAsset_NotFound(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectRollback()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/cde/assets/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Network Segments
// =============================================================================

func TestListCDESegments_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cde_network_segments`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM cde_network_segments`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(segCols).AddRow(
			"seg-001", "org-001", "CDE Core", "100", "10.0.0.0/24", "cde", "firewall",
			now, now,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/segments", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.EqualValues(t, 1, resp["meta"].(map[string]interface{})["total"])
}

func TestCreateCDESegment_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`INSERT INTO cde_network_segments`).
		WithArgs(
			sqlmock.AnyArg(), "org-001", "CDE Core", nil, "10.0.0.0/24", "cde", nil,
		).
		WillReturnRows(sqlmock.NewRows(segCols).AddRow(
			"new-seg", "org-001", "CDE Core", nil, "10.0.0.0/24", "cde", nil, now, now,
		))

	body := `{"name": "CDE Core", "subnet": "10.0.0.0/24", "segment_type": "cde"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/segments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateCDESegment_InvalidType(t *testing.T) {
	router, _ := setupCDERouter()

	body := `{"name": "Seg", "segment_type": "bogus"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/segments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCDESegment_NotFound(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT .* FROM cde_network_segments WHERE id = \$1 AND org_id = \$2`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows(segCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/segments/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteCDESegment_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectExec(`DELETE FROM cde_network_segments`).
		WithArgs("seg-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/cde/segments/seg-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteCDESegment_NotFound(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectExec(`DELETE FROM cde_network_segments`).
		WithArgs("missing", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 0))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/cde/segments/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// Data Flows
// =============================================================================

func TestCreateCDEDataFlow_Success(t *testing.T) {
	router, mock := setupCDERouter()

	// Source asset exists.
	mock.ExpectQuery(`SELECT EXISTS.*cde_assets.*\$1.*\$2`).
		WithArgs("asset-src", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Dest asset exists.
	mock.ExpectQuery(`SELECT EXISTS.*cde_assets.*\$1.*\$2`).
		WithArgs("asset-dst", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`INSERT INTO cde_data_flows`).
		WithArgs(
			sqlmock.AnyArg(), "org-001",
			"asset-src", "asset-dst",
			nil, nil, nil, nil,
		).
		WillReturnRows(sqlmock.NewRows(flowCols).AddRow(
			"flow-001", "org-001", "asset-src", "asset-dst", nil, nil, nil, nil, now, now,
		))

	body := `{"source_asset_id": "asset-src", "dest_asset_id": "asset-dst"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/data-flows", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "flow-001", data["id"])
}

func TestCreateCDEDataFlow_SameSourceAndDest(t *testing.T) {
	router, _ := setupCDERouter()

	body := `{"source_asset_id": "asset-x", "dest_asset_id": "asset-x"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/data-flows", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateCDEDataFlow_InvalidPort(t *testing.T) {
	router, _ := setupCDERouter()

	body := `{"source_asset_id": "src", "dest_asset_id": "dst", "port": 99999}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/data-flows", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListCDEDataFlows_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cde_data_flows`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM cde_data_flows`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(flowCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/data-flows", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteCDEDataFlow_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectExec(`DELETE FROM cde_data_flows WHERE id`).
		WithArgs("flow-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/cde/data-flows/flow-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// =============================================================================
// Segmentation Tests
// =============================================================================

func TestCreateCDESegmentationTest_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`INSERT INTO segmentation_tests`).
		WithArgs(
			sqlmock.AnyArg(), // id
			"org-001",
			sqlmock.AnyArg(), // test_date
			"Jane QSA",
			nil, // methodology
			nil, // segment_id
			"pass",
			nil, // findings
			nil, // next_test_date
		).
		WillReturnRows(sqlmock.NewRows(segTestCols).AddRow(
			"test-001", "org-001", now, "Jane QSA", nil, nil,
			"pass", nil, nil, now, now,
		))

	body := fmt.Sprintf(`{"test_date": "%s", "tester": "Jane QSA", "result": "pass"}`,
		now.Format(time.RFC3339))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/segmentation-tests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "pass", resp["data"].(map[string]interface{})["result"])
}

func TestCreateCDESegmentationTest_InvalidResult(t *testing.T) {
	router, _ := setupCDERouter()

	body := `{"test_date": "2026-01-15", "tester": "Jane", "result": "unknown"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/segmentation-tests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateCDESegmentationTest_InvalidDate(t *testing.T) {
	router, _ := setupCDERouter()

	body := `{"test_date": "not-a-date", "tester": "Jane", "result": "pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/cde/segmentation-tests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListCDESegmentationTests_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM segmentation_tests`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .* FROM segmentation_tests`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(segTestCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/segmentation-tests", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCDESegmentationTest_NotFound(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectQuery(`SELECT .* FROM segmentation_tests WHERE id`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows(segTestCols))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/segmentation-tests/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteCDESegmentationTest_Success(t *testing.T) {
	router, mock := setupCDERouter()

	mock.ExpectExec(`DELETE FROM segmentation_tests`).
		WithArgs("test-001", "org-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/cde/segmentation-tests/test-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// =============================================================================
// Scope Summary
// =============================================================================

func TestGetCDEScopeSummary_Success(t *testing.T) {
	router, mock := setupCDERouter()

	// Asset scope aggregation.
	mock.ExpectQuery(`SELECT scope_status, COUNT\(\*\) FROM cde_assets`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"scope_status", "count"}).
			AddRow("in_scope", 5).
			AddRow("out_of_scope", 2).
			AddRow("unclassified", 1))

	// Segment type aggregation.
	mock.ExpectQuery(`SELECT segment_type, COUNT\(\*\) FROM cde_network_segments`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"segment_type", "count"}).
			AddRow("cde", 1).
			AddRow("connected", 2))

	// Data flow aggregation.
	mock.ExpectQuery(`SELECT COUNT\(\*\),.*FROM cde_data_flows`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"total", "encrypted"}).AddRow(10, 8))

	// Segmentation test stats.
	mock.ExpectQuery(`SELECT.*COUNT\(\*\).*FROM segmentation_tests`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"total", "last_test_date", "next_test_date",
			"overdue_count", "pass_count", "fail_count",
		}).AddRow(3, now, nil, 0, 3, 0))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/cde/scope-summary", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	data := resp["data"].(map[string]interface{})
	assets := data["assets"].(map[string]interface{})
	assert.EqualValues(t, 8, assets["total"])
	assert.EqualValues(t, 5, assets["in_scope"])
	assert.EqualValues(t, 2, assets["out_of_scope"])

	flows := data["data_flows"].(map[string]interface{})
	assert.EqualValues(t, 10, flows["total"])
	assert.EqualValues(t, 8, flows["encrypted"])

	segtests := data["segmentation_tests"].(map[string]interface{})
	assert.EqualValues(t, 3, segtests["total"])
	assert.EqualValues(t, 3, segtests["pass_count"])
}
