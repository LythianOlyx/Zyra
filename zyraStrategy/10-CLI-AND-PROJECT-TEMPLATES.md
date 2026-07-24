# 10 — CLI `zyra create` & Daftar Template Project

## Alur Interaktif `zyra create`

```
$ zyra create

? Nama project: my-app
? Pilih template:
   1) blank              - Kosong, minimal, belajar dari nol
   2) saas-starter        - Auth + billing + dashboard (paling lengkap)
   3) dashboard-admin      - Admin panel + data table + RBAC
   4) landing-page          - Marketing site, SEO-first, blog
   5) ecommerce              - Storefront + cart + checkout
   6) ai-chat                 - Chat app streaming ala ChatGPT
   7) blog-cms                 - Blog/konten dengan MDX
   8) realtime-collab            - Kanban/chat real-time
   9) api-only                     - Headless, tanpa halaman React
  10) portfolio                     - Website personal, sangat sederhana

? Pilih database: [Postgres / MySQL / SQLite / MongoDB / Firebase / Supabase / Skip]
? Aktifkan auth bawaan? [Yes/No]
? Aktifkan observability (OpenTelemetry+Prometheus)? [Yes/No, default Yes]
? Inisialisasi git repo? [Yes/No]

✓ Project dibuat di ./my-app
✓ Dependency Go terpasang
✓ Binary Tailwind standalone siap
✓ Migration awal dijalankan

Langkah berikutnya:
  cd my-app
  zyra dev
```

Setiap pilihan di atas menampilkan deskripsi singkat + (di masa depan, lewat website) link preview screenshot — supaya user, terutama pemula, tahu apa yang akan dia dapat sebelum memilih.

## Detail Setiap Template

### 1. `blank`
- **Target:** belajar konsep dasar Zyra dari nol.
- **Isi:** satu halaman `pages/index.tsx`, satu contoh `// +zyraaction`, tanpa auth/database aktif.
- **Kenapa penting:** framework yang bagus untuk pemula harus punya titik awal yang tidak membanjiri dengan fitur — supaya konsep inti (routing, action, rendering mode) terlihat jelas tanpa "noise" dari fitur SaaS yang belum relevan.

### 2. `saas-starter` (flagship template)
- **Target:** indie hacker/startup yang ingin cepat punya produk SaaS jadi.
- **Isi:** auth lengkap (register/login/OAuth/reset password), integrasi Stripe (checkout + webhook + halaman billing), dashboard pelanggan, halaman landing page dasar, RBAC (admin vs user biasa), email transaksional (`zyra.Mail`) untuk welcome/invoice.
- **Rendering mode:** landing = `ssg`, dashboard = `csr` (di balik auth, tidak butuh SEO), halaman blog (opsional) = `ssg`.

### 3. `dashboard-admin`
- **Target:** internal tools/admin panel.
- **Isi:** data table dengan sorting/filter/pagination server-side, form CRUD lengkap dengan validasi, RBAC granular per-menu, grafik dasar (chart komponen).

### 4. `landing-page`
- **Target:** marketing site, portofolio produk, halaman pra-peluncuran.
- **Isi:** hero section, pricing table, testimonial, FAQ, blog dengan MDX, semua `ssg`, dilengkapi sitemap+meta+JSON-LD otomatis, skor Lighthouse tinggi dari awal sebagai contoh nyata (dogfooding klaim performa Zyra).

### 5. `ecommerce`
- **Target:** toko online skala kecil-menengah.
- **Isi:** katalog produk (`ssg`/`ssr` campuran), keranjang belanja (state client + Go Action), checkout dengan Stripe, halaman admin produk dasar, webhook pembayaran.

### 6. `ai-chat`
- **Target:** aplikasi chat berbasis LLM (asisten AI, chatbot customer service, dst).
- **Isi:** UI chat siap pakai, streaming response lewat `// +zyrastream` (SSE), contoh integrasi ke API LLM (placeholder yang mudah diganti ke OpenAI/Anthropic/model lain), riwayat percakapan tersimpan di database.

### 7. `blog-cms`
- **Target:** blog personal/tim kecil, dokumentasi produk.
- **Isi:** penulisan konten via file MDX (atau opsional CMS sederhana berbasis database + admin form), `ssg` dengan `revalidate`, RSS feed otomatis, syntax highlighting kode bawaan.

### 8. `realtime-collab`
- **Target:** showcase kemampuan real-time Zyra (kanban board kolaboratif atau chat tim).
- **Isi:** WebSocket/SSE, presence indicator ("siapa yang online"), optimistic update (`useZyraMutation`), demonstrasi `zyra.Broadcast`.

### 9. `api-only`
- **Target:** tim yang sudah punya frontend terpisah (mobile app native, atau frontend lain), hanya butuh backend Go yang type-safe dan cepat.
- **Isi:** tanpa folder `pages/`, hanya `actions/`, otomatis men-generate **OpenAPI spec** dari anotasi `// +zyraaction` (selain TypeScript client) — supaya tetap bisa dikonsumsi klien non-TypeScript (Swift/Kotlin/dll) lewat tooling generate klien dari OpenAPI standar.
- **Kenapa penting secara strategi:** ini memperluas total addressable market Zyra ke tim yang tidak butuh React sama sekali dari Zyra, hanya butuh backend Go yang powerful — tanpa melanggar prinsip "React-only frontend" karena di mode ini memang tidak ada frontend yang di-serve oleh Zyra sama sekali.

### 10. `portfolio`
- **Target:** pemula yang baru belajar, ingin portfolio pribadi sederhana.
- **Isi:** satu halaman, form kontak (`// +zyraaction` + `zyra.Mail`), sangat minimal tapi menunjukkan pola lengkap end-to-end (halaman → action → email) dalam satu contoh kecil yang mudah dipahami — ideal jadi "tutorial pertama" resmi di website.

## Struktur Folder per Template (konsisten di semua template)

```
my-app/
  actions/            # Go Actions (+zyraaction, +zyrastream, +zyracron)
  pages/               # File-based routing React
  components/           # Komponen React kustom project (di luar UI kit bawaan)
  migrations/             # SQL migration
  public/                  # Aset statis
  zyra.config.ts            # Config utama, typed
  .env.example                # Daftar env var wajib (dibaca zyra.doctor)
  main.go                       # Entry point, wiring minimal
```

## Prinsip Kualitas Template

- Semua template **wajib** lolos `zyra audit` tanpa temuan kritis sejak digenerate — template resmi tidak boleh mencontohkan praktik tidak aman.
- Semua template **wajib** menyertakan minimal 1 contoh test (unit Action + component) — supaya sejak hari pertama, pemula sudah melihat pola testing yang benar, bukan belajar testing belakangan sebagai renungan.
- Setiap template dites end-to-end sebagai bagian CI framework itu sendiri (`zyra create --template X` dijalankan di CI, lalu `zyra build` harus sukses) — mencegah "template rusak karena API framework berubah tapi template lupa diupdate", masalah umum di banyak framework yang punya banyak starter kit.
