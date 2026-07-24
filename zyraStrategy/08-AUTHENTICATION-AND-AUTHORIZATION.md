# 08 — Authentication & Authorization

## Kenapa ini modul asli, bukan sekadar template

Di kebanyakan framework Go, auth adalah "silakan pilih library sendiri dan wire manual". Di Zyra, auth adalah **modul resmi** (`internal/auth`, diekspos lewat `pkg/zyra`) yang teruji, di-maintain, dan terintegrasi langsung dengan router, Go Actions, dan hook React — sama seriusnya dengan Devise (Rails) atau Laravel Breeze/Fortify.

## Fitur yang wajib ada di v1

1. **Email + Password** dengan hashing Argon2id (default) / bcrypt (opsi migrasi legacy).
2. **Session-based auth** (default, cocok untuk SSR/SSG) dan **JWT** (opsi, cocok untuk kasus API-only/mobile client) — dipilih lewat config, API pemakaian di kode tetap sama.
3. **OAuth2**: Google & GitHub sebagai provider resmi bawaan (adapter generik supaya provider lain mudah ditambah lewat plugin).
4. **Email verification flow** (kirim email verifikasi otomatis lewat `zyra.Mail`, token expiring).
5. **Password reset flow** (token sekali pakai, expiring, invalidasi token lama saat reset berhasil).
6. **RBAC (Role-Based Access Control)**: role & permission sebagai konsep bawaan, bukan diserahkan penuh ke developer untuk desain dari nol.
7. **2FA/TOTP** (aplikasi authenticator standar) sebagai opsi yang bisa diaktifkan per-user.
8. **Brute-force protection**: lockout sementara/backoff setelah gagal login berkali-kali, terintegrasi dengan rate limiter (`05-SECURITY.md`).
9. **Magic link login** (passwordless) sebagai opsi strategi login tambahan.

## API Permukaan (contoh pemakaian)

### Setup awal (dari `zyra add auth` atau sudah bawaan di template `saas-starter`)
```ts
// zyra.config.ts
auth: {
  strategy: "session",       // atau "jwt"
  oauth: ["google", "github"],
  twoFactor: "optional",     // "off" | "optional" | "required"
}
```

### Go Action
```go
// +zyraaction
func Register(ctx context.Context, input RegisterInput) (*zyra.Session, error) {
    return zyra.Auth.Register(ctx, input)
}

// +zyraaction
// +zyraauth requireRole="admin"
func DeleteUser(ctx context.Context, userID string) error {
    return db.Users.Delete(ctx, userID)
}
```
Direktif `+zyraauth` di-scan oleh codegen yang sama seperti `+zyraaction` — kalau tidak ada direktif ini, default-nya adalah **memerlukan sesi login yang valid** (bukan public), sesuai prinsip fail-safe default di `05-SECURITY.md`. Untuk Action yang memang harus publik, developer wajib menulis `+zyrapublic` secara eksplisit — supaya "sengaja publik" terlihat jelas dibaca reviewer kode, bukan default yang bisa lolos tanpa sadar.

### React
```tsx
const { user, isAuthenticated, loading, login, logout, register } = useZyraAuth();

<ZyraProtectedRoute role="admin">
  <AdminPanel />
</ZyraProtectedRoute>
```

### Proteksi di level halaman (file-based routing)
```tsx
// pages/dashboard/index.tsx
export const requireAuth = true;
export const requireRole = "member"; // opsional
```
Router men-redirect otomatis ke halaman login (dikonfigurasi path-nya) kalau syarat tidak terpenuhi — dicek di server side untuk mode `ssr`/`ssg` (supaya tidak ada "flash of protected content"), dan di client untuk mode `csr`.

## RBAC Data Model (default, bisa di-extend)

```
users        (id, email, password_hash, ...)
roles        (id, name)
permissions  (id, name)
role_permissions (role_id, permission_id)
user_roles   (user_id, role_id)
```
Migration untuk tabel-tabel ini sudah termasuk bawaan modul `zyra add auth` — developer tidak perlu desain schema RBAC dari nol, tapi tetap bisa extend field tambahan sesuai kebutuhan.

## Keamanan Sesi

- Cookie sesi: `HttpOnly`, `Secure` (production), `SameSite=Lax`, ID sesi diregenerasi setiap login berhasil (mencegah session fixation).
- Sesi disimpan di database (default, konsisten dengan prinsip "zero infra tambahan") atau Redis (opsional, untuk deployment skala besar/multi-instance).
- Logout **selalu** menghapus sesi di server (bukan hanya menghapus cookie di client).

## Testing Auth

`zyra.Test.LoginAs(ctx, user)` — helper untuk unit/integration test yang butuh konteks user terautentikasi tanpa harus benar-benar melewati flow login HTTP lengkap setiap test.

## Yang secara sadar TIDAK masuk v1 (didokumentasikan sebagai keputusan, bukan lupa)

- SAML/Enterprise SSO — kompleksitasnya tinggi dan kebutuhannya spesifik ke segmen enterprise besar; jadikan target plugin resmi pasca-v1.0.0 (lihat `13-PLUGIN-SYSTEM.md`), bukan bagian core.
- WebAuthn/Passkey — sangat diinginkan tapi ekosistem tooling Go untuk ini masih berkembang; taruh sebagai kandidat kuat `v1.1` kalau waktu v1.0.0 sudah sangat ketat. Catat eksplisit di `15-ROADMAP-AND-MILESTONES.md` sebagai keputusan yang harus direview ulang, jangan sampai "keputusan sementara" ini terlupakan begitu saja.
