package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/animalpride/partners/services/auth/internal/config"
	"github.com/animalpride/partners/services/auth/internal/models"
	"github.com/animalpride/partners/services/auth/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository provides methods for user management
type UserRepository struct {
	db           *gorm.DB
	cfg          *config.Config
	emailService *services.EmailService
	rbacRepo     *RBACRepository
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB, cfg *config.Config, emailService *services.EmailService, rbacRepo *RBACRepository) *UserRepository {
	return &UserRepository{
		db:           db,
		cfg:          cfg,
		emailService: emailService,
		rbacRepo:     rbacRepo,
	}
}

// GetAllUsers retrieves all users
func (s *UserRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		log.Printf("GetAllUsers: query failed: %v", err)
		return nil, err
	}
	return users, nil
}

// GetUserByID retrieves a user by ID
func (s *UserRepository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		log.Printf("GetUserByID: query failed: %v", err)
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (s *UserRepository) CreateUser(user *models.User) error {
	if err := s.db.Create(user).Error; err != nil {
		log.Printf("CreateUser: create failed: %v", err)
		return err
	}
	return nil
}

// UpdateUser updates an existing user
func (s *UserRepository) UpdateUser(user *models.User) error {
	if err := s.db.Save(user).Error; err != nil {
		log.Printf("UpdateUser: save failed: %v", err)
		return err
	}
	return nil
}

// DeleteUser deletes a user by ID
func (s *UserRepository) DeleteUser(id string) error {
	if err := s.db.Delete(&models.User{}, id).Error; err != nil {
		log.Printf("DeleteUser: delete failed: %v", err)
		return err
	}
	return nil
}

// ChangePassword changes the password of a user
func (s *UserRepository) ChangePassword(userID int, newPassword string) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("ChangePassword: lookup failed: %v", err)
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ChangePassword: hash failed: %v", err)
		return err
	}
	user.PasswordHash = string(hashedPassword)
	if err := s.db.Save(&user).Error; err != nil {
		log.Printf("ChangePassword: save failed: %v", err)
		return err
	}
	return nil
}

// ResetPassword resets the password of a user
func (s *UserRepository) ResetPassword(userID int, newPassword string) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("ResetPassword: lookup failed: %v", err)
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ResetPassword: hash failed: %v", err)
		return err
	}
	user.PasswordHash = string(hashedPassword)
	if err := s.db.Save(&user).Error; err != nil {
		log.Printf("ResetPassword: save failed: %v", err)
		return err
	}
	return nil
}

// InviteUser sends an invitation to a user
func (s *UserRepository) InviteUser(email string) error {
	var user models.User
	now := time.Now()

	// Generate invitation token
	token, err := s.generateInvitationToken()
	if err != nil {
		log.Printf("InviteUser: token generation failed: %v", err)
		return fmt.Errorf("failed to generate invitation token: %v", err)
	}

	// Check if user already exists
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new user
			user = models.User{
				Email:           email,
				Active:          0, // Inactive until they accept invitation
				InvitedAt:       &now,
				InvitationToken: token,
				CreatedAt:       now,
			}
			if err := s.db.Create(&user).Error; err != nil {
				log.Printf("InviteUser: create user failed: %v", err)
				return fmt.Errorf("failed to create user: %v", err)
			}
		} else {
			log.Printf("InviteUser: lookup failed: %v", err)
			return fmt.Errorf("database error: %v", err)
		}
	} else {
		// User exists, update invitation details
		user.InvitedAt = &now
		user.InvitationToken = token
		user.InvitationAcceptedAt = nil // Reset acceptance if re-inviting
		if err := s.db.Save(&user).Error; err != nil {
			log.Printf("InviteUser: update user failed: %v", err)
			return fmt.Errorf("failed to update user: %v", err)
		}
	}

	// Send invitation email
	recipientName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if recipientName == " " {
		recipientName = ""
	}

	if err := s.emailService.SendInvitationEmail(email, recipientName, "", token); err != nil {
		// Log the error but don't fail the invitation - the user was created/updated
		log.Printf("InviteUser: failed to send invitation email to %s: %v", email, err)
	}

	return nil
}

// AcceptInvitation accepts an invitation for a user
func (s *UserRepository) AcceptInvitation(userID string, token string) error {
	var user models.User
	if err := s.db.Where("id = ? AND invitation_token = ?", userID, token).First(&user).Error; err != nil {
		log.Printf("AcceptInvitation: lookup failed: %v", err)
		return err
	}

	now := time.Now()
	user.InvitationAcceptedAt = &now
	if err := s.db.Save(&user).Error; err != nil {
		log.Printf("AcceptInvitation: save failed: %v", err)
		return err
	}
	return nil
}

// GetUserByEmail retrieves a user by email
func (s *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("GetUserByEmail: lookup failed: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetUserByToken retrieves a user by token
func (s *UserRepository) GetUserByToken(token string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("invitation_token = ?", token).First(&user).Error; err != nil {
		log.Printf("GetUserByToken: lookup failed: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetUserByEmailAndToken retrieves a user by email and token
func (s *UserRepository) GetUserByEmailAndToken(email string, token string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ? AND invitation_token = ?", email, token).First(&user).Error; err != nil {
		log.Printf("GetUserByEmailAndToken: lookup failed: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetUserByEmailAndPassword retrieves a user by email and password
func (s *UserRepository) GetUserByEmailAndPassword(email string, password string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("GetUserByEmailAndPassword: lookup failed: %v", err)
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Printf("GetUserByEmailAndPassword: password mismatch: %v", err)
		return nil, err
	}
	return &user, nil
}

// Authenticate checks if the user exists and verifies the password
func (s *UserRepository) Authenticate(username, password string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", username).First(&user).Error; err != nil {
		log.Printf("Authenticate: lookup failed: %v", err)
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if user.Active != 1 {
		log.Printf("Authenticate: user deactivated: %s", username)
		return nil, fmt.Errorf("user is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Printf("Authenticate: invalid password: %v", err)
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	return &user, nil
}

// ValidateToken checks if the token is valid
func (s *UserRepository) ValidateToken(token string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("invitation_token = ?", token).First(&user).Error; err != nil {
		log.Printf("ValidateToken: lookup failed: %v", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return &user, nil
}

// ResendInvitation resends an invitation to a user
func (s *UserRepository) ResendInvitation(email string) error {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("ResendInvitation: lookup failed: %v", err)
		return fmt.Errorf("user not found: %v", err)
	}

	// Generate new invitation token
	token, err := s.generateInvitationToken()
	if err != nil {
		log.Printf("ResendInvitation: token generation failed: %v", err)
		return fmt.Errorf("failed to generate invitation token: %v", err)
	}

	now := time.Now()
	user.InvitedAt = &now
	user.InvitationToken = token
	user.InvitationAcceptedAt = nil // Reset acceptance

	if err := s.db.Save(&user).Error; err != nil {
		log.Printf("ResendInvitation: update user failed: %v", err)
		return fmt.Errorf("failed to update user: %v", err)
	}

	// Send invitation email
	recipientName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if recipientName == " " {
		recipientName = ""
	}

	if err := s.emailService.SendInvitationEmail(email, recipientName, "", token); err != nil {
		// Log the error but don't fail the resend - the user was updated
		log.Printf("ResendInvitation: failed to send invitation email to %s: %v", email, err)
	}

	return nil
}

// ValidateTokenById checks if the token is valid by user ID
func (s *UserRepository) ValidateTokenById(userID string, token string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND invitation_token = ?", userID, token).First(&user).Error; err != nil {
		log.Printf("ValidateTokenById: lookup failed: %v", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return &user, nil
}

// BlacklistToken blacklists a token
func (s *UserRepository) BlacklistToken(token string) error {
	var user models.User
	if err := s.db.Where("invitation_token = ?", token).First(&user).Error; err != nil {
		log.Printf("BlacklistToken: lookup failed: %v", err)
		return err
	}
	user.InvitationToken = ""
	if err := s.db.Save(&user).Error; err != nil {
		log.Printf("BlacklistToken: save failed: %v", err)
		return err
	}
	return nil
}

// DeleteBlacklistedToken deletes a blacklisted token
func (s *UserRepository) DeleteBlacklistedToken(token string) error {
	var user models.User
	if err := s.db.Where("invitation_token = ?", token).First(&user).Error; err != nil {
		log.Printf("DeleteBlacklistedToken: lookup failed: %v", err)
		return err
	}
	user.InvitationToken = ""
	if err := s.db.Save(&user).Error; err != nil {
		log.Printf("DeleteBlacklistedToken: save failed: %v", err)
		return err
	}
	return nil
}

// GetBlacklistedTokens retrieves all blacklisted tokens
func (s *UserRepository) GetBlacklistedTokens() ([]models.JwtBlacklist, error) {
	var tokens []models.JwtBlacklist
	if err := s.db.Find(&tokens).Error; err != nil {
		log.Printf("GetBlacklistedTokens: query failed: %v", err)
		return nil, err
	}
	return tokens, nil
}

// UpdateUserStatus updates the active status of a user
func (s *UserRepository) UpdateUserStatus(userID int, active int) error {
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("active", active).Error; err != nil {
		log.Printf("UpdateUserStatus: update failed: %v", err)
		return err
	}
	return nil
}

// UpdateUserThemeColor updates user's theme color preference
func (s *UserRepository) UpdateUserThemeColor(userID int, themeColor string) error {
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("theme_color", themeColor).Error; err != nil {
		log.Printf("UpdateUserThemeColor: update failed: %v", err)
		return err
	}
	return nil
}

// generateInvitationToken generates a random invitation token
func (s *UserRepository) generateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("generateInvitationToken: rand read failed: %v", err)
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RemoveInvitation removes a pending invitation by deleting the user if they haven't accepted
func (s *UserRepository) RemoveInvitation(email string) error {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("RemoveInvitation: lookup failed: %v", err)
		return fmt.Errorf("user not found: %v", err)
	}

	// Only allow removing invitations that haven't been accepted yet
	if user.InvitationAcceptedAt != nil {
		log.Printf("RemoveInvitation: invitation already accepted for %s", email)
		return fmt.Errorf("cannot remove invitation for user who has already accepted")
	}

	// Delete the user record since they were only created for the invitation
	if err := s.db.Delete(&user).Error; err != nil {
		log.Printf("RemoveInvitation: delete failed: %v", err)
		return fmt.Errorf("failed to remove invitation: %v", err)
	}

	return nil
}

// CreateUserManually creates a user with default password and basic role
func (s *UserRepository) CreateUserManually(email string) error {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		log.Printf("CreateUserManually: email already exists: %s", email)
		return fmt.Errorf("user with email %s already exists", email)
	}

	// Hash default password "changethis"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("changethis"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("CreateUserManually: hash failed: %v", err)
		return fmt.Errorf("failed to hash password: %v", err)
	}

	now := time.Now()
	user := models.User{
		Email:                email,
		PasswordHash:         string(hashedPassword),
		Active:               1, // Active immediately
		MustChangePassword:   1, // Require password change on first login
		InvitedAt:            &now,
		InvitationAcceptedAt: &now, // Mark as accepted since manually created
		CreatedAt:            now,
	}

	if err := s.db.Create(&user).Error; err != nil {
		log.Printf("CreateUserManually: create failed: %v", err)
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Assign 'user' role to the newly created user
	userRole, err := s.rbacRepo.GetRoleByName("user")
	if err != nil {
		// Log the error but don't fail user creation - they just won't have any roles initially
		log.Printf("CreateUserManually: role lookup failed: %v", err)
		return nil
	}

	if err := s.rbacRepo.AssignRoleToUser(user.ID, userRole.ID); err != nil {
		// Log the error but don't fail user creation
		log.Printf("CreateUserManually: assign role failed for user %d: %v", user.ID, err)
	}

	return nil
}

// UpdateUserProfile updates user profile information and clears must_change_password flag
func (s *UserRepository) UpdateUserProfile(userID int, firstName, lastName string) error {
	updates := map[string]interface{}{
		"first_name":           firstName,
		"last_name":            lastName,
		"must_change_password": 0, // Clear the flag once profile is updated
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		log.Printf("UpdateUserProfile: update failed: %v", err)
		return err
	}
	return nil
}

// ChangePasswordAndClearFlag changes password and clears must_change_password flag
func (s *UserRepository) ChangePasswordAndClearFlag(userID int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ChangePasswordAndClearFlag: hash failed: %v", err)
		return fmt.Errorf("failed to hash password: %v", err)
	}

	updates := map[string]interface{}{
		"password_hash":        string(hashedPassword),
		"must_change_password": 0, // Clear the flag once password is changed
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		log.Printf("ChangePasswordAndClearFlag: update failed: %v", err)
		return err
	}
	return nil
}
