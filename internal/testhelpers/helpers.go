package testhelpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

// MockConfig returns a test configuration instance.
func MockConfig() zyra.Config {
	cfg := zyra.DefaultConfig()
	cfg.Env = "test"
	return cfg
}

// MakeJSONRequest creates a new HTTP test request with JSON body.
func MakeJSONRequest(method, target string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// AssertStatusCode checks that the response recorder matches expected HTTP status code.
func AssertStatusCode(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if rec.Code != expected {
		t.Fatalf("expected HTTP status %d, got %d. Body: %s", expected, rec.Code, rec.Body.String())
	}
}

// AssertJSONField checks a string field value in a JSON response body.
func AssertJSONField(t *testing.T, body []byte, key string, expected string) {
	t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}
	val, exists := data[key]
	if !exists {
		t.Fatalf("expected JSON key %s not found in response", key)
	}
	if strVal, ok := val.(string); ok {
		if strVal != expected {
			t.Fatalf("expected JSON key %s to be '%s', got '%s'", key, expected, strVal)
		}
	}
}
