package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// tokenBucket is a simple token bucket rate limiter for a single key.
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func newTokenBucket(maxTokens, refillRate float64) *tokenBucket {
	return &tokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow returns true if a token is available and consumes it.
func (b *tokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(b.maxTokens, b.tokens+elapsed*b.refillRate)
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RateLimiter manages per-key token buckets.
type RateLimiter struct {
	buckets    map[string]*tokenBucket
	mu         sync.RWMutex
	maxTokens  float64
	refillRate float64
	// cleanup
	lastClean time.Time
}

func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		buckets:    map[string]*tokenBucket{},
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastClean:  time.Now(),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.RLock()
	b, ok := rl.buckets[key]
	rl.mu.RUnlock()

	if !ok {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		if b, ok = rl.buckets[key]; !ok {
			b = newTokenBucket(rl.maxTokens, rl.refillRate)
			rl.buckets[key] = b
		}
		rl.mu.Unlock()
	}

	// Periodically evict old buckets (every 10 minutes)
	if time.Since(rl.lastClean) > 10*time.Minute {
		go rl.cleanup()
	}

	return b.Allow()
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastClean = time.Now()
	// Remove buckets that are full (idle keys)
	for key, b := range rl.buckets {
		b.mu.Lock()
		full := b.tokens >= b.maxTokens
		b.mu.Unlock()
		if full {
			delete(rl.buckets, key)
		}
	}
}

// extractIP returns the client IP from a request, checking X-Forwarded-For first.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

// retryAfterSeconds computes how many seconds until the bucket refills by 1 token.
func retryAfterSeconds(refillRate float64) string {
	secs := int(1.0/refillRate) + 1
	return strconv.Itoa(secs)
}

// RateLimitMiddleware wraps a handler with IP-based rate limiting.
func RateLimitMiddleware(rl *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !rl.Allow(ip) {
			w.Header().Set("Retry-After", retryAfterSeconds(rl.refillRate))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// Preset limiters used by the handler.
var (
	// roomCreateLimiter: 5 rooms per minute per IP (5 tokens, refill 5/60 per second)
	roomCreateLimiter = NewRateLimiter(5, 5.0/60.0)
	// joinLimiter: 20 join attempts per minute per IP
	joinLimiter = NewRateLimiter(20, 20.0/60.0)
)

// RequestIDMiddleware attaches a unique X-Request-ID header to every request and response.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
