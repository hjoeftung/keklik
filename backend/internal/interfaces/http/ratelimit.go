package httpapi

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/hjoeftung/keklik/internal/apperror"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipEntry
	r        rate.Limit
	b        int
}

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
	rl := &ipRateLimiter{
		limiters: make(map[string]*ipEntry),
		r:        r,
		b:        b,
	}
	go rl.cleanup()
	return rl
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		entry = &ipEntry{limiter: rate.NewLimiter(rl.r, rl.b)}
		rl.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (rl *ipRateLimiter) cleanup() {
	for range time.Tick(3 * time.Minute) {
		rl.mu.Lock()
		for ip, e := range rl.limiters {
			if time.Since(e.lastSeen) > 3*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// rateLimitByIP wraps a handler with per-IP token-bucket rate limiting.
func rateLimitByIP(rl *ipRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			writeError(w, apperror.New(apperror.CodeTooManyRequests, "too many requests"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP returns the originating IP from X-Forwarded-For or RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.IndexByte(xff, ','); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
