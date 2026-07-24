# 06 — SEO & Performance Bawaan

Prinsip: SEO dan performa yang baik harus jadi **hasil otomatis** dari memakai Zyra dengan wajar, bukan pekerjaan tambahan yang harus dikonfigurasi manual satu-satu seperti di kebanyakan framework SPA murni.

## SEO

### Meta tag & Open Graph per halaman
Konvensi `export function meta({ props })` (lihat `02-ARCHITECTURE.md`) otomatis di-inject ke `<head>` oleh Rendering Engine — jalan konsisten baik di mode `ssg`, `ssr`, maupun `csr` (untuk CSR, meta tag dasar tetap di-render di HTML shell awal dari data yang tersedia sebelum hydration, supaya crawler yang tidak eksekusi JS pun tetap melihatnya).

### Sitemap & robots.txt otomatis
`zyra generate sitemap` (dan otomatis dijalankan sebagai bagian `zyra build`) memindai seluruh route hasil file-based routing + halaman dinamis dari `getStaticProps` (misal daftar slug blog), menghasilkan `sitemap.xml` valid dan `robots.txt` default yang aman (tidak accidentally block seluruh site, kesalahan umum yang sering terjadi).

### Structured Data (JSON-LD)
Komponen resmi `<ZyraSchema type="Article" data={...} />` untuk generate JSON-LD terstruktur tanpa developer perlu tahu spesifikasi schema.org secara detail.

### Canonical URL
Otomatis di-generate dari `siteUrl` di config + path halaman, dengan opsi override manual per halaman untuk kasus konten duplikat (misal parameter query/pagination).

### Rendering mode & SEO
Dokumentasi wajib menjelaskan dengan tegas: **halaman publik yang butuh terindeks mesin pencari sebaiknya pakai `ssg` atau `ssr`, bukan `csr`.** CLI `zyra audit` memberi peringatan (bukan error keras) kalau mendeteksi halaman yang terlihat seperti halaman publik (tidak ada `requireAuth`) tapi memakai mode `csr` tanpa SEO meta — supaya developer pemula tidak tanpa sadar merilis landing page yang tidak SEO-friendly.

## Performance

### Core Web Vitals sebagai target eksplisit
- **LCP (Largest Contentful Paint):** didorong oleh SSR/SSG default untuk halaman publik + `<ZyraImage>` yang otomatis set `width`/`height` (mencegah layout shift) dan lazy-load gambar di luar viewport.
- **CLS (Cumulative Layout Shift):** font loading strategy bawaan (`font-display: optional/swap` dikonfigurasi otomatis), reserved space untuk gambar/iklan/embed.
- **INP (Interaction to Next Paint):** code-splitting otomatis per-route (esbuild) supaya JS yang di-download & dieksekusi di setiap halaman minimal.

### `<ZyraImage>` — Komponen Gambar Teroptimasi
```tsx
// Width & height opsional untuk gambar dinamis/lokal!
<ZyraImage src="/uploads/cover.jpg" alt="Cover" priority />
```
Di balik layar: 
1. **Aturan anti-Next.js:** Berbeda dari Next.js `next/image` yang memaksakan `width` & `height` wajib diisi manual (merepotkan untuk gambar dinamis/CMS), `<ZyraImage>` membuat `width` dan `height` **opsional**. Backend Go Zyra otomatis membaca dimensi gambar asli dari file/storage buffer dan men-generate aspect ratio CSS yang tepat untuk mencegah CLS (Cumulative Layout Shift).
2. Resize + convert ke WebP/AVIF via pipeline pure-Go (bagian dari Rendering Engine, lihat `03-RENDERING-ENGINE.md`).
3. Hasil di-cache dengan content-hash, `srcset` responsif digenerate otomatis, `loading="lazy"` default kecuali `priority` di-set untuk gambar above-the-fold.

### `<ZyraLink>` — Prefetching Cerdas
```tsx
<ZyraLink href="/pricing" prefetch="hover">Pricing</ZyraLink>
```
Prefetch data halaman tujuan saat link terlihat di viewport atau di-hover (mirip Next.js `<Link>`), sehingga navigasi berikutnya terasa instan.

### Critical CSS
Untuk halaman `ssg`/`ssr`, CSS kritikal (yang dipakai di atas fold) di-inline langsung ke `<head>` hasil render, sisanya di-load async — mengurangi render-blocking request.

### Performance Budget di Build
```bash
zyra build --budget
```
Build **gagal** (atau warning keras, dikonfigurasi tingkat strictness-nya) kalau ukuran bundle JS per-route melebihi threshold yang dikonfigurasi (misal 200KB gzip) — mencegah "bundle bloat" menumpuk tanpa disadari seiring project bertambah besar.

### Lighthouse CI Template
Starter template menyediakan workflow GitHub Actions siap pakai yang menjalankan Lighthouse CI di setiap PR, dengan threshold skor minimum yang bisa dikonfigurasi per-project.

## Edge & CDN Friendliness

- Header `Cache-Control` yang benar secara default untuk tiap jenis konten (asset immutable vs halaman `ssg` dengan `revalidate` vs halaman `ssr` yang umumnya `no-store` kecuali dikonfigurasi cache eksplisit).
- Dokumentasi resmi cara menaruh Zyra di belakang CDN (Cloudflare/Fastly) termasuk cara handle purge cache saat `revalidate` SSG terjadi.

## Accessibility (a11y) sebagai bagian dari "performa yang dianggap serius"

- Seluruh komponen UI bawaan (`09-UI-COMPONENT-LIBRARY.md`) wajib lolos audit ARIA dasar & keyboard navigation sebelum masuk rilis resmi — bukan best-effort.
- `zyra audit` juga bisa menjalankan pengecekan a11y dasar terhadap output HTML hasil SSG (misal: gambar tanpa `alt`, kontras warna terlalu rendah pada tema default).
