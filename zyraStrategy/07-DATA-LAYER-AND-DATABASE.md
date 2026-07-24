# 07 — Data Layer & Database

## Filosofi

Zyra **tidak** membangun ORM besar dengan lazy-loading kompleks ala Hibernate/Prisma (sumber bug performa N+1 klasik). Zyra memakai **typed query layer yang tipis**: developer menulis query yang jelas, Go memberi tipe hasil yang aman, tanpa magic yang menyembunyikan apa yang sebenarnya terjadi di database.

## Dukungan Database (dipertahankan dari v1)

| Database | Driver | Catatan |
|---|---|---|
| PostgreSQL | `pgx`/`lib/pq` | Rekomendasi utama untuk production |
| MySQL/MariaDB | `go-sql-driver/mysql` | |
| SQLite | `modernc.org/sqlite` (pure Go, no CGO) | Cocok untuk indie hacker/deployment "zero infra" — database ikut ke dalam binary/volume |
| MongoDB | `mongo-driver` | Adapter terpisah, non-relational |
| Firebase/Firestore | `firebase.google.com/go/v4` | Adapter terpisah |
| Supabase | Pakai adapter Postgres + client Supabase opsional untuk fitur Storage/Realtime | |

## Migration System

- Berbasis `golang-migrate/migrate` sebagai **library** (bukan CLI eksternal terpisah) — tetap terkompilasi ke dalam binary Zyra, jadi tidak menambah dependency runtime.
- File migration (`migrations/0001_create_users.up.sql`, `.down.sql`) di-embed via `//go:embed` ke binary hasil `zyra build`.
- CLI: `zyra migrate create <name>`, `zyra migrate up`, `zyra migrate down`, `zyra migrate status`.
- **Auto-migrate on boot (opsional, dikontrol config)**: `zyra.AutoMigrate(ctx)` dipanggil di awal `main()` template starter — cocok untuk deployment sederhana. Pada deployment multi-replica (horizontal scaling), `zyra.AutoMigrate` menggunakan **PostgreSQL/MySQL Advisory Locks** untuk menjamin hanya 1 replica yang mengeksekusi migration pada satu waktu (mencegah *race condition* schema migration).
- `zyra db:seed` untuk data awal (development/demo).
- **Automated Database Backup Helper (`zyra.Backup`)**: Helper built-in `zyra.Backup.Run(ctx, destinationStorage)` untuk melakukan *online WAL backup* pada SQLite atau streaming dump pada Postgres langsung ke S3/R2 storage — fitur berharga bagi *indie hacker* & SaaS skala kecil.
- **Multi-Tenancy Isolation (`zyra.Tenant`)**: Context helper `zyra.ContextWithTenant(ctx, tenantID)` yang otomatis menginjeksikan filter `tenant_id` ke dalam query repository & path storage untuk mencegah kebocoran data antar tenant pada aplikasi SaaS.

## Query Layer

Pendekatan yang direkomendasikan (dua opsi, keduanya masuk v1 — lihat catatan effort di `15-ROADMAP-AND-MILESTONES.md`):

### Opsi dasar (wajib ada, effort sedang): Repository pattern tipis di atas `sqlx`
```go
type UserRepo struct{ db *zyra.DB }

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
    var u User
    err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
    return &u, err
}
```
Semua query **wajib** pakai parameter placeholder (`$1`, `?`) — tidak ada API yang menerima string SQL hasil concatenation, ini juga bagian dari kontrak security (`05-SECURITY.md`).

### Opsi lanjutan (stretch di v1, boleh paralel dikerjakan sub-agent terpisah): codegen ala `sqlc`
Developer menulis SQL murni beranotasi di file `.sql`:
```sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
```
`zyra generate db` men-scan file ini dan menghasilkan fungsi Go bertipe otomatis (`GetUserByEmail(ctx, email) (User, error)`) — konsisten dengan filosofi codegen yang sama dipakai untuk Go Actions. Ini murni tambahan di atas opsi dasar, bukan pengganti — kalau waktu terbatas, opsi dasar saja sudah cukup untuk `v1.0.0`, opsi ini bisa masuk sebagai `v1.1` **tanpa** melanggar prinsip "all-in v1" karena bukan termasuk killer feature inti (lihat `14-V1-KILLER-FEATURES-ALL-IN.md` untuk daftar final mana yang benar-benar wajib).

## Transaction Helper
```go
err := zyra.DB.Transaction(ctx, func(tx *zyra.Tx) error {
    if err := tx.Exec(...); err != nil { return err }
    if err := tx.Exec(...); err != nil { return err }
    return nil
})
```
Rollback otomatis kalau closure mengembalikan error, commit otomatis kalau sukses — pola yang sudah familiar dari banyak bahasa lain, mengurangi bug lupa rollback/commit.

## Konvensi Model

- Timestamp otomatis (`created_at`, `updated_at`) lewat helper/hook opsional, tidak dipaksa.
- Soft delete opsional (`deleted_at`) sebagai konvensi yang didukung helper query (`WithTrashed()`, `OnlyTrashed()`) mirip Laravel — familiar bagi banyak developer yang pindah dari ekosistem PHP.

## Koneksi & Observability

- Connection pool dikonfigurasi sensible default (`max_open_conns`, `max_idle_conns`) dengan override di config.
- Health check koneksi database terintegrasi ke endpoint `/readyz` (reuse pola v1) — load balancer/orchestrator tahu kapan instance benar-benar siap terima traffic.
- **N+1 Query Auto-Detector (Anti-Laravel N+1 Pain Point):** Di mode development, Zyra men-scan pola query berulang yang dieksekusi dalam loop pada request yang sama. Jika terdeteksi query identik dijalankan N kali berulang dalam satu request handler, DevTools Panel & terminal dev log menampilkan peringatan keras **N+1 Query Warning** beserta file & baris kode tempat query tersebut dipanggil.
- Setiap query lambat (di atas threshold yang dikonfigurasi) otomatis di-log sebagai warning terstruktur lewat `internal/observability` — membantu menemukan index yang hilang sebelum jadi masalah production.

## Testing Data Layer

- `zyra.Test.WithTestDB(t)` — helper yang menyiapkan database sementara (SQLite in-memory untuk test cepat, atau container Postgres kalau test butuh fitur spesifik Postgres), menjalankan migration otomatis, dan membersihkan setelah test selesai. Detail lengkap ada di `11-TESTING-STRATEGY.md`.
