package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/config"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/models"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/repository"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo         *repository.UserRepository
	jwtService       *services.JWTService
	refreshRepo      *repository.RefreshSessionRepository
	resetRepo        *repository.PasswordResetRepository
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	rotationInterval time.Duration
}

func NewAuthHandler(us *repository.UserRepository, js *services.JWTService, rs *repository.RefreshSessionRepository, resetRepo *repository.PasswordResetRepository, authSession config.AuthSession) *AuthHandler {
	return &AuthHandler{
		userRepo:         us,
		jwtService:       js,
		refreshRepo:      rs,
		resetRepo:        resetRepo,
		accessTokenTTL:   authSession.AccessTokenTTL,
		refreshTokenTTL:  authSession.RefreshTokenTTL,
		rotationInterval: authSession.RefreshRotationInterval,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		log.Printf("Invalid login input: %v", err)
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}
	log.Printf("Login attempt for user: %s", loginRequest.Email)
	user, err := h.userRepo.Authenticate(loginRequest.Email, loginRequest.Password)
	if err != nil {
		log.Printf("Authentication failed for user %s: %v", loginRequest.Email, err)
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	if _, err := h.resetRepo.InvalidateActiveTokensOnLogin(user.ID, c.ClientIP()); err != nil {
		log.Printf("Login: reset token invalidation failed for user %d: %v", user.ID, err)
	}

	h.issueSession(c, user)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, err := h.userIDFromAccessOrHeader(c)
	if err != nil {
		userID, err = h.userIDFromRefreshCookie(c)
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	_ = h.refreshRepo.RevokeByUserID(userID, time.Now())
	clearAuthCookies(c)
	c.JSON(200, gin.H{"message": "Logged out successfully"})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		log.Printf("Refresh token missing or empty: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
		return
	}

	claims, err := h.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Printf("Invalid refresh token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	userID, err := userIDFromClaims(claims)
	if err != nil {
		log.Printf("Failed to extract user ID from refresh token claims: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	familyID, _ := claims["family_id"].(string)
	if familyID == "" {
		log.Printf("Family ID missing in refresh token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	session, err := h.refreshRepo.GetByUserID(userID)
	if err != nil || session == nil {
		log.Printf("Refresh session not found for user %d: %v", userID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
		return
	}
	if session.FamilyID != familyID || session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		log.Printf("Refresh session invalid for user %d", userID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		return
	}

	providedHash := hashToken(refreshToken)
	previousHash := ""
	if session.PreviousTokenHash != nil {
		previousHash = *session.PreviousTokenHash
	}

	if providedHash != session.CurrentTokenHash && providedHash != previousHash {
		_ = h.refreshRepo.RevokeByUserID(userID, time.Now())
		log.Printf("Refresh token reuse detected for user %d", userID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh reuse detected"})
		return
	}

	if providedHash == previousHash && time.Since(session.LastRotatedAt) >= h.rotationInterval {
		_ = h.refreshRepo.RevokeByUserID(userID, time.Now())
		log.Printf("Refresh token reuse detected for user %d", userID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh reuse detected"})
		return
	}

	accessToken, err := h.jwtService.GenerateAccessToken(userID, h.accessTokenTTL)
	if err != nil {
		log.Printf("Failed to generate access token for user %d: %v", userID, err)
		c.JSON(500, gin.H{"error": "Failed to generate access token"})
		return
	}
	refreshTokenID, err := randomHex(32)
	if err != nil {
		log.Printf("Failed to generate refresh token ID for user %d: %v", userID, err)
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	newRefreshToken, err := h.jwtService.GenerateRefreshToken(userID, familyID, refreshTokenID, h.refreshTokenTTL)
	if err != nil {
		log.Printf("Failed to generate refresh token for user %d: %v", userID, err)
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	now := time.Now()
	newHash := hashToken(newRefreshToken)
	if now.Sub(session.LastRotatedAt) >= h.rotationInterval {
		prev := session.CurrentTokenHash
		session.PreviousTokenHash = &prev
		session.CurrentTokenHash = newHash
		session.LastRotatedAt = now
	} else {
		session.PreviousTokenHash = &newHash
	}
	session.LastUsedAt = &now
	session.ExpiresAt = now.Add(h.refreshTokenTTL)
	if err := h.refreshRepo.Update(session); err != nil {
		log.Printf("Failed to update refresh session for user %d: %v", userID, err)
		c.JSON(500, gin.H{"error": "Failed to update session"})
		return
	}

	setAccessCookie(c, accessToken, h.accessTokenTTL)
	setRefreshCookie(c, newRefreshToken, h.refreshTokenTTL)
	setCSRFCookie(c, generateCSRFToken(), h.refreshTokenTTL)

	c.JSON(200, gin.H{"message": "Token refreshed successfully"})
}

// CSRF issues a new CSRF cookie for double-submit protection.
func (h *AuthHandler) CSRF(c *gin.Context) {
	setCSRFCookie(c, generateCSRFToken(), h.refreshTokenTTL)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) issueSession(c *gin.Context, user *models.User) {
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, h.accessTokenTTL)
	if err != nil {
		log.Printf("Failed to generate access token for user %d: %v", user.ID, err)
		c.JSON(500, gin.H{"error": "Failed to generate access token"})
		return
	}

	familyID, err := randomHex(32)
	if err != nil {
		log.Printf("Failed to generate family ID for user %d: %v", user.ID, err)
		c.JSON(500, gin.H{"error": "Failed to generate session"})
		return
	}
	refreshTokenID, err := randomHex(32)
	if err != nil {
		log.Printf("Failed to generate refresh token ID for user %d: %v", user.ID, err)
		c.JSON(500, gin.H{"error": "Failed to generate session"})
		return
	}
	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID, familyID, refreshTokenID, h.refreshTokenTTL)
	if err != nil {
		log.Printf("Failed to generate refresh token for user %d: %v", user.ID, err)
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	now := time.Now()
	_ = h.refreshRepo.RevokeByUserID(user.ID, now)

	refreshSession := &models.RefreshSession{
		UserID:           user.ID,
		FamilyID:         familyID,
		CurrentTokenHash: hashToken(refreshToken),
		LastRotatedAt:    now,
		LastUsedAt:       &now,
		ExpiresAt:        now.Add(h.refreshTokenTTL),
	}
	if err := h.refreshRepo.Upsert(refreshSession); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create session"})
		return
	}

	setAccessCookie(c, accessToken, h.accessTokenTTL)
	setRefreshCookie(c, refreshToken, h.refreshTokenTTL)
	setCSRFCookie(c, generateCSRFToken(), h.refreshTokenTTL)

	response := gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
		},
	}

	if user.MustChangePassword == 1 {
		response["must_change_password"] = true
	}

	c.JSON(200, response)
}

func (h *AuthHandler) userIDFromAccessOrHeader(c *gin.Context) (int, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return h.jwtService.ValidateAccessToken(authHeader[7:])
	}
	accessToken, err := c.Cookie("access_token")
	if err != nil || accessToken == "" {
		log.Printf("Access token missing or empty: %v", err)
		return 0, err
	}
	return h.jwtService.ValidateAccessToken(accessToken)
}

func (h *AuthHandler) userIDFromRefreshCookie(c *gin.Context) (int, error) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		log.Printf("Refresh token missing or empty: %v", err)
		return 0, err
	}
	claims, err := h.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Printf("Invalid refresh token: %v", err)
		return 0, err
	}
	return userIDFromClaims(claims)
}

func setAccessCookie(c *gin.Context, token string, ttl time.Duration) {
	secure := cookieSecure(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func setRefreshCookie(c *gin.Context, token string, ttl time.Duration) {
	secure := cookieSecure(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/auth/v1/refresh",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func setCSRFCookie(c *gin.Context, token string, ttl time.Duration) {
	secure := cookieSecure(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookies(c *gin.Context) {
	secure := cookieSecure(c)
	clearCookie := func(name, path string, httpOnly bool) {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     path,
			MaxAge:   0,
			HttpOnly: httpOnly,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
	}
	clearCookie("access_token", "/", true)
	clearCookie("refresh_token", "/api/auth/v1/refresh", true)
	clearCookie("csrf_token", "/", false)
}

func generateCSRFToken() string {
	val, err := randomHex(32)
	if err != nil {
		return ""
	}
	return val
}

func cookieSecure(_ *gin.Context) bool {
	return false
}

func randomHex(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func userIDFromClaims(claims map[string]interface{}) (int, error) {
	if userIDFloat, ok := claims["user_id"].(float64); ok {
		return int(userIDFloat), nil
	}
	if userIDInt, ok := claims["user_id"].(int); ok {
		return userIDInt, nil
	}
	if userIDStr, ok := claims["user_id"].(string); ok {
		return strconv.Atoi(userIDStr)
	}
	return 0, fmt.Errorf("invalid user_id claim")
}
