package middleware

import (
	"net/http"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

// SecurityHeaders returns an HTTP middleware that injects security headers.
func SecurityHeaders(cfg zyra.HeaderConfig, env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// X-Frame-Options
			frameOpt := cfg.FrameOptions
			if frameOpt == "" {
				frameOpt = "DENY"
			}
			w.Header().Set("X-Frame-Options", frameOpt)

			// X-Content-Type-Options
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Referrer-Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			// Content-Security-Policy (CSP)
			csp := cfg.ContentSecurityPolicy
			if csp == "" {
				if env == "production" {
					csp = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self';"
				} else {
					// Relaxed CSP for development (allows unsafe-eval for devtools & websocket live reload)
					csp = "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:;"
				}
			}
			w.Header().Set("Content-Security-Policy", csp)

			// Strict-Transport-Security (HSTS) - Enabled automatically in production or if configured
			if cfg.HSTS || env == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			next.ServeHTTP(w, r)
		})
	}
}
