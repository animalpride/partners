package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorLogger returns a Gin middleware that only logs requests with a status
// code >= 400.  Successful healthcheck noise (2xx/3xx) is silently dropped.
// Use in place of gin.Logger() on all services.
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		c.Next()

		status := c.Writer.Status()
		if status >= 400 {
			fmt.Printf("[GIN] %s | %3d | %12v | %s | %-7s %s\n",
				start.Format("2006/01/02 - 15:04:05"),
				status,
				time.Since(start),
				c.ClientIP(),
				c.Request.Method,
				path,
			)
		}
	}
}
