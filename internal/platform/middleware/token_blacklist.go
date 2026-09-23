package middleware

import (
	"sync"
	"time"
)

// TokenBlacklist maintains an in-memory thread-safe blacklist of revoked JWT tokens.
type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

// GlobalTokenBlacklist is the singleton token blacklist instance.
var GlobalTokenBlacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

// Revoke adds a token string to the blacklist with its natural expiration time.
func (b *TokenBlacklist) Revoke(token string, exp time.Time) {
	if token == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[token] = exp
}

// IsRevoked checks if a token has been revoked.
func (b *TokenBlacklist) IsRevoked(token string) bool {
	if token == "" {
		return false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	exp, exists := b.tokens[token]
	if !exists {
		return false
	}
	// If the token's natural exp has passed, it is already invalid anyway
	if time.Now().After(exp) {
		return false
	}
	return true
}

// cleanupPeriodic removes expired tokens every 10 minutes to prevent memory leaks.
func (b *TokenBlacklist) cleanupPeriodic() {
	for {
		time.Sleep(10 * time.Minute)
		b.mu.Lock()
		now := time.Now()
		for t, exp := range b.tokens {
			if now.After(exp) {
				delete(b.tokens, t)
			}
		}
		b.mu.Unlock()
	}
}

func init() {
	go GlobalTokenBlacklist.cleanupPeriodic()
}
