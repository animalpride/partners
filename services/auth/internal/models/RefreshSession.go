package models

import "time"

type RefreshSession struct {
	ID                int        `gorm:"column:id" json:"id"`
	UserID            int        `gorm:"column:user_id" json:"user_id"`
	FamilyID          string     `gorm:"column:family_id" json:"family_id"`
	CurrentTokenHash  string     `gorm:"column:current_token_hash" json:"current_token_hash"`
	PreviousTokenHash *string    `gorm:"column:previous_token_hash" json:"previous_token_hash"`
	LastRotatedAt     time.Time  `gorm:"column:last_rotated_at" json:"last_rotated_at"`
	LastUsedAt        *time.Time `gorm:"column:last_used_at" json:"last_used_at"`
	ExpiresAt         time.Time  `gorm:"column:expires_at" json:"expires_at"`
	RevokedAt         *time.Time `gorm:"column:revoked_at" json:"revoked_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type RefreshSessionUpdate struct {
	FamilyID          string     `gorm:"column:family_id" json:"family_id"`
	CurrentTokenHash  string     `gorm:"column:current_token_hash" json:"current_token_hash"`
	PreviousTokenHash *string    `gorm:"column:previous_token_hash" json:"previous_token_hash"`
	LastRotatedAt     time.Time  `gorm:"column:last_rotated_at" json:"last_rotated_at"`
	LastUsedAt        *time.Time `gorm:"column:last_used_at" json:"last_used_at"`
	ExpiresAt         time.Time  `gorm:"column:expires_at" json:"expires_at"`
	RevokedAt         *time.Time `gorm:"column:revoked_at" json:"revoked_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (RefreshSession) TableName() string {
	return "refresh_sessions"
}
