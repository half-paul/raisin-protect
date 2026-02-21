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

func setupPolicyRouter() (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()

	// Ensure auditDB is nil so LogAudit is a no-op in tests
	middleware.SetAuditDB(nil)

	// Protected group with auth context set
	protected := router.Group("/api/v1")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "test@acme.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	// Policy CRUD
	protected.GET("/policies", ListPolicies)
	protected.POST("/policies", CreatePolicy)
	protected.GET("/policies/search", SearchPolicies)
	protected.GET("/policies/stats", GetPolicyStats)
	protected.GET("/policies/:id", GetPolicy)
	protected.PUT("/policies/:id", UpdatePolicy)
	protected.POST("/policies/:id/archive", ArchivePolicy)
	protected.POST("/policies/:id/submit-for-review", SubmitForReview)
	protected.POST("/policies/:id/publish", PublishPolicy)

	// Policy Versions
	protected.GET("/policies/:id/versions", ListPolicyVersions)
	protected.GET("/policies/:id/versions/compare", CompareVersions)
	protected.GET("/policies/:id/versions/:version_number", GetPolicyVersion)
	protected.POST("/policies/:id/versions", CreatePolicyVersion)

	// Policy Sign-offs
	protected.GET("/policies/:id/signoffs", ListPolicySignoffs)
	protected.POST("/policies/:id/signoffs/:signoff_id/approve", ApproveSignoff)
	protected.POST("/policies/:id/signoffs/:signoff_id/reject", RejectSignoff)
	protected.POST("/policies/:id/signoffs/:signoff_id/withdraw", WithdrawSignoff)
	protected.POST("/policies/:id/signoffs/remind", RemindSignoffs)
	protected.GET("/signoffs/pending", ListPendingSignoffs)

	// Policy Controls
	protected.GET("/policies/:id/controls", ListPolicyControls)
	protected.POST("/policies/:id/controls", LinkPolicyControl)
	protected.POST("/policies/:id/controls/bulk", BulkLinkPolicyControls)
	protected.DELETE("/policies/:id/controls/:control_id", UnlinkPolicyControl)

	// Templates
	protected.GET("/policy-templates", ListPolicyTemplates)
	protected.POST("/policy-templates/:id/clone", ClonePolicyTemplate)

	// Gap detection
	protected.GET("/policy-gap", GetPolicyGap)
	protected.GET("/policy-gap/by-framework", GetPolicyGapByFramework)

	return router, mock
}

// ==================== Policy CRUD Tests ====================

func TestCreatePolicy_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Check identifier uniqueness
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "POL-TEST-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Insert policy
	mock.ExpectExec("INSERT INTO policies").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Insert version
	mock.ExpectExec("INSERT INTO policy_versions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Update current_version_id
	mock.ExpectExec("UPDATE policies SET current_version_id").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// Audit logs

	body := `{
		"identifier": "POL-TEST-001",
		"title": "Test Policy",
		"category": "information_security",
		"content": "<h1>Test Policy</h1><p>This is a test policy.</p>"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "POL-TEST-001", data["identifier"])
	assert.Equal(t, "Test Policy", data["title"])
	assert.Equal(t, "draft", data["status"])
	assert.NotNil(t, data["current_version"])
}

func TestCreatePolicy_DuplicateIdentifier(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "POL-DUP-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	body := `{
		"identifier": "POL-DUP-001",
		"title": "Duplicate Policy",
		"category": "access_control",
		"content": "<p>Content</p>"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "DUPLICATE_IDENTIFIER", errObj["code"])
}

func TestCreatePolicy_InvalidCategory(t *testing.T) {
	router, _ := setupPolicyRouter()

	body := `{
		"identifier": "POL-BAD-001",
		"title": "Bad Category",
		"category": "not_a_real_category",
		"content": "<p>Content</p>"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePolicy_MissingRequiredFields(t *testing.T) {
	router, _ := setupPolicyRouter()

	body := `{"title": "No identifier"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePolicy_ContentTooLarge(t *testing.T) {
	router, _ := setupPolicyRouter()

	// Generate content larger than 1MB
	largeContent := make([]byte, 1024*1024+1)
	for i := range largeContent {
		largeContent[i] = 'A'
	}

	body := `{
		"identifier": "POL-BIG-001",
		"title": "Big Policy",
		"category": "information_security",
		"content": "` + string(largeContent) + `"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePolicy_HTMLSanitization(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "POL-XSS-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// The content with script tag should be sanitized
	mock.ExpectExec("INSERT INTO policies").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO policy_versions").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE policies SET current_version_id").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := `{
		"identifier": "POL-XSS-001",
		"title": "XSS Test",
		"category": "information_security",
		"content": "<h1>Policy</h1><script>alert('xss')</script><p>Safe content</p>"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUpdatePolicy_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Fetch current policy
	mock.ExpectQuery("SELECT owner_id, status FROM policies").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id", "status"}).AddRow("user-001", "draft"))

	// Update
	mock.ExpectQuery("UPDATE policies SET").
		WillReturnRows(sqlmock.NewRows([]string{"id", "identifier", "title", "status", "updated_at"}).
			AddRow("policy-001", "POL-TEST-001", "Updated Title", "draft", time.Now()))

	// Audit log

	body := `{"title": "Updated Title"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/policies/policy-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Title", data["title"])
}

func TestUpdatePolicy_ArchivedPolicy(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT owner_id, status FROM policies").
		WithArgs("policy-archived", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id", "status"}).AddRow("user-001", "archived"))

	body := `{"title": "Can't update archived"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/policies/policy-archived", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestArchivePolicy_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT status FROM policies").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("published"))

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE policies SET status = 'archived'").
		WillReturnRows(sqlmock.NewRows([]string{"identifier", "updated_at"}).AddRow("POL-TEST-001", time.Now()))
	mock.ExpectExec("UPDATE policy_signoffs").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/archive", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "archived", data["status"])
}

func TestArchivePolicy_AlreadyArchived(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT status FROM policies").
		WithArgs("policy-archived", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("archived"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-archived/archive", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Policy Version Tests ====================

func TestCreatePolicyVersion_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Fetch policy
	mock.ExpectQuery("SELECT status, owner_id FROM policies").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "owner_id"}).AddRow("draft", "user-001"))

	mock.ExpectBegin()

	// Get max version
	mock.ExpectQuery("SELECT COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(1))

	// Mark previous as not current
	mock.ExpectExec("UPDATE policy_versions SET is_current = FALSE").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Insert new version
	mock.ExpectExec("INSERT INTO policy_versions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Update current version id
	mock.ExpectExec("UPDATE policies SET current_version_id").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// Audit log

	body := `{
		"content": "<h1>Updated Policy</h1><p>Version 2 content.</p>",
		"change_summary": "Updated section 3",
		"change_type": "minor"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/versions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["version_number"])
	assert.Equal(t, true, data["is_current"])
}

func TestCreatePolicyVersion_ArchivedPolicy(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT status, owner_id FROM policies").
		WithArgs("policy-archived", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "owner_id"}).AddRow("archived", "user-001"))

	body := `{
		"content": "<p>New version</p>",
		"change_summary": "Can't add to archived"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-archived/versions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreatePolicyVersion_RevertsApproved(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Policy is approved — should revert to draft after new version
	mock.ExpectQuery("SELECT status, owner_id FROM policies").
		WithArgs("policy-approved", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "owner_id"}).AddRow("approved", "user-001"))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))
	mock.ExpectExec("UPDATE policy_versions SET is_current = FALSE").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO policy_versions").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE policies SET current_version_id").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Revert status to draft
	mock.ExpectExec("UPDATE policies SET status = 'draft'").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := `{
		"content": "<p>New content</p>",
		"change_summary": "Major update"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-approved/versions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// ==================== Policy Sign-off Tests ====================

func TestSubmitForReview_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Fetch policy
	mock.ExpectQuery("SELECT status, owner_id, current_version_id FROM policies").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "owner_id", "current_version_id"}).
			AddRow("draft", "user-001", "version-001"))

	// Validate signers
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("signer-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectBegin()

	// Update status
	mock.ExpectExec("UPDATE policies SET status = 'in_review'").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Get signer info
	mock.ExpectQuery("SELECT role, first_name, last_name FROM users").
		WithArgs("signer-001").
		WillReturnRows(sqlmock.NewRows([]string{"role", "first_name", "last_name"}).AddRow("ciso", "David", "CISO"))

	// Insert signoff
	mock.ExpectExec("INSERT INTO policy_signoffs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	body := `{
		"signer_ids": ["signer-001"],
		"due_date": "2026-03-01"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/submit-for-review", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "in_review", data["status"])
	assert.Equal(t, float64(1), data["signoffs_created"])
}

func TestSubmitForReview_InvalidStatus(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Policy is published — can't submit for review
	mock.ExpectQuery("SELECT status, owner_id, current_version_id FROM policies").
		WithArgs("policy-pub", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "owner_id", "current_version_id"}).
			AddRow("published", "user-001", "version-001"))

	body := `{"signer_ids": ["signer-001"]}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-pub/submit-for-review", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitForReview_TooManySigners(t *testing.T) {
	router, _ := setupPolicyRouter()

	signers := make([]string, 11)
	for i := range signers {
		signers[i] = "signer-" + string(rune('0'+i))
	}
	bodyMap := map[string]interface{}{
		"signer_ids": signers,
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/submit-for-review", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApproveSignoff_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Fetch signoff — user is the signer
	mock.ExpectQuery("SELECT signer_id, status, policy_version_id").
		WithArgs("signoff-001", "policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"signer_id", "status", "policy_version_id"}).
			AddRow("user-001", "pending", "version-001"))

	// Update signoff
	mock.ExpectExec("UPDATE policy_signoffs SET status = 'approved'").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Audit

	// Check if all approved
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Get version number for approved_version
	mock.ExpectQuery("SELECT version_number FROM policy_versions").
		WillReturnRows(sqlmock.NewRows([]string{"version_number"}).AddRow(1))

	// Auto-approve policy
	mock.ExpectExec("UPDATE policies SET status = 'approved'").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Audit for policy status change

	body := `{"comments": "Looks good"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/signoffs/signoff-001/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "approved", data["status"])
	assert.Equal(t, true, data["all_signoffs_complete"])
	assert.Equal(t, "approved", data["policy_status"])
}

func TestApproveSignoff_NotSigner(t *testing.T) {
	router, mock := setupPolicyRouter()

	// User is not the signer
	mock.ExpectQuery("SELECT signer_id, status, policy_version_id").
		WithArgs("signoff-001", "policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"signer_id", "status", "policy_version_id"}).
			AddRow("other-user", "pending", "version-001"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/signoffs/signoff-001/approve", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRejectSignoff_RequiresComments(t *testing.T) {
	router, _ := setupPolicyRouter()

	body := `{}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/signoffs/signoff-001/reject", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	assert.Equal(t, "REJECTION_REQUIRES_COMMENTS", errObj["code"])
}

func TestRejectSignoff_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT signer_id, status").
		WithArgs("signoff-001", "policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"signer_id", "status"}).AddRow("user-001", "pending"))

	mock.ExpectExec("UPDATE policy_signoffs SET status = 'rejected'").
		WillReturnResult(sqlmock.NewResult(1, 1))


	body := `{"comments": "Section 4.2 needs revision"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/signoffs/signoff-001/reject", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "rejected", data["status"])
	assert.Equal(t, "in_review", data["policy_status"])
}

func TestPublishPolicy_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT status, review_frequency_days, current_version_id").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "review_frequency_days", "current_version_id"}).
			AddRow("approved", 365, "version-001"))

	// Get version number
	mock.ExpectQuery("SELECT version_number FROM policy_versions").
		WillReturnRows(sqlmock.NewRows([]string{"version_number"}).AddRow(1))

	// Update to published
	mock.ExpectQuery("UPDATE policies SET").
		WillReturnRows(sqlmock.NewRows([]string{"published_at"}).AddRow(time.Now()))

	// Audit

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/publish", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "published", data["status"])
}

func TestPublishPolicy_NotApproved(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT status, review_frequency_days, current_version_id").
		WithArgs("policy-draft", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "review_frequency_days", "current_version_id"}).
			AddRow("draft", nil, "version-001"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-draft/publish", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Policy Control Mapping Tests ====================

func TestLinkPolicyControl_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Get policy owner
	mock.ExpectQuery("SELECT owner_id FROM policies").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id"}).AddRow("user-001"))

	// Check control exists
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Check not already linked
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "policy-001", "ctrl-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Insert
	mock.ExpectExec("INSERT INTO policy_controls").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Audit

	// Get control info
	mock.ExpectQuery("SELECT identifier, title FROM controls").
		WillReturnRows(sqlmock.NewRows([]string{"identifier", "title"}).AddRow("CTRL-AC-001", "MFA"))

	body := `{"control_id": "ctrl-001", "coverage": "full", "notes": "Section 3.1"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/controls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLinkPolicyControl_AlreadyLinked(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT owner_id FROM policies").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id"}).AddRow("user-001"))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "policy-001", "ctrl-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := `{"control_id": "ctrl-001"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/policy-001/controls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUnlinkPolicyControl_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT owner_id FROM policies").
		WithArgs("policy-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id"}).AddRow("user-001"))

	mock.ExpectExec("DELETE FROM policy_controls").
		WillReturnResult(sqlmock.NewResult(0, 1))


	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/policies/policy-001/controls/ctrl-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// ==================== Policy Template Tests ====================

func TestCloneTemplate_Success(t *testing.T) {
	router, mock := setupPolicyRouter()

	// Fetch template — use sqlmock.AnyArg for flexible matching
	mock.ExpectQuery("SELECT identifier, title, description, category").
		WillReturnRows(sqlmock.NewRows([]string{
			"identifier", "title", "description", "category",
			"review_frequency_days", "tags", "current_version_id",
		}).AddRow("TPL-IS-001", "Info Security Template", "Template desc", "information_security",
			365, "{soc2,template}", "tpl-version-001"))

	// Check identifier uniqueness
	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Get template version content
	mock.ExpectQuery("SELECT content, content_format, word_count").
		WillReturnRows(sqlmock.NewRows([]string{"content", "content_format", "word_count"}).
			AddRow("<h1>Template Content</h1>", "html", 100))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO policies").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO policy_versions").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE policies SET current_version_id").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := `{
		"identifier": "POL-IS-001",
		"title": "Acme Corp IS Policy"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policy-templates/template-001/clone", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "POL-IS-001", data["identifier"])
	assert.Equal(t, "draft", data["status"])
	assert.Equal(t, "template-001", data["cloned_from_policy_id"])
}

func TestCloneTemplate_NotFound(t *testing.T) {
	router, mock := setupPolicyRouter()

	mock.ExpectQuery("SELECT identifier, title, description, category").
		WithArgs("nonexistent", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"identifier", "title", "description", "category",
			"review_frequency_days", "tags", "current_version_id",
		}))

	body := `{"identifier": "POL-NEW-001"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policy-templates/nonexistent/clone", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ==================== Search and Stats Tests ====================

func TestSearchPolicies_MissingQuery(t *testing.T) {
	router, _ := setupPolicyRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/policies/search", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompareVersions_SameVersion(t *testing.T) {
	router, _ := setupPolicyRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/policies/policy-001/versions/compare?v1=1&v2=1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompareVersions_MissingParams(t *testing.T) {
	router, _ := setupPolicyRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/policies/policy-001/versions/compare", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Helper Tests ====================

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		absent   string
	}{
		{
			name:   "strips script tags",
			input:  "<h1>Title</h1><script>alert('xss')</script><p>Content</p>",
			absent: "script",
		},
		{
			name:   "strips iframe tags",
			input:  "<p>Text</p><iframe src='evil.com'></iframe>",
			absent: "iframe",
		},
		{
			name:   "strips event handlers",
			input:  `<img src="test.jpg" onerror="alert('xss')">`,
			absent: "onerror",
		},
		{
			name:   "strips javascript URLs",
			input:  `<a href="javascript:alert('xss')">Click</a>`,
			absent: "javascript",
		},
		{
			name:     "preserves safe HTML",
			input:    "<h1>Title</h1><p>Paragraph</p><ul><li>Item</li></ul>",
			contains: "<h1>Title</h1>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeHTML(tt.input)
			if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			}
			if tt.absent != "" {
				assert.NotContains(t, result, tt.absent)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	assert.Equal(t, 3, countWords("<p>Hello world today</p>"))
	assert.Equal(t, 5, countWords("<h1>Title</h1><p>Three more words here.</p>"))
	assert.Equal(t, 0, countWords(""))
}

func TestComputeReviewStatus(t *testing.T) {
	assert.Equal(t, "no_schedule", computeReviewStatus(nil))

	past := time.Now().Add(-24 * time.Hour)
	assert.Equal(t, "overdue", computeReviewStatus(&past))

	soon := time.Now().Add(15 * 24 * time.Hour)
	assert.Equal(t, "due_soon", computeReviewStatus(&soon))

	future := time.Now().Add(60 * 24 * time.Hour)
	assert.Equal(t, "on_track", computeReviewStatus(&future))
}

func TestSuggestPolicyCategories(t *testing.T) {
	cats := suggestPolicyCategories("technical")
	require.NotEmpty(t, cats)
	assert.Contains(t, cats, "encryption")

	cats = suggestPolicyCategories("administrative")
	assert.Contains(t, cats, "compliance")

	cats = suggestPolicyCategories("unknown")
	assert.Contains(t, cats, "information_security")
}
