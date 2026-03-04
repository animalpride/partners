package models

import "time"

// Role represents a role in the system
type Role struct {
	ID          int       `gorm:"column:id;primaryKey" json:"id"`
	Name        string    `gorm:"column:name;unique;not null" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Active      int       `gorm:"column:active;default:1" json:"active"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}

// Permission represents a permission in the system
type Permission struct {
	ID          int       `gorm:"column:id;primaryKey" json:"id"`
	Name        string    `gorm:"column:name;unique;not null" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Resource    string    `gorm:"column:resource" json:"resource"` // e.g., "users", "roles", "invitations"
	Action      string    `gorm:"column:action" json:"action"`     // e.g., "create", "read", "update", "delete"
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID        int       `gorm:"column:id;primaryKey" json:"id"`
	UserID    int       `gorm:"column:user_id;not null" json:"user_id"`
	RoleID    int       `gorm:"column:role_id;not null" json:"role_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	ID           int       `gorm:"column:id;primaryKey" json:"id"`
	RoleID       int       `gorm:"column:role_id;not null" json:"role_id"`
	PermissionID int       `gorm:"column:permission_id;not null" json:"permission_id"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	Role       Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
