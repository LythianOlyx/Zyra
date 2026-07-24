package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/action"
	"github.com/LythianOlyx/Zyra/internal/render/goja"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
	"github.com/LythianOlyx/Zyra/pkg/zyra/app"
)

const benchmarkSSRBundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	return '<div class="page"><h1>' + (props.title || 'Benchmark') + '</h1><p>Count: ' + (props.count || 0) + '</p></div>';
}
`

// BenchmarkRPCLatency measures Go Action RPC call throughput & latency.
func BenchmarkRPCLatency(b *testing.B) {
	reg := action.NewRegistry(true)
	reg.Register("actions", "Calculate", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input struct {
			A int `json:"a"`
			B int `json:"b"`
		}
		if err := json.Unmarshal(payload, &input); err != nil {
			return nil, err
		}
		return map[string]int{"result": input.A + input.B}, nil
	})

	body, _ := json.Marshal(map[string]int{"a": 42, "b": 58})
	payloadBytes := bytes.Repeat(body, 1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/_zyra/action/actions/Calculate", bytes.NewReader(payloadBytes))
		rec := httptest.NewRecorder()
		reg.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("RPC call failed with status %d", rec.Code)
		}
	}
}

// BenchmarkSSRThroughput measures Goja pool SSR rendering performance.
func BenchmarkSSRThroughput(b *testing.B) {
	pool, err := goja.NewPool(benchmarkSSRBundle, goja.Options{
		Size: runtime.NumCPU(),
	})
	if err != nil {
		b.Fatalf("failed to create Goja SSR pool: %v", err)
	}
	defer pool.Close()

	propsJSON := `{"title":"High-Speed SSR","count":1000}`

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			html, err := pool.Render(context.Background(), "/", propsJSON)
			if err != nil || html == "" {
				b.Fatalf("SSR render failed: %v", err)
			}
		}
	})
}

// BenchmarkRouterThroughput measures fullstack HTTP router and security middleware overhead.
func BenchmarkRouterThroughput(b *testing.B) {
	cfg := zyra.Config{
		Env:  "production",
		Port: 8080,
		Security: zyra.SecurityConfig{
			CSRF:           zyra.CSRFConfig{Enabled: false},
			RateLimit:      zyra.RateLimitConfig{Enabled: true, Requests: 1000000},
			SecurityHeader: zyra.HeaderConfig{Enabled: true, HSTS: true},
		},
	}

	pageRouter := zyra.NewRouter()
	_ = pageRouter.RegisterRoute("pages/index.tsx", zyra.RenderModeSSG, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>SSG Page</body></html>"))
	})

	srv := app.NewServer(app.ServerOptions{
		Config: cfg,
		Router: pageRouter,
	})

	handler := srv.Handler()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				b.Fatalf("router request failed with status %d", rec.Code)
			}
		}
	})
}

// TestRAMFootprint verifies framework baseline memory consumption under load.
func TestRAMFootprint(t *testing.T) {
	var mBefore, mAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&mBefore)

	// Spin up server & pool
	pool, err := goja.NewPool(benchmarkSSRBundle, goja.Options{Size: 4})
	if err != nil {
		t.Fatalf("failed to init pool: %v", err)
	}
	defer pool.Close()

	// Execute 1,000 SSR iterations
	for i := 0; i < 1000; i++ {
		_, err := pool.Render(context.Background(), "/", `{"title":"RAM Test"}`)
		if err != nil {
			t.Fatalf("render error: %v", err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&mAfter)

	allocMB := float64(mAfter.Alloc) / (1024 * 1024)
	sysMB := float64(mAfter.Sys) / (1024 * 1024)

	t.Logf("📊 Zyra Framework Baseline RAM Footprint:")
	t.Logf("   - Heap Alloc: %.2f MB", allocMB)
	t.Logf("   - Total Sys Mem: %.2f MB", sysMB)

	if allocMB > 50.0 {
		t.Errorf("baseline memory footprint exceeded limit (> 50MB): %.2f MB", allocMB)
	}
}

func TestRunBenchmarkSuite(t *testing.T) {
	RunSuite()
}
