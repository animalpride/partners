package repository

import (
	"github.com/animalpride/partners/services/core/internal/models"
	"gorm.io/gorm"
)

type LeadRepository struct {
	db *gorm.DB
}

func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

func (r *LeadRepository) Create(lead *models.LeadSubmission) error {
	return r.db.Create(lead).Error
}

func (r *LeadRepository) List(limit int) ([]models.LeadSubmission, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var leads []models.LeadSubmission
	err := r.db.Order("created_at desc").Limit(limit).Find(&leads).Error
	return leads, err
}
