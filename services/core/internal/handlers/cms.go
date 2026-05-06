package handlers

import (
	"errors"
	"net/http"

	"github.com/animalpride/partners/services/core/internal/models"
	"github.com/animalpride/partners/services/core/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CMSHandler struct {
	cmsRepo  *repository.CMSRepository
	leadRepo *repository.LeadRepository
}

func NewCMSHandler(cmsRepo *repository.CMSRepository, leadRepo *repository.LeadRepository) *CMSHandler {
	return &CMSHandler{cmsRepo: cmsRepo, leadRepo: leadRepo}
}

func (h *CMSHandler) GetAllPages(c *gin.Context) {
	pages, err := h.cmsRepo.GetAllPages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pages"})
		return
	}
	c.JSON(http.StatusOK, pages)
}

func (h *CMSHandler) GetPageBySlug(c *gin.Context) {
	slug := c.Param("slug")
	page, err := h.cmsRepo.GetPageBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch page"})
		return
	}
	if page == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *CMSHandler) UpdatePage(c *gin.Context) {
	slug := c.Param("slug")
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		ContentJSON string `json:"content_json" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	page, err := h.cmsRepo.UpdatePage(slug, req.Title, req.Description, req.ContentJSON, 0)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update page"})
		return
	}

	c.JSON(http.StatusOK, page)
}

func (h *CMSHandler) SubmitLead(c *gin.Context) {
	var req struct {
		OrganizationName string `json:"organization_name" binding:"required"`
		ContactName      string `json:"contact_name" binding:"required"`
		Email            string `json:"email" binding:"required,email"`
		Phone            string `json:"phone"`
		Website          string `json:"website"`
		MonthlyTraffic   string `json:"monthly_traffic"`
		CurrentStore     string `json:"current_store"`
		Goals            string `json:"goals"`
		Notes            string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	lead := &models.LeadSubmission{
		OrganizationName: req.OrganizationName,
		ContactName:      req.ContactName,
		Email:            req.Email,
		Phone:            req.Phone,
		Website:          req.Website,
		MonthlyTraffic:   req.MonthlyTraffic,
		CurrentStore:     req.CurrentStore,
		Goals:            req.Goals,
		Notes:            req.Notes,
	}

	if err := h.leadRepo.Create(lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit application"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "application submitted"})
}

func (h *CMSHandler) ListLeads(c *gin.Context) {
	leads, err := h.leadRepo.List(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch submissions"})
		return
	}
	c.JSON(http.StatusOK, leads)
}
