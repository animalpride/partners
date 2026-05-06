package models

import "time"

type PartnerApplication struct {
	ID               string    `gorm:"type:char(36);primaryKey" json:"id"`
	OrganizationName string    `gorm:"size:255;not null" json:"organization_name"`
	ContactName      string    `gorm:"size:255;not null" json:"contact_name"`
	Email            string    `gorm:"size:255;not null" json:"email"`
	Phone            string    `gorm:"size:50;not null" json:"phone"`
	AddressLine1     string    `gorm:"size:255;not null" json:"address_line1"`
	AddressLine2     string    `gorm:"size:255" json:"address_line2"`
	City             string    `gorm:"size:100;not null" json:"city"`
	CityLookupID     *uint     `gorm:"index" json:"city_lookup_id,omitempty"`
	State            string    `gorm:"size:100;not null" json:"state"`
	StateCode        string    `gorm:"size:20" json:"state_code"`
	PostalCode       string    `gorm:"size:20;not null" json:"postal_code"`
	Country          string    `gorm:"size:100;not null" json:"country"`
	CountryCode      string    `gorm:"type:char(2);index" json:"country_code"`
	Website          string    `gorm:"size:255" json:"website"`
	MonthlyTraffic   string    `gorm:"size:100" json:"monthly_traffic"`
	CurrentStore     string    `gorm:"size:255" json:"current_store"`
	Goals            string    `gorm:"type:text" json:"goals"`
	Notes            string    `gorm:"type:text" json:"notes"`
	DenopsSyncStatus string    `gorm:"size:20;not null;default:pending" json:"denops_sync_status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (PartnerApplication) TableName() string {
	return "partner_applications"
}
