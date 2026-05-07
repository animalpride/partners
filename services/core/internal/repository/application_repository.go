package repository

import (
	"strings"

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

func (r *PartnerApplicationRepository) ListByStatus(status string, limit int) ([]models.PartnerApplication, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var apps []models.PartnerApplication
	err := r.db.
		Where("denops_sync_status = ?", strings.TrimSpace(strings.ToLower(status))).
		Order("created_at desc").
		Limit(limit).
		Find(&apps).Error
	return apps, err
}

func (r *PartnerApplicationRepository) GetByID(id string) (*models.PartnerApplication, error) {
	var app models.PartnerApplication
	err := r.db.Where("id = ?", strings.TrimSpace(id)).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *PartnerApplicationRepository) UpdateStatus(id, status string) error {
	result := r.db.Model(&models.PartnerApplication{}).
		Where("id = ?", strings.TrimSpace(id)).
		Update("denops_sync_status", strings.TrimSpace(strings.ToLower(status)))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
