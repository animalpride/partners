package models

import "time"

type Invitation struct {
	ID              int        `gorm:"column:id" json:"id"`
	Email           string     `gorm:"column:email" json:"email"`
	RoleID          int        `gorm:"column:role_id" json:"role_id"`
	Status          string     `gorm:"column:status" json:"status"`
	ExpiresAt       time.Time  `gorm:"column:expires_at" json:"expires_at"`
	TokenHash       string     `gorm:"column:token_hash" json:"-"`
	TokenNonce      string     `gorm:"column:token_nonce" json:"-"`
	InvitedByUserID *int       `gorm:"column:invited_by_user_id" json:"invited_by_user_id,omitempty"`
	AcceptedAt      *time.Time `gorm:"column:accepted_at" json:"accepted_at,omitempty"`
	RevokedAt       *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Invitation) TableName() string {
	return "invitations"
}
