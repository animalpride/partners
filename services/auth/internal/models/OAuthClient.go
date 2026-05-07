package models

import "time"

type OAuthClient struct {
	ID               int       `gorm:"column:id;primaryKey" json:"id"`
	ClientID         string    `gorm:"column:client_id;unique;not null" json:"client_id"`
	Name             string    `gorm:"column:name;not null" json:"name"`
	Description      string    `gorm:"column:description" json:"description"`
	ClientSecretHash string    `gorm:"column:client_secret_hash;not null" json:"-"`
	Active           int       `gorm:"column:active;default:1" json:"active"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

type ClientPermission struct {
	ID            int       `gorm:"column:id;primaryKey" json:"id"`
	OAuthClientID int       `gorm:"column:oauth_client_id;not null" json:"oauth_client_id"`
	PermissionID  int       `gorm:"column:permission_id;not null" json:"permission_id"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ClientPermission) TableName() string {
	return "client_permissions"
}
