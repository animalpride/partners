package handlers

import (
	"net/http"

	"github.com/animalpride/animalpride-core/services/core/internal/models"
	"github.com/animalpride/animalpride-core/services/core/internal/repository"
	"github.com/gin-gonic/gin"
)

type PartnerLeadMailer interface {
	SendPartnerLead(lead *models.LeadSubmission) error
}

type PartnerHandler struct {
	leadRepo *repository.LeadRepository
	mailer   PartnerLeadMailer
}

func NewPartnerHandler(leadRepo *repository.LeadRepository, mailer PartnerLeadMailer) *PartnerHandler {
	return &PartnerHandler{leadRepo: leadRepo, mailer: mailer}
}

func (h *PartnerHandler) SubmitLead(c *gin.Context) {
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

	if err := h.mailer.SendPartnerLead(lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "application saved but failed to send notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "application submitted"})
}
