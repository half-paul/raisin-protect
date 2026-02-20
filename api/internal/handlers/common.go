// Package handlers provides HTTP handlers for the Raisin Protect API.
package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/half-paul/raisin-protect/api/internal/auth"
	"github.com/half-paul/raisin-protect/api/internal/db"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
)

var (
	database   *db.DB
	redisDB    *db.RedisClient
	jwtMgr     *auth.JWTManager
	bcryptCost int = 12
)

// SetDB sets the database connection for handlers.
func SetDB(d *db.DB) {
	database = d
}

// SetRedis sets the Redis client for handlers.
func SetRedis(r *db.RedisClient) {
	redisDB = r
}

// SetJWTManager sets the JWT manager for handlers.
func SetJWTManager(m *auth.JWTManager) {
	jwtMgr = m
}

// SetBcryptCost sets the bcrypt cost for password hashing.
func SetBcryptCost(cost int) {
	bcryptCost = cost
}

// successResponse creates a standard success response.
func successResponse(c *gin.Context, data interface{}) gin.H {
	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	return gin.H{
		"data": data,
		"meta": gin.H{
			"request_id": reqID,
		},
	}
}

// listResponse creates a standard paginated list response.
func listResponse(c *gin.Context, data interface{}, total int, page, perPage int) gin.H {
	reqID, _ := c.Get(middleware.ContextKeyRequestID)
	return gin.H{
		"data": data,
		"meta": gin.H{
			"total":      total,
			"page":       page,
			"per_page":   perPage,
			"request_id": reqID,
		},
	}
}

// errorResponse creates a standard error response.
func errorResponse(code, message string) gin.H {
	return gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}

// errorResponseWithDetails creates an error response with field-level details.
func errorResponseWithDetails(code, message string, details []gin.H) gin.H {
	return gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"details": details,
		},
	}
}
