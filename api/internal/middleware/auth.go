package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/auth"
)

// Context keys for authenticated user info.
const (
	ContextKeyUserID = "user_id"
	ContextKeyOrgID  = "org_id"
	ContextKeyEmail  = "email"
	ContextKeyRole   = "role"
	ContextKeyClaims = "claims"
)

var jwtManager *auth.JWTManager

// SetJWTManager configures the JWT manager for auth middleware.
func SetJWTManager(manager *auth.JWTManager) {
	jwtManager = manager
}

// GetJWTManager returns the configured JWT manager.
func GetJWTManager() *auth.JWTManager {
	return jwtManager
}

// AuthRequired validates JWT Bearer tokens and populates context with user claims.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if jwtManager == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "Authentication not configured"},
			})
			return
		}

		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Authorization header required"},
			})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid authorization header format"},
			})
			return
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid or expired token"},
			})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyOrgID, claims.OrgID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// GetUserID extracts the user ID from context.
func GetUserID(c *gin.Context) string {
	if v, ok := c.Get(ContextKeyUserID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetOrgID extracts the org ID from context.
func GetOrgID(c *gin.Context) string {
	if v, ok := c.Get(ContextKeyOrgID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetUserRole extracts the user role from context.
func GetUserRole(c *gin.Context) string {
	if v, ok := c.Get(ContextKeyRole); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
