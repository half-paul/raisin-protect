package handlers

import (
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

// setupAuthContext adds fake auth context to test requests.
func setupAuthContext(router *gin.Engine, handler gin.HandlerFunc, method, path string) *gin.Engine {
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u0000000-0000-0000-0000-000000000001")
		c.Set(middleware.ContextKeyOrgID, "a0000000-0000-0000-0000-000000000001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Set(middleware.ContextKeyEmail, "alice@acme.com")
		c.Next()
	})
	router.Handle(method, path, handler)
	return router
}

func TestListFrameworks(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, ListFrameworks, "GET", "/api/v1/frameworks")

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "identifier", "name", "description", "category", "website_url", "logo_url", "versions_count", "created_at",
	}).AddRow(
		"f0000000-0000-0000-0000-000000000001", "soc2", "SOC 2",
		"Service Organization Control 2", "security_privacy",
		"https://www.aicpa.org/soc2", nil, 1, now,
	).AddRow(
		"f0000000-0000-0000-0000-000000000003", "pci_dss", "PCI DSS",
		"Payment Card Industry Data Security Standard", "payment",
		"https://www.pcisecuritystandards.org/", nil, 1, now,
	)

	mock.ExpectQuery("SELECT f.id, f.identifier, f.name").WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/frameworks", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)

	first := data[0].(map[string]interface{})
	assert.Equal(t, "soc2", first["identifier"])
	assert.Equal(t, "SOC 2", first["name"])
	assert.Equal(t, "security_privacy", first["category"])
}

func TestListFrameworks_WithCategoryFilter(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, ListFrameworks, "GET", "/api/v1/frameworks")

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "identifier", "name", "description", "category", "website_url", "logo_url", "versions_count", "created_at",
	}).AddRow(
		"f0000000-0000-0000-0000-000000000003", "pci_dss", "PCI DSS",
		"Payment Card Industry", "payment", nil, nil, 1, now,
	)

	mock.ExpectQuery("SELECT f.id").WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/frameworks?category=payment", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	first := data[0].(map[string]interface{})
	assert.Equal(t, "pci_dss", first["identifier"])
}

func TestGetFramework_Success(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, GetFramework, "GET", "/api/v1/frameworks/:id")

	now := time.Now()

	// Framework query
	fwRow := sqlmock.NewRows([]string{
		"id", "identifier", "name", "description", "category", "website_url", "logo_url", "created_at",
	}).AddRow(
		"f0000000-0000-0000-0000-000000000001", "soc2", "SOC 2",
		"Service Organization Control 2", "security_privacy",
		"https://www.aicpa.org/soc2", nil, now,
	)
	mock.ExpectQuery("SELECT id, identifier, name").WillReturnRows(fwRow)

	// Versions query
	versRows := sqlmock.NewRows([]string{
		"id", "version", "display_name", "status", "effective_date", "sunset_date", "total_requirements", "created_at",
	}).AddRow(
		"v0000000-0000-0000-0000-000000000001", "2024", "SOC 2 (2024 TSC)",
		"active", "2024-01-01", nil, 64, now,
	)
	mock.ExpectQuery("SELECT id, version, display_name").WillReturnRows(versRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/frameworks/f0000000-0000-0000-0000-000000000001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "soc2", data["identifier"])

	versions := data["versions"].([]interface{})
	assert.Len(t, versions, 1)
	v := versions[0].(map[string]interface{})
	assert.Equal(t, "2024", v["version"])
	assert.Equal(t, float64(64), v["total_requirements"])
}

func TestGetFramework_NotFound(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, GetFramework, "GET", "/api/v1/frameworks/:id")

	mock.ExpectQuery("SELECT id, identifier, name").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/frameworks/nonexistent-id", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetFrameworkVersion_Success(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, GetFrameworkVersion, "GET", "/api/v1/frameworks/:id/versions/:vid")

	now := time.Now()
	row := sqlmock.NewRows([]string{
		"id", "framework_id", "identifier", "name", "version", "display_name",
		"status", "effective_date", "sunset_date", "changelog", "total_requirements", "created_at",
	}).AddRow(
		"v0000000-0000-0000-0000-000000000003", "f0000000-0000-0000-0000-000000000003",
		"pci_dss", "PCI DSS", "4.0.1", "PCI DSS v4.0.1",
		"active", "2024-06-11", nil, nil, 280, now,
	)
	mock.ExpectQuery("SELECT fv.id, fv.framework_id").WillReturnRows(row)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		"/api/v1/frameworks/f0000000-0000-0000-0000-000000000003/versions/v0000000-0000-0000-0000-000000000003", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "PCI DSS v4.0.1", data["display_name"])
	assert.Equal(t, float64(280), data["total_requirements"])
}

func TestGetFrameworkVersion_NotFound(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, GetFrameworkVersion, "GET", "/api/v1/frameworks/:id/versions/:vid")

	mock.ExpectQuery("SELECT fv.id").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/frameworks/bad-id/versions/bad-vid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListRequirements_Flat(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, ListRequirements, "GET", "/api/v1/frameworks/:id/versions/:vid/requirements")

	// Verify version exists
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	// Count
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(3),
	)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "identifier", "title", "description", "guidance", "parent_id",
		"depth", "section_order", "is_assessable", "created_at",
	}).AddRow(
		"r001", "6", "Develop and Maintain Secure Systems", nil, nil, nil, 0, 6, false, now,
	).AddRow(
		"r002", "6.4", "Public-Facing Web Apps", nil, nil, "r001", 1, 4, false, now,
	).AddRow(
		"r003", "6.4.3", "Payment page scripts managed", nil, nil, "r002", 2, 3, true, now,
	)

	mock.ExpectQuery("SELECT r.id, r.identifier, r.title").WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		"/api/v1/frameworks/fw-id/versions/v-id/requirements?format=flat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 3)

	third := data[2].(map[string]interface{})
	assert.Equal(t, "6.4.3", third["identifier"])
	assert.Equal(t, true, third["is_assessable"])
}

func TestListRequirements_VersionNotFound(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, ListRequirements, "GET", "/api/v1/frameworks/:id/versions/:vid/requirements")

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(false),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		"/api/v1/frameworks/bad-fw/versions/bad-v/requirements", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListRequirements_Tree(t *testing.T) {
	router, mock := setupTestRouter()
	setupAuthContext(router, ListRequirements, "GET", "/api/v1/frameworks/:id/versions/:vid/requirements")

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	rows := sqlmock.NewRows([]string{
		"id", "parent_id", "identifier", "title", "depth", "section_order", "is_assessable",
	}).AddRow("r001", nil, "6", "Develop Secure Systems", 0, 6, false).
		AddRow("r002", "r001", "6.4", "Public-Facing Web Apps", 1, 4, false).
		AddRow("r003", "r002", "6.4.3", "Payment page scripts", 2, 3, true)

	mock.ExpectQuery("WITH RECURSIVE").WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		"/api/v1/frameworks/fw-id/versions/v-id/requirements?format=tree", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1, "Should have one root node")

	root := data[0].(map[string]interface{})
	assert.Equal(t, "6", root["identifier"])
	children := root["children"].([]interface{})
	assert.Len(t, children, 1)

	child := children[0].(map[string]interface{})
	assert.Equal(t, "6.4", child["identifier"])
	grandchildren := child["children"].([]interface{})
	assert.Len(t, grandchildren, 1)
	assert.Equal(t, "6.4.3", grandchildren[0].(map[string]interface{})["identifier"])
}
