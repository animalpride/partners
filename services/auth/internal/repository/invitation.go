package repository

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/animalpride/partners/services/auth/internal/config"
	"github.com/animalpride/partners/services/auth/internal/models"
	"github.com/animalpride/partners/services/auth/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	invitationStatusPending  = "pending"
	invitationStatusAccepted = "accepted"
	invitationStatusRevoked  = "revoked"
	invitationStatusExpired  = "expired"
	invitationTTL            = 48 * time.Hour
)

type InvitationRepository struct {
	db           *gorm.DB
	cfg          *config.Config
	emailService *services.EmailService
	rbacRepo     *RBACRepository
	auditRepo    *AuditRepository
}

func NewInvitationRepository(db *gorm.DB, cfg *config.Config, emailService *services.EmailService, rbacRepo *RBACRepository, auditRepo *AuditRepository) *InvitationRepository {
	return &InvitationRepository{
		db:           db,
		cfg:          cfg,
		emailService: emailService,
		rbacRepo:     rbacRepo,
		auditRepo:    auditRepo,
	}
}

func (r *InvitationRepository) CreateInvitation(email string, roleID int, inviterID int) error {
	if err := r.ensureNoUser(email); err != nil {
		log.Printf("CreateInvitation: ensure user failed: %v", err)
		return err
	}

	now := time.Now()
	if err := r.expireStaleInvitations(now); err != nil {
		log.Printf("CreateInvitation: expire stale invitations failed: %v", err)
		return err
	}

	var existing models.Invitation
	if err := r.db.Where("email = ? AND status = ?", email, invitationStatusPending).First(&existing).Error; err == nil {
		log.Printf("CreateInvitation: pending invitation already exists for %s", email)
		return errors.New("pending invitation already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("CreateInvitation: lookup pending invitation failed: %v", err)
		return err
	}

	role, err := r.rbacRepo.GetRoleByID(roleID)
	if err != nil {
		log.Printf("CreateInvitation: role lookup failed: %v", err)
		return fmt.Errorf("role not found: %v", err)
	}

	token, hash, nonce, err := r.generateInvitationToken()
	if err != nil {
		log.Printf("CreateInvitation: token generation failed: %v", err)
		return err
	}

	inviter := inviterID
	invitation := &models.Invitation{
		Email:           email,
		RoleID:          roleID,
		Status:          invitationStatusPending,
		ExpiresAt:       now.Add(invitationTTL),
		TokenHash:       hash,
		TokenNonce:      nonce,
		InvitedByUserID: &inviter,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := r.db.Create(invitation).Error; err != nil {
		log.Printf("CreateInvitation: create invitation failed: %v", err)
		return err
	}

	if err := r.emailService.SendInvitationEmail(email, "", role.Name, token); err != nil {
		fmt.Printf("Warning: Failed to send invitation email to %s: %v\n", email, err)
	}

	_ = r.auditRepo.CreateEvent(
		"invitation.created",
		&inviter,
		nil,
		&email,
		map[string]any{"role_id": role.ID, "role_name": role.Name},
	)

	return nil
}

func (r *InvitationRepository) ResendInvitation(email string, inviterID int) error {
	if err := r.expireStaleInvitations(time.Now()); err != nil {
		log.Printf("ResendInvitation: expire stale invitations failed: %v", err)
		return err
	}

	var invitation models.Invitation
	if err := r.db.Where("email = ? AND status = ?", email, invitationStatusPending).First(&invitation).Error; err != nil {
		log.Printf("ResendInvitation: pending invitation lookup failed: %v", err)
		return fmt.Errorf("pending invitation not found: %v", err)
	}

	token, err := r.rebuildToken(invitation.TokenNonce)
	if err != nil {
		log.Printf("ResendInvitation: rebuild token failed: %v", err)
		return err
	}
	if hashToken(token) != invitation.TokenHash {
		log.Printf("ResendInvitation: invitation token hash mismatch for %s", email)
		return errors.New("invitation token mismatch")
	}

	resendRole, err := r.rbacRepo.GetRoleByID(invitation.RoleID)
	resendRoleName := ""
	if err != nil {
		log.Printf("ResendInvitation: role lookup failed (non-fatal): %v", err)
	} else {
		resendRoleName = resendRole.Name
	}

	if err := r.emailService.SendInvitationEmail(email, "", resendRoleName, token); err != nil {
		fmt.Printf("Warning: Failed to resend invitation email to %s: %v\n", email, err)
	}

	_ = r.auditRepo.CreateEvent("invitation.resent", &inviterID, nil, &email, nil)
	return nil
}

func (r *InvitationRepository) RevokeInvitation(email string, actorID int) error {
	if err := r.expireStaleInvitations(time.Now()); err != nil {
		log.Printf("RevokeInvitation: expire stale invitations failed: %v", err)
		return err
	}

	var invitation models.Invitation
	if err := r.db.Where("email = ? AND status = ?", email, invitationStatusPending).First(&invitation).Error; err != nil {
		log.Printf("RevokeInvitation: pending invitation lookup failed: %v", err)
		return fmt.Errorf("pending invitation not found: %v", err)
	}

	now := time.Now()
	updates := map[string]any{
		"status":     invitationStatusRevoked,
		"revoked_at": now,
	}
	if err := r.db.Model(&models.Invitation{}).Where("id = ?", invitation.ID).Updates(updates).Error; err != nil {
		log.Printf("RevokeInvitation: update invitation failed: %v", err)
		return err
	}

	_ = r.auditRepo.CreateEvent("invitation.revoked", &actorID, nil, &email, nil)
	return nil
}

func (r *InvitationRepository) ValidateInvitation(token string) (*models.Invitation, *models.Role, error) {
	if token == "" {
		log.Printf("ValidateInvitation: missing token")
		return nil, nil, errors.New("token is required")
	}

	now := time.Now()
	if err := r.expireStaleInvitations(now); err != nil {
		log.Printf("ValidateInvitation: expire stale invitations failed: %v", err)
		return nil, nil, err
	}

	var invitation models.Invitation
	if err := r.db.Where("token_hash = ?", hashToken(token)).First(&invitation).Error; err != nil {
		log.Printf("ValidateInvitation: invitation token lookup failed: %v", err)
		return nil, nil, errors.New("invalid invitation token")
	}

	if invitation.Status != invitationStatusPending {
		log.Printf("ValidateInvitation: invitation status invalid: %s", invitation.Status)
		return nil, nil, errors.New("invitation is no longer valid")
	}

	if now.After(invitation.ExpiresAt) {
		_ = r.db.Model(&models.Invitation{}).Where("id = ?", invitation.ID).
			Updates(map[string]any{"status": invitationStatusExpired}).Error
		log.Printf("ValidateInvitation: invitation expired")
		return nil, nil, errors.New("invitation expired")
	}

	role, err := r.rbacRepo.GetRoleByID(invitation.RoleID)
	if err != nil {
		log.Printf("ValidateInvitation: role lookup failed: %v", err)
		return nil, nil, fmt.Errorf("role not found: %v", err)
	}

	return &invitation, role, nil
}

func (r *InvitationRepository) RegisterInvitation(token, firstName, lastName, password string) (*models.User, *models.Role, error) {
	if token == "" {
		log.Printf("RegisterInvitation: missing token")
		return nil, nil, errors.New("token is required")
	}

	if strings.TrimSpace(password) == "" {
		log.Printf("RegisterInvitation: missing password")
		return nil, nil, errors.New("password is required")
	}

	now := time.Now()
	if err := r.expireStaleInvitations(now); err != nil {
		log.Printf("RegisterInvitation: expire stale invitations failed: %v", err)
		return nil, nil, err
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("RegisterInvitation: begin transaction failed: %v", tx.Error)
		return nil, nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var invitation models.Invitation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", hashToken(token)).First(&invitation).Error; err != nil {
		tx.Rollback()
		log.Printf("RegisterInvitation: invitation token lookup failed: %v", err)
		return nil, nil, errors.New("invalid invitation token")
	}

	if invitation.Status != invitationStatusPending {
		tx.Rollback()
		log.Printf("RegisterInvitation: invitation status invalid: %s", invitation.Status)
		return nil, nil, errors.New("invitation is no longer valid")
	}

	if now.After(invitation.ExpiresAt) {
		_ = tx.Model(&models.Invitation{}).Where("id = ?", invitation.ID).
			Updates(map[string]any{"status": invitationStatusExpired}).Error
		tx.Rollback()
		log.Printf("RegisterInvitation: invitation expired")
		return nil, nil, errors.New("invitation expired")
	}

	var existingUser models.User
	if err := tx.Where("email = ?", invitation.Email).First(&existingUser).Error; err == nil {
		tx.Rollback()
		log.Printf("RegisterInvitation: email already in use")
		return nil, nil, errors.New("email already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		log.Printf("RegisterInvitation: lookup user failed: %v", err)
		return nil, nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		log.Printf("RegisterInvitation: password hash failed: %v", err)
		return nil, nil, errors.New("failed to hash password")
	}

	user := &models.User{
		Email:        invitation.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Active:       1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		log.Printf("RegisterInvitation: create user failed: %v", err)
		return nil, nil, err
	}

	userRole := &models.UserRole{
		UserID: user.ID,
		RoleID: invitation.RoleID,
	}
	if err := tx.Create(userRole).Error; err != nil {
		tx.Rollback()
		log.Printf("RegisterInvitation: assign role failed: %v", err)
		return nil, nil, err
	}

	acceptUpdates := map[string]any{
		"status":      invitationStatusAccepted,
		"accepted_at": now,
	}
	if err := tx.Model(&models.Invitation{}).Where("id = ?", invitation.ID).Updates(acceptUpdates).Error; err != nil {
		tx.Rollback()
		log.Printf("RegisterInvitation: update invitation failed: %v", err)
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("RegisterInvitation: commit failed: %v", err)
		return nil, nil, err
	}

	role, err := r.rbacRepo.GetRoleByID(invitation.RoleID)
	if err != nil {
		log.Printf("RegisterInvitation: role lookup failed: %v", err)
		return user, nil, fmt.Errorf("role not found: %v", err)
	}

	_ = r.auditRepo.CreateEvent(
		"invitation.accepted",
		invitation.InvitedByUserID,
		&user.ID,
		&invitation.Email,
		map[string]any{"role_id": invitation.RoleID, "role_name": role.Name},
	)

	return user, role, nil
}

// InvitationSummary is a view of an Invitation enriched with the role name.
type InvitationSummary struct {
	models.Invitation
	RoleName string `json:"role_name"`
}

func (r *InvitationRepository) ListPendingInvitations() ([]InvitationSummary, error) {
	now := time.Now()
	if err := r.expireStaleInvitations(now); err != nil {
		log.Printf("ListPendingInvitations: expire stale invitations failed: %v", err)
		return nil, err
	}

	var invitations []models.Invitation
	if err := r.db.Where("status = ?", invitationStatusPending).
		Where("expires_at > ?", now).
		Order("created_at desc").
		Find(&invitations).Error; err != nil {
		log.Printf("ListPendingInvitations: query failed: %v", err)
		return nil, err
	}

	// Collect unique role IDs and bulk-fetch roles.
	roleIDSet := make(map[int]struct{})
	for _, inv := range invitations {
		roleIDSet[inv.RoleID] = struct{}{}
	}
	roleMap := make(map[int]string, len(roleIDSet))
	for roleID := range roleIDSet {
		role, err := r.rbacRepo.GetRoleByID(roleID)
		if err != nil {
			log.Printf("ListPendingInvitations: role lookup failed for id %d: %v", roleID, err)
			roleMap[roleID] = ""
		} else {
			roleMap[roleID] = role.Name
		}
	}

	summaries := make([]InvitationSummary, len(invitations))
	for i, inv := range invitations {
		summaries[i] = InvitationSummary{
			Invitation: inv,
			RoleName:   roleMap[inv.RoleID],
		}
	}
	return summaries, nil
}

func (r *InvitationRepository) ensureNoUser(email string) error {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err == nil {
		log.Printf("ensureNoUser: email already in use: %s", email)
		return errors.New("email already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("ensureNoUser: lookup failed: %v", err)
		return err
	}
	return nil
}

func (r *InvitationRepository) expireStaleInvitations(now time.Time) error {
	if err := r.db.Model(&models.Invitation{}).
		Where("status = ? AND expires_at <= ?", invitationStatusPending, now).
		Updates(map[string]any{"status": invitationStatusExpired}).Error; err != nil {
		log.Printf("expireStaleInvitations: update failed: %v", err)
		return err
	}
	return nil
}

func (r *InvitationRepository) generateInvitationToken() (string, string, string, error) {
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		log.Printf("generateInvitationToken: rand read failed: %v", err)
		return "", "", "", err
	}
	nonce := hex.EncodeToString(nonceBytes)
	token, err := r.rebuildToken(nonce)
	if err != nil {
		log.Printf("generateInvitationToken: rebuild token failed: %v", err)
		return "", "", "", err
	}
	return token, hashToken(token), nonce, nil
}

func (r *InvitationRepository) rebuildToken(nonce string) (string, error) {
	mac := hmac.New(sha256.New, []byte(r.cfg.JWTSecret))
	if _, err := mac.Write([]byte(nonce)); err != nil {
		log.Printf("rebuildToken: hmac write failed: %v", err)
		return "", err
	}
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", nonce, signature), nil
}
