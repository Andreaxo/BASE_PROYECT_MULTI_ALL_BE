package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter holds a rate limiter and its last-seen time for a single IP.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// loginRateLimiter is the global registry of per-IP rate limiters for the login endpoint.
var loginRateLimiter = &rateLimiterStore{
	limiters: make(map[string]*ipLimiter),
	// Allow 5 login attempts per minute per IP (1 token every 12s, burst of 5).
	ratePerSec: rate.Every(12 * time.Second),
	burst:      5,
}

// rateLimiterStore manages per-IP rate limiters with periodic cleanup.
type rateLimiterStore struct {
	mu         sync.Mutex
	limiters   map[string]*ipLimiter
	ratePerSec rate.Limit
	burst      int
}

func (s *rateLimiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.limiters[ip]
	if !exists {
		lim := rate.NewLimiter(s.ratePerSec, s.burst)
		s.limiters[ip] = &ipLimiter{limiter: lim, lastSeen: time.Now()}
		return lim
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

// cleanupOldEntries removes IP entries not seen for more than 10 minutes.
// Call this in a background goroutine.
func (s *rateLimiterStore) cleanupOldEntries() {
	for {
		time.Sleep(5 * time.Minute)
		s.mu.Lock()
		for ip, entry := range s.limiters {
			if time.Since(entry.lastSeen) > 10*time.Minute {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

func init() {
	// Start the cleanup goroutine when the package is loaded
	go loginRateLimiter.cleanupOldEntries()
}

// LoginRateLimit is a Gin middleware that limits login attempts to 5 per minute per IP.
// Exceeding the limit returns HTTP 429 Too Many Requests with a Retry-After header.
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := loginRateLimiter.getLimiter(ip)
		if !limiter.Allow() {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Demasiados intentos de inicio de sesión. Por favor, espera un momento antes de intentarlo de nuevo.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
