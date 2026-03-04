package repository

import (
	"errors"
	"log"
	"time"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/models"
	"gorm.io/gorm"
)

type RefreshSessionRepository struct {
	db *gorm.DB
}

func NewRefreshSessionRepository(db *gorm.DB) *RefreshSessionRepository {
	return &RefreshSessionRepository{db: db}
}

func (r *RefreshSessionRepository) GetByUserID(userID int) (*models.RefreshSession, error) {
	var session models.RefreshSession
	if err := r.db.Where("user_id = ?", userID).First(&session).Error; err != nil {
		log.Printf("GetByUserID: query failed: %v", err)
		return nil, err
	}
	return &session, nil
}

func (r *RefreshSessionRepository) GetByFamilyID(familyID string) (*models.RefreshSession, error) {
	var session models.RefreshSession
	if err := r.db.Where("family_id = ?", familyID).First(&session).Error; err != nil {
		log.Printf("GetByFamilyID: query failed: %v", err)
		return nil, err
	}
	return &session, nil
}

func (r *RefreshSessionRepository) Upsert(session *models.RefreshSession) error {
	var existing models.RefreshSession
	err := r.db.Where("user_id = ?", session.UserID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.db.Omit("CreatedAt", "UpdatedAt").Create(session).Error; err != nil {
				log.Printf("Upsert: create failed: %v", err)
				return err
			}
			return nil
		}
		log.Printf("Upsert: lookup failed: %v", err)
		return err
	}

	session.ID = existing.ID
	update := models.RefreshSessionUpdate{
		FamilyID:          session.FamilyID,
		CurrentTokenHash:  session.CurrentTokenHash,
		PreviousTokenHash: session.PreviousTokenHash,
		LastRotatedAt:     session.LastRotatedAt,
		LastUsedAt:        session.LastUsedAt,
		ExpiresAt:         session.ExpiresAt,
		RevokedAt:         session.RevokedAt,
		UpdatedAt:         time.Now(),
	}
	if err := r.db.Model(&models.RefreshSession{}).
		Where("id = ?", existing.ID).
		Select(
			"family_id",
			"current_token_hash",
			"previous_token_hash",
			"last_rotated_at",
			"last_used_at",
			"expires_at",
			"revoked_at",
			"updated_at",
		).
		Updates(update).Error; err != nil {
		log.Printf("Upsert: update failed: %v", err)
		return err
	}
	return nil
}

func (r *RefreshSessionRepository) Update(session *models.RefreshSession) error {
	update := models.RefreshSessionUpdate{
		FamilyID:          session.FamilyID,
		CurrentTokenHash:  session.CurrentTokenHash,
		PreviousTokenHash: session.PreviousTokenHash,
		LastRotatedAt:     session.LastRotatedAt,
		LastUsedAt:        session.LastUsedAt,
		ExpiresAt:         session.ExpiresAt,
		RevokedAt:         session.RevokedAt,
		UpdatedAt:         time.Now(),
	}
	if err := r.db.Model(&models.RefreshSession{}).
		Where("id = ?", session.ID).
		Select(
			"family_id",
			"current_token_hash",
			"previous_token_hash",
			"last_rotated_at",
			"last_used_at",
			"expires_at",
			"revoked_at",
			"updated_at",
		).
		Updates(update).Error; err != nil {
		log.Printf("Update: update failed: %v", err)
		return err
	}
	return nil
}

func (r *RefreshSessionRepository) RevokeByUserID(userID int, revokedAt time.Time) error {
	if err := r.db.Model(&models.RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]interface{}{
			"revoked_at": revokedAt,
			"updated_at": time.Now(),
		}).Error; err != nil {
		log.Printf("RevokeByUserID: update failed: %v", err)
		return err
	}
	return nil
}
