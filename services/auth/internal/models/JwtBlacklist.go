package models

import (
	"time"
)

type JwtBlacklist struct {
	ID        int       `gorm:"column:id" json:"id"`
	UserID    int       `gorm:"column:user_id" json:"user_id"`
	Token     string    `gorm:"column:token" json:"token"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (JwtBlacklist) TableName() string {
	return "jwt_blacklist"
}
