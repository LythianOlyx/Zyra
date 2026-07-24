# 09 — UI Component Library

## Filosofi: Ownership, bukan Dependency

Sama seperti shadcn/ui, komponen di-**eject** (disalin sumber kodenya) ke project user via `zyra add ui <component>`, bukan diinstal sebagai package npm. Ini menghindari:
- Ketergantungan ke versi library UI pihak ketiga yang bisa breaking change.
- Bundle size yang menggembung karena fitur library yang tidak dipakai.
- Kesulitan kustomisasi mendalam (kalau cuma import dari `node_modules`, styling/behaviour susah diubah total).

Karena Zyra berprinsip zero-Node-dependency, komponen ini juga harus tidak butuh package npm tambahan apapun di runtime — styling murni Tailwind + sedikit JS/React murni (tanpa Radix/Headless UI dkk sebagai dependency, walau boleh **belajar** dari pattern accessibility mereka saat menulis komponen sendiri).

## Daftar Komponen UI Kit (target: 40+)

Kategori dasar (harus 100% selesai & accessible sebelum rilis):
`button`, `input`, `textarea`, `select`, `switch`, `checkbox`, `radio-group`, `slider`, `badge`, `alert`, `toast`, `skeleton`, `tooltip`, `progress`, `spinner`, `avatar`.

Kategori data & navigasi:
`data-table` (sorting, filtering, pagination bawaan), `tabs`, `accordion`, `breadcrumbs`, `pagination`, `command-palette`, `dropdown-menu`, `context-menu`, `sidebar-nav`.

Kategori overlay:
`modal`, `drawer`, `popover`, `dialog-confirm` (khusus konfirmasi aksi destruktif, dengan pola aman "ketik untuk konfirmasi" untuk aksi berbahaya).

Kategori form lanjutan (terintegrasi langsung dengan sistem validasi `04-DEVELOPER-EXPERIENCE.md`):
`form` (wrapper yang otomatis sinkron dengan schema validasi Go), `date-picker`, `combobox`, `file-upload` (terintegrasi `zyra.Storage`), `tag-input`.

Kategori feedback & status:
`empty-state`, `error-boundary-ui`, `loading-overlay`, `banner`.

Kategori khusus SEO/marketing:
`hero-section`, `pricing-table`, `testimonial-card`, `faq-accordion`, `cta-section` — komponen siap pakai untuk landing page, karena target user termasuk indie hacker yang butuh landing page cepat jadi tanpa desain dari nol.

## Sistem Theming

- Design token berbasis CSS variables (`--zyra-color-primary`, `--zyra-radius`, dst), diatur di satu file `theme.css` yang digenerate saat `zyra create`.
- Dark mode bawaan: `useZyraTheme()` hook + toggle otomatis tersimpan di `localStorage`, tanpa flicker saat reload (inline script kecil di HTML shell untuk set class sebelum React hydrate — teknik yang sudah umum dipakai, diterapkan otomatis oleh Zyra tanpa developer perlu tahu detailnya).
- Preset tema siap pakai (`zyra theme use <preset>`) untuk yang tidak ingin desain sendiri dari nol.

## Standar Kualitas Wajib per Komponen

1. **Aksesibilitas:** ARIA role & attribute benar, keyboard navigation lengkap (Tab/Escape/Arrow sesuai jenis komponen), focus trap untuk overlay (`modal`/`drawer`).
2. **Responsif:** diuji di breakpoint mobile/tablet/desktop.
3. **Test otomatis:** setiap komponen minimal punya test render dasar + test interaksi kunci (Vitest + React Testing Library).
4. **Dokumentasi:** setiap komponen di website punya contoh kode, daftar props, dan preview live.
5. **Zero dependency tambahan:** hanya React + Tailwind, tidak menambah package npm baru ke project user.

## Alur `zyra add ui`

```bash
zyra add ui data-table
```
1. CLI mendeteksi struktur project (lokasi folder komponen dari `zyra.config.ts`).
2. Menyalin file sumber komponen + file test-nya ke project.
3. Menambahkan entry Tailwind content path kalau belum ada (supaya class Tailwind di komponen baru tidak ke-purge).
4. Menampilkan ringkasan props/contoh pemakaian langsung di terminal.

## Komponen yang terikat langsung ke fitur DX lain (bukan komponen visual biasa)

- `<ZyraBoundary>` — lihat `04-DEVELOPER-EXPERIENCE.md`.
- `<ZyraImage>`, `<ZyraLink>` — lihat `06-SEO-AND-PERFORMANCE.md`.
- `<ZyraForm>`, `<ZyraProtectedRoute>` — lihat dokumen auth & DX.
- `<ZyraSchema>` (JSON-LD) — lihat dokumen SEO.

Ini penting dicatat supaya jelas: sebagian "komponen UI" sebenarnya adalah **integrasi arsitektural**, bukan cuma styling — jangan dikerjakan tim/agent yang berbeda tanpa koordinasi dengan modul yang mereka bungkus.
