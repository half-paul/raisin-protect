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
)

func authMiddleware(userID, orgID, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, userID)
		c.Set(middleware.ContextKeyOrgID, orgID)
		c.Set(middleware.ContextKeyRole, role)
		c.Set(middleware.ContextKeyEmail, "test@example.com")
		c.Next()
	}
}

func TestListUsers_Success(t *testing.T) {
	router, mock := setupTestRouter()

	router.GET("/api/v1/users",
		authMiddleware("user-1", "org-1", "compliance_manager"),
		ListUsers,
	)

	// Count
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("org-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Users
	now := time.Now()
	mock.ExpectQuery("SELECT id, email, first_name, last_name, role, status, mfa_enabled, last_login_at, created_at, updated_at").
		WithArgs("org-1", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "mfa_enabled", "last_login_at", "created_at", "updated_at"}).
			AddRow("user-1", "alice@acme.com", "Alice", "Compliance", "compliance_manager", "active", false, now, now, now).
			AddRow("user-2", "bob@acme.com", "Bob", "Security", "security_engineer", "active", false, nil, now, now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)

	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["total"])
	assert.Equal(t, float64(1), meta["page"])
}

func TestListUsers_WithFilters(t *testing.T) {
	router, mock := setupTestRouter()

	router.GET("/api/v1/users",
		authMiddleware("user-1", "org-1", "compliance_manager"),
		ListUsers,
	)

	// Count with status filter
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("org-1", "active").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	now := time.Now()
	mock.ExpectQuery("SELECT id, email").
		WithArgs("org-1", "active", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "mfa_enabled", "last_login_at", "created_at", "updated_at"}).
			AddRow("user-1", "alice@acme.com", "Alice", "Compliance", "compliance_manager", "active", false, now, now, now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users?status=active", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUser_Success(t *testing.T) {
	router, mock := setupTestRouter()

	router.GET("/api/v1/users/:id",
		authMiddleware("user-1", "org-1", "compliance_manager"),
		GetUser,
	)

	now := time.Now()
	mock.ExpectQuery("SELECT id, email, first_name, last_name, role, status, mfa_enabled, last_login_at, created_at, updated_at").
		WithArgs("user-2", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "mfa_enabled", "last_login_at", "created_at", "updated_at"}).
			AddRow("user-2", "bob@acme.com", "Bob", "Security", "security_engineer", "active", false, now, now, now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/user-2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "bob@acme.com", data["email"])
	assert.Equal(t, "security_engineer", data["role"])
}

func TestGetUser_NotFound(t *testing.T) {
	router, mock := setupTestRouter()

	router.GET("/api/v1/users/:id",
		authMiddleware("user-1", "org-1", "compliance_manager"),
		GetUser,
	)

	mock.ExpectQuery("SELECT id, email").
		WithArgs("nonexistent", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "mfa_enabled", "last_login_at", "created_at", "updated_at"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateUser_Success(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/users",
		authMiddleware("admin-1", "org-1", "compliance_manager"),
		CreateUser,
	)

	// Check email uniqueness
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("newuser@acme.com", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Insert user
	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))

	// Audit log
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"email": "newuser@acme.com",
		"password": "SecureP@ss123",
		"first_name": "New",
		"last_name": "User",
		"role": "security_engineer"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "newuser@acme.com", data["email"])
	assert.Equal(t, "security_engineer", data["role"])
	assert.Equal(t, "invited", data["status"])
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/users",
		authMiddleware("admin-1", "org-1", "compliance_manager"),
		CreateUser,
	)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("existing@acme.com", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := `{
		"email": "existing@acme.com",
		"password": "SecureP@ss123",
		"first_name": "Existing",
		"last_name": "User",
		"role": "security_engineer"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateUser_InvalidRole(t *testing.T) {
	router, _ := setupTestRouter()

	router.POST("/api/v1/users",
		authMiddleware("admin-1", "org-1", "compliance_manager"),
		CreateUser,
	)

	body := `{
		"email": "newuser@acme.com",
		"password": "SecureP@ss123",
		"first_name": "New",
		"last_name": "User",
		"role": "superadmin"
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeactivateUser_Success(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/users/:id/deactivate",
		authMiddleware("admin-1", "org-1", "ciso"),
		DeactivateUser,
	)

	mock.ExpectQuery("SELECT status").
		WithArgs("user-2", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))

	mock.ExpectExec("UPDATE users SET status").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE refresh_tokens SET revoked_at").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/user-2/deactivate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "deactivated", data["status"])
}

func TestDeactivateUser_CannotSelf(t *testing.T) {
	router, _ := setupTestRouter()

	router.POST("/api/v1/users/:id/deactivate",
		authMiddleware("user-1", "org-1", "ciso"),
		DeactivateUser,
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/user-1/deactivate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestReactivateUser_Success(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/users/:id/reactivate",
		authMiddleware("admin-1", "org-1", "ciso"),
		ReactivateUser,
	)

	mock.ExpectQuery("SELECT status").
		WithArgs("user-2", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("deactivated"))

	mock.ExpectExec("UPDATE users SET status").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/user-2/reactivate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReactivateUser_NotDeactivated(t *testing.T) {
	router, mock := setupTestRouter()

	router.POST("/api/v1/users/:id/reactivate",
		authMiddleware("admin-1", "org-1", "ciso"),
		ReactivateUser,
	)

	mock.ExpectQuery("SELECT status").
		WithArgs("user-2", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/user-2/reactivate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestChangeUserRole_Success(t *testing.T) {
	router, mock := setupTestRouter()

	router.PUT("/api/v1/users/:id/role",
		authMiddleware("admin-1", "org-1", "ciso"),
		ChangeUserRole,
	)

	mock.ExpectQuery("SELECT role, email").
		WithArgs("user-2", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"role", "email"}).AddRow("security_engineer", "bob@acme.com"))

	mock.ExpectExec("UPDATE users SET role").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO audit_log").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"role": "ciso"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/users/user-2/role", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "ciso", data["role"])
	assert.Equal(t, "security_engineer", data["previous_role"])
}

func TestChangeUserRole_CannotSelf(t *testing.T) {
	router, _ := setupTestRouter()

	router.PUT("/api/v1/users/:id/role",
		authMiddleware("user-1", "org-1", "ciso"),
		ChangeUserRole,
	)

	body := `{"role": "auditor"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/users/user-1/role", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestChangeUserRole_SameRole(t *testing.T) {
	router, mock := setupTestRouter()

	router.PUT("/api/v1/users/:id/role",
		authMiddleware("admin-1", "org-1", "ciso"),
		ChangeUserRole,
	)

	mock.ExpectQuery("SELECT role, email").
		WithArgs("user-2", "org-1").
		WillReturnRows(sqlmock.NewRows([]string{"role", "email"}).AddRow("ciso", "bob@acme.com"))

	body := `{"role": "ciso"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/users/user-2/role", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
