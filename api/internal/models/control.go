package models

import "time"

// Control represents a compliance control in the org's library.
type Control struct {
	ID                      string    `json:"id"`
	OrgID                   string    `json:"org_id"`
	Identifier              string    `json:"identifier"`
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	ImplementationGuidance  *string   `json:"implementation_guidance"`
	Category                string    `json:"category"`
	Status                  string    `json:"status"`
	OwnerID                 *string   `json:"owner_id"`
	SecondaryOwnerID        *string   `json:"secondary_owner_id"`
	EvidenceRequirements    *string   `json:"evidence_requirements"`
	TestCriteria            *string   `json:"test_criteria"`
	IsCustom                bool      `json:"is_custom"`
	SourceTemplateID        *string   `json:"source_template_id"`
	Metadata                string    `json:"metadata"` // raw JSON
	MappingsCount           int       `json:"mappings_count,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// Valid control categories.
var ValidControlCategories = []string{"technical", "administrative", "physical", "operational"}

// Valid control statuses.
var ValidControlStatuses = []string{"draft", "active", "under_review", "deprecated"}

// ControlCreateRoles can create controls.
var ControlCreateRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// ControlStatusRoles can change control status.
var ControlStatusRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// ControlMappingRoles can create/delete control mappings.
var ControlMappingRoles = []string{RoleCISO, RoleComplianceManager, RoleSecurityEngineer}

// IsValidControlCategory checks if the given category is valid.
func IsValidControlCategory(cat string) bool {
	for _, c := range ValidControlCategories {
		if c == cat {
			return true
		}
	}
	return false
}

// IsValidControlStatus checks if the given status is valid.
func IsValidControlStatus(status string) bool {
	for _, s := range ValidControlStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// ValidStatusTransitions defines allowed control status changes.
var ValidStatusTransitions = map[string][]string{
	"draft":        {"active", "deprecated"},
	"active":       {"under_review", "deprecated"},
	"under_review": {"active", "deprecated"},
	"deprecated":   {"draft"},
}

// IsValidStatusTransition checks if a status transition is valid.
func IsValidStatusTransition(from, to string) bool {
	allowed, ok := ValidStatusTransitions[from]
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

// CreateControlRequest is the request for creating a control.
type CreateControlRequest struct {
	Identifier             string                  `json:"identifier" binding:"required"`
	Title                  string                  `json:"title" binding:"required"`
	Description            string                  `json:"description" binding:"required"`
	ImplementationGuidance *string                 `json:"implementation_guidance"`
	Category               string                  `json:"category" binding:"required"`
	Status                 *string                 `json:"status"`
	OwnerID                *string                 `json:"owner_id"`
	SecondaryOwnerID       *string                 `json:"secondary_owner_id"`
	EvidenceRequirements   *string                 `json:"evidence_requirements"`
	TestCriteria           *string                 `json:"test_criteria"`
	Metadata               map[string]interface{}  `json:"metadata"`
}

// UpdateControlRequest is the request for updating a control.
type UpdateControlRequest struct {
	Title                  *string                 `json:"title"`
	Description            *string                 `json:"description"`
	ImplementationGuidance *string                 `json:"implementation_guidance"`
	Category               *string                 `json:"category"`
	EvidenceRequirements   *string                 `json:"evidence_requirements"`
	TestCriteria           *string                 `json:"test_criteria"`
	Metadata               map[string]interface{}  `json:"metadata"`
}

// ChangeOwnerRequest is the request for changing control ownership.
type ChangeOwnerRequest struct {
	OwnerID          *string `json:"owner_id"`
	SecondaryOwnerID *string `json:"secondary_owner_id"`
}

// ChangeStatusRequest is the request for changing control status.
type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// BulkStatusRequest is the request for bulk status change.
type BulkStatusRequest struct {
	ControlIDs []string `json:"control_ids" binding:"required"`
	Status     string   `json:"status" binding:"required"`
}
