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
	"github.com/half-paul/raisin-protect/api/internal/services"
	"github.com/half-paul/raisin-protect/api/internal/workers"
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

	// MinIO
	minioSvc, err := services.NewMinIOService(services.MinIOConfig{
		Endpoint:  cfg.MinIOEndpoint,
		AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey,
		Bucket:    cfg.MinIOBucket,
		UseSSL:    cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to MinIO — evidence uploads may be degraded")
	} else {
		if err := minioSvc.EnsureBucket(context.Background()); err != nil {
			log.Warn().Err(err).Msg("Failed to ensure MinIO bucket")
		}
		handlers.SetMinIO(minioSvc)
		log.Info().Str("bucket", cfg.MinIOBucket).Msg("MinIO connected successfully")
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

			// === Sprint 2: Frameworks & Controls ===

			// Framework catalog (system-level, read-only)
			fw := protected.Group("/frameworks")
			{
				fw.GET("", handlers.ListFrameworks)
				fw.GET("/:id", handlers.GetFramework)
				fw.GET("/:id/versions/:vid", handlers.GetFrameworkVersion)
				fw.GET("/:id/versions/:vid/requirements", handlers.ListRequirements)
			}

			// Org frameworks (per-org activation)
			of := protected.Group("/org-frameworks")
			{
				of.GET("", handlers.ListOrgFrameworks)
				of.POST("", middleware.RequireRoles(models.OrgFrameworkRoles...), handlers.ActivateFramework)
				of.PUT("/:id", middleware.RequireRoles(models.OrgFrameworkRoles...), handlers.UpdateOrgFramework)
				of.DELETE("/:id", middleware.RequireRoles(models.OrgFrameworkRoles...), handlers.DeactivateFramework)
				of.GET("/:id/coverage", handlers.GetCoverage)
				of.GET("/:id/scoping", handlers.ListScoping)
				of.PUT("/:id/requirements/:rid/scope", middleware.RequireRoles(models.OrgFrameworkRoles...), handlers.SetScope)
				of.DELETE("/:id/requirements/:rid/scope", middleware.RequireRoles(models.OrgFrameworkRoles...), handlers.ResetScope)
			}

			// Controls (per-org library)
			ctrl := protected.Group("/controls")
			{
				ctrl.GET("", handlers.ListControls)
				ctrl.POST("", middleware.RequireRoles(models.ControlCreateRoles...), handlers.CreateControl)
				ctrl.GET("/stats", handlers.GetControlStats)
				ctrl.POST("/bulk-status", middleware.RequireRoles(models.AdminRoles...), handlers.BulkControlStatus)
				ctrl.GET("/:id", handlers.GetControl)
				ctrl.PUT("/:id", handlers.UpdateControl) // owner check in handler
				ctrl.PUT("/:id/owner", middleware.RequireRoles(models.AdminRoles...), handlers.ChangeControlOwner)
				ctrl.PUT("/:id/status", middleware.RequireRoles(models.ControlStatusRoles...), handlers.ChangeControlStatus)
				ctrl.DELETE("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.DeprecateControl)
				ctrl.GET("/:id/mappings", handlers.ListControlMappings)
				ctrl.POST("/:id/mappings", middleware.RequireRoles(models.ControlMappingRoles...), handlers.CreateControlMappings)
				ctrl.DELETE("/:id/mappings/:mid", middleware.RequireRoles(models.ControlMappingRoles...), handlers.DeleteControlMapping)
			}

			// Mapping matrix
			protected.GET("/mapping-matrix", handlers.GetMappingMatrix)

			// === Sprint 3: Evidence Management ===

			ev := protected.Group("/evidence")
			{
				ev.GET("", handlers.ListEvidence)
				ev.POST("", middleware.RequireRoles(models.EvidenceUploadRoles...), handlers.CreateEvidence)
				ev.GET("/staleness", handlers.GetStalenessAlerts)
				ev.GET("/freshness-summary", handlers.GetFreshnessSummary)
				ev.GET("/search", handlers.SearchEvidence)

				ev.GET("/:id", handlers.GetEvidence)
				ev.PUT("/:id", handlers.UpdateEvidence) // uploader check in handler
				ev.DELETE("/:id", middleware.RequireRoles(models.EvidenceStatusRoles...), handlers.DeleteEvidence)
				ev.PUT("/:id/status", middleware.RequireRoles(models.EvidenceStatusRoles...), handlers.ChangeEvidenceStatus)

				// Upload flow
				ev.POST("/:id/confirm", handlers.ConfirmEvidenceUpload) // uploader check in handler
				ev.POST("/:id/upload", handlers.GetUploadURL)          // uploader check in handler
				ev.GET("/:id/download", handlers.GetDownloadURL)

				// Versioning
				ev.POST("/:id/versions", middleware.RequireRoles(models.EvidenceUploadRoles...), handlers.CreateEvidenceVersion)
				ev.GET("/:id/versions", handlers.ListEvidenceVersions)

				// Links
				ev.GET("/:id/links", handlers.ListEvidenceLinks)
				ev.POST("/:id/links", middleware.RequireRoles(models.EvidenceLinkRoles...), handlers.CreateEvidenceLinks)
				ev.DELETE("/:id/links/:lid", middleware.RequireRoles(models.EvidenceLinkRoles...), handlers.DeleteEvidenceLink)

				// Evaluations
				ev.GET("/:id/evaluations", handlers.ListEvidenceEvaluations)
				ev.POST("/:id/evaluations", middleware.RequireRoles(models.EvidenceEvalRoles...), handlers.CreateEvidenceEvaluation)
			}

			// Evidence on existing resources
			ctrl.GET("/:id/evidence", handlers.ListControlEvidence)

			// Requirements evidence
			req := protected.Group("/requirements")
			{
				req.GET("/:id/evidence", handlers.ListRequirementEvidence)
			}

			// === Sprint 4: Continuous Monitoring Engine ===

			// Tests (test definitions)
			tests := protected.Group("/tests")
			{
				tests.GET("", handlers.ListTests)
				tests.POST("", middleware.RequireRoles(models.TestCreateRoles...), handlers.CreateTest)
				tests.GET("/:id", handlers.GetTest)
				tests.PUT("/:id", middleware.RequireRoles(models.TestCreateRoles...), handlers.UpdateTest)
				tests.PUT("/:id/status", middleware.RequireRoles(models.TestStatusRoles...), handlers.ChangeTestStatus)
				tests.DELETE("/:id", middleware.RequireRoles(models.TestDeleteRoles...), handlers.DeleteTest)
				tests.GET("/:id/results", handlers.ListTestResultsByTest)
			}

			// Test Runs (execution sweeps)
			runs := protected.Group("/test-runs")
			{
				runs.POST("", middleware.RequireRoles(models.TestRunCreateRoles...), handlers.CreateTestRun)
				runs.GET("", handlers.ListTestRuns)
				runs.GET("/:id", handlers.GetTestRun)
				runs.POST("/:id/cancel", middleware.RequireRoles(models.TestRunCancelRoles...), handlers.CancelTestRun)
				runs.GET("/:id/results", handlers.ListTestRunResults)
				runs.GET("/:id/results/:rid", handlers.GetTestRunResult)
			}

			// Control test results (cross-resource query)
			ctrl.GET("/:id/test-results", handlers.ListControlTestResults)

			// Alerts
			alerts := protected.Group("/alerts")
			{
				alerts.GET("", handlers.ListAlerts)
				alerts.GET("/:id", handlers.GetAlert)
				alerts.PUT("/:id/status", middleware.RequireRoles(models.AlertStatusRoles...), handlers.ChangeAlertStatus)
				alerts.PUT("/:id/assign", middleware.RequireRoles(models.AlertAssignRoles...), handlers.AssignAlert)
				alerts.PUT("/:id/resolve", middleware.RequireRoles(models.AlertResolveRoles...), handlers.ResolveAlert)
				alerts.PUT("/:id/suppress", middleware.RequireRoles(models.AlertSuppressRoles...), handlers.SuppressAlert)
				alerts.PUT("/:id/close", middleware.RequireRoles(models.AlertSuppressRoles...), handlers.CloseAlert)
				alerts.POST("/:id/deliver", middleware.RequireRoles(models.AlertDeliveryRoles...), handlers.RedeliverAlert)
				alerts.POST("/test-delivery", middleware.RequireRoles(models.AlertSuppressRoles...), handlers.TestAlertDelivery)
			}

			// Alert Rules
			rules := protected.Group("/alert-rules")
			{
				rules.GET("", middleware.RequireRoles(models.AlertRuleViewRoles...), handlers.ListAlertRules)
				rules.POST("", middleware.RequireRoles(models.AlertRuleCreateRoles...), handlers.CreateAlertRule)
				rules.GET("/:id", middleware.RequireRoles(models.AlertRuleViewRoles...), handlers.GetAlertRule)
				rules.PUT("/:id", middleware.RequireRoles(models.AlertRuleCreateRoles...), handlers.UpdateAlertRule)
				rules.DELETE("/:id", middleware.RequireRoles(models.AlertRuleCreateRoles...), handlers.DeleteAlertRule)
			}

			// Monitoring Dashboard
			monitoring := protected.Group("/monitoring")
			{
				monitoring.GET("/heatmap", handlers.GetControlHealthHeatmap)
				monitoring.GET("/posture", handlers.GetCompliancePosture)
				monitoring.GET("/summary", handlers.GetMonitoringSummary)
				monitoring.GET("/alert-queue", handlers.GetAlertQueue)
			}
		}
	}

	// Start monitoring worker (background)
	if database != nil {
		workerCtx, workerCancel := context.WithCancel(context.Background())
		defer workerCancel()
		monitoringWorker := workers.NewMonitoringWorker(database.DB, 30*time.Second)
		go monitoringWorker.Run(workerCtx)
		log.Info().Msg("Monitoring worker started in background")
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
