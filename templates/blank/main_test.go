//go:build zyratemplate

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"

	"[[.ModulePath]]/actions"
)

func TestBuildServer_GreetActionRPC(t *testing.T) {
	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false // simplifies direct RPC calls in this test

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	body, _ := json.Marshal(actions.GreetInput{Name: "Zyra"})
	req := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/Greet", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// srv.Handler() runs the full built-in security middleware stack
	// (headers/rate-limit/CSRF), matching what Start() actually serves —
	// unlike calling srv.ServeHTTP directly, which is the raw dispatcher.
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		OK   bool               `json:"ok"`
		Data actions.GreetOutput `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true response")
	}
	if resp.Data.Message == "" {
		t.Errorf("expected a non-empty greeting message")
	}
}

func TestBuildServer_IndexPageServesCSRShell(t *testing.T) {
	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}
}
