package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/LythianOlyx/Zyra/internal/data"
)

var (
	sqliteMutex sync.Mutex
	LockName    = "zyra_migration_advisory_lock"
)

// StatusInfo holds status metrics for database migrations.
type StatusInfo struct {
	CurrentVersion uint
	Dirty          bool
	AppliedCount   int
}

type noCloseDBDriver struct {
	database.Driver
}

func (d *noCloseDBDriver) Close() error {
	// Prevents golang-migrate from closing the shared *data.DB connection pool
	return nil
}

// AutoMigrate runs pending migrations using advisory locks to prevent race conditions.
func AutoMigrate(ctx context.Context, db *data.DB, migrationFS fs.FS, dir string) error {
	return WithAdvisoryLock(ctx, db, func() error {
		return Up(ctx, db, migrationFS, dir)
	})
}

// Up applies all pending up migrations.
func Up(ctx context.Context, db *data.DB, migrationFS fs.FS, dir string) error {
	m, err := createMigrator(db, migrationFS, dir)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("zyra/migrate: up failed: %w", err)
	}
	return nil
}

// Down rolls back the last applied migration.
func Down(ctx context.Context, db *data.DB, migrationFS fs.FS, dir string) error {
	m, err := createMigrator(db, migrationFS, dir)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("zyra/migrate: down failed: %w", err)
	}
	return nil
}

// Status returns current migration version and status.
func Status(ctx context.Context, db *data.DB, migrationFS fs.FS, dir string) (*StatusInfo, error) {
	m, err := createMigrator(db, migrationFS, dir)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return nil, fmt.Errorf("zyra/migrate: failed to fetch status: %w", err)
	}

	return &StatusInfo{
		CurrentVersion: version,
		Dirty:          dirty,
	}, nil
}

// Create creates a new set of .up.sql and .down.sql migration files on disk.
func Create(name string, dir string) error {
	if dir == "" {
		dir = "migrations"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("zyra/migrate: failed to create migrations directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, name)

	upFile := filepath.Join(dir, fmt.Sprintf("%s.up.sql", baseName))
	downFile := filepath.Join(dir, fmt.Sprintf("%s.down.sql", baseName))

	upHeader := fmt.Sprintf("-- Migration Up: %s\n", name)
	downHeader := fmt.Sprintf("-- Migration Down: %s\n", name)

	if err := os.WriteFile(upFile, []byte(upHeader), 0644); err != nil {
		return fmt.Errorf("zyra/migrate: failed to write %s: %w", upFile, err)
	}
	if err := os.WriteFile(downFile, []byte(downHeader), 0644); err != nil {
		return fmt.Errorf("zyra/migrate: failed to write %s: %w", downFile, err)
	}

	fmt.Printf("✓ Created migration files:\n  - %s\n  - %s\n", upFile, downFile)
	return nil
}

// WithAdvisoryLock wraps fn with database-specific advisory locks across Postgres, MySQL, and SQLite.
func WithAdvisoryLock(ctx context.Context, db *data.DB, fn func() error) error {
	switch db.Driver {
	case "pgx", "postgres", "postgresql":
		return postgresAdvisoryLock(ctx, db.SqlxDB.DB, fn)
	case "mysql":
		return mysqlAdvisoryLock(ctx, db.SqlxDB.DB, fn)
	case "sqlite":
		sqliteMutex.Lock()
		defer sqliteMutex.Unlock()
		return fn()
	default:
		return fn()
	}
}

func postgresAdvisoryLock(ctx context.Context, sqlDB *sql.DB, fn func() error) error {
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("zyra/migrate: failed to acquire DB connection for advisory lock: %w", err)
	}
	defer conn.Close()

	lockID := stringToLockID(LockName)

	var acquired bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired); err != nil {
		return fmt.Errorf("zyra/migrate: failed to execute pg_try_advisory_lock: %w", err)
	}

	if !acquired {
		return fmt.Errorf("zyra/migrate: could not acquire postgres advisory lock (another replica running migration)")
	}

	defer func() {
		var unlocked bool
		_ = conn.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1)", lockID).Scan(&unlocked)
	}()

	return fn()
}

func mysqlAdvisoryLock(ctx context.Context, sqlDB *sql.DB, fn func() error) error {
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("zyra/migrate: failed to acquire DB connection for mysql advisory lock: %w", err)
	}
	defer conn.Close()

	var acquired sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 10)", LockName).Scan(&acquired); err != nil {
		return fmt.Errorf("zyra/migrate: failed to execute GET_LOCK: %w", err)
	}

	if !acquired.Valid || acquired.Int64 != 1 {
		return fmt.Errorf("zyra/migrate: could not acquire mysql advisory lock (another replica running migration)")
	}

	defer func() {
		var released sql.NullInt64
		_ = conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", LockName).Scan(&released)
	}()

	return fn()
}

func createMigrator(db *data.DB, migrationFS fs.FS, dir string) (*migrate.Migrate, error) {
	if dir == "" {
		dir = "."
	}
	sourceDriver, err := iofs.New(migrationFS, dir)
	if err != nil {
		return nil, fmt.Errorf("zyra/migrate: failed to create iofs source driver: %w", err)
	}

	var dbDriver database.Driver
	switch db.Driver {
	case "pgx", "postgres", "postgresql":
		dbDriver, err = postgres.WithInstance(db.SqlxDB.DB, &postgres.Config{})
	case "mysql":
		dbDriver, err = mysql.WithInstance(db.SqlxDB.DB, &mysql.Config{})
	case "sqlite":
		dbDriver, err = sqlite.WithInstance(db.SqlxDB.DB, &sqlite.Config{})
	default:
		return nil, fmt.Errorf("zyra/migrate: unsupported migration database driver %q", db.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("zyra/migrate: failed to initialize database migration driver: %w", err)
	}

	wrappedDriver := &noCloseDBDriver{Driver: dbDriver}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, db.Driver, wrappedDriver)
	if err != nil {
		return nil, fmt.Errorf("zyra/migrate: failed to create migrator instance: %w", err)
	}
	return m, nil
}

func stringToLockID(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return int64(h.Sum64())
}
