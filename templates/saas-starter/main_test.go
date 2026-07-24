//go:build zyratemplate

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
	"github.com/LythianOlyx/Zyra/pkg/zyra/app"
)

func newTestServer(t *testing.T) *app.Server {
	t.Helper()
	zyra.InitAuth("test-secret")

	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false // simplifies direct RPC/API calls in this test

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}
	return srv
}

func TestSaasStarter_LandingPageServesSSGShell(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}
}

func TestSaasStarter_DashboardRequiresAuth(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401 for an unauthenticated dashboard request, got %d", rec.Code)
	}
}

func TestSaasStarter_RegisterLoginAndAccessDashboard(t *testing.T) {
	srv := newTestServer(t)

	// 1. Register
	regBody, _ := json.Marshal(map[string]string{"email": "founder@example.com", "password": "correct-horse-battery-staple"})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	regRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201 from register, got %d: %s", regRec.Code, regRec.Body.String())
	}

	// 2. Login
	loginBody, _ := json.Marshal(map[string]string{"email": "founder@example.com", "password": "correct-horse-battery-staple"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 from login, got %d: %s", loginRec.Code, loginRec.Body.String())
	}

	var sessionCookie *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == "_zyra_session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected login to set a _zyra_session cookie")
	}

	// 3. Access the auth-gated dashboard page using the session cookie.
	dashReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashReq.AddCookie(sessionCookie)
	dashRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(dashRec, dashReq)
	if dashRec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for authenticated dashboard request, got %d", dashRec.Code)
	}

	// 4. Call the authenticated CreateCheckoutSession Go Action using the
	// same session cookie, proving zyra.ResolveAuth() correctly threads
	// the current user into Action RPC context.
	checkoutBody, _ := json.Marshal(map[string]string{"plan": "pro"})
	checkoutReq := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/CreateCheckoutSession", bytes.NewReader(checkoutBody))
	checkoutReq.AddCookie(sessionCookie)
	checkoutRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(checkoutRec, checkoutReq)
	if checkoutRec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 from CreateCheckoutSession, got %d: %s", checkoutRec.Code, checkoutRec.Body.String())
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			CheckoutURL string `json:"checkoutUrl"`
		} `json:"data"`
	}
	if err := json.NewDecoder(checkoutRec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode checkout response: %v", err)
	}
	if !resp.OK || resp.Data.CheckoutURL == "" {
		t.Fatalf("expected a successful checkout response with a URL, got %+v", resp)
	}
}

func TestSaasStarter_CheckoutActionRejectsAnonymousCaller(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"plan": "pro"})
	req := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/CreateCheckoutSession", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatal("expected an anonymous CreateCheckoutSession call to fail")
	}
}
