package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
	rate     time.Duration
	capacity int
}

type Visitor struct {
	tokens   int
	lastSeen time.Time
	mutex    sync.Mutex
}

func NewRateLimiter(rate time.Duration, capacity int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		capacity: capacity,
	}

	// Clean up old visitors every minute
	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			tokens:   rl.capacity,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
	}

	return visitor
}

func (rl *RateLimiter) allow(ip string) bool {
	visitor := rl.getVisitor(ip)
	visitor.mutex.Lock()
	defer visitor.mutex.Unlock()

	now := time.Now()
	tokensToAdd := int(now.Sub(visitor.lastSeen) / rl.rate)
	visitor.tokens += tokensToAdd

	if visitor.tokens > rl.capacity {
		visitor.tokens = rl.capacity
	}

	visitor.lastSeen = now

	if visitor.tokens > 0 {
		visitor.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mutex.Lock()

		for ip, visitor := range rl.visitors {
			if time.Since(visitor.lastSeen) > time.Hour {
				delete(rl.visitors, ip)
			}
		}

		rl.mutex.Unlock()
	}
}

func RateLimitMiddleware(rate time.Duration, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded",
				"error":   "Too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
