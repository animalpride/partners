package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the incoming Authorization bearer token or session
// cookie by forwarding a GET /permissions request to the denops-auth service.
// If the upstream returns HTTP 200, the request is allowed through.
func AuthMiddleware(authBaseURL string) gin.HandlerFunc {
	client := &http.Client{Timeout: 3 * time.Second}
	trimmedBase := strings.TrimRight(authBaseURL, "/")

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		cookieHeader := c.GetHeader("Cookie")
		if authHeader == "" && cookieHeader == "" {
			log.Printf("Authorization required: no Authorization or Cookie header provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		if authHeader != "" {
			if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				log.Printf("Invalid authorization format: %s", authHeader)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
				c.Abort()
				return
			}
		}

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			fmt.Sprintf("%s/permissions", trimmedBase),
			nil,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build auth request"})
			c.Abort()
			return
		}
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		} else if cookieHeader != "" {
			req.Header.Set("Cookie", cookieHeader)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to validate session: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to validate session"})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Session not authorized: %d", resp.StatusCode)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session not authorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
