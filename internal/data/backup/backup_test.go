package backup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zyra-framework/zyra/internal/data"
	"github.com/zyra-framework/zyra/internal/data/backup"
)

func TestSQLiteOnlineBackup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zyra_backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbFile := filepath.Join(tempDir, "original.db")
	cfg := data.DatabaseConfig{
		Driver:       "sqlite",
		URL:          dbFile,
		MaxOpenConns: 1,
	}

	db, err := data.Open(cfg)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	_, err = db.ExecContext(ctx, "CREATE TABLE items (id INT PRIMARY KEY, name TEXT);")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO items VALUES (1, 'Widget A'), (2, 'Widget B');")
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	backupDir := filepath.Join(tempDir, "backups")
	fileDest := &backup.FileDestination{Dir: backupDir}

	bm := backup.NewBackupManager(db)
	if err := bm.Run(ctx, fileDest); err != nil {
		t.Fatalf("Backup.Run failed: %v", err)
	}

	// Verify backup file created in backupDir
	entries, err := os.ReadDir(backupDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 backup file in destination, got %d (err: %v)", len(entries), err)
	}

	// Verify restored backup DB contents
	backupFilePath := filepath.Join(backupDir, entries[0].Name())
	backupDB, err := data.Open(data.DatabaseConfig{Driver: "sqlite", URL: backupFilePath})
	if err != nil {
		t.Fatalf("failed to open backup database: %v", err)
	}
	defer backupDB.Close()

	var count int
	if err := backupDB.GetContext(ctx, &count, "SELECT COUNT(*) FROM items"); err != nil {
		t.Fatalf("failed to query count from backup DB: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 items in backup DB, got %d", count)
	}
}
