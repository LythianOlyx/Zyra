package migration_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/LythianOlyx/Zyra/internal/data"
	"github.com/LythianOlyx/Zyra/internal/data/migration"
)

func setupTestDB(t *testing.T) *data.DB {
	cfg := data.DatabaseConfig{
		Driver:       "sqlite",
		URL:          "file::memory:?mode=memory&cache=shared",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	}

	db, err := data.Open(cfg)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	return db
}

func TestMigrationLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	testFS := fstest.MapFS{
		"0001_create_posts.up.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT NOT NULL);`),
		},
		"0001_create_posts.down.sql": &fstest.MapFile{
			Data: []byte(`DROP TABLE posts;`),
		},
	}

	// 1. Run AutoMigrate
	if err := migration.AutoMigrate(ctx, db, testFS, "."); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// 2. Check Status
	status, err := migration.Status(ctx, db, testFS, ".")
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status.CurrentVersion != 1 || status.Dirty {
		t.Errorf("expected version=1, dirty=false, got version=%d, dirty=%v", status.CurrentVersion, status.Dirty)
	}

	// Verify table created
	_, err = db.ExecContext(ctx, "INSERT INTO posts (title) VALUES (?)", "Hello Zyra")
	if err != nil {
		t.Fatalf("failed to insert into posts: %v", err)
	}

	// 3. Run Down
	if err := migration.Down(ctx, db, testFS, "."); err != nil {
		t.Fatalf("Down failed: %v", err)
	}

	// Verify table dropped
	_, err = db.ExecContext(ctx, "INSERT INTO posts (title) VALUES (?)", "Hello Zyra")
	if err == nil {
		t.Errorf("expected error inserting into dropped posts table")
	}
}

func TestAdvisoryLockWrapper(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	executed := false

	err := migration.WithAdvisoryLock(ctx, db, func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Fatalf("WithAdvisoryLock failed: %v", err)
	}
	if !executed {
		t.Fatalf("expected fn inside advisory lock to execute")
	}
}
