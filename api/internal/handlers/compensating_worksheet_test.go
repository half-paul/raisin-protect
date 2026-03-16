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

var cwNow = time.Now()

func setupWorksheetRouter(mock sqlmock.Sqlmock) *gin.Engine {
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(middleware.RequestID())
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Set(middleware.ContextKeyEmail, "cm@acme.com")
		c.Next()
	})
	middleware.SetAuditDB(nil)
	r.GET("/api/v1/controls/:id/compensating-worksheet", GetCompensatingWorksheet)
	r.PUT("/api/v1/controls/:id/compensating-worksheet", UpdateCompensatingWorksheet)
	r.GET("/api/v1/controls", ListControls)
	return r
}

// =============================================================================
// GetCompensatingWorksheet
// =============================================================================

func TestGetCompensatingWorksheet_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupWorksheetRouter(mock)

	worksheetJSON := `{"original_requirement":"Req 8.3","constraint":"Legacy system","objective":"Protect credentials","compensating_control":"MFA enforced via proxy","validation":"Quarterly pen test","risk_assessment":"Low residual risk","maintenance_plan":"Annual review"}`

	mock.ExpectQuery(`SELECT is_compensating, compensating_worksheet::text FROM controls WHERE id = \$1 AND org_id = \$2`).
		WithArgs("ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"is_compensating", "compensating_worksheet"}).
			AddRow(true, worksheetJSON))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls/ctrl-001/compensating-worksheet", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "ctrl-001", data["control_id"])
	assert.True(t, data["is_compensating"].(bool))
	ws := data["worksheet"].(map[string]interface{})
	assert.Equal(t, "Req 8.3", ws["original_requirement"])
	assert.Equal(t, "MFA enforced via proxy", ws["compensating_control"])
}

func TestGetCompensatingWorksheet_NotCompensating(t *testing.T) {
	// A control that exists but has no worksheet yet.
	_, mock := setupTestRouter()
	r := setupWorksheetRouter(mock)

	mock.ExpectQuery(`SELECT is_compensating, compensating_worksheet::text FROM controls WHERE id = \$1 AND org_id = \$2`).
		WithArgs("ctrl-002", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"is_compensating", "compensating_worksheet"}).
			AddRow(false, nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls/ctrl-002/compensating-worksheet", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.False(t, data["is_compensating"].(bool))
	assert.Nil(t, data["worksheet"])
}

func TestGetCompensatingWorksheet_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupWorksheetRouter(mock)

	mock.ExpectQuery(`SELECT is_compensating, compensating_worksheet::text FROM controls WHERE id = \$1 AND org_id = \$2`).
		WithArgs("missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"is_compensating", "compensating_worksheet"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls/missing/compensating-worksheet", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// UpdateCompensatingWorksheet
// =============================================================================

func TestUpdateCompensatingWorksheet_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupWorksheetRouter(mock)

	returnedJSON := `{"original_requirement":"Req 8.3","constraint":"Legacy system","objective":"Protect credentials","compensating_control":"MFA enforced via proxy","validation":"Quarterly pen test","risk_assessment":"Low residual risk","maintenance_plan":"Annual review"}`

	mock.ExpectQuery(`UPDATE controls SET is_compensating = TRUE`).
		WithArgs(sqlmock.AnyArg(), "ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"is_compensating", "compensating_worksheet"}).
			AddRow(true, returnedJSON))

	body := `{
		"original_requirement": "Req 8.3",
		"constraint": "Legacy system",
		"objective": "Protect credentials",
		"compensating_control": "MFA enforced via proxy",
		"validation": "Quarterly pen test",
		"risk_assessment": "Low residual risk",
		"maintenance_plan": "Annual review"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/controls/ctrl-001/compensating-worksheet",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "ctrl-001", data["control_id"])
	assert.True(t, data["is_compensating"].(bool))
	ws := data["worksheet"].(map[string]interface{})
	assert.Equal(t, "Req 8.3", ws["original_requirement"])
}

func TestUpdateCompensatingWorksheet_PartialFields(t *testing.T) {
	// Partial update — only required fields filled in, rest are null.
	_, mock := setupTestRouter()
	r := setupWorksheetRouter(mock)

	returnedJSON := `{"original_requirement":"Req 6.3","constraint":null,"objective":null,"compensating_control":"WAF in blocking mode","validation":null,"risk_assessment":null,"maintenance_plan":null}`

	mock.ExpectQuery(`UPDATE controls SET is_compensating = TRUE`).
		WithArgs(sqlmock.AnyArg(), "ctrl-003", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"is_compensating", "compensating_worksheet"}).
			AddRow(true, returnedJSON))

	body := `{"original_requirement": "Req 6.3", "compensating_control": "WAF in blocking mode"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/controls/ctrl-003/compensating-worksheet",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.True(t, data["is_compensating"].(bool))
}

func TestUpdateCompensatingWorksheet_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupWorksheetRouter(mock)

	mock.ExpectQuery(`UPDATE controls SET is_compensating = TRUE`).
		WithArgs(sqlmock.AnyArg(), "missing", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"is_compensating", "compensating_worksheet"}))

	body := `{"original_requirement": "Req 8.3", "compensating_control": "MFA via proxy"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/controls/missing/compensating-worksheet",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

// =============================================================================
// ListControls — is_compensating filter
// =============================================================================

func TestListControls_FilterByIsCompensating(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupWorksheetRouter(mock)

	ctrlCols := []string{
		"id", "identifier", "title", "description", "category", "status", "is_custom",
		"owner_id", "owner_name", "owner_email", "secondary_owner_id",
		"mappings_count", "is_compensating", "compensating_worksheet",
		"created_at", "updated_at",
	}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM controls c WHERE`).
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .* FROM controls c`).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows(ctrlCols).
			AddRow("ctrl-001", "CTRL-001", "MFA Compensating", "desc", "technical", "active", true,
				nil, "", "", nil, 0, true, nil, cwNow, cwNow))

	// Frameworks query for the control
	mock.ExpectQuery(`SELECT DISTINCT f.name FROM control_mappings`).
		WithArgs("ctrl-001").
		WillReturnRows(sqlmock.NewRows([]string{"name"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/controls?is_compensating=true", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	require.Len(t, data, 1)
	assert.Equal(t, "ctrl-001", data[0].(map[string]interface{})["id"])
}
