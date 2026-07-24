# Zyra Benchmark Suite

Automated performance benchmarks measuring:
1. Process Memory (RAM) Footprint
2. Go Action RPC Throughput & Latency
3. Embedded Goja SSR Throughput & Latency
4. Fullstack HTTP Router & Security Middleware Overhead

## How to Run

```bash
go test -v ./benchmarks/...
```
For detailed results, see [results.md](./results.md).
