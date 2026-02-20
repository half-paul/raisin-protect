package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Simple in-memory rate limiter (Redis-backed rate limiting can be added later).
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old entries
	timestamps := rl.requests[key]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

var (
	publicLimiter = newRateLimiter(10, time.Minute)  // 10/min per IP
	authLimiter   = newRateLimiter(100, time.Minute)  // 100/min per user
)

// RateLimitPublic applies rate limiting based on IP for public endpoints.
func RateLimitPublic() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !publicLimiter.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "Too many requests"},
			})
			return
		}
		c.Next()
	}
}

// RateLimitAuth applies rate limiting based on user ID for authenticated endpoints.
func RateLimitAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		key := userID
		if key == "" {
			key = c.ClientIP()
		}
		if !authLimiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "Too many requests"},
			})
			return
		}
		c.Next()
	}
}
