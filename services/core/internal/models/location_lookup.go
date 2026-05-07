package models

import "time"

type LocationCountry struct {
	Code       string    `gorm:"type:char(2);primaryKey" json:"code"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	SearchName string    `gorm:"size:120;not null;index" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (LocationCountry) TableName() string {
	return "location_countries"
}

type LocationState struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CountryCode string    `gorm:"type:char(2);not null;index:idx_location_states_country_name,priority:1" json:"country_code"`
	StateCode   string    `gorm:"size:20;not null;uniqueIndex:ux_location_states_country_state_code,priority:2" json:"state_code"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	SearchName  string    `gorm:"size:160;not null;index:idx_location_states_country_name,priority:2" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (LocationState) TableName() string {
	return "location_states"
}

type LocationCity struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CountryCode string    `gorm:"type:char(2);not null;index:idx_location_cities_country_state,priority:1;index:idx_location_cities_country_search,priority:1" json:"country_code"`
	StateCode   string    `gorm:"size:20;not null;index:idx_location_cities_country_state,priority:2" json:"state_code"`
	StateName   string    `gorm:"size:120;not null" json:"state_name"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	SearchName  string    `gorm:"size:200;not null;index:idx_location_cities_country_search,priority:2" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (LocationCity) TableName() string {
	return "location_cities"
}

type LocationCityAlias struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CityID      uint      `gorm:"not null;index" json:"city_id"`
	Alias       string    `gorm:"size:160;not null" json:"alias"`
	SearchAlias string    `gorm:"size:160;not null;index" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

func (LocationCityAlias) TableName() string {
	return "location_city_aliases"
}
