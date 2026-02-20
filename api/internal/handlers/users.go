package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	internalAuth "github.com/half-paul/raisin-protect/api/internal/auth"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// ListUsers returns paginated users in the caller's organization.
func ListUsers(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	statusFilter := c.Query("status")
	roleFilter := c.Query("role")
	search := c.Query("search")
	sortField := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// Validate sort/order
	allowedSort := map[string]bool{"created_at": true, "email": true, "last_name": true, "role": true, "last_login_at": true}
	if !allowedSort[sortField] {
		sortField = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Build query
	where := []string{"org_id = $1"}
	args := []interface{}{orgID}
	argN := 2

	if statusFilter != "" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, statusFilter)
		argN++
	}
	if roleFilter != "" {
		where = append(where, fmt.Sprintf("role = $%d", argN))
		args = append(args, roleFilter)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause), countArgs...).Scan(&total)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count users")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Query users
	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, email, first_name, last_name, role, status, mfa_enabled, last_login_at, created_at, updated_at
		FROM users WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, order, argN, argN+1)
	args = append(args, perPage, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query users")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer rows.Close()

	users := []gin.H{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Role, &u.Status,
			&u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan user row")
			continue
		}
		user := gin.H{
			"id":         u.ID,
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"role":       u.Role,
			"status":     u.Status,
			"mfa_enabled": u.MFAEnabled,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		}
		if u.LastLoginAt != nil {
			user["last_login_at"] = u.LastLoginAt
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, listResponse(c, users, total, page, perPage))
}

// GetUser returns a specific user by ID.
func GetUser(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := c.Param("id")

	var u models.User
	err := database.QueryRow(`
		SELECT id, email, first_name, last_name, role, status, mfa_enabled, last_login_at, created_at, updated_at
		FROM users WHERE id = $1 AND org_id = $2
	`, userID, orgID).Scan(
		&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Role, &u.Status,
		&u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	resp := gin.H{
		"id":          u.ID,
		"email":       u.Email,
		"first_name":  u.FirstName,
		"last_name":   u.LastName,
		"role":        u.Role,
		"status":      u.Status,
		"mfa_enabled": u.MFAEnabled,
		"created_at":  u.CreatedAt,
		"updated_at":  u.UpdatedAt,
	}
	if u.LastLoginAt != nil {
		resp["last_login_at"] = u.LastLoginAt
	}

	c.JSON(http.StatusOK, successResponse(c, resp))
}

// CreateUser creates a new user in the organization.
func CreateUser(c *gin.Context) {
	orgID := middleware.GetOrgID(c)

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate
	if !internalAuth.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
			{"field": "email", "message": "must be a valid email address"},
		}))
		return
	}
	if len(req.FirstName) > 100 || len(req.LastName) > 100 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Name fields must be at most 100 characters"))
		return
	}
	if !models.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
			{"field": "role", "message": "must be a valid GRC role"},
		}))
		return
	}
	if err := internalAuth.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
			{"field": "password", "message": err.Error()},
		}))
		return
	}

	// Check email uniqueness within org
	var exists bool
	database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND org_id = $2)", req.Email, orgID).Scan(&exists)
	if exists {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Email already exists in this organization"))
		return
	}

	// Hash password
	passwordHash, err := internalAuth.HashPassword(req.Password, bcryptCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	userID := uuid.New().String()
	_, err = database.Exec(`
		INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'invited')
	`, userID, orgID, req.Email, passwordHash, req.FirstName, req.LastName, req.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	actorID := middleware.GetUserID(c)
	middleware.LogAudit(c, "user.register", "user", &userID, map[string]interface{}{
		"email":      req.Email,
		"role":       req.Role,
		"invited_by": actorID,
	})

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"id":         userID,
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"role":       req.Role,
		"status":     "invited",
		"mfa_enabled": false,
	}))
}

// UpdateUser updates a user's profile.
func UpdateUser(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	callerID := middleware.GetUserID(c)
	callerRole := middleware.GetUserRole(c)
	targetID := c.Param("id")

	// Check authorization: admins can update anyone, users can update themselves
	isAdmin := models.HasRole(callerRole, models.AdminRoles) || models.HasRole(callerRole, []string{models.RoleITAdmin})
	isSelf := callerID == targetID
	if !isAdmin && !isSelf {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Not authorized to update this user"))
		return
	}

	// Verify target exists in org
	var targetExists bool
	database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND org_id = $2)", targetID, orgID).Scan(&targetExists)
	if !targetExists {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found"))
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	changes := map[string]interface{}{}

	if req.FirstName != nil {
		if len(*req.FirstName) > 100 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "First name must be at most 100 characters"))
			return
		}
		database.Exec("UPDATE users SET first_name = $1 WHERE id = $2", *req.FirstName, targetID)
		changes["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		if len(*req.LastName) > 100 {
			c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Last name must be at most 100 characters"))
			return
		}
		database.Exec("UPDATE users SET last_name = $1 WHERE id = $2", *req.LastName, targetID)
		changes["last_name"] = *req.LastName
	}
	if req.Email != nil {
		if !internalAuth.ValidateEmail(*req.Email) {
			c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
				{"field": "email", "message": "must be a valid email address"},
			}))
			return
		}
		// Check uniqueness
		var emailExists bool
		database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND org_id = $2 AND id != $3)", *req.Email, orgID, targetID).Scan(&emailExists)
		if emailExists {
			c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Email already in use"))
			return
		}
		database.Exec("UPDATE users SET email = $1 WHERE id = $2", *req.Email, targetID)
		changes["email"] = *req.Email
	}

	if len(changes) > 0 {
		middleware.LogAudit(c, "user.updated", "user", &targetID, changes)
	}

	// Return updated user
	c.Params = gin.Params{{Key: "id", Value: targetID}}
	GetUser(c)
}

// DeactivateUser deactivates a user account.
func DeactivateUser(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	callerID := middleware.GetUserID(c)
	targetID := c.Param("id")

	if callerID == targetID {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Cannot deactivate yourself"))
		return
	}

	var status string
	err := database.QueryRow("SELECT status FROM users WHERE id = $1 AND org_id = $2", targetID, orgID).Scan(&status)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	_, err = database.Exec("UPDATE users SET status = 'deactivated' WHERE id = $1", targetID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to deactivate user")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Revoke all refresh tokens
	_, _ = database.Exec("UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL", targetID)

	middleware.LogAudit(c, "user.deactivated", "user", &targetID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":      targetID,
		"status":  "deactivated",
		"message": "User deactivated successfully",
	}))
}

// ReactivateUser reactivates a deactivated user.
func ReactivateUser(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	targetID := c.Param("id")

	var status string
	err := database.QueryRow("SELECT status FROM users WHERE id = $1 AND org_id = $2", targetID, orgID).Scan(&status)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user status")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	if status != "deactivated" {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "User is not deactivated"))
		return
	}

	_, err = database.Exec("UPDATE users SET status = 'active' WHERE id = $1", targetID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reactivate user")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "user.reactivated", "user", &targetID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":      targetID,
		"status":  "active",
		"message": "User reactivated successfully",
	}))
}

// ChangeUserRole changes a user's GRC role.
func ChangeUserRole(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	callerID := middleware.GetUserID(c)
	targetID := c.Param("id")

	if callerID == targetID {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "Cannot change your own role"))
		return
	}

	var req models.RoleChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	if !models.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
			{"field": "role", "message": "must be a valid GRC role"},
		}))
		return
	}

	var currentRole, email string
	err := database.QueryRow("SELECT role, email FROM users WHERE id = $1 AND org_id = $2", targetID, orgID).Scan(&currentRole, &email)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user role")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	if currentRole == req.Role {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "User already has this role"))
		return
	}

	_, err = database.Exec("UPDATE users SET role = $1 WHERE id = $2", req.Role, targetID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to change user role")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	middleware.LogAudit(c, "user.role_assigned", "user", &targetID, map[string]interface{}{
		"old_role": currentRole,
		"new_role": req.Role,
	})

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"id":            targetID,
		"email":         email,
		"role":          req.Role,
		"previous_role": currentRole,
		"message":       "Role updated successfully",
	}))
}
