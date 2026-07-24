//go:build zyratemplate

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestAiChat_SendMessageAndGetHistory(t *testing.T) {
	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	// 1. Send Message Action
	msgBody, _ := json.Marshal(map[string]string{"conversationId": "test_conv", "content": "Hello AI"})
	reqMsg := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/SendMessage", bytes.NewReader(msgBody))
	recMsg := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recMsg, reqMsg)
	if recMsg.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for SendMessage, got %d", recMsg.Code)
	}

	time.Sleep(150 * time.Millisecond) // Allow mock assistant reply to generate

	// 2. Get History Action
	histBody, _ := json.Marshal(map[string]string{"conversationId": "test_conv"})
	reqHist := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/GetHistory", bytes.NewReader(histBody))
	recHist := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recHist, reqHist)
	if recHist.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GetHistory, got %d", recHist.Code)
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"data"`
	}
	_ = json.NewDecoder(recHist.Body).Decode(&resp)
	if !resp.OK || len(resp.Data) < 2 {
		t.Fatalf("expected at least 2 messages in history, got %+v", resp)
	}
}
