package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/auth"
	"github.com/half-paul/raisin-protect/api/internal/db"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	mockDB, mock, _ := sqlmock.New()
	database = &db.DB{DB: mockDB}

	jwtMgr = auth.NewJWTManager(auth.JWTConfig{
		Secret:        "test-secret-key-for-unit-tests-32ch",
		AccessExpiry:  900000000000,  // 15 min in ns
		RefreshExpiry: 604800000000000, // 7 days in ns
		Issuer:        "test",
	})
	middleware.SetJWTManager(jwtMgr)
	bcryptCost = 4 // Low cost for fast tests

	router := gin.New()
	router.Use(middleware.RequestID())

	return router, mock
}

func TestHealthCheck(t *testing.T) {
	router, _ := setupTestRouter()
	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "0.1.0", resp["version"])
}

func TestRegister_Success(t *testing.T) {
	router, mock := setupTestRouter()
	router.POST("/api/v1/auth/register", Register)

	// Mock: check email doesn't exist
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("alice@acme.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock: transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO organizations").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Mock: store refresh token
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock: audit log entries
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"email": "alice@acme.com",
		"password": "SecureP@ss123",
		"first_name": "Alice",
		"last_name": "Compliance",
		"org_name": "Acme Corporation"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])

	user := data["user"].(map[string]interface{})
	assert.Equal(t, "alice@acme.com", user["email"])
	assert.Equal(t, "compliance_manager", user["role"])
}

func TestRegister_DuplicateEmail(t *testing.T) {
	router, mock := setupTestRouter()
	router.POST("/api/v1/auth/register", Register)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("alice@acme.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := `{
		"email": "alice@acme.com",
		"password": "SecureP@ss123",
		"first_name": "Alice",
		"last_name": "Compliance",
		"org_name": "Acme Corp"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_WeakPassword(t *testing.T) {
	router, _ := setupTestRouter()
	router.POST("/api/v1/auth/register", Register)

	body := `{
		"email": "alice@acme.com",
		"password": "weak",
		"first_name": "Alice",
		"last_name": "Compliance",
		"org_name": "Acme Corp"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_InvalidEmail(t *testing.T) {
	router, _ := setupTestRouter()
	router.POST("/api/v1/auth/register", Register)

	body := `{
		"email": "not-an-email",
		"password": "SecureP@ss123",
		"first_name": "Alice",
		"last_name": "Compliance",
		"org_name": "Acme Corp"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_MissingFields(t *testing.T) {
	router, _ := setupTestRouter()
	router.POST("/api/v1/auth/register", Register)

	body := `{"email": "alice@acme.com"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	router, mock := setupTestRouter()
	router.POST("/api/v1/auth/login", Login)

	// Hash for "SecureP@ss123" at cost 4
	hash, _ := auth.HashPassword("SecureP@ss123", 4)

	mock.ExpectQuery("SELECT id, org_id, email, password_hash, first_name, last_name, role, status, mfa_enabled").
		WithArgs("alice@acme.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "email", "password_hash", "first_name", "last_name", "role", "status", "mfa_enabled"}).
			AddRow("user-1", "org-1", "alice@acme.com", hash, "Alice", "Compliance", "compliance_manager", "active", false))

	mock.ExpectExec("UPDATE users SET last_login_at").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"email": "alice@acme.com", "password": "SecureP@ss123"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])

	user := data["user"].(map[string]interface{})
	assert.Equal(t, "alice@acme.com", user["email"])
}

func TestLogin_WrongPassword(t *testing.T) {
	router, mock := setupTestRouter()
	router.POST("/api/v1/auth/login", Login)

	hash, _ := auth.HashPassword("SecureP@ss123", 4)

	mock.ExpectQuery("SELECT id, org_id").
		WithArgs("alice@acme.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "email", "password_hash", "first_name", "last_name", "role", "status", "mfa_enabled"}).
			AddRow("user-1", "org-1", "alice@acme.com", hash, "Alice", "Compliance", "compliance_manager", "active", false))

	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"email": "alice@acme.com", "password": "WrongPassword1!"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	router, mock := setupTestRouter()
	router.POST("/api/v1/auth/login", Login)

	mock.ExpectQuery("SELECT id, org_id").
		WithArgs("nobody@acme.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "email", "password_hash", "first_name", "last_name", "role", "status", "mfa_enabled"}))

	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"email": "nobody@acme.com", "password": "SecureP@ss123"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_InactiveAccount(t *testing.T) {
	router, mock := setupTestRouter()
	router.POST("/api/v1/auth/login", Login)

	hash, _ := auth.HashPassword("SecureP@ss123", 4)

	mock.ExpectQuery("SELECT id, org_id").
		WithArgs("alice@acme.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "email", "password_hash", "first_name", "last_name", "role", "status", "mfa_enabled"}).
			AddRow("user-1", "org-1", "alice@acme.com", hash, "Alice", "Compliance", "compliance_manager", "deactivated", false))

	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"email": "alice@acme.com", "password": "SecureP@ss123"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestChangePassword_Success(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/auth/change-password",
		func(c *gin.Context) {
			c.Set(middleware.ContextKeyUserID, "user-1")
			c.Set(middleware.ContextKeyOrgID, "org-1")
			c.Set(middleware.ContextKeyRole, "compliance_manager")
			c.Next()
		},
		ChangePassword,
	)

	currentHash, _ := auth.HashPassword("OldP@ss123", 4)

	mock.ExpectQuery("SELECT password_hash").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(currentHash))

	mock.ExpectExec("UPDATE users SET password_hash").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE refresh_tokens SET revoked_at").WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"current_password": "OldP@ss123", "new_password": "NewSecureP@ss456"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Password changed successfully", data["message"])
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/auth/change-password",
		func(c *gin.Context) {
			c.Set(middleware.ContextKeyUserID, "user-1")
			c.Set(middleware.ContextKeyOrgID, "org-1")
			c.Next()
		},
		ChangePassword,
	)

	currentHash, _ := auth.HashPassword("OldP@ss123", 4)

	mock.ExpectQuery("SELECT password_hash").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(currentHash))

	body := `{"current_password": "WrongP@ss123", "new_password": "NewSecureP@ss456"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestChangePassword_SamePassword(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/auth/change-password",
		func(c *gin.Context) {
			c.Set(middleware.ContextKeyUserID, "user-1")
			c.Set(middleware.ContextKeyOrgID, "org-1")
			c.Next()
		},
		ChangePassword,
	)

	currentHash, _ := auth.HashPassword("SecureP@ss123", 4)

	mock.ExpectQuery("SELECT password_hash").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(currentHash))

	body := `{"current_password": "SecureP@ss123", "new_password": "SecureP@ss123"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestSlugGeneration(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Acme Corporation", "acme-corporation"},
		{"My Cool Org!", "my-cool-org"},
		{"  spaces  ", "spaces"},
		{"already-slug", "already-slug"},
		{"MiXeD CaSe 123", "mixed-case-123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"SecureP@ss123", true},
		{"short", false},           // Too short
		{"alllowercase1!", false},   // No uppercase (note: has ! which is special)
		{"ALLUPPERCASE1!", false},   // No lowercase
		{"NoDigits@Here", false},    // No digit
		{"NoSpecial123Ab", false},   // No special char
		{"ValidP@ss1", true},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := auth.ValidatePassword(tt.password)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJWTTokenPair(t *testing.T) {
	mgr := auth.NewJWTManager(auth.JWTConfig{
		Secret:        "test-secret-key-for-unit-tests-32ch",
		AccessExpiry:  900000000000,
		RefreshExpiry: 604800000000000,
		Issuer:        "test",
	})

	pair, err := mgr.GenerateTokenPair("user-1", "org-1", "test@example.com", "compliance_manager")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)

	// Validate access token
	claims, err := mgr.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)
	assert.Equal(t, "org-1", claims.OrgID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "compliance_manager", claims.Role)

	// Validate refresh token
	rClaims, err := mgr.ValidateRefreshToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, "user-1", rClaims.UserID)

	// Access token should NOT validate as refresh
	_, err = mgr.ValidateRefreshToken(pair.AccessToken)
	assert.Error(t, err)

	// Refresh token should NOT validate as access
	_, err = mgr.ValidateAccessToken(pair.RefreshToken)
	assert.Error(t, err)
}
