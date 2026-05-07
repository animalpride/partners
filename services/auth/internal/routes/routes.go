package routes

import (
	"github.com/animalpride/partners/services/auth/internal/config"
	"github.com/animalpride/partners/services/auth/internal/handlers"
	"github.com/animalpride/partners/services/auth/internal/middleware"
	"github.com/animalpride/partners/services/auth/internal/repository"
	"github.com/animalpride/partners/services/auth/internal/services"
	sharedmw "github.com/animalpride/partners/services/shared/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter initializes the Gin router and sets up the routes
func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	// Initialize services
	jwtService := services.NewJWTService(cfg)
	emailService := services.NewEmailService(cfg)

	// Initialize repositories
	rbacRepo := repository.NewRBACRepository(db)
	userService := repository.NewUserRepository(db, cfg, emailService, rbacRepo)
	refreshRepo := repository.NewRefreshSessionRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	invitationRepo := repository.NewInvitationRepository(db, cfg, emailService, rbacRepo, auditRepo)
	passwordResetRepo := repository.NewPasswordResetRepository(db, cfg, emailService, auditRepo)
	oauthClientRepo := repository.NewOAuthClientRepository(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, jwtService, refreshRepo, passwordResetRepo, cfg.AuthSession)
	oauthHandler := handlers.NewOAuthHandler(oauthClientRepo, jwtService, int(cfg.AuthSession.AccessTokenTTL.Seconds()))
	userHandler := handlers.NewUserHandler(userService)
	adminHandler := handlers.NewAdminHandler(rbacRepo, userService, invitationRepo, auditRepo, refreshRepo)
	invitationHandler := handlers.NewInvitationHandler(invitationRepo, authHandler)
	passwordResetHandler := handlers.NewPasswordResetHandler(passwordResetRepo, emailService)

	// Create a new Gin router
	router := gin.New()
	router.Use(gin.Recovery(), sharedmw.ErrorLogger())
	router.SetTrustedProxies(nil)

	// Middleware for CSRF (double-submit) on unsafe methods
	router.Use(middleware.CSRFMiddleware())

	// Public routes (no authentication required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/csrf", authHandler.CSRF)
	router.POST("/oauth/token", oauthHandler.Token)
	router.POST("/login", authHandler.Login)
	router.POST("/refresh", authHandler.Refresh)
	router.POST("/logout", authHandler.Logout)
	router.POST("/password-reset/request", middleware.RateLimitMiddleware(5, 5), passwordResetHandler.RequestReset)
	router.POST("/password-reset/validate", middleware.RateLimitMiddleware(30, 10), passwordResetHandler.ValidateReset)
	router.POST("/password-reset/complete", middleware.RateLimitMiddleware(30, 10), passwordResetHandler.CompleteReset)
	// Invite-only registration
	router.POST("/invitations/validate", middleware.RateLimitMiddleware(30, 10), invitationHandler.ValidateInvitation)
	router.POST("/invitations/register", middleware.RateLimitMiddleware(10, 5), invitationHandler.RegisterInvitation)

	// Middleware for JWT authentication (applies to routes below)
	router.Use(middleware.RBACMiddleware(rbacRepo, jwtService))

	// User permissions endpoint - accessible to authenticated users
	router.GET("/permissions", adminHandler.GetUserPermissions)
	router.GET("/permissions/machine", adminHandler.GetMachinePermissions)

	// Admin routes - require admin role
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.RequireRole(rbacRepo, "admin"))
	{
		adminGroup.GET("/dashboard", adminHandler.GetAdminDashboard)
		adminGroup.GET("/users", adminHandler.GetUsersWithRoles)
		adminGroup.PUT("/users/:id/activate", adminHandler.ActivateUser)
		adminGroup.POST("/roles", adminHandler.CreateRole)
		adminGroup.GET("/roles", adminHandler.GetAllRoles)
		adminGroup.POST("/assign-role", adminHandler.AssignRoleToUser)
		adminGroup.POST("/remove-role", adminHandler.RemoveRoleFromUser)
		adminGroup.GET("/oauth/clients", oauthHandler.ListClients)
		adminGroup.POST("/oauth/clients", oauthHandler.CreateClient)
		adminGroup.PUT("/oauth/clients/:id/status", oauthHandler.UpdateClientStatus)
		adminGroup.POST("/oauth/clients/:id/rotate-secret", oauthHandler.RotateClientSecret)
		// Token blacklist management - admin only
		adminGroup.POST("/blacklist", userHandler.BlacklistToken)
		adminGroup.GET("/blacklist", userHandler.GetBlacklistedTokens)
		adminGroup.DELETE("/blacklist/:id", userHandler.DeleteBlacklistedToken)
	}

	// Routes for user management
	userGroup := router.Group("/users")

	// Self-service routes — any authenticated user may call these for their own account.
	userGroup.POST("/update-profile", userHandler.UpdateUserProfile)

	// Admin-only user management routes.
	userAdminGroup := userGroup.Group("")
	userAdminGroup.Use(middleware.RequireRole(rbacRepo, "admin"))
	userAdminGroup.GET("/", userHandler.GetAllUsers)
	userAdminGroup.GET("/:id", userHandler.GetUserByID)
	// POST /users/ creates a user with default role assignment via invitation flow.
	// POST /users/create-manually creates a user bypassing the invitation system (admin tool).
	userAdminGroup.POST("/", userHandler.CreateUser)
	userAdminGroup.PUT("/:id", userHandler.UpdateUser)
	userAdminGroup.DELETE("/:id", userHandler.DeleteUser)
	userAdminGroup.POST("/change-password", userHandler.ChangePassword)
	userAdminGroup.POST("/change-password-clear-flag", userHandler.ChangePasswordAndClearFlag)
	userAdminGroup.POST("/reset-password", userHandler.ResetPassword)
	userAdminGroup.POST("/create-manually", userHandler.CreateUserManually)
	userAdminGroup.POST("/validate-token", userHandler.ValidateToken)
	userAdminGroup.GET("/validate-token/:token", userHandler.ValidateTokenById)
	userAdminGroup.PUT("/theme-color", userHandler.UpdateUserThemeColor)

	invitationGroup := router.Group("/invitations")
	invitationGroup.Use(middleware.RequireRole(rbacRepo, "admin"))
	invitationGroup.GET("/pending", invitationHandler.ListPendingInvitations)
	invitationGroup.POST("", middleware.RateLimitMiddleware(10, 5), invitationHandler.CreateInvitation)
	invitationGroup.POST("/resend", middleware.RateLimitMiddleware(10, 5), invitationHandler.ResendInvitation)
	invitationGroup.POST("/revoke", invitationHandler.RevokeInvitation)
	return router
}
