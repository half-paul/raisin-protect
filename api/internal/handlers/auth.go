package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	internalAuth "github.com/half-paul/raisin-protect/api/internal/auth"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog/log"
)

// Register creates a new user and organization.
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Validate fields
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
	if len(req.OrgName) > 255 {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Organization name must be at most 255 characters"))
		return
	}
	if err := internalAuth.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, errorResponseWithDetails("VALIDATION_ERROR", "Validation failed", []gin.H{
			{"field": "password", "message": err.Error()},
		}))
		return
	}

	// Check if email already exists
	var exists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check email existence")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Email already registered"))
		return
	}

	// Hash password
	passwordHash, err := internalAuth.HashPassword(req.Password, bcryptCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Create org and user in a transaction
	tx, err := database.Begin()
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}
	defer tx.Rollback()

	orgID := uuid.New().String()
	slug := generateSlug(req.OrgName)
	now := time.Now().UTC()

	_, err = tx.Exec(`
		INSERT INTO organizations (id, name, slug, status, settings, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', '{}', $4, $4)
	`, orgID, req.OrgName, slug, now)
	if err != nil {
		if strings.Contains(err.Error(), "uq_organizations_slug") || strings.Contains(err.Error(), "organizations_slug_key") {
			c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Organization name already taken"))
			return
		}
		log.Error().Err(err).Msg("Failed to create organization")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	userID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO users (id, org_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'compliance_manager', 'active', $7, $7)
	`, userID, orgID, req.Email, passwordHash, req.FirstName, req.LastName, now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Generate tokens
	tokenPair, err := jwtMgr.GenerateTokenPair(userID, orgID, req.Email, models.RoleComplianceManager)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Store refresh token hash
	storeRefreshToken(userID, orgID, tokenPair.RefreshToken, c)

	// Audit log
	middleware.LogAuditWithOrg(&orgID, &userID, "user.register", "user", &userID,
		map[string]interface{}{"email": req.Email}, c.ClientIP(), c.GetHeader("User-Agent"))
	middleware.LogAuditWithOrg(&orgID, nil, "org.created", "organization", &orgID,
		map[string]interface{}{"name": req.OrgName}, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusCreated, successResponse(c, gin.H{
		"user": gin.H{
			"id":         userID,
			"email":      req.Email,
			"first_name": req.FirstName,
			"last_name":  req.LastName,
			"role":       models.RoleComplianceManager,
			"status":     "active",
			"created_at": now.Format(time.RFC3339),
		},
		"organization": gin.H{
			"id":         orgID,
			"name":       req.OrgName,
			"slug":       slug,
			"status":     "active",
			"created_at": now.Format(time.RFC3339),
		},
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	}))
}

// Login authenticates a user with email/password.
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	var user models.User
	err := database.QueryRow(`
		SELECT id, org_id, email, password_hash, first_name, last_name, role, status, mfa_enabled
		FROM users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Role, &user.Status, &user.MFAEnabled,
	)
	if err == sql.ErrNoRows {
		// Audit failed login
		middleware.LogAuditWithOrg(nil, nil, "user.login_failed", "user", nil,
			map[string]interface{}{"email": req.Email, "reason": "user_not_found"},
			c.ClientIP(), c.GetHeader("User-Agent"))
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Invalid email or password"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to query user")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Check account status
	if user.Status != "active" {
		middleware.LogAuditWithOrg(&user.OrgID, &user.ID, "user.login_failed", "user", &user.ID,
			map[string]interface{}{"email": req.Email, "reason": "account_not_active"},
			c.ClientIP(), c.GetHeader("User-Agent"))
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Account is not active"))
		return
	}

	// Verify password
	if !internalAuth.CheckPassword(req.Password, user.PasswordHash) {
		middleware.LogAuditWithOrg(&user.OrgID, nil, "user.login_failed", "user", &user.ID,
			map[string]interface{}{"email": req.Email, "reason": "invalid_password"},
			c.ClientIP(), c.GetHeader("User-Agent"))
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Invalid email or password"))
		return
	}

	// Update last_login_at
	_, _ = database.Exec("UPDATE users SET last_login_at = NOW() WHERE id = $1", user.ID)

	// Generate tokens
	tokenPair, err := jwtMgr.GenerateTokenPair(user.ID, user.OrgID, user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Store refresh token hash
	storeRefreshToken(user.ID, user.OrgID, tokenPair.RefreshToken, c)

	// Audit log
	middleware.LogAuditWithOrg(&user.OrgID, &user.ID, "user.login", "user", &user.ID,
		map[string]interface{}{"email": user.Email, "method": "password"},
		c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
			"org_id":     user.OrgID,
			"status":     user.Status,
		},
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	}))
}

// RefreshToken exchanges a valid refresh token for a new token pair.
func RefreshToken(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	// Hash the provided token to look up in DB
	tokenHash := hashToken(req.RefreshToken)

	var tokenRecord models.RefreshToken
	var userRole, userEmail, userOrgID string
	err := database.QueryRow(`
		SELECT rt.id, rt.user_id, rt.org_id, rt.token_hash, rt.expires_at, rt.revoked_at,
			   u.role, u.email, u.org_id
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1
	`, tokenHash).Scan(
		&tokenRecord.ID, &tokenRecord.UserID, &tokenRecord.OrgID, &tokenRecord.TokenHash,
		&tokenRecord.ExpiresAt, &tokenRecord.RevokedAt,
		&userRole, &userEmail, &userOrgID,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Invalid refresh token"))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to query refresh token")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// If token was already revoked, this is potential theft — revoke all tokens for the user
	if tokenRecord.RevokedAt != nil {
		_, _ = database.Exec("UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL", tokenRecord.UserID)
		log.Warn().Str("user_id", tokenRecord.UserID).Msg("Refresh token reuse detected — all tokens revoked")
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Token has been revoked"))
		return
	}

	// Check expiry
	if time.Now().After(tokenRecord.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Refresh token expired"))
		return
	}

	// Revoke the old refresh token
	_, _ = database.Exec("UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1", tokenRecord.ID)

	// Generate new token pair
	tokenPair, err := jwtMgr.GenerateTokenPair(tokenRecord.UserID, userOrgID, userEmail, userRole)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Store new refresh token (inherits original expiry window, NOT extended)
	storeRefreshTokenWithExpiry(tokenRecord.UserID, tokenRecord.OrgID, tokenPair.RefreshToken, tokenRecord.ExpiresAt, c)

	// Audit
	middleware.LogAuditWithOrg(&userOrgID, &tokenRecord.UserID, "token.refreshed", "token", &tokenRecord.ID,
		nil, c.ClientIP(), c.GetHeader("User-Agent"))

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	}))
}

// Logout revokes the provided refresh token.
func Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	tokenHash := hashToken(req.RefreshToken)
	result, err := database.Exec("UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL", tokenHash)
	if err != nil {
		log.Error().Err(err).Msg("Failed to revoke refresh token")
	}

	if rows, _ := result.RowsAffected(); rows > 0 {
		middleware.LogAudit(c, "user.logout", "token", nil, nil)
	}

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"message": "Logged out successfully",
	}))
}

// ChangePassword changes the authenticated user's password.
func ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	userID := middleware.GetUserID(c)

	// Get current hash
	var currentHash string
	err := database.QueryRow("SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentHash)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user password")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Verify current password
	if !internalAuth.CheckPassword(req.CurrentPassword, currentHash) {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "Current password is incorrect"))
		return
	}

	// Validate new password
	if err := internalAuth.ValidatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusUnprocessableEntity, errorResponseWithDetails("UNPROCESSABLE", "Validation failed", []gin.H{
			{"field": "new_password", "message": err.Error()},
		}))
		return
	}

	// Check that new password differs from current
	if internalAuth.CheckPassword(req.NewPassword, currentHash) {
		c.JSON(http.StatusUnprocessableEntity, errorResponse("UNPROCESSABLE", "New password must be different from current password"))
		return
	}

	// Hash new password
	newHash, err := internalAuth.HashPassword(req.NewPassword, bcryptCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash new password")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Update password and revoke all refresh tokens
	_, err = database.Exec("UPDATE users SET password_hash = $1, password_changed_at = NOW() WHERE id = $2", newHash, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	// Revoke ALL refresh tokens (forces re-login on all devices)
	_, _ = database.Exec("UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL", userID)

	middleware.LogAudit(c, "user.password_changed", "user", &userID, nil)

	c.JSON(http.StatusOK, successResponse(c, gin.H{
		"message": "Password changed successfully",
	}))
}

// Helper: generate slug from text
func generateSlug(input string) string {
	slug := strings.ToLower(strings.TrimSpace(input))
	// Replace non-alphanumeric with hyphens
	result := make([]byte, 0, len(slug))
	for i := 0; i < len(slug); i++ {
		ch := slug[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			result = append(result, ch)
		} else {
			if len(result) > 0 && result[len(result)-1] != '-' {
				result = append(result, '-')
			}
		}
	}
	return strings.Trim(string(result), "-")
}

// Helper: hash a token using SHA-256
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Helper: store refresh token in the database
func storeRefreshToken(userID, orgID, rawToken string, c *gin.Context) {
	tokenHash := hashToken(rawToken)
	id := uuid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	_, err := database.Exec(`
		INSERT INTO refresh_tokens (id, user_id, org_id, token_hash, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::inet, $7, NOW())
	`, id, userID, orgID, tokenHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
	}
}

// Helper: store refresh token with a specific expiry (for rotation).
func storeRefreshTokenWithExpiry(userID, orgID, rawToken string, expiresAt time.Time, c *gin.Context) {
	tokenHash := hashToken(rawToken)
	id := uuid.New().String()
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	_, err := database.Exec(`
		INSERT INTO refresh_tokens (id, user_id, org_id, token_hash, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::inet, $7, NOW())
	`, id, userID, orgID, tokenHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
	}
}
