package models

import "time"

// Test represents a test definition in the database.
type Test struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"org_id"`
	Identifier          string     `json:"identifier"`
	Title               string     `json:"title"`
	Description         *string    `json:"description"`
	TestType            string     `json:"test_type"`
	Severity            string     `json:"severity"`
	Status              string     `json:"status"`
	ControlID           string     `json:"control_id"`
	ScheduleCron        *string    `json:"schedule_cron"`
	ScheduleIntervalMin *int       `json:"schedule_interval_min"`
	NextRunAt           *time.Time `json:"next_run_at"`
	LastRunAt           *time.Time `json:"last_run_at"`
	TestScript          *string    `json:"test_script,omitempty"`
	TestScriptLanguage  *string    `json:"test_script_language,omitempty"`
	TestConfig          string     `json:"test_config"` // raw JSON
	TimeoutSeconds      int        `json:"timeout_seconds"`
	RetryCount          int        `json:"retry_count"`
	RetryDelaySeconds   int        `json:"retry_delay_seconds"`
	Tags                []string   `json:"tags"`
	CreatedBy           *string    `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// Valid test types.
var ValidTestTypes = []string{
	"configuration", "access_control", "endpoint", "vulnerability",
	"data_protection", "network", "logging", "custom",
}

// Valid test statuses.
var ValidTestStatuses = []string{"draft", "active", "paused", "deprecated"}

// Valid test severities.
var ValidTestSeverities = []string{"critical", "high", "medium", "low", "informational"}

// Valid test status transitions.
var ValidTestStatusTransitions = map[string][]string{
	"draft":  {"active"},
	"active": {"paused", "deprecated"},
	"paused": {"active", "deprecated"},
	// deprecated is terminal
}

// TestCreateRoles can create/update tests.
var TestCreateRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer, RoleDevOpsEngineer}

// TestStatusRoles can change test status.
var TestStatusRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// TestDeleteRoles can soft-delete (deprecate) tests.
var TestDeleteRoles = []string{RoleCISO, RoleComplianceManager}

// IsValidTestType checks if the given type is valid.
func IsValidTestType(t string) bool {
	for _, v := range ValidTestTypes {
		if v == t {
			return true
		}
	}
	return false
}

// IsValidTestStatus checks if the given status is valid.
func IsValidTestStatus(s string) bool {
	for _, v := range ValidTestStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidTestSeverity checks if the given severity is valid.
func IsValidTestSeverity(s string) bool {
	for _, v := range ValidTestSeverities {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidTestStatusTransition checks if a test status transition is allowed.
func IsValidTestStatusTransition(from, to string) bool {
	allowed, ok := ValidTestStatusTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

// CreateTestRequest is the request body for creating a test.
type CreateTestRequest struct {
	Identifier          string                 `json:"identifier" binding:"required"`
	Title               string                 `json:"title" binding:"required"`
	Description         *string                `json:"description"`
	TestType            string                 `json:"test_type" binding:"required"`
	Severity            *string                `json:"severity"`
	ControlID           string                 `json:"control_id" binding:"required"`
	ScheduleCron        *string                `json:"schedule_cron"`
	ScheduleIntervalMin *int                   `json:"schedule_interval_min"`
	TimeoutSeconds      *int                   `json:"timeout_seconds"`
	RetryCount          *int                   `json:"retry_count"`
	RetryDelaySeconds   *int                   `json:"retry_delay_seconds"`
	TestConfig          map[string]interface{} `json:"test_config"`
	TestScript          *string                `json:"test_script"`
	TestScriptLanguage  *string                `json:"test_script_language"`
	Tags                []string               `json:"tags"`
}

// UpdateTestRequest is the request body for updating a test.
type UpdateTestRequest struct {
	Title               *string                `json:"title"`
	Description         *string                `json:"description"`
	Severity            *string                `json:"severity"`
	ScheduleCron        *string                `json:"schedule_cron"`
	ScheduleIntervalMin *int                   `json:"schedule_interval_min"`
	TimeoutSeconds      *int                   `json:"timeout_seconds"`
	RetryCount          *int                   `json:"retry_count"`
	RetryDelaySeconds   *int                   `json:"retry_delay_seconds"`
	TestConfig          map[string]interface{} `json:"test_config"`
	TestScript          *string                `json:"test_script"`
	TestScriptLanguage  *string                `json:"test_script_language"`
	Tags                []string               `json:"tags"`
}

// ChangeTestStatusRequest is the request body for changing a test's status.
type ChangeTestStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
