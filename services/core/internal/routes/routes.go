package routes

import (
	"log"

	"github.com/animalpride/animalpride-core/services/core/internal/config"
	"github.com/animalpride/animalpride-core/services/core/internal/handlers"
	"github.com/animalpride/animalpride-core/services/core/internal/middleware"
	"github.com/animalpride/animalpride-core/services/core/internal/repository"
	"github.com/animalpride/animalpride-core/services/core/internal/services"
	sharedmw "github.com/animalpride/animalpride-core/services/shared/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	cmsRepo := repository.NewCMSRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	leadEmailService := services.NewLeadEmailService(cfg)
	if err := cmsRepo.EnsureDefaults(); err != nil {
		log.Printf("failed to ensure default CMS pages: %v", err)
	}

	cmsHandler := handlers.NewCMSHandler(cmsRepo, leadRepo)
	partnerHandler := handlers.NewPartnerHandler(leadRepo, leadEmailService)

	router := gin.New()
	router.Use(gin.Recovery(), sharedmw.ErrorLogger())
	router.SetTrustedProxies(nil)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	public := router.Group("/cms")
	public.GET("/pages", cmsHandler.GetAllPages)
	public.GET("/pages/:slug", cmsHandler.GetPageBySlug)
	public.POST("/application", cmsHandler.SubmitLead)
	router.POST("/partners/leads", partnerHandler.SubmitLead)

	admin := router.Group("/cms/admin")
	admin.Use(sharedmw.AuthMiddleware(cfg.Auth.BaseURL))
	admin.Use(middleware.RequirePermission(cfg.Auth.BaseURL, "cms", "edit"))
	admin.PUT("/pages/:slug", cmsHandler.UpdatePage)
	admin.GET("/applications", cmsHandler.ListLeads)

	return router
}
