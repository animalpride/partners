package repository

import (
	"github.com/animalpride/partners/services/core/internal/models"
	"gorm.io/gorm"
)

type PartnerApplicationRepository struct {
	db *gorm.DB
}

func NewPartnerApplicationRepository(db *gorm.DB) *PartnerApplicationRepository {
	return &PartnerApplicationRepository{db: db}
}

func (r *PartnerApplicationRepository) Create(app *models.PartnerApplication) error {
	return r.db.Create(app).Error
}

func (r *PartnerApplicationRepository) List(limit int) ([]models.PartnerApplication, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var apps []models.PartnerApplication
	err := r.db.Order("created_at desc").Limit(limit).Find(&apps).Error
	return apps, err
}
