# 14 — Daftar Final Killer Feature v1.0.0 ("All-In")

Dokumen ini adalah **kontrak cakupan rilis**. Kalau sebuah fitur ada di daftar ini, fitur itu **wajib** ada dan berfungsi penuh sebelum tag `v1.0.0` dibuat. Tidak ada "nanti saja di v1.1" untuk item-item di bawah ini.

Yang boleh fleksibel adalah **urutan pengerjaan internal** — lihat `15-ROADMAP-AND-MILESTONES.md`. Sequencing pengerjaan ≠ pengurangan cakupan rilis publik.

## Daftar Killer Feature v1.0.0

1. **Satu binary Go, nol dependensi runtime** — tidak butuh Node/Bun/Deno terpasang di server produksi untuk kasus apapun (termasuk saat memakai mode `ssr`, karena JS engine ter-embed lewat `goja`).
2. **Tiga mode rendering per halaman** (`csr`/`ssg`/`ssr`) yang dipilih developer per file, dengan konvensi familiar ala Next.js Pages Router.
3. **Go Action RPC Bridge type-safe** — anotasi `// +zyraaction` menghasilkan client TypeScript + hook React otomatis, termasuk skema validasi form yang konsisten client/server.
4. **Realtime Streams** (`// +zyrastream`) + helper `zyra.Broadcast`.
5. **Background Jobs & Cron** (`zyra.Jobs`, `// +zyracron`) tanpa butuh infrastruktur queue tambahan untuk kasus dasar.
6. **Sistem Auth resmi lengkap**: session/JWT, OAuth (Google, GitHub), RBAC, email verification, reset password, 2FA/TOTP, brute-force protection.
7. **Security by default**: CSRF, rate limiting, security headers, validasi input ketat, fail-safe default "protected unless explicitly public" pada Go Actions, `zyra audit` CLI.
8. **SEO bawaan**: meta/OG per halaman, sitemap & robots.txt otomatis, JSON-LD helper, canonical URL.
9. **Performance bawaan**: `<ZyraImage>` teroptimasi, `<ZyraLink>` prefetching, code-splitting otomatis, performance budget di build.
10. **Data layer**: multi-DB (Postgres/MySQL/SQLite/MongoDB/Firebase/Supabase), migration system ter-embed, transaction helper, query layer typed dan aman dari SQL injection secara desain.
11. **DX helper "one function away"**: Mail, Storage/Upload, Cache, Pagination, Feature Flags, Env validation, error terstruktur — seluruh daftar di `04-DEVELOPER-EXPERIENCE.md`.
12. **AI-Ready Error Overlay** — fitur diferensiasi unik, error overlay dev-mode dengan tombol "copy prompt untuk AI".
13. **40+ komponen UI** zero-dependency, ejectable, accessible, terintegrasi form/validasi.
14. **10 starter template production-grade** yang dipilih interaktif di `zyra create`, masing-masing lolos `zyra audit` dan punya test bawaan.
15. **Testing story resmi lengkap**: unit Action, component React, e2e Playwright, contract test type-drift guard — semua terkonfigurasi otomatis di template, bukan opsional yang harus disetup manual.
16. **Observability lengkap**: OpenTelemetry, Prometheus, health checks, structured logging, Grafana dashboard template.
17. **Deployment siap produksi**: Dockerfile multi-stage kecil, panduan konkret ke platform populer, graceful shutdown, `zyra audit --production` sebagai gate rilis.
18. **Plugin system** dengan minimal plugin resmi (`stripe`, `resend`/`ses`, `sentry`, `analytics`) sudah tersedia sejak hari rilis.
19. **CLI `zyra doctor`** untuk diagnosa environment developer secara instan.
20. **Tailwind CSS penuh tanpa Node.js** lewat standalone binary manager.

## Yang Secara Sadar Diletakkan SETELAH v1.0.0 (dengan alasan eksplisit, bukan kelupaan)

| Fitur | Kenapa ditunda | Target |
|---|---|---|
| Codegen query ala `sqlc` dari file `.sql` | Query layer dasar (repository + `sqlx`) sudah cukup aman & produktif untuk v1; ini optimisasi DX tambahan, bukan blocker fungsional | v1.1 |
| SAML/Enterprise SSO | Kompleksitas tinggi, kebutuhan sangat spesifik ke enterprise besar, lebih cocok jadi plugin resmi terpisah | Plugin pasca-v1.0.0 |
| WebAuthn/Passkey | Ekosistem tooling Go untuk ini belum sematang OAuth2/TOTP | v1.1, direview lagi saat itu |
| Streaming SSR (React 18 `renderToPipeableStream`) | `renderToString` sinkron sudah cukup untuk seluruh use-case v1; streaming butuh event-loop lebih matang di goja | v1.x, setelah dasar SSR stabil |
| Playground online berbasis WASM di website | Nilai tambah besar tapi bukan blocker adopsi awal, dan bergantung pada website yang sengaja dikerjakan belakangan | Pasca website v1 |
| Visual page builder / no-code layer | Di luar misi inti "framework untuk developer", risiko melebarkan fokus terlalu jauh | Dievaluasi ulang, bukan komitmen |

Catatan penting: tabel di atas kecil dan sengaja dibuat pendek — bukti bahwa memang **hampir semua hal masuk v1**, sesuai instruksi "all-in". Item yang ditunda adalah item yang secara objektif berisiko tinggi/kompleksitas tidak proporsional kalau dipaksakan masuk v1 dengan waktu terbatas, bukan sekadar "males duluan".
