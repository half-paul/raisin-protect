// Package main is the entry point for the Raisin Protect API server.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/half-paul/raisin-protect/api/internal/auth"
	"github.com/half-paul/raisin-protect/api/internal/config"
	"github.com/half-paul/raisin-protect/api/internal/db"
	"github.com/half-paul/raisin-protect/api/internal/handlers"
	"github.com/half-paul/raisin-protect/api/internal/middleware"
	"github.com/half-paul/raisin-protect/api/internal/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load()

	// Logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("RP_LOG_LEVEL") == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// JWT
	jwtManager := auth.NewJWTManager(auth.JWTConfig{
		Secret:        cfg.JWTSecret,
		AccessExpiry:  cfg.JWTAccessExpiry,
		RefreshExpiry: cfg.JWTRefreshExpiry,
		Issuer:        cfg.JWTIssuer,
	})
	middleware.SetJWTManager(jwtManager)
	handlers.SetJWTManager(jwtManager)
	handlers.SetBcryptCost(cfg.BcryptCost)
	log.Info().
		Dur("access_expiry", cfg.JWTAccessExpiry).
		Dur("refresh_expiry", cfg.JWTRefreshExpiry).
		Msg("JWT authentication configured")

	// Database
	database, err := db.Connect(db.Config{
		URL:          cfg.DatabaseURL,
		MaxOpenConns: 25,
		MaxIdleConns: 5,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to database — running in limited mode")
	} else {
		defer database.Close()
		handlers.SetDB(database)
		middleware.SetAuditDB(database.DB)
		log.Info().Msg("Database connected successfully")
	}

	// Redis
	redisClient, err := db.ConnectRedis(cfg.RedisURL)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Redis — rate limiting may be degraded")
	} else {
		defer redisClient.Close()
		handlers.SetRedis(redisClient)
		log.Info().Msg("Redis connected successfully")
	}

	// Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(cfg.CORSOrigins))

	// Health endpoints (public, no auth)
	router.GET("/health", handlers.HealthCheck)
	router.GET("/ready", handlers.ReadyCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Public auth routes with rate limiting
		authRoutes := v1.Group("/auth")
		authRoutes.Use(middleware.RateLimitPublic())
		{
			authRoutes.POST("/register", handlers.Register)
			authRoutes.POST("/login", handlers.Login)
			authRoutes.POST("/refresh", handlers.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired())
		protected.Use(middleware.RateLimitAuth())
		{
			// Auth (any authenticated user)
			protected.POST("/auth/logout", handlers.Logout)
			protected.POST("/auth/change-password", handlers.ChangePassword)

			// Organizations (all roles can read; admin roles can update)
			protected.GET("/organizations/current", handlers.GetCurrentOrganization)
			protected.PUT("/organizations/current", middleware.RequireAdmin(), handlers.UpdateCurrentOrganization)

			// Users (all roles can list/get)
			protected.GET("/users", handlers.ListUsers)
			protected.GET("/users/:id", handlers.GetUser)

			// Users (restricted roles can create/update)
			protected.POST("/users", middleware.RequireRoles(models.UserCreateRoles...), handlers.CreateUser)
			protected.PUT("/users/:id", handlers.UpdateUser) // self-edit + admin handled inside handler

			// User lifecycle (admin roles only)
			protected.POST("/users/:id/deactivate", middleware.RequireAdmin(), handlers.DeactivateUser)
			protected.POST("/users/:id/reactivate", middleware.RequireAdmin(), handlers.ReactivateUser)
			protected.PUT("/users/:id/role", middleware.RequireAdmin(), handlers.ChangeUserRole)

			// Audit log (admin + auditor)
			protected.GET("/audit-log", middleware.RequireRoles(models.AuditViewRoles...), handlers.ListAuditLogs)
		}
	}

	// HTTP server
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info().Str("addr", addr).Str("env", cfg.Environment).Msg("Starting Raisin Protect API server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}
	log.Info().Msg("Server stopped")
}
