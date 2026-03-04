package models

import "time"

type PasswordResetToken struct {
	ID        int        `gorm:"column:id" json:"id"`
	UserID    int        `gorm:"column:user_id" json:"user_id"`
	TokenHash string     `gorm:"column:token_hash" json:"token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
