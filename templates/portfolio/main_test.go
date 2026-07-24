//go:build zyratemplate

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

func TestPortfolio_PageAndContactAction(t *testing.T) {
	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	// 1. Home page CSR shell
	reqHome := httptest.NewRequest(http.MethodGet, "/", nil)
	recHome := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recHome, reqHome)
	if recHome.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for index page, got %d", recHome.Code)
	}

	// 2. SubmitContactForm Action
	body, _ := json.Marshal(map[string]string{
		"name":    "Visitor",
		"email":   "visitor@example.com",
		"message": "Nice portfolio!",
	})
	reqContact := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/SubmitContactForm", bytes.NewReader(body))
	recContact := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recContact, reqContact)
	if recContact.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for SubmitContactForm action, got %d", recContact.Code)
	}
}
