package middleware

import (
	"net/http"
	"strconv"

	"github.com/animalpride/partners/services/auth/internal/repository"
	"github.com/animalpride/partners/services/auth/internal/services"
	"github.com/gin-gonic/gin"
)

// RBACMiddleware creates middleware for role-based access control
func RBACMiddleware(rbacRepo *repository.RBACRepository, jwtService *services.JWTService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Prefer Authorization header for service-to-service calls; fallback to cookie for browser sessions.
		tokenString := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
				c.Abort()
				return
			}
		} else {
			cookieToken, err := c.Cookie("access_token")
			if err != nil || cookieToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
				c.Abort()
				return
			}
			tokenString = cookieToken
		}

		principalType, principalID, err := jwtService.ValidatePrincipalToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("principal_type", principalType)
		if principalType == "user" {
			userID, err := strconv.Atoi(principalID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			c.Set("user_id", userID)
		} else {
			c.Set("client_id", principalID)
		}
		c.Next()
	})
}

// RequireRole middleware ensures user has a specific role
func RequireRole(rbacRepo *repository.RBACRepository, roleName string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		hasRole, err := rbacRepo.UserHasRole(userID.(int), roleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user role"})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	})
}
