# 11 — Testing Strategy

Testing harus punya "jalur emas" (golden path) yang jelas dan didokumentasikan resmi — bukan "silakan cari sendiri cara test aplikasi Zyra", yang jadi alasan klasik tim enterprise ragu memakai framework baru.

## Lapisan Testing

### 1. Unit Test — Go Actions
```go
func TestCreateUser(t *testing.T) {
    ctx := zyra.Test.NewContext(t, zyra.Test.WithDB(t))
    user, err := CreateUser(ctx, CreateUserInput{Email: "a@b.com"})
    require.NoError(t, err)
    require.Equal(t, "a@b.com", user.Email)
}
```
`internal/testhelpers` (reuse dari v1, extend) menyediakan:
- `zyra.Test.NewContext()` — context tiruan lengkap dengan request ID, logger, dsb.
- `zyra.Test.WithDB()` — database sementara (SQLite in-memory default) + migration otomatis.
- `zyra.Test.LoginAs()` — konteks user terautentikasi (lihat `08-AUTHENTICATION-AND-AUTHORIZATION.md`).
- `zyra.Test.CallAction()` — helper memanggil Action seolah lewat HTTP (termasuk middleware auth/validasi), untuk test integrasi yang lebih menyeluruh dari sekadar unit test fungsi murni.

### 2. Component Test — React
- Setup resmi: **Vitest + React Testing Library**, sudah terkonfigurasi otomatis di setiap template hasil `zyra create` (bukan sesuatu yang harus dipasang manual oleh user).
- Konvensi: file test bertetangga dengan komponennya (`Button.tsx` + `Button.test.tsx`).
- Mocking Go Actions di sisi test React disediakan lewat helper `mockZyraAction(GetUserProfile, mockData)` — supaya component test tidak butuh backend Go benar-benar berjalan.

### 3. End-to-End Test
- Rekomendasi resmi: **Playwright**, dengan template konfigurasi siap pakai (`playwright.config.ts`) yang otomatis menjalankan `zyra build && zyra start` sebelum test jalan.
- Setiap template starter (`10-CLI-AND-PROJECT-TEMPLATES.md`) menyertakan minimal satu skenario e2e contoh (misal: `saas-starter` punya test "user bisa register lalu login").

### 4. Contract Test (Type Drift Guard)
- `zyra generate --check` — menjalankan ulang seluruh codegen (Go Actions → TypeScript, config schema, dst) dan **gagal** (exit code non-zero) kalau hasilnya berbeda dari file `.generated/` yang sudah ter-commit.
- Wajib dijalankan di CI setiap PR — mencegah kondisi "lupa run generate setelah ubah struct Go", bug klasik di framework berbasis codegen.

### 5. Load/Performance Test
- Folder `benchmarks/` dilengkapi dengan skrip **k6** siap pakai untuk load-test endpoint Action dan halaman SSR/SSG.
- Hasil benchmark (RPS, latency p50/p95/p99, RAM footprint) didokumentasikan di website sebagai bukti klaim performa — jangan cuma jadi angka di README tanpa cara reproduksi (lihat catatan di `01-VISION-AND-PHILOSOPHY.md` soal kredibilitas klaim).

### 6. Security Test
- Test suite security (`csrf_test.go`, `rate_limiter_test.go`, `security_headers_test.go`, `scanner_test.go`) dijadikan baseline wajib lolos di v1.0.0.
- Tambahkan test otomatis untuk `zyra audit` — pastikan setiap rule di checklist `05-SECURITY.md` benar-benar terdeteksi oleh test dengan skenario negatif (config yang sengaja tidak aman harus terdeteksi).

## CI Pipeline Wajib (GitHub Actions)

1. `go test -race ./...` — seluruh test Go, dengan race detector aktif.
2. `zyra generate --check` — contract test.
3. Vitest (component test frontend).
4. `govulncheck ./...` — keamanan dependency.
5. Build seluruh template starter (`zyra create --template <x>` lalu `zyra build`) — regression guard.
6. (opsional per-PR label) Playwright e2e — karena lebih lambat, boleh dijalankan penuh hanya di merge ke `main`/sebelum rilis, bukan di setiap commit kecil.

## Target Coverage

- Pertahankan standar tinggi dari v1 (≥90% untuk `internal/` yang bukan glue code CLI) — jangan turunkan standar hanya karena rewrite total.
- Coverage adalah alat bantu, bukan tujuan akhir — prioritaskan test pada logic yang berisiko tinggi (auth, security middleware, migration, query layer) di atas sekadar mengejar angka persentase.
