# Zyra Framework v1 — Strategy & Blueprint

Folder ini adalah **single source of truth** untuk Zyra: fullstack framework Go + React (murni React, tanpa Node/Bun di runtime), production-ready sejak v1.0.0 pertama kali dirilis — bukan "produk MVP yang dilengkapi belakangan".

## Prinsip folder ini

- Setiap file punya satu topik yang jelas. Jangan gabung keputusan lintas topik ke satu file — supaya mudah di-maintain dan mudah dijadikan referensi terpisah oleh AI coding agent.
- Ini dokumen hidup. Kalau ada keputusan berubah saat development, **update file terkait**, jangan buat file baru yang menduplikasi.

## Urutan baca yang disarankan

| # | File | Isi |
|---|---|---|
| 01 | `01-VISION-AND-PHILOSOPHY.md` | Kenapa Zyra ada, siapa target user, prinsip desain, non-goals |
| 02 | `02-ARCHITECTURE.md` | Arsitektur teknis inti, struktur folder, request lifecycle |
| 03 | `03-RENDERING-ENGINE.md` | CSR/SSG/SSR tanpa Bun, strategi Tailwind tanpa Node |
| 04 | `04-DEVELOPER-EXPERIENCE.md` | Daftar pain point developer + solusi "satu fungsi" di Zyra |
| 05 | `05-SECURITY.md` | Security blueprint lengkap |
| 06 | `06-SEO-AND-PERFORMANCE.md` | SEO & performa bawaan |
| 07 | `07-DATA-LAYER-AND-DATABASE.md` | ORM/query layer, migration, multi-DB |
| 08 | `08-AUTHENTICATION-AND-AUTHORIZATION.md` | Sistem auth & RBAC lengkap |
| 09 | `09-UI-COMPONENT-LIBRARY.md` | Strategi komponen UI |
| 10 | `10-CLI-AND-PROJECT-TEMPLATES.md` | `zyra create`, daftar template, struktur masing-masing |
| 11 | `11-TESTING-STRATEGY.md` | Strategi testing unit/komponen/e2e |
| 12 | `12-DEPLOYMENT-AND-PRODUCTION.md` | Deployment, observability, single binary |
| 13 | `13-PLUGIN-SYSTEM.md` | Sistem plugin & extensibility |
| 14 | `14-V1-KILLER-FEATURES-ALL-IN.md` | Daftar final fitur yang WAJIB ada di v1.0.0 |
| 15 | `15-ROADMAP-AND-MILESTONES.md` | Urutan eksekusi realistis (bukan pengurangan fitur) |
| 16 | `16-WEBSITE-STRATEGY.md` | Website dokumentasi/tutorial/changelog di Cloudflare Pages |
| 17 | `17-AI-EXECUTION-PROMPTS.md` | Master Prompt & daftar prompt eksekusi Bahasa Inggris per-fase untuk AI coding agent |

## Keputusan besar yang sudah final (jangan diubah-ubah lagi tanpa alasan kuat)

1. **React-only.** Satu frontend React, dikerjakan dengan sangat matang.
2. **Zero runtime dependency di production.** Tidak ada Node.js, tidak ada Bun, tidak ada Deno yang wajib terpasang di server produksi. Satu binary Go, titik.
3. **Tailwind tetap didukung penuh**, tapi lewat Tailwind Standalone CLI binary yang di-manage sendiri oleh Zyra — bukan lewat `npx`/`node_modules`.
4. **Tiga mode rendering per halaman** (CSR default, SSG, SSR opsional via embedded JS engine) — developer pilih per halaman, bukan per-project.
5. **"All-in" di v1.0.0.** Semua fitur di `14-V1-KILLER-FEATURES-ALL-IN.md` WAJIB ada sebelum tag `v1.0.0` dirilis ke publik. Yang boleh diatur ulang adalah **urutan pengerjaan internal** (lihat `15-ROADMAP-AND-MILESTONES.md`), bukan cakupan fiturnya.
6. **Struktur Kode Internal yang Terorganisir.** Modul di `internal/middleware`, `internal/observability`, `internal/router`, `internal/codegen`, `internal/config`, `internal/testhelpers` dikembangkan secara terstruktur dengan kualitas produksi.

## Cara pakai folder ini dengan AI coding agent

1. Mulai project baru (repo/folder kosong atau repo baru).
2. Copy seluruh folder `zyraStrategy/` ke root project baru.
3. Tempel **Master System Prompt** dari [17-AI-EXECUTION-PROMPTS.md](file:///home/lythian/MyProject/Framework/Zyra/zyraStrategy/17-AI-EXECUTION-PROMPTS.md) pada awal sesi AI agent.
4. Salin prompt fase spesifik (misal: *Phase 0 Prompt*, *Phase 1 Prompt*) untuk instruksi pengerjaan per-fase.
5. Kerjakan per-fase sesuai [15-ROADMAP-AND-MILESTONES.md](file:///home/lythian/MyProject/Framework/Zyra/zyraStrategy/15-ROADMAP-AND-MILESTONES.md) dan pastikan *Definition of Done* (DoD) terpenuhi sebelum lanjut ke fase berikutnya.

## Status Verifikasi Blueprint v1 (100% Kedap Air)

- **Audit & Penyelarasan:** Seluruh 17 file blueprint di folder ini telah diverifikasi dan diselaraskan secara menyeluruh.
- **Garansi Anti-Pain-Point:** Mengeliminasi secara eksplisit **340+ pain point** dari **24 framework fullstack utama** (Next.js, Nuxt, SvelteKit, Remix, Angular, Laravel, Astro, SolidStart, Rails, Django, Phoenix, Blazor, Gatsby, AdonisJS, Spring, Meteor, RedwoodJS, Qwik, Fresh, Flask, Express, Hydrogen, NestJS, Symfony).
- **Kaya Utilitas DX:** Dilengkapi dengan **45 helper fungsi ringkas** (`zyra.Slice`, `zyra.PDF`, `zyra.Excel`, `zyra.Jobs`, `zyra.I18n`, `zyra.Backup`, `zyra.Parallel`, dll.) untuk memangkas *boilerplate* harian developer.
- **Status Eksekusi:** Siap dieksekusi dari **Fase 0 (Fondasi Kode Core)** sesuai petunjuk [15-ROADMAP-AND-MILESTONES.md](file:///home/lythian/MyProject/Framework/Zyra/zyraStrategy/15-ROADMAP-AND-MILESTONES.md).
