package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/LythianOlyx/Zyra/internal/data/nplusone"
)

// DatabaseConfig specifies database driver and connection info for internal/data.
type DatabaseConfig struct {
	Driver          string        `json:"driver"`
	URL             string        `json:"url"`
	MaxOpenConns    int           `json:"maxOpenConns"`
	MaxIdleConns    int           `json:"maxIdleConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime"`
}

// DB represents the thin Zyra repository & database wrapper on top of sqlx.DB.
type DB struct {
	SqlxDB *sqlx.DB
	Driver string
	URL    string
}

// Open initializes a database connection pool matching cfg.
func Open(cfg DatabaseConfig) (*DB, error) {
	driverName := cfg.Driver
	switch driverName {
	case "sqlite", "sqlite3":
		driverName = "sqlite" // modernc.org/sqlite driver name
	case "postgres", "postgresql", "pgx":
		driverName = "pgx" // jackc/pgx/v5/stdlib driver name
	case "mysql":
		driverName = "mysql"
	default:
		return nil, fmt.Errorf("zyra/db: unsupported database driver %q", cfg.Driver)
	}

	sqlxDB, err := sqlx.Open(driverName, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("zyra/db: failed to open database connection: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		sqlxDB.SetMaxOpenConns(cfg.MaxOpenConns)
	} else {
		sqlxDB.SetMaxOpenConns(25)
	}

	if cfg.MaxIdleConns > 0 {
		sqlxDB.SetMaxIdleConns(cfg.MaxIdleConns)
	} else {
		sqlxDB.SetMaxIdleConns(5)
	}

	if cfg.ConnMaxLifetime > 0 {
		sqlxDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := sqlxDB.Ping(); err != nil {
		_ = sqlxDB.Close()
		return nil, fmt.Errorf("zyra/db: ping failed for driver %q: %w", driverName, err)
	}

	return &DB{
		SqlxDB: sqlxDB,
		Driver: driverName,
		URL:    cfg.URL,
	}, nil
}

// Close closes the underlying database connection pool.
func (db *DB) Close() error {
	if db.SqlxDB != nil {
		return db.SqlxDB.Close()
	}
	return nil
}

// Ping checks if the database connection is alive.
func (db *DB) Ping(ctx context.Context) error {
	return db.SqlxDB.PingContext(ctx)
}

// ExecContext executes a query without returning any rows.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return db.SqlxDB.ExecContext(ctx, query, args...)
}

// GetContext executes a query that returns a single row into dest.
func (db *DB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return db.SqlxDB.GetContext(ctx, dest, query, args...)
}

// SelectContext executes a query returning multiple rows into dest slice.
func (db *DB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return db.SqlxDB.SelectContext(ctx, dest, query, args...)
}

// QueryxContext executes a query that returns sqlx.Rows.
func (db *DB) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return db.SqlxDB.QueryxContext(ctx, query, args...)
}

// QueryRowxContext executes a query expected to return at most one row.
func (db *DB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return db.SqlxDB.QueryRowxContext(ctx, query, args...)
}
