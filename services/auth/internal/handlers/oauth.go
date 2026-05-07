package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/animalpride/partners/services/auth/internal/models"
	"github.com/animalpride/partners/services/auth/internal/repository"
	"github.com/animalpride/partners/services/auth/internal/services"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OAuthHandler struct {
	clientRepo *repository.OAuthClientRepository
	jwtService *services.JWTService
	tokenTTL   int
}

func NewOAuthHandler(clientRepo *repository.OAuthClientRepository, jwtService *services.JWTService, tokenTTLSeconds int) *OAuthHandler {
	return &OAuthHandler{clientRepo: clientRepo, jwtService: jwtService, tokenTTL: tokenTTLSeconds}
}

func (h *OAuthHandler) Token(c *gin.Context) {
	grantType := strings.TrimSpace(c.PostForm("grant_type"))
	clientID := strings.TrimSpace(c.PostForm("client_id"))
	clientSecret := c.PostForm("client_secret")

	if grantType != "client_credentials" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
		return
	}
	if clientID == "" || strings.TrimSpace(clientSecret) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}

	client, err := h.clientRepo.GetByClientID(clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate client"})
		return
	}
	if client.Active != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
		return
	}
	if err := h.clientRepo.ValidateSecret(client, clientSecret); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
		return
	}

	permissions, err := h.clientRepo.GetPermissions(client.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve client permissions"})
		return
	}
	accessToken, err := h.jwtService.GenerateMachineAccessToken(client.ClientID, timeDurationFromSeconds(h.tokenTTL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   h.tokenTTL,
		"scope":        permissionsToScopeString(permissions),
	})
}

func (h *OAuthHandler) ListClients(c *gin.Context) {
	clients, err := h.clientRepo.List()
	if err != nil {
		log.Printf("ListClients: failed to list clients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch oauth clients"})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func (h *OAuthHandler) CreateClient(c *gin.Context) {
	var req struct {
		ClientID    string   `json:"client_id" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Scopes      []string `json:"scopes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	secret, secretHash, err := generateClientSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate client secret"})
		return
	}

	client := &models.OAuthClient{
		ClientID:         strings.TrimSpace(req.ClientID),
		Name:             strings.TrimSpace(req.Name),
		Description:      strings.TrimSpace(req.Description),
		ClientSecretHash: secretHash,
		Active:           1,
	}
	if err := h.clientRepo.Create(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create oauth client"})
		return
	}
	if err := h.clientRepo.SetPermissionsByScope(client.ID, req.Scopes); err != nil {
		log.Printf("CreateClient: failed to set permissions for client %d: %v", client.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	permissions, err := h.clientRepo.GetPermissions(client.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "client created but failed to load permissions"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"client":        client,
		"client_secret": secret,
		"scope":         permissionsToScopeString(permissions),
	})
}

func (h *OAuthHandler) RotateClientSecret(c *gin.Context) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
		return
	}

	secret, secretHash, err := generateClientSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate client secret"})
		return
	}
	if err := h.clientRepo.UpdateSecretHash(id, secretHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate client secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"client_secret": secret})
}

func (h *OAuthHandler) UpdateClientStatus(c *gin.Context) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
		return
	}

	var req struct {
		Active *int `json:"active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	if *req.Active != 0 && *req.Active != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "active must be 0 or 1"})
		return
	}
	if err := h.clientRepo.UpdateActive(id, *req.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update client status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client status updated successfully"})
}

func permissionsToScopeString(permissions []models.Permission) string {
	scopes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		scopes = append(scopes, permission.Resource+":"+permission.Action)
	}
	return strings.Join(scopes, " ")
}

func generateClientSecret() (string, string, error) {
	secretBytes := make([]byte, 24)
	if _, err := rand.Read(secretBytes); err != nil {
		log.Printf("generateClientSecret: rand failed: %v", err)
		return "", "", err
	}
	secret := "opc_" + hex.EncodeToString(secretBytes)
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("generateClientSecret: hash failed: %v", err)
		return "", "", err
	}
	return secret, string(hash), nil
}

func timeDurationFromSeconds(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
