package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/internal/action"
	"github.com/zyra-framework/zyra/internal/router"
	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestServer_ActionDispatching(t *testing.T) {
	actions := action.NewRegistry(false)
	actions.Register("test", "Ping", func(ctx context.Context, payload []byte) (interface{}, error) {
		return map[string]string{"pong": "true"}, nil
	})

	r := router.NewRouter()
	cfg := zyra.DefaultConfig()

	srv := New(Options{
		Config:  cfg,
		Router:  r,
		Actions: actions,
	})

	req := httptest.NewRequest(http.MethodPost, "/_zyra/action/test/Ping", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", rec.Code)
	}
}

func TestServer_PageRouting(t *testing.T) {
	r := router.NewRouter()
	err := r.RegisterRoute("/about", router.RenderModeSSG, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<h1>About Zyra</h1>"))
	})
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	cfg := zyra.DefaultConfig()
	srv := New(Options{
		Config: cfg,
		Router: r,
	})

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", rec.Code)
	}
	if rec.Body.String() != "<h1>About Zyra</h1>" {
		t.Errorf("Unexpected body: %s", rec.Body.String())
	}
}
