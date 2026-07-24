package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/router"
)

func TestRouter_Matching(t *testing.T) {
	r := router.NewRouter()

	err := r.RegisterRoute("pages/index.tsx", router.RenderModeCSR, func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("home"))
	})
	if err != nil {
		t.Fatalf("failed to register index route: %v", err)
	}

	err = r.RegisterRoute("pages/blog/[slug].tsx", router.RenderModeSSG, func(w http.ResponseWriter, req *http.Request) {
		params := router.GetParams(req)
		w.Write([]byte("post: " + params["slug"]))
	})
	if err != nil {
		t.Fatalf("failed to register dynamic route: %v", err)
	}

	err = r.RegisterRoute("pages/docs/[...catchAll].tsx", router.RenderModeSSR, func(w http.ResponseWriter, req *http.Request) {
		params := router.GetParams(req)
		w.Write([]byte("docs: " + params["catchAll"]))
	})
	if err != nil {
		t.Fatalf("failed to register wildcard route: %v", err)
	}

	// Test 1: Home page
	recHome := httptest.NewRecorder()
	reqHome := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(recHome, reqHome)
	if recHome.Body.String() != "home" {
		t.Errorf("expected 'home', got '%s'", recHome.Body.String())
	}

	// Test 2: Dynamic param
	recBlog := httptest.NewRecorder()
	reqBlog := httptest.NewRequest(http.MethodGet, "/blog/hello-world", nil)
	r.ServeHTTP(recBlog, reqBlog)
	if recBlog.Body.String() != "post: hello-world" {
		t.Errorf("expected 'post: hello-world', got '%s'", recBlog.Body.String())
	}

	// Test 3: Catch-all wildcard
	recDocs := httptest.NewRecorder()
	reqDocs := httptest.NewRequest(http.MethodGet, "/docs/api/v1/overview", nil)
	r.ServeHTTP(recDocs, reqDocs)
	if recDocs.Body.String() != "docs: api/v1/overview" {
		t.Errorf("expected 'docs: api/v1/overview', got '%s'", recDocs.Body.String())
	}

	// Test 4: 404 Not Found
	rec404 := httptest.NewRecorder()
	req404 := httptest.NewRequest(http.MethodGet, "/unknown-page", nil)
	r.ServeHTTP(rec404, req404)
	if rec404.Code != http.StatusNotFound {
		t.Errorf("expected 404 StatusNotFound, got %d", rec404.Code)
	}
}
