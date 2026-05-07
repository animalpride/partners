package repository

import (
	"fmt"
	"log"
	"strings"

	"github.com/animalpride/partners/services/auth/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OAuthClientRepository struct {
	db *gorm.DB
}

func NewOAuthClientRepository(db *gorm.DB) *OAuthClientRepository {
	return &OAuthClientRepository{db: db}
}

func (r *OAuthClientRepository) Create(client *models.OAuthClient) error {
	if err := r.db.Create(client).Error; err != nil {
		log.Printf("CreateOAuthClient: create failed: %v", err)
		return err
	}
	return nil
}

func (r *OAuthClientRepository) GetByID(id int) (*models.OAuthClient, error) {
	var client models.OAuthClient
	if err := r.db.First(&client, id).Error; err != nil {
		log.Printf("GetOAuthClientByID: query failed: %v", err)
		return nil, err
	}
	return &client, nil
}

func (r *OAuthClientRepository) GetByClientID(clientID string) (*models.OAuthClient, error) {
	var client models.OAuthClient
	if err := r.db.Where("client_id = ?", strings.TrimSpace(clientID)).First(&client).Error; err != nil {
		log.Printf("GetOAuthClientByClientID: query failed: %v", err)
		return nil, err
	}
	return &client, nil
}

func (r *OAuthClientRepository) List() ([]models.OAuthClient, error) {
	var clients []models.OAuthClient
	if err := r.db.Order("created_at desc").Find(&clients).Error; err != nil {
		log.Printf("ListOAuthClients: query failed: %v", err)
		return nil, err
	}
	return clients, nil
}

func (r *OAuthClientRepository) ValidateSecret(client *models.OAuthClient, secret string) error {
	return bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(secret))
}

func (r *OAuthClientRepository) UpdateSecretHash(id int, secretHash string) error {
	if err := r.db.Model(&models.OAuthClient{}).Where("id = ?", id).Update("client_secret_hash", secretHash).Error; err != nil {
		log.Printf("UpdateOAuthClientSecretHash: update failed: %v", err)
		return err
	}
	return nil
}

func (r *OAuthClientRepository) UpdateActive(id int, active int) error {
	if err := r.db.Model(&models.OAuthClient{}).Where("id = ?", id).Update("active", active).Error; err != nil {
		log.Printf("UpdateOAuthClientActive: update failed: %v", err)
		return err
	}
	return nil
}

func (r *OAuthClientRepository) GetPermissions(clientID int) ([]models.Permission, error) {
	var permissions []models.Permission
	if err := r.db.Table("permissions").
		Joins("JOIN client_permissions ON permissions.id = client_permissions.permission_id").
		Where("client_permissions.oauth_client_id = ?", clientID).
		Order("permissions.resource asc, permissions.action asc").
		Find(&permissions).Error; err != nil {
		log.Printf("GetOAuthClientPermissions: query failed: %v", err)
		return nil, err
	}
	return permissions, nil
}

func (r *OAuthClientRepository) GetPermissionsByClientID(clientID string) ([]models.Permission, error) {
	client, err := r.GetByClientID(clientID)
	if err != nil {
		return nil, err
	}
	return r.GetPermissions(client.ID)
}

func (r *OAuthClientRepository) SetPermissionsByScope(id int, scopes []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("oauth_client_id = ?", id).Delete(&models.ClientPermission{}).Error; err != nil {
			return err
		}
		if len(scopes) == 0 {
			return nil
		}

		permissionIDs := make([]int, 0, len(scopes))
		seen := make(map[int]struct{}, len(scopes))
		for _, scope := range scopes {
			resource, action, err := parseScope(scope)
			if err != nil {
				return err
			}
			var permission models.Permission
			if err := tx.Where("resource = ? AND action = ?", resource, action).First(&permission).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return fmt.Errorf("unknown permission scope: %s", scope)
				}
				return err
			}
			if _, exists := seen[permission.ID]; exists {
				continue
			}
			seen[permission.ID] = struct{}{}
			permissionIDs = append(permissionIDs, permission.ID)
		}

		for _, permissionID := range permissionIDs {
			clientPermission := &models.ClientPermission{OAuthClientID: id, PermissionID: permissionID}
			if err := tx.Create(clientPermission).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func parseScope(scope string) (string, string, error) {
	normalized := strings.TrimSpace(strings.ToLower(scope))
	parts := strings.SplitN(normalized, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid permission scope: %s", scope)
	}
	return parts[0], parts[1], nil
}
