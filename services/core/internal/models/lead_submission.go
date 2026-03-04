package models

import "time"

type LeadSubmission struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	OrganizationName string    `gorm:"size:255;not null" json:"organization_name"`
	ContactName      string    `gorm:"size:255;not null" json:"contact_name"`
	Email            string    `gorm:"size:255;not null" json:"email"`
	Phone            string    `gorm:"size:50" json:"phone"`
	Website          string    `gorm:"size:255" json:"website"`
	MonthlyTraffic   string    `gorm:"size:100" json:"monthly_traffic"`
	CurrentStore     string    `gorm:"size:255" json:"current_store"`
	Goals            string    `gorm:"type:text" json:"goals"`
	Notes            string    `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
}

func (LeadSubmission) TableName() string {
	return "partner_lead_submissions"
}
