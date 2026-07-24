# 16 — Strategi Website (Cloudflare Pages)

**Catatan prioritas:** sesuai keputusanmu, website dikerjakan **setelah** framework inti selesai (Fase 10 di `15-ROADMAP-AND-MILESTONES.md`). Dokumen ini disiapkan sekarang supaya keputusan arsitektur website tidak improvisasi mendadak nanti, dan supaya struktur kontennya bisa dicicil (misal nulis draft tutorial) kapan saja tanpa menunggu coding-nya dimulai.

## Klarifikasi penting: Node/Bun di website TIDAK melanggar prinsip "zero dependency" framework

Prinsip "zero runtime dependency" di `01-VISION-AND-PHILOSOPHY.md` berlaku untuk **binary hasil `zyra build` yang dijalankan user Zyra di server produksi mereka**. Website dokumentasi ini adalah **proyek terpisah** (situs statis marketing/dokumentasi), bukan aplikasi yang dibangun memakai runtime Zyra. Memakai Vite + Bun/Node **hanya saat proses build di CI Cloudflare Pages** (yang hasilnya file statis) tidak bertentangan sama sekali dengan klaim framework — ini penting dijelaskan juga nanti di FAQ website itu sendiri, supaya tidak menimbulkan kebingungan yang sama seperti yang kamu tanyakan ke saya.

## Kondisi Existing yang Perlu Dipertahankan

Repo sudah punya fondasi di folder `website/`: Vite + React + TypeScript + Tailwind + Cloudflare Pages (`wrangler.json`, `_headers`, `_routes.json`). Strategi ini melanjutkan stack tersebut, tidak mengganti dari nol.

## Sitemap Informasi Website

1. **Landing/Home** — hero singkat, killer feature (ambil langsung dari `14-V1-KILLER-FEATURES-ALL-IN.md`), tabel perbandingan kompetitor (dari `01-VISION-AND-PHILOSOPHY.md`), CTA `zyra create`.
2. **Docs** — referensi teknis lengkap, tersusun per kategori (Getting Started, Routing & Rendering, Go Actions, Auth, Database, Security, Deployment, CLI Reference, API Reference `pkg/zyra`).
3. **Tutorials** — konten step-by-step, bukan referensi kering. Wajib ada minimal:
   - "Build your first Zyra app in 10 minutes" (pakai template `portfolio`, cocok untuk pemula absolut).
   - "Build a SaaS with auth & billing in under an hour" (pakai `saas-starter`).
   - "Migrating from Next.js to Zyra".
   - "Migrating from Express + React to Zyra".
4. **Templates Gallery** — showcase 10 starter template dengan screenshot/live preview, tombol copy command `zyra create --template <x>`.
5. **Changelog** — digenerate dari `CHANGELOG.md` di repo utama (parse otomatis saat build website, jangan tulis ulang manual dua kali di dua tempat berbeda).
6. **Blog** (opsional, growth-oriented) — artikel SEO seperti "Alternatif Next.js berbasis Go", studi kasus benchmark, dsb.
7. **Community** — link Discord/GitHub Discussions **yang benar-benar aktif** (jangan publish placeholder link — ini poin kredibilitas yang sangat penting).
8. **Playground** (stretch goal, boleh menyusul) — coba Zyra langsung di browser.

## Arsitektur Konten Docs/Tutorial

- Simpan konten sebagai file Markdown/MDX di `website/src/content/`, bukan hardcode di komponen React — memisahkan penulisan konten dari coding tampilan (memudahkan kontribusi komunitas ke dokumentasi lewat PR markdown biasa).
- **Docs harus SEO-friendly**, jadi tiap halaman docs di-*prerender* jadi HTML statis saat build website (Vite SSG plugin, atau prerender script sederhana) — bukan SPA murni yang kosong sampai JS jalan, karena docs adalah sumber traffic organik terbesar untuk framework baru.
- **Versioning docs sejak hari pertama** (`/docs/v1/...`), supaya saat `v2.0.0` framework rilis suatu saat nanti, docs lama tidak perlu migrasi struktur URL yang menyakitkan.
- Search berbasis **Pagefind** (static search index, tidak butuh server/API berbayar, cocok untuk hosting statis di Cloudflare Pages).

## Desain & Performa Website (harus jadi bukti hidup klaim framework)

- Website ini sendiri harus mendapat skor Lighthouse tinggi (90+) — kalau website Zyra sendiri lambat, itu kontradiksi langsung dengan pesan marketing performa.
- Dark mode, code block dengan tombol copy, syntax highlighting untuk Go & TSX.
- OG image otomatis per halaman docs/tutorial (untuk share ke social media terlihat profesional).

## SEO Website

- Sitemap.xml + robots.txt (ironisnya, ini kesempatan bagus untuk nanti "dogfood" fitur SEO milik Zyra kalau website di-generate ulang memakai Zyra sendiri suatu saat — catat sebagai ide menarik jangka panjang, bukan komitmen v1).
- JSON-LD `TechArticle`/`SoftwareApplication` di halaman relevan.
- Target keyword long-tail di konten blog: "Go fullstack framework", "Next.js alternative Go", "tRPC alternative Go", "framework Go untuk React".

## Analytics

- Pakai analytics privacy-friendly (Cloudflare Web Analytics) — tidak butuh cookie consent banner yang mengganggu UX, selaras juga dengan citra "developer-first, tidak ribet".

## Deployment

- Lanjutkan setup `wrangler.json` yang sudah ada.
- `_headers` diisi security header dasar untuk situs statis ini juga (CSP, HSTS, dst) — supaya website resmi Zyra sendiri jadi contoh nyata praktik yang direkomendasikan frameworknya.
- `_routes.json` diatur untuk SPA fallback + exclude asset statis dari fallback routing.

## Urutan Pengerjaan Website (saat Fase 10 dimulai)

1. Restrukturisasi konten docs jadi Markdown/MDX + prerender pipeline.
2. Landing page (paling penting untuk first impression & konversi).
3. Docs inti (Getting Started dulu, baru referensi lengkap menyusul).
4. Tutorials (mulai dari "10 minutes" karena paling menentukan kesan pertama pemula).
5. Templates Gallery.
6. Changelog otomatis.
7. Blog & Playground (paling akhir, nice-to-have).
