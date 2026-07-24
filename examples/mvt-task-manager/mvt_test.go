package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LythianOlyx/Zyra/examples/mvt-task-manager/actions"
	"github.com/LythianOlyx/Zyra/internal/action"
	"github.com/LythianOlyx/Zyra/internal/render"
	"github.com/LythianOlyx/Zyra/internal/render/goja"
	"github.com/LythianOlyx/Zyra/internal/router"
	"github.com/LythianOlyx/Zyra/internal/server"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

func setupTestServer(t *testing.T) *server.Server {
	_ = actions.InitDB(":memory:")

	actionsReg := action.NewRegistry(false)
	actionsReg.Register("actions", "CreateTask", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateTaskInput
		_ = unmarshalPayload(payload, &input)
		return actions.CreateTask(ctx, input)
	})
	actionsReg.Register("actions", "ListTasks", func(ctx context.Context, payload []byte) (interface{}, error) {
		return actions.ListTasks(ctx, struct{}{})
	})
	actionsReg.Register("actions", "UpdateTaskStatus", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.UpdateTaskInput
		_ = unmarshalPayload(payload, &input)
		return actions.UpdateTaskStatus(ctx, input)
	})
	actionsReg.Register("actions", "DeleteTask", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.DeleteTaskInput
		_ = unmarshalPayload(payload, &input)
		return actions.DeleteTask(ctx, input)
	})

	ssrPool, err := goja.NewPool(mockSSRBundle, goja.Options{Size: 2})
	if err != nil {
		t.Fatalf("Failed to initialize Goja pool: %v", err)
	}

	engine := render.NewEngine(render.EngineOptions{SSR: ssrPool})
	pageRouter := router.NewRouter()

	_ = engine.RegisterPage(pageRouter, render.PageConfig{FilePath: "pages/index.tsx", Mode: router.RenderModeCSR})
	_ = engine.RegisterPage(pageRouter, render.PageConfig{
		FilePath: "pages/about.tsx",
		Mode:     router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{"framework": "Zyra v1"}, 3600 * time.Second, nil
		},
	})
	_ = engine.RegisterPage(pageRouter, render.PageConfig{
		FilePath: "pages/stats.tsx",
		Mode:     router.RenderModeSSR,
		GetServerSideProps: func(r *http.Request) (interface{}, error) {
			return map[string]interface{}{"ssrEngine": "Goja Pool", "systemStatus": "HEALTHY"}, nil
		},
	})

	cfg := zyra.DefaultConfig()
	cfg.Security.CSRF.Enabled = false // Disabled for direct integration unit test calls

	return server.New(server.Options{
		Config:  cfg,
		Router:  pageRouter,
		Engine:  engine,
		Actions: actionsReg,
	})
}

func TestMVT_RPCLatencyUnder10ms(t *testing.T) {
	srv := setupTestServer(t)

	// Test Task Creation RPC
	body, _ := json.Marshal(actions.CreateTaskInput{Title: "Sub-10ms Benchmark Task", Description: "Testing Go Action RPC Speed"})
	req := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/CreateTask", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	start := time.Now()
	srv.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("RPC CreateTask latency was %v, expected < 10ms", elapsed)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", rec.Code)
	}

	var resp action.ActionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode RPC response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected ActionResponse.OK to be true, got error: %+v", resp.Error)
	}

	t.Logf("✓ CreateTask RPC Latency: %v (<10ms DoD met)", elapsed)
}

func TestMVT_ThreeRenderModes(t *testing.T) {
	srv := setupTestServer(t)

	// 1. CSR Page Test
	reqCSR := httptest.NewRequest(http.MethodGet, "/", nil)
	recCSR := httptest.NewRecorder()
	srv.ServeHTTP(recCSR, reqCSR)
	if recCSR.Code != http.StatusOK {
		t.Errorf("CSR page returned status %d, expected 200", recCSR.Code)
	}

	// 2. SSG Page Test
	reqSSG := httptest.NewRequest(http.MethodGet, "/about", nil)
	recSSG := httptest.NewRecorder()
	srv.ServeHTTP(recSSG, reqSSG)
	if recSSG.Code != http.StatusOK {
		t.Errorf("SSG page returned status %d, expected 200", recSSG.Code)
	}
	if !strings.Contains(recSSG.Body.String(), "About Zyra") {
		t.Errorf("SSG page content missing expected markup: %s", recSSG.Body.String())
	}

	// 3. SSR Page Test
	reqSSR := httptest.NewRequest(http.MethodGet, "/stats", nil)
	recSSR := httptest.NewRecorder()
	srv.ServeHTTP(recSSR, reqSSR)
	if recSSR.Code != http.StatusOK {
		t.Errorf("SSR page returned status %d, expected 200", recSSR.Code)
	}
	if !strings.Contains(recSSR.Body.String(), "Real-Time Stats") {
		t.Errorf("SSR page content missing expected markup: %s", recSSR.Body.String())
	}
}
