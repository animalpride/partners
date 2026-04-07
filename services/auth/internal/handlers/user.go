package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/models"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepository *repository.UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepository *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepository: userRepository,
	}
}

// GetAllUsers retrieves all users
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userRepository.GetAllUsers()
	if err != nil {
		log.Printf("GetAllUsers: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUserByID retrieves a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userRepository.GetUserByID(id)
	if err != nil {
		log.Printf("GetUserByID: db error for id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("CreateUser: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user.CreatedAt = time.Now()
	if err := h.userRepository.CreateUser(&user); err != nil {
		log.Printf("CreateUser: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("UpdateUser: invalid input for id %s: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	userId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("UpdateUser: invalid user ID %q: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	user.ID = userId
	if err := h.userRepository.UpdateUser(&user); err != nil {
		log.Printf("UpdateUser: db error for user %d: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.userRepository.DeleteUser(id); err != nil {
		log.Printf("DeleteUser: db error for id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ChangePassword changes a user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var request struct {
		UserId      int    `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ChangePassword: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.ChangePassword(request.UserId, request.NewPassword); err != nil {
		log.Printf("ChangePassword: db error for user %d: %v", request.UserId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// ResetPassword resets a user's password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var request struct {
		UserId      int    `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ResetPassword: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.ResetPassword(request.UserId, request.NewPassword); err != nil {
		log.Printf("ResetPassword: db error for user %d: %v", request.UserId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// InviteUser invites a new user
func (h *UserHandler) InviteUser(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("InviteUser: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.InviteUser(request.Email); err != nil {
		log.Printf("InviteUser: db error for %s: %v", request.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invite user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User invited successfully"})
}

// AcceptInvitation accepts a user invitation
func (h *UserHandler) AcceptInvitation(c *gin.Context) {
	var request struct {
		UserId string `json:"user_id"`
		Token  string `json:"token"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("AcceptInvitation: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.AcceptInvitation(request.UserId, request.Token); err != nil {
		log.Printf("AcceptInvitation: db error for user %s: %v", request.UserId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted successfully"})
}

// ResendInvitation resends a user invitation
func (h *UserHandler) ResendInvitation(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ResendInvitation: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.ResendInvitation(request.Email); err != nil {
		log.Printf("ResendInvitation: db error for %s: %v", request.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend invitation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation resent successfully"})
}

// BlacklistToken blacklists a JWT token
func (h *UserHandler) BlacklistToken(c *gin.Context) {
	var request struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("BlacklistToken: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.BlacklistToken(request.Token); err != nil {
		log.Printf("BlacklistToken: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to blacklist token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Token blacklisted successfully"})
}

// GetBlacklistedTokens retrieves all blacklisted tokens
func (h *UserHandler) GetBlacklistedTokens(c *gin.Context) {
	tokens, err := h.userRepository.GetBlacklistedTokens()
	if err != nil {
		log.Printf("GetBlacklistedTokens: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve blacklisted tokens"})
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// DeleteBlacklistedToken deletes a blacklisted token
func (h *UserHandler) DeleteBlacklistedToken(c *gin.Context) {
	id := c.Param("id")
	if err := h.userRepository.DeleteBlacklistedToken(id); err != nil {
		log.Printf("DeleteBlacklistedToken: db error for id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blacklisted token"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ValidateToken validates a JWT token
func (h *UserHandler) ValidateToken(c *gin.Context) {
	var request struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ValidateToken: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user, err := h.userRepository.ValidateToken(request.Token)
	if err != nil {
		log.Printf("ValidateToken: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Token is valid", "user": user})
}

// ValidateTokenById validates a JWT token by ID
func (h *UserHandler) ValidateTokenById(c *gin.Context) {
	id := c.Param("id")
	token := c.Param("token")
	user, err := h.userRepository.ValidateTokenById(id, token)
	if err != nil {
		log.Printf("ValidateTokenById: db error for id %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// RegisterUserRequest represents the registration request
type RegisterUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// RegisterUser creates a new user with password hashing
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("RegisterUser: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("RegisterUser: bcrypt error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user model
	now := time.Now()
	user := models.User{
		Email:                req.Email,
		PasswordHash:         string(hashedPassword),
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		Active:               1, // Active by default
		InvitedAt:            &now,
		InvitationAcceptedAt: &now,
		CreatedAt:            time.Now(),
	}

	if err := h.userRepository.CreateUser(&user); err != nil {
		log.Printf("RegisterUser: db error for %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Don't return password hash in response
	user.PasswordHash = ""
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// RemoveInvitation removes a pending invitation
func (h *UserHandler) RemoveInvitation(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("RemoveInvitation: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.RemoveInvitation(request.Email); err != nil {
		log.Printf("RemoveInvitation: db error for %s: %v", request.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove invitation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation removed successfully"})
}

// CreateUserManually creates a user manually with default password
func (h *UserHandler) CreateUserManually(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("CreateUserManually: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.CreateUserManually(request.Email); err != nil {
		log.Printf("CreateUserManually: db error for %s: %v", request.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user manually"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User created manually successfully"})
}

// UpdateUserProfile updates user profile information and clears must_change_password flag
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	var request struct {
		FirstName          string `json:"first_name" binding:"required"`
		LastName           string `json:"last_name" binding:"required"`
		ChangePassword     bool   `json:"change_password"`
		NewPassword        string `json:"new_password"`
		NewPasswordConfirm string `json:"new_password_confirm"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("UpdateUserProfile: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if request.ChangePassword {
		if request.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password is required"})
			return
		}
		if len(request.NewPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
			return
		}
		if request.NewPassword != request.NewPasswordConfirm {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
			return
		}
	}

	if err := h.userRepository.UpdateUserProfile(userID, request.FirstName, request.LastName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	if request.ChangePassword {
		if err := h.userRepository.ChangePasswordAndClearFlag(userID, request.NewPassword); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}
	}

	user, err := h.userRepository.GetUserByID(strconv.Itoa(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile updated successfully",
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
		},
	})
}

// ChangePasswordAndClearFlag changes password and clears must_change_password flag
func (h *UserHandler) ChangePasswordAndClearFlag(c *gin.Context) {
	var request struct {
		UserId      int    `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.ChangePasswordAndClearFlag(request.UserId, request.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// UpdateUserThemeColor updates user's theme color preference
func (h *UserHandler) UpdateUserThemeColor(c *gin.Context) {
	var request struct {
		UserId     int    `json:"user_id"`
		ThemeColor string `json:"theme_color"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.userRepository.UpdateUserThemeColor(request.UserId, request.ThemeColor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theme color"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Theme color updated successfully"})
}
