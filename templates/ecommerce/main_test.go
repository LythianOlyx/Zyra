//go:build zyratemplate

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestEcommerce_StorefrontAndCart(t *testing.T) {
	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	// 1. Storefront SSG page
	reqStore := httptest.NewRequest(http.MethodGet, "/", nil)
	recStore := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recStore, reqStore)
	if recStore.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for storefront, got %d", recStore.Code)
	}

	// 2. Validate Cart Action
	cartBody, _ := json.Marshal(map[string]any{
		"items": []map[string]any{
			{"productId": "prod_1", "quantity": 1},
		},
	})
	reqCart := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/ValidateCart", bytes.NewReader(cartBody))
	recCart := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recCart, reqCart)
	if recCart.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for ValidateCart action, got %d", recCart.Code)
	}

	// 3. Webhook endpoint
	reqHook := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader([]byte("{}")))
	recHook := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recHook, reqHook)
	if recHook.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for stripe webhook, got %d", recHook.Code)
	}
}
