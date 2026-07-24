package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/zyra-framework/zyra/internal/data"
)

// DestinationStorage is an interface for writing database backup artifacts (e.g. S3, local disk, custom stream).
type DestinationStorage interface {
	WriteBackup(ctx context.Context, filename string, reader io.Reader) error
}

// FileDestination implements DestinationStorage for local filesystem destinations.
type FileDestination struct {
	Dir string
}

func (f *FileDestination) WriteBackup(ctx context.Context, filename string, reader io.Reader) error {
	if err := os.MkdirAll(f.Dir, 0755); err != nil {
		return fmt.Errorf("zyra/backup: failed to create target directory: %w", err)
	}

	targetPath := filepath.Join(f.Dir, filename)
	outFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("zyra/backup: failed to create backup file %s: %w", targetPath, err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("zyra/backup: failed to write backup stream: %w", err)
	}
	return nil
}

// BackupManager orchestrates automated online SQLite WAL backups and Postgres dumps.
type BackupManager struct {
	db *data.DB
}

// NewBackupManager initializes a backup manager for the given database connection.
func NewBackupManager(db *data.DB) *BackupManager {
	return &BackupManager{db: db}
}

// Run executes an online backup based on the database driver type and streams to destinationStorage.
func (bm *BackupManager) Run(ctx context.Context, dest DestinationStorage) error {
	if bm.db == nil {
		return fmt.Errorf("zyra/backup: database connection is nil")
	}

	timestamp := time.Now().Format("20060102_150405")

	switch bm.db.Driver {
	case "sqlite":
		filename := fmt.Sprintf("zyra_sqlite_backup_%s.db", timestamp)
		return bm.runSQLiteBackup(ctx, dest, filename)
	case "pgx", "postgres", "postgresql":
		filename := fmt.Sprintf("zyra_postgres_dump_%s.sql", timestamp)
		return bm.runPostgresBackup(ctx, dest, filename)
	case "mysql":
		filename := fmt.Sprintf("zyra_mysql_dump_%s.sql", timestamp)
		return bm.runMySQLBackup(ctx, dest, filename)
	default:
		return fmt.Errorf("zyra/backup: unsupported database driver %q for backup", bm.db.Driver)
	}
}

func (bm *BackupManager) runSQLiteBackup(ctx context.Context, dest DestinationStorage, filename string) error {
	tempFile, err := os.CreateTemp("", "zyra_sqlite_vacuum_*.db")
	if err != nil {
		return fmt.Errorf("zyra/backup: failed to create temporary backup file: %w", err)
	}
	tempPath := tempFile.Name()
	_ = tempFile.Close()
	defer os.Remove(tempPath)

	// Execute SQLite online WAL snapshot using VACUUM INTO
	vacuumQuery := fmt.Sprintf("VACUUM INTO '%s'", tempPath)
	if _, err := bm.db.ExecContext(ctx, vacuumQuery); err != nil {
		return fmt.Errorf("zyra/backup: online SQLite VACUUM INTO failed: %w", err)
	}

	r, err := os.Open(tempPath)
	if err != nil {
		return fmt.Errorf("zyra/backup: failed to read temp vacuum file: %w", err)
	}
	defer r.Close()

	return dest.WriteBackup(ctx, filename, r)
}

func (bm *BackupManager) runPostgresBackup(ctx context.Context, dest DestinationStorage, filename string) error {
	// Execute pg_dump if available, fallback to streaming table export
	pgDumpPath, err := exec.LookPath("pg_dump")
	if err == nil {
		cmd := exec.CommandContext(ctx, pgDumpPath, bm.db.URL)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("zyra/backup: failed to create stdout pipe for pg_dump: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("zyra/backup: failed to start pg_dump process: %w", err)
		}

		if err := dest.WriteBackup(ctx, filename, stdout); err != nil {
			_ = cmd.Process.Kill()
			return err
		}

		return cmd.Wait()
	}

	// Fallback SQL schema/data dump query if pg_dump CLI is not present on host
	return fmt.Errorf("zyra/backup: pg_dump CLI tool not found on system PATH for Postgres dump")
}

func (bm *BackupManager) runMySQLBackup(ctx context.Context, dest DestinationStorage, filename string) error {
	mysqldumpPath, err := exec.LookPath("mysqldump")
	if err != nil {
		return fmt.Errorf("zyra/backup: mysqldump CLI tool not found on system PATH for MySQL dump")
	}

	cmd := exec.CommandContext(ctx, mysqldumpPath, bm.db.URL)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("zyra/backup: failed to create stdout pipe for mysqldump: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("zyra/backup: failed to start mysqldump process: %w", err)
	}

	if err := dest.WriteBackup(ctx, filename, stdout); err != nil {
		_ = cmd.Process.Kill()
		return err
	}

	return cmd.Wait()
}
