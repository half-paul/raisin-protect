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
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func evidenceAuthRouter(mock sqlmock.Sqlmock) *gin.Engine {
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(middleware.RequestID())
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u0000000-0000-0000-0000-000000000001")
		c.Set(middleware.ContextKeyOrgID, "a0000000-0000-0000-0000-000000000001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Set(middleware.ContextKeyEmail, "ciso@acme.com")
		c.Next()
	})
	return r
}

func TestCreateEvidence_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence", CreateEvidence)

	mock.ExpectExec("INSERT INTO evidence_artifacts").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"title": "Okta MFA Config Export",
		"evidence_type": "configuration_export",
		"file_name": "okta-mfa-config.json",
		"file_size": 15234,
		"mime_type": "application/json",
		"collection_date": "2026-02-15",
		"freshness_period_days": 90,
		"tags": ["mfa", "okta"]
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Okta MFA Config Export", data["title"])
	assert.Equal(t, "draft", data["status"])
	assert.Equal(t, float64(1), data["version"])
	assert.NotEmpty(t, data["object_key"])
}

func TestCreateEvidence_InvalidMIMEType(t *testing.T) {
	_, _ = setupTestRouter()
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u001")
		c.Set(middleware.ContextKeyOrgID, "a001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Next()
	})
	r.POST("/api/v1/evidence", CreateEvidence)

	body := `{
		"title": "Bad File",
		"evidence_type": "other",
		"file_name": "malware.exe",
		"file_size": 1024,
		"mime_type": "application/x-executable",
		"collection_date": "2026-02-15"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEvidence_FutureCollectionDate(t *testing.T) {
	_, _ = setupTestRouter()
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u001")
		c.Set(middleware.ContextKeyOrgID, "a001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Next()
	})
	r.POST("/api/v1/evidence", CreateEvidence)

	body := `{
		"title": "Future Evidence",
		"evidence_type": "screenshot",
		"file_name": "future.png",
		"file_size": 1024,
		"mime_type": "image/png",
		"collection_date": "2030-01-01"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateEvidence_FileTooLarge(t *testing.T) {
	_, _ = setupTestRouter()
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u001")
		c.Set(middleware.ContextKeyOrgID, "a001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Next()
	})
	r.POST("/api/v1/evidence", CreateEvidence)

	body := `{
		"title": "Huge File",
		"evidence_type": "audit_report",
		"file_name": "huge.pdf",
		"file_size": 200000000,
		"mime_type": "application/pdf",
		"collection_date": "2026-02-15"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEvidence_MissingRequired(t *testing.T) {
	_, _ = setupTestRouter()
	r := gin.New()
	gin.SetMode(gin.TestMode)
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "u001")
		c.Set(middleware.ContextKeyOrgID, "a001")
		c.Set(middleware.ContextKeyRole, "ciso")
		c.Next()
	})
	r.POST("/api/v1/evidence", CreateEvidence)

	body := `{"title": "Incomplete"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetEvidence_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.GET("/api/v1/evidence/:id", GetEvidence)

	now := time.Now()
	expires := now.Add(60 * 24 * time.Hour)
	freshDays := 90

	row := sqlmock.NewRows([]string{
		"id", "title", "description", "evidence_type", "status",
		"collection_method", "file_name", "file_size", "mime_type",
		"object_key", "checksum_sha256",
		"parent_artifact_id", "version", "is_current",
		"collection_date", "expires_at", "freshness_period_days",
		"source_system", "uploaded_by",
		"uploader_name", "uploader_email",
		"tags", "created_at", "updated_at",
	}).AddRow(
		"e001", "Okta MFA Config", nil, "configuration_export", "approved",
		"system_export", "okta.json", 15234, "application/json",
		"a001/e001/1/okta.json", nil,
		nil, 1, true,
		"2026-02-15", &expires, &freshDays,
		"okta", "u001",
		"Bob Security", "bob@acme.com",
		pq.Array([]string{"mfa", "okta"}), now, now,
	)
	mock.ExpectQuery("SELECT ea.id, ea.title").WillReturnRows(row)

	// Total versions
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Links
	mock.ExpectQuery("SELECT el.id, el.target_type").WillReturnRows(sqlmock.NewRows(nil))

	// Latest evaluation
	mock.ExpectQuery("SELECT ee.id, ee.verdict").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/evidence/e001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Okta MFA Config", data["title"])
	assert.Equal(t, "approved", data["status"])
	assert.Equal(t, float64(1), data["version"])
	assert.Equal(t, "fresh", data["freshness_status"])
}

func TestGetEvidence_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.GET("/api/v1/evidence/:id", GetEvidence)

	mock.ExpectQuery("SELECT ea.id, ea.title").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/evidence/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChangeEvidenceStatus_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.PUT("/api/v1/evidence/:id/status", ChangeEvidenceStatus)

	mock.ExpectQuery("SELECT status, title").WillReturnRows(
		sqlmock.NewRows([]string{"status", "title"}).AddRow("draft", "Test Evidence"),
	)
	mock.ExpectExec("UPDATE evidence_artifacts SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"status": "pending_review"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/evidence/e001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "pending_review", data["status"])
	assert.Equal(t, "draft", data["previous_status"])
}

func TestChangeEvidenceStatus_InvalidTransition(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.PUT("/api/v1/evidence/:id/status", ChangeEvidenceStatus)

	mock.ExpectQuery("SELECT status, title").WillReturnRows(
		sqlmock.NewRows([]string{"status", "title"}).AddRow("draft", "Test"),
	)

	body := `{"status": "approved"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/evidence/e001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestDeleteEvidence_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.DELETE("/api/v1/evidence/:id", DeleteEvidence)

	mock.ExpectQuery("SELECT title").WillReturnRows(
		sqlmock.NewRows([]string{"title"}).AddRow("Evidence to Delete"),
	)
	mock.ExpectExec("UPDATE evidence_artifacts SET status").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/evidence/e001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "superseded", data["status"])
}

func TestDeleteEvidence_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.DELETE("/api/v1/evidence/:id", DeleteEvidence)

	mock.ExpectQuery("SELECT title").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/evidence/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConfirmUpload_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/confirm", ConfirmEvidenceUpload)

	userID := "u0000000-0000-0000-0000-000000000001"
	mock.ExpectQuery("SELECT status, object_key, uploaded_by, file_size").WillReturnRows(
		sqlmock.NewRows([]string{"status", "object_key", "uploaded_by", "file_size"}).
			AddRow("draft", "a001/e001/1/test.pdf", &userID, 1024),
	)

	body := `{"checksum_sha256": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/confirm", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// MinIO not available in test, so it accepts without verification
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["file_verified"])
}

func TestConfirmUpload_AlreadyConfirmed(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/confirm", ConfirmEvidenceUpload)

	userID := "u0000000-0000-0000-0000-000000000001"
	mock.ExpectQuery("SELECT status, object_key, uploaded_by, file_size").WillReturnRows(
		sqlmock.NewRows([]string{"status", "object_key", "uploaded_by", "file_size"}).
			AddRow("pending_review", "a001/e001/1/test.pdf", &userID, 1024),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/confirm", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestListEvidenceLinks_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.GET("/api/v1/evidence/:id/links", ListEvidenceLinks)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)
	mock.ExpectQuery("SELECT el.id, el.target_type").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/evidence/e001/links", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 0)
}

func TestCreateEvidenceLink_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/links", CreateEvidenceLinks)

	// Artifact exists
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)
	// Control exists
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(true),
	)
	// No duplicate
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"exists"}).AddRow(false),
	)
	// Insert
	mock.ExpectExec("INSERT INTO evidence_links").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"target_type": "control",
		"control_id": "c001",
		"strength": "primary",
		"notes": "Direct evidence"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["created"])
}

func TestCreateEvidenceLink_Duplicate(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/links", CreateEvidenceLinks)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true)) // duplicate!

	body := `{"target_type": "control", "control_id": "c001"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDeleteEvidenceLink_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.DELETE("/api/v1/evidence/:id/links/:lid", DeleteEvidenceLink)

	controlID := "c001"
	mock.ExpectQuery("SELECT target_type, control_id, requirement_id").WillReturnRows(
		sqlmock.NewRows([]string{"target_type", "control_id", "requirement_id"}).
			AddRow("control", &controlID, nil),
	)
	mock.ExpectExec("DELETE FROM evidence_links").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/evidence/e001/links/l001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteEvidenceLink_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.DELETE("/api/v1/evidence/:id/links/:lid", DeleteEvidenceLink)

	mock.ExpectQuery("SELECT target_type").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/evidence/e001/links/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListEvidenceEvaluations_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.GET("/api/v1/evidence/:id/evaluations", ListEvidenceEvaluations)

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT ee.id, ee.verdict").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/evidence/e001/evaluations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateEvaluation_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/evaluations", CreateEvidenceEvaluation)

	mock.ExpectQuery("SELECT status").WillReturnRows(
		sqlmock.NewRows([]string{"status"}).AddRow("pending_review"),
	)
	mock.ExpectExec("INSERT INTO evidence_evaluations").WillReturnResult(sqlmock.NewResult(1, 1))
	// Auto status change
	mock.ExpectExec("UPDATE evidence_artifacts SET status").WillReturnResult(sqlmock.NewResult(1, 1))
	// Get evaluator name
	mock.ExpectQuery("SELECT first_name").WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("Alice CISO"),
	)

	body := `{
		"verdict": "sufficient",
		"confidence": "high",
		"comments": "Evidence fully satisfies MFA requirements."
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/evaluations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "sufficient", data["verdict"])
	assert.Equal(t, "high", data["confidence"])
}

func TestCreateEvaluation_InvalidVerdict(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/evaluations", CreateEvidenceEvaluation)

	mock.ExpectQuery("SELECT status").WillReturnRows(
		sqlmock.NewRows([]string{"status"}).AddRow("pending_review"),
	)

	body := `{"verdict": "maybe", "comments": "Not sure"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/evaluations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEvaluation_AutoReject(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/evaluations", CreateEvidenceEvaluation)

	mock.ExpectQuery("SELECT status").WillReturnRows(
		sqlmock.NewRows([]string{"status"}).AddRow("pending_review"),
	)
	mock.ExpectExec("INSERT INTO evidence_evaluations").WillReturnResult(sqlmock.NewResult(1, 1))
	// Auto status change to rejected
	mock.ExpectExec("UPDATE evidence_artifacts SET status").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT first_name").WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("Alice"),
	)

	body := `{
		"verdict": "insufficient",
		"comments": "Missing date and signature.",
		"missing_elements": ["date", "signature"],
		"remediation_notes": "Please add date and get it signed."
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/evaluations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "insufficient", data["verdict"])
	missing := data["missing_elements"].([]interface{})
	assert.Len(t, missing, 2)
}

func TestListEvidence_Empty(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.GET("/api/v1/evidence", ListEvidence)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT ea.id").WillReturnRows(sqlmock.NewRows(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/evidence", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 0)
}

func TestEvidenceStatusTransitions(t *testing.T) {
	tests := []struct {
		from  string
		to    string
		valid bool
	}{
		{"draft", "pending_review", true},
		{"draft", "approved", false},
		{"pending_review", "approved", true},
		{"pending_review", "rejected", true},
		{"pending_review", "draft", false},
		{"rejected", "pending_review", true},
		{"rejected", "approved", false},
		{"approved", "expired", true},
		{"approved", "draft", false},
		{"expired", "pending_review", true},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			transitions := map[string][]string{
				"draft":          {"pending_review"},
				"pending_review": {"approved", "rejected"},
				"rejected":       {"pending_review"},
				"approved":       {"expired"},
				"expired":        {"pending_review"},
			}
			allowed, ok := transitions[tt.from]
			valid := false
			if ok {
				for _, a := range allowed {
					if a == tt.to {
						valid = true
						break
					}
				}
			}
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestCreateEvidenceVersion_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/versions", CreateEvidenceVersion)

	freshDays := 90
	mock.ExpectQuery("SELECT title, evidence_type").WillReturnRows(
		sqlmock.NewRows([]string{"title", "evidence_type", "collection_method", "source_system", "freshness_period_days", "version"}).
			AddRow("MFA Config v1", "configuration_export", "system_export", "okta", &freshDays, 1),
	)
	mock.ExpectQuery("SELECT parent_artifact_id").WillReturnRows(
		sqlmock.NewRows([]string{"parent_artifact_id"}).AddRow(nil),
	)
	mock.ExpectExec("UPDATE evidence_artifacts SET is_current").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO evidence_artifacts").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO evidence_links").WillReturnResult(sqlmock.NewResult(0, 0))

	body := `{
		"file_name": "okta-mfa-config-v2.json",
		"file_size": 16500,
		"mime_type": "application/json",
		"collection_date": "2026-02-20"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/e001/versions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["version"])
	assert.Equal(t, true, data["is_current"])
	assert.Equal(t, "draft", data["status"])

	prevVersion := data["previous_version"].(map[string]interface{})
	assert.Equal(t, "superseded", prevVersion["status"])
}

func TestCreateEvidenceVersion_NotFound(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.POST("/api/v1/evidence/:id/versions", CreateEvidenceVersion)

	mock.ExpectQuery("SELECT title, evidence_type").WillReturnRows(sqlmock.NewRows(nil))

	body := `{
		"file_name": "test.pdf",
		"file_size": 1024,
		"mime_type": "application/pdf",
		"collection_date": "2026-02-20"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evidence/nonexistent/versions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFreshnessSummary_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.GET("/api/v1/evidence/freshness-summary", GetFreshnessSummary)

	orgID := "a0000000-0000-0000-0000-000000000001"

	mock.ExpectQuery("SELECT COUNT").WithArgs(orgID).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(42),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"fresh", "expiring_soon", "expired", "no_expiry"}).
			AddRow(30, 5, 3, 4),
	)
	mock.ExpectQuery("SELECT status, COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).
			AddRow("approved", 30).AddRow("draft", 2).AddRow("pending_review", 5),
	)
	mock.ExpectQuery("SELECT evidence_type, COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"evidence_type", "count"}).
			AddRow("configuration_export", 10).AddRow("screenshot", 8),
	)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(280),
	)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(210),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/evidence/freshness-summary", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(42), data["total_evidence"])

	coverage := data["coverage"].(map[string]interface{})
	assert.Equal(t, float64(280), coverage["total_active_controls"])
	assert.Equal(t, float64(210), coverage["controls_with_evidence"])
	assert.Equal(t, float64(75), coverage["evidence_coverage_pct"])
}

func TestListVersions_Success(t *testing.T) {
	_, mock := setupTestRouter()
	r := evidenceAuthRouter(mock)
	r.GET("/api/v1/evidence/:id/versions", ListEvidenceVersions)

	// Find root
	mock.ExpectQuery("SELECT id, parent_artifact_id").WillReturnRows(
		sqlmock.NewRows([]string{"id", "parent_artifact_id"}).AddRow("e001", nil),
	)

	now := time.Now()
	mock.ExpectQuery("SELECT ea.id, ea.version").WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "version", "is_current", "title", "status",
			"file_name", "file_size", "collection_date",
			"uploaded_by", "uploader_name", "created_at",
		}).
			AddRow("e002", 2, true, "V2 Title", "draft", "v2.json", 16500, "2026-02-20", nil, "", now).
			AddRow("e001", 1, false, "V1 Title", "superseded", "v1.json", 15000, "2026-02-15", nil, "", now),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/evidence/e001/versions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["total_versions"])
	assert.Equal(t, float64(2), meta["current_version"])
}
