# Zyra v1.0.0 Benchmark Results & Performance Report

This document contains reproducible benchmark metrics for **Zyra v1.0.0**, measuring memory footprint (RAM), Go Action RPC bridge latency, embedded Goja SSR engine throughput, and HTTP router & security middleware overhead.

## Benchmark Environment

- **OS / Architecture**: Linux / amd64 (Bare Server / Alpine container compatible)
- **Go Version**: Go 1.23+ (`CGO_ENABLED=0`)
- **Node.js / Bun / Deno**: **0 (Zero runtime dependencies)**
- **Concurrency**: 12 CPU cores parallel execution

---

## 📊 Summary Performance Metrics

| Metric Category | Measured Benchmark Value | Framework Target / Limit | Status |
| :--- | :--- | :--- | :--- |
| **Idle Memory Footprint (Allocated Heap)** | **0.77 MB** | < 10.0 MB | ✅ PASSED |
| **Total Process OS Memory (Sys)** | **13.94 MB** | < 50.0 MB | ✅ PASSED |
| **Go Action RPC Throughput** | **102,020 req/sec** | > 10,000 req/sec | ✅ PASSED |
| **Go Action RPC Avg Latency** | **9.80 µs (0.0098 ms)** | < 1.00 ms | ✅ PASSED |
| **Embedded Goja SSR Throughput** | **232,114 req/sec** | > 15,000 req/sec | ✅ PASSED |
| **Embedded Goja SSR Avg Latency** | **4.31 µs (0.0043 ms)** | < 0.50 ms | ✅ PASSED |
| **HTTP Router & Middleware Throughput** | **161,199 req/sec** | > 20,000 req/sec | ✅ PASSED |
| **HTTP Router Avg Latency** | **6.20 µs (0.0062 ms)** | < 0.10 ms | ✅ PASSED |

---

## 1. Memory Footprint (RAM) Analysis

Zyra's zero-dependency pure-Go runtime uses `modernc.org/sqlite` and an embedded `goja` JS engine pool. Unlike Node.js or Bun applications that require 150MB - 300MB baseline memory just to start, Zyra's production binary runs with:

- **Baseline Heap**: `0.77 MB`
- **Goja SSR Pool Allocation (4 workers)**: `1.49 MB`
- **Total System RSS / Sys Mem**: `13.94 MB`

## 2. Go Action RPC Bridge Performance

Go Actions use typed binary/JSON RPC over HTTP (`/_zyra/action/{package}/{action}`) with zero dynamic reflection overhead in hot paths:

- **Total Execution Duration**: `490.09 ms` for 50,000 requests
- **Throughput**: `102,020 requests / sec`
- **Average Latency**: `9.80 microseconds`

## 3. Embedded Goja SSR Engine Throughput

Server-Side Rendering (SSR) pages render React component shells in an isolated pure-Go Goja VM pool without calling external Node/Bun sub-processes:

- **Total Execution Duration**: `86.13 ms` for 19,992 rendered pages
- **Throughput**: `232,114 page renders / sec`
- **Average Render Latency**: `4.31 microseconds`

## 4. Reproducing the Benchmarks

To run the reproducible benchmark suite locally:

```bash
# Run the benchmark suite via Go tests
go test -v ./benchmarks/...

# Or run standard Go micro-benchmarks
go test -bench=. -benchmem ./benchmarks/...
```
