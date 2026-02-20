package models

import "time"

// EvidenceLink represents a link between an evidence artifact and a control or requirement.
type EvidenceLink struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	ArtifactID    string    `json:"artifact_id"`
	TargetType    string    `json:"target_type"`
	ControlID     *string   `json:"control_id"`
	RequirementID *string   `json:"requirement_id"`
	Notes         *string   `json:"notes"`
	Strength      string    `json:"strength"`
	LinkedBy      *string   `json:"linked_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateLinkRequest is the request for creating a single evidence link.
type CreateLinkRequest struct {
	TargetType    string  `json:"target_type" binding:"required"`
	ControlID     *string `json:"control_id"`
	RequirementID *string `json:"requirement_id"`
	Strength      *string `json:"strength"`
	Notes         *string `json:"notes"`
}

// BulkCreateLinksRequest supports both single and bulk link creation.
type BulkCreateLinksRequest struct {
	// Single link fields
	TargetType    string  `json:"target_type"`
	ControlID     *string `json:"control_id"`
	RequirementID *string `json:"requirement_id"`
	Strength      *string `json:"strength"`
	Notes         *string `json:"notes"`

	// Bulk links
	Links []CreateLinkRequest `json:"links"`
}
