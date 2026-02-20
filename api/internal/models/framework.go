package models

import "time"

// Framework represents a compliance framework in the system catalog.
type Framework struct {
	ID            string    `json:"id"`
	Identifier    string    `json:"identifier"`
	Name          string    `json:"name"`
	Description   *string   `json:"description"`
	Category      string    `json:"category"`
	WebsiteURL    *string   `json:"website_url"`
	LogoURL       *string   `json:"logo_url"`
	IsCustom      bool      `json:"is_custom"`
	VersionsCount int       `json:"versions_count,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// FrameworkVersion represents a specific version of a framework.
type FrameworkVersion struct {
	ID                  string    `json:"id"`
	FrameworkID         string    `json:"framework_id"`
	FrameworkIdentifier string    `json:"framework_identifier,omitempty"`
	FrameworkName       string    `json:"framework_name,omitempty"`
	Version             string    `json:"version"`
	DisplayName         string    `json:"display_name"`
	Status              string    `json:"status"`
	EffectiveDate       *string   `json:"effective_date"`
	SunsetDate          *string   `json:"sunset_date"`
	Changelog           *string   `json:"changelog,omitempty"`
	TotalRequirements   int       `json:"total_requirements"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Valid framework categories.
var ValidFrameworkCategories = []string{
	"security_privacy", "payment", "data_privacy", "ai_governance", "industry", "custom",
}

// IsValidFrameworkCategory checks if the given category is valid.
func IsValidFrameworkCategory(cat string) bool {
	for _, c := range ValidFrameworkCategories {
		if c == cat {
			return true
		}
	}
	return false
}
