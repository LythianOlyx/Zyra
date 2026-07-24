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

func TestLandingPage_SSGPages(t *testing.T) {
	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false

	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	// 1. Home page SSG test
	reqHome := httptest.NewRequest(http.MethodGet, "/", nil)
	recHome := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recHome, reqHome)
	if recHome.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for home page, got %d", recHome.Code)
	}

	// 2. Blog index SSG test
	reqBlog := httptest.NewRequest(http.MethodGet, "/blog", nil)
	recBlog := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recBlog, reqBlog)
	if recBlog.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for blog index, got %d", recBlog.Code)
	}

	// 3. Dynamic blog post SSG test
	reqPost := httptest.NewRequest(http.MethodGet, "/blog/introducing-zyra-v1", nil)
	recPost := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for blog post, got %d", recPost.Code)
	}

	// 4. Contact action test
	contactBody, _ := json.Marshal(map[string]string{"name": "Lead", "email": "lead@example.com", "message": "Demo inquiry"})
	reqContact := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/SubmitContact", bytes.NewReader(contactBody))
	recContact := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recContact, reqContact)
	if recContact.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for SubmitContact action, got %d", recContact.Code)
	}
}
