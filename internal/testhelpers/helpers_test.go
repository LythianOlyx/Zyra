package testhelpers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/testhelpers"
)

func TestTestHelpers(t *testing.T) {
	cfg := testhelpers.MockConfig()
	if cfg.Env != "test" {
		t.Errorf("expected env 'test', got %s", cfg.Env)
	}

	payload := map[string]string{"name": "zyra"}
	req, err := testhelpers.MakeJSONRequest(http.MethodPost, "/api/test", payload)
	if err != nil {
		t.Fatalf("failed to make JSON request: %v", err)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json")
	}

	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	rec.Write([]byte(`{"status":"ok","name":"zyra"}`))

	testhelpers.AssertStatusCode(t, rec, http.StatusOK)
	testhelpers.AssertJSONField(t, rec.Body.Bytes(), "status", "ok")
}
