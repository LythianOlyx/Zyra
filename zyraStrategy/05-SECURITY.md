# 05 — Security Blueprint

Prinsip utama: **developer tidak boleh bisa secara tidak sengaja men-deploy aplikasi yang tidak aman.** Semua yang tercantum di sini adalah default, bukan opt-in yang harus dicari-cari di dokumentasi.

## Komponen Keamanan Utama

- `internal/middleware/csrf.go` — double-submit cookie pattern, token 32-byte via `crypto/rand`, constant-time compare.
- `internal/middleware/rate_limiter.go` — sliding window, `Retry-After` header, `TrustedProxies` whitelist.
- `internal/middleware/security_headers.go` — CSP, `X-Frame-Options`, HSTS, `Referrer-Policy`, `Permissions-Policy`.
- Checklist OWASP Top 10 — jalankan audit sebelum rilis `v1.0.0`.

## Spesifikasi Keamanan Mandatori

### A01 — Broken Access Control
- RBAC bawaan (`08-AUTHENTICATION-AND-AUTHORIZATION.md`) dengan middleware `zyra.RequireRole()`, `zyra.RequirePermission()`.
- Setiap Go Action **default menolak** akses kalau tidak ada anotasi eksplisit soal siapa yang boleh memanggilnya — opt-in ke "public", bukan opt-out dari "protected" (fail-safe default).

### A02 — Cryptographic Failures
- Password wajib di-hash dengan Argon2id (bukan bcrypt sebagai default, karena Argon2id direkomendasikan OWASP untuk deployment baru) — bcrypt tetap didukung sebagai opsi migrasi dari sistem lama.
- Cookie sesi: `HttpOnly`, `Secure` (wajib di production, auto-detect via `X-Forwarded-Proto`/TLS), `SameSite=Lax` (atau `Strict` jika tidak ada flow cross-site yang legit).
- Secret (JWT signing key, dsb) wajib dibaca dari environment variable, CLI `zyra doctor` memperingatkan jika mendeteksi secret di-hardcode di source code (regex scan sederhana untuk pattern seperti `sk_live_`, `AKIA`, dsb).

### A03 — Injection
- Query layer (`07-DATA-LAYER-AND-DATABASE.md`) **hanya** mendukung parameterized query. Tidak ada API resmi yang menerima raw string concatenation untuk SQL — kalau user butuh raw query, tetap harus lewat placeholder parameter, bukan string interpolation.
- Validasi & sanitasi input terpusat lewat tag `validate` (lihat DX doc), termasuk validasi karakter untuk nama Action/tipe (reuse pola v1 di `scanner.go`).
- Auto-escaping HTML di semua output template (default JSX sudah aman; untuk SSR shell, gunakan `html/template` yang auto-escape, bukan `text/template`).

### A04 — Insecure Design
- Rate limiting per-route bisa di-override (misal endpoint login butuh limit lebih ketat dari endpoint biasa) — dikonfigurasi lewat `zyra.config.ts`.
- Brute-force protection bawaan di modul auth: exponential backoff/lockout sementara setelah N kali gagal login dari IP/akun yang sama.

### A05 — Security Misconfiguration
- `zyra audit` CLI command wajib ada di v1: mendeteksi mode debug aktif di production, CORS `*` di production, cookie tanpa `Secure` flag di production, TLS tidak dipaksa, header security dinonaktifkan.
- Konfigurasi default **berbeda otomatis** antara `development` dan `production` (CSP lebih longgar di dev untuk devtools, sangat ketat di production) — dikontrol oleh `zyra.Env.Current()`, bukan manual toggle developer.

### A06 — Vulnerable Components
- CI template resmi (`.github/workflows/security-scan.yml`, reuse dari v1) menjalankan `govulncheck` setiap push/PR.
- Karena tidak ada lagi dependency npm/Bun di runtime, permukaan serangan supply-chain dari sisi JS **berkurang drastis** dibanding v1 — ini poin marketing sekaligus poin security nyata, sebut ini di dokumentasi.

### A07 — Identification & Authentication Failures
- Dukungan 2FA/TOTP sebagai bagian modul auth resmi (lihat `08-AUTHENTICATION-AND-AUTHORIZATION.md`) — bukan hanya password+session.
- Session fixation protection: regenerate session ID setiap kali login berhasil.

### A08 — Software & Data Integrity
- Subresource Integrity (SRI) otomatis untuk asset yang di-load, karena semua asset sudah content-hashed dari build pipeline.
- Binary rilis tetap ditandatangani GPG + checksum SHA256 (reuse `.goreleaser.yml` v1).

### A09 — Security Logging & Monitoring
- Structured audit log untuk event sensitif (login gagal/berhasil, perubahan role, akses admin) sebagai bagian dari `internal/observability`, dengan format yang mudah di-ship ke SIEM (JSON terstruktur).

### A10 — Server-Side Request Forgery (SSRF)
- Helper `zyra.Storage`/`zyra.Mail`/webhook outbound wajib punya validasi URL tujuan (blok akses ke IP privat/metadata cloud seperti `169.254.169.254`) secara default kalau helper tersebut me-nugaskan input URL dari client.

### A11 — Environment Variable Leakage & Code Boundary Protection (Anti-Next.js Leak)
- **Environment Variable Isolation:** Seluruh variabel environment tanpa awalan `PUBLIC_` (misal: `DATABASE_URL`, `STRIPE_SECRET_KEY`) **dihapus secara fisik oleh bundler `esbuild`** saat mengkompilasi bundle client. Hanya variabel berawalan `PUBLIC_` yang diizinkan masuk ke browser.
- **Server/Client Code Boundary Enforcement:** Kode server (Go Actions, database queries) terisolasi secara total di binary Go. Komponen React client tidak pernah mengeksekusi kode server secara langsung di browser, mencegah kebocoran kredensial atau logic rahasia secara otomatis.

## File Upload Security

- Validasi MIME type berdasarkan **konten file** (magic bytes), bukan cuma ekstensi/`Content-Type` header dari client (yang mudah dipalsukan).
- Limit ukuran default yang wajar, wajib eksplisit di-override kalau ingin lebih besar.
- File yang di-upload disimpan dengan nama yang di-generate ulang (bukan nama asli dari client) untuk mencegah path traversal.

## Kebijakan Dependency

- Setiap dependency Go baru yang ditambah harus dicek dulu lewat `govulncheck` dan dipertimbangkan bobot binary-nya (selaras dengan prinsip single-binary yang ringan).
- Dependency dev-only (testing, tooling) dipisah jelas dari dependency yang benar-benar terkompilasi ke binary production.

## `zyra audit` — Checklist yang Wajib Dicek Otomatis

1. Debug mode / verbose error page aktif di environment production.
2. CORS mengizinkan `*` origin di production.
3. Cookie sesi tanpa `Secure`/`HttpOnly`.
4. CSRF protection dinonaktifkan tanpa alasan eksplisit di config.
5. Rate limiting dinonaktifkan di endpoint auth.
6. Dependency dengan known CVE (`govulncheck`).
7. Environment variable wajib yang belum di-set.
8. Migration database yang belum dijalankan di target environment.
9. Header security (CSP/HSTS) dinonaktifkan.
10. Password hashing memakai algoritma lemah/deprecated (kalau ada kode custom yang override default).

Command ini harus bisa dijalankan manual (`zyra audit`) dan juga otomatis sebagai bagian dari `zyra build --production` (build gagal/warning keras kalau ada temuan level kritis).
