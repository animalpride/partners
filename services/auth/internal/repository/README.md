# repository/

This directory contains repository implementations for data access logic, typically using GORM and MySQL.

## Example

```go
package repository

import (
	"gorm.io/gorm"
)

// User represents a user model.
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string
	Email string
}

// UserRepository provides access to user storage.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
```

Place your GORM-based data access code here, organized by domain entity.
