// SPDX-License-Identifier: MIT

package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Rate limiting configuration constants
const (
	// DefaultRateLimitRequests is the maximum number of requests allowed per period
	DefaultRateLimitRequests = 100
	// DefaultRateLimitPeriod is the time window for rate limiting
	DefaultRateLimitPeriod = 1 * time.Minute
)

// RateLimiter is a thread-safe, in-memory rate limiter using the token bucket algorithm.
// It enforces rate limits per key (typically a hashed IP address) and is designed for
// web API rate limiting with privacy compliance.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex

	// Configuration
	requestsPerPeriod int
	period            time.Duration
}

// visitor tracks the token bucket state for a single key (IP hash)
type visitor struct {
	tokens     int
	lastAccess time.Time
}

var (
	oneRateLimiter    sync.Once
	globalRateLimiter *RateLimiter
	cleanupDone       = make(chan struct{})
)

// NewRateLimitMiddleware creates a rate limiter to prevent abuse
// Default: 100 requests per minute per IP address
// Uses custom token bucket algorithm for simplicity and to avoid external dependencies
func NewRateLimitMiddleware() gin.HandlerFunc {
	oneRateLimiter.Do(func() {
		// This is used by the HTTP middleware for rate limiting requests
		globalRateLimiter = NewRateLimiter(DefaultRateLimitRequests, DefaultRateLimitPeriod)
		globalRateLimiter.StartCleanup(5*time.Minute, cleanupDone)
	})

	return func(ctx *gin.Context) {
		// Use hashed IP for privacy compliance (GDPR, CCPA)
		// Same IP always produces the same hash, but original IP cannot be derived
		ipHash := "ip:" + hashIP(ctx.ClientIP())

		// Check if request is allowed
		if !globalRateLimiter.Allow(ipHash) {
			ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", globalRateLimiter.GetLimit()))
			ctx.Header("X-RateLimit-Remaining", "0")
			ctx.Header("Retry-After", fmt.Sprintf("%.0f", globalRateLimiter.GetPeriod().Seconds()))
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		// Get current token count for headers
		remaining := globalRateLimiter.GetRemaining(ipHash)

		ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", globalRateLimiter.GetLimit()))
		ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		ctx.Next()
	}
}

// NewRateLimiter creates a new RateLimiter with the specified configuration.
// If requestsPerPeriod or period are zero or negative, default values are used.
func NewRateLimiter(requestsPerPeriod int, period time.Duration) *RateLimiter {
	if requestsPerPeriod <= 0 {
		requestsPerPeriod = DefaultRateLimitRequests
	}
	if period <= 0 {
		period = DefaultRateLimitPeriod
	}

	return &RateLimiter{
		visitors:          make(map[string]*visitor),
		requestsPerPeriod: requestsPerPeriod,
		period:            period,
	}
}

// hashIP generates a SHA-256 hash of an IP address for privacy-preserving rate limiting.
// This ensures we can rate limit by IP without storing the actual IP address,
// complying with privacy regulations like GDPR and CCPA.
// The same IP address will always produce the same hash, but the original IP
// cannot be derived from the hash.
func hashIP(ip string) string {
	h := sha256.New()
	h.Write([]byte(ip))
	return hex.EncodeToString(h.Sum(nil))
}

// Allow checks if a request with the given key (typically an IP address) should be allowed.
// Returns true if the request is allowed, false if it should be rate limited.
// This is thread-safe and uses the token bucket algorithm:
// - New keys start with a full bucket of tokens
// - Tokens are consumed on each request
// - Tokens replenish over time at the configured rate
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	v, exists := rl.visitors[key]
	if !exists {
		// New visitor, add with max tokens minus one (since we're consuming one now)
		rl.visitors[key] = &visitor{
			tokens:     rl.requestsPerPeriod - 1,
			lastAccess: now,
		}
		return true
	}

	// Replenish tokens based on elapsed time
	elapsed := now.Sub(v.lastAccess)
	v.tokens += int(elapsed.Seconds() * float64(rl.requestsPerPeriod) / rl.period.Seconds())

	// Cap tokens at maximum
	if v.tokens > rl.requestsPerPeriod {
		v.tokens = rl.requestsPerPeriod
	}

	// Update last access time
	v.lastAccess = now

	// Check if we have tokens available
	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// GetRemaining returns the number of remaining requests for a given key.
// This is useful for populating rate limit headers.
// Returns 0 if the key doesn't exist (shouldn't happen in normal use).
func (rl *RateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	v, exists := rl.visitors[key]
	if !exists {
		return 0
	}

	// Ensure we don't return negative values
	if v.tokens < 0 {
		return 0
	}

	return v.tokens
}

// GetLimit returns the maximum number of requests allowed per period.
func (rl *RateLimiter) GetLimit() int {
	return rl.requestsPerPeriod
}

// GetPeriod returns the rate limit period.
func (rl *RateLimiter) GetPeriod() time.Duration {
	return rl.period
}

// Cleanup removes old entries that haven't been accessed in a while.
// This prevents memory leaks from accumulating stale rate limit entries.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// Remove entries that haven't been accessed in 2x the rate limit period
	cleanupThreshold := 2 * rl.period

	for key, v := range rl.visitors {
		if now.Sub(v.lastAccess) > cleanupThreshold {
			delete(rl.visitors, key)
		}
	}
}

// StartCleanup starts a background goroutine that periodically cleans up old entries.
// It runs cleanup every cleanupInterval and stops when the done channel is closed.
func (rl *RateLimiter) StartCleanup(cleanupInterval time.Duration, done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rl.Cleanup()
			case <-done:
				return
			}
		}
	}()
}
