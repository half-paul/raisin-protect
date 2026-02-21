package models

import "time"

// TestRun represents a batch test execution sweep.
type TestRun struct {
	ID              string     `json:"id"`
	OrgID           string     `json:"org_id"`
	RunNumber       int        `json:"run_number"`
	Status          string     `json:"status"`
	TriggerType     string     `json:"trigger_type"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	DurationMs      *int       `json:"duration_ms"`
	TotalTests      int        `json:"total_tests"`
	Passed          int        `json:"passed"`
	Failed          int        `json:"failed"`
	Errors          int        `json:"errors"`
	Skipped         int        `json:"skipped"`
	Warnings        int        `json:"warnings"`
	TriggeredBy     *string    `json:"triggered_by"`
	TriggerMetadata string     `json:"trigger_metadata"` // raw JSON
	WorkerID        *string    `json:"worker_id"`
	ErrorMessage    *string    `json:"error_message"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Valid test run statuses.
var ValidTestRunStatuses = []string{"pending", "running", "completed", "failed", "cancelled"}

// Valid trigger types.
var ValidTriggerTypes = []string{"scheduled", "manual", "on_change", "webhook"}

// TestRunCreateRoles can trigger manual test runs.
var TestRunCreateRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleDevOpsEngineer}

// TestRunCancelRoles can cancel test runs.
var TestRunCancelRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// CreateTestRunRequest is the request body for triggering a test run.
type CreateTestRunRequest struct {
	TestIDs         []string               `json:"test_ids"`
	TriggerMetadata map[string]interface{} `json:"trigger_metadata"`
}
