package tailwind_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zyra-framework/zyra/internal/render/tailwind"
)

func TestResolveLatestVersion_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua == "" {
			t.Error("expected User-Agent header to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v4.3.3"}`))
	}))
	t.Cleanup(srv.Close)

	version, err := tailwind.ResolveLatestVersionFromURL(context.Background(), srv.Client(), "", srv.URL)
	if err != nil {
		t.Fatalf("ResolveLatestVersion failed: %v", err)
	}
	if version != "4.3.3" {
		t.Errorf("version = %q, want %q", version, "4.3.3")
	}
}

func TestResolveLatestVersion_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	_, err := tailwind.ResolveLatestVersionFromURL(context.Background(), srv.Client(), "", srv.URL)
	if err == nil {
		t.Fatal("expected error on rate limit")
	}
	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestResolveLatestVersion_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	_, err := tailwind.ResolveLatestVersionFromURL(context.Background(), srv.Client(), "", srv.URL)
	if err == nil {
		t.Fatal("expected error on 500 status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got: %v", err)
	}
}

func TestResolveLatestVersion_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	t.Cleanup(srv.Close)

	_, err := tailwind.ResolveLatestVersionFromURL(context.Background(), srv.Client(), "", srv.URL)
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestResolveLatestVersion_MissingTagName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name": "Release 4.3.3"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := tailwind.ResolveLatestVersionFromURL(context.Background(), srv.Client(), "", srv.URL)
	if err == nil {
		t.Fatal("expected error when tag_name is missing")
	}
	if !strings.Contains(err.Error(), "missing tag_name") {
		t.Errorf("expected missing tag_name error, got: %v", err)
	}
}

func TestResolveLatestVersion_BearerTokenHeader(t *testing.T) {
	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v4.3.3"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := tailwind.ResolveLatestVersionFromURL(context.Background(), srv.Client(), "secret-token", srv.URL)
	if err != nil {
		t.Fatalf("ResolveLatestVersion failed: %v", err)
	}
	if authHeader != "Bearer secret-token" {
		t.Errorf("Authorization header = %q, want %q", authHeader, "Bearer secret-token")
	}
}
