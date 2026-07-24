package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/internal/observability"
)

func TestInitLogger(t *testing.T) {
	logger, err := observability.InitLogger("development")
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
	if logger == nil {
		t.Fatalf("expected logger to be non-nil")
	}

	observability.LogAudit("user_login", "usr_123", "127.0.0.1", map[string]interface{}{
		"success": true,
	})
}

func TestTracer(t *testing.T) {
	ctx := context.Background()
	ctx, span := observability.StartSpan(ctx, "test_operation")
	defer span.End()

	if span == nil {
		t.Fatalf("expected span to be created")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	handler := observability.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected StatusOK 200, got %d", rec.Code)
	}

	// Metrics handler endpoint
	metricsReq := httptest.NewRequest(http.MethodGet, "/_zyra/metrics", nil)
	metricsRec := httptest.NewRecorder()
	observability.MetricsHandler().ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Errorf("expected metrics handler StatusOK 200, got %d", metricsRec.Code)
	}
}
