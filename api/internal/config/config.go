// Package config loads application configuration from environment variables.
// All variables use the RP_ prefix for namespacing (Raisin Protect).
package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the API server.
type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string

	// Security
	BcryptCost     int
	CORSOrigins    string
	TrustedProxies []string
}

// Load reads configuration from RP_* environment variables.
func Load() (*Config, error) {
	accessExpiry, err := time.ParseDuration(getEnv("RP_JWT_ACCESS_TTL", "15m"))
	if err != nil {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry, err := time.ParseDuration(getEnv("RP_JWT_REFRESH_TTL", "168h"))
	if err != nil {
		refreshExpiry = 7 * 24 * time.Hour
	}

	bcryptCost, err := strconv.Atoi(getEnv("RP_BCRYPT_COST", "12"))
	if err != nil || bcryptCost < 10 || bcryptCost > 15 {
		bcryptCost = 12
	}

	cfg := &Config{
		Port:             getEnv("RP_PORT", "8090"),
		Environment:      getEnv("RP_ENV", "development"),
		DatabaseURL:      getEnv("RP_DB_URL", "postgres://rp:rp_dev_password@localhost:5433/raisin_protect?sslmode=disable"),
		RedisURL:         getEnv("RP_REDIS_URL", "redis://localhost:6380"),
		JWTSecret:        getEnv("RP_JWT_SECRET", ""),
		JWTAccessExpiry:  accessExpiry,
		JWTRefreshExpiry: refreshExpiry,
		JWTIssuer:        getEnv("RP_JWT_ISSUER", "raisin-protect"),
		BcryptCost:       bcryptCost,
		CORSOrigins:      getEnv("RP_CORS_ORIGINS", "http://localhost:3010"),
	}

	// Validate JWT secret
	if cfg.JWTSecret == "" {
		if cfg.Environment == "production" || cfg.Environment == "staging" {
			return nil, fmt.Errorf("RP_JWT_SECRET is required in production/staging")
		}
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, fmt.Errorf("failed to generate random JWT secret: %w", err)
		}
		cfg.JWTSecret = base64.StdEncoding.EncodeToString(randomBytes)
		log.Warn().Msg("Using randomly generated JWT secret — tokens won't persist across restarts")
	} else if len(cfg.JWTSecret) < 32 && (cfg.Environment == "production" || cfg.Environment == "staging") {
		return nil, fmt.Errorf("RP_JWT_SECRET must be at least 32 characters in production/staging (got %d)", len(cfg.JWTSecret))
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
