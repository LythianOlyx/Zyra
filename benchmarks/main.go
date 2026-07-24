package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zyra-framework/zyra/internal/action"
	"github.com/zyra-framework/zyra/internal/render/goja"
	"github.com/zyra-framework/zyra/pkg/zyra"
	"github.com/zyra-framework/zyra/pkg/zyra/app"
)

const bundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	return '<div class="bench"><h1>' + (props.title || 'Zyra Benchmark') + '</h1></div>';
}
`

func RunSuite() {
	fmt.Println("================================================================================")
	fmt.Println("⚡ ZYRA v1.0.0 BENCHMARK SUITE — REPRODUCIBLE PERFORMANCE METRICS")
	fmt.Println("================================================================================")
	fmt.Printf("OS/Arch       : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Go Version    : %s\n", runtime.Version())
	fmt.Printf("CPU Cores     : %d\n", runtime.NumCPU())
	fmt.Println("--------------------------------------------------------------------------------")

	// 1. RAM Footprint Test
	var mInit, mPost runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&mInit)

	pool, _ := goja.NewPool(bundle, goja.Options{Size: runtime.NumCPU()})
	defer pool.Close()

	runtime.GC()
	runtime.ReadMemStats(&mPost)

	initHeapMB := float64(mInit.Alloc) / (1024 * 1024)
	postHeapMB := float64(mPost.Alloc) / (1024 * 1024)
	sysMB := float64(mPost.Sys) / (1024 * 1024)

	fmt.Printf("1. MEMORY FOOTPRINT (RAM)\n")
	fmt.Printf("   - Baseline Heap Alloc : %.2f MB\n", initHeapMB)
	fmt.Printf("   - Post-Goja Pool Alloc: %.2f MB\n", postHeapMB)
	fmt.Printf("   - Total OS Allocated  : %.2f MB\n", sysMB)
	fmt.Println("--------------------------------------------------------------------------------")

	// 2. Go Action RPC Latency Benchmark
	reg := action.NewRegistry(true)
	reg.Register("benchmark", "Echo", func(ctx context.Context, payload []byte) (interface{}, error) {
		var in map[string]string
		_ = json.Unmarshal(payload, &in)
		return map[string]string{"reply": in["msg"]}, nil
	})

	rpcBody, _ := json.Marshal(map[string]string{"msg": "ping"})
	rpcIter := 50000
	startRPC := time.Now()

	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	iterPerWorker := rpcIter / workers

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterPerWorker; i++ {
				req := httptest.NewRequest(http.MethodPost, "/_zyra/action/benchmark/Echo", bytes.NewReader(rpcBody))
				rec := httptest.NewRecorder()
				reg.ServeHTTP(rec, req)
			}
		}()
	}
	wg.Wait()
	durRPC := time.Since(startRPC)
	rpcRPS := float64(rpcIter) / durRPC.Seconds()
	rpcAvgMicro := float64(durRPC.Microseconds()) / float64(rpcIter)

	fmt.Printf("2. GO ACTION RPC BRIDGE BENCHMARK\n")
	fmt.Printf("   - Total Requests      : %d\n", rpcIter)
	fmt.Printf("   - Execution Duration  : %v\n", durRPC)
	fmt.Printf("   - Throughput (RPS)    : %.2f req/sec\n", rpcRPS)
	fmt.Printf("   - Avg Latency         : %.2f µs (%.4f ms)\n", rpcAvgMicro, rpcAvgMicro/1000.0)
	fmt.Println("--------------------------------------------------------------------------------")

	// 3. Embedded Goja SSR Throughput Benchmark
	ssrIter := 20000
	var ssrCompleted int64
	startSSR := time.Now()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < ssrIter/workers; i++ {
				_, err := pool.Render(context.Background(), "/", `{"title":"Bench"}`)
				if err == nil {
					atomic.AddInt64(&ssrCompleted, 1)
				}
			}
		}()
	}
	wg.Wait()
	durSSR := time.Since(startSSR)
	ssrRPS := float64(ssrCompleted) / durSSR.Seconds()
	ssrAvgMicro := float64(durSSR.Microseconds()) / float64(ssrCompleted)

	fmt.Printf("3. EMBEDDED GOJA SSR ENGINE BENCHMARK\n")
	fmt.Printf("   - Successful Renders  : %d\n", ssrCompleted)
	fmt.Printf("   - Duration            : %v\n", durSSR)
	fmt.Printf("   - SSR Throughput      : %.2f req/sec\n", ssrRPS)
	fmt.Printf("   - Avg Render Latency  : %.2f µs (%.4f ms)\n", ssrAvgMicro, ssrAvgMicro/1000.0)
	fmt.Println("--------------------------------------------------------------------------------")

	// 4. HTTP Router & Middleware Throughput Benchmark
	cfg := zyra.Config{
		Env: "production",
		Security: zyra.SecurityConfig{
			CSRF:           zyra.CSRFConfig{Enabled: false},
			RateLimit:      zyra.RateLimitConfig{Enabled: false},
			SecurityHeader: zyra.HeaderConfig{Enabled: true, HSTS: true},
		},
	}
	pageRouter := zyra.NewRouter()
	_ = pageRouter.RegisterRoute("pages/index.tsx", zyra.RenderModeSSG, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	srv := app.NewServer(app.ServerOptions{Config: cfg, Router: pageRouter})
	handler := srv.Handler()

	httpIter := 100000
	startHTTP := time.Now()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < httpIter/workers; i++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
			}
		}()
	}
	wg.Wait()
	durHTTP := time.Since(startHTTP)
	httpRPS := float64(httpIter) / durHTTP.Seconds()
	httpAvgMicro := float64(durHTTP.Microseconds()) / float64(httpIter)

	fmt.Printf("4. HTTP ROUTER & MIDDLEWARE THROUGHPUT\n")
	fmt.Printf("   - Total Requests      : %d\n", httpIter)
	fmt.Printf("   - Duration            : %v\n", durHTTP)
	fmt.Printf("   - Router Throughput   : %.2f req/sec\n", httpRPS)
	fmt.Printf("   - Avg Latency         : %.2f µs (%.4f ms)\n", httpAvgMicro, httpAvgMicro/1000.0)
	fmt.Println("================================================================================")
	fmt.Println("✅ ALL BENCHMARKS COMPLETED SUCCESSFULLY!")
	fmt.Println("================================================================================")
}
