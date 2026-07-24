package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zyra-framework/zyra/internal/middleware"
	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestCSRFMiddleware(t *testing.T) {
	cfg := zyra.CSRFConfig{
		Enabled:    true,
		CookieName: "_zyra_csrf",
		HeaderName: "X-CSRF-Token",
	}

	handler := middleware.CSRF(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// GET request should succeed and issue cookie
	reqGET := httptest.NewRequest(http.MethodGet, "/", nil)
	recGET := httptest.NewRecorder()
	handler.ServeHTTP(recGET, reqGET)

	if recGET.Code != http.StatusOK {
		t.Fatalf("expected GET 200, got %d", recGET.Code)
	}

	cookies := recGET.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "_zyra_csrf" {
			csrfCookie = c
			break
		}
	}
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatalf("expected CSRF cookie to be set")
	}

	// Unsafe POST without matching header should fail 403
	reqPOSTFail := httptest.NewRequest(http.MethodPost, "/", nil)
	reqPOSTFail.AddCookie(csrfCookie)
	recPOSTFail := httptest.NewRecorder()
	handler.ServeHTTP(recPOSTFail, reqPOSTFail)

	if recPOSTFail.Code != http.StatusForbidden {
		t.Errorf("expected POST without header to fail 403, got %d", recPOSTFail.Code)
	}

	// Unsafe POST with matching header should succeed 200
	reqPOSTSuccess := httptest.NewRequest(http.MethodPost, "/", nil)
	reqPOSTSuccess.AddCookie(csrfCookie)
	reqPOSTSuccess.Header.Set("X-CSRF-Token", csrfCookie.Value)
	recPOSTSuccess := httptest.NewRecorder()
	handler.ServeHTTP(recPOSTSuccess, reqPOSTSuccess)

	if recPOSTSuccess.Code != http.StatusOK {
		t.Errorf("expected POST with matching CSRF header to succeed 200, got %d", recPOSTSuccess.Code)
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	cfg := zyra.RateLimitConfig{
		Enabled:  true,
		Requests: 2,
		Window:   "1m",
	}

	handler := middleware.RateLimiter(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Req 1 - OK
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req)
	if rec1.Code != http.StatusOK {
		t.Errorf("request 1 expected 200, got %d", rec1.Code)
	}

	// Req 2 - OK
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Errorf("request 2 expected 200, got %d", rec2.Code)
	}

	// Req 3 - Rate limited 429
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("request 3 expected 429, got %d", rec3.Code)
	}
	if rec3.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	cfg := zyra.HeaderConfig{
		Enabled:      true,
		FrameOptions: "DENY",
	}

	handler := middleware.SecurityHeaders(cfg, "production")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	headers := rec.Header()
	if headers.Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options DENY, got %s", headers.Get("X-Frame-Options"))
	}
	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options nosniff, got %s", headers.Get("X-Content-Type-Options"))
	}
	if !strings.Contains(headers.Get("Strict-Transport-Security"), "max-age=") {
		t.Errorf("expected HSTS header in production mode")
	}
}
