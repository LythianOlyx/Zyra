package action

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type TestInput struct {
	Name string `json:"name"`
}

type TestOutput struct {
	Greeting string `json:"greeting"`
}

func TestActionRegistry_Success(t *testing.T) {
	reg := NewRegistry(false)
	reg.Register("actions", "Greet", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input TestInput
		if err := json.Unmarshal(payload, &input); err != nil {
			return nil, err
		}
		return TestOutput{Greeting: "Hello " + input.Name}, nil
	})

	body, _ := json.Marshal(TestInput{Name: "Zyra"})
	req := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/Greet", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	start := time.Now()
	reg.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("RPC execution took %v, expected < 10ms", elapsed)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", rec.Code)
	}

	var resp ActionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected resp.OK == true, got error: %+v", resp.Error)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok || dataMap["greeting"] != "Hello Zyra" {
		t.Errorf("Unexpected data output: %+v", resp.Data)
	}
}

func TestActionRegistry_NotFound(t *testing.T) {
	reg := NewRegistry(false)
	req := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/Unknown", nil)
	rec := httptest.NewRecorder()

	reg.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected HTTP 404, got %d", rec.Code)
	}
}

func TestActionRegistry_MethodNotAllowed(t *testing.T) {
	reg := NewRegistry(false)
	req := httptest.NewRequest(http.MethodGet, "/_zyra/action/actions/Greet", nil)
	rec := httptest.NewRecorder()

	reg.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected HTTP 405, got %d", rec.Code)
	}
}
