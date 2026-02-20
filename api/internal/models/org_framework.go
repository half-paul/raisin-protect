package models

import "time"

// OrgFramework represents an org's activation of a framework.
type OrgFramework struct {
	ID              string     `json:"id"`
	OrgID           string     `json:"org_id"`
	FrameworkID     string     `json:"framework_id"`
	ActiveVersionID string     `json:"active_version_id"`
	Status          string     `json:"status"`
	TargetDate      *string    `json:"target_date"`
	Notes           *string    `json:"notes"`
	ActivatedAt     time.Time  `json:"activated_at"`
	DeactivatedAt   *time.Time `json:"deactivated_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ActivateFrameworkRequest is the request for activating a framework.
type ActivateFrameworkRequest struct {
	FrameworkID  string  `json:"framework_id" binding:"required"`
	VersionID    string  `json:"version_id" binding:"required"`
	TargetDate   *string `json:"target_date"`
	Notes        *string `json:"notes"`
	SeedControls *bool   `json:"seed_controls"`
}

// UpdateOrgFrameworkRequest is the request for updating an org framework.
type UpdateOrgFrameworkRequest struct {
	VersionID  *string `json:"version_id"`
	TargetDate *string `json:"target_date"`
	Notes      *string `json:"notes"`
	Status     *string `json:"status"`
}

// OrgFrameworkRoles can activate/deactivate frameworks.
var OrgFrameworkRoles = []string{RoleCISO, RoleComplianceManager}
