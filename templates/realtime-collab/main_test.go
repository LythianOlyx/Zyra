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

func TestRealtimeCollab_ActionsAndBoard(t *testing.T) {
	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	// 1. List Board Action
	reqList := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/ListBoard", nil)
	recList := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for ListBoard, got %d", recList.Code)
	}

	// 2. Add Card Action
	addBody, _ := json.Marshal(map[string]string{"title": "Test Realtime Card", "column": "todo"})
	reqAdd := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/AddCard", bytes.NewReader(addBody))
	recAdd := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recAdd, reqAdd)
	if recAdd.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for AddCard, got %d", recAdd.Code)
	}

	// 3. Heartbeat Action
	hbBody, _ := json.Marshal(map[string]string{"displayName": "Tester"})
	reqHB := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/Heartbeat", bytes.NewReader(hbBody))
	recHB := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recHB, reqHB)
	if recHB.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for Heartbeat, got %d", recHB.Code)
	}
}
