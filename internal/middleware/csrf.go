package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

// GenerateCSRFToken generates a 32-byte secure random hex string.
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CSRF Middleware implements Double-Submit Cookie pattern.
func CSRF(cfg zyra.CSRFConfig) func(http.Handler) http.Handler {
	cookieName := cfg.CookieName
	if cookieName == "" {
		cookieName = "_zyra_csrf"
	}
	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = "X-CSRF-Token"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Retrieve or issue CSRF cookie token
			var cookieToken string
			cookie, err := r.Cookie(cookieName)
			if err == nil && cookie.Value != "" {
				cookieToken = cookie.Value
			} else {
				newToken, err := GenerateCSRFToken()
				if err != nil {
					http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
					return
				}
				cookieToken = newToken
				http.SetCookie(w, &http.Cookie{
					Name: cookieName,
					Value: cookieToken,
					Path: "/",
					// Deliberately NOT HttpOnly: the Double-Submit Cookie pattern
					// requires client-side JavaScript (runtime/client's
					// getCsrfTokenCookie()) to read this value and echo it back
					// as the X-CSRF-Token header on unsafe requests — a
					// cross-origin attacker cannot read it themselves due to the
					// browser's Same-Origin Policy, which is precisely what makes
					// this pattern work. This cookie only ever carries an
					// unguessable random token, never session/auth data, so it
					// being JS-readable is safe by design (unlike the actual
					// session cookie, which stays HttpOnly — see internal/auth).
					HttpOnly: false,
					Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
					SameSite: http.SameSiteLaxMode,
				})
			}

			// Safe methods skip validation
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
				next.ServeHTTP(w, r)
				return
			}

			// Validate token for unsafe HTTP methods (POST, PUT, DELETE, PATCH)
			headerToken := r.Header.Get(headerName)
			if headerToken == "" {
				// Also allow checking form parameter if header is absent
				headerToken = r.FormValue("csrf_token")
			}

			if headerToken == "" || subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
				http.Error(w, `{"ok":false,"error":{"code":"CSRF_INVALID","message":"Invalid or missing CSRF token"}}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IsSafeMethod checks if the HTTP method is read-only.
func IsSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
