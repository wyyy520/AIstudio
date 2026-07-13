package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Token Bucket Algorithm ---

// bucket represents a token bucket for a single client.
type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// rateLimiter manages per-IP token buckets.
type rateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	burst    int     // max burst size
	cleanupInterval time.Duration
}

// newRateLimiter creates a rate limiter with the given rate (tokens/sec) and burst.
func newRateLimiter(rate float64, burst int) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
		cleanupInterval: 5 * time.Minute,
	}
	// Periodic cleanup of stale buckets
	go rl.cleanup()
	return rl
}

// allow checks if a request from the given key is allowed.
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[key]
	now := time.Now()

	if !exists {
		rl.buckets[key] = &bucket{
			tokens:    float64(rl.burst) - 1,
			lastCheck: now,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}
	b.lastCheck = now

	// Try to consume a token
	if b.tokens >= 1.0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes stale buckets that haven't been accessed for a while.
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.buckets {
			if now.Sub(b.lastCheck) > rl.cleanupInterval*2 {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// --- Global Rate Limiter Instance ---

// Default rate limit: 100 requests per minute per IP.
const (
	defaultRate  = 100.0 / 60.0 // ~1.67 tokens/sec = 100 requests/min
	defaultBurst = 100
)

var globalLimiter = newRateLimiter(defaultRate, defaultBurst)

// RateLimitConfig allows customization of rate limiting behavior.
type RateLimitConfig struct {
	// Rate is the number of tokens added per second.
	Rate float64
	// Burst is the maximum number of tokens in the bucket.
	Burst int
	// KeyFunc extracts the client identifier (default: client IP).
	KeyFunc func(*gin.Context) string
	// Message is the response body for rate-limited requests.
	Message string
}

// DefaultRateLimitConfig returns the default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:    defaultRate,
		Burst:   defaultBurst,
		KeyFunc: nil, // uses default IP-based key
		Message: "too many requests, please try again later",
	}
}

// RateLimit returns a Gin middleware that limits request rates per client IP.
// Default limit: 100 requests per minute.
func RateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig())
}

// RateLimitWithConfig returns a rate limit middleware with the given configuration.
func RateLimitWithConfig(cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.Rate <= 0 {
		cfg.Rate = defaultRate
	}
	if cfg.Burst <= 0 {
		cfg.Burst = defaultBurst
	}
	if cfg.Message == "" {
		cfg.Message = "too many requests, please try again later"
	}

	keyFunc := cfg.KeyFunc
	if keyFunc == nil {
		keyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	limiter := newRateLimiter(cfg.Rate, cfg.Burst)

	return func(c *gin.Context) {
		key := keyFunc(c)

		// Skip rate limiting for OPTIONS preflight requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		if !limiter.allow(key) {
			// Set retry-after header
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    -1,
				"message": cfg.Message,
			})
			return
		}

		c.Next()
	}
}