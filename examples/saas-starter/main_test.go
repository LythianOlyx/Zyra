package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestSaaSStarterDogfooding(t *testing.T) {
	cfg := zyra.Config{
		Env:  "testing",
		Port: 8080,
		Auth: zyra.AuthConfig{
			Strategy: "session",
		},
		Security: zyra.SecurityConfig{
			CSRF:      zyra.CSRFConfig{Enabled: false}, // CSRF disabled in test for raw HTTP POST convenience
			RateLimit: zyra.RateLimitConfig{Enabled: true, Requests: 1000},
			SecurityHeader: zyra.HeaderConfig{Enabled: true, HSTS: false},
		},
	}

	zyra.InitAuth("test-secret-key-32-bytes-minimum-length")

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("failed to build server: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	client := ts.Client()

	// 1. Test SSG Landing Page
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("failed to GET /: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 2. Test Unauthenticated Access to Protected Page /dashboard
	resp, err = client.Get(ts.URL + "/dashboard")
	if err != nil {
		t.Fatalf("failed to GET /dashboard: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for /dashboard, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 3. Test Registration
	regBody, _ := json.Marshal(map[string]string{
		"email":    "dogfood@zyra.dev",
		"password": "SecurePassword123!",
	})
	resp, err = client.Post(ts.URL+"/api/auth/register", "application/json", bytes.NewReader(regBody))
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created for register, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 4. Test Login & Session Cookie
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "dogfood@zyra.dev",
		"password": "SecurePassword123!",
	})
	resp, err = client.Post(ts.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("failed to login user: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for login, got %d", resp.StatusCode)
	}

	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "_zyra_session" {
			sessionCookie = c
			break
		}
	}
	resp.Body.Close()

	if sessionCookie == nil {
		t.Fatal("expected _zyra_session cookie after login")
	}

	// 5. Test Authenticated Access to /api/auth/me using Session Cookie
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/auth/me", nil)
	req.AddCookie(sessionCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to GET /api/auth/me: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/auth/me, got %d", resp.StatusCode)
	}

	var meResp struct {
		Ok   bool `json:"ok"`
		Data struct {
			Email string `json:"email"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&meResp)
	resp.Body.Close()

	if !meResp.Ok || meResp.Data.Email != "dogfood@zyra.dev" {
		t.Fatalf("unexpected me response: %+v", meResp)
	}

	// 6. Test Go Action RPC Bridge /_zyra/action/actions.CreateCheckoutSession
	actionBody, _ := json.Marshal(map[string]any{
		"planId":     "pro_monthly",
		"successUrl": "http://localhost:8080/billing/success",
		"cancelUrl":  "http://localhost:8080/billing/cancel",
	})
	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/_zyra/action/actions/CreateCheckoutSession", bytes.NewReader(actionBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to call Action RPC: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for Action RPC, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 7. Test Health Endpoint /healthz
	resp, err = client.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("failed to GET /healthz: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /healthz, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 8. Test Prometheus Metrics Endpoint /metrics
	resp, err = client.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("failed to GET /metrics: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /metrics, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
