package models

import "time"

// TestResult represents an individual test outcome within a test run.
type TestResult struct {
	ID             string     `json:"id"`
	OrgID          string     `json:"org_id"`
	TestRunID      string     `json:"test_run_id"`
	TestID         string     `json:"test_id"`
	ControlID      string     `json:"control_id"`
	Status         string     `json:"status"`
	Severity       string     `json:"severity"`
	Message        *string    `json:"message"`
	Details        string     `json:"details"`     // raw JSON
	OutputLog      *string    `json:"output_log,omitempty"`
	ErrorMessage   *string    `json:"error_message"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	DurationMs     *int       `json:"duration_ms"`
	AlertGenerated bool       `json:"alert_generated"`
	AlertID        *string    `json:"alert_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Valid test result statuses.
var ValidTestResultStatuses = []string{"pass", "fail", "error", "skip", "warning"}

// IsValidTestResultStatus checks if the given status is valid.
func IsValidTestResultStatus(s string) bool {
	for _, v := range ValidTestResultStatuses {
		if v == s {
			return true
		}
	}
	return false
}
