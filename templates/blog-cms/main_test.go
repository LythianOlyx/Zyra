//go:build zyratemplate

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestBlogCms_SSGAndRSS(t *testing.T) {
	cfg := zyra.DefaultConfig()
	srv, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}

	// 1. Blog index page (SSG)
	reqIndex := httptest.NewRequest(http.MethodGet, "/", nil)
	recIndex := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recIndex, reqIndex)
	if recIndex.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for blog index, got %d", recIndex.Code)
	}

	// 2. Blog post page (SSG)
	reqPost := httptest.NewRequest(http.MethodGet, "/blog/getting-started-with-zyra", nil)
	recPost := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for blog post, got %d", recPost.Code)
	}

	// 3. RSS 2.0 Feed XML endpoint
	reqRSS := httptest.NewRequest(http.MethodGet, "/rss.xml", nil)
	recRSS := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recRSS, reqRSS)
	if recRSS.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for RSS feed, got %d", recRSS.Code)
	}
	if !strings.Contains(recRSS.Body.String(), "<rss version=\"2.0\">") {
		t.Errorf("expected RSS XML body, got %s", recRSS.Body.String())
	}
}
