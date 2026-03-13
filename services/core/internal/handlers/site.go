package handlers

import (
	"net/http"
	"strings"

	"github.com/animalpride/animalpride-core/services/core/internal/config"
	"github.com/gin-gonic/gin"
)

type SiteHandler struct {
	cfg *config.Config
}

func NewSiteHandler(cfg *config.Config) *SiteHandler {
	return &SiteHandler{cfg: cfg}
}

func (h *SiteHandler) GetComingSoonState(c *gin.Context) {
	comingSoonCfg := h.cfg.Site.ComingSoon
	unlocked := h.hasPreviewAccess(c)

	c.JSON(http.StatusOK, gin.H{
		"enabled":          comingSoonCfg.Enabled,
		"preview_unlocked": unlocked,
		"message":          comingSoonCfg.Message,
	})
}

func (h *SiteHandler) UnlockPreview(c *gin.Context) {
	comingSoonCfg := h.cfg.Site.ComingSoon
	token := strings.TrimSpace(c.Param("token"))
	expected := strings.TrimSpace(strings.TrimPrefix(comingSoonCfg.PreviewPath, "/"))

	if expected == "" || token == "" || token != expected {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	maxAgeSeconds := comingSoonCfg.PreviewCookieTTLHour * 3600
	c.SetCookie(
		comingSoonCfg.PreviewCookieName,
		"1",
		maxAgeSeconds,
		"/",
		"",
		comingSoonCfg.PreviewCookieSecure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"unlocked": true})
}

func (h *SiteHandler) hasPreviewAccess(c *gin.Context) bool {
	if !h.cfg.Site.ComingSoon.Enabled {
		return true
	}

	cookieVal, err := c.Cookie(h.cfg.Site.ComingSoon.PreviewCookieName)
	return err == nil && cookieVal == "1"
}
