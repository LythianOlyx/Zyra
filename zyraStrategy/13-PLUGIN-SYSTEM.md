# 13 — Plugin System & Extensibility

## Kenapa Perlu Plugin System Sejak v1

Tidak semua kebutuhan bisa (atau sebaiknya) masuk core framework. Plugin system yang jelas sejak awal:
- Mencegah core membengkak dengan fitur niche (SAML SSO, provider pembayaran spesifik regional, dst).
- Memberi jalan bagi komunitas untuk berkontribusi tanpa harus lewat proses review core team untuk setiap ide.
- Jadi jalan keluar resmi untuk fitur yang "diputuskan sadar tidak masuk v1" (lihat catatan di `08-AUTHENTICATION-AND-AUTHORIZATION.md` soal SAML/WebAuthn).

## Interface Plugin (Go)

```go
type Plugin interface {
    Name() string
    OnInit(app *zyra.App) error       // saat app boot, sebelum server listen
    OnBuild(ctx *zyra.BuildContext) error // saat "zyra build", bisa inject asset/codegen tambahan
    OnRequest(next http.Handler) http.Handler // middleware chain, komposabel
    OnShutdown(ctx context.Context) error
}
```

Plugin didaftarkan eksplisit di `main.go` atau `zyra.config.ts` (tidak ada "magic auto-discovery" dari `node_modules`-like folder, supaya jelas apa yang benar-benar jalan di aplikasi — selaras prinsip "tidak ada yang tersembunyi").

## CLI untuk Plugin

```bash
zyra plugin add stripe       # plugin resmi
zyra plugin add github.com/someone/zyra-plugin-algolia   # plugin komunitas
zyra plugin list
```

## Plugin Resmi yang Disiapkan Zyra Sendiri (sejak v1, karena sangat umum dibutuhkan)

- `@zyra/stripe` — checkout, subscription, webhook handler siap pakai.
- `@zyra/resend` / `@zyra/ses` — adapter provider email (di luar SMTP dasar yang sudah built-in).
- `@zyra/sentry` — adapter error tracking, menyambungkan `zyra.Error` ke Sentry otomatis.
- `@zyra/analytics` — wrapper privacy-friendly analytics (Plausible/Cloudflare Web Analytics).

Plugin resmi ini dipisah repo/module dari core (`internal/plugin` hanya menyediakan mekanisme loader-nya) — supaya core tetap kecil dan plugin bisa dirilis/diupdate independen dari siklus rilis framework inti.

## Middleware Publik

Selain plugin, `pkg/zyra` juga mengekspos cara menambah middleware HTTP standar untuk kasus yang lebih sederhana dari full plugin:
```go
app.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // custom logic
        next.ServeHTTP(w, r)
    })
})
```

## Registry Komunitas (fase lanjut, terhubung ke `16-WEBSITE-STRATEGY.md`)

Direktori plugin komunitas ditampilkan di website resmi (bukan sekadar "cari di GitHub sendiri") — meningkatkan discoverability dan kepercayaan (misal dengan badge "terverifikasi maintainer" untuk plugin yang direview core team).
