package models

import "time"

// ControlMapping links a control to a requirement.
type ControlMapping struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	ControlID     string    `json:"control_id"`
	RequirementID string    `json:"requirement_id"`
	Strength      string    `json:"strength"`
	Notes         *string   `json:"notes"`
	MappedBy      *string   `json:"mapped_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// Valid mapping strengths.
var ValidMappingStrengths = []string{"primary", "supporting", "partial"}

// IsValidMappingStrength checks if the given strength is valid.
func IsValidMappingStrength(s string) bool {
	for _, v := range ValidMappingStrengths {
		if v == s {
			return true
		}
	}
	return false
}

// CreateMappingRequest is a single mapping creation request.
type CreateMappingRequest struct {
	RequirementID string  `json:"requirement_id" binding:"required"`
	Strength      *string `json:"strength"`
	Notes         *string `json:"notes"`
}

// BulkCreateMappingRequest supports both single and bulk mapping creation.
type BulkCreateMappingRequest struct {
	// Single mapping fields
	RequirementID string  `json:"requirement_id"`
	Strength      *string `json:"strength"`
	Notes         *string `json:"notes"`

	// Bulk mapping field
	Mappings []CreateMappingRequest `json:"mappings"`
}
