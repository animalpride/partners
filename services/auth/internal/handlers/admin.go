package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/models"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/repository"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	rbacRepo    *repository.RBACRepository
	userRepo    *repository.UserRepository
	invRepo     *repository.InvitationRepository
	auditRepo   *repository.AuditRepository
	refreshRepo *repository.RefreshSessionRepository
}

func NewAdminHandler(
	rbacRepo *repository.RBACRepository,
	userRepo *repository.UserRepository,
	invRepo *repository.InvitationRepository,
	auditRepo *repository.AuditRepository,
	refreshRepo *repository.RefreshSessionRepository,
) *AdminHandler {
	return &AdminHandler{
		rbacRepo:    rbacRepo,
		userRepo:    userRepo,
		invRepo:     invRepo,
		auditRepo:   auditRepo,
		refreshRepo: refreshRepo,
	}
}

// GetAdminDashboard returns admin dashboard data
func (h *AdminHandler) GetAdminDashboard(c *gin.Context) {
	// Get total users count
	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		log.Printf("GetAdminDashboard: failed to get users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users count"})
		return
	}

	// Get total roles count
	roles, err := h.rbacRepo.GetAllRoles()
	if err != nil {
		log.Printf("GetAdminDashboard: failed to get roles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles count"})
		return
	}

	// Count active users
	activeUsers := 0
	for _, user := range users {
		if user.Active == 1 {
			activeUsers++
		}
	}

	invites, err := h.invRepo.ListPendingInvitations()
	if err != nil {
		log.Printf("GetAdminDashboard: failed to get invitations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invitations"})
		return
	}
	pendingInvitations := len(invites)

	dashboardData := map[string]interface{}{
		"total_users":         len(users),
		"active_users":        activeUsers,
		"total_roles":         len(roles),
		"pending_invitations": pendingInvitations,
	}

	c.JSON(http.StatusOK, dashboardData)
}

// CreateRole creates a new role
func (h *AdminHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateRole: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		Active:      1,
	}

	if err := h.rbacRepo.CreateRole(role); err != nil {
		log.Printf("CreateRole: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetAllRoles returns all roles
func (h *AdminHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.rbacRepo.GetAllRoles()
	if err != nil {
		log.Printf("GetAllRoles: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// AssignRoleToUser assigns a role to a user
func (h *AdminHandler) AssignRoleToUser(c *gin.Context) {
	var req struct {
		UserID int `json:"user_id" binding:"required"`
		RoleID int `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("AssignRoleToUser: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.rbacRepo.AssignRoleToUser(req.UserID, req.RoleID); err != nil {
		log.Printf("AssignRoleToUser: failed to assign role %d to user %d: %v", req.RoleID, req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

// RemoveRoleFromUser removes a role from a user
func (h *AdminHandler) RemoveRoleFromUser(c *gin.Context) {
	var req struct {
		UserID int `json:"user_id" binding:"required"`
		RoleID int `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("RemoveRoleFromUser: invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.rbacRepo.RemoveRoleFromUser(req.UserID, req.RoleID); err != nil {
		log.Printf("RemoveRoleFromUser: failed to remove role %d from user %d: %v", req.RoleID, req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetUsersWithRoles returns all users with their roles
func (h *AdminHandler) GetUsersWithRoles(c *gin.Context) {
	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		log.Printf("GetUsersWithRoles: failed to get users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	var usersWithRoles []map[string]interface{}
	for _, user := range users {
		roles, err := h.rbacRepo.GetUserRoles(user.ID)
		if err != nil {
			log.Printf("GetUsersWithRoles: failed to get roles for user %d: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
			return
		}

		userWithRoles := map[string]interface{}{
			"id":                     user.ID,
			"email":                  user.Email,
			"first_name":             user.FirstName,
			"last_name":              user.LastName,
			"active":                 user.Active,
			"must_change_password":   user.MustChangePassword,
			"invitation_accepted_at": user.InvitationAcceptedAt,
			"created_at":             user.CreatedAt,
			"roles":                  roles,
		}
		usersWithRoles = append(usersWithRoles, userWithRoles)
	}

	c.JSON(http.StatusOK, usersWithRoles)
}

// ActivateUser activates or deactivates a user
func (h *AdminHandler) ActivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("ActivateUser: invalid user ID %q: %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Active *int `json:"active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ActivateUser: invalid input for user %d: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.userRepo.UpdateUserStatus(userID, *req.Active); err != nil {
		log.Printf("ActivateUser: failed to update status for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	if *req.Active == 0 {
		_ = h.refreshRepo.RevokeByUserID(userID, time.Now())
		actorID := c.GetInt("user_id")
		_ = h.auditRepo.CreateEvent("user.revoked", &actorID, &userID, nil, nil)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}

// GetUserPermissions returns user's current permissions
func (h *AdminHandler) GetUserPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	permissions, err := h.rbacRepo.GetUserPermissions(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}

	c.JSON(http.StatusOK, permissions)
}
