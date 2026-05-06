package repository

import (
	"errors"
	"log"

	"github.com/animalpride/partners/services/auth/internal/models"
	"gorm.io/gorm"
)

type RBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

// Role operations
func (r *RBACRepository) CreateRole(role *models.Role) error {
	if err := r.db.Create(role).Error; err != nil {
		log.Printf("CreateRole: create failed: %v", err)
		return err
	}
	return nil
}

func (r *RBACRepository) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Where("active = 1").Find(&roles).Error
	if err != nil {
		log.Printf("GetAllRoles: query failed: %v", err)
	}
	return roles, err
}

func (r *RBACRepository) GetRoleByID(id int) (*models.Role, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		log.Printf("GetRoleByID: query failed: %v", err)
		return nil, err
	}
	return &role, nil
}

func (r *RBACRepository) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		log.Printf("GetRoleByName: query failed: %v", err)
		return nil, err
	}
	return &role, nil
}

func (r *RBACRepository) UpdateRole(role *models.Role) error {
	if err := r.db.Save(role).Error; err != nil {
		log.Printf("UpdateRole: save failed: %v", err)
		return err
	}
	return nil
}

func (r *RBACRepository) DeleteRole(id int) error {
	if err := r.db.Model(&models.Role{}).Where("id = ?", id).Update("active", 0).Error; err != nil {
		log.Printf("DeleteRole: update failed: %v", err)
		return err
	}
	return nil
}

// User role operations
func (r *RBACRepository) AssignRoleToUser(userID, roleID int) error {
	userRole := &models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	if err := r.db.Create(userRole).Error; err != nil {
		log.Printf("AssignRoleToUser: create failed: %v", err)
		return err
	}
	return nil
}

func (r *RBACRepository) RemoveRoleFromUser(userID, roleID int) error {
	if err := r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error; err != nil {
		log.Printf("RemoveRoleFromUser: delete failed: %v", err)
		return err
	}
	return nil
}

func (r *RBACRepository) GetUserRoles(userID int) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.active = 1", userID).
		Find(&roles).Error
	if err != nil {
		log.Printf("GetUserRoles: query failed: %v", err)
	}
	return roles, err
}

func (r *RBACRepository) GetUsersWithRole(roleName string) ([]models.User, error) {
	var users []models.User
	err := r.db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("roles.name = ? AND users.active = 1", roleName).
		Find(&users).Error
	if err != nil {
		log.Printf("GetUsersWithRole: query failed: %v", err)
	}
	return users, err
}

// Permission operations
func (r *RBACRepository) GetUserPermissions(userID int) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct().
		Find(&permissions).Error
	if err != nil {
		log.Printf("GetUserPermissions: query failed: %v", err)
	}
	return permissions, err
}

func (r *RBACRepository) UserHasPermission(userID int, resource, action string) (bool, error) {
	var count int64
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.resource = ? AND permissions.action = ?", userID, resource, action).
		Count(&count).Error

	if err != nil {
		log.Printf("UserHasPermission: query failed: %v", err)
		return false, err
	}
	return count > 0, nil
}

func (r *RBACRepository) UserHasRole(userID int, roleName string) (bool, error) {
	var count int64
	err := r.db.Table("user_roles").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.name = ? AND roles.active = 1", userID, roleName).
		Count(&count).Error

	if err != nil {
		log.Printf("UserHasRole: query failed: %v", err)
		return false, err
	}
	return count > 0, nil
}

// Role permission operations
func (r *RBACRepository) AssignPermissionToRole(roleID, permissionID int) error {
	rolePermission := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	if err := r.db.Create(rolePermission).Error; err != nil {
		log.Printf("AssignPermissionToRole: create failed: %v", err)
		return err
	}
	return nil
}

func (r *RBACRepository) GetRolePermissions(roleID int) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		log.Printf("GetRolePermissions: query failed: %v", err)
	}
	return permissions, err
}

// Initialize default admin user
func (r *RBACRepository) EnsureAdminUser(userEmail string) error {
	// Get or create admin role
	adminRole, err := r.GetRoleByName("admin")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("EnsureAdminUser: admin role not found")
			return errors.New("admin role not found - please run migrations first")
		}
		log.Printf("EnsureAdminUser: role lookup failed: %v", err)
		return err
	}

	// Find user by email
	var user models.User
	err = r.db.Where("email = ?", userEmail).First(&user).Error
	if err != nil {
		log.Printf("EnsureAdminUser: user lookup failed: %v", err)
		return err
	}

	// Check if user already has admin role
	hasAdminRole, err := r.UserHasRole(user.ID, "admin")
	if err != nil {
		log.Printf("EnsureAdminUser: role check failed: %v", err)
		return err
	}

	if !hasAdminRole {
		if err := r.AssignRoleToUser(user.ID, adminRole.ID); err != nil {
			log.Printf("EnsureAdminUser: assign role failed: %v", err)
			return err
		}
		return nil
	}

	return nil
}
