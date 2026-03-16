package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/db"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func setupAccessEntryRouter() (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	mockDB, mock, _ := sqlmock.New()
	database = &db.DB{DB: mockDB}
	middleware.SetAuditDB(nil)

	router := gin.New()
	router.Use(middleware.RequestID())

	protected := router.Group("/api/v1")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "test@acme.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	ar := protected.Group("/access-reviews")
	{
		entries := ar.Group("/entries")
		{
			entries.GET("", ListAccessEntries)
			entries.GET("/anomalies", GetAccessEntryAnomalies)
			entries.GET("/:id", GetAccessEntry)
		}
	}

	return router, mock
}

func TestListAccessEntries_Success(t *testing.T) {
	router, mock := setupAccessEntryRouter()

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT e.id").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "identity_provider_id", "resource_id", "external_user_id", "user_email",
			"user_display_name", "user_department", "user_title", "user_manager_email", "internal_user_id",
			"role_name", "access_level", "permissions", "is_privileged", "expected_role", "expected_access_level",
			"has_role_drift", "granted_at", "last_used_at", "last_login_at", "status", "mfa_enabled",
			"anomalies", "last_sync_at", "is_service_account", "notes", "created_at", "updated_at",
			"resource_name", "resource_criticality",
		}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/access-reviews/entries", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 0)
}

func TestGetAccessEntry_NotFound(t *testing.T) {
	router, mock := setupAccessEntryRouter()

	mock.ExpectQuery("SELECT e.id, e.org_id").
		WithArgs("nonexistent", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/access-reviews/entries/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAccessEntryAnomalies_Success(t *testing.T) {
	router, mock := setupAccessEntryRouter()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"total", "with_anomalies"}).AddRow(100, 5))

	mock.ExpectQuery("SELECT a.value").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"anomaly_type", "cnt"}).
			AddRow("stale_access", 3).
			AddRow("role_drift", 2))

	mock.ExpectQuery("SELECT r.criticality").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"criticality", "count"}).AddRow("high", 3))

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"department", "count"}).AddRow("Engineering", 5))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/access-reviews/entries/anomalies", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(100), data["total_entries"])
	assert.Equal(t, float64(5), data["entries_with_anomalies"])
}
