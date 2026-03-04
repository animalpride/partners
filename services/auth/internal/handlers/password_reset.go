package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/repository"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/services"
	"github.com/gin-gonic/gin"
)

type PasswordResetHandler struct {
	resetRepo    *repository.PasswordResetRepository
	emailService *services.EmailService
}

func NewPasswordResetHandler(resetRepo *repository.PasswordResetRepository, emailService *services.EmailService) *PasswordResetHandler {
	return &PasswordResetHandler{
		resetRepo:    resetRepo,
		emailService: emailService,
	}
}

func (h *PasswordResetHandler) RequestReset(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	requestIP := c.ClientIP()
	if err := h.resetRepo.RequestReset(strings.TrimSpace(request.Email), requestIP); err != nil {
		log.Printf("RequestReset: processing error: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email is registered, a reset link will be sent shortly.",
	})
}

func (h *PasswordResetHandler) ValidateReset(c *gin.Context) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, _, err := h.resetRepo.ValidateToken(strings.TrimSpace(request.Token), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset link is invalid or expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reset link is valid"})
}

func (h *PasswordResetHandler) CompleteReset(c *gin.Context) {
	var request struct {
		Token           string `json:"token" binding:"required"`
		Password        string `json:"password" binding:"required"`
		PasswordConfirm string `json:"password_confirm" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if request.Password != request.PasswordConfirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	if !passwordMeetsComplexity(request.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password does not meet complexity requirements"})
		return
	}

	user, err := h.resetRepo.CompleteReset(strings.TrimSpace(request.Token), request.Password, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset link is invalid or expired"})
		return
	}

	if err := h.emailService.SendPasswordChangedEmail(user.Email); err != nil {
		log.Printf("CompleteReset: password changed email failed for %s: %v", user.Email, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully. Please sign in with your new password.",
	})
}
