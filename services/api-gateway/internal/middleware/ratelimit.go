package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	defaultRateLimit     = 100
	defaultWindowSeconds = 60
	evictionInterval     = 5 * time.Minute
)

type rateLimitEntry struct {
	count     int
	windowEnd time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	limit   int
	window  time.Duration
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   requestsPerMinute,
		window:  time.Duration(defaultWindowSeconds) * time.Second,
	}
	go rl.runEviction()
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	entry, ok := rl.entries[key]
	if !ok || now.After(entry.windowEnd) {
		rl.entries[key] = &rateLimitEntry{count: 1, windowEnd: now.Add(rl.window)}
		return true
	}
	if entry.count >= rl.limit {
		return false
	}
	entry.count++
	return true
}

func (rl *RateLimiter) runEviction() {
	ticker := time.NewTicker(evictionInterval)
	defer ticker.Stop()
	for range ticker.C {
		rl.evict()
	}
}

func (rl *RateLimiter) evict() {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for key, entry := range rl.entries {
		if now.After(entry.windowEnd) {
			delete(rl.entries, key)
		}
	}
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			if !limiter.Allow(ip) {
				http.Error(w, "troppe richieste", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
