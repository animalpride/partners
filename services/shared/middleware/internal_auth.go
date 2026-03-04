package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// InternalAuthMiddleware validates internal service calls using a shared token
// passed via the X-Internal-Token request header.
func InternalAuthMiddleware(token string) gin.HandlerFunc {
	trimmed := strings.TrimSpace(token)

	return func(c *gin.Context) {
		if trimmed == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "internal auth not configured"})
			c.Abort()
			return
		}

		provided := strings.TrimSpace(c.GetHeader("X-Internal-Token"))
		if provided == "" || provided != trimmed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid internal token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InternalOrAuthMiddleware allows requests that carry a valid X-Internal-Token
// header, and falls back to full session auth via AuthMiddleware for all others.
// Use this for endpoints that must be callable by both internal services and
// authenticated end-users.
func InternalOrAuthMiddleware(authBaseURL, internalToken string) gin.HandlerFunc {
	internal := strings.TrimSpace(internalToken)
	auth := AuthMiddleware(authBaseURL)

	return func(c *gin.Context) {
		provided := strings.TrimSpace(c.GetHeader("X-Internal-Token"))

		if internal != "" && provided != "" {
			if provided == internal {
				log.Printf("[InternalAuth] Internal token validated for %s %s", c.Request.Method, c.Request.URL.Path)
				c.Next()
				return
			}
			log.Printf("[InternalAuth] Token mismatch for %s %s - expected (first 8): %s, provided (first 8): %s",
				c.Request.Method,
				c.Request.URL.Path,
				MaskToken(internal),
				MaskToken(provided),
			)
		} else if provided != "" {
			log.Printf("[InternalAuth] Internal token provided but no expected token configured")
		} else {
			log.Printf("[InternalAuth] No internal token, falling back to auth for %s %s", c.Request.Method, c.Request.URL.Path)
		}

		auth(c)
	}
}
