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
	"github.com/half-paul/raisin-protect/api/internal/db"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func setupAccessReviewRouterWithRole(role string) (*gin.Engine, sqlmock.Sqlmock) {
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

	// Sprint 8 routes
	ar := protected.Group("/access-reviews")
	{
		idp := ar.Group("/identity-providers")
		{
			idp.GET("", ListIdentityProviders)
			idp.POST("", CreateIdentityProvider)
		}
		camp := ar.Group("/campaigns")
		{
			camp.POST("", CreateCampaign)
			camp.GET("/:id", GetCampaign)
			camp.PUT("/:id", UpdateCampaign)
			camp.POST("/:id/complete", CompleteCampaign)
			camp.POST("/:id/cancel", CancelCampaign)
			camp.GET("/:id/stats", GetCampaignStats)
			camp.GET("/:id/reviews", ListCampaignReviews)
			camp.GET("/:id/reviews/:rid", GetReviewDetail)
			camp.POST("/:id/reviews/:rid/decide", DecideReviewNested)
			camp.POST("/:id/reviews/bulk-decide", BulkDecideReviews)
			camp.POST("/:id/reviews/:rid/delegate", DelegateReview)
			camp.POST("/:id/reviews/:rid/escalate", EscalateReview)
			camp.POST("/:id/reviews/:rid/revocation", MarkRevocation)
			camp.GET("/:id/certification-report", GetCertificationReport)
		}
		ar.GET("/dashboard", GetAccessReviewDashboard)
	}

	return router, mock
}

func TestListIdentityProviders_Success(t *testing.T) {
	router, mock := setupAccessReviewRouterWithRole("compliance_manager")

	mock.ExpectQuery("SELECT COUNT(*) FROM identity_providers WHERE org_id = $1").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	query := "SELECT id, org_id, name, provider_type, status, config, last_sync_at, last_sync_status, last_sync_error, last_sync_stats, sync_interval_mins, description, created_at, updated_at FROM identity_providers WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"

	mock.ExpectQuery(query).
		WithArgs("org-001", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "name", "provider_type", "status", "config", "last_sync_at",
			"last_sync_status", "last_sync_error", "last_sync_stats", "sync_interval_mins",
			"description", "created_at", "updated_at",
		}).AddRow(
			"id-001", "org-001", "Okta", "okta", "connected", []byte("{}"), nil, nil, nil, []byte("{}"), 360, nil, time.Now(), time.Now(),
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/access-reviews/identity-providers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateCampaign_Success(t *testing.T) {
	router, mock := setupAccessReviewRouterWithRole("compliance_manager")

	query := "INSERT INTO access_review_campaigns ( id, org_id, name, description, status, cadence, scope, reviewer_strategy, default_reviewer_id, deadline, escalation_config, created_by, tags ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"

	mock.ExpectExec(query).
		WithArgs(sqlmock.AnyArg(), "org-001", "Q1 Review", sqlmock.AnyArg(), "draft", "quarterly", []byte("{}"), "resource_owner", sqlmock.AnyArg(), sqlmock.AnyArg(), []byte("{}"), "user-001", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(map[string]interface{}{
		"name":              "Q1 Review",
		"cadence":           "quarterly",
		"reviewer_strategy": "resource_owner",
		"deadline":          time.Now().Add(30 * 24 * time.Hour),
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func setupAccessReviewRouterRegex(role string) (*gin.Engine, sqlmock.Sqlmock) {
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

	ar := protected.Group("/access-reviews")
	{
		camp := ar.Group("/campaigns")
		{
			camp.POST("/:id/complete", CompleteCampaign)
			camp.POST("/:id/cancel", CancelCampaign)
			camp.POST("/:id/reviews/:rid/decide", DecideReviewNested)
			camp.POST("/:id/reviews/bulk-decide", BulkDecideReviews)
			camp.POST("/:id/reviews/:rid/delegate", DelegateReview)
			camp.POST("/:id/reviews/:rid/escalate", EscalateReview)
			camp.POST("/:id/reviews/:rid/revocation", MarkRevocation)
		}
		ar.GET("/dashboard", GetAccessReviewDashboard)
	}

	return router, mock
}

func TestCompleteCampaign_Success(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("compliance_manager")

	mock.ExpectQuery("SELECT status FROM access_review_campaigns").
		WithArgs("camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE access_reviews SET decision").
		WithArgs("camp-001", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 2))

	mock.ExpectExec("UPDATE access_review_campaigns").
		WithArgs("camp-001", "org-001", nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	mock.ExpectQuery("SELECT total_reviews").
		WithArgs("camp-001").
		WillReturnRows(sqlmock.NewRows([]string{"total", "completed", "approved", "revoked", "flagged"}).
			AddRow(10, 8, 6, 1, 1))

	body, _ := json.Marshal(map[string]interface{}{
		"expire_pending": true,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/complete", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "completed", data["status"])
}

func TestCompleteCampaign_InvalidState(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("compliance_manager")

	mock.ExpectQuery("SELECT status FROM access_review_campaigns").
		WithArgs("camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("draft"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/complete", bytes.NewBuffer([]byte("{}")))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelCampaign_Success(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("compliance_manager")

	mock.ExpectQuery("SELECT status FROM access_review_campaigns").
		WithArgs("camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE access_reviews SET decision").
		WithArgs("camp-001", "org-001").
		WillReturnResult(sqlmock.NewResult(0, 5))

	mock.ExpectExec("UPDATE access_review_campaigns").
		WithArgs("camp-001", "org-001", "Scope changed").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	body, _ := json.Marshal(map[string]interface{}{
		"reason": "Scope changed",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/cancel", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCancelCampaign_MissingReason(t *testing.T) {
	router, _ := setupAccessReviewRouterRegex("compliance_manager")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/cancel", bytes.NewBuffer([]byte("{}")))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDelegateReview_Success(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("compliance_manager")

	mock.ExpectQuery("SELECT decision FROM access_reviews").
		WithArgs("rev-001", "camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"decision"}).AddRow("pending"))

	mock.ExpectQuery("SELECT full_name FROM users").
		WithArgs("user-002", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"full_name"}).AddRow("Charlie TeamLead"))

	mock.ExpectExec("UPDATE access_reviews").
		WithArgs("user-002", "Not my area", "rev-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO access_reviews").
		WithArgs("user-002", "rev-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(map[string]interface{}{
		"delegate_to_id": "user-002",
		"reason":         "Not my area",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/reviews/rev-001/delegate", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "delegated", data["decision"])
	assert.Equal(t, "Charlie TeamLead", data["delegated_to_name"])
}

func TestMarkRevocation_Success(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("ciso")

	mock.ExpectQuery("SELECT decision, revocation_executed").
		WithArgs("rev-001", "camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"decision", "revocation_executed"}).AddRow("revoked", false))

	mock.ExpectExec("UPDATE access_reviews").
		WithArgs("Access removed", "rev-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(map[string]interface{}{
		"executed": true,
		"notes":    "Access removed",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/reviews/rev-001/revocation", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMarkRevocation_NotRevoked(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("ciso")

	mock.ExpectQuery("SELECT decision, revocation_executed").
		WithArgs("rev-001", "camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"decision", "revocation_executed"}).AddRow("approved", false))

	body, _ := json.Marshal(map[string]interface{}{
		"executed": true,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/reviews/rev-001/revocation", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBulkDecideReviews_EmptyArray(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("compliance_manager")

	mock.ExpectQuery("SELECT status FROM access_review_campaigns").
		WithArgs("camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))

	body, _ := json.Marshal(map[string]interface{}{
		"reviews": []interface{}{},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/reviews/bulk-decide", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEscalateReview_Success(t *testing.T) {
	router, mock := setupAccessReviewRouterRegex("ciso")

	mock.ExpectQuery("SELECT decision FROM access_reviews").
		WithArgs("rev-001", "camp-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"decision"}).AddRow("pending"))

	mock.ExpectQuery("SELECT full_name FROM users").
		WithArgs("user-003", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"full_name"}).AddRow("Diana Director"))

	mock.ExpectExec("UPDATE access_reviews").
		WithArgs("user-003", "rev-001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(map[string]interface{}{
		"escalate_to_id": "user-003",
		"reason":         "No response in 5 days",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/access-reviews/campaigns/camp-001/reviews/rev-001/escalate", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["is_escalated"])
}
