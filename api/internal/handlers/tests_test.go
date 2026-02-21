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
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Definition Tests ---

func TestCreateTest_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/tests", CreateTest)

	// Check identifier uniqueness
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("a0000000-0000-0000-0000-000000000001", "TST-AC-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Check control exists and is active
	mock.ExpectQuery("SELECT identifier, title, status FROM controls").
		WithArgs(sqlmock.AnyArg(), "a0000000-0000-0000-0000-000000000001").
		WillReturnRows(sqlmock.NewRows([]string{"identifier", "title", "status"}).
			AddRow("CTRL-AC-001", "Multi-Factor Authentication", "active"))

	// Insert
	mock.ExpectExec("INSERT INTO tests").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"identifier": "TST-AC-001",
		"title": "MFA Enforcement Verification",
		"test_type": "access_control",
		"severity": "critical",
		"control_id": "c0000000-0000-0000-0000-000000000001",
		"schedule_cron": "0 * * * *",
		"tags": ["mfa", "access-control"]
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "TST-AC-001", data["identifier"])
	assert.Equal(t, "MFA Enforcement Verification", data["title"])
	assert.Equal(t, "draft", data["status"])
	assert.Equal(t, "critical", data["severity"])
}

func TestCreateTest_DuplicateIdentifier(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/tests", CreateTest)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	body := `{
		"identifier": "TST-AC-001",
		"title": "Test",
		"test_type": "access_control",
		"control_id": "c001"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateTest_InvalidTestType(t *testing.T) {
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
	r.POST("/api/v1/tests", CreateTest)

	body := `{
		"identifier": "TST-XX-001",
		"title": "Test",
		"test_type": "invalid_type",
		"control_id": "c001"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTest_InactiveControl(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/tests", CreateTest)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(false),
	)
	mock.ExpectQuery("SELECT identifier, title, status FROM controls").
		WillReturnRows(sqlmock.NewRows([]string{"identifier", "title", "status"}).
			AddRow("CTRL-AC-001", "Test", "draft"))

	body := `{
		"identifier": "TST-AC-002",
		"title": "Test",
		"test_type": "access_control",
		"control_id": "c001"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetTest_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.GET("/api/v1/tests/:id", GetTest)

	now := time.Now()
	row := sqlmock.NewRows([]string{
		"id", "identifier", "title", "description", "test_type", "severity", "status",
		"control_id", "c_identifier", "c_title", "c_category", "c_status",
		"schedule_cron", "schedule_interval_min", "next_run_at", "last_run_at",
		"test_config", "test_script", "test_script_language",
		"timeout_seconds", "retry_count", "retry_delay_seconds",
		"tags", "created_by", "created_by_name",
		"created_at", "updated_at",
	}).AddRow(
		"t001", "TST-AC-001", "MFA Enforcement", "Check MFA", "access_control", "critical", "active",
		"c001", "CTRL-AC-001", "Multi-Factor Authentication", "technical", "active",
		"0 * * * *", nil, now, now,
		`{"check": "mfa_enforced"}`, nil, nil,
		120, 1, 30,
		pq.Array([]string{"mfa", "pci"}), "u001", "Alice Compliance",
		now, now,
	)
	mock.ExpectQuery("SELECT t.id, t.identifier, t.title").WillReturnRows(row)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tests/t001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "TST-AC-001", data["identifier"])
	assert.Equal(t, "active", data["status"])
	assert.Equal(t, "critical", data["severity"])
}

func TestGetTest_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.GET("/api/v1/tests/:id", GetTest)

	mock.ExpectQuery("SELECT t.id").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tests/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeTestStatus_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/tests/:id/status", ChangeTestStatus)

	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(
		sqlmock.NewRows([]string{"status", "identifier"}).AddRow("draft", "TST-AC-001"),
	)
	mock.ExpectExec("UPDATE tests SET").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"status": "active"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/tests/t001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "active", data["status"])
	assert.Equal(t, "draft", data["previous_status"])
}

func TestChangeTestStatus_InvalidTransition(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/tests/:id/status", ChangeTestStatus)

	mock.ExpectQuery("SELECT status, identifier").WillReturnRows(
		sqlmock.NewRows([]string{"status", "identifier"}).AddRow("deprecated", "TST-AC-001"),
	)

	body := `{"status": "active"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/tests/t001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestDeleteTest_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.DELETE("/api/v1/tests/:id", DeleteTest)

	mock.ExpectQuery("SELECT identifier").WillReturnRows(
		sqlmock.NewRows([]string{"identifier"}).AddRow("TST-AC-001"),
	)
	mock.ExpectExec("UPDATE tests SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/tests/t001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "deprecated", data["status"])
}

func TestDeleteTest_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.DELETE("/api/v1/tests/:id", DeleteTest)

	mock.ExpectQuery("SELECT identifier").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/tests/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Test Run Tests ---

func TestCreateTestRun_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/test-runs", CreateTestRun)

	// Check no existing run
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(0),
	)
	// Count active tests
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(8),
	)
	// Insert run
	mock.ExpectQuery("INSERT INTO test_runs").WillReturnRows(
		sqlmock.NewRows([]string{"run_number"}).AddRow(1),
	)

	body := `{}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test-runs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "pending", data["status"])
	assert.Equal(t, "manual", data["trigger_type"])
	assert.Equal(t, float64(8), data["total_tests"])
}

func TestCreateTestRun_AlreadyInProgress(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/test-runs", CreateTestRun)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(1),
	)

	body := `{}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test-runs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCancelTestRun_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/test-runs/:id/cancel", CancelTestRun)

	mock.ExpectQuery("SELECT status, run_number").WillReturnRows(
		sqlmock.NewRows([]string{"status", "run_number"}).AddRow("running", 42),
	)
	mock.ExpectExec("UPDATE test_runs SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test-runs/r001/cancel", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "cancelled", data["status"])
	assert.Equal(t, "running", data["previous_status"])
}

func TestCancelTestRun_AlreadyCompleted(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/test-runs/:id/cancel", CancelTestRun)

	mock.ExpectQuery("SELECT status, run_number").WillReturnRows(
		sqlmock.NewRows([]string{"status", "run_number"}).AddRow("completed", 42),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/test-runs/r001/cancel", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// --- Alert Tests ---

func TestChangeAlertStatus_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/alerts/:id/status", ChangeAlertStatus)

	mock.ExpectQuery("SELECT status, alert_number").WillReturnRows(
		sqlmock.NewRows([]string{"status", "alert_number"}).AddRow("open", 7),
	)
	mock.ExpectExec("UPDATE alerts SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"status": "acknowledged"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/a001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "acknowledged", data["status"])
	assert.Equal(t, "open", data["previous_status"])
}

func TestChangeAlertStatus_InvalidTransition(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/alerts/:id/status", ChangeAlertStatus)

	mock.ExpectQuery("SELECT status, alert_number").WillReturnRows(
		sqlmock.NewRows([]string{"status", "alert_number"}).AddRow("open", 7),
	)

	body := `{"status": "resolved"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/a001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestResolveAlert_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/alerts/:id/resolve", ResolveAlert)

	mock.ExpectQuery("SELECT status, alert_number").WillReturnRows(
		sqlmock.NewRows([]string{"status", "alert_number"}).AddRow("in_progress", 7),
	)
	mock.ExpectExec("UPDATE alerts SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"resolution_notes": "Fixed the security group rules."}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/a001/resolve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "resolved", data["status"])
}

func TestSuppressAlert_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.PUT("/api/v1/alerts/:id/suppress", SuppressAlert)

	mock.ExpectQuery("SELECT status, alert_number").WillReturnRows(
		sqlmock.NewRows([]string{"status", "alert_number"}).AddRow("open", 7),
	)
	mock.ExpectExec("UPDATE alerts SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"suppressed_until": "2026-03-20T16:00:00Z",
		"suppression_reason": "Known issue during scheduled infrastructure migration. Expected resolution by March 15."
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/a001/suppress", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuppressAlert_ReasonTooShort(t *testing.T) {
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
	r.PUT("/api/v1/alerts/:id/suppress", SuppressAlert)

	body := `{
		"suppressed_until": "2026-03-20T16:00:00Z",
		"suppression_reason": "Too short"
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/a001/suppress", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Alert Rule Tests ---

func TestCreateAlertRule_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/alert-rules", CreateAlertRule)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(false),
	)
	mock.ExpectExec("INSERT INTO alert_rules").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"name": "Critical Test Failures",
		"alert_severity": "critical",
		"delivery_channels": ["slack", "email", "in_app"],
		"match_severities": ["critical"],
		"consecutive_failures": 1,
		"cooldown_minutes": 60,
		"sla_hours": 4,
		"priority": 10
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/alert-rules", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Critical Test Failures", data["name"])
	assert.Equal(t, "critical", data["alert_severity"])
}

func TestCreateAlertRule_DuplicateName(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.POST("/api/v1/alert-rules", CreateAlertRule)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)

	body := `{
		"name": "Critical Test Failures",
		"alert_severity": "critical",
		"delivery_channels": ["in_app"]
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/alert-rules", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDeleteAlertRule_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.DELETE("/api/v1/alert-rules/:id", DeleteAlertRule)

	mock.ExpectQuery("SELECT name").WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("Critical Test Failures"),
	)
	mock.ExpectExec("DELETE FROM alert_rules").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/alert-rules/ar001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteAlertRule_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := setupAuthRouter(mock)
	r.DELETE("/api/v1/alert-rules/:id", DeleteAlertRule)

	mock.ExpectQuery("SELECT name").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/alert-rules/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Model Tests ---

func TestTestStatusTransitions(t *testing.T) {
	tests := []struct {
		from  string
		to    string
		valid bool
	}{
		{"draft", "active", true},
		{"draft", "paused", false},
		{"draft", "deprecated", false},
		{"active", "paused", true},
		{"active", "deprecated", true},
		{"active", "draft", false},
		{"paused", "active", true},
		{"paused", "deprecated", true},
		{"paused", "draft", false},
		{"deprecated", "active", false},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			assert.Equal(t, tt.valid, models.IsValidTestStatusTransition(tt.from, tt.to))
		})
	}
}

func TestAlertStatusTransitions(t *testing.T) {
	tests := []struct {
		from  string
		to    string
		valid bool
	}{
		{"open", "acknowledged", true},
		{"open", "in_progress", true},
		{"open", "suppressed", true},
		{"open", "closed", true},
		{"open", "resolved", false},
		{"acknowledged", "in_progress", true},
		{"acknowledged", "suppressed", true},
		{"acknowledged", "closed", true},
		{"in_progress", "resolved", true},
		{"in_progress", "suppressed", true},
		{"in_progress", "closed", true},
		{"resolved", "closed", true},
		{"resolved", "open", true},
		{"suppressed", "open", true},
		{"suppressed", "closed", true},
		{"closed", "open", true},
		{"closed", "resolved", false},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			assert.Equal(t, tt.valid, models.IsValidAlertStatusTransition(tt.from, tt.to))
		})
	}
}

func TestValidTestTypes(t *testing.T) {
	assert.True(t, models.IsValidTestType("configuration"))
	assert.True(t, models.IsValidTestType("access_control"))
	assert.True(t, models.IsValidTestType("custom"))
	assert.False(t, models.IsValidTestType("invalid"))
}

func TestValidAlertSeverities(t *testing.T) {
	assert.True(t, models.IsValidAlertSeverity("critical"))
	assert.True(t, models.IsValidAlertSeverity("high"))
	assert.True(t, models.IsValidAlertSeverity("medium"))
	assert.True(t, models.IsValidAlertSeverity("low"))
	assert.False(t, models.IsValidAlertSeverity("informational"))
}

func TestValidDeliveryChannels(t *testing.T) {
	assert.True(t, models.IsValidAlertDeliveryChannel("slack"))
	assert.True(t, models.IsValidAlertDeliveryChannel("email"))
	assert.True(t, models.IsValidAlertDeliveryChannel("webhook"))
	assert.True(t, models.IsValidAlertDeliveryChannel("in_app"))
	assert.False(t, models.IsValidAlertDeliveryChannel("sms"))
}
