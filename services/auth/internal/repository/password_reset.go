package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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

const passwordResetTTL = 15 * time.Minute

// PasswordResetRepository manages password reset tokens and flow.
type PasswordResetRepository struct {
	db           *gorm.DB
	emailService *services.EmailService
	auditRepo    *AuditRepository
}

func NewPasswordResetRepository(db *gorm.DB, cfg *config.Config, emailService *services.EmailService, auditRepo *AuditRepository) *PasswordResetRepository {
	return &PasswordResetRepository{
		db:           db,
		emailService: emailService,
		auditRepo:    auditRepo,
	}
}

func (r *PasswordResetRepository) RequestReset(email, requestIP string) error {
	cleanEmail := strings.TrimSpace(email)
	if cleanEmail == "" {
		return errors.New("email is required")
	}

	now := time.Now()
	var user models.User
	err := r.db.Where("email = ?", cleanEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = r.auditRepo.CreateEvent(
				"password_reset.requested",
				nil,
				nil,
				&cleanEmail,
				map[string]any{"email_found": false, "request_ip": requestIP},
			)
			return nil
		}
		log.Printf("RequestReset: lookup failed: %v", err)
		_ = r.auditRepo.CreateEvent(
			"password_reset.request_failed",
			nil,
			nil,
			&cleanEmail,
			map[string]any{"error": err.Error(), "request_ip": requestIP},
		)
		return err
	}

	_ = r.auditRepo.CreateEvent(
		"password_reset.requested",
		nil,
		&user.ID,
		&cleanEmail,
		map[string]any{"email_found": true, "request_ip": requestIP},
	)

	if user.Active != 1 {
		_ = r.auditRepo.CreateEvent(
			"password_reset.request_ignored",
			nil,
			&user.ID,
			&cleanEmail,
			map[string]any{"reason": "inactive_user", "request_ip": requestIP},
		)
		return nil
	}

	if count, err := r.invalidateActiveTokens(user.ID, now, "new_request", requestIP); err != nil {
		log.Printf("RequestReset: invalidate tokens failed: %v", err)
	} else if count > 0 {
		_ = r.auditRepo.CreateEvent(
			"password_reset.tokens_invalidated",
			nil,
			&user.ID,
			&cleanEmail,
			map[string]any{"reason": "new_request", "count": count, "request_ip": requestIP},
		)
	}

	token, hash, err := generateResetToken()
	if err != nil {
		log.Printf("RequestReset: token generation failed: %v", err)
		_ = r.auditRepo.CreateEvent(
			"password_reset.request_failed",
			nil,
			&user.ID,
			&cleanEmail,
			map[string]any{"error": err.Error(), "request_ip": requestIP},
		)
		return err
	}

	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: now.Add(passwordResetTTL),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.Create(resetToken).Error; err != nil {
		log.Printf("RequestReset: token create failed: %v", err)
		_ = r.auditRepo.CreateEvent(
			"password_reset.request_failed",
			nil,
			&user.ID,
			&cleanEmail,
			map[string]any{"error": err.Error(), "request_ip": requestIP},
		)
		return err
	}

	_ = r.auditRepo.CreateEvent(
		"password_reset.token_issued",
		nil,
		&user.ID,
		&cleanEmail,
		map[string]any{"expires_at": resetToken.ExpiresAt, "request_ip": requestIP},
	)

	if err := r.emailService.SendPasswordResetEmail(cleanEmail, token); err != nil {
		log.Printf("RequestReset: email send failed to %s: %v", cleanEmail, err)
		_ = r.auditRepo.CreateEvent(
			"password_reset.email_failed",
			nil,
			&user.ID,
			&cleanEmail,
			map[string]any{"error": err.Error(), "request_ip": requestIP},
		)
		return err
	}

	_ = r.auditRepo.CreateEvent(
		"password_reset.email_sent",
		nil,
		&user.ID,
		&cleanEmail,
		map[string]any{"request_ip": requestIP},
	)

	return nil
}

func (r *PasswordResetRepository) ValidateToken(token, requestIP string) (*models.PasswordResetToken, *models.User, error) {
	if strings.TrimSpace(token) == "" {
		return nil, nil, errors.New("token is required")
	}

	now := time.Now()
	var resetToken models.PasswordResetToken
	if err := r.db.Where("token_hash = ?", hashToken(token)).First(&resetToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = r.auditRepo.CreateEvent(
				"password_reset.validation_failed",
				nil,
				nil,
				nil,
				map[string]any{"reason": "token_not_found", "request_ip": requestIP},
			)
			return nil, nil, errors.New("reset link is invalid or expired")
		}
		log.Printf("ValidateToken: lookup failed: %v", err)
		return nil, nil, err
	}

	if resetToken.UsedAt != nil {
		_ = r.auditRepo.CreateEvent(
			"password_reset.validation_failed",
			nil,
			&resetToken.UserID,
			nil,
			map[string]any{"reason": "token_used", "request_ip": requestIP},
		)
		return nil, nil, errors.New("reset link is invalid or expired")
	}

	if now.After(resetToken.ExpiresAt) {
		_ = r.auditRepo.CreateEvent(
			"password_reset.validation_failed",
			nil,
			&resetToken.UserID,
			nil,
			map[string]any{"reason": "token_expired", "request_ip": requestIP},
		)
		return nil, nil, errors.New("reset link is invalid or expired")
	}

	var user models.User
	if err := r.db.Where("id = ?", resetToken.UserID).First(&user).Error; err != nil {
		log.Printf("ValidateToken: user lookup failed: %v", err)
		return nil, nil, errors.New("reset link is invalid or expired")
	}

	_ = r.auditRepo.CreateEvent(
		"password_reset.validated",
		nil,
		&user.ID,
		&user.Email,
		map[string]any{"request_ip": requestIP},
	)

	return &resetToken, &user, nil
}

func (r *PasswordResetRepository) CompleteReset(token, newPassword, requestIP string) (*models.User, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("token is required")
	}

	now := time.Now()
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var resetToken models.PasswordResetToken
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", hashToken(token)).First(&resetToken).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = r.auditRepo.CreateEvent(
				"password_reset.completed_failed",
				nil,
				nil,
				nil,
				map[string]any{"reason": "token_not_found", "request_ip": requestIP},
			)
			return nil, errors.New("reset link is invalid or expired")
		}
		log.Printf("CompleteReset: lookup failed: %v", err)
		return nil, err
	}

	if resetToken.UsedAt != nil || now.After(resetToken.ExpiresAt) {
		tx.Rollback()
		_ = r.auditRepo.CreateEvent(
			"password_reset.completed_failed",
			nil,
			&resetToken.UserID,
			nil,
			map[string]any{"reason": "token_invalid", "request_ip": requestIP},
		)
		return nil, errors.New("reset link is invalid or expired")
	}

	var user models.User
	if err := tx.Where("id = ?", resetToken.UserID).First(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("CompleteReset: user lookup failed: %v", err)
		return nil, errors.New("reset link is invalid or expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		log.Printf("CompleteReset: password hash failed: %v", err)
		return nil, errors.New("failed to update password")
	}

	updates := map[string]any{
		"password_hash":        string(hashedPassword),
		"must_change_password": 0,
		"updated_at":           now,
	}
	if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		tx.Rollback()
		log.Printf("CompleteReset: update user failed: %v", err)
		return nil, errors.New("failed to update password")
	}

	if err := tx.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", user.ID).
		Updates(map[string]any{"used_at": now, "updated_at": now}).Error; err != nil {
		tx.Rollback()
		log.Printf("CompleteReset: update tokens failed: %v", err)
		return nil, errors.New("failed to update password")
	}

	if err := tx.Model(&models.RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL", user.ID).
		Updates(map[string]any{"revoked_at": now, "updated_at": now}).Error; err != nil {
		tx.Rollback()
		log.Printf("CompleteReset: revoke sessions failed: %v", err)
		return nil, errors.New("failed to update password")
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("CompleteReset: commit failed: %v", err)
		return nil, err
	}

	_ = r.auditRepo.CreateEvent(
		"password_reset.completed",
		nil,
		&user.ID,
		&user.Email,
		map[string]any{"request_ip": requestIP},
	)

	return &user, nil
}

func (r *PasswordResetRepository) InvalidateActiveTokensOnLogin(userID int, requestIP string) (int64, error) {
	now := time.Now()
	count, err := r.invalidateActiveTokens(userID, now, "login", requestIP)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		_ = r.auditRepo.CreateEvent(
			"password_reset.invalidated_on_login",
			nil,
			&userID,
			nil,
			map[string]any{"count": count, "request_ip": requestIP},
		)
	}
	return count, nil
}

func (r *PasswordResetRepository) invalidateActiveTokens(userID int, now time.Time, reason, requestIP string) (int64, error) {
	res := r.db.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL AND expires_at > ?", userID, now).
		Updates(map[string]any{"used_at": now, "updated_at": now})
	if res.Error != nil {
		log.Printf("invalidateActiveTokens: update failed: %v", res.Error)
		_ = r.auditRepo.CreateEvent(
			"password_reset.invalidated_failed",
			nil,
			&userID,
			nil,
			map[string]any{"reason": reason, "error": res.Error.Error(), "request_ip": requestIP},
		)
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func generateResetToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(bytes)
	return token, hashToken(token), nil
}
