package routes

import (
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/config"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/handlers"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/middleware"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/repository"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/services"
	sharedmw "github.com/animalpride/animalpride-core/services/shared/middleware"
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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, jwtService, refreshRepo, passwordResetRepo, cfg.AuthSession)
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
		// Token blacklist management - admin only
		adminGroup.POST("/blacklist", userHandler.BlacklistToken)
		adminGroup.GET("/blacklist", userHandler.GetBlacklistedTokens)
		adminGroup.DELETE("/blacklist/:id", userHandler.DeleteBlacklistedToken)
	}

	// Routes for user management
	userGroup := router.Group("/users")
	userGroup.GET("/", userHandler.GetAllUsers)
	userGroup.GET("/:id", userHandler.GetUserByID) // change to only allow int
	// POST /users/ creates a user with default role assignment via invitation flow.
	// POST /users/create-manually creates a user bypassing the invitation system (admin tool).
	userGroup.POST("/", userHandler.CreateUser)
	userGroup.PUT("/:id", userHandler.UpdateUser)    // change to only allow int
	userGroup.DELETE("/:id", userHandler.DeleteUser) // change to only allow int
	userGroup.POST("/change-password", userHandler.ChangePassword)
	userGroup.POST("/change-password-clear-flag", userHandler.ChangePasswordAndClearFlag)
	userGroup.POST("/update-profile", userHandler.UpdateUserProfile)
	userGroup.POST("/reset-password", userHandler.ResetPassword)
	userGroup.POST("/create-manually", userHandler.CreateUserManually)
	userGroup.POST("/validate-token", userHandler.ValidateToken)
	userGroup.GET("/validate-token/:token", userHandler.ValidateTokenById)
	userGroup.PUT("/theme-color", userHandler.UpdateUserThemeColor)

	invitationGroup := router.Group("/invitations")
	invitationGroup.Use(middleware.RequireRole(rbacRepo, "admin"))
	invitationGroup.GET("/pending", invitationHandler.ListPendingInvitations)
	invitationGroup.POST("", middleware.RateLimitMiddleware(10, 5), invitationHandler.CreateInvitation)
	invitationGroup.POST("/resend", middleware.RateLimitMiddleware(10, 5), invitationHandler.ResendInvitation)
	invitationGroup.POST("/revoke", invitationHandler.RevokeInvitation)
	return router
}
