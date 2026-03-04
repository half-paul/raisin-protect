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
		}
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
