package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/animalpride/partners/services/auth/internal/repository"
	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	invitationRepo *repository.InvitationRepository
	authHandler    *AuthHandler
}

func NewInvitationHandler(invitationRepo *repository.InvitationRepository, authHandler *AuthHandler) *InvitationHandler {
	return &InvitationHandler{
		invitationRepo: invitationRepo,
		authHandler:    authHandler,
	}
}

func (h *InvitationHandler) CreateInvitation(c *gin.Context) {
	var request struct {
		Email  string `json:"email" binding:"required,email"`
		RoleID int    `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("CreateInvitation: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	actorID := c.GetInt("user_id")
	if err := h.invitationRepo.CreateInvitation(strings.TrimSpace(request.Email), request.RoleID, actorID); err != nil {
		log.Printf("CreateInvitation: repo error for %s: %v", request.Email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation sent"})
}

func (h *InvitationHandler) ResendInvitation(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ResendInvitation: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	actorID := c.GetInt("user_id")
	if err := h.invitationRepo.ResendInvitation(strings.TrimSpace(request.Email), actorID); err != nil {
		log.Printf("ResendInvitation: repo error for %s: %v", request.Email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation resent"})
}

func (h *InvitationHandler) RevokeInvitation(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("RevokeInvitation: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	actorID := c.GetInt("user_id")
	if err := h.invitationRepo.RevokeInvitation(strings.TrimSpace(request.Email), actorID); err != nil {
		log.Printf("RevokeInvitation: repo error for %s: %v", request.Email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation revoked"})
}

func (h *InvitationHandler) ListPendingInvitations(c *gin.Context) {
	invitations, err := h.invitationRepo.ListPendingInvitations()
	if err != nil {
		log.Printf("ListPendingInvitations: repo error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invitations"})
		return
	}
	c.JSON(http.StatusOK, invitations)
}

func (h *InvitationHandler) ValidateInvitation(c *gin.Context) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ValidateInvitation: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	invitation, role, err := h.invitationRepo.ValidateInvitation(strings.TrimSpace(request.Token))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"email":      invitation.Email,
		"role_id":    invitation.RoleID,
		"role_name":  role.Name,
		"expires_at": invitation.ExpiresAt,
	})
}

func (h *InvitationHandler) RegisterInvitation(c *gin.Context) {
	var request struct {
		Token           string `json:"token" binding:"required"`
		FirstName       string `json:"first_name" binding:"required"`
		LastName        string `json:"last_name" binding:"required"`
		Password        string `json:"password" binding:"required,min=8"`
		PasswordConfirm string `json:"password_confirm" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("RegisterInvitation: invalid input: %v", err)
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

	user, _, err := h.invitationRepo.RegisterInvitation(
		strings.TrimSpace(request.Token),
		strings.TrimSpace(request.FirstName),
		strings.TrimSpace(request.LastName),
		request.Password,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.authHandler.issueSession(c, user)
}
