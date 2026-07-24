//go:build zyratemplate

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestApiOnly_HeadlessActionsAndBearerAuth(t *testing.T) {
	zyra.InitAuth("test-secret")
	seedAPIUser()

	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	// 1. Unauthenticated CreateTask Action fails
	createBody, _ := json.Marshal(map[string]string{"title": "Unauth Task"})
	reqCreate := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/CreateTask", bytes.NewReader(createBody))
	recCreate := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recCreate, reqCreate)
	if recCreate.Code == http.StatusOK {
		t.Fatal("expected unauthenticated CreateTask call to fail")
	}

	// 2. API Login returning session token
	loginBody, _ := json.Marshal(map[string]string{"email": "client@example.com", "password": "api-password-secret"})
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	recLogin := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recLogin, reqLogin)
	if recLogin.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 from API login, got %d", recLogin.Code)
	}

	var loginResp struct {
		OK    bool   `json:"ok"`
		Token string `json:"token"`
	}
	_ = json.NewDecoder(recLogin.Body).Decode(&loginResp)
	if !loginResp.OK || loginResp.Token == "" {
		t.Fatalf("expected valid login token, got %+v", loginResp)
	}

	// 3. Authenticated CreateTask using Bearer Token
	reqAuthCreate := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/CreateTask", bytes.NewReader(createBody))
	reqAuthCreate.Header.Set("Authorization", fmt.Sprintf("Bearer %s", loginResp.Token))
	recAuthCreate := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recAuthCreate, reqAuthCreate)
	if recAuthCreate.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for authenticated CreateTask, got %d: %s", recAuthCreate.Code, recAuthCreate.Body.String())
	}
}
