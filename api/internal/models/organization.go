package models

import "time"

// Organization represents a tenant organization.
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Domain    *string   `json:"domain,omitempty"`
	Status    string    `json:"status"`
	Settings  string    `json:"settings"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrganizationResponse is the API response format for organizations.
type OrganizationResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Domain    *string   `json:"domain,omitempty"`
	Status    string    `json:"status"`
	Settings  any       `json:"settings"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateOrganizationRequest is the request body for updating an organization.
type UpdateOrganizationRequest struct {
	Name     *string `json:"name"`
	Domain   *string `json:"domain"`
	Settings any     `json:"settings"`
}
