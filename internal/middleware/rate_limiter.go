package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"sync"
	"time"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

type clientRecord struct {
	timestamps []time.Time
}

// RateLimiter creates a sliding window rate limiter middleware.
func RateLimiter(cfg zyra.RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	maxRequests := cfg.Requests
	if maxRequests <= 0 {
		maxRequests = 100
	}

	windowDuration, err := time.ParseDuration(cfg.Window)
	if err != nil || windowDuration <= 0 {
		windowDuration = time.Minute
	}

	var mu sync.Mutex
	clients := make(map[string]*clientRecord)

	// Clean up stale client entries periodically
	go func() {
		for {
			time.Sleep(windowDuration)
			mu.Lock()
			now := time.Now()
			for ip, rec := range clients {
				cutoff := now.Add(-windowDuration)
				var valid []time.Time
				for _, t := range rec.timestamps {
					if t.After(cutoff) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(clients, ip)
				} else {
					rec.timestamps = valid
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := extractIP(r, cfg.TrustedProxies)

			mu.Lock()
			now := time.Now()
			cutoff := now.Add(-windowDuration)

			rec, exists := clients[clientIP]
			if !exists {
				rec = &clientRecord{}
				clients[clientIP] = rec
			}

			// Filter out entries outside current sliding window
			var valid []time.Time
			for _, t := range rec.timestamps {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}

			if len(valid) >= maxRequests {
				mu.Unlock()
				retryAfterSec := int(windowDuration.Seconds())
				if len(valid) > 0 {
					oldestInWindow := valid[0]
					retryAfterSec = int(oldestInWindow.Add(windowDuration).Sub(now).Seconds()) + 1
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSec))
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, `{"ok":false,"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests. Please try again later."}}`, http.StatusTooManyRequests)
				return
			}

			rec.timestamps = append(valid, now)
			remaining := maxRequests - len(rec.timestamps)
			mu.Unlock()

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request, trustedProxies []string) string {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	if isTrustedProxy(remoteIP, trustedProxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			clientIP := strings.TrimSpace(parts[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}
	return remoteIP
}

func isTrustedProxy(ip string, trusted []string) bool {
	if len(trusted) == 0 {
		return false
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, entry := range trusted {
		if entry == ip {
			return true
		}
		if _, cidr, err := net.ParseCIDR(entry); err == nil {
			if cidr.Contains(parsedIP) {
				return true
			}
		}
	}
	return false
}
