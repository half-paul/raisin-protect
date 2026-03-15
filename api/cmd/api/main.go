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

			// === Sprint 5: Policy Management ===

			// Policies (CRUD + status transitions)
			policies := protected.Group("/policies")
			{
				policies.GET("", handlers.ListPolicies)
				policies.POST("", middleware.RequireRoles(models.PolicyCreateRoles...), handlers.CreatePolicy)
				policies.GET("/search", handlers.SearchPolicies)
				policies.GET("/stats", handlers.GetPolicyStats)

				policies.GET("/:id", handlers.GetPolicy)
				policies.PUT("/:id", handlers.UpdatePolicy) // owner check in handler
				policies.POST("/:id/archive", middleware.RequireRoles(models.PolicyArchiveRoles...), handlers.ArchivePolicy)
				policies.POST("/:id/submit-for-review", handlers.SubmitForReview) // owner check in handler
				policies.POST("/:id/publish", middleware.RequireRoles(models.PolicyPublishRoles...), handlers.PublishPolicy)

				// Policy Versions
				policies.GET("/:id/versions", handlers.ListPolicyVersions)
				policies.GET("/:id/versions/compare", handlers.CompareVersions)
				policies.GET("/:id/versions/:version_number", handlers.GetPolicyVersion)
				policies.POST("/:id/versions", handlers.CreatePolicyVersion) // owner check in handler

				// Policy Sign-offs
				policies.GET("/:id/signoffs", handlers.ListPolicySignoffs)
				policies.POST("/:id/signoffs/remind", handlers.RemindSignoffs) // owner check in handler
				policies.POST("/:id/signoffs/:signoff_id/approve", handlers.ApproveSignoff) // signer check in handler
				policies.POST("/:id/signoffs/:signoff_id/reject", handlers.RejectSignoff)   // signer check in handler
				policies.POST("/:id/signoffs/:signoff_id/withdraw", handlers.WithdrawSignoff) // requester check in handler

				// Policy-to-Control Mapping
				policies.GET("/:id/controls", handlers.ListPolicyControls)
				policies.POST("/:id/controls", handlers.LinkPolicyControl) // owner check in handler
				policies.POST("/:id/controls/bulk", handlers.BulkLinkPolicyControls) // owner check in handler
				policies.DELETE("/:id/controls/:control_id", handlers.UnlinkPolicyControl) // owner check in handler
			}

			// Pending sign-offs (cross-policy, per-user)
			protected.GET("/signoffs/pending", handlers.ListPendingSignoffs)

			// Policy Templates
			templates := protected.Group("/policy-templates")
			{
				templates.GET("", handlers.ListPolicyTemplates)
				templates.POST("/:id/clone", middleware.RequireRoles(models.PolicyCreateRoles...), handlers.ClonePolicyTemplate)
			}

			// Policy Gap Detection
			policyGap := protected.Group("/policy-gap")
			{
				policyGap.GET("", middleware.RequireRoles(models.PolicyGapRoles...), handlers.GetPolicyGap)
				policyGap.GET("/by-framework", middleware.RequireRoles(models.PolicyGapRoles...), handlers.GetPolicyGapByFramework)
			}

			// === Sprint 6: Risk Register ===

			// Risks (CRUD + status transitions)
			risks := protected.Group("/risks")
			{
				risks.GET("", handlers.ListRisks)
				risks.POST("", middleware.RequireRoles(models.RiskCreateRoles...), handlers.CreateRisk)
				risks.GET("/heat-map", handlers.GetRiskHeatMap)
				risks.GET("/gaps", middleware.RequireRoles(models.RiskGapRoles...), handlers.GetRiskGaps)
				risks.GET("/search", handlers.SearchRisks)
				risks.GET("/stats", handlers.GetRiskStats)

				risks.GET("/:id", handlers.GetRisk)
				risks.PUT("/:id", handlers.UpdateRisk) // owner check in handler
				risks.POST("/:id/archive", middleware.RequireRoles(models.RiskArchiveRoles...), handlers.ArchiveRisk)
				risks.PUT("/:id/status", handlers.ChangeRiskStatus) // owner + role check in handler
				risks.POST("/:id/recalculate", middleware.RequireRoles(models.RiskRecalcRoles...), handlers.RecalculateRiskScores)

				// Risk Assessments
				risks.GET("/:id/assessments", handlers.ListRiskAssessments)
				risks.POST("/:id/assessments", handlers.CreateRiskAssessment) // owner + role check in handler

				// Risk Treatments
				risks.GET("/:id/treatments", handlers.ListRiskTreatments)
				risks.POST("/:id/treatments", handlers.CreateRiskTreatment) // owner + role check in handler
				risks.PUT("/:id/treatments/:treatment_id", handlers.UpdateRiskTreatment) // owner check in handler
				risks.POST("/:id/treatments/:treatment_id/complete", handlers.CompleteTreatment) // owner check in handler

				// Risk-to-Control Linkage
				risks.GET("/:id/controls", handlers.ListRiskControls)
				risks.POST("/:id/controls", handlers.LinkRiskControl) // owner + role check in handler
				risks.PUT("/:id/controls/:control_id", handlers.UpdateRiskControl) // owner + role check in handler
				risks.DELETE("/:id/controls/:control_id", handlers.UnlinkRiskControl) // owner + role check in handler
			}

			// === Sprint 7: Audit Hub ===

			// Audit Request Templates (PBC list, global)
			auditTemplates := protected.Group("/audit-request-templates")
			{
				auditTemplates.GET("", middleware.RequireRoles(models.AuditRequestCreateRoles...), handlers.ListAuditRequestTemplates)
			}

			// Audits (CRUD + status + auditor management)
			audits := protected.Group("/audits")
			{
				audits.GET("", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.ListAudits)
				audits.POST("", middleware.RequireRoles(models.AuditCreateRoles...), handlers.CreateAudit)
				audits.GET("/dashboard", middleware.RequireRoles(models.AuditDashboardRoles...), handlers.GetAuditDashboard)

				audits.GET("/:id", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.GetAudit)
				audits.PUT("/:id", middleware.RequireRoles(models.AuditCreateRoles...), handlers.UpdateAudit)
				audits.PUT("/:id/status", middleware.RequireRoles(models.AuditCreateRoles...), handlers.ChangeAuditStatus)
				audits.POST("/:id/auditors", middleware.RequireRoles(models.AuditCreateRoles...), handlers.AddAuditAuditor)
				audits.DELETE("/:id/auditors/:user_id", middleware.RequireRoles(models.AuditCreateRoles...), handlers.RemoveAuditAuditor)

				// Per-audit stats and readiness
				audits.GET("/:id/stats", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.GetAuditStats)
				audits.GET("/:id/readiness", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.GetAuditReadiness)

				// Audit Requests (evidence request/response workflow)
				audits.GET("/:id/requests", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.ListAuditRequests)
				audits.GET("/:id/requests/:rid", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.GetAuditRequest)
				audits.POST("/:id/requests", middleware.RequireRoles(models.AuditRequestCreateRoles...), handlers.CreateAuditRequest)
				audits.PUT("/:id/requests/:rid", middleware.RequireRoles(models.AuditRequestCreateRoles...), handlers.UpdateAuditRequest)
				audits.PUT("/:id/requests/:rid/assign", middleware.RequireRoles(models.AuditRequestAssignRoles...), handlers.AssignAuditRequest)
				audits.PUT("/:id/requests/:rid/submit", middleware.RequireRoles(models.AuditEvidenceSubmitRoles...), handlers.SubmitAuditRequest)
				audits.PUT("/:id/requests/:rid/review", middleware.RequireRoles(models.AuditEvidenceReviewRoles...), handlers.ReviewAuditRequest)
				audits.PUT("/:id/requests/:rid/close", middleware.RequireRoles(models.AuditRequestCreateRoles...), handlers.CloseAuditRequest)
				audits.POST("/:id/requests/bulk", middleware.RequireRoles(models.AuditRequestCreateRoles...), handlers.BulkCreateAuditRequests)
				audits.POST("/:id/requests/from-template", middleware.RequireRoles(models.AuditRequestCreateRoles...), handlers.CreateFromTemplate)

				// Evidence submission for requests
				audits.GET("/:id/requests/:rid/evidence", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.ListRequestEvidence)
				audits.POST("/:id/requests/:rid/evidence", middleware.RequireRoles(models.AuditEvidenceSubmitRoles...), handlers.SubmitRequestEvidence)
				audits.PUT("/:id/requests/:rid/evidence/:lid/review", middleware.RequireRoles(models.AuditEvidenceReviewRoles...), handlers.ReviewRequestEvidence)
				audits.DELETE("/:id/requests/:rid/evidence/:lid", handlers.RemoveRequestEvidence) // auth check in handler

				// Audit Findings
				audits.GET("/:id/findings", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.ListAuditFindings)
				audits.GET("/:id/findings/:fid", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.GetAuditFinding)
				audits.POST("/:id/findings", middleware.RequireRoles(models.AuditFindingCreateRoles...), handlers.CreateAuditFinding)
				audits.PUT("/:id/findings/:fid", middleware.RequireRoles(models.AuditFindingCreateRoles...), handlers.UpdateAuditFinding)
				audits.PUT("/:id/findings/:fid/status", handlers.ChangeFindingStatus) // role check in handler per transition
				audits.PUT("/:id/findings/:fid/management-response", middleware.RequireRoles(models.AuditManagementResponseRoles...), handlers.SubmitManagementResponse)

				// Audit Comments
				audits.GET("/:id/comments", middleware.RequireRoles(models.AuditHubViewRoles...), handlers.ListAuditComments)
				audits.POST("/:id/comments", middleware.RequireRoles(models.AuditCommentCreateRoles...), handlers.CreateAuditComment)
				audits.PUT("/:id/comments/:cid", handlers.UpdateAuditComment) // author check in handler
				audits.DELETE("/:id/comments/:cid", handlers.DeleteAuditComment) // author + admin check in handler
			}

			// === Sprint 8: User Access Reviews ===
			ar := protected.Group("/access-reviews")
			{
				// Identity Providers
				idp := ar.Group("/identity-providers")
				{
					idp.GET("", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.ListIdentityProviders)
					idp.POST("", middleware.RequireRoles(models.IdPManageRoles...), handlers.CreateIdentityProvider)
					idp.GET("/:id", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetIdentityProvider)
					idp.PUT("/:id", middleware.RequireRoles(models.IdPManageRoles...), handlers.UpdateIdentityProvider)
					idp.DELETE("/:id", middleware.RequireRoles(models.IdPManageRoles...), handlers.DeleteIdentityProvider)
					idp.POST("/:id/sync", middleware.RequireRoles(models.IdPManageRoles...), handlers.SyncIdentityProvider)
					idp.GET("/:id/sync-stats", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetIdentityProviderSyncStats)
				}

				// Access Resources
				res := ar.Group("/resources")
				{
					res.GET("", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.ListAccessResources)
					res.POST("", middleware.RequireRoles(models.ResourceManageRoles...), handlers.CreateAccessResource)
					res.GET("/stats", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetAccessResourceStats)
					res.GET("/:id", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetAccessResource)
					res.PUT("/:id", middleware.RequireRoles(models.ResourceManageRoles...), handlers.UpdateAccessResource)
					res.DELETE("/:id", middleware.RequireRoles(models.ResourceManageRoles...), handlers.DeleteAccessResource)
					res.GET("/:id/users", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.ListResourceUsers)
				}

				// Access Entries
				entries := ar.Group("/entries")
				{
					entries.GET("", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.ListAccessEntries)
					entries.GET("/anomalies", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetAccessEntryAnomalies)
					entries.POST("/detect-anomalies", middleware.RequireRoles(models.ResourceManageRoles...), handlers.DetectAnomalies)
					entries.GET("/:id", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetAccessEntry)
				}

				// Campaigns
				camp := ar.Group("/campaigns")
				{
					camp.GET("", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.ListCampaigns)
					camp.POST("", middleware.RequireRoles(models.CampaignManageRoles...), handlers.CreateCampaign)
					camp.GET("/:id", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetCampaign)
					camp.PUT("/:id", middleware.RequireRoles(models.CampaignManageRoles...), handlers.UpdateCampaign)
					camp.POST("/:id/launch", middleware.RequireRoles(models.CampaignManageRoles...), handlers.LaunchCampaign)
					camp.POST("/:id/complete", middleware.RequireRoles(models.AccessReviewAdminRoles...), handlers.CompleteCampaign)
					camp.POST("/:id/cancel", middleware.RequireRoles(models.AccessReviewAdminRoles...), handlers.CancelCampaign)
					camp.GET("/:id/stats", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetCampaignStats)
					camp.GET("/:id/certification-report", middleware.RequireRoles("compliance_manager", "ciso", "auditor"), handlers.GetCertificationReport)

					// Reviews within campaigns
					camp.GET("/:id/reviews", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.ListCampaignReviews)
					camp.GET("/:id/reviews/:rid", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetReviewDetail)
					camp.POST("/:id/reviews/:rid/decide", middleware.RequireRoles(models.AccessReviewReviewerRoles...), handlers.DecideReviewNested)
					camp.POST("/:id/reviews/bulk-decide", middleware.RequireRoles(models.AccessReviewReviewerRoles...), handlers.BulkDecideReviews)
					camp.POST("/:id/reviews/:rid/delegate", middleware.RequireRoles(models.AccessReviewReviewerRoles...), handlers.DelegateReview)
					camp.POST("/:id/reviews/:rid/escalate", middleware.RequireRoles(models.AccessReviewAdminRoles...), handlers.EscalateReview)
					camp.POST("/:id/reviews/:rid/revocation", middleware.RequireRoles("it_admin", "ciso"), handlers.MarkRevocation)
				}

				// Individual Reviews (legacy path)
				rev := ar.Group("/reviews")
				{
					rev.PUT("/:id", middleware.RequireRoles(models.AccessReviewReviewerRoles...), handlers.DecideReview)
				}

				// Dashboard
				ar.GET("/dashboard", middleware.RequireRoles(models.AccessReviewViewRoles...), handlers.GetAccessReviewDashboard)

				// Personal Queue
				ar.GET("/my-reviews", handlers.ListMyReviews)
			}

			// === Sprint 9: Integration Engine ===

			// Integration catalog (system-level, read-only)
			integrations := protected.Group("/integrations")
			integrations.Use(middleware.RequireRoles(models.IntegrationViewRoles...))
			{
				integrations.GET("", handlers.ListIntegrations)
				integrations.GET("/dashboard", handlers.IntegrationDashboard)
				integrations.GET("/dashboard/sync-activity", handlers.IntegrationSyncActivity)
				integrations.GET("/:id", handlers.GetIntegration)
			}

			// Integration connections (per-org)
			connections := protected.Group("/integration-connections")
			{
				connections.GET("", middleware.RequireRoles(models.IntegrationViewRoles...), handlers.ListConnections)
				connections.GET("/:id", middleware.RequireRoles(models.IntegrationViewRoles...), handlers.GetConnection)
				connections.POST("", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.CreateConnection)
				connections.PUT("/:id", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.UpdateConnection)
				connections.DELETE("/:id", middleware.RequireRoles(models.IntegrationDeleteRoles...), handlers.DeleteConnection)
				connections.POST("/:id/test", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.TestConnection)
				connections.POST("/:id/enable", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.EnableConnection)
				connections.POST("/:id/disable", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.DisableConnection)
				connections.POST("/:id/sync", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.TriggerSync)
				connections.GET("/:id/runs", middleware.RequireRoles(models.IntegrationViewRoles...), handlers.ListRuns)
				connections.GET("/:id/runs/:rid", middleware.RequireRoles(models.IntegrationViewRoles...), handlers.GetRun)
				connections.POST("/:id/runs/:rid/cancel", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.CancelRun)
				connections.GET("/:id/runs/:rid/logs", middleware.RequireRoles(models.IntegrationViewRoles...), handlers.GetRunLogs)
				connections.GET("/:id/health", middleware.RequireRoles(models.IntegrationViewRoles...), handlers.GetConnectionHealth)
				connections.POST("/:id/health-check", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.TriggerHealthCheck)
				connections.GET("/:id/webhooks", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.ListWebhooks)
				connections.POST("/:id/webhooks", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.CreateWebhook)
				connections.DELETE("/:id/webhooks/:wid", middleware.RequireRoles(models.IntegrationManageRoles...), handlers.DeleteWebhook)
				connections.POST("/:id/webhooks/:wid/rotate-secret", middleware.RequireRoles(models.IntegrationDeleteRoles...), handlers.RotateWebhookSecret)
				connections.GET("/:id/preview", middleware.RequireRoles(models.IntegrationViewRoles...), handlers.IntegrationPreview)
			}
		}

		// CDE Scoping Module (PCI DSS Req 1, 11.4) — Sprint 11
		cde := protected.Group("/cde")
		{
			// Assets
			cdeAssets := cde.Group("/assets")
			{
				cdeAssets.GET("", middleware.RequireRoles(models.AdminRoles...), handlers.ListCDEAssets)
				cdeAssets.POST("", middleware.RequireRoles(models.AdminRoles...), handlers.CreateCDEAsset)
				cdeAssets.GET("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.GetCDEAsset)
				cdeAssets.PUT("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.UpdateCDEAsset)
				cdeAssets.DELETE("/:id", middleware.RequireAdmin(), handlers.DeleteCDEAsset)
			}
			// Network Segments
			cdeSegs := cde.Group("/segments")
			{
				cdeSegs.GET("", middleware.RequireRoles(models.AdminRoles...), handlers.ListCDESegments)
				cdeSegs.POST("", middleware.RequireRoles(models.AdminRoles...), handlers.CreateCDESegment)
				cdeSegs.GET("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.GetCDESegment)
				cdeSegs.PUT("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.UpdateCDESegment)
				cdeSegs.DELETE("/:id", middleware.RequireAdmin(), handlers.DeleteCDESegment)
			}
			// Data Flows
			cdeFlows := cde.Group("/data-flows")
			{
				cdeFlows.GET("", middleware.RequireRoles(models.AdminRoles...), handlers.ListCDEDataFlows)
				cdeFlows.POST("", middleware.RequireRoles(models.AdminRoles...), handlers.CreateCDEDataFlow)
				cdeFlows.GET("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.GetCDEDataFlow)
				cdeFlows.PUT("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.UpdateCDEDataFlow)
				cdeFlows.DELETE("/:id", middleware.RequireAdmin(), handlers.DeleteCDEDataFlow)
			}
			// Segmentation Tests
			cdeSegTests := cde.Group("/segmentation-tests")
			{
				cdeSegTests.GET("", middleware.RequireRoles(models.AdminRoles...), handlers.ListCDESegmentationTests)
				cdeSegTests.POST("", middleware.RequireRoles(models.AdminRoles...), handlers.CreateCDESegmentationTest)
				cdeSegTests.GET("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.GetCDESegmentationTest)
				cdeSegTests.PUT("/:id", middleware.RequireRoles(models.AdminRoles...), handlers.UpdateCDESegmentationTest)
				cdeSegTests.DELETE("/:id", middleware.RequireAdmin(), handlers.DeleteCDESegmentationTest)
			}
			// Scope Summary
			cde.GET("/scope-summary", middleware.RequireRoles(models.AdminRoles...), handlers.GetCDEScopeSummary)
		}

		// Service Provider / Vendor Management (PCI DSS Req 12.8, 12.9) — Sprint 12
		sp := protected.Group("/service-providers")
		{
			sp.GET("", middleware.RequireRoles(models.SPViewRoles...), handlers.ListServiceProviders)
			sp.POST("", middleware.RequireRoles(models.SPManageRoles...), handlers.CreateServiceProvider)
			sp.GET("/compliance-summary", middleware.RequireRoles(models.SPViewRoles...), handlers.GetSPComplianceSummary)
			sp.GET("/:id", middleware.RequireRoles(models.SPViewRoles...), handlers.GetServiceProvider)
			sp.PUT("/:id", middleware.RequireRoles(models.SPManageRoles...), handlers.UpdateServiceProvider)
			sp.DELETE("/:id", middleware.RequireRoles(models.SPManageRoles...), handlers.DeleteServiceProvider)

			// Compliance Documents
			spDocs := sp.Group("/:id/documents")
			{
				spDocs.GET("", middleware.RequireRoles(models.SPViewRoles...), handlers.ListSPComplianceDocs)
				spDocs.POST("", middleware.RequireRoles(models.SPManageRoles...), handlers.CreateSPComplianceDoc)
				spDocs.GET("/:docId", middleware.RequireRoles(models.SPViewRoles...), handlers.GetSPComplianceDoc)
				spDocs.DELETE("/:docId", middleware.RequireRoles(models.SPManageRoles...), handlers.DeleteSPComplianceDoc)
			}

			// Responsibility Matrix
			spResp := sp.Group("/:id/responsibilities")
			{
				spResp.GET("", middleware.RequireRoles(models.SPViewRoles...), handlers.ListSPResponsibilities)
				spResp.PUT("/:reqCode", middleware.RequireRoles(models.SPManageRoles...), handlers.UpsertSPResponsibility)
				spResp.DELETE("/:reqCode", middleware.RequireRoles(models.SPManageRoles...), handlers.DeleteSPResponsibility)
			}
		}

		// ASV Scan Management (PCI DSS Req 11.3.2) — Sprint 12
		asv := protected.Group("/asv-scans")
		{
			asv.GET("", middleware.RequireRoles(models.ASVViewRoles...), handlers.ListASVScans)
			asv.POST("", middleware.RequireRoles(models.ASVManageRoles...), handlers.CreateASVScan)
			asv.GET("/quarterly-status", middleware.RequireRoles(models.ASVViewRoles...), handlers.GetASVQuarterlyStatus)
			asv.POST("/import", middleware.RequireRoles(models.ASVManageRoles...), handlers.ImportASVScan)
			asv.GET("/:id", middleware.RequireRoles(models.ASVViewRoles...), handlers.GetASVScan)
			asv.PUT("/:id", middleware.RequireRoles(models.ASVManageRoles...), handlers.UpdateASVScan)
			asv.DELETE("/:id", middleware.RequireRoles(models.ASVManageRoles...), handlers.DeleteASVScan)
		}

		// Public webhook receiver (no JWT auth, HMAC verification)
		v1.POST("/webhooks/receive/:id", handlers.ReceiveWebhook)
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
