//go:build zyratemplate

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
	"github.com/zyra-framework/zyra/pkg/zyra/app"
)

func newTestServer(t *testing.T) *app.Server {
	t.Helper()
	zyra.InitAuth("test-secret")
	seedAdminAccount()

	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}
	return srv
}

func TestDashboardAdmin_RouteRBAC(t *testing.T) {
	srv := newTestServer(t)

	// 1. Anonymous access to admin index is 401/403
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("expected 401/403 for anonymous admin index request, got %d", rec.Code)
	}

	// 2. Login as Admin
	loginBody, _ := json.Marshal(map[string]string{"email": "admin@example.com", "password": "change-this-password-now"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 from admin login, got %d", loginRec.Code)
	}

	var sessionCookie *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == "_zyra_session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected _zyra_session cookie")
	}

	// 3. Admin access to index & users succeeds
	for _, path := range []string{"/", "/users", "/reports"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("expected HTTP 200 for path %s, got %d", path, w.Code)
		}
	}
}
