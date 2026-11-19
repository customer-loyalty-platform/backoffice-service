package http

import (
	"time"

	"backoffice-service/config"
	"backoffice-service/internal/middleware"
	adminUserHandler "backoffice-service/internal/modules/admin_user/handler"
	adminUserRepo "backoffice-service/internal/modules/admin_user/repository"
	adminUserService "backoffice-service/internal/modules/admin_user/service"
	authHandler "backoffice-service/internal/modules/auth/handler"
	authRepo "backoffice-service/internal/modules/auth/repository"
	authService "backoffice-service/internal/modules/auth/service"
	campaignHandler "backoffice-service/internal/modules/campaign/handler"
	campaignRepo "backoffice-service/internal/modules/campaign/repository"
	campaignService "backoffice-service/internal/modules/campaign/service"
	earningRuleHandler "backoffice-service/internal/modules/earning_rule/handler"
	earningRuleRepo "backoffice-service/internal/modules/earning_rule/repository"
	earningRuleService "backoffice-service/internal/modules/earning_rule/service"
	makerCheckerHandler "backoffice-service/internal/modules/maker_checker/handler"
	makerCheckerRepo "backoffice-service/internal/modules/maker_checker/repository"
	makerCheckerService "backoffice-service/internal/modules/maker_checker/service"
	memberHandler "backoffice-service/internal/modules/member/handler"
	memberRepo "backoffice-service/internal/modules/member/repository"
	memberService "backoffice-service/internal/modules/member/service"
	monitoringHandler "backoffice-service/internal/modules/monitoring/handler"
	monitoringRepo "backoffice-service/internal/modules/monitoring/repository"
	monitoringService "backoffice-service/internal/modules/monitoring/service"
	settlementHandler "backoffice-service/internal/modules/settlement/handler"
	settlementRepo "backoffice-service/internal/modules/settlement/repository"
	settlementService "backoffice-service/internal/modules/settlement/service"
	tierHandler "backoffice-service/internal/modules/tier/handler"
	tierRepo "backoffice-service/internal/modules/tier/repository"
	tierService "backoffice-service/internal/modules/tier/service"
	voucherHandler "backoffice-service/internal/modules/voucher_catalog/handler"
	voucherRepo "backoffice-service/internal/modules/voucher_catalog/repository"
	voucherService "backoffice-service/internal/modules/voucher_catalog/service"
	"backoffice-service/internal/pkg/database"
	pkgJWT "backoffice-service/internal/pkg/jwt"
	"backoffice-service/internal/pkg/validator"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(cfg *config.Config, db *database.PostgresDB, log *zap.Logger) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global Middlewares
	r.Use(middleware.Recovery(log))

	// CORS Configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	r.Use(cors.New(corsConfig))

	// JWT Manager
	jwtManager := pkgJWT.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	// JSON Schema Validator
	schemaValidator, err := validator.NewJSONSchemaValidator()
	if err != nil {
		log.Warn("Failed to compile earning rule JSON schema meta-validator", zap.Error(err))
	}

	// Module Repositories
	authR := authRepo.NewAuthRepository(db.Master)
	memberR := memberRepo.NewMemberRepository(db.Master, db.Replica)
	tierR := tierRepo.NewTierRepository(db.Master)
	earningRuleR := earningRuleRepo.NewEarningRuleRepository(db.Master)
	voucherR := voucherRepo.NewVoucherRepository(db.Master)
	campaignR := campaignRepo.NewCampaignRepository(db.Master)
	settlementR := settlementRepo.NewSettlementRepository(db.Master)
	makerCheckerR := makerCheckerRepo.NewMakerCheckerRepository(db.Master)
	monitoringR := monitoringRepo.NewMonitoringRepository(db.Replica)
	adminUserR := adminUserRepo.NewAdminUserRepository(db.Master)

	// Module Services
	authS := authService.NewAuthService(authR, jwtManager, cfg.JWT.AccessExpiry)
	memberS := memberService.NewMemberService(memberR)
	tierS := tierService.NewTierService(tierR)
	earningRuleS := earningRuleService.NewEarningRuleService(earningRuleR, schemaValidator)
	voucherS := voucherService.NewVoucherService(voucherR)
	campaignS := campaignService.NewCampaignService(campaignR)
	settlementS := settlementService.NewSettlementService(settlementR)
	makerCheckerS := makerCheckerService.NewMakerCheckerService(makerCheckerR)
	monitoringS := monitoringService.NewMonitoringService(monitoringR)
	adminUserS := adminUserService.NewAdminUserService(adminUserR)

	// Module Handlers
	authH := authHandler.NewAuthHandler(authS)
	memberH := memberHandler.NewMemberHandler(memberS)
	tierH := tierHandler.NewTierHandler(tierS)
	earningRuleH := earningRuleHandler.NewEarningRuleHandler(earningRuleS)
	voucherH := voucherHandler.NewVoucherHandler(voucherS)
	campaignH := campaignHandler.NewCampaignHandler(campaignS)
	settlementH := settlementHandler.NewSettlementHandler(settlementS)
	makerCheckerH := makerCheckerHandler.NewMakerCheckerHandler(makerCheckerS)
	monitoringH := monitoringHandler.NewMonitoringHandler(monitoringS)
	adminUserH := adminUserHandler.NewAdminUserHandler(adminUserS)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		Success(c, 200, gin.H{"status": "healthy", "service": "backoffice-service"})
	})

	// API Route Group: /api/v1/admin
	api := r.Group("/api/v1/admin")
	{
		// Public Auth Routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", authH.Login)
			authGroup.POST("/refresh", authH.Refresh)
		}

		// Protected Routes
		protected := api.Group("")
		protected.Use(middleware.Auth(jwtManager))
		protected.Use(middleware.AuditLogger(db.Master, log))
		{
			// Auth
			protected.POST("/auth/logout", authH.Logout)
			protected.GET("/auth/me", authH.Me)

			// Members
			protected.GET("/members", memberH.GetMembers)
			protected.GET("/members/:id", memberH.GetMemberByID)
			protected.PATCH("/members/:id/status", memberH.UpdateStatus)
			protected.POST("/members/:id/adjustments", memberH.SubmitAdjustment)
			protected.POST("/members/:id/points/adjust", memberH.SubmitAdjustment)
			protected.GET("/members/:id/transactions", memberH.GetMemberTransactions)

			// Tiers
			protected.GET("/tiers", tierH.GetTiers)
			protected.POST("/tiers", tierH.CreateTier)
			protected.PUT("/tiers/:id", tierH.UpdateTier)
			protected.DELETE("/tiers/:id", tierH.DeleteTier)

			// Earning Rules
			protected.GET("/rules", earningRuleH.GetRules)
			protected.POST("/rules", earningRuleH.CreateRule)
			protected.PUT("/rules/:id", earningRuleH.UpdateRule)
			protected.PATCH("/rules/:id/status", earningRuleH.UpdateRuleStatus)
			protected.DELETE("/rules/:id", earningRuleH.DeleteRule)

			// Voucher Batches
			protected.GET("/voucher-batches", voucherH.GetBatches)
			protected.POST("/voucher-batches", voucherH.CreateBatch)
			protected.POST("/voucher-batches/:id/upload-codes", voucherH.UploadCodes)
			protected.GET("/voucher-batches/:id/inventory", voucherH.GetBatchInventory)
			protected.POST("/voucher-batches/:id/revoke", voucherH.RevokeVouchers)

			// Campaigns
			protected.GET("/campaigns", campaignH.GetCampaigns)
			protected.POST("/campaigns", campaignH.CreateCampaign)
			protected.PUT("/campaigns/:id", campaignH.UpdateCampaign)
			protected.PATCH("/campaigns/:id/status", campaignH.UpdateStatus)

			// Partners & Settlement
			protected.GET("/partners", settlementH.GetPartners)
			protected.POST("/partners", settlementH.CreatePartner)
			protected.PUT("/partners/:id", settlementH.UpdatePartner)
			protected.GET("/partners/settlement-report", settlementH.GetSettlementReport)

			// Maker-Checker Approvals
			protected.GET("/pending-actions", makerCheckerH.GetPendingActions)
			protected.POST("/pending-actions/:id/approve", makerCheckerH.Approve)
			protected.POST("/pending-actions/:id/reject", makerCheckerH.Reject)

			// Dashboard & Analytics
			protected.GET("/dashboard/summary", monitoringH.GetSummary)
			protected.GET("/dashboard/transaction-feed", monitoringH.GetTransactionFeed)
			protected.GET("/dashboard/charts/earning-trend", monitoringH.GetEarningTrend)
			protected.GET("/audit-logs", monitoringH.GetAuditLogs)

			// Admin Users (SUPER_ADMIN)
			protected.GET("/users", adminUserH.GetAdminUsers)
			protected.POST("/users", adminUserH.CreateAdminUser)
			protected.PUT("/users/:id", adminUserH.UpdateAdminUser)
		}
	}

	return r
}

