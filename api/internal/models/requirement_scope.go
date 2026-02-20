package models

import "time"

// RequirementScope represents an org's scoping decision for a requirement.
type RequirementScope struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	RequirementID string    `json:"requirement_id"`
	InScope       bool      `json:"in_scope"`
	Justification *string   `json:"justification"`
	ScopedBy      *string   `json:"scoped_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SetScopeRequest is the request for setting a requirement scope.
type SetScopeRequest struct {
	InScope       bool    `json:"in_scope"`
	Justification *string `json:"justification"`
}
