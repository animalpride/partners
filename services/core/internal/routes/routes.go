package routes

import (
	"log"

	"github.com/animalpride/partners/services/core/internal/config"
	"github.com/animalpride/partners/services/core/internal/handlers"
	"github.com/animalpride/partners/services/core/internal/middleware"
	"github.com/animalpride/partners/services/core/internal/repository"
	"github.com/animalpride/partners/services/core/internal/services"
	sharedmw "github.com/animalpride/partners/services/shared/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	cmsRepo := repository.NewCMSRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	appRepo := repository.NewPartnerApplicationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	leadEmailService := services.NewLeadEmailService(cfg)
	if err := cmsRepo.EnsureDefaults(); err != nil {
		log.Printf("failed to ensure default CMS pages: %v", err)
	}

	cmsHandler := handlers.NewCMSHandler(cmsRepo, leadRepo)
	partnerHandler := handlers.NewPartnerHandler(appRepo, locationRepo, leadEmailService)
	locationHandler := handlers.NewLocationHandler(locationRepo)
	siteHandler := handlers.NewSiteHandler(cfg)

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

	router.GET("/site/coming-soon", siteHandler.GetComingSoonState)
	router.POST("/site/coming-soon/unlock/:token", siteHandler.UnlockPreview)
	router.POST("/partners/leads", partnerHandler.SubmitLead)
	router.GET("/partners/locations/countries", locationHandler.GetCountries)
	router.GET("/partners/locations/city-states", locationHandler.SearchCityStates)

	admin := router.Group("/cms/admin")
	admin.Use(sharedmw.AuthMiddleware(cfg.Auth.BaseURL))
	admin.Use(middleware.RequirePermission(cfg.Auth.BaseURL, "cms", "edit"))
	admin.Use(sharedmw.CSRFMiddleware())
	admin.PUT("/pages/:slug", cmsHandler.UpdatePage)
	admin.GET("/applications", cmsHandler.ListLeads)

	return router
}
