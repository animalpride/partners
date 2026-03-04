package models

import (
	"time"
)

type User struct {
	ID                   int        `gorm:"column:id" json:"id"`
	Email                string     `gorm:"column:email" json:"email"`
	PasswordHash         string     `gorm:"column:password_hash" json:"password_hash"`
	FirstName            string     `gorm:"column:first_name" json:"first_name"`
	LastName             string     `gorm:"column:last_name" json:"last_name"`
	Active               int        `gorm:"column:active" json:"active"`
	MustChangePassword   int        `gorm:"column:must_change_password" json:"must_change_password"`
	ThemeColor           string     `gorm:"column:theme_color;default:'#0d6efd'" json:"theme_color"`
	InvitedAt            *time.Time `gorm:"column:invited_at" json:"invited_at,omitempty"`
	InvitationToken      string     `gorm:"column:invitation_token" json:"invitation_token,omitempty"`
	InvitationAcceptedAt *time.Time `gorm:"column:invitation_accepted_at" json:"invitation_accepted_at,omitempty"`
	CreatedAt            time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
