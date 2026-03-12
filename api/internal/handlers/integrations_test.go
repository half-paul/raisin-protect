package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/db"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupIntegrationRouter(role string) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	mockDB, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	database = &db.DB{DB: mockDB}
	middleware.SetAuditDB(nil)

	router := gin.New()
	router.Use(middleware.RequestID())

	protected := router.Group("/api/v1")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "test@acme.com")
		c.Set(middleware.ContextKeyRole, role)
		c.Next()
	})

	// Integration catalog
	integrations := protected.Group("/integrations")
	{
		integrations.GET("", ListIntegrations)
		integrations.GET("/dashboard", IntegrationDashboard)
		integrations.GET("/dashboard/sync-activity", IntegrationSyncActivity)
		integrations.GET("/:id", GetIntegration)
	}

	// Integration connections
	connections := protected.Group("/integration-connections")
	{
		connections.GET("", ListConnections)
		connections.GET("/:id", GetConnection)
		connections.POST("", CreateConnection)
		connections.PUT("/:id", UpdateConnection)
		connections.DELETE("/:id", DeleteConnection)
		connections.POST("/:id/test", TestConnection)
		connections.POST("/:id/enable", EnableConnection)
		connections.POST("/:id/disable", DisableConnection)
		connections.POST("/:id/sync", TriggerSync)
		connections.GET("/:id/runs", ListRuns)
		connections.GET("/:id/runs/:rid", GetRun)
		connections.POST("/:id/runs/:rid/cancel", CancelRun)
		connections.GET("/:id/runs/:rid/logs", GetRunLogs)
		connections.GET("/:id/health", GetConnectionHealth)
		connections.POST("/:id/health-check", TriggerHealthCheck)
		connections.GET("/:id/webhooks", ListWebhooks)
		connections.POST("/:id/webhooks", CreateWebhook)
		connections.DELETE("/:id/webhooks/:wid", DeleteWebhook)
		connections.POST("/:id/webhooks/:wid/rotate-secret", RotateWebhookSecret)
		connections.GET("/:id/preview", IntegrationPreview)
	}

	// Public webhook receiver
	router.POST("/api/v1/webhooks/receive/:id", ReceiveWebhook)

	return router, mock
}

func setupIntegrationRouterRegex(role string) (*gin.Engine, sqlmock.Sqlmock) {
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
		c.Set(middleware.ContextKeyRole, role)
		c.Next()
	})

	integrations := protected.Group("/integrations")
	{
		integrations.GET("", ListIntegrations)
		integrations.GET("/dashboard", IntegrationDashboard)
		integrations.GET("/dashboard/sync-activity", IntegrationSyncActivity)
		integrations.GET("/:id", GetIntegration)
	}

	connections := protected.Group("/integration-connections")
	{
		connections.GET("", ListConnections)
		connections.GET("/:id", GetConnection)
		connections.POST("", CreateConnection)
		connections.PUT("/:id", UpdateConnection)
		connections.DELETE("/:id", DeleteConnection)
		connections.POST("/:id/test", TestConnection)
		connections.POST("/:id/enable", EnableConnection)
		connections.POST("/:id/disable", DisableConnection)
		connections.POST("/:id/sync", TriggerSync)
		connections.GET("/:id/runs", ListRuns)
		connections.GET("/:id/runs/:rid", GetRun)
		connections.POST("/:id/runs/:rid/cancel", CancelRun)
		connections.GET("/:id/runs/:rid/logs", GetRunLogs)
		connections.GET("/:id/health", GetConnectionHealth)
		connections.POST("/:id/health-check", TriggerHealthCheck)
		connections.GET("/:id/webhooks", ListWebhooks)
		connections.POST("/:id/webhooks", CreateWebhook)
		connections.DELETE("/:id/webhooks/:wid", DeleteWebhook)
		connections.POST("/:id/webhooks/:wid/rotate-secret", RotateWebhookSecret)
		connections.GET("/:id/preview", IntegrationPreview)
	}

	router.POST("/api/v1/webhooks/receive/:id", ReceiveWebhook)

	return router, mock
}

// ========== Integration Catalog Tests ==========

func TestListIntegrations_Success(t *testing.T) {
	router, mock := setupIntegrationRouter("compliance_manager")

	mock.ExpectQuery("SELECT COUNT(*) FROM integration_definitions WHERE is_active = true").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery("SELECT id, name, slug, provider, category, description, short_description, icon_url, documentation_url, website_url, auth_type, config_schema, capabilities, is_active, is_beta, version, tags, created_at, updated_at FROM integration_definitions WHERE is_active = true ORDER BY name ASC LIMIT $1 OFFSET $2").
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "slug", "provider", "category", "description", "short_description",
			"icon_url", "documentation_url", "website_url", "auth_type", "config_schema",
			"capabilities", "is_active", "is_beta", "version", "tags", "created_at", "updated_at",
		}).AddRow(
			"def-001", "GitHub", "github", "github", "version_control", "GitHub integration", "Code monitoring",
			"/icons/github.svg", nil, nil, "oauth2", []byte("{}"),
			pq.Array([]string{"evidence_collection"}), true, false, "1.0.0", pq.Array([]string{"devops"}), time.Now(), time.Now(),
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/integrations", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetIntegration_Success(t *testing.T) {
	router, mock := setupIntegrationRouter("it_admin")

	mock.ExpectQuery("SELECT id, name, slug, provider, category, description, short_description, icon_url, documentation_url, website_url, auth_type, config_schema, capabilities, is_active, is_beta, version, tags, created_at, updated_at\n\t\tFROM integration_definitions WHERE id = $1").
		WithArgs("def-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "slug", "provider", "category", "description", "short_description",
			"icon_url", "documentation_url", "website_url", "auth_type", "config_schema",
			"capabilities", "is_active", "is_beta", "version", "tags", "created_at", "updated_at",
		}).AddRow(
			"def-001", "GitHub", "github", "github", "version_control", "GitHub integration", "Code monitoring",
			"/icons/github.svg", nil, nil, "oauth2", []byte("{}"),
			pq.Array([]string{"evidence_collection"}), true, false, "1.0.0", pq.Array([]string{"devops"}), time.Now(), time.Now(),
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/integrations/def-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetIntegration_NotFound(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	mock.ExpectQuery("SELECT id, name, slug, provider, category.*FROM integration_definitions WHERE id =.*").
		WithArgs("bad-id").
		WillReturnRows(sqlmock.NewRows([]string{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/integrations/bad-id", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ========== Connection CRUD Tests ==========

func TestCreateConnection_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	// Check definition exists
	mock.ExpectQuery("SELECT EXISTS.*FROM integration_definitions WHERE id =.*").
		WithArgs("def-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Insert
	mock.ExpectExec("INSERT INTO integration_connections.*").
		WithArgs(sqlmock.AnyArg(), "org-001", "def-001", "My GitHub", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false, 360, sqlmock.AnyArg(), sqlmock.AnyArg(), "user-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(map[string]interface{}{
		"definition_id": "def-001",
		"name":          "My GitHub",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateConnection_MissingName(t *testing.T) {
	router, _ := setupIntegrationRouter("it_admin")

	body, _ := json.Marshal(map[string]interface{}{
		"definition_id": "def-001",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteConnection_Success(t *testing.T) {
	router, mock := setupIntegrationRouter("ciso")

	mock.ExpectQuery("SELECT COUNT(*) FROM integration_runs WHERE connection_id = $1 AND org_id = $2 AND status IN ('pending', 'running')").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec("DELETE FROM integration_connections WHERE id = $1 AND org_id = $2").
		WithArgs("conn-001", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/integration-connections/conn-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteConnection_SyncInProgress(t *testing.T) {
	router, mock := setupIntegrationRouter("ciso")

	mock.ExpectQuery("SELECT COUNT(*) FROM integration_runs WHERE connection_id = $1 AND org_id = $2 AND status IN ('pending', 'running')").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/integration-connections/conn-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// ========== Connection Lifecycle Tests ==========

func TestEnableConnection_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	mock.ExpectExec("UPDATE integration_connections SET status.*WHERE id =.*AND org_id =.*AND status !=.*").
		WithArgs("conn-001", "org-001", "connected").
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDisableConnection_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	mock.ExpectExec("UPDATE integration_connections SET status.*sync_enabled.*WHERE id =.*AND org_id =.*").
		WithArgs("conn-001", "org-001", "disabled").
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Sync / Run Tests ==========

func TestTriggerSync_Success(t *testing.T) {
	router, mock := setupIntegrationRouter("it_admin")

	mock.ExpectQuery("SELECT status FROM integration_connections WHERE id = $1 AND org_id = $2").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("connected"))

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT COUNT(*) FROM integration_runs WHERE connection_id = $1 AND org_id = $2 AND status IN ('pending', 'running') FOR UPDATE").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec("INSERT INTO integration_runs (id, org_id, connection_id, trigger, triggered_by, status)\n\t\tVALUES ($1, $2, $3, $4, $5, $6)").
		WithArgs(sqlmock.AnyArg(), "org-001", "conn-001", "manual", "user-001", "pending").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE integration_connections SET total_runs = total_runs + 1, updated_at = NOW() WHERE id = $1 AND org_id = $2").
		WithArgs("conn-001", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/sync", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestTriggerSync_NotConnected(t *testing.T) {
	router, mock := setupIntegrationRouter("it_admin")

	mock.ExpectQuery("SELECT status FROM integration_connections WHERE id = $1 AND org_id = $2").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("disabled"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/sync", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTriggerSync_AlreadyRunning(t *testing.T) {
	router, mock := setupIntegrationRouter("it_admin")

	mock.ExpectQuery("SELECT status FROM integration_connections WHERE id = $1 AND org_id = $2").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("connected"))

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT COUNT(*) FROM integration_runs WHERE connection_id = $1 AND org_id = $2 AND status IN ('pending', 'running') FOR UPDATE").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectRollback()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/sync", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCancelRun_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	mock.ExpectExec("UPDATE integration_runs SET status.*WHERE id =.*AND connection_id =.*AND org_id =.*AND status IN.*").
		WithArgs("run-001", "conn-001", "org-001", "cancelled").
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/runs/run-001/cancel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Health Tests ==========

func TestTriggerHealthCheck_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	mock.ExpectQuery("SELECT EXISTS.*FROM integration_connections WHERE id =.*AND org_id =.*").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec("UPDATE integration_connections SET.*last_health_check_at.*health.*WHERE id =.*AND org_id =.*").
		WithArgs("conn-001", "org-001", sqlmock.AnyArg(), "healthy").
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/health-check", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "healthy", data["health"])
}

// ========== Webhook Tests ==========

func TestCreateWebhook_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	mock.ExpectQuery("SELECT EXISTS.*FROM integration_connections WHERE id =.*AND org_id =.*").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec("INSERT INTO integration_webhooks.*").
		WithArgs(sqlmock.AnyArg(), "org-001", "conn-001", "Push Events", sqlmock.AnyArg(), sqlmock.AnyArg(), "X-Hub-Signature-256", "sha256", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(map[string]interface{}{
		"name": "Push Events",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/webhooks", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data["webhook_secret"].(string), "whsec_")
}

func TestDeleteWebhook_Success(t *testing.T) {
	router, mock := setupIntegrationRouter("it_admin")

	mock.ExpectExec("DELETE FROM integration_webhooks WHERE id = $1 AND connection_id = $2 AND org_id = $3").
		WithArgs("wh-001", "conn-001", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/integration-connections/conn-001/webhooks/wh-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRotateWebhookSecret_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("ciso")

	mock.ExpectExec("UPDATE integration_webhooks SET webhook_secret.*WHERE id =.*AND connection_id =.*AND org_id =.*").
		WithArgs("wh-001", "conn-001", "org-001", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/webhooks/wh-001/rotate-secret", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data["webhook_secret"].(string), "whsec_")
}

// ========== Webhook Receiver (HMAC) Tests ==========

func TestReceiveWebhook_ValidSignature(t *testing.T) {
	router, mock := setupIntegrationRouter("ciso") // role doesn't matter for public endpoint

	secret := "test_secret_123"
	payload := []byte(`{"event":"push","repo":"acme/api"}`)

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	mock.ExpectQuery("SELECT webhook_secret, signature_header, signature_algo, status, connection_id, org_id\n\t\tFROM integration_webhooks WHERE id = $1").
		WithArgs("wh-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"webhook_secret", "signature_header", "signature_algo", "status", "connection_id", "org_id",
		}).AddRow(secret, "X-Hub-Signature-256", "sha256", "active", "conn-001", "org-001"))

	mock.ExpectExec("UPDATE integration_webhooks SET\n\t\t\ttotal_received = total_received + 1, total_processed = total_processed + 1,\n\t\t\tlast_received_at = $2, updated_at = NOW()\n\t\tWHERE id = $1").
		WithArgs("wh-001", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/receive/wh-001", bytes.NewBuffer(payload))
	req.Header.Set("X-Hub-Signature-256", "sha256="+sig)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiveWebhook_InvalidSignature(t *testing.T) {
	router, mock := setupIntegrationRouter("ciso")

	mock.ExpectQuery("SELECT webhook_secret, signature_header, signature_algo, status, connection_id, org_id\n\t\tFROM integration_webhooks WHERE id = $1").
		WithArgs("wh-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"webhook_secret", "signature_header", "signature_algo", "status", "connection_id", "org_id",
		}).AddRow("real_secret", "X-Hub-Signature-256", "sha256", "active", "conn-001", "org-001"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/receive/wh-001", bytes.NewBufferString(`{"event":"push"}`))
	req.Header.Set("X-Hub-Signature-256", "sha256=bad_signature")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReceiveWebhook_MissingSignature(t *testing.T) {
	router, mock := setupIntegrationRouter("ciso")

	mock.ExpectQuery("SELECT webhook_secret, signature_header, signature_algo, status, connection_id, org_id\n\t\tFROM integration_webhooks WHERE id = $1").
		WithArgs("wh-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"webhook_secret", "signature_header", "signature_algo", "status", "connection_id", "org_id",
		}).AddRow("real_secret", "X-Hub-Signature-256", "sha256", "active", "conn-001", "org-001"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/receive/wh-001", bytes.NewBufferString(`{}`))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReceiveWebhook_InactiveWebhook(t *testing.T) {
	router, mock := setupIntegrationRouter("ciso")

	mock.ExpectQuery("SELECT webhook_secret, signature_header, signature_algo, status, connection_id, org_id\n\t\tFROM integration_webhooks WHERE id = $1").
		WithArgs("wh-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"webhook_secret", "signature_header", "signature_algo", "status", "connection_id", "org_id",
		}).AddRow("secret", "X-Hub-Signature-256", "sha256", "inactive", "conn-001", "org-001"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/receive/wh-001", bytes.NewBufferString(`{}`))
	req.Header.Set("X-Hub-Signature-256", "sha256=abc")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Dashboard Tests ==========

func TestIntegrationDashboard_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("compliance_manager")

	mock.ExpectQuery("SELECT COUNT.*FILTER.*FROM integration_connections WHERE org_id =.*").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"total", "connected", "disconnected", "error", "disabled"}).AddRow(3, 2, 0, 1, 0))

	mock.ExpectQuery("SELECT COUNT.*FILTER.*FROM integration_connections WHERE org_id =.*AND status =.*").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"healthy", "degraded", "unhealthy"}).AddRow(2, 0, 0))

	mock.ExpectQuery("SELECT COUNT.*FROM integration_runs.*WHERE org_id =.*AND status =.*AND created_at.*").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT COUNT.*FROM integration_runs.*WHERE org_id =.*AND status IN.*").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT COUNT.*FROM integration_connections.*WHERE org_id =.*AND sync_enabled.*").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/integrations/dashboard", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Secret Masking Tests ==========

func TestMaskSecrets(t *testing.T) {
	config := map[string]interface{}{
		"domain":    "acme.okta.com",
		"api_token": "very_secret_token_12345",
		"password":  "mypassword",
		"region":    "us-east-1",
	}

	maskSecrets(config)

	assert.Equal(t, "acme.okta.com", config["domain"])
	assert.Equal(t, "us-east-1", config["region"])
	assert.Contains(t, config["api_token"].(string), "****")
	assert.NotEqual(t, "very_secret_token_12345", config["api_token"])
	assert.Contains(t, config["password"].(string), "****")
}

// ========== Multi-tenancy Tests ==========

func TestListConnections_OrgIsolation(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("compliance_manager")

	// The query should filter by org-001 from the JWT context
	mock.ExpectQuery("SELECT COUNT.*FROM integration_connections.*WHERE ic.org_id =.*").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT ic.id.*FROM integration_connections.*WHERE ic.org_id =.*ORDER BY.*LIMIT.*OFFSET.*").
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "definition_id", "name", "description", "instance_label",
			"status", "health", "config", "sync_enabled", "sync_interval_mins", "sync_cron",
			"next_sync_at", "last_sync_at", "last_sync_status", "last_sync_error",
			"last_health_check_at", "last_health_status", "consecutive_failures",
			"total_runs", "successful_runs", "failed_runs", "created_by",
			"tags", "metadata", "created_at", "updated_at",
			"def_name", "def_slug", "def_category", "def_icon_url",
		}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/integration-connections", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Test Connection Tests ==========

func TestTestConnection_Success(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	// getConnectionByID query
	mock.ExpectQuery("SELECT ic.id.*FROM integration_connections ic.*JOIN integration_definitions id.*WHERE ic.id =.*AND ic.org_id =.*").
		WithArgs("conn-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "definition_id", "name", "description", "instance_label",
			"status", "health", "config", "sync_enabled", "sync_interval_mins", "sync_cron",
			"next_sync_at", "last_sync_at", "last_sync_status", "last_sync_error",
			"last_health_check_at", "last_health_status", "consecutive_failures",
			"total_runs", "successful_runs", "failed_runs", "created_by",
			"tags", "metadata", "created_at", "updated_at",
			"def_name", "def_slug", "def_category", "def_icon_url",
		}).AddRow(
			"conn-001", "org-001", "def-001", "Test Conn", nil, nil,
			"pending", "unknown", []byte("{}"), false, 360, nil,
			nil, nil, nil, nil,
			nil, nil, 0,
			0, 0, 0, nil,
			pq.Array([]string{}), []byte("{}"), time.Now(), time.Now(),
			"GitHub", "github", "version_control", "",
		))

	// update health
	mock.ExpectExec("UPDATE integration_connections SET.*last_health_check_at.*health.*WHERE id =.*AND org_id =.*").
		WithArgs("conn-001", "org-001", sqlmock.AnyArg(), "healthy").
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections/conn-001/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Validation Tests ==========

func TestCreateConnection_InvalidDefinition(t *testing.T) {
	router, mock := setupIntegrationRouterRegex("it_admin")

	mock.ExpectQuery("SELECT EXISTS.*FROM integration_definitions WHERE id =.*").
		WithArgs("bad-def").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	body, _ := json.Marshal(map[string]interface{}{
		"definition_id": "bad-def",
		"name":          "Bad Connection",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/integration-connections", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
