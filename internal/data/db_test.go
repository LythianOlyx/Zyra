package data_test

import (
	"context"
	"errors"
	"testing"

	"github.com/zyra-framework/zyra/internal/data"
)

type User struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

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

	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	return db
}

func TestDBOperations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	res, err := db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	id, _ := res.LastInsertId()

	var u User
	if err := db.GetContext(ctx, &u, "SELECT * FROM users WHERE id = ?", id); err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if u.Name != "Alice" {
		t.Errorf("expected name Alice, got %s", u.Name)
	}

	var users []User
	if err := db.SelectContext(ctx, &users, "SELECT * FROM users"); err != nil {
		t.Fatalf("failed to select users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestTransactionCommit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	err := db.Transaction(ctx, func(tx *data.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Bob")
		return err
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	var u User
	if err := db.GetContext(ctx, &u, "SELECT * FROM users WHERE name = ?", "Bob"); err != nil {
		t.Fatalf("expected Bob in database, but got err: %v", err)
	}
}

func TestTransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	dummyErr := errors.New("something went wrong")
	err := db.Transaction(ctx, func(tx *data.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Charlie")
		if err != nil {
			return err
		}
		return dummyErr // Return error to trigger rollback
	})

	if !errors.Is(err, dummyErr) {
		t.Fatalf("expected dummyErr, got %v", err)
	}

	var count int
	if err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE name = ?", "Charlie"); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected Charlie to be rolled back, but found count=%d", count)
	}
}

func TestTransactionPanicRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic in transaction test")
		}

		var count int
		if err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE name = ?", "Dave"); err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if count != 0 {
			t.Errorf("expected Dave to be rolled back on panic, but found count=%d", count)
		}
	}()

	_ = db.Transaction(ctx, func(tx *data.Tx) error {
		_, _ = tx.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Dave")
		panic("unexpected crisis")
	})
}
