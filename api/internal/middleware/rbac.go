package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/models"
)

// RequireRoles returns middleware that restricts access to specific GRC roles.
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"},
			})
			return
		}

		if !models.HasRole(userRole, roles) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions",
				},
			})
			return
		}

		c.Next()
	}
}

// RequireAdmin restricts to ciso and compliance_manager.
func RequireAdmin() gin.HandlerFunc {
	return RequireRoles(models.AdminRoles...)
}
