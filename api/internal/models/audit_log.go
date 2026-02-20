package models

import "time"

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID           string     `json:"id"`
	OrgID        string     `json:"org_id"`
	ActorID      *string    `json:"actor_id,omitempty"`
	ActorEmail   string     `json:"actor_email,omitempty"`
	ActorName    string     `json:"actor_name,omitempty"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *string    `json:"resource_id,omitempty"`
	Metadata     string     `json:"metadata"`
	IPAddress    *string    `json:"ip_address,omitempty"`
	UserAgent    *string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// AuditLogFilter holds query parameters for listing audit logs.
type AuditLogFilter struct {
	OrgID        string
	Action       string
	ActorID      string
	ResourceType string
	ResourceID   string
	From         *time.Time
	To           *time.Time
	Page         int
	PerPage      int
	Sort         string
	Order        string
}
