package models

import "time"

// GRC role constants from spec §1.2.
const (
	RoleCISO              = "ciso"
	RoleComplianceManager = "compliance_manager"
	RoleSecurityEngineer  = "security_engineer"
	RoleITAdmin           = "it_admin"
	RoleDevOpsEngineer    = "devops_engineer"
	RoleAuditor           = "auditor"
	RoleVendorManager     = "vendor_manager"
)

// AdminRoles can manage org and users.
var AdminRoles = []string{RoleCISO, RoleComplianceManager}

// UserCreateRoles can create new users.
var UserCreateRoles = []string{RoleCISO, RoleComplianceManager, RoleITAdmin}

// AuditViewRoles can view audit logs.
var AuditViewRoles = []string{RoleCISO, RoleComplianceManager, RoleAuditor}

// AllRoles is the complete list of GRC roles.
var AllRoles = []string{
	RoleCISO, RoleComplianceManager, RoleSecurityEngineer,
	RoleITAdmin, RoleDevOpsEngineer, RoleAuditor, RoleVendorManager,
}

// IsValidRole checks if a role string is a valid GRC role.
func IsValidRole(role string) bool {
	for _, r := range AllRoles {
		if r == role {
			return true
		}
	}
	return false
}

// HasRole checks if the given role is in the allowed list.
func HasRole(role string, allowed []string) bool {
	for _, r := range allowed {
		if r == role {
			return true
		}
	}
	return false
}

// User represents a user in the database.
type User struct {
	ID                string     `json:"id"`
	OrgID             string     `json:"org_id"`
	Email             string     `json:"email"`
	PasswordHash      string     `json:"-"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	Role              string     `json:"role"`
	Status            string     `json:"status"`
	MFAEnabled        bool       `json:"mfa_enabled"`
	MFASecret         *string    `json:"-"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UserResponse is the API response format (no sensitive fields).
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	MFAEnabled  bool       `json:"mfa_enabled"`
	OrgID       string     `json:"org_id,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts a User to its API response.
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Role:        u.Role,
		Status:      u.Status,
		MFAEnabled:  u.MFAEnabled,
		OrgID:       u.OrgID,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
}

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	OrgName   string `json:"org_name" binding:"required"`
}

// LoginRequest is the request body for login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the request body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest is the request body for logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest is the request body for changing password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// RoleChangeRequest is the request body for changing a user's role.
type RoleChangeRequest struct {
	Role string `json:"role" binding:"required"`
}
