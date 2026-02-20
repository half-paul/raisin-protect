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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthRouter(mock sqlmock.Sqlmock) *gin.Engine {
	router := gin.New()
	gin.SetMode(gin.TestMode)
	router.Use(middleware.RequestID())
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u0000000-0000-0000-0000-000000000001")
		c.Set(middleware.ContextKeyOrgID, "a0000000-0000-0000-0000-000000000001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Set(middleware.ContextKeyEmail, "alice@acme.com")
		c.Next()
	})
	return router
}

func TestCreateControl_Success(t *testing.T) {
	router, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	_ = router // unused, we use r with auth
	r.POST("/api/v1/controls", CreateControl)

	// Check identifier uniqueness
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("a0000000-0000-0000-0000-000000000001", "CTRL-NW-015").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Insert
	mock.ExpectExec("INSERT INTO controls").WillReturnResult(sqlmock.NewResult(1, 1))
	// Note: audit log writes to auditDB (nil in tests), not database — no mock needed

	body := `{
		"identifier": "CTRL-NW-015",
		"title": "Web Application Firewall",
		"description": "Deploy and maintain WAF in front of all public-facing web applications.",
		"category": "technical"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/controls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "CTRL-NW-015", data["identifier"])
	assert.Equal(t, "Web Application Firewall", data["title"])
	assert.Equal(t, "technical", data["category"])
	assert.Equal(t, "draft", data["status"])
	assert.Equal(t, true, data["is_custom"])
}

func TestCreateControl_DuplicateIdentifier(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/controls", CreateControl)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	body := `{
		"identifier": "CTRL-AC-001",
		"title": "Duplicate",
		"description": "Already exists",
		"category": "technical"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/controls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateControl_InvalidCategory(t *testing.T) {
	_, _ = setupTestRouter()
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(middleware.RequestID())
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u001")
		c.Set(middleware.ContextKeyOrgID, "a001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Next()
	})
	r.POST("/api/v1/controls", CreateControl)

	body := `{
		"identifier": "CTRL-XX-001",
		"title": "Test",
		"description": "Test description",
		"category": "invalid_category"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/controls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateControl_MissingRequired(t *testing.T) {
	_, _ = setupTestRouter()
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u001")
		c.Set(middleware.ContextKeyOrgID, "a001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Next()
	})
	r.POST("/api/v1/controls", CreateControl)

	body := `{"title": "Missing identifier and category"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/controls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetControl_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.GET("/api/v1/controls/:id", GetControl)

	now := time.Now()
	row := sqlmock.NewRows([]string{
		"id", "identifier", "title", "description", "implementation_guidance",
		"category", "status", "is_custom", "source_template_id",
		"owner_id", "owner_name", "owner_email", "secondary_owner_id",
		"evidence_requirements", "test_criteria", "metadata",
		"created_at", "updated_at",
	}).AddRow(
		"c001", "CTRL-AC-001", "Multi-Factor Authentication",
		"Enforce MFA for all user accounts", nil,
		"technical", "active", false, "TPL-AC-001",
		nil, "", "", nil,
		nil, nil, "{}",
		now, now,
	)
	mock.ExpectQuery("SELECT c.id, c.identifier, c.title").WillReturnRows(row)

	// Mappings query
	mappingRows := sqlmock.NewRows([]string{
		"cm_id", "r_id", "r_identifier", "r_title", "f_name", "fv_version", "strength", "notes",
	}).AddRow(
		"m001", "r001", "8.3.1", "All user access", "PCI DSS", "4.0.1", "primary", nil,
	)
	mock.ExpectQuery("SELECT cm.id, r.id, r.identifier").WillReturnRows(mappingRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls/c001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "CTRL-AC-001", data["identifier"])
	assert.Equal(t, "active", data["status"])
	assert.Equal(t, false, data["is_custom"])

	mappings := data["mappings"].([]interface{})
	assert.Len(t, mappings, 1)
	m := mappings[0].(map[string]interface{})
	assert.Equal(t, "primary", m["strength"])
}

func TestGetControl_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.GET("/api/v1/controls/:id", GetControl)

	mock.ExpectQuery("SELECT c.id").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeControlStatus_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/controls/:id/status", ChangeControlStatus)

	// Get current status
	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(
		sqlmock.NewRows([]string{"status", "identifier"}).AddRow("draft", "CTRL-AC-001"),
	)

	// Update
	mock.ExpectExec("UPDATE controls SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"status": "active"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/controls/c001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "active", data["status"])
	assert.Equal(t, "draft", data["previous_status"])
}

func TestChangeControlStatus_InvalidTransition(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/controls/:id/status", ChangeControlStatus)

	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(
		sqlmock.NewRows([]string{"status", "identifier"}).AddRow("draft", "CTRL-AC-001"),
	)

	body := `{"status": "under_review"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/controls/c001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestDeprecateControl_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.DELETE("/api/v1/controls/:id", DeprecateControl)

	mock.ExpectQuery("SELECT identifier").WillReturnRows(
		sqlmock.NewRows([]string{"identifier"}).AddRow("CTRL-AC-001"),
	)
	mock.ExpectExec("UPDATE controls SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/controls/c001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "deprecated", data["status"])
}

func TestDeprecateControl_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.DELETE("/api/v1/controls/:id", DeprecateControl)

	mock.ExpectQuery("SELECT identifier").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/controls/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBulkControlStatus_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/controls/bulk-status", BulkControlStatus)

	// First control
	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(
		sqlmock.NewRows([]string{"status", "identifier"}).AddRow("draft", "CTRL-AC-001"),
	)
	mock.ExpectExec("UPDATE controls SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	// Second control
	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(
		sqlmock.NewRows([]string{"status", "identifier"}).AddRow("draft", "CTRL-AC-002"),
	)
	mock.ExpectExec("UPDATE controls SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"control_ids": ["c001", "c002"],
		"status": "active"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/controls/bulk-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["updated"])
	assert.Equal(t, float64(0), data["failed"])
}

func TestBulkControlStatus_PartialFailure(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/controls/bulk-status", BulkControlStatus)

	// First control succeeds
	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(
		sqlmock.NewRows([]string{"status", "identifier"}).AddRow("draft", "CTRL-AC-001"),
	)
	mock.ExpectExec("UPDATE controls SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	// Second control not found
	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(sqlmock.NewRows(nil))

	body := `{
		"control_ids": ["c001", "nonexistent"],
		"status": "active"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/controls/bulk-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["updated"])
	assert.Equal(t, float64(1), data["failed"])
}

func TestGetControlStats(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.GET("/api/v1/controls/stats", GetControlStats)

	orgID := "a0000000-0000-0000-0000-000000000001"

	// Total
	mock.ExpectQuery("SELECT COUNT").WithArgs(orgID).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(312),
	)

	// By status
	mock.ExpectQuery("SELECT status, COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).
			AddRow("draft", 15).AddRow("active", 280).
			AddRow("under_review", 12).AddRow("deprecated", 5),
	)

	// By category
	mock.ExpectQuery("SELECT category, COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"category", "count"}).
			AddRow("technical", 145).AddRow("administrative", 98).
			AddRow("physical", 22).AddRow("operational", 47),
	)

	// Custom count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(18),
	)

	// Unowned count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(3),
	)

	// Unmapped count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(8),
	)

	// Frameworks coverage
	mock.ExpectQuery("SELECT f.name, fv.version, fv.id").WillReturnRows(
		sqlmock.NewRows([]string{"name", "version", "id"}),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(312), data["total"])
	assert.Equal(t, float64(18), data["custom_count"])
	assert.Equal(t, float64(294), data["library_count"])
}

func TestListControls_EmptyResult(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.GET("/api/v1/controls", ListControls)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(0),
	)
	mock.ExpectQuery("SELECT c.id").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 0)
}

func TestControlMappings_DeleteNotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.DELETE("/api/v1/controls/:id/mappings/:mid", DeleteControlMapping)

	mock.ExpectQuery("SELECT c.identifier, r.identifier, f.identifier").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/controls/c001/mappings/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestControlMappings_ListForControl(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.GET("/api/v1/controls/:id/mappings", ListControlMappings)

	// Control exists check
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"cm_id", "r_id", "r_identifier", "r_title",
		"f_identifier", "f_name", "fv_version",
		"strength", "notes", "mapped_by", "mapped_by_name", "created_at",
	}).AddRow(
		"m001", "r001", "8.3.1", "All user access",
		"pci_dss", "PCI DSS", "4.0.1",
		"primary", nil, nil, "", now,
	)
	mock.ExpectQuery("SELECT cm.id").WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls/c001/mappings", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		from    string
		to      string
		valid   bool
	}{
		{"draft", "active", true},
		{"draft", "deprecated", true},
		{"draft", "under_review", false},
		{"active", "under_review", true},
		{"active", "deprecated", true},
		{"active", "draft", false},
		{"under_review", "active", true},
		{"under_review", "deprecated", true},
		{"under_review", "draft", false},
		{"deprecated", "draft", true},
		{"deprecated", "active", false},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			result := isValidStatusTransitionTest(tt.from, tt.to)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func isValidStatusTransitionTest(from, to string) bool {
	transitions := map[string][]string{
		"draft":        {"active", "deprecated"},
		"active":       {"under_review", "deprecated"},
		"under_review": {"active", "deprecated"},
		"deprecated":   {"draft"},
	}
	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}
