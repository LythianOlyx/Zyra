# 15 — Roadmap & Urutan Eksekusi Realistis

Dokumen ini menjawab pertanyaan: **"Semua fitur di `14-V1-KILLER-FEATURES-ALL-IN.md` wajib ada, tapi harus dikerjakan dengan urutan apa, mengingat waktu saya sangat terbatas (kuliah pagi–sore, kerja malam)?"**

Prinsip: fase di bawah adalah urutan **konstruksi internal**, bukan rilis bertahap ke publik. Publik hanya melihat `v1.0.0` yang sudah lengkap di akhir Fase 9.

## Fase 0 — Fondasi Kode Core

**Tujuan:** Siapkan fondasi modul internal yang teruji dengan bersih dan efisien.

- Susun modul internal: `internal/middleware/*`, `internal/observability/*`, `internal/router/page_router.go`, `internal/codegen/*`, `internal/config/*`, `internal/testhelpers/*`.
- Pastikan arsitektur berfokus murni pada React & Go tanpa dependensi Node.js atau Bun.
- Setup `cmd/zyra` sebagai **satu-satunya** entry point CLI.

**Definition of Done:** `go build ./...` sukses, unit test hijau, CLI bisa jalan (`zyra --help`).

## Fase 1 — Rendering Engine Baru

- Implementasi pool `goja` untuk SSR (lihat `03-RENDERING-ENGINE.md`).
- Implementasi Tailwind Standalone Binary Manager.
- Wiring esbuild untuk client bundle + code splitting per route.
- Implementasi 3 mode rendering (`csr`/`ssg`/`ssr`) di router.

**DoD:** halaman contoh bisa dirender di ketiga mode, terukur waktu render-nya, fallback CSR-saat-SSR-gagal teruji.

## Fase 2 — Data Layer

- Migration runner (`golang-migrate` sebagai library) + embed.
- Repository pattern dasar di atas `sqlx` untuk Postgres/MySQL/SQLite.
- Transaction helper, connection health check ke `/readyz`.

**DoD:** `zyra migrate up/down/status` jalan di 3 database utama, transaction helper punya test rollback yang jelas.

### MVT (Minimum Viable Testbed) Gate — Validasi Dogfooding Awal

Sebelum melangkah ke pembangunan Auth Module & UI Library, bangun 1 aplikasi mini percobaan (misal: *Task Manager* sederhana memakai Go Actions RPC + SQLite + 3 mode rendering + Tailwind Standalone) untuk memvalidasi fondasi RPC, router, esbuild, dan Goja pool di skenario penggunaan nyata.

**DoD MVT:** Mini-app berjalan lancar tanpa bug fondasi, latency RPC < 10ms (lokal), build single binary sukses, dan `zyra dev` HMR berjalan stabil.

## Fase 3 — Auth Module

- Session + JWT strategy, OAuth Google/GitHub, RBAC schema & middleware, email verification, reset password, 2FA/TOTP, brute-force protection.

**DoD:** skenario end-to-end "register → verifikasi email → login → akses halaman ber-role" lolos test integrasi.

## Fase 4 — DX Helpers

- `zyra.Mail`, `zyra.Storage`, `zyra.Cache`, `zyra.Jobs` (+`+zyracron`), `zyra.Paginate`, `zyra.Flags`, `zyra.Env.MustLoad`, error terstruktur.

**DoD:** setiap helper punya minimal satu contoh dipakai nyata di template `saas-starter` (bukan cuma unit test terisolasi).

## Fase 5 — Security & SEO Hardening Pass

- Terapkan seluruh checklist `05-SECURITY.md` dan `06-SEO-AND-PERFORMANCE.md` ke atas hasil Fase 1–4.
- Bangun `zyra doctor` dan `zyra audit`.

**DoD:** audit keamanan komprehensif dijalankan untuk codebase baru, hasil "PASS" di semua item OWASP Top 10, plus item baru (SSR/goja, Tailwind binary) ikut diaudit.

## Fase 6 — UI Component Library

- Buat 40+ komponen UI kit bawaan, terapkan standar aksesibilitas & test sesuai `09-UI-COMPONENT-LIBRARY.md`.

**DoD:** seluruh komponen lolos audit a11y dasar, punya test, dan terdaftar di `zyra add ui`.

## Fase 7 — CLI Templates (10 starter)

- Bangun satu-per-satu, prioritas: `blank` → `saas-starter` (paling kompleks, sekaligus jadi dogfood test seluruh fitur) → sisanya.

**DoD:** setiap template lolos `zyra audit`, punya test bawaan, dan berhasil `zyra build` di CI.

## Fase 8 — Plugin System + Fitur Diferensiasi

- Plugin interface + 4 plugin resmi (`stripe`, `resend`/`ses`, `sentry`, `analytics`).
- AI-Ready Error Overlay.

**DoD:** minimal satu plugin komunitas contoh (dummy) berhasil dipasang lewat `zyra plugin add` tanpa mengubah core.

## Fase 9 — Dogfooding, Benchmark, Rilis

- Bangun satu aplikasi nyata (bukan contoh mainan) memakai template `saas-starter` untuk memvalidasi seluruh klaim.
- Jalankan benchmark RAM/latency dengan skrip reproducible di `benchmarks/`, publikasikan hasil apa adanya.
- Ulangi security audit final (`zyra audit --production`).
- Commit seluruh project, buat tag release `v1.0.0`, dan push ke GitHub repository (`https://github.com/LythianOlyx/Zyra.git`).

## Fase 10 — Website (Sengaja Terakhir, Sesuai Prioritasmu)

- Bangun sesuai `16-WEBSITE-STRATEGY.md`. Boleh mulai dicicil paralel di late Fase 7–9 kalau ada waktu luang, tapi tidak menghambat rilis `v1.0.0` framework itu sendiri.

## Cara Realistis Mengerjakan Ini dengan Waktu Terbatas

- Setiap fase di atas punya **write-scope yang jelas dan terpisah** — ini persis kondisi ideal untuk delegasi ke AI coding agent per-fase (atau bahkan per-modul dalam satu fase), karena tidak saling tumpang tindih file yang diubah.
- Kerjakan dalam sesi pendek 30–90 menit per hari kerja (malam setelah kerja) + sesi lebih panjang di akhir pekan untuk fase yang butuh fokus (misal Fase 1 rendering engine).
- Review hasil AI agent tetap wajib dilakukan manual sebelum lanjut fase berikutnya — jangan menumpuk fase tanpa verifikasi, karena bug di fondasi (Fase 0–2) akan menggelinding jadi masalah besar di fase UI/template.
- Instruksikan AI agent untuk membaca seluruh folder `zyraStrategy/` tiap kali membuka sesi baru untuk fase tertentu — cukup sebutkan "lanjutkan ke Fase N sesuai `15-ROADMAP-AND-MILESTONES.md`".
