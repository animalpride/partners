package models

import "time"

type CMSPage struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Slug        string    `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	ContentJSON string    `gorm:"type:longtext;not null" json:"content_json"`
	UpdatedBy   *int      `gorm:"column:updated_by" json:"updated_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (CMSPage) TableName() string {
	return "cms_pages"
}
