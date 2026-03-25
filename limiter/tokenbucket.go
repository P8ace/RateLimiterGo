package limiter

import (
	"net/http"
	"sync"
	"time"
)

// compile time interface implementation check
var _ RateLimiter = (*TokenBucketLimiter)(nil)

type TokenBucketLimiter struct {
	tokens     uint64    // current no.of tokens
	capacity   uint64    // max no.of tokens that can be held
	refillRate float64   // rate at which tokens can be refilled
	lastRefill time.Time // time at which the bucket was refilled
	mu         sync.Mutex
}

func NewTokenBucketLimiter(refillRate float64, burst uint64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		tokens:     burst,
		capacity:   burst,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (self *TokenBucketLimiter) Allow() bool {

	now := time.Now()
	elapsedTimeinSecs := now.Sub(self.lastRefill).Seconds()
	toAdd := uint64(elapsedTimeinSecs * self.refillRate)

	self.mu.Lock()
	defer self.mu.Unlock()
	// check if tokens need to be refilled
	if toAdd > 0 {
		self.tokens = min(self.capacity, self.tokens+toAdd)
		self.lastRefill = now
	}

	// if enough tokens are available allow the request and consume 1 token
	if self.tokens > 0 {
		self.tokens -= 1
		return true
	}
	return false
}

func (self *TokenBucketLimiter) TokenBucketMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !self.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(w, r)
	})
}
