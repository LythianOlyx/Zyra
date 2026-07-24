# 04 — Developer Experience: "One Function Away"

Ini dokumen paling penting untuk misi "developer suka pakai Zyra". Setiap baris di bawah adalah: **pain point nyata di framework lain → solusi satu-fungsi di Zyra**. Semua helper ini hidup di `internal/dx/` (implementasi) dan diekspos lewat `pkg/zyra` (API publik).

## Prinsip penulisan helper

- Nama fungsi harus tebak-tebakan pertama developer benar (`zyra.Mail.Send`, bukan `zyra.Communication.Dispatch.EmailMessage`).
- Setiap helper punya **default aman** yang jalan tanpa config tambahan di local dev (misal: mail dev-mode otomatis tulis ke console/log alih-alih benar-benar mengirim).
- Setiap helper punya **escape hatch** — kalau default tidak cukup, developer bisa akses layer di bawahnya (raw connection, raw request, dst) tanpa harus fork.

## Daftar Pain Point → Solusi

### 1. Duplikasi tipe data Frontend/Backend
**Masalah di framework lain:** definisikan struct/model di backend, lalu tulis ulang interface TypeScript yang sama persis di frontend. Gampang out-of-sync.
**Solusi Zyra:** Anotasi `// +zyraaction` di fungsi Go otomatis menghasilkan client TypeScript type-safe + hook React (`useZyraAction`, `useZyraQuery`, `useZyraMutation`).

### 2. Autentikasi dari nol
**Masalah:** hashing password, session/JWT, proteksi route, OAuth — biasanya harus rakit sendiri atau integrasi library pihak ketiga yang punya opini sendiri.
**Solusi:**
```go
// +zyraaction
func Login(ctx context.Context, email, password string) (*zyra.Session, error) {
    return zyra.Auth.Login(ctx, email, password)
}
```
```tsx
const { user, isAuthenticated, login, logout } = useZyraAuth();
```
Proteksi halaman: `export const requireAuth = true;` atau `export const requireRole = "admin";` di file halaman — router otomatis redirect ke halaman login kalau belum memenuhi syarat. Detail penuh di `08-AUTHENTICATION-AND-AUTHORIZATION.md`.

### 3. Upload file
**Masalah:** parsing multipart, validasi ukuran/tipe file, pilih storage (lokal/S3/R2) — biasanya 50+ baris boilerplate per endpoint.
**Solusi:**
```go
url, err := zyra.Storage.Upload(ctx, file, zyra.UploadOptions{
    Folder:      "avatars",
    MaxSizeMB:   5,
    AllowedMIME: []string{"image/png", "image/jpeg"},
})
```
Storage adapter (local disk / S3 / Cloudflare R2) dikonfigurasi sekali di `zyra.config.ts`, kode di atas tidak berubah walau backend storage diganti.

### 4. Kirim email
**Masalah:** setup SMTP/API provider, template HTML, testing tanpa benar-benar mengirim email saat development.
**Solusi:**
```go
zyra.Mail.Send(ctx, zyra.Email{
    To:       user.Email,
    Template: "welcome",
    Data:     map[string]any{"Name": user.Name},
})
```
Mode `dev`: email otomatis ditulis ke log + preview HTML di DevTools Panel browser, tidak pernah terkirim asli. Mode `production`: pakai adapter SMTP/Resend/SES yang dikonfigurasi.

### 5. Validasi form (client + server)
**Masalah:** validasi ditulis dua kali — sekali di backend (keamanan), sekali di frontend (UX) — sering pesan errornya beda/tidak konsisten.
**Solusi:** satu definisi struct tag di Go:
```go
type RegisterInput struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```
Codegen otomatis menghasilkan skema validasi TypeScript yang identik (pesan error sama persis), dipakai otomatis oleh komponen form Zyra (`<ZyraForm schema={RegisterInputSchema}>`) untuk validasi real-time di client **sebelum** request dikirim, sekaligus tetap divalidasi ulang di server (never trust client).

### 6. Real-time feature (chat, notifikasi, live dashboard)
**Masalah:** setup WebSocket/SSE manual, reconnect logic, state management di client.
**Solusi:** Penggunaan anotasi `// +zyrastream` untuk real-time stream:
```go
// +zyrastream
func LiveNotifications(ctx context.Context, userID string) (<-chan Notification, error) { ... }

zyra.Broadcast("user:123:notifications", notification) // dari action lain manapun
```
```tsx
const { data, connected } = useZyraStream(LiveNotifications, [userID]);
```

### 7. Background jobs & cron
**Masalah:** butuh infra tambahan (queue server) hanya untuk "kirim email 5 menit setelah signup" atau "jalankan setiap jam".
**Solusi:** job queue berbasis tabel di database yang sudah dipakai project (default) atau Redis (opsional, kalau sudah dikonfigurasi) — **tidak butuh infrastruktur tambahan** untuk kasus dasar:
```go
zyra.Jobs.Enqueue(ctx, "send-welcome-email", payload, zyra.JobOptions{Delay: 5 * time.Minute})

// +zyracron "0 * * * *"
func HourlyCleanup(ctx context.Context) error { ... }
```

### 8. Caching hasil query/komputasi mahal
**Masalah:** implementasi cache manual, invalidation logic berulang di banyak tempat.
**Solusi:**
```go
stats, err := zyra.Cache.Remember(ctx, "dashboard:stats", 5*time.Minute, func() (Stats, error) {
    return computeExpensiveStats(ctx)
})
```
Backend cache (in-memory LRU default / Redis opsional) transparan bagi kode di atas.

### 9. Pagination & infinite scroll
**Masalah:** logic offset/cursor, state "loading more", deduplikasi data — ditulis ulang di setiap list view.
**Solusi:**
```go
page, err := zyra.Paginate(ctx, query, zyra.PageRequest{Page: 2, PerPage: 20})
```
```tsx
const { items, loadMore, hasMore, loading } = useZyraInfiniteQuery(ListPosts, []);
```

### 10. Feature flags
**Masalah:** rilis fitur baru butuh deploy terpisah atau setup layanan pihak ketiga mahal untuk sekadar toggle sederhana.
**Solusi:**
```go
if zyra.Flags.IsEnabled(ctx, "new-checkout-flow") { ... }
```
Dikelola lewat file config atau tabel database sederhana, dashboard toggle di DevTools Panel saat development.

### 11. Environment variable & config yang rawan `nil`/panic saat runtime
**Masalah:** lupa set env var di production baru ketahuan saat runtime crash di tengah request user.
**Solusi:** validasi env **saat boot**, bukan saat dipakai:
```go
var cfg = zyra.Env.MustLoad[AppEnv]() // app gagal start dengan pesan jelas kalau ada field wajib kosong
```

### 12. Error handling tidak konsisten antara Go dan TypeScript
**Masalah:** error di backend jadi `"internal server error"` generik di frontend, developer tidak tahu apa yang salah.
**Solusi:** tipe error terstruktur (`zyra.Error{Code, Message, Status}`) yang otomatis dipetakan ke exception bertipe di client, sehingga `catch` di React tahu persis kode error apa yang terjadi, bukan cuma string bebas.

### 13. Loading state & skeleton boilerplate
**Masalah:** setiap komponen data-fetching menulis ulang `if (loading) return <Spinner />` secara manual.
**Solusi:** komponen `<ZyraBoundary fallback={<Skeleton />}>` yang otomatis sinkron dengan state hook `useZyraQuery`/`useZyraAction`.

### 14. Internationalization (i18n) & Multi-Bahasa
**Masalah:** implementasi i18n di React/Node biasanya rumit, butuh sync translation file, router locale parsing, dan SSR hydration mismatch.
**Solusi:** `zyra.I18n` — deteksi locale dari URL path (`/en/about`), cookie, atau header `Accept-Language`. Translation key didefinisikan di Go (`locales/en.json`, `id.json`), digenerate otomatis jadi TypeScript autocomplete type (`t("welcome.title")`), serta ter-inject otomatis di HTML SSR (`<html lang="id">` & meta `hreflang`).

### 15. Realtime Streams & WebSockets tanpa Boilerplate
**Masalah:** setup SSE/WebSocket, reconnection logic di client, dan room broadcasting di server biasanya butuh 100+ baris kode.
**Solusi:** anotasi `// +zyrastream` pada Go function:
```go
// +zyrastream
func WatchPrice(ctx context.Context, symbol string, stream *zyra.Stream[PriceData]) error {
    return zyra.Broadcast.Subscribe(ctx, "crypto:"+symbol, stream)
}
```
Client React hook:
```tsx
const { data, isConnected, error } = useZyraStream(WatchPrice, { symbol: "BTC" });
```
Otomatis menangani reconnection heartbeat, auth token validation, dan room subscription.

### 16. External HTTP Fetching (`zyra.HTTP.FetchJSON`)
**Masalah:** memanggil API HTTP eksternal di Go biasanya butuh 25+ baris (bikin client, setup timeout context, check status code 200, `io.ReadAll`, `json.Unmarshal`).
**Solusi:** 1 baris generic helper:
```go
res, err := zyra.HTTP.FetchJSON[WeatherResponse](ctx, "https://api.weather.com/v1", zyra.FetchOptions{
    Timeout: 3 * time.Second,
    Headers: map[string]string{"Authorization": "Bearer token"},
})
```

### 17. CSV Import / Export (`zyra.CSV`)
**Masalah:** parsing file CSV ke struct array atau export struct ke download CSV butuh 35+ baris (header parsing, row iteration, writer flush, HTTP header attachment).
**Solusi:**
```go
// Export 1 baris
err := zyra.CSV.Export(w, "users-export.csv", usersList)

// Import 1 baris
err := zyra.CSV.Import(fileHeader, &importedUsers)
```

### 18. Hashing & Cryptography (`zyra.Crypto`)
**Masalah:** argon2/bcrypt setup, salt generation, hex encoding, timing-safe string comparison butuh 20+ baris.
**Solusi:**
```go
hash, err := zyra.Crypto.HashPassword("user-secret")
match := zyra.Crypto.VerifyPassword("user-secret", hash)
token := zyra.Crypto.RandomToken(32) // secure 32-byte hex token
```

### 19. Retry & Resiliency (`zyra.Resilience.Retry`)
**Masalah:** mencoba ulang (*retry*) API yang flaky dengan *exponential backoff* dan *jitter* biasanya butuh 30+ baris loop `time.Sleep`.
**Solusi:**
```go
err := zyra.Resilience.Retry(ctx, 3, time.Second, func() error {
    return callExternalGateway(ctx)
})
```

### 20. HTML-to-PDF Generator (`zyra.PDF`)
**Masalah:** membuat file PDF dari HTML template (misal Invoice/Struk) butuh 40+ baris setup engine PDF, buffer management, dan HTTP header disposition.
**Solusi:**
```go
err := zyra.PDF.Generate(w, "templates/invoice.html", invoiceData, zyra.PDFOptions{
    Filename: "Invoice-1001.pdf",
    PaperSize: "A4",
})
```

### 21. QR Code Generator (`zyra.QRCode`)
**Masalah:** membuat gambar QR code (untuk payment/link) butuh 30+ baris matriks encoder, PNG renderer, dan HTTP writer.
**Solusi:**
```go
err := zyra.QRCode.Generate(w, "https://example.com/pay/123", zyra.QROptions{Size: 256})
```

### 22. GeoIP Location Lookup (`zyra.Geo`)
**Masalah:** mencari negara/kota dari IP request user butuh 25+ baris GeoIP database reader dan error handling.
**Solusi:**
```go
loc, err := zyra.Geo.IPToLocation(ctx, req.RemoteAddr) // loc.Country, loc.City, loc.Timezone
```

### 23. Slug Generator (`zyra.Slug`)
**Masalah:** mengubah judul artikel jadi URL slug ramah SEO (`"Halo Dunia! #1"` -> `"halo-dunia-1"`) biasanya butuh 15+ baris regex sanitasi dan Unicode handling.
**Solusi:**
```go
slug := zyra.Slug.Make("Halo Dunia! #1") // "halo-dunia-1"
```

### 24. Excel Spreadsheet Import / Export (`zyra.Excel`)
**Masalah:** buat file .xlsx dengan header, styling, dan multiple sheet atau membaca Excel yang diupload user butuh 40+ baris kode.
**Solusi:**
```go
err := zyra.Excel.Export(w, "Laporan-Penjualan.xlsx", salesData)
err := zyra.Excel.Import(fileHeader, &importedSales)
```

### 25. HTML XSS Sanitizer (`zyra.Sanitize`)
**Masalah:** membersihkan input Rich-Text Editor (WYSIWYG) dari skrip XSS jahat butuh 25+ baris setup policy parser.
**Solusi:**
```go
cleanHTML := zyra.Sanitize.HTML(userRichTextHTML)
```

### 26. High-Performance Unique IDs (`zyra.ID`)
**Masalah:** butuh ID unik aman (ULID time-sortable untuk primary key DB, UUIDv4, atau Short NanoID) tapi harus impor library terpisah.
**Solusi:**
```go
id := zyra.ID.ULID()   // 01ARZ3NDEKTSV4RRFFQ69G5FAV (default time-sortable DB PK)
uuid := zyra.ID.UUID() // f47ac10b-58cc-4372-a567-0e02b2c3d479
short := zyra.ID.Nano(10) // V1StGXR8_Z
```

### 27. Format Mata Uang (`zyra.Money`)
**Masalah:** format angka ke mata uang lokal (`15000000` -> `"Rp 15.000.000"`) butuh 15+ baris locale precision dan thousand separator logic.
**Solusi:**
```go
formatted := zyra.Money.Format(15000000, "IDR") // "Rp 15.000.000"
formattedUSD := zyra.Money.Format(1500, "USD")  // "$1,500.00"
```

### 28. Relative Human Time (`zyra.DateTime`)
**Masalah:** mengubah timestamp ke teks relatif ramah manusia (`"5 menit yang lalu"`, `"2 hari yang lalu"`) sesuai locale user.
**Solusi:**
```go
text := zyra.DateTime.HumanAgo(createdAt, "id") // "5 menit yang lalu"
```

### 29. ZIP Archive Creation & Extraction (`zyra.Archive`)
**Masalah:** membuat archive .zip dari beberapa file atau mengekstrak zip yang diupload butuh 40+ baris zip writer/reader & file walking.
**Solusi:**
```go
err := zyra.Archive.Zip(w, "files.zip", []string{"file1.pdf", "file2.png"})
err := zyra.Archive.Unzip(zipFile, "./extracted")
```

### 30. Instant Alert & Webhook Notifications (`zyra.Notification`)
**Masalah:** kirim notifikasi instant alert dev/admin saat ada error/order baru ke Telegram/Slack/Discord butuh 30+ baris HTTP request payload.
**Solusi:**
```go
err := zyra.Notification.Telegram(ctx, botToken, chatID, "🚨 Order baru masuk: #1002")
err := zyra.Notification.Slack(ctx, webhookURL, "🔥 Error di server: "+err.Error())
```

### 31. Simplified JWT Token Management (`zyra.JWT`)
**Masalah:** membuat dan verifikasi JWT token dengan custom claims, expiration, dan signature checking butuh 25+ baris.
**Solusi:**
```go
tokenStr, err := zyra.JWT.Sign(claimsMap, secretKey, 24*time.Hour)
claims, err := zyra.JWT.Verify[MyClaims](tokenStr, secretKey)
```

### 32. Security Audit Logging (`zyra.AuditLog`)
**Masalah:** mencatat aktivitas sensitif user (misal: "User A mengubah password User B") dengan IP address, timestamp, dan metadata ke DB audit trail.
**Solusi:**
```go
zyra.AuditLog.Record(ctx, "user.password_reset", targetUserID, metadataMap)
```

### 33. Functional Slice Operations (`zyra.Slice`)
**Masalah di Go standar:** melakukan `filter`, `map`, `find`, `grouping`, `unique`, `chunk`, atau `reduce` pada array/slice butuh menulis ulang `for` loop 7-10 baris dengan `if` statement berulang kali.
**Solusi 1 baris generic (Zero Library, murni Go 1.18+ Generics):**
```go
// Filter 1 baris (pengganti 7-10 baris for+if)
activeUsers := zyra.Slice.Filter(users, func(u User) bool { return u.IsActive && u.Age >= 18 })

// Map / Transform 1 baris
emails := zyra.Slice.Map(users, func(u User) string { return u.Email })

// Find 1 baris
user, found := zyra.Slice.Find(users, func(u User) bool { return u.ID == targetID })

// GroupBy 1 baris (menghasilkan map[string][]User)
byRole := zyra.Slice.GroupBy(users, func(u User) string { return u.Role })

// Unique / Deduplikasi 1 baris
uniqueIDs := zyra.Slice.Unique(categoryIDs)

// Chunking 1 baris (membagi slice besar jadi batch kecil, misal 100 per batch untuk DB Bulk Insert)
batches := zyra.Slice.Chunk(largeUserList, 100) // [][]User

// Reduce / Sum 1 baris
totalRevenue := zyra.Slice.Reduce(orders, 0, func(acc int, o Order) int { return acc + o.TotalAmount })
```

### 34. Map Utilities (`zyra.Map`)
**Masalah:** mengambil daftar keys/values dari Go `map` atau menggabungkan dua `map` butuh loop `for k, v := range m` berulang.
**Solusi:**
```go
keys := zyra.Map.Keys(userRoleMap)     // []string
values := zyra.Map.Values(userRoleMap) // []Role
merged := zyra.Map.Merge(mapA, mapB)   // map[K]V
```

### 35. Pointer Helpers (`zyra.Ptr`)
**Masalah di Go:** menulis `&"string_literal"` atau `&123` adalah sintaks **ilegal** di Go. Developer biasanya terpaksa membuat variabel sementara 3 baris hanya untuk mengisi struct field opsional ber-tipe pointer `*string`.
**Solusi 1 baris:**
```go
// Buat pointer *T instan tanpa variabel sementara
user := User{ Role: zyra.Ptr.To("admin"), Age: zyra.Ptr.To(25) }

// Ambil nilai pointer dengan fallback aman jika nil
role := zyra.Ptr.Val(user.Role, "guest") // jika user.Role == nil, kembalikan "guest"
```

### 36. Inline Conditional & Fallback (`zyra.Ternary` & `zyra.Coalesce`)
**Masalah di Go:** Go **tidak** memiliki operator ternari (`a ? b : c`). Developer harus menulis 5 baris `if / else` hanya untuk menentukan nilai satu variabel sederhana.
**Solusi 1 baris:**
```go
// Inline Ternary
label := zyra.Ternary(user.Age >= 18, "Dewasa", "Anak-anak")

// Coalesce (Mengambil nilai non-kosong pertama)
displayName := zyra.Coalesce(user.Nickname, user.FullName, "Anonim")
```

### 37. Slice Partitioning (`zyra.Slice.Partition`)
**Masalah:** memisah 1 array menjadi 2 kelompok (misal: user aktif vs non-aktif) butuh 12 baris inisialisasi 2 slice + loop `for` + `if/else`.
**Solusi 1 baris:**
```go
active, inactive := zyra.Slice.Partition(users, func(u User) bool { return u.IsActive })
```

### 38. Slice KeyBy Indexing (`zyra.Slice.KeyBy`)
**Masalah:** mengubah array `[]User` menjadi dictionary `map[ID]User` yang di-index berdasarkan ID user butuh 8 baris loop `map[u.ID] = u`.
**Solusi 1 baris:**
```go
userMap := zyra.Slice.KeyBy(users, func(u User) string { return u.ID }) // map[string]User
```

### 39. Slice Flatten (`zyra.Slice.Flatten`)
**Masalah:** meratakan array 2D `[][]T` (misal: gabungan list tags dari banyak kategori) menjadi 1D `[]T` butuh 8 baris nested `for` loop.
**Solusi 1 baris:**
```go
allTags := zyra.Slice.Flatten(tagsPerCategory) // []string
```

### 40. Slice Sampling & Shuffle (`zyra.Slice.Sample` & `Shuffle`)
**Masalah:** mengambil 1 item acak dari array atau mengacak urutan array (*shuffle*) dengan seed cryptographically secure.
**Solusi 1 baris:**
```go
winner := zyra.Slice.Sample(participants) // 1 item acak *T
shuffledList := zyra.Slice.Shuffle(deck)   // array teracak
```

### 41. Slice Set Difference & Intersection (`zyra.Slice.Difference` & `Intersect`)
**Masalah:** mencari elemen yang ada di Slice A tapi tidak di B (selisih), atau elemen yang sama-sama ada di A & B (irisan) butuh 20+ baris nested loop / map lookup.
**Solusi 1 baris:**
```go
addedTags := zyra.Slice.Difference(newTags, existingTags) // elemen di A yang tidak ada di B
commonTags := zyra.Slice.Intersect(userATags, userBTags)  // irisan A dan B
```

### 42. High-Speed Concurrency Parallel Map (`zyra.Parallel.Map`)
**Masalah di Go:** memproses 1000 item secara paralel dengan *worker pool* (misal: panggil 1000 API eksternal dengan max 10 goroutine concurrent) biasanya butuh 40+ baris `sync.WaitGroup`, channel job, channel result, dan worker loop.
**Solusi 1 baris:**
```go
// Memproses slice secara paralel dengan 10 goroutine concurrent
results, errs := zyra.Parallel.Map(ctx, urlList, 10, func(url string) (Data, error) {
    return fetchSingleUrl(ctx, url)
})
```

### 43. Function Throttling (`zyra.Throttle.Once`)
**Masalah:** mencegah fungsi berat dieksekusi terlalu sering dalam rentang waktu tertentu per-key (misal: re-calculate analytics per user max 1x per 10 detik).
**Solusi 1 baris:**
```go
executed := zyra.Throttle.Once("calc:"+userID, 10*time.Second, func() {
    recalculateUserScore(userID)
})
```

### 44. Struct <-> Map Conversion (`zyra.Struct`)
**Masalah:** mengonversi struct Go ke `map[string]any` atau sebaliknya dengan menghormati `json:"..."` struct tag.
**Solusi 1 baris:**
```go
dataMap, err := zyra.Struct.ToMap(userStruct)
err := zyra.Struct.FromMap(inputMap, &targetStruct)
```

### 45. Smart String Truncate (`zyra.String.Truncate`)
**Masalah:** memotong teks panjang agar rapi di UI tanpa memotong kata di tengah-tengah + menambahkan ellipsis `...`.
**Solusi 1 baris:**
```go
shortText := zyra.String.Truncate("Belajar framework Go + React mudah", 20, "...") 
// Hasil: "Belajar framework..." (potong rapi di batas kata)
```

## CLI Ergonomics — "Generate, Jangan Tulis Manual"

| Command | Fungsi |
|---|---|
| `zyra generate action <Name>` | Scaffold fungsi Go Action + file test-nya |
| `zyra generate page <route>` | Scaffold file halaman baru sesuai konvensi file-based routing, termasuk contoh `meta()` |
| `zyra generate component <Name>` | Scaffold komponen React + file test Vitest |
| `zyra add auth` | Inject modul auth lengkap (login/register/reset password/OAuth) ke project yang sudah ada |
| `zyra add ui <component>` | Eject komponen UI ke project |
| `zyra add stripe` / `zyra add resend` | Inject integrasi pihak ketiga populer sebagai plugin resmi |
| `zyra doctor` | Cek kesehatan environment: versi Go, binary Tailwind tersedia, port bebas, env var wajib, koneksi database |
| `zyra audit` | Scan konfigurasi project untuk kesalahan umum (debug mode aktif di production, CORS wildcard, secret ter-commit) |
| `zyra migrate up/down/status/create` | Migration database |

## Fitur unik: "AI-Ready Error Overlay"

Saat terjadi error di dev mode (baik dari Go Action, render SSR, maupun error React di client), overlay error di browser menampilkan tombol **"Copy prompt untuk AI"** yang menyalin ke clipboard sebuah prompt siap-pakai berisi: pesan error lengkap, file & baris yang relevan, snippet kode sekitar error, dan konteks framework (versi Zyra, mode render halaman). Developer tinggal paste ke ChatGPT/Claude/Zed Agent tanpa perlu menyusun konteks manual.

Ini fitur kecil secara teknis, tapi sangat relevan dengan cara developer bekerja hari ini (2026) dan belum ada framework mainstream yang menyediakannya secara resmi — jadi ini genuine differentiator, bukan gimmick.

## Editor Tooling

- Extension VS Code/Zed (`syntaxes/zyra.tmLanguage.json`): highlight direktif (`// +zyraaction`, `// +zyrastream`, `// +zyracron`), autocomplete nama Go Action di file TSX, "Go to Generated Type" code lens dari Go struct ke file `.generated/*.ts`.
- Perintah "Preview generated TypeScript" langsung dari hover di atas fungsi Go yang beranotasi.

## Progressive Disclosure dalam Dokumentasi

Setiap halaman dokumentasi fitur DX wajib punya 2 bagian:
1. **"Quick Way"** — cara paling singkat pakai helper bawaan (untuk pemula/kasus umum).
2. **"Escape Hatch"** — cara akses layer mentah di baliknya kalau butuh kontrol penuh (untuk profesional/kasus khusus).

## Matriks Garansi Eliminasi Pain Point Framework Lain

Tabel berikut adalah jaminan arsitektural Zyra dalam mengeliminasi pain point paling banyak dikeluhkan dari Next.js, Nuxt, SvelteKit, Remix, Angular Universal, dan Laravel:

| Kategori Framework | Pain Point Asli di Framework Lain | Jawaban Kongkret Arsitektur Zyra |
|---|---|---|
| **Next.js** | Directive `"use client"`/`"use server"` membingungkan & memecah kode | **Bebas Directive:** Pilihan `renderMode = "csr" \| "ssg" \| "ssr"` ditetapkan per-file halaman, tanpa perlu directive `"use client"` di tengah file. |
| **Next.js** | Caching berlapis rahasia (Data Cache/Router Cache) susah di-debug | **Transparent Cache:** Cache bersifat eksplisit via `export const cache = { ttl: 60 }`. Tidak ada hidden caching layer. |
| **Next.js** | Middleware terbatas Edge Runtime (tidak bisa baca body / DB) | **Full Go Middleware:** Middleware berjalan di HTTP pipeline Go murni (bisa akses DB, baca/ubah body, tanpa limit durasi). |
| **Next.js** | `next/image` wajib `width` & `height` manual | **Optional Dimensions:** `<ZyraImage>` membaca dimensi otomatis di backend Go untuk mencegah CLS tanpa paksaan. |
| **Next.js** | Kebocoran Env Variable & Paket Server ke Client | **Hard Compile Filtering:** Variabel tanpa `PUBLIC_` secara fisik dibuang dari client bundle oleh esbuild saat build. |
| **Next.js** | WebSocket/Realtime butuh custom server & gagal di serverless | **Built-in Streams:** Anotasi `// +zyrastream` & `useZyraStream` native terkompilasi di binary Go. |
| **Next.js** | Self-hosting On-demand ISR sangat rumit tanpa Vercel | **VPS-Native ISR:** Single binary Zyra menjalankan `revalidate: N` secara otomatis di VPS $5/bulan manapun. |
| **Nuxt** | Auto-import magic membingungkan & konflik nama | **Explicit Codegen:** Codegen menghasilkan file `.generated/*.ts` yang diimpor secara eksplisit tanpa magic misterius. |
| **Nuxt / SvelteKit** | Hydration Error pada library browser-only (chart, editor) | **`<ZyraClientOnly>`:** Wrapper resmi untuk merender komponen browser-only secara aman tanpa hydration mismatch. |
| **SvelteKit / Remix** | Caching & Form Action handling manual / butuh Zod | **Single Struct Tag:** Tag `validate:"required,email"` di Go generate baik client validation schema maupun server validation. |
| **Remix** | Waterfall Data Fetching pada nested routes | **Concurrent RPC:** Go Actions dipanggil secara paralel/concurrent di server/client tanpa waterfall. |
| **Astro** | Mencampur UI framework jadi berantakan, `client:load` / island state rumit | **Unified React State:** Bebas dari kebingungan island directive. 1 frontend React terintegrasi mulus. |
| **Astro / Gatsby** | Waktu build melambat drastis di proyek besar / GraphQL wajib | **Fast Go Build:** Bundling esbuild + `goja` pool SSR berkecepatan sub-milidetik, tanpa paksaan GraphQL. |
| **SolidStart** | Signal leak antar-request di SSR & `server$` terbatas | **Isolated Memory Pool:** Runtime Goja terisolasi per-request dalam pool Go, bebas dari memory leak antar request. |
| **Rails / Django** | Background job (Sidekiq/Celery) wajib Redis & Celery broker | **Zero-Redis Jobs:** `zyra.Jobs` (+ `// +zyracron`) mengeksekusi job & cron langsung di dalam binary Go tanpa Redis. |
| **Rails / Django** | Asset pipeline rumit (Webpack/Importmap/collectstatic) | **Built-in Asset Pipeline:** Tailwind standalone + esbuild ter-embed tanpa setup Webpack atau collectstatic manual. |
| **Phoenix / Blazor** | Blazor WASM bundle > 1MB / Phoenix LiveView process state rumit | **Lightweight Client Bundle:** Bundle client React ringan (~150KB gzip), RPC latency ultra-rendah (<10ms). |
| **AdonisJS / Spring** | Template engine non-type-safe / restart Spring lambat | **Fast HMR + Full Type-Safety:** Instant `zyra dev` HMR + 45 helper ringkas (`zyra.PDF`, `zyra.Excel`, `zyra.QRCode`). |
| **Meteor.js** | Tightly coupled dengan MongoDB / pub-sub data leak | **Multi-DB SQL & Strict Auth:** Support Relational SQL (Postgres/MySQL/SQLite) + Mongo, fail-safe protected by default. |
| **RedwoodJS** | GraphQL wajib, Cells kaku, Prisma rollback rumit | **No-GraphQL Simplicity:** Go Actions RPC type-safe tanpa GraphQL boilerplate, embedded migration dengan rollback `down.sql`. |
| **Qwik City** | Resumability curve terliku, routeLoader$ un-typed, plugin minim | **Standard React Hydration:** React state tree standar, 100% type-safe RPC codegen, 45 helper & plugin resmi terstruktur. |
| **Fresh (Deno)** | Tanpa build step = bundle besar, tidak ada state global & auth | **Single Binary Opt:** Bundling esbuild ultra-cepat, state React terpadu, modul Auth resmi (`pkg/zyra/auth`) bawaan. |
| **Flask / Express** | Tidak ada struktur standar, Passport.js/Flask-Login boilerplate | **Sensible Structure & Built-in Auth:** File-based routing standar, Auth module bawaan, CSRF & Rate-limit otomatis. |
| **NestJS / Symfony** | Overload decorator (@Controller, @Module) & Twig non-type-safe | **Clean Annotations:** Cukup 1 anotasi `// +zyraaction`, React TSX template 100% type-safe tanpa boilerplate decorator berlebih. |
