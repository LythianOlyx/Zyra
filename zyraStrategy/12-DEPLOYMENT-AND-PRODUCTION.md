# 12 — Deployment & Production Readiness

## Single Binary, Sungguhan Satu File

`zyra build` menghasilkan satu executable (`dist/zyra-server`) yang sudah meng-embed:
- Bundle client (JS/CSS hasil esbuild + Tailwind).
- Bundle SSR (untuk halaman bermode `ssr`, dijalankan lewat pool `goja`).
- File migration SQL.
- Template email bawaan.
- Aset publik.

Server produksi cukup punya binary ini + environment variable (`DATABASE_URL`, secret, dst). Tidak ada folder `node_modules`, tidak ada instalasi Node/Bun, tidak ada langkah build tambahan di server.

## Dockerfile Resmi (template, disediakan di setiap starter)

- Multi-stage build: stage pertama compile Go + jalankan `zyra build`, stage kedua **distroless** atau `scratch` (karena SQLite driver pure-Go dan tidak ada CGO, ini benar-benar memungkinkan) — target ukuran image final di bawah 20MB.
- Non-root user di dalam container secara default (praktik keamanan container dasar yang sering diabaikan template lain).

## Rekomendasi Platform Deploy

Dokumentasikan langkah konkret (bukan cuma "Zyra bisa dideploy di mana saja") untuk:
- **VM murah/VPS** — systemd unit file template + reverse proxy Caddy/Nginx contoh config dengan HTTPS otomatis (Caddy: cukup beberapa baris).
- **Fly.io / Railway / Render** — `fly.toml`/config setara siap pakai, karena platform ini populer di kalangan indie hacker (persona target utama Zyra).
- **Kubernetes** — manifest minimal + `/healthz`/`/readyz` sudah otomatis kompatibel dengan liveness/readiness probe K8s tanpa konfigurasi tambahan.
- **SQLite mode "zero infra"** — panduan khusus untuk deployment tanpa database server terpisah (volume persistent + strategi backup file sederhana), sangat relevan untuk MVP/side project.

## Environment & Config

- Cascading config: `.env` → `.env.local` → `.env.production`/`.env.development` sesuai `NODE_ENV`-equivalent (`ZYRA_ENV`).
- `zyra.Env.MustLoad[T]()` (lihat `04-DEVELOPER-EXPERIENCE.md`) memastikan **boot gagal cepat dengan pesan jelas** kalau ada env var wajib yang kosong — dibanding gagal random di tengah request produksi.
- `.env.example` di setiap template selalu sinkron dengan struct env yang benar-benar dipakai kode (`zyra doctor` memverifikasi ini, memperingatkan kalau ada drift).

## Observability di Production

- OpenTelemetry tracing + Prometheus metrics (`/metrics`) — reuse penuh dari v1.
- Structured logging (`zap`) dengan level yang berbeda default antara dev (verbose, human-readable) dan production (JSON terstruktur untuk log aggregator).
- Grafana dashboard template siap pakai (reuse & extend `examples/grafana-dashboard.json` dari v1) — panel untuk latency Action, SSR render time, cache hit rate, error rate.
- Health checks: `/healthz` (liveness), `/readyz` (readiness, termasuk cek koneksi database).
- Graceful shutdown: server berhenti menerima request baru, menyelesaikan request in-flight, menutup koneksi database dengan rapi saat menerima `SIGTERM` — wajib teruji, bukan asumsi.

## Backup & Disaster Recovery

- Panduan resmi backup untuk mode SQLite (file-based, snapshot berkala) dan mode Postgres/MySQL (`pg_dump`/`mysqldump` terjadwal, contoh cron job/GitHub Action).
- Dokumentasi rollback: karena deployment berupa satu binary, rollback = ganti binary versi sebelumnya + (kalau perlu) `zyra migrate down` — proses yang jauh lebih sederhana dibanding rollback deployment Node multi-service.

## Zero-Downtime Deploy

- Dokumentasi pola **rolling deploy** dua instance di belakang load balancer (binary baru naik, health check lolos, traffic dialihkan, binary lama dimatikan) — tidak butuh fitur canggih di level framework, cukup dipastikan `/readyz` benar-benar akurat merepresentasikan kesiapan instance.

## Checklist "Production Ready" Final (harus lolos semua sebelum project user disebut siap production, dicek otomatis lewat `zyra audit --production`)

1. Semua environment variable wajib ter-set.
2. TLS aktif / berada di belakang reverse proxy yang menghandle TLS.
3. `zyra audit` tanpa temuan kritis.
4. Migration database berjalan sukses di environment target.
5. Health check (`/healthz`, `/readyz`) merespons benar.
6. Observability (tracing/metrics) aktif dan terhubung ke collector yang benar.
7. Backup strategy terdokumentasi/terjadwal.
8. Build dijalankan dengan flag production (`zyra build --production`), bukan hasil `zyra dev`.
