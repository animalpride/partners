package models

import "time"

type AuditEvent struct {
	ID           int       `gorm:"column:id" json:"id"`
	EventType    string    `gorm:"column:event_type" json:"event_type"`
	ActorUserID  *int      `gorm:"column:actor_user_id" json:"actor_user_id,omitempty"`
	TargetUserID *int      `gorm:"column:target_user_id" json:"target_user_id,omitempty"`
	TargetEmail  *string   `gorm:"column:target_email" json:"target_email,omitempty"`
	Metadata     *string   `gorm:"column:metadata" json:"metadata,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AuditEvent) TableName() string {
	return "audit_events"
}
