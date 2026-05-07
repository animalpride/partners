package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware enforces double-submit CSRF protection for unsafe HTTP methods.
// It validates that the csrf_token cookie matches the X-CSRF-Token request header.
// The /login path is exempt because the login form cannot set the token before
// the user has authenticated.
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Login endpoint is exempt — no session cookie exists yet at that point.
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/oauth/token" {
			c.Next()
			return
		}

		csrfCookie, err := c.Cookie("csrf_token")
		csrfHeader := strings.TrimSpace(c.GetHeader("X-CSRF-Token"))
		if err != nil || csrfCookie == "" || csrfHeader == "" || csrfCookie != csrfHeader {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
