package repository

import (
	"encoding/json"
	"log"

	"github.com/animalpride/partners/services/auth/internal/models"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) CreateEvent(eventType string, actorUserID *int, targetUserID *int, targetEmail *string, metadata map[string]any) error {
	var payload *string
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			encoded := string(b)
			payload = &encoded
		} else {
			log.Printf("CreateEvent: metadata marshal failed: %v", err)
		}
	}
	entry := &models.AuditEvent{
		EventType:    eventType,
		ActorUserID:  actorUserID,
		TargetUserID: targetUserID,
		TargetEmail:  targetEmail,
		Metadata:     payload,
	}
	if err := r.db.Create(entry).Error; err != nil {
		log.Printf("CreateEvent: create failed: %v", err)
		return err
	}
	return nil
}
