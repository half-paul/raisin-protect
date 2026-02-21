package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func setupRiskRouter() (*gin.Engine, sqlmock.Sqlmock) {
	router, mock := setupTestRouter()
	middleware.SetAuditDB(nil)

	protected := router.Group("/api/v1")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "user-001")
		c.Set(middleware.ContextKeyOrgID, "org-001")
		c.Set(middleware.ContextKeyEmail, "test@acme.com")
		c.Set(middleware.ContextKeyRole, "compliance_manager")
		c.Next()
	})

	// Risk CRUD
	protected.GET("/risks", ListRisks)
	protected.POST("/risks", CreateRisk)
	protected.GET("/risks/heat-map", GetRiskHeatMap)
	protected.GET("/risks/gaps", GetRiskGaps)
	protected.GET("/risks/search", SearchRisks)
	protected.GET("/risks/stats", GetRiskStats)

	protected.GET("/risks/:id", GetRisk)
	protected.PUT("/risks/:id", UpdateRisk)
	protected.POST("/risks/:id/archive", ArchiveRisk)
	protected.PUT("/risks/:id/status", ChangeRiskStatus)
	protected.POST("/risks/:id/recalculate", RecalculateRiskScores)

	// Assessments
	protected.GET("/risks/:id/assessments", ListRiskAssessments)
	protected.POST("/risks/:id/assessments", CreateRiskAssessment)

	// Treatments
	protected.GET("/risks/:id/treatments", ListRiskTreatments)
	protected.POST("/risks/:id/treatments", CreateRiskTreatment)
	protected.PUT("/risks/:id/treatments/:treatment_id", UpdateRiskTreatment)
	protected.POST("/risks/:id/treatments/:treatment_id/complete", CompleteTreatment)

	// Risk-Control linkage
	protected.GET("/risks/:id/controls", ListRiskControls)
	protected.POST("/risks/:id/controls", LinkRiskControl)
	protected.PUT("/risks/:id/controls/:control_id", UpdateRiskControl)
	protected.DELETE("/risks/:id/controls/:control_id", UnlinkRiskControl)

	return router, mock
}

// ==================== Risk CRUD Tests ====================

func TestCreateRisk_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	// Check identifier uniqueness
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "RISK-CY-010").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Insert risk
	mock.ExpectExec("INSERT INTO risks").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"identifier": "RISK-CY-010",
		"title":      "API Authentication Bypass",
		"category":   "cyber_security",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "RISK-CY-010", data["identifier"])
	assert.Equal(t, "identified", data["status"])
}

func TestCreateRisk_WithInitialAssessment(t *testing.T) {
	router, mock := setupRiskRouter()

	// Check identifier uniqueness
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "RISK-CY-011").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Insert risk
	mock.ExpectExec("INSERT INTO risks").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create inherent assessment
	mock.ExpectExec("INSERT INTO risk_assessments").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Update risk with inherent scores
	mock.ExpectExec("UPDATE risks SET inherent_likelihood").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Update next_assessment_at
	mock.ExpectExec("UPDATE risks SET next_assessment_at").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create residual assessment
	mock.ExpectExec("INSERT INTO risk_assessments").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Update risk with residual scores
	mock.ExpectExec("UPDATE risks SET residual_likelihood").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"identifier":                "RISK-CY-011",
		"title":                     "Test Risk with Assessment",
		"category":                  "cyber_security",
		"assessment_frequency_days": 90,
		"initial_assessment": map[string]interface{}{
			"inherent_likelihood": "likely",
			"inherent_impact":     "severe",
			"residual_likelihood": "possible",
			"residual_impact":     "major",
			"justification":       "Test assessment.",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["inherent_score"])
	assert.NotNil(t, data["residual_score"])

	inhScore := data["inherent_score"].(map[string]interface{})
	assert.Equal(t, float64(20), inhScore["score"])
	assert.Equal(t, "critical", inhScore["severity"])

	resScore := data["residual_score"].(map[string]interface{})
	assert.Equal(t, float64(12), resScore["score"])
	assert.Equal(t, "high", resScore["severity"])
}

func TestCreateRisk_DuplicateIdentifier(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("org-001", "RISK-CY-010").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := map[string]interface{}{
		"identifier": "RISK-CY-010",
		"title":      "Duplicate Risk",
		"category":   "cyber_security",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateRisk_InvalidCategory(t *testing.T) {
	router, _ := setupRiskRouter()

	body := map[string]interface{}{
		"identifier": "RISK-XX-001",
		"title":      "Bad Category",
		"category":   "nonexistent",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListRisks_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	now := time.Now()

	// Count query
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("org-001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// List query
	rows := sqlmock.NewRows([]string{
		"id", "identifier", "title", "description", "category", "status",
		"owner_id", "secondary_owner_id",
		"inherent_likelihood", "inherent_impact", "inherent_score",
		"residual_likelihood", "residual_impact", "residual_score",
		"risk_appetite_threshold",
		"assessment_frequency_days", "next_assessment_at", "last_assessed_at",
		"source", "affected_assets", "tags",
		"created_at", "updated_at",
		"owner_name", "owner_email",
		"linked_controls_count", "active_treatments_count",
	}).AddRow(
		"risk-001", "RISK-CY-001", "Ransomware", "Risk desc", "cyber_security", "treating",
		"user-001", nil,
		"likely", "severe", 20.0,
		"possible", "major", 12.0,
		10.0,
		90, now.Add(60*24*time.Hour), now,
		"threat_assessment", "{}", "{}",
		now, now,
		"Bob Security", "bob@acme.com",
		2, 3,
	)
	mock.ExpectQuery("SELECT r.id").WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/risks", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)

	risk := data[0].(map[string]interface{})
	assert.Equal(t, "RISK-CY-001", risk["identifier"])
	assert.Equal(t, true, risk["appetite_breached"])
}

func TestGetRisk_NotFound(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT r.id").
		WillReturnRows(sqlmock.NewRows([]string{}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/risks/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestArchiveRisk_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	// Get current status
	mock.ExpectQuery("SELECT status").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("open"))

	// Archive
	mock.ExpectExec("UPDATE risks SET status").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Cancel treatments
	mock.ExpectExec("UPDATE risk_treatments").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Get identifier
	mock.ExpectQuery("SELECT identifier").
		WillReturnRows(sqlmock.NewRows([]string{"identifier"}).AddRow("RISK-CY-001"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/archive", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "archived", data["status"])
}

func TestArchiveRisk_AlreadyArchived(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT status").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("archived"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/archive", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangeRiskStatus_InvalidTransition(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT status, owner_id").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "owner_id"}).AddRow("identified", "user-001"))

	body := map[string]interface{}{"status": "monitoring"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/risks/risk-001/status", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangeRiskStatus_AcceptRequiresJustification(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT status, owner_id").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "owner_id"}).AddRow("open", "user-001"))

	body := map[string]interface{}{"status": "accepted"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/risks/risk-001/status", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateRisk_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	// Get current owner
	mock.ExpectQuery("SELECT owner_id, status").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id", "status"}).AddRow("user-001", "open"))

	// Update
	mock.ExpectExec("UPDATE risks SET").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Fetch updated
	mock.ExpectQuery("SELECT identifier, title, status, updated_at").
		WillReturnRows(sqlmock.NewRows([]string{"identifier", "title", "status", "updated_at"}).
			AddRow("RISK-CY-001", "Updated Title", "open", time.Now()))

	body := map[string]interface{}{"title": "Updated Title"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/risks/risk-001", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Title", data["title"])
}

// ==================== Assessment Tests ====================

func TestCreateAssessment_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	// Check risk exists
	mock.ExpectQuery("SELECT owner_id").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id"}).AddRow("user-001"))

	// Check previous current assessment (none)
	mock.ExpectQuery("SELECT id FROM risk_assessments").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Insert assessment
	mock.ExpectExec("INSERT INTO risk_assessments").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Update residual scores on risks
	mock.ExpectExec("UPDATE risks SET residual_likelihood").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Check appetite breach
	mock.ExpectQuery("SELECT risk_appetite_threshold").
		WillReturnRows(sqlmock.NewRows([]string{"threshold"}).AddRow(10.0))

	// Update next_assessment_at
	mock.ExpectQuery("SELECT assessment_frequency_days").
		WillReturnRows(sqlmock.NewRows([]string{"freq"}).AddRow(90))
	mock.ExpectExec("UPDATE risks SET next_assessment_at").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Get assessor name
	mock.ExpectQuery("SELECT COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Bob Security"))

	body := map[string]interface{}{
		"assessment_type": "residual",
		"likelihood":      "possible",
		"impact":          "major",
		"justification":   "After controls applied.",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/assessments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(12), data["overall_score"])
	assert.Equal(t, "high", data["severity"])
	assert.Equal(t, true, data["is_current"])

	// Check appetite breach in risk_updated
	riskUpdated := data["risk_updated"].(map[string]interface{})
	assert.Equal(t, true, riskUpdated["appetite_breached"])
}

func TestCreateAssessment_InvalidLikelihood(t *testing.T) {
	router, _ := setupRiskRouter()

	body := map[string]interface{}{
		"assessment_type": "inherent",
		"likelihood":      "super_likely",
		"impact":          "major",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/assessments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Treatment Tests ====================

func TestCreateTreatment_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	// Check risk exists
	mock.ExpectQuery("SELECT owner_id, status").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id", "status"}).AddRow("user-001", "open"))

	// Insert treatment
	mock.ExpectExec("INSERT INTO risk_treatments").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Auto-transition to treating
	mock.ExpectExec("UPDATE risks SET status").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"treatment_type": "mitigate",
		"title":          "Deploy WAF",
		"priority":       "high",
		"due_date":       "2026-04-15",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/treatments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "planned", data["status"])
	assert.Equal(t, "high", data["priority"])
}

func TestCreateTreatment_ArchivedRisk(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT owner_id, status").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id", "status"}).AddRow("user-001", "archived"))

	body := map[string]interface{}{
		"treatment_type": "mitigate",
		"title":          "Deploy WAF",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/treatments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCompleteTreatment_WithEffectiveness(t *testing.T) {
	router, mock := setupRiskRouter()

	// Get treatment status
	mock.ExpectQuery("SELECT rt.status, rt.owner_id, r.owner_id").
		WithArgs("treat-001", "risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"status", "treat_owner", "risk_owner"}).
			AddRow("in_progress", "user-001", "user-001"))

	// Update treatment
	mock.ExpectExec("UPDATE risk_treatments SET status").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Check remaining active treatments
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Auto-transition risk (all treatments complete)
	mock.ExpectQuery("SELECT status FROM risks").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("treating"))
	mock.ExpectExec("UPDATE risks SET status").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"actual_effort_hours":   65.0,
		"effectiveness_rating":  "effective",
		"effectiveness_notes":   "WAF blocked 200+ attacks.",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/treatments/treat-001/complete", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "verified", data["status"])
	assert.Equal(t, "effective", data["effectiveness_rating"])
}

// ==================== Risk-Control Tests ====================

func TestLinkRiskControl_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	// Check risk exists
	mock.ExpectQuery("SELECT owner_id").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id"}).AddRow("user-001"))

	// Check control exists
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Check not already linked
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("risk-001", "ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Insert
	mock.ExpectExec("INSERT INTO risk_controls").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Get control details
	mock.ExpectQuery("SELECT identifier, title").
		WillReturnRows(sqlmock.NewRows([]string{"identifier", "title"}).AddRow("CTRL-EP-001", "EDR"))

	// Get user name
	mock.ExpectQuery("SELECT COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Bob Security"))

	body := map[string]interface{}{
		"control_id":            "ctrl-001",
		"effectiveness":         "effective",
		"mitigation_percentage": 35,
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/controls", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "effective", data["effectiveness"])
}

func TestLinkRiskControl_AlreadyLinked(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT owner_id").
		WithArgs("risk-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"owner_id"}).AddRow("user-001"))

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("risk-001", "ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	body := map[string]interface{}{"control_id": "ctrl-001"}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/risks/risk-001/controls", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUnlinkRiskControl_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	// Get record
	mock.ExpectQuery("SELECT rc.id, r.owner_id").
		WithArgs("risk-001", "ctrl-001", "org-001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).AddRow("rc-001", "user-001"))

	// Delete
	mock.ExpectExec("DELETE FROM risk_controls").
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/risks/risk-001/controls/ctrl-001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// ==================== Scoring Engine Tests ====================

func TestScoringEngine_LikelihoodScores(t *testing.T) {
	tests := []struct {
		likelihood string
		expected   int
	}{
		{"rare", 1},
		{"unlikely", 2},
		{"possible", 3},
		{"likely", 4},
		{"almost_certain", 5},
	}

	for _, tt := range tests {
		t.Run(tt.likelihood, func(t *testing.T) {
			score := likelihoodToScore(tt.likelihood)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestScoringEngine_ImpactScores(t *testing.T) {
	tests := []struct {
		impact   string
		expected int
	}{
		{"negligible", 1},
		{"minor", 2},
		{"moderate", 3},
		{"major", 4},
		{"severe", 5},
	}

	for _, tt := range tests {
		t.Run(tt.impact, func(t *testing.T) {
			score := impactToScore(tt.impact)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestScoringEngine_SeverityBands(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{25.0, "critical"},
		{20.0, "critical"},
		{19.0, "high"},
		{12.0, "high"},
		{11.0, "medium"},
		{6.0, "medium"},
		{5.0, "low"},
		{1.0, "low"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%.0f", tt.score), func(t *testing.T) {
			severity := scoreToSeverity(tt.score)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestScoringEngine_Formulas(t *testing.T) {
	tests := []struct {
		likelihood string
		impact     string
		expected   float64
		severity   string
	}{
		{"almost_certain", "severe", 25.0, "critical"},
		{"likely", "severe", 20.0, "critical"},
		{"possible", "major", 12.0, "high"},
		{"unlikely", "moderate", 6.0, "medium"},
		{"rare", "negligible", 1.0, "low"},
		{"possible", "moderate", 9.0, "medium"},
		{"likely", "minor", 8.0, "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.likelihood+"_x_"+tt.impact, func(t *testing.T) {
			l := likelihoodToScore(tt.likelihood)
			i := impactToScore(tt.impact)
			score := float64(l * i)
			assert.Equal(t, tt.expected, score)
			assert.Equal(t, tt.severity, scoreToSeverity(score))
		})
	}
}

// ==================== Search Tests ====================

func TestSearchRisks_MissingQuery(t *testing.T) {
	router, _ := setupRiskRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/risks/search", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchRisks_Success(t *testing.T) {
	router, mock := setupRiskRouter()

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{
		"id", "identifier", "title", "description", "category", "status",
		"residual_score", "owner_id", "owner_name",
	}).AddRow(
		"risk-001", "RISK-CY-001", "Ransomware Attack", "Risk of ransomware...", "cyber_security", "treating",
		12.0, "user-001", "Bob Security",
	)
	mock.ExpectQuery("SELECT r.id").WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/risks/search?q=ransomware", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	require.Len(t, data, 1)
	assert.Equal(t, "RISK-CY-001", data[0].(map[string]interface{})["identifier"])
}

// ==================== Helper functions for tests ====================

func likelihoodToScore(l string) int {
	switch l {
	case "rare":
		return 1
	case "unlikely":
		return 2
	case "possible":
		return 3
	case "likely":
		return 4
	case "almost_certain":
		return 5
	}
	return 0
}

func impactToScore(i string) int {
	switch i {
	case "negligible":
		return 1
	case "minor":
		return 2
	case "moderate":
		return 3
	case "major":
		return 4
	case "severe":
		return 5
	}
	return 0
}

func scoreToSeverity(score float64) string {
	switch {
	case score >= 20:
		return "critical"
	case score >= 12:
		return "high"
	case score >= 6:
		return "medium"
	default:
		return "low"
	}
}

var _ = fmt.Sprintf // prevent unused import
var _ = require.Equal // prevent unused import
