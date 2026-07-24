# 01 — Visi & Filosofi Zyra

## Misi

**Zyra adalah framework fullstack Go + React yang menghilangkan seluruh boilerplate dan pain point klasik pengembangan web modern, dikemas dalam satu binary Go yang bisa dijalankan tanpa dependensi runtime apapun.**

Zyra ada bukan untuk jadi "framework Go yang ke-38", tapi untuk jadi jawaban konkret atas pertanyaan: *"Kenapa membangun aplikasi fullstack production-grade masih terasa seperti merakit 15 tools yang tidak saling kenal satu sama lain?"*

## Masalah yang benar-benar diselesaikan Zyra

Developer modern yang membangun aplikasi fullstack biasanya harus merakit sendiri:

- Backend framework (Express/Fiber/Gin) + frontend framework (React/Vue) — dua dunia terpisah, dua bahasa, sering dua repo.
- Layer API manual (REST/GraphQL) hanya untuk menjembatani tipe data antara backend dan frontend — sering rawan out-of-sync.
- Auth dari nol atau dari library pihak ketiga yang harus diintegrasikan manual.
- Security header, CSRF, rate limiting — sering baru dipikirkan setelah insiden, bukan sejak awal.
- Observability (logging, tracing, metrics) — sering ditambahkan belakangan saat sudah production dan kebakaran.
- Migration database — dikerjakan dengan tool terpisah yang tidak nyambung dengan sisa stack.
- Deployment — bingung antara container Node + container Go, atau harus jalankan dua proses runtime sekaligus.

Zyra menyatukan semua ini menjadi satu pengalaman koheren, dengan Go sebagai fondasi (cepat, hemat resource, cross-compile mudah, satu binary) dan React sebagai satu-satunya frontend (fokus, matang, tidak setengah-setengah dukung banyak framework).

## Prinsip Desain (Design Principles)

Prinsip ini adalah **kontrak** yang harus dipatuhi setiap kali menambah fitur baru. Kalau sebuah fitur melanggar salah satu prinsip ini, fitur itu harus didesain ulang, bukan prinsipnya yang dilanggar.

### 1. "One Function Away" — Setiap pain point harus punya solusi satu panggilan fungsi

Kalau developer di framework lain butuh menulis 40 baris boilerplate untuk hal umum (upload file, kirim email, cache hasil query, validasi form), di Zyra itu harus jadi satu import + satu pemanggilan fungsi dengan default yang aman dan masuk akal. Detail lengkap ada di `04-DEVELOPER-EXPERIENCE.md`.

### 2. "Secure by Default, Impossible to Forget"

Developer tidak boleh bisa "lupa" mengaktifkan security dasar. CSRF, rate limiting, security header, validasi input — semua **aktif secara default** sejak `zyra create`. Mematikannya harus eksplisit dan sengaja (opt-out yang jelas terlihat di config), bukan sebaliknya.

### 3. "Zero Runtime Dependency"

Binary hasil `zyra build` harus bisa jalan di server kosong (fresh Ubuntu/Alpine/Windows Server) tanpa install Node.js, Bun, Deno, Python, atau runtime apapun selain OS itu sendiri. Titik.

### 4. "Type-Safety End to End, Tanpa Duplikasi"

Definisikan tipe data **satu kali** di Go. TypeScript, validasi form, dokumentasi API — semua digenerate otomatis dari sumber itu. Developer tidak boleh pernah menulis interface TypeScript yang menduplikasi struct Go secara manual.

### 5. "Progressive Complexity" — Ramah pemula, tidak membatasi profesional

- Pemula harus bisa `zyra create my-app && cd my-app && zyra dev` dan langsung punya aplikasi fullstack yang jalan, tanpa paham konsep rendering mode, migration, atau observability dulu.
- Profesional harus bisa membongkar setiap layer default (ganti rendering engine, ganti query layer, tulis middleware custom, akses raw `*http.Request`) tanpa harus fork framework.
- Setiap fitur "ajaib" (magic) harus punya jalur "escape hatch" yang didokumentasikan berdampingan.

### 6. "Boringly Reliable"

Zyra menyasar juga developer profesional yang alergi terhadap breaking change tiap minggu. Setelah v1.0.0:
- Semantic Versioning dipatuhi ketat.
- Breaking change hanya di major version, dengan migration guide otomatis (`zyra upgrade --migrate`).
- API publik (`pkg/zyra`) dipisah tegas dari `internal/` — hanya API publik yang punya stability guarantee.

### 7. "Convention Familiar, Bukan Convention Aneh"

Kalau konvensi yang sudah dikenal jutaan developer React (misalnya pola `getStaticProps`/`getServerSideProps` ala Next.js Pages Router, file-based routing ala Next.js/Remix) tidak bertentangan dengan arsitektur Zyra, **pakai konvensi yang familiar**. Tujuannya mengurangi kurva belajar — developer harus merasa "oh ini mirip yang aku tahu, cuma lebih enak", bukan "wah harus belajar paradigma baru dari nol".

### 8. "Ownership, Bukan Dependency" untuk UI

Komponen UI di-*eject* langsung ke project user (`zyra add ui button` menyalin kode sumber ke project, bukan menambah package npm). User punya kontrol penuh, tidak terkunci ke versi library UI pihak ketiga, dan tidak menambah node_modules yang menggemuk.

## Target Pengguna (Personas)

Framework ini harus terasa dibuat khusus untuk **keempat** persona ini secara bersamaan — bukan kompromi di tengah-tengah yang tidak memuaskan siapapun:

1. **Mahasiswa/pemula belajar fullstack.** Butuh: instruksi jelas langkah demi langkah, error message yang mengajarkan (bukan stack trace mentah), starter template yang bisa langsung dipelajari strukturnya, dokumentasi tutorial gaya "ikuti langkah 1-2-3".
2. **Indie hacker / freelancer membangun SaaS sendirian.** Butuh: kecepatan dari 0 ke produk jadi, auth+billing+dashboard sudah jadi lewat template, deployment simpel (satu binary, satu server kecil sudah cukup), biaya infra rendah (SQLite + single binary = tidak butuh banyak service terpisah).
3. **Software engineer profesional di startup.** Butuh: performa nyata (bukan cuma klaim marketing), observability untuk debugging production, kemampuan mengganti/extend bagian tertentu tanpa fork, testing story yang jelas.
4. **Tim enterprise/regulated industry.** Butuh: security audit trail, RBAC, compliance-friendly logging, stabilitas API jangka panjang, dukungan database yang sudah mereka pakai.

## Non-Goals (Sengaja TIDAK dikerjakan di v1, dan mungkin selamanya)

Batasan ini penting supaya fokus tidak melebar dan kualitas tidak terkorbankan:

- **Tidak multi-frontend-framework.** Tidak akan mendukung Vue/Svelte/Angular. React-only, dikerjakan sangat matang.
- **Tidak multi-bahasa-backend.** Go-only. Tidak akan ada mode "pakai backend Node/Python/Rust".
- **Tidak membangun ORM raksasa ala Prisma/Hibernate dari nol.** Zyra akan pakai query layer yang typed dan ringan (lihat `07-DATA-LAYER-AND-DATABASE.md`), bukan ORM penuh dengan lazy-loading kompleks yang sering jadi sumber bug performa.
- **Tidak mencoba jadi mobile framework.** Zyra adalah web framework. Integrasi dengan mobile app dimungkinkan lewat mode `api-only` (lihat `10-CLI-AND-PROJECT-TEMPLATES.md`), tapi Zyra tidak generate kode mobile.
- **Tidak mengejar "100% kompatibel dengan ekosistem npm React".** Zyra punya komponen sendiri; kompatibilitas dengan library React pihak ketiga adalah best-effort, bukan garansi.

## Posisi Kompetitif

| Dimensi | Zyra | Next.js / Remix | Astro / Gatsby | Laravel / Rails / Django | Blazor / Spring |
|---|---|---|---|---|---|
| Runtime produksi | Go binary saja | Node / Vercel Edge | Node / Static Host | PHP / Ruby / Python + Web Server | .NET / JVM Runtime |
| Frontend | React (fokus) | React | Multi-Island (kompleks) | Blade / ERB / Django Template | Blazor / Thymeleaf |
| Type-safety FE↔BE | Otomatis dari Go | Manual / tRPC tambahan | Manual | Tidak ada / Perlu manual DTO | C# / Java DTO |
| RAM footprint | Sangat rendah (~20MB) | Tinggi (Node) | Sedang | Sedang | Tinggi (.NET / JVM) |
| Auth bawaan matang | Ya (modul resmi) | Tidak (pihak ke-3) | Tidak | Ya | Kompleks (Identity / Spring Security) |
| Background Jobs | Ya (Zero-Redis) | Tidak | Tidak | Perlu Redis / Celery / Sidekiq | Perlu Quartz / Hangfire |
| Security default | Fail-safe protected | Perlu setup manual | Perlu setup manual | Cukup baik | Kompleks |
| Target deployment | Server apapun, 1 binary | Vercel / Node host | CDN / Node host | Web Server khusus | IIS / Kestrel / Tomcat |

**Kesimpulan posisi:** Zyra menargetkan "kenikmatan developer experience ala Laravel/Rails (batteries-included)" digabung dengan "performa & simplicity deployment ala Go", plus "DX modern React" — kombinasi yang belum ada satupun framework populer saat ini yang menawarkannya secara utuh.

## "Unfair Advantages" yang harus terus dijaga

Ini adalah alasan orang akan pindah ke Zyra — jangan sampai hilang saat development:

1. Satu binary, nol dependency runtime, RAM sangat kecil dibanding stack Node manapun.
2. RPC type-safe otomatis dari anotasi Go — tidak ada framework Go+React lain yang punya ini semulus Zyra.
3. Security & observability yang built-in sejak hari pertama, bukan checklist yang harus dicari-cari plugin-nya.
4. Auth, migration, UI kit, dan starter template yang benar-benar lengkap dan terintegrasi — bukan "silakan install library X, Y, Z sendiri".
