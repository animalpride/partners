package middleware

import (
	sharedmw "github.com/animalpride/partners/services/shared/middleware"
	"github.com/gin-gonic/gin"
)

// CSRFMiddleware delegates to the shared implementation.
func CSRFMiddleware() gin.HandlerFunc {
	return sharedmw.CSRFMiddleware()
}
