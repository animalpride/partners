package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

func newRateLimiter(r rate.Limit, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

func (r *rateLimiter) getLimiter(key string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if v, ok := r.visitors[key]; ok {
		v.lastSeen = time.Now()
		return v.limiter
	}

	limiter := rate.NewLimiter(r.rate, r.burst)
	r.visitors[key] = &visitor{limiter: limiter, lastSeen: time.Now()}
	return limiter
}

func (r *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		r.mu.Lock()
		for key, v := range r.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(r.visitors, key)
			}
		}
		r.mu.Unlock()
	}
}

func RateLimitMiddleware(requestsPerMinute int, burst int) gin.HandlerFunc {
	limiter := newRateLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), burst)
	return func(c *gin.Context) {
		key := c.ClientIP() + "|" + c.FullPath()
		if !limiter.getLimiter(key).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
