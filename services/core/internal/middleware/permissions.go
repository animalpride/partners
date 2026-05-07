package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type permissionResponse struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

func requirePermissionForPath(authBaseURL, permissionPath, resource, action string) gin.HandlerFunc {
	client := &http.Client{Timeout: 3 * time.Second}
	trimmedBase := strings.TrimRight(authBaseURL, "/")

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		cookieHeader := c.GetHeader("Cookie")
		if authHeader == "" && cookieHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			fmt.Sprintf("%s%s", trimmedBase, permissionPath),
			nil,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build auth request"})
			c.Abort()
			return
		}

		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		if cookieHeader != "" {
			req.Header.Set("Cookie", cookieHeader)
		}

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to validate session"})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session not authorized"})
			c.Abort()
			return
		}

		var permissions []permissionResponse
		if err := json.NewDecoder(resp.Body).Decode(&permissions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse permission payload"})
			c.Abort()
			return
		}

		for _, permission := range permissions {
			if permission.Resource == resource && permission.Action == action {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

func RequirePermission(authBaseURL, resource, action string) gin.HandlerFunc {
	return requirePermissionForPath(authBaseURL, "/permissions", resource, action)
}

func RequireMachinePermission(authBaseURL, resource, action string) gin.HandlerFunc {
	return requirePermissionForPath(authBaseURL, "/permissions/machine", resource, action)
}
